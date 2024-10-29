package model

import (
	"database/sql"
	"log"
	"os"
	"strings"
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
		return 0, err
	}
	log.SetOutput(f)

	for _, dbName := range db.Names {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var query string
			query = "BACKUP DATABASE [" + dbName + "] TO DISK = '" + db.Path + "/" + dbName + "=" + time.Now().Format("2006-01-02") + "_" + time.Now().Format("15-04-05") + ".bak' WITH FORMAT;"

			_, err = con.Query(query)
			if err != nil {
				log.Printf("Erro: %v Banco de dados: %v", err, dbName)
				return
			}
			sumBackupDbs += 1
			log.Printf(query)
		}()
	}
	wg.Wait()
	defer f.Close()

	return sumBackupDbs, nil
}

type dataFile struct {
	logicalName          string
	physicalName         string
	fileType             string
	fileGroupName        string
	size                 string
	maxSize              string
	fileId               string
	createLSN            string
	dropLSN              string
	uniqueId             string
	readOnlyLSN          string
	readWriteLSN         string
	backupSizeInBytes    string
	sourceBlockSize      string
	fileGroupId          string
	logGroupGUID         string
	differentialBaseLSN  string
	differentialBaseGUID string
	isReadOnly           string
	isPresent            string
	TDEThumbprint        string
}

func (db Database) Restore(con *sql.DB, backupFileList *[]string) (int, error) {
	var sumRestoreDbs int
	f, err := os.OpenFile("restoreDatabase.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return 0, err
	}
	log.SetOutput(f)

	for _, backupFileName := range *backupFileList {
		if strings.Contains(backupFileName, ".bak") {
			var query, masterPhysicalName, mastlogPhysicalName string
			var defaultLogPath, defaultDataPath = "", ""
			var backupFiles []dataFile

			dbName := strings.Split(backupFileName, "=")[0]

			query = "SELECT serverproperty('InstanceDefaultDataPath'), serverproperty('InstanceDefaultLogPath');"
			defaultFilePaths, err := con.Query(query)
			if err != nil {
				log.Printf("Erro: %v Query: %v", err, query)
				return 0, err
			}

			for defaultFilePaths.Next() {
				defaultFilePaths.Scan(&defaultDataPath, &defaultLogPath)
			}

			if defaultDataPath == "" {
				query = "SELECT physical_name FROM sys.master_files WHERE name='master';"
				err := con.QueryRow(query, 1).Scan(&masterPhysicalName)
				if err != nil {
					log.Printf("Erro: %v Query: %v", err, query)
					return 0, err
				}
				defaultDataPath = strings.Split(masterPhysicalName, "master.mdf")[0]
			}
			if defaultLogPath == "" {
				query = "SELECT physical_name FROM sys.master_files WHERE name='mastlog';"
				err := con.QueryRow(query, 1).Scan(&mastlogPhysicalName)
				if err != nil {
					log.Printf("Erro: %v Query: %v", err, query)
					return 0, err
				}
				defaultLogPath = strings.Split(mastlogPhysicalName, "mastlog.ldf")[0]
			}

			query = "RESTORE FILELISTONLY FROM DISK = '" + db.Path + "/" + backupFileName + "'; "
			dataFiles, err := con.Query(query)
			if err != nil {
				log.Printf("Erro: %v Query: %v", err, query)
				return 0, err
			}
			for dataFiles.Next() {
				var dataFile = dataFile{"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", ""}

				dataFiles.Scan(&dataFile.logicalName, &dataFile.physicalName, &dataFile.fileType, &dataFile.fileGroupName, &dataFile.size,
					&dataFile.maxSize, &dataFile.fileId, &dataFile.createLSN, &dataFile.dropLSN, &dataFile.uniqueId, &dataFile.readOnlyLSN, &dataFile.readWriteLSN,
					&dataFile.backupSizeInBytes, &dataFile.sourceBlockSize, &dataFile.fileGroupId, &dataFile.logGroupGUID, &dataFile.differentialBaseLSN,
					&dataFile.differentialBaseGUID, &dataFile.isReadOnly, &dataFile.isPresent, &dataFile.TDEThumbprint)

				backupFiles = append(backupFiles, dataFile)
			}

			query = "RESTORE DATABASE [" + dbName + "] FROM DISK = '" + db.Path + "/" + backupFileName + "' WITH "
			for _, dataFile := range backupFiles {
				if dataFile.fileType == "D" {
					query += "MOVE '" + dataFile.logicalName + "' TO '" + defaultDataPath + dbName + ".mdf', "
				} else if dataFile.fileType == "L" {
					query += "MOVE '" + dataFile.logicalName + "' TO '" + defaultLogPath + dbName + "_log.ldf', "
				}
			}
			query += "RECOVERY;"
			_, err = con.Query(query)
			if err != nil {
				log.Printf("Erro: %v Query: %v", err, query)
				return 0, err
			}

			sumRestoreDbs += 1
			log.Printf("%v\n", query)
		}
	}

	return sumRestoreDbs, nil
}

func (db Database) GetAllDatabases(con *sql.DB) (*[]string, error) {
	var databases []string
	dbList, err := con.Query("SELECT name FROM sys.databases WHERE name not in ('master', 'model', 'msdb', 'tempdb') AND state_desc='ONLINE' order by name;")
	if err != nil {
		return &databases, err
	}

	for dbList.Next() {
		var dbName string
		err := dbList.Scan(&dbName)
		if err != nil {
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
		return "", nil
	}
	return defaultBackupPath, nil
}
