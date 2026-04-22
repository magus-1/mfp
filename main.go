package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/magus-1/mfp/internal/tui"
)

func main() {
	p := tea.NewProgram(
		tui.NewModel(),
		tea.WithAltScreen(),       // takes over the full terminal
		tea.WithMouseCellMotion(), // enables mouse scroll
	)

	f, _ := tea.LogToFile("debug.log", "debug")
	defer f.Close()

	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error running program:", err)
		log.Fatal(err)
	}
}
