package chron

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/gookit/color"
	"github.com/shopspring/decimal"
)

type Entry struct {
	ID      int64           `json:"-"`
	Date    string          `json:"date,omitempty"`
	Begin   time.Time       `json:"begin,omitempty"`
	Finish  time.Time       `json:"finish,omitempty"`
	Project string          `json:"project,omitempty"`
	Hours   decimal.Decimal `json:"hours,omitempty"`
	Task    string          `json:"task,omitempty"`
	Notes   string          `json:"notes,omitempty"`
	Running bool            `json:"-"`
}

type EntryDB struct {
	ID      string
	Date    string
	Begin   string
	Finish  string
	Project string
	Hours   string
	Task    string
	Notes   string
	Running bool
}

func (edb *EntryDB) FormatTimes() error {
	parsedFinish, err := dateparse.ParseAny(edb.Finish)
	if err != nil {
		fmt.Printf("%s could not convert finish time to standard. Error: %s\n", CharError, err.Error())
		return err
	}
	parsedBegining, err := dateparse.ParseAny(edb.Begin)
	if err != nil {
		fmt.Printf("%s could not convert finish time to standard. Error: %s\n", CharError, err.Error())
		return err
	}
	edb.Finish = parsedFinish.Format("2006-01-02 15:04:05")
	edb.Begin = parsedBegining.Format("2006-01-02 15:04:05")
	return nil
}

func (edb *EntryDB) ConvertToEntry() (*Entry, error) {
	entry := Entry{}
	idParsed, err := strconv.Atoi(edb.ID)
	if err != nil {
		return nil, err
	}
	beginParsed, err := dateparse.ParseAny(edb.Begin)
	if err != nil {
		return nil, err
	}
	finishParsed, err := dateparse.ParseAny(edb.Finish)
	if err != nil {
		return nil, err

	}
	hoursParsed, err := decimal.NewFromString(edb.Hours)
	if err != nil {
		return nil, err
	}
	entry.ID = int64(idParsed)
	entry.Date = edb.Date
	entry.Begin = beginParsed
	entry.Finish = finishParsed
	entry.Project = edb.Project
	entry.Hours = hoursParsed
	entry.Task = edb.Task
	entry.Notes = edb.Notes
	entry.Running = edb.Running

	return &entry, nil

}

type EntriesGroupedByDay struct {
	Date     string
	Projects int8
	Tasks    int8
	Hours    decimal.Decimal
}

func NewEntry(project string, task string) Entry {

	newEntry := Entry{}

	newEntry.Project = project
	newEntry.Task = task

	newEntry.SetBeginingToNow()
	newEntry.SetDateFromBegining()
	return newEntry
}

func (entry *Entry) SetDateFromBegining() {
	entry.Date = entry.Begin.Format("02-01-2006")
}

func (entry *Entry) SetBeginingToNow() error {
	formatedTime, err := time.Parse("2006-01-02 15:04", time.Now().Truncate(0).Format("2006-01-02 15:04"))
	if err != nil {
		return err
	}
	entry.Begin = formatedTime
	entry.Running = true
	return nil
}

func (entry *Entry) GetOutputStrLong() string {
	return fmt.Sprintf(`Task: %s on Project: %s started at: %s finished at: %s and in total has %s hours`,
		entry.Task, entry.Project, entry.Begin.String(), entry.Finish.String(), entry.Hours.String())
}

func (entry *Entry) GetOutputStrShort() string {
	return fmt.Sprintf(`Task: %s | Project: %s |  Dated: %s | Hours: %s `,
		entry.Task, entry.Project, entry.Date, entry.Hours.String())
}

func (entry *Entry) GetStartTrackingStr() string {
	return fmt.Sprintf(`Started tracking --> Task: %s on Project: %s `,
		entry.Task, entry.Project)
}

func (entry *Entry) SetBeginFromString(begin string) (time.Time, error) {
	var beginTime time.Time
	var err error

	if begin == "" {
		beginTime = time.Now().Truncate(0)
	} else {
		beginTime, err = ParseTime(begin)
		if err != nil {
			return beginTime, err
		}
	}

	entry.Begin = beginTime
	return beginTime, nil
}
func (entry *Entry) SetFinish() error {
	formatedTime, err := time.Parse("2006-01-02 15:04", time.Now().Truncate(0).Format("2006-01-02 15:04"))
	if err != nil {
		return err
	}

	entry.Finish = formatedTime
	entry.Running = false
	return nil
}

func (entry *Entry) SetFinishFromString(finish string) (time.Time, error) {
	var finishTime time.Time
	var err error

	if finish != "" {
		finishTime, err = ParseTime(finish)
		if err != nil {
			return finishTime, err
		}
	}

	entry.Finish = finishTime
	return finishTime, nil
}

func (entry *Entry) GetCSVHeaderAllData() []string {
	return []string{"date", "begin", "finish", "project", "task", "hours", "notes"}
}
func (entry *Entry) GetCSVHeaderShortData() []string {
	return []string{"date", "project", "task", "hours"}
}

func (entry *Entry) IsFinishedAfterBegan() bool {
	return (entry.Finish.IsZero() || entry.Begin.Before(entry.Finish))
}

