# Phase 7 — Distribution

Ship Thakkali as a one-line install for macOS, Linux, and Windows so the author and colleagues can install and update it without building from source.

## Goals

- Users on macOS can run `brew install <tap>/thakkali` and get a working binary.
- Users on any platform can grab a prebuilt archive from the GitHub Releases page.
- A new release is triggered by pushing a git tag (`vX.Y.Z`) — no manual build steps.
- Versioning is baked into the binary (`thakkali -version` reports the tag and commit).

## Non-goals (for this phase)

- Linux packages (`.deb`, `.rpm`, AUR, snap, flatpak). Tarball is enough for now.
- Signed / notarised macOS builds. Users will `chmod +x` or accept Gatekeeper prompts on first run.
- Auto-update from within the app. Users upgrade via `brew upgrade` or re-downloading.
- Publishing the Go module for `go install`. The binary path is the only supported install route.

## Blockers to confirm before implementation

1. **License choice** — likely MIT. Required because GoReleaser bundles `LICENSE` into each archive and Homebrew formulae expect one.
2. **Module path** — `go.mod` says `github.com/praveensrinivasan/thakkali`; actual repo is `github.com/praveenkumarsrinivasan/Thakkali`. Align for cleanliness.
3. **Homebrew tap repo** — decide whether to set up the tap in this phase (see "Delivery paths" below).

## Delivery paths

**Path A — binaries only.** GoReleaser builds cross-platform archives and publishes them to GitHub Releases on tag push. Users download, extract, put the binary on `PATH`. No tap, no formula.

**Path B — full tap (recommended).** Path A plus: GoReleaser pushes an auto-generated Homebrew formula to a companion repo (e.g. `praveenkumarsrinivasan/homebrew-thakkali`) on each release. Colleagues on macOS install with:

```bash
brew tap praveenkumarsrinivasan/thakkali
brew install thakkali
```

Path B is the original distribution story from Phase 0 — only picking between A and B for *when* to do it, not whether.

## Deliverables

1. **`LICENSE`** at repo root (MIT, author "Praveen Kumar Srinivasan", year 2026).
2. **`.goreleaser.yaml`** configuring:
   - Builds: `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64`, `windows/amd64`.
   - Archives: `tar.gz` for Unix, `zip` for Windows; include `README.md` and `LICENSE`.
   - `ldflags` injecting version, commit, build date into `main.go`.
   - Checksums file (`thakkali_<version>_checksums.txt`).
   - (Path B only) `brews:` section pointing at the tap repo.
3. **`.github/workflows/release.yml`**:
   - Triggers on tag push matching `v*`.
   - Sets up Go, checks out the repo, runs `goreleaser release --clean`.
   - Uses `GITHUB_TOKEN` for release upload; `HOMEBREW_TAP_GITHUB_TOKEN` (PAT) for tap push if Path B.
4. **`-version` flag in `main.go`** wired to the `ldflags`-injected variables.
5. **Homebrew tap repo** (Path B only) — a fresh empty repo named `homebrew-thakkali`. GoReleaser populates `Formula/thakkali.rb` automatically.
6. **README updates**:
   - Replace "Coming in a future release" with the real `brew install` command.
   - Add a "Download" section pointing at GitHub Releases for non-mac users.

## Implementation steps

1. Add `LICENSE`.
2. Fix the module path in `go.mod`; run `go mod tidy`; update the import path if it appears anywhere.
3. Add version variables to `main.go` and a `-version` flag.
4. Write `.goreleaser.yaml`. Test locally with `goreleaser release --snapshot --clean` to confirm archives build for all targets.
5. Write `.github/workflows/release.yml`.
6. For Path B: create the `homebrew-thakkali` repo on GitHub; create a PAT with `repo` scope; add it as `HOMEBREW_TAP_GITHUB_TOKEN` in the Thakkali repo's Actions secrets; add the `brews:` block to `.goreleaser.yaml`.
7. Commit everything. Tag `v0.1.0`. Push the tag. Watch the Actions run.
8. Verify:
   - Release page shows archives and checksums for every target.
   - `brew install praveenkumarsrinivasan/thakkali/thakkali` (Path B) installs and runs.
   - `thakkali -version` prints the tag.

## Risks and mitigations

- **GoReleaser config drift** — keep the config minimal; regenerate with `goreleaser init` only once and diff carefully on upgrades.
- **Tap PAT expiry** — fine-grained tokens expire in ≤1 year. Set a calendar reminder; rotate before expiry.
- **macOS Gatekeeper** — unsigned binaries trigger a quarantine prompt on first run. Acceptable for a personal project; document the `xattr -d com.apple.quarantine` workaround in README if anyone hits it.
- **Windows build surface** — untested path. First release may need a follow-up fix for CRLF line endings or path handling; none expected but noting it.

## Follow-ups (not in this phase)

- Signed macOS builds via an Apple Developer ID (eliminates Gatekeeper prompt).
- Linux package repos (Debian / Arch).
- `thakkali upgrade` self-updater.
- Beta / pre-release channel via separate tap.
