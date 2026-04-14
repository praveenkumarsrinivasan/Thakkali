# Thakkali documentation site

MkDocs Material source. Auto-deploys to GitHub Pages on push to `main`
via `.github/workflows/docs.yml`.

## Local preview

Requires Python ≥ 3.10:

```bash
cd website
python -m venv .venv
source .venv/bin/activate
pip install mkdocs-material
mkdocs serve               # http://127.0.0.1:8000 with live reload
```

## Build

```bash
mkdocs build --strict      # output to website/site/
```

CI uses the same command; `--strict` fails on broken links or warnings
so preview deploys catch regressions.

## Structure

```
website/
├── mkdocs.yml             # site config, nav, theme
└── docs/
    ├── index.md           # landing
    ├── install.md
    ├── quickstart.md
    ├── timers/            # countdown, pomodoro, stopwatch
    ├── tasks/             # CLI, todo TUI, kanban, bulk-edit, timer integration
    ├── viz/               # stats, gantt, activity
    ├── reference/         # keybindings, config, file formats, CLI
    ├── faq.md
    └── stylesheets/
        └── extra.css      # tomato-red accent overrides
```

## Adding a page

1. Create the `.md` under the appropriate section directory.
2. Add an entry to the `nav:` tree in `mkdocs.yml`.
3. Run `mkdocs serve` to check rendering.
4. Push — CI deploys.
