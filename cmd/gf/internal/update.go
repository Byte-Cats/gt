package internal

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ShowHelpMsg toggles help screen
type ShowHelpMsg bool

// RenamePromptMsg initiates rename operation
type RenamePromptMsg struct {
	Type string // "file", "dir", or "new"
}

// NewFilePromptMsg initiates new file creation
type NewFilePromptMsg struct{}

// NewDirPromptMsg initiates new directory creation
type NewDirPromptMsg struct{}

// Update handles and responds to messages
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		cmds []tea.Cmd
	)

	// Handle image preview mode separately
	if m.IsInImagePreviewMode {
		return handleImagePreviewMode(m, msg)
	}

	// Handle help screen mode
	if m.ShowHelp {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if key.Matches(msg, m.Keymap.Help) || key.Matches(msg, m.Keymap.Quit) || key.Matches(msg, m.Keymap.Back) {
				m.ShowHelp = false
				m.Viewport.SetContent(m.RenderEntries())
			}
		}
		return m, nil
	}

	// Handle confirmation screen
	if m.ShowConfirm {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, m.Keymap.ConfirmYes):
				absPath, err := filepath.Abs(m.FinalPath)
				if err != nil {
					m.Err = fmt.Errorf("Error getting absolute path: %w", err)
					m.ShowConfirm = false
					m.FinalPath = ""
				} else {
					// This is for non-image files only
					fmt.Println(absPath)
					_ = os.Stdout.Sync()
					return m, tea.Quit // Quit after printing path
				}
			case key.Matches(msg, m.Keymap.ConfirmNo), key.Matches(msg, m.Keymap.ClearFilter):
				m.ShowConfirm = false
				m.FinalPath = ""
			}
			m.Viewport.SetContent(m.RenderEntries())
			return m, nil
		}
	}

	// Handle file operation confirmation
	if m.ConfirmOperation {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, m.Keymap.ConfirmYes):
				// Execute the confirmed action
				if m.ConfirmAction != nil {
					err := m.ConfirmAction()
					if err != nil {
						m.Err = err
					}
					// Refresh directory after operation
					m.ReadDir(m.Cwd)
				}
				m.ConfirmOperation = false
				m.ConfirmAction = nil
				m.ConfirmPrompt = ""
			case key.Matches(msg, m.Keymap.ConfirmNo), key.Matches(msg, m.Keymap.ClearFilter):
				m.ConfirmOperation = false
				m.ConfirmAction = nil
				m.ConfirmPrompt = ""
			}
			m.Viewport.SetContent(m.RenderEntries())
			return m, nil
		}
	}

	// Get current entries list based on filter state
	currentEntries := m.Entries
	if m.Filtering || m.FilterInput.Value() != "" {
		currentEntries = m.FilteredEntries
	}

	// Process general messages
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return handleWindowResize(m, msg)

	case RenamePromptMsg:
		m.Filtering = true
		m.FilterInput.Reset()
		m.FilterInput.Focus()
		m.FilterInput.Placeholder = fmt.Sprintf("New name for %s:", msg.Type)
		return m, textinput.Blink

	case NewFilePromptMsg:
		m.Filtering = true
		m.FilterInput.Reset()
		m.FilterInput.Focus()
		m.FilterInput.Placeholder = "New file name:"
		return m, textinput.Blink

	case NewDirPromptMsg:
		m.Filtering = true
		m.FilterInput.Reset()
		m.FilterInput.Focus()
		m.FilterInput.Placeholder = "New directory name:"
		return m, textinput.Blink

	case ShowHelpMsg:
		m.ShowHelp = bool(msg)
		return m, nil

	case tea.KeyMsg:
		// Always handle quit first
		if key.Matches(msg, m.Keymap.Quit) {
			// Save state before quitting
			state, _ := LoadState()
			state.LastDirectory = m.Cwd
			state.Bookmarks = m.Bookmarks
			_ = SaveState(&state)
			return m, tea.Quit
		}

		// Handle filter input keys if filtering
		if m.Filtering {
			return handleFilteringMode(m, msg)
		}

		// Handle bookmarks view
		if m.ShowBookmarks {
			return handleBookmarksMode(m, msg)
		}

		// Handle regular navigation and actions (when not filtering)
		return handleNavigationMode(m, msg, currentEntries)
	}

	// Handle keyboard and mouse events in the viewport
	var viewportCmd tea.Cmd
	m.Viewport, viewportCmd = m.Viewport.Update(msg)
	cmds = append(cmds, viewportCmd)
	
	if m.ShowPreview && m.PreviewViewport.Height > 0 {
		var previewCmd tea.Cmd
		m.PreviewViewport, previewCmd = m.PreviewViewport.Update(msg)
		cmds = append(cmds, previewCmd)
	}

	return m, tea.Batch(cmds...)
}

