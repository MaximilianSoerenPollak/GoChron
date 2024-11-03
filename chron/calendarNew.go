package chron

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/NimbleMarkets/ntcharts/barchart"
	"github.com/araddon/dateparse"
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
// 3. Can enable filters and then filter depending on that.
// 		-> We should always query all the data as it's fast, and then filter just the output
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
	cf      calendarTimeFrame
}

type calendarTimeFrame struct {
	since time.Time // Filter from
	until time.Time // Filter until
}

func createDefaultCalendarTimeFrame() calendarTimeFrame {
	twoWeeksAgo := time.Now().UTC().Add(-336 * time.Hour)
	return calendarTimeFrame{
		since: twoWeeksAgo,
		until: time.Now().UTC(),
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
		cf:      createDefaultCalendarTimeFrame(),
	}
	cm.chart = createDefaultBarChartModel(database, createDefaultCalendarTimeFrame(), cm.dump)
	cm.chart.Draw()
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
	dateStr := dateRangeBottomBorderStyle.Render(getCurrentDateRangeString(m.cf))
	return lipgloss.JoinVertical(lipgloss.Center, dateStr, m.chart.View())
}

func createDefaultBarChartModel(db *Database, cf calendarTimeFrame, dump io.Writer) barchart.Model {
	dailyHrs, err := db.GetHoursTrackedPerDay(cf)
	if err != nil {
		fmt.Printf("%s could not get hours tracked per day. Error: %s\n", CharError, err.Error())
		os.Exit(1)
	}
	spew.Fdump(dump, dailyHrs)
	bc := barchart.New(termWidth/2, termHeight/2)
	var data []barchart.BarData
	for i, v := range dailyHrs {
		bd := barchart.BarData{
			Label: v.Date[:len(v.Date)-5],
			Values: []barchart.BarValue{
				{
					Name:  fmt.Sprintf("Item%d", i),
					Value: v.Hours.InexactFloat64(),
					Style: barchartDefaultStyle,
				},
			},
		}
		data = append(data, bd)

	}
	bc.PushAll(data)
	bc.AutoMaxValue = true
	bc.AutoBarWidth = true 

	// Hardcoded for now I guess
	bc.SetBarGap(1)
	// bc.Resize(termHeight - extraBuffer - 5)
	return bc

}

// Calcualte all data available as 'total' and 'perProject' per day.
func calculatePerDayData(entries []Entry) ([]barchart.BarData, []barchart.BarData) {

	// Calculating all hours and hours per day
	var totalHours decimal.Decimal

	// TODO: Is there a better way than this?
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
			Label: date,
			Values: []barchart.BarValue{
				{
					Name:  "totalHours",
					Value: hoursMap["totalHours"].InexactFloat64(),
					Style: barchartDefaultStyle,
				},
			},
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

//          ╭─────────────────────────────────────────────────────────╮
//          │                  STYLING RELATED STUFF                  │
//          ╰─────────────────────────────────────────────────────────╯

func getCurrentDateRangeString(cf calendarTimeFrame) string {
	parsedDateSince, err := dateparse.ParseAny(cf.since.String())
	if err != nil {
		fmt.Printf("%s error parsing calendar time range 'since'. CalendarTimeFrame: %+v\n, Error: %s",
			CharError,
			cf,
			err.Error(),
		)
		return "Gotten Error parsing time. Check logs"
	}
	sinceFormated := parsedDateSince.Format("01-02-2006")
	parsedDateUntil, err := dateparse.ParseAny(cf.until.String())
	if err != nil {
		fmt.Printf("%s error parsing calendar time range 'until'. CalendarTimeFrame: %+v\n, Error: %s",
			CharError,
			cf,
			err.Error(),
		)
		return "Gotten Error parsing time. Check logs"
	}
	untilFormated := parsedDateUntil.Format("01-02-2006")
	dateRangeStr := fmt.Sprintf("Date range\n\n %s --- %s", sinceFormated, untilFormated)
	return dateRangeStr
}

// func (m listModel) setSize(dateRangeStr string) {
// 	widthDR, heightDR := lipgloss.Size(dateRangeStr)
// 	m.chart.
// }
