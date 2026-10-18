# tmarkdown

A minimal markdown reader for the terminal. It shows the rendered,
human-readable version of a file in a full-screen viewer, with a green and
grey "Matrix" theme.

## Install

```
go install github.com/stephanventer/tmarkdown@latest
```

## Use

```
tmarkdown README.md          open a file
cat notes.md | tmarkdown     read from stdin
tmarkdown --plain file.md    print the rendered text and exit
```

When output is piped, tmarkdown prints instead of opening the viewer.
Flags: `--dark`, `--light` (the theme is auto-detected otherwise), `--plain`.

## Features

- Headings told apart by weight and rules instead of `#` marks
- Tables, syntax-highlighted code blocks, word wrap (max 100 columns, centred)
- Search as you type, with every match highlighted
- Heading outline you can filter and jump from
- Link URLs hidden in the text; select a link to see it and open it
- Reloads when the file changes on disk, keeping your place

## Keys

**Scrolling**

| Action | Keys |
|---|---|
| Line up / down | <kbd>↑</kbd> <kbd>↓</kbd> or <kbd>k</kbd> <kbd>j</kbd> |
| 10 lines up / down | <kbd>Shift</kbd>+<kbd>↑</kbd> <kbd>↓</kbd> or <kbd>K</kbd> <kbd>J</kbd> |
| Half page up / down | <kbd>Ctrl</kbd>+<kbd>u</kbd> <kbd>Ctrl</kbd>+<kbd>d</kbd> |
| Page up / down | <kbd>PgUp</kbd> <kbd>PgDn</kbd> or <kbd>b</kbd> <kbd>Space</kbd> |
| Top / bottom | <kbd>g</kbd> <kbd>G</kbd> or <kbd>Home</kbd> <kbd>End</kbd> |
| Scroll / scroll fast | mouse wheel, <kbd>Shift</kbd>+wheel |

**Finding things**

| Action | Keys |
|---|---|
| Search (as you type) | <kbd>/</kbd> |
| Next / previous match | <kbd>n</kbd> <kbd>N</kbd> |
| Heading outline | <kbd>o</kbd> |
| Select next / previous link | <kbd>Tab</kbd> <kbd>Shift</kbd>+<kbd>Tab</kbd> |
| Open selected link | <kbd>Enter</kbd> |

**General**

| Action | Keys |
|---|---|
| Back out one level | <kbd>Esc</kbd> |
| Reload file | <kbd>r</kbd> |
| Help | <kbd>?</kbd> |
| Quit | <kbd>q</kbd> or <kbd>Ctrl</kbd>+<kbd>c</kbd> |

Only `http`, `https` and `mailto` links are opened.

## Line spacing

Terminals set line height, not programs. In Windows Terminal, add
`"cellHeight": "1.4"` to a profile's `font` settings for airier text.

Built with [glamour](https://github.com/charmbracelet/glamour) and
[Bubble Tea](https://github.com/charmbracelet/bubbletea).
