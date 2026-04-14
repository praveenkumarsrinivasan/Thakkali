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
- **Stopwatch / timer mode** (`-timer` / `-T`) — count up to track a task, optional soft target
- **Animated tomato** — rolls continuously with a shimmering outer ring, blinking eyes, and a periodic 360° spin-and-jump trick (Ghostty-style layered ASCII)
- **Big ANSI Shadow digits** for the timer, matching the logo font
- **Cross-platform desktop notifications + beep** when a phase ends (macOS system sounds supported)
- **Task tagging** per work session, logged for later review
- **JSON config file** auto-created on first run
- **JSON-lines session log** for stats, scripting, or export
- **Minimal mode** — hide the logo and animation when you just want the timer

## Install

### Homebrew (macOS / Linux)

```bash
brew install praveenkumarsrinivasan/thakkali/thakkali
```

Upgrade later with `brew upgrade thakkali`.

### Prebuilt binaries

Grab the archive for your platform from the [Releases page](https://github.com/praveenkumarsrinivasan/Thakkali/releases), extract it, and put the `thakkali` binary somewhere on your `PATH`.

### From source

Requires Go 1.21+.

```bash
git clone https://github.com/praveenkumarsrinivasan/Thakkali.git
cd Thakkali
go build -o thakkali .
./thakkali
```

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
| `-timer`      | `-T`  | false   | Stopwatch mode — count up to track a task      |
| `-target`     |       | —       | Soft goal for `-timer` (e.g. `45m`, `1h30m`)   |
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

# Stopwatch — open-ended tracking for a task
thakkali -T -t "code review"

# Stopwatch with a 45-minute soft target (beeps and keeps running)
thakkali -T -target 45m -t "debug prod issue"
```

## Keybindings

| Key       | Action                                |
|-----------|---------------------------------------|
| `space`   | Pause / resume                        |
| `r`       | Reset current phase timer             |
| `s`       | Skip to next phase (Pomodoro mode)    |
| `m`       | Toggle minimal mode (hide logo + tomato) |
| `h`       | Toggle footer help                       |
| `+` / `=` | Add 1 minute (phase duration, or `-timer` target) |
| `-` / `_` | Subtract 1 minute (phase duration, or `-timer` target) |
| `1`       | Switch to countdown mode                       |
| `2`       | Switch to Pomodoro mode                        |
| `3`       | Switch to timer / stopwatch mode               |
| `q`       | Quit (saves in-progress `-timer` session) |

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

### Timer ↔ task integration

Pass a task ID to `-task` (or `-t`) and the session gets tied to the
tracked task:

```bash
thakkali -w 25 -t TSK-3          # countdown session tagged to TSK-3
thakkali -T -t TSK-3 -target 1h  # stopwatch run tied to TSK-3
```

The title shown in-app and the log entry both pick up the task's title
and `#project`. `thakkali stats` then rolls up time per task (prefixed
with `TSK-N`) and per project, alongside the existing free-text tags.

## Stats

```bash
thakkali stats                         # both sections — Pomodoro then Timer
thakkali stats -days 30                # custom window
thakkali stats -mode pomodoro          # only Pomodoro / countdown sessions
thakkali stats -mode timer             # only stopwatch sessions
thakkali stats -m timer -days 14       # short form
```

Each section prints today's total, a per-day ASCII bar chart (independently scaled), top tasks (prefixed with `TSK-N` when tied to a tracked task), top projects, and an all-time total — all read from `log.jsonl`.

## Tasks

Thakkali also tracks *work*, not just time. `thakkali task ...` is a
lightweight task manager that reads and writes a plain markdown file so
you can commit it to your repo and collaborate with colleagues.

```bash
thakkali task add "Review auth PR" -p auth -d 2026-05-01
thakkali task add "Draft phase-9 RFC" -p thakkali -s 2026-04-19 -d 2026-04-22
thakkali task list                         # active (todo + doing) tasks
thakkali task list -state all -p thakkali  # all tasks in a project
thakkali task move TSK-1 doing             # auto-stamps @begin
thakkali task done TSK-2                   # auto-stamps @done
thakkali task show TSK-1                   # full details
thakkali task rm TSK-3
thakkali task bulk                         # open $EDITOR for bulk capture
thakkali todo                              # interactive TUI for the task file
thakkali kanban                            # three-column TODO/DOING/DONE board
```

**TUI (`thakkali todo`).** Interactive view of the same markdown file,
grouped DOING → TODO → DONE. Every mutation writes to disk immediately
and auto-stamps `@begin` / `@done` on state transitions — so you can
leave the TUI open alongside `task add` from another shell and both
stay consistent as long as you press `r` to reload.

| key            | action                                      |
|----------------|---------------------------------------------|
| `j`/`k`, `↑↓`  | move cursor                                 |
| `g`/`G`        | jump to top / bottom                        |
| `space`, `⏎`   | cycle state: todo → doing → done → todo     |
| `n`            | new task (inline input)                     |
| `e`            | edit the selected task's title              |
| `d`            | delete the selected task                    |
| `/`            | filter by title or project (live)           |
| `c`            | clear the active filter                     |
| `r`            | reload from disk                            |
| `?`            | toggle keymap footer                        |
| `q`, `esc`     | quit                                        |

