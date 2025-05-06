package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- Configuration ---

type Config struct {
	Style StyleConfig `toml:"style"`
}

type StyleConfig struct {
	DirColor       string `toml:"dir_color"`
	FileColor      string `toml:"file_color"`
	SelectedPrefix string `toml:"selected_prefix"`
	SelectedColor  string `toml:"selected_color"`
	ErrorColor     string `toml:"error_color"`
	HeaderColor    string `toml:"header_color"`
	FooterColor    string `toml:"footer_color"`
	BorderColor    string `toml:"border_color"`
	BorderType     string `toml:"border_type"`
}

func loadConfig() Config {
	// Default config
	cfg := Config{
		Style: StyleConfig{
			DirColor:       "blue",
			FileColor:      "white",
			SelectedPrefix: "> ",
			SelectedColor:  "yellow",
			ErrorColor:     "red",
			HeaderColor:    "cyan",
			FooterColor:    "gray",
			BorderColor:    "magenta",
			BorderType:     "rounded", // Default border type
		},
	}

	configPath := "config.toml"
	if _, err := os.Stat(configPath); err == nil {
		if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
			// Log error but continue with defaults
			log.Printf("Error decoding config file %s: %v. Using defaults.", configPath, err)
		}
	} else if !os.IsNotExist(err) {
		// Log other stat errors but continue with defaults
		log.Printf("Error checking config file %s: %v. Using defaults.", configPath, err)
	}

	return cfg
}

// --- Bubble Tea Model ---

type model struct {
	config          Config
	styles          Styles // Pre-rendered styles
	keymap          KeyMap // Keybindings
	cwd             string
	entries         []fs.DirEntry // All entries in the current directory
	filteredEntries []fs.DirEntry // Entries matching the filter
	cursor          int           // Index of the selected item in the *currently displayed* list (entries or filteredEntries)
	err             error         // To display errors to the user
	viewport        viewport.Model
	ready           bool            // Indicates if viewport is ready
	finalPath       string          // Stores the final selected path before exiting
	showConfirm     bool            // Confirmation screen for file selection
	filterInput     textinput.Model // Input field for filtering
	filtering       bool            // Are we currently filtering?
}

// Styles struct to hold pre-rendered lipgloss styles
type Styles struct {
	Base           lipgloss.Style
	Header         lipgloss.Style
	Footer         lipgloss.Style
	Dir            lipgloss.Style
	File           lipgloss.Style
	SelectedDir    lipgloss.Style
	SelectedFile   lipgloss.Style
	Error          lipgloss.Style
	SelectedPrefix string
}

// KeyMap defines custom keybindings
type KeyMap struct {
	Up          key.Binding
	Down        key.Binding
	PageUp      key.Binding
	PageDown    key.Binding
	GoToTop     key.Binding
	GoToBottom  key.Binding
	SelectEnter key.Binding
	SelectSpace key.Binding
	Back        key.Binding // Go to parent directory
	Quit        key.Binding
	ConfirmYes  key.Binding
	ConfirmNo   key.Binding
	StartFilter key.Binding // New: Start filtering
	ClearFilter key.Binding // New: Clear/cancel filter (ESC)
}

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
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	SelectSpace: key.NewBinding( // Alias for select
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
		key.WithKeys("n", "esc"), // Note: Esc is now overloaded, handled in Update
		key.WithHelp("n/esc", "cancel"),
	),
	StartFilter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
	ClearFilter: key.NewBinding( // Esc clears filter or cancels confirmation
		key.WithKeys("esc"),
		key.WithHelp("esc", "clear filter/cancel"),
	),
}

