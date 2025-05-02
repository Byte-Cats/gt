package buffer

import (
	// "fmt"
	"strconv"
	// "strings"
	"unicode/utf8" // Needed for rune handling

	"github.com/mattn/go-runewidth"
)

// --- Color Types ---
const (
	ColorTypeStandard = "standard"
	ColorType256      = "256"
	ColorTypeTrue     = "truecolor" // Placeholder for future
)

// --- SGR Constants ---
const (
	// Attributes
	AttrReset        = 0
	AttrBold         = 1
	AttrUnderline    = 4
	AttrReverse      = 7
	AttrBoldOff      = 22 // Normal intensity, cancels Bold/Dim
	AttrUnderlineOff = 24
	AttrReverseOff   = 27
	// TODO: Add dim, italic, blink, hidden

	// Standard Colors (30-37 Foreground, 40-47 Background)
	FgBlack   = 30
	FgRed     = 31
	FgGreen   = 32
	FgYellow  = 33
	FgBlue    = 34
	FgMagenta = 35
	FgCyan    = 36
	FgWhite   = 37
	FgDefault = 39

	BgBlack   = 40
	BgRed     = 41
	BgGreen   = 42
	BgYellow  = 43
	BgBlue    = 44
	BgMagenta = 45
	BgCyan    = 46
	BgWhite   = 47
	BgDefault = 49

	// Extended color codes
	SetFgColorExt    = 38
	SetBgColorExt    = 48
	ExtColorMode256  = 5
	ExtColorModeTrue = 2 // Placeholder

	// TODO: Add bright colors (90-97 FG, 100-107 BG) and 256/TrueColor support
)

// defaultFgColor represents the default foreground color code.
const defaultFgColor = FgDefault

// defaultBgColor represents the default background color code.
const defaultBgColor = BgDefault

// defaultColorType represents the default color type.
const defaultColorType = ColorTypeStandard

const defaultMaxScrollback = 1000 // Default number of lines to keep in scrollback

// Cell represents a single character cell on the terminal screen
type Cell struct {
	Char        rune
	Width       int // 0=Continuation of wide char, 1=Standard, 2=Start of wide char
	Fg          int
	Bg          int
	FgColorType string
	BgColorType string
	Bold        bool
	Underline   bool
	Reverse     bool
}

// newDefaultCell creates a cell with default attributes.
func newDefaultCell() Cell {
	return Cell{
		Char:        ' ',
		Width:       1, // Default width
		Fg:          defaultFgColor,
		Bg:          defaultBgColor,
		FgColorType: defaultColorType,
		BgColorType: defaultColorType,
		Bold:        false,
		Underline:   false,
		Reverse:     false,
	}
}

// Output represents the state of the terminal screen buffer.
type Output struct {
	grid    [][]Cell // The screen grid [row][col]
	rows    int
	cols    int
	cursorX int
	cursorY int

	// Scrollback
	scrollback      [][]Cell // Circular buffer for scrollback lines
	maxScrollback   int      // Max lines in scrollback
	scrollbackLines int      // Current number of lines stored in scrollback
	scrollbackHead  int      // Index in scrollback slice where the NEXT scrolled line will go
	viewOffset      int      // How many lines the view is scrolled back (0 = live view)

	// Internal state for parsing escape sequences
	parserState   string // e.g., "GROUND", "ESC", "CSI"
	params        []int  // Parameters for CSI sequence
	currentParam  string // Current parameter being built
	privateMarker rune   // Private marker like ?, >, etc.

	// Current graphic rendition attributes
	currentFg          int
	currentBg          int
	currentFgColorType string
	currentBgColorType string
	currentBold        bool
	currentUnderline   bool
	currentReverse     bool
	// TODO: Add other current attributes

	// TODO: Add scrollback buffer
	// TODO: Add SGR state (current colors, bold, etc.)
}

const (
	esc         = '\x1b'
	stateGround = "GROUND"
	stateEsc    = "ESC"
	stateCsi    = "CSI"
)

