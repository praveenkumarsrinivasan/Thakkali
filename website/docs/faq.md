# FAQ

## Why another Pomodoro timer?

Because the existing options either have too few features
(barebones terminal timers with no log), too many (full-fledged desktop
apps with accounts and subscriptions), or no task integration. Thakkali
sits in the middle — one binary, plain-text state, tracks both time
and work.

## Will my task file conflict with other tools?

The file format is a subset of Obsidian's checklist syntax with a few
extra `@key:value` fields. It renders fine in any markdown previewer,
and Obsidian itself treats the tasks as regular to-do items. If you
use another task manager, you can usually import via
`- [ ] <title> #<project> @due:<date>` lines.

## Can multiple people edit `thakkali.md` at once?

Not safely from the same directory at the same time — Thakkali doesn't
lock. But because the file is plain markdown, `git` handles concurrent
edits well: commit your changes, pull, resolve any merge conflicts
(which will look like normal text conflicts), push. That's the
intended workflow for team use.

## Will my `log.jsonl` get huge?

At 10 sessions per day, you'll accumulate ~3,650 lines per year —
maybe 0.5 MB. Stats and activity scan the whole file but neither
operation is expensive until you're in the millions. If you ever want
to prune, `jq` to a date range and rewrite.

## How do I back up my data?

Three files matter:

- `thakkali.md` (possibly per-project)
- `~/Library/Application Support/thakkali/log.jsonl` (and the global
  `tasks.md` if you use it)
- `~/Library/Application Support/thakkali/config.json`

All plain text. Commit, rsync, or copy them anywhere.

## Can I change the tomato color / animation?

Not via config (yet). The palette and animation are hardcoded to keep
the binary dependency-free and the look consistent. If you want
something different, the whole app is ~1,400 lines of single-file Go —
fork and tweak.

## Why does `thakkali task add TSK-3 doing` not move a task?

That's the old syntax (from early drafts). Use `thakkali task move`:

```bash
thakkali task move TSK-3 doing
```

`task add` only creates new tasks.

## Why is my new task showing up as `THAK-1` and not `TSK-1`?

Because you're in a directory called `Thakkali/` (or similar). Thakkali
derives the ID prefix from the containing directory when the task file
is project-local. See
[ID prefixes](reference/file-formats.md#id-prefixes) for the full rules.

## Can I rebind keys in the TUIs?

Not currently. Keybindings are hardcoded to match vim conventions
(`h j k l`, `g G`, `/`, `?`). Happy to consider feature requests.

## Does it work on Windows?

Yes — the binary cross-compiles cleanly. `afplay` sounds won't work
(those are macOS-only) but the cross-platform beep fallback kicks in.

## The gantt chart is wrong / looks weird in my terminal

Make sure your terminal supports Unicode block characters (`█`, `─`,
`│`, `┬`) and ANSI colors. Most modern terminals do. If you see
`?` or `□` instead, your font is missing glyphs — try JetBrains Mono,
Fira Code, or any "Nerd Font" variant.

## How do I uninstall?

```bash
# Homebrew
brew uninstall thakkali

# go install
rm "$(which thakkali)"

# Manual install
rm /usr/local/bin/thakkali     # or wherever you put it

# Optional: remove state
rm -rf ~/Library/Application\ Support/thakkali
```

## Where do I report bugs / feature requests?

GitHub Issues:
<https://github.com/praveenkumarsrinivasan/Thakkali/issues>

Pull requests welcome — the whole app is in one `main.go` so there's
no architecture to learn.
