package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var heatLevels = []lipgloss.Color{
	"#1a1a2e",
	"#2d5a3d",
	"#3f8f5c",
	"#52c17d",
	"#6fe89e",
}

func heatColor(count, max int) lipgloss.Color {
	if count == 0 || max == 0 {
		return heatLevels[0]
	}
	ratio := float64(count) / float64(max)
	switch {
	case ratio > 0.75:
		return heatLevels[4]
	case ratio > 0.5:
		return heatLevels[3]
	case ratio > 0.25:
		return heatLevels[2]
	default:
		return heatLevels[1]
	}
}

func renderHeatmap(daily map[string]int, weeks int) string {
	now := time.Now()
	end := now
	for end.Weekday() != time.Saturday {
		end = end.AddDate(0, 0, 1)
	}
	start := end.AddDate(0, 0, -weeks*7+1)

	max := 0
	for _, v := range daily {
		if v > max {
			max = v
		}
	}

	cols := make([][]string, weeks)
	monthLabels := make([]string, weeks)
	lastMonth := ""
	day := start
	for w := 0; w < weeks; w++ {
		col := make([]string, 7)
		for d := 0; d < 7; d++ {
			key := day.Format("2006-01-02")
			if day.After(now) {
				col[d] = lipgloss.NewStyle().Foreground(lipgloss.Color("#0d0d15")).Render("  ")
			} else {
				c := daily[key]
				style := lipgloss.NewStyle().Foreground(heatColor(c, max))
				col[d] = style.Render("██")
			}
			day = day.AddDate(0, 0, 1)
		}
		cols[w] = col
		m := day.AddDate(0, 0, -7).Format("Jan")
		if m != lastMonth {
			monthLabels[w] = m
			lastMonth = m
		}
	}

	var monthRow strings.Builder
	for _, m := range monthLabels {
		if m == "" {
			monthRow.WriteString("  ")
		} else {
			monthRow.WriteString(lipgloss.NewStyle().Foreground(subtle).Render(fmt.Sprintf("%-2s", m)))
		}
	}

	dayLabels := []string{"", "Mon", "", "Wed", "", "Fri", ""}
	var rows []string
	for d := 0; d < 7; d++ {
		var row strings.Builder
		label := lipgloss.NewStyle().Foreground(subtle).Width(4).Render(dayLabels[d])
		row.WriteString(label)
		for w := 0; w < weeks; w++ {
			row.WriteString(cols[w][d])
		}
		rows = append(rows, row.String())
	}

	grid := "    " + monthRow.String() + "\n" + strings.Join(rows, "\n")
	return grid
}

func renderLegend() string {
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Foreground(subtle).Render("less "))
	for _, c := range heatLevels {
		b.WriteString(lipgloss.NewStyle().Foreground(c).Render("██"))
	}
	b.WriteString(lipgloss.NewStyle().Foreground(subtle).Render(" more"))
	return b.String()
}
