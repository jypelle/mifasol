package ui

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/jypelle/mifasol/internal/cli/config"
	"github.com/jypelle/mifasol/internal/localdb"
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/jypelle/mifasol/restClientV1"
	"github.com/sirupsen/logrus"
	"gitlab.com/tslocum/cview"
	"strconv"
)

type App struct {
	config.ClientConfig
	restClient *restClientV1.RestClient

	// In memory db
	localDb *localdb.LocalDb

	// region View component
	cviewApp         *cview.Application
	mainLayout       *cview.Flex
	globalLayout     *cview.Flex
	pagesComponent   *cview.Pages
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
		localDb:      localdb.NewLocalDb(restClient, clientConfig.Collator()),
	}

	app.cviewApp = cview.NewApplication()

	app.libraryComponent = NewLibraryComponent(app)

	app.currentComponent = NewCurrentComponent(app)

	app.helpComponent = NewHelpComponent(app)

	app.playerComponent = NewPlayerComponent(app, 100)

	app.messageComponent = NewMessageComponent(app)

	app.pagesComponent = cview.NewPages()

	// Activate mouse
	//app.cviewApp.EnableMouse(true)

	mainLayoutFlex := cview.NewFlex()
	mainLayoutFlex.SetDirection(cview.FlexColumn)
	mainLayoutFlex.AddItem(app.libraryComponent, 0, 1, true)
	mainLayoutFlex.AddItem(app.currentComponent, 0, 1, false)

	app.mainLayout = cview.NewFlex()
	app.mainLayout.SetDirection(cview.FlexRow)
	app.mainLayout.AddItem(mainLayoutFlex, 0, 1, false)
	app.mainLayout.AddItem(app.playerComponent, 1, 0, false)

	app.pagesComponent.AddAndSwitchToPage("main", app.mainLayout, true)

	globalLayoutFlex := cview.NewFlex()
	globalLayoutFlex.SetDirection(cview.FlexRow)
	globalLayoutFlex.AddItem(app.pagesComponent, 0, 1, true)
	globalLayoutFlex.AddItem(app.messageComponent, 1, 0, false)

	app.globalLayout = cview.NewFlex()
	app.globalLayout.SetDirection(cview.FlexColumn)
	app.globalLayout.AddItem(globalLayoutFlex, 0, 2, true)

	app.cviewApp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if app.mainLayout.HasFocus() && !app.libraryComponent.nameFilterInputField.HasFocus() {
			switch {
			case event.Key() == tcell.KeyF5:
				app.Reload()
			case event.Key() == tcell.KeyEsc:
				app.ConfirmExit()
				return nil
			case event.Modifiers()&tcell.ModCtrl > 0 && event.Key() == tcell.KeyLeft:
				app.playerComponent.GoBackward()
				return nil
			case event.Modifiers()&tcell.ModCtrl > 0 && event.Key() == tcell.KeyRight:
				app.playerComponent.GoForward()
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
					app.cviewApp.SetFocus(app.currentComponent)
					return nil
				case app.currentComponent.HasFocus():
					app.libraryComponent.Enable()
					app.currentComponent.Disable()
					app.cviewApp.SetFocus(app.libraryComponent)
					return nil
				}
			case event.Key() == tcell.KeyBacktab:
				switch {
				case app.libraryComponent.HasFocus():
					app.libraryComponent.Disable()
					app.currentComponent.Enable()
					app.cviewApp.SetFocus(app.currentComponent)
					return nil
				case app.currentComponent.HasFocus():
					app.libraryComponent.Enable()
					app.currentComponent.Disable()
					app.cviewApp.SetFocus(app.libraryComponent)
					return nil
				}
			}
		}
		return event
	})

	app.cviewApp.SetRoot(app.globalLayout, true)

	app.libraryComponent.Enable()
	app.currentComponent.Disable()
	app.cviewApp.SetFocus(app.libraryComponent)

	return app
}

func (a *App) Start() {
	logrus.Debugf("Starting console user interface ...")

	// Refresh Db from Server
	fmt.Println("Syncing...")
	a.Reload()

	// Start event loop
	a.cviewApp.SetFocus(a.libraryComponent)
	if err := a.cviewApp.Run(); err != nil {
		panic(err)
	}

}

func (a *App) ConfirmExit() {
	currentFocus := a.cviewApp.GetFocus()

	modal := cview.NewModal()
	modal.SetText("Do you want to quit ?")
	modal.AddButtons([]string{"Quit", "Cancel"})
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Quit" {
			a.cviewApp.Stop()
		} else {
			a.pagesComponent.HidePage("exitConfirm")
			a.pagesComponent.RemovePage("exitConfirm")
			a.cviewApp.SetFocus(currentFocus)
		}
	})

	a.pagesComponent.AddPage("exitConfirm", modal, false, true)
}

