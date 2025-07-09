package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"

	"github.com/RenanMonteiroS/MaestroSQLWeb/model"
	_ "github.com/microsoft/go-mssqldb"
)

// Creates a connection pool using the provided connection information. It don't uses encryption and connects by default to the [master] database
func ConnDb(connInfo model.ConnInfo) (*sql.DB, error) {
	queryParams := url.Values{}
	var u *url.URL

	if connInfo.Encryption == "" {
		queryParams.Add("encrypt", "mandatory")
	} else {
		queryParams.Add("encrypt", connInfo.Encryption)
	}

	if connInfo.TrustServerCertificate == nil {
		queryParams.Add("trustServerCertificate", "false")
	} else {
		queryParams.Add("trustServerCertificate", strconv.FormatBool(*connInfo.TrustServerCertificate))
	}

	queryParams.Add("database", "master")

	u = &url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(connInfo.User, connInfo.Password),
		Host:     connInfo.Host,
		RawQuery: queryParams.Encode(),
	}

	if connInfo.Port != "" {
		u.Host = fmt.Sprintf("%s:%s", connInfo.Host, connInfo.Port)
	} else if connInfo.Instance != "" {
		u.Path = connInfo.Instance
	}

	db, err := sql.Open("sqlserver", u.String())
	if err != nil {
		return nil, err
	}

	slog.Info("Trying to connect to the database: ", "ConnInfo", connInfo)

	if connInfo.MaxConnections != 0 {
		db.SetMaxOpenConns(connInfo.MaxConnections)
	}

	// Checks the database connection
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
