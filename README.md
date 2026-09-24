# gitpulse

A terminal analytics dashboard for any git repository. Point it at a repo and get a GitHub-style contribution heatmap, commit velocity, language breakdown, and top contributors, all rendered live in your terminal.

No API keys, no network calls. It reads local git history and the working tree, nothing else.

This fork of [arshnah/gitpulse](https://github.com/arshnah/gitpulse) adds JSON
reports and more reliable working-tree accounting. The original dashboard and
license are retained.

## Install

Build this fork from source with Go 1.27 or newer:

```
git clone https://github.com/goonerlogy-cyber/gitpulse
cd gitpulse
go build -o gitpulse .
```

## Usage

```
gitpulse [--json] [path]
```

Defaults to the current directory if no path is given.

### JSON reports

Use `--json` before the path to produce a report without opening a terminal UI:

```sh
gitpulse --json . > report.json
gitpulse --json ./src
gitpulse --json -- ./-unusual-repo-name
```

Paths inside a repository resolve to its root, so reporting from a subdirectory
includes the complete working tree. A new repository with no commits returns zero
counts and `null` first/last dates. The dashboard also handles this empty state.

The report has `schema_version: 1`, the absolute `repository` path, `name`,
`total_commits`, `contributors` (author names to counts), `daily_counts`
(`YYYY-MM-DD` to counts), `language_bytes`, and RFC 3339 `first_commit` and
`last_commit` timestamps. Raw commit subjects and email addresses are omitted.
Output is a single JSON object; failures return a nonzero exit code.

Counts include non-merge commits reachable from all local refs. Daily buckets use
each commit author's recorded time zone. Language bytes reflect the current sizes
of recognized tracked regular files, including staged files; deleted files,
symlinks, submodules, and untracked files are excluded. This is an extension-based
working-tree estimate, not GitHub Linguist or a historical language breakdown.

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

## Development

```sh
go test ./...
go vet ./...
go build ./...
```

Tests use temporary Git repositories and cover empty history, subdirectory paths,
quoted/unicode filenames, output failures, and non-interactive JSON output. CI
runs on Linux and Windows; symlink tests skip when Windows cannot create them.
