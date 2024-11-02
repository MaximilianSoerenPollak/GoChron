package chron

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/davecgh/go-spew/spew"
	"github.com/evertras/bubble-table/table"
)

const (
	columnKeyID      = "id"
	columnKeyDate    = "date"
	columnKeyStart   = "start"
	columnKeyFinish  = "finish"
	columnKeyProject = "project"
	columnKeyTask    = "task"
	columnKeyHours   = "hours"
	columnKeyNotes   = "notes"
)

func (lm listModel) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{lm.keys.RowDown, lm.keys.RowUp, lm.keys.RowSelectToggle},
		{lm.keys.PageDown, lm.keys.PageUp, lm.keys.PageFirst, lm.keys.PageLast},
		{lm.keys.Filter, lm.keys.FilterBlur, lm.keys.FilterClear, lm.keys.ScrollRight, lm.keys.ScrollLeft},
	}
}

func (lm listModel) ShortHelp() []key.Binding {
	return []key.Binding{
		lm.keys.RowDown,
		lm.keys.RowUp,
		lm.keys.RowSelectToggle,
		lm.keys.PageDown,
		lm.keys.PageUp,
		lm.keys.Filter,
		lm.keys.FilterBlur,
		lm.keys.FilterClear,
	}
}

type listModel struct {
	table       table.Model
	keys        table.KeyMap
	entries     []EntryDB
	db          *Database
	help        help.Model
	compactView bool
	dump        io.Writer
}

func (m listModel) Init() tea.Cmd {
	return nil
}

