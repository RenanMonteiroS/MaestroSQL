package controller

import (
	"net/http"

	"github.com/RenanMonteiroS/MaestroSQLWeb/model"
	"github.com/RenanMonteiroS/MaestroSQLWeb/service"
	"github.com/gin-gonic/gin"
)

type DatabaseController struct {
	service service.DatabaseService
}

func NewDatabaseController(sv service.DatabaseService) DatabaseController {
	return DatabaseController{service: sv}
}

func (dc DatabaseController) GetDatabases(ctx *gin.Context) {
	databases, err := dc.service.GetDatabases()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, databases)
	return
}

func (dc DatabaseController) BackupDatabase(ctx *gin.Context) {

	type PostRequired struct {
		Databases []model.Database `json:"databases" binding:"required"`
		Path      string           `json:"path" binding:"required"`
	}

	var postData PostRequired

	err := ctx.BindJSON(&postData)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	databaseBackupList, err := dc.service.BackupDatabase(postData.Databases, postData.Path)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, databaseBackupList)
	return
}
