package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/RenanMonteiroS/MaestroSQLWeb/controller"
	"github.com/RenanMonteiroS/MaestroSQLWeb/repository"
	"github.com/RenanMonteiroS/MaestroSQLWeb/service"
	"github.com/gin-gonic/gin"
)

func main() {
	server := gin.Default()
	baseFolder, err := os.Getwd()
	if err != nil {
		fmt.Println("Erro ao encontrar o caminho base")
	}

	server.LoadHTMLGlob(filepath.Join(baseFolder, "templates", "*"))

	DatabaseRepository := repository.NewDatabaseRepository(nil)
	DatabaseService := service.NewDatabaseService(DatabaseRepository)
	DatabaseController := controller.NewDatabaseController(DatabaseService)

	server.POST("/connect", DatabaseController.ConnectDatabase)

	server.GET("/databases", DatabaseController.GetDatabases)

	server.POST("/backup", DatabaseController.BackupDatabase)

	server.POST("/restore", DatabaseController.RestoreDatabase)

	server.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "backupForm.html", gin.H{})
	})

	go openFile("http://localhost:8000/")
	server.Run(":8000")

}

func openFile(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}
