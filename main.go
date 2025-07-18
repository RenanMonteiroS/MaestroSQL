//go:generate go-winres make --in winres.json
package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/RenanMonteiroS/MaestroSQLWeb/config"
	"github.com/RenanMonteiroS/MaestroSQLWeb/controller"
	"github.com/RenanMonteiroS/MaestroSQLWeb/middleware"
	"github.com/RenanMonteiroS/MaestroSQLWeb/repository"
	"github.com/RenanMonteiroS/MaestroSQLWeb/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/template/html/v2"
	"github.com/gofiber/utils"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// Embeds HTML templates. Allows files within the templates folder to be referenced without having them on the computer when the application is compiled.

//go:embed templates
var TemplateFS embed.FS

//go:embed static
var StaticFS embed.FS

//go:embed locales/*.json
var LocaleFS embed.FS

func main() {
	// Create logs
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

	//Get server network information
	serverIP := config.AppHost
	serverAddr := fmt.Sprintf(config.AppHost + ":" + fmt.Sprint(config.AppPort))
	localIP := getOutboundIP()
	var serverProtocol string

	// Set the created templates in the Fiber Server Engine
	templateSub, err := fs.Sub(TemplateFS, ".")
	if err != nil {
		slog.Error("Error creating subfilesystem for templates", "Error", err)
		os.Exit(1)
	}
	templateEngine := html.NewFileSystem(http.FS(templateSub), ".html")

	// Create an instance of the Fiber Server Engine to be runned
	server := fiber.New(fiber.Config{
		Views: templateEngine,
	})

	// Configure CORS usage
	if config.AppCORSUsage {
		server.Use(cors.New(cors.Config{
			AllowOrigins:     config.AppCORSAllowOrigins,
			AllowMethods:     "GET,POST",
			AllowHeaders:     "Content-Type, Authorization, Accept-Language, X-Csrf-Token",
			AllowCredentials: true,
		}))
	}

	// Starts a new language bundle
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	if _, err := bundle.LoadMessageFileFS(LocaleFS, "locales/en-US.json"); err != nil {
		slog.Error("Failed to load en-US messages", "Error", err)
	}
	if _, err := bundle.LoadMessageFileFS(LocaleFS, "locales/pt-BR.json"); err != nil {
		slog.Error("Failed to load pt-BR messages", "Error", err)
	}

	// Initialize the auth layers instances
	AuthService := service.NewAuthService()
	AuthController := controller.NewAuthController(AuthService)

	// Initialize the database layers instances
	DatabaseRepository := repository.NewDatabaseRepository(nil)
	DatabaseService := service.NewDatabaseService(DatabaseRepository)
	DatabaseController := controller.NewDatabaseController(DatabaseService)

	// Create subfilesystem to serve static
	staticSub, err := fs.Sub(StaticFS, "static")
	if err != nil {
		slog.Error("Error creating subfilesystem", "Error", err)
	}

	// Starts a Session Cookie Store
	store := session.New(session.Config{
		KeyLookup:      "cookie:maestro-sessions",
		CookieDomain:   "",
		CookiePath:     "/",
		CookieSecure:   config.AppCertificateUsage,
		CookieHTTPOnly: true,
		CookieSameSite: "Lax",
		Expiration:     time.Hour * 24,
		KeyGenerator: func() string {
			return utils.UUID()
		},
	})

	//Middlewares session

	//Inserts the session into the app scope
	server.Use(middleware.SessionMiddleware(store))

	// Configure CSRF token usage
	if config.AppCSRFTokenUsage {
		server.Use(middleware.CsrfMiddleware())
	}

	// Starts a middleware to handle multilinguals
	server.Use(middleware.LanguageMiddleware(bundle))

	// HTTP Routes session

	// Auth routes
	authRoutes := server.Group("/")
	{
		authRoutes.All("/login", AuthController.LoginHandler)
		authRoutes.Get("/logout", AuthController.LogoutHandler)
		authRoutes.Get("/session", AuthController.SessionHandler)
	}

	// OAuth2 Routes
	oAuth2Routes := server.Group("/")
	{
		oAuth2Routes.Get("/auth/google/callback", AuthController.GoogleCallBackHandler)
		oAuth2Routes.Get("/auth/microsoft/callback", AuthController.MicrosoftCallBackHandler)
	}

	server.Use("/static", filesystem.New(filesystem.Config{
		Root: http.FS(staticSub),
	}))
	// Initialize the "/" HTTP route, serving the HTML template file, providing some variables to the template
	server.Get("/", func(ctx *fiber.Ctx) error {
		localizer := ctx.Locals("localizer").(*i18n.Localizer)
		var varToServe fiber.Map
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

		varToServe = fiber.Map{
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
			varToServe["csrfToken"] = ctx.Locals("csrf")
		}
		return ctx.Render("backupForm.html", varToServe)
	})

	// If the app uses some sort of authentication, starts a security middleware
	protected := server.Group("/api")
	if len(config.AuthenticationMethods) > 0 {
		protected.Use(middleware.AuthMiddleware())
	}

	{
		protected.Post("/connect", DatabaseController.ConnectDatabase)
		protected.Get("/databases", DatabaseController.GetDatabases)
		protected.Post("/backup", DatabaseController.BackupDatabase)
		protected.Post("/restore", DatabaseController.RestoreDatabase)
		protected.Post("/list-backups", DatabaseController.ListBackups)
	}

	// Not found route
	server.Use(func(ctx *fiber.Ctx) error {
		localizer := ctx.Locals("localizer").(*i18n.Localizer)

		// If is an API request...
		if strings.Contains(ctx.Get("Accept"), "application/json") {
			return ctx.Status(http.StatusNotFound).JSON(map[string]any{"msg": "Route not found"})
		}

		varsToServe := fiber.Map{
			"T": func(translationID string) string {
				return localizer.MustLocalize(&i18n.LocalizeConfig{
					MessageID: translationID,
				})
			},
		}

		return ctx.Status(http.StatusNotFound).Render("404.html", varsToServe)
	})

	// Gets the server network information
	serverIP = localIP
	serverAddr = fmt.Sprintf(config.AppHost + ":" + fmt.Sprint(config.AppPort))

	if config.AppCertificateUsage {
		serverProtocol = "https"
	} else {
		serverProtocol = "http"
	}

	if config.AppOpenOnceRunned {
		// Opens the URL in the browser and starts the server
		if config.AppHost == "0.0.0.0" {
			go openFile(fmt.Sprintf("%v://%v:%v/", serverProtocol, localIP, config.AppPort))
		} else {
			go openFile(fmt.Sprintf("%v://%v/", serverProtocol, serverAddr))
		}
	}

	fmt.Printf("MaestroSQL started. Your application is running at: http://%v/", serverAddr)
	logger.Info(fmt.Sprintf("MaestroSQL started. Your application is running at: http://%v/", serverAddr))

	if config.AppCertificateUsage {
		server.ListenTLS(serverAddr, config.AppCertificateLocation, config.AppCertificateKeyLocation)
	} else {
		server.Listen(serverAddr)
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
