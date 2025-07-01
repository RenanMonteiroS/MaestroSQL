package service

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/RenanMonteiroS/MaestroSQLWeb/config"
	"github.com/RenanMonteiroS/MaestroSQLWeb/model"
	"github.com/RenanMonteiroS/MaestroSQLWeb/repository"
)

// Struct responsible for manage authentication logic, business rules, data transformation and logs. Requires a DatabaseRepository.
// Related to database objects
type DatabaseService struct {
	repository repository.DatabaseRepository
}

// Creates an instance of DatabaseService struct
func NewDatabaseService(rp repository.DatabaseRepository) DatabaseService {
	return DatabaseService{repository: rp}
}

// ErrPortAndInstanceEmpty is returned when both instance and port are empty.
var (
	ErrPortAndInstanceEmpty = errors.New("Instance and port are both empty")
)

// Makes a HTTP request to the authenticator's /isValid endpoint.
// If the JWT into the Authorization header is not valid, it returns an error. Else, it returns nil
func (ds *DatabaseService) IsAuth(authorization *[]string) error {
	type ResponseBody struct {
		Msg    string `json:"msg"`
		Status string `json:"status"`
	}

	client := &http.Client{Timeout: 10 * time.Second}

	var responseBody ResponseBody

	if config.AuthenticatorUsage != true {
		return nil
	}

	if len(*authorization) == 0 {
		slog.Error("Authorization header not setted. Try to login into the system.")
		return errors.New("Authorization header not setted")
	}

	auth := *authorization
	req, err := http.NewRequest("GET", fmt.Sprintf("%v/isValid", config.AuthenticatorURL), nil)
	req.Header.Set("Authorization", auth[0])
	req.Header.Set("Content-Type", "application/json")

	slog.Info("Trying to connect to: ", "URL: ", fmt.Sprintf("%v/isValid", config.AuthenticatorURL))
	res, err := client.Do(req)
	if err != nil {
		slog.Error("Cannot connect to authenticator: ", "Error: ", err)
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		err := json.NewDecoder(res.Body).Decode(&responseBody)
		if err != nil {
			slog.Error("Cannot decode response body: ", "Error: ", err)
			return err
		}

		slog.Error("Authentication failed: ", "Error: ", responseBody.Msg)
		return errors.New(fmt.Sprintf("Authentication failed: %v", responseBody.Msg))
	}

	slog.Info("Authentication completed sucessfully")
	return nil

}

// Establish a connection with a database.
// Args: connInfo -> A struct with connection params (host, port, user, password)
func (ds *DatabaseService) ConnectDatabase(connInfo model.ConnInfo) (*sql.DB, error) {
	if connInfo.Port == "" && connInfo.Instance == "" {
		slog.Error("Cannot connect to database: ", "Error: ", ErrPortAndInstanceEmpty)
		return nil, ErrPortAndInstanceEmpty
	}

	conn, err := ds.repository.ConnectDatabase(connInfo)
	if err != nil {
		slog.Error("Cannot connect to database: ", "Error: ", err)
		return nil, err
	}

	return conn, nil
}

// Checks if the connection poll is set and running
func (ds *DatabaseService) CheckDbConn() error {
	err := ds.repository.CheckDbConn()
	if err != nil {
		slog.Error("Cannot connect to database: ", "Error: ", err)
		return err
	}

	return nil
}

// Gets all server databases
// Before it calls the repository.GetDatabases() function, it checks if the connection is set.
func (ds *DatabaseService) GetDatabases() ([]model.Database, error) {
	err := ds.CheckDbConn()
	if err != nil {
		slog.Error("Cannot connect to database: ", "Error: ", err)
		return []model.Database{}, err
	}

	slog.Info("Getting databases...")
	dbListAux, err := ds.repository.GetDatabases()
	if err != nil {
		slog.Error("Cannot get databases: ", "Error: ", err)
		return nil, err
	}
	slog.Info("Databases read successfully")
	var dbObj model.Database
	var dbFile model.DatabaseFile
	var dbList []model.Database
	var found bool

	for _, dbData := range dbListAux {
		dbObj = model.Database{}
		dbObj.ID = dbData.DatabaseId
		dbObj.Name = dbData.DatabaseName

		for key, dbListData := range dbList {
			if dbListData.ID == dbData.DatabaseId {
				found = true
				dbFile.LogicalName = dbData.LogicalName
				dbFile.PhysicalName = dbData.PhysicalName
				dbFile.FileType = dbData.FileType
				dbList[key].Files = append(dbList[key].Files, dbFile)
			}
		}
		if found != true {
			dbFile.LogicalName = dbData.LogicalName
			dbFile.PhysicalName = dbData.PhysicalName
			dbFile.FileType = dbData.FileType
			dbObj.Files = append(dbObj.Files, dbFile)
			dbList = append(dbList, dbObj)
		}
		found = false
	}

	return dbList, nil
}

