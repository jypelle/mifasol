package ui

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

var ColorTitle = tcell.NewHexColor(0x441800)
var ColorTitleStr = "#441800"

var ColorTitleBackground = tcell.NewHexColor(0xf0f0f0)
var ColorTitleUnfocusedBackground = tcell.NewHexColor(0xa0a0a0)

var ColorArtist = tcell.NewHexColor(0xA0A9CC)
var ColorArtistStr = "#A0A9CC"
var ColorAlbum = tcell.NewHexColor(0x5ADFDF)
var ColorAlbumStr = "#5ADFDF"
var ColorPlaylist = tcell.NewHexColor(0xFFB500)
var ColorPlaylistStr = "#FFB500"
var ColorSong = tcell.NewHexColor(0xFFFFE5)
var ColorSongStr = "#FFFFE5"
var ColorUser = tcell.NewHexColor(0xFFFACD)
var ColorUserStr = "#FFFACD"

var ColorSelected = tcell.NewHexColor(0x602020)
var ColorUnfocusedSelected = tcell.NewHexColor(0x402020)

var ColorEnabled = tcell.NewHexColor(0)
var ColorDisabled = tcell.NewHexColor(0x202024)

var ColorHelpTitleStr = "#402020"
var ColorHelpTitle2Str = "#111111"
var ColorHelpText = tcell.NewHexColor(0)
var ColorHelpTextStr = "#000000"

func init() {
	tview.Styles = tview.Theme{
		PrimitiveBackgroundColor:    tcell.ColorBlack,
		ContrastBackgroundColor:     tcell.NewHexColor(0x403030),
		MoreContrastBackgroundColor: tcell.ColorGreen,
		BorderColor:                 tcell.NewHexColor(0x808080),
		TitleColor:                  tcell.ColorWhite,
		GraphicsColor:               tcell.ColorWhite,
		PrimaryTextColor:            tcell.ColorWhite,
		SecondaryTextColor:          tcell.NewHexColor(0xFF6040),
		TertiaryTextColor:           tcell.ColorGreen,
		InverseTextColor:            tcell.ColorBlue,
		ContrastSecondaryTextColor:  tcell.ColorDarkCyan,
	}
}

/*

#####  Palette URL: http://paletton.com/#uid=30k0D0kllll5fBtdirKtoeWLh8x

*** Primary color:

   shade 0 = #AA6039 = rgb(170, 96, 57) = rgba(170, 96, 57,1) = rgb0(0.667,0.376,0.224)
   shade 1 = #FFE4D5 = rgb(255,228,213) = rgba(255,228,213,1) = rgb0(1,0.894,0.835)
   shade 2 = #DDA181 = rgb(221,161,129) = rgba(221,161,129,1) = rgb0(0.867,0.631,0.506)
   shade 3 = #77300A = rgb(119, 48, 10) = rgba(119, 48, 10,1) = rgb0(0.467,0.188,0.039)
   shade 4 = #441800 = rgb( 68, 24,  0) = rgba( 68, 24,  0,1) = rgb0(0.267,0.094,0)

*** Secondary color (1):

   shade 0 = #2E4272 = rgb( 46, 66,114) = rgba( 46, 66,114,1) = rgb0(0.18,0.259,0.447)
   shade 1 = #A0A9BC = rgb(160,169,188) = rgba(160,169,188,1) = rgb0(0.627,0.663,0.737)
   shade 2 = #5D6E94 = rgb( 93,110,148) = rgba( 93,110,148,1) = rgb0(0.365,0.431,0.58)
   shade 3 = #0E2250 = rgb( 14, 34, 80) = rgba( 14, 34, 80,1) = rgb0(0.055,0.133,0.314)
   shade 4 = #020F2D = rgb(  2, 15, 45) = rgba(  2, 15, 45,1) = rgb0(0.008,0.059,0.176)

*** Secondary color (2):

   shade 0 = #6B9A33 = rgb(107,154, 51) = rgba(107,154, 51,1) = rgb0(0.42,0.604,0.2)
   shade 1 = #DAECC5 = rgb(218,236,197) = rgba(218,236,197,1) = rgb0(0.855,0.925,0.773)
   shade 2 = #A2C975 = rgb(162,201,117) = rgba(162,201,117,1) = rgb0(0.635,0.788,0.459)
   shade 3 = #3E6C09 = rgb( 62,108,  9) = rgba( 62,108,  9,1) = rgb0(0.243,0.424,0.035)
   shade 4 = #213E00 = rgb( 33, 62,  0) = rgba( 33, 62,  0,1) = rgb0(0.129,0.243,0)

*/
