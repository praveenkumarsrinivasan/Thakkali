# Phase 8 — Task/todo manager (A1: storage + CLI CRUD)

## Context

Thakkali today tracks *time* (Pomodoro, countdown, stopwatch) but not *work*. `-task "string"` tags a session in `log.jsonl` but there's no concept of an independent task that exists across sessions, has state (todo/doing/done), or ties to a project.

We want a lightweight task tracker built into the same binary — similar to an Obsidian daily-task file or a terminal Jira — so the user can:
- List what they're working on, what's up next, and what's done
- Eventually see a kanban view and a gantt view
- Track actuals (when a task moved to doing/done) and planned dates
- Commit a project-local task file to git so colleagues collaborating on the same repo share the same view
- Bulk-edit in neovim when capturing many tasks at once

This phase ships **A1**: the foundational data model, the markdown storage layer, and the CLI CRUD surface. No TUI. No timer integration yet. Those are A2 and A3 respectively.

Rationale for splitting: each sub-phase is independently shippable and gives the user something usable. A1 gives a working CLI task manager; A2 adds interactivity; A3 closes the loop back to time tracking.

## Goals (A1)

- A markdown file (`./thakkali.md` or `./.thakkali/tasks.md` in cwd, else `~/.../thakkali/tasks.md`) is the source of truth for tasks.
- `thakkali task add | list | move | done | rm | bulk | show` covers the full CLI CRUD flow.
- Tasks have: id, state, title, project, planned start, planned due, actual begin, actual done, created-at.
- Obsidian-style syntax: `#project` tags; `@key:value` date fields.
- State changes auto-stamp `@begin` / `@done` on save when stamp is missing — uniform behavior whether the change came from CLI, nvim-bulk, or (future) TUI.
- Project-local file wins over global, so teams can commit tasks to the repo.

## Non-goals (A1)

- No TUI (`thakkali todo`) — that's A2.
- No timer/stats integration (`-task TSK-N`, per-task rollups) — that's A3.
- No kanban, gantt, or contribution heatmap — those are Phases B and C.
- No username / multi-owner tracking — deferred per user's direction ("initial version need not have the username field").
- No per-project ID scoping — v2 ("if there is a project level markdown document then lets use project level scope for task numbers").

## Storage format

File format: markdown with a flat checklist. One task per line.

```
# Thakkali tasks

- [ ] TSK-12 Review auth PR #auth @start:2026-04-20 @due:2026-05-01
- [*] TSK-13 Draft phase-8 RFC #thakkali @begin:2026-04-19T10:30:00Z
- [x] TSK-14 Refactor config #thakkali @done:2026-04-18T14:22:00Z
- [ ] Buy milk
```

- **State:** `[ ]` todo, `[*]` doing, `[x]` done.
- **ID:** `TSK-N` monotonic, global, auto-assigned on save if missing. Parser tolerates lines without an ID (free-text capture works naturally in nvim-bulk).
- **Project:** `#project` (single tag per task, first `#word` wins; extra `#tags` ignored in A1).
- **Dates:** `@start:YYYY-MM-DD` (planned), `@due:YYYY-MM-DD` (planned), `@begin:<RFC3339>` (actual), `@done:<RFC3339>` (actual), `@created:<RFC3339>`. Planned dates are user-set; actuals are auto-stamped.
- **Title:** everything between the ID (or its absence) and the first tag/field.

File discovery:
1. `./thakkali.md` in cwd
2. `./.thakkali/tasks.md` in cwd
3. Fallback: `<thakkaliDir()>/tasks.md` (global)

Created on first write with a single `# Thakkali tasks` heading if absent.

## CLI surface

All under a new `task` subcommand, dispatched from `main()` at the same spot `stats` is handled today (main.go:1267–1269).

```
thakkali task add "title" [-p project] [-s YYYY-MM-DD] [-d YYYY-MM-DD]
thakkali task list [-state todo|doing|done|all] [-p project]
thakkali task move <id> <todo|doing|done>
thakkali task done <id>                # alias for `move <id> done`
thakkali task rm <id>
thakkali task bulk                     # opens the file in $EDITOR (fallback: nvim, then vim)
thakkali task show <id>                # show full details for one task
```

`task list` default: show non-done tasks in insertion order, grouped by state (DOING → TODO → DONE if `-state all`), colorized with the existing `lipgloss` palette (red accent like stats).

`task bulk` spawns `$EDITOR $TASK_FILE` synchronously; on exit, reparse + rewrite (which auto-stamps actuals for any state changes).

## Implementation

Single-file per existing convention (CLAUDE.md: "Single-file app. Everything lives in main.go"). Add roughly 300–400 LOC to `main.go`:

**New types**
```go
type taskState uint8
const (stateTodo taskState = iota; stateDoing; stateDone)

type task struct {
    ID        int          // 0 = unassigned
    State     taskState
    Title     string
    Project   string
    Start     string       // YYYY-MM-DD, optional
    Due       string
    Begin     time.Time    // zero = unset
    Done      time.Time
    Created   time.Time
    Raw       string       // original line for round-trip fidelity on unknown extras
}
```

