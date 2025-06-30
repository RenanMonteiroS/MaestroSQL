package main

import (
	"embed"
	"fmt"
	"html/template"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/RenanMonteiroS/MaestroSQLWeb/config"
	"github.com/RenanMonteiroS/MaestroSQLWeb/controller"
	"github.com/RenanMonteiroS/MaestroSQLWeb/repository"
	"github.com/RenanMonteiroS/MaestroSQLWeb/service"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"
)

// Embeds HTML templates. Allows files within the templates folder to be referenced without having them on the computer when the application is compiled.

//go:embed templates/*
var TemplateFS embed.FS

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
	var serverProtocol string

	// Create an instance of the Gin Server Engine to be runned
	server := gin.Default()

	if config.AppCSRFTokenUsage {
		store := cookie.NewStore([]byte(config.AppCSRFCookieSecret))
		store.Options(sessions.Options{
			Path:     "/",
			MaxAge:   86400,
			HttpOnly: true,
			Secure:   config.AppCertificateUsage,
			SameSite: http.SameSiteLaxMode,
		})

		server.Use(sessions.Sessions("maestro-sessions", store))
		server.Use(csrf.Middleware(csrf.Options{
			Secret: config.AppCSRFTokenSecret,
			ErrorFunc: func(c *gin.Context) {
				slog.Error("CSRF token mismatch")
				c.JSON(400, gin.H{"msg": "CSRF token mismatch"})
				c.Abort()
			},
		}))
	}

	// Creates the templates that will be served. It uses template.ParseFS to read the system's fs instead of the OS's fs. The embed templates will be read
	tmpl := template.Must(template.ParseFS(TemplateFS, "templates/*"))
	// Set the created templates in the Gin Server Engine
	server.SetHTMLTemplate(tmpl)

	// Initialize the layers instances
	DatabaseRepository := repository.NewDatabaseRepository(nil)
	DatabaseService := service.NewDatabaseService(DatabaseRepository)
	DatabaseController := controller.NewDatabaseController(DatabaseService)

	// Initialize the HTTP routes
	server.POST("/connect", DatabaseController.ConnectDatabase)
	server.GET("/databases", DatabaseController.GetDatabases)
	server.POST("/backup", DatabaseController.BackupDatabase)
	server.POST("/restore", DatabaseController.RestoreDatabase)

	if config.AppCertificateUsage {
		serverProtocol = "https"
	} else {
		serverProtocol = "http"
	}

	// Opens the URL in the browser and starts the server
	if config.AppHost == "0.0.0.0" {
		localIP := getOutboundIP()
		serverIP = localIP
		serverAddr = fmt.Sprintf(config.AppHost + ":" + fmt.Sprint(config.AppPort))
		go openFile(fmt.Sprintf("%v://%v:%v/", serverProtocol, localIP, config.AppPort))
	} else {
		go openFile(fmt.Sprintf("%v://%v/", serverProtocol, serverAddr))
	}

	// Initialize the "/" HTTP route, serving the HTML template file, providing some variables to the template
	server.GET("/", func(c *gin.Context) {
		var varToServe gin.H
		if config.AppCSRFTokenUsage {
			varToServe = gin.H{
				"appHost":             serverIP,
				"appPort":             config.AppPort,
				"appCertificateUsage": config.AppCertificateUsage,
				"appCSRFTokenUsage":   config.AppCSRFTokenUsage,
				"authenticatorUsage":  config.AuthenticatorUsage,
				"authenticatorURL":    config.AuthenticatorURL,
				"csrfToken":           csrf.GetToken(c),
			}
		} else {
			varToServe = gin.H{
				"appHost":             serverIP,
				"appPort":             config.AppPort,
				"appCertificateUsage": config.AppCertificateUsage,
				"appCSRFTokenUsage":   config.AppCSRFTokenUsage,
				"authenticatorUsage":  config.AuthenticatorUsage,
				"authenticatorURL":    config.AuthenticatorURL,
			}
		}
		c.HTML(http.StatusOK, "backupForm.html", varToServe)
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
