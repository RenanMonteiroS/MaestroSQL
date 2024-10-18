package main

import (
	"fmt"
	"log"
	"os"

	//"os"
	"gioui.org/app"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	db "github.com/RenanMonteiroS/MaestroSQL/model"
	u "github.com/RenanMonteiroS/MaestroSQL/utils"
	_ "github.com/microsoft/go-mssqldb"
)

func main() {
	go func() {
		w := new(app.Window)
		w.Option(app.Title("MaestroSQL"))
		w.Option(app.Size(unit.Dp(1000), unit.Dp(600)))
		if err := draw(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()

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

	defer con.Close()

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

/*
	 func openLogFile(path string) (*os.File, error) {
		logFile, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0777)
		if err != nil {
			return nil, err
		}
		return logFile, nil
	}
*/
var hostInput widget.Editor
var sendButtonVar widget.Clickable

func draw(window *app.Window) error {
	theme := material.NewTheme()
	var ops op.Ops
	for {
		switch e := window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			sendButton := material.Button(theme, &sendButtonVar, "Start")
			sendButton.Layout(gtx)

			if sendButton.Button.Clicked(gtx) {
				//inputString := hostInput.Text()
				fmt.Println("hey")
			}
			e.Frame(gtx.Ops)
			/* layout.Flex{
				Axis:    layout.Vertical,
				Spacing: layout.SpaceStart,
			}.Layout(gtx,
				layout.Rigid(
					func(gtx layout.Context) layout.Dimensions {
						// ONE: First define margins around the button using layout.Inset ...
						margins := layout.Inset{
							Top:    unit.Dp(25),
							Bottom: unit.Dp(25),
							Right:  unit.Dp(35),
							Left:   unit.Dp(35),
						}
						// TWO: ... then we lay out those margins ...
						return margins.Layout(gtx,
							// THREE: ... and finally within the margins, we ddefine and lay out the button
							func(gtx layout.Context) layout.Dimensions {
								input := material.Editor(theme, &hostInput, "teste")
								return input.Layout(gtx)
							},
						)
					},
				),
				layout.Rigid(
					func(gtx layout.Context) layout.Dimensions {

						margins := layout.Inset{
							Top:    unit.Dp(25),
							Bottom: unit.Dp(25),
							Right:  unit.Dp(35),
							Left:   unit.Dp(35),
						}

						return margins.Layout(gtx,
							func(gtx layout.Context) layout.Dimensions {
								sendButton := material.Button(theme, &sendButton, "Start")
								return sendButton.Layout(gtx)
							},
						)
					},
				),
			) */

			/* if sendButton.Clicked(gtx) {
				//inputString := hostInput.Text()
				fmt.Println("hey")
			}
			e.Frame(gtx.Ops) */
		}
	}
}
