package internal

import (
	"database/sql"
	"fmt"
)

func BackupDB(con *sql.DB, database string) (string, error) {
	var query string = "BACKUP DATABASE [" + database + "] TO DISK = '/var/opt/mssql/data/" + database + ".bak' WITH FORMAT;"
	_, err := con.Query(query)
	if err != nil {
		fmt.Println("Erro: ", err)
		return "", err
	}

	return query, nil
}
