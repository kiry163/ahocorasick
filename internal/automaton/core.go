package automaton

type machine interface {
	MatchKind() *matchKind
	StartState() stateID
	MaxPatternLen() int
	PatternCount() int
	OverlappingFindAt(*prefilterState, []byte, int, *stateID, *int) *rawMatch
	FindAtNoState(*prefilterState, []byte, int) *rawMatch
}

// Kind selects the match semantics used by an Automaton. It is internal to
// the module; the public package exposes behavior-oriented methods instead.
type Kind uint8

const (
	Standard Kind = iota
	LeftmostLongest
)

// Automaton is the narrow facade used by the public package.
type Automaton struct {
	machine machine
}

// Match is a byte-oriented match produced by the automaton.
type Match struct {
	Pattern    int
	RuneLength int
	ByteStart  int
	ByteEnd    int
}

// Compile builds an Aho-Corasick automaton and chooses its representation
// within a bounded DFA memory budget.
func Compile(patterns [][]byte, kind Kind, asciiInsensitive bool) *Automaton {
	return compile(patterns, kind, asciiInsensitive, backendAuto)
}

type backend uint8

const (
	backendAuto backend = iota
	backendNFA
	backendDFA
)

const maxDFATransitions = 1 << 20

func compile(patterns [][]byte, kind Kind, asciiInsensitive bool, selected backend) *Automaton {
	internalKind := leftmostLongestMatch
	if kind == Standard {
		internalKind = standardMatch
	}
	nfa := newNFABuilder(internalKind, asciiInsensitive).build(patterns)
	useDFA := selected == backendDFA
	if selected == backendAuto {
		alphabetLen := nfa.byteClasses.alphabetLen()
		useDFA = alphabetLen > 0 && len(nfa.states) <= maxDFATransitions/alphabetLen
	}
	if useDFA {
		return &Automaton{machine: newDFABuilder().build(nfa)}
	}
	return &Automaton{machine: nfa}
}

// PatternCount returns the number of compiled patterns.
func (a *Automaton) PatternCount() int {
	return a.machine.PatternCount()
}

type matchKind uint8

const (
	standardMatch matchKind = iota
	leftmostFirstMatch
	leftmostLongestMatch
)

func (m matchKind) isStandard() bool {
	return m == standardMatch
}

func (m matchKind) isLeftmost() bool {
	return m == leftmostFirstMatch || m == leftmostLongestMatch
}

func (m matchKind) isLeftmostFirst() bool {
	return m == leftmostFirstMatch
}

type rawMatch struct {
	pattern int
	length  int
	runeLen int
	end     int
}

func (m rawMatch) start() int {
	return m.end - m.length
}

type stateID uint

const (
	failedStateID stateID = 0
	deadStateID   stateID = 1
)
