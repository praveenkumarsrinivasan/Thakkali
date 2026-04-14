# Pomodoro mode

Full Pomodoro cycle: work → short break → work → short break → ... →
long break, repeating across a configurable number of rounds.

## Basic usage

```bash
thakkali -p                              # 25 / 5 / 15, 4 rounds — classic
thakkali -p -t "ship phase 8"            # with a task tag (work phases only)
thakkali -p -w 50 -s 10 -l 20 -r 3       # custom cadence
thakkali -p -w 1 -s 1 -l 1 -r 2          # 5-minute smoke test
thakkali -p -m -S Hero                   # minimal Pomodoro with Hero sound
```

## Cycle structure

With `-r 4` (the default):

```
work → short → work → short → work → short → work → LONG
 1      1       2     2       3     3       4     end
```

The long break replaces the short break on the last round.

## Pomodoro-specific flags

| Flag                | Default | Effect                               |
|---------------------|---------|--------------------------------------|
| `-p`, `-pomodoro`   | off     | Enable Pomodoro mode.                |
| `-s`, `-short <m>`  | 5       | Short break length in minutes.       |
| `-l`, `-long <m>`   | 15      | Long break length in minutes.        |
| `-r`, `-rounds <n>` | 4       | Work rounds before the long break.   |
| `-w`, `-work <m>`   | 25      | Each work round's length in minutes. |

## What gets logged

Only **work** phases are logged to `log.jsonl`. Short and long breaks
are not tracked — they're ergonomic, not billable.

```jsonl
{"timestamp":"2026-04-15T09:00:00Z","phase":"work","duration_sec":1500,"task":"Ship docs"}
{"timestamp":"2026-04-15T09:30:00Z","phase":"work","duration_sec":1500,"task":"Ship docs"}
```

## Live interaction

Same keybindings as [countdown mode](countdown.md). Extras worth
highlighting:

| Key                     | Action                                                    |
|-------------------------|-----------------------------------------------------------|
| ++space++               | Pause / resume the current phase.                         |
| ++r++                   | Reset the current phase (work or break).                  |
| ++1++ / ++2++ / ++3++   | Switch mode live (countdown / Pomodoro / stopwatch).      |

Switching out of Pomodoro mid-cycle is safe — the current phase's time
to that point is logged if it was a work phase.

## Round counter

The status line shows `work (2/4)` during a work phase and
`short break` / `long break` during breaks. The round counter lets you
see at a glance how close you are to the long break.

## Tying it to a task

```bash
thakkali -p -t THAK-3
```

The task tag applies to **every work round** in the cycle — when the
cycle completes, you'll have N entries in `log.jsonl` for `THAK-3`.
Handy for "I want to spend two full Pomodoro cycles on this task".
