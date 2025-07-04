package ui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/glazed/pkg/help"
)

// RunUI starts the interactive help UI
func RunUI(helpSystem *help.HelpSystem) error {
	model := New(helpSystem)
	
	p := tea.NewProgram(model, tea.WithAltScreen())
	
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running help UI: %w", err)
	}
	
	return nil
}

// RunUIWithOutput starts the interactive help UI and returns the selected section
func RunUIWithOutput(helpSystem *help.HelpSystem) (*help.Section, error) {
	model := New(helpSystem)
	
	// Create a program that captures the final model state
	p := tea.NewProgram(model, tea.WithAltScreen())
	
	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("error running help UI: %w", err)
	}
	
	// Extract the selected section if any
	if m, ok := finalModel.(*Model); ok && m.currentSection != nil {
		return m.currentSection, nil
	}
	
	return nil, nil
}

// MustRunUI runs the UI and exits on error
func MustRunUI(helpSystem *help.HelpSystem) {
	if err := RunUI(helpSystem); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
