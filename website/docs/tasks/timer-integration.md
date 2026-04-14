# Timer ↔ task integration

The two halves of Thakkali — time tracking and task tracking — come
together when you pass a task ID to the timer's `-task` / `-t` flag.

```bash
thakkali -w 25 -t THAK-3           # countdown session tagged to THAK-3
thakkali -p -t THAK-3              # Pomodoro cycle, every work round logged
thakkali -T -t THAK-3 -target 1h   # stopwatch tied to THAK-3
```

## What happens on start

When `-t` looks like a task reference (`THAK-3`, `TSK-7`,
`AUTH-12` — case-insensitive, any `[A-Z]+-N` prefix):

1. **Resolve.** Thakkali reads the current task file and finds the
   task by numeric ID. If no task matches, the timer exits with a
   clear error — no partial run, no silent skip.
2. **Replace the title.** The `m.task` string becomes the task's
   actual title (not the raw `THAK-3`), so the status line reads
   `task: THAK-3 Ship v2 storage #thakkali`.
3. **Auto-promote.** If the task is currently `todo`, Thakkali flips
   it to `doing` and stamps `@begin` before the timer starts running.
   So the `@begin` stamp marks when work *actually started*, not when
   the task was captured.
4. **Carry the metadata.** The resolved `task_id`, `task_prefix`, and
   `#project` are threaded into every `log.jsonl` entry for the
   session.

## What the log looks like

```jsonl
{"timestamp":"2026-04-15T09:00:00Z","phase":"work","duration_sec":1500,"task":"Ship v2 storage","task_id":3,"task_prefix":"THAK","project":"thakkali"}
```

Those extra fields are what [`thakkali stats`](../viz/stats.md) uses
to render:

- **Top tasks** prefixed with the task ID (`THAK-3 Ship v2 storage`).
- **Top projects** aggregated from `project`.
- A **tracked tasks** table joining log time with the current task
  file — state, due date, session count, and a red `[overdue]` badge
  for tasks past their due date.

## Free-text tags still work

Not every session is tied to a tracked task:

```bash
thakkali -w 45 -t "ship the docs"
```

A free-text `-task` value just lands in `"task"` — no resolve, no
auto-promote, no `task_id`. Top tasks still aggregates these by
string, so your ad-hoc tags don't disappear.

## Auto-promote: the details

Auto-promotion only fires when the task is currently in `todo` state.
If the task is already `doing` or `done`, the existing `@begin` stamp
stays intact — Thakkali assumes you know what you're doing.

!!! example "Why this matters"
    A typical workflow:

    ```bash
    thakkali task add "Ship phase 8" -d 2026-05-01   # creates THAK-3 in todo
    ...later...
    thakkali -p -t THAK-3                            # @begin stamps here
    ```

    Without auto-promote, `@begin` would stay empty until you ran
    `task move THAK-3 doing` manually — which you'd usually forget.

## Resuming a task across sessions

Just run the timer again with the same ID:

```bash
thakkali -p -t THAK-3   # Monday morning — @begin is stamped
...
thakkali -p -t THAK-3   # Tuesday morning — @begin unchanged, new sessions logged
```

Every Pomodoro work round or stopwatch session adds a new log entry;
the task stays `doing` until you run `task done THAK-3` or cycle it
from the TUI.

## Deleted tasks in the log

If you delete a task (`task rm THAK-3`) after sessions have already
been logged for it, those log entries stay. `thakkali stats` shows the
task with `[deleted]` in place of a state, so historical time isn't
lost.

## Stats mental model

| Field in `log.jsonl`            | Where it shows in stats       |
|---------------------------------|-------------------------------|
| `task`                          | Top tasks (free-text rollup)  |
| `task_id` + `task_prefix`       | Top tasks (with `THAK-3` id)  |
| `project`                       | Top projects                  |
| `task_id` joined to task file   | Tracked tasks table           |
