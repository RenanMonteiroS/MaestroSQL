package service

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
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

// Makes a HTTP request to the authenticator's /isValid endpoint.
// If the JWT into the Authorization header is not valid, it returns an error. Else, it returns nil
func (ds *DatabaseService) IsAuth(authorization *[]string) error {
	type ResponseBody struct {
		Msg    string `json:"msg"`
		Status string `json:"status"`
	}

	client := &http.Client{Timeout: 10 * time.Second}

	var responseBody ResponseBody

	if config.UseAuthentication != true {
		return nil
	}

	if len(*authorization) == 0 {
		return errors.New("Authorization header not setted")
	}

	auth := *authorization
	req, err := http.NewRequest("GET", fmt.Sprintf("%v/isValid", config.AuthenticatorURL), nil)
	req.Header.Set("Authorization", auth[0])
	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		err := json.NewDecoder(res.Body).Decode(&responseBody)
		if err != nil {
			return err
		}

		return errors.New(fmt.Sprintf("Authentication failed: %v", responseBody.Msg))
	}

	return nil

}

// Establish a connection with a database.
// Args: connInfo -> A struct with connection params (host, port, user, password)
func (ds *DatabaseService) ConnectDatabase(connInfo model.ConnInfo) (*sql.DB, error) {
	conn, err := ds.repository.ConnectDatabase(connInfo)
	if err != nil {
		f, errFile := os.OpenFile("fatal.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if errFile != nil {
			fmt.Println("Error: ", errFile)
			return nil, errFile
		}
		defer f.Close()
		log.SetOutput(f)
		log.Printf("Error: %v: \n", err)

		return nil, err
	}

	return conn, nil
}

// Checks if the connection poll is set and running
func (ds *DatabaseService) CheckDbConn() error {
	err := ds.repository.CheckDbConn()
	if err != nil {
		f, errFile := os.OpenFile("fatal.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if errFile != nil {
			fmt.Println("Error: ", errFile)
			return errFile
		}
		defer f.Close()
		log.SetOutput(f)
		log.Printf("Error: %v: \n", err)
		return err
	}

	return nil
}

// Gets all server databases
// Before it calls the repository.GetDatabases() function, it checks if the connection is set.
func (ds *DatabaseService) GetDatabases() ([]model.Database, error) {
	err := ds.CheckDbConn()
	if err != nil {
		return []model.Database{}, err
	}

	dbListAux, err := ds.repository.GetDatabases()
	if err != nil {
		f, errFile := os.OpenFile("fatal.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if errFile != nil {
			fmt.Println("Error: ", errFile)
			return nil, errFile
		}
		defer f.Close()
		log.SetOutput(f)
		log.Printf("Error: %v: \n", err)
		return nil, err
	}

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
		return []model.Database{}, []error{err}
	}

	f, err := os.OpenFile("backup.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, []error{err}
	}
	defer f.Close()
	log.SetOutput(f)
	log.Printf("-------------------//-------------------//-------------------//-------------------")

	backupDbDoneList, errBackup := ds.repository.BackupDatabase(backupDbList, backupPath)
	if errBackup != nil {
		return backupDbDoneList, errBackup
	}

	return backupDbDoneList, nil
}

// Starts the backup, for each database selected, storing into the backup path choosed.
// Before it calls the repository.RestoreDatabase() function, it checks if the connection is set, gets the backup file data, mounts the database object and gets the default data files path
func (ds *DatabaseService) RestoreDatabase(backupFilesPath string) ([]model.RestoreDb, []error, error) {
	err := ds.CheckDbConn()
	if err != nil {
		return []model.RestoreDb{}, []error{}, errors.New("Connection was not set. Try to call /connect with the connection parameters")
	}

	var backupsFullPathList []string

	var database model.RestoreDb
	var restoreDatabaseList []model.RestoreDb
	var databaseFile model.DatabaseFile

	f, err := os.OpenFile("restore.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	log.SetOutput(f)
	log.Printf("-------------------//-------------------//-------------------//-------------------")

	dir, err := os.ReadDir(backupFilesPath)
	if err != nil {
		return nil, nil, err
	}

	for _, file := range dir {
		if filepath.Ext(file.Name()) == ".bak" {
			backupsFullPathList = append(backupsFullPathList, fmt.Sprintf("%s%s", backupFilesPath, file.Name()))
		} else if filepath.Ext(file.Name()) == ".BAK" {
			backupsFullPathList = append(backupsFullPathList, fmt.Sprintf("%s%s", backupFilesPath, strings.Split(file.Name(), ".BAK")[0])+".bak")
		}
	}

	backupFilesData, err := ds.repository.GetBackupFilesData(backupsFullPathList)
	if err != nil {
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
		return nil, nil, err
	}

	restoredDatabases, errRestoreList := ds.repository.RestoreDatabase(restoreDatabaseList, dataPath, logPath)
	if errRestoreList != nil {
		return restoredDatabases, errRestoreList, nil
	}

	return restoredDatabases, nil, nil

}
