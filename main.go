package main

import (
	"fmt"

	//"log"
	//"os"
	i "github.com/RenanMonteiroS/MaestroSQL/internal"
	db "github.com/RenanMonteiroS/MaestroSQL/model"
	u "github.com/RenanMonteiroS/MaestroSQL/utils"
	_ "github.com/microsoft/go-mssqldb"
)

func main() {
	var ope string
	var dbConInfo = db.DatabaseCon{Port: 1433, Instance: "SQLEXPRESS"}
	var databases []string
	var path string

	//file, err := openLogFile("./sqlLog.log")

	fmt.Println("Digite a operacao desejada: (Backup, Restore)")
	fmt.Scanf("%s\n", &ope)

	fmt.Println("Digite as informacoes relativas a conexao com o banco de dados:")
	fmt.Println("Host:")
	fmt.Scanf("%s\n", &dbConInfo.Host)
	fmt.Println("Instancia (SQLEXPRESS):")
	fmt.Scanf("%s\n", &dbConInfo.Instance)
	fmt.Println("Porta (1433):")
	fmt.Scanf("%d\n", &dbConInfo.Port)
	fmt.Println("Usuario:")
	fmt.Scanf("%s\n", &dbConInfo.User)
	fmt.Println("Senha:")
	fmt.Scanf("%s\n", &dbConInfo.Pwd)

	con, err := u.DbCon(&dbConInfo)

	dbList, err := con.Query("SELECT name FROM sys.databases WHERE name not in ('master', 'model', 'msdb', 'tempdb');")
	if err != nil {
		fmt.Println("Erro: ", err)
		return
	}

	for dbList.Next() {
		var dbName string
		err := dbList.Scan(&dbName)
		if err != nil {
			fmt.Println("Erro: ", err)
			return
		}
		databases = append(databases, dbName)
	}

	defaultbackuppath, err := con.Query("SELECT SERVERPROPERTY('instancedefaultbackuppath');")
	for defaultbackuppath.Next() {
		err := defaultbackuppath.Scan(&path)
		if err != nil {
			fmt.Println("Erro: ", err)
			return
		}
	}

	fmt.Printf("Caminho onde serao salvos os backups: %v/", path)
	fmt.Scanf("%s\n", &path)

	for _, database := range databases {
		result, err := i.BackupDB(con, database, path)
		if err != nil {
			fmt.Println("Erro: ", err)
			return
		}
		fmt.Println(result)
	}

}

/* func openLogFile(path string) (*os.File, error) {
	logFile, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}
	return logFile, nil
} */
