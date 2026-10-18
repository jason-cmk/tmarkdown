package main

import (
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

type match struct {
	line, start, end int
}

// findMatches searches plain lines. Case-insensitive unless q has a capital.
func findMatches(plain [][]rune, q string) []match {
	if q == "" {
		return nil
	}
	fold := strings.ToLower(q) == q
	needle := []rune(q)
	var out []match
	for ln, line := range plain {
		for i := 0; i+len(needle) <= len(line); i++ {
			ok := true
			for j, r := range needle {
				c := line[i+j]
				if fold {
					c = unicode.ToLower(c)
				}
				if c != r {
					ok = false
					break
				}
			}
			if ok {
				out = append(out, match{ln, i, i + len(needle)})
				i += len(needle) - 1
			}
		}
	}
	return out
}

// span marks rune columns [start,end) of a plain line. on is written at the
// start, off at the end. If off is empty the span is a style highlight: on
// is re-applied after any escape code inside the span, and the line's own
// style is restored afterwards, so the rest of the line keeps its colours.
type span struct {
	start, end int
	on, off    string
}

func isSGR(seq string) bool {
	return len(seq) >= 3 && seq[1] == '[' && seq[len(seq)-1] == 'm'
}

func decorate(line string, spans []span) string {
	if len(spans) == 0 {
		return line
	}
	sort.Slice(spans, func(a, b int) bool { return spans[a].start < spans[b].start })
	var b strings.Builder
	b.Grow(len(line) + 16*len(spans))
	active := "" // SGR codes since the last reset
	col, si := 0, 0
	open := -1
	closeSpan := func() {
		s := spans[open]
		if s.off != "" {
			b.WriteString(s.off)
		} else {
			b.WriteString("\x1b[0m" + active)
		}
		open = -1
		si++
	}
	atCol := func() {
		if open >= 0 && spans[open].end == col {
			closeSpan()
		}
		// Skip spans that overlap one already drawn, or are empty.
		for open < 0 && si < len(spans) && (spans[si].start < col || spans[si].end <= spans[si].start) {
			si++
		}
		if open < 0 && si < len(spans) && spans[si].start == col {
			open = si
			b.WriteString(spans[si].on)
		}
	}
	for i := 0; i < len(line); {
		if line[i] == 0x1b {
			n := escLen(line, i)
			seq := line[i : i+n]
			b.WriteString(seq)
			if isSGR(seq) {
				if seq == "\x1b[0m" || seq == "\x1b[m" {
					active = ""
				} else {
					active += seq
				}
				if open >= 0 && spans[open].off == "" {
					b.WriteString(spans[open].on)
				}
			}
			i += n
			continue
		}
		atCol()
		_, n := utf8.DecodeRuneInString(line[i:])
		b.WriteString(line[i : i+n])
		i += n
		col++
	}
	atCol()
	if open >= 0 {
		closeSpan()
	}
	return b.String()
}