**Kanban (`thakkali kanban`).** Same data, three columns side-by-side.
The focused column has a bright border; everything else is dim.

| key            | action                                            |
|----------------|---------------------------------------------------|
| `h`/`l`, `←→`  | switch focused column                             |
| `j`/`k`, `↑↓`  | move cursor within column                         |
| `>`/`<`, `L`/`H` | shift the selected task one column right/left   |
| `space`, `⏎`   | cycle state (same effect as `>` with wrap)        |
| `n`            | new task in the focused column's state            |
| `e`/`d`        | edit / delete the selected task                   |
| `/`/`c`/`r`    | filter / clear filter / reload                    |
| `?`            | toggle keymap footer                              |
| `q`, `esc`     | quit                                              |

Stamps stay consistent with the state column on every save: a task
moved back from `done` to `doing` loses its `@done` stamp, and a task
demoted to `todo` loses both `@begin` and `@done`. So the markdown
file always reflects the current truth, not a stale history.

### Gantt and activity

```bash
thakkali gantt                 # default month view
thakkali gantt -view week      # 14-day window
thakkali gantt -view year      # 12-month window
thakkali activity              # GitHub-style 52-week heatmap
thakkali activity -weeks 12    # smaller window
```

`gantt` reads the task file and plots one row per task with a `@start`,
`@due`, `@begin`, or `@done` stamp. Planned ranges render in dim red,
actuals (begin → done, or begin → now if still in progress) in bright
red. A green `│` marks today's column.

`activity` reads `log.jsonl` and renders a Sun-rows × N-week-columns
heatmap of total tracked time per day. Cell intensity scales relative
to the busiest day in the window; future cells stay blank.

### Tracked-task rollup and auto-start

When the timer runs with `-t TSK-N`:

- If the task is currently `todo`, it is auto-promoted to `doing` and
  `@begin` is stamped on that first tagged session — so the task's
  `@begin` marks when work actually started, not when it was captured.
- The session log entry records the `task_id` and `#project`, so
  `thakkali stats` can join it against the current task file.

`stats` now ends each mode section with a **tracked tasks** table
showing per-task time, session count, current state, project and due
date. Overdue tasks (past their `@due` and not yet `done`) are flagged
in red. Tasks deleted from the task file still appear as
`[deleted]` — the log is authoritative for historical time.



**Storage.** The task file is plain markdown — one task per line in
Obsidian-style checklist format:

```
# Thakkali tasks

- [ ] TSK-12 Review auth PR #auth @start:2026-04-20 @due:2026-05-01
- [*] TSK-13 Draft phase-9 RFC #thakkali @begin:2026-04-19T10:30:00Z
- [x] TSK-14 Refactor config #thakkali @done:2026-04-18T14:22:00Z
- [ ] Buy milk
```

- State: `[ ]` todo, `[*]` doing, `[x]` done.
- `#project` — single project tag per task.
- `@start:YYYY-MM-DD`, `@due:YYYY-MM-DD` — user-set planned dates.
- `@begin:<RFC3339>`, `@done:<RFC3339>` — auto-stamped on state change.
- IDs like `TSK-12` are monotonic and auto-assigned on save; free-text
  lines without an ID also work — capture first, tag later. The prefix
  is derived per file: the global task file uses `TSK-N`; a
  project-local file derives its prefix from the containing directory
  name (e.g. `THAK-N` for `~/repos/Thakkali/thakkali.md`,
  `AUTH-N` for an auth-service repo). Mixed prefixes on the same file
  round-trip cleanly — a line imported as `AUTH-99` keeps its `AUTH`
  prefix; new lines inherit the file's default. The CLI is forgiving
  on input (`task move tsk-3 done` still resolves to `THAK-3` when
  that's what the file contains).

**File discovery.** `./thakkali.md` wins, then `./.thakkali/tasks.md`,
then the global `~/Library/Application Support/thakkali/tasks.md`
(linux/windows paths follow `os.UserConfigDir()`). On first write in a
fresh directory, Thakkali creates `./thakkali.md` so project-local is
the default — check it into git and your whole team sees the same list.

`thakkali task bulk` opens the file in `$EDITOR` (fallback: `nvim`,
`vim`). On exit, the file is reparsed and rewritten so new lines get
IDs and any state changes get `@begin` / `@done` stamps.

## Roadmap

- macOS system-sound customization (`afplay ~/System/Library/Sounds/*.aiff`)
- Homebrew tap with single-binary distribution
- Additional animations (other fruit? different styles?)

## Inspirations

- [Ghostty](https://ghostty.org) — for the layered ASCII animation style
- [GSD / get-shit-done](https://github.com/gsd-build/get-shit-done) — for the logo font
- [pymodoro](https://github.com/emson/pymodoro) — for feature inspiration

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lip Gloss](https://github.com/charmbracelet/lipgloss).
