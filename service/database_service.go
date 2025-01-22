package service

import (
	"github.com/RenanMonteiroS/MaestroSQLWeb/model"
	"github.com/RenanMonteiroS/MaestroSQLWeb/repository"
)

type DatabaseService struct {
	repository repository.DatabaseRepository
}

func NewDatabaseService(rp repository.DatabaseRepository) DatabaseService {
	return DatabaseService{repository: rp}
}

func (ds DatabaseService) GetDatabases() ([]model.Database, error) {
	dbListAux, err := ds.repository.GetDatabases()
	if err != nil {
		return []model.Database{}, err
	}

	var dbObj model.Database
	var dbFile model.DatabaseFile
	var dbList []model.Database
	var found bool

	for _, dbData := range dbListAux {
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
			dbList = append(dbList, dbObj)
		}
		found = false
	}

	return dbList, nil
}
