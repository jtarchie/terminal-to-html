package terminal

import (
	"cmp"
	"strconv"
)

type style uint64

// style encoding:
// 0... ...23  24... ...47  48...57     58     59   60....63
// [fg color]  [bg color]   [flags]  element  link  [unused]
// flags = bold, faint, etc

const (
	sbFGColorX1 = style(1) << (48 + iota)
	sbFGColorX2
	sbBGColorX1
	sbBGColorX2
	sbBold
	sbFaint
	sbItalic
	sbUnderline
	sbStrike
	sbBlink
	sbElement   // meaning: this node is actually an element
	sbHyperlink // this node is styled with an OSC 8 (iTerm-style) link
)

const (
	sbFGColorX = sbFGColorX1 | sbFGColorX2
	sbBGColorX = sbBGColorX1 | sbBGColorX2
)

const (
	colorNone = uint8(iota)
	colorSGR
	color8Bit
	color24Bit
)

// Used for comparing styles - ignores the element bit, link bit, and unused bits.
const styleComparisonMask = 0x03ff_ffff_ffff_ffff

// isPlain reports if there is no style information. elements (that have no
// other style set) are also considered plain.
func (s style) isPlain() bool { return s&styleComparisonMask == 0 }

func (s style) fgColor() uint32    { return uint32(s & 0x0000_00ff_ffff) }
func (s style) fgColorType() uint8 { return uint8((s & sbFGColorX) >> 48) }
func (s style) bgColor() uint32    { return uint32((s & 0xffff_ff00_0000) >> 24) }
func (s style) bgColorType() uint8 { return uint8((s & sbBGColorX) >> 50) }
func (s style) bold() bool         { return s&sbBold != 0 }
func (s style) faint() bool        { return s&sbFaint != 0 }
func (s style) italic() bool       { return s&sbItalic != 0 }
func (s style) underline() bool    { return s&sbUnderline != 0 }
func (s style) strike() bool       { return s&sbStrike != 0 }
func (s style) blink() bool        { return s&sbBlink != 0 }
func (s style) element() bool      { return s&sbElement != 0 }
func (s style) hyperlink() bool    { return s&sbHyperlink != 0 }

func (s *style) resetFGColor() { *s = (*s &^ 0x3_0000_00ff_ffff) }
func (s *style) setFGColorSGR(v uint8) {
	*s = (*s &^ 0x3_0000_00ff_ffff) | style(v) | (style(colorSGR) << 48)
}
func (s *style) setFGColor8Bit(v uint8) {
	*s = (*s &^ 0x3_0000_00ff_ffff) | style(v) | (style(color8Bit) << 48)
}
func (s *style) setFGColor24Bit(rgb [3]uint8) {
	*s = (*s &^ 0x3_0000_00ff_ffff) | (style(rgb[0]) << 16) | (style(rgb[1]) << 8) | style(rgb[2]) | (style(color24Bit) << 48)
}

func (s *style) resetBGColor() { *s = (*s &^ 0xc_ffff_ff00_0000) }
func (s *style) setBGColorSGR(v uint8) {
	*s = (*s &^ 0xc_ffff_ff00_0000) | (style(v) << 24) | (style(colorSGR) << 50)
}
func (s *style) setBGColor8Bit(v uint8) {
	*s = (*s &^ 0xc_ffff_ff00_0000) | (style(v) << 24) | (style(color8Bit) << 50)
}
func (s *style) setBGColor24Bit(rgb [3]uint8) {
	*s = (*s &^ 0xc_ffff_ff00_0000) | (style(rgb[0]) << 40) | (style(rgb[1]) << 32) | (style(rgb[2]) << 24) | (style(color24Bit) << 50)
}

func (s *style) setBold(v bool)      { *s = (*s &^ sbBold) | booln(v, sbBold) }
func (s *style) setFaint(v bool)     { *s = (*s &^ sbFaint) | booln(v, sbFaint) }
func (s *style) setItalic(v bool)    { *s = (*s &^ sbItalic) | booln(v, sbItalic) }
func (s *style) setUnderline(v bool) { *s = (*s &^ sbUnderline) | booln(v, sbUnderline) }
func (s *style) setStrike(v bool)    { *s = (*s &^ sbStrike) | booln(v, sbStrike) }
func (s *style) setBlink(v bool)     { *s = (*s &^ sbBlink) | booln(v, sbBlink) }
func (s *style) setElement(v bool)   { *s = (*s &^ sbElement) | booln(v, sbElement) }
func (s *style) setHyperlink(v bool) { *s = (*s &^ sbHyperlink) | booln(v, sbHyperlink) }

// States for parsing extended colors, e.g. 38;5;150 or 48;2;r;g;b
const (
	colorModeNormal = iota
	colorModeGot38
	colorModeGot48
	colorModeGot38_5
	colorModeGot48_5
	colorModeGot38_2
	colorModeGot48_2
)

