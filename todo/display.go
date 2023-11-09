package todo

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	todos    []Todo
	cursor   int
	selected map[int]struct{}
}

func InitialModel(todos []Todo) model {
	return model{
		todos: todos,
	}
}

// Init is the first function that will be called. It returns an optional
// initial command. To not perform an initial command return nil.
func (m model) Init() tea.Cmd {
	return nil
}

// Update is called when a message is received. Use it to inspect messages
// and, in response, update the model and/or send a command.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.todos)-1 {
				m.cursor--
			}
		case " ":
			_, found := m.selected[m.cursor]
			if found {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}

	return m, nil
}

// View renders the program's UI, which is just a string. The view is
// rendered after every Update.
func (m model) View() string {
	strTodos := make([]string, len(m.todos))
	for i, todo := range m.todos {
		strTodos[i] = listItemStyle.Render(fmt.Sprintf("%s: %d;%d", todo.filePath, todo.startPoint.Row, todo.startPoint.Column))
	}
	list := listStyle.Render(strTodos...)
	return list
}

// Styles
var (
	listStyle     = lipgloss.NewStyle()
	listItemStyle = lipgloss.NewStyle()
)
