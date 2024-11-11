package chron

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/NimbleMarkets/ntcharts/barchart"
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
	barData []barchart.BarData
	keys    help.KeyMap
	debug   map[string]map[string]decimal.Decimal
	entries []EntryDB
	db      *Database
	help    help.Model
	dump    io.Writer
	cf      calendarTimeFrame
}

type calendarTimeFrame struct {
	since string // Filter from
	until string // Filter until
}

func createDefaultCalendarTimeFrame() calendarTimeFrame {
	since := time.Now().UTC().Add(-336 * time.Hour).Format("2006-01-02")
	until := time.Now().UTC().Format("2006-01-02")
	return calendarTimeFrame{
		since: since,
		until: until,
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
	data := createBarData(database, createDefaultCalendarTimeFrame())
	cm.chart = createDefaultBarChartModel(data)
	cm.barData = data
	// cm.chart.Canvas.SetString()
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
	dateStr := m.generateDateRangeStr()
	mmaStr := m.generateMinMaxAvgStr()
	mmaChart := lipgloss.JoinHorizontal(lipgloss.Top, m.chart.View(), mmaStr)
	return lipgloss.JoinVertical(lipgloss.Center, dateStr, mmaChart)
}
func createBarData(db *Database, cf calendarTimeFrame) []barchart.BarData {
	dailyHrs, err := db.GetHoursTrackedPerDay(cf)
	if err != nil {
		fmt.Printf("%s could not get hours tracked per day. Error: %s\n", CharError, err.Error())
		os.Exit(1)
	}
	var data []barchart.BarData
	if dailyHrs == nil {
		// TODO: Need to make this into a new information box not like this.
		fmt.Printf("error, no data for selected date range.")
		os.Exit(1)
	}
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
	return data

}
func createDefaultBarChartModel(data []barchart.BarData) barchart.Model {
	bc := barchart.New(termWidth/2, termHeight/2)
	bc.PushAll(data)
	bc.AutoMaxValue = true
	bc.AutoBarWidth = true

	// Hardcoded for now I guess
	bc.SetBarGap(1)
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

func (m calendarModel) generateMinMaxAvgStr() string {
	data := m.barData
	maxValue := m.chart.MaxValue()
	var minValue  float64
	var total float64 
	for i, dp := range data {
		for _, x := range dp.Values {
			total += x.Value
			if i == 0 {
				minValue = x.Value 
			}
			if x.Value < minValue {
				minValue = x.Value
			}
		}	
	}
	avg := total / float64(len(data))
	mmaStr := fmt.Sprintf("Max: %.2fh\nMin: %.2fh\nAvg: %.2fh",maxValue, minValue, avg)
	strHeight := lipgloss.Height(mmaStr)
	vertialStr :=  lipgloss.PlaceVertical(strHeight, 0.5, mmaStr)
	return dateRangeBottomBorderStyle.Render(vertialStr)
}

func (m calendarModel) generateDateRangeStr() string {
	dateRangeStr := m.cf.since + "  -  " + m.cf.until
	dateStr := lipgloss.JoinVertical(lipgloss.Center, "Queried Dates\n", dateRangeStr)
	return  dateRangeBottomBorderStyle.Render(dateStr)
}

// func (m calendarModel) generateHourAxis() string {
// 	maxVal := m.chart.MaxValue()
// 	return  ""
// }

// func (m listModel) setSize(dateRangeStr string) {
// 	widthDR, heightDR := lipgloss.Size(dateRangeStr)
// 	m.chart.
// }
