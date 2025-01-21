package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	_ "github.com/microsoft/go-mssqldb"
)

type ConnInfo struct {
	Host     string
	Port     string
	User     string
	Password string
	DbName   string
}

func ConnDb() (*sql.DB, error) {
	var connInfo ConnInfo

	file, err := os.Open("../config/config.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	decoder.Decode(&connInfo)

	fmt.Println(connInfo)

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
