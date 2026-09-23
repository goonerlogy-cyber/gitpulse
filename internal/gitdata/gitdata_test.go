package gitdata

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	git(t, dir, "init", "--quiet")
	git(t, dir, "config", "user.name", "Test Author")
	git(t, dir, "config", "user.email", "test@example.invalid")
	return dir
}

func TestLoadFromSubdirectoryCountsExactTrackedNames(t *testing.T) {
	dir := initRepo(t)
	names := []string{" leading.go", "café.go", "nested/hello.go"}
	if runtime.GOOS != "windows" {
		names = append(names, "tab\tname.go", "line\nbreak.go", "trailing .go")
	}
	for _, name := range names {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("hello"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	git(t, dir, "add", "--all")
	git(t, dir, "commit", "--quiet", "--no-gpg-sign", "-m", "Add sample files")
	if err := os.WriteFile(filepath.Join(dir, "untracked.go"), []byte("ignored"), 0644); err != nil {
		t.Fatal(err)
	}
	data, err := Load(filepath.Join(dir, "nested"))
	if err != nil {
		t.Fatal(err)
	}
	gotRoot, err := os.Stat(data.RepoPath)
	if err != nil {
		t.Fatal(err)
	}
	wantRoot, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	// Git can expand a Windows 8.3 path (RUNNER~1) to its long spelling.
	if !os.SameFile(gotRoot, wantRoot) {
		t.Fatalf("repository = %q; want %q", data.RepoPath, dir)
	}
	if got, want := data.Languages["Go"], int64(5*len(names)); got != want {
		t.Fatalf("Go bytes = %d; want %d", got, want)
	}
	if len(data.Commits) != 1 || data.Contributors["Test Author"] != 1 {
		t.Fatalf("unexpected history: %+v", data)
	}
	var output bytes.Buffer
	if err := WriteReport(&output, data); err != nil {
		t.Fatal(err)
	}
	var report Report
	if err := json.Unmarshal(output.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.SchemaVersion != 1 || report.TotalCommits != 1 || report.FirstCommit == nil || report.LastCommit == nil {
		t.Fatalf("unexpected report: %+v", report)
	}
	if strings.Contains(output.String(), "test@example.invalid") || strings.Contains(output.String(), "Add sample files") {
		t.Fatal("report exposed raw commit details")
	}
}

func TestEmptyRepositoryReport(t *testing.T) {
	data, err := Load(initRepo(t))
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := WriteReport(&output, data); err != nil {
		t.Fatal(err)
	}
	var report map[string]any
	if err := json.Unmarshal(output.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report["total_commits"] != float64(0) || report["first_commit"] != nil || report["last_commit"] != nil {
		t.Fatalf("unexpected empty report: %s", output.String())
	}
	for _, key := range []string{"contributors", "daily_counts", "language_bytes"} {
		if fields, ok := report[key].(map[string]any); !ok || len(fields) != 0 {
			t.Fatalf("%s must be an empty object: %s", key, output.String())
		}
	}
}

func TestLoadIgnoresTrackedSymlinks(t *testing.T) {
	dir := initRepo(t)
	outside := filepath.Join(t.TempDir(), "outside.go")
	if err := os.WriteFile(outside, []byte("must not count outside files"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(dir, "linked.go")); err != nil {
		if runtime.GOOS == "windows" {
			t.Skip("Windows symlink creation unavailable")
		}
		t.Fatal(err)
	}
	git(t, dir, "add", "--all")
	data, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(data.Languages) != 0 {
		t.Fatalf("symlink counted as source: %v", data.Languages)
	}
}

func TestLoadRejectsNonRepository(t *testing.T) {
	if _, err := Load(t.TempDir()); err == nil {
		t.Fatal("expected an error for a non-repository")
	}
}

func TestLoadIgnoresSymlinkCheckedOutAsRegularFile(t *testing.T) {
	dir := initRepo(t)
	git(t, dir, "config", "core.symlinks", "false")
	cmd := exec.Command("git", "-C", dir, "hash-object", "-w", "--stdin")
	cmd.Stdin = strings.NewReader("elsewhere.go")
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	git(t, dir, "update-index", "--add", "--cacheinfo", "120000,"+strings.TrimSpace(string(out))+",linked.go")
	git(t, dir, "checkout-index", "--", "linked.go")
	info, err := os.Lstat(filepath.Join(dir, "linked.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !info.Mode().IsRegular() {
		t.Fatal("fixture must materialize a tracked symlink as a regular file")
	}
	data, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(data.Languages) != 0 {
		t.Fatalf("tracked symlink counted as source: %v", data.Languages)
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("output unavailable") }
func TestReportPropagatesOutputFailure(t *testing.T) {
	if err := WriteReport(failingWriter{}, &Data{}); err == nil {
		t.Fatal("expected a writer error")
	}
}
