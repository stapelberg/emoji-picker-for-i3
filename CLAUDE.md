# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Test

```bash
go install ./cmd/emoji-picker-for-i3  # install binary to $GOBIN
go test ./...                          # run all tests
go test ./internal/picker/             # run tests for the picker package
go test ./internal/picker/ -run TestLoadEmojis  # run a single test
```

Nix: `nix build` produces the package. After `.nix` file changes: `nix fmt && nix build`.

Regenerate emoji data from upstream Unicode/CLDR:

```bash
go run ./cmd/generate-emoji -- -o internal/picker/_data/emoji.txt
```

## Architecture

On-demand CLI (`emoji-picker-for-i3`) invoked from an i3 keybinding. Shows a rofi menu of all Unicode emojis, types the selected one via xdotool.

**Key components:**
- `picker.go` — main entry point (`Run()`): loads emojis, state (recent/frecency), builds rofi input, handles selection (including numbered recent items), types result via xdotool.
- `emoji.go` — embeds `_data/emoji.txt` and `_data/keywords.txt`, parses and merges them into `[]Emoji`. Custom keywords extend CLDR annotations for better search.
- `frecency.go` — frequency-based sorting: tracks per-emoji use counts, persists atomically via renameio.
- `searchlog.go` — append-only log of searches (timestamp, query, selected emoji or cancelled).

**Code generator:** `cmd/generate-emoji/main.go` fetches Unicode emoji-test.txt and CLDR annotations to regenerate `_data/emoji.txt`.

**External tools** (wrapped into PATH by Nix): `rofi`, `xdotool`.
