package main

import (
	"os"
	"strings"
	"testing"
	"time"
)

func sampleDoc(t *testing.T) *Doc {
	t.Helper()
	src, err := os.ReadFile("sample.md")
	if err != nil {
		t.Fatal(err)
	}
	d, err := buildDoc(src, 80, 3, "dark", false)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestHeadingsMapped(t *testing.T) {
	d := sampleDoc(t)
	want := []string{"tmarkdown sample", "Lists", "Quote", "Code", "Tables", "Deeper heading", "Final section"}
	if len(d.headings) != len(want) {
		t.Fatalf("got %d headings, want %d", len(d.headings), len(want))
	}
	for i, h := range d.headings {
		if h.text != want[i] || !strings.Contains(strings.ToLower(string(d.plain[h.line])), strings.ToLower(h.text)) {
			t.Errorf("heading %q mapped to line %d: %q", h.text, h.line, string(d.plain[h.line]))
		}
	}
}

func TestHeadingStyles(t *testing.T) {
	d := sampleDoc(t)
	for _, p := range d.plain {
		if strings.Contains(string(p), "##") {
			t.Fatalf("hash marks left: %q", string(p))
		}
	}
	rule := func(h heading) string { return strings.TrimSpace(string(d.plain[h.line+1])) }
	if r := rule(d.headings[1]); r != strings.Repeat("━", len("Lists")) {
		t.Errorf("H2 rule = %q", r)
	}
	if r := rule(d.headings[5]); r != strings.Repeat("─", len("Deeper heading")) {
		t.Errorf("H3 rule = %q", r)
	}
	if !strings.Contains(string(d.plain[d.headings[0].line]), "TMARKDOWN SAMPLE") {
		t.Error("H1 should be uppercase")
	}
}

func TestLinksHiddenAndLocated(t *testing.T) {
	d := sampleDoc(t)
	all := strings.Join(d.lines, "\n")
	if strings.ContainsAny(all, string([]rune{markLink, markAutoStart, markAutoEnd})) {
		t.Fatal("marker runes left in output")
	}
	if strings.Contains(all, "example.com/docs") || strings.Contains(all, "Textualize") {
		t.Error("regular link URL is visible")
	}
	if !strings.Contains(all, "https://github.com/charmbracelet/glamour") {
		t.Error("bare URL should stay visible")
	}
	if len(d.links) != 8 {
		t.Fatalf("got %d links, want 8", len(d.links))
	}
	for _, l := range d.links {
		got := string(d.plain[l.line][l.start:l.end])
		if got == "" || !strings.HasSuffix(l.text, got) || strings.HasPrefix(got, " ") {
			t.Errorf("link %q at line %d [%d,%d) covers %q", l.text, l.line, l.start, l.end, got)
		}
	}
	if d.links[4].url != "./README.md" {
		t.Errorf("local link url = %q", d.links[4].url)
	}
}

func TestSearchSmartCase(t *testing.T) {
	plain := [][]rune{[]rune("Go go GO"), []rune("nothing")}
	if n := len(findMatches(plain, "go")); n != 3 {
		t.Errorf("lowercase query: %d matches, want 3", n)
	}
	if m := findMatches(plain, "GO"); len(m) != 1 || m[0].start != 6 {
		t.Errorf("capital query: %+v", m)
	}
}

func TestDecorateKeepsColours(t *testing.T) {
	line := "\x1b[31mred\x1b[0m \x1b[32mgreen text\x1b[0m"
	got := decorate(line, []span{{start: 4, end: 9, on: "\x1b[7m"}})
	want := "\x1b[31mred\x1b[0m \x1b[32m\x1b[7mgreen\x1b[0m\x1b[32m text\x1b[0m"
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
	if string(stripANSI(got)) != "red green text" {
		t.Error("decorate changed the text")
	}
	// A reset inside the span re-applies the highlight.
	got = decorate("\x1b[31mab\x1b[0mcd", []span{{start: 1, end: 3, on: "\x1b[7m"}})
	if !strings.Contains(got, "\x1b[0m\x1b[7mc") {
		t.Errorf("highlight not re-applied after reset: %q", got)
	}
}

func TestLooksBinary(t *testing.T) {
	if looksBinary([]byte("# hi\n")) || !looksBinary([]byte("MZ\x00\x01")) {
		t.Error("binary detection wrong")
	}
}

func TestRenderSpeed(t *testing.T) {
	sample, _ := os.ReadFile("sample.md")
	var b strings.Builder
	for b.Len() < 1<<20 {
		b.Write(sample)
		b.WriteString("\n")
	}
	start := time.Now()
	if _, err := buildDoc([]byte(b.String()), 96, 2, "dark", false); err != nil {
		t.Fatal(err)
	}
	t.Logf("1 MB render: %v", time.Since(start))
}

func TestWheelAcceleration(t *testing.T) {
	if wheelStepFor(0) != wheelStep || wheelStepFor(5) != 2*wheelStep || wheelStepFor(20) != fastWheel {
		t.Error("wheel steps wrong")
	}
	m := newModel("", "x", nil, "dark")
	m.vp.Height = 10
	m.vp.SetContent(strings.Repeat("x\n", 500))
	for i := 0; i < 12; i++ { // a quick spin
		m.wheel(1, false)
	}
	if want := 3*3 + 5*6 + 4*15; m.vp.YOffset != want {
		t.Errorf("after fast spin YOffset = %d, want %d", m.vp.YOffset, want)
	}
}
