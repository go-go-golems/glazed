package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/glazed/pkg/help"
)

const (
	stateSearch = iota
	stateViewing
)

type Model struct {
	helpSystem *help.HelpSystem
	state      int
	
	// Search state
	searchInput string
	cursor      int
	results     []*help.Section
	selected    int
	
	// View state
	currentSection *help.Section
	
	// UI dimensions
	width  int
	height int
	
	// Error state
	err error
}

func New(helpSystem *help.HelpSystem) *Model {
	return &Model{
		helpSystem: helpSystem,
		state:      stateSearch,
		results:    []*help.Section{},
		selected:   0,
	}
}

func (m *Model) Init() tea.Cmd {
	// Load initial results (all sections)
	return m.search("")
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
		
	case tea.KeyMsg:
		switch m.state {
		case stateSearch:
			return m.updateSearch(msg)
		case stateViewing:
			return m.updateViewing(msg)
		}
		
	case searchResultsMsg:
		m.results = msg.results
		m.err = msg.err
		if len(m.results) > 0 && m.selected >= len(m.results) {
			m.selected = 0
		}
		return m, nil
	}
	
	return m, nil
}

func (m *Model) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
		
	case "enter":
		if len(m.results) > 0 && m.selected < len(m.results) {
			m.currentSection = m.results[m.selected]
			m.state = stateViewing
		}
		return m, nil
		
	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}
		return m, nil
		
	case "down", "j":
		if m.selected < len(m.results)-1 {
			m.selected++
		}
		return m, nil
		
	case "backspace":
		if len(m.searchInput) > 0 {
			m.searchInput = m.searchInput[:len(m.searchInput)-1]
			return m, m.search(m.searchInput)
		}
		return m, nil
		
	default:
		if len(msg.String()) == 1 {
			m.searchInput += msg.String()
			return m, m.search(m.searchInput)
		}
	}
	
	return m, nil
}

func (m *Model) updateViewing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
		
	case "esc", "backspace":
		m.state = stateSearch
		return m, nil
	}
	
	return m, nil
}

func (m *Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress q to quit", m.err)
	}
	
	switch m.state {
	case stateSearch:
		return m.viewSearch()
	case stateViewing:
		return m.viewSection()
	}
	
	return ""
}

func (m *Model) viewSearch() string {
	var s strings.Builder
	
	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		MarginBottom(1)
	
	s.WriteString(headerStyle.Render("Glazed Help System"))
	s.WriteString("\n\n")
	
	// Search box
	searchStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("8")).
		Padding(0, 1)
	
	searchPrompt := fmt.Sprintf("Search: %s", m.searchInput)
	s.WriteString(searchStyle.Render(searchPrompt))
	s.WriteString("\n\n")
	
	// Results
	if len(m.results) == 0 {
		s.WriteString("No results found.\n")
	} else {
		s.WriteString(fmt.Sprintf("Results (%d):\n", len(m.results)))
		
		maxResults := 10
		if m.height > 0 {
			maxResults = m.height - 8 // Leave room for header, search box, and instructions
		}
		
		for i, section := range m.results {
			if i >= maxResults {
				break
			}
			
			prefix := "  "
			if i == m.selected {
				prefix = "▶ "
			}
			
			title := section.Title
			if title == "" {
				title = section.Slug
			}
			
			typeStr := ""
			if section.SectionType != 0 {
				typeStr = fmt.Sprintf(" [%s]", section.SectionType.String())
			}
			
			if i == m.selected {
				selectedStyle := lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("11"))
				s.WriteString(selectedStyle.Render(fmt.Sprintf("%s%s%s", prefix, title, typeStr)))
			} else {
				s.WriteString(fmt.Sprintf("%s%s%s", prefix, title, typeStr))
			}
			s.WriteString("\n")
		}
		
		if len(m.results) > maxResults {
			s.WriteString(fmt.Sprintf("... and %d more\n", len(m.results)-maxResults))
		}
	}
	
	// Instructions
	s.WriteString("\n")
	instructionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Italic(true)
	
	instructions := "↑/k up • ↓/j down • enter select • type to search • q quit"
	s.WriteString(instructionStyle.Render(instructions))
	
	return s.String()
}

func (m *Model) viewSection() string {
	if m.currentSection == nil {
		return "No section selected"
	}
	
	var s strings.Builder
	
	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		MarginBottom(1)
	
	title := m.currentSection.Title
	if title == "" {
		title = m.currentSection.Slug
	}
	
	s.WriteString(headerStyle.Render(title))
	s.WriteString("\n\n")
	
	// Metadata
	metaStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8"))
	
	if m.currentSection.SectionType != 0 {
		s.WriteString(metaStyle.Render(fmt.Sprintf("Type: %s", m.currentSection.SectionType.String())))
		s.WriteString("\n")
	}
	
	if len(m.currentSection.Topics) > 0 {
		s.WriteString(metaStyle.Render(fmt.Sprintf("Topics: %v", m.currentSection.Topics)))
		s.WriteString("\n")
	}
	
	if len(m.currentSection.Commands) > 0 {
		s.WriteString(metaStyle.Render(fmt.Sprintf("Commands: %v", m.currentSection.Commands)))
		s.WriteString("\n")
	}
	
	if len(m.currentSection.Flags) > 0 {
		s.WriteString(metaStyle.Render(fmt.Sprintf("Flags: %v", m.currentSection.Flags)))
		s.WriteString("\n")
	}
	
	s.WriteString("\n")
	
	// Content
	contentStyle := lipgloss.NewStyle().
		Width(m.width - 2).
		MarginLeft(1)
	
	content := m.currentSection.Content
	if content == "" {
		content = "No content available"
	}
	
	s.WriteString(contentStyle.Render(content))
	s.WriteString("\n\n")
	
	// Instructions
	instructionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Italic(true)
	
	instructions := "esc/backspace back • q quit"
	s.WriteString(instructionStyle.Render(instructions))
	
	return s.String()
}

type searchResultsMsg struct {
	results []*help.Section
	err     error
}

func (m *Model) search(query string) tea.Cmd {
	return func() tea.Msg {
		var results []*help.Section
		var err error
		
		if query == "" {
			// Return all sections
			results, err = m.helpSystem.QuerySections("")
		} else {
			// Search using the query
			results, err = m.helpSystem.QuerySections(query)
		}
		
		return searchResultsMsg{
			results: results,
			err:     err,
		}
	}
}
