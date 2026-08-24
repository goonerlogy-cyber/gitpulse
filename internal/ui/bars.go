package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var barPalette = []lipgloss.Color{
	"#6fe89e", "#52c1e8", "#c17de8", "#e8b552", "#e85287", "#52e8d1", "#e87d52", "#8fe852",
}

var sparkChars = []rune("▁▂▃▄▅▆▇█")

func renderSparkline(values []int) string {
	max := 0
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	var b strings.Builder
	for _, v := range values {
		if max == 0 {
			b.WriteRune(sparkChars[0])
			continue
		}
		idx := v * (len(sparkChars) - 1) / max
		b.WriteRune(sparkChars[idx])
	}
	return lipgloss.NewStyle().Foreground(accent).Render(b.String())
}

func renderBarChart(labels []string, values []int64, width int) string {
	max := int64(0)
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	labelWidth := 0
	for _, l := range labels {
		if len(l) > labelWidth {
			labelWidth = len(l)
		}
	}
	if labelWidth > 14 {
		labelWidth = 14
	}

	var rows []string
	for i, l := range labels {
		barLen := 0
		if max > 0 {
			barLen = int(values[i] * int64(width) / max)
		}
		if barLen == 0 && values[i] > 0 {
			barLen = 1
		}
		color := barPalette[i%len(barPalette)]
		bar := lipgloss.NewStyle().Foreground(color).Render(strings.Repeat("█", barLen))
		pad := strings.Repeat(" ", width-barLen)
		labelStr := l
		if len(labelStr) > labelWidth {
			labelStr = labelStr[:labelWidth-1] + "…"
		}
		label := lipgloss.NewStyle().Width(labelWidth).Render(labelStr)
		rows = append(rows, fmt.Sprintf("%s %s%s %s", label, bar, pad, lipgloss.NewStyle().Foreground(subtle).Render(fmt.Sprintf("%d", values[i]))))
	}
	return strings.Join(rows, "\n")
}

func renderIntBarChart(labels []string, values []int, width int) string {
	v64 := make([]int64, len(values))
	for i, v := range values {
		v64[i] = int64(v)
	}
	return renderBarChart(labels, v64, width)
}

func sortedLangKeys(langs map[string]int64) ([]string, []int64) {
	type kv struct {
		K string
		V int64
	}
	var list []kv
	for k, v := range langs {
		list = append(list, kv{k, v})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].V > list[j].V })
	if len(list) > 8 {
		list = list[:8]
	}
	labels := make([]string, len(list))
	values := make([]int64, len(list))
	for i, kv := range list {
		labels[i] = kv.K
		values[i] = kv.V
	}
	return labels, values
}
