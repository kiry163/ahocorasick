package automaton

type automaton interface {
	Repr() *iRepr
	MatchKind() *matchKind
	Anchored() bool
	Prefilter() prefilter
	StartState() stateID
	IsValid(stateID) bool
	IsMatchState(stateID) bool
	IsMatchOrDeadState(stateID) bool
	GetMatch(stateID, int, int) *rawMatch
	MatchCount(stateID) int
	NextStateNoFail(stateID, byte) stateID
}

func isMatchOrDeadState(a automaton, si stateID) bool {
	return si == deadStateID || a.IsMatchState(si)
}

func standardFindAt(a automaton, prestate *prefilterState, haystack []byte, at int, sID *stateID) *rawMatch {
	pre := a.Prefilter()
	return standardFindAtImp(a, prestate, pre, haystack, at, sID)
}

func standardFindAtImp(a automaton, prestate *prefilterState, prefilter prefilter, haystack []byte, at int, sID *stateID) *rawMatch {
	for at < len(haystack) {
		if prefilter != nil {
			startState := a.StartState()
			if prestate.IsEffective(at) && *sID == startState {
				c := nextPrefilter(prestate, prefilter, haystack, at)
				if c == noneCandidate {
					return nil
				} else {
					at = c
				}
			}
		}
		*sID = a.NextStateNoFail(*sID, haystack[at])
		at += 1

		if a.IsMatchOrDeadState(*sID) {
			if *sID == deadStateID {
				return nil
			} else {
				return a.GetMatch(*sID, 0, at)
			}
		}
	}
	return nil
}

// These stateful variants remain private adapters for the concrete
// representations. Public iteration uses the no-state paths below.
func leftmostFindAt(a automaton, prestate *prefilterState, haystack []byte, at int, sID *stateID) *rawMatch {
	return leftmostFindAtImp(a, prestate, a.Prefilter(), haystack, at, sID)
}

func leftmostFindAtImp(a automaton, prestate *prefilterState, prefilter prefilter, haystack []byte, at int, sID *stateID) *rawMatch {
	if a.Anchored() && at > 0 && *sID == a.StartState() {
		return nil
	}
	lastMatch := a.GetMatch(*sID, 0, at)
	for at < len(haystack) {
		if prefilter != nil {
			startState := a.StartState()
			if prestate.IsEffective(at) && *sID == startState {
				candidate := nextPrefilter(prestate, prefilter, haystack, at)
				if candidate == noneCandidate {
					return nil
				}
				at = candidate
			}
		}
		*sID = a.NextStateNoFail(*sID, haystack[at])
		at++
		if a.IsMatchOrDeadState(*sID) {
			if *sID == deadStateID {
				return lastMatch
			}
			lastMatch = a.GetMatch(*sID, 0, at)
		}
	}
	return lastMatch
}

func leftmostFindAtNoState(a automaton, prestate *prefilterState, haystack []byte, at int) *rawMatch {
	return leftmostFindAtNoStateImp(a, prestate, a.Prefilter(), haystack, at)
}

func leftmostFindAtNoStateImp(a automaton, prestate *prefilterState, prefilter prefilter, haystack []byte, at int) *rawMatch {
	if a.Anchored() && at > 0 {
		return nil
	}
	if prefilter != nil && !prefilter.ReportsFalsePositives() {
		c := prefilter.NextCandidate(prestate, haystack, at)
		if c == noneCandidate {
			return nil
		}
	}

	stateID := a.StartState()
	lastMatch := a.GetMatch(stateID, 0, at)

	for at < len(haystack) {
		if prefilter != nil && prestate.IsEffective(at) && stateID == a.StartState() {
			c := prefilter.NextCandidate(prestate, haystack, at)
			if c == noneCandidate {
				return nil
			} else {
				at = c
			}
		}

		stateID = a.NextStateNoFail(stateID, haystack[at])
		at += 1

		if a.IsMatchOrDeadState(stateID) {
			if stateID == deadStateID {
				return lastMatch
			}
			lastMatch = a.GetMatch(stateID, 0, at)
		}
	}

	return lastMatch
}

func overlappingFindAt(a automaton, prestate *prefilterState, haystack []byte, at int, id *stateID, matchIndex *int) *rawMatch {
	if a.Anchored() && at > 0 && *id == a.StartState() {
		return nil
	}

	matchCount := a.MatchCount(*id)

	if *matchIndex < matchCount {
		result := a.GetMatch(*id, *matchIndex, at)
		*matchIndex += 1
		return result
	}

	*matchIndex = 0
	match := standardFindAt(a, prestate, haystack, at, id)

	if match == nil {
		return nil
	}

	*matchIndex = 1
	return match
}

func earliestFindAt(a automaton, prestate *prefilterState, haystack []byte, at int, id *stateID) *rawMatch {
	if *id == a.StartState() {
		if a.Anchored() && at > 0 {
			return nil
		}
		match := a.GetMatch(*id, 0, at)
		if match != nil {
			return match
		}
	}
	return standardFindAt(a, prestate, haystack, at, id)
}

func findAt(a automaton, prestate *prefilterState, haystack []byte, at int, id *stateID) *rawMatch {
	kind := a.MatchKind()
	if kind == nil {
		return nil
	}
	if *kind == standardMatch {
		return earliestFindAt(a, prestate, haystack, at, id)
	}
	return leftmostFindAt(a, prestate, haystack, at, id)
}

func findAtNoState(a automaton, prestate *prefilterState, haystack []byte, at int) *rawMatch {
	kind := a.MatchKind()
	if kind == nil {
		return nil
	}
	switch *kind {
	case standardMatch:
		state := a.StartState()
		return earliestFindAt(a, prestate, haystack, at, &state)
	case leftmostFirstMatch, leftmostLongestMatch:
		return leftmostFindAtNoState(a, prestate, haystack, at)
	}
	return nil
}
