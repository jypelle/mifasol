package cliwa

import (
	"github.com/jypelle/mifasol/internal/cliwa/jst"
	"github.com/jypelle/mifasol/restApiV1"
	"sort"
	"strings"
	"syscall/js"
)

type HomePlaylistEditComponent struct {
	app          *App
	playlistId   restApiV1.PlaylistId
	playlistMeta *restApiV1.PlaylistMeta
	closed       bool
}

func NewHomePlaylistEditComponent(app *App, playlistId restApiV1.PlaylistId, playlistMeta *restApiV1.PlaylistMeta) *HomePlaylistEditComponent {
	c := &HomePlaylistEditComponent{
		app:          app,
		playlistId:   playlistId,
		playlistMeta: playlistMeta.Copy(),
	}

	return c
}

func (c *HomePlaylistEditComponent) Show() {
	div := jst.Id("homeMainModal")
	div.Set("innerHTML", c.app.RenderTemplate(
		c.playlistMeta, "home/playlistEdit/index"),
	)

	form := jst.Id("playlistEditForm")
	form.Call("addEventListener", "submit", c.app.AddEventFuncPreventDefault(c.saveAction))
	cancelButton := jst.Id("playlistEditCancelButton")
	cancelButton.Call("addEventListener", "click", c.app.AddEventFunc(c.cancelAction))

	// Owners
	ownerCurrentList := jst.Id("playlistEditOwnerCurrentList")
	ownerSearchInput := jst.Id("playlistEditOwnerSearchInput")
	ownerSearchClean := jst.Id("playlistEditOwnerSearchClean")
	ownerSearchList := jst.Id("playlistEditOwnerSearchList")

	// Remove owner
	ownerCurrentList.Call("addEventListener", "click", c.app.AddRichEventFunc(func(this js.Value, i []js.Value) {
		link := i[0].Get("target").Call("closest", ".userLink")
		if !link.Truthy() {
			return
		}
		dataset := link.Get("dataset")

		switch link.Get("className").String() {
		case "userLink":
			userId := restApiV1.UserId(dataset.Get("userid").String())

			for idx, ownerUserId := range c.playlistMeta.OwnerUserIds {
				if ownerUserId == userId {
					if idx == len(c.playlistMeta.OwnerUserIds)-1 {
						c.playlistMeta.OwnerUserIds = c.playlistMeta.OwnerUserIds[0:idx]
					} else {
						c.playlistMeta.OwnerUserIds = append(c.playlistMeta.OwnerUserIds[0:idx], c.playlistMeta.OwnerUserIds[idx+1:]...)
					}

					break
				}
			}

			// Refresh current owners
			c.refreshCurrentOwnerAction()
		}
	}))

	// Search owner
	ownerSearchInput.Call("addEventListener", "keypress", c.app.AddBlockingRichEventFunc(func(this js.Value, i []js.Value) {
		if i[0].Get("which").Int() == 13 {
			i[0].Call("preventDefault")
		}
	}))
	ownerSearchInput.Call("addEventListener", "input", c.app.AddEventFunc(c.ownerSearchAction))
	ownerSearchInput.Call("addEventListener", "focusout", c.app.AddRichEventFunc(func(this js.Value, i []js.Value) {
		relatedTarget := i[0].Get("relatedTarget")
		if relatedTarget.Truthy() && relatedTarget.Call("closest", ".userLink").Truthy() {
			return
		}
		// Clear search input
		ownerSearchInput.Set("value", "")
		c.ownerSearchAction()
	}))
	ownerSearchClean.Call("addEventListener", "click", c.app.AddEventFunc(func() {
		// Clear search input
		ownerSearchInput.Set("value", "")
		c.ownerSearchAction()
	}))

	// Add owner
	ownerSearchList.Call("addEventListener", "click", c.app.AddRichEventFunc(func(this js.Value, i []js.Value) {
		link := i[0].Get("target").Call("closest", ".userLink")
		if !link.Truthy() {
			return
		}
		dataset := link.Get("dataset")

		switch link.Get("className").String() {
		case "userLink":
			userId := restApiV1.UserId(dataset.Get("userid").String())
			c.playlistMeta.OwnerUserIds = append(c.playlistMeta.OwnerUserIds, userId)

			// Clear search input
			ownerSearchInput.Set("value", "")
			c.ownerSearchAction()

			// Refresh current owners
			c.refreshCurrentOwnerAction()
		}
	}))

	c.refreshCurrentOwnerAction()
}

