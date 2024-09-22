package chron

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss/list"
	"github.com/davecgh/go-spew/spew"
)

type taskForm struct {
	form *huh.Form
	dump io.Writer
}

var oldProject bool
var newEntry Entry
var taskRunning bool
var stopTask bool

func initMainForm(dump io.Writer) taskForm {
	uniqueProjects, err := database.GetUniqueProjects()
	if err != nil {
		fmt.Printf("%s %+v\n", CharError, err)
		os.Exit(1)
	}
	oldProjectSelection := huh.NewSelect[string]().
		Title("Select Project").
		Options(huh.NewOptions(uniqueProjects...)...).
		Value(&newEntry.Project)
	newProjectEntry := huh.NewInput().Title("Enter Project").Value(&newEntry.Project).Validate(huh.ValidateNotEmpty())
	// Have to declare here otherwise form will be out of scope.
	var form *huh.Form
	if oldProject {
		form = huh.NewForm(
			huh.NewGroup(
				huh.NewInput().Title("Task").Value(&newEntry.Task).Validate(huh.ValidateNotEmpty()),
				oldProjectSelection,
				huh.NewInput().Title("Notes").Value(&newEntry.Notes),
			),
		)
	} else {
		form = huh.NewForm(
			huh.NewGroup(
				huh.NewInput().Title("Task").Value(&newEntry.Task).Validate(huh.ValidateNotEmpty()),
				newProjectEntry,
				huh.NewInput().Title("Notes").Value(&newEntry.Notes),
			),
		)
	}
	return taskForm{form: form, dump: dump}
}

func (m taskForm) Init() tea.Cmd {
	taskRunning = true
	runningEntry, err := database.GetRunningEntry()
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			taskRunning = false
		default:
			fmt.Printf("something went wrong getting current runnign entries. Error: %s", err.Error())
			os.Exit(1)
		}
	}
	if !taskRunning {
		return m.form.Init()
	}
	err = huh.NewForm(huh.NewGroup(huh.NewConfirm().Title("Stop running task and start new one?").Value(&stopTask))).Run()
	if err != nil {
		fmt.Printf("something went wrong running the confirm form. Error: %s", err.Error())
		os.Exit(1)
	}
	if !stopTask {
		return func() tea.Msg { return switchToListModel{} }
	}
	err = runningEntry.SetFinish()
	if err != nil {
		fmt.Printf("%s could not convert finish time to standard. Error: %s\n", CharError, err.Error())
		os.Exit(1)
	}
	err = database.AddFinishToEntry(*runningEntry)
	if err != nil {
		fmt.Printf("%s could not set finish on entry in DB. Error: %s\n", CharError, err.Error())
		os.Exit(1)
	}
	return m.form.Init()
}

// ...
func (m taskForm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.dump != nil {
		spew.Fdump(m.dump, fmt.Sprintf("taskForm: %s", msg))
	}

	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
	}
	if m.form.State == huh.StateCompleted {
		newEntry.SetBeginingToNow()
		newEntry.SetDateFromBegining()
		err := database.AddEntry(&newEntry, true)
		if err != nil {
			fmt.Printf("%s could not add entry to the DB. Error: %s\n", CharError, err.Error())
			os.Exit(1)
		}
		cmds = append(cmds, func() tea.Msg { return switchToListModel{} })
		// return m, func() tea.Msg { return switchToListModel{} }
	}

	return m, tea.Batch(cmds...)
}

func (m taskForm) View() string {
	if m.form.State == huh.StateCompleted {
		return fmt.Sprintf("Task added successfully:\n %s", createNewlyAddedTaskList())
	}
	return m.form.View()
}

func createNewlyAddedTaskList() *list.List {
	return list.New(
		"ID", list.New(newEntry.ID),
		"Date", list.New(newEntry.Date),
		"Start", list.New(newEntry.Begin),
		"Finish", list.New("Task is running"),
		"Task", list.New(newEntry.Task),
		"Project", list.New(newEntry.Project),
		"Notes", list.New(newEntry.Notes),
	)
}
