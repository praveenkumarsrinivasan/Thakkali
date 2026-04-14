# Thakkali

**Terminal Pomodoro timer with a rolling-tomato ASCII animation — and a
lightweight task tracker built into the same binary.**

Thakkali (தக்காளி, Tamil for *tomato*) is a single-binary CLI that
does two things well:

1. **Tracks time** — countdown, Pomodoro cycles, and count-up stopwatch
   with a Ghostty-inspired rolling-tomato sprite.
2. **Tracks work** — an Obsidian-style markdown task file with CLI,
   interactive TUI, kanban board, and gantt / activity visualizations.

All in one binary, with a `log.jsonl` session log and a
`thakkali.md` task file that you can commit to your repo so your whole
team shares the same view.

```bash
$ thakkali -p -t TSK-3            # Pomodoro tied to a tracked task
$ thakkali todo                   # interactive TUI
$ thakkali kanban                 # three-column board
$ thakkali stats -days 30         # time + tracked tasks rollup
$ thakkali activity               # GitHub-style heatmap from log.jsonl
```

---

## Why Thakkali?

- **One binary, no services, no accounts.** Drop it in `$PATH` and go.
- **Plain-text state.** `thakkali.md` is markdown, `log.jsonl` is
  JSON Lines — grep, `jq`, commit to git, edit in `nvim`.
- **Works offline, works on planes.** No cloud, no API keys.
- **Crisp terminal UI.** Bubble Tea + Lip Gloss; matches Ghostty's visual
  language where it fits.
- **Time and work in one place.** `-task TSK-3` auto-starts the tracked
  task, stamps `@begin`, and rolls up per-task / per-project time in
  `stats`.

---

## Where to go next

<div class="grid cards" markdown>

- :material-rocket-launch: **[Quickstart](quickstart.md)** — a 30-second tour.
- :material-download:    **[Install](install.md)** — Homebrew, go install, or build from source.
- :material-timer:       **[Timers](timers/index.md)** — countdown, Pomodoro, stopwatch.
- :material-check-all:   **[Tasks](tasks/index.md)** — CLI CRUD, interactive TUI, kanban.
- :material-chart-line:  **[Visualization](viz/index.md)** — stats, gantt, activity heatmap.
- :material-book-open:   **[Reference](reference/index.md)** — flags, keybindings, file formats.

</div>

---

## Screenshot tour

### Pomodoro mode

```text
████████╗██╗  ██╗ █████╗ ██╗  ██╗██╗  ██╗ █████╗ ██╗     ██╗
╚══██╔══╝██║  ██║██╔══██╗██║ ██╔╝██║ ██╔╝██╔══██╗██║     ██║
   ██║   ███████║███████║█████╔╝ █████╔╝ ███████║██║     ██║
   ██║   ██╔══██║██╔══██║██╔═██╗ ██╔═██╗ ██╔══██║██║     ██║
   ██║   ██║  ██║██║  ██║██║  ██╗██║  ██╗██║  ██║███████╗██║
   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚══════╝╚═╝

              work (1/4) · task: TSK-3 Ship phase D

                       ╭─────────────────╮
                       │                 │
                       │      24:37      │
                       │                 │
                       ╰─────────────────╯

             (the tomato spins, blinks, and hops here)

          space pause · r reset · m minimal · q quit
```

### Todo TUI

```text
THAKKALI · todo
/repo/thakkali.md

DOING
  ▸ THAK-3  Ship v2 storage             (#thakkali · due 2026-04-22 · began 4m ago)

TODO
    THAK-4  Write docs                  (#thakkali)
    THAK-5  Polish activity heatmap     (#thakkali · due 2026-05-01)

DONE
    THAK-1  Scaffold task model         (#thakkali · done 2h ago)
    THAK-2  Parse markdown tasks        (#thakkali · done 1h ago)

? help · q quit
```

### Kanban board

```text
THAKKALI · kanban

╭──────────────────╮╭──────────────────╮╭──────────────────╮
│ TODO  (2)        ││ DOING  (1)       ││ DONE  (2)        │
│ ──────────────── ││ ──────────────── ││ ──────────────── │
│ ▸ THAK-4 Write   ││   THAK-3 Ship v2 ││   THAK-1 Scaffol │
│     #thakkali    ││     #thakkali    ││     #thakkali    │
│   THAK-5 Polish  ││                  ││   THAK-2 Parse m │
│     #thakkali    ││                  ││     #thakkali    │
╰──────────────────╯╰──────────────────╯╰──────────────────╯

h/l switch col · j/k move · </> shift · space cycle · n new · e edit
```

### Activity heatmap

```text
THAKKALI · activity (52 weeks)

         May     Jun     Jul     Aug     Sep     Oct     Nov     Dec     Jan     Feb     Mar     Apr
    ■ · · ■ ■ ■ ■ ■ ■ · ■ ■ · ■ ■ · · · · ■ ■ · · ■ ■ · ■ ■ · ■ ■ · ■ ■ ■ ■ ■ · ■ ■ · ■ · ■ ■ ■ ■ ■ ■ · ·
Mon ■ · ■ ■ ■ · ■ ■ ■ ■ ■ ■ · ■ · ■ · · ■ · · · ■ ■ · · ■ · ■ · ■ · · ■ ■ ■ ■ ■ · ■ · · ■ ■ · ■ ■ ■ · · ■
    · ■ · ■ ■ · · ■ · ■ · ■ · ■ · ■ ■ ■ ■ ■ ■ ■ · ■ · ■ ■ · · ■ ■ ■ ■ · ■ · ■ · · ■ · ■ ■ ■ · · ■ ■ · · ■
Wed · ■ · ■ ■ ■ · · ■ · ■ ■ ■ ■ ■ ■ ■ ■ · ■ · ■ · ■ ■ · ■ · ■ ■ ■ ■ ■ ■ ■ ■ ■ · · ■ · ■ ■ ■ ■ ■ ■ · · ■ ·
    · ■ · ■ ■ ■ ■ ■ ■ ■ · ■ · ■ ■ ■ · · · · · · ■ ■ ■ · ■ ■ ■ · ■ ■ ■ ■ ■ ■ ■ · · ■ · · ■ · ■ · ■ ■ · ■
Fri ■ · ■ ■ ■ ■ ■ ■ ■ · · · ■ ■ ■ ■ · ■ ■ ■ · ■ ■ ■ ■ · ■ ■ ■ · · ■ ■ ■ ■ ■ ■ ■ ■ ■ · · ■ ■ · · ■ · ■ ■
    ■ · ■ · ■ ■ · ■ · ■ ■ ■ ■ ■ ■ ■ ■ ■ · ■ · · ■ ■ ■ ■ ■ · · ■ ■ · ■ ■ ■ ■ ■ ■ ■ ■ · ■ · · ■ ■ · ■ ■ ■ ■

total: 200h 10m   active days: 237

less · ■ ■ ■ ■ more
```

---

## License

MIT. See [the repository](https://github.com/praveenkumarsrinivasan/Thakkali/blob/main/LICENSE).
