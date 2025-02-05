package db

import (
	"database/sql"
	"fmt"
	"net/url"

	"github.com/RenanMonteiroS/MaestroSQLWeb/model"
	_ "github.com/microsoft/go-mssqldb"
)

func ConnDb(connInfo model.ConnInfo) (*sql.DB, error) {
	queryParams := url.Values{}
	queryParams.Add("database", connInfo.DbName)
	queryParams.Add("encrypt", "disable")

	u := &url.URL{
		Scheme: "sqlserver",
		User:   url.UserPassword(connInfo.User, connInfo.Password),
		Host:   fmt.Sprintf("%s:%s", connInfo.Host, connInfo.Port),
		//Path:     dbConInfo.Instance, // if connecting to an instance instead of a port
		RawQuery: queryParams.Encode(),
	}

	db, err := sql.Open("sqlserver", u.String())
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	fmt.Println("Connected to database")

	return db, nil
}
