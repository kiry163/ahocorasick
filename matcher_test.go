package ahocorasick_test

import (
	"reflect"
	"strings"
	"sync"
	"testing"
	"unicode/utf8"

	ac "github.com/kiry163/ahocorasick"
)

func TestCompileValidation(t *testing.T) {
	t.Parallel()

	if _, err := ac.Compile(nil, ac.Options{}); err == nil {
		t.Fatal("Compile(nil) succeeded")
	}
	if _, err := ac.Compile([]string{"ok", ""}, ac.Options{}); err == nil {
		t.Fatal("Compile accepted an empty pattern")
	}
	if _, err := ac.Compile([]string{string([]byte{0xff})}, ac.Options{}); err == nil {
		t.Fatal("Compile accepted invalid UTF-8")
	}
}

func TestFindAllCoordinates(t *testing.T) {
	t.Parallel()

	matcher := mustCompile(t, []string{"好", "Go", "世界"}, ac.Options{})
	matches := matcher.FindAll("你好吗 Go 世界")
	want := []matchView{
		{pattern: 0, start: 1, end: 2, byteStart: 3, byteEnd: 6},
		{pattern: 1, start: 4, end: 6, byteStart: 10, byteEnd: 12},
		{pattern: 2, start: 7, end: 9, byteStart: 13, byteEnd: 19},
	}
	assertMatches(t, matches, want)

	first, ok := matcher.Find("你好吗 Go 世界")
	if !ok {
		t.Fatal("Find returned no match")
	}
	assertMatches(t, []ac.Match{first}, want[:1])
}

func TestInvalidUTF8HaystackCoordinates(t *testing.T) {
	t.Parallel()

	haystack := string([]byte{'a', 0xff, 'b'})
	matcher := mustCompile(t, []string{"b"}, ac.Options{})
	assertMatches(t, matcher.FindAll(haystack), []matchView{
		{pattern: 0, start: 2, end: 3, byteStart: 2, byteEnd: 3},
	})
}

func TestLeftmostLongest(t *testing.T) {
	t.Parallel()

	matcher := mustCompile(t, []string{"he", "hers", "her"}, ac.Options{})
	assertMatches(t, matcher.FindAll("hers"), []matchView{
		{pattern: 1, start: 0, end: 4, byteStart: 0, byteEnd: 4},
	})
}

func TestFindAllIsNonOverlapping(t *testing.T) {
	t.Parallel()

	matcher := mustCompile(t, []string{"aa"}, ac.Options{})
	assertMatches(t, matcher.FindAll("aaa"), []matchView{
		{pattern: 0, start: 0, end: 2, byteStart: 0, byteEnd: 2},
	})
}

func TestOverlappingOrder(t *testing.T) {
	t.Parallel()

	matcher := mustCompile(t, []string{"he", "her", "hers", "ers"}, ac.Options{})
	assertMatches(t, matcher.FindAllOverlapping("hers"), []matchView{
		{pattern: 2, start: 0, end: 4, byteStart: 0, byteEnd: 4},
		{pattern: 1, start: 0, end: 3, byteStart: 0, byteEnd: 3},
		{pattern: 0, start: 0, end: 2, byteStart: 0, byteEnd: 2},
		{pattern: 3, start: 1, end: 4, byteStart: 1, byteEnd: 4},
	})
}

func TestUnicodeWholeWords(t *testing.T) {
	t.Parallel()

	matcher := mustCompile(t, []string{"猫", "cat", "foo"}, ac.Options{WholeWords: true})
	matches := matcher.FindAll("黑猫 cat 猫咪 foo_bar foo.")
	if got := matchedText("黑猫 cat 猫咪 foo_bar foo.", matches); !reflect.DeepEqual(got, []string{"cat", "foo"}) {
		t.Fatalf("matched text = %q", got)
	}
}

func TestWholeWordFallsBackToShorterCandidate(t *testing.T) {
	t.Parallel()

	matcher := mustCompile(t, []string{"foo", "foo!"}, ac.Options{WholeWords: true})
	matches := matcher.FindAll("foo!bar")
	if got := matchedText("foo!bar", matches); !reflect.DeepEqual(got, []string{"foo"}) {
		t.Fatalf("matched text = %q", got)
	}
}

func TestASCIIInsensitive(t *testing.T) {
	t.Parallel()

	matcher := mustCompile(t, []string{"Go"}, ac.Options{ASCIIInsensitive: true})
	if got := matchedText("go GO Go", matcher.FindAll("go GO Go")); !reflect.DeepEqual(got, []string{"go", "GO", "Go"}) {
		t.Fatalf("matched text = %q", got)
	}
}

