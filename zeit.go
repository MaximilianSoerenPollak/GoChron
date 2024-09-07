package main

import (
	"fmt"
	"os"
	"github.com/MaximilianSoerenPollak/zeit/z"
	"github.com/charmbracelet/bubbletea"
)

func main() {
    p := tea.NewProgram(z.InitialModel(), tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
        fmt.Printf("Alas, there's been an error: %v", err)
        os.Exit(1)
    }
	// z.Execute()
}
