package controller

import (
	"net/http"

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
