# CLI reference

Exhaustive list of every subcommand and flag. For the "how do I…"
view, see [Quickstart](../quickstart.md) and the topic pages in
[Timers](../timers/index.md) and [Tasks](../tasks/index.md).

## Top-level

```
thakkali [flags]
thakkali stats    [-days N] [-mode all|pomodoro|timer]
thakkali task     add|list|move|done|rm|bulk|show ...
thakkali todo
thakkali kanban
thakkali gantt    [-view week|month|year]
thakkali activity [-weeks N]
```

### Global flags (timer modes)

| Long          | Short | Default          | Effect                                         |
|---------------|-------|------------------|------------------------------------------------|
| `-work`       | `-w`  | 25               | Work / countdown length in minutes.            |
| `-pomodoro`   | `-p`  | off              | Enable Pomodoro cycle.                         |
| `-timer`      | `-T`  | off              | Enable stopwatch mode.                         |
| `-target`     | —     | —                | Soft goal for `-timer` (`45m`, `1h30m`, …).    |
| `-short`      | `-s`  | 5                | Pomodoro short break (minutes).                |
| `-long`       | `-l`  | 15               | Pomodoro long break (minutes).                 |
| `-rounds`     | `-r`  | 4                | Pomodoro rounds before a long break.           |
| `-task`       | `-t`  | —                | Session tag. Free text, or `TSK-N` / `THAK-N`. |
| `-minimal`    | `-m`  | off              | Hide logo and tomato animation.                |
| `-sound`      | `-S`  | `""` (beep)      | macOS system sound on completion (`Glass`, …). |
| `-version`    | `-v`  | —                | Print version and exit.                        |
| `-examples`   | `-e`  | —                | Print worked examples and exit.                |

Mutually exclusive: `-timer` and `-pomodoro`.

## `thakkali stats`

```
thakkali stats [-days N] [-mode all|pomodoro|timer]
```

| Flag              | Short | Default | Effect                                           |
|-------------------|-------|---------|--------------------------------------------------|
| `-days`           | —     | 7       | Recent-window size.                              |
| `-mode`           | `-m`  | `all`   | Filter: `all`, `pomodoro`, or `timer`.           |

See the [stats page](../viz/stats.md).

## `thakkali task`

All IDs are case-insensitive, any `[A-Z]+-N` prefix accepted.

### `task add`

```
thakkali task add "title" [-p project] [-s YYYY-MM-DD] [-d YYYY-MM-DD]
```

| Flag   | Meaning                       |
|--------|-------------------------------|
| `-p`   | `#project` tag (single).      |
| `-s`   | Planned start date.           |
| `-d`   | Planned due date.             |

### `task list`

```
thakkali task list [-state STATE] [-p PROJECT]
```

| Flag      | Default  | Values                                        |
|-----------|----------|-----------------------------------------------|
| `-state`  | `active` | `todo`, `doing`, `done`, `active`, `all`.     |
| `-p`      | —        | Only tasks tagged `#<project>`.               |

### `task move`

```
thakkali task move <id> <todo|doing|done>
```

Aliases accepted for state (case-insensitive):

| State    | Aliases                        |
|----------|--------------------------------|
| `todo`   | `t`                            |
| `doing`  | `d`, `in-progress`, `wip`      |
| `done`   | `x`                            |

### `task done`

```
thakkali task done <id>
```

Shortcut for `task move <id> done`.

### `task rm`

```
thakkali task rm <id>
```

### `task show`

```
thakkali task show <id>
```

### `task bulk`

```
thakkali task bulk
```

Opens `$EDITOR` on the file (fallbacks: `nvim`, `vim`). See the
[bulk-edit page](../tasks/bulk-edit.md).

## `thakkali todo`, `thakkali kanban`

Interactive TUIs. No flags — all interaction via keybindings, see
the [keybindings reference](keybindings.md).

## `thakkali gantt`

```
thakkali gantt [-view week|month|year]
```

| Flag    | Short | Default | Values                   |
|---------|-------|---------|--------------------------|
| `-view` | `-v`  | `month` | `week`, `month`, `year`. |

See the [gantt page](../viz/gantt.md).

## `thakkali activity`

```
thakkali activity [-weeks N]
```

| Flag     | Short | Default | Range   |
|----------|-------|---------|---------|
| `-weeks` | `-w`  | 52      | 4 – 104 |

See the [activity page](../viz/activity.md).

## Exit codes

| Code | Used by            | Meaning                                              |
|------|--------------------|------------------------------------------------------|
| 0    | all                | Success                                              |
| 1    | all                | Runtime / I/O error (file couldn't be read, etc.)    |
| 2    | `task` subcommands | Usage error — bad argument, unknown id, bad state.   |

## Version and examples

```bash
thakkali -v
# thakkali 0.2.0 (c9f4f00) built 2026-04-14T13:49:58Z
# (released builds stamp version/commit/date via -ldflags; dev builds
#  print "dev (none) built unknown")

thakkali -e
# (prints every example from every mode, styled, with keymaps at the bottom)
```
