package ahocorasick

import (
	"errors"
	"sort"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/kiry163/ahocorasick/internal/automaton"
)

// Options controls matching semantics. The automaton representation is chosen
// internally and is not part of the public API.
type Options struct {
	ASCIIInsensitive bool
	WholeWords       bool
}

// Matcher searches for a fixed set of patterns. A Matcher is safe for
// concurrent use.
type Matcher struct {
	main         *automaton.Automaton
	patternCount int
	options      Options
	overlap      *lazyAutomaton
}

type lazyAutomaton struct {
	once             sync.Once
	patterns         [][]byte
	asciiInsensitive bool
	machine          *automaton.Automaton
}

// Match identifies a pattern and its half-open location in the haystack.
// Start and End are rune offsets; ByteStart and ByteEnd are byte offsets.
type Match struct {
	pattern            int
	start, end         int
	byteStart, byteEnd int
}

// Pattern returns the index of the matching pattern passed to Compile.
func (m Match) Pattern() int { return m.pattern }

// Start returns the inclusive rune offset of the match.
func (m Match) Start() int { return m.start }

// End returns the exclusive rune offset of the match.
func (m Match) End() int { return m.end }

// ByteStart returns the inclusive byte offset of the match.
func (m Match) ByteStart() int { return m.byteStart }

// ByteEnd returns the exclusive byte offset of the match.
func (m Match) ByteEnd() int { return m.byteEnd }

// Compile builds a matcher for patterns. At least one non-empty pattern is
// required.
func Compile(patterns []string, options Options) (*Matcher, error) {
	if len(patterns) == 0 {
		return nil, errors.New("ahocorasick: at least one pattern is required")
	}

	bytePatterns := make([][]byte, len(patterns))
	for i, pattern := range patterns {
		if len(pattern) == 0 {
			return nil, errors.New("ahocorasick: empty patterns are not supported")
		}
		if !utf8.ValidString(pattern) {
			return nil, errors.New("ahocorasick: patterns must be valid UTF-8")
		}
		bytePatterns[i] = []byte(pattern)
	}

	return &Matcher{
		main:         automaton.Compile(bytePatterns, automaton.LeftmostLongest, options.ASCIIInsensitive),
		patternCount: len(patterns),
		options:      options,
		overlap: &lazyAutomaton{
			patterns:         bytePatterns,
			asciiInsensitive: options.ASCIIInsensitive,
		},
	}, nil
}

// Find returns the first leftmost-longest match.
func (m *Matcher) Find(haystack string) (Match, bool) {
	if !m.options.WholeWords {
		iter := newMatchIterator(m.main, []byte(haystack), false)
		match, ok := iter.next()
		if !ok {
			return Match{}, false
		}
		return match, true
	}

	matches := m.findWholeWordMatches(haystack, false)
	if len(matches) == 0 {
		return Match{}, false
	}
	return matches[0], true
}

// FindAll returns non-overlapping leftmost-longest matches.
func (m *Matcher) FindAll(haystack string) []Match {
	var matches []Match
	m.forEachMatch(haystack, func(match Match) bool {
		matches = append(matches, match)
		return true
	})
	return matches
}

// FindAllOverlapping returns every match, including overlapping matches.
func (m *Matcher) FindAllOverlapping(haystack string) []Match {
	return m.findWholeWordMatches(haystack, true)
}

func (m *Matcher) overlappingMachine() *automaton.Automaton {
	m.overlap.once.Do(func() {
		m.overlap.machine = automaton.Compile(
			m.overlap.patterns,
			automaton.Standard,
			m.overlap.asciiInsensitive,
		)
	})
	return m.overlap.machine
}

func (m *Matcher) forEachMatch(haystack string, yield func(Match) bool) {
	if m.options.WholeWords {
		for _, match := range m.findWholeWordMatches(haystack, false) {
			if !yield(match) {
				return
			}
		}
		return
	}

	iter := newMatchIterator(m.main, []byte(haystack), false)
	for match, ok := iter.next(); ok; match, ok = iter.next() {
		if !yield(match) {
			return
		}
	}
}

func (m *Matcher) findWholeWordMatches(haystack string, keepOverlapping bool) []Match {
	bytes := []byte(haystack)
	iter := newMatchIterator(m.overlappingMachine(), bytes, true)
	var matches []Match
	for match, ok := iter.next(); ok; match, ok = iter.next() {
		if m.options.WholeWords && !hasWordBoundaries(bytes, match.byteStart, match.byteEnd) {
			continue
		}
		matches = append(matches, match)
	}

	if keepOverlapping {
		sortMatches(matches)
		return matches
	}
	if len(matches) < 2 {
		return matches
	}
	return leftmostLongest(matches)
}

func leftmostLongest(matches []Match) []Match {
	sortMatches(matches)

	selected := matches[:0]
	nextStart := 0
	for _, match := range matches {
		if match.byteStart < nextStart {
			continue
		}
		selected = append(selected, match)
		nextStart = match.byteEnd
	}
	return selected
}

func sortMatches(matches []Match) {
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].byteStart != matches[j].byteStart {
			return matches[i].byteStart < matches[j].byteStart
		}
		if matches[i].byteEnd != matches[j].byteEnd {
			return matches[i].byteEnd > matches[j].byteEnd
		}
		return matches[i].pattern < matches[j].pattern
	})
}

func hasWordBoundaries(haystack []byte, start, end int) bool {
	if start > 0 {
		r, _ := utf8.DecodeLastRune(haystack[:start])
		if isWordRune(r) {
			return false
		}
	}
	if end < len(haystack) {
		r, _ := utf8.DecodeRune(haystack[end:])
		if isWordRune(r) {
			return false
		}
	}
	return true
}

func isWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsMark(r) || unicode.Is(unicode.Pc, r)
}
