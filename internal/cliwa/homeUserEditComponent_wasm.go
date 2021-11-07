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
	div.Set("innerHTML", c.app.RenderTemplate(
		c.userMetaComplete, "home/userEdit/index"),
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

	c.app.ShowLoader("Updating all songs of the user")

	userName := jst.Id("userEditUserName")
	c.userMetaComplete.Name = userName.Get("value").String()

	if c.userId != "" {
		_, cliErr := c.app.restClient.UpdateUser(c.userId, c.userMetaComplete)
		if cliErr != nil {
			c.app.HomeComponent.MessageComponent.ClientErrorMessage("Unable to update the user", cliErr)
		}

		// Update username/password stored in config on self edit
		if c.app.ConnectedUserId() == c.userId {
			c.app.config.ClientEditableConfig.Username = c.userMetaComplete.Name
			if c.userMetaComplete.Password != "" {
				c.app.config.ClientEditableConfig.Password = c.userMetaComplete.Password
			}
		}
	} else {
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
