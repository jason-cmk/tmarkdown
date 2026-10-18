# tmarkdown sample

A paragraph with **bold**, *italic*, `inline code` and ~~strike~~. Here is a
[regular link](https://example.com/docs) in the middle of a sentence, and a
[second link with a much longer label that should wrap across lines](https://go.dev/doc/effective_go).
A bare URL: https://github.com/charmbracelet/glamour and an email <someone@example.com>.
An [anchor link](#tables) and a [local file](./README.md).

## Lists

- First item
- Second item with a [nested link](https://pkg.go.dev)
  - Nested item
- [x] Done task
- [ ] Open task

1. One
2. Two
3. Three

## Quote

> Markdown is intended to be as easy-to-read and easy-to-write as is feasible.
> — John Gruber

## Code

```go
package main

import "fmt"

func main() {
	fmt.Println("hello, world")
}
```

```python
def greet(name: str) -> str:
    return f"hello, {name}"
```

## Tables

| Tool      | Language | Link                                   |
|-----------|----------|----------------------------------------|
| glamour   | Go       | [repo](https://github.com/charmbracelet/glamour) |
| rich      | Python   | [repo](https://github.com/Textualize/rich) |

### Deeper heading

Some text under a level-three heading.

---

## Final section

The end.