func TestReplace(t *testing.T) {
	t.Parallel()

	matcher := mustCompile(t, []string{"世界", "Go"}, ac.Options{})
	got, err := matcher.ReplaceAll("你好，世界 Go", []string{"地球", "Golang"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "你好，地球 Golang" {
		t.Fatalf("ReplaceAll = %q", got)
	}

	got = matcher.ReplaceAllFunc("世界 Go", func(match ac.Match) string {
		return strings.ToUpper([]string{"earth", "go"}[match.Pattern()])
	})
	if got != "EARTH GO" {
		t.Fatalf("ReplaceAllFunc = %q", got)
	}

	if _, err := matcher.ReplaceAll("世界", []string{"only one"}); err == nil {
		t.Fatal("ReplaceAll accepted the wrong replacement count")
	}
}

func TestMatcherConcurrentUse(t *testing.T) {
	matcher := mustCompile(t, []string{"he", "hers", "世界"}, ac.Options{})

	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if len(matcher.FindAll("hers 世界")) != 2 {
				t.Error("FindAll returned an unexpected result")
			}
			if len(matcher.FindAllOverlapping("hers 世界")) != 3 {
				t.Error("FindAllOverlapping returned an unexpected result")
			}
		}()
	}
	wg.Wait()
}

func FuzzLeftmostLongest(f *testing.F) {
	f.Add("ushers", "he", "hers")
	f.Add("你好世界", "好", "世界")
	f.Add(string([]byte{'a', 0xff, 'b'}), "b", "a")

	f.Fuzz(func(t *testing.T, haystack, first, second string) {
		if first == "" || second == "" || !utf8.ValidString(first) || !utf8.ValidString(second) {
			t.Skip()
		}
		patterns := []string{first, second}
		matcher := mustCompile(t, patterns, ac.Options{})
		got := matcher.FindAll(haystack)
		want := naiveLeftmostLongest(haystack, patterns)
		if !equalByteViews(byteViews(got), want) {
			t.Fatalf("got %#v, want %#v", byteViews(got), want)
		}
	})
}

func BenchmarkFindAll(b *testing.B) {
	matcher, err := ac.Compile([]string{"alpha", "beta", "世界", "needle"}, ac.Options{})
	if err != nil {
		b.Fatal(err)
	}
	haystack := strings.Repeat("alpha haystack 世界 without much content ", 100)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = matcher.FindAll(haystack)
	}
}

func BenchmarkCompile(b *testing.B) {
	patterns := make([]string, 100)
	for i := range patterns {
		patterns[i] = strings.Repeat("prefix", i%8) + string(rune('a'+i%26))
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := ac.Compile(patterns, ac.Options{}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReplaceAll(b *testing.B) {
	matcher, err := ac.Compile([]string{"alpha", "世界"}, ac.Options{})
	if err != nil {
		b.Fatal(err)
	}
	haystack := strings.Repeat("alpha and 世界 ", 100)
	replacements := []string{"a", "world"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = matcher.ReplaceAll(haystack, replacements)
	}
}

type matchView struct {
	pattern, start, end, byteStart, byteEnd int
}

func mustCompile(t testing.TB, patterns []string, options ac.Options) *ac.Matcher {
	t.Helper()
	matcher, err := ac.Compile(patterns, options)
	if err != nil {
		t.Fatal(err)
	}
	return matcher
}

func assertMatches(t testing.TB, matches []ac.Match, want []matchView) {
	t.Helper()
	got := make([]matchView, len(matches))
	for i, match := range matches {
		got[i] = matchView{match.Pattern(), match.Start(), match.End(), match.ByteStart(), match.ByteEnd()}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("matches = %#v, want %#v", got, want)
	}
}

func matchedText(haystack string, matches []ac.Match) []string {
	result := make([]string, len(matches))
	for i, match := range matches {
		result[i] = haystack[match.ByteStart():match.ByteEnd()]
	}
	return result
}

type byteView struct {
	pattern, start, end int
}

func byteViews(matches []ac.Match) []byteView {
	views := make([]byteView, len(matches))
	for i, match := range matches {
		views[i] = byteView{match.Pattern(), match.ByteStart(), match.ByteEnd()}
	}
	return views
}

func equalByteViews(left, right []byteView) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func naiveLeftmostLongest(haystack string, patterns []string) []byteView {
	var matches []byteView
	for at := 0; at <= len(haystack); {
		best := byteView{pattern: -1, start: len(haystack) + 1}
		for patternID, pattern := range patterns {
			relative := strings.Index(haystack[at:], pattern)
			if relative < 0 {
				continue
			}
			start := at + relative
			end := start + len(pattern)
			if start < best.start || start == best.start && (end > best.end || end == best.end && patternID < best.pattern) {
				best = byteView{patternID, start, end}
			}
		}
		if best.pattern < 0 {
			break
		}
		matches = append(matches, best)
		at = best.end
	}
	return matches
}
