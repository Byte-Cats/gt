package internal

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Config holds application configuration
type Config struct {
	Style    StyleConfig    `toml:"style"`
	Behavior BehaviorConfig `toml:"behavior"`
}

// StyleConfig defines the visual appearance of the application
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

// BehaviorConfig defines application behavior settings
type BehaviorConfig struct {
	ShowHiddenByDefault   bool `toml:"show_hidden_by_default"`
	ConfirmFileOperations bool `toml:"confirm_file_operations"`
	PreviewMaxSizeKB      int  `toml:"preview_max_size_kb"`
	RememberLastDirectory bool `toml:"remember_last_directory"`
}

// PersistentState represents data that is saved between sessions
type PersistentState struct {
	LastDirectory string   `json:"last_directory"`
	Bookmarks     []string `json:"bookmarks"`
	RecentDirs    []string `json:"recent_dirs"`
}

// Model represents the application state
type Model struct {
	Config                   Config
	Styles                   Styles // Pre-rendered styles
	Keymap                   KeyMap // Keybindings
	Cwd                      string
	Entries                  []fs.DirEntry // All entries in the current directory
	FilteredEntries          []fs.DirEntry // Entries matching the filter
	Cursor                   int           // Index of the selected item in the *currently displayed* list
	Err                      error         // To display errors to the user
	Viewport                 viewport.Model
	PreviewViewport          viewport.Model
	Ready                    bool            // Indicates if viewport is ready
	FinalPath                string          // Stores the final selected path before exiting (for non-image files)
	ShowConfirm              bool            // Confirmation screen for file selection
	FilterInput              textinput.Model // Input field for filtering
	Filtering                bool            // Are we currently filtering?
	IsInImagePreviewMode     bool
	ImageFilesInDir          []fs.DirEntry       // Cache of image files in the current directory view
	CurrentPreviewImageIndex int                 // Index into imageFilesInDir
	ShowPreview              bool                // Whether to show the preview pane
	PreviewContent           string              // Content to display in preview pane
	ShowHidden               bool                // Whether to show hidden files
	Bookmarks                []string            // List of bookmarked paths
	ShowBookmarks            bool                // Whether we're in bookmarks view
	Clipboard                string              // Path for copy/cut operations
	ClipboardOp              string              // "copy" or "cut"
	ConfirmOperation         bool                // Whether we're confirming a file operation
	ConfirmPrompt            string              // Prompt text for confirmation
	ConfirmAction            func() error        // Action to execute on confirmation
	ShowHelp                 bool                // Whether to show help screen
	CurrentImageOptions      ImageDisplayOptions // Current image display settings
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

// NewModel initializes and returns a new model with default values
func NewModel() Model {
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

	// Load persistent state
	state, _ := LoadState() // Ignore error, will use defaults

	// Use last directory from state if configured to do so
	if cfg.Behavior.RememberLastDirectory && state.LastDirectory != "" {
		// Verify the directory exists before using it
		if _, err := os.Stat(state.LastDirectory); err == nil {
			cwd = state.LastDirectory
		}
	}

	m := Model{
		Config:              cfg,
		Styles:              styles,
		Keymap:              DefaultKeyMap,
		Cwd:                 cwd,
		Err:                 nil, // Initial state has no error
		FilterInput:         filterInput,
		Filtering:           false, // Start not filtering
		ShowHidden:          cfg.Behavior.ShowHiddenByDefault,
		Bookmarks:           state.Bookmarks,
		CurrentImageOptions: DefaultImageOptions(), // Initialize here
	}
	m.ReadDir(cwd) // Load initial directory contents

	return m
}

// loadConfig loads configuration from the config.toml file
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
		Behavior: BehaviorConfig{
			ShowHiddenByDefault:   false,
			ConfirmFileOperations: true,
			PreviewMaxSizeKB:      100,
			RememberLastDirectory: true,
		},
	}

	configPath := "config.toml"
	if _, err := os.Stat(configPath); err == nil {
		if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
			// Log error but continue with defaults
			fmt.Printf("Error decoding config file %s: %v. Using defaults.\n", configPath, err)
		}
	} else if !os.IsNotExist(err) {
		// Log other stat errors but continue with defaults
		fmt.Printf("Error checking config file %s: %v. Using defaults.\n", configPath, err)
	}

	return cfg
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

