package automaton

import (
	"reflect"
	"strings"
	"testing"
)

func TestNFAAndDFAAgree(t *testing.T) {
	t.Parallel()

	patterns := [][]byte{[]byte("he"), []byte("hers"), []byte("世界")}
	haystack := []byte("ushers 世界")

	for _, kind := range []Kind{Standard, LeftmostLongest} {
		var reference []Match
		for _, selected := range []backend{backendNFA, backendDFA} {
			automaton := compile(patterns, kind, false, selected)
			matches := collect(automaton.Iterate(haystack, kind == Standard))
			if reference == nil {
				reference = matches
				continue
			}
			if !reflect.DeepEqual(matches, reference) {
				t.Fatalf("kind %v: DFA matches %#v, NFA matches %#v", kind, matches, reference)
			}
		}
	}
}

func BenchmarkBackends(b *testing.B) {
	patterns := [][]byte{[]byte("alpha"), []byte("beta"), []byte("世界"), []byte("needle")}
	haystack := []byte(strings.Repeat("alpha haystack 世界 without much content ", 100))
	for _, benchmark := range []struct {
		name    string
		backend backend
	}{
		{name: "NFA", backend: backendNFA},
		{name: "DFA", backend: backendDFA},
	} {
		b.Run(benchmark.name, func(b *testing.B) {
			automaton := compile(patterns, LeftmostLongest, false, benchmark.backend)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				iterator := automaton.Iterate(haystack, false)
				for _, ok := iterator.Next(); ok; _, ok = iterator.Next() {
				}
			}
		})
	}
}

func TestStandardSearchUsesPrefilter(t *testing.T) {
	t.Parallel()

	patterns := [][]byte{[]byte("needle")}
	haystack := append(make([]byte, 256), []byte("needle")...)
	for i := range haystack[:256] {
		haystack[i] = 'x'
	}

	for _, selected := range []backend{backendNFA, backendDFA} {
		automaton := compile(patterns, Standard, false, selected)
		iterator := automaton.Iterate(haystack, true)
		match, ok := iterator.Next()
		if !ok || match.ByteStart != 256 {
			t.Fatalf("backend %v returned %#v, %v", selected, match, ok)
		}
		if iterator.prefilter.skips == 0 {
			t.Fatalf("backend %v did not use its prefilter", selected)
		}
	}
}

func collect(iterator *Iterator) []Match {
	var matches []Match
	for match, ok := iterator.Next(); ok; match, ok = iterator.Next() {
		matches = append(matches, match)
	}
	return matches
}
