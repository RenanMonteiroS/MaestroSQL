package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/RenanMonteiroS/MaestroSQLWeb/db"
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

func (ds *DatabaseRepository) ConnectDatabase(connInfo model.ConnInfo) (*sql.DB, error) {
	conn, err := db.ConnDb(connInfo)
	if err != nil {
		return nil, err
	}

	ds.connection = conn

	return conn, nil
}

func (ds *DatabaseRepository) CheckDbConn() error {
	if ds.connection == nil {
		return errors.New("Connection was not set. Try to call /connect with the connection parameters")
	}

	err := ds.connection.Ping()
	if err != nil {
		return err
	}

	return nil

}

func (dr *DatabaseRepository) GetDatabases() ([]model.MergedDatabaseFileInfo, error) {
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

	return dbListAux, nil
}

func (dr *DatabaseRepository) BackupDatabase(backupDbList []model.Database, backupPath string) ([]model.Database, []error) {
	if dr.connection == nil {
		return []model.Database{}, []error{errors.New("Connection was not set. Try to call /connect with the connection parameters")}
	}

	var wg sync.WaitGroup
	ch := make(chan model.Database, len(backupDbList))
	chError := make(chan error, len(backupDbList))

	var dbDoneList []model.Database
	var errorsList []error

	for _, db := range backupDbList {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()

			query := fmt.Sprintf("BACKUP DATABASE [%s] TO DISK = @Path", db.Name)
			path := fmt.Sprintf("%s/%s=%v_%v.bak", backupPath, db.Name, time.Now().Format("2006-01-02"), time.Now().Format("15-04-05"))

			stmt, err := dr.connection.Prepare(query)
			if err != nil {
				chError <- err
				return
			}

			_, err = stmt.ExecContext(ctx, sql.Named("Path", path))
			if err != nil {
				chError <- err
				return
			}

			ch <- db

			return
		}(&wg)

	}

	go func() {
		wg.Wait()
		close(ch)
		close(chError)
	}()

	for db := range ch {
		dbDoneList = append(dbDoneList, db)
	}

	for err := range chError {
		errorsList = append(errorsList, err)
	}

	return dbDoneList, errorsList

}

func (dr *DatabaseRepository) RestoreDatabase(restoreDbList []model.RestoreDb, dataPath string, logPath string) ([]model.RestoreDb, []error) {

	var restoreDoneDbList []model.RestoreDb
	var errorsList []error

	var wg sync.WaitGroup
	ch := make(chan model.RestoreDb, len(restoreDbList))
	chError := make(chan error, len(restoreDbList))

	for _, db := range restoreDbList {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
			defer cancel()

			query := fmt.Sprintf("RESTORE DATABASE [%s] FROM DISK = @Path WITH ", db.Database.Name)
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

			stmt, err := dr.connection.Prepare(query)
			if err != nil {
				chError <- err
				return
			}
			_, err = stmt.ExecContext(ctx, sql.Named("Path", db.BackupPath))
			if err != nil {
				chError <- err
				return
			}

			ch <- db

			return

		}(&wg)

	}

	go func() {
		wg.Wait()
		close(ch)
		close(chError)
	}()

	for doneDb := range ch {
		restoreDoneDbList = append(restoreDoneDbList, doneDb)
	}

	for err := range chError {
		errorsList = append(errorsList, err)
	}

	return restoreDoneDbList, errorsList

}

func (dr *DatabaseRepository) GetDefaultFilesPath() (string, string, error) {
	var dataPath, logPath string

	query := "SELECT SERVERPROPERTY(@Prop);"

	stmt, err := dr.connection.Prepare(query)
	if err != nil {
		return "", "", err
	}

	err = stmt.QueryRow(sql.Named("Prop", "instancedefaultdatapath")).Scan(&dataPath)
	if err != nil {
		return "", "", err
	}

	query = "SELECT SERVERPROPERTY(@Prop);"

	stmt, err = dr.connection.Prepare(query)
	if err != nil {
		return "", "", err
	}

	err = stmt.QueryRow(sql.Named("Prop", "instancedefaultlogpath")).Scan(&logPath)
	if err != nil {
		return "", "", err
	}

	return dataPath, logPath, nil
}

func (dr *DatabaseRepository) GetBackupFilesData(backupFiles []string) ([]model.DatabaseFromBackupFile, error) {
	var restoreDatabaseInfo model.BackupDataFile
	var restoreDatabaseInfoList []model.DatabaseFromBackupFile

	var restoreDatabase model.DatabaseFromBackupFile

	for _, backupFile := range backupFiles {
		query := "RESTORE FILELISTONLY FROM DISK = @Path;"

		stmt, err := dr.connection.Prepare(query)
		if err != nil {
			return nil, err
		}

		rows, err := stmt.Query(sql.Named("Path", backupFile))
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

	return restoreDatabaseInfoList, nil

}