func (entry *Entry) GetOutputForTrack(isRunning bool, wasRunning bool) string {
	var outputSuffix = ""
	var outputPrefix = ""

	now := time.Now().Truncate(0)
	trackDiffNow := now.Sub(entry.Begin)
	durationString := fmtDuration(trackDiffNow)

	if isRunning && !wasRunning {
		outputPrefix = "began tracking"
	} else if isRunning && wasRunning {
		outputPrefix = "tracking"
		outputSuffix = fmt.Sprintf(" for %sh", color.FgLightWhite.Render(durationString))
	} else if !isRunning && !wasRunning {
		outputPrefix = "tracked"
	}

	if entry.Task != "" && entry.Project != "" {
		return fmt.Sprintf("%s %s %s on %s%s\n", CharTrack, outputPrefix, color.FgLightWhite.Render(entry.Task), color.FgLightWhite.Render(entry.Project), outputSuffix)
	} else if entry.Task != "" && entry.Project == "" {
		return fmt.Sprintf("%s %s %s%s\n", CharTrack, outputPrefix, color.FgLightWhite.Render(entry.Task), outputSuffix)
	} else if entry.Task == "" && entry.Project != "" {
		return fmt.Sprintf("%s %s task on %s%s\n", CharTrack, outputPrefix, color.FgLightWhite.Render(entry.Project), outputSuffix)
	}

	return fmt.Sprintf("%s %s task%s\n", CharTrack, outputPrefix, outputSuffix)
}

func (entry *Entry) GetDuration() decimal.Decimal {
	duration := entry.Finish.Sub(entry.Begin)
	if duration < 0 {
		duration = time.Since(entry.Begin)
	}
	return decimal.NewFromFloat(duration.Hours())
}

func (entry *Entry) ConvertToCSVAllData() []string {
	return []string{
		entry.Date,
		entry.Begin.String(),
		entry.Finish.String(),
		entry.Project,
		entry.Task,
		entry.Hours.String(),
		entry.Notes,
	}
}

func (entry *Entry) ConvertToCSVShortData() []string {
	return []string{
		entry.Date,
		entry.Project,
		entry.Task,
		entry.Hours.String(),
	}
}

func (entry *Entry) GetOutputForFinish() string {
	var outputSuffix = ""

	trackDiff := entry.Finish.Sub(entry.Begin)
	taskDuration := fmtDuration(trackDiff)

	outputSuffix = fmt.Sprintf(" for %sh", color.FgLightWhite.Render(taskDuration))

	if entry.Task != "" && entry.Project != "" {
		return fmt.Sprintf("%s finished tracking %s on %s%s\n", CharFinish, color.FgLightWhite.Render(entry.Task), color.FgLightWhite.Render(entry.Project), outputSuffix)
	} else if entry.Task != "" && entry.Project == "" {
		return fmt.Sprintf("%s finished tracking %s%s\n", CharFinish, color.FgLightWhite.Render(entry.Task), outputSuffix)
	} else if entry.Task == "" && entry.Project != "" {
		return fmt.Sprintf("%s finished tracking task on %s%s\n", CharFinish, color.FgLightWhite.Render(entry.Project), outputSuffix)
	}

	return fmt.Sprintf("%s finished tracking task%s\n", CharFinish, outputSuffix)
}

func (entry *Entry) GetOutput(full bool) string {
	var output = ""
	var entryFinish time.Time
	var isRunning = ""

	if entry.Finish.IsZero() {
		entryFinish = time.Now().Truncate(0)
		isRunning = "[running]"
	} else {
		entryFinish = entry.Finish
	}

	trackDiff := entryFinish.Sub(entry.Begin)
	taskDuration := fmtDuration(trackDiff)
	if !full {

		output = fmt.Sprintf("%s %s on %s from %s to %s (%sh) %s",
			color.FgGray.Render(entry.ID),
			color.FgLightWhite.Render(entry.Task),
			color.FgLightWhite.Render(entry.Project),
			color.FgLightWhite.Render(entry.Begin.Format("2006-01-02 15:04")),
			color.FgLightWhite.Render(entryFinish.Format("2006-01-02 15:04")),
			color.FgLightWhite.Render(taskDuration),
			color.FgLightYellow.Render(isRunning),
		)
	} else {
		output = fmt.Sprintf("%s\n   %s on %s\n   %sh from %s to %s %s\n\n   Notes:\n   %s\n",
			color.FgGray.Render(entry.ID),
			color.FgLightWhite.Render(entry.Task),
			color.FgLightWhite.Render(entry.Project),
			color.FgLightWhite.Render(taskDuration),
			color.FgLightWhite.Render(entry.Begin.Format("2006-01-02 15:04")),
			color.FgLightWhite.Render(entryFinish.Format("2006-01-02 15:04")),
			color.FgLightYellow.Render(isRunning),
			color.FgLightWhite.Render(strings.Replace(entry.Notes, "\n", "\n   ", -1)),
		)
	}

	return output
}

func GetFilteredEntries(entries []Entry, project string, task string, since time.Time, until time.Time) ([]Entry, error) {
	var filteredEntries []Entry

	for _, entry := range entries {
		if project != "" && GetIdFromName(entry.Project) != GetIdFromName(project) {
			continue
		}

		if task != "" && GetIdFromName(entry.Task) != GetIdFromName(task) {
			continue
		}

		if !since.IsZero() && !since.Before(entry.Begin) && !since.Equal(entry.Begin) {
			continue
		}

		if !until.IsZero() && !until.After(entry.Finish) && !until.Equal(entry.Finish) {
			continue
		}

		filteredEntries = append(filteredEntries, entry)
	}

	return filteredEntries, nil
}
