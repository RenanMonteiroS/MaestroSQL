package main

import (
	"embed"
	"html/template"
	"net/http"
	"os/exec"
	"runtime"

	"github.com/RenanMonteiroS/MaestroSQLWeb/config"
	"github.com/RenanMonteiroS/MaestroSQLWeb/controller"
	"github.com/RenanMonteiroS/MaestroSQLWeb/repository"
	"github.com/RenanMonteiroS/MaestroSQLWeb/service"
	"github.com/gin-gonic/gin"
)

// Embeds HTML templates. Allows files within the templates folder to be referenced without having them on the computer when the application is compiled.

//go:embed templates/*
var TemplateFS embed.FS

func main() {
	// Create an instance of the Gin Server Engine to be runned
	server := gin.Default()

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

	// Initialize the "/" HTTP route, serving the HTML template file, providing some variables to the template
	server.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "backupForm.html", gin.H{
			"useAuth":          config.UseAuthentication,
			"authenticatorURL": config.AuthenticatorURL,
		})
	})

	// Opens the URL in the browser and starts the server
	go openFile("http://localhost:8000/")
	server.Run(":8000")
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
