package primitive

import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"math"
	"strconv"
	"strings"
)

// richListItem represents one item in a RichList.
type richListItem struct {
	MainText string // The main text of the list item.
}

// RichList displays rows of items, each of which can be selected.
//
// See https://github.com/rivo/tview/wiki/List for an example.
type RichList struct {
	*tview.Box

	//
	infiniteScroll bool

	// The items of the list.
	items []*richListItem

	// The index of the currently selected item.
	currentItem int

	// The item main text color.
	mainTextColor tcell.Color

	// The background color for unfocused selected items.
	unfocusedSelectedBackgroundColor tcell.Color

	// The background color for selected items.
	selectedBackgroundColor tcell.Color

	// If true, the entire row is highlighted when selected.
	highlightFullLine bool

	// If true, the row is prefixed with the line number
	prefixWithLineNumber bool

	// The number of list items skipped at the top before the first item is drawn.
	offset int

	// An optional function which is called when the user has navigated to a list
	// item.
	changed func(index int, mainText string)

	highlightedMainText func(index int, mainText string) string
}

// NewRichList returns a new list.
func NewRichList() *RichList {
	return &RichList{
		Box:                              tview.NewBox(),
		infiniteScroll:                   false,
		mainTextColor:                    tview.Styles.PrimaryTextColor,
		selectedBackgroundColor:          tview.Styles.PrimaryTextColor,
		unfocusedSelectedBackgroundColor: tview.Styles.PrimitiveBackgroundColor,
	}
}

// SetCurrentItem sets the currently selected item by its index, starting at 0
// for the first item. If a negative index is provided, items are referred to
// from the back (-1 = last item, -2 = second-to-last item, and so on). Out of
// range indices are clamped to the beginning/end.
//
// Calling this function triggers a "changed" event if the selection changes.
func (l *RichList) SetCurrentItem(index int) *RichList {
	if index < 0 {
		index = len(l.items) + index
	}
	if index >= len(l.items) {
		index = len(l.items) - 1
	}
	if index < 0 {
		index = 0
	}

	if index != l.currentItem && l.changed != nil {
		item := l.items[index]
		l.changed(index, item.MainText)
	}

	l.currentItem = index

	return l
}

// GetCurrentItem returns the index of the currently selected list item,
// starting at 0 for the first item.
func (l *RichList) GetCurrentItem() int {
	return l.currentItem
}

// RemoveItem removes the item with the given index (starting at 0) from the
// list. If a negative index is provided, items are referred to from the back
// (-1 = last item, -2 = second-to-last item, and so on). Out of range indices
// are clamped to the beginning/end, i.e. unless the list is empty, an item is
// always removed.
//
// The currently selected item is shifted accordingly. If it is the one that is
// removed, a "changed" event is fired.
func (l *RichList) RemoveItem(index int) *RichList {
	if len(l.items) == 0 {
		return l
	}

	// Adjust index.
	if index < 0 {
		index = len(l.items) + index
	}
	if index >= len(l.items) {
		index = len(l.items) - 1
	}
	if index < 0 {
		index = 0
	}

	// Remove item.
	l.items = append(l.items[:index], l.items[index+1:]...)

	// If there is nothing left, we're done.
	if len(l.items) == 0 {
		return l
	}

	// Shift current item.
	previousCurrentItem := l.currentItem
	if l.currentItem > index {
		l.currentItem--
	}

	// Fire "changed" event for removed items.
	if previousCurrentItem == index && l.changed != nil {
		item := l.items[l.currentItem]
		l.changed(l.currentItem, item.MainText)
	}

	return l
}

// SetInfiniteScroll sets a flag which determines whether the scroll should be endless or not.
func (l *RichList) SetInfiniteScroll(infiniteScroll bool) *RichList {
	l.infiniteScroll = infiniteScroll
	return l
}

// SetMainTextColor sets the color of the items' main text.
func (l *RichList) SetMainTextColor(color tcell.Color) *RichList {
	l.mainTextColor = color
	return l
}

// SetUnfocusedSelectedBackgroundColor sets the background color of unfocused selected items.
func (l *RichList) SetUnfocusedSelectedBackgroundColor(color tcell.Color) *RichList {
	l.unfocusedSelectedBackgroundColor = color
	return l
}

