package ui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/mattn/go-isatty"
)

// RunUI starts the interactive help UI
func RunUI(helpSystem *help.HelpSystem) error {
	model := New(helpSystem)

	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("error running help UI: %w", err)
	}

	// Check if we should output the selected section
	if m, ok := finalModel.(*Model); ok && m.quitWithOutput && m.currentSection != nil {
		// Check if output is piped or we're in an interactive terminal
		if isatty.IsTerminal(os.Stdout.Fd()) {
			// Interactive terminal - render with glamour
			content := m.currentSection.Content
			if content != "" {
				renderer, err := glamour.NewTermRenderer(
					glamour.WithAutoStyle(),
					glamour.WithWordWrap(120),
				)
				if err == nil {
					rendered, err := renderer.Render(content)
					if err == nil {
						fmt.Print(rendered)
						return nil
					}
				}
			}
		}
		// Fallback to plain text for piped output or if glamour fails
		fmt.Print(m.currentSection.Content)
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
		// Check if output is piped or we're in an interactive terminal
		if isatty.IsTerminal(os.Stdout.Fd()) {
			// Interactive terminal - render with glamour
			content := m.currentSection.Content
			if content != "" {
				renderer, err := glamour.NewTermRenderer(
					glamour.WithAutoStyle(),
					glamour.WithWordWrap(120),
				)
				if err == nil {
					rendered, err := renderer.Render(content)
					if err == nil {
						fmt.Print(rendered)
						return m.currentSection, nil
					}
				}
			}
		}
		// Fallback to plain text for piped output or if glamour fails
		fmt.Print(m.currentSection.Content)
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
