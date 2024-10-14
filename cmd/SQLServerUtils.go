package main

import "fmt"

type DatabaseCon struct {
	Host string
	Port int16
	User string
	Pwd  string
}

func main() {
	var ope string
	var dbCon = DatabaseCon{Port: 1433}

	fmt.Println("Digite a operacao desejada: (Backup, Restore)")
	fmt.Scanf("%s\n", &ope)

	fmt.Println("Digite as informacoes relativas a conexao com o banco de dados:")
	fmt.Println("Host:")
	fmt.Scanf("%s\n", &dbCon.Host)
	fmt.Println("Porta:")
	fmt.Scanf("%d\n", &dbCon.Port)
	fmt.Println("Usuario:")
	fmt.Scanf("%s\n", &dbCon.User)
	fmt.Println("Senha:")
	fmt.Scanf("%s\n", &dbCon.Pwd)

	fmt.Println(dbCon)
}
