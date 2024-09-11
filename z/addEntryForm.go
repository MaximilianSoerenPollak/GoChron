package z

import (
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
		cmds = append(cmds, func() tea.Msg { return switchToListModel{} } )
		// return m, func() tea.Msg { return switchToListModel{} }
	}

	return m, tea.Batch(cmds...)
}

func (m taskForm) View() string {
	if m.form.State == huh.StateCompleted {
		newEntry.SetBegining()
		newEntry.SetDateFromBegining()
		database.AddEntry(&newEntry, true)
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
