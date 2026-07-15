package ahocorasick

import (
	"errors"
	"strings"
)

// ReplaceAll replaces each match with the replacement at its pattern index.
func (m *Matcher) ReplaceAll(haystack string, replacements []string) (string, error) {
	if len(replacements) != m.patternCount {
		return "", errors.New("ahocorasick: replacement count must equal pattern count")
	}
	return m.replaceAll(haystack, func(match Match) string {
		return replacements[match.Pattern()]
	}), nil
}

// ReplaceAllFunc replaces each match using replace.
func (m *Matcher) ReplaceAllFunc(haystack string, replace func(Match) string) string {
	return m.replaceAll(haystack, replace)
}

func (m *Matcher) replaceAll(haystack string, replace func(Match) string) string {
	var result strings.Builder
	result.Grow(len(haystack))
	last := 0
	found := false
	m.forEachMatch(haystack, func(match Match) bool {
		found = true
		result.WriteString(haystack[last:match.ByteStart()])
		result.WriteString(replace(match))
		last = match.ByteEnd()
		return true
	})
	if !found {
		return haystack
	}
	result.WriteString(haystack[last:])
	return result.String()
}
