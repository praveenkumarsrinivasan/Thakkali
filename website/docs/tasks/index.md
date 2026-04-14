# Tasks

Thakkali also tracks *work* — not just time. A plain markdown file acts
as your source of truth, and the binary wraps it with four surfaces:

| Surface                                      | Command                             |
|----------------------------------------------|-------------------------------------|
| [CLI CRUD](cli.md)                           | `thakkali task add \| list \| ...`  |
| [Interactive TUI](todo-tui.md)               | `thakkali todo`                     |
| [Kanban board](kanban.md)                    | `thakkali kanban`                   |
| [Bulk edit](bulk-edit.md) in your editor     | `thakkali task bulk`                |

All four surfaces read and write the same `thakkali.md` file with
uniform `@begin` / `@done` auto-stamping, so you can freely switch
between them.

## The task file at a glance

```markdown
# Thakkali tasks

- [ ] THAK-12 Review auth PR #auth @start:2026-04-20 @due:2026-05-01
- [*] THAK-13 Draft phase-8 RFC #thakkali @begin:2026-04-19T10:30:00Z
- [x] THAK-14 Refactor config #thakkali @done:2026-04-18T14:22:00Z
- [ ] Buy milk
```

- **State:** `[ ]` todo · `[*]` doing · `[x]` done.
- **ID:** `THAK-12` (prefix derived from the directory — `TSK-12`
  for the global file). Monotonic within the file; auto-assigned on
  save. Free-text lines work too — they get an ID on the next write.
- **Project:** `#project` — a single tag per task.
- **Dates:** planned `@start:YYYY-MM-DD`, `@due:YYYY-MM-DD`; actuals
  `@begin:<RFC3339>`, `@done:<RFC3339>`, `@created:<RFC3339>`.
- **Title:** everything between the ID and the first `#tag` / `@field`.

See the full [file-format reference](../reference/file-formats.md).

## Where the file lives

Discovery order when a command runs:

1. `./thakkali.md` in the current directory.
2. `./.thakkali/tasks.md` in the current directory.
3. Fallback to a global `~/Library/Application Support/thakkali/tasks.md`
   (or the platform equivalent of `os.UserConfigDir()/thakkali/tasks.md`).

`task add` writes to `./thakkali.md` if nothing exists anywhere, so
project-local is the default — commit the file to git and your whole
team shares the same list.

## Pick your surface

<div class="grid cards" markdown>

- :material-console: **[CLI CRUD](cli.md)**
  Add, list, move, done, rm, show. Best for one-off edits from a
  scripting context or terminal flow.

- :material-view-list: **[`thakkali todo`](todo-tui.md)**
  Interactive list TUI grouped DOING / TODO / DONE. Best for a
  quick "what's on my plate today".

- :material-view-column: **[`thakkali kanban`](kanban.md)**
  Three-column board with task shifting. Best for sprint-style
  planning.

- :material-file-edit: **[Bulk edit](bulk-edit.md)**
  Open the file in `$EDITOR` (nvim, vim, …). Best for pasting in a
  dump from a meeting or reorganizing in bulk.

</div>

## Closing the loop back to the timer

Sessions run with `-t THAK-N` are recorded against the task, and the
task auto-promotes `todo → doing` on the first tagged session. See the
[timer ↔ task integration](timer-integration.md) page for the full
behavior.