// handleWindowResize updates layout on window size change
func handleWindowResize(m Model, msg tea.WindowSizeMsg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	headerHeight := lipgloss.Height(m.HeaderView())
	statusHeight := 1 // Status bar height
	filterHeight := 0
	if m.Filtering {
		filterHeight = lipgloss.Height(m.FilterInput.View()) + 1 // +1 for padding
	}
	footerHeight := lipgloss.Height(m.FooterView())
	verticalMarginHeight := headerHeight + statusHeight + filterHeight + footerHeight

	if !m.Ready {
		// Initialize viewport
		m.Viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
		m.Viewport.YPosition = headerHeight + statusHeight + filterHeight
		m.Viewport.HighPerformanceRendering = false
		m.FilterInput.Width = msg.Width - 4 // Adjust filter input width slightly less than full
		
		// Initialize preview viewport
		m.PreviewViewport = viewport.New(msg.Width/2, msg.Height-verticalMarginHeight)
		m.PreviewViewport.YPosition = headerHeight + statusHeight + filterHeight
		
		m.ApplyFilter() // Apply initial (empty) filter now that viewport is sized
		m.Ready = true

		// Apply border style
		if m.Styles.Base.GetBorderStyle() != lipgloss.HiddenBorder() {
			borderSize := m.Styles.Base.GetHorizontalBorderSize()
			m.Viewport.Style = m.Styles.Base.Copy().
				Width(msg.Width - borderSize).
				Height(msg.Height - verticalMarginHeight - m.Styles.Base.GetVerticalBorderSize())
			
			m.PreviewViewport.Style = m.Styles.Base.Copy().
				Width(msg.Width/2 - borderSize).
				Height(msg.Height - verticalMarginHeight - m.Styles.Base.GetVerticalBorderSize())
		}
		
		cmds = append(cmds, textinput.Blink)
	} else {
		// Update viewport and filter input size on resize
		m.Viewport.Width = msg.Width
		m.Viewport.Height = msg.Height - verticalMarginHeight
		m.Viewport.YPosition = headerHeight + statusHeight + filterHeight
		m.FilterInput.Width = msg.Width - 4
		
		// Update preview viewport
		previewWidth := msg.Width / 2
		if !m.ShowPreview {
			previewWidth = 0
		}
		m.PreviewViewport.Width = previewWidth
		m.PreviewViewport.Height = msg.Height - verticalMarginHeight
		m.PreviewViewport.YPosition = headerHeight + statusHeight + filterHeight

		// Update border size on resize too
		if m.Styles.Base.GetBorderStyle() != lipgloss.HiddenBorder() {
			borderSize := m.Styles.Base.GetHorizontalBorderSize()
			m.Viewport.Style = m.Styles.Base.Copy().
				Width(msg.Width - borderSize).
				Height(msg.Height - verticalMarginHeight - m.Styles.Base.GetVerticalBorderSize())
			
			m.PreviewViewport.Style = m.Styles.Base.Copy().
				Width(previewWidth - borderSize).
				Height(msg.Height - verticalMarginHeight - m.Styles.Base.GetVerticalBorderSize())
		}
	}

	return m, tea.Batch(cmds...)
}

// handleImagePreviewMode handles keypresses in image preview mode
func handleImagePreviewMode(m Model, msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		keyStr := msg.String()
		switch {
		case keyStr == "k" || keyStr == "up":
			if len(m.ImageFilesInDir) > 0 {
				m.CurrentPreviewImageIndex--
				if m.CurrentPreviewImageIndex < 0 {
					m.CurrentPreviewImageIndex = len(m.ImageFilesInDir) - 1
				}
				if err := m.DisplayCurrentImageInGT(); err != nil {
					m.Err = err
				}
			}
			return m, nil
		case keyStr == "j" || keyStr == "down":
			if len(m.ImageFilesInDir) > 0 {
				m.CurrentPreviewImageIndex++
				if m.CurrentPreviewImageIndex >= len(m.ImageFilesInDir) {
					m.CurrentPreviewImageIndex = 0
				}
				if err := m.DisplayCurrentImageInGT(); err != nil {
					m.Err = err
				}
			}
			return m, nil
		case key.Matches(msg, m.Keymap.Quit) || key.Matches(msg, m.Keymap.ClearFilter) || key.Matches(msg, m.Keymap.Back) || keyStr == "q" || keyStr == "escape":
			m.IsInImagePreviewMode = false
			fmt.Print("\x1b[2J\x1b[H") // Clear screen
			_ = os.Stdout.Sync()
			m.Err = nil

			// If there was filter text, ensure the filter input is active again
			if m.FilterInput.Value() != "" {
				m.Filtering = true
				m.FilterInput.Focus()
				return m, textinput.Blink
			} else {
				m.Filtering = false
			}
			m.Viewport.SetContent(m.RenderEntries())
			return m, nil
		}
		return m, nil
	}
	return m, nil
}

