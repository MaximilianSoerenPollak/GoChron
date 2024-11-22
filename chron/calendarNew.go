package chron

import (
	"fmt"
	"io"
	"os"
	"strconv"
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
	CurrWeek int = iota
	LastWeek
	CurrMonth
	LastMonth
	CurrQurater
	CurrYear
)

type calendarModel struct {
	barChart barChartResult
	keys     help.KeyMap
	debug    map[string]map[string]decimal.Decimal
	entries  []EntryDB
	db       *Database
	dump     io.Writer
	cf       calendarTimeFrame
	ctf      int //current time window
	err      error
}

type barChartResult struct {
	chart barchart.Model
	data  []barchart.BarData
}

type calendarTimeFrame struct {
	since string // Filter from
	until string // Filter until
}

func createDefaultCalendarTimeFrame() calendarTimeFrame {
	since := time.Now().UTC().Add(-1500 * time.Hour).Format(time.DateOnly)
	until := time.Now().UTC().Format(time.DateOnly)
	return calendarTimeFrame{
		since: since,
		until: until,
	}
}

// Gets the first day of the Month
func getMonthStart(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location()).UTC()
}

// Gets the last day of the current Month
func getMonthEnd(date time.Time) time.Time {
	nextMonth := getMonthStart(date).AddDate(0, 1, 0)
	return nextMonth.Add(-1 * time.Second)
}

func getWeekStart(date time.Time) time.Time {
	sinceMon := int(date.Weekday()) + 1
	return date.AddDate(0, 0, -sinceMon)
}

func getWeekEnd(date time.Time) time.Time {
	daysUntilMonday := 7 - int(date.Weekday())
	return date.AddDate(0, 0, daysUntilMonday)
}

// Returns start & end dates for the current week
func getWeekDates(date time.Time) (time.Time, time.Time) {
	start := getWeekStart(date)
	end := getWeekEnd(date)
	return start, end
}
func getCurrQuaterDates(date time.Time) (time.Time, time.Time) {
	var quaterStart time.Time
	var quaterEnd time.Time
	m := date.Month()
	switch {
	case m >= 1 && m <= 3:
		quaterStart = time.Date(date.Year(), 1, 1, 0, 0, 0, 0, date.Location()).UTC()
		quaterEnd = time.Date(date.Year(), 3, 31, 0, 0, 0, 0, date.Location()).UTC()
	case m >= 4 && m <= 6:
		quaterStart = time.Date(date.Year(), 4, 1, 0, 0, 0, 0, date.Location()).UTC()
		quaterEnd = time.Date(date.Year(), 6, 30, 0, 0, 0, 0, date.Location()).UTC()
	case m >= 7 && m <= 9:
		quaterStart = time.Date(date.Year(), 7, 1, 0, 0, 0, 0, date.Location()).UTC()
		quaterEnd = time.Date(date.Year(), 9, 30, 0, 0, 0, 0, date.Location()).UTC()
	case m >= 10 && m <= 12:
		quaterStart = time.Date(date.Year(), 10, 1, 0, 0, 0, 0, date.Location()).UTC()
		quaterEnd = time.Date(date.Year(), 12, 31, 0, 0, 0, 0, date.Location()).UTC()
	}
	return quaterStart, quaterEnd
}

func createCalendarTimeFrame(ctf int) calendarTimeFrame {
	var since time.Time
	var until time.Time
	now := time.Now().UTC()
	switch ctf {
	case CurrWeek:
		since, until = getWeekDates(now)
	case LastWeek:
		lastWeek := getWeekStart(now.AddDate(0, 0, -1))
		since, until = getWeekDates(lastWeek)
	case CurrMonth:
		since = getMonthStart(now)
		until = getMonthEnd(now)
	case LastMonth:
		t := now.AddDate(0, -1, 0)
		since = getMonthStart(t)
		until = getMonthEnd(t)
	case CurrQurater:
		since, until = getCurrQuaterDates(now)
	case CurrYear:
		since = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location()).UTC()
		until = now
	}
	return calendarTimeFrame{
		since: since.Format("2006-01-02 15:03"),
		until: until.Format("2006-01-02 15:03"),
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
		dump:    dump,
		cf:      createDefaultCalendarTimeFrame(),
	}

	cm.ctf = CurrMonth
	cm.cf = createDefaultCalendarTimeFrame()
	// cm.cf = createCalendarTimeFrame(CurrWeek)
	cm.barChart, cm.err = updateBarChart(database, cm.cf, "months")
	cm.barChart.chart.Draw()
	return cm
}

