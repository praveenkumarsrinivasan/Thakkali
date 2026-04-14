# Countdown mode

Simple countdown timer — the default. Counts down from *N* minutes to
zero, plays a sound, logs the session.

## Basic usage

```bash
thakkali                             # 25-minute countdown
thakkali -w 45                       # 45-minute block
thakkali -w 50 -t "deep work"        # tagged session — shows in stats
thakkali -w 30 -m                    # minimal mode: no logo/animation
thakkali -w 25 -S Glass              # macOS Glass sound on completion
```

## With a tracked task

```bash
thakkali -w 50 -t THAK-3
```

Thakkali will:

1. Look up `THAK-3` in your task file.
2. If it's `todo`, promote it to `doing` and stamp `@begin`.
3. Show `task: THAK-3 <title> #project` in the status line.
4. On completion, log the session with `task_id` and `project`.

See [timer ↔ task integration](../tasks/timer-integration.md) for details.

## Live interaction

Once the timer is running:

| Key                        | Action                                       |
|----------------------------|----------------------------------------------|
| ++space++                  | Pause / resume                               |
| ++plus++                   | Add 5 minutes                                |
| ++minus++                  | Subtract 5 minutes (floor 1 minute)          |
| ++r++                      | Reset the current block                      |
| ++m++                      | Toggle minimal mode                          |
| ++h++                      | Toggle the bottom help line                  |
| ++1++ / ++2++ / ++3++      | Switch to countdown / Pomodoro / stopwatch   |
| ++q++                      | Quit (writes the session if the block ended) |

## When the timer ends

- A native desktop notification fires (via `beeep`).
- On macOS, `afplay` plays the named system sound (default `Glass`);
  elsewhere a terminal beep plays.
- A single line is appended to `~/Library/Application Support/thakkali/log.jsonl`
  (or your platform's equivalent — see
  [file formats](../reference/file-formats.md)).

## Why countdown and not Pomodoro by default?

Most sessions aren't the full four-round Pomodoro ritual — they're
"sit down, focus for 25 / 45 / 60 minutes, done". The countdown is the
lightest thing that covers that case. Reach for [Pomodoro](pomodoro.md)
when you want the break cadence; reach for [stopwatch](stopwatch.md)
when you don't know how long the work will take.
