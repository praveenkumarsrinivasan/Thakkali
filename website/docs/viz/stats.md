# `thakkali stats`

The primary retrospective view. Reads `log.jsonl` and joins it against
your current task file to produce a per-section summary.

```bash
thakkali stats                  # both sections: Pomodoro then Timer
thakkali stats -days 30         # wider window (default 7)
thakkali stats -mode pomodoro   # only Pomodoro / countdown sessions
thakkali stats -m timer         # only stopwatch sessions
thakkali stats -m timer -d 14   # stopwatch, last 14 days (short flags)
```

## Flags

| Flag                | Default  | Effect                                                           |
|---------------------|----------|------------------------------------------------------------------|
| `-days`, `-d <n>`   | 7        | Size of the recent-window bar chart.                             |
| `-mode`, `-m <m>`   | `all`    | Which sections to render: `all`, `pomodoro`, or `timer`.         |

## Output structure

Each mode section has up to six blocks:

1. **today** — total for today with session count.
2. **last N days** — total across the window.
3. **Bar chart** — one row per day with an ASCII bar and total.
4. **top tasks** — aggregated by tag (prefixed with `THAK-N` for
   tracked tasks, free-text for untagged strings).
5. **top projects** — aggregated from log `project` field.
6. **tracked tasks** — joins log by `task_id` with the task file:

    ```text
    tracked tasks
      THAK-1  Ship phase D              [doing]       50m    2 sess  #thakkali · due 2026-04-22
      THAK-2  Review PRs                [overdue]     25m    1 sess  #reviews  · due 2026-04-10
    ```

    - **State badges:** `[todo]` / `[doing]` / `[done]` in dim; red
      `[overdue]` when a non-done task is past its `@due`; `[deleted]`
      if the task was removed from the file but the log has entries.

7. **all-time total** — sum across every session of this mode, ever.

## A full sample

```text
████████╗██╗  ██╗ █████╗ ██╗  ██╗██╗  ██╗ █████╗ ██╗     ██╗
╚══██╔══╝██║  ██║██╔══██╗██║ ██╔╝██║ ██╔╝██╔══██╗██║     ██║
   ██║   ███████║███████║█████╔╝ █████╔╝ ███████║██║     ██║
   ██║   ██╔══██║██╔══██║██╔═██╗ ██╔═██╗ ██╔══██║██║     ██║
   ██║   ██║  ██║██║  ██║██║  ██╗██║  ██╗██║  ██║███████╗██║
   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚══════╝╚═╝

██████████████████████████ STATS ███████████████████████████

POMODORO

today
  work   2h 00m  (4 sessions)

last 7 days
  total  11h 25m

  Thu Apr 09  ████████████░░░░░░░░░░░░  1h 15m
  Fri Apr 10  ███████████████░░░░░░░░░  1h 35m
  Sat Apr 11  ░░░░░░░░░░░░░░░░░░░░░░░░  0m
  Sun Apr 12  ████░░░░░░░░░░░░░░░░░░░░  25m
  Mon Apr 13  ████████████████████████  2h 30m
  Tue Apr 14  ██████████████░░░░░░░░░░  1h 40m
  Wed Apr 15  █████████████████████░░░  2h 00m ← today

top tasks
  THAK-3 Ship v2 storage          3h 15m
  THAK-1 Review auth PR           2h 10m
  deep work                       1h 45m
  THAK-4 Polish heatmap           55m

top projects
  #thakkali                       5h 05m
  #auth                           2h 10m

tracked tasks
  THAK-3  Ship v2 storage           [doing]   3h 15m   7 sess  #thakkali · due 2026-04-22
  THAK-1  Review auth PR            [doing]   2h 10m   5 sess  #auth · due 2026-05-01
  THAK-4  Polish activity heatmap   [overdue]   55m   2 sess  #thakkali · due 2026-04-10

all-time work: 38h 20m


TIMER

today
  timer  1h 30m  (2 sessions)

last 7 days
  total  4h 45m

  ...

all-time timer: 21h 05m
```

## Tips for reading it

- **Bars are scaled per section** — a full bar in the Pomodoro section
  is that section's busiest day, not a global max. Pomodoro and timer
  bars aren't comparable across sections.
- **Top tasks show tracked IDs inline** — `THAK-3` in the top-tasks
  list and in the tracked-tasks table are the same task. The
  tracked-tasks table adds the joined metadata.
- **Deleted ≠ lost.** Remove `THAK-3` from `thakkali.md` and its log
  entries still show up under top tasks (by title). The tracked-tasks
  table flags it `[deleted]` so you know it's a historical reference.
- **No sessions yet?** The top-level command prints a friendly
  `no sessions logged yet — run thakkali to get started.`

## Scripting

`thakkali stats` prints human-friendly output. For programmatic access,
read `log.jsonl` directly:

```bash
# total minutes today
jq -s '
  map(select(.phase == "work" and (.timestamp | startswith("2026-04-15"))))
  | (map(.duration_sec) | add / 60)
' < ~/Library/Application\ Support/thakkali/log.jsonl
```

See [file formats](../reference/file-formats.md#logjsonl) for the
complete schema.
