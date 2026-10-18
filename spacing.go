package main

import (
	"sort"
	"strings"
)

// spaceLines adds a blank row between neighbouring text lines for airier
// reading. Lines that already have a blank next to them get no extra gap,
// and a heading stays joined to its own rule or wrapped lines. rows[i] is
// the row where doc line i ends up.
func spaceLines(lines []string, plain [][]rune, headings []heading, on bool) (out []string, rows []int) {
	rows = make([]int, len(lines))
	if !on {
		for i := range rows {
			rows[i] = i
		}
		return lines, rows
	}
	blank := func(i int) bool { return strings.TrimSpace(string(plain[i])) == "" }
	// Lines inside a heading block (text, wraps, rule) keep no gap after them.
	joined := map[int]bool{}
	for _, h := range headings {
		for i := h.line; i+1 < len(plain) && !blank(i+1); i++ {
			joined[i] = true
		}
	}
	out = make([]string, 0, len(lines)*2)
	for i, l := range lines {
		rows[i] = len(out)
		out = append(out, l)
		if i+1 < len(lines) && !blank(i) && !blank(i+1) && !joined[i] {
			out = append(out, "")
		}
	}
	return out, rows
}

// toRow turns a doc line into a viewport row.
func (m *model) toRow(line int) int {
	if len(m.rows) == 0 {
		return line
	}
	return m.rows[min(max(line, 0), len(m.rows)-1)]
}

// toLine turns a viewport row into the doc line shown there (or just above).
func (m *model) toLine(row int) int {
	if len(m.rows) == 0 {
		return row
	}
	return max(sort.Search(len(m.rows), func(i int) bool { return m.rows[i] > row })-1, 0)
}
