package picker

import (
	"bytes"
	"cmp"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"slices"
	"strings"
)

func defaultDataDir() string {
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		log.Printf("UserHomeDir: %v, falling back to $HOME", err)
		home = os.Getenv("HOME")
	}
	if home == "" {
		log.Fatal("cannot determine home directory: set $HOME or $XDG_DATA_HOME")
	}
	return filepath.Join(home, ".local", "share")
}

// Run is the main entry point for the emoji picker.
func Run() error {
	var (
		recentPath = flag.String("recent_path",
			filepath.Join(defaultDataDir(), "rofimoji", "recent"),
			"path to recent emoji file")

		frequencyPath = flag.String("frequency_path",
			filepath.Join(defaultDataDir(), "rofimoji", "frecency"),
			"path to frequency file")

		logDir = flag.String("log_dir",
			filepath.Join(defaultDataDir(), "emoji-picker-for-i3"),
			"directory for log and search log files")

		dpi = flag.String("dpi", "288", "DPI for rofi")

		version = flag.Bool("version", false, "print version and exit")
	)
	flag.Parse()

	if *version {
		fmt.Printf("emoji-picker-for-i3 %s\n", vcsRevision())
		return nil
	}

	// Set up logging to file + stderr
	if err := os.MkdirAll(*logDir, 0755); err != nil {
		log.Printf("creating log dir: %v", err)
	} else {
		logPath := filepath.Join(*logDir, "picker.log")
		f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			log.Printf("opening log file: %v", err)
		} else {
			defer f.Close()
			log.SetOutput(io.MultiWriter(os.Stderr, f))
		}
	}

	emojis := LoadEmojis()
	log.Printf("emoji-picker-for-i3 %s started, %d emojis loaded", vcsRevision(), len(emojis))

	// Load state
	recents := LoadRecent(*recentPath)
	frequencies := LoadFrecency(*frequencyPath)

	// Sort emojis by frequency bucket (most-used first), stable to preserve
	// emoji.txt order within each bucket. Bucketing by order of magnitude
	// keeps positional muscle memory: a few extra uses no longer reorder
	// the list — an emoji has to be used roughly 10x more often to win.
	slices.SortStableFunc(emojis, func(a, b Emoji) int {
		return cmp.Compare(
			frequencyBucket(frequencies[b.Char]),
			frequencyBucket(frequencies[a.Char]),
		)
	})

	// Build recent pairings for -mesg display
	pairings := make([]string, len(recents))
	for idx, recent := range recents {
		pairings[idx] = fmt.Sprintf("%c: %s", '0'+((idx+1)%10), recent)
	}

	// Build rofi input: numbered recent items first, then all emojis
	var rofiInput strings.Builder
	for idx, recent := range recents {
		digit := (idx + 1) % 10
		// Find this emoji's name for display
		name := recent
		for _, e := range emojis {
			if e.Char == recent {
				name = recent + " " + e.Name
				break
			}
		}
		fmt.Fprintf(&rofiInput, "%d  %s\n", digit, name)
	}
	for _, e := range emojis {
		rofiInput.WriteString(e.RofiLine() + "\n")
	}

	ctx := context.Background()

	args := []string{
		"-dmenu",
		"-dpi", *dpi,
		"-markup-rows",
		"-matching", "normal",
		"-no-custom", // prevent typing arbitrary text
		"-p", "emoji",
		"-format", "f\tp", // filter query + selected (stripped of pango)
	}
	if len(pairings) > 0 {
		args = append(args, "-mesg", strings.Join(pairings, " | "))
	}
	rofi := exec.CommandContext(ctx, "rofi", args...)
	var stdout bytes.Buffer
	rofi.Stdout = &stdout
	w, err := rofi.StdinPipe()
	if err != nil {
		return err
	}
	if err := rofi.Start(); err != nil {
		return err
	}

	go func() {
		defer w.Close()
		io.WriteString(w, rofiInput.String())
	}()

	searchLogPath := filepath.Join(*logDir, "searches.log")
	var char string
	var query string

	if err := rofi.Wait(); err != nil {
		ee, ok := errors.AsType[*exec.ExitError](err)
		if !ok {
			return err
		}
		if ee.ExitCode() == 1 {
			// User cancelled — try to log what they typed
			output := stdout.String()
			if parts := strings.SplitN(output, "\t", 2); len(parts) > 0 {
				query = parts[0]
			}
			if query != "" {
				log.Printf("cancelled, query=%q", query)
				LogSearch(searchLogPath, query, "")
			}
			return nil
		} else {
			return err
		}
	} else {
		output := stdout.String()
		parts := strings.SplitN(output, "\t", 2)
		if len(parts) > 0 {
			query = strings.TrimSpace(parts[0])
		}
		selected := ""
		if len(parts) > 1 {
			selected = strings.TrimSpace(parts[1])
		}

		// Extract the emoji character from the selected line
		fields := strings.Fields(selected)
		if len(fields) == 0 {
			return fmt.Errorf("no selection from rofi")
		}

		// Check if this is a numbered recent item (e.g. "1  🥳 party face")
		first := fields[0]
		if len(first) == 1 && first[0] >= '0' && first[0] <= '9' && len(fields) >= 2 {
			char = fields[1]
		} else {
			char = first
		}
	}

	log.Printf("selected %s, query=%q", char, query)

	// Type the emoji
	xdotool := exec.CommandContext(ctx, "xdotool", "type", "--clearmodifiers", char)
	xdotool.Stdout = os.Stdout
	xdotool.Stderr = os.Stderr
	if err := xdotool.Run(); err != nil {
		return err
	}

	// Log search
	if err := LogSearch(searchLogPath, query, char); err != nil {
		log.Printf("logging search: %v", err)
	}

	// Update frecency
	frequencies[char]++
	log.Printf("frecency updated: %s now at %d", char, frequencies[char])
	if err := SaveFrecency(*frequencyPath, frequencies); err != nil {
		log.Printf("saving frecency: %v", err)
	}

	// Update recent
	if err := SaveRecent(*recentPath, char, recents); err != nil {
		log.Printf("saving recent: %v", err)
	}

	return nil
}

func vcsRevision() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return "(no build info)"
	}
	var rev, t string
	var dirty bool
	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.time":
			t = s.Value
		case "vcs.modified":
			dirty = s.Value == "true"
		}
	}
	if rev == "" {
		return "(no vcs info)"
	}
	if len(rev) > 12 {
		rev = rev[:12]
	}
	if dirty {
		rev += "-dirty"
	}
	return rev + " " + t
}
