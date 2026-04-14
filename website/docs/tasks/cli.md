# Task CLI

Seven subcommands. All of them read and write the same markdown file,
auto-stamp `@begin` / `@done`, and preserve round-trip fidelity of
hand-edited content.

```bash
thakkali task add    "title" [-p project] [-s YYYY-MM-DD] [-d YYYY-MM-DD]
thakkali task list   [-state todo|doing|done|active|all] [-p project]
thakkali task move   <id> <todo|doing|done>
thakkali task done   <id>                # alias for `move <id> done`
thakkali task rm     <id>
thakkali task show   <id>
thakkali task bulk                       # open the file in $EDITOR
```

IDs can be typed in any of these forms — the prefix is ergonomic and
case-insensitive:

```bash
thakkali task move THAK-3 doing
thakkali task move thak-3 doing
thakkali task move tsk-3  doing   # prefix mismatch is tolerated
thakkali task move 3      doing
```

## `task add`

```bash
thakkali task add "Ship phase 8" -p thakkali -d 2026-04-22
thakkali task add "Buy milk"
thakkali task add "Plan offsite" -p mgmt -s 2026-05-01 -d 2026-05-03
```

- `-p`, `-project` — the `#project` tag (one per task).
- `-s`, `-start`   — planned start date in `YYYY-MM-DD` form.
- `-d`, `-due`     — planned due date in `YYYY-MM-DD` form.

The new task is appended to the file in `todo` state with an
auto-assigned ID and a `@created` stamp.

!!! tip "Flag order is flexible"
    Title can come before or after the flags:
    `task add -p auth "Review PR"` works the same as
    `task add "Review PR" -p auth`.

## `task list`

```bash
thakkali task list                          # default: active (todo + doing)
thakkali task list -state all               # include done tasks
thakkali task list -state done              # only done
thakkali task list -state all -p thakkali   # scope by project
```

Output is grouped `DOING → TODO → DONE` and the file path is printed
at the top so you always know which file is in play.

```text
file: /repo/thakkali.md

DOING
  THAK-3  Ship v2 storage            (#thakkali · due 2026-04-22 · began 4m ago)

TODO
  THAK-4  Write docs                 (#thakkali)
  THAK-5  Polish activity heatmap    (#thakkali · due 2026-05-01)
```

## `task move` and `task done`

```bash
thakkali task move THAK-3 doing      # todo → doing, auto-stamps @begin
thakkali task move THAK-3 todo       # back to todo; @begin and @done cleared
thakkali task done THAK-3            # shortcut for `move THAK-3 done`
```

State transitions auto-stamp or clear the actuals:

| Target state | `@begin`                      | `@done`           |
|--------------|-------------------------------|-------------------|
| `todo`       | cleared                       | cleared           |
| `doing`      | stamped if missing            | cleared           |
| `done`       | stamped if missing            | stamped if missing|

So the file always reflects the current state, never a stale history.

## `task rm`

```bash
thakkali task rm THAK-3
```

Deletes the line. IDs are `max + 1`, not `count + 1`, so removing
`THAK-3` won't cause the next new task to inherit that ID — historical
log references to `THAK-3` stay meaningful.

## `task show`

```bash
thakkali task show THAK-3
```

```text
THAK-3  Ship v2 storage
  state:   doing
  project: #thakkali
  due:     2026-04-22
  began:   2026-04-15T09:00:00+12:00 (4m ago)
  created: 2026-04-15T08:55:00+12:00
```

Useful when you want the full metadata for one task without scanning
the list.

## `task bulk`

```bash
thakkali task bulk
```

Opens the task file in `$EDITOR` (fallback: `nvim`, then `vim`), waits
for you to save and exit, then re-parses and rewrites. That rewrite is
important — it assigns IDs to any free-text lines you added and
auto-stamps `@begin` / `@done` for any state changes you made by hand.

See the [bulk-edit page](bulk-edit.md) for the full round-trip behavior.

## Exit codes

| Code | Meaning                                                |
|------|--------------------------------------------------------|
| 0    | Success.                                               |
| 1    | I/O error (couldn't read / write the file).            |
| 2    | Usage error — bad argument, unknown id, invalid state. |

Useful for scripting:

```bash
if thakkali task move THAK-99 done 2>/dev/null; then
  echo "marked done"
else
  echo "no such task"
fi
```
