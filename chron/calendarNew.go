package chron

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/NimbleMarkets/ntcharts/barchart"
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	lgTable "github.com/charmbracelet/lipgloss/table"
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

const (
	LastWeek int = iota
	LastTwoWeeks
	ThisMonth
	LastMonth
)

type calendarModel struct {
	chart   barchart.Model
	barData []barchart.BarData
	keys    help.KeyMap
	debug   map[string]map[string]decimal.Decimal
	entries []EntryDB
	db      *Database
	dump    io.Writer
	cf      calendarTimeFrame
	ctf     string //current time window
}

type calendarTimeFrame struct {
	since string // Filter from
	until string // Filter until
}

func createDefaultCalendarTimeFrame() calendarTimeFrame {
	since := time.Now().UTC().Add(-1336 * time.Hour).Format("2006-01-02")
	until := time.Now().UTC().Format(time.DateOnly)
	return calendarTimeFrame{
		since: since,
		until: until,
	}
}

func createCalendarTimeFrame(ctf int){
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
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "ctrl+v":
			// Switch from showing 'start / finish' to not showing it.
			// 	case "enter":
			// 		return m, tea.Batch(
			// 			tea.Printf("Let's go to %v!", m.table.HighlightedRow().Data),
			// 		)
			// 	}
		}
	}
	// }
	return m, cmd

}

func (m calendarModel) View() string {
	dateStr := m.generateDateRangeStr()
	mmaStr := m.generateMinMaxAvgStr()
	keyMapStr := generateKeyMapStr(mmaStr, m.chart.Height())
	sepStr := strings.Repeat("=", lipgloss.Width(keyMapStr))
	keyMapSepStr := lipgloss.JoinVertical(lipgloss.Center, lipgloss.NewStyle().Margin(1).Render(sepStr), keyMapStr)
	mmaKmpStr := lipgloss.JoinVertical(lipgloss.Left, mmaStr, keyMapSepStr)
	mmaKmpStrPad := lipgloss.NewStyle().MarginLeft(2).Render(mmaKmpStr)
	mmaChart := lipgloss.JoinHorizontal(lipgloss.Top, m.chart.View(), mmaKmpStrPad)
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
	// TODO: Get a better way to determin the width of this thing
	bc := barchart.New(termWidth/2, termHeight/2)
	bc.PushAll(data)
	bc.AutoMaxValue = true
	bc.AutoBarWidth = true

	return bc

}

//          ╭─────────────────────────────────────────────────────────╮
//          │                  STYLING RELATED STUFF                  │
//          ╰─────────────────────────────────────────────────────────╯

func (m calendarModel) generateMinMaxAvgStr() string {
	data := m.barData
	maxValue := m.chart.MaxValue()
	var minValue float64
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
	mmaStr := fmt.Sprintf("Max: %.2fh\nMin: %.2fh\nAvg: %.2fh", maxValue, minValue, avg)
	strHeight := lipgloss.Height(mmaStr)
	verticalStr := lipgloss.PlaceVertical(strHeight, 0.5, mmaStr)
	return dateRangeBottomBorderStyle.Render(verticalStr)
}

func (m calendarModel) generateDateRangeStr() string {
	dateRangeStr := m.cf.since + "  -  " + m.cf.until
	dateStr := lipgloss.JoinVertical(lipgloss.Center, "Queried Dates\n", dateRangeStr)
	return dateRangeBottomBorderStyle.Render(dateStr)
}

// Unsure where I should put these keymaps, for now they are hardcoded
func generateKeyMapStr(minMaxStr string, tableHeight int) string {
	rows := [][]string{
		{" ctrl+1 ", " Last Week "},
		{" ctrl+2 ", " Last 2 Weeks "},
		{" ctrl+m ", " Current Month "},
		{" ctrl+l ", " Last Month "},
	}
	maxHeight := lipgloss.Height(minMaxStr) + tableHeight
	lg := lgTable.New().
		Headers(" KEYMAP ", " QUERIED DATES ").
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return keyMapShortCutTableHeaderStyle
			}
			return lipgloss.NewStyle()
		},
		).
		BorderRow(true).
		BorderColumn(true).
		Height(maxHeight)
	// keyMapStr := "Quick select queried dates\n\n"

	return lg.Render()
	// return keyMapStr + fmt.Sprint(lg)
}
