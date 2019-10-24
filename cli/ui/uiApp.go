package ui

import (
	"github.com/gdamore/tcell"
	"github.com/jypelle/mifasol/cli/config"
	"github.com/jypelle/mifasol/cli/db"
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/jypelle/mifasol/restClientV1"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"strconv"
)

type UIApp struct {
	config.ClientConfig
	restClient *restClientV1.RestClient

	// In memory db
	localDb *db.LocalDb

	// region View component
	tviewApp         *tview.Application
	mainLayout       *tview.Flex
	globalLayout     *tview.Flex
	pagesComponent   *tview.Pages
	libraryComponent *LibraryComponent
	currentComponent *CurrentComponent
	playerComponent  *PlayerComponent
	messageComponent *MessageComponent
	helpComponent    *HelpComponent
	// endregion

	showHelp bool
}

func NewUIApp(clientConfig config.ClientConfig, restClient *restClientV1.RestClient) *UIApp {
	uiApp := &UIApp{
		ClientConfig: clientConfig,
		restClient:   restClient,
		localDb:      db.NewLocalDb(restClient, clientConfig.Collator()),
	}

	uiApp.tviewApp = tview.NewApplication()

	uiApp.libraryComponent = NewLibraryComponent(uiApp)

	uiApp.currentComponent = NewCurrentComponent(uiApp)

	uiApp.helpComponent = NewHelpComponent(uiApp)

	uiApp.playerComponent = NewPlayerComponent(uiApp, 100)

	uiApp.messageComponent = NewMessageComponent(uiApp)

	uiApp.pagesComponent = tview.NewPages()

	uiApp.mainLayout = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(
			tview.NewFlex().
				SetDirection(tview.FlexColumn).
				AddItem(uiApp.libraryComponent, 0, 1, true).
				AddItem(uiApp.currentComponent, 0, 1, false),
			0, 1, false).
		AddItem(uiApp.playerComponent, 2, 0, false)

	uiApp.pagesComponent.AddAndSwitchToPage("main", uiApp.mainLayout, true)

	uiApp.globalLayout = tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(
			tview.NewFlex().
				SetDirection(tview.FlexRow).
				AddItem(uiApp.pagesComponent, 0, 1, true).
				AddItem(uiApp.messageComponent, 1, 0, false),
			0, 2, true)

	uiApp.tviewApp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if uiApp.mainLayout.HasFocus() {
			switch {
			case event.Key() == tcell.KeyF5:
				uiApp.Reload()
			case event.Key() == tcell.KeyEsc:
				uiApp.ConfirmExit()
				return nil
			case event.Key() == tcell.KeyRune:
				switch event.Rune() {
				case 'h':
					uiApp.switchHelpView()
					return nil
				case 'p':
					uiApp.playerComponent.PauseResume()
					return nil
				case '+':
					uiApp.playerComponent.VolumeUp()
					return nil
				case '-':
					uiApp.playerComponent.VolumeDown()
					return nil
				}
			case event.Key() == tcell.KeyTab:
				switch {
				case uiApp.libraryComponent.HasFocus():
					uiApp.libraryComponent.Disable()
					uiApp.currentComponent.Enable()
					uiApp.playerComponent.Disable()
					uiApp.tviewApp.SetFocus(uiApp.currentComponent)
					return nil
				case uiApp.currentComponent.HasFocus():
					uiApp.libraryComponent.Disable()
					uiApp.currentComponent.Disable()
					uiApp.playerComponent.Enable()
					uiApp.tviewApp.SetFocus(uiApp.playerComponent)
					return nil
				case uiApp.playerComponent.HasFocus():
					uiApp.libraryComponent.Enable()
					uiApp.currentComponent.Disable()
					uiApp.playerComponent.Disable()
					uiApp.tviewApp.SetFocus(uiApp.libraryComponent)
					return nil
				}
			case event.Key() == tcell.KeyBacktab:
				switch {
				case uiApp.playerComponent.HasFocus():
					uiApp.libraryComponent.Disable()
					uiApp.currentComponent.Enable()
					uiApp.playerComponent.Disable()
					uiApp.tviewApp.SetFocus(uiApp.currentComponent)
					return nil
				case uiApp.currentComponent.HasFocus():
					uiApp.libraryComponent.Enable()
					uiApp.currentComponent.Disable()
					uiApp.playerComponent.Disable()
					uiApp.tviewApp.SetFocus(uiApp.libraryComponent)
					return nil
				case uiApp.libraryComponent.HasFocus():
					uiApp.libraryComponent.Disable()
					uiApp.currentComponent.Disable()
					uiApp.playerComponent.Enable()
					uiApp.tviewApp.SetFocus(uiApp.playerComponent)
					return nil
				}
			}
		}
		return event
	})

	uiApp.tviewApp.SetRoot(uiApp.globalLayout, true)

	uiApp.libraryComponent.Enable()
	uiApp.currentComponent.Disable()
	uiApp.playerComponent.Disable()
	uiApp.tviewApp.SetFocus(uiApp.libraryComponent)

	return uiApp
}

