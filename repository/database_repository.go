package repository

import (
	"database/sql"
	"fmt"
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
