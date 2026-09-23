package gitdata

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Commit struct {
	Hash    string
	Author  string
	Email   string
	Date    time.Time
	Subject string
}

type Data struct {
	RepoPath     string
	RepoName     string
	Commits      []Commit
	Contributors map[string]int
	DailyCounts  map[string]int
	Languages    map[string]int64
	FirstCommit  time.Time
	LastCommit   time.Time
}

const logSep = "\x1f"

func Load(path string) (*Data, error) {
	abs, err := repoRoot(path)
	if err != nil {
		return nil, err
	}
	commits, err := loadCommits(abs)
	if err != nil {
		return nil, err
	}
	langs, err := loadLanguages(abs)
	if err != nil {
		return nil, fmt.Errorf("load tracked files: %w", err)
	}

	d := &Data{
		RepoPath:     abs,
		RepoName:     filepath.Base(abs),
		Commits:      commits,
		Contributors: map[string]int{},
		DailyCounts:  map[string]int{},
		Languages:    langs,
	}
	if len(commits) > 0 {
		d.FirstCommit = commits[len(commits)-1].Date
		d.LastCommit = commits[0].Date
	}

	for _, c := range commits {
		d.Contributors[c.Author]++
		d.DailyCounts[c.Date.Format("2006-01-02")]++
	}

	return d, nil
}

func repoRoot(path string) (string, error) {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("%s is not a readable git working tree: %w", path, err)
	}
	return filepath.Clean(strings.TrimSuffix(string(out), "\n")), nil
}

func loadCommits(path string) ([]Commit, error) {
	format := strings.Join([]string{"%H", "%an", "%ae", "%aI", "%s"}, logSep)
	cmd := exec.Command("git", "-C", path, "log", "--all", "--no-merges", "--pretty=format:"+format)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log failed: %w", err)
	}

	commits := make([]Commit, 0)
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, logSep, 5)
		if len(parts) != 5 {
			continue
		}
		t, err := time.Parse(time.RFC3339, parts[3])
		if err != nil {
			continue
		}
		commits = append(commits, Commit{
			Hash:    parts[0],
			Author:  parts[1],
			Email:   parts[2],
			Date:    t,
			Subject: parts[4],
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read git log: %w", err)
	}
	sort.Slice(commits, func(i, j int) bool { return commits[i].Date.After(commits[j].Date) })
	return commits, nil
}

var extLang = map[string]string{
	".go":    "Go",
	".ts":    "TypeScript",
	".tsx":   "TypeScript",
	".js":    "JavaScript",
	".jsx":   "JavaScript",
	".py":    "Python",
	".rb":    "Ruby",
	".rs":    "Rust",
	".java":  "Java",
	".c":     "C",
	".h":     "C",
	".cpp":   "C++",
	".hpp":   "C++",
	".cc":    "C++",
	".cs":    "C#",
	".php":   "PHP",
	".swift": "Swift",
	".kt":    "Kotlin",
	".m":     "Objective-C",
	".sh":    "Shell",
	".zsh":   "Shell",
	".bash":  "Shell",
	".html":  "HTML",
	".css":   "CSS",
	".scss":  "CSS",
	".sql":   "SQL",
	".lua":   "Lua",
	".vim":   "VimScript",
	".ex":    "Elixir",
	".exs":   "Elixir",
	".dart":  "Dart",
	".yaml":  "YAML",
	".yml":   "YAML",
	".json":  "JSON",
	".md":    "Markdown",
	".vue":   "Vue",
	".zig":   "Zig",
	".hs":    "Haskell",
	".scala": "Scala",
	".erl":   "Erlang",
	".r":     "R",
	".pl":    "Perl",
	".clj":   "Clojure",
	".ml":    "OCaml",
	".jl":    "Julia",
}

func loadLanguages(path string) (map[string]int64, error) {
	cmd := exec.Command("git", "-C", path, "ls-files", "--stage", "-z")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	langs := map[string]int64{}
	seen := make(map[string]bool)
	for _, record := range strings.Split(string(out), "\x00") {
		if record == "" {
			continue
		}
		header, f, ok := strings.Cut(record, "\t")
		fields := strings.Fields(header)
		if !ok || len(fields) != 3 {
			return nil, fmt.Errorf("invalid git ls-files record")
		}
		// Git can check out mode-120000 links as ordinary files when
		// core.symlinks=false. The index, not the filesystem, identifies them.
		if fields[0] != "100644" && fields[0] != "100755" {
			continue
		}
		if seen[f] {
			continue
		}
		seen[f] = true
		ext := strings.ToLower(filepath.Ext(f))
		lang, ok := extLang[ext]
		if !ok {
			continue
		}
		fi, err := os.Lstat(filepath.Join(path, f))
		if err != nil || !fi.Mode().IsRegular() {
			continue
		}
		langs[lang] += fi.Size()
	}
	return langs, nil
}

func (d *Data) WeeklyVelocity(weeks int) []int {
	buckets := make([]int, weeks)
	now := time.Now()
	for _, c := range d.Commits {
		daysAgo := int(now.Sub(c.Date).Hours() / 24)
		week := daysAgo / 7
		idx := weeks - 1 - week
		if idx >= 0 && idx < weeks {
			buckets[idx]++
		}
	}
	return buckets
}

func (d *Data) TopContributors(n int) []struct {
	Name  string
	Count int
} {
	type kv struct {
		Name  string
		Count int
	}
	var list []kv
	for k, v := range d.Contributors {
		list = append(list, kv{k, v})
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].Count != list[j].Count {
			return list[i].Count > list[j].Count
		}
		return list[i].Name < list[j].Name
	})
	if len(list) > n {
		list = list[:n]
	}
	result := make([]struct {
		Name  string
		Count int
	}, len(list))
	for i, kv := range list {
		result[i] = struct {
			Name  string
			Count int
		}{kv.Name, kv.Count}
	}
	return result
}

func FormatCount(n int) string {
	if n < 1000 {
		return strconv.Itoa(n)
	}
	return fmt.Sprintf("%.1fk", float64(n)/1000)
}
