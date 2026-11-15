/*
Package terminal converts ANSI input to HTML output.

The generated HTML needs to be used with the stylesheet at
https://raw.githubusercontent.com/jtarchie/terminal-to-html/main/internal/assets/terminal.css
and wrapped in a term-container div.

You can call this library from the command line with terminal-to-html:
go install github.com/jtarchie/terminal-to-html/v3/cmd/terminal-to-html@latest
*/
package terminal

import "fmt"

// Render converts ANSI to HTML and returns the result.
func Render(input []byte) string {
	screen, err := NewScreen()
	if err != nil {
		// This shouldn't happen! (famous last words)
		panic(fmt.Sprintf("NewScreen error: %v", err))
	}
	screen.Write(input)
	return screen.AsHTML()
}
