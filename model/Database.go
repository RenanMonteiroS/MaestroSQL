package model

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type Database struct {
	Names []string
	Path  string
}

func (db Database) Backup(con *sql.DB) (int, error) {
	var wg sync.WaitGroup

	var sumBackupDbs int
	f, err := os.OpenFile("backupDatabase.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	log.SetOutput(f)

	for _, dbName := range db.Names {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var query string
			query = "BACKUP DATABASE [" + dbName + "] TO DISK = '" + db.Path + "/" + dbName + "_" + time.Now().Format("2006-01-02") + "_" + time.Now().Format("15-04-05") + ".bak' WITH FORMAT;"

			_, err = con.Query(query)
			if err != nil {
				log.Printf("Erro: %v Banco de dados: %v", err, dbName)
				return
			}
			sumBackupDbs += 1
			log.Printf(query)
			//queries = append(queries, query)
		}()
	}
	wg.Wait()
	defer f.Close()

	return sumBackupDbs, nil
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
