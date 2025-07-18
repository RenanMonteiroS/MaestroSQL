package controller

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/RenanMonteiroS/MaestroSQLWeb/model"
	"github.com/RenanMonteiroS/MaestroSQLWeb/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
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
func (dc *DatabaseController) ConnectDatabase(ctx *fiber.Ctx) error {
	var connInfo model.ConnInfo

	sess, ok := ctx.Locals("session").(*session.Session)
	if !ok {
		return ctx.Status(http.StatusInternalServerError).JSON(model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "Internal server error: session not found", Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
	}

	err := ctx.BodyParser(&connInfo)
	if err != nil {
		slog.Error("Cannot bind JSON from request body", "Origin", ctx.IP(), "User", sess.Get("userEmail"), "Error", err.Error())
		return ctx.Status(http.StatusInternalServerError).JSON(model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "Cannot bind JSON from request body", Errors: map[string]any{"bindJson": err.Error()}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
	}

	_, err = dc.service.ConnectDatabase(connInfo)
	if err != nil {
		if errors.Is(err, service.ErrPortAndInstanceEmpty) {
			slog.Error("Cannot connect to database", "Origin", ctx.IP(), "User", sess.Get("userEmail"), "Error", err.Error())
			return ctx.Status(http.StatusBadRequest).JSON(model.APIResponse{Status: "error", Code: http.StatusBadRequest, Message: "Cannot connect to the server - Port and Instance empty", Errors: map[string]any{"connect": err.Error()}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
		} else {
			slog.Error("Cannot connect to database", "Origin", ctx.IP(), "User", sess.Get("userEmail"), "Error", err.Error())
			return ctx.Status(http.StatusInternalServerError).JSON(model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "Cannot connect to the server", Errors: map[string]any{"connect": err.Error()}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
		}
	}

	slog.Info("Connection done successfully", "Origin", ctx.IP(), "User", sess.Get("userEmail"), "Host", connInfo.Host)
	return ctx.Status(http.StatusOK).JSON(model.APIResponse{Status: "success", Code: http.StatusOK, Message: "Connection done successfully", Data: map[string]any{"server": connInfo.Host}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
}

// Handles the GET /databases endpoint.
// Gets all databases allocated in the database server instance. For each request, it checks if the user is authenticated.
func (dc *DatabaseController) GetDatabases(ctx *fiber.Ctx) error {
	sess, ok := ctx.Locals("session").(*session.Session)
	if !ok {
		return ctx.Status(http.StatusInternalServerError).JSON(model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "Internal server error: session not found", Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
	}

	databases, err := dc.service.GetDatabases()
	if err != nil {
		slog.Error("Cannot get databases", "Origin", ctx.IP(), "User", sess.Get("userEmail"), "Error", err.Error())
		return ctx.Status(http.StatusInternalServerError).JSON(model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "Cannot get databases", Errors: map[string]any{"databases": err.Error()}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
	}

	slog.Info("Databases collected successfully", "Origin", ctx.IP(), "User", sess.Get("userEmail"))
	return ctx.Status(http.StatusOK).JSON(model.APIResponse{Status: "success", Code: 200, Message: "Databases collected successfully", Data: map[string]any{"databases": databases}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
}

// Handles the POST /backup endpoint.
// Calls the backup functions. For each request, it checks if the user is authenticated.
func (dc *DatabaseController) BackupDatabase(ctx *fiber.Ctx) error {
	type BackupPostRequired struct {
		Databases     []model.Database `json:"databases" binding:"required"`
		Path          string           `json:"path" binding:"required"`
		ConcurrentOpe *int             `json:"concurrentOpe,omitempty"`
	}

	var postData BackupPostRequired

	sess, ok := ctx.Locals("session").(*session.Session)
	if !ok {
		return ctx.Status(http.StatusInternalServerError).JSON(model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "Internal server error: session not found", Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
	}

	err := ctx.BodyParser(&postData)
	if err != nil {
		slog.Error("Cannot bind JSON from request body", "Origin", ctx.IP(), "User", sess.Get("userEmail"), "Error", err.Error())
		return ctx.Status(http.StatusInternalServerError).JSON(model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "Cannot bind JSON from request body", Errors: map[string]any{"bindJSON": err.Error()}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
	}

	databaseBackupList, errBackup, err, totalTime := dc.service.BackupDatabase(postData.Databases, postData.Path, postData.ConcurrentOpe)
	if err != nil {
		slog.Error("No backup was completed", "Origin", ctx.IP(), "User", sess.Get("userEmail"), "Error", err)
		return ctx.Status(http.StatusInternalServerError).JSON(model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "No backup was completed", Errors: map[string]any{"connect": err.Error()}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
	}

	if errBackup != nil && len(databaseBackupList) != 0 {
		slog.Error("Backup completed with errors.", "Origin", ctx.IP(), "User", sess.Get("userEmail"), "Error", errBackup)
		return ctx.Status(http.StatusMultiStatus).JSON(model.APIResponse{Status: "error", Code: http.StatusMultiStatus, Message: "Backup completed with errors.", Data: map[string]any{"backupDone": databaseBackupList, "totalTime": totalTime, "backupPath": postData.Path, "totalBackup": len(databaseBackupList)}, Errors: map[string]any{"backupErrors": errBackup, "totalBackupErrors": len(errBackup)}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
	} else if errBackup != nil && len(databaseBackupList) == 0 {
		var errStringList []string
		for _, value := range errBackup {
			errStringList = append(errStringList, value.Error())
		}

		slog.Error("No backup was completed.", "Origin", ctx.IP(), "User", sess.Get("userEmail"), "Error", errBackup)
		return ctx.Status(http.StatusInternalServerError).JSON(model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "No backup was completed.", Errors: map[string]any{"backupErrors": errBackup, "totalBackupErrors": len(errBackup)}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
	}

	slog.Info("Backup done successfully.", "Origin", ctx.IP(), "User", sess.Get("userEmail"))
	return ctx.Status(http.StatusOK).JSON(model.APIResponse{Status: "success", Code: http.StatusOK, Message: "Backup done successfully.", Data: map[string]any{"backupDone": databaseBackupList, "backupPath": postData.Path, "totalBackup": len(databaseBackupList), "totalTime": totalTime}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
}

// Handles the POST /restore endpoint.
// Calls the restore functions. For each request, it checks if the user is authenticated.
func (dc *DatabaseController) RestoreDatabase(ctx *fiber.Ctx) error {
	var postData model.RestorePostRequired

	sess, ok := ctx.Locals("session").(*session.Session)
	if !ok {
		return ctx.Status(http.StatusInternalServerError).JSON(model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "Internal server error: session not found", Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
	}

	err := ctx.BodyParser(&postData)
	if err != nil {
		slog.Error("Cannot bind JSON from request body", "Origin", ctx.IP(), "User", sess.Get("userEmail"), "Error", err.Error())
		return ctx.Status(http.StatusInternalServerError).JSON(model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "Cannot bind JSON from request body", Errors: map[string]any{"bindJSON": err.Error()}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
	}

	restoredDatabases, errRestore, err, totalTime := dc.service.RestoreDatabase(postData.Databases, postData.ConcurrentOpe)
	if err != nil {
		slog.Error("No restore was completed", "Origin", ctx.IP(), "User", sess.Get("userEmail"), "Error", err.Error())
		return ctx.Status(http.StatusInternalServerError).JSON(model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "Restore operation error", Errors: map[string]any{"restore": err.Error()}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
	}

	if errRestore != nil && len(restoredDatabases) > 0 {
		slog.Error("Restore operation completed with errors", "Origin", ctx.IP(), "User", sess.Get("userEmail"), "Error", errRestore)
		return ctx.Status(http.StatusMultiStatus).JSON(model.APIResponse{Status: "error", Code: http.StatusMultiStatus, Message: "Restore operation completed with errors", Data: map[string]any{"restoreDone": restoredDatabases, "totalRestore": len(restoredDatabases), "totalTime": totalTime}, Errors: map[string]any{"restoreErrors": errRestore, "totalRestoreErrors": len(errRestore)}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
	} else if errRestore != nil && len(restoredDatabases) == 0 {
		slog.Error("No restore was completed.", "Origin", ctx.IP(), "User", sess.Get("userEmail"), "Error", errRestore)
		return ctx.Status(http.StatusInternalServerError).JSON(model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "No restore was completed.", Errors: map[string]any{"restoreErrors": errRestore, "totalRestoreErrors": len(errRestore)}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
	}

	slog.Info("Restore done successfully.", "Origin", ctx.IP(), "User", sess.Get("userEmail"))
	return ctx.Status(http.StatusOK).JSON(model.APIResponse{Status: "success", Code: http.StatusOK, Message: "Restore done successfully.", Data: map[string]any{"restoreDone": restoredDatabases, "totalRestore": len(restoredDatabases), "totalTime": totalTime}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
}

func (dc *DatabaseController) ListBackups(ctx *fiber.Ctx) error {
	type listBackupPostRequired struct {
		Path string `json:"backupFilesPath" binding:"required"`
	}

	var postData listBackupPostRequired

	sess, ok := ctx.Locals("session").(*session.Session)
	if !ok {
		return ctx.Status(http.StatusInternalServerError).JSON(model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "Internal server error: session not found", Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
	}

	err := ctx.BodyParser(&postData)
	if err != nil {
		slog.Error("Cannot bind JSON from request body", "Origin", ctx.IP(), "User", sess.Get("userEmail"), "Error", err.Error())
		return ctx.Status(http.StatusInternalServerError).JSON(model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "Cannot bind JSON from request body", Errors: map[string]any{"bindJson": err.Error()}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})

	}

	backupFiles, err := dc.service.ListBackupFiles(postData.Path)
	if err != nil {
		slog.Error("Cannot list backup files", "Origin", ctx.IP(), "User", sess.Get("userEmail"), "Error", err.Error())
		return ctx.Status(http.StatusInternalServerError).JSON(model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "Cannot list backup files", Errors: map[string]any{"listBackups": err.Error()}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})

	}

	slog.Info("Backup files listed successfully", "Origin", ctx.IP(), "User", sess.Get("userEmail"))
	return ctx.Status(http.StatusOK).JSON(model.APIResponse{Status: "success", Code: http.StatusOK, Message: "Backup files listed successfully", Data: map[string]any{"backupFiles": backupFiles}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
}
