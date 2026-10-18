package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	xansi "github.com/charmbracelet/x/ansi"
)

// Terminals can't change font size, so heading levels are told apart by
// weight and rules instead of glamour's "## " marks:
//
//	H1  full-width coloured banner, uppercase
//	H2  bold, heavy rule (━) underneath
//	H3  bold, thin rule (─) underneath
//	H4+ bold in a softer colour
type headingLook struct {
	text string // SGR for the heading text
	rule string // SGR for the rule; "" means no rule
	char string // rule character
}

func headingLooks(style string) map[int]headingLook {
	p := paletteFor(style)
	esc := func(params ...string) string { return "\x1b[" + strings.Join(params, ";") + "m" }
	return map[int]headingLook{
		1: {text: esc("1", fgSGR(p.onAcc), bgSGR(p.bright))},
		2: {text: esc("1", fgSGR(p.bright)), rule: esc(fgSGR(p.green)), char: "━"},
		3: {text: esc("1", fgSGR(p.green)), rule: esc(fgSGR(p.faint)), char: "─"},
		4: {text: esc("1", fgSGR(p.soft))},
	}
}

// styleHeadings replaces glamour's heading lines with our own and shifts
// the line numbers of everything below. width is glamour's wrap width,
// which includes its 2-column margin on each side.
func styleHeadings(d *Doc, width, pad int, style string) {
	looks := headingLooks(style)
	indent := strings.Repeat(" ", pad+2)
	inner := max(width-4, 10)

	for hi := len(d.headings) - 1; hi >= 0; hi-- {
		h := d.headings[hi]
		look, ok := looks[h.level]
		if !ok {
			look = looks[4]
		}
		// The heading may have wrapped over several lines; it ends at a blank.
		first, last := h.line, h.line
		for last+1 < len(d.plain) && strings.TrimSpace(string(d.plain[last+1])) != "" {
			last++
		}

		var out []string
		if h.level == 1 {
			for _, l := range strings.Split(xansi.Wordwrap(strings.ToUpper(h.text), inner-2, ""), "\n") {
				l = " " + l + strings.Repeat(" ", max(inner-1-lipgloss.Width(l), 0))
				out = append(out, indent+look.text+l+"\x1b[0m")
			}
		} else {
			wrapped := strings.Split(xansi.Wordwrap(h.text, inner, ""), "\n")
			widest := 0
			for _, l := range wrapped {
				out = append(out, indent+look.text+l+"\x1b[0m")
				widest = max(widest, lipgloss.Width(l))
			}
			if look.rule != "" {
				out = append(out, indent+look.rule+strings.Repeat(look.char, widest)+"\x1b[0m")
			}
		}

		plain := make([][]rune, len(out))
		for i, l := range out {
			plain[i] = stripANSI(l)
		}
		n := last - first + 1
		shift := len(out) - n
		d.lines = append(d.lines[:first], append(out, d.lines[last+1:]...)...)
		d.plain = append(d.plain[:first], append(plain, d.plain[last+1:]...)...)

		for j := hi + 1; j < len(d.headings); j++ {
			d.headings[j].line += shift
		}
		// Links inside a heading lose their position; links below move down.
		kept := d.links[:0]
		for _, l := range d.links {
			switch {
			case l.line > last:
				l.line += shift
				kept = append(kept, l)
			case l.line < first:
				kept = append(kept, l)
			}
		}
		d.links = kept
	}
}