// NewOutputBuffer creates a new terminal buffer with given dimensions and scrollback.
func NewOutputBuffer(rows, cols int) *Output {
	grid := make([][]Cell, rows)
	for r := range grid {
		grid[r] = make([]Cell, cols)
		for c := range grid[r] {
			grid[r][c] = newDefaultCell()
		}
	}
	// Initialize scrollback with capacity
	scrollback := make([][]Cell, defaultMaxScrollback)

	return &Output{
		grid:               grid,
		rows:               rows,
		cols:               cols,
		cursorX:            0,
		cursorY:            0,
		scrollback:         scrollback,
		maxScrollback:      defaultMaxScrollback,
		scrollbackLines:    0,
		scrollbackHead:     0,
		viewOffset:         0,
		parserState:        stateGround,
		params:             make([]int, 0, 16),
		currentFg:          defaultFgColor,
		currentBg:          defaultBgColor,
		currentFgColorType: defaultColorType,
		currentBgColorType: defaultColorType,
		currentBold:        false,
		currentUnderline:   false,
		currentReverse:     false,
	}
}

// Write processes incoming bytes, handling printable characters and escape sequences.
func (o *Output) Write(p []byte) (n int, err error) {
	bytesProcessed := 0
	for i := 0; i < len(p); {
		byteVal := p[i]
		consumed := 1 // How many bytes consumed in this iteration

		// Always handle escape char first regardless of state
		if byteVal == byte(esc) {
			o.enterState(stateEsc)
			i += consumed
			bytesProcessed += consumed
			continue
		}

		switch o.parserState {
		case stateGround:
			// In ground state, decode runes and handle standard chars
			r, size := utf8.DecodeRune(p[i:])
			consumed = size // Update consumed bytes
			if r == utf8.RuneError {
				// Skip invalid UTF-8
			} else {
				o.handleGroundChar(r)
			}
		case stateEsc:
			o.handleEsc(byteVal)
		case stateCsi:
			o.handleCsi(byteVal)
		default:
			// Should not happen, reset state
			o.enterState(stateGround)
		}

		i += consumed
		bytesProcessed += consumed
	}
	return bytesProcessed, nil
}

// handleGroundChar processes a character when in the default state.
func (o *Output) handleGroundChar(r rune) {
	switch r {
	case '\n': // Newline (Line Feed)
		o.cursorY++
		if o.cursorY >= o.rows {
			o.scrollUp()
			o.cursorY = o.rows - 1
		}
	case '\r': // Carriage Return
		o.cursorX = 0
	case '\b': // Backspace (ASCII BS)
		if o.cursorX > 0 {
			o.cursorX--
		}
		// Note: This just moves the cursor. Some terminals might also erase the character
		// at the previous position, but often the shell handles the erase via subsequent sequences.
	// TODO: Handle other C0 controls like BEL, TAB, etc.
	default:
		o.putChar(r)
	}
}

// putChar places a character at the current cursor position and advances the cursor,
// applying the current attributes and handling rune width.
func (o *Output) putChar(r rune) {
	// Clamp cursor Y first
	if o.cursorY < 0 {
		o.cursorY = 0
	}
	if o.cursorY >= o.rows {
		o.cursorY = o.rows - 1
	}

	rWidth := runewidth.RuneWidth(r)
	if rWidth == 0 {
		// Handle zero-width characters (e.g., combining marks)
		// TODO: Ideally, apply to the character in the *previous* cell.
		// For now, just ignore them to avoid breaking layout.
		return
	}

	// Check for line wrap *before* placing character, considering its width
	if o.cursorX+rWidth > o.cols {
		o.cursorX = 0
		o.cursorY++
		if o.cursorY >= o.rows {
			o.scrollUp()
			o.cursorY = o.rows - 1
		}
	}

	// Clamp cursor X after potential wrap
	if o.cursorX < 0 {
		o.cursorX = 0
	}

	// Place character applying current attributes
	o.grid[o.cursorY][o.cursorX] = Cell{
		Char:        r,
		Width:       rWidth,
		Fg:          o.currentFg,
		Bg:          o.currentBg,
		FgColorType: o.currentFgColorType,
		BgColorType: o.currentBgColorType,
		Bold:        o.currentBold,
		Underline:   o.currentUnderline,
		Reverse:     o.currentReverse,
	}

	// If it was a wide character, mark the next cell as a continuation
	if rWidth == 2 {
		if o.cursorX+1 < o.cols {
			o.grid[o.cursorY][o.cursorX+1] = Cell{
				Char:  ' ', // Or a special marker?
				Width: 0,   // Mark as continuation
				// Inherit background color? Usually yes.
				Bg:          o.currentBg,
				BgColorType: o.currentBgColorType,
				// Other attributes usually default for continuation cell
				Fg:          defaultFgColor,
				FgColorType: defaultColorType,
			}
		}
	}

	o.cursorX += rWidth // Advance cursor by rune width
}

