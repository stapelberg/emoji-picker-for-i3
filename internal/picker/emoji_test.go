package picker

import (
	"strings"
	"testing"
)

func TestParseLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		wantChar string
		wantName string
		wantTags []string
	}{
		{
			name:     "emoji with tags",
			line:     "😀 grinning face <small>(cheerful, face, grin)</small>",
			wantChar: "😀",
			wantName: "grinning face",
			wantTags: []string{"cheerful", "face", "grin"},
		},
		{
			name:     "emoji without tags",
			line:     "☺️ smiling face",
			wantChar: "☺️",
			wantName: "smiling face",
			wantTags: nil,
		},
		{
			name:     "ZWJ sequence without tags",
			line:     "😶\u200d🌫️ face in clouds",
			wantChar: "😶\u200d🌫️",
			wantName: "face in clouds",
			wantTags: nil,
		},
		{
			name:     "single tag",
			line:     "🤖 robot <small>(face)</small>",
			wantChar: "🤖",
			wantName: "robot",
			wantTags: []string{"face"},
		},
		{
			name:     "empty line",
			line:     "",
			wantChar: "",
			wantName: "",
			wantTags: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseLine(tt.line)
			if got.Char != tt.wantChar {
				t.Errorf("Char = %q, want %q", got.Char, tt.wantChar)
			}
			if got.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", got.Name, tt.wantName)
			}
			if len(got.Tags) != len(tt.wantTags) {
				t.Fatalf("Tags = %v (len %d), want %v (len %d)", got.Tags, len(got.Tags), tt.wantTags, len(tt.wantTags))
			}
			for i, tag := range got.Tags {
				if tag != tt.wantTags[i] {
					t.Errorf("Tags[%d] = %q, want %q", i, tag, tt.wantTags[i])
				}
			}
		})
	}
}

func TestParseKeywords(t *testing.T) {
	// parseKeywords operates on the embedded keywordsData, so we test
	// its output indirectly through known entries in keywords.txt.
	kw := parseKeywords()

	// Single-word keywords
	thumbs, ok := kw["👍"]
	if !ok {
		t.Fatal("👍 not found in keywords")
	}
	for _, want := range []string{"lgtm", "approve", "ack"} {
		found := false
		for _, k := range thumbs {
			if k == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("👍 missing keyword %q, got %v", want, thumbs)
		}
	}

	// Multi-word keywords
	clap, ok := kw["👏"]
	if !ok {
		t.Fatal("👏 not found in keywords")
	}
	found := false
	for _, k := range clap {
		if k == "well done" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("👏 missing multi-word keyword \"well done\", got %v", clap)
	}

	// Comment lines and blanks should be skipped (no key starting with #)
	for k := range kw {
		if strings.HasPrefix(k, "#") {
			t.Errorf("comment leaked as keyword key: %q", k)
		}
	}
}

func TestLoadEmojisSkipsComments(t *testing.T) {
	emojis := LoadEmojis()
	for _, e := range emojis {
		if strings.HasPrefix(e.Char, "#") {
			t.Errorf("comment line parsed as emoji: %+v", e)
		}
	}
	if len(emojis) == 0 {
		t.Fatal("no emojis loaded")
	}
}

func TestLoadEmojisPreservesExistingTags(t *testing.T) {
	emojis := LoadEmojis()
	// Find 😅 grinning face with sweat — has CLDR tags like "nervous", "sweat"
	// and keywords.txt adds "embarrassed", "oops", "relief"
	for _, e := range emojis {
		if e.Char == "😅" {
			// CLDR tags must still be present
			for _, want := range []string{"nervous", "sweat"} {
				found := false
				for _, tag := range e.Tags {
					if tag == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("😅 missing CLDR tag %q, tags: %v", want, e.Tags)
				}
			}
			// Custom keywords must be merged in
			for _, want := range []string{"embarrassed", "oops", "relief"} {
				found := false
				for _, tag := range e.Tags {
					if tag == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("😅 missing keyword %q, tags: %v", want, e.Tags)
				}
			}
			return
		}
	}
	t.Fatal("😅 not found in emoji list")
}

func TestMultiWordKeywords(t *testing.T) {
	emojis := LoadEmojis()
	for _, e := range emojis {
		if e.Char == "👏" {
			// "well done" is a multi-word keyword that was broken before the fix
			found := false
			for _, tag := range e.Tags {
				if tag == "well done" {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("👏 missing multi-word keyword \"well done\", tags: %v", e.Tags)
			}
			// "kudos" comes after the space in "well done,kudos" —
			// it was lost before the fix
			found = false
			for _, tag := range e.Tags {
				if tag == "kudos" {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("👏 missing keyword \"kudos\" (was lost due to space-split bug), tags: %v", e.Tags)
			}
			return
		}
	}
	t.Fatal("👏 not found in emoji list")
}

func TestEmotionalKeywordsSearchable(t *testing.T) {
	emojis := LoadEmojis()

	// Build a map of emotion word -> emojis that match
	tests := []struct {
		query   string
		wantAny []string // at least one of these emojis should have the keyword
	}{
		{"embarrassed", []string{"😅", "🫢", "🫣", "🙈", "🙊"}},
		{"proud", []string{"😤", "🥲"}},
		{"sarcastic", []string{"🙃", "🙄"}},
		{"cringe", []string{"😳", "🫠", "🫣", "🙈", "😬"}},
		{"grateful", []string{"😌", "🥲", "🫶"}},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			found := false
			for _, e := range emojis {
				line := e.RofiLine()
				if strings.Contains(strings.ToLower(line), tt.query) {
					for _, want := range tt.wantAny {
						if e.Char == want {
							found = true
							break
						}
					}
				}
				if found {
					break
				}
			}
			if !found {
				t.Errorf("searching %q did not match any of %v", tt.query, tt.wantAny)
			}
		})
	}
}
