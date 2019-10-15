package ui

import (
	"github.com/rivo/tview"
	"mifasol/restApiV1"
)

type UserEditComponent struct {
	*tview.Form
	nameInputField     *tview.InputField
	passwordInputField *tview.InputField
	adminCheckBox      *tview.Checkbox
	uiApp              *UIApp
	userId             string
	userMeta           *restApiV1.UserMeta
	originPrimitive    tview.Primitive
}

func OpenUserCreateComponent(uiApp *UIApp, originPrimitive tview.Primitive) {
	OpenUserEditComponent(uiApp, "", &restApiV1.UserMeta{}, originPrimitive)
}

func OpenUserEditComponent(uiApp *UIApp, userId string, userMeta *restApiV1.UserMeta, originPrimitive tview.Primitive) {

	// Only admin can create or edit another user
	if uiApp.ConnectedUserId() != userId && !uiApp.IsConnectedUserAdmin() {
		uiApp.WarningMessage("Only administrator can create or edit another user")
		return
	}

	c := &UserEditComponent{
		uiApp:           uiApp,
		userId:          userId,
		userMeta:        userMeta,
		originPrimitive: originPrimitive,
	}

	c.Form = tview.NewForm()
	c.nameInputField = tview.NewInputField().
		SetLabel("Name").
		SetText(userMeta.Name).
		SetFieldWidth(50)
	c.Form.AddFormItem(c.nameInputField)

	c.passwordInputField = tview.NewInputField().
		SetLabel("Password").
		SetText("").
		SetFieldWidth(50)
	c.Form.AddFormItem(c.passwordInputField)

	c.adminCheckBox = tview.NewCheckbox().
		SetLabel("Administrator").
		SetChecked(userMeta.AdminFg)
	if uiApp.IsConnectedUserAdmin() {
		c.Form.AddFormItem(c.adminCheckBox)
	}

	c.Form.AddButton("Save", c.save)
	c.Form.AddButton("Cancel", c.cancel)
	if c.userId != "" {
		c.Form.SetBorder(true).SetTitle("Edit user")
	} else {
		c.Form.SetBorder(true).SetTitle("Create user")
	}

	uiApp.pagesComponent.AddAndSwitchToPage("userEdit", c, true)
}

func (c *UserEditComponent) save() {
	userMetaComplete := &restApiV1.UserMetaComplete{*c.userMeta, c.passwordInputField.GetText()}
	userMetaComplete.Name = c.nameInputField.GetText()

	// Non-admin user can't change admin user flag
	if c.uiApp.IsConnectedUserAdmin() {
		userMetaComplete.AdminFg = c.adminCheckBox.IsChecked()
	}

	if c.userId != "" {
		c.uiApp.restClient.UpdateUser(c.userId, userMetaComplete)
		// Update username/password stored in config file on self edit
		if c.uiApp.ConnectedUserId() == c.userId {
			c.uiApp.ClientEditableConfig.Username = userMetaComplete.Name
			if userMetaComplete.Password != "" {
				c.uiApp.ClientEditableConfig.Password = userMetaComplete.Password
			}
			c.uiApp.ClientConfig.Save()
		}

	} else {

		// Can't create user without password
		if userMetaComplete.Password == "" {
			return
		}

		_, err := c.uiApp.restClient.CreateUser(userMetaComplete)
		if err != nil {
			return
		}
	}
	c.uiApp.Reload()

	c.close()
}

func (c *UserEditComponent) cancel() {
	c.close()
}

func (c *UserEditComponent) close() {
	c.uiApp.pagesComponent.RemovePage("userEdit")
	c.uiApp.tviewApp.SetFocus(c.originPrimitive)
}
