package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/glazed/pkg/help"
)

const (
	stateNormal     = iota // Default navigation mode
	stateSearch            // Active search input mode
	stateViewing           // Viewing selected entry
	stateHelp              // Help screen
	stateCheatsheet        // DSL cheatsheet
)

type Model struct {
	helpSystem *help.HelpSystem
	state      int

	// Search state
	searchInput string
	list        list.Model
	results     []*help.Section

	// View state
	currentSection *help.Section
	viewport       viewport.Model

	// UI dimensions
	width  int
	height int

	// Glamour renderer
	glamourRenderer *glamour.TermRenderer

	// Error state
	err error

	// Cursor state
	cursorVisible bool
	cursorBlink   time.Time

	// Clipboard/output state
	copyMessage    string
	copyMessageTs  time.Time
	quitWithOutput bool
}

// listItem represents a help section in the list
type listItem struct {
	section *help.Section
}

func (i listItem) Title() string {
	title := i.section.Title
	if title == "" {
		title = i.section.Slug
	}
	return title
}

func (i listItem) Description() string {
	var parts []string

	if i.section.SectionType != 0 {
		parts = append(parts, i.section.SectionType.String())
	}

	if len(i.section.Topics) > 0 {
		parts = append(parts, fmt.Sprintf("Topics: %s", strings.Join(i.section.Topics, ", ")))
	}

	if len(i.section.Commands) > 0 {
		parts = append(parts, fmt.Sprintf("Commands: %s", strings.Join(i.section.Commands, ", ")))
	}

	if len(parts) == 0 {
		// Show first line of content as description
		lines := strings.Split(strings.TrimSpace(i.section.Content), "\n")
		if len(lines) > 0 && lines[0] != "" {
			desc := lines[0]
			if len(desc) > 80 {
				desc = desc[:77] + "..."
			}
			return desc
		}
		return "No description available"
	}

	return strings.Join(parts, " • ")
}

func (i listItem) FilterValue() string {
	var filterText []string
	filterText = append(filterText, i.section.Title, i.section.Slug)

	if i.section.SectionType != 0 {
		filterText = append(filterText, i.section.SectionType.String())
	}

	filterText = append(filterText, i.section.Topics...)
	filterText = append(filterText, i.section.Commands...)
	filterText = append(filterText, i.section.Flags...)

	return strings.Join(filterText, " ")
}

func New(helpSystem *help.HelpSystem) *Model {
	// Initialize glamour renderer
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
	)
	if err != nil {
		// Fallback to no styling if glamour fails
		renderer = nil
	}

	// Create list
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Glazed Help System"
	l.SetFilteringEnabled(false) // We'll handle search ourselves
	l.SetShowStatusBar(true)
	l.SetShowPagination(true)
	l.SetShowHelp(false) // We'll show our own help

	return &Model{
		helpSystem:      helpSystem,
		state:           stateNormal,
		searchInput:     "",
		list:            l,
		results:         []*help.Section{},
		glamourRenderer: renderer,
		cursorVisible:   true,
		cursorBlink:     time.Now(),
	}
}

func (m *Model) Init() tea.Cmd {
	// Load initial results (all sections)
	return tea.Batch(
		m.search(""),
		m.tickCursor(),
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update list size
		listHeight := m.height - 5 // Leave room for search input and status
		m.list.SetSize(m.width, listHeight)

		// Update viewport size
		m.viewport.Width = m.width - 4
		m.viewport.Height = m.height - 8

		return m, nil

	case tea.KeyMsg:
		// Global keys available from any state
		switch msg.String() {
		case "ctrl+c":
			// Check if we're in normal mode with a selected section and want to quit with output
			if m.state == stateNormal && len(m.results) > 0 {
				selected := m.list.Index()
				if selected >= 0 && selected < len(m.results) {
					m.currentSection = m.results[selected]
					m.quitWithOutput = true
				}
			}
			return m, tea.Quit
		}

		switch m.state {
		case stateNormal:
			return m.updateNormal(msg)
		case stateSearch:
			return m.updateSearch(msg)
		case stateViewing:
			return m.updateViewing(msg)
		case stateHelp:
			return m.updateHelp(msg)
		case stateCheatsheet:
			return m.updateCheatsheet(msg)
		}

	case searchResultsMsg:
		m.results = msg.results
		m.err = msg.err

		// Convert results to list items
		items := make([]list.Item, len(m.results))
		for i, section := range m.results {
			items[i] = listItem{section: section}
		}

		cmd = m.list.SetItems(items)
		return m, cmd

	case cursorTickMsg:
		m.cursorVisible = !m.cursorVisible
		return m, m.tickCursor()

	case copySuccessMsg:
		m.copyMessage = "Copied to clipboard!"
		m.copyMessageTs = time.Now()
		return m, nil

	case copyErrorMsg:
		m.copyMessage = fmt.Sprintf("Copy failed: %v", msg.err)
		m.copyMessageTs = time.Now()
		return m, nil
	}

	return m, nil
}

