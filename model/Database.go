package model

import (
	"database/sql"
	"fmt"
	"time"
)

type Database struct {
	Names []string
	Path  string
}

func (db Database) Backup(con *sql.DB) (*[]string, error) {
	var queries []string
	for _, dbName := range db.Names {
		var query string
		query = "BACKUP DATABASE [" + dbName + "] TO DISK = '" + db.Path + "/" + dbName + "_" + time.Now().Format("2006-01-02") + "_" + time.Now().Format("15-04-05") + ".bak' WITH FORMAT;"
		queries = append(queries, query)
		fmt.Println(queries)
		_, err := con.Query(query)
		if err != nil {
			fmt.Println("Erro: ", err)
			return &queries, err
		}
	}
	return &queries, nil
}

func (db Database) GetAllDatabases(con *sql.DB) (*[]string, error) {
	var databases []string
	dbList, err := con.Query("SELECT name FROM sys.databases WHERE name not in ('master', 'model', 'msdb', 'tempdb') AND state_desc='ONLINE' order by name;")
	if err != nil {
		fmt.Println("Erro: ", err)
		return &databases, err
	}

	for dbList.Next() {
		var dbName string
		err := dbList.Scan(&dbName)
		if err != nil {
			fmt.Println("Erro: ", err)
			return &databases, nil
		}
		databases = append(databases, dbName)
	}

	return &databases, nil
}

func (db Database) GetDefaultBackupPath(con *sql.DB) (string, error) {
	var defaultBackupPath string
	result, err := con.Query("SELECT SERVERPROPERTY('instancedefaultbackuppath');")
	if err != nil {
		return "", err
	}
	for result.Next() {
		err := result.Scan(&defaultBackupPath)

		if err != nil {
			return "", err
		}
	}

	if defaultBackupPath == "" {
		fmt.Println("Hey2")
		return "", nil
	}
	return defaultBackupPath, nil
}
