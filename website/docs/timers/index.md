# Timers

Thakkali has three time-tracking modes, all sharing the same
animation, keybindings, and log format.

| Mode                      | Command            | What it's for                              |
|---------------------------|--------------------|--------------------------------------------|
| [Countdown](countdown.md) | `thakkali`         | Single fixed-length focus block (default). |
| [Pomodoro](pomodoro.md)   | `thakkali -p`      | Full work + breaks cycle over N rounds.    |
| [Stopwatch](stopwatch.md) | `thakkali -T`      | Count up, optionally with a soft target.   |

## Common flags across all modes

| Flag                       | Effect                                                      |
|----------------------------|-------------------------------------------------------------|
| `-w`, `-work <min>`        | Length of a work block in minutes (default 25).             |
| `-t`, `-task <string|id>`  | Tag the session. Free text, or `TSK-N` for a tracked task.  |
| `-m`, `-minimal`           | Hide the logo and tomato animation.                         |
| `-S`, `-sound <name>`      | macOS system sound (`Glass`, `Ping`, `Hero`, ...) or empty. |
| `-v`, `-version`           | Print version and exit.                                     |
| `-e`, `-examples`          | Print worked examples for every mode.                       |

## Shared in-app keybindings

| Key                         | Action                                               |
|-----------------------------|------------------------------------------------------|
| ++space++                   | Pause / resume                                       |
| ++r++                       | Reset current block                                  |
| ++plus++ / ++minus++        | Add / remove 5 minutes from the current block        |
| ++1++ / ++2++ / ++3++       | Switch to countdown / Pomodoro / stopwatch live      |
| ++m++                       | Toggle minimal mode (hides logo + animation)         |
| ++h++                       | Toggle the footer help line                          |
| ++q++, ++esc++, ++ctrl+c++  | Quit                                                 |

See the full list in the [keybindings reference](../reference/keybindings.md).

## What gets logged

Every completed work session (or stopwatch segment on quit/reset) is
appended to `log.jsonl`:

```jsonl
{"timestamp":"2026-04-15T09:00:00Z","phase":"work","duration_sec":1500,"task":"Ship docs","task_id":3,"task_prefix":"THAK","project":"thakkali"}
```

That file is the authoritative source for everything in
[`thakkali stats`](../viz/stats.md) and
[`thakkali activity`](../viz/activity.md).