// handleFilteringMode handles keypresses when in filtering mode
func handleFilteringMode(m Model, msg tea.KeyMsg) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	
	switch {
	// Clear filter and exit filtering mode (ESC key)
	case key.Matches(msg, m.Keymap.ClearFilter):
		m.Filtering = false
		m.FilterInput.Blur()
		
		// Only reset filter if it's a filter operation, not a rename or other operation
		if strings.HasPrefix(m.FilterInput.Placeholder, "Filter") {
			m.FilterInput.Reset() // Clears text
			m.ApplyFilter()       // Shows all entries
		}

	// Commit filter and return to list navigation (ENTER key)
	case msg.Type == tea.KeyEnter:
		m.Filtering = false
		m.FilterInput.Blur()
		
		// Check if this is a filter or another operation
		if strings.HasPrefix(m.FilterInput.Placeholder, "New file name:") {
			ops := NewFileOperations(&m)
			if err := ops.CreateFile(m.FilterInput.Value()); err != nil {
				m.Err = err
			} else {
				m.ReadDir(m.Cwd) // Refresh directory
			}
			m.FilterInput.Reset()
		} else if strings.HasPrefix(m.FilterInput.Placeholder, "New directory name:") {
			ops := NewFileOperations(&m)
			if err := ops.CreateDirectory(m.FilterInput.Value()); err != nil {
				m.Err = err
			} else {
				m.ReadDir(m.Cwd) // Refresh directory
			}
			m.FilterInput.Reset()
		} else if strings.HasPrefix(m.FilterInput.Placeholder, "New name for") {
			ops := NewFileOperations(&m)
			if err := ops.RenameFile(m.FilterInput.Value()); err != nil {
				m.Err = err
			} else {
				m.ReadDir(m.Cwd) // Refresh directory
			}
			m.FilterInput.Reset()
		} else {
			// Regular filtering - keep the filter value but apply it
			m.ApplyFilter()
		}

	// Update filter input (for character keys)
	default:
		var filterCmd tea.Cmd
		m.FilterInput, filterCmd = m.FilterInput.Update(msg)
		if strings.HasPrefix(m.FilterInput.Placeholder, "Filter") {
			m.ApplyFilter() // Re-apply filter on every keystroke for filter operations
		}
		cmds = append(cmds, filterCmd)
	}
	
	// Ensure viewport updates after filter actions
	m.Viewport.SetContent(m.RenderEntries())
	return m, tea.Batch(cmds...)
}

// handleBookmarksMode handles keypresses when in bookmarks view
func handleBookmarksMode(m Model, msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.Keymap.Back), key.Matches(msg, m.Keymap.ClearFilter):
		m.ShowBookmarks = false
		m.Viewport.SetContent(m.RenderEntries())
		return m, nil
		
	case key.Matches(msg, m.Keymap.Up):
		if m.Cursor > 0 {
			m.Cursor--
			m.Viewport.SetContent(m.RenderBookmarks())
		}
		
	case key.Matches(msg, m.Keymap.Down):
		if m.Cursor < len(m.Bookmarks)-1 {
			m.Cursor++
			m.Viewport.SetContent(m.RenderBookmarks())
		}
		
	case key.Matches(msg, m.Keymap.SelectEnter), key.Matches(msg, m.Keymap.SelectSpace):
		if len(m.Bookmarks) > 0 && m.Cursor < len(m.Bookmarks) {
			selectedPath := m.Bookmarks[m.Cursor]
			m.ShowBookmarks = false
			m.ReadDir(selectedPath)
		}
		
	case key.Matches(msg, m.Keymap.Delete):
		if len(m.Bookmarks) > 0 && m.Cursor < len(m.Bookmarks) {
			ops := NewFileOperations(&m)
			if err := ops.RemoveBookmark(m.Cursor); err != nil {
				m.Err = err
			}
			// Adjust cursor if needed
			if m.Cursor >= len(m.Bookmarks) && len(m.Bookmarks) > 0 {
				m.Cursor = len(m.Bookmarks) - 1
			}
			m.Viewport.SetContent(m.RenderBookmarks())
		}
	}
	
	return m, nil
}

