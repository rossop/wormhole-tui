# wormhole-tui

A terminal visualizer for the "read the string through wormholes" puzzle: walk
a string left to right, and every time you hit `(` or `)` you teleport to its
matching partner and reverse direction. Everything else you read gets appended
to the result.

For the classic example `(ed(et(oc))el)`, that walk spells out `leetcode`.

## Demo

```
🕳  Wormhole Parentheses Visualizer

╭──────────────────────────────────────────────╮
│ Result                                       │
│ so far: "leet"                               │
│ final:  "leetcode"  (step 3/6)               │
╰──────────────────────────────────────────────╯
╭──────────────────────────────────────────────╮
│ Process                                      │
│   (ed(et(oc))el)                             │
│             ^<-                              │
│      @~~~~~~@      jump 10 → 3               │
│                                               │
│ ▶ (ed(et(oc))el)                             │
│       ->^                                    │
│         @~~@       jump 6 → 9                │
╰──────────────────────────────────────────────╯

j/k step • g/G first/last • r restart • q/ctrl+c quit
```

Each block is one leg of the walk: the arrow track shows how you got to the
jump character, and the `@~~~@` tunnel line spans the exact columns between
the paren and its match. The current step is colored (cyan for moving right,
yellow for moving left, pink for the jump); earlier steps fade to gray.

## Run it

Requires Go (see `go.mod` for the version).

```
go run .
```

Type a string containing parentheses and press Enter (blank uses the default
`(ed(et(oc))el)`). Unbalanced parens are rejected with an inline error.

## Controls

| Key(s)          | Action                          |
| --------------- | -------------------------------- |
| `j` / `↓`        | Step forward one leg of the walk |
| `k` / `↑`        | Step back one leg                |
| `g`              | Jump to the first step           |
| `G`              | Jump to the last step (final result) |
| `r`              | Restart with a new string        |
| `q` / `ctrl+c`   | Quit                             |

## How it works

The whole walk is precomputed up front (`computeSegments` in `main.go`) into
a list of legs, so stepping with `j`/`k` just moves a cursor through that
list instead of re-running or animating the algorithm live. Each leg records
the arrow track leading up to its jump (or to the end of the string), the
tunnel span for the jump itself, and the result string built so far — which
is what lets the UI show progress and the final answer at the same time.

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and
[Lipgloss](https://github.com/charmbracelet/lipgloss).
