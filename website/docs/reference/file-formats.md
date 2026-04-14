# File formats

Thakkali writes two files and reads a third. All three are plain text
— grep them, `jq` them, commit them to git, edit them in `nvim`.

## `thakkali.md` — the task file

Obsidian-style markdown checklist. One task per line.

```markdown
# Thakkali tasks

- [ ] THAK-12 Review auth PR #auth @start:2026-04-20 @due:2026-05-01
- [*] THAK-13 Draft phase-8 RFC #thakkali @begin:2026-04-19T10:30:00Z
- [x] THAK-14 Refactor config #thakkali @done:2026-04-18T14:22:00Z @begin:2026-04-18T13:10:00Z
- [ ] Buy milk
```

### Line grammar

```
- [<state>] [<id> ]<title>[ #<project>][ @<key>:<value>]*
```

- `<state>` is `<space>` (todo), `*` (doing), or `x` (done).
- `<id>` is optional. When present, it matches `^[A-Z]+-\d+$`
  (e.g. `THAK-12`, `AUTH-99`, `TSK-3`).
- `<title>` is every word between the ID (or its absence) and the
  first `#tag` or `@key:value` field.
- `<project>` is the first `#word` token; extra `#tags` are
  preserved verbatim as extras.
- `@key:value` fields are parsed case-sensitively. Unknown keys are
  preserved so hand-edited extras round-trip cleanly.

### Recognized `@` fields

| Key        | Value format                 | Meaning                                     |
|------------|------------------------------|---------------------------------------------|
| `@start`   | `YYYY-MM-DD`                 | Planned start date.                         |
| `@due`     | `YYYY-MM-DD`                 | Planned due date.                           |
| `@begin`   | RFC3339 (`2026-04-19T10:30:00Z`) | Actual start timestamp (auto-stamped).  |
| `@done`    | RFC3339                      | Actual completion timestamp (auto-stamped). |
| `@created` | RFC3339                      | Creation timestamp (auto-stamped).          |

Anything else — `@blocked:true`, `@effort:3h`, `@link:https://...` —
is preserved on every rewrite but not used by Thakkali itself.

### File discovery

When you run any `task` subcommand (or `thakkali todo` /
`thakkali kanban`), Thakkali searches in order:

1. `./thakkali.md` in the current directory.
2. `./.thakkali/tasks.md` in the current directory.
3. Global: `~/Library/Application Support/thakkali/tasks.md`
   (or the platform equivalent).

If nothing exists and you're running a write command (`task add`,
`task bulk`), Thakkali creates `./thakkali.md` — project-local is the
default so the file can be committed to your repo.

### ID prefixes

The ID prefix is derived per file:

- **Global file** (under `UserConfigDir/thakkali/tasks.md`) → `TSK-N`.
- **Project-local file** → prefix derived from the containing
  directory name: uppercase alphanumeric, truncated to 4 characters.
    - `~/repos/Thakkali/thakkali.md` → `THAK-1`, `THAK-2`, …
    - `~/repos/auth-service/thakkali.md` → `AUTH-1`, …
    - `~/repos/x/thakkali.md` → falls back to `TSK-1` (name too short).
- **Existing prefixes are preserved.** An imported `AUTH-99` line stays
  `AUTH-99` even when the file's default prefix is `THAK`. The parser
  accepts any `[A-Z]+-\d+` ID on input.

### Round-trip guarantees

Every rewrite preserves:

- The task's original order in the file.
- Non-task lines (blank lines, comments, prose) at their original
  positions.
- Per-task prefixes (for imports with mixed prefixes).
- Unknown `@key:value` fields and extra `#tags`.

Every rewrite normalizes:

- State stamps — `@begin` and `@done` are set or cleared to match the
  current state.
- Missing IDs — any checklist line without a `<PREFIX>-N` token gets
  the next available integer ID and the file's default prefix.
- Missing `@created` on any task — stamped to "now" on first save.

## `log.jsonl` — the session log {#logjsonl}

One JSON object per line, appended on session completion.

```jsonl
{"timestamp":"2026-04-15T09:00:00Z","phase":"work","duration_sec":1500,"task":"Ship docs","task_id":3,"task_prefix":"THAK","project":"thakkali"}
{"timestamp":"2026-04-15T09:30:00Z","phase":"short_break","duration_sec":300}
{"timestamp":"2026-04-15T13:42:10Z","phase":"timer","duration_sec":4320,"task":"code review"}
```

### Schema

| Field            | Type   | Required | Meaning                                        |
|------------------|--------|----------|------------------------------------------------|
| `timestamp`      | string | yes      | RFC3339 UTC timestamp of session completion.   |
| `phase`          | string | yes      | `work`, `short_break`, `long_break`, `timer`.  |
| `duration_sec`   | int    | yes      | Session length in seconds.                     |
| `task`           | string | no       | Free-text tag or resolved task title.          |
| `task_id`        | int    | no       | Numeric ID from the task file (if `-t TSK-N`). |
| `task_prefix`    | string | no       | ID prefix (`TSK`, `THAK`, `AUTH`, …).          |
| `project`        | string | no       | `#project` tag carried from the tracked task.  |

### Notes

- **Append-only.** Thakkali never rewrites `log.jsonl`. If you want to
  prune or migrate, do it yourself with `jq`.
- **Breaks are logged too** (for Pomodoro) — `phase: short_break` /
  `long_break` — though stats currently only aggregates work phases.
- **Free-text and tracked entries coexist.** You'll see entries with
  just `task`, entries with `task` + `task_id`, and entries with none
  of the above. All are valid.

### Scripting examples

```bash
LOG=~/Library/Application\ Support/thakkali/log.jsonl

# total tracked time today (work + timer)
jq -s 'map(select(.timestamp | startswith("2026-04-15"))
        | select(.phase != "short_break" and .phase != "long_break"))
       | (map(.duration_sec) | add / 60)' $LOG

# all sessions for THAK-3
jq -c 'select(.task_id == 3)' $LOG

# entries missing task metadata (candidates for backfill)
jq -c 'select(.task_id == null and (.task // "") == "")' $LOG
```

## `config.json` — the config file

See the [config reference](config.md) for the full schema.