// enterState transitions the parser to a new state, resetting intermediates.
func (o *Output) enterState(newState string) {
	o.parserState = newState
	if newState != stateCsi { // Reset CSI params unless entering CSI
		o.params = o.params[:0] // Clear slice while keeping capacity
		o.currentParam = ""
		o.privateMarker = 0
	}
}

// handleEsc processes a byte following an ESC character.
func (o *Output) handleEsc(b byte) {
	switch b {
	case '[': // Control Sequence Introducer (CSI)
		o.enterState(stateCsi)
	// TODO: Handle other ESC sequences (e.g., OSC, non-CSI controls)
	default:
		// Unhandled ESC sequence, return to ground state
		o.enterState(stateGround)
	}
}

// handleCsi processes a byte within a CSI escape sequence.
func (o *Output) handleCsi(b byte) {
	switch {
	case b >= '0' && b <= '9': // Parameter digit
		o.currentParam += string(b)
	case b == ';': // Parameter separator
		val, err := strconv.Atoi(o.currentParam)
		if err != nil {
			val = 0 // Default parameter if conversion fails or empty
		}
		o.params = append(o.params, val)
		o.currentParam = "" // Reset for next param
	case b >= 0x3c && b <= 0x3f: // Private marker characters (?, >, etc.)
		o.privateMarker = rune(b)
	case b >= 0x40 && b <= 0x7e: // Final byte (command character)
		// Final parameter needs to be added
		val, err := strconv.Atoi(o.currentParam)
		if err != nil {
			if o.currentParam == "" && len(o.params) == 0 {
				// No params provided often implies 1 or default behaviour
				// Handled per-command below
			} else {
				val = 0 // Default param if conversion fails
			}
		}
		o.params = append(o.params, val)

		// Dispatch command
		o.dispatchCsi(rune(b))

		// Return to ground state after processing command
		o.enterState(stateGround)
	// TODO: Intermediate bytes (0x20-0x2F) - not handled here
	default:
		// Invalid byte in CSI sequence, ignore or reset?
		// For now, just return to ground state
		o.enterState(stateGround)
	}
}

// getParam gets the nth parameter (1-based index) or a default value.
func (o *Output) getParam(n int, defaultVal int) int {
	if n <= 0 || n > len(o.params) {
		return defaultVal
	}
	// Handle 0 vs omitted param: CSI often treats 0 and omitted(1) differently.
	// Our current parsing treats empty as 0. For simplicity now, stick with this.
	// A more correct parser might need to distinguish.
	val := o.params[n-1]
	if val == 0 && n == 1 && len(o.params) == 1 && o.currentParam == "" {
		// If only one param was possible and it was empty, often means 1
		// This heuristic is basic and might be wrong for some commands.
		// return 1 // Let's use defaultVal for simplicity now.
	}
	if val == 0 {
		// Some commands treat 0 as 1, others don't. Use defaultVal logic.
	}
	return val
}