func (a *App) ConfirmSongDelete(song *restApiV1.Song) {
	// Only admin can delete a song
	if !a.IsConnectedUserAdmin() {
		a.WarningMessage("Only administrator can delete this song")
		return
	}

	currentFocus := a.cviewApp.GetFocus()

	modal := cview.NewModal()
	modal.SetText("Do you want to delete \"" + song.Name + "\" ?")
	modal.AddButtons([]string{"Yes", "Cancel"})
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Yes" {
			a.restClient.DeleteSong(song.Id)
			a.Reload()
		}
		a.pagesComponent.HidePage("songDeleteConfirm")
		a.pagesComponent.RemovePage("songDeleteConfirm")
		a.cviewApp.SetFocus(currentFocus)
	})

	a.pagesComponent.AddPage(
		"songDeleteConfirm",
		modal,
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

	currentFocus := a.cviewApp.GetFocus()
	modal := cview.NewModal()
	modal.SetText("Do you want to delete \"" + artist.Name + "\" ?")
	modal.AddButtons([]string{"Yes", "Cancel"})
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Yes" {
			_, cliErr := a.restClient.DeleteArtist(artist.Id)
			if cliErr != nil {
				if cliErr.Code() == restApiV1.DeleteArtistWithSongsErrorCode {
					a.WarningMessage("You should first delete or unlink artist's songs")
				} else {
					a.ClientErrorMessage("Unable to delete the artist", cliErr)
				}
			} else {
				a.Reload()
			}
		}
		a.pagesComponent.HidePage("artistDeleteConfirm")
		a.pagesComponent.RemovePage("artistDeleteConfirm")
		a.cviewApp.SetFocus(currentFocus)
	})

	a.pagesComponent.AddPage("artistDeleteConfirm", modal, false, true)
}

func (a *App) ConfirmAlbumDelete(album *restApiV1.Album) {
	// Only admin can delete an album
	if !a.IsConnectedUserAdmin() {
		a.WarningMessage("Only administrator can delete this album")
		return
	}

	currentFocus := a.cviewApp.GetFocus()
	modal := cview.NewModal()
	modal.SetText("Do you want to delete \"" + album.Name + "\" ?")
	modal.AddButtons([]string{"Yes", "Cancel"})
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Yes" {
			_, cliErr := a.restClient.DeleteAlbum(album.Id)
			if cliErr != nil {
				if cliErr.Code() == restApiV1.DeleteAlbumWithSongsErrorCode {
					a.WarningMessage("You should first delete or unlink album's songs")
				} else {
					a.ClientErrorMessage("Unable to delete the album", cliErr)
				}
			} else {
				a.Reload()
			}
		}
		a.pagesComponent.HidePage("albumDeleteConfirm")
		a.pagesComponent.RemovePage("albumDeleteConfirm")
		a.cviewApp.SetFocus(currentFocus)
	})
	a.pagesComponent.AddPage("albumDeleteConfirm", modal, false, true)
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

	currentFocus := a.cviewApp.GetFocus()
	modal := cview.NewModal()
	modal.SetText("Do you want to delete \"" + playlist.Name + "\" ?")
	modal.AddButtons([]string{"Yes", "Cancel"})
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Yes" {
			a.restClient.DeletePlaylist(playlist.Id)
			a.Reload()
		}
		a.pagesComponent.HidePage("playlistDeleteConfirm")
		a.pagesComponent.RemovePage("playlistDeleteConfirm")
		a.cviewApp.SetFocus(currentFocus)
	})
	a.pagesComponent.AddPage("playlistDeleteConfirm", modal, false, true)
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

	currentFocus := a.cviewApp.GetFocus()
	modal := cview.NewModal()
	modal.SetText("Do you want to delete \"" + user.Name + "\" ?")
	modal.AddButtons([]string{"Yes", "Cancel"})
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Yes" {
			a.restClient.DeleteUser(user.Id)
			a.Reload()
		}
		a.pagesComponent.HidePage("userDeleteConfirm")
		a.pagesComponent.RemovePage("userDeleteConfirm")
		a.cviewApp.SetFocus(currentFocus)
	})
	a.pagesComponent.AddPage("userDeleteConfirm", modal, false, true)
}

func (a *App) Message(message string) {
	a.messageComponent.SetMessage(message)
}
func (a *App) ForceMessage(message string) {
	a.Message(message)
	a.cviewApp.Draw()
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
		a.ClientErrorMessage("Unable to load data from mifasolsrv", cliErr)
		return
	}

	a.libraryComponent.RefreshView()
	a.currentComponent.RefreshView()

	a.Message(strconv.Itoa(len(a.localDb.Songs)) + " songs, " + strconv.Itoa(len(a.localDb.Artists)) + " artists, " + strconv.Itoa(len(a.localDb.Albums)) + " albums, " + strconv.Itoa(len(a.localDb.Playlists)) + " playlists ready to be played for " + strconv.Itoa(len(a.localDb.Users)) + " users.")
}

func (a *App) LocalDb() *localdb.LocalDb {
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
