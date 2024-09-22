package chron

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	heightOffset  = 20
	widthOffset = 20
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

var tableHeaderStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240")).
	BorderBottom(true).
	Bold(false)

var selectedEntryStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("229")).
	Background(lipgloss.Color("57")).
	Bold(false)
