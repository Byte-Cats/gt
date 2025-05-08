package internal

import (
	"github.com/charmbracelet/bubbles/key"
)

// KeyMap defines custom keybindings
type KeyMap struct {
	Up            key.Binding
	Down          key.Binding
	PageUp        key.Binding
	PageDown      key.Binding
	GoToTop       key.Binding
	GoToBottom    key.Binding
	SelectEnter   key.Binding
	SelectSpace   key.Binding
	Back          key.Binding // Go to parent directory
	Quit          key.Binding
	ConfirmYes    key.Binding
	ConfirmNo     key.Binding
	StartFilter   key.Binding // Start filtering
	ClearFilter   key.Binding // Clear/cancel filter (ESC)
	AddBookmark   key.Binding // Add current directory to bookmarks
	ShowBookmarks key.Binding // Show bookmarks list
	ToggleHidden  key.Binding // Toggle showing hidden files
	Copy          key.Binding // Copy selected file/directory
	Cut           key.Binding // Cut selected file/directory
	Paste         key.Binding // Paste copied/cut file/directory
	Delete        key.Binding // Delete selected file/directory
	Rename        key.Binding // Rename selected file/directory
	TogglePreview key.Binding // Toggle preview pane
	Help          key.Binding // Show help screen

	// Image preview mode bindings
	ImgPrevNext       key.Binding // Next image
	ImgPrevPrev       key.Binding // Previous image
	ImgPrevSizeUp     key.Binding // Increase image size
	ImgPrevSizeDown   key.Binding // Decrease image size
	ImgPrevAlignL     key.Binding // Align image left
	ImgPrevAlignC     key.Binding // Align image center
	ImgPrevAlignR     key.Binding // Align image right
	ImgPrevAspect     key.Binding // Toggle aspect ratio preservation
	ImgPrevPersist    key.Binding // Toggle persistence
	ImgPrevReset      key.Binding // Reset image options
	ImgPrevWidthMode  key.Binding // Switch to width adjustment mode
	ImgPrevHeightMode key.Binding // Switch to height adjustment mode
	ImgPrevMaxHeight  key.Binding // Toggle max-height limit
	ImgPrevZIndex     key.Binding // Adjust z-index
	ImgFitToWidth     key.Binding // Fit image to terminal width, preserving aspect ratio
}

// DefaultKeyMap returns the default keybindings
var DefaultKeyMap = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	PageUp: key.NewBinding(
		key.WithKeys("pgup"),
		key.WithHelp("pgup", "page up"),
	),
	PageDown: key.NewBinding(
		key.WithKeys("pgdown"),
		key.WithHelp("pgdown", "page down"),
	),
	GoToTop: key.NewBinding(
		key.WithKeys("home", "g"),
		key.WithHelp("home/g", "go to top"),
	),
	GoToBottom: key.NewBinding(
		key.WithKeys("end", "G"),
		key.WithHelp("end/G", "go to bottom"),
	),
	SelectEnter: key.NewBinding(
		key.WithKeys("enter", "l"),
		key.WithHelp("enter/l", "select"),
	),
	SelectSpace: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "select"),
	),
	Back: key.NewBinding(
		key.WithKeys("backspace", "h", "left"),
		key.WithHelp("←/h/bs", "parent dir"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q/ctrl+c", "quit"),
	),
	ConfirmYes: key.NewBinding(
		key.WithKeys("y"),
		key.WithHelp("y", "confirm"),
	),
	ConfirmNo: key.NewBinding(
		key.WithKeys("n", "esc"),
		key.WithHelp("n/esc", "cancel"),
	),
	StartFilter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
	ClearFilter: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "clear filter/cancel"),
	),
	AddBookmark: key.NewBinding(
		key.WithKeys("b"),
		key.WithHelp("b", "bookmark"),
	),
	ShowBookmarks: key.NewBinding(
		key.WithKeys("B"),
		key.WithHelp("B", "bookmarks"),
	),
	ToggleHidden: key.NewBinding(
		key.WithKeys("H"),
		key.WithHelp("H", "toggle hidden"),
	),
	Copy: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "copy"),
	),
	Cut: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "cut"),
	),
	Paste: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "paste"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete"),
	),
	Rename: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "rename"),
	),
	TogglePreview: key.NewBinding(
		key.WithKeys("v"),
		key.WithHelp("v", "preview"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),

	// Image preview mode bindings
	ImgPrevNext: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("j/↓", "next image"),
	),
	ImgPrevPrev: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("k/↑", "previous image"),
	),
	ImgPrevSizeUp: key.NewBinding(
		key.WithKeys("+", "="),
		key.WithHelp("+/=", "increase size"),
	),
	ImgPrevSizeDown: key.NewBinding(
		key.WithKeys("-"),
		key.WithHelp("-", "decrease size"),
	),
	ImgPrevAlignL: key.NewBinding(
		key.WithKeys("l"),
		key.WithHelp("l", "align left"),
	),
	ImgPrevAlignC: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "align center"),
	),
	ImgPrevAlignR: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "align right"),
	),
	ImgPrevAspect: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "toggle aspect ratio"),
	),
	ImgPrevPersist: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "toggle persistence"),
	),
	ImgPrevReset: key.NewBinding(
		key.WithKeys("0"),
		key.WithHelp("0", "reset options"),
	),
	ImgPrevWidthMode: key.NewBinding(
		key.WithKeys("w"),
		key.WithHelp("w", "width mode"),
	),
	ImgPrevHeightMode: key.NewBinding(
		key.WithKeys("h"),
		key.WithHelp("h", "height mode"),
	),
	ImgPrevMaxHeight: key.NewBinding(
		key.WithKeys("m"),
		key.WithHelp("m", "toggle max-height"),
	),
	ImgPrevZIndex: key.NewBinding(
		key.WithKeys("z"),
		key.WithHelp("z", "increase z-index"),
	),
	ImgFitToWidth: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "fit to width"),
	),
}