func (m *Model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "/":
		// Enter search mode
		m.state = stateSearch
		return m, nil

	case "?":
		m.state = stateHelp
		return m, nil

	case "f1":
		m.state = stateCheatsheet
		return m, nil

	case "enter":
		if len(m.results) > 0 {
			selected := m.list.Index()
			if selected >= 0 && selected < len(m.results) {
				m.currentSection = m.results[selected]
				m.state = stateViewing
				return m, m.setupViewport()
			}
		}
		return m, nil

	case "y":
		// Copy selected section to clipboard
		if len(m.results) > 0 {
			selected := m.list.Index()
			if selected >= 0 && selected < len(m.results) {
				return m, m.copySection(m.results[selected])
			}
		}
		return m, nil

	case "o":
		// Quit and output selected section
		if len(m.results) > 0 {
			selected := m.list.Index()
			if selected >= 0 && selected < len(m.results) {
				m.currentSection = m.results[selected]
				m.quitWithOutput = true
				return m, tea.Quit
			}
		}
		return m, nil

	case "up", "k":
		m.list, cmd = m.list.Update(msg)
		return m, cmd

	case "down", "j":
		m.list, cmd = m.list.Update(msg)
		return m, cmd

	case "pgup", "pgdown":
		m.list, cmd = m.list.Update(msg)
		return m, cmd

	default:
		// Handle other navigation keys with the list
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}
}

func (m *Model) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Execute search and exit search mode
		m.state = stateNormal
		return m, m.search(m.searchInput)

	case "esc":
		// Cancel search and exit search mode
		m.state = stateNormal
		return m, nil

	case "backspace", "ctrl+h":
		if len(m.searchInput) > 0 {
			m.searchInput = m.searchInput[:len(m.searchInput)-1]
			return m, m.search(m.searchInput)
		}
		return m, nil

	default:
		// Handle character input for search - check for runes first
		if msg.Type == tea.KeyRunes {
			for _, r := range msg.Runes {
				if r >= 32 && r <= 126 { // printable ASCII
					m.searchInput += string(r)
				}
			}
			return m, m.search(m.searchInput)
		}

		// Handle space key explicitly - it might come as a separate key type
		if msg.String() == " " {
			m.searchInput += " "
			return m, m.search(m.searchInput)
		}

		// Handle other printable characters that might not be in KeyRunes
		if len(msg.String()) == 1 {
			r := []rune(msg.String())[0]
			if r >= 32 && r <= 126 {
				m.searchInput += msg.String()
				return m, m.search(m.searchInput)
			}
		}

		return m, nil
	}
}

func (m *Model) updateViewing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "?":
		m.state = stateHelp
		return m, nil

	case "ctrl+h", "f1":
		m.state = stateCheatsheet
		return m, nil

	case "esc", "backspace", "q":
		m.state = stateNormal
		return m, nil

	case "y":
		// Copy current section to clipboard
		if m.currentSection != nil {
			return m, m.copySection(m.currentSection)
		}
		return m, nil

	case "o":
		// Quit and output current section
		if m.currentSection != nil {
			m.quitWithOutput = true
			return m, tea.Quit
		}
		return m, nil

	default:
		// Let viewport handle scrolling
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}
}

func (m *Model) updateHelp(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "?", "q", "enter", "backspace":
		m.state = stateNormal
		return m, nil
	}
	return m, nil
}

func (m *Model) updateCheatsheet(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+h", "f1", "q", "enter", "backspace":
		m.state = stateNormal
		return m, nil
	}
	return m, nil
}

func (m *Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress Ctrl+C to quit", m.err)
	}

	switch m.state {
	case stateNormal:
		return m.viewNormal()
	case stateSearch:
		return m.viewSearch()
	case stateViewing:
		return m.viewSection()
	case stateHelp:
		return m.viewHelp()
	case stateCheatsheet:
		return m.viewCheatsheet()
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

	s.WriteString(headerStyle.Render("Glazed Help System - Search Mode"))
	s.WriteString("\n")

	// Search box - active with cursor
	searchStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("12")). // Highlight active search
		Padding(0, 1).
		Width(m.width - 4)

	cursor := ""
	if m.cursorVisible {
		cursor = "█"
	}
	searchPrompt := fmt.Sprintf("Search: %s%s", m.searchInput, cursor)
	s.WriteString(searchStyle.Render(searchPrompt))
	s.WriteString("\n")

	// Copy message if present
	if m.copyMessage != "" && time.Since(m.copyMessageTs) < 2*time.Second {
		copyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Italic(true)
		s.WriteString(copyStyle.Render(m.copyMessage))
		s.WriteString("\n")
	}
	s.WriteString("\n")

	// List
	s.WriteString(m.list.View())
	s.WriteString("\n")

	// Instructions
	instructionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Italic(true)

	instructions := "enter: execute search • esc: cancel • type to search • ctrl+c: quit"
	s.WriteString(instructionStyle.Render(instructions))

	return s.String()
}

