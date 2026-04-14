# Thakkali — context for Claude Code

A terminal Pomodoro timer with a Ghostty-inspired rolling-tomato ASCII animation. Personal project of Praveen Kumar Srinivasan. Intended to be installed on multiple machines and shared with colleagues — keep dependencies minimal and the install story simple.

## Stack

- **Go** (single binary, cross-platform)
- **Bubble Tea** — TUI event loop
- **Lip Gloss** — styling and layout
- **beeep** — cross-platform beep and desktop notifications
- macOS-only: shells out to `afplay` for system-sound support

No other runtime deps. Adding a dep should clear a high bar — every dep is friction for install/distribution.

## Layout

Single-file app. Everything lives in `main.go`:

- **Timer state + Bubble Tea model**
- **Tomato frame generation** — procedural, two pre-baked frame sets (idle shimmer + full-rotation spin)
- **ANSI Shadow digit font** for the timer display, matching the header logo
- **Config + session log** under `os.UserConfigDir()/thakkali/`
- **`stats` subcommand** reads `log.jsonl` and prints totals + ASCII bar chart

Supporting files:

- `README.md` — user-facing docs
- `docs/` — internal planning docs (phase plans, design notes)
- `go.mod`, `go.sum` — module metadata; binary is `./thakkali`

## Build and run

```bash
go build -o thakkali .    # always run after source changes
./thakkali                 # simple 25-min timer (default)
./thakkali -p              # full Pomodoro cycle (work + breaks + rounds)
./thakkali -T              # stopwatch / count-up timer for task tracking
./thakkali -T -target 45m  # stopwatch with a soft target
./thakkali stats           # both sections — Pomodoro then Timer
./thakkali stats -m timer  # only stopwatch chart
```

There is no test suite yet. Verification is visual: build, run, observe the animation and timer behaviour.

## Design principles

- **Crisp over pretty.** The tomato is meant to read cleanly at a glance — layered shading with a dark-red halo, a clean gap, and a dense bright-red body, mimicking Ghostty's ghost. Avoid character ramps that produce fuzzy edges.
- **Ghostty vocabulary where it fits.** Halo ramp `x = + * % $ @`; body ramp `@ $ %`; eyes are empty cells; stem is green.
- **Defaults for the common case.** Simple timer is the default; Pomodoro is opt-in via `-pomodoro` / `-p`. Every flag has a long and a short form.
- **Don't over-engineer.** Single-file Go, JSON config, JSONL log. No framework ceremony. No unused abstractions.

## Phase status (as of this writing)

Phases 1–6 shipped and on `main`:

1. Core timer + logo + keybindings
2. Rolling tomato animation (Ghostty-style layered sprite, blink, spin, jump)
3. Pomodoro cycle, session counter, CLI flags
4. Config file + task tagging + JSONL session log
5. `stats` subcommand
6. macOS system-sound support (`afplay`) with cross-platform beep fallback

Post-phase polish also landed: ANSI Shadow digit font for the timer, solo/pomodoro flip (simple timer is default), minimal mode (hide logo + animation), duration hotkeys (`+` / `-`), short flag aliases, and a count-up stopwatch mode (`-timer` / `-T`) with optional soft target (`-target`) and a split `stats -mode` view (pomodoro | timer | all).

Pending:

7. Distribution — GoReleaser, GitHub Actions release workflow, Homebrew tap. Plan in `docs/phase-7-distribution.md`.

## Conventions

- **Commit style**: imperative mood, focused on *why*. First commit is the reference.
- **User email for this repo only**: `praveen.sxi@gmail.com`, name "Praveen Kumar Srinivasan" — set locally, not globally.
- **Never skip git hooks** or bypass signing on commits.
- **Don't add tests or type checks that weren't asked for.** When the user asks for a feature, ship the feature, not a framework.
- **Don't write code comments that merely restate what the code does.** Comment *why* (a non-obvious constraint, a subtle invariant, a workaround).
