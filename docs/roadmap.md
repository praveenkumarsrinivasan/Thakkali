# Thakkali — Roadmap

The full phased plan for Thakkali, from first commit to distributable. Phases 1–6 are shipped; Phase 7 is pending. Post-phase polish items are listed at the end.

## Phase 0 — Decisions (shipped)

The choices that shaped everything else.

- **Language: Go.** Single static binary, no runtime for colleagues to install, trivial cross-compile.
- **TUI: Bubble Tea + Lip Gloss.** Purpose-built for exactly this kind of animated terminal UI.
- **Notifications: `beeep`** (cross-platform) for default beep + desktop notification; optional macOS `afplay` for system-sound customisation.
- **Distribution: Homebrew tap + GitHub Releases** via GoReleaser (Phase 7).
- **Name: Thakkali.** Tamil for "tomato".

## Phase 1 — Skeleton + core timer (shipped)

A usable 25-minute timer in a Bubble Tea shell.

**Scope**
- `go mod init`, Bubble Tea / Lip Gloss / beeep deps.
- Red ANSI Shadow "THAKKALI" logo header.
- Work countdown `mm:ss` inside a rounded red border.
- Keybindings: `space` pause/resume, `r` reset, `q` quit.
- Cross-platform beep + desktop notification on completion.

**Outcome**
- `./thakkali` and `./thakkali -work 1` both worked end-to-end.

## Phase 2 — Rolling tomato animation (shipped)

Bring the logo to life with a continuous ambient animation.

**Scope**
- Design an ASCII tomato sprite (stem + body).
- Add an animation ticker (separate from the timer ticker).
- Track terminal width via `tea.WindowSizeMsg`.
- Scroll the sprite horizontally, wrap around continuously.
- Render in a horizontal strip between logo and timer.

**Outcome**
- A small red tomato rolled left-to-right below the logo.

## Phase 2.5 — Ghostty-style rework (shipped)

The initial line-art tomato had gaps and read as simple ASCII. We reworked it to match the Ghostty website vocabulary.

**Scope**
- Research: subagent sweep of `ghostty.org` page source plus the renderer chunk, confirming 235 pre-baked frames at 32 fps with a custom density ramp.
- Switch to a discrete-layer design: dark-red outer halo, 1-cell dark gap, dense bright-red body.
- Borrow the Ghostty character sets — halo uses `x = + * % $ @`; body uses `@ $ %`.
- Eyes are empty cells (no character) for crisp contrast against the body.
- Pre-baked frame sets: 8 idle-shimmer variants and 24 spin-rotation frames.
- Behaviour: continuous gentle bounce, eye blink every ~1.4 s, periodic 360° spin + parabolic jump every ~8 s.

**Outcome**
- The tomato reads as a distinct character with a face; the spin-and-jump trick is the visual hook.

## Phase 3 — Full Pomodoro cycle (shipped)

From one timer to a disciplined cycle.

**Scope**
- Phase enum (`work`, `short_break`, `long_break`).
- Auto-transition on timer hit-zero.
- Session counter (`work 2 / 4`).
- Phase-aware label colour (red for work, green for breaks).
- CLI flags: `-work`, `-short`, `-long`, `-rounds`.
- `s` key to skip to next phase.
- `+` / `-` adjust the current phase's duration.
- Phase-specific notification message on transition.

**Outcome**
- `./thakkali -work 1 -short 1 -long 1 -rounds 2` cycled through work→break→work→long in ~5 minutes.

## Phase 4 — Config file + task logging (shipped)

Persistence beyond a single session.

**Scope**
- JSON config at `os.UserConfigDir()/thakkali/config.json`, auto-created with defaults on first run.
- Precedence: CLI flags > config values > hard-coded defaults.
- `-t "task"` flag to tag the current session; shows under the phase label.
- Append every completed phase (work and breaks) as one JSON object per line to `log.jsonl` in the same directory.
- Work entries include the task tag; breaks don't.

**Outcome**
- Config survives across runs; the log file grew over time and became the stats source for Phase 5.

## Phase 5 — Stats subcommand (shipped)

Read `log.jsonl`; surface patterns.

**Scope**
- `thakkali stats` subcommand, dispatched from `main()` before the TUI boots.
- Aggregations: today's work time + session count, last-N-days total, per-day bar chart (defaults to 7 days, `-days` flag to override), top 5 tasks by total time, all-time work total.
- ASCII bar chart using `█` and `░`, scaled to the max day in the window.
- Colour via Lip Gloss; output is still plain-text friendly for piping.

**Outcome**
- `thakkali stats` and `thakkali stats -days 30` both work and look good.

## Phase 6 — macOS system sounds (shipped)

A small quality-of-life addition.

**Scope**
- Add `sound` field to config and `-sound` / `-S` CLI flag.
- On macOS: shell out to `afplay /System/Library/Sounds/<name>.aiff` (supports `Glass`, `Ping`, `Hero`, `Submarine`, etc.). Absolute paths work too.
- On Linux and Windows: always fall back to the cross-platform beep.
- Empty string / `"default"` / `"beep"` → cross-platform beep on all platforms.
- Sound plays asynchronously so the UI doesn't block.

**Outcome**
- `./thakkali -sound Glass` — proper macOS completion sound.

## Post-phase polish (shipped)

Small changes that landed between phases based on feedback.

- **ANSI Shadow digit font** for the timer, matching the logo.
- **Solo vs Pomodoro flip.** The default run is now a simple 25-minute timer; the full cycle is opt-in via `-pomodoro` / `-p`.
- **Minimal mode** (`-minimal` / `-m`, or `m` at runtime) hides the logo and animation for a compact view.
- **Duration hotkeys** — `+` and `-` adjust the *current* phase duration (so breaks can be bumped too).
- **Short flag aliases** — every long flag has a single-letter counterpart (`-w`, `-s`, `-l`, `-r`, `-t`, `-m`, `-S`).
- **Help line** uses en-dashes between key and action for readability.
- **Notification messages** are phase-aware ("Break's over — back to work!", "All rounds done — enjoy a long break!").

## Phase 7 — Distribution (pending)

Ship binaries and a Homebrew formula via tag-triggered CI. Full plan in [`docs/phase-7-distribution.md`](./phase-7-distribution.md).

Summary:

- `LICENSE` (MIT, pending confirmation).
- Module path alignment with the GitHub repo URL.
- `-version` flag wired to `ldflags`-injected variables.
- `.goreleaser.yaml` for darwin/linux/windows on amd64 + arm64.
- `.github/workflows/release.yml` triggered by `v*` tag push.
- `homebrew-thakkali` tap repo + PAT-scoped GitHub Actions secret.
- README updates pointing at the real `brew install` command.

## Deferred (not yet scheduled)

- **Additional animations.** Phase 2 locked in one animation (rolling tomato with periodic spin). Other styles — different fruit, different motion patterns, ambient scene — deferred to post-v1.
- **Linux package repos** (`.deb`, `.rpm`, AUR, snap, flatpak). Tarball is enough for now.
- **Signed macOS builds** via an Apple Developer ID to eliminate Gatekeeper prompts.
- **In-app self-updater** (`thakkali upgrade`). Homebrew handles this for mac users already; unclear if worth building for Linux.
- **Test suite.** Currently verification is visual — build, run, observe. Worth adding once distribution is in place and the code needs protecting from regressions.
