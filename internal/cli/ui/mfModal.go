package ui

import "code.rocketnine.space/tslocum/cview"

type MfModal struct {
	*cview.Modal
	currentFocus cview.Primitive
	app          *App
}

func (a *App) OpenModalMessage(message string) *MfModal {
	mfModal := &MfModal{
		Modal:        cview.NewModal(),
		currentFocus: a.cviewApp.GetFocus(),
		app:          a,
	}

	mfModal.SetText(message)

	a.pagesComponent.AddPage(
		"modalMessage",
		mfModal.Modal,
		false,
		true,
	)

	return mfModal
}
func (m *MfModal) Close() {
	m.app.pagesComponent.HidePage("modalMessage")
	m.app.pagesComponent.RemovePage("modalMessage")
	m.app.cviewApp.SetFocus(m.currentFocus)
	m.app.cviewApp.Draw()
}
