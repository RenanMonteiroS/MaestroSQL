package controller

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/RenanMonteiroS/MaestroSQLWeb/model"
	"github.com/RenanMonteiroS/MaestroSQLWeb/service"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Struct responsible for handle the HTTP requests. Requires a DatabaseService.
// Related to database objects
type DatabaseController struct {
	service service.DatabaseService
}

// Creates an instance of DatabaseController struct
func NewDatabaseController(sv service.DatabaseService) DatabaseController {
	return DatabaseController{service: sv}
}

// Handles the POST /connect endpoint.
// Starts a connection pool for a database server instance. For each request, it checks if the user is authenticated.
func (dc *DatabaseController) ConnectDatabase(ctx *gin.Context) {
	var connInfo model.ConnInfo
	err := ctx.BindJSON(&connInfo)
	if err != nil {
		slog.Error("Cannot bind JSON from request body", "Origin", ctx.ClientIP(), "User", sessions.Default(ctx).Get("userEmail"), "Error", err.Error())
		ctx.JSON(http.StatusInternalServerError, model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "Cannot bind JSON from request body", Errors: map[string]any{"bindJson": err.Error()}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Request.URL.Path})
		return
	}

	_, err = dc.service.ConnectDatabase(connInfo)
	if err != nil {
		if errors.Is(err, service.ErrPortAndInstanceEmpty) {
			slog.Error("Cannot connect to database", "Origin", ctx.ClientIP(), "User", sessions.Default(ctx).Get("userEmail"), "Error", err.Error())
			ctx.JSON(http.StatusBadRequest, model.APIResponse{Status: "error", Code: http.StatusBadRequest, Message: "Cannot connect to the server - Port and Instance empty", Errors: map[string]any{"connect": err.Error()}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Request.URL.Path})
		} else {
			slog.Error("Cannot connect to database", "Origin", ctx.ClientIP(), "User", sessions.Default(ctx).Get("userEmail"), "Error", err.Error())
			ctx.JSON(http.StatusInternalServerError, model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "Cannot connect to the server", Errors: map[string]any{"connect": err.Error()}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Request.URL.Path})
		}
		return
	}

	slog.Info("Connection done successfully", "Origin", ctx.ClientIP(), "User", sessions.Default(ctx).Get("userEmail"), "Host", connInfo.Host)
	ctx.JSON(http.StatusOK, model.APIResponse{Status: "success", Code: http.StatusOK, Message: "Connection done successfully", Data: map[string]any{"server": connInfo.Host}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Request.URL.Path})
	return
}

// Handles the GET /databases endpoint.
// Gets all databases allocated in the database server instance. For each request, it checks if the user is authenticated.
func (dc *DatabaseController) GetDatabases(ctx *gin.Context) {
	databases, err := dc.service.GetDatabases()
	if err != nil {
		slog.Error("Cannot get databases", "Origin", ctx.ClientIP(), "User", sessions.Default(ctx).Get("userEmail"), "Error", err.Error())
		ctx.JSON(http.StatusInternalServerError, model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "Cannot get databases", Errors: map[string]any{"databases": err.Error()}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Request.URL.Path})
		return
	}

	slog.Info("Databases collected successfully", "Origin", ctx.ClientIP(), "User", sessions.Default(ctx).Get("userEmail"))
	ctx.JSON(http.StatusOK, model.APIResponse{Status: "success", Code: 200, Message: "Databases collected successfully", Data: map[string]any{"databases": databases}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Request.URL.Path})
	return
}

