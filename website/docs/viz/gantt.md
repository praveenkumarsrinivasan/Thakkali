# `thakkali gantt`

ASCII Gantt chart of the tasks in your task file. Planned date ranges
(`@start` → `@due`) render in dim red; actuals (`@begin` → `@done`, or
`@begin` → *now* if still in progress) render in bright red. A green
`│` marks today's column.

```bash
thakkali gantt                  # default: month view
thakkali gantt -view week       # 14-day window, 4 chars per day
thakkali gantt -view month      # ~45-day window, 2 chars per day
thakkali gantt -view year       # ~12-month window, ~4 days per char
thakkali gantt -v year          # short form
```

## Views

| `-view`   | Window                          | Density            |
|-----------|---------------------------------|--------------------|
| `week`    | Today − 3d  → Today + 11d       | 4 chars / day      |
| `month`   | Today − 7d  → Today + 38d       | ~2 chars / day     |
| `year`    | Today − 30d → Today + 11 months | ~4 days / char     |

## Sample output

```text
THAKKALI · gantt (month view)
file: /repo/thakkali.md

                                Apr 08       Apr 15       Apr 23       Apr 30       May 08       May 15
                                ┬───────────││┬────────────┬────────────┬────────────┬────────────┬───────────
THAK-4 Long-running effort      ████████████████████████████████████████████████████████████████████████████   #thakkali 2026-04-01 → 2026-06-30 [todo]
THAK-2 Review PRs                  ████████████████████████████████████                                       #reviews 2026-04-10 → 2026-04-30 [todo]
THAK-1 Ship phase C                         █████████████                                                      #thakkali 2026-04-15 → 2026-04-22 [doing]
THAK-3 Quarterly planning                   │                ████████████████████████████████████              #mgmt 2026-04-25 → 2026-05-15 [todo]

legend: █ planned  █ actual  │ today
```

## What gets included

A task shows up in the gantt if it has at least one of:

- `@start:YYYY-MM-DD`
- `@due:YYYY-MM-DD`
- `@begin:<RFC3339>` (actuals)
- `@done:<RFC3339>`  (actuals)

If it has only planned dates, the bar is dim red. If it has only
actuals, it's bright red. If it has both, you get two overlapping bars
in one row — the actuals overlay the planned range so you can see
slippage at a glance.

## Tasks are sorted by start

Earliest-starting tasks appear on top. If a task has no `@start`,
Thakkali falls back to `@begin`. This keeps the chart reading like a
calendar — older items on top, newer below.

## Today marker

The green `│` sits at today's column. Bars to the left of it are past;
bars to the right are future. It only renders where there's no bar —
when a task is actively in progress across today, the bar takes
precedence.

## Out-of-window tasks

Tasks whose ranges fall entirely outside the view window still render,
clipped to the visible window's left/right edges. So a task that
started three months ago and is due next month shows up in the `week`
view as a full-width bar — still useful context.

## When there's nothing to show

```text
no tasks with @start / @due / @begin / @done — nothing to plot.
file: /repo/thakkali.md
```

Easy fix — add dates to your tasks:

```bash
thakkali task add "Ship v2 docs" -p thakkali -s 2026-04-15 -d 2026-04-22
```

Or bulk-add dates by hand in `task bulk`:

```markdown
- [ ] THAK-3 Ship v2 docs #thakkali @start:2026-04-15 @due:2026-04-22
```
