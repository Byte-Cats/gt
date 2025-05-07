package main

import (
	"encoding/base64"
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
	config                   Config
	styles                   Styles // Pre-rendered styles
	keymap                   KeyMap // Keybindings
	cwd                      string
	entries                  []fs.DirEntry // All entries in the current directory
	filteredEntries          []fs.DirEntry // Entries matching the filter
	cursor                   int           // Index of the selected item in the *currently displayed* list (entries or filteredEntries)
	err                      error         // To display errors to the user
	viewport                 viewport.Model
	ready                    bool            // Indicates if viewport is ready
	finalPath                string          // Stores the final selected path before exiting (for non-image files)
	showConfirm              bool            // Confirmation screen for file selection
	filterInput              textinput.Model // Input field for filtering
	filtering                bool            // Are we currently filtering?
	isInImagePreviewMode     bool
	imageFilesInDir          []fs.DirEntry // Cache of image files in the current directory view
	currentPreviewImageIndex int           // Index into imageFilesInDir
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

	if m.isInImagePreviewMode {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			keyStr := msg.String()
			switch {
			case keyStr == "k" || keyStr == "up":
				if len(m.imageFilesInDir) > 0 {
					m.currentPreviewImageIndex--
					if m.currentPreviewImageIndex < 0 {
						m.currentPreviewImageIndex = len(m.imageFilesInDir) - 1
					}
					if err := m.displayCurrentImageInGT(); err != nil {
						m.err = err
					}
				}
				return m, nil
			case keyStr == "j" || keyStr == "down":
				if len(m.imageFilesInDir) > 0 {
					m.currentPreviewImageIndex++
					if m.currentPreviewImageIndex >= len(m.imageFilesInDir) {
						m.currentPreviewImageIndex = 0
					}
					if err := m.displayCurrentImageInGT(); err != nil {
						m.err = err
					}
				}
				return m, nil
			case key.Matches(msg, m.keymap.Quit) || key.Matches(msg, m.keymap.ClearFilter) || key.Matches(msg, m.keymap.Back) || keyStr == "q" || keyStr == "escape":
				m.isInImagePreviewMode = false
				fmt.Print("\x1b[2J\x1b[H")
				_ = os.Stdout.Sync()
				m.err = nil

				// If there was filter text, ensure the filter input is active again.
				if m.filterInput.Value() != "" {
					m.filtering = true // Make the filter input visible
					m.filterInput.Focus()
					cmds = append(cmds, textinput.Blink) // Add blink command
				} else {
					m.filtering = false // No filter text, ensure filter mode is off
				}
				// m.applyFilter() // Re-applying filter here might be redundant if content isn't changing, but renderEntries will use current filter.
				m.viewport.SetContent(m.renderEntries())
				return m, tea.Batch(cmds...)
			}
		}
		return m, nil
	}

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
				absPath, err := filepath.Abs(m.finalPath)
				if err != nil {
					m.err = fmt.Errorf("Error getting absolute path: %w", err)
					m.showConfirm = false
					m.finalPath = ""
				} else {
					// This is for non-image files only now
					fmt.Println(absPath)
					_ = os.Stdout.Sync()
					return m, tea.Quit // Quit after printing path
				}
			case key.Matches(msg, m.keymap.ConfirmNo), key.Matches(msg, m.keymap.ClearFilter):
				m.showConfirm = false
				m.finalPath = ""
			}
			m.viewport.SetContent(m.renderEntries())
			return m, nil
		}

		// Handle filter input keys if filtering
		if m.filtering {
			switch {
			// Clear filter and exit filtering mode (ESC key)
			case key.Matches(msg, m.keymap.ClearFilter):
				m.filtering = false
				m.filterInput.Blur()
				m.filterInput.Reset() // Clears text
				m.applyFilter()       // Shows all entries
				// cmds = append(cmds, textinput.Blink) // Blink is usually tied to focus; blurring should handle it.
				// Or, if we want to ensure a blink command isn't processed, ensure cmds is managed correctly.
				// For now, let's rely on Blur(). If a specific "stop blink" cmd is needed, it's usually implicit.

			// Commit filter and return to list navigation (ENTER key)
			case msg.Type == tea.KeyEnter:
				m.filtering = false  // Stop active typing in filter
				m.filterInput.Blur() // Remove cursor from input; this should also stop blinking.
				// DO NOT RESET m.filterInput.Value() - this keeps the filter active
				m.applyFilter() // Ensure filteredEntries is up-to-date with current filter text
				// No specific cmd needed to stop blink, Blur() should handle it.

			// Default: Update filter input (for other character keys)
			default:
				var filterCmd tea.Cmd
				m.filterInput, filterCmd = m.filterInput.Update(msg)
				m.applyFilter()                // Re-apply filter on every keystroke
				cmds = append(cmds, filterCmd) // This will include textinput.Blink if input is focused and active
			}
			// Ensure viewport updates after filter actions
			m.viewport.SetContent(m.renderEntries())
			return m, tea.Batch(cmds...)
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
			absPath := filepath.Join(m.cwd, selectedEntry.Name()) // Calculate absPath here

			if selectedEntry.IsDir() {
				m.readDir(absPath) // Navigate into directory (will reset filter)
			} else {
				m.finalPath = absPath
				if isImageFile(absPath) {
					m.isInImagePreviewMode = true
					m.showConfirm = false
					m.imageFilesInDir = []fs.DirEntry{}

					sourceEntries := m.entries
					if m.filtering || m.filterInput.Value() != "" {
						sourceEntries = m.filteredEntries
					}

					selectedIndexInImages := -1
					for _, entry := range sourceEntries {
						if entry.IsDir() {
							continue
						}
						entryAbsPath := filepath.Join(m.cwd, entry.Name())
						if isImageFile(entryAbsPath) {
							m.imageFilesInDir = append(m.imageFilesInDir, entry)
							if entryAbsPath == absPath {
								selectedIndexInImages = len(m.imageFilesInDir) - 1
							}
						}
					}

					if len(m.imageFilesInDir) > 0 {
						if selectedIndexInImages != -1 {
							m.currentPreviewImageIndex = selectedIndexInImages
						} else {
							m.currentPreviewImageIndex = 0
						}
						errDisplay := m.displayCurrentImageInGT()
						if errDisplay != nil {
							m.err = errDisplay
							m.isInImagePreviewMode = false
						}
					} else {
						m.err = fmt.Errorf("No images found in current view.")
						m.isInImagePreviewMode = false
					}
					return m, nil
				} else {
					m.showConfirm = true
				}
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

	return m, tea.Batch(cmds...) // This should be the correct end of the Update function
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
	if m.keymap.SelectSpace.Enabled() { // Check if space is enabled as a key
		help.WriteString(m.keymap.SelectSpace.Help().Key + " sel ")
	}
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

// displayCurrentImageInGT is a helper to load and send image sequence to gt
func (m *model) displayCurrentImageInGT() error {
	if len(m.imageFilesInDir) == 0 {
		return fmt.Errorf("No images available in current directory listing")
	}
	if m.currentPreviewImageIndex < 0 || m.currentPreviewImageIndex >= len(m.imageFilesInDir) {
		// This case should ideally be handled by wrapping logic before calling,
		// but as a safeguard:
		m.currentPreviewImageIndex = 0   // Reset to first image if out of bounds
		if len(m.imageFilesInDir) == 0 { // Double check after reset
			return fmt.Errorf("No images to display after index reset")
		}
	}

	imgPath := filepath.Join(m.cwd, m.imageFilesInDir[m.currentPreviewImageIndex].Name())
	data, err := os.ReadFile(imgPath) // Use os.ReadFile
	if err != nil {
		return fmt.Errorf("Error reading image %s: %w", filepath.Base(imgPath), err)
	}
	b64data := base64.StdEncoding.EncodeToString(data)

	fmt.Print("\x1b[2J\x1b[H") // Clear screen + home
	fmt.Printf("\x1b]1337;File=inline=1:%s\a", b64data)
	_ = os.Stdout.Sync()
	return nil
}

// isImageFile checks if the filename has a common image extension.
func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".bmp": // Added .bmp as another common one
		return true
	default:
		return false
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

// View is a part of tea.Model - THIS IS THE CORRECT ONE
func (m model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	if m.isInImagePreviewMode {
		var status string
		if m.err != nil { // Prioritize error display
			status = m.styles.Error.Render(m.err.Error())
		} else if len(m.imageFilesInDir) > 0 &&
			m.currentPreviewImageIndex >= 0 &&
			m.currentPreviewImageIndex < len(m.imageFilesInDir) {
			currentImageName := filepath.Base(m.imageFilesInDir[m.currentPreviewImageIndex].Name())
			status = fmt.Sprintf("Preview: %s (%d/%d) | Cycle: j/k,↑/↓ | Exit: Esc/q/Bksp | Scroll in GT: (mouse wheel/keys if needed)",
				currentImageName,
				m.currentPreviewImageIndex+1,
				len(m.imageFilesInDir))
		} else {
			status = "Image preview active. No image loaded or index error."
		}
		return "\n\n" + m.styles.Footer.Render(status)
	}

	// Normal view rendering
	filterStr := m.filterView()
	viewContent := m.viewport.View()

	// Add padding below filter if it's shown and border is hidden
	if filterStr != "" && m.styles.Base.GetBorderStyle() == lipgloss.HiddenBorder() {
		filterStr += "\n"
	}

	formatString := "%s\n%s%s\n%s"
	return fmt.Sprintf(formatString,
		m.headerView(),
		filterStr,
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

	m := newModel() // m is of type model

	// If Init, Update, View are defined on (m model), then pass m directly.
	// If they are on (m *model), then pass &m.
	// The linter errors suggest they are on (m model) due to "missing method View" on *model.
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if err := p.Start(); err != nil { // Start() will block until tea.Quit is received
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

	// Outputting is now handled directly in the Update function before tea.Quit.
	/*
		// After TUI exits, print the final path if one was selected
		// This logic was commented out in previous steps as Update handles printing.
		// If finalModel was intended to be used, it should be the result of p.Run()
		// or another mechanism to get the final model state.
		// For now, keeping it commented as per earlier decisions.
	*/
}