// ReadDir reads the contents of the directory specified by path
func (m *Model) ReadDir(path string) {
	m.Err = nil             // Clear previous error
	m.Filtering = false     // Reset filtering state when changing directory
	m.FilterInput.Reset()   // Clear filter text
	m.ShowBookmarks = false // Exit bookmark view if active

	// Validate path before attempting to read
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			m.Err = fmt.Errorf("Directory no longer exists: %s", path)
		} else if os.IsPermission(err) {
			m.Err = fmt.Errorf("Permission denied: %s", path)
		} else {
			m.Err = fmt.Errorf("Error accessing directory %s: %w", path, err)
		}
		return
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		m.Err = fmt.Errorf("Error reading directory %s: %w", path, err)
		m.Entries = []fs.DirEntry{}
		m.FilteredEntries = []fs.DirEntry{}
		m.Cursor = 0
		return
	}

	m.Cwd = path // Update Cwd only on successful read
	m.Entries = []fs.DirEntry{}
	m.Cursor = 0 // Reset cursor

	// Separate dirs and files, filter hidden files if needed
	dirs := []fs.DirEntry{}
	files := []fs.DirEntry{}
	for _, entry := range entries {
		// Skip hidden files if ShowHidden is false
		if !m.ShowHidden && strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		if entry.IsDir() {
			dirs = append(dirs, entry)
		} else {
			files = append(files, entry)
		}
	}

	m.Entries = append(dirs, files...)
	m.ApplyFilter() // Apply filter (which will be empty initially)
	m.Viewport.SetContent(m.RenderEntries())
	m.Viewport.GotoTop()

	// Add to recent directories if it's not already the most recent
	state, _ := LoadState()
	if len(state.RecentDirs) == 0 || state.RecentDirs[0] != path {
		// Add to beginning and limit to 10 entries
		state.RecentDirs = append([]string{path}, state.RecentDirs...)
		if len(state.RecentDirs) > 10 {
			state.RecentDirs = state.RecentDirs[:10]
		}

		// Update last directory
		state.LastDirectory = path

		// Save state
		_ = SaveState(&state)
	}
}

// ApplyFilter updates the FilteredEntries based on the FilterInput value
func (m *Model) ApplyFilter() {
	filterText := strings.ToLower(m.FilterInput.Value())

	// Pre-allocate capacity for filtered entries
	if filterText == "" {
		m.FilteredEntries = m.Entries // Just point to full list for empty filter
		m.Cursor = 0
		m.Viewport.SetContent(m.RenderEntries())
		m.Viewport.GotoTop()
		return
	}

	// If filter text isn't empty, create a new slice with estimated capacity
	m.FilteredEntries = make([]fs.DirEntry, 0, len(m.Entries)/2) // Estimate half will match

	for _, entry := range m.Entries {
		if strings.Contains(strings.ToLower(entry.Name()), filterText) {
			m.FilteredEntries = append(m.FilteredEntries, entry)
		}
	}

	m.Cursor = 0
	m.Viewport.SetContent(m.RenderEntries())
	m.Viewport.GotoTop()
}

// RenderEntries generates the string content for the viewport
func (m *Model) RenderEntries() string {
	if m.ShowBookmarks {
		return m.RenderBookmarks()
	}

	var builder strings.Builder
	// Use FilteredEntries if filtering, otherwise use all entries
	entriesToRender := m.Entries
	if m.Filtering || m.FilterInput.Value() != "" { // Show filtered list if filtering mode OR if text exists (even if not focused)
		entriesToRender = m.FilteredEntries
	}

	for i, entry := range entriesToRender {
		name := entry.Name()
		var line string
		isDir := entry.IsDir()

		// Pre-render the base style
		styledName := ""
		if isDir {
			styledName = m.Styles.Dir.Render(name + "/")
		} else {
			styledName = m.Styles.File.Render(name)
		}

		// Apply selection style if this item is the cursor
		if i == m.Cursor {
			prefix := m.Styles.SelectedPrefix
			if isDir {
				// Apply selected style to the pre-rendered dir name
				line = m.Styles.SelectedDir.Render(prefix + name + "/")
			} else {
				// Apply selected style to the pre-rendered file name
				line = m.Styles.SelectedFile.Render(prefix + name)
			}
		} else {
			// Render without prefix, add padding using the base styledName
			line = "  " + styledName // Add padding to the already styled name
		}

		builder.WriteString(line)
		builder.WriteRune('\n')
	}

	if len(entriesToRender) == 0 && (m.Filtering || m.FilterInput.Value() != "") {
		builder.WriteString("\n  (No matching entries)")
	} else if len(m.Entries) == 0 { // Check original entries for empty dir message
		builder.WriteString("\n  (Directory is empty)")
	}

	return builder.String()
}

// RenderBookmarks generates the bookmarks view content
func (m *Model) RenderBookmarks() string {
	var builder strings.Builder

	builder.WriteString("BOOKMARKS:\n\n")

	if len(m.Bookmarks) == 0 {
		builder.WriteString("  (No bookmarks saved)")
		return builder.String()
	}

	for i, bookmark := range m.Bookmarks {
		var line string

		// Apply selection style if this item is the cursor
		if i == m.Cursor {
			line = m.Styles.SelectedDir.Render(m.Styles.SelectedPrefix + bookmark)
		} else {
			// Render without prefix
			line = "  " + m.Styles.Dir.Render(bookmark)
		}

		builder.WriteString(line)
		builder.WriteRune('\n')
	}

	return builder.String()
}

