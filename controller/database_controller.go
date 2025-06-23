package controller

import (
	"fmt"
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

func (dc *DatabaseController) ConnectDatabase(ctx *gin.Context) {
	authorization := ctx.Request.Header["Authorization"]

	err := dc.service.IsAuth(&authorization)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, map[string]any{"msg": err.Error()})
		return
	}

	var connInfo model.ConnInfo
	err = ctx.BindJSON(&connInfo)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]any{"msg": err.Error()})
		return
	}

	conn, err := dc.service.ConnectDatabase(connInfo)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]any{"msg": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, conn)
	return
}

func (dc *DatabaseController) GetDatabases(ctx *gin.Context) {
	authorization := ctx.Request.Header["Authorization"]

	err := dc.service.IsAuth(&authorization)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, map[string]any{"msg": err.Error()})
		return
	}

	databases, err := dc.service.GetDatabases()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]any{"msg": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, databases)
	return
}

func (dc *DatabaseController) BackupDatabase(ctx *gin.Context) {
	authorization := ctx.Request.Header["Authorization"]

	err := dc.service.IsAuth(&authorization)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, map[string]any{"msg": err.Error()})
		return
	}

	type PostRequired struct {
		Databases []model.Database `json:"databases" binding:"required"`
		Path      string           `json:"path" binding:"required"`
	}

	var postData PostRequired

	err = ctx.BindJSON(&postData)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	databaseBackupList, errBackup := dc.service.BackupDatabase(postData.Databases, postData.Path)
	if errBackup != nil && len(databaseBackupList) != 0 {
		ctx.JSON(http.StatusInternalServerError, map[string]any{"msg": "Completed with errors.", "backupErrors": errBackup, "backupCompleted": databaseBackupList})
		return
	} else if errBackup != nil && len(databaseBackupList) == 0 {
		var errStringList []string
		for _, value := range errBackup {
			errStringList = append(errStringList, value.Error())
		}

		ctx.JSON(http.StatusInternalServerError, map[string]any{"msg": "No backup was completed.", "backupErrors": errStringList})
		return
	}

	ctx.JSON(http.StatusOK, databaseBackupList)
	return
}

func (dc *DatabaseController) RestoreDatabase(ctx *gin.Context) {
	authorization := ctx.Request.Header["Authorization"]

	err := dc.service.IsAuth(&authorization)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, map[string]any{"msg": err.Error()})
		return
	}

	var backupFiles model.BackupFiles

	err = ctx.BindJSON(&backupFiles)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	restoredDatabases, errRestore, err := dc.service.RestoreDatabase(backupFiles.Path)
	if err != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusInternalServerError, map[string]any{"msg": "No restore was completed.", "restoreErrors": err.Error()})
		return
	}

	if errRestore != nil && len(restoredDatabases) > 0 {
		ctx.JSON(http.StatusInternalServerError, map[string]any{"msg": "Completed with errors.", "restoreErrors": errRestore, "restoreCompleted": restoredDatabases})
		return
	} else if errRestore != nil && len(restoredDatabases) == 0 {
		ctx.JSON(http.StatusInternalServerError, map[string]any{"msg": "No restore was completed.", "restoreErrors": errRestore})
		return
	}

	ctx.JSON(http.StatusOK, restoredDatabases)
	return
}
