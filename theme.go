package main

import (
	"fmt"
	"strconv"

	gansi "github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/glamour/styles"
)

// palette is the single source of colour. Matrix theme: greens for
// structure and emphasis, greys for body text and secondary detail.
type palette struct {
	text   string // body text: grey with a green tint
	muted  string // bullets, quotes, comments, rules
	faint  string // thin rules, table borders
	bright string // the signature matrix green: H1, H2, keywords
	green  string // H3, links, H2 rule
	soft   string // H4+, strings in code
	pale   string // bold text, functions: near-white green
	codeBg string // background for inline code and code blocks
	barBg  string // status bar background
	keyBg  string // keycaps on the help screen
	onAcc  string // text on a bright-green background
}

var matrixDark = palette{
	text:   "#B4C4B4",
	muted:  "#6B7B6B",
	faint:  "#344034",
	bright: "#00FF41",
	green:  "#00C832",
	soft:   "#5FAF6F",
	pale:   "#D0FFD0",
	codeBg: "#0C170C",
	barBg:  "#0F2410",
	keyBg:  "#1C2E1C",
	onAcc:  "#000000",
}

// Matrix on a light background: same roles, darker inks.
var matrixLight = palette{
	text:   "#253525",
	muted:  "#667566",
	faint:  "#B8C4B8",
	bright: "#00801F",
	green:  "#00661A",
	soft:   "#3F7F4F",
	pale:   "#0B3D14",
	codeBg: "#E6F2E6",
	barBg:  "#D4EBD4",
	keyBg:  "#D0E2D0",
	onAcc:  "#FFFFFF",
}

func paletteFor(style string) palette {
	if style == "light" {
		return matrixLight
	}
	return matrixDark
}

// fgSGR and bgSGR turn "#RRGGBB" into 24-bit colour escape parameters.
func rgb(hex string) (int, int, int) {
	v, _ := strconv.ParseUint(hex[1:], 16, 32)
	return int(v >> 16 & 0xff), int(v >> 8 & 0xff), int(v & 0xff)
}

func fgSGR(hex string) string {
	r, g, b := rgb(hex)
	return fmt.Sprintf("38;2;%d;%d;%d", r, g, b)
}

func bgSGR(hex string) string {
	r, g, b := rgb(hex)
	return fmt.Sprintf("48;2;%d;%d;%d", r, g, b)
}

func sp(s string) *string { return &s }
func bp(b bool) *bool     { return &b }

// styleConfig starts from glamour's dark layout (margins, bullets, indents)
// and replaces every colour with the palette.
func styleConfig(style string) gansi.StyleConfig {
	p := paletteFor(style)
	c := styles.DarkStyleConfig

	c.Document.Color = sp(p.text)
	c.BlockQuote.Color = sp(p.muted)
	c.BlockQuote.Italic = bp(true)
	c.Heading.Color = sp(p.bright) // replaced by headings.go, kept for safety
	c.H1.Color, c.H1.BackgroundColor = sp(p.onAcc), sp(p.bright)
	c.H6.Color = sp(p.soft)
	c.Strong.Color = sp(p.pale)
	c.Emph.Color = sp(p.text)
	c.Strikethrough.Color = sp(p.muted)
	c.HorizontalRule.Color = sp(p.faint)
	c.HorizontalRule.Format = "\n" + "────────────────────────────────" + "\n"
	c.Item.Color = sp(p.text)
	c.Enumeration.Color = sp(p.muted)
	c.Task.Color = sp(p.green)

	c.Link.Color = sp(p.green) // bare URLs; regular link URLs are hidden
	c.Link.Underline = bp(true)
	c.Link.Format = "{{tmdHref .text}}"
	c.LinkText.Color = sp(p.green)
	c.LinkText.Bold = bp(false)
	c.LinkText.Underline = bp(true)
	c.Image.Color = sp(p.soft)
	c.ImageText.Color = sp(p.muted)

	c.Code.Color, c.Code.BackgroundColor = sp(p.bright), sp(p.codeBg)
	c.CodeBlock.Color = sp(p.text)
	c.CodeBlock.Chroma = &gansi.Chroma{
		Text:                gansi.StylePrimitive{Color: sp(p.text)},
		Error:               gansi.StylePrimitive{Color: sp(p.pale), BackgroundColor: sp(p.faint)},
		Comment:             gansi.StylePrimitive{Color: sp(p.muted), Italic: bp(true)},
		CommentPreproc:      gansi.StylePrimitive{Color: sp(p.muted)},
		Keyword:             gansi.StylePrimitive{Color: sp(p.bright), Bold: bp(true)},
		KeywordReserved:     gansi.StylePrimitive{Color: sp(p.bright), Bold: bp(true)},
		KeywordNamespace:    gansi.StylePrimitive{Color: sp(p.green)},
		KeywordType:         gansi.StylePrimitive{Color: sp(p.green)},
		Operator:            gansi.StylePrimitive{Color: sp(p.text)},
		Punctuation:         gansi.StylePrimitive{Color: sp(p.muted)},
		Name:                gansi.StylePrimitive{Color: sp(p.text)},
		NameBuiltin:         gansi.StylePrimitive{Color: sp(p.green)},
		NameTag:             gansi.StylePrimitive{Color: sp(p.green)},
		NameAttribute:       gansi.StylePrimitive{Color: sp(p.soft)},
		NameClass:           gansi.StylePrimitive{Color: sp(p.pale), Bold: bp(true)},
		NameDecorator:       gansi.StylePrimitive{Color: sp(p.soft)},
		NameFunction:        gansi.StylePrimitive{Color: sp(p.pale)},
		LiteralNumber:       gansi.StylePrimitive{Color: sp(p.pale)},
		LiteralString:       gansi.StylePrimitive{Color: sp(p.soft)},
		LiteralStringEscape: gansi.StylePrimitive{Color: sp(p.pale)},
		GenericDeleted:      gansi.StylePrimitive{Color: sp(p.muted), CrossedOut: bp(true)},
		GenericEmph:         gansi.StylePrimitive{Italic: bp(true)},
		GenericInserted:     gansi.StylePrimitive{Color: sp(p.bright)},
		GenericStrong:       gansi.StylePrimitive{Bold: bp(true)},
		GenericSubheading:   gansi.StylePrimitive{Color: sp(p.muted)},
		Background:          gansi.StylePrimitive{BackgroundColor: sp(p.codeBg)},
	}
	c.Table.Color = sp(p.text)
	c.DefinitionDescription.Color = sp(p.text)
	return c
}
