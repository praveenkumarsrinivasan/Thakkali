# `thakkali activity`

GitHub-style contribution heatmap of your logged sessions. One cell
per day, columns are weeks (Sunday-anchored), rows are days of the
week.

```bash
thakkali activity              # default: 52 weeks
thakkali activity -weeks 12    # narrower window
thakkali activity -w 26        # 6 months (short flag)
```

## Flags

| Flag                | Default | Range   | Effect                            |
|---------------------|---------|---------|-----------------------------------|
| `-weeks`, `-w <n>`  | 52      | 4 – 104 | How many weeks back from today.   |

The right-most column of the grid is always anchored on the current
week's Saturday; the grid fills in weeks backward from there.

## Sample output

```text
THAKKALI · activity (52 weeks)
range: 2025-04-20 → 2026-04-18

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

## Reading the grid

- Each column is one **week**, starting Sunday.
- Each row is one **day of the week** — Sun, Mon, Tue, Wed, Thu, Fri,
  Sat from top to bottom. Alternating rows are labeled (Mon / Wed /
  Fri) to reduce visual clutter.
- The month labels above the grid sit at the column for each month's
  first-Sunday.
- Cells are shaded by intensity relative to the busiest day in the
  window:
    - `·` — no activity
    - dim `■` — 1–24% of the peak day
    - medium `■` — 25–49%
    - bright `■` — 50–74%
    - bold bright `■` — 75–100%
- Future cells (within the rightmost week if today isn't Saturday)
  stay blank.

## Legend and totals

Below the grid:

- **total** — sum of every session's duration within the window.
- **active days** — number of days with at least one logged session.
- A color-level legend (`less · ■ ■ ■ ■ more`) so you can calibrate.

## What counts as activity

Every line in `log.jsonl` with a non-zero `duration_sec`. That
includes:

- Pomodoro work phases
- Stopwatch runs

Short / long breaks are **not** logged, so they don't show up here.

## Scoping by mode or task

The current `activity` command aggregates everything. If you want
mode-specific or task-specific heatmaps, pipe `jq` + your own
visualizer — the log is JSON Lines and easy to process.

```bash
# activity just for task THAK-3
jq -c 'select(.task_id == 3)' \
  < ~/Library/Application\ Support/thakkali/log.jsonl
```

Feature requests welcome.

## Tips

- If every cell looks the same dim gray, you probably have one outlier
  day that's dragging the scale. Try a narrower `-weeks 12` window —
  the peak recalculates per run.
- The heatmap scales to terminal width. On narrow terminals (< 110
  cols), consider `-weeks 26` or smaller.
