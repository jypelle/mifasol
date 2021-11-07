package cliwa

import (
	"fmt"
	"github.com/jypelle/mifasol/internal/cliwa/jst"
	"github.com/jypelle/mifasol/restApiV1"
	"github.com/sirupsen/logrus"
	"html"
	"strconv"
)

type HomePlayerComponent struct {
	app                   *App
	volume                float64
	muted                 bool
	autoRefreshSeekSlider bool
}

func NewHomePlayerComponent(app *App) *HomePlayerComponent {
	c := &HomePlayerComponent{
		app:                   app,
		volume:                1,
		muted:                 false,
		autoRefreshSeekSlider: true,
	}

	return c
}

func (c *HomePlayerComponent) Show() {
	playerPlayButton := jst.Id("playerPlayButton")
	playerNextButton := jst.Id("playerNextButton")
	playerAudio := jst.Id("playerAudio")
	playerSeekSlider := jst.Id("playerSeekSlider")
	playerCurrentTime := jst.Id("playerCurrentTime")
	playerDuration := jst.Id("playerDuration")
	playerMuteButton := jst.Id("playerMuteButton")
	playerVolumeSlider := jst.Id("playerVolumeSlider")

	playerAudio.Call("addEventListener", "ended", c.app.AddEventFunc(c.app.HomeComponent.CurrentComponent.PlayNextSongAction))
	playerAudio.Call("addEventListener", "loadedmetadata", c.app.AddEventFunc(func() {
		duration := playerAudio.Get("duration").Int()
		logrus.Infof("duration: %d", duration)
		playerDuration.Set("innerHTML", fmt.Sprintf("%d:%02d", duration/60, duration%60))
		playerSeekSlider.Set("max", duration)
		playerSeekSlider.Set("value", 0)
	}))
	playerAudio.Call("addEventListener", "timeupdate", c.app.AddEventFunc(func() {
		currentTime := playerAudio.Get("currentTime").Int()
		playerCurrentTime.Set("innerHTML", fmt.Sprintf("%d:%02d", currentTime/60, currentTime%60))
		if c.autoRefreshSeekSlider {
			playerSeekSlider.Set("value", currentTime)
		}
	}))

	playerPlayButton.Call("addEventListener", "click", c.app.AddEventFunc(func() {
		if playerAudio.Get("paused").Bool() {
			c.ResumeSongAction()
		} else {
			c.PauseSongAction()
		}
	}))

	playerNextButton.Call("addEventListener", "click", c.app.AddEventFunc(c.app.HomeComponent.CurrentComponent.PlayNextSongAction))

	playerSeekSlider.Call("addEventListener", "input", c.app.AddEventFunc(func() {
		c.autoRefreshSeekSlider = false
	}))
	playerSeekSlider.Call("addEventListener", "change", c.app.AddEventFunc(func() {
		newTime, _ := strconv.ParseInt(playerSeekSlider.Get("value").String(), 10, 64)
		logrus.Infof("newTime: %d", newTime)
		playerAudio.Set("currentTime", newTime)
		c.autoRefreshSeekSlider = true
	}))

	playerMuteButton.Call("addEventListener", "click", c.app.AddEventFunc(func() {
		if !c.muted {
			playerMuteButton.Set("innerHTML", `<i class="fas fa-volume-up"></i>`)
			playerAudio.Set("volume", 0)
			c.muted = true
		} else {
			playerMuteButton.Set("innerHTML", `<i class="fas fa-volume-off"></i>`)
			playerAudio.Set("volume", c.volume)
			c.muted = false
		}
	}))

	adjustVolumeFunc := c.app.AddEventFunc(func() {
		c.volume, _ = strconv.ParseFloat(playerVolumeSlider.Get("value").String(), 64)
		if c.muted {
			playerMuteButton.Set("innerHTML", `<i class="fas fa-volume-off"></i>`)
			c.muted = false
		}
		playerAudio.Set("volume", c.volume)
	})
	playerVolumeSlider.Call("addEventListener", "change", adjustVolumeFunc)
	playerVolumeSlider.Call("addEventListener", "input", adjustVolumeFunc)

}

func (c *HomePlayerComponent) PlaySongAction(songId restApiV1.SongId) {
	token, cliErr := c.app.restClient.GetToken()

	if cliErr != nil {
		return
	}

	playerPlayButton := jst.Id("playerPlayButton")
	playerPlayButton.Set("innerHTML", `<i class="fas fa-pause"></i>`)

	player := jst.Id("playerAudio")
	player.Set("src", "/api/v1/songContents/"+string(songId)+"?bearer="+token.AccessToken)
	player.Call("play")

	c.app.HomeComponent.MessageComponent.Message(fmt.Sprintf(`Playing <span class="songLink">%s</span> ...`, html.EscapeString(c.app.localDb.Songs[songId].Name)))

	return
}

func (c *HomePlayerComponent) PauseSongAction() {
	playerPlayButton := jst.Id("playerPlayButton")
	playerPlayButton.Set("innerHTML", `<i class="fas fa-play"></i>`)
	player := jst.Id("playerAudio")
	player.Call("pause")
}

func (c *HomePlayerComponent) ResumeSongAction() {
	playerPlayButton := jst.Id("playerPlayButton")
	playerPlayButton.Set("innerHTML", `<i class="fas fa-pause"></i>`)
	player := jst.Id("playerAudio")
	player.Call("play")
}
