package repository

import (
	"database/sql"
	"fmt"
	"path/filepath"
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

	for _, value := range backupDbList {
		query += fmt.Sprintf("BACKUP DATABASE %s TO DISK = '%s/%s=%v_%v.bak'; ", value.Name, backupPath, value.Name, time.Now().Format("2006-01-02"), time.Now().Format("15-04-05"))
	}

	_, err := dr.connection.Query(query)
	if err != nil {
		return []model.Database{}, err
	}

	return backupDbList, nil

}

func (dr *DatabaseRepository) RestoreDatabase(restoreDbList []model.Database, backupPath string, dataPath string, logPath string) ([]model.Database, error) {
	var query string
	fmt.Println(restoreDbList)
	for _, db := range restoreDbList {
		query += fmt.Sprintf("RESTORE DATABASE [%s] FROM DISK = '%s' WITH ", db.Name, backupPath)
		for _, file := range db.Files {
			if file.FileType == "ROWS" {
				query += fmt.Sprintf("MOVE '%s' TO '%s%s.mdf' , ", file.LogicalName, dataPath, db.Name)
			} else if file.FileType == "LOG" {
				query += fmt.Sprintf("MOVE '%s' TO '%s%s.ldf' , ", file.LogicalName, logPath, db.Name)
			}
		}
	}
	query += "RECOVERY;"
	fmt.Println(query)
	_, err := dr.connection.Query(query)
	if err != nil {
		return []model.Database{}, err
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
				&restoreDatabaseInfo.SourceBlockSize, &restoreDatabaseInfo.SourceBlockSize, &restoreDatabaseInfo.FileGroupId, &restoreDatabaseInfo.LogGroupGUID,
				&restoreDatabaseInfo.DifferentialBaseLSN, &restoreDatabaseInfo.IsReadOnly, &restoreDatabaseInfo.IsPresent, &restoreDatabaseInfo.TDEThumbprint,
				&restoreDatabaseInfo.SnapshotUrl)
			if err != nil {
				fmt.Println(err)
				return []model.DatabaseFromBackupFile{}, err
			}
			fmt.Println(restoreDatabaseInfo)
			restoreDatabase.Name = filepath.Base(backupFile)
			restoreDatabase.BackupFileInfo = append(restoreDatabase.BackupFileInfo, restoreDatabaseInfo)

			restoreDatabaseInfoList = append(restoreDatabaseInfoList, restoreDatabase)
		}

	}

	return restoreDatabaseInfoList, nil

}