func newModel() model {
	cfg := loadConfig()
	styles := createStyles(cfg.Style)

	cwd, err := os.Getwd()
	if err != nil {
		// This error is critical, handle it before starting Bubble Tea
		fmt.Printf("Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	// Initialize filter input
	filterInput := textinput.New()
	filterInput.Placeholder = "Filter..."
	filterInput.CharLimit = 156
	filterInput.Width = 20 // Initial width, will resize

	m := model{
		config:      cfg,
		styles:      styles,
		keymap:      DefaultKeyMap,
		cwd:         cwd,
		err:         nil, // Initial state has no error
		filterInput: filterInput,
		filtering:   false, // Start not filtering
	}
	m.readDir(cwd) // Load initial directory contents

	return m
}

// createStyles initializes lipgloss styles based on the loaded config
func createStyles(cfg StyleConfig) Styles {
	borderStyle := lipgloss.HiddenBorder() // Default to no border
	if cfg.BorderType != "" && cfg.BorderType != "hidden" {
		switch cfg.BorderType {
		case "single":
			borderStyle = lipgloss.NormalBorder()
		case "double":
			borderStyle = lipgloss.DoubleBorder()
		case "rounded":
			borderStyle = lipgloss.RoundedBorder()
		case "thick":
			borderStyle = lipgloss.ThickBorder()
		}
	}

	s := Styles{
		Base: lipgloss.NewStyle().
			BorderStyle(borderStyle).
			BorderForeground(lipgloss.Color(cfg.BorderColor)),
		Header: lipgloss.NewStyle().
			Foreground(lipgloss.Color(cfg.HeaderColor)).
			Bold(true).
			Padding(0, 1),
		Footer: lipgloss.NewStyle().
			Foreground(lipgloss.Color(cfg.FooterColor)).
			Padding(0, 1),
		Dir: lipgloss.NewStyle().
			Foreground(lipgloss.Color(cfg.DirColor)),
		File: lipgloss.NewStyle().
			Foreground(lipgloss.Color(cfg.FileColor)),
		SelectedDir: lipgloss.NewStyle().
			Foreground(lipgloss.Color(cfg.SelectedColor)).
			Bold(true),
		SelectedFile: lipgloss.NewStyle().
			Foreground(lipgloss.Color(cfg.SelectedColor)).
			Bold(true),
		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color(cfg.ErrorColor)),
		SelectedPrefix: cfg.SelectedPrefix,
	}
	// Apply selected prefix style
	s.SelectedDir = s.SelectedDir.SetString(cfg.SelectedPrefix + "%s")
	s.SelectedFile = s.SelectedFile.SetString(cfg.SelectedPrefix + "%s")

	return s
}

// readDir reads the contents of the directory specified by path
func (m *model) readDir(path string) {
	m.err = nil           // Clear previous error
	m.filtering = false   // Reset filtering state when changing directory
	m.filterInput.Reset() // Clear filter text

	entries, err := os.ReadDir(path)
	if err != nil {
		m.err = fmt.Errorf("Error reading directory %s: %w", path, err)
		m.entries = []fs.DirEntry{}
		m.filteredEntries = []fs.DirEntry{}
		m.cursor = 0
		return
	}

	m.cwd = path // Update cwd only on successful read
	m.entries = []fs.DirEntry{}
	m.cursor = 0 // Reset cursor

	// Separate dirs and files, then combine (dirs first)
	dirs := []fs.DirEntry{}
	files := []fs.DirEntry{}
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry)
		} else {
			files = append(files, entry)
		}
	}
	m.entries = append(dirs, files...)
	m.applyFilter() // Apply filter (which will be empty initially)
	m.viewport.SetContent(m.renderEntries())
	m.viewport.GotoTop()
}

// applyFilter updates the filteredEntries based on the filterInput value
func (m *model) applyFilter() {
	filterText := strings.ToLower(m.filterInput.Value())
	m.filteredEntries = []fs.DirEntry{} // Reset filtered list

	if filterText == "" {
		m.filteredEntries = m.entries // No filter, show all
	} else {
		for _, entry := range m.entries {
			if strings.Contains(strings.ToLower(entry.Name()), filterText) {
				m.filteredEntries = append(m.filteredEntries, entry)
			}
		}
	}

	m.cursor = 0 // Reset cursor to the top of the new list
	m.viewport.SetContent(m.renderEntries())
	m.viewport.GotoTop()
}

// renderEntries generates the string content for the viewport
func (m *model) renderEntries() string {
	var builder strings.Builder
	// Use filteredEntries if filtering, otherwise use all entries
	entriesToRender := m.entries
	if m.filtering || m.filterInput.Value() != "" { // Show filtered list if filtering mode OR if text exists (even if not focused)
		entriesToRender = m.filteredEntries
	}

	for i, entry := range entriesToRender {
		name := entry.Name()
		var line string
		isDir := entry.IsDir()

		// Pre-render the base style
		styledName := ""
		if isDir {
			styledName = m.styles.Dir.Render(name + "/")
		} else {
			styledName = m.styles.File.Render(name)
		}

		// Apply selection style if this item is the cursor
		if i == m.cursor {
			prefix := m.styles.SelectedPrefix
			if isDir {
				// Apply selected style to the pre-rendered dir name
				line = m.styles.SelectedDir.Render(prefix + name + "/")
			} else {
				// Apply selected style to the pre-rendered file name
				line = m.styles.SelectedFile.Render(prefix + name)
			}
		} else {
			// Render without prefix, add padding using the base styledName
			line = "  " + styledName // Add padding to the already styled name
		}

		builder.WriteString(line)
		builder.WriteRune('\n')
	}

	if len(entriesToRender) == 0 && (m.filtering || m.filterInput.Value() != "") {
		builder.WriteString("\n  (No matching entries)")
	} else if len(m.entries) == 0 { // Check original entries for empty dir message
		builder.WriteString("\n  (Directory is empty)")
	}

	return builder.String()
}

