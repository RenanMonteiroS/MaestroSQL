package model

import (
	"encoding/json"
	"fmt"
)

type SqlErr struct {
	Database string
	Err      error
}

func NewSqlErr(database string, err error) *SqlErr {
	if err == nil {
		return nil
	}

	return &SqlErr{
		Database: database,
		Err:      err,
	}
}

func (se *SqlErr) Error() string {
	if se.Err == nil {
		return fmt.Sprintf("Error on database %v: <nil>", se.Database)
	}

	return fmt.Sprintf("Error on database %v: %v", se.Database, se.Err.Error())
}

func (se *SqlErr) Unwrap() error {
	return se.Err
}

func (se *SqlErr) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Database string `json:"database"`
		Error    string `json:"error"`
	}{
		Database: se.Database,
		Error:    se.Err.Error(),
	})
}
