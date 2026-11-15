package terminal

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// countingWriter records how many Write calls RenderHTML made, to prove the
// large fixtures exercise chunked flushing.
type countingWriter struct {
	strings.Builder
	writes int
}

func (w *countingWriter) Write(p []byte) (int, error) {
	w.writes++
	return w.Builder.Write(p)
}

// TestRenderHTMLChunks: a large fixture is written in several chunks and the
// joined result is still the golden output.
func TestRenderHTMLChunks(t *testing.T) {
	s, err := NewScreen()
	if err != nil {
		t.Fatal(err)
	}
	s.Write(loadFixture(t, "npm.sh", "raw"))

	var w countingWriter
	if err := s.RenderHTML(&w, true); err != nil {
		t.Fatal(err)
	}
	if w.writes < 2 {
		t.Errorf("RenderHTML wrote %d chunks, want several", w.writes)
	}
	if diff := cmp.Diff(w.String(), string(loadFixture(t, "npm.sh", "rendered"))); diff != "" {
		t.Errorf("diff (-got +want):\n%s", diff)
	}
}

func TestStrict(t *testing.T) {
	tests := []struct {
		name, input, want string
	}{
		{
			name:  "OSC 8 http link kept",
			input: "\x1b]8;;https://example.com\x1b\\hi\x1b]8;;\x1b\\",
			want:  `<a href="https://example.com">hi</a>`,
		},
		{
			name:  "OSC 8 data link unlinked",
			input: "\x1b]8;;data:text/html,x\x1b\\hi\x1b]8;;\x1b\\",
			want:  "hi",
		},
		{
			name:  "OSC 8 relative link unlinked",
			input: "\x1b]8;;/admin/delete\x1b\\hi\x1b]8;;\x1b\\",
			want:  "hi",
		},
		{
			name:  "1339 http link kept",
			input: "\x1b]1339;url=http://example.com;content=ex\a",
			want:  `<a href="http://example.com">ex</a>`,
		},
		{
			name:  "1339 vbscript link becomes escaped text",
			input: "\x1b]1339;url=vbscript:x;content=<b>\a",
			want:  "&lt;b&gt;",
		},
		{
			name:  "1338 image dropped",
			input: "a\x1b]1338;url=https://example.com/x.png\ab",
			want:  "ab",
		},
		{
			name:  "malformed element dropped, not echoed as an error line",
			input: "a\x1b]1338;nourl=<b>x</b>\ab",
			want:  "ab",
		},
		{
			name:  "1337 image dropped",
			input: "a\x1b]1337;File=name=MS5naWY=;inline=1:AA==\ab",
			want:  "ab",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s, err := NewScreen(WithStrict())
			if err != nil {
				t.Fatal(err)
			}
			s.Write([]byte(tc.input))
			if diff := cmp.Diff(s.AsHTML(), tc.want); diff != "" {
				t.Errorf("diff (-got +want):\n%s", diff)
			}
		})
	}
}