// dispatchCsi executes the command based on the final CSI character.
func (o *Output) dispatchCsi(cmd rune) {
	switch cmd {
	case 'A': // CUU - Cursor Up
		n := o.getParam(1, 1)  // Default is 1 line
		o.cursorY -= max(1, n) // Move at least 1 line
		o.clampCursor()
	case 'B': // CUD - Cursor Down
		n := o.getParam(1, 1)
		o.cursorY += max(1, n)
		o.clampCursor()
	case 'C': // CUF - Cursor Forward
		n := o.getParam(1, 1)
		o.cursorX += max(1, n)
		o.clampCursor()
	case 'D': // CUB - Cursor Backward
		n := o.getParam(1, 1)
		o.cursorX -= max(1, n)
		o.clampCursor()
	case 'H', 'f': // CUP - Cursor Position / HVP
		row := o.getParam(1, 1)   // Default row 1
		col := o.getParam(2, 1)   // Default col 1
		o.cursorY = max(0, row-1) // Convert to 0-based
		o.cursorX = max(0, col-1) // Convert to 0-based
		o.clampCursor()
	case 'J': // ED - Erase in Display
		mode := o.getParam(1, 0) // Default 0
		o.eraseInDisplay(mode)
	case 'K': // EL - Erase in Line
		mode := o.getParam(1, 0) // Default 0
		o.eraseInLine(mode)
	case 'm': // SGR - Select Graphic Rendition
		o.handleSgr()
	// TODO: Add more CSI commands (Scrolling, Insert/Delete Chars/Lines, etc.)
	default:
		// Unhandled CSI command
		// fmt.Printf("Unhandled CSI: params=%v, private=%c, cmd=%c\n", o.params, o.privateMarker, cmd)
		break
	}
}

// handleSgr processes SGR (Select Graphic Rendition) escape codes.
func (o *Output) handleSgr() {
	if len(o.params) == 0 {
		o.params = append(o.params, AttrReset)
	}

	i := 0
	for i < len(o.params) {
		param := o.params[i]
		switch param {
		case AttrReset: // Reset all attributes
			o.currentFg = defaultFgColor
			o.currentBg = defaultBgColor
			o.currentFgColorType = defaultColorType
			o.currentBgColorType = defaultColorType
			o.currentBold = false
			o.currentUnderline = false
			o.currentReverse = false
			// TODO: Reset other attributes
		case AttrBold: // Set bold
			o.currentBold = true
		case AttrUnderline: // Set underline
			o.currentUnderline = true
		case AttrReverse: // Set reverse video
			o.currentReverse = true
		case AttrBoldOff: // Turn off Bold/Dim
			o.currentBold = false
		case AttrUnderlineOff: // Turn off underline
			o.currentUnderline = false
		case AttrReverseOff: // Turn off reverse video
			o.currentReverse = false

		// Standard Foreground Colors
		case FgBlack, FgRed, FgGreen, FgYellow, FgBlue, FgMagenta, FgCyan, FgWhite:
			o.currentFg = param
			o.currentFgColorType = ColorTypeStandard
		case FgDefault: // Default foreground color
			o.currentFg = defaultFgColor
			o.currentFgColorType = ColorTypeStandard

		// Standard Background Colors
		case BgBlack, BgRed, BgGreen, BgYellow, BgBlue, BgMagenta, BgCyan, BgWhite:
			o.currentBg = param
			o.currentBgColorType = ColorTypeStandard
		case BgDefault: // Default background color
			o.currentBg = defaultBgColor
			o.currentBgColorType = ColorTypeStandard

		// Extended Colors (38 / 48)
		case SetFgColorExt: // Set foreground color (extended)
			// Need at least mode + 1 value (256) or mode + 3 values (truecolor)
			if i+1 < len(o.params) {
				mode := o.params[i+1]
				if mode == ExtColorMode256 && i+2 < len(o.params) {
					colorIndex := o.params[i+2]
					if colorIndex >= 0 && colorIndex <= 255 {
						o.currentFg = colorIndex
						o.currentFgColorType = ColorType256
					}
					i += 2 // Consumed mode and index
				} else if mode == ExtColorModeTrue && i+4 < len(o.params) {
					r := o.params[i+2]
					g := o.params[i+3]
					b := o.params[i+4]
					if r >= 0 && r <= 255 && g >= 0 && g <= 255 && b >= 0 && b <= 255 {
						o.currentFg = (r << 16) | (g << 8) | b // Pack RGB
						o.currentFgColorType = ColorTypeTrue
					}
					i += 4 // Consumed mode, r, g, b
				} else {
					// Invalid extended color mode or missing params, consume only mode
					i += 1
				}
			}
		case SetBgColorExt: // Set background color (extended)
			if i+1 < len(o.params) {
				mode := o.params[i+1]
				if mode == ExtColorMode256 && i+2 < len(o.params) {
					colorIndex := o.params[i+2]
					if colorIndex >= 0 && colorIndex <= 255 {
						o.currentBg = colorIndex
						o.currentBgColorType = ColorType256
					}
					i += 2
				} else if mode == ExtColorModeTrue && i+4 < len(o.params) {
					r := o.params[i+2]
					g := o.params[i+3]
					b := o.params[i+4]
					if r >= 0 && r <= 255 && g >= 0 && g <= 255 && b >= 0 && b <= 255 {
						o.currentBg = (r << 16) | (g << 8) | b // Pack RGB
						o.currentBgColorType = ColorTypeTrue
					}
					i += 4
				} else {
					i += 1
				}
			}

		// TODO: Handle 256-color and true-color modes (e.g., 38;5;n, 48;5;n, 38;2;r;g;b, 48;2;r;g;b)
		// These require consuming more parameters (i needs careful updating).

		default:
			// Ignored SGR code
		}
		i++ // Move to the next parameter
	}
}

