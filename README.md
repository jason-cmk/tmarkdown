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

| Key | Action |
|---|---|
| ↑ ↓ / j k | scroll a line |
| Shift+↑ ↓ / J K | scroll 10 lines |
| PgUp PgDn / b space, Ctrl-u Ctrl-d | page / half page |
| g G / Home End | top / bottom |
| mouse wheel, Shift+wheel | scroll, scroll fast |
| `/` then `n` `N` | search, next / previous match |
| `o` | heading outline |
| Tab Shift-Tab, Enter | select a link, open it |
| `r` | reload |
| Esc | back out one level |
| `?` | help |
| `q` Ctrl-c | quit |

Only `http`, `https` and `mailto` links are opened.

## Line spacing

Terminals set line height, not programs. In Windows Terminal, add
`"cellHeight": "1.4"` to a profile's `font` settings for airier text.

Built with [glamour](https://github.com/charmbracelet/glamour) and
[Bubble Tea](https://github.com/charmbracelet/bubbletea).
