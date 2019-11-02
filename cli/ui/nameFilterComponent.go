package ui

import (
	"github.com/rivo/tview"
	"strings"
)

type NameFilterComponent struct {
	*tview.Form
	nameFilterInputField *tview.InputField
	uiApp                *UIApp
}

func OpenNameFilterComponent(uiApp *UIApp) {

	c := &NameFilterComponent{
		uiApp: uiApp,
	}

	c.nameFilterInputField = tview.NewInputField().
		SetLabel("Name").
		SetFieldWidth(30)

	if uiApp.libraryComponent.currentFilter().contains != nil {
		c.nameFilterInputField.SetText(*uiApp.libraryComponent.currentFilter().contains)
	}

	c.Form = tview.NewForm()
	c.Form.AddFormItem(c.nameFilterInputField)
	c.Form.AddButton("Filter", c.filter)
	c.Form.AddButton("Cancel", c.cancel)
	c.Form.SetBorder(true).SetTitle("Filter")
	uiApp.pagesComponent.AddAndSwitchToPage("nameFilter", c, false)

}

func (c *NameFilterComponent) filter() {
	nameFilter := strings.ToLower(strings.TrimSpace(c.nameFilterInputField.GetText()))
	if nameFilter != "" {
		c.uiApp.libraryComponent.currentFilter().contains = &nameFilter
	} else {
		c.uiApp.libraryComponent.currentFilter().contains = nil
	}
	c.uiApp.libraryComponent.RefreshView()
	c.close()
}

func (c *NameFilterComponent) cancel() {
	c.close()
}

func (c *NameFilterComponent) close() {
	c.uiApp.pagesComponent.RemovePage("nameFilter")
	c.uiApp.tviewApp.SetFocus(c.uiApp.libraryComponent)
}
