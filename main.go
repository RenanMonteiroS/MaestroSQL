//go:generate goversioninfo -icon=conductor.ico
package main

import (
	"fmt"
	"os"
	"time"

	//"image/color"

	//"log"

	//"gioui.org/app"
	//"gioui.org/layout"
	//"gioui.org/op"
	//"gioui.org/text"
	//"gioui.org/unit"
	//"gioui.org/widget"
	//"gioui.org/widget/material"
	//"gioui.org/x/component"
	i "github.com/RenanMonteiroS/MaestroSQL/internal"
	db "github.com/RenanMonteiroS/MaestroSQL/model"
	u "github.com/RenanMonteiroS/MaestroSQL/utils"
	_ "github.com/RenanMonteiroS/MaestroSQL/views"
	_ "github.com/microsoft/go-mssqldb"
)

/* var hostInput widget.Editor
var portInput widget.Editor
var userInput widget.Editor
var passwordInput widget.Editor
var locationBackupInput widget.Editor
var sendButton widget.Clickable
var appbar component.AppBar */

func main() {
	var ope string

	var dbConInfo = db.DatabaseCon{Port: "1433", Instance: "SQLEXPRESS"}
	var databases = new(db.Database)
	/* go func() {
		w := new(app.Window)
		w.Option(app.Title("MaestroSQL"))
		w.Option(app.Size(unit.Dp(1000), unit.Dp(600)))
		if err := draw(w, &dbConInfo); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main() */

	//file, err := openLogFile("./sqlLog.log")

	fmt.Println("Digite a operacao desejada: (Backup, Restore)")
	fmt.Scanf("%s\n", &ope)

	switch ope {
	case "Backup":
		err := i.PrintBackupForm(&dbConInfo)
		if err != nil {
			fmt.Println("Erro: ", err)
			f, errFile := os.OpenFile("fatal.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			if errFile != nil {
				fmt.Println("Erro: ", errFile)
				time.Sleep(time.Second * 5)
				return
			}
			defer f.Close()
			fmt.Fprintf(f, "Erro: %v", err)
			time.Sleep(time.Second * 5)
			return
		}

		con, err := u.DbCon(&dbConInfo)

		if err != nil {
			fmt.Printf("Erro: %v", err)
			f, errFile := os.OpenFile("fatal.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			defer f.Close()
			if errFile != nil {
				fmt.Println("Erro: ", errFile)
				time.Sleep(time.Second * 5)
				return
			}

			fmt.Fprintf(f, "Erro: %v", err)
			time.Sleep(time.Second * 5)
			return
		}

		defer con.Close()

		dbList, err := databases.GetAllDatabases(con)
		if err != nil {
			fmt.Println("Erro: ", err)
			f, errFile := os.OpenFile("fatal.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			defer f.Close()
			if errFile != nil {
				fmt.Println("Erro: ", errFile)
				time.Sleep(time.Second * 5)
				return
			}
			fmt.Fprintf(f, "Erro: %v", err)
			time.Sleep(time.Second * 5)
			return
		}

		for _, db := range *dbList {
			databases.Names = append(databases.Names, db)
		}

		databases.Path, err = databases.GetDefaultBackupPath(con)
		if err != nil {
			if err.Error() != "sql: Scan error on column index 0, name \"\": converting NULL to string is unsupported" {
				f, errFile := os.OpenFile("fatal.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
				defer f.Close()
				if errFile != nil {
					fmt.Printf("Erro: %v\n", err)
					time.Sleep(time.Second * 5)
					return
				}
				fmt.Fprintf(f, "Erro: %v", err)
				fmt.Println("Erro: ", err)
				time.Sleep(time.Second * 5)
				return
			}
			databases.Path = ""
		}

		fmt.Printf("Caminho onde serao salvos os backups: (%v/)\n", databases.Path)
		fmt.Scanf("%s\n", &databases.Path)

		t0 := time.Now()
		backupQty, err := databases.Backup(con)
		if err != nil {
			fmt.Printf("Erro: %v\n", err)
			f, errFile := os.OpenFile("fatal.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			defer f.Close()
			if errFile != nil {
				fmt.Printf("Erro: %v\n", errFile)
				time.Sleep(time.Second * 5)
				return
			}
			fmt.Fprintf(f, "Erro: %v", err)
			time.Sleep(time.Second * 5)
			return
		}

		f, errFile := os.OpenFile("backupDatabase.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if errFile != nil {
			fmt.Printf("Erro: %v\n", errFile)
			time.Sleep(time.Second * 5)
			return
		}

		fmt.Fprintf(f, "-------------------//-------------------//-------------------//-------------------")
		fmt.Fprintf(f, "\nData: %v", time.Now().Format("2006-01-02"))
		fmt.Fprintf(f, "\nTotal de backups realizados: %v", backupQty)
		fmt.Fprintf(f, "\nLocal: %v", databases.Path)
		fmt.Fprintf(f, "\nTempo total: %v", time.Since(t0))
		fmt.Fprintf(f, "\n-------------------//-------------------//-------------------//-------------------\n")
		defer f.Close()

	case "Restore":
		fmt.Println("Digite o local onde os backups estão alocados:")
		fmt.Scanf("%s\n", &databases.Path)
		backupFileList, err := i.CountBackupFiles(databases.Path)
		if err != nil {
			fmt.Println("Erro: ", err)
			f, errFile := os.OpenFile("fatal.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			defer f.Close()
			if errFile != nil {
				fmt.Println("Erro: ", errFile)
				time.Sleep(time.Second * 5)
				return
			}
			fmt.Fprintf(f, "Erro: %v", err)
			time.Sleep(time.Second * 5)
			return
		}

		err = i.PrintBackupForm(&dbConInfo)
		if err != nil {
			fmt.Println("Erro: ", err)
			f, errFile := os.OpenFile("fatal.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			defer f.Close()
			if errFile != nil {
				fmt.Println("Erro: ", errFile)
				time.Sleep(time.Second * 5)
				return
			}
			fmt.Fprintf(f, "Erro: %v", err)
			time.Sleep(time.Second * 5)
			return
		}

		con, err := u.DbCon(&dbConInfo)
		if err != nil {
			fmt.Printf("Erro: %v", err)
			f, errFile := os.OpenFile("fatal.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			defer f.Close()
			if errFile != nil {
				fmt.Println("Erro: ", errFile)
				time.Sleep(time.Second * 5)
				return
			}
			fmt.Fprintf(f, "Erro: hey%v", err)
			time.Sleep(time.Second * 5)
			return
		}
		defer con.Close()

		t0 := time.Now()

		restoreQty, err := databases.Restore(con, &backupFileList)
		if err != nil {
			fmt.Printf("Erro: %v", err)
			f, errFile := os.OpenFile("restoreDatabase.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			defer f.Close()
			if errFile != nil {
				fmt.Printf("Erro: %v\n", errFile)
				time.Sleep(time.Second * 5)
				return
			}
			fmt.Fprintf(f, "Erro: %v", err)
			time.Sleep(time.Second * 5)
			return
		}

		f, errFile := os.OpenFile("restoreDatabase.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		defer f.Close()
		if errFile != nil {
			fmt.Printf("Erro: %v\n", errFile)
			time.Sleep(time.Second * 5)
			return
		}
		fmt.Fprintf(f, "-------------------//-------------------//-------------------//-------------------")
		fmt.Fprintf(f, "\nData: %v", time.Now().Format("2006-01-02"))
		fmt.Fprintf(f, "\nTotal de backups realizados: %v", restoreQty)
		fmt.Fprintf(f, "\nLocal: %v", databases.Path)
		fmt.Fprintf(f, "\nTempo total: %v", time.Since(t0))
		fmt.Fprintf(f, "\n-------------------//-------------------//-------------------//-------------------\n")

	default:
		f, err := os.OpenFile("fatal.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			fmt.Println("Erro: %v\n", err)
			time.Sleep(time.Second * 5)
			return
		}
		fmt.Fprintf(f, "Erro: Operação não autorizada:\n")
		fmt.Println("Erro: Operação não autorizada:\n")
		time.Sleep(time.Second * 5)
		defer f.Close()
		return

	}

}
