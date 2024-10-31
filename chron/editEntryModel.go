package chron

import (
	"fmt"
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/davecgh/go-spew/spew"
	"golang.org/x/term"
)

var entryToChange EntryDB

type changeEntryModel struct {
	form     *huh.Form
	dump     io.Writer
	oldEntry EntryDB
}

func (m changeEntryModel) Init() tea.Cmd {
	return m.form.Init()
}

func initChangeEntryForm(dump io.Writer, entry EntryDB) changeEntryModel {
	// Not sure if this is a good way ?
	entryToChange.ID = entry.ID
	entryToChange.Date = entry.Date
	entryToChange.Hours = entry.Hours
	entryToChange.Begin = entry.Begin
	entryToChange.Finish = entry.Finish
	entryToChange.Task = entry.Task
	entryToChange.Project = entry.Project
	entryToChange.Notes = entry.Notes
	f := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Changing Task").
				Description("You can change any field of the Task. If you like to cancel just press 'esc' to go back or 'ctrl+c' to exit"),
			huh.NewNote().
				Title("Date").
				Description(fmt.Sprintf("If you want to change the date, change the 'start' field. Current Task date: %s ", entry.Date)),
			huh.NewNote().
				Title("Hours").
				Description(fmt.Sprintf("The hours the task was tracked for. This changes automatically when 'start' or 'finish' changes. Current hours: %s", entry.Hours)),
			huh.NewInput().
				Title("Start").
				Description("When the task started tracking").
				Placeholder(entry.Begin).
				Value(&entryToChange.Begin),
			huh.NewInput().
				Title("Finish").
				Description("When the task stopped tracking").
				Placeholder(entry.Finish).
				Value(&entryToChange.Finish),
			huh.NewInput().
				Title("Task").
				Placeholder(entry.Task).
				Value(&entryToChange.Task),
			huh.NewInput().
				Title("Project").
				Description("What project the task currently is assigned to").
				Placeholder(entry.Project).
				Value(&entryToChange.Project),
			huh.NewInput().
				Title("Notes").
				Description("The notes associated with the task").
				Placeholder(entry.Notes).
				Value(&entryToChange.Notes),
			huh.NewConfirm().
				Title("Confirm Changes")))
	termWidth, termHeight, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		fmt.Printf("%s could not get terminal size. Error: %s\n", CharError, err.Error())
		os.Exit(1)
	}
	f.WithHeight(termHeight)
	f.WithWidth(termWidth)
	return changeEntryModel{form: f, dump: dump, oldEntry: entry}
}

func (m changeEntryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	if m.dump != nil {
		spew.Fdump(m.dump, msg)
	}
	// We only check global keys here.
	// Ctrl+c for example
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit
			// This seems to never execute?
		case "esc":
			return m, func() tea.Msg { return switchToListModel{} }
		}
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
	}
	if m.form.State == huh.StateCompleted {
		entryToChange.FormatTimes()
		convEntry, err := entryToChange.ConvertToEntry()
		if err != nil {
			fmt.Printf("%s could not convert entryDB to entry. Error: %s\n", CharError, err.Error())
			os.Exit(1)
		}
		convEntry.SetDateFromBegining()
		convEntry.Hours = convEntry.GetDuration()
		err = database.UpdateEntry(*convEntry)
		if err != nil {
			fmt.Printf("%s could not add entry to the DB. Error: %s\n", CharError, err.Error())
			os.Exit(1)
		}
		cmds = append(cmds, func() tea.Msg { return switchToListModel{} })
	}
	return m, tea.Batch(cmds...)
}

func (m changeEntryModel) View() string {
	if m.form.State == huh.StateCompleted {
		return fmt.Sprintf("Task added successfully:\n %s", createNewlyAddedTaskList())
	}
	return m.form.View()
}
