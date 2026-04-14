# Thakkali

**Thakkali** (Tamil for "tomato") is a terminal Pomodoro timer with a Ghostty-inspired ASCII animation. Live in a spare terminal window while you work.

```
████████╗██╗  ██╗ █████╗ ██╗  ██╗██╗  ██╗ █████╗ ██╗     ██╗
╚══██╔══╝██║  ██║██╔══██╗██║ ██╔╝██║ ██╔╝██╔══██╗██║     ██║
   ██║   ███████║███████║█████╔╝ █████╔╝ ███████║██║     ██║
   ██║   ██╔══██║██╔══██║██╔═██╗ ██╔═██╗ ██╔══██║██║     ██║
   ██║   ██║  ██║██║  ██║██║  ██╗██║  ██╗██║  ██║███████╗██║
   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚══════╝╚═╝
```

## Features

- **Simple timer by default** — pick a duration, hit go. No ceremony.
- **Opt-in Pomodoro mode** (`-pomodoro` / `-p`) — work, short break, long break, configurable rounds
- **Animated tomato** — rolls continuously with a shimmering outer ring, blinking eyes, and a periodic 360° spin-and-jump trick (Ghostty-style layered ASCII)
- **Big ANSI Shadow digits** for the timer, matching the logo font
- **Cross-platform desktop notifications + beep** when a phase ends (macOS system sounds supported)
- **Task tagging** per work session, logged for later review
- **JSON config file** auto-created on first run
- **JSON-lines session log** for stats, scripting, or export
- **Minimal mode** — hide the logo and animation when you just want the timer

## Install

### From source

Requires Go 1.21+.

```bash
git clone <repo-url> thakkali
cd thakkali
go build -o thakkali .
./thakkali
```

Move the binary anywhere on your `PATH` (e.g. `/usr/local/bin`) to run it as `thakkali` from any directory.

### Homebrew

Coming in a future release.

## Usage

```
thakkali [flags]
```

### Flags

Every flag has a short form.

| Flag        | Default | Description                               |
|-------------|---------|-------------------------------------------|
| Long          | Short | Default | Description                                    |
|---------------|-------|---------|------------------------------------------------|
| `-work`       | `-w`  | 25      | Timer length (minutes)                         |
| `-pomodoro`   | `-p`  | false   | Enable full Pomodoro cycle (breaks + rounds)   |
| `-short`      | `-s`  | 5       | Short break length (Pomodoro mode)             |
| `-long`       | `-l`  | 15      | Long break length (Pomodoro mode)              |
| `-rounds`     | `-r`  | 4       | Work rounds before a long break (Pomodoro)     |
| `-task`       | `-t`  | —       | Task description to tag the session            |
| `-minimal`    | `-m`  | false   | Hide logo and tomato animation                 |
| `-sound`      | `-S`  | beep    | macOS system sound name (see below)            |

### Examples

```bash
# Default: simple 25-minute timer
thakkali

# 45-minute timer with a task tag
thakkali -work 45 -t "deep work"

# Full Pomodoro cycle — 25/5/15, 4 rounds
thakkali -p

# Custom Pomodoro — longer work, fewer rounds
thakkali -p -work 50 -short 10 -rounds 3

# Quick smoke test — one full Pomodoro cycle in ~5 minutes
thakkali -p -work 1 -short 1 -long 1 -rounds 2
```

## Keybindings

| Key       | Action                                |
|-----------|---------------------------------------|
| `space`   | Pause / resume                        |
| `r`       | Reset current phase timer             |
| `s`       | Skip to next phase (Pomodoro mode)    |
| `m`       | Toggle minimal mode (hide logo + tomato) |
| `+` / `=` | Add 1 minute to current phase         |
| `-` / `_` | Subtract 1 minute from current phase  |
| `q`       | Quit                                  |

## Config

A config file is created on first run at:

- **macOS:** `~/Library/Application Support/thakkali/config.json`
- **Linux:** `~/.config/thakkali/config.json`
- **Windows:** `%AppData%\thakkali\config.json`

```json
{
  "work": 25,
  "short": 5,
  "long": 15,
  "rounds": 4,
  "sound": ""
}
```

Edit it to change your defaults. CLI flags always override config values.

### Notification sounds

- **Default** (empty string, `"default"`, or `"beep"`) — cross-platform beep
- **macOS**: set `sound` to any system sound name (the `.aiff` file basename), e.g. `"Glass"`, `"Ping"`, `"Hero"`, `"Submarine"`. Full path also works. Available on macOS 15:
  `Basso`, `Blow`, `Bottle`, `Frog`, `Funk`, `Glass`, `Hero`, `Morse`, `Ping`, `Pop`, `Purr`, `Sosumi`, `Submarine`, `Tink`
- **Linux / Windows** — always use the cross-platform beep (custom sounds TBD)

```bash
thakkali -sound Glass       # override from the command line
```

## Session log

Every completed phase (work *and* break) is appended as one JSON object per line to `log.jsonl` in the same directory as `config.json`:

```json
{"timestamp":"2026-04-14T17:30:00Z","phase":"work","duration_sec":1500,"task":"Ship Phase 4"}
{"timestamp":"2026-04-14T17:55:00Z","phase":"short_break","duration_sec":300}
```

This format is easy to grep, pipe to `jq`, or load into any tool for your own stats.

## Stats

```bash
thakkali stats              # today + last 7 days
thakkali stats -days 30     # custom window
```

Prints today's total, a per-day ASCII bar chart, top tasks by time, and all-time work total — all read from `log.jsonl`.

## Roadmap

- macOS system-sound customization (`afplay ~/System/Library/Sounds/*.aiff`)
- Homebrew tap with single-binary distribution
- Additional animations (other fruit? different styles?)

## Inspirations

- [Ghostty](https://ghostty.org) — for the layered ASCII animation style
- [GSD / get-shit-done](https://github.com/gsd-build/get-shit-done) — for the logo font
- [pymodoro](https://github.com/emson/pymodoro) — for feature inspiration

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lip Gloss](https://github.com/charmbracelet/lipgloss).