// SetSelectedBackgroundColor sets the background color of selected items.
func (l *RichList) SetSelectedBackgroundColor(color tcell.Color) *RichList {
	l.selectedBackgroundColor = color
	return l
}

// SetHighlightFullLine sets a flag which determines whether the colored
// background of selected items spans the entire width of the view. If set to
// true, the highlight spans the entire view. If set to false, only the text of
// the selected item from beginning to end is highlighted.
func (l *RichList) SetHighlightFullLine(highlight bool) *RichList {
	l.highlightFullLine = highlight
	return l
}

func (l *RichList) SetPrefixWithLineNumber(prefixWithLineNumber bool) *RichList {
	l.prefixWithLineNumber = prefixWithLineNumber
	return l
}

// SetChangedFunc sets the function which is called when the user navigates to
// a list item. The function receives the item's index in the list of items
// (starting with 0), its main text, secondary text, and its shortcut rune.
//
// This function is also called when the first item is added or when
// SetCurrentItem() is called.
func (l *RichList) SetChangedFunc(handler func(index int, mainText string)) *RichList {
	l.changed = handler
	return l
}

func (l *RichList) SetHighlightedMainTextFunc(handler func(index int, mainText string) string) *RichList {
	l.highlightedMainText = handler
	return l
}

// AddItem calls InsertItem() with an index of -1.
func (l *RichList) AddItem(mainText string) *RichList {
	l.InsertItem(-1, mainText)
	return l
}

// InsertItem adds a new item to the list at the specified index. An index of 0
// will insert the item at the beginning, an index of 1 before the second item,
// and so on. An index of GetItemCount() or higher will insert the item at the
// end of the list. Negative indices are also allowed: An index of -1 will
// insert the item at the end of the list, an index of -2 before the last item,
// and so on. An index of -GetItemCount()-1 or lower will insert the item at the
// beginning.
//
// An item has a main text which will be highlighted when selected. It also has
// a secondary text which is shown underneath the main text (if it is set to
// visible) but which may remain empty.
//
// The shortcut is a key binding. If the specified rune is entered, the item
// is selected immediately. Set to 0 for no binding.
//
// The "selected" callback will be invoked when the user selects the item. You
// may provide nil if no such callback is needed or if all events are handled
// through the selected callback set with SetSelectedFunc().
//
// The currently selected item will shift its position accordingly. If the list
// was previously empty, a "changed" event is fired because the new item becomes
// selected.
func (l *RichList) InsertItem(index int, mainText string) *RichList {
	item := &richListItem{
		MainText: mainText,
	}

	// Shift index to range.
	if index < 0 {
		index = len(l.items) + index + 1
	}
	if index < 0 {
		index = 0
	} else if index > len(l.items) {
		index = len(l.items)
	}

	// Shift current item.
	if l.currentItem < len(l.items) && l.currentItem >= index {
		l.currentItem++
	}

	// Insert item (make space for the new item, then shift and insert).
	l.items = append(l.items, nil)
	if index < len(l.items)-1 { // -1 because l.items has already grown by one item.
		copy(l.items[index+1:], l.items[index:])
	}
	l.items[index] = item

	// Fire a "change" event for the first item in the list.
	if len(l.items) == 1 && l.changed != nil {
		item := l.items[0]
		l.changed(0, item.MainText)
	}

	return l
}

// GetItemCount returns the number of items in the list.
func (l *RichList) GetItemCount() int {
	return len(l.items)
}

// GetItemText returns an item's texts (main). Panics if the index
// is out of range.
func (l *RichList) GetItemText(index int) (main string) {
	return l.items[index].MainText
}

// SetItemText sets an item's main text. Panics if the index is
// out of range.
func (l *RichList) SetItemText(index int, main string) *RichList {
	item := l.items[index]
	item.MainText = main
	return l
}

