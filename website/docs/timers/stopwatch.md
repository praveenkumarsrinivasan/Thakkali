# Stopwatch mode

Count *up* from zero instead of down. Useful when you don't know how
long a task will take and just want to log the actual duration — code
reviews, debugging, meetings, sometimes deep work.

## Basic usage

```bash
thakkali -T                                # open-ended stopwatch
thakkali -T -t "code review"               # tagged
thakkali -T -target 45m -t "debug prod"    # soft 45-minute target
thakkali -T -target 1h30m -t "design doc"  # any time.ParseDuration
thakkali -T -m -t "meeting"                # minimal + tagged
```

## The `-target` flag

`-target` sets a **soft goal** (not a cutoff). When elapsed time hits
the target:

- a notification fires and the target sound plays
- the target indicator flips from red to green
- the timer **keeps counting** — there's no auto-stop

So if you planned 45 minutes and worked 72, your log reflects 72. No
fudging.

Accepted formats: anything `time.ParseDuration` understands —
`45m`, `1h30m`, `2h`, `15m30s`.

## When sessions get logged

Unlike countdown / Pomodoro, stopwatch sessions have no natural
completion point. Thakkali logs the session when:

- You quit (++q++, ++esc++, ++ctrl+c++).
- You reset with ++r++ — the current run is flushed, then a new one starts.
- You switch modes with ++1++ / ++2++ / ++3++.

Runs under 1 second are discarded.

Each logged entry has `"phase": "timer"`:

```jsonl
{"timestamp":"2026-04-15T13:42:10Z","phase":"timer","duration_sec":4320,"task":"code review"}
```

## Seeing stopwatch stats separately

```bash
thakkali stats                  # both sections: Pomodoro then Timer
thakkali stats -m timer         # only stopwatch sessions
thakkali stats -m pomodoro      # only work-phase sessions
thakkali stats -m timer -days 14   # stopwatch, last 14 days
```

See the [stats page](../viz/stats.md) for the full output.

## When to use which

| Situation                                    | Mode                         |
|----------------------------------------------|------------------------------|
| Scheduled focus block — "45 minutes on this" | [Countdown](countdown.md)    |
| Multi-hour session with rest breaks          | [Pomodoro](pomodoro.md)      |
| Open-ended — "however long it takes"         | Stopwatch                    |
| Meeting you want to bill accurately          | Stopwatch with a soft target |
