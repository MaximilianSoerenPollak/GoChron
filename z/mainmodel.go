package z

import (

	tea "github.com/charmbracelet/bubbletea"
)

type MainModel struct {
	models  []tea.Model
	activeModel tea.Model
	state        int // 0=Normal, 1=Add, 2=Edit, 3=DetailedView, 4=StatsView, 5=CalendarView, 6=ExportView
}

func InitialModel() MainModel {
	entryModel := initEntryListModel()
	return MainModel{
		models: []tea.Model{entryModel},
		activeModel: entryModel,
		state:  0,
	}
}

func (m MainModel) Init() tea.Cmd {
	return nil
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit
		}
		default:
			return m.activeModel.Update(msg)
	}

	return m, nil
}

func (m MainModel) View() string {
	switch m.state {
	case 0:
		return m.activeModel.View()
	}
	return ""
}
