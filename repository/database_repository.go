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

func (dr *DatabaseRepository) GetDatabases() (*[]model.MergedDatabaseFileInfo, error) {
	query := "SELECT d.database_id, d.name DatabaseName, " +
		"f.name LogicalName, f.physical_name AS PhysicalName, f.type_desc TypeofFile " +
		"FROM sys.master_files f " +
		"INNER JOIN sys.databases " +
		"d ON d.database_id = f.database_id;"

	rows, err := dr.connection.Query(query)
	if err != nil {
		return nil, err
	}

	var dbObjAux model.MergedDatabaseFileInfo
	var dbListAux []model.MergedDatabaseFileInfo

	for rows.Next() {
		err = rows.Scan(&dbObjAux.DatabaseId, &dbObjAux.DatabaseName, &dbObjAux.LogicalName,
			&dbObjAux.PhysicalName, &dbObjAux.File_type)

		if err != nil {
			return nil, err
		}

		dbListAux = append(dbListAux, dbObjAux)
	}

	return &dbListAux, nil
}

func (dr *DatabaseRepository) BackupDatabase(backupDbList *[]model.Database, backupPath string) (*[]model.Database, error) {

	var wg sync.WaitGroup
	ch := make(chan model.Database)

	var dbDoneList []model.Database

	for _, db := range *backupDbList {
		wg.Add(1)
		go func(db *model.Database, connection *sql.DB, wg *sync.WaitGroup, ch chan model.Database) {
			defer wg.Done()

			var query string
			query = fmt.Sprintf("BACKUP DATABASE %s TO DISK = '%s/%s=%v_%v.bak'; ", db.Name, backupPath, db.Name, time.Now().Format("2006-01-02"), time.Now().Format("15-04-05"))

			_, err := connection.Query(query)
			if err != nil {
				return
			}

			ch <- *db

			return
		}(&db, dr.connection, &wg, ch)

	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for db := range ch {
		dbDoneList = append(dbDoneList, db)
	}

	return &dbDoneList, nil

}

func (dr *DatabaseRepository) RestoreDatabase(restoreDbList *[]model.RestoreDb, dataPath string, logPath string) (*[]model.RestoreDb, error) {

	var restoreDoneDbList []model.RestoreDb

	var wg sync.WaitGroup
	ch := make(chan model.RestoreDb)

	for _, db := range *restoreDbList {
		wg.Add(1)
		go func(db model.RestoreDb, wg *sync.WaitGroup, ch chan model.RestoreDb) {
			defer wg.Done()

			var query string

			query = fmt.Sprintf("RESTORE DATABASE [%s] FROM DISK = '%s' WITH ", db.Database.Name, db.BackupPath)
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
			_, err := dr.connection.Query(query)
			if err != nil {
				return
			}

			ch <- db

			return

		}(db, &wg, ch)

	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for doneDb := range ch {
		restoreDoneDbList = append(restoreDoneDbList, doneDb)
	}

	return &restoreDoneDbList, nil

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

func (dr *DatabaseRepository) GetBackupFilesData(backupFiles *[]string) (*[]model.DatabaseFromBackupFile, error) {
	var query string

	var restoreDatabaseInfo model.BackupDataFile
	var restoreDatabaseInfoList []model.DatabaseFromBackupFile

	var restoreDatabase model.DatabaseFromBackupFile

	for _, backupFile := range *backupFiles {
		query = fmt.Sprintf("RESTORE FILELISTONLY FROM DISK = '%s'; ", backupFile)

		rows, err := dr.connection.Query(query)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			err = rows.Scan(&restoreDatabaseInfo.LogicalName, &restoreDatabaseInfo.PhysicalName, &restoreDatabaseInfo.FileType, &restoreDatabaseInfo.FileGroupName,
				&restoreDatabaseInfo.Size, &restoreDatabaseInfo.MaxSize, &restoreDatabaseInfo.FileId, &restoreDatabaseInfo.CreateLSN, &restoreDatabaseInfo.DropLSN,
				&restoreDatabaseInfo.UniqueId, &restoreDatabaseInfo.ReadOnlyLSN, &restoreDatabaseInfo.ReadWriteLSN, &restoreDatabaseInfo.BackupSizeInBytes,
				&restoreDatabaseInfo.SourceBlockSize, &restoreDatabaseInfo.FileGroupId, &restoreDatabaseInfo.LogGroupGUID, &restoreDatabaseInfo.DifferentialBaseLSN,
				&restoreDatabaseInfo.DifferentialBaseGUID, &restoreDatabaseInfo.IsReadOnly, &restoreDatabaseInfo.IsPresent, &restoreDatabaseInfo.TDEThumbprint,
				&restoreDatabaseInfo.SnapshotUrl)
			if err != nil {
				return nil, err
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

	return &restoreDatabaseInfoList, nil

}
