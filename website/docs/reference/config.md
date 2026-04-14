# Config

Thakkali persists a tiny `config.json` in your user config directory
and writes defaults on first launch.

## Where it lives

| Platform        | Path                                                                   |
|-----------------|------------------------------------------------------------------------|
| macOS           | `~/Library/Application Support/thakkali/config.json`                   |
| Linux           | `~/.config/thakkali/config.json` (respects `$XDG_CONFIG_HOME`)         |
| Windows         | `%AppData%\thakkali\config.json`                                       |

Same directory also holds `log.jsonl` and the *global* task file
(`tasks.md`).

## Schema

```json
{
  "work": 25,
  "short": 5,
  "long": 15,
  "rounds": 4,
  "sound": ""
}
```

| Field     | Type   | Default | Effect                                                           |
|-----------|--------|---------|------------------------------------------------------------------|
| `work`    | int    | 25      | Default work / countdown length in minutes.                      |
| `short`   | int    | 5       | Pomodoro short break in minutes.                                 |
| `long`    | int    | 15      | Pomodoro long break in minutes.                                  |
| `rounds`  | int    | 4       | Work rounds before a long break in Pomodoro mode.                |
| `sound`   | string | `""`    | macOS system sound (`Glass`, `Ping`, `Hero`, …). `""` = beep.    |

CLI flags override the config on a per-run basis without touching the
file:

```bash
thakkali -w 45 -S Hero       # 45-minute timer with Hero sound, just this run
```

## Editing it

Just open the file — it's plain JSON:

```bash
$EDITOR ~/Library/Application\ Support/thakkali/config.json
```

Changes take effect the next time you run `thakkali`. No "reload"
command needed.

## Deleting it

```bash
rm ~/Library/Application\ Support/thakkali/config.json
```

The next launch recreates it with defaults, so this is a safe "reset".

## Sounds on macOS

`sound` accepts any file name from `/System/Library/Sounds/`
(minus the `.aiff` extension). Common options:

```
Basso     Blow      Bottle    Frog      Funk
Glass     Hero      Morse     Ping      Pop
Purr      Sosumi    Submarine Tink
```

On Linux / Windows the field is ignored and a cross-platform terminal
beep plays instead.
