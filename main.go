package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/arshnah/gitpulse/internal/gitdata"
	"github.com/arshnah/gitpulse/internal/ui"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "gitpulse: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("gitpulse", flag.ContinueOnError)
	flags.SetOutput(stderr)
	jsonOutput := flags.Bool("json", false, "write a JSON report without opening the dashboard")
	flags.Usage = func() { printUsage(stdout) }
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() > 1 {
		return fmt.Errorf("expected at most one repository path")
	}
	path := "."
	if flags.NArg() == 1 {
		path = flags.Arg(0)
	}
	data, err := gitdata.Load(path)
	if err != nil {
		return err
	}
	if *jsonOutput {
		return gitdata.WriteReport(stdout, data)
	}

	m := ui.New(data)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithOutput(stdout))
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, `gitpulse - a terminal analytics dashboard for any git repo

Usage:
  gitpulse [--json] [path]

Arguments:
  path    Path to a git repository (defaults to the current directory)

Options:
  --json       write a JSON report, suitable for scripts and CI
  -h, --help   show this help
  --           end options, for paths beginning with a hyphen

Controls:
  ←/→ or h/l    switch tabs
  1-4           jump to a tab
  q / ctrl+c    quit`)
}
