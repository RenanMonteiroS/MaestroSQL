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
	return ds.repository.GetDatabases()
}
