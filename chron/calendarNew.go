package chron

import (
	"fmt"
	"io"
	"os"

	"github.com/NimbleMarkets/ntcharts/barchart"
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shopspring/decimal"
)

// QUESTIONS:
// 1. We need to get make sure we have a default 'time period' we display.
// 2. We need a way to get all worked hours per day.
// 3. We need to be able to filter the hours displayed by project if wanted
// 4. Need to be able to widen the displayed times beyond 2 weeks. -> Month -> 1-2-3 Months? quarters? year?
// 5. If we have all projects -> color each one in differently.
// 6. How to organize stuff?
// ANSWERS:
// 1. Default Period: last 2 weeks
// 2. Probably best to have a 'calc hours' function or so
// 3. Can enable filters and then filter depending on that. -> We should always query all the data as it's fast, and then filter just the output
// 4. Recalculate / redraw depending on buttonpress
// 5. Can just pick 4 colors -> randomly color them in
// 6. We should probably make it a BubbleTea model I think

type calendarModel struct {
	chart   barchart.Model
	keys    help.KeyMap
	entries []EntryDB
	db      *Database
	help    help.Model
	dump    io.Writer
}

func (m calendarModel) Init() tea.Cmd {
	return nil
}

func initCalendarModel(dump io.Writer) calendarModel {
	database, err := InitDB()
	if err != nil {
		fmt.Printf("%s %+v\n", CharError, err)
		os.Exit(1)
	}
	entries, err := database.GetAllEntriesAsString()
	if err != nil {
		fmt.Printf("%s %+v\n", CharError, err)
		os.Exit(1)
	}
	return calendarModel{
		chart:   defaultCalendar(),
		keys:    createChronDefaultKeyMap(),
		entries: entries,
		db:      database,
		help:    help.New(),
		dump:    dump,
	}
}