// ImageDisplayOptions stores parameters for image rendering
type ImageDisplayOptions struct {
	Width               string // Can be in cells, pixels (px), or percentage (%)
	Height              string // Can be in cells, pixels (px), or percentage (%)
	MaxWidth            string // Maximum width in pixels
	MaxHeight           string // Maximum height in pixels
	PreserveAspectRatio bool   // Whether to maintain aspect ratio
	Align               string // horizontal alignment: "left", "center", "right"
	ZIndex              int    // For layering images (higher = in front)
	Persistent          bool   // Whether to keep image when cursor moves
	Name                string // Reference name for the image
	AdjustMode          string // Which dimension to adjust: "width", "height", or "both"
	BackgroundColor     string // Optional background color (hex: "#RRGGBB")
	Rotation            int    // Rotation in degrees (0, 90, 180, 270)
	Quality             int    // JPEG quality (1-100, only applies to JPEG images)
	AnimationDuration   int    // Duration between frames for animated GIFs (ms)
}

// DefaultImageOptions returns standard display options
func DefaultImageOptions() ImageDisplayOptions {
	return ImageDisplayOptions{
		Width:               "90%",
		Height:              "auto",
		MaxWidth:            "0", // Unlimited
		MaxHeight:           "0", // Unlimited
		PreserveAspectRatio: true,
		Align:               "center",
		ZIndex:              0,
		Persistent:          false,
		Name:                "preview",
		AdjustMode:          "width",
		BackgroundColor:     "",
		Rotation:            0,
		Quality:             85,
		AnimationDuration:   100,
	}
}

// DisplayCurrentImageInGT is a helper to load and send image sequence to gt
func (m *Model) DisplayCurrentImageInGT() error {
	if len(m.ImageFilesInDir) == 0 {
		return fmt.Errorf("No images available in current directory listing")
	}
	if m.CurrentPreviewImageIndex < 0 || m.CurrentPreviewImageIndex >= len(m.ImageFilesInDir) {
		// This case should ideally be handled by wrapping logic before calling,
		// but as a safeguard:
		m.CurrentPreviewImageIndex = 0   // Reset to first image if out of bounds
		if len(m.ImageFilesInDir) == 0 { // Double check after reset
			return fmt.Errorf("No images to display after index reset")
		}
	}

	imgPath := filepath.Join(m.Cwd, m.ImageFilesInDir[m.CurrentPreviewImageIndex].Name())

	// Get file info first to check size
	fileInfo, err := os.Stat(imgPath)
	if err != nil {
		return fmt.Errorf("Error checking image file %s: %w", filepath.Base(imgPath), err)
	}

	// For very large images, consider warning or alternative handling
	const maxSizeForDirectDisplay = 10 * 1024 * 1024 // 10MB
	if fileInfo.Size() > maxSizeForDirectDisplay {
		// Display a warning before loading
		fmt.Printf("\x1b[2J\x1b[H") // Clear screen + home
		fmt.Printf("Warning: Large image file (%d MB). Loading...\n", fileInfo.Size()/(1024*1024))
	}

	data, err := os.ReadFile(imgPath)
	if err != nil {
		return fmt.Errorf("Error reading image %s: %w", filepath.Base(imgPath), err)
	}

	b64data := base64.StdEncoding.EncodeToString(data)
	fmt.Print("\x1b[2J\x1b[H") // Clear screen + home

	// Construct options string based on m.CurrentImageOptions
	opts := []string{"inline=1"}

	// Add all options
	if m.CurrentImageOptions.Width != "" {
		opts = append(opts, fmt.Sprintf("width=%s", m.CurrentImageOptions.Width))
	}

	if m.CurrentImageOptions.Height != "" && m.CurrentImageOptions.Height != "auto" {
		opts = append(opts, fmt.Sprintf("height=%s", m.CurrentImageOptions.Height))
	}

	if m.CurrentImageOptions.MaxWidth != "" && m.CurrentImageOptions.MaxWidth != "0" {
		opts = append(opts, fmt.Sprintf("max-width=%s", m.CurrentImageOptions.MaxWidth))
	}

	if m.CurrentImageOptions.MaxHeight != "" && m.CurrentImageOptions.MaxHeight != "0" {
		opts = append(opts, fmt.Sprintf("max-height=%s", m.CurrentImageOptions.MaxHeight))
	}

	if !m.CurrentImageOptions.PreserveAspectRatio {
		opts = append(opts, "preserveaspectratio=0")
	}

	if m.CurrentImageOptions.Align != "" {
		opts = append(opts, fmt.Sprintf("align=%s", m.CurrentImageOptions.Align))
	}

	if m.CurrentImageOptions.ZIndex > 0 {
		opts = append(opts, fmt.Sprintf("z-index=%d", m.CurrentImageOptions.ZIndex))
	}

	if m.CurrentImageOptions.Persistent {
		opts = append(opts, "persistent=1")
	}

	if m.CurrentImageOptions.Name != "" {
		opts = append(opts, fmt.Sprintf("name=%s", m.CurrentImageOptions.Name))
	}

	// Add extended options
	if m.CurrentImageOptions.BackgroundColor != "" {
		opts = append(opts, fmt.Sprintf("background-color=%s", m.CurrentImageOptions.BackgroundColor))
	}

	if m.CurrentImageOptions.Rotation > 0 {
		opts = append(opts, fmt.Sprintf("rotation=%d", m.CurrentImageOptions.Rotation))
	}

	if m.CurrentImageOptions.Quality > 0 && m.CurrentImageOptions.Quality <= 100 {
		opts = append(opts, fmt.Sprintf("quality=%d", m.CurrentImageOptions.Quality))
	}

	if m.CurrentImageOptions.AnimationDuration > 0 {
		opts = append(opts, fmt.Sprintf("animation-duration=%d", m.CurrentImageOptions.AnimationDuration))
	}

	// Construct and send the escape sequence
	fmt.Printf("\x1b]1337;File=%s:%s\a", strings.Join(opts, ";"), b64data)
	_ = os.Stdout.Sync()
	return nil
}