func updateBarChart(db *Database, cf calendarTimeFrame, grouping string) (barChartResult, error) {
	var queriedEntries []GroupedEntries
	var err error
	switch grouping {
	// Quaterly view
	case "weeks":
		queriedEntries, err = db.GetHoursTrackerPerWeek(cf)
		if err != nil {
			errMsg := fmt.Errorf("Encountered an error while getting entries grouped by week. Error: %w", err)
			return barChartResult{barchart.Model{}, nil}, errMsg
		}
	// Yearly view
	case "months":
		queriedEntries, err = db.GetHoursTrackerPerMonth(cf)
		if err != nil {
			errMsg := fmt.Errorf("Encountered an error while getting entries grouped by month. Error: %w", err)
			return barChartResult{barchart.Model{}, nil}, errMsg
		}
	default:
		queriedEntries, err = db.GetHoursTrackedPerDay(cf)
		if err != nil {
			errMsg := fmt.Errorf("Encountered an error while getting entries grouped by day. Error: %w", err)
			return barChartResult{barchart.Model{}, nil}, errMsg
		}

	}
	if queriedEntries == nil {
		// TODO: Need to make this into a new information box not like this.
		errMsg := fmt.Errorf("error, no data for selected date range. Please select another date range")
		return barChartResult{barchart.Model{}, nil}, errMsg
	}
	data := createDayGroupedBarData(queriedEntries, grouping)
	return barChartResult{
		chart: createBarChartModel(data),
		data:  data,
	}, nil

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

		case "1":
			m.ctf = CurrWeek
			m.cf = createCalendarTimeFrame(CurrWeek)
			m.barChart, m.err = updateBarChart(m.db, m.cf, "days")
		case "2":
			m.ctf = LastWeek
			m.cf = createCalendarTimeFrame(LastWeek)
			m.barChart, m.err = updateBarChart(m.db, m.cf, "days")
		case "ctrl+m":
			m.ctf = CurrMonth
			m.cf = createCalendarTimeFrame(CurrMonth)
			m.barChart, m.err = updateBarChart(m.db, m.cf, "days")
		case "ctrl+l":
			m.ctf = LastMonth
			m.cf = createCalendarTimeFrame(LastMonth)
			m.barChart, m.err = updateBarChart(m.db, m.cf, "days")
		case "ctrl+q":
			m.ctf = CurrQurater
			m.cf = createCalendarTimeFrame(CurrQurater)
			m.barChart, m.err = updateBarChart(m.db, m.cf, "weeks")
		case "ctrl+y":
			m.ctf = CurrYear
			m.cf = createCalendarTimeFrame(CurrYear)
			m.barChart, m.err = updateBarChart(m.db, m.cf, "months")
		}
	}
	return m, cmd
}

// func (m calendarModel) View() string {
// 	dateStr := m.generateDateRangeStr()
// 	minMaxStr := m.generateMinMaxAvgStr()
// 	keyMapStr := generateKeyMapStr(minMaxStr, m.barChart.chart.Height(), m.ctf)
// 	sepStr := strings.Repeat("=", lipgloss.Width(keyMapStr))
// 	keyMapSepStr := lipgloss.JoinVertical(lipgloss.Center, lipgloss.NewStyle().Margin(1).Render(sepStr), keyMapStr)
// 	minMaxKeyMapStr := lipgloss.JoinVertical(lipgloss.Left, minMaxStr, keyMapSepStr)
// 	minMaxKeymapStrPadded := lipgloss.NewStyle().MarginLeft(2).Render(minMaxKeyMapStr)
// 	if m.err != nil {
// 		styleHeight, styleWidth := lipgloss.Size(minMaxKeymapStrPadded)
// 		emptyStyle := lipgloss.NewStyle().Width(termWidth-styleWidth).Height(termHeight-styleHeight)
// 		return placeOverlay(emptyStyle.Render(""), m.err.Error(), styleWidth, styleHeight, styleWidth)
// 	}
// 	barChartView := m.barChart.chart.View()
// 	minMaxChartStr := lipgloss.JoinHorizontal(lipgloss.Top, barChartView, minMaxKeymapStrPadded)
// 	finalStr := lipgloss.JoinVertical(lipgloss.Center, dateStr, minMaxChartStr)
// 	return finalStr
// }

