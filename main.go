package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/arshnah/gitpulse/internal/gitdata"
	"github.com/arshnah/gitpulse/internal/ui"
)

func main() {
	path := "."
	if len(os.Args) > 1 {
		if os.Args[1] == "-h" || os.Args[1] == "--help" {
			printUsage()
			return
		}
		path = os.Args[1]
	}

	data, err := gitdata.Load(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gitpulse: %v\n", err)
		os.Exit(1)
	}

	m := ui.New(data)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "gitpulse: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`gitpulse - a terminal analytics dashboard for any git repo

Usage:
  gitpulse [path]

Arguments:
  path    Path to a git repository (defaults to the current directory)

Controls:
  ←/→ or h/l    switch tabs
  1-4           jump to a tab
  q / ctrl+c    quit`)
}
