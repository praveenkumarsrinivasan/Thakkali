# Bulk edit in `$EDITOR`

Because the task file is plain markdown, the fastest way to capture
a batch of tasks — say, from a meeting — is just to paste them in with
your editor.

```bash
thakkali task bulk
```

That opens `thakkali.md` in:

1. `$EDITOR` if set
2. otherwise `nvim`
3. otherwise `vim`

When you save and exit, Thakkali reparses the file and rewrites it —
assigning IDs to any free-text lines, auto-stamping `@begin` / `@done`
for state transitions, and preserving unknown tags verbatim.

## What you can do while bulk-editing

### Capture new tasks without worrying about IDs

Add plain lines:

```markdown
- [ ] Buy milk
- [ ] Schedule 1:1 with Alice #mgmt
- [ ] Investigate flaky test #flaky @due:2026-05-01
```

On save, they become:

```markdown
- [ ] THAK-12 Buy milk @created:2026-04-15T09:00:00Z
- [ ] THAK-13 Schedule 1:1 with Alice #mgmt @created:2026-04-15T09:00:00Z
- [ ] THAK-14 Investigate flaky test #flaky @due:2026-05-01 @created:2026-04-15T09:00:00Z
```

### Transition a bunch of tasks at once

Change `[ ]` to `[*]` on every task you're starting today, save, and
they all get `@begin` stamped in one shot.

### Add free-form notes

Any line that doesn't look like a task is preserved verbatim:

```markdown
# Thakkali tasks

<!-- sprint goals: ship docs + land v2 storage -->

- [*] THAK-3 Ship v2 storage
- [ ] THAK-4 Write docs
```

Comments, blank lines, and random prose all survive the round-trip.

### Import from another project

Paste lines with a different prefix — they keep their prefix:

```markdown
- [ ] AUTH-99 Rotate signing key #auth @due:2026-05-15
```

On save, `AUTH-99` stays `AUTH-99` (not renumbered to `THAK-N`).
See [v2 storage](../reference/file-formats.md#id-prefixes) for the
full semantics.

## What happens on save

1. Thakkali reads the entire file.
2. Every `- [ ]` / `- [*]` / `- [x]` line is parsed into a task.
3. Any task without an ID gets the next available `max(id) + 1`.
4. `@begin` / `@done` stamps are reconciled to the current state
   (see [CLI stamping rules](cli.md#task-move-and-task-done)).
5. Every task inherits the file's default prefix if its own is empty.
6. The file is rewritten. Non-task lines keep their original position
   and text.

## Concurrent editing

The bulk-edit workflow does **not** detect concurrent edits. If you
have the file open in your editor **and** run `thakkali task add` from
another shell, the later save wins.

For now, the workaround is: finish bulk edits before running other
commands. A future release may add an `mtime` check.

## Can I just open the file directly?

Yes! `thakkali task bulk` is a convenience — there's nothing magic
about the editor invocation. These two are equivalent:

```bash
thakkali task bulk
# vs.
$EDITOR ./thakkali.md && thakkali task list
```

Running any `task` subcommand afterwards triggers the same reparse +
rewrite, so IDs and stamps get reconciled either way.
