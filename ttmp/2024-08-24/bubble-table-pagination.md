

## 1. Import the necessary packages

Bubble-table is a powerful and flexible library for creating interactive tables in Go using the Bubble Tea framework. This guide will walk you through the process of creating a paginated table with bubble-table.

Import the required packages, including bubble-table.

```go
import (
    "github.com/charmbracelet/bubbletea"
    "github.com/evertras/bubble-table/table"
)
```

## 2. Define your table structure

Create a struct to hold your table model and other necessary data.

```go
type Model struct {
    table    table.Model
    rowCount int
}
```

## 3. Generate table data

Create functions to generate columns and rows for your table.

```go
func genColumns() []table.Column {
    // Define your columns here
}

func genRows(rowCount int) []table.Row {
    // Generate your rows here
}
```

## 4. Initialize the table model

Create a function to initialize your table model with pagination.

```go
func NewModel() Model {
    return Model{
        rowCount: 100,
        table:    table.New(genColumns()).WithRows(genRows(100)).WithPageSize(10).Focused(true),
    }
}
```

## 5. Implement the tea.Model interface

Implement the Init, Update, and View methods for your model.

```go
func (m Model) Init() tea.Cmd {
    return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    m.table, cmd = m.table.Update(msg)
    return m, cmd
}

func (m Model) View() string {
    return m.table.View()
}
```

## 6. Run the program

Create a main function to run your Bubble Tea program.

```go
func main() {
    p := tea.NewProgram(NewModel())
    if err := p.Start(); err != nil {
        log.Fatal(err)
    }
}
```

For a complete example of pagination implementation, you can refer to the following lines in the provided code:


```go
func NewModel() Model {
	const startingRowCount = 105

	m := Model{
		rowCount:            startingRowCount,
		tableDefault:        genTable(3, startingRowCount).WithPageSize(10).Focused(true),
		tableWithRowIndices: genTable(3, startingRowCount).WithPageSize(10).Focused(false),
	}

	m.regenTableRows()

	return m
}

func (m *Model) regenTableRows() {
	m.tableDefault = m.tableDefault.WithRows(genRows(3, m.rowCount))
	m.tableWithRowIndices = m.tableWithRowIndices.WithRows(genRows(3, m.rowCount))
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			cmds = append(cmds, tea.Quit)

		case "a":
			m.tableDefault = m.tableDefault.Focused(true)
			m.tableWithRowIndices = m.tableWithRowIndices.Focused(false)

		case "b":
			m.tableDefault = m.tableDefault.Focused(false)
			m.tableWithRowIndices = m.tableWithRowIndices.Focused(true)

		case "u":
			m.tableDefault = m.tableDefault.WithPageSize(m.tableDefault.PageSize() - 1)
			m.tableWithRowIndices = m.tableWithRowIndices.WithPageSize(m.tableWithRowIndices.PageSize() - 1)

		case "i":
			m.tableDefault = m.tableDefault.WithPageSize(m.tableDefault.PageSize() + 1)
			m.tableWithRowIndices = m.tableWithRowIndices.WithPageSize(m.tableWithRowIndices.PageSize() + 1)

		case "r":
			m.tableDefault = m.tableDefault.WithCurrentPage(rand.Intn(m.tableDefault.MaxPages()) + 1)
			m.tableWithRowIndices = m.tableWithRowIndices.WithCurrentPage(rand.Intn(m.tableWithRowIndices.MaxPages()) + 1)

		case "z":
			if m.rowCount < 10 {
				break
			}

			m.rowCount -= 10
			m.regenTableRows()

		case "x":
			m.rowCount += 10
			m.regenTableRows()
		}
	}

	m.tableDefault, cmd = m.tableDefault.Update(msg)
	cmds = append(cmds, cmd)

	m.tableWithRowIndices, cmd = m.tableWithRowIndices.Update(msg)
	cmds = append(cmds, cmd)

	// Write a custom footer
	start, end := m.tableWithRowIndices.VisibleIndices()
	m.tableWithRowIndices = m.tableWithRowIndices.WithStaticFooter(
		fmt.Sprintf("%d-%d of %d", start+1, end+1, m.tableWithRowIndices.TotalRows()),
	)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	body := strings.Builder{}

	body.WriteString("Table demo with pagination! Press left/right to move pages, or use page up/down, or 'r' to jump to a random page\nPress 'a' for left table, 'b' for right table\nPress 'z' to reduce rows by 10, 'y' to increase rows by 10\nPress 'u' to decrease page size by 1, 'i' to increase page size by 1\nPress q or ctrl+c to quit\n\n")

	pad := lipgloss.NewStyle().Padding(1)

	tables := []string{
		lipgloss.JoinVertical(lipgloss.Center, "A", pad.Render(m.tableDefault.View())),
		lipgloss.JoinVertical(lipgloss.Center, "B", pad.Render(m.tableWithRowIndices.View())),
	}

	body.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, tables...))

	return body.String()
}

func main() {
	p := tea.NewProgram(NewModel())

	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}
```