func (a *UIApp) Start() {
	logrus.Debugf("Starting console user interface ...")

	// Refresh Db from Server
	a.Reload()

	// Start event loop
	if err := a.tviewApp.SetFocus(a.libraryComponent).Run(); err != nil {
		panic(err)
	}

}

func (a *UIApp) ConfirmExit() {
	currentFocus := a.tviewApp.GetFocus()
	a.pagesComponent.AddPage(
		"exitConfirm",
		tview.NewModal().
			SetText("Do you want to quit ?").
			AddButtons([]string{"Quit", "Cancel"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				if buttonLabel == "Quit" {
					a.tviewApp.Stop()
				} else {
					a.pagesComponent.HidePage("exitConfirm").RemovePage("exitConfirm")
					a.tviewApp.SetFocus(currentFocus)
				}
			}),
		false,
		true,
	)
}

func (a *UIApp) ConfirmSongDelete(song *restApiV1.Song) {
	// Only admin can delete a song
	if !a.IsConnectedUserAdmin() {
		a.WarningMessage("Only administrator can delete this song")
		return
	}

	currentFocus := a.tviewApp.GetFocus()
	a.pagesComponent.AddPage(
		"songDeleteConfirm",
		tview.NewModal().
			SetText("Do you want to delete \""+song.Name+"\" ?").
			AddButtons([]string{"Yes", "Cancel"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				if buttonLabel == "Yes" {
					a.restClient.DeleteSong(song.Id)
					a.Reload()
				}
				a.pagesComponent.HidePage("songDeleteConfirm").RemovePage("songDeleteConfirm")
				a.tviewApp.SetFocus(currentFocus)
			}),
		false,
		true,
	)
}

func (a *UIApp) ConfirmArtistDelete(artist *restApiV1.Artist) {
	// Only admin can delete an artist
	if !a.IsConnectedUserAdmin() {
		a.WarningMessage("Only administrator can delete this artist")
		return
	}

	currentFocus := a.tviewApp.GetFocus()
	a.pagesComponent.AddPage(
		"artistDeleteConfirm",
		tview.NewModal().
			SetText("Do you want to delete \""+artist.Name+"\" ?").
			AddButtons([]string{"Yes", "Cancel"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				if buttonLabel == "Yes" {
					a.restClient.DeleteArtist(artist.Id)
					a.Reload()
				}
				a.pagesComponent.HidePage("artistDeleteConfirm").RemovePage("artistDeleteConfirm")
				a.tviewApp.SetFocus(currentFocus)
			}),
		false,
		true,
	)
}

func (a *UIApp) ConfirmAlbumDelete(album *restApiV1.Album) {
	// Only admin can delete an album
	if !a.IsConnectedUserAdmin() {
		a.WarningMessage("Only administrator can delete this album")
		return
	}

	currentFocus := a.tviewApp.GetFocus()
	a.pagesComponent.AddPage(
		"albumDeleteConfirm",
		tview.NewModal().
			SetText("Do you want to delete \""+album.Name+"\" ?").
			AddButtons([]string{"Yes", "Cancel"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				if buttonLabel == "Yes" {
					a.restClient.DeleteAlbum(album.Id)
					a.Reload()
				}
				a.pagesComponent.HidePage("albumDeleteConfirm").RemovePage("albumDeleteConfirm")
				a.tviewApp.SetFocus(currentFocus)
			}),
		false,
		true,
	)
}