func (m *Model) viewNormal() string {
	var s strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		MarginBottom(1)

	s.WriteString(headerStyle.Render("Glazed Help System"))
	s.WriteString("\n")

	// Search box - inactive in normal mode
	searchStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("8")).
		Padding(0, 1).
		Width(m.width - 4)

	searchPrompt := fmt.Sprintf("Search: %s (press / to search)", m.searchInput)
	s.WriteString(searchStyle.Render(searchPrompt))
	s.WriteString("\n")

	// Copy message if present
	if m.copyMessage != "" && time.Since(m.copyMessageTs) < 2*time.Second {
		copyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Italic(true)
		s.WriteString(copyStyle.Render(m.copyMessage))
		s.WriteString("\n")
	}
	s.WriteString("\n")

	// List
	s.WriteString(m.list.View())
	s.WriteString("\n")

	// Instructions
	instructionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Italic(true)

	instructions := "/: search • enter: select • y: copy • o: output • ?: help • ctrl+h/F1: query help • ctrl+c: quit"
	s.WriteString(instructionStyle.Render(instructions))

	return s.String()
}

func (m *Model) viewSection() string {
	if m.currentSection == nil {
		return "No section selected"
	}

	var s strings.Builder

	// Header with back instruction
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12"))

	title := m.currentSection.Title
	if title == "" {
		title = m.currentSection.Slug
	}

	s.WriteString(headerStyle.Render(title))
	s.WriteString("\n\n")

	// Viewport with content
	contentStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("8")).
		Padding(1).
		Width(m.width - 2).
		Height(m.height - 6)

	s.WriteString(contentStyle.Render(m.viewport.View()))
	s.WriteString("\n")

	// Copy message if present
	if m.copyMessage != "" && time.Since(m.copyMessageTs) < 2*time.Second {
		copyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Italic(true)
		s.WriteString(copyStyle.Render(m.copyMessage))
		s.WriteString("\n")
	}

	// Instructions
	instructionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Italic(true)

	instructions := "↑/↓: scroll • y: copy • o: output • esc/q: back • ?: help • ctrl+h/F1: query help • ctrl+c: quit"
	s.WriteString(instructionStyle.Render(instructions))

	return s.String()
}

func (m *Model) viewHelp() string {
	helpStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("12")).
		Padding(1).
		Width(m.width - 2).
		Height(m.height - 4)

	help := `
GLAZED HELP SYSTEM - KEYBOARD SHORTCUTS

NORMAL MODE (default):
  /                Enter search mode
  ↑/↓ or j/k       Navigate results
  Enter            View selected entry
  Y                Copy selected entry to clipboard
  O                Quit and output selected entry
  ?                Show this help
  Ctrl+H or F1     Show query DSL help
  Ctrl+C           Quit

SEARCH MODE (activated by /):
  Type             Search for text
  Enter            Execute search and return to normal mode
  Esc              Cancel search and return to normal mode
  Backspace        Delete search character
  Ctrl+C           Quit

VIEWING MODE:
  ↑/↓ or j/k       Scroll content
  Page Up/Down     Scroll page
  Y                Copy current entry to clipboard
  O                Quit and output current entry
  q/Esc/Backspace  Return to normal mode
  ?                Show this help
  Ctrl+H or F1     Show query DSL help
  Ctrl+C           Quit

HELP/CHEATSHEET MODE:
  Any key          Return to normal mode
  Ctrl+C           Quit

QUERY DSL EXAMPLES:
  database tutorial           Search for "database tutorial"
  type:example               Show examples only
  topic:templates            Show template-related content
  'quoted text'              Search for exact phrase
  tutorial AND type:example  Boolean combination
  
Press any key to return to search...`

	return helpStyle.Render(help)
}