func (m model) Init() tea.Cmd {
	return textinput.Blink // Start the text input blinking
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	currentEntries := m.entries
	if m.filtering || m.filterInput.Value() != "" {
		currentEntries = m.filteredEntries
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		filterHeight := 0
		if m.filtering {
			filterHeight = lipgloss.Height(m.filterInput.View()) + 1 // +1 for potential newline/padding
		}
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + filterHeight + footerHeight

		if !m.ready {
			// Initialize viewport
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight + filterHeight // Adjust Y position based on filter visibility
			m.viewport.HighPerformanceRendering = false
			m.filterInput.Width = msg.Width - 4 // Adjust filter input width slightly less than full
			m.applyFilter()                     // Apply initial (empty) filter now that viewport is sized
			m.ready = true

			// Apply border style
			if m.styles.Base.GetBorderStyle() != lipgloss.HiddenBorder() {
				m.viewport.Style = m.styles.Base.Copy().
					Width(msg.Width - m.styles.Base.GetHorizontalBorderSize()).
					Height(msg.Height - verticalMarginHeight - m.styles.Base.GetVerticalBorderSize())
			}
		} else {
			// Update viewport and filter input size on resize
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
			m.viewport.YPosition = headerHeight + filterHeight // Update Y position
			m.filterInput.Width = msg.Width - 4

			// Update border size on resize too
			if m.styles.Base.GetBorderStyle() != lipgloss.HiddenBorder() {
				m.viewport.Style = m.styles.Base.Copy().
					Width(msg.Width - m.styles.Base.GetHorizontalBorderSize()).
					Height(msg.Height - verticalMarginHeight - m.styles.Base.GetVerticalBorderSize())
			}
		}

	case tea.KeyMsg:
		// Always handle quit first
		if key.Matches(msg, m.keymap.Quit) {
			return m, tea.Quit
		}

		// Handle confirmation screen keys
		if m.showConfirm {
			switch {
			case key.Matches(msg, m.keymap.ConfirmYes):
				m.finalPath, m.err = filepath.Abs(m.finalPath)
				if m.err != nil {
					m.showConfirm = false
				} else {
					return m, tea.Quit
				}
			case key.Matches(msg, m.keymap.ConfirmNo), key.Matches(msg, m.keymap.ClearFilter): // Esc cancels confirm
				m.showConfirm = false
				m.finalPath = ""
			}
			m.viewport.SetContent(m.renderEntries()) // Re-render after confirm action
			return m, nil
		}

		// Handle filter input keys if filtering
		if m.filtering {
			switch {
			// Clear filter and exit filtering mode
			case key.Matches(msg, m.keymap.ClearFilter):
				m.filtering = false
				m.filterInput.Blur()
				m.filterInput.Reset()
				m.applyFilter()                      // Re-apply to show all entries
				cmds = append(cmds, textinput.Blink) // Stop blinking

			// Default: Update filter input
			default:
				var filterCmd tea.Cmd
				m.filterInput, filterCmd = m.filterInput.Update(msg)
				m.applyFilter() // Re-apply filter on every keystroke
				cmds = append(cmds, filterCmd)

			}
			// Ensure viewport updates after filter actions
			m.viewport.SetContent(m.renderEntries())
			return m, tea.Batch(cmds...) // Return early after handling filter keys
		}

		// Handle regular navigation and actions (when not filtering)
		switch {
		// Start filtering
		case key.Matches(msg, m.keymap.StartFilter):
			m.filtering = true
			m.filterInput.Focus()
			cmds = append(cmds, textinput.Blink)

		// Navigation keys (Up, Down, etc.) - Operate on currentEntries
		case key.Matches(msg, m.keymap.Up):
			if m.cursor > 0 {
				m.cursor--
				// Scroll viewport if needed (simplified check)
				if m.viewport.YOffset > m.cursor {
					m.viewport.SetYOffset(m.cursor)
				}
			}
		case key.Matches(msg, m.keymap.Down):
			if m.cursor < len(currentEntries)-1 {
				m.cursor++
				// Scroll viewport if needed (simplified check)
				if m.cursor >= m.viewport.YOffset+m.viewport.Height {
					m.viewport.LineDown(1)
				}
			}
		case key.Matches(msg, m.keymap.PageUp):
			m.viewport.ViewUp()
			m.cursor = max(0, m.cursor-m.viewport.Height) // Adjust cursor based on viewport jump
		case key.Matches(msg, m.keymap.PageDown):
			m.viewport.ViewDown()
			m.cursor = min(len(currentEntries)-1, m.cursor+m.viewport.Height) // Adjust cursor
		case key.Matches(msg, m.keymap.GoToTop):
			m.viewport.GotoTop()
			m.cursor = 0
		case key.Matches(msg, m.keymap.GoToBottom):
			m.viewport.GotoBottom()
			m.cursor = len(currentEntries) - 1

		// Select item
		case key.Matches(msg, m.keymap.SelectEnter), key.Matches(msg, m.keymap.SelectSpace):
			if len(currentEntries) == 0 || m.cursor >= len(currentEntries) { // Check bounds
				break // Nothing to select
			}
			selectedEntry := currentEntries[m.cursor]
			newPath := filepath.Join(m.cwd, selectedEntry.Name())

			if selectedEntry.IsDir() {
				m.readDir(newPath) // Navigate into directory (will reset filter)
			} else {
				// File selected - prepare for confirmation
				m.finalPath = newPath
				m.showConfirm = true
			}

		// Go back
		case key.Matches(msg, m.keymap.Back):
			parentDir := filepath.Dir(m.cwd)
			if parentDir != m.cwd {
				m.readDir(parentDir) // Go up (will reset filter)
			} else {
				m.err = fmt.Errorf("Already at root directory")
			}

		// If Esc is pressed when not filtering and not confirming, do nothing silently
		case key.Matches(msg, m.keymap.ClearFilter):
			break

		}
		m.viewport.SetContent(m.renderEntries()) // Update view after action
	}

	// Handle keyboard and mouse events in the viewport (allow scrolling even when filtering)
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) headerView() string {
	title := m.styles.Header.Render("Current: " + m.cwd)
	// Don't repeat the line if border is enabled, as viewport will have top border
	if m.styles.Base.GetBorderStyle() == lipgloss.HiddenBorder() {
		line := strings.Repeat("─", m.viewport.Width)
		return lipgloss.JoinVertical(lipgloss.Left, title, line)
	}
	return title // Just the title if border is present
}