func (c *HomePlaylistEditComponent) saveAction() {
	if c.closed {
		return
	}

	c.app.ShowLoader("Updating playlist")

	playlistName := jst.Id("playlistEditPlaylistName")
	c.playlistMeta.Name = playlistName.Get("value").String()

	_, cliErr := c.app.restClient.UpdatePlaylist(c.playlistId, c.playlistMeta)
	if cliErr != nil {
		c.app.HomeComponent.MessageComponent.ClientErrorMessage("Unable to update the playlist", cliErr)
	}

	c.close()
	c.app.HomeComponent.Reload()
	c.app.HideLoader()
}

func (c *HomePlaylistEditComponent) cancelAction() {
	if c.closed {
		return
	}
	c.close()
}

func (c *HomePlaylistEditComponent) close() {
	c.closed = true
	c.app.HomeComponent.CloseModal()
}

func (c *HomePlaylistEditComponent) refreshCurrentOwnerAction() {
	type OwnerCurrentItem struct {
		UserId   restApiV1.UserId
		UserName string
	}

	var resultOwnerList []*OwnerCurrentItem

	for _, userId := range c.playlistMeta.OwnerUserIds {
		ownerCurrentItem := &OwnerCurrentItem{
			UserId:   userId,
			UserName: c.app.localDb.Users[userId].Name,
		}

		resultOwnerList = append(resultOwnerList, ownerCurrentItem)
	}

	ownerCurrentList := jst.Id("playlistEditOwnerCurrentList")
	ownerCurrentList.Set("innerHTML", c.app.RenderTemplate(
		resultOwnerList, "home/playlistEdit/ownerCurrentList"),
	)
}

func (c *HomePlaylistEditComponent) ownerSearchAction() {
	ownerSearchInput := jst.Id("playlistEditOwnerSearchInput")
	ownerSearchList := jst.Id("playlistEditOwnerSearchList")

	nameFilter := strings.TrimSpace(ownerSearchInput.Get("value").String())

	type UserSearchItem struct {
		UserId   restApiV1.UserId
		UserName string
	}

	var resultUserList []*UserSearchItem

	if nameFilter != "" {
		lowerNameFilter := strings.ToLower(nameFilter)
		for _, user := range c.app.localDb.OrderedUsers {

			if user == nil || !strings.Contains(strings.ToLower(user.Name), lowerNameFilter) {
				continue
			}

			ownerOfCurrentPlaylist := false
			for _, ownerUserId := range c.playlistMeta.OwnerUserIds {
				if user.Id == ownerUserId {
					ownerOfCurrentPlaylist = true
					break
				}
			}
			if ownerOfCurrentPlaylist {
				continue
			}

			userSearchItem := &UserSearchItem{
				UserId:   user.Id,
				UserName: user.Name,
			}

			resultUserList = append(resultUserList, userSearchItem)
		}

		sort.SliceStable(resultUserList, func(i, j int) bool {
			return len(resultUserList[i].UserName) < len(resultUserList[j].UserName)
		})

		if len(resultUserList) > 100 {
			resultUserList = resultUserList[0:100]
		}

		ownerSearchList.Set("innerHTML", c.app.RenderTemplate(
			struct {
				UserList   []*UserSearchItem
				NameFilter string
			}{
				UserList:   resultUserList,
				NameFilter: nameFilter,
			}, "home/playlistEdit/userSearchList"),
		)
		ownerSearchList.Get("style").Set("display", "block")
	} else {
		ownerSearchList.Set("innerHTML", "")
		ownerSearchList.Get("style").Set("display", "none")
	}
}
