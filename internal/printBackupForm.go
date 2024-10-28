package internal

import (
	"fmt"
	"syscall"

	"golang.org/x/term"

	db "github.com/RenanMonteiroS/MaestroSQL/model"
)

func PrintBackupForm(dbConInfo *db.DatabaseCon) error {

	fmt.Println("Digite as informacoes relativas a conexao com o banco de dados:")

	fmt.Println("Host:")
	fmt.Scanf("%s\n", &dbConInfo.Host)

	fmt.Println("Instancia (SQLEXPRESS):")
	fmt.Scanf("%s\n", &dbConInfo.Instance)

	fmt.Println("Porta (1433):")
	fmt.Scanf("%s\n", &dbConInfo.Port)

	fmt.Println("Usuario:")
	fmt.Scanf("%s\n", &dbConInfo.User)

	fmt.Println("Senha:")
	bytePwd, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Println(err)
		return err
	}
	dbConInfo.Pwd = string(bytePwd)

	return nil
}