// May be "" for a non-plain style: 24-bit colors have no class.
func (s style) asClasses() string {
	b := make([]byte, 0, 64)
	add := func(prefix string, n uint32) {
		if len(b) > 0 {
			b = append(b, ' ')
		}
		b = append(b, prefix...)
		b = strconv.AppendUint(b, uint64(n), 10)
	}

	switch fg := s.fgColor(); s.fgColorType() {
	case colorSGR:
		if fg > 29 && fg < 38 {
			add("term-fg", fg)
		}
		if fg > 89 && fg < 98 {
			add("term-fgi", fg)
		}
	case color8Bit:
		add("term-fgx", fg)
	}

	switch bg := s.bgColor(); s.bgColorType() {
	case colorSGR:
		if bg > 39 && bg < 48 {
			add("term-bg", bg)
		}
		if bg > 99 && bg < 108 {
			add("term-bgi", bg)
		}
	case color8Bit:
		add("term-bgx", bg)
	}

	if s.bold() {
		add("term-fg", 1)
	}
	if s.faint() {
		add("term-fg", 2)
	}
	if s.italic() {
		add("term-fg", 3)
	}
	if s.underline() {
		add("term-fg", 4)
	}
	if s.blink() {
		add("term-fg", 5)
	}
	if s.strike() {
		add("term-fg", 9)
	}

	return string(b)
}

// Add colours to an existing style, returning a new style.
func (s style) color(colors []string) style {
	if len(colors) == 0 || (len(colors) == 1 && (colors[0] == "0" || colors[0] == "")) {
		// s with all normal styles masked out
		return s &^ styleComparisonMask
	}

	colorMode := colorModeNormal
	var rgb [3]uint8
	var rgbIndex int

	for _, ccs := range colors {
		// If multiple colors are defined, i.e. \e[30;42m\e then loop through each
		// one, and assign it to s.fgColor or s.bgColor
		// An omitted parameter means 0, e.g. \e[;1m is \e[0;1m.
		cc, err := strconv.ParseUint(cmp.Or(ccs, "0"), 10, 8)
		if err != nil {
			continue
		}

		// State machine for XTerm colors, eg 38;5;150
		switch colorMode {
		case colorModeGot38:
			switch cc {
			case 5:
				colorMode = colorModeGot38_5
			case 2:
				colorMode = colorModeGot38_2
				rgbIndex = 0
			default:
				colorMode = colorModeNormal
			}
			continue
		case colorModeGot48:
			switch cc {
			case 5:
				colorMode = colorModeGot48_5
			case 2:
				colorMode = colorModeGot48_2
				rgbIndex = 0
			default:
				colorMode = colorModeNormal
			}
			continue
		case colorModeGot38_5:
			s.setFGColor8Bit(uint8(cc))
			colorMode = colorModeNormal
			continue
		case colorModeGot48_5:
			s.setBGColor8Bit(uint8(cc))
			colorMode = colorModeNormal
			continue
		case colorModeGot38_2:
			rgb[rgbIndex] = uint8(cc)
			if rgbIndex == len(rgb)-1 {
				s.setFGColor24Bit(rgb)
				colorMode = colorModeNormal
				continue
			}
			rgbIndex++
			continue
		case colorModeGot48_2:
			rgb[rgbIndex] = uint8(cc)
			if rgbIndex == len(rgb)-1 {
				s.setBGColor24Bit(rgb)
				colorMode = colorModeNormal
				continue
			}
			rgbIndex++
			continue
		}

		switch cc {
		case 0:
			// Reset all styles
			s &^= styleComparisonMask
		case 1:
			s.setBold(true)
			s.setFaint(false)
		case 2:
			s.setFaint(true)
			s.setBold(false)
		case 3:
			s.setItalic(true)
		case 4:
			s.setUnderline(true)
		case 5, 6:
			s.setBlink(true)
		case 9:
			s.setStrike(true)
		case 21, 22:
			s.setBold(false)
			s.setFaint(false)
		case 23:
			s.setItalic(false)
		case 24:
			s.setUnderline(false)
		case 25:
			s.setBlink(false)
		case 29:
			s.setStrike(false)
		case 38:
			colorMode = colorModeGot38
		case 39:
			s.resetFGColor()
		case 48:
			colorMode = colorModeGot48
		case 49:
			s.resetBGColor()
		case 30, 31, 32, 33, 34, 35, 36, 37, 90, 91, 92, 93, 94, 95, 96, 97:
			s.setFGColorSGR(uint8(cc))
		case 40, 41, 42, 43, 44, 45, 46, 47, 100, 101, 102, 103, 104, 105, 106, 107:
			s.setBGColorSGR(uint8(cc))
		}
	}
	return s
}

// false, true => 0, t
func booln(b bool, t style) style {
	if b {
		return t
	}
	return 0
}
