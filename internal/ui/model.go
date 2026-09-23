package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/arshnah/gitpulse/internal/gitdata"
)

type tab int

const (
	tabOverview tab = iota
	tabActivity
	tabLanguages
	tabContributors
	tabCount
)

var tabNames = []string{"Overview", "Activity", "Languages", "Contributors"}

type Model struct {
	data   *gitdata.Data
	active tab
	width  int
	height int
}

func New(data *gitdata.Data) Model {
	return Model{data: data, active: tabOverview, width: 100, height: 30}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "right", "l", "tab":
			m.active = (m.active + 1) % tabCount
		case "left", "h", "shift+tab":
			m.active = (m.active - 1 + tabCount) % tabCount
		case "1":
			m.active = tabOverview
		case "2":
			m.active = tabActivity
		case "3":
			m.active = tabLanguages
		case "4":
			m.active = tabContributors
		}
	}
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder

	header := titleStyle.Render(" gitpulse ") + "  " + subtitleStyle.Render(m.data.RepoName)
	b.WriteString(header)
	b.WriteString("\n\n")

	var tabs []string
	for i, name := range tabNames {
		if tab(i) == m.active {
			tabs = append(tabs, tabActiveStyle.Render(name))
		} else {
			tabs = append(tabs, tabInactiveStyle.Render(name))
		}
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, tabs...))
	b.WriteString("\n\n")

	switch m.active {
	case tabOverview:
		b.WriteString(m.viewOverview())
	case tabActivity:
		b.WriteString(m.viewActivity())
	case tabLanguages:
		b.WriteString(m.viewLanguages())
	case tabContributors:
		b.WriteString(m.viewContributors())
	}

	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("←/→ switch tabs   1-4 jump   q quit"))

	return b.String()
}

func (m Model) viewOverview() string {
	d := m.data
	if len(d.Commits) == 0 {
		return subtitleStyle.Render("No commits yet. Tracked files are available in the Languages tab.")
	}
	days := int(d.LastCommit.Sub(d.FirstCommit).Hours()/24) + 1
	if days < 1 {
		days = 1
	}
	avgPerDay := float64(len(d.Commits)) / float64(days)

	stats := []struct{ label, value string }{
		{"Total commits", gitdata.FormatCount(len(d.Commits))},
		{"Contributors", fmt.Sprintf("%d", len(d.Contributors))},
		{"Active since", d.FirstCommit.Format("Jan 2, 2006")},
		{"Avg commits/day", fmt.Sprintf("%.1f", avgPerDay)},
	}

	var statLines []string
	for _, s := range stats {
		statLines = append(statLines, fmt.Sprintf("%s  %s",
			statValueStyle.Render(s.value),
			statLabelStyle.Render(s.label)))
	}
	statsBlock := lipgloss.JoinHorizontal(lipgloss.Top, statLines[0], "    ", statLines[1], "    ", statLines[2], "    ", statLines[3])

	heat := renderHeatmap(d.DailyCounts, 26)

	var out strings.Builder
	out.WriteString(statsBlock)
	out.WriteString("\n\n")
	out.WriteString(sectionTitleStyle.Render("Contribution activity"))
	out.WriteString("\n\n")
	out.WriteString(heat)
	out.WriteString("\n\n")
	out.WriteString(renderLegend())
	return out.String()
}

func (m Model) viewActivity() string {
	d := m.data
	weeks := 24
	vel := d.WeeklyVelocity(weeks)

	var out strings.Builder
	out.WriteString(sectionTitleStyle.Render(fmt.Sprintf("Commit velocity (last %d weeks)", weeks)))
	out.WriteString("\n\n")
	out.WriteString(renderSparkline(vel))
	out.WriteString("\n\n")

	total := 0
	for _, v := range vel {
		total += v
	}
	avg := float64(total) / float64(weeks)
	out.WriteString(subtitleStyle.Render(fmt.Sprintf("avg %.1f commits/week", avg)))
	out.WriteString("\n\n")

	out.WriteString(sectionTitleStyle.Render("Recent commits"))
	out.WriteString("\n\n")
	n := 10
	if len(d.Commits) < n {
		n = len(d.Commits)
	}
	for _, c := range d.Commits[:n] {
		age := formatAge(c.Date)
		subject := c.Subject
		if len(subject) > 60 {
			subject = subject[:59] + "…"
		}
		line := fmt.Sprintf("%s %s %s",
			lipgloss.NewStyle().Foreground(subtle).Width(10).Render(age),
			lipgloss.NewStyle().Foreground(accent).Render(c.Author),
			subject)
		out.WriteString(line)
		out.WriteString("\n")
	}
	return out.String()
}

func (m Model) viewLanguages() string {
	d := m.data
	if len(d.Languages) == 0 {
		return subtitleStyle.Render("No recognized source files found.")
	}
	labels, values := sortedLangKeys(d.Languages)

	var total int64
	for _, v := range values {
		total += v
	}

	var out strings.Builder
	out.WriteString(sectionTitleStyle.Render("Language breakdown (by bytes)"))
	out.WriteString("\n\n")
	out.WriteString(renderBarChart(labels, values, 30))
	return out.String()
}

func (m Model) viewContributors() string {
	d := m.data
	top := d.TopContributors(8)
	labels := make([]string, len(top))
	values := make([]int, len(top))
	for i, c := range top {
		labels[i] = c.Name
		values[i] = c.Count
	}

	var out strings.Builder
	out.WriteString(sectionTitleStyle.Render("Top contributors"))
	out.WriteString("\n\n")
	out.WriteString(renderIntBarChart(labels, values, 30))
	return out.String()
}

func formatAge(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	default:
		return t.Format("Jan 2, 2006")
	}
}
