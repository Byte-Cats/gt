package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"gf/internal"
	"github.com/charmbracelet/bubbles/textinput"
)

// Model wraps the internal model for the TUI
type Model struct {
	internal internal.Model
	showHelp bool
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" && !m.internal.Filtering && !m.internal.ShowConfirm && !m.internal.IsInImagePreviewMode {
			return m, tea.Quit
		}
	}
	
	// Pass message to internal model
	var cmd tea.Cmd
	m.internal, cmd = m.internal.Update(msg)
	
	// Return updated model and command
	return m, cmd
}

func (m Model) View() string {
	return m.internal.View()
}

func main() {
	// Create internal model
	internalModel := internal.NewModel()
	
	// Wrap it in our model
	m := Model{
		internal: internalModel,
		showHelp: false,
	}
	
	// Create and start program
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running application: %v\n", err)
		os.Exit(1)
	}
	
	// The path output is now handled in internal.Update before calling tea.Quit
}