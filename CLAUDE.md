# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

Go library + CLI (`github.com/buildkite/terminal-to-html/v3`) converting ANSI terminal output to HTML. Used by Buildkite to render job logs.

## Commands

- Test (root package only, what CI runs): `make test` / `go test`
- All packages: `go test ./...`
- Single test: `go test -run 'TestRendererAgainstFixtures/pikachu' .`
- Bench: `make bench` (`go test -bench . -benchmem`). The `npm` fixture benchmark is the one that matters most per README.
- Build binary: `make` (tests + `terminal-to-html` binary); cross-compile: `make dist`
- Try it: `cat fixtures/pikachu.sh.raw | go run ./cmd/terminal-to-html -preview > out.html`
- CI (`.buildkite/pipeline.yml`) runs test/bench/dist inside `docker-compose run terminal`.

The Makefile `_bindata.go` target is dead — CSS is now `go:embed`ed in `internal/assets`.

## Architecture

Pipeline: `parser` (byte-level ANSI state machine) → `Screen` (virtual terminal grid) → `output` (HTML/plain rendering).

- `parser.go` — streaming state machine (normal / escape / CSI / OSC / APC / charset modes; big comment above `parseToScreen` documents it). Handles input split across `Write` calls via `remainder`/`join`. Dispatches CSI instructions to `Screen.applyEscape`.
- `screen.go` — `Screen` holds cursor, current `style`, lines of `node`s, window size. Public API: `NewScreen(opts...)`, `Write`, `AsHTML`/`AsPlainText` (+ `...WithTimestamps`), `ScrollOutFunc`/`ScrollOutPlainFunc` callbacks. With `WithMaxSize(_, maxLines)` lines scroll out of the top and are emitted via the callback → streaming mode for huge inputs (CLI `--buffer-max-lines`). Node slices of scrolled-out lines are recycled.
- `style.go` — `style` is a bit-packed `uint64` (fg/bg color + color type, bold, italic, …, element, hyperlink flags). Style changes compare whole words; keep new attributes inside this packing and update `styleComparisonMask` accordingly.
- `node.go` — one cell: rune + style (+ element/link pointer).
- `element.go` — iTerm2 inline images (OSC 1337), URL images (1338), links (1339). `url.go` sanitizes URLs.
- `bk.go` — Buildkite APC `ESC _bk;t=…;dt=… ST` line metadata; `t` absolute ms, `dt` delta from last `t`. Rendered as a `<time datetime="…">` tag at line start (`output.go` `appendMeta`) unless timestamps disabled.
- `output.go` — HTML emission (`term-fg*`/`term-bg*`/etc. CSS classes, styles in `internal/assets/terminal.css`).
- `cmd/terminal-to-html` — CLI (urfave/cli): stdin→stdout, `-http` server mode (`/terminal`), `-preview`, `-format plain`, `-log-stats-to-stderr` (uses `internal/rusage`).
- `cmd/termplayer` — dev tool replaying raw Buildkite logs at their embedded timestamps.

Output is not XSS-hardened by design (README); don't assume it is safe HTML.

## Tests

- Golden fixtures: `fixtures/*.raw` input → `*.rendered` expected. Each fixture is checked twice: plain `Render` and streaming mode (`WithMaxSize(-1, 300)` + `ScrollOutFunc`); both must produce identical output. New fixtures must be added to `TestFiles` in `terminal_test.go`. No auto-update flag — regenerate `.rendered` manually and review the diff.
- Inline cases live in `rendererTestCases` (`terminal_test.go`) and also run in both modes.
