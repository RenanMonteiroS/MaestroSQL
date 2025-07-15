package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/RenanMonteiroS/MaestroSQLWeb/db"
	"github.com/RenanMonteiroS/MaestroSQLWeb/model"
)

// Struct responsible for manage database access, like SELECT, BACKUP and RESTORE statements. Requires a sql connection pool object [sql.DB]
// Related to database objects
type DatabaseRepository struct {
	connection *sql.DB
}

// Creates an instance of DatabaseRepository struct
func NewDatabaseRepository(connection *sql.DB) DatabaseRepository {
	return DatabaseRepository{
		connection: connection,
	}
}

// Establish a connection with a database.
// Args: connInfo -> A struct with connection params (host, port, user, password)
func (ds *DatabaseRepository) ConnectDatabase(connInfo model.ConnInfo) (*sql.DB, error) {
	conn, err := db.ConnDb(connInfo)
	if err != nil {
		return nil, err
	}

	ds.connection = conn

	return conn, nil
}

// Checks if the connection poll is set and running
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

// Performs a SELECT in [master] database, to list all server databases
func (dr *DatabaseRepository) GetDatabases() ([]model.MergedDatabaseFileInfo, error) {
	query := "SELECT d.database_id, d.name DatabaseName, " +
		"f.name LogicalName, f.physical_name AS PhysicalName, f.type_desc TypeofFile " +
		"FROM sys.master_files f " +
		"INNER JOIN sys.databases " +
		"d ON d.database_id = f.database_id ORDER BY d.name;"

	rows, err := dr.connection.Query(query)
	if err != nil {
		return nil, err
	}

	var dbObjAux model.MergedDatabaseFileInfo
	var dbListAux []model.MergedDatabaseFileInfo

	for rows.Next() {
		err = rows.Scan(&dbObjAux.DatabaseId, &dbObjAux.DatabaseName, &dbObjAux.LogicalName,
			&dbObjAux.PhysicalName, &dbObjAux.FileType)

		if err != nil {
			return nil, err
		}

		dbListAux = append(dbListAux, dbObjAux)
	}

	return dbListAux, nil
}

// Performs a BACKUP DATABASE statement, for each database selected, storing into the backup path choosed.
// The BACKUP DATABASE statements are executed in goroutines, which makes them concurrent
func (dr *DatabaseRepository) BackupDatabase(backupDbList []model.Database, backupPath string) ([]model.Database, []model.SqlErr) {
	t0 := time.Now()
	type BackupResult struct {
		Database model.Database
		Error    error
		Success  bool
	}

	var wg sync.WaitGroup
	resultCh := make(chan BackupResult, len(backupDbList))

	var dbDoneList []model.Database
	var errorsList []model.SqlErr

	backupLogFile, err := os.OpenFile("backup.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		slog.Error("Cannot open backup log file: ", "Error: ", err)
	}
	defer backupLogFile.Close()

	backupLogger := slog.New(slog.NewJSONHandler(backupLogFile, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}))

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
				backupLogger.Error("Error preparing BACKUP query: ", "Query: ", query, "Error: ", err)
				resultCh <- BackupResult{db, err, false}
				return
			}

			_, err = stmt.ExecContext(ctx, sql.Named("Path", path))
			if err != nil {
				backupLogger.Error("Error executing BACKUP query: ", "Query: ", query, "Error: ", err)
				resultCh <- BackupResult{db, err, false}
				return
			}

			backupLogger.Info(fmt.Sprintf("Backup related to [%v] database completed", db.Name), "Database:", db.Name)
			resultCh <- BackupResult{db, nil, true}

			return
		}(&wg)

	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	results := make([]BackupResult, len(backupDbList))
	for result := range resultCh {
		results = append(results, result)
	}

	for _, result := range results {
		if result.Success {
			dbDoneList = append(dbDoneList, result.Database)
		} else {
			sqlErr := model.NewSqlErr(result.Database.Name, result.Error)
			if sqlErr != nil {
				errorsList = append(errorsList, *sqlErr)
			}
		}
	}

	backupLogger.Info(fmt.Sprintf("Total Time: %v", time.Since(t0)))
	backupLogger.Info(fmt.Sprintf("Path: %v", backupPath))
	backupLogger.Info(fmt.Sprintf("Total Backups: %v", len(dbDoneList)))

	return dbDoneList, errorsList

}

