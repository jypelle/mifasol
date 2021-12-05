package ui

import (
	"code.rocketnine.space/tslocum/cview"
	"github.com/jypelle/mifasol/restApiV1"
)

type UserEditComponent struct {
	*cview.Form
	nameInputField     *cview.InputField
	passwordInputField *cview.InputField
	hideExplicitBox    *cview.CheckBox
	adminCheckBox      *cview.CheckBox
	uiApp              *App
	userId             restApiV1.UserId
	userMeta           *restApiV1.UserMeta
	originPrimitive    cview.Primitive
}

func OpenUserCreateComponent(uiApp *App, originPrimitive cview.Primitive) {
	OpenUserEditComponent(uiApp, "", &restApiV1.UserMeta{}, originPrimitive)
}

func OpenUserEditComponent(uiApp *App, userId restApiV1.UserId, userMeta *restApiV1.UserMeta, originPrimitive cview.Primitive) {

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

	c.Form = cview.NewForm()
	c.Form.SetFieldTextColorFocused(cview.Styles.PrimitiveBackgroundColor)
	c.Form.SetFieldBackgroundColorFocused(cview.Styles.PrimaryTextColor)

	c.nameInputField = cview.NewInputField()
	c.nameInputField.SetLabel("Name")
	c.nameInputField.SetText(userMeta.Name)
	c.nameInputField.SetFieldWidth(50)
	c.Form.AddFormItem(c.nameInputField)

	c.passwordInputField = cview.NewInputField()
	c.passwordInputField.SetLabel("Password")
	c.passwordInputField.SetText("")
	c.passwordInputField.SetFieldWidth(50)
	c.Form.AddFormItem(c.passwordInputField)

	c.hideExplicitBox = cview.NewCheckBox()
	c.hideExplicitBox.SetLabel("Hide explicit songs")
	c.hideExplicitBox.SetChecked(userMeta.HideExplicitFg)
	if uiApp.IsConnectedUserAdmin() {
		c.Form.AddFormItem(c.hideExplicitBox)
	}

	c.adminCheckBox = cview.NewCheckBox()
	c.adminCheckBox.SetLabel("Administrator")
	c.adminCheckBox.SetChecked(userMeta.AdminFg)
	if uiApp.IsConnectedUserAdmin() {
		c.Form.AddFormItem(c.adminCheckBox)
	}

	c.Form.AddButton("Save", c.save)
	c.Form.AddButton("Cancel", c.cancel)
	if c.userId != "" {
		c.Form.SetBorder(true)
		c.Form.SetTitle("Edit user")
	} else {
		c.Form.SetBorder(true)
		c.Form.SetTitle("Create user")
	}

	uiApp.pagesComponent.AddAndSwitchToPage("userEdit", c, true)
}

func (c *UserEditComponent) save() {
	userMetaComplete := &restApiV1.UserMetaComplete{*c.userMeta, c.passwordInputField.GetText()}
	userMetaComplete.Name = c.nameInputField.GetText()

	// Non-admin user can't change *hide explicit* or *admin user* flag
	if c.uiApp.IsConnectedUserAdmin() {
		userMetaComplete.AdminFg = c.adminCheckBox.IsChecked()
		userMetaComplete.HideExplicitFg = c.hideExplicitBox.IsChecked()
	}

	if c.userId != "" {
		_, cliErr := c.uiApp.restClient.UpdateUser(c.userId, userMetaComplete)
		if cliErr != nil {
			c.uiApp.ClientErrorMessage("Unable to update the user", cliErr)
			return
		}

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

		_, cliErr := c.uiApp.restClient.CreateUser(userMetaComplete)
		if cliErr != nil {
			c.uiApp.ClientErrorMessage("Unable to create the user", cliErr)
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
	c.uiApp.cviewApp.SetFocus(c.originPrimitive)
}
