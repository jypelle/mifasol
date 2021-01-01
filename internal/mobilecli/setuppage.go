package mobilecli

import (
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"image/color"
	"strconv"
)

type SetupPage struct {
	mobileApp *MobileApp

	title                    *material.LabelStyle
	serverHostnameEditor     *widget.Editor
	serverPortEditor         *widget.Editor
	serverSslCheckbox        *widget.Bool
	serverSelfSignedCheckbox *widget.Bool
	usernameEditor           *widget.Editor
	passwordEditor           *widget.Editor
	editorBorder             *widget.Border
	backBtn                  *widget.Clickable
	mainWidget               layout.Widget
}

func NewSetupPage(mobileApp *MobileApp) *SetupPage {

	p := &SetupPage{
		mobileApp: mobileApp,
	}

	label := material.H5(mobileApp.th, "Setup")
	p.title = &label

	p.serverHostnameEditor = &widget.Editor{
		SingleLine: true,
	}
	p.serverHostnameEditor.SetText(mobileApp.config.ServerHostname)

	p.serverPortEditor = &widget.Editor{
		SingleLine: true,
	}
	p.serverPortEditor.SetText(strconv.FormatInt(mobileApp.config.ServerPort, 10))

	p.serverSslCheckbox = new(widget.Bool)
	p.serverSslCheckbox.Value = mobileApp.config.ServerSsl

	p.serverSelfSignedCheckbox = new(widget.Bool)
	p.serverSelfSignedCheckbox.Value = mobileApp.config.ServerSelfSigned

	p.usernameEditor = &widget.Editor{
		SingleLine: true,
		Submit:     true,
	}
	p.usernameEditor.SetText(mobileApp.config.Username)

	p.passwordEditor = &widget.Editor{
		SingleLine: true,
		Submit:     true,
		Mask:       '*',
	}
	p.passwordEditor.SetText(mobileApp.config.Password)

	p.editorBorder = &widget.Border{Color: color.NRGBA{A: 0xdd}, CornerRadius: unit.Dp(8), Width: unit.Px(1)}

	p.backBtn = new(widget.Clickable)

	widgets := []layout.Widget{
		func(gtx layout.Context) layout.Dimensions {
			e := material.Editor(mobileApp.th, p.serverHostnameEditor, "Server Host")
			return p.editorBorder.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(unit.Dp(8)).Layout(gtx, e.Layout)
			})
		},
		func(gtx layout.Context) layout.Dimensions {
			e := material.Editor(mobileApp.th, p.serverPortEditor, "Server Port")
			return p.editorBorder.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(unit.Dp(8)).Layout(gtx, e.Layout)
			})
		},
		func(gtx layout.Context) layout.Dimensions {
			return material.CheckBox(mobileApp.th, p.serverSslCheckbox, "SSL").Layout(gtx)
		},
		func(gtx layout.Context) layout.Dimensions {
			return material.CheckBox(mobileApp.th, p.serverSelfSignedCheckbox, "Self Signed").Layout(gtx)
		},
		func(gtx layout.Context) layout.Dimensions {
			e := material.Editor(mobileApp.th, p.usernameEditor, "Username")
			return p.editorBorder.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(unit.Dp(8)).Layout(gtx, e.Layout)
			})
		},
		func(gtx layout.Context) layout.Dimensions {
			e := material.Editor(mobileApp.th, p.passwordEditor, "Password")
			return p.editorBorder.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(unit.Dp(8)).Layout(gtx, e.Layout)
			})
		},
	}

	list := &layout.List{
		Axis: layout.Vertical,
	}

	backIcon, _ := widget.NewIcon(icons.HardwareKeyboardArrowLeft)

	p.mainWidget = func(gtx layout.Context) layout.Dimensions {

		return layout.Flex{WeightSum: 1, Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{WeightSum: 1, Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(material.IconButton(mobileApp.th, p.backBtn, backIcon).Layout),
					layout.Flexed(1,
						func(gtx layout.Context) layout.Dimensions {
							return layout.UniformInset(unit.Dp(8)).Layout(gtx, p.title.Layout)
						}),
				)
			}),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{WeightSum: 0, Axis: layout.Vertical, Spacing: layout.SpaceSides}.Layout(gtx,
					layout.Rigid(
						func(gtx layout.Context) layout.Dimensions {
							return list.Layout(
								gtx,
								len(widgets),
								func(gtx layout.Context, i int) layout.Dimensions {
									return layout.UniformInset(unit.Dp(8)).Layout(gtx, widgets[i])
								},
							)
						}),
				)
			}),
		)

	}

	return p
}

func (p *SetupPage) display(gtx layout.Context, systemInsets system.Insets) {
	for p.backBtn.Clicked() {
		p.back()
	}

	inset := &layout.Inset{
		Top:    systemInsets.Top,
		Right:  systemInsets.Right,
		Bottom: systemInsets.Bottom,
		Left:   systemInsets.Left,
	}

	inset.Layout(
		gtx,
		p.mainWidget,
	)

	return
}

func (p *SetupPage) back() {
	p.mobileApp.windowType = SyncPageType
	p.mobileApp.config.ServerHostname = p.serverHostnameEditor.Text()
	serverPort, err := strconv.ParseInt(p.serverPortEditor.Text(), 10, 64)
	if err == nil {
		p.mobileApp.config.ServerPort = serverPort
	}
	p.serverPortEditor.SetText(strconv.FormatInt(p.mobileApp.config.ServerPort, 10))
	p.mobileApp.config.ServerSsl = p.serverSslCheckbox.Value
	p.mobileApp.config.ServerSelfSigned = p.serverSelfSignedCheckbox.Value
	p.mobileApp.config.Username = p.usernameEditor.Text()
	p.mobileApp.config.Password = p.passwordEditor.Text()
	p.mobileApp.config.Save()
}