// Performs a RESTORE DATABASE statement, for all backup files inside the backup path.
// The database name is based on the backup file name, as well as the name of the database files (.mdf, .ldf, .ndf)
// The RESTORE DATABASE statements are executed in goroutines, which makes them concurrent
func (dr *DatabaseRepository) RestoreDatabase(restoreDbList []model.RestoreDb, dataPath string, logPath string) ([]model.RestoreDb, []model.SqlErr) {
	t0 := time.Now()

	var restoreDoneDbList []model.RestoreDb
	var errorsList []model.SqlErr

	type RestoreResult struct {
		Database model.RestoreDb
		Error    error
		Success  bool
	}

	var wg sync.WaitGroup
	resultCh := make(chan RestoreResult, len(restoreDbList))

	restoreLogFile, err := os.OpenFile("restore.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		slog.Error("Cannot open restore log file: ", "Error: ", err)
	}
	defer restoreLogFile.Close()

	restoreLogger := slog.New(slog.NewJSONHandler(restoreLogFile, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}))

	for i, db := range restoreDbList {
		wg.Add(1)
		go func(db model.RestoreDb, index int) {
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
				restoreLogger.Error("Error preparing RESTORE query: ", "Query: ", query, "Error: ", err)
				resultCh <- RestoreResult{Database: db, Error: err, Success: false}
				return
			}
			_, err = stmt.ExecContext(ctx, sql.Named("Path", db.BackupPath))
			if err != nil {
				restoreLogger.Error("Error executing RESTORE query: ", "Query: ", query, "Error: ", err)
				resultCh <- RestoreResult{Database: db, Error: err, Success: false}
				return
			}

			restoreLogger.Info(fmt.Sprintf("Restore related to [%v] database completed", db.Database.Name), "Database:", db.Database.Name)
			resultCh <- RestoreResult{Database: db, Error: nil, Success: true}

			return

		}(db, i)

	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	results := make([]RestoreResult, len(restoreDbList))
	for result := range resultCh {
		results = append(results, result)
	}

	for _, result := range results {
		if result.Success {
			restoreDoneDbList = append(restoreDoneDbList, result.Database)
		} else {
			sqlErr := model.NewSqlErr(result.Database.Database.Name, result.Error)
			if sqlErr != nil {
				errorsList = append(errorsList, *sqlErr)
			}
		}
	}

	restoreLogger.Info(fmt.Sprintf("Total Time: %v", time.Since(t0)))
	restoreLogger.Info(fmt.Sprintf("Total Restores: %v", len(restoreDoneDbList)))

	return restoreDoneDbList, errorsList

}

// Gets the default data path and log path, set as a server property
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

// Performs a RESTORE FILELISTONLY for each backup file. It gets all the information about the related database backup file.
// RESTORE FILELISTONLY is necessary because if RESTORE DATABASE is run without setting the name and location of the database files,
// it will restore the database using the previous data. Therefore, if the database was previously located in /var/opt/mssql/,
// even if the restore is being performed on a Windows server, it will attempt to restore the files in /var/opt/mssql/. Also, RESTORE DATABASE expects the original
// logical name of the database file. That's when RESTORE FILELISTONLY helps.
func (dr *DatabaseRepository) GetBackupFilesData(backupFiles []string) ([]model.DatabaseFromBackupFile, error) {
	var restoreDatabaseInfo model.BackupDataFile
	var restoreDatabaseInfoList []model.DatabaseFromBackupFile

	var restoreDatabase model.DatabaseFromBackupFile

	for _, backupFile := range backupFiles {
		query := "RESTORE FILELISTONLY FROM DISK = @Path;"

		stmt, err := dr.connection.Prepare(query)
		if err != nil {
			slog.Error("Error preparing RESTORE FILELISTONLY query: ", "Query: ", query, "Error: ", err)
			return nil, err
		}

		rows, err := stmt.Query(sql.Named("Path", backupFile))
		if err != nil {
			slog.Error("Error executing RESTORE FILELISTONLY query: ", "Query: ", query, "Error: ", err)
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
