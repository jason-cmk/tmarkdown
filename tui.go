package main

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	xansi "github.com/charmbracelet/x/ansi"
)

type mode int

const (
	modeNormal mode = iota
	modeSearch      // typing a search query
	modeOutline
	modeHelp
)

const (
	maxWrap   = 100
	fastLines = 10 // Shift+arrow / J K
	fastWheel = 15 // Shift+wheel (normal wheel is 3)
)

type (
	fileChangedMsg struct{}
	reloadMsg      struct{ gen int }
	resizeMsg      struct{ gen int }
	clearFlashMsg  struct{ gen int }
)

type model struct {
	path, name string // path is "" for stdin
	src        []byte
	style      string
	th         uiTheme

	w, h  int
	pad   int
	doc   *Doc
	vp    viewport.Model
	ready bool
	mode  mode

	query   string
	prevQ   string // query before this search started, restored on Esc
	matches []match
	cur     int

	sel int // selected link, -1 for none

	filter string // outline filter
	oItems []int  // indexes into doc.headings
	oSel   int
	oTop   int

	flash     string
	flashGen  int
	reloadGen int
	resizeGen int
	deleted   bool
	errText   string
}

func newModel(path, name string, src []byte, style string) model {
	return model{path: path, name: name, src: src, style: style, sel: -1, th: newUITheme(style)}
}

func (m model) Init() tea.Cmd { return nil }

// ---- rendering and layout ----

func (m *model) rerender() {
	wrap := min(m.w-4, maxWrap)
	if wrap < 20 {
		wrap = max(m.w, 10)
	}
	m.pad = max((m.w-wrap)/2, 0)
	d, err := buildDoc(m.src, wrap, m.pad, m.style, false)
	if err != nil {
		m.errText = err.Error()
		d = &Doc{}
	}
	m.doc = d
	m.matches = findMatches(d.plain, m.query)
	if m.cur >= len(m.matches) {
		m.cur = 0
	}
	if m.sel >= len(d.links) {
		m.sel = -1
	}
	m.refresh()
}

// refresh rebuilds viewport content with search and link highlights.
func (m *model) refresh() {
	if len(bytes.TrimSpace(m.src)) == 0 {
		msg := "(empty file)"
		top := strings.Repeat("\n", max(m.vp.Height/2-1, 0))
		m.vp.SetContent(top + strings.Repeat(" ", max((m.w-len(msg))/2, 0)) + msg)
		return
	}
	spans := map[int][]span{}
	for i, mt := range m.matches {
		code := m.th.match
		if i == m.cur {
			code = m.th.current
		}
		spans[mt.line] = append(spans[mt.line], span{start: mt.start, end: mt.end, on: code})
	}
	if m.sel >= 0 {
		l := m.doc.links[m.sel]
		spans[l.line] = append(spans[l.line], span{start: l.start, end: l.end, on: m.th.link})
	}
	lines := make([]string, len(m.doc.lines))
	copy(lines, m.doc.lines)
	for ln, sp := range spans {
		lines[ln] = decorate(lines[ln], sp)
	}
	if m.sel >= 0 {
		ln := m.doc.links[m.sel].line
		if m.pad >= 2 && strings.HasPrefix(lines[ln], strings.Repeat(" ", m.pad)) {
			lines[ln] = strings.Repeat(" ", m.pad-2) + "▶" + lines[ln][m.pad-1:]
		}
	}
	m.vp.SetContent(strings.Join(lines, "\n"))
}

// anchor remembers the position as "n lines below heading i" so it
// survives re-wrapping and edits above it.
type anchor struct {
	head, off int
	text      string
	pct       float64
}

func (m *model) anchor() anchor {
	a := anchor{head: -1, pct: m.vp.ScrollPercent()}
	if i := m.currentHeading(); i >= 0 {
		h := m.doc.headings[i]
		a.head, a.off, a.text = i, m.vp.YOffset-h.line, h.text
	}
	return a
}

func (m *model) restore(a anchor) {
	if a.head >= 0 {
		hs := m.doc.headings
		idx := -1
		if a.head < len(hs) && hs[a.head].text == a.text {
			idx = a.head
		} else {
			for i, h := range hs {
				if h.text == a.text {
					idx = i
					break
				}
			}
		}
		if idx >= 0 {
			m.vp.SetYOffset(hs[idx].line + a.off)
			return
		}
	}
	total := max(m.vp.TotalLineCount()-m.vp.Height, 0)
	m.vp.SetYOffset(int(a.pct * float64(total)))
}

func (m *model) currentHeading() int {
	if m.doc == nil {
		return -1
	}
	idx := -1
	for i, h := range m.doc.headings {
		if h.line > m.vp.YOffset {
			break
		}
		idx = i
	}
	return idx
}

