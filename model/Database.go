package model

import (
	"database/sql"
	"fmt"
)

type Database struct {
	Names []string
	Path  string
	Con   *sql.DB
}

func (db Database) Backup() (*[]string, error) {
	var queries []string
	for _, dbName := range db.Names {
		var query string
		query = "BACKUP DATABASE [" + dbName + "] TO DISK = '" + db.Path + "/" + dbName + ".bak' WITH FORMAT;"
		queries = append(queries, query)
		_, err := db.Con.Query(query)
		if err != nil {
			fmt.Println("Erro: ", err)
			return &queries, err
		}
	}

	return &queries, nil
}
