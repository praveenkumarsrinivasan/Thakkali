# Quickstart

A 30-second tour: timer, task, stats. You'll come out the other side
understanding the whole app.

## 1. Run a timer

```bash
thakkali                 # 25-minute countdown, Ghostty-style tomato
```

Inside the app:

- ++space++ pauses / resumes
- ++plus++ / ++minus++ adds or removes 5 minutes
- ++m++ toggles minimal mode (hides the logo and animation)
- ++q++ quits

When the timer finishes, Thakkali appends a line to `log.jsonl` and
plays a sound.

## 2. Tag your session

```bash
thakkali -w 45 -t "ship the docs"
```

Now the session shows up under "top tasks" in `thakkali stats`.

## 3. Capture a real task

```bash
thakkali task add "Ship the docs site" -p thakkali -d 2026-04-22
thakkali task add "Review auth PR"     -p auth     -d 2026-05-01
thakkali task list
```

That creates `./thakkali.md` in your current directory — commit it to
your repo and your colleagues see the same list. The IDs are scoped to
the containing directory name, so in a folder called `Thakkali/` you'll
get `THAK-1`, `THAK-2`, … (`TSK-N` in the global task file).

## 4. Tie a timer to that task

```bash
thakkali -p -t THAK-1
```

Thakkali:

- resolves `THAK-1` to **Ship the docs site** (the title shows up in-app)
- auto-promotes the task from `todo` → `doing` and stamps `@begin`
- records `task_id` + `project` in `log.jsonl` so stats can roll up
  time per task and per project.

When you finish, mark it done:

```bash
thakkali task done THAK-1
```

## 5. See what you've done

```bash
thakkali stats                       # today + last 7 days, both modes
thakkali stats -days 30 -m pomodoro  # Pomodoro-only, wider window
thakkali gantt -view week            # planned vs. actual over 14 days
thakkali activity                    # GitHub-style 52-week heatmap
```

## 6. Browse your tasks interactively

```bash
thakkali todo     # list TUI: j/k, space, n, e, d, /, ?, q
thakkali kanban   # three-column board: h/l switch col, </> shift task
```

## Where next?

- **[Timers](timers/index.md)** — every flag, every mode.
- **[Tasks](tasks/index.md)** — full CRUD, the TUIs, bulk edit.
- **[Visualization](viz/index.md)** — stats / gantt / activity in depth.
- **[Reference](reference/index.md)** — flags, keybindings, file formats.
