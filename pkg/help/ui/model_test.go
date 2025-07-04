package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/stretchr/testify/assert"
)

func TestViewingModeYAndOShortcuts(t *testing.T) {
	// Create a test help system with a sample section
	hs := help.NewHelpSystem()
	section := &help.Section{
		Section: &model.Section{
			Slug:    "test-section",
			Title:   "Test Section",
			Content: "This is a test section content",
		},
	}
	hs.AddSection(section)

	// Create model
	model := New(hs)
	model.SetSize(80, 24)

	// Set up viewing mode with the test section
	model.state = stateViewing
	model.currentSection = section

	// Test 'y' key (copy)
	yMsg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'y'},
	}

	updatedModel, cmd := model.Update(yMsg)
	assert.NotNil(t, updatedModel)
	assert.NotNil(t, cmd)

	// The model should remain in viewing state
	m := updatedModel.(*Model)
	assert.Equal(t, stateViewing, m.state)
	assert.Equal(t, section, m.currentSection)

	// Test 'o' key (quit with output)
	oMsg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'o'},
	}

	updatedModel, _ = model.Update(oMsg)
	assert.NotNil(t, updatedModel)

	// The model should quit with output
	m = updatedModel.(*Model)
	assert.True(t, m.quitWithOutput)
	assert.Equal(t, section, m.currentSection)
}

func TestViewingModeStillHandlesOtherKeys(t *testing.T) {
	// Create a test help system with a sample section
	hs := help.NewHelpSystem()
	section := &help.Section{
		Section: &model.Section{
			Slug:    "test-section",
			Title:   "Test Section",
			Content: "This is a test section content",
		},
	}
	hs.AddSection(section)

	// Create model
	model := New(hs)
	model.SetSize(80, 24)

	// Set up viewing mode with the test section
	model.state = stateViewing
	model.currentSection = section

	// Test 'q' key (quit to normal mode)
	qMsg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'q'},
	}

	updatedModel, _ := model.Update(qMsg)
	assert.NotNil(t, updatedModel)

	// The model should return to normal state
	m := updatedModel.(*Model)
	assert.Equal(t, stateNormal, m.state)
	assert.False(t, m.quitWithOutput)
}

func TestNormalModeYAndOShortcuts(t *testing.T) {
	// Create a test help system with a sample section
	hs := help.NewHelpSystem()
	section := &help.Section{
		Section: &model.Section{
			Slug:    "test-section",
			Title:   "Test Section",
			Content: "This is a test section content",
		},
	}
	hs.AddSection(section)

	// Create model and initialize
	model := New(hs)
	model.SetSize(80, 24)

	// Simulate having results and a selected item
	model.results = []*help.Section{section}
	model.state = stateNormal

	// Test 'y' key (copy)
	yMsg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'y'},
	}

	updatedModel, cmd := model.Update(yMsg)
	assert.NotNil(t, updatedModel)
	assert.NotNil(t, cmd)

	// The model should remain in normal state
	m := updatedModel.(*Model)
	assert.Equal(t, stateNormal, m.state)

	// Test 'o' key (quit with output)
	oMsg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'o'},
	}

	updatedModel, _ = model.Update(oMsg)
	assert.NotNil(t, updatedModel)

	// The model should quit with output
	m = updatedModel.(*Model)
	assert.True(t, m.quitWithOutput)
	assert.Equal(t, section, m.currentSection)
}
