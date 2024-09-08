package z

import (
	"database/sql"
	"errors"
	"fmt"
	"io"

	"github.com/davecgh/go-spew/spew"

	tea "github.com/charmbracelet/bubbletea"
)


type MainModel struct {
	activeModel tea.Model
	state       int // 0=Normal, 1=Add, 2=Edit, 3=DetailedView, 4=StatsView, 5=CalendarView, 6=ExportView
	err         error
	dump        io.Writer
}

func InitialModel(dump io.Writer) MainModel {
	return MainModel{
		activeModel: initEntryListModel(dump),
		state:       0,
		dump:        dump,
	}
}

func (m MainModel) Init() tea.Cmd {
	return nil
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.dump != nil {
		spew.Fdump(m.dump, msg)
	}
	// We only check global keys here. 
	// Ctrl+c for example
	if m.activeModel.state == 1 {
		m.Update(msg)	
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit
		case "a":
			oldProject = true
			_, err := database.GetUniqueProjects()
			if errors.Is(err, sql.ErrNoRows) {
				tea.Println("There are currently no projects. Please create one")
				oldProject = false
			}
			m.state = 1
		case "ctrl+a":
			oldProject = false
			m.state = 1
		}
	}
	switch m.state {
	case 0:
		_, ok := m.activeModel.(entryModel)
		if !ok {
			m.activeModel = initEntryListModel(m.dump)
		}
	case 1:
		form, ok := m.activeModel.(taskForm)
		if !ok {
			m.activeModel = initMainForm(m.dump)
			return m.activeModel.Update(msg)
		}
		if ok && form.form.State == 1 {
			m.state = 0
		}
	}
	return m, nil
	// THIS IS THE ISSUE.
	// return m.Update(msg)
}

func (m MainModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("An error occured. Error: %s", m.err.Error())
	}
	return m.activeModel.View()
}