// LoadFilePreview loads a preview of a text file
func (m *Model) LoadFilePreview(filePath string) {
	m.ShowPreview = false
	m.PreviewContent = ""

	// Check if file exists
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		m.Err = fmt.Errorf("Error accessing file: %v", err)
		return
	}

	// Check file size
	maxSize := int64(m.Config.Behavior.PreviewMaxSizeKB * 1024)
	if fileInfo.Size() > maxSize {
		m.PreviewContent = fmt.Sprintf("File too large to preview (%d KB, max %d KB)",
			fileInfo.Size()/1024, m.Config.Behavior.PreviewMaxSizeKB)
		m.ShowPreview = true
		m.PreviewViewport.SetContent(m.PreviewContent)
		return
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		m.Err = fmt.Errorf("Error reading file: %v", err)
		return
	}

	// For binary files, show a message
	if !IsTextFile(content) {
		m.PreviewContent = "Binary file - preview not available"
		m.ShowPreview = true
		m.PreviewViewport.SetContent(m.PreviewContent)
		return
	}

	// Set preview content
	m.PreviewContent = string(content)
	m.ShowPreview = true
	m.PreviewViewport.SetContent(m.PreviewContent)
}

// IsTextFile attempts to determine if a file is a text file by checking content
func IsTextFile(content []byte) bool {
	// Check for null bytes (common in binary files)
	for _, b := range content {
		if b == 0 {
			return false
		}
	}

	// Check for non-printable characters (simple heuristic)
	nonPrintable := 0
	for _, b := range content {
		if (b < 32 || b > 126) && b != '\n' && b != '\r' && b != '\t' {
			nonPrintable++
		}
	}

	// If more than 5% non-printable, probably binary
	return float64(nonPrintable)/float64(len(content)) < 0.05
}

// IsImageFile checks if the filename has a common image extension
func IsImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".bmp", ".svg", ".tiff", ".tif":
		return true
	default:
		return false
	}
}

// IsTextFileExt checks if the filename has a common text file extension
func IsTextFileExt(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".txt", ".md", ".go", ".py", ".js", ".html", ".css", ".json", ".toml", ".yaml", ".yml",
		".c", ".cpp", ".h", ".hpp", ".java", ".sh", ".bat", ".ps1", ".rb", ".php", ".xml", ".csv":
		return true
	default:
		return false
	}
}

// Init initializes the model for bubbletea
func (m *Model) Init() tea.Cmd {
	// Reset image options to defaults
	m.CurrentImageOptions = DefaultImageOptions()
	return textinput.Blink // Start the text input blinking
}

// SaveState saves the current state to a file
func SaveState(state *PersistentState) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".config", "gf")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	stateFile := filepath.Join(configDir, "state.json")
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(stateFile, data, 0644)
}

// LoadState loads the state from a file
func LoadState() (PersistentState, error) {
	var state PersistentState

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return state, err
	}

	stateFile := filepath.Join(homeDir, ".config", "gf", "state.json")
	data, err := os.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return state, nil // Return empty state, not an error
		}
		return state, err
	}

	err = json.Unmarshal(data, &state)
	return state, err
}

// Helper functions
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
