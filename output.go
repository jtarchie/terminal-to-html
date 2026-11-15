package terminal

import (
	"bytes"
	"html"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// []byte rather than strings.Builder so lines can be trimmed in place and the buffer reused.
type outputBuffer []byte

func (b *outputBuffer) WriteString(s string) { *b = append(*b, s...) }
func (b *outputBuffer) WriteRune(r rune)     { *b = utf8.AppendRune(*b, r) }

func (b *outputBuffer) appendAnchor(url string) {
	b.WriteString(`<a href="`)
	b.WriteString(html.EscapeString(sanitizeURL(url)))
	b.WriteString(`">`)
}

// Append a character to our outputbuffer, escaping HTML bits as necessary.
func (b *outputBuffer) appendChar(char rune) {
	switch char {
	case '&':
		b.WriteString("&amp;")
	case '\'':
		b.WriteString("&#39;")
	case '<':
		b.WriteString("&lt;")
	case '>':
		b.WriteString("&gt;")
	case '"':
		b.WriteString("&quot;")
	case '/':
		b.WriteString("&#47;")
	default:
		b.WriteRune(char)
	}
}

// Last part's t wins.
func lineTimestamp(parts []screenLine) (time.Time, bool) {
	var t string
	for _, l := range parts {
		if v, ok := l.metadata[bkNamespace]["t"]; ok {
			t = v
		}
	}
	millis, err := strconv.ParseInt(t, 10, 64)
	if err != nil {
		return time.Time{}, false
	}
	return time.UnixMilli(millis).UTC(), true
}

// Wrapped screen lines (newline unset) continue the logical line.
func logicalLineLen(lines []screenLine) int {
	for i, l := range lines {
		if l.newline {
			return i + 1
		}
	}
	return len(lines)
}

// Assumes all parts but the last are wrapped; output ends in \n.
func (b *outputBuffer) appendLine(parts []screenLine, timestamps bool) {
	start := len(*b)

	if timestamps {
		if ts, ok := lineTimestamp(parts); ok {
			// One of the formats accepted by the <time> tag:
			datetime := ts.Format("2006-01-02T15:04:05.999Z")
			b.WriteString(`<time datetime="`)
			b.WriteString(datetime)
			b.WriteString(`">`)
			b.WriteString(datetime)
			b.WriteString(`</time>`)
		}
	}

	// tagStack is used as a stack of open tags, so they can be closed in the
	// right order. We only have two kinds of tag, so the stack should be tiny,
	// but the algorithm can be extended later if needed.
	tagStack := make([]int, 0, 2)
	const (
		tagAnchor = iota
		tagSpan
	)

	closeFrom := func(idx int) {
		for i := len(tagStack) - 1; i >= idx; i-- {
			switch tagStack[i] {
			case tagAnchor:
				b.WriteString("</a>")
			case tagSpan:
				b.WriteString("</span>")
			}
		}
		tagStack = tagStack[:idx]
	}

	// The zero value for node has a plain style and no hyperlink.
	var previous node

	for _, l := range parts {
		for x, current := range l.nodes {
			// A set of flags for which tags need changing.
			tagChanged := [...]bool{
				// The anchor tag needs changing if the link "style" has changed,
				// or if they are both links the link URLs are different.
				// (Note that the x-1 index into the hyperlinks map returns "".)
				tagAnchor: current.style.hyperlink() != previous.style.hyperlink() ||
					(current.style.hyperlink() && l.hyperlinks[x-1] != l.hyperlinks[x]),

				// The span tag needs changing if the style has changed.
				tagSpan: !current.hasSameStyle(previous),
			}

			// HTML tags can't interleave, so closing a changed tag closes everything opened after it.
			closeFromIdx := len(tagStack)
			for i, ot := range tagStack {
				if tagChanged[ot] {
					closeFromIdx = i
					break
				}
			}

			closeFrom(closeFromIdx)

			if !slices.Contains(tagStack, tagAnchor) && current.style.hyperlink() {
				b.appendAnchor(l.hyperlinks[x])
				tagStack = append(tagStack, tagAnchor)
			}
			// 24-bit colors have no class; skip the empty span.
			if !slices.Contains(tagStack, tagSpan) && !current.style.isPlain() {
				if classes := current.style.asClasses(); classes != "" {
					b.WriteString(`<span class="`)
					b.WriteString(classes)
					b.WriteString(`">`)
					tagStack = append(tagStack, tagSpan)
				}
			}

			if current.style.element() {
				b.WriteString(l.elements[current.blob].asHTML())
			} else {
				b.appendChar(current.blob)
			}

			previous = current
		}
	}

	closeFrom(0)

	line := bytes.TrimRight((*b)[start:], " \t")
	*b = (*b)[:start+len(line)]
	if len(line) == 0 {
		b.WriteString("&nbsp;")
	}
	b.WriteString("\n")
}

// Plain-text twin of appendLine.
func lineToPlain(parts []screenLine, timestamps bool) string {
	var buf strings.Builder

	if timestamps {
		if ts, ok := lineTimestamp(parts); ok {
			buf.WriteString(ts.Format("2006-01-02T15:04:05Z"))
			buf.WriteString(" ")
		}
	}

	for _, l := range parts {
		for _, node := range l.nodes {
			if !node.style.element() {
				buf.WriteRune(node.blob)
			}
		}
	}

	return strings.TrimRight(buf.String(), " \t") + "\n"
}