func (m model) filterView() string {
	if !m.filtering {
		return "" // Don't show if not actively filtering
	}
	return m.filterInput.View() // Return the text input view
}

func (m model) footerView() string {
	var help strings.Builder

	// Basic navigation help
	help.WriteString(m.keymap.Up.Help().Key + "/" + m.keymap.Down.Help().Key + " nav ")
	help.WriteString(m.keymap.SelectEnter.Help().Key + " sel ")
	help.WriteString(m.keymap.Back.Help().Key + " back ")

	if m.filtering {
		help.WriteString(m.keymap.ClearFilter.Help().Key + " clear ")
	} else {
		help.WriteString(m.keymap.StartFilter.Help().Key + " filter ")
	}
	help.WriteString(m.keymap.Quit.Help().Key + " quit")

	footerText := help.String()

	if m.showConfirm {
		footerText = fmt.Sprintf("Confirm selection: %s? (%s/%s)",
			filepath.Base(m.finalPath),
			m.keymap.ConfirmYes.Help().Key,
			m.keymap.ConfirmNo.Help().Key, // Show 'n' from ConfirmNo binding
		)
	} else if m.err != nil {
		footerText = m.styles.Error.Render(m.err.Error())
	}

	// Don't repeat the line if border is enabled
	if m.styles.Base.GetBorderStyle() == lipgloss.HiddenBorder() {
		line := strings.Repeat("─", m.viewport.Width)
		return lipgloss.JoinVertical(lipgloss.Left, line, m.styles.Footer.Render(footerText))
	}
	return m.styles.Footer.Render(footerText) // Just the text if border is present
}

func (m model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	filterStr := m.filterView()
	viewContent := m.viewport.View()

	// Add padding below filter if it's shown and border is hidden
	if filterStr != "" && m.styles.Base.GetBorderStyle() == lipgloss.HiddenBorder() {
		filterStr += "\n" // Add spacing below filter input
	}

	// Combine header, filter (if active), viewport, and footer
	formatString := "%s\n%s%s\n%s"
	return fmt.Sprintf(formatString,
		m.headerView(),
		filterStr, // Filter view (includes newline if needed)
		viewContent,
		m.footerView())
}

// --- Main Function ---

func main() {
	// Enable mouse events
	// if _, ok := os.LookupEnv("DISABLE_MOUSE"); !ok {
	// 	fmt.Println("Enabling mouse...") // Debug
	// 	tea.EnterAltScreen() // Not strictly necessary for mouse but often used together
	// }

	m := newModel()
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion()) // Enable Alt Screen and Mouse

	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

	// After TUI exits, print the final path if one was selected
	if m.finalPath != "" {
		fmt.Println(m.finalPath) // Print final absolute path to stdout
	} else if m.err != nil && !m.showConfirm { // Print error if we exited due to one (and not confirming)
		fmt.Fprintf(os.Stderr, "Error: %v\n", m.err)
		os.Exit(1) // Exit with error code if there was an error state
	}
}

// Helper functions (consider moving to a utils package later)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
