# tmarkdown

A minimal markdown reader for the terminal. It shows the rendered,
human-readable version of a file in a full-screen viewer, with a green and
grey "Matrix" theme.

## Install

tmarkdown is installed with Go (version 1.27 or newer). Install Go first,
then tmarkdown.

### 1. Install Go

**macOS**

```
brew install go
echo 'export PATH="$PATH:$HOME/go/bin"' >> ~/.zshrc
source ~/.zshrc
```

**Linux**

```
sudo apt install golang-go      # Debian, Ubuntu
sudo dnf install golang         # Fedora
sudo pacman -S go               # Arch
```

**Windows**

```
winget install GoLang.Go
```

Open a new terminal afterwards and check it worked:

```
go version
```

### 2. Install tmarkdown

```
go install github.com/stephanventer/tmarkdown@latest
```

Check it with `tmarkdown --version` or `tmarkdown --help`.

### Check the version

```
tmarkdown --version
```

This prints something like `tmarkdown v0.1.2`. To see the newest release,
look at the [releases page](https://github.com/stephanventer/tmarkdown/releases).

### Update

Run the install command again to get the latest release:

```
go install github.com/stephanventer/tmarkdown@latest
```

Just after a release the version may not change; install it directly with `@v0.1.2` instead of `@latest`.

## Use

```
tmarkdown README.md          open a file
cat notes.md | tmarkdown     read from stdin
tmarkdown --plain file.md    print the rendered text and exit
```

When output is piped, tmarkdown prints instead of opening the viewer.
Flags: `--dark`, `--light` (the theme is auto-detected otherwise), `--plain`,
`--double` (start with double line spacing).

## Features

- Headings set apart by weight and underlines, not `#` marks
- Tables, task lists, blockquotes and syntax-highlighted code blocks
- Word wrap at up to 100 columns, centred
- Search as you type, with every match highlighted
- Heading outline you can filter and jump from
- Links stay clean in the text; select one to see its URL and open it
- Reloads when the file changes on disk, keeping your place
- Dark and light themes, auto-detected
- Reads from a file or stdin; prints plain text when piped
- Optional double line spacing

## Keys

**Scrolling**

| Action | Keys |
|---|---|
| Line up / down | <kbd>↑</kbd> <kbd>↓</kbd> or <kbd>k</kbd> <kbd>j</kbd> |
| 10 lines up / down | <kbd>Shift</kbd>+<kbd>↑</kbd> <kbd>↓</kbd> or <kbd>K</kbd> <kbd>J</kbd> |
| Half page up / down | <kbd>Ctrl</kbd>+<kbd>u</kbd> <kbd>Ctrl</kbd>+<kbd>d</kbd> |
| Page up / down | <kbd>PgUp</kbd> <kbd>PgDn</kbd> or <kbd>b</kbd> <kbd>Space</kbd> |
| Top / bottom | <kbd>g</kbd> <kbd>G</kbd> or <kbd>Home</kbd> <kbd>End</kbd> |
| Scroll | mouse wheel |
| Scroll fast | <kbd>Alt</kbd>+wheel |

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
| Double / single spacing | <kbd>s</kbd> |
| Help | <kbd>?</kbd> or <kbd>h</kbd> |
| Quit | <kbd>q</kbd> or <kbd>Ctrl</kbd>+<kbd>c</kbd> |

Only `http`, `https` and `mailto` links are opened.

## Line spacing

Press <kbd>s</kbd> in the viewer for double spacing: a blank row goes
between neighbouring lines, while headings stay joined to their rules.
Press it again for single spacing, or start double-spaced with `--double`.

Built with [glamour](https://github.com/charmbracelet/glamour) and
[Bubble Tea](https://github.com/charmbracelet/bubbletea).