**New functions**
- `taskFilePath() (string, error)` — cwd-first discovery, then fallback to `filepath.Join(thakkaliDir(), "tasks.md")`. Reuses existing `thakkaliDir()` at main.go:319.
- `readTasks() ([]task, error)` — parse the file; tolerate any lines that don't match the task shape (pass through as `Raw` comments or skip).
- `writeTasks([]task) error` — serialize back with the `# Thakkali tasks` header. Auto-stamp actuals before serializing: if `t.State == stateDoing && t.Begin.IsZero()` → stamp `now()`; same for `Done`.
- `nextTaskID([]task) int` — max existing ID + 1.
- `parseTaskLine(line string) (task, bool)` — hand-roll tokenizer. Patterns: `^\s*- \[([ *x])\]\s+(.*)$`; then within the trailing text pull `TSK-N`, `#proj`, `@k:v` tokens with a simple split-and-classify loop. No regex library beyond stdlib.
- `renderTaskLine(t task) string` — inverse.
- `runTask(args []string)` — top-level dispatcher modelled after `runStats` (main.go:1109). Sub-sub-commands via a `switch args[0]`.
- Sub-handlers: `taskAdd`, `taskList`, `taskMove`, `taskRm`, `taskBulk`, `taskShow`.
- `openEditor(path string) error` — resolve `$EDITOR` → `nvim` → `vim`, spawn with `exec.Command`, wire stdin/stdout/stderr.

**Main dispatch**
```go
if len(os.Args) > 1 && os.Args[1] == "task" {
    runTask(os.Args[2:])
    return
}
```

**Help text**
- Extend `flag.Usage` to include `thakkali task ...` line in the Usage block.
- Extend `printExamples()` with a new `section("task management", [...])` block showing 5–6 examples.

## Verification

End-to-end test sequence after building:

```bash
go build -o thakkali .

# create + list
./thakkali task add "Review auth PR" -p auth -d 2026-05-01
./thakkali task add "Draft RFC" -p thakkali -s 2026-04-19 -d 2026-04-22
./thakkali task add "Buy milk"
./thakkali task list

# state transitions (verify @begin / @done auto-stamp)
./thakkali task move TSK-1 doing
./thakkali task list -state doing
./thakkali task done TSK-2
cat thakkali.md     # confirm @begin / @done stamps appear

# bulk edit round-trip
./thakkali task bulk     # opens $EDITOR; change [ ] to [*] on a line, save, quit
./thakkali task list     # confirm state change picked up and auto-stamped

# cwd-vs-global discovery
cd /tmp && ./thakkali task list     # should use global ~/Library/Application Support/thakkali/tasks.md
cd /tmp/foo && touch thakkali.md && ./thakkali task list     # should use local file

# delete + free-text fidelity
./thakkali task rm TSK-3
# manually add a free-text line without ID in the file, then:
./thakkali task list     # free-text line should appear, get auto-assigned an ID only on next save

# error paths
./thakkali task move TSK-99 done     # unknown id → clear error, exit 2
./thakkali task add                  # missing title → usage hint
./thakkali task move TSK-1 bogus     # invalid state → clear error
```

Visual verification: color/formatting match the existing stats output (red headers, bright-red accent, dim secondary text).

Parser robustness check: hand-edit `tasks.md` in nvim to include (a) lines without IDs, (b) extra `#extra-tags`, (c) unknown `@future:keys`, (d) blank lines, (e) the heading. Re-running `task list` must preserve the file on rewrite except for the intended state changes.

## Critical files

- `main.go` — all new code lives here (types, parser, CLI handlers, usage, examples)
- `README.md` — add a "Tasks" section after "Stats"; usage + syntax cheatsheet
- `CLAUDE.md` — bump phase status line mentioning Phase 8 / A1 shipped

## Follow-ups (A2, A3, B, C, D)

- **A2** — `thakkali todo` list TUI with `n/e/space/d//` hotkeys; reuses `readTasks` / `writeTasks`.
- **A3** — `-task TSK-N` resolves to a tracked task; `logEntry` gains `TaskID *int`; stats grows per-task / per-project rollups.
- **B** — `thakkali kanban` three-column board.
- **C** — `thakkali gantt -view week|month|year` + `thakkali activity` (GitHub-style 52-week contribution heatmap).
- **D** — deeper stats integration; optional actual-start auto-stamp on first tagged session.
- **v2 storage** — per-project ID scoping when a project-local file is used.
- **Username** — stamp tasks with `$USER` once multi-user collaboration is a proven need.

## Risks and mitigations

- **Parser fragility on hand-edited files.** Mitigation: every recognized token is optional; the parser skips lines it doesn't recognize instead of erroring; unknown `@key:value` fields are preserved in `task.Raw` residue and re-emitted on rewrite to keep nvim-bulk round-trip safe.
- **Concurrent edits (nvim open + `task add` from another shell).** v1 ignores — documented as a known limitation. Plausible future fix: file mtime check before rewrite.
- **Project-local file surprises** (user `cd`s into a random directory with `thakkali.md`). Mitigation: `task list` always prints the resolved file path at the top, dim-styled, so the user sees which file is in play.
- **ID reuse after deletes.** Mitigation: IDs are max+1, not count+1, so deleting TSK-3 doesn't reassign it. Historical log references stay valid.
