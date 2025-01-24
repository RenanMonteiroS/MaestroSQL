package service

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/RenanMonteiroS/MaestroSQLWeb/model"
	"github.com/RenanMonteiroS/MaestroSQLWeb/repository"
)

type DatabaseService struct {
	repository repository.DatabaseRepository
}

func NewDatabaseService(rp repository.DatabaseRepository) DatabaseService {
	return DatabaseService{repository: rp}
}

func (ds *DatabaseService) GetDatabases() ([]model.Database, error) {
	dbListAux, err := ds.repository.GetDatabases()
	if err != nil {
		return []model.Database{}, err
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

	return dbList, nil
}

func (ds *DatabaseService) BackupDatabase(backupDbList []model.Database, backupPath string) ([]model.Database, error) {
	backupDbList, err := ds.repository.BackupDatabase(backupDbList, backupPath)
	if err != nil {
		return []model.Database{}, err
	}

	return backupDbList, nil
}

func (ds *DatabaseService) RestoreDatabase(backupFilesPath string) ([]model.Database, error) {
	var backupsFullPathList []string

	var restoreDatabase model.Database
	var restoreDatabaseList []model.Database
	var databaseFile model.DatabaseFile

	dir, err := os.ReadDir(backupFilesPath)
	if err != nil {
		return []model.Database{}, err
	}

	for _, file := range dir {
		if filepath.Ext(file.Name()) == ".bak" || filepath.Ext(file.Name()) == ".BAK" {
			backupsFullPathList = append(backupsFullPathList, fmt.Sprintf("%s%s", backupFilesPath, file.Name()))
		}
	}

	backupFilesData, err := ds.repository.GetBackupFilesData(backupsFullPathList)
	if err != nil {
		return []model.Database{}, err
	}

	for _, backupFileData := range backupFilesData {
		fmt.Println(backupFileData)
		restoreDatabase.Name = backupFileData.Name
		for _, backupFileInfo := range backupFileData.BackupFileInfo {
			if backupFileInfo.FileType == "D" {
				databaseFile.FileType = "ROWS"
				databaseFile.LogicalName = backupFileInfo.LogicalName
				restoreDatabase.Files = append(restoreDatabase.Files, databaseFile)
			} else if backupFileInfo.FileType == "L" {
				databaseFile.FileType = "LOG"
				databaseFile.LogicalName = backupFileInfo.LogicalName

				restoreDatabase.Files = append(restoreDatabase.Files, databaseFile)
			}
		}
		restoreDatabaseList = append(restoreDatabaseList, restoreDatabase)
	}

	dataPath, logPath, err := ds.repository.GetDefaultFilesPath()
	if err != nil {
		fmt.Println(err)
		return []model.Database{}, err
	}

	restoredDatabases, err := ds.repository.RestoreDatabase(restoreDatabaseList, backupFilesPath, dataPath, logPath)
	if err != nil {
		return []model.Database{}, err
	}

	return restoredDatabases, nil

}
