package chron

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/NimbleMarkets/ntcharts/barchart"
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/davecgh/go-spew/spew"
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
// ===============================================================
// ===============================================================
// ===============================================================
// What to do
// ===============================================================
// Each 'day' will be one BarData struct point.
// Default 2 weeks? How to split it up?
// Let's pretend we start at Monday
//
//
//				21.10.2024 - 27.10.2024
//	--------------------------------------------
//  14h|
//  12h|
//  10h|		   ##
//  8h | ##        ##
//  6h | ##		   ##					   ##
//  4h | ##   ##   ##				 ##	   ##
//  2h | ##   ##   ##    ##          ##    ##
//	  -----------------------------------------
//       Mon  Tue  Wed   Thu   Fri   Sat   Sun
// ===============================================================
// ===============================================================

type calendarModel struct {
	chart   barchart.Model
	keys    help.KeyMap
	debug   map[string]map[string]decimal.Decimal
	entries []EntryDB
	db      *Database
	help    help.Model
	dump    io.Writer
}

type calendarFilter struct {
	since         time.Time // Filter from
	until         time.Time // Filter until
	period        string    // Filter from now back x time. -> '1w', 2w, 1m, quater, year ???
	projectFilter bool      // Color in projects or just display time per day

}

func createDefaultCalendarFilter() calendarFilter {
	twoWeeksAgo := time.Now().UTC().Add(-336 * time.Hour)
	return calendarFilter{
		since:         twoWeeksAgo,
		until:         time.Now().UTC(),
		period:        "2w",
		projectFilter: false,
	}
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
	cm := calendarModel{
		keys:    createChronDefaultKeyMap(),
		entries: entries,
		db:      database,
		help:    help.New(),
		dump:    dump,
	}
	cm.createDefaultCalendarChart()

	return cm
}

func (m calendarModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	if m.dump != nil {
		spew.Fdump(m.dump, msg)
	}
	return m, cmd

}

func (m calendarModel) View() string {
	b, _ := json.MarshalIndent(m.debug, "", "  ")
	return fmt.Sprintln(string(b))
}

// Function errors
func (m *calendarModel) createDefaultCalendarChart() {
	// Getting all entries.
	entries, err := m.db.GetAllEntries()
	if err != nil {
		fmt.Printf("%s %+v\n", CharError, err)
		os.Exit(1)
	}
	// Filtering entries 
	defaultFilter := createDefaultCalendarFilter()
	filteredEntries := filterEntries(defaultFilter, entries)
	// TODO: We need to filter the entries before this
	barDataDayTotal, barDataDayProject := calculatePerDayData(filteredEntries)
	//

}

func filterEntries(filter calendarFilter, entries []Entry) []Entry {
	

	
}

// Calcualte all data available as 'total' and 'perProject' per day.
func calculatePerDayData(entries []Entry) ([]barchart.BarData, []barchart.BarData) {
	// Calculating all hours and hours per day
	var totalHours decimal.Decimal
	// map[string]map[string]decimal.Decimal
	// { "20.01.2024": { "Project A": 20, "Project B": 10 },}
	var hoursPerDay = make(map[string]map[string]decimal.Decimal)
	// Question: How do we filter???
	for _, v := range entries {
		// Hours on that day.
		// date = map
		hoursPerDay[v.Date] = make(map[string]decimal.Decimal)
		hoursPerDay[v.Date][v.Project] = hoursPerDay[v.Date][v.Project].Add(v.Hours)
		hoursPerDay[v.Date]["totalHours"] = hoursPerDay[v.Date]["totalHours"].Add(v.Hours)
		totalHours.Add(v.Hours)
	}
	var barDataDayTotal []barchart.BarData
	var barDataDayProject []barchart.BarData
	for date, hoursMap := range hoursPerDay {
		bcTotal := barchart.BarData{
			Label:  date,
			Values: []barchart.BarValue{{Name: "totalHours", Value: hoursMap["totalHours"].InexactFloat64(), Style: barchartDefaultStyle}},
		}
		barDataDayTotal = append(barDataDayTotal, bcTotal)

		// Needed for when we want to show the Projects individually
		bcProject := barchart.BarData{
			Label: date,
		}
		var bcProjectValues []barchart.BarValue
		for k, v := range hoursMap {
			bv := barchart.BarValue{
				Name:  k,
				Value: v.InexactFloat64(),
				Style: barchartDefaultStyle,
			}
			bcProjectValues = append(bcProjectValues, bv)
		}
		bcProject.Values = bcProjectValues
		barDataDayProject = append(barDataDayProject, bcProject)
	}
	return barDataDayTotal, barDataDayProject
}

func calculatePerWeekData(hoursPerDay map[string]map[string]decimal.Decimal) ([]barchart.BarData, []barchart.BarData) {
	return nil, nil
}
