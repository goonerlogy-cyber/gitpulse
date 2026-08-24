# gitpulse

A terminal analytics dashboard for any git repository. Point it at a repo and get a GitHub-style contribution heatmap, commit velocity, language breakdown, and top contributors, all rendered live in your terminal.

No API keys, no network calls. It reads local git history and the working tree, nothing else.

## Install

```
go install github.com/arshnah/gitpulse@latest
```

Or build from source:

```
git clone https://github.com/arshnah/gitpulse
cd gitpulse
go build -o gitpulse .
```

## Usage

```
gitpulse [path]
```

Defaults to the current directory if no path is given.

Controls:

- `←`/`→` or `h`/`l` to switch tabs
- `1`-`4` to jump to a tab
- `q` or `ctrl+c` to quit

## Tabs

- **Overview** - total commits, contributor count, repo age, and a 26-week contribution heatmap
- **Activity** - a commit velocity sparkline over the last 24 weeks and a feed of recent commits
- **Languages** - a breakdown of tracked files by extension, weighted by file size
- **Contributors** - top committers ranked by commit count

## How it works

gitpulse shells out to `git log` and `git ls-files` to build its dataset, then renders everything with [bubbletea](https://github.com/charmbracelet/bubbletea) and [lipgloss](https://github.com/charmbracelet/lipgloss). No git library dependency, no server, no telemetry.