// clampCursor ensures cursor coordinates are within grid boundaries.
func (o *Output) clampCursor() {
	if o.cursorX < 0 {
		o.cursorX = 0
	}
	if o.cursorX >= o.cols {
		o.cursorX = o.cols - 1 // Clamp to last column
	}
	if o.cursorY < 0 {
		o.cursorY = 0
	}
	if o.cursorY >= o.rows {
		o.cursorY = o.rows - 1 // Clamp to last row
	}
}

// eraseInDisplay handles ED (Erase in Display) commands.
// Ensures cleared cells have default attributes.
func (o *Output) eraseInDisplay(mode int) {
	switch mode {
	case 0: // Erase from cursor to end of screen
		for c := o.cursorX; c < o.cols; c++ {
			o.grid[o.cursorY][c] = newDefaultCell()
		}
		for r := o.cursorY + 1; r < o.rows; r++ {
			for c := 0; c < o.cols; c++ {
				o.grid[r][c] = newDefaultCell()
			}
		}
	case 1: // Erase from beginning of screen to cursor
		for r := 0; r < o.cursorY; r++ {
			for c := 0; c < o.cols; c++ {
				o.grid[r][c] = newDefaultCell()
			}
		}
		for c := 0; c <= o.cursorX && c < o.cols; c++ {
			o.grid[o.cursorY][c] = newDefaultCell()
		}
	case 2: // Erase entire screen
		for r := 0; r < o.rows; r++ {
			for c := 0; c < o.cols; c++ {
				o.grid[r][c] = newDefaultCell()
			}
		}
		o.cursorX = 0
		o.cursorY = 0
	case 3: // Erase scrollback buffer (not implemented)
		// No-op
	}
}

// eraseInLine handles EL (Erase in Line) commands.
// Ensures cleared cells have default attributes.
func (o *Output) eraseInLine(mode int) {
	switch mode {
	case 0: // Erase from cursor to end of line
		for c := o.cursorX; c < o.cols; c++ {
			o.grid[o.cursorY][c] = newDefaultCell()
		}
	case 1: // Erase from beginning of line to cursor
		for c := 0; c <= o.cursorX && c < o.cols; c++ {
			o.grid[o.cursorY][c] = newDefaultCell()
		}
	case 2: // Erase entire line
		for c := 0; c < o.cols; c++ {
			o.grid[o.cursorY][c] = newDefaultCell()
		}
	}
}

// scrollUp shifts all lines in the grid up by one, saving the top line to scrollback,
// and clearing the new bottom line.
func (o *Output) scrollUp() {
	// If view is scrolled back, just decrease offset instead of scrolling content
	if o.viewOffset > 0 {
		o.viewOffset--
		return
	}

	// Save the top line to the scrollback buffer (circularly)
	// Need to copy the line data, not just the slice header
	topLineCopy := make([]Cell, o.cols)
	copy(topLineCopy, o.grid[0])
	o.scrollback[o.scrollbackHead] = topLineCopy
	o.scrollbackHead = (o.scrollbackHead + 1) % o.maxScrollback
	if o.scrollbackLines < o.maxScrollback {
		o.scrollbackLines++
	}

	// Shift grid lines up
	copy(o.grid[0:], o.grid[1:])
	// Clear the last line
	o.grid[o.rows-1] = make([]Cell, o.cols)
	for c := range o.grid[o.rows-1] {
		o.grid[o.rows-1][c] = newDefaultCell()
	}
}

