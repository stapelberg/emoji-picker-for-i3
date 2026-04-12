package picker

import (
	_ "embed"
	"html"
	"strings"
)

//go:embed _data/emoji.txt
var emojiData []byte

//go:embed _data/keywords.txt
var keywordsData []byte

// Emoji represents a single emoji with its metadata.
type Emoji struct {
	Char string
	Name string
	Tags []string
}

// RofiLine formats the emoji for rofi display with pango markup.
// Names and tags are XML-escaped to avoid breaking pango parsing
// (e.g. "Antigua & Barbuda", tag "<3").
func (e Emoji) RofiLine() string {
	name := html.EscapeString(e.Name)
	if len(e.Tags) == 0 {
		return e.Char + " " + name
	}
	escaped := make([]string, len(e.Tags))
	for i, t := range e.Tags {
		escaped[i] = html.EscapeString(t)
	}
	return e.Char + " " + name + " <small>(" + strings.Join(escaped, ", ") + ")</small>"
}

// LoadEmojis parses the embedded emoji data and merges custom keywords.
func LoadEmojis() []Emoji {
	keywords := parseKeywords()
	var emojis []Emoji
	for _, line := range strings.Split(strings.TrimSpace(string(emojiData)), "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		e := parseLine(line)
		if extra, ok := keywords[e.Char]; ok {
			e.Tags = append(e.Tags, extra...)
		}
		emojis = append(emojis, e)
	}
	return emojis
}

// parseLine parses a single emoji.txt line:
//
//	🎃 jack-o-lantern <small>(celebration, halloween, jack)</small>
func parseLine(line string) Emoji {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return Emoji{}
	}
	char := fields[0]

	// Find name (text before <small>) and tags (inside parentheses in <small>)
	rest := strings.TrimSpace(strings.TrimPrefix(line, char))
	name := rest
	var tags []string

	if idx := strings.Index(rest, "<small>"); idx >= 0 {
		name = strings.TrimSpace(rest[:idx])
		tagStr := rest[idx:]
		tagStr = strings.TrimPrefix(tagStr, "<small>")
		tagStr = strings.TrimSuffix(tagStr, "</small>")
		tagStr = strings.TrimPrefix(tagStr, "(")
		tagStr = strings.TrimSuffix(tagStr, ")")
		for _, t := range strings.Split(tagStr, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				tags = append(tags, t)
			}
		}
	}

	return Emoji{
		Char: char,
		Name: name,
		Tags: tags,
	}
}

// parseKeywords parses the keywords.txt file.
// Format: emoji_char keyword1,keyword2,keyword3
func parseKeywords() map[string][]string {
	m := make(map[string][]string)
	for _, line := range strings.Split(strings.TrimSpace(string(keywordsData)), "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		char := fields[0]
		for _, kw := range strings.Split(strings.Join(fields[1:], " "), ",") {
			kw = strings.TrimSpace(kw)
			if kw != "" {
				m[char] = append(m[char], kw)
			}
		}
	}
	return m
}
