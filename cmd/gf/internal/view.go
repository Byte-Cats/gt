package internal

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// HeaderView renders the header section
func (m *Model) HeaderView() string {
	title := m.Styles.Header.Render("Current: " + m.Cwd)
	
	// Don't repeat the line if border is enabled, as viewport will have top border
	if m.Styles.Base.GetBorderStyle() == lipgloss.HiddenBorder() {
		line := strings.Repeat("─", m.Viewport.Width)
		return lipgloss.JoinVertical(lipgloss.Left, title, line)
	}
	return title // Just the title if border is present
}

// FilterView renders the filter input field
func (m *Model) FilterView() string {
	if !m.Filtering {
		return "" // Don't show if not actively filtering
	}
	return m.FilterInput.View() // Return the text input view
}

// FooterView renders the footer section with help text
func (m *Model) FooterView() string {
	var help strings.Builder

	// Different footer views for different modes
	if m.IsInImagePreviewMode {
		help.WriteString("j/k navigate • esc/q exit")
	} else if m.ShowBookmarks {
		help.WriteString(m.Keymap.Up.Help().Key + "/" + m.Keymap.Down.Help().Key + " nav • ")
		help.WriteString(m.Keymap.SelectEnter.Help().Key + " select • ")
		help.WriteString(m.Keymap.Back.Help().Key + " back • ")
		help.WriteString(m.Keymap.Quit.Help().Key + " quit")
	} else if m.ConfirmOperation {
		help.WriteString(m.Keymap.ConfirmYes.Help().Key + " confirm • ")
		help.WriteString(m.Keymap.ConfirmNo.Help().Key + " cancel")
	} else {
		// Basic navigation help
		help.WriteString(m.Keymap.Up.Help().Key + "/" + m.Keymap.Down.Help().Key + " nav • ")
		help.WriteString(m.Keymap.SelectEnter.Help().Key + " sel • ")
		
		if m.Filtering {
			help.WriteString(m.Keymap.ClearFilter.Help().Key + " clear • ")
		} else {
			help.WriteString(m.Keymap.StartFilter.Help().Key + " filter • ")
			help.WriteString(m.Keymap.ToggleHidden.Help().Key + " hidden • ")
			help.WriteString(m.Keymap.AddBookmark.Help().Key + "/" + m.Keymap.ShowBookmarks.Help().Key + " bkmrk • ")
		}
		
		help.WriteString(m.Keymap.Help.Help().Key + " help • ")
		help.WriteString(m.Keymap.Quit.Help().Key + " quit")
	}

	footerText := help.String()

	if m.ShowConfirm {
		footerText = fmt.Sprintf("Confirm selection: %s? (%s/%s)",
			filepath.Base(m.FinalPath),
			m.Keymap.ConfirmYes.Help().Key,
			m.Keymap.ConfirmNo.Help().Key,
		)
	} else if m.ConfirmOperation {
		footerText = m.ConfirmPrompt + " (y/n)"
	} else if m.Err != nil {
		footerText = m.Styles.Error.Render(m.Err.Error())
	}

	// Don't repeat the line if border is enabled
	if m.Styles.Base.GetBorderStyle() == lipgloss.HiddenBorder() {
		line := strings.Repeat("─", m.Viewport.Width)
		return lipgloss.JoinVertical(lipgloss.Left, line, m.Styles.Footer.Render(footerText))
	}
	return m.Styles.Footer.Render(footerText) // Just the text if border is present
}

// HelpView renders the help screen
func (m *Model) HelpView() string {
	var b strings.Builder
	
	b.WriteString("# Keyboard Shortcuts\n\n")
	
	// Navigation
	b.WriteString("## Navigation\n")
	b.WriteString(fmt.Sprintf("%-15s %s\n", m.Keymap.Up.Help().Key, "Move cursor up"))
	b.WriteString(fmt.Sprintf("%-15s %s\n", m.Keymap.Down.Help().Key, "Move cursor down"))
	b.WriteString(fmt.Sprintf("%-15s %s\n", m.Keymap.PageUp.Help().Key, "Page up"))
	b.WriteString(fmt.Sprintf("%-15s %s\n", m.Keymap.PageDown.Help().Key, "Page down"))
	b.WriteString(fmt.Sprintf("%-15s %s\n", m.Keymap.GoToTop.Help().Key, "Go to top"))
	b.WriteString(fmt.Sprintf("%-15s %s\n", m.Keymap.GoToBottom.Help().Key, "Go to bottom"))
	b.WriteString(fmt.Sprintf("%-15s %s\n", m.Keymap.SelectEnter.Help().Key, "Select item"))
	b.WriteString(fmt.Sprintf("%-15s %s\n", m.Keymap.Back.Help().Key, "Go to parent directory"))
	
	// Filtering
	b.WriteString("\n## Filtering\n")
	b.WriteString(fmt.Sprintf("%-15s %s\n", m.Keymap.StartFilter.Help().Key, "Start filtering"))
	b.WriteString(fmt.Sprintf("%-15s %s\n", m.Keymap.ClearFilter.Help().Key, "Clear filter"))
	
	// Bookmarks
	b.WriteString("\n## Bookmarks\n")
	b.WriteString(fmt.Sprintf("%-15s %s\n", m.Keymap.AddBookmark.Help().Key, "Add bookmark"))
	b.WriteString(fmt.Sprintf("%-15s %s\n", m.Keymap.ShowBookmarks.Help().Key, "Show bookmarks"))
	
	// File Operations
	b.WriteString("\n## File Operations\n")
	b.WriteString(fmt.Sprintf("%-15s %s\n", m.Keymap.Copy.Help().Key, "Copy file/directory"))
	b.WriteString(fmt.Sprintf("%-15s %s\n", m.Keymap.Cut.Help().Key, "Cut file/directory"))
	b.WriteString(fmt.Sprintf("%-15s %s\n", m.Keymap.Paste.Help().Key, "Paste file/directory"))
	b.WriteString(fmt.Sprintf("%-15s %s\n", m.Keymap.Delete.Help().Key, "Delete file/directory"))
	b.WriteString(fmt.Sprintf("%-15s %s\n", m.Keymap.Rename.Help().Key, "Rename file/directory"))
	
	// Display
	b.WriteString("\n## Display\n")
	b.WriteString(fmt.Sprintf("%-15s %s\n", m.Keymap.ToggleHidden.Help().Key, "Toggle hidden files"))
	b.WriteString(fmt.Sprintf("%-15s %s\n", m.Keymap.TogglePreview.Help().Key, "Toggle preview pane"))
	
	// General
	b.WriteString("\n## General\n")
	b.WriteString(fmt.Sprintf("%-15s %s\n", m.Keymap.Help.Help().Key, "Show/hide help"))
	b.WriteString(fmt.Sprintf("%-15s %s\n", m.Keymap.Quit.Help().Key, "Quit"))
	
	return b.String()
}

