package main

import (
	"fmt"
	"os"

	"github.com/MaximilianSoerenPollak/zeit/z"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	var dump *os.File
	if _, ok := os.LookupEnv("DEBUG"); ok {
		fmt.Printf("Gotten debug")
		var err error
		dump, err = os.OpenFile("messages.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
		if err != nil {
			os.Exit(1)
		}
	}
	m := z.InitialModel(dump)
	p := tea.NewProgram(&m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
	// z.Execute()
}
