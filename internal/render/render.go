package render

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tylerkeyes/jot-me/internal/database"
)

// model will store the application state
type model struct {
	db       *database.Service // db connection
	notes    []string          // notes that have been written
	cursor   int               // which note the cursor is pointing at
	selected int               // which note was selected
}

// initialModel defines the initial state.
func InitialModel(db *database.Service, groupName string) model {
	notes, err := (*db).ReadTable(groupName)
	if err != nil {
		fmt.Printf("encountered an issue reading the notes from the group: %v\n", err)
		os.Exit(1)
	}
	return model{
		notes: notes,
	}
}

// Init can perform some initial I/O.
// TODO: may be useful later to read data from the db
func (m model) Init() tea.Cmd {
	return nil
}

// Update is called "when things happen".
// This will change the model state in response to things happening.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// check what key was pressed
		switch msg.String() {
		// exit the program
		case "ctrl+c", "q":
			fmt.Printf("\n\n%v\n", m.notes[m.selected])
			return m, tea.Quit

		// "up" & "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// "down" & "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.notes)-1 {
				m.cursor++
			}

		// "enter" & spacebar select with item to highlight
		case "enter", " ":
			m.selected = m.cursor
		}
	}

	// return the updated model for processing
	return m, nil
}

// View renders the UI to display the application model.
func (m model) View() string {
	// header
	s := "Select a note\n\n"

	// iterate over the choices
	for i, note := range m.notes {
		// set the cursor
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // set the cursor
		}

		// set the selected item
		checked := " "
		if m.selected == i {
			checked = "X"
		}

		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, note)
	}

	// footer
	s += "\nPress q to quit.\n"

	// send to the UI for rendering
	return s
}