// ScrollUp moves the view offset up (further into scrollback).
func (o *Output) ScrollUp(lines int) {
	newOffset := o.viewOffset + lines
	// Clamp offset to the number of lines actually in scrollback
	if newOffset > o.scrollbackLines {
		newOffset = o.scrollbackLines
	}
	o.viewOffset = newOffset
}

// ScrollDown moves the view offset down (towards the live view).
func (o *Output) ScrollDown(lines int) {
	newOffset := o.viewOffset - lines
	if newOffset < 0 {
		newOffset = 0 // Clamp at live view
	}
	o.viewOffset = newOffset
}

// GetVisibleGrid returns the slice of rows representing the current view,
// considering the viewOffset into the scrollback buffer.
func (o *Output) GetVisibleGrid() [][]Cell {
	visibleGrid := make([][]Cell, o.rows)

	if o.viewOffset == 0 {
		// Live view: just return the main grid (or a copy?)
		// Return direct for now, renderer only reads
		return o.grid
	} else {
		// Viewing scrollback
		// Calculate the index of the topmost visible line in the scrollback buffer.
		// This is the Nth most recent line, where N = viewOffset.
		scrollbackStartIdx := (o.scrollbackHead - o.viewOffset + o.maxScrollback) % o.maxScrollback

		scrollRowsToShow := min(o.rows, o.viewOffset)
		gridRowsToShow := o.rows - scrollRowsToShow

		// Copy lines from scrollback (handle wrap-around and nil entries)
		readIdx := scrollbackStartIdx
		for i := 0; i < scrollRowsToShow; i++ {
			if o.scrollback[readIdx] != nil {
				visibleGrid[i] = o.scrollback[readIdx]
			} else {
				// If scrollback line doesn't exist (nil), create an empty line
				emptyLine := make([]Cell, o.cols)
				for c := range emptyLine {
					emptyLine[c] = newDefaultCell()
				}
				visibleGrid[i] = emptyLine
			}
			readIdx = (readIdx + 1) % o.maxScrollback
		}

		// Copy lines from live grid
		for i := 0; i < gridRowsToShow; i++ {
			visibleGrid[scrollRowsToShow+i] = o.grid[i]
		}
	}

	return visibleGrid
}

// GetGrid returns the current state of the screen grid.
// Used by the renderer.
func (o *Output) GetGrid() [][]Cell {
	// Return a copy to prevent external modification? For now, return direct.
	return o.grid
}

// GetCursorPos returns the current cursor position.
func (o *Output) GetCursorPos() (int, int) {
	return o.cursorX, o.cursorY
}

// Resize changes the buffer dimensions, preserving content where possible.
func (o *Output) Resize(newRows, newCols int) {
	if newRows == o.rows && newCols == o.cols {
		return // No change
	}

	newGrid := make([][]Cell, newRows)
	for r := range newGrid {
		newGrid[r] = make([]Cell, newCols)
		for c := range newGrid[r] {
			newGrid[r][c] = newDefaultCell()
		}
	}

	// Copy existing content
	rowsToCopy := min(o.rows, newRows)
	colsToCopy := min(o.cols, newCols)
	for r := 0; r < rowsToCopy; r++ {
		copy(newGrid[r][:colsToCopy], o.grid[r][:colsToCopy])
	}

	o.grid = newGrid
	o.rows = newRows
	o.cols = newCols

	// Adjust cursor position if it's now out of bounds
	if o.cursorX >= o.cols {
		o.cursorX = o.cols - 1
	}
	if o.cursorY >= o.rows {
		o.cursorY = o.rows - 1
	}
	if o.cursorX < 0 {
		o.cursorX = 0
	}
	if o.cursorY < 0 {
		o.cursorY = 0
	}

}

// min returns the smaller of two integers.
// Added locally as it's not worth an import for just this in Go < 1.21
// If using Go 1.21+, use the built-in min.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the larger of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// IsLiveView returns true if the view offset is zero (viewing the live screen).
func (o *Output) IsLiveView() bool {
	return o.viewOffset == 0
}
