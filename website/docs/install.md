# Install

Thakkali ships as a single, dependency-free binary. Pick whichever path
is most comfortable for you.

## Homebrew (macOS, Linuxbrew)

```bash
brew install praveenkumarsrinivasan/tap/thakkali
```

That's it — `thakkali` is now on your `$PATH`.

To upgrade later:

```bash
brew upgrade thakkali
```

## go install

If you have a Go toolchain (`go` ≥ 1.21):

```bash
go install github.com/praveenkumarsrinivasan/Thakkali@latest
```

The binary lands in `$(go env GOBIN)` (commonly `~/go/bin`). Make sure
that's on your `$PATH`.

## Prebuilt binaries

Every release publishes tarballs for macOS (Intel + Apple Silicon),
Linux (amd64 + arm64), and Windows under the [releases page][releases].
Download, extract, move `thakkali` to a directory on your `$PATH`.

[releases]: https://github.com/praveenkumarsrinivasan/Thakkali/releases

```bash
# macOS Apple Silicon example
curl -L https://github.com/praveenkumarsrinivasan/Thakkali/releases/latest/download/thakkali_Darwin_arm64.tar.gz \
  | tar xz
mv thakkali /usr/local/bin/
```

## Build from source

```bash
git clone https://github.com/praveenkumarsrinivasan/Thakkali.git
cd Thakkali
go build -o thakkali .
./thakkali -v
```

## Verify the install

```bash
thakkali -v          # print version
thakkali -h          # list flags and subcommands
thakkali -e          # print worked examples for every mode
```

If `-v` reports a version, you're done — head to the
[Quickstart](quickstart.md).
