# Visualization

Three read-only commands turn your `log.jsonl` and `thakkali.md` into
something you can look at.

| Command                                          | Data source                  | What it shows                                             |
|--------------------------------------------------|------------------------------|-----------------------------------------------------------|
| [`thakkali stats`](stats.md)                     | `log.jsonl` + `thakkali.md`  | Totals, bar chart, top tasks / projects, tracked tasks.   |
| [`thakkali gantt`](gantt.md)                     | `thakkali.md`                | Planned vs. actual date ranges over a week / month / year.|
| [`thakkali activity`](activity.md)               | `log.jsonl`                  | GitHub-style 52-week contribution heatmap.                |

All three are pure readers — they never modify your files.

## When to use which

- **What did I spend time on?** → [`stats`](stats.md).
- **What's on my plate over the next month?** → [`gantt`](gantt.md).
- **How consistent has my focus been this year?** → [`activity`](activity.md).

## What each one needs

| Command     | Needs                                             | Still works without              |
|-------------|---------------------------------------------------|----------------------------------|
| `stats`     | At least one completed session.                   | No task file (free-text only).   |
| `gantt`     | At least one task with `@start` / `@due` / actuals. | Empty log.                      |
| `activity`  | At least one logged session.                      | No task file.                    |

If a command has nothing to show, it prints a friendly reminder instead
of failing.
