# Aho-Corasick

A fast, feature-rich Go implementation of the [Aho-Corasick](https://en.wikipedia.org/wiki/Aho%E2%80%93Corasick_algorithm) string matching algorithm.

## Features

- **Multiple matching modes** — `StandardMatch`, `LeftMostFirstMatch`, and `LeftMostLongestMatch`
- **DFA and NFA backends** — choose between the deterministic (faster, more memory) and non-deterministic (smaller) automaton via the `DFA` option
- **Overlapping matches** — find all overlapping occurrences via `IterOverlapping`
- **Streaming / iterator API** — iterate through matches one at a time with `Iter` / `IterByte`
- **Byte-level API** — work directly with `[]byte` for zero-allocation use cases
- **Find + Replace** — `Replacer` supports `ReplaceAll`, `ReplaceAllFunc`, and `ReplaceAllWith`
- **Prefilter** — built-in rare-byte prefilter for fast skipping over non-matching regions
- **Whole-word matching** — optional `MatchOnlyWholeWords` filtering
- **ASCII case-insensitive matching** — `AsciiCaseInsensitive` option

## Installation

```bash
go get github.com/kiry163/ahocorasick
```

## Quick Start

```go
package main

import (
    "fmt"

    ac "github.com/kiry163/aho-corasick"
)

func main() {
    // Build a matcher
    builder := ac.NewAhoCorasickBuilder(ac.Opts{
        MatchKind: ac.LeftMostLongestMatch,
        DFA:       true,
    })
    matcher := builder.Build([]string{"hello", "world", "hello world"})

    // Find all matches
    matches := matcher.FindAll("hello world!")
    for _, m := range matches {
        fmt.Printf("Pattern %d matched at [%d..%d]\n", m.Pattern(), m.Start(), m.End())
    }

    // Replace
    replacer := ac.NewReplacer(matcher)
    result := replacer.ReplaceAll("hello world!", []string{"HELLO", "WORLD", "HELLO WORLD"})
    fmt.Println(result) // "HELLO WORLD!"
}
```

## API Overview

### AhoCorasick

| Method | Description |
|--------|-------------|
| `FindAll(haystack string)` | Returns all non-overlapping matches |
| `Iter(haystack string)` | Returns a lazy iterator over matches |
| `IterByte(haystack []byte)` | Byte-slice variant of `Iter` |
| `IterOverlapping(haystack string)` | Iterator that includes overlapping matches |
| `IterOverlappingByte(haystack []byte)` | Byte-slice variant of `IterOverlapping` |
| `PatternCount()` | Returns the number of patterns |

### AhoCorasickBuilder

| Option | Type | Description |
|--------|------|-------------|
| `MatchKind` | `matchKind` | `StandardMatch`, `LeftMostFirstMatch`, or `LeftMostLongestMatch` |
| `DFA` | `bool` | Use the DFA backend (faster, more memory) |
| `AsciiCaseInsensitive` | `bool` | Case-insensitive matching (ASCII only) |
| `MatchOnlyWholeWords` | `bool` | Filter matches to whole words only |

### Replacer

| Method | Description |
|--------|-------------|
| `ReplaceAll(haystack, replacements)` | Replace each pattern with the corresponding replacement string |
| `ReplaceAllFunc(haystack, func)` | Replace each match using a callback function |
| `ReplaceAllWith(haystack, replacement)` | Replace all matches with a single replacement string |

## License

MIT