func initEntryListModel(dump io.Writer) listModel {
	setTerminalSize()
	database, err := InitDB()
	if err != nil {
		fmt.Printf("%s %+v\n", CharError, err)
		os.Exit(1)
	}
	entries, err := database.GetAllEntriesAsStringWithFormatedTime()
	if err != nil {
		fmt.Printf("%s %+v\n", CharError, err)
		os.Exit(1)
	}
	compactTable := createCompactTable(entries)
	return listModel{
		table:       compactTable,
		db:          database,
		keys:        table.DefaultKeyMap(),
		entries:     entries,
		help:        help.New(),
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
			if m.table.GetFocused() {
				m.table.Focused(false)
			} else {
				m.table.Focused(true)
			}
		case "ctrl+v":
			// Switch from showing 'start / finish' to not showing it.
			if m.compactView {
				expTable := createExpandedTable(m.entries)
				m.table = expTable
				m.compactView = false
			} else {
				compTable := createCompactTable(m.entries)
				m.table = compTable
				m.compactView = true
			}
		case "enter":
			return m, tea.Batch(
				tea.Printf("Let's go to %v!", m.table.HighlightedRow().Data),
			)
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
		case "e":
			entrySelected := m.getRowAsEntryDB()
			return m, func() tea.Msg { return switchToEditModel{entry: entrySelected} }
		case "t":
			return m, func() tea.Msg { return switchToCalendarModel{} }
		case "i":
			return m, func() tea.Msg { return switchToImportModel{} }
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m listModel) View() string {
	helpView := m.help.View(m)
	return baseStyle.Render(m.table.View() + "\n" + helpView)
}

func createExpandedTable(entries []EntryDB) table.Model {
	columns := []table.Column{
		table.NewFlexColumn(columnKeyStart, "Start", 2).WithFiltered(true),
		table.NewFlexColumn(columnKeyFinish, "Finish", 2).WithFiltered(true),
		table.NewFlexColumn(columnKeyDate, "Date", 1).WithFiltered(true),
		table.NewFlexColumn(columnKeyProject, "Project", 2).WithFiltered(true),
		table.NewFlexColumn(columnKeyTask, "Task", 3).WithFiltered(true),
		table.NewFlexColumn(columnKeyHours, "Hours", 1).WithFormatString("%.2f"),
		table.NewFlexColumn(columnKeyNotes, "Notes", 3).WithFiltered(true),
	}
	var rows []table.Row
	for _, v := range entries {
		fmtHrs, err := strconv.ParseFloat(v.Hours, 64)
		if err != nil {
			fmt.Printf(
				"%s could not convert hours into float. Hours: %s Error: %s",
				CharError,
				v.Hours,
				err.Error(),
			)
			os.Exit(1)
		}
		r := table.NewRow(
			table.RowData{
				columnKeyID:      v.ID,
				columnKeyStart:   v.Begin,
				columnKeyFinish:  v.Finish,
				columnKeyDate:    v.Date,
				columnKeyProject: v.Project,
				columnKeyTask:    v.Task,
				columnKeyHours:   fmtHrs,
				columnKeyNotes:   v.Notes,
			})
		rows = append(rows, r)
	}
	keys := table.DefaultKeyMap()
	pageSize := calculatePages(entries)
	t := table.New(columns).
		WithRows(rows).
		Filtered(true).
		WithBaseStyle(baseStyle).
		WithTargetWidth(termWidth).
		Focused(true).
		WithKeyMap(keys).
		WithPageSize(pageSize)
	t.WithMaxTotalWidth(termWidth)
	t.WithMinimumHeight(termHeight / 3)
	return t
}

func createCompactTable(entries []EntryDB) table.Model {
	columns := []table.Column{
		table.NewFlexColumn(columnKeyDate, "Date", 1).WithFiltered(true),
		table.NewFlexColumn(columnKeyProject, "Project", 2).WithFiltered(true),
		table.NewFlexColumn(columnKeyTask, "Task", 3).WithFiltered(true),
		table.NewFlexColumn(columnKeyHours, "Hours", 1).WithFormatString("%.2f"),
		table.NewFlexColumn(columnKeyNotes, "Notes", 3).WithFiltered(true),
	}
	var rows []table.Row
	for _, v := range entries {
		fmtHrs, err := strconv.ParseFloat(v.Hours, 64)
		if err != nil {
			fmt.Printf(
				"%s could not convert hours into float. Hours: %s Error: %s",
				CharError,
				v.Hours,
				err.Error(),
			)
			os.Exit(1)
		}
		r := table.NewRow(
			table.RowData{
				columnKeyID:      v.ID,
				columnKeyDate:    v.Date,
				columnKeyStart:   v.Begin,
				columnKeyProject: v.Project,
				columnKeyTask:    v.Task,
				columnKeyHours:   fmtHrs,
				columnKeyNotes:   v.Notes,
			})
		rows = append(rows, r)
	}
	keys := table.DefaultKeyMap()
	pageSize := calculatePages(entries)
	t := table.New(columns).
		Filtered(true).
		WithRows(rows).
		Focused(true).
		WithBaseStyle(baseStyle).
		WithTargetWidth(termWidth).
		WithKeyMap(keys).
		WithMaxTotalWidth(termWidth - 2).
		WithPageSize(pageSize)

	return t
}

func (m listModel) getRowAsEntryDB() EntryDB {
	selected := m.table.HighlightedRow()
	idConv, err := strconv.Atoi(selected.Data[columnKeyID].(string))
	if err != nil {
		fmt.Printf("%s could not convert id. Error: %s", CharError, err.Error())
		os.Exit(1)
	}
	entry, err := m.db.GetEntryAsString(int64(idConv))
	if err != nil {
		fmt.Printf("%s could not convert id. Error: %s", CharError, err.Error())
		os.Exit(1)
	}
	return *entry
}

func calculatePages(entries []EntryDB) int {
	// TODO: Get it from model do not hardcode it
	styleHeight, _ := calculateTotalStyleSize(baseStyle)
	// Need to calculate max rows we can display.
	// -4 height for filter & help display
	// -2 as buffer
	fotterAndHelpDisplay := 5
	extraBuffer := 3
	totalPadding := styleHeight + fotterAndHelpDisplay + extraBuffer
	maxRows := termHeight - totalPadding
	return maxRows

}
