package repository

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/RenanMonteiroS/MaestroSQLWeb/model"
)

type DatabaseRepository struct {
	connection *sql.DB
}

func NewDatabaseRepository(connection *sql.DB) DatabaseRepository {
	return DatabaseRepository{
		connection: connection,
	}
}

func (dr *DatabaseRepository) GetDatabases() ([]model.MergedDatabaseFileInfo, error) {
	query := "SELECT d.database_id, d.name DatabaseName, " +
		"f.name LogicalName, f.physical_name AS PhysicalName, f.type_desc TypeofFile " +
		"FROM sys.master_files f " +
		"INNER JOIN sys.databases " +
		"d ON d.database_id = f.database_id;"

	rows, err := dr.connection.Query(query)
	if err != nil {
		return []model.MergedDatabaseFileInfo{}, err
	}

	var dbObjAux model.MergedDatabaseFileInfo
	var dbListAux []model.MergedDatabaseFileInfo

	for rows.Next() {
		err = rows.Scan(&dbObjAux.DatabaseId, &dbObjAux.DatabaseName, &dbObjAux.LogicalName,
			&dbObjAux.PhysicalName, &dbObjAux.File_type)

		if err != nil {
			return []model.MergedDatabaseFileInfo{}, err
		}

		dbListAux = append(dbListAux, dbObjAux)
	}

	return dbListAux, nil
}

func (dr *DatabaseRepository) BackupDatabase(backupDbList []model.Database, backupPath string) ([]model.Database, error) {
	var query string

	//var dbDoneList []model.Database

	var wg sync.WaitGroup
	//chann := make(chan model.Database)
	//var channList []chan model.Database

	var channList []chan string

	for _, db := range backupDbList {
		wg.Add(1)
		go func(wg *sync.WaitGroup, db *model.Database, channList *[]chan string) {
			defer wg.Done()

			chann := make(chan string)

			query = fmt.Sprintf("BACKUP DATABASE %s TO DISK = '%s/%s=%v_%v.bak'; ", db.Name, backupPath, db.Name, time.Now().Format("2006-01-02"), time.Now().Format("15-04-05"))
			fmt.Println("Teste")

			_, err := dr.connection.Query(query)
			if err != nil {
				return
			}
			fmt.Println("Teste2")

			//chann <- *db
			chann <- "a"
			fmt.Println("Teste3")
			*channList = append(*channList, chann)
			fmt.Println("Teste4")

			return
		}(&wg, &db, &channList)

	}

	wg.Wait()

	/* for key, chann := range channList {
		dbDoneList[key] = <- chann
	} */

	return backupDbList, nil

}

func (dr *DatabaseRepository) RestoreDatabase(restoreDbList []model.RestoreDb, dataPath string, logPath string) ([]model.RestoreDb, error) {
	var query string

	for _, db := range restoreDbList {

		query += fmt.Sprintf("RESTORE DATABASE [%s] FROM DISK = '%s' WITH ", db.Database.Name, db.BackupPath)
		for _, file := range db.Database.Files {
			if file.FileType == "ROWS" {
				if strings.Contains(file.PhysicalName, ".mdf") {
					query += fmt.Sprintf("MOVE '%s' TO '%s%s.mdf' , ", file.LogicalName, dataPath, db.Database.Name)
				} else if strings.Contains(file.PhysicalName, ".ndf") {
					query += fmt.Sprintf("MOVE '%s' TO '%s%s.ndf' , ", file.LogicalName, dataPath, db.Database.Name)
				}

			} else if file.FileType == "LOG" {
				query += fmt.Sprintf("MOVE '%s' TO '%s%s.ldf' , ", file.LogicalName, logPath, db.Database.Name)
			}
		}
		query += "RECOVERY;"
	}

	_, err := dr.connection.Query(query)
	fmt.Println(query)
	if err != nil {
		return []model.RestoreDb{}, err
	}

	return restoreDbList, nil

}

func (dr *DatabaseRepository) GetDefaultFilesPath() (string, string, error) {
	var dataPath, logPath string

	query := "SELECT SERVERPROPERTY('instancedefaultdatapath');"

	err := dr.connection.QueryRow(query).Scan(&dataPath)
	if err != nil {
		return "", "", err
	}

	query = "SELECT SERVERPROPERTY('instancedefaultlogpath');"

	err = dr.connection.QueryRow(query).Scan(&logPath)
	if err != nil {
		return "", "", err
	}

	return dataPath, logPath, nil
}

func (dr *DatabaseRepository) GetBackupFilesData(backupFiles []string) ([]model.DatabaseFromBackupFile, error) {
	var query string

	var restoreDatabaseInfo model.BackupDataFile
	var restoreDatabaseInfoList []model.DatabaseFromBackupFile

	var restoreDatabase model.DatabaseFromBackupFile

	for _, backupFile := range backupFiles {
		query = fmt.Sprintf("RESTORE FILELISTONLY FROM DISK = '%s'; ", backupFile)

		rows, err := dr.connection.Query(query)
		if err != nil {
			return []model.DatabaseFromBackupFile{}, err
		}

		for rows.Next() {
			err = rows.Scan(&restoreDatabaseInfo.LogicalName, &restoreDatabaseInfo.PhysicalName, &restoreDatabaseInfo.FileType, &restoreDatabaseInfo.FileGroupName,
				&restoreDatabaseInfo.Size, &restoreDatabaseInfo.MaxSize, &restoreDatabaseInfo.FileId, &restoreDatabaseInfo.CreateLSN, &restoreDatabaseInfo.DropLSN,
				&restoreDatabaseInfo.UniqueId, &restoreDatabaseInfo.ReadOnlyLSN, &restoreDatabaseInfo.ReadWriteLSN, &restoreDatabaseInfo.BackupSizeInBytes,
				&restoreDatabaseInfo.SourceBlockSize, &restoreDatabaseInfo.FileGroupId, &restoreDatabaseInfo.LogGroupGUID, &restoreDatabaseInfo.DifferentialBaseLSN,
				&restoreDatabaseInfo.DifferentialBaseGUID, &restoreDatabaseInfo.IsReadOnly, &restoreDatabaseInfo.IsPresent, &restoreDatabaseInfo.TDEThumbprint,
				&restoreDatabaseInfo.SnapshotUrl)
			if err != nil {
				return []model.DatabaseFromBackupFile{}, err
			}
			restoreDatabase.Name = filepath.Base(backupFile)
			restoreDatabase.BackupFilePath = backupFile

			restoreDatabase.BackupFileInfo = append(restoreDatabase.BackupFileInfo, restoreDatabaseInfo)

			if len(restoreDatabaseInfoList) > 0 {
				if restoreDatabaseInfoList[len(restoreDatabaseInfoList)-1].Name == restoreDatabase.Name {
					restoreDatabaseInfoList[len(restoreDatabaseInfoList)-1] = restoreDatabase
				} else {
					restoreDatabaseInfoList = append(restoreDatabaseInfoList, restoreDatabase)
				}
			} else {
				restoreDatabaseInfoList = append(restoreDatabaseInfoList, restoreDatabase)
			}

		}
		restoreDatabase = model.DatabaseFromBackupFile{}

	}

	return restoreDatabaseInfoList, nil

}
