# terminal-to-html

> [!NOTE]
> This is an independent fork of [buildkite/terminal-to-html](https://github.com/buildkite/terminal-to-html). It is not affiliated with or endorsed by Buildkite. The original copyright and MIT license are retained below and in [LICENSE](LICENSE).

Terminal is a Go library for converting arbitrary shell output (with ANSI) into beautifully rendered HTML. See http://en.wikipedia.org/wiki/ANSI_escape_code for more information about ANSI Terminal Control Escape Sequences.

It provides a single command, `terminal-to-html`, that can be used to convert terminal output via STDIN, as well as via a simple web server.

## Usage

> [!WARNING]
> `terminal-to-html` is **not** hardened against cross-site scripting (XSS) or other attacks. When
> run on user-generated content, output should be passed through a HTML sanitizer prior to
> displaying in a browser.
>
> Strict mode (`-strict`, or `terminal.WithStrict()` in Go) narrows the attack surface for
> untrusted input: inline images and malformed element sequences are dropped, and links are kept only for absolute `http(s)` URLs.
> Text is always HTML-escaped. It has not been externally audited; sanitize anyway if it matters.

Piping in terminal output via the command line:

``` bash
cat fixtures/pikachu.sh.raw | terminal-to-html -preview > out.html
```

Posting terminal content via HTTP:

```bash
terminal-to-html -http=:6060 &
curl --data-binary "@fixtures/pikachu.sh.raw" http://localhost:6060/terminal > out.html
```

From Go, `Screen` is an `io.Writer`; render with `RenderHTML(w, timestamps)` /
`RenderPlainText(w, timestamps)` to stream output to an `io.Writer` without building it as one
string (or `AsHTML()` / `AsPlainText()` for a string):

```go
screen, _ := terminal.NewScreen(terminal.WithStrict())
io.Copy(screen, src)
screen.RenderHTML(dst, false)
```

Clear-screen (`ESC [2J`) pushes the current display into scrollback rather than erasing it, and
clear-scrollback (`ESC [3J`) is ignored: in a log, history is the point.

For coloring you can use the sample [terminal.css](/internal/assets/terminal.css) stylesheet and wrap the output in an element with class `term-container` (e.g. `<div class="term-container"><!-- terminal output --></div>`).

### iTerm2 Image support

Terminal has basic support for [iTerm2 inline images](http://iterm2.com/images.html). Only control sequences with `inline=1` will be rendered and `preserveAspectRatio` is not supported.

#### URL-based images

Terminal also provides a way to refer to images from the internet rather than transmitted via ANSI. The format is similar to iTerm2 inline images but uses the escape code `1338`:

`1338;url=http://imgur.com/foo.gif;width=100%;height=50px;alt=My Image`

You can use the provided `image.sh` to produce this escape sequence.

#### Links

Terminal can also render hyperlinks:

`1339;url=https://google.com;content=Google Search`

You can use the provided `link.sh` to produce this escape sequence.

Links which contain semicolons can be surrounded by either single or double quotation marks:

`1339;url='https://example.com/link-with;semicolon?argument=something';content=Example`

## Installation

With Go installed, install the `terminal-to-html` command into `$GOPATH/bin`:

```bash
$ go install github.com/jtarchie/terminal-to-html/v3/cmd/terminal-to-html@latest
```

## Developing

To get a bash prompt with all the go cross-compilation tools set up for you already:

```
$ docker build -t terminal . && docker run -it --rm -v $(pwd):/src -w /src terminal bash
```

## Benchmarking

Run `go test -bench .` to see raw Go performance. The `npm` test is the focus: this best represents the kind of use cases the original code was developed against.

## Contributing

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create a new Pull Request

## Licence

> Copyright (c) 2019 Keith Pitt, Tim Lucas, Michael Pearson
>
> MIT License
>
> Permission is hereby granted, free of charge, to any person obtaining
> a copy of this software and associated documentation files (the
> "Software"), to deal in the Software without restriction, including
> without limitation the rights to use, copy, modify, merge, publish,
> distribute, sublicense, and/or sell copies of the Software, and to
> permit persons to whom the Software is furnished to do so, subject to
> the following conditions:
>
> The above copyright notice and this permission notice shall be
> included in all copies or substantial portions of the Software.
>
> THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
> EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
> MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
> NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
> LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
> OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
> WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