func (m *Model) viewCheatsheet() string {
	cheatStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("11")).
		Padding(1).
		Width(m.width - 2).
		Height(m.height - 4)

	cheat := `
QUERY DSL CHEATSHEET

TEXT SEARCH:
  hello world                 Search for "hello world" in content
  'quoted text'               Search for exact phrase with single quotes
  "quoted text"               Search for exact phrase with double quotes

FIELD QUERIES:
  type:example                Find examples only
  type:tutorial               Find tutorials only
  type:topic                  Find general topics
  type:application            Find applications
  topic:database              Find database-related content
  command:json                Find JSON command help
  flag:--output               Find flag documentation
  slug:help-system            Find specific section

BOOLEAN OPERATIONS:
  tutorial AND type:example   Both conditions must match
  examples OR tutorials       Either condition can match
  NOT type:application        Exclude applications
  (A OR B) AND C              Use parentheses for grouping

METADATA QUERIES:
  toplevel:true               Top-level sections only
  default:false               Non-default sections

PRACTICAL EXAMPLES:
  database tutorial           Find database tutorials
  json AND type:example       JSON examples
  topic:templates OR topic:themes    Template or theme help
  'error handling' AND NOT type:application    Error docs, no apps
  
Press any key to return to search...`

	return cheatStyle.Render(cheat)
}

type searchResultsMsg struct {
	results []*help.Section
	err     error
}

type cursorTickMsg struct{}

type copySuccessMsg struct{}

type copyErrorMsg struct {
	err error
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

func (m *Model) setupViewport() tea.Cmd {
	if m.currentSection == nil {
		return nil
	}

	content := m.renderContent(m.currentSection)
	m.viewport.SetContent(content)
	m.viewport.GotoTop()

	return nil
}

func (m *Model) renderContent(section *help.Section) string {
	var s strings.Builder

	// Metadata
	metaStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Italic(true)

	var metadata []string
	if section.SectionType != 0 {
		metadata = append(metadata, fmt.Sprintf("Type: %s", section.SectionType.String()))
	}
	if len(section.Topics) > 0 {
		metadata = append(metadata, fmt.Sprintf("Topics: %s", strings.Join(section.Topics, ", ")))
	}
	if len(section.Commands) > 0 {
		metadata = append(metadata, fmt.Sprintf("Commands: %s", strings.Join(section.Commands, ", ")))
	}
	if len(section.Flags) > 0 {
		metadata = append(metadata, fmt.Sprintf("Flags: %s", strings.Join(section.Flags, ", ")))
	}

	if len(metadata) > 0 {
		s.WriteString(metaStyle.Render(strings.Join(metadata, " • ")))
		s.WriteString("\n\n")
	}

	// Content
	content := section.Content
	if content == "" {
		content = "No content available"
	}

	// Try to render with glamour
	if m.glamourRenderer != nil {
		rendered, err := m.glamourRenderer.Render(content)
		if err == nil {
			content = rendered
		}
	}

	s.WriteString(content)

	return s.String()
}

func (m *Model) tickCursor() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return cursorTickMsg{}
	})
}

func (m *Model) copySection(section *help.Section) tea.Cmd {
	return func() tea.Msg {
		content := m.formatSectionForCopy(section)
		err := clipboard.WriteAll(content)
		if err != nil {
			return copyErrorMsg{err: err}
		}
		return copySuccessMsg{}
	}
}

func (m *Model) formatSectionForCopy(section *help.Section) string {
	var s strings.Builder

	// Title
	title := section.Title
	if title == "" {
		title = section.Slug
	}
	s.WriteString(fmt.Sprintf("# %s\n\n", title))

	// Metadata
	var metadata []string
	if section.SectionType != 0 {
		metadata = append(metadata, fmt.Sprintf("Type: %s", section.SectionType.String()))
	}
	if len(section.Topics) > 0 {
		metadata = append(metadata, fmt.Sprintf("Topics: %s", strings.Join(section.Topics, ", ")))
	}
	if len(section.Commands) > 0 {
		metadata = append(metadata, fmt.Sprintf("Commands: %s", strings.Join(section.Commands, ", ")))
	}
	if len(section.Flags) > 0 {
		metadata = append(metadata, fmt.Sprintf("Flags: %s", strings.Join(section.Flags, ", ")))
	}

	if len(metadata) > 0 {
		s.WriteString(fmt.Sprintf("*%s*\n\n", strings.Join(metadata, " • ")))
	}

	// Content
	s.WriteString(section.Content)

	return s.String()
}

// SetSize sets the UI dimensions for testing
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// FormatSectionForCopy exposes the internal method for testing
func (m *Model) FormatSectionForCopy(section *help.Section) string {
	return m.formatSectionForCopy(section)
}

// State returns the current state for testing
func (m *Model) State() int {
	return m.state
}
