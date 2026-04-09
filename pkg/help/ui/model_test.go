package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/stretchr/testify/assert"
)

func TestViewingModeYAndOShortcuts(t *testing.T) {
	hs := help.NewHelpSystem()
	section := &model.Section{
		Slug:    "test-section",
		Title:   "Test Section",
		Content: "This is a test section content",
	}
	hs.AddSection(section)

	uiModel := New(hs)
	uiModel.SetSize(80, 24)
	uiModel.state = stateViewing
	uiModel.CurrentSection = section

	// Test 'y' key (copy)
	yMsg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'y'},
	}

	updatedModel, cmd := uiModel.Update(yMsg)
	assert.NotNil(t, updatedModel)
	assert.NotNil(t, cmd)

	m := updatedModel.(*Model)
	assert.Equal(t, stateViewing, m.state)
	assert.Equal(t, section, m.CurrentSection)

	// Test 'o' key (quit with output)
	oMsg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'o'},
	}

	updatedModel, _ = uiModel.Update(oMsg)
	assert.NotNil(t, updatedModel)

	m = updatedModel.(*Model)
	assert.True(t, m.QuitWithOutput)
	assert.Equal(t, section, m.CurrentSection)
}

func TestViewingModeStillHandlesOtherKeys(t *testing.T) {
	hs := help.NewHelpSystem()
	section := &model.Section{
		Slug:    "test-section",
		Title:   "Test Section",
		Content: "This is a test section content",
	}
	hs.AddSection(section)

	uiModel := New(hs)
	uiModel.SetSize(80, 24)
	uiModel.state = stateViewing
	uiModel.CurrentSection = section

	qMsg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'q'},
	}

	updatedModel, _ := uiModel.Update(qMsg)
	assert.NotNil(t, updatedModel)

	m := updatedModel.(*Model)
	assert.Equal(t, stateNormal, m.state)
	assert.False(t, m.QuitWithOutput)
}

func TestNormalModeYAndOShortcuts(t *testing.T) {
	hs := help.NewHelpSystem()
	section := &model.Section{
		Slug:    "test-section",
		Title:   "Test Section",
		Content: "This is a test section content",
	}
	hs.AddSection(section)

	uiModel := New(hs)
	uiModel.SetSize(80, 24)
	uiModel.results = []*model.Section{section}
	uiModel.state = stateNormal

	yMsg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'y'},
	}

	updatedModel, cmd := uiModel.Update(yMsg)
	assert.NotNil(t, updatedModel)
	assert.NotNil(t, cmd)

	m := updatedModel.(*Model)
	assert.Equal(t, stateNormal, m.state)

	oMsg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'o'},
	}

	updatedModel, _ = uiModel.Update(oMsg)
	assert.NotNil(t, updatedModel)

	m = updatedModel.(*Model)
	assert.True(t, m.QuitWithOutput)
}
