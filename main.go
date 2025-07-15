//go:generate go-winres make --in winres.json
package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"strings"

	"github.com/RenanMonteiroS/MaestroSQLWeb/config"
	"github.com/RenanMonteiroS/MaestroSQLWeb/controller"
	"github.com/RenanMonteiroS/MaestroSQLWeb/middleware"
	"github.com/RenanMonteiroS/MaestroSQLWeb/repository"
	"github.com/RenanMonteiroS/MaestroSQLWeb/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	csrf "github.com/utrack/gin-csrf"
	"golang.org/x/text/language"
)

// Embeds HTML templates. Allows files within the templates folder to be referenced without having them on the computer when the application is compiled.

//go:embed templates/*
var TemplateFS embed.FS

//go:embed static
var StaticFS embed.FS

//go:embed locales/*.json
var LocaleFS embed.FS

func main() {
	logFile, err := os.OpenFile("app.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		os.Exit(1)
	}
	defer logFile.Close()

	logger := slog.New(slog.NewJSONHandler(logFile, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	serverIP := config.AppHost
	serverAddr := fmt.Sprintf(config.AppHost + ":" + fmt.Sprint(config.AppPort))
	localIP := getOutboundIP()
	var serverProtocol string

	// Create an instance of the Gin Server Engine to be runned
	server := gin.Default()

	// Configure CORS usage
	if config.CORSUsage {
		server.Use(cors.New(cors.Config{
			AllowOrigins:     config.CORSAllowOrigins,
			AllowMethods:     []string{"GET", "POST"},
			AllowHeaders:     []string{"Content-Type", "Authorization", "Accept-Language", "X-Csrf-Token"},
			AllowCredentials: true,
		}))
	}

	// Starts a Session Cookie Store
	store := cookie.NewStore([]byte(config.AppSessionSecret))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		Secure:   config.AppCertificateUsage,
		SameSite: http.SameSiteLaxMode,
	})
	server.Use(sessions.Sessions("maestro-sessions", store))

	var csrfMiddleware gin.HandlerFunc
	// Configure CSRF token usage
	if config.AppCSRFTokenUsage {
		csrfMiddleware = csrf.Middleware(csrf.Options{
			Secret: config.AppCSRFTokenSecret,
			ErrorFunc: func(c *gin.Context) {
				slog.Error("CSRF token mismatch")
				c.JSON(400, gin.H{"msg": "CSRF token mismatch"})
				c.Abort()
			},
		})
	}

	//server.Static("/static", "./static")

	// Creates the templates that will be served. It uses template.ParseFS to read the system's fs instead of the OS's fs. The embed templates will be read
	tmpl := template.Must(template.ParseFS(TemplateFS, "templates/*"))
	// Set the created templates in the Gin Server Engine
	server.SetHTMLTemplate(tmpl)

	AuthService := service.NewAuthService()
	AuthController := controller.NewAuthController(AuthService)

	// Initialize the layers instances
	DatabaseRepository := repository.NewDatabaseRepository(nil)
	DatabaseService := service.NewDatabaseService(DatabaseRepository)
	DatabaseController := controller.NewDatabaseController(DatabaseService)

	// Initialize the HTTP routes

	// If the app uses CSRF protection, starts a CSRF security middleware into the authentication routes
	authRoutes := server.Group("/")
	if config.AppCSRFTokenUsage {
		authRoutes.Use(csrfMiddleware)
	}
	{
		authRoutes.Any("/login", AuthController.LoginHandler)
		authRoutes.GET("/logout", AuthController.LogoutHandler)
		authRoutes.GET("/session", AuthController.SessionHandler)
	}

	oAuth2Routes := server.Group("/")
	oAuth2Routes.GET("/auth/google/callback", AuthController.GoogleCallBackHandler)
	oAuth2Routes.GET("/auth/microsoft/callback", AuthController.MicrosoftCallBackHandler)

	// If the app uses some sort of authentication, starts a security middleware
	protected := server.Group("/")
	if len(config.AuthenticationMethods) > 0 {
		protected.Use(middleware.AuthMiddleware())
	}

	// If the app uses CSRF protection, starts a CSRF security middleware into the general routes
	if config.AppCSRFTokenUsage {
		protected.Use(csrfMiddleware)
	}
	{
		protected.POST("/connect", DatabaseController.ConnectDatabase)
		protected.GET("/databases", DatabaseController.GetDatabases)
		protected.POST("/backup", DatabaseController.BackupDatabase)
		protected.POST("/restore", DatabaseController.RestoreDatabase)
	}

	if config.AppCertificateUsage {
		serverProtocol = "https"
	} else {
		serverProtocol = "http"
	}

	// Opens the URL in the browser and starts the server
	if config.AppHost == "0.0.0.0" {
		serverIP = localIP
		serverAddr = fmt.Sprintf(config.AppHost + ":" + fmt.Sprint(config.AppPort))
		go openFile(fmt.Sprintf("%v://%v:%v/", serverProtocol, localIP, config.AppPort))
	} else {
		go openFile(fmt.Sprintf("%v://%v/", serverProtocol, serverAddr))
	}

	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	bundle.LoadMessageFileFS(LocaleFS, `locales/en-US.json`)
	bundle.LoadMessageFileFS(LocaleFS, `locales/pt-BR.json`)

	// Starts a middleware to handle multilinguals
	server.Use(func(c *gin.Context) {
		// Gets the Accept Language header
		lang := c.GetHeader("Accept-Language")

		// Starts a new localizer looks up messages in the bundle according to the language preferences in langs
		localizer := i18n.NewLocalizer(bundle, lang)

		// Sets a key-value with the localizer
		c.Set("localizer", localizer)

		c.Next()
	})

	// Create subfilesystem to serve static
	staticSub, err := fs.Sub(StaticFS, "static")
	if err != nil {
		slog.Error("Error creating subfilesystem", "Error", err)
	}
	server.StaticFS("/static", http.FS(staticSub))

	// If the app uses CSRF protection, starts a CSRF security middleware into the "/" route
	if config.AppCSRFTokenUsage {
		server.Use(csrfMiddleware)
	}

	// Initialize the "/" HTTP route, serving the HTML template file, providing some variables to the template
	server.GET("/", func(c *gin.Context) {
		localizer := c.MustGet("localizer").(*i18n.Localizer)
		var varToServe gin.H
		var authenticationUsage bool
		var authenticationOSIUsage bool
		var authenticationGoogleOAuth2Usage bool
		var authenticationMicrosoftOAuth2Usage bool

		if slices.Contains(config.AuthenticationMethods, "OSI") {
			authenticationOSIUsage = true
			authenticationUsage = true
		}

		if slices.Contains(config.AuthenticationMethods, "OAUTH2GOOGLE") {
			authenticationGoogleOAuth2Usage = true
			authenticationUsage = true
		}

		if slices.Contains(config.AuthenticationMethods, "OAUTH2MICROSOFT") {
			authenticationMicrosoftOAuth2Usage = true
			authenticationUsage = true
		}

		varToServe = gin.H{
			"appHost":                            serverIP,
			"appPort":                            config.AppPort,
			"appCertificateUsage":                config.AppCertificateUsage,
			"appCSRFTokenUsage":                  config.AppCSRFTokenUsage,
			"authenticatorURL":                   config.AuthenticatorURL,
			"authenticationUsage":                authenticationUsage,
			"authenticationMethods":              config.AuthenticationMethods,
			"authenticationOSIUsage":             authenticationOSIUsage,
			"authenticationGoogleOAuth2Usage":    authenticationGoogleOAuth2Usage,
			"authenticationMicrosoftOAuth2Usage": authenticationMicrosoftOAuth2Usage,
			"T": func(translationID string) string {
				return localizer.MustLocalize(&i18n.LocalizeConfig{
					MessageID: translationID,
				})
			},
		}

		if config.AppCSRFTokenUsage {
			varToServe["csrfToken"] = csrf.GetToken(c)
		}
		c.HTML(http.StatusOK, "backupForm.html", varToServe)
	})

	// Not found route
	server.NoRoute(func(ctx *gin.Context) {
		localizer := ctx.MustGet("localizer").(*i18n.Localizer)

		// If is an API request...
		if strings.Contains(ctx.GetHeader("Accept"), "application/json") {
			ctx.JSON(http.StatusNotFound, map[string]any{"msg": "Route not found"})
		}

		varsToServe := gin.H{
			"T": func(translationID string) string {
				return localizer.MustLocalize(&i18n.LocalizeConfig{
					MessageID: translationID,
				})
			},
		}

		ctx.HTML(http.StatusNotFound, "404.html", varsToServe)
	})

	fmt.Printf("MaestroSQL started. Your application is running at: http://%v/", serverAddr)
	logger.Info(fmt.Sprintf("MaestroSQL started. Your application is running at: http://%v/", serverAddr))

	if config.AppCertificateUsage {
		server.RunTLS(serverAddr, config.AppCertificateLocation, config.AppCertificateKeyLocation)
	} else {
		server.Run(serverAddr)
	}

}