// handleNavigationMode handles regular navigation mode
func handleNavigationMode(m Model, msg tea.KeyMsg, currentEntries []fs.DirEntry) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	
	switch {
	// Help screen
	case key.Matches(msg, m.Keymap.Help):
		m.ShowHelp = true
		
	// Start filtering
	case key.Matches(msg, m.Keymap.StartFilter):
		m.Filtering = true
		m.FilterInput.Placeholder = "Filter..."
		m.FilterInput.Focus()
		cmds = append(cmds, textinput.Blink)
		
	// Toggle hidden files
	case key.Matches(msg, m.Keymap.ToggleHidden):
		m.ShowHidden = !m.ShowHidden
		m.ReadDir(m.Cwd) // Re-read directory with new setting
		
	// Toggle preview pane
	case key.Matches(msg, m.Keymap.TogglePreview):
		m.ShowPreview = !m.ShowPreview
		if m.ShowPreview {
			// Try loading a preview for the current selection
			selectedEntry, err := getSelectedEntry(m)
			if err == nil && !selectedEntry.IsDir() {
				filePath := filepath.Join(m.Cwd, selectedEntry.Name())
				m.LoadFilePreview(filePath)
			}
		}
		
	// Bookmark operations
	case key.Matches(msg, m.Keymap.AddBookmark):
		ops := NewFileOperations(&m)
		if err := ops.AddBookmark(); err != nil {
			m.Err = err
		} else {
			m.Err = fmt.Errorf("Bookmarked: %s", m.Cwd)
		}
		
	case key.Matches(msg, m.Keymap.ShowBookmarks):
		if len(m.Bookmarks) == 0 {
			m.Err = fmt.Errorf("No bookmarks saved")
		} else {
			m.ShowBookmarks = true
			m.Cursor = 0 // Reset cursor for bookmark view
			m.Viewport.SetContent(m.RenderBookmarks())
		}
		
	// File operations
	case key.Matches(msg, m.Keymap.Copy):
		ops := NewFileOperations(&m)
		if err := ops.CopyFile(); err != nil {
			m.Err = err
		}
		
	case key.Matches(msg, m.Keymap.Cut):
		ops := NewFileOperations(&m)
		if err := ops.CutFile(); err != nil {
			m.Err = err
		}
		
	case key.Matches(msg, m.Keymap.Paste):
		ops := NewFileOperations(&m)
		if err := ops.PasteFile(); err != nil {
			m.Err = err
		}
		
	case key.Matches(msg, m.Keymap.Delete):
		ops := NewFileOperations(&m)
		if err := ops.DeleteFile(); err != nil {
			m.Err = err
		}
		
	case key.Matches(msg, m.Keymap.Rename):
		selectedEntry, err := getSelectedEntry(m)
		if err != nil {
			m.Err = err
		} else {
			entryType := "file"
			if selectedEntry.IsDir() {
				entryType = "directory"
			}
			return m, func() tea.Msg {
				return RenamePromptMsg{Type: entryType}
			}
		}

	// Navigation keys
	case key.Matches(msg, m.Keymap.Up):
		if m.Cursor > 0 {
			m.Cursor--
			// Update preview if enabled
			if m.ShowPreview {
				updatePreviewForCurrentSelection(m)
			}
		}
		
	case key.Matches(msg, m.Keymap.Down):
		if m.Cursor < len(currentEntries)-1 {
			m.Cursor++
			// Update preview if enabled
			if m.ShowPreview {
				updatePreviewForCurrentSelection(m)
			}
		}
		
	case key.Matches(msg, m.Keymap.PageUp):
		m.Viewport.HalfViewUp()
		m.Cursor = Max(0, m.Cursor-m.Viewport.Height/2)
		if m.ShowPreview {
			updatePreviewForCurrentSelection(m)
		}
		
	case key.Matches(msg, m.Keymap.PageDown):
		m.Viewport.HalfViewDown()
		m.Cursor = Min(len(currentEntries)-1, m.Cursor+m.Viewport.Height/2)
		if m.ShowPreview {
			updatePreviewForCurrentSelection(m)
		}
		
	case key.Matches(msg, m.Keymap.GoToTop):
		m.Viewport.GotoTop()
		m.Cursor = 0
		if m.ShowPreview {
			updatePreviewForCurrentSelection(m)
		}
		
	case key.Matches(msg, m.Keymap.GoToBottom):
		m.Viewport.GotoBottom()
		m.Cursor = len(currentEntries) - 1
		if m.ShowPreview {
			updatePreviewForCurrentSelection(m)
		}

	// Select item
	case key.Matches(msg, m.Keymap.SelectEnter), key.Matches(msg, m.Keymap.SelectSpace):
		if len(currentEntries) == 0 || m.Cursor >= len(currentEntries) {
			break // Nothing to select
		}
		
		selectedEntry := currentEntries[m.Cursor]
		absPath := filepath.Join(m.Cwd, selectedEntry.Name())

		if selectedEntry.IsDir() {
			m.ReadDir(absPath) // Navigate into directory
		} else {
			m.FinalPath = absPath
			if IsImageFile(absPath) {
				m = handleImageSelection(m, absPath)
			} else {
				m.ShowConfirm = true
			}
		}

	// Go back
	case key.Matches(msg, m.Keymap.Back):
		parentDir := filepath.Dir(m.Cwd)
		if parentDir != m.Cwd {
			m.ReadDir(parentDir) // Go up
		} else {
			m.Err = fmt.Errorf("Already at root directory")
		}
	}
	
	m.Viewport.SetContent(m.RenderEntries())
	return m, tea.Batch(cmds...)
}

