//go:generate goversioninfo -icon=conductor.ico
package main

import (
	"fmt"
	"os"
	"syscall"
	"time"

	//"image/color"

	//"log"

	"golang.org/x/term"

	//"gioui.org/app"
	//"gioui.org/layout"
	//"gioui.org/op"
	//"gioui.org/text"
	//"gioui.org/unit"
	//"gioui.org/widget"
	//"gioui.org/widget/material"
	//"gioui.org/x/component"
	db "github.com/RenanMonteiroS/MaestroSQL/model"
	u "github.com/RenanMonteiroS/MaestroSQL/utils"
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
	//var ope string

	var err error
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

	/* fmt.Println("Digite a operacao desejada: (Backup, Restore)")
	ope, err = reader.ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return
	} */

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
		return
	}
	dbConInfo.Pwd = string(bytePwd)

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
	if err != nil {
		if err.Error() != "sql: Scan error on column index 0, name \"\": converting NULL to string is unsupported" {
			fmt.Println("Erro: ", err)
			return
		}
		databases.Path = ""
	}

	fmt.Printf("Caminho onde serao salvos os backups: (%v/)\n", databases.Path)
	fmt.Scanf("%s\n", &databases.Path)

	t0 := time.Now()
	backupQty, err := databases.Backup(con)

	f, err := os.OpenFile("backupDatabase.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	fmt.Fprintf(f, "-------------------//-------------------//-------------------//-------------------")
	fmt.Fprintf(f, "\nData: %v", time.Now().Format("2006-01-02"))
	fmt.Fprintf(f, "\nTotal de backups realizados: %v", backupQty)
	fmt.Fprintf(f, "\nLocal: %v", databases.Path)
	fmt.Fprintf(f, "\nTempo total: %v", time.Since(t0))
	fmt.Fprintf(f, "\n-------------------//-------------------//-------------------//-------------------\n")
}

/* func draw(window *app.Window, dbConInfo *db.DatabaseCon) error {
	theme := material.NewTheme()

	var ops op.Ops
	for {
		switch e := window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			if sendButton.Clicked(gtx) {
				dbConInfo.Host = strings.TrimSpace(hostInput.Text())
				dbConInfo.Port = strings.TrimSpace(portInput.Text())
				dbConInfo.User = strings.TrimSpace(userInput.Text())
				dbConInfo.Pwd = strings.TrimSpace(passwordInput.Text())
				//locationBackupString := strings.TrimSpace(locationBackupInput.Text())

			}

			layout.Flex{
				Axis: layout.Vertical,
				//Espaco no inicio da tela
				Spacing: layout.SpaceStart,
			}.Layout(gtx,

				layout.Rigid(
					func(gtx layout.Context) layout.Dimensions {

						hostInput.Alignment = text.Middle
						hostInput.SingleLine = true
						input := material.Editor(theme, &hostInput, "Host")

						margins := layout.Inset{
							Top:    unit.Dp(25),
							Bottom: unit.Dp(25),
							Right:  unit.Dp(300),
							Left:   unit.Dp(300),
						}

						border := widget.Border{
							Color:        color.NRGBA{63, 81, 181, 255},
							CornerRadius: unit.Dp(3),
							Width:        unit.Dp(2),
						}

						return margins.Layout(gtx,
							func(gtx layout.Context) layout.Dimensions {
								return border.Layout(gtx, input.Layout)
							},
						)
					},
				),
				layout.Rigid(
					func(gtx layout.Context) layout.Dimensions {

						input := material.Editor(theme, &portInput, "Porta")

						portInput.Alignment = text.Middle
						portInput.SingleLine = true

						margins := layout.Inset{
							Top:    unit.Dp(25),
							Bottom: unit.Dp(25),
							Right:  unit.Dp(300),
							Left:   unit.Dp(300),
						}

						border := widget.Border{
							Color:        color.NRGBA{63, 81, 181, 255},
							CornerRadius: unit.Dp(3),
							Width:        unit.Dp(2),
						}

						return margins.Layout(gtx,
							func(gtx layout.Context) layout.Dimensions {
								return border.Layout(gtx, input.Layout)
							},
						)
					},
				),
				layout.Rigid(
					func(gtx layout.Context) layout.Dimensions {
						userInput.Alignment = text.Middle
						userInput.SingleLine = true
						input := material.Editor(theme, &userInput, "Usuario")

						margins := layout.Inset{
							Top:    unit.Dp(25),
							Bottom: unit.Dp(25),
							Right:  unit.Dp(300),
							Left:   unit.Dp(300),
						}

						border := widget.Border{
							Color:        color.NRGBA{63, 81, 181, 255},
							CornerRadius: unit.Dp(3),
							Width:        unit.Dp(2),
						}

						return margins.Layout(gtx,
							func(gtx layout.Context) layout.Dimensions {
								return border.Layout(gtx, input.Layout)
							},
						)
					},
				),
				layout.Rigid(
					func(gtx layout.Context) layout.Dimensions {
						input := material.Editor(theme, &passwordInput, "Senha")

						passwordInput.Alignment = text.Middle
						passwordInput.SingleLine = true

						margins := layout.Inset{
							Top:    unit.Dp(25),
							Bottom: unit.Dp(25),
							Right:  unit.Dp(300),
							Left:   unit.Dp(300),
						}

						border := widget.Border{
							Color:        color.NRGBA{63, 81, 181, 255},
							CornerRadius: unit.Dp(3),
							Width:        unit.Dp(2),
						}

						return margins.Layout(gtx,
							func(gtx layout.Context) layout.Dimensions {
								return border.Layout(gtx, input.Layout)
							},
						)
					},
				),
				layout.Rigid(
					func(gtx layout.Context) layout.Dimensions {

						input := material.Editor(theme, &locationBackupInput, "Localizacao dos backups")

						locationBackupInput.Alignment = text.Middle
						locationBackupInput.SingleLine = true

						margins := layout.Inset{
							Top:    unit.Dp(25),
							Bottom: unit.Dp(25),
							Right:  unit.Dp(300),
							Left:   unit.Dp(300),
						}

						border := widget.Border{
							Color:        color.NRGBA{63, 81, 181, 255},
							CornerRadius: unit.Dp(3),
							Width:        unit.Dp(2),
						}

						return margins.Layout(gtx,
							func(gtx layout.Context) layout.Dimensions {
								return border.Layout(gtx, input.Layout)
							},
						)
					},
				),
				layout.Rigid(
					func(gtx layout.Context) layout.Dimensions {

						margins := layout.Inset{
							Top:    unit.Dp(25),
							Bottom: unit.Dp(25),
							Right:  unit.Dp(300),
							Left:   unit.Dp(300),
						}

						return margins.Layout(gtx,
							func(gtx layout.Context) layout.Dimensions {
								sendButton := material.Button(theme, &sendButton, "Start")
								return sendButton.Layout(gtx)
							},
						)
					},
				),
			)

			e.Frame(gtx.Ops)
		}
	}
} */
