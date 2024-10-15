package utils

import (
	"database/sql"
	"fmt"
	"net/url"

	db "github.com/RenanMonteiroS/MaestroSQL/model"
	_ "github.com/microsoft/go-mssqldb"
)

func DbCon(dbConInfo *db.DatabaseCon) (*sql.DB, error) {
	queryParams := url.Values{}
	queryParams.Add("database", "master")
	queryParams.Add("encrypt", "disable")

	u := &url.URL{
		Scheme: "sqlserver",
		User:   url.UserPassword(dbConInfo.User, dbConInfo.Pwd),
		Host:   fmt.Sprintf("%s:%d", dbConInfo.Host, dbConInfo.Port),
		// Path:  instance, // if connecting to an instance instead of a port
		RawQuery: queryParams.Encode(),
	}

	con, err := sql.Open("sqlserver", u.String())
	return con, err
}
