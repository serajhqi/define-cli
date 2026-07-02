# define

Terminal dictionary — look up English word definitions without leaving the CLI.

## Quick install

```bash
curl -fSL https://github.com/serajhqi/define-cli/releases/latest/download/define-linux-amd64 -o ~/.local/bin/define && chmod +x ~/.local/bin/define
```

If `~/.local/bin` is not on your `PATH`, add it:

**bash** (`~/.bashrc`):
```bash
export PATH="$HOME/.local/bin:$PATH"
```

**fish** (`~/.config/fish/config.fish`):
```fish
fish_add_path ~/.local/bin
```

## Usage

```bash
define hello               # Look up "hello" (colored ANSI tree, exits after print)
define --play hello         # Look up "hello" and play pronunciation aloud
define --plain hello        # Plain text output, no ANSI colors
define -f hello             # Force fresh API call, skip cache
define                      # Interactive TUI history browser
```

In the TUI: type to fuzzy-filter your lookup history, `↑`/`↓` to navigate, `Enter` to pick a word, `f` to force-refresh, `d` to delete entries, `p` to play audio, `?` for full help.

Piped output is automatically plain text — no flag needed.

## Features

- **Cache-first.** Every lookup is saved to `~/.cache/define/cache.json`. Repeat lookups are instant and work offline.
- **Interactive TUI.** Browse your history with fuzzy search, scroll through definitions, play audio — all inside the terminal.
- **Audio pronunciation.** Auto-detects `paplay`, `ffplay`, or `mpv` on your system. Press `p` to hear the word.
- **Pipe-friendly.** Plain text output with `--plain`, or auto-detected when stdout is not a TTY.
- **Synonyms and antonyms.** Shown inline under each definition when available from the API.
- **Inflection highlighting.** Searched word and its inflections are bolded in context (e.g. "aim" highlights "aiming", "aimed", "aims").
- **Force refresh.** `-f` flag or `f` key bypasses cache and fetches fresh data from the API.

## Requirements

- Linux
- One of `paplay`, `ffplay`, or `mpv` for audio playback (optional — definitions work without it)

## Install from source

```bash
go install github.com/serajhqi/define-cli/cmd/define@latest
```
