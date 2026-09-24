package gitdata

import (
	"encoding/json"
	"io"
	"time"
)

// Report is the versioned, non-interactive view of the dashboard dataset.
// It excludes commit subjects and email addresses; contributors are author names.
type Report struct {
	SchemaVersion int              `json:"schema_version"`
	Repository    string           `json:"repository"`
	Name          string           `json:"name"`
	TotalCommits  int              `json:"total_commits"`
	Contributors  map[string]int   `json:"contributors"`
	DailyCounts   map[string]int   `json:"daily_counts"`
	LanguageBytes map[string]int64 `json:"language_bytes"`
	FirstCommit   *time.Time       `json:"first_commit"`
	LastCommit    *time.Time       `json:"last_commit"`
}

// WriteReport writes one JSON object followed by a newline, without terminal codes.
func WriteReport(w io.Writer, data *Data) error {
	report := Report{
		SchemaVersion: 1,
		Repository:    data.RepoPath,
		Name:          data.RepoName,
		TotalCommits:  len(data.Commits),
		Contributors:  data.Contributors,
		DailyCounts:   data.DailyCounts,
		LanguageBytes: data.Languages,
	}
	if len(data.Commits) > 0 {
		report.FirstCommit = &data.FirstCommit
		report.LastCommit = &data.LastCommit
	}
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}
