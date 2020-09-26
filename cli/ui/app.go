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

type App struct {
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

func NewApp(clientConfig config.ClientConfig, restClient *restClientV1.RestClient) *App {
	app := &App{
		ClientConfig: clientConfig,
		restClient:   restClient,
		localDb:      db.NewLocalDb(restClient, clientConfig.Collator()),
	}

	app.tviewApp = tview.NewApplication()

	app.libraryComponent = NewLibraryComponent(app)

	app.currentComponent = NewCurrentComponent(app)

	app.helpComponent = NewHelpComponent(app)

	app.playerComponent = NewPlayerComponent(app, 100)

	app.messageComponent = NewMessageComponent(app)

	app.pagesComponent = tview.NewPages()

	app.mainLayout = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(
			tview.NewFlex().
				SetDirection(tview.FlexColumn).
				AddItem(app.libraryComponent, 0, 1, true).
				AddItem(app.currentComponent, 0, 1, false),
			0, 1, false).
		AddItem(app.playerComponent, 2, 0, false)

	app.pagesComponent.AddAndSwitchToPage("main", app.mainLayout, true)

	app.globalLayout = tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(
			tview.NewFlex().
				SetDirection(tview.FlexRow).
				AddItem(app.pagesComponent, 0, 1, true).
				AddItem(app.messageComponent, 1, 0, false),
			0, 2, true)

	app.tviewApp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if app.mainLayout.HasFocus() && !app.libraryComponent.nameFilterInputField.HasFocus() {
			switch {
			case event.Key() == tcell.KeyF5:
				app.Reload()
			case event.Key() == tcell.KeyEsc:
				app.ConfirmExit()
				return nil
			case event.Key() == tcell.KeyRune:
				switch event.Rune() {
				case 'h':
					app.switchHelpView()
					return nil
				case 'p':
					app.playerComponent.PauseResume()
					return nil
				case '+':
					app.playerComponent.VolumeUp()
					return nil
				case '-':
					app.playerComponent.VolumeDown()
					return nil
				}
			case event.Key() == tcell.KeyTab:
				switch {
				case app.libraryComponent.HasFocus():
					app.libraryComponent.Disable()
					app.currentComponent.Enable()
					app.playerComponent.Disable()
					app.tviewApp.SetFocus(app.currentComponent)
					return nil
				case app.currentComponent.HasFocus():
					app.libraryComponent.Disable()
					app.currentComponent.Disable()
					app.playerComponent.Enable()
					app.tviewApp.SetFocus(app.playerComponent)
					return nil
				case app.playerComponent.HasFocus():
					app.libraryComponent.Enable()
					app.currentComponent.Disable()
					app.playerComponent.Disable()
					app.tviewApp.SetFocus(app.libraryComponent)
					return nil
				}
			case event.Key() == tcell.KeyBacktab:
				switch {
				case app.playerComponent.HasFocus():
					app.libraryComponent.Disable()
					app.currentComponent.Enable()
					app.playerComponent.Disable()
					app.tviewApp.SetFocus(app.currentComponent)
					return nil
				case app.currentComponent.HasFocus():
					app.libraryComponent.Enable()
					app.currentComponent.Disable()
					app.playerComponent.Disable()
					app.tviewApp.SetFocus(app.libraryComponent)
					return nil
				case app.libraryComponent.HasFocus():
					app.libraryComponent.Disable()
					app.currentComponent.Disable()
					app.playerComponent.Enable()
					app.tviewApp.SetFocus(app.playerComponent)
					return nil
				}
			}
		}
		return event
	})

	app.tviewApp.SetRoot(app.globalLayout, true)

	app.libraryComponent.Enable()
	app.currentComponent.Disable()
	app.playerComponent.Disable()
	app.tviewApp.SetFocus(app.libraryComponent)

	return app
}

func (a *App) Start() {
	logrus.Debugf("Starting console user interface ...")

	// Refresh Db from Server
	a.Reload()

	// Start event loop
	if err := a.tviewApp.SetFocus(a.libraryComponent).Run(); err != nil {
		panic(err)
	}

}

func (a *App) ConfirmExit() {
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

func (a *App) ConfirmSongDelete(song *restApiV1.Song) {
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

func (a *App) ConfirmArtistDelete(artist *restApiV1.Artist) {
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

func (a *App) ConfirmAlbumDelete(album *restApiV1.Album) {
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

func (a *App) ConfirmPlaylistDelete(playlist *restApiV1.Playlist) {
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

func (a *App) ConfirmUserDelete(user *restApiV1.User) {
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

func (a *App) Message(message string) {
	a.messageComponent.SetMessage(message)
}
func (a *App) ForceMessage(message string) {
	a.Message(message)
	a.tviewApp.ForceDraw()
}
func (a *App) WarningMessage(message string) {
	a.messageComponent.SetWarningMessage("! " + message)
}
func (a *App) ClientErrorMessage(message string, cliErr restClientV1.ClientError) {
	a.messageComponent.SetWarningMessage("! " + message + " (" + cliErr.Code().String() + ")")
}

func (a *App) Reload() {

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

func (a *App) LocalDb() *db.LocalDb {
	return a.localDb
}

func (a *App) Play(songId restApiV1.SongId) {
	a.playerComponent.Play(songId)
}

func (a *App) CurrentComponent() *CurrentComponent {
	return a.currentComponent
}

func (a *App) IsConnectedUserAdmin() bool {
	if user, ok := a.localDb.Users[a.ConnectedUserId()]; ok == true {
		return user.AdminFg
	}
	return false
}

func (a *App) HideExplicitSongForConnectedUser() bool {
	if user, ok := a.localDb.Users[a.ConnectedUserId()]; ok == true {
		return user.HideExplicitFg
	}
	return false
}

func (a *App) ConnectedUserId() restApiV1.UserId {
	return a.restClient.UserId()
}

func (a *App) switchHelpView() {
	if a.showHelp {
		a.globalLayout.RemoveItem(a.helpComponent)
	} else {
		a.globalLayout.AddItem(a.helpComponent, 0, 1, false)
	}
	a.showHelp = !a.showHelp

}
