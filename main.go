package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/magus-1/mfp/internal/tui"
)

// main.go
func main() {
	// truncate on each run so you always see a fresh log
	_ = os.Truncate("debug.log", 0)

	f, err := tea.LogToFile("debug.log", "")
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not open log file:", err)
		os.Exit(1)
	}
	defer f.Close()

	log.SetFlags(log.Ltime | log.Lshortfile) // timestamp + file:line, no date noise

	log.Println("--- app start ---")

	p := tea.NewProgram(tui.NewModel(), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
