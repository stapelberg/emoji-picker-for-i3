package picker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFrecencyRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "frecency")

	want := map[string]int64{
		"😀": 42,
		"🎉": 10,
		"👍": 1,
	}
	if err := SaveFrecency(path, want); err != nil {
		t.Fatalf("SaveFrecency: %v", err)
	}

	got := LoadFrecency(path)
	for k, v := range want {
		if got[k] != v {
			t.Errorf("frecency[%s] = %d, want %d", k, got[k], v)
		}
	}
	if len(got) != len(want) {
		t.Errorf("got %d entries, want %d", len(got), len(want))
	}
}

func TestFrecencyEmptyMap(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "frecency")

	if err := SaveFrecency(path, map[string]int64{}); err != nil {
		t.Fatalf("SaveFrecency: %v", err)
	}
	got := LoadFrecency(path)
	if len(got) != 0 {
		t.Errorf("got %d entries, want 0", len(got))
	}
}

func TestFrecencyMissingFile(t *testing.T) {
	got := LoadFrecency(filepath.Join(t.TempDir(), "nonexistent"))
	if len(got) != 0 {
		t.Errorf("got %d entries from missing file, want 0", len(got))
	}
}

func TestFrecencyCreatesDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "dir", "frecency")

	if err := SaveFrecency(path, map[string]int64{"😀": 1}); err != nil {
		t.Fatalf("SaveFrecency with nested dir: %v", err)
	}
	got := LoadFrecency(path)
	if got["😀"] != 1 {
		t.Errorf("frecency[😀] = %d, want 1", got["😀"])
	}
}

func TestRecentRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "recent")

	if err := SaveRecent(path, "🎉", []string{"😀", "👍"}); err != nil {
		t.Fatalf("SaveRecent: %v", err)
	}

	got := LoadRecent(path)
	want := []string{"🎉", "😀", "👍"}
	if len(got) != len(want) {
		t.Fatalf("got %d recent, want %d", len(got), len(want))
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("recent[%d] = %q, want %q", i, got[i], w)
		}
	}
}

func TestRecentDeduplication(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "recent")

	if err := SaveRecent(path, "😀", []string{"🎉", "😀", "👍"}); err != nil {
		t.Fatalf("SaveRecent: %v", err)
	}

	got := LoadRecent(path)
	want := []string{"😀", "🎉", "👍"}
	if len(got) != len(want) {
		t.Fatalf("got %d recent, want %d", len(got), len(want))
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("recent[%d] = %q, want %q", i, got[i], w)
		}
	}
}

func TestRecentCapsAt10(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "recent")

	existing := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"}
	if err := SaveRecent(path, "new", existing); err != nil {
		t.Fatalf("SaveRecent: %v", err)
	}

	got := LoadRecent(path)
	if len(got) != 10 {
		t.Errorf("got %d recent, want 10 (cap)", len(got))
	}
	if got[0] != "new" {
		t.Errorf("recent[0] = %q, want %q", got[0], "new")
	}
}

func TestRecentMissingFile(t *testing.T) {
	got := LoadRecent(filepath.Join(t.TempDir(), "nonexistent"))
	if got != nil {
		t.Errorf("got %v from missing file, want nil", got)
	}
}

func TestRecentCreatesDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "dir", "recent")

	if err := SaveRecent(path, "😀", nil); err != nil {
		t.Fatalf("SaveRecent with nested dir: %v", err)
	}
	got := LoadRecent(path)
	if len(got) != 1 || got[0] != "😀" {
		t.Errorf("got %v, want [😀]", got)
	}
}

func TestFrequencyBucket(t *testing.T) {
	for _, tt := range []struct {
		count int64
		want  int
	}{
		{-5, 0},
		{0, 0},
		{1, 1},
		{9, 1},
		{10, 2},
		{99, 2},
		{100, 3},
		{999, 3},
		{1000, 4},
		{9999, 4},
		{10000, 5},
	} {
		if got := frequencyBucket(tt.count); got != tt.want {
			t.Errorf("frequencyBucket(%d) = %d, want %d", tt.count, got, tt.want)
		}
	}
}

func TestFrecencyCorruptLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "frecency")

	// Write a file with some corrupt lines mixed in
	data := "42 😀\nbadline\n\n10 🎉\nnot_a_number 👍\n"
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	got := LoadFrecency(path)
	if got["😀"] != 42 {
		t.Errorf("frecency[😀] = %d, want 42", got["😀"])
	}
	if got["🎉"] != 10 {
		t.Errorf("frecency[🎉] = %d, want 10", got["🎉"])
	}
	if _, ok := got["👍"]; ok {
		t.Error("corrupt line should have been skipped")
	}
}