// Handles the POST /backup endpoint.
// Calls the backup functions. For each request, it checks if the user is authenticated.
func (dc *DatabaseController) BackupDatabase(ctx *gin.Context) {
	type BackupPostRequired struct {
		Databases     []model.Database `json:"databases" binding:"required"`
		Path          string           `json:"path" binding:"required"`
		ConcurrentOpe *int             `json:"concurrentOpe,omitempty"`
	}

	var postData BackupPostRequired

	err := ctx.BindJSON(&postData)
	if err != nil {
		slog.Error("Cannot bind JSON from request body", "Origin", ctx.ClientIP(), "User", sessions.Default(ctx).Get("userEmail"), "Error", err.Error())
		ctx.JSON(http.StatusInternalServerError, model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "Cannot bind JSON from request body", Errors: map[string]any{"bindJSON": err.Error()}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Request.URL.Path})
		return
	}

	databaseBackupList, errBackup, err, totalTime := dc.service.BackupDatabase(postData.Databases, postData.Path, postData.ConcurrentOpe)
	if err != nil {
		slog.Error("No backup was completed", "Origin", ctx.ClientIP(), "User", sessions.Default(ctx).Get("userEmail"), "Error", err)
		ctx.JSON(http.StatusInternalServerError, model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "No backup was completed", Errors: map[string]any{"connect": err.Error()}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Request.URL.Path})
		return
	}

	if errBackup != nil && len(databaseBackupList) != 0 {
		slog.Error("Backup completed with errors.", "Origin", ctx.ClientIP(), "User", sessions.Default(ctx).Get("userEmail"), "Error", errBackup)
		ctx.JSON(http.StatusMultiStatus, model.APIResponse{Status: "error", Code: http.StatusMultiStatus, Message: "Backup completed with errors.", Data: map[string]any{"backupDone": databaseBackupList, "totalTime": totalTime, "backupPath": postData.Path, "totalBackup": len(databaseBackupList)}, Errors: map[string]any{"backupErrors": errBackup, "totalBackupErrors": len(errBackup)}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Request.URL.Path})
		return
	} else if errBackup != nil && len(databaseBackupList) == 0 {
		var errStringList []string
		for _, value := range errBackup {
			errStringList = append(errStringList, value.Error())
		}

		slog.Error("No backup was completed.", "Origin", ctx.ClientIP(), "User", sessions.Default(ctx).Get("userEmail"), "Error", errBackup)
		ctx.JSON(http.StatusInternalServerError, model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "No backup was completed.", Errors: map[string]any{"backupErrors": errBackup, "totalBackupErrors": len(errBackup)}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Request.URL.Path})
		return
	}

	slog.Info("Backup done successfully.", "Origin", ctx.ClientIP(), "User", sessions.Default(ctx).Get("userEmail"))
	ctx.JSON(http.StatusOK, model.APIResponse{Status: "success", Code: http.StatusOK, Message: "Backup done successfully.", Data: map[string]any{"backupDone": databaseBackupList, "backupPath": postData.Path, "totalBackup": len(databaseBackupList), "totalTime": totalTime}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Request.URL.Path})
	return
}

// Handles the POST /restore endpoint.
// Calls the restore functions. For each request, it checks if the user is authenticated.
func (dc *DatabaseController) RestoreDatabase(ctx *gin.Context) {
	type RestorePostRequired struct {
		Path          string `json:"backupFilesPath" binding:"required"`
		ConcurrentOpe *int   `json:"concurrentOpe,omitempty"`
	}

	var postData RestorePostRequired

	err := ctx.BindJSON(&postData)
	if err != nil {
		slog.Error("Cannot bind JSON from request body", "Origin", ctx.ClientIP(), "User", sessions.Default(ctx).Get("userEmail"), "Error", err.Error())
		ctx.JSON(http.StatusInternalServerError, model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "Cannot bind JSON from request body", Errors: map[string]any{"bindJSON": err.Error()}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Request.URL.Path})
		return
	}

	restoredDatabases, errRestore, err, totalTime := dc.service.RestoreDatabase(postData.Path, postData.ConcurrentOpe)
	if err != nil {
		slog.Error("No restore was completed", "Origin", ctx.ClientIP(), "User", sessions.Default(ctx).Get("userEmail"), "Error", err.Error())
		ctx.JSON(http.StatusInternalServerError, model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "Restore operation error", Errors: map[string]any{"restore": err.Error()}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Request.URL.Path})
		return
	}

	if errRestore != nil && len(restoredDatabases) > 0 {
		slog.Error("Restore operation completed with errors", "Origin", ctx.ClientIP(), "User", sessions.Default(ctx).Get("userEmail"), "Error", errRestore)
		ctx.JSON(http.StatusMultiStatus, model.APIResponse{Status: "error", Code: http.StatusMultiStatus, Message: "Restore operation completed with errors", Data: map[string]any{"restoreDone": restoredDatabases, "backupPath": postData.Path, "totalRestore": len(restoredDatabases), "totalTime": totalTime}, Errors: map[string]any{"restoreErrors": errRestore, "totalRestoreErrors": len(errRestore)}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Request.URL.Path})
		return
	} else if errRestore != nil && len(restoredDatabases) == 0 {
		slog.Error("No restore was completed.", "Origin", ctx.ClientIP(), "User", sessions.Default(ctx).Get("userEmail"), "Error", errRestore)
		ctx.JSON(http.StatusInternalServerError, model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "No restore was completed.", Errors: map[string]any{"restoreErrors": errRestore, "totalRestoreErrors": len(errRestore)}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Request.URL.Path})
		return
	}

	slog.Info("Restore done successfully.", "Origin", ctx.ClientIP(), "User", sessions.Default(ctx).Get("userEmail"))
	ctx.JSON(http.StatusOK, model.APIResponse{Status: "success", Code: http.StatusOK, Message: "Restore done successfully.", Data: map[string]any{"restoreDone": restoredDatabases, "backupPath": postData.Path, "totalRestore": len(restoredDatabases), "totalTime": totalTime}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Request.URL.Path})
	return
}
