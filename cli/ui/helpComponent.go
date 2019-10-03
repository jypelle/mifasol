package ui

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type HelpComponent struct {
	*tview.Flex
	title   *tview.TextView
	content *tview.TextView

	uiApp *UIApp
}

func NewHelpComponent(uiApp *UIApp) *HelpComponent {

	c := &HelpComponent{
		uiApp: uiApp,
	}

	c.title = tview.NewTextView()
	c.title.SetDynamicColors(true)
	c.title.SetText("[" + ColorTitleStr + "] Help")
	c.title.SetBackgroundColor(tcell.NewHexColor(0xd0d0d0))

	c.content = tview.NewTextView()
	c.content.SetDynamicColors(true)
	c.content.SetBackgroundColor(tcell.NewHexColor(0xd0d0d0))
	c.content.SetTextColor(ColorHelpText)
	c.content.SetText(`
[` + ColorHelpTitleStr + `::u]Global shortcuts

'h'    : Show/Hide this sideview
'+'    : Increase volume
'-'    : Decrease Volume
<TAB>  : Change view
<F5>   : Get updates from lyrasrv
<ESC>  : Quit

[` + ColorHelpTitle2Str + `::u]"Library" shortcuts

'c'    : Create album / artist
'e'    : Edit song / album / artist / playlist
'd'    : Delete song / album / artist / playlist
'a'    : Add song / album / artist / playlist to Current playlist
'l'    : Load song / album / artist / playlist to Current playlist
<LEFT> : Previous item
<RIGHT>: Next item
<ENTER>: Play song / Artist's songs / Album's songs / Playlist's songs
<BACK> : Go back

[` + ColorHelpTitle2Str + `::u]"Playlist" shortcuts

'c'    : Clear
'r'    : Shuffle songs
'd'    : Remove song
's'    : Save to existing or new playlist
'8'    : Move up highlighted song
'2'    : Move down highlighted song
<ENTER>: Play song

[` + ColorHelpTitle2Str + `::u]"Player" shortcuts

...
`,
	)
	c.content.SetScrollable(true)
	c.content.ScrollToBeginning()

	c.Flex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(c.title, 1, 0, false).
		AddItem(c.content, 0, 1, false)

	return c
}
