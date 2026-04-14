# Keybindings

Every hotkey Thakkali responds to, grouped by mode.

## Timer (countdown / Pomodoro / stopwatch)

| Key                              | Action                                                          |
|----------------------------------|-----------------------------------------------------------------|
| ++space++                        | Pause / resume                                                  |
| ++r++                            | Reset the current block (stopwatch: flushes the session first)  |
| ++s++                            | Skip to the next Pomodoro phase (Pomodoro mode only; no-op elsewhere) |
| ++plus++ / ++equal++             | Add 1 minute to the work block (or the stopwatch target)        |
| ++minus++ / ++underscore++       | Subtract 1 minute (floor: 1 minute)                             |
| ++1++                            | Switch to countdown mode                                        |
| ++2++                            | Switch to Pomodoro mode                                         |
| ++3++                            | Switch to stopwatch mode                                        |
| ++m++                            | Toggle minimal mode (hides logo + tomato animation)             |
| ++h++                            | Toggle the help line at the bottom                              |
| ++q++ / ++ctrl+c++               | Quit (writes the session log if the block ended)                |

!!! note "ESC does *not* quit the timer"
    ESC is only a quit key in the `todo` and `kanban` TUIs, and for
    canceling inline inputs. The timer only quits on ++q++ or
    ++ctrl+c++ — this is deliberate, so you can't accidentally lose a
    stopwatch session by mashing ESC.

## `thakkali todo` (list TUI)

### Navigation

| Key                         | Action                       |
|-----------------------------|------------------------------|
| ++j++ / ++down++            | Cursor down                  |
| ++k++ / ++up++              | Cursor up                    |
| ++g++ / ++home++            | Jump to top                  |
| ++shift+g++ / ++end++       | Jump to bottom               |

### State

| Key                    | Action                                      |
|------------------------|---------------------------------------------|
| ++space++ / ++enter++  | Cycle state: todo → doing → done → todo     |

### Edit

| Key       | Action                                          |
|-----------|-------------------------------------------------|
| ++n++     | New task (inline input)                         |
| ++e++     | Edit the selected task's title                  |
| ++d++     | Delete the selected task                        |

### Filter & reload

| Key            | Action                                                   |
|----------------|----------------------------------------------------------|
| ++slash++      | Live filter on title and `#project`                      |
| ++c++          | Clear the active filter                                  |
| ++r++          | Reload from disk                                         |

### Meta

| Key                         | Action                               |
|-----------------------------|--------------------------------------|
| ++question++                | Toggle the full keymap footer        |
| ++q++, ++esc++, ++ctrl+c++  | Quit                                 |

### Input mode (while typing a new / edited title)

| Key              | Action                                         |
|------------------|------------------------------------------------|
| *letters*        | Typed into the input                           |
| ++backspace++    | Delete the previous character                  |
| ++enter++        | Commit                                         |
| ++esc++          | Cancel (discards new / edit; keeps filter)     |

## `thakkali kanban` (board TUI)

All of `thakkali todo`'s keys plus:

### Column navigation

| Key                  | Action                        |
|----------------------|-------------------------------|
| ++h++ / ++left++     | Focus previous column         |
| ++l++ / ++right++    | Focus next column             |

### Move a task across columns

| Key                            | Action                                                  |
|--------------------------------|---------------------------------------------------------|
| ++greater-than++ / ++shift+l++ | Shift the selected task to the **next** column          |
| ++less-than++ / ++shift+h++    | Shift the selected task to the **previous** column      |

## On the command line

| Key                         | Action                                                  |
|-----------------------------|---------------------------------------------------------|
| ++ctrl+c++                  | Abort any running command                               |
| ++tab++                     | (not yet supported) future home for shell completion    |
