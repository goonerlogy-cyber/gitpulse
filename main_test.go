package main

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
)

func TestJSONCommandWorksWithoutTerminal(t *testing.T) {
	dir := t.TempDir()
	if out, err := exec.Command("git", "-C", dir, "init", "--quiet").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, out)
	}
	var stdout, stderr bytes.Buffer
	if err := run([]string{"--json", dir}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if !json.Valid(stdout.Bytes()) || stderr.Len() != 0 || strings.Contains(stdout.String(), "\x1b") {
		t.Fatalf("unexpected output: stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

func TestHelpDoesNotLoadRepository(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := run([]string{"--help"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "--json") {
		t.Fatal("missing JSON usage")
	}
}

func TestInvalidArguments(t *testing.T) {
	for _, args := range [][]string{{"--unknown"}, {"--json", "one", "two"}} {
		var stdout, stderr bytes.Buffer
		if err := run(args, &stdout, &stderr); err == nil {
			t.Fatalf("expected error for %v", args)
		}
	}
}
