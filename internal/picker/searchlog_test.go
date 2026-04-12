package picker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLogSearchRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "searches.log")

	if err := LogSearch(path, "rocket", "🚀"); err != nil {
		t.Fatalf("LogSearch: %v", err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading log: %v", err)
	}
	line := strings.TrimSpace(string(b))
	parts := strings.Split(line, "\t")
	if len(parts) != 3 {
		t.Fatalf("got %d fields, want 3: %q", len(parts), line)
	}
	if parts[1] != "rocket" {
		t.Errorf("query = %q, want %q", parts[1], "rocket")
	}
	if parts[2] != "🚀" {
		t.Errorf("selected = %q, want %q", parts[2], "🚀")
	}
}

func TestLogSearchCancelled(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "searches.log")

	if err := LogSearch(path, "test query", ""); err != nil {
		t.Fatalf("LogSearch: %v", err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading log: %v", err)
	}
	line := strings.TrimSpace(string(b))
	parts := strings.Split(line, "\t")
	if len(parts) != 3 {
		t.Fatalf("got %d fields, want 3: %q", len(parts), line)
	}
	if parts[2] != "(cancelled)" {
		t.Errorf("status = %q, want %q", parts[2], "(cancelled)")
	}
}

func TestLogSearchAppends(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "searches.log")

	if err := LogSearch(path, "first", "😀"); err != nil {
		t.Fatalf("LogSearch 1: %v", err)
	}
	if err := LogSearch(path, "second", "🎉"); err != nil {
		t.Fatalf("LogSearch 2: %v", err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading log: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(b)), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2", len(lines))
	}
}

func TestLogSearchCreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "searches.log")

	if err := LogSearch(path, "q", "😀"); err != nil {
		t.Fatalf("LogSearch: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Errorf("log file not created: %v", err)
	}
}
