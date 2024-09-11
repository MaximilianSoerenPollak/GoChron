package z

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/davecgh/go-spew/spew"
)

type listModel struct {
	table       table.Model
	db          *Database
	compactView bool
	dump        io.Writer
}

func (m listModel) Init() tea.Cmd {
	return nil
}

func initEntryListModel(dump io.Writer) listModel {
	database, err := InitDB()
	if err != nil {
		fmt.Printf("%s %+v\n", CharError, err)
		os.Exit(1)
	}
	compactTable := createCompactTable(*database)
	return listModel{
		table:       compactTable,
		db:          database,
		compactView: true,
		dump:        dump,
	}
}

func (m listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.dump != nil {
		spew.Fdump(m.dump, fmt.Sprintf("EntryModel: %s", msg))
	}
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "enter":
			return m, tea.Batch(
				tea.Printf("Let's go to %s!", m.table.SelectedRow()[1]),
			)
		case "ctrl+v":
			// Switch from showing 'start / finish' to not showing it.
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
		case "a":
			oldProject = true
			_, err := database.GetUniqueProjects()
			if errors.Is(err, sql.ErrNoRows) {
				tea.Println("There are currently no projects. Please create one")
				oldProject = false
			}
			return m, func() tea.Msg { return switchToAddEntryModel{} }
		case "ctrl+a":
			oldProject = false
			return m, func() tea.Msg { return switchToAddEntryModel{} }
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m listModel) View() string {
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