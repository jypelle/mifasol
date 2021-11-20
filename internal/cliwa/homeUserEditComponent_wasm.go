package cliwa

import (
	"github.com/jypelle/mifasol/internal/cliwa/jst"
	"github.com/jypelle/mifasol/restApiV1"
)

type HomeUserEditComponent struct {
	app              *App
	userId           restApiV1.UserId
	userMetaComplete *restApiV1.UserMetaComplete
	closed           bool
}

func NewHomeUserCreateComponent(app *App) *HomeUserEditComponent {
	return NewHomeUserEditComponent(app, "", &restApiV1.UserMeta{})
}

func NewHomeUserEditComponent(app *App, userId restApiV1.UserId, userMeta *restApiV1.UserMeta) *HomeUserEditComponent {
	c := &HomeUserEditComponent{
		app:              app,
		userId:           userId,
		userMetaComplete: &restApiV1.UserMetaComplete{*userMeta.Copy(), ""},
	}

	return c
}

func (c *HomeUserEditComponent) Show() {
	div := jst.Id("homeMainModal")

	userItem := struct {
		*restApiV1.UserMetaComplete
		IsNewUser            bool
		IsConnectedUserAdmin bool
	}{
		UserMetaComplete:     c.userMetaComplete,
		IsNewUser:            c.userId == "",
		IsConnectedUserAdmin: c.app.IsConnectedUserAdmin(),
	}
	div.Set("innerHTML", c.app.RenderTemplate(
		&userItem, "home/userEdit/index"),
	)

	form := jst.Id("userEditForm")
	form.Call("addEventListener", "submit", c.app.AddEventFuncPreventDefault(c.saveAction))
	cancelButton := jst.Id("userEditCancelButton")
	cancelButton.Call("addEventListener", "click", c.app.AddEventFunc(c.cancelAction))

}

func (c *HomeUserEditComponent) saveAction() {
	if c.closed {
		return
	}

	// User name
	userName := jst.Id("userEditUserName")
	c.userMetaComplete.Name = userName.Get("value").String()
	if c.userMetaComplete.Name == "" {
		c.app.HomeComponent.MessageComponent.WarningMessage("Empty user name")
		return
	}

	// Password
	password := jst.Id("userEditPassword")
	c.userMetaComplete.Password = password.Get("value").String()
	if c.userId == "" && c.userMetaComplete.Password == "" {
		c.app.HomeComponent.MessageComponent.WarningMessage("Empty password")
		return
	}

	// Non-admin user can't change *hide explicit* or *admin user* flag
	if c.app.IsConnectedUserAdmin() {
		// Hide explicit songs flag
		c.userMetaComplete.HideExplicitFg = jst.Id("userEditHideExplicitFg").Get("checked").Bool()

		// Administrator flag
		c.userMetaComplete.AdminFg = jst.Id("userEditAdminFg").Get("checked").Bool()
	}

	if c.userId != "" {
		c.app.ShowLoader("Updating user")
		_, cliErr := c.app.restClient.UpdateUser(c.userId, c.userMetaComplete)
		if cliErr != nil {
			c.app.HomeComponent.MessageComponent.ClientErrorMessage("Unable to update the user", cliErr)
		}

		// Update username/password stored on self edit
		if c.app.ConnectedUserId() == c.userId {
			c.app.config.ClientEditableConfig.Username = c.userMetaComplete.Name
			if c.userMetaComplete.Password != "" {
				c.app.config.ClientEditableConfig.Password = c.userMetaComplete.Password
			}
		}
	} else {
		c.app.ShowLoader("Creating user")
		_, cliErr := c.app.restClient.CreateUser(c.userMetaComplete)
		if cliErr != nil {
			c.app.HomeComponent.MessageComponent.ClientErrorMessage("Unable to create the user", cliErr)
		}
	}
	c.close()
	c.app.HomeComponent.Reload()
	c.app.HideLoader()
}

func (c *HomeUserEditComponent) cancelAction() {
	if c.closed {
		return
	}
	c.close()
}

func (c *HomeUserEditComponent) close() {
	c.closed = true
	c.app.HomeComponent.CloseModal()
}
