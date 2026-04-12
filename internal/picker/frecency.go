package picker

import (
	"cmp"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/google/renameio/v2"
)

// frequencyBucket returns the order-of-magnitude bucket of count.
// 0 -> 0, 1..9 -> 1, 10..99 -> 2, 100..999 -> 3, etc. Negative counts
// are treated as 0 defensively. Sorting by bucket (instead of raw count)
// keeps the all-time-favorites order stable: an emoji needs roughly 10x
// more uses than another to swap positions.
func frequencyBucket(count int64) int {
	if count <= 0 {
		return 0
	}
	b := 0
	for count > 0 {
		b++
		count /= 10
	}
	return b
}

// LoadFrecency reads the frecency file and returns emoji→count mapping.
func LoadFrecency(path string) map[string]int64 {
	freq := make(map[string]int64)
	b, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("reading frecency: %v", err)
		}
		return freq
	}
	for _, line := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		n, err := strconv.ParseInt(parts[0], 0, 64)
		if err != nil {
			log.Printf("parsing frecency line %q: %v", line, err)
			continue
		}
		freq[parts[1]] = n
	}
	return freq
}

// SaveFrecency writes the frecency file atomically, sorted by count descending.
func SaveFrecency(path string, freq map[string]int64) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating frecency dir: %w", err)
	}

	// Sort by count descending
	type entry struct {
		char  string
		count int64
	}
	entries := make([]entry, 0, len(freq))
	for ch, n := range freq {
		entries = append(entries, entry{ch, n})
	}
	slices.SortFunc(entries, func(a, b entry) int {
		return cmp.Compare(b.count, a.count) // descending
	})

	pending, err := renameio.NewPendingFile(path)
	if err != nil {
		return fmt.Errorf("creating pending frecency file: %w", err)
	}
	defer pending.Cleanup()

	for _, e := range entries {
		if _, err := fmt.Fprintf(pending, "%d %s\n", e.count, e.char); err != nil {
			return fmt.Errorf("writing frecency entry: %w", err)
		}
	}
	return pending.CloseAtomicallyReplace()
}

// LoadRecent reads the recent emojis file (one emoji per line).
func LoadRecent(path string) []string {
	b, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("reading recent: %v", err)
		}
		return nil
	}
	var recent []string
	for _, line := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		if line != "" {
			recent = append(recent, line)
		}
	}
	return recent
}

// SaveRecent writes the recent emojis file atomically.
// Prepends char, deduplicates, and caps at 10.
func SaveRecent(path string, char string, existing []string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating recent dir: %w", err)
	}

	recent := []string{char}
	for _, r := range existing {
		if r != char && len(recent) < 10 {
			recent = append(recent, r)
		}
	}

	pending, err := renameio.NewPendingFile(path)
	if err != nil {
		return fmt.Errorf("creating pending recent file: %w", err)
	}
	defer pending.Cleanup()

	for _, r := range recent {
		if _, err := fmt.Fprintln(pending, r); err != nil {
			return fmt.Errorf("writing recent entry: %w", err)
		}
	}
	return pending.CloseAtomicallyReplace()
}
