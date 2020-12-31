package ui

import (
	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cview"
)

type HelpComponent struct {
	*cview.Flex
	//	title   *cview.TextView
	content *cview.TextView

	uiApp *App
}

func NewHelpComponent(uiApp *App) *HelpComponent {

	c := &HelpComponent{
		uiApp: uiApp,
	}
	c.content = cview.NewTextView()
	c.content.SetDynamicColors(true)
	c.content.SetBackgroundColor(tcell.NewHexColor(0xd0d0d0))
	c.content.SetTextColor(ColorHelpText)
	c.content.SetText(`[` + ColorHelpTitleStr + `::u]Global shortcuts[-::-]

'h'    : Show / Hide this sideview
'p'    : Play / Pause
'+'    : Increase volume
'-'    : Decrease Volume
<TAB>  : Switch view
<F5>   : Get updates from mifasolsrv
<ESC>  : Quit

[` + ColorHelpTitle2Str + `::u]"Library" shortcuts[-::-]

'c'    : Create album / artist
'e'    : Edit song / album / artist / playlist
'd'    : Delete song / album / artist / playlist
'a'    : Add song / album / artist / playlist to current playlist
'l'    : Load song / album / artist / playlist to current playlist
'f'    : Add to / Remove from favorite songs / playlists
'/'    : Filter by song / album / artist name
<LEFT> : Previous item
<RIGHT>: Next item
<ENTER>: Play song / Artist's songs / Album's songs / Playlist's songs
<BACK> : Go back

[` + ColorHelpTitle2Str + `::u]"Playlist" shortcuts[-::-]

'c'    : Clear
'r'    : Shuffle songs
'd'    : Remove song
's'    : Save to existing or new playlist
'8'    : Move up highlighted song
'2'    : Move down highlighted song
<ENTER>: Play song

[` + ColorHelpTitle2Str + `::u]"Player" shortcuts[-::-]

...
`,
	)
	c.content.SetScrollable(true)
	c.content.ScrollToBeginning()

	c.Flex = cview.NewFlex()
	c.Flex.SetDirection(cview.FlexRow)
	//		c.Flex.AddItem(c.title, 1, 0, false)
	c.Flex.AddItem(c.content, 0, 1, false)

	return c
}
