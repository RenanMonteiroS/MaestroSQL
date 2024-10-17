package main

import (
	"fmt"

	//"log"
	//"os"

	db "github.com/RenanMonteiroS/MaestroSQL/model"
	u "github.com/RenanMonteiroS/MaestroSQL/utils"
	_ "github.com/microsoft/go-mssqldb"
)

func main() {
	var ope string
	var dbConInfo = db.DatabaseCon{Port: 1433, Instance: "SQLEXPRESS"}
	var databases = new(db.Database)

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
	if err != nil {
		fmt.Println("Erro: ", err)
		return
	}

	dbList, err := databases.GetAllDatabases(con)
	for _, db := range *dbList {
		databases.Names = append(databases.Names, db)
	}

	databases.Path, err = databases.GetDefaultBackupPath(con)

	fmt.Printf("Caminho onde serao salvos os backups: %v/", databases.Path)
	fmt.Scanf("%s\n", &databases.Path)

	result, err := databases.Backup(con)
	if err != nil {
		fmt.Println("Erro: ", err)
		return
	}
	fmt.Println(*result)
}

/* func openLogFile(path string) (*os.File, error) {
	logFile, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}
	return logFile, nil
} */