// FindItems searches the main and secondary texts for the given strings and
// returns a list of item indices in which those strings are found. One of the
// two search strings may be empty, it will then be ignored. Indices are always
// returned in ascending order.
//
// If mustContainBoth is set to true, mainSearch must be contained in the main
// text AND secondarySearch must be contained in the secondary text. If it is
// false, only one of the two search strings must be contained.
//
// Set ignoreCase to true for case-insensitive search.
func (l *RichList) FindItems(mainSearch string, ignoreCase bool) (indices []int) {
	if mainSearch == "" {
		return
	}

	if ignoreCase {
		mainSearch = strings.ToLower(mainSearch)
	}

	for index, item := range l.items {
		mainText := item.MainText
		if ignoreCase {
			mainText = strings.ToLower(mainText)
		}

		// strings.Contains() always returns true for a "" search.
		mainContained := strings.Contains(mainText, mainSearch)
		if mainText != "" && mainContained {
			indices = append(indices, index)
		}
	}

	return
}

// Clear removes all items from the list.
func (l *RichList) Clear() *RichList {
	l.items = nil
	l.currentItem = 0
	l.offset = 0
	return l
}

// Draw draws this primitive onto the screen.
func (l *RichList) Draw(screen tcell.Screen) {
	l.Box.Draw(screen)

	// Determine the dimensions.
	x, y, width, height := l.GetInnerRect()
	bottomLimit := y + height

	// Adjust offset to keep the current selection in view.
	if l.currentItem < l.offset {
		l.offset = l.currentItem
	} else {
		if l.currentItem-l.offset >= height {
			l.offset = l.currentItem + 1 - height
		}
	}

	// Draw the list items.
	libraryListSizeDigit := strconv.Itoa(int(math.Log10(float64(len(l.items)))) + 1)

	for index, item := range l.items {
		if index < l.offset {
			continue
		}

		if y >= bottomLimit {
			break
		}

		var mainText string

		// Prefix with line number
		if l.prefixWithLineNumber {
			mainText = "[#505050]" + fmt.Sprintf("%"+libraryListSizeDigit+"d", index+1) + " [white]"
		}

		// Background color of selected text.
		if index == l.currentItem {
			// Main text.
			if l.highlightedMainText == nil {
				mainText += item.MainText
			} else {
				mainText += l.highlightedMainText(index, item.MainText)
			}
			tview.Print(screen, mainText, x, y, width, tview.AlignLeft, l.mainTextColor)

			textWidth := width
			if !l.highlightFullLine {
				if w := tview.TaggedStringWidth(item.MainText); w < textWidth {
					textWidth = w
				}
			}

			for bx := 0; bx < textWidth; bx++ {
				m, c, style, _ := screen.GetContent(x+bx, y)
				fg, _, _ := style.Decompose()
				if l.HasFocus() {
					style = style.Background(l.selectedBackgroundColor).Foreground(fg)
				} else {
					style = style.Background(l.unfocusedSelectedBackgroundColor).Foreground(fg)
				}
				screen.SetContent(x+bx, y, m, c, style)
			}
		} else {
			// Main text.
			mainText += item.MainText
			tview.Print(screen, mainText, x, y, width, tview.AlignLeft, l.mainTextColor)
		}

		y++

		if y >= bottomLimit {
			break
		}

	}
}

// InputHandler returns the handler for this primitive.
func (l *RichList) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return l.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		previousItem := l.currentItem

		switch key := event.Key(); key {
		case tcell.KeyDown:
			l.currentItem++
		case tcell.KeyUp:
			l.currentItem--
		case tcell.KeyHome:
			l.currentItem = 0
		case tcell.KeyEnd:
			l.currentItem = len(l.items) - 1
		case tcell.KeyPgDn:
			l.currentItem += 10
		case tcell.KeyPgUp:
			l.currentItem -= 10
		}

		if l.infiniteScroll {
			if l.currentItem < 0 {
				l.currentItem = len(l.items) - 1
			} else if l.currentItem >= len(l.items) {
				l.currentItem = 0
			}
		} else {
			if l.currentItem < 0 {
				l.currentItem = 0
			} else if l.currentItem >= len(l.items) {
				l.currentItem = len(l.items) - 1
			}
		}
		if l.currentItem < 0 {
			l.currentItem = 0
		}

		if l.currentItem != previousItem && l.currentItem < len(l.items) && l.changed != nil {
			item := l.items[l.currentItem]
			l.changed(l.currentItem, item.MainText)
		}
	})
}