func (a *UIApp) ConfirmPlaylistDelete(playlist *restApiV1.Playlist) {
	// Only admin or playlist owner can delete a playlist
	if !a.IsConnectedUserAdmin() && !a.localDb.IsPlaylistOwnedBy(playlist.Id, a.ConnectedUserId()) {
		a.WarningMessage("Only administrator or owner can delete this playlist")
		return
	}

	if playlist.Id == restApiV1.IncomingPlaylistId {
		a.WarningMessage("Incoming playlist cannot be deleted")
		return
	}

	currentFocus := a.tviewApp.GetFocus()
	a.pagesComponent.AddPage(
		"playlistDeleteConfirm",
		tview.NewModal().
			SetText("Do you want to delete \""+playlist.Name+"\" ?").
			AddButtons([]string{"Yes", "Cancel"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				if buttonLabel == "Yes" {
					a.restClient.DeletePlaylist(playlist.Id)
					a.Reload()
				}
				a.pagesComponent.HidePage("playlistDeleteConfirm").RemovePage("playlistDeleteConfirm")
				a.tviewApp.SetFocus(currentFocus)
			}),
		false,
		true,
	)
}

func (a *UIApp) ConfirmUserDelete(user *restApiV1.User) {
	// Only admin can delete a user
	if !a.IsConnectedUserAdmin() {
		a.WarningMessage("Only administrator can delete a user")
		return
	}
	// You can't delete yourself
	if a.ConnectedUserId() == user.Id {
		a.WarningMessage("Sorry, you can't delete yourself")
		return
	}

	currentFocus := a.tviewApp.GetFocus()
	a.pagesComponent.AddPage(
		"userDeleteConfirm",
		tview.NewModal().
			SetText("Do you want to delete \""+user.Name+"\" ?").
			AddButtons([]string{"Yes", "Cancel"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				if buttonLabel == "Yes" {
					a.restClient.DeleteUser(user.Id)
					a.Reload()
				}
				a.pagesComponent.HidePage("userDeleteConfirm").RemovePage("userDeleteConfirm")
				a.tviewApp.SetFocus(currentFocus)
			}),
		false,
		true,
	)
}

func (a *UIApp) Message(message string) {
	a.messageComponent.SetMessage(message)
}
func (a *UIApp) ForceMessage(message string) {
	a.Message(message)
	a.tviewApp.ForceDraw()
}
func (a *UIApp) WarningMessage(message string) {
	a.messageComponent.SetWarningMessage("! " + message)
}
func (a *UIApp) ClientErrorMessage(message string, cliErr restClientV1.ClientError) {
	a.messageComponent.SetWarningMessage("! " + message + " (" + cliErr.Code().String() + ")")
}

func (a *UIApp) Reload() {

	a.ForceMessage("Syncing...")
	// Refresh In memory Db
	cliErr := a.localDb.Refresh()
	if cliErr != nil {
		a.ClientErrorMessage("Unable to refresh data from mifasolsrv", cliErr)
		return
	}

	a.libraryComponent.RefreshView()
	a.currentComponent.RefreshView()

	a.Message(strconv.Itoa(len(a.localDb.Songs)) + " songs, " + strconv.Itoa(len(a.localDb.Artists)) + " artists, " + strconv.Itoa(len(a.localDb.Albums)) + " albums, " + strconv.Itoa(len(a.localDb.Playlists)) + " playlists ready to be played for " + strconv.Itoa(len(a.localDb.Users)) + " users.")
}

func (a *UIApp) LocalDb() *db.LocalDb {
	return a.localDb
}

func (a *UIApp) Play(songId restApiV1.SongId) {
	a.playerComponent.Play(songId)
}

func (a *UIApp) CurrentComponent() *CurrentComponent {
	return a.currentComponent
}

func (a *UIApp) IsConnectedUserAdmin() bool {
	if user, ok := a.localDb.Users[a.ConnectedUserId()]; ok == true {
		return user.AdminFg
	}
	return false
}

func (a *UIApp) ConnectedUserId() restApiV1.UserId {
	return a.restClient.UserId()
}

func (a *UIApp) switchHelpView() {
	if a.showHelp {
		a.globalLayout.RemoveItem(a.helpComponent)
	} else {
		a.globalLayout.AddItem(a.helpComponent, 0, 1, false)
	}
	a.showHelp = !a.showHelp

}
