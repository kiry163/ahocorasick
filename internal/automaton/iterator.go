package automaton

// Iterator walks matches in a single haystack.
type Iterator struct {
	machine     machine
	prefilter   prefilterState
	haystack    []byte
	position    int
	state       stateID
	matchIndex  int
	overlapping bool
}

// Iterate creates an iterator. The automaton must use Standard semantics when
// overlapping is true.
func (a *Automaton) Iterate(haystack []byte, overlapping bool) *Iterator {
	return &Iterator{
		machine: a.machine,
		prefilter: prefilterState{
			maxMatchLen: a.machine.MaxPatternLen(),
		},
		haystack:    haystack,
		state:       a.machine.StartState(),
		overlapping: overlapping,
	}
}

// Next returns the next match.
func (i *Iterator) Next() (Match, bool) {
	if i.position > len(i.haystack) {
		return Match{}, false
	}

	var result *rawMatch
	if i.overlapping {
		result = i.machine.OverlappingFindAt(
			&i.prefilter,
			i.haystack,
			i.position,
			&i.state,
			&i.matchIndex,
		)
		if result != nil {
			i.position = result.end
		}
	} else {
		result = i.machine.FindAtNoState(&i.prefilter, i.haystack, i.position)
		if result != nil {
			i.position = result.end
		}
	}
	if result == nil {
		return Match{}, false
	}
	return Match{
		Pattern:    result.pattern,
		RuneLength: result.runeLen,
		ByteStart:  result.start(),
		ByteEnd:    result.end,
	}, true
}
