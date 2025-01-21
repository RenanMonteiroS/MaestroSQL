package repository

import (
	"database/sql"
	"fmt"

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

func (dr DatabaseRepository) GetDatabases() ([]model.Database, error) {
	query := "SELECT d.database_id, d.name DatabaseName, " +
		"f.name LogicalName, f.physical_name AS PhysicalName, f.type_desc TypeofFile " +
		"FROM sys.master_files f " +
		"INNER JOIN sys.databases " +
		"d ON d.database_id = f.database_id;"

	rows, err := dr.connection.Query(query)
	if err != nil {
		return []model.Database{}, err
	}

	var dbObjAux model.MergedDatabaseFileInfo
	var dbListAux []model.MergedDatabaseFileInfo

	for rows.Next() {
		err = rows.Scan(&dbObjAux.DatabaseId, &dbObjAux.DatabaseName, &dbObjAux.LogicalName,
			&dbObjAux.PhysicalName, &dbObjAux.File_type)
		fmt.Println(dbObjAux)
		if err != nil {
			fmt.Println(query)
			return []model.Database{}, err
		}

		dbListAux = append(dbListAux, dbObjAux)
	}

	var dbObj model.Database
	var dbFile model.DatabaseFile
	var dbList []model.Database
	var found bool

	for _, dbData := range dbListAux {
		dbObj.ID = dbObjAux.DatabaseId
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
			dbList = append(dbList, dbObj)
		}
		found = false
	}

	return dbList, nil
}

/* func (dr DatabaseRepository) BackupDatabase() ([]model.Database, error){

}  */