// show scrolls so line is visible, about a third from the top.
func (m *model) show(line int) {
	if line < m.vp.YOffset || line >= m.vp.YOffset+m.vp.Height {
		m.vp.SetYOffset(line - m.vp.Height/3)
	}
}

func (m *model) setFlash(s string) tea.Cmd {
	m.flash = s
	m.flashGen++
	gen := m.flashGen
	return tea.Tick(2*time.Second, func(time.Time) tea.Msg { return clearFlashMsg{gen} })
}

// ---- update ----

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		if !m.ready {
			m.vp = viewport.New(m.w, m.h-1)
			m.ready = true
			m.rerender()
			var cmd tea.Cmd
			if len(m.src) > largeFile {
				cmd = m.setFlash("large file")
			}
			return m, cmd
		}
		// Re-rendering a big file is slow, so wait until resizing stops.
		m.vp.Width, m.vp.Height = m.w, m.h-1
		m.resizeGen++
		gen := m.resizeGen
		return m, tea.Tick(60*time.Millisecond, func(time.Time) tea.Msg { return resizeMsg{gen} })

	case resizeMsg:
		if msg.gen == m.resizeGen {
			a := m.anchor()
			m.rerender()
			m.restore(a)
		}
		return m, nil

	case fileChangedMsg:
		m.reloadGen++
		gen := m.reloadGen
		return m, tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg { return reloadMsg{gen} })

	case reloadMsg:
		if msg.gen != m.reloadGen {
			return m, nil
		}
		return m.reload(false)

	case clearFlashMsg:
		if msg.gen == m.flashGen {
			m.flash = ""
		}
		return m, nil

	case tea.MouseMsg:
		if m.mode == modeNormal {
			// Shift+wheel scrolls faster. Some terminals turn Shift+wheel
			// into a sideways wheel, so treat left/right as fast up/down.
			if msg.Action == tea.MouseActionPress {
				switch {
				case msg.Button == tea.MouseButtonWheelLeft,
					msg.Button == tea.MouseButtonWheelUp && msg.Shift:
					m.vp.ScrollUp(fastWheel)
					return m, nil
				case msg.Button == tea.MouseButtonWheelRight,
					msg.Button == tea.MouseButtonWheelDown && msg.Shift:
					m.vp.ScrollDown(fastWheel)
					return m, nil
				}
			}
			var cmd tea.Cmd
			m.vp, cmd = m.vp.Update(msg)
			return m, cmd
		}
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		switch m.mode {
		case modeSearch:
			return m.updateSearch(msg)
		case modeOutline:
			return m.updateOutline(msg)
		case modeHelp:
			m.mode = modeNormal
			return m, nil
		}
		return m.updateNormal(msg)
	}
	return m, nil
}

func (m model) reload(manual bool) (tea.Model, tea.Cmd) {
	if m.path == "" {
		return m, m.setFlash("stdin can't be reloaded")
	}
	data, err := os.ReadFile(m.path)
	if err != nil {
		if os.IsNotExist(err) {
			m.deleted = true
			return m, m.setFlash("file deleted")
		}
		return m, m.setFlash("reload failed: " + unwrapPathErr(err).Error())
	}
	m.deleted = false
	if bytes.Equal(data, m.src) && !manual {
		return m, nil
	}
	a := m.anchor()
	m.src = data
	m.sel = -1
	m.rerender()
	m.restore(a)
	return m, m.setFlash("reloaded")
}

func (m model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "esc":
		if m.sel >= 0 {
			m.sel = -1
		} else {
			m.query, m.matches = "", nil
		}
		m.refresh()
		return m, nil
	case "/":
		m.mode, m.prevQ, m.query = modeSearch, m.query, ""
		m.matches = nil
		m.refresh()
		return m, nil
	case "n", "N":
		if len(m.matches) == 0 {
			if m.query != "" {
				return m, m.setFlash("no matches")
			}
			return m, nil
		}
		if msg.String() == "n" {
			m.cur = (m.cur + 1) % len(m.matches)
		} else {
			m.cur = (m.cur - 1 + len(m.matches)) % len(m.matches)
		}
		m.refresh()
		m.show(m.matches[m.cur].line)
		return m, nil
	case "o":
		if len(m.doc.headings) == 0 {
			return m, m.setFlash("no headings")
		}
		m.mode, m.filter = modeOutline, ""
		m.buildOutline()
		m.oSel = 0
		cur := m.currentHeading()
		for i, h := range m.oItems {
			if h == cur {
				m.oSel = i
			}
		}
		return m, nil
	case "tab", "shift+tab":
		return m.stepLink(msg.String() == "tab")
	case "enter":
		if m.sel >= 0 {
			return m, m.setFlash(openLink(m.doc.links[m.sel].url))
		}
		return m, nil
	case "r":
		return m.reload(true)
	case "?", "h":
		m.mode = modeHelp
		return m, nil
	case "g", "home":
		m.vp.GotoTop()
		return m, nil
	case "G", "end":
		m.vp.GotoBottom()
		return m, nil
	case "shift+up", "K":
		m.vp.ScrollUp(fastLines)
		return m, nil
	case "shift+down", "J":
		m.vp.ScrollDown(fastLines)
		return m, nil
	}
	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(msg)
	return m, cmd
}

