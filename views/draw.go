package views

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
