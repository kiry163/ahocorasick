// Package ahocorasick implements multi-pattern string matching with an
// Aho-Corasick automaton.
//
// Match locations are half-open ranges. Match.Start and Match.End report rune
// offsets, while Match.ByteStart and Match.ByteEnd report byte offsets suitable
// for slicing the original string. Patterns must be valid UTF-8; invalid bytes
// in haystacks follow utf8.RuneCountInString behavior.
package ahocorasick
