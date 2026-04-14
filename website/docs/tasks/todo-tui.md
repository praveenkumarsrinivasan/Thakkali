# `thakkali todo` — the task TUI

Interactive list view over the task file, grouped DOING / TODO / DONE.
Every mutation writes to disk immediately and auto-stamps `@begin` /
`@done` — so the file is always in sync with what you see.

```bash
thakkali todo
```

## What it looks like

```text
THAKKALI · todo
/repo/thakkali.md

DOING
  ▸ THAK-3  Ship v2 storage            (#thakkali · due 2026-04-22 · began 4m ago)

TODO
    THAK-4  Write docs                 (#thakkali)
    THAK-5  Polish activity heatmap    (#thakkali · due 2026-05-01)

DONE
    THAK-1  Scaffold task model        (#thakkali · done 2h ago)
    THAK-2  Parse markdown tasks       (#thakkali · done 1h ago)

? help · q quit
```

## Keybindings

### Navigation

| Key                               | Action                          |
|-----------------------------------|---------------------------------|
| ++j++ / ++down++                  | Move cursor down                |
| ++k++ / ++up++                    | Move cursor up                  |
| ++g++ / ++home++                  | Jump to top                     |
| ++shift+g++ / ++end++             | Jump to bottom                  |

### State

| Key                       | Action                                              |
|---------------------------|-----------------------------------------------------|
| ++space++ / ++enter++     | Cycle state: todo → doing → done → todo             |

Cycling auto-stamps `@begin` when the task enters `doing`, and both
`@begin` + `@done` when it enters `done`. Moving back to `todo` clears
both.

### Edit

| Key       | Action                                                               |
|-----------|----------------------------------------------------------------------|
| ++n++     | New task — opens an inline input; ++enter++ commits, ++esc++ cancels |
| ++e++     | Edit the selected task's title (inline input)                        |
| ++d++     | Delete the selected task                                             |

### Filter and reload

| Key   | Action                                                                 |
|-------|------------------------------------------------------------------------|
| ++slash++ | Start a live filter — typing matches title and project tags         |
| ++c++ | Clear the active filter                                                |
| ++r++ | Reload from disk (if another tool changed the file)                    |

### Help and quit

| Key                        | Action                            |
|----------------------------|-----------------------------------|
| ++question++               | Toggle the full keymap footer     |
| ++q++, ++esc++, ++ctrl+c++ | Quit                              |

## Behavior notes

- **Cursor preservation.** After every save → reload cycle, the TUI
  restores the cursor to the same task ID, even if the state change
  moved it to a different group. So cycling a task from DOING to DONE
  doesn't leave your cursor stranded.
- **Inline input.** While editing or creating, letter keys type into
  the prompt — ++n++ in edit mode becomes the letter `n`, not a new
  task. ++esc++ cancels; ++enter++ commits.
- **Live filter.** Typing after ++slash++ filters every keystroke —
  backspace shrinks the match, ++esc++ exits typing but keeps the
  filter active (use ++c++ to clear).
- **External changes.** If you edit `thakkali.md` from another tool
  while the TUI is open, press ++r++ to pick up the changes. The TUI
  doesn't watch the file.

## Tips

- Add new tasks quickly by tapping ++n++ repeatedly — after each
  ++enter++, the input clears and you're back in normal mode.
- Use ++slash++ + partial project name for a quick per-project view
  without leaving the TUI.
- If you prefer a column layout, the same keys (plus a few extras)
  work in [`thakkali kanban`](kanban.md).
