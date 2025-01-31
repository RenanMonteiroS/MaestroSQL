package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/RenanMonteiroS/MaestroSQLWeb/model"
	"github.com/RenanMonteiroS/MaestroSQLWeb/repository"
)

type DatabaseService struct {
	repository repository.DatabaseRepository
}

func NewDatabaseService(rp repository.DatabaseRepository) DatabaseService {
	return DatabaseService{repository: rp}
}

func (ds *DatabaseService) GetDatabases() (*[]model.Database, error) {
	dbListAux, err := ds.repository.GetDatabases()
	if err != nil {
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
				dbFile.FileType = dbData.File_type
				dbList[key].Files = append(dbList[key].Files, dbFile)
			}
		}
		if found != true {
			dbFile.LogicalName = dbData.LogicalName
			dbFile.PhysicalName = dbData.PhysicalName
			dbFile.FileType = dbData.File_type
			dbObj.Files = append(dbObj.Files, dbFile)
			dbList = append(dbList, dbObj)
		}
		found = false
	}

	return &dbList, nil
}

func (ds *DatabaseService) BackupDatabase(backupDbList []model.Database, backupPath string) ([]model.Database, error) {
	backupDbDoneList, err := ds.repository.BackupDatabase(backupDbList, backupPath)
	if err != nil {
		return nil, err
	}

	return backupDbDoneList, nil
}

func (ds *DatabaseService) RestoreDatabase(backupFilesPath string) ([]model.RestoreDb, error) {
	var backupsFullPathList []string

	var database model.RestoreDb
	var restoreDatabaseList []model.RestoreDb
	var databaseFile model.DatabaseFile

	dir, err := os.ReadDir(backupFilesPath)
	if err != nil {
		return nil, err
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
		return nil, err
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
		return nil, err
	}

	restoredDatabases, err := ds.repository.RestoreDatabase(restoreDatabaseList, dataPath, logPath)
	if err != nil {
		return nil, err
	}

	return restoredDatabases, nil

}
