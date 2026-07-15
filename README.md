# Aho-Corasick for Go

`ahocorasick` is a compact multi-pattern string matcher based on the
[Aho-Corasick algorithm](https://en.wikipedia.org/wiki/Aho%E2%80%93Corasick_algorithm).
It uses leftmost-longest matching by default and keeps NFA/DFA representation
details internal.

## Install

```bash
go get github.com/kiry163/ahocorasick
```

## Usage

```go
package main

import (
	"fmt"
	"log"

	"github.com/kiry163/ahocorasick"
)

func main() {
	matcher, err := ahocorasick.Compile(
		[]string{"世界", "Go"},
		ahocorasick.Options{ASCIIInsensitive: true},
	)
	if err != nil {
		log.Fatal(err)
	}

	haystack := "你好，世界! go"
	for _, match := range matcher.FindAll(haystack) {
		fmt.Printf(
			"pattern %d: runes [%d,%d), bytes [%d,%d), text %q\n",
			match.Pattern(),
			match.Start(),
			match.End(),
			match.ByteStart(),
			match.ByteEnd(),
			haystack[match.ByteStart():match.ByteEnd()],
		)
	}
}
```

## Matching Semantics

- `Find` and `FindAll` use non-overlapping leftmost-longest matching.
- `FindAllOverlapping` returns all matches ordered by start position, then by
  longest match and pattern index.
- All ranges are half-open: `[start, end)`.
- `Start` and `End` are rune offsets.
- `ByteStart` and `ByteEnd` are byte offsets suitable for slicing the original
  string.
- Patterns must be valid UTF-8. Invalid UTF-8 in a haystack follows Go's
  `utf8.RuneCountInString` behavior.
- Empty pattern sets and empty patterns are rejected by `Compile`.

`Options.WholeWords` uses Unicode letters, digits, marks, and connector
punctuation (including `_`) as word characters. `Options.ASCIIInsensitive`
folds ASCII letters only.

## Replacement

```go
result, err := matcher.ReplaceAll(
	"你好，世界! go",
	[]string{"地球", "Golang"},
)
```

The replacement slice must contain one entry for every compiled pattern.
`ReplaceAllFunc` can compute replacements from each `Match`.

`Matcher` is immutable after compilation and safe for concurrent use.

## License

MIT
