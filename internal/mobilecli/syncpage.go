package mobilecli

import (
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"image/color"
)

type SyncPage struct {
	mobileApp *MobileApp

	editorBorder *widget.Border

	title     *material.LabelStyle
	configBtn *widget.Clickable

	statusLabel    *material.LabelStyle
	progressionBar *material.ProgressBarStyle
	syncBtn        *widget.Clickable

	mainWidget layout.Widget
}

func NewSyncPage(mobileApp *MobileApp) *SyncPage {
	p := &SyncPage{
		mobileApp: mobileApp,
	}

	label := material.H5(mobileApp.th, "Mifasol")
	p.title = &label
	p.editorBorder = &widget.Border{Color: color.NRGBA{A: 0xdd}, CornerRadius: unit.Dp(8), Width: unit.Px(1)}

	statusLabel := material.H5(mobileApp.th, "filepath...")
	p.statusLabel = &statusLabel

	progressBar := material.ProgressBar(mobileApp.th, 12)
	p.progressionBar = &progressBar

	p.syncBtn = new(widget.Clickable)
	p.configBtn = new(widget.Clickable)

	syncIcon, _ := widget.NewIcon(icons.NotificationSync)

	p.mainWidget = func(gtx layout.Context) layout.Dimensions {

		return layout.Flex{WeightSum: 1, Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{WeightSum: 1, Alignment: layout.Middle}.Layout(gtx,
					layout.Flexed(1,
						func(gtx layout.Context) layout.Dimensions {
							return layout.UniformInset(unit.Dp(8)).Layout(gtx, p.title.Layout)
						}),
					// Rigid widget.
					layout.Rigid(material.Button(mobileApp.th, p.configBtn, "Config").Layout),
				)
			}),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{WeightSum: 0, Axis: layout.Vertical, Spacing: layout.SpaceSides}.Layout(gtx,
					layout.Rigid(
						func(gtx layout.Context) layout.Dimensions {
							return layout.UniformInset(unit.Dp(20)).Layout(gtx, statusLabel.Layout)
						}),
					layout.Rigid(
						func(gtx layout.Context) layout.Dimensions {
							return layout.UniformInset(unit.Dp(20)).Layout(gtx, progressBar.Layout)
						}),
				)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(unit.Dp(30)).Layout(gtx,
					func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{WeightSum: 4, Spacing: layout.SpaceSides}.Layout(gtx,
							layout.Flexed(2, material.IconButton(mobileApp.th, p.syncBtn, syncIcon).Layout),
						)
					},
				)
			}),
		)

	}

	return p
}

func (p *SyncPage) display(gtx layout.Context, systemInsets system.Insets) {
	for p.syncBtn.Clicked() {
	}
	for p.configBtn.Clicked() {
		p.mobileApp.windowType = SetupPageType
	}

	//p.title.Text = msg
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
