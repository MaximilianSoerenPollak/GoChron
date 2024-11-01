package chron

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/davecgh/go-spew/spew"
)

// Q: File picker or import model?
type importModel struct {
	filepicker   filepicker.Model
	selectedFile string
	err          error
	dump         io.Writer
}

// Unsure if this is needed?
type clearErrorMsg struct{}

func clearErrorAfter(t time.Duration) tea.Cmd {
	return tea.Tick(t, func(_ time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

func initImportModel(dump io.Writer) importModel {
	// Create new simple filepicker MOdle

	fp := filepicker.New()
	// TODO: also allow 'zeit' input
	fp.AllowedTypes = []string{".csv"}
	fp.CurrentDirectory, _ = os.UserHomeDir()
	fp.Height = termHeight / 2
	// fp.AutoHeight = true

	return importModel{
		filepicker: fp,
		dump:       dump,
	}
}

func (m importModel) Init() tea.Cmd {
	return m.filepicker.Init()
}

func (m importModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.dump != nil {
		spew.Fdump(m.dump, msg)
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "ctrl+d":
			return m, func() tea.Msg { return switchToListModel{} }
		}

	case clearErrorMsg:
		m.err = nil
	}

	var cmd tea.Cmd
	m.filepicker, cmd = m.filepicker.Update(msg)

	// Did the user select a file?
	if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
		// Get the path of the selected file.
		m.selectedFile = path

		currDir := m.filepicker.CurrentDirectory
		cleanedStr := strings.Replace(path, currDir, "", 1)
		var confirm bool
		huh.NewConfirm().
			Title(fmt.Sprintf("Import file: %s from Dir: %s ?", cleanedStr, currDir)).
			Value(&confirm).
			Run()

		// If 'not confirmed'
		if !confirm {
			m.err = errors.New("import not confirmed.")
			m.selectedFile = ""
			return m, tea.Batch(cmd, clearErrorAfter(1*time.Second))
		}
		contents := readCSVFile(m.selectedFile)
		parseCSVContents(contents)

	}

	// Did the user select a disabled file?
	// This is only necessary to display an error to the user.
	if didSelect, path := m.filepicker.DidSelectDisabledFile(msg); didSelect {
		// Let's clear the selectedFile and display an error.
		m.err = errors.New(path + " is not valid.")
		m.selectedFile = ""
		return m, tea.Batch(cmd, clearErrorAfter(2*time.Second))
	}

	return m, cmd

}

func (m importModel) View() string {
	var s strings.Builder
	s.WriteString("\n  ")
	if m.err != nil {
		s.WriteString(m.filepicker.Styles.DisabledFile.Render(m.err.Error()))
	} else if m.selectedFile == "" {
		s.WriteString("Pick a file:")
	} else {
		s.WriteString("Selected file: " + m.filepicker.Styles.Selected.Render(m.selectedFile))
	}
	s.WriteString("\n\n" + m.filepicker.View() + "\n")
	return s.String()
}

func readCSVFile(path string) [][]string {
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("%s The choosen file '%s' could not be opened. Error: %s\n", CharError, path, err.Error())
		os.Exit(1)
	}
	csvReader := csv.NewReader(file)
	csvReader.Comma = ','
	content, err := csvReader.ReadAll()
	if err != nil {
		fmt.Printf("%s Something went wrong reading the choosen file '%s'. Error: %s\n", CharError, path, err.Error())
		os.Exit(1)
	}
	return content
}

func parseCSVContents(content [][]string) {
	// We are skipping the header
	for _, v := range content[1:] {
		if len(v) != 7 {
			fmt.Printf("%s The csv has the wrong format, please make sure you export the csv with all fields\n", CharError)
			fmt.Printf("%s The fields expected in THIS order: date, start, finish, project, task, hours, notes\n", CharInfo)
			os.Exit(1)
		}
		var eDB EntryDB
		// Hack to make it work, probably should fix this someday
		eDB.ID = "1"
		eDB.Date = v[0]
		eDB.Begin = v[1]
		eDB.Finish = v[2]
		eDB.Project = v[3]
		eDB.Task = v[4]
		eDB.Hours = v[5]
		eDB.Notes = v[6]
		eDB.Running = false
		entryConv, err := eDB.ConvertToEntry()
		if err != nil {
			fmt.Printf("%s Could not convert entryDB '%+v' to entry. Error: %s\n", CharError, eDB, err.Error())
			os.Exit(1)
		}
		err = database.AddEntry(entryConv, false)
		if err != nil {
			fmt.Printf("%s Could not add entry '%+v' to the database. Error: %s\n", CharError, entryConv, err.Error())
			os.Exit(1)
		}
		if verbose {
			fmt.Printf("%s added Entry: '%s' to the database\n", CharInfo, entryConv.GetOutputStrShort())
		}
	}
	fmt.Printf("%s added all entries to the database\n", CharInfo)
}
