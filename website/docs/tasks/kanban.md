# `thakkali kanban` — three-column board

Same data as [`thakkali todo`](todo-tui.md), laid out as a
TODO / DOING / DONE board. The focused column has a bright border;
the others are dimmed.

```bash
thakkali kanban
```

## What it looks like

```text
THAKKALI · kanban
/repo/thakkali.md

╭──────────────────╮╭──────────────────╮╭──────────────────╮
│ TODO  (2)        ││ DOING  (1)       ││ DONE  (2)        │
│ ──────────────── ││ ──────────────── ││ ──────────────── │
│ ▸ THAK-4 Write   ││   THAK-3 Ship v2 ││   THAK-1 Scaffol │
│     #thakkali    ││     #thakkali    ││     #thakkali    │
│   THAK-5 Polish  ││                  ││   THAK-2 Parse m │
│     #thakkali    ││                  ││     #thakkali    │
╰──────────────────╯╰──────────────────╯╰──────────────────╯

h/l switch col · j/k move · </> shift · space cycle · n new · e edit · d delete
```

## Keybindings

### Column navigation

| Key                  | Action                              |
|----------------------|-------------------------------------|
| ++h++ / ++left++     | Focus previous column               |
| ++l++ / ++right++    | Focus next column                   |

### Within-column navigation

| Key                      | Action                         |
|--------------------------|--------------------------------|
| ++j++ / ++down++         | Move cursor down in the column |
| ++k++ / ++up++           | Move cursor up in the column   |
| ++g++ / ++home++         | Jump to top of column          |
| ++shift+g++ / ++end++    | Jump to bottom of column       |

### Moving tasks between columns

| Key                               | Action                                                     |
|-----------------------------------|------------------------------------------------------------|
| ++greater-than++ / ++shift+l++    | Shift the selected task to the **next** column (no wrap)   |
| ++less-than++ / ++shift+h++       | Shift the selected task to the **previous** column         |
| ++space++ / ++enter++             | Cycle through all three states (wraps: done → todo)        |

`>` / `<` are good when you want precise control ("move this to DOING,
not DONE"). `space` is good for "toggle through and stop where I want".

Either way, `@begin` / `@done` are auto-stamped or cleared to match the
new state.

### Edit / filter

Same as the [todo TUI](todo-tui.md):

| Key           | Action                                                      |
|---------------|-------------------------------------------------------------|
| ++n++         | New task — created in the **focused column's** state        |
| ++e++         | Edit the selected task's title                              |
| ++d++         | Delete the selected task                                    |
| ++slash++     | Live filter (title and project)                             |
| ++c++         | Clear the filter                                            |
| ++r++         | Reload from disk                                            |
| ++question++  | Toggle the keymap footer                                    |
| ++q++, ++esc++| Quit                                                        |

## Behavior notes

- **New-task column.** `n` creates the task in whichever column is
  focused. So if you want a new DOING task, focus DOING first.
- **Cursor follows the task.** When you shift a task across columns
  with `>` / `<`, the focus and cursor move with it — so you can chain
  `>>` to hop straight from TODO to DONE.
- **Terminal width.** Columns auto-size to `(width - 8) / 3` with a
  minimum of 18 characters each. Long titles get truncated visually
  but the underlying markdown is untouched.
- **Border color signals focus.** The focused column has a bright-red
  rounded border; the others are dim. No ambiguity about which column
  the keys apply to.

## When to reach for kanban vs. todo

| You're…                                    | Use                 |
|--------------------------------------------|---------------------|
| Scanning what's active right now           | [todo](todo-tui.md) |
| Triaging a sprint — lots of state changes  | `kanban`            |
| On a narrow terminal (< 60 cols)           | [todo](todo-tui.md) |
| Doing retro — "what did I ship this week"  | `kanban`            |
