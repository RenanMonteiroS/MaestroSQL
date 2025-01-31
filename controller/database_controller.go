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

func (dc *DatabaseController) GetDatabases(ctx *gin.Context) {
	databases, err := dc.service.GetDatabases()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, databases)
	return
}

func (dc *DatabaseController) BackupDatabase(ctx *gin.Context) {

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

	databaseBackupList, errBackup := dc.service.BackupDatabase(postData.Databases, postData.Path)
	if errBackup != nil && databaseBackupList != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]any{"msg": "Completed with errors.", "backupErrors": errBackup, "backupCompleted": databaseBackupList})
		return
	} else if errBackup != nil && databaseBackupList == nil {
		ctx.JSON(http.StatusInternalServerError, map[string]any{"msg": "No backup was completed.", "backupErrors": errBackup})
		return
	}

	ctx.JSON(http.StatusOK, databaseBackupList)
	return
}

func (dc *DatabaseController) RestoreDatabase(ctx *gin.Context) {

	var backupFiles model.BackupFiles

	err := ctx.BindJSON(&backupFiles)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	restoredDatabases, errRestore, err := dc.service.RestoreDatabase(backupFiles.Path)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}
	if errRestore != nil && restoredDatabases != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]any{"msg": "Completed with errors.", "restoreErrors": errRestore, "restoreCompleted": restoredDatabases})
	} else if errRestore != nil && restoredDatabases == nil {
		ctx.JSON(http.StatusInternalServerError, map[string]any{"msg": "No restore was completed.", "restoreErrors": errRestore})
	}

	ctx.JSON(http.StatusOK, restoredDatabases)
	return
}