// Opens the browser to some URL. The command executed depends on the type of operating system.
func openFile(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default:
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

// Checks if there's any IP related do ethernet.
func getOutboundIP() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		slog.Warn("Failed to get network interfaces", "Error", err)
		return "127.0.0.1"
	}

	var potentialIPs []string

	for _, iface := range interfaces {
		// Ignores disabled and loopback interfaces
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// Get IP addresses
		addrs, err := iface.Addrs()
		if err != nil {
			slog.Warn("Failed to get network interfaces", "Interface", iface.Name, "Error", err)
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			if ipnet, ok := addr.(*net.IPNet); ok {
				ip = ipnet.IP
			}

			// Checks if the IP is valid, or not LoopBack
			if ip != nil && !ip.IsLoopback() && ip.To4() != nil {
				// Checks if the interface is related to the ethernet to priorize it
				if strings.Contains(strings.ToLower(iface.Name), "ethernet") {
					return ip.String() // If some Ethernet IP is found, it is returned
				}
				// Collect other potential ethernet IPS
				potentialIPs = append(potentialIPs, ip.String())
			}
		}
	}

	// If no Ethernet IP is found, it returns other potential IP
	if len(potentialIPs) > 0 {
		return potentialIPs[0]
	}

	// If none IP is found, it returns the loopback
	return "127.0.0.1"
}
