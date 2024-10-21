package chron

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	heightOffset = 20
	widthOffset  = 20
)

var paddingStyle = lipgloss.NewStyle().
	PaddingTop(2).
	PaddingRight(4).
	PaddingBottom(2).
	PaddingLeft(4)

var centerAlignStyle = lipgloss.NewStyle().Align(lipgloss.Center)

var barchartDefaultStyle = lipgloss.NewStyle().BorderForeground(lipgloss.Color("240")).Align(lipgloss.Center)

// BorderBackground(lipgloss.Color("#81b29a")).
//
//	BorderForeground(lipgloss.Color("#81b29a")).
//
// Background(lipgloss.Color("#3d405b")).
var baseStyle = lipgloss.NewStyle().
	Align(lipgloss.Center)

var tableHeaderStyle = lipgloss.NewStyle().
	BorderForeground(lipgloss.Color("240")).
	BorderBottom(true).
	AlignVertical(lipgloss.Center).
	AlignHorizontal(lipgloss.Center).
	Bold(false)

var selectedEntryStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("229")).
	Background(lipgloss.Color("57")).
	AlignVertical(lipgloss.Center).
	AlignHorizontal(lipgloss.Center).
	Bold(false)
