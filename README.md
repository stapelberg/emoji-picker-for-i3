# emoji-picker-for-i3

A fast emoji picker for i3 that uses rofi for selection and xdotool to type
the chosen emoji.

Includes 1900+ Unicode 16.0 emojis with CLDR annotations and custom search
keywords for emotional/contextual searches (e.g. "embarrassed", "celebrate",
"lgtm").

## Try it

```bash
nix run github:stapelberg/emoji-picker-for-i3
```

This fetches and runs the picker without installing anything — `rofi` and
`xdotool` are wrapped in automatically. Bind it to a key in your i3 config
once you like it (see [Installation](#installation)).

## Features

- Frecency-based sorting: emojis you use often float to the top
- Recent emojis shown in the rofi message bar for quick number-key selection
- Custom keywords file for better search (searchable by emotion, reaction, etc.)
- Search log for analyzing usage patterns
- Atomic state file writes (no corruption on crash)

## How it works

`emoji-picker-for-i3` is a short-lived CLI launched on each keypress. The flow:

1. **Load emoji data.** The full list is embedded in the binary at build time
   from `internal/picker/_data/emoji.txt` (Unicode + CLDR annotations) and
   `keywords.txt` (custom synonyms).
2. **Load state** from `$XDG_DATA_HOME`: a frecency file (per-emoji use counts)
   and a recent file (last N picks). Both are written atomically via renameio.
3. **Sort** emojis by frequency *bucket* (order of magnitude), stable to keep
   the underlying emoji.txt order within each bucket. Bucketing preserves
   positional muscle memory — an emoji has to be used ~10× more often to move
   up.
4. **Spawn rofi** in `-dmenu` mode, piping the sorted list on stdin. Recent
   picks are rendered in rofi's `-mesg` bar prefixed with digits 1–9,0 so you
   can grab them by number key.
5. **Read rofi's selection** (`-format "f\tp"` gives filter query + picked
   line), extract the emoji character, and pipe it to `xdotool type
   --clearmodifiers` which injects it into the focused X11 window.
6. **Update state**: bump the frecency counter, push onto the recent list,
   append a line to the search log (timestamp, query, selected emoji — or
   blank for cancellations).

So rofi provides the UI + fuzzy matching, xdotool is the keyboard-injection
backend, and the Go program is the glue that owns the data, sorting, and
persistent state.

## Installation

### NixOS (recommended)

Add the flake input in your machine's `flake.nix`:

```nix
inputs.emoji-picker-for-i3 = {
  url = "github:stapelberg/emoji-picker-for-i3";
  inputs.nixpkgs.follows = "nixpkgs";
};
```

Pass it through to your `nixosConfigurations` and add the module:

```nix
modules = [
  emoji-picker-for-i3.nixosModules.default
  # ...
];
```

Then enable the program in your machine config:

```nix
programs.emoji-picker-for-i3.enable = true;
```

### From source

Requires Go 1.26+, rofi, and xdotool.

```bash
go install github.com/stapelberg/emoji-picker-for-i3/cmd/emoji-picker-for-i3@latest
```

Make sure `rofi` and `xdotool` are in your `PATH`.

## Keybinding

Add to your i3 config:

```
bindsym $mod+period exec emoji-picker-for-i3
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-dpi` | `288` | DPI for rofi |
| `-recent_path` | `$XDG_DATA_HOME/rofimoji/recent` | Path to recent emoji file |
| `-frequency_path` | `$XDG_DATA_HOME/rofimoji/frecency` | Path to frequency file |
| `-log_dir` | `$XDG_DATA_HOME/emoji-picker-for-i3` | Directory for log files |

## Custom keywords

Edit `internal/picker/_data/keywords.txt` to add search keywords. Format:

```
👍 ok,yes,approve,agreed,lgtm,+1
```

Keywords are merged with CLDR annotations at build time.

## Regenerating emoji data

To update the emoji database from upstream Unicode/CLDR sources:

```bash
go run ./cmd/generate-emoji -- -o internal/picker/_data/emoji.txt
```

## License

0BSD. See [LICENSE](LICENSE).

Emoji data is derived from Unicode sources under the Unicode License Agreement.
See [NOTICE](NOTICE).