func (m calendarModel) View() string {
	// Generate all the common layout elements
	dateStr := m.generateDateRangeStr()
	minMaxStr := m.generateMinMaxAvgStr()
	keyMapStr := generateKeyMapStr(minMaxStr, m.barChart.chart.Height(), m.ctf)
	sepStr := strings.Repeat("=", lipgloss.Width(keyMapStr))
	keyMapSepStr := lipgloss.JoinVertical(lipgloss.Center, lipgloss.NewStyle().Margin(1).Render(sepStr), keyMapStr)
	minMaxKeyMapStr := lipgloss.JoinVertical(lipgloss.Left, minMaxStr, keyMapSepStr)
	minMaxKeymapStrPadded := lipgloss.NewStyle().MarginLeft(2).Render(minMaxKeyMapStr)

	if m.err != nil {
		// Create a more compact and elegant error box
		chartHeight := m.barChart.chart.Height()
		availableWidth := termWidth - lipgloss.Width(minMaxKeymapStrPadded) - 10
		messageWidth := availableWidth / 3 // Make the error box smaller

		// Simplified error box style
		errorStyle := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF7777")).
			Padding(1, 2).
			Width(messageWidth).
			Align(lipgloss.Center).
			Background(lipgloss.Color("#2A1215")).
			Foreground(lipgloss.Color("#FF8888")).
			Bold(true)

		// Simple container just for height matching
		containerStyle := lipgloss.NewStyle().
			Width(availableWidth).
			Height(chartHeight).
			Align(lipgloss.Center)

		// Render error message
		errorMsg := "⚠️  " + m.err.Error()
		errorBox := errorStyle.Render(errorMsg)

		// Center in container
		chartArea := containerStyle.Render(errorBox)

		// Join with the rest of the layout
		errorWithKeymap := lipgloss.JoinHorizontal(lipgloss.Top, chartArea, minMaxKeymapStrPadded)
		return lipgloss.JoinVertical(lipgloss.Center, dateStr, errorWithKeymap)
	}

	// Normal rendering without error
	barChartView := m.barChart.chart.View()
	minMaxChartStr := lipgloss.JoinHorizontal(lipgloss.Top, barChartView, minMaxKeymapStrPadded)
	finalStr := lipgloss.JoinVertical(lipgloss.Center, dateStr, minMaxChartStr)
	return finalStr
}

func createDayGroupedBarData(groupedEntries []GroupedEntries, labelStyle string) []barchart.BarData {
	var data []barchart.BarData
	for i, v := range groupedEntries {
		var labelStr string
		// fmt.Printf("LabelStyle: %s || v.Date: %s. || Selected v.date1: %s || Selected2: %s\n", labelStr, v.Date, v.Date[:2], v.Date[3:5])
		switch labelStyle {
		case "days":
			labelStr = v.Date[:2]
		case "weeks":
			labelStr = fmt.Sprintf("CW:  %s", v.Date)
		case "months":
			m, err := strconv.Atoi(v.Date)
			if err != nil {
				fmt.Printf("Errored converting months: %s", err.Error())
				os.Exit(1)
			}
			labelStr = time.Month(m).String()
		}
		bd := barchart.BarData{
			Label: labelStr,
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

func createBarChartModel(data []barchart.BarData) barchart.Model {
	// TODO: Get a better way to determin the width of this thing
	bc := barchart.New(termWidth/2, termHeight/2)
	bc.PushAll(data)
	bc.AutoMaxValue = true
	bc.ShowAxis()
	bc.AutoBarWidth = true
	bc.Draw()

	return bc

}

//          ╭─────────────────────────────────────────────────────────╮
//          │                  STYLING RELATED STUFF                  │
//          ╰─────────────────────────────────────────────────────────╯

func (m calendarModel) generateMinMaxAvgStr() string {
	data := m.barChart.data
	maxValue := m.barChart.chart.MaxValue()
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
func generateKeyMapStr(minMaxStr string, tableHeight int, cft int) string {
	rows := [][]string{
		{" ctrl+1 ", " Current Week "},
		{" ctrl+2 ", " Last Week "},
		{" ctrl+m ", " Current Month "},
		{" ctrl+l ", " Last Month "},
		{" ctrl+q ", " Current Quarter "},
		{" ctrl+y ", " Current Year "},
	}
	maxHeight := lipgloss.Height(minMaxStr) + tableHeight
	lg := lgTable.New().
		Headers(" KEYMAP ", " QUERIED DATES ").
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return keyMapShortCutTableHeaderStyle
			}
			if row == cft+1 {
				return keyMapShortCutTableHighlightStyle
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
