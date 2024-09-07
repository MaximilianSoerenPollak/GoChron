package z

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type entryModel struct {
	table       table.Model
	db          *Database
	compactView bool
}

func (m entryModel) Init() tea.Cmd {
	return nil
}

func initEntryListModel() entryModel {
	database, err := InitDB()
	if err != nil {
		fmt.Printf("%s %+v\n", CharError, err)
		os.Exit(1)
	}
	compactTable := createCompactTable(*database)
	return entryModel{
		table:       compactTable,
		db:          database,
		compactView: true,
	}
}

func (m entryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			return m, tea.Batch(
				tea.Printf("Let's go to %s!", m.table.SelectedRow()[1]),
			)
		case "ctrl+v":
			if m.compactView {
				expTable := createExpandedTable(*m.db)
				m.table = expTable
				m.table.UpdateViewport()
				m.compactView = false
			} else {
				compTable := createCompactTable(*m.db)
				m.table = compTable
				m.table.UpdateViewport()
				m.compactView = true
			}

		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m entryModel) View() string {
	return baseStyle.Render(m.table.View()) + "\n  " + m.table.HelpView() + "\n"
}

func createExpandedTable(db Database) table.Model {
	entries, err := db.GetAllEntriesAsString()
	if err != nil {
		fmt.Printf("%s %+v\n", CharError, err)
		os.Exit(1)
	}
	columns := []table.Column{
		{Title: "ID", Width: 5},
		{Title: "Date", Width: 10},
		{Title: "Start", Width: 20},
		{Title: "Finish", Width: 20},
		{Title: "Project", Width: 30},
		{Title: "Task", Width: 40},
		{Title: "Hours", Width: 20},
		{Title: "Notes", Width: 40},
	}
	var rows []table.Row
	for _, v := range entries {
		err := v.FormatTimes()
		if err != nil {
			fmt.Printf("Encountered an error formatting times")
			fmt.Printf("%s %+v\n", CharError, err)
			os.Exit(1)
		}
		r := table.Row{
			v.ID, v.Date, v.Begin, v.Finish, v.Project, v.Task, v.Hours, v.Notes}
		rows = append(rows, r)
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	s := table.DefaultStyles()
	s.Header = tableHeaderStyle
	s.Selected = selectedEntryStyle
	t.SetStyles(s)
	return t
}

func createCompactTable(db Database) table.Model {
	entries, err := db.GetAllEntriesAsString()
	if err != nil {
		fmt.Printf("%s %+v\n", CharError, err)
		os.Exit(1)
	}
	columns := []table.Column{
		{Title: "Date", Width: 20},
		{Title: "Project", Width: 30},
		{Title: "Task", Width: 40},
		{Title: "Hours", Width: 20},
		{Title: "Notes", Width: 40},
	}
	var rows []table.Row
	for _, v := range entries {
		r := table.Row{
			v.Date, v.Project, v.Task, v.Hours, v.Notes}
		rows = append(rows, r)
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	s := table.DefaultStyles()
	s.Header = tableHeaderStyle
	s.Selected = selectedEntryStyle
	t.SetStyles(s)
	return t
}
