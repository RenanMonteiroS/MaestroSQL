package internal

import (
	"database/sql"
	"fmt"
)

func BackupDB(con *sql.DB, database string, path string) (string, error) {
	var query string = "BACKUP DATABASE [" + database + "] TO DISK = '" + path + "/" + database + ".bak' WITH FORMAT;"
	_, err := con.Query(query)
	if err != nil {
		fmt.Println("Erro: ", err)
		return "", err
	}

	return query, nil
}
