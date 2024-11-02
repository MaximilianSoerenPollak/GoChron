package chron

import (
	"fmt"
	"io"

	"github.com/davecgh/go-spew/spew"

	tea "github.com/charmbracelet/bubbletea"
)

// type SwitchToAddEntryModelMsg tea.Msg
type switchToAddEntryModel struct{}
type switchToListModel struct{}
type switchToEditModel struct{ entry EntryDB }
type switchToCalendarModel struct{}
type switchToImportModel struct{}

type MainModel struct {
	activeModel tea.Model
	err         error
	dump        io.Writer
}

func InitialModel(dump io.Writer) MainModel {
	return MainModel{
		activeModel: initEntryListModel(dump),
		dump:        dump,
	}
}

func (m MainModel) Init() tea.Cmd {
	return nil
}

func (m *MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	setTerminalSize()
	if m.dump != nil {
		spew.Fdump(m.dump, msg)
	}
	// We only check global keys here.
	// Ctrl+c for example
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit
			// This seems to never execute?
		}
	case switchToAddEntryModel:
		m.activeModel = initMainForm(m.dump)
		return m, m.activeModel.Init()
	case switchToListModel:
		m.activeModel = initEntryListModel(m.dump)
		return m, m.activeModel.Init()
	case switchToEditModel:
		m.activeModel = initChangeEntryForm(m.dump, msg.entry)
		return m, m.activeModel.Init()
	case switchToCalendarModel:
		m.activeModel = initCalendarModel(m.dump)
		return m, m.activeModel.Init()
	case switchToImportModel:
		m.activeModel = initImportModel(m.dump)
		return m, m.activeModel.Init()
	}
	m.activeModel, cmd = m.activeModel.Update(msg)
	return m, cmd
}

func (m MainModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("An error occured. Error: %s", m.err.Error())
	}
	return m.activeModel.View()
}