// StatusBar renders a compact status bar for the current application state
func (m *Model) StatusBar() string {
	var status []string
	
	// Current directory indicator
	dirName := filepath.Base(m.Cwd)
	if dirName == "." {
		dirName = m.Cwd
	}
	status = append(status, fmt.Sprintf("Dir: %s", dirName))
	
	// Item count
	totalItems := len(m.Entries)
	visibleItems := totalItems
	if m.Filtering || m.FilterInput.Value() != "" {
		visibleItems = len(m.FilteredEntries)
	}
	status = append(status, fmt.Sprintf("Items: %d/%d", visibleItems, totalItems))
	
	// Filter indicator
	if m.FilterInput.Value() != "" {
		status = append(status, fmt.Sprintf("Filter: %s", m.FilterInput.Value()))
	}
	
	// Hidden files indicator
	if m.ShowHidden {
		status = append(status, "Hidden: On")
	}
	
	// Clipboard indicator
	if m.Clipboard != "" {
		clipboardOp := "Copy"
		if m.ClipboardOp == "cut" {
			clipboardOp = "Cut"
		}
		status = append(status, fmt.Sprintf("%s: %s", clipboardOp, filepath.Base(m.Clipboard)))
	}
	
	return strings.Join(status, " | ")
}

// ImagePreviewStatusBar renders status information while in image preview mode
func (m *Model) ImagePreviewStatusBar() string {
	if !m.IsInImagePreviewMode || len(m.ImageFilesInDir) == 0 {
		return "Image preview: No images available"
	}
	
	// Make sure index is valid
	if m.CurrentPreviewImageIndex < 0 || m.CurrentPreviewImageIndex >= len(m.ImageFilesInDir) {
		return "Image preview: Index out of range"
	}
	
	currentImage := m.ImageFilesInDir[m.CurrentPreviewImageIndex]
	return fmt.Sprintf(
		"Image: %s (%d/%d) | Use j/k to navigate, Esc to exit",
		currentImage.Name(),
		m.CurrentPreviewImageIndex+1,
		len(m.ImageFilesInDir),
	)
}

// PreviewPane renders the preview pane for the selected file
func (m *Model) PreviewPane() string {
	if !m.ShowPreview {
		return ""
	}
	
	return m.PreviewViewport.View()
}

// View renders the main application interface
func (m *Model) View() string {
	if !m.Ready {
		return "\n  Initializing..."
	}

	// Image preview mode has a special view
	if m.IsInImagePreviewMode {
		return "\n\n" + m.Styles.Footer.Render(m.ImagePreviewStatusBar())
	}
	
	// Help screen mode
	if m.ShowHelp {
		return m.HelpView()
	}
	
	// Main layout
	filterStr := m.FilterView()
	viewContent := m.Viewport.View()
	previewContent := m.PreviewPane()
	
	// Status bar (below header)
	statusBar := m.StatusBar()
	
	// Add padding below filter if it's shown and border is hidden
	if filterStr != "" && m.Styles.Base.GetBorderStyle() == lipgloss.HiddenBorder() {
		filterStr += "\n"
	}
	
	// Basic layout without preview
	if !m.ShowPreview || previewContent == "" {
		return fmt.Sprintf("%s\n%s\n%s\n%s\n%s",
			m.HeaderView(),
			m.Styles.Header.Render(statusBar),
			filterStr,
			viewContent,
			m.FooterView())
	}
	
	// Layout with preview pane (split screen)
	// This is a simplified version - for a more advanced split layout
	// you'd need to calculate widths and handle resizing better
	return fmt.Sprintf("%s\n%s\n%s\n%s | %s\n%s",
		m.HeaderView(),
		m.Styles.Header.Render(statusBar),
		filterStr,
		viewContent,
		previewContent,
		m.FooterView())
}