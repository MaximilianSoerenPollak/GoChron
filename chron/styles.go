package chron

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

var (
	totalStyleHeight = 4
	totalStyleWidth  = 4
	termWidth        int
	termHeight       int
)

var paddingStyle = lipgloss.NewStyle().
	PaddingTop(2).
	PaddingRight(2).
	PaddingBottom(2).
	PaddingLeft(2)

var centerAlignStyle = lipgloss.NewStyle().Align(lipgloss.Center)

var barchartDefaultStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))

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

// top, left, right ,bottom
func getStyleBorderSize(style lipgloss.Style) (int, int, int, int) {
	return style.GetBorderTopSize(),
		style.GetBorderLeftSize(),
		style.GetBorderRightSize(),
		style.GetBorderBottomSize()
}

// top, left, right ,bottom
func getStylePadding(style lipgloss.Style) (int, int, int, int) {
	return style.GetPaddingTop(),
		style.GetPaddingLeft(),
		style.GetPaddingRight(),
		style.GetPaddingBottom()
}

func setTerminalSize() {
	tW, tH, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		fmt.Printf("%s could not get terminal size. Error: %s\n", CharError, err.Error())
		os.Exit(1)
	}
	termWidth = tW
	termHeight = tH
}


// totalHeight, totalWidth
func calculateTotalStyleSize(styles ...lipgloss.Style) (int, int) {
	totalHeight := 0
	totalWidth := 0
	for _, style := range styles {
		a,b,c,d := getStyleBorderSize(style)
		e,f,g,h := getStylePadding(style)
		totalHeight += a + d + e + h
		totalWidth +=  b + c + f + g
	}
	return totalHeight, totalWidth
}


