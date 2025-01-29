package main

import (
	"fmt"

	"github.com/RenanMonteiroS/MaestroSQLWeb/controller"
	"github.com/RenanMonteiroS/MaestroSQLWeb/db"
	"github.com/RenanMonteiroS/MaestroSQLWeb/repository"
	"github.com/RenanMonteiroS/MaestroSQLWeb/service"
	"github.com/gin-gonic/gin"
)

func main() {
	server := gin.Default()

	conn, err := db.ConnDb()
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}
	defer conn.Close()

	DatabaseRepository := repository.NewDatabaseRepository(conn)
	DatabaseService := service.NewDatabaseService(DatabaseRepository)
	DatabaseController := controller.NewDatabaseController(DatabaseService)

	server.GET("/databases", DatabaseController.GetDatabases)

	server.POST("/backup", DatabaseController.BackupDatabase)

	server.POST("/restore", DatabaseController.RestoreDatabase)

	server.Run(":8000")

}