// handleImageSelection prepares for image preview mode
func handleImageSelection(m Model, absPath string) Model {
	m.IsInImagePreviewMode = true
	m.ShowConfirm = false
	m.ImageFilesInDir = []fs.DirEntry{}

	sourceEntries := m.Entries
	if m.Filtering || m.FilterInput.Value() != "" {
		sourceEntries = m.FilteredEntries
	}

	selectedIndexInImages := -1
	for _, entry := range sourceEntries {
		if entry.IsDir() {
			continue
		}
		entryAbsPath := filepath.Join(m.Cwd, entry.Name())
		if IsImageFile(entryAbsPath) {
			m.ImageFilesInDir = append(m.ImageFilesInDir, entry)
			if entryAbsPath == absPath {
				selectedIndexInImages = len(m.ImageFilesInDir) - 1
			}
		}
	}

	if len(m.ImageFilesInDir) > 0 {
		if selectedIndexInImages != -1 {
			m.CurrentPreviewImageIndex = selectedIndexInImages
		} else {
			m.CurrentPreviewImageIndex = 0
		}
		if err := m.DisplayCurrentImageInGT(); err != nil {
			m.Err = err
			m.IsInImagePreviewMode = false
		}
	} else {
		m.Err = fmt.Errorf("No images found in current view")
		m.IsInImagePreviewMode = false
	}
	return m
}

// updatePreviewForCurrentSelection updates the preview content for the current selection
func updatePreviewForCurrentSelection(m Model) Model {
	selectedEntry, err := getSelectedEntry(m)
	if err != nil {
		m.ShowPreview = false
		return m
	}
	
	if selectedEntry.IsDir() {
		// For directories, show a simple message or directory info
		m.PreviewContent = fmt.Sprintf("Directory: %s", selectedEntry.Name())
		m.PreviewViewport.SetContent(m.PreviewContent)
		return m
	}
	
	// For files, try to show content preview
	filePath := filepath.Join(m.Cwd, selectedEntry.Name())
	m.LoadFilePreview(filePath)
	return m
}

// getSelectedEntry returns the currently selected entry
func getSelectedEntry(m Model) (fs.DirEntry, error) {
	// Check if there are any entries
	entriesToUse := m.Entries
	if m.Filtering || m.FilterInput.Value() != "" {
		entriesToUse = m.FilteredEntries
	}

	if len(entriesToUse) == 0 {
		return nil, fmt.Errorf("No entries to select from")
	}

	// Check cursor bounds
	if m.Cursor < 0 || m.Cursor >= len(entriesToUse) {
		return nil, fmt.Errorf("Invalid selection")
	}

	return entriesToUse[m.Cursor], nil
}