func (m model) stepLink(forward bool) (tea.Model, tea.Cmd) {
	links := m.doc.links
	if len(links) == 0 {
		return m, m.setFlash("no links")
	}
	switch {
	case m.sel < 0 && forward:
		m.sel = 0
		for i, l := range links {
			if l.line >= m.vp.YOffset {
				m.sel = i
				break
			}
		}
	case m.sel < 0:
		m.sel = len(links) - 1
		for i := len(links) - 1; i >= 0; i-- {
			if links[i].line < m.vp.YOffset+m.vp.Height {
				m.sel = i
				break
			}
		}
	case forward:
		m.sel = (m.sel + 1) % len(links)
	default:
		m.sel = (m.sel - 1 + len(links)) % len(links)
	}
	m.refresh()
	m.show(links[m.sel].line)
	return m, nil
}

func (m model) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.mode, m.query = modeNormal, m.prevQ
		m.matches = findMatches(m.doc.plain, m.query)
		m.cur = 0
		m.refresh()
		return m, nil
	case tea.KeyEnter:
		m.mode = modeNormal
		if m.query != "" && len(m.matches) == 0 {
			return m, m.setFlash("no matches")
		}
		return m, nil
	case tea.KeyBackspace:
		if r := []rune(m.query); len(r) > 0 {
			m.query = string(r[:len(r)-1])
		}
	case tea.KeyRunes, tea.KeySpace:
		m.query += string(msg.Runes)
		if msg.Type == tea.KeySpace {
			m.query += " "
		}
	default:
		return m, nil
	}
	// Search as you type: jump to the first match at or below the view top.
	m.matches = findMatches(m.doc.plain, m.query)
	m.cur = 0
	for i, mt := range m.matches {
		if mt.line >= m.vp.YOffset {
			m.cur = i
			break
		}
	}
	m.refresh()
	if len(m.matches) > 0 {
		m.show(m.matches[m.cur].line)
	}
	return m, nil
}

func (m *model) buildOutline() {
	m.oItems = m.oItems[:0]
	f := strings.ToLower(m.filter)
	for i, h := range m.doc.headings {
		if f == "" || strings.Contains(strings.ToLower(h.text), f) {
			m.oItems = append(m.oItems, i)
		}
	}
	m.oSel, m.oTop = 0, 0
}

func (m model) updateOutline(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	rows := m.outlineRows()
	switch msg.Type {
	case tea.KeyEsc:
		m.mode = modeNormal
		return m, nil
	case tea.KeyEnter:
		if len(m.oItems) > 0 {
			m.vp.SetYOffset(m.doc.headings[m.oItems[m.oSel]].line)
		}
		m.mode = modeNormal
		return m, nil
	case tea.KeyUp, tea.KeyCtrlP:
		m.oSel = max(m.oSel-1, 0)
	case tea.KeyDown, tea.KeyCtrlN:
		m.oSel = min(m.oSel+1, max(len(m.oItems)-1, 0))
	case tea.KeyPgUp:
		m.oSel = max(m.oSel-rows, 0)
	case tea.KeyPgDown:
		m.oSel = min(m.oSel+rows, max(len(m.oItems)-1, 0))
	case tea.KeyBackspace:
		if r := []rune(m.filter); len(r) > 0 {
			m.filter = string(r[:len(r)-1])
			m.buildOutline()
		}
	case tea.KeyRunes, tea.KeySpace:
		m.filter += string(msg.Runes)
		if msg.Type == tea.KeySpace {
			m.filter += " "
		}
		m.buildOutline()
	}
	if m.oSel < m.oTop {
		m.oTop = m.oSel
	}
	if m.oSel >= m.oTop+rows {
		m.oTop = m.oSel - rows + 1
	}
	return m, nil
}

func (m model) outlineRows() int { return max(m.h-4, 1) }

// openLink opens web and mail links in the default app. Other schemes and
// local paths are not opened: a markdown file shouldn't be able to launch
// programs.
func openLink(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return "bad link: " + raw
	}
	switch strings.ToLower(u.Scheme) {
	case "http", "https", "mailto":
	default:
		return "not opened (only web/mail links): " + raw
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", u.String())
	case "darwin":
		cmd = exec.Command("open", u.String())
	default:
		cmd = exec.Command("xdg-open", u.String())
	}
	if err := cmd.Start(); err != nil {
		return "open failed: " + err.Error()
	}
	go cmd.Wait() //nolint:errcheck
	return "opened in browser"
}