// Starts the backup, for each database selected, storing into the backup path choosed.
// Before it calls the repository.BackupDatabase() function, it checks if the connection is set.
func (ds *DatabaseService) BackupDatabase(backupDbList []model.Database, backupPath string) ([]model.Database, []error) {
	err := ds.CheckDbConn()
	if err != nil {
		slog.Error("Cannot connect to database: ", "Error: ", err)
		return []model.Database{}, []error{err}
	}

	slog.Info("Starting backup...", "Databases: ", backupDbList, "Backup path: ", backupPath)
	backupDbDoneList, errBackup := ds.repository.BackupDatabase(backupDbList, backupPath)
	if errBackup != nil {
		slog.Warn("Backup completed with errors: ", "Completed backups: ", backupDbDoneList, "Errors: ", errBackup)
		return backupDbDoneList, errBackup
	}

	slog.Info("Backup completed sucessfully: ", "Completed backups: ", backupDbDoneList)
	return backupDbDoneList, nil
}

// Starts the backup, for each database selected, storing into the backup path choosed.
// Before it calls the repository.RestoreDatabase() function, it checks if the connection is set, gets the backup file data, mounts the database object and gets the default data files path
func (ds *DatabaseService) RestoreDatabase(backupFilesPath string) ([]model.RestoreDb, []error, error) {
	err := ds.CheckDbConn()
	if err != nil {
		slog.Error("Cannot connect to database: ", "Error: ", err)
		return []model.RestoreDb{}, []error{}, errors.New("Connection was not set. Try to call /connect with the connection parameters")
	}

	var backupsFullPathList []string

	var database model.RestoreDb
	var restoreDatabaseList []model.RestoreDb
	var databaseFile model.DatabaseFile

	// Gets the dir where the backup files are allocated
	dir, err := os.ReadDir(backupFilesPath)
	if err != nil {
		slog.Error("Cannot get the directory: ", "Error: ", err)
		return nil, nil, err
	}

	// For each file with ".bak" extension, inserts the file into a list
	for _, file := range dir {
		if filepath.Ext(file.Name()) == ".bak" {
			backupsFullPathList = append(backupsFullPathList, fmt.Sprintf("%s%s", backupFilesPath, file.Name()))
		} else if filepath.Ext(file.Name()) == ".BAK" {
			backupsFullPathList = append(backupsFullPathList, fmt.Sprintf("%s%s", backupFilesPath, strings.Split(file.Name(), ".BAK")[0])+".bak")
		}
	}

	slog.Info("backup files read successfully: ", "Files: ", "backupsFullPathList")

	backupFilesData, err := ds.repository.GetBackupFilesData(backupsFullPathList)
	if err != nil {
		slog.Warn("Cannot get backup files data (RESTORE FILELISTONLY): ", "Error: ", err)
		return nil, nil, err
	}

	for _, backupFileData := range backupFilesData {
		database.Database.Name = strings.Split(backupFileData.Name, ".bak")[0]
		database.BackupPath = backupFileData.BackupFilePath

		for _, backupFileInfo := range backupFileData.BackupFileInfo {
			if backupFileInfo.FileType == "D" {
				databaseFile.FileType = "ROWS"
				databaseFile.LogicalName = backupFileInfo.LogicalName
				databaseFile.PhysicalName = backupFileInfo.PhysicalName

				database.Database.Files = append(database.Database.Files, databaseFile)
			} else if backupFileInfo.FileType == "L" {
				databaseFile.FileType = "LOG"
				databaseFile.LogicalName = backupFileInfo.LogicalName
				databaseFile.PhysicalName = backupFileInfo.PhysicalName

				database.Database.Files = append(database.Database.Files, databaseFile)
			}

		}

		if len(restoreDatabaseList) > 0 {
			if database.Database.Name == restoreDatabaseList[len(restoreDatabaseList)-1].Database.Name {
				restoreDatabaseList[len(restoreDatabaseList)-1] = database
			} else {
				restoreDatabaseList = append(restoreDatabaseList, database)
			}
		} else {
			restoreDatabaseList = append(restoreDatabaseList, database)
		}

		database = model.RestoreDb{}
	}

	dataPath, logPath, err := ds.repository.GetDefaultFilesPath()
	if err != nil {
		slog.Error("Cannot get default files path: ", "Error: ", err)
		return nil, nil, err
	}

	slog.Info("Starting restore...", "Databases: ", restoreDatabaseList, "Data path: ", dataPath, "Log path:", "Log path")
	restoredDatabases, errRestoreList := ds.repository.RestoreDatabase(restoreDatabaseList, dataPath, logPath)
	if errRestoreList != nil {
		slog.Warn("Restore completed with errors: ", "Completed restores: ", restoredDatabases, "Errors: ", errRestoreList)
		return restoredDatabases, errRestoreList, nil
	}

	slog.Info("Restore completed sucessfully: ", "Completed restores: ", restoredDatabases)
	return restoredDatabases, nil, nil

}
