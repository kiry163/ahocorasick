package ahocorasick

import (
	"unicode/utf8"

	"github.com/kiry163/ahocorasick/internal/automaton"
)

type matchIterator struct {
	inner        *automaton.Iterator
	haystack     []byte
	runeBytePos  int
	runePosition int
}

func newMatchIterator(machine *automaton.Automaton, haystack []byte, overlapping bool) matchIterator {
	return matchIterator{
		inner:    machine.Iterate(haystack, overlapping),
		haystack: haystack,
	}
}

func (i *matchIterator) next() (Match, bool) {
	for {
		raw, ok := i.inner.Next()
		if !ok {
			return Match{}, false
		}
		if !isRuneBoundary(i.haystack, raw.ByteStart) || !isRuneBoundary(i.haystack, raw.ByteEnd) {
			continue
		}
		if raw.ByteEnd > i.runeBytePos {
			i.runePosition += utf8.RuneCount(i.haystack[i.runeBytePos:raw.ByteEnd])
			i.runeBytePos = raw.ByteEnd
		}
		return Match{
			pattern:   raw.Pattern,
			start:     i.runePosition - raw.RuneLength,
			end:       i.runePosition,
			byteStart: raw.ByteStart,
			byteEnd:   raw.ByteEnd,
		}, true
	}
}

func isRuneBoundary(haystack []byte, offset int) bool {
	if offset == 0 || offset == len(haystack) {
		return true
	}
	if utf8.Valid(haystack) {
		return utf8.RuneStart(haystack[offset])
	}

	for at := 0; at < offset; {
		_, size := utf8.DecodeRune(haystack[at:])
		at += size
		if at == offset {
			return true
		}
		if at > offset {
			return false
		}
	}
	return false
}