// ---- view ----

// uiTheme holds the viewer's own colours, all taken from the palette.
type uiTheme struct {
	match, current, link string // SGR codes for highlights in the text
	bar, dim, sel, head  lipgloss.Style
}

func newUITheme(style string) uiTheme {
	p := paletteFor(style)
	esc := func(params ...string) string { return "[" + strings.Join(params, ";") + "m" }
	return uiTheme{
		match:   esc(fgSGR(p.pale), bgSGR(p.faint)),
		current: esc("1", fgSGR(p.onAcc), bgSGR(p.bright)),
		link:    esc(fgSGR(p.onAcc), bgSGR(p.green)),
		bar:     lipgloss.NewStyle().Foreground(lipgloss.Color(p.bright)).Background(lipgloss.Color(p.barBg)),
		dim:     lipgloss.NewStyle().Foreground(lipgloss.Color(p.muted)),
		sel:     lipgloss.NewStyle().Foreground(lipgloss.Color(p.onAcc)).Background(lipgloss.Color(p.green)),
		head:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(p.bright)),
	}
}

func (m model) View() string {
	if !m.ready {
		return ""
	}
	var body string
	switch m.mode {
	case modeOutline:
		body = m.outlineView()
	case modeHelp:
		body = m.helpView()
	default:
		body = m.vp.View()
	}
	return body + "\n" + m.statusBar()
}

func (m model) statusBar() string {
	var left, right string
	switch {
	case m.mode == modeSearch:
		left = "/" + m.query + "█"
		if m.query != "" {
			right = fmt.Sprintf("%d/%d", min(m.cur+1, len(m.matches)), len(m.matches))
		}
	case m.mode == modeOutline:
		left = "Outline: type to filter · ↑↓ move · Enter jump · Esc close"
	case m.flash != "":
		left = m.name + " · " + m.flash
	case m.sel >= 0:
		left = "→ " + m.doc.links[m.sel].url
		right = "Enter open · Tab next · Esc"
	default:
		left = m.name
		if m.deleted {
			left += " · file deleted"
		}
		if i := m.currentHeading(); i >= 0 {
			left += " · " + m.doc.headings[i].text
		}
		if m.query != "" && len(m.matches) > 0 {
			right = fmt.Sprintf("/%s %d/%d · ", m.query, m.cur+1, len(m.matches))
		}
		right += fmt.Sprintf("%3.f%% · ? help", m.vp.ScrollPercent()*100)
	}
	if m.errText != "" {
		left = "render error: " + m.errText
	}
	left, right = " "+left, right+" "
	room := m.w - lipgloss.Width(right) - 1
	if room < 1 {
		right, room = "", m.w
	}
	left = xansi.Truncate(left, room, "…")
	gap := max(m.w-lipgloss.Width(left)-lipgloss.Width(right), 0)
	return m.th.bar.Render(left + strings.Repeat(" ", gap) + right)
}

func (m model) outlineView() string {
	rows := m.outlineRows()
	lines := []string{"", m.th.head.Render("  Outline") + m.th.dim.Render("  "+m.filter), ""}
	cur := m.currentHeading()
	if len(m.oItems) == 0 {
		lines = append(lines, m.th.dim.Render("  no matching headings"))
	}
	for i := m.oTop; i < len(m.oItems) && i < m.oTop+rows; i++ {
		h := m.doc.headings[m.oItems[i]]
		text := strings.Repeat("  ", h.level-1) + h.text
		mark := "  "
		if m.oItems[i] == cur {
			mark = "• "
		}
		text = xansi.Truncate(mark+text, max(m.w-6, 1), "…")
		if i == m.oSel {
			text = m.th.sel.Render(" " + text + " ")
		} else {
			text = " " + text
		}
		lines = append(lines, "  "+text)
	}
	return fill(lines, m.h-1)
}

const helpText = `
  Keys

  ↑ ↓  j k            scroll a line
  Shift+↑↓  J K       scroll 10 lines
  Shift+wheel         scroll fast
  PgUp PgDn  b space  scroll a page
  Ctrl-u Ctrl-d       scroll half a page
  g G  Home End       top / bottom
  mouse wheel         scroll

  /                   search (as you type; Enter keeps it)
  n N                 next / previous match
  o                   outline of headings
  Tab Shift-Tab       select next / previous link
  Enter               open the selected link
  r                   reload the file
  Esc                 back out (clear selection, then search)
  ?  h                this help
  q  Ctrl-c           quit

  The file reloads by itself when it changes on disk.

  press any key to close`

func (m model) helpView() string {
	return fill(strings.Split(helpText, "\n"), m.h-1)
}

// fill pads or cuts lines to exactly n rows.
func fill(lines []string, n int) string {
	if len(lines) > n {
		lines = lines[:max(n, 0)]
	}
	for len(lines) < n {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}
