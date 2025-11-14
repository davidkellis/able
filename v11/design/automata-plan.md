# Able v10 Automata Plan

## Goals
- Provide reusable NFA/DFA builders and executors implemented entirely in Able, specialised for UTF-8 `String` input.
- Serve as the foundation for the future regex engine, while remaining useful to other text algorithms (e.g., token filters, deterministic scanners).
- Keep the API ergonomic: programmatic construction, epsilon transitions, subset determinisation, and helper matchers.

## Constraints
- No parser infrastructure yet; automata must be buildable via explicit builder calls.
- Execution operates over Able `String` instances using the new `chars()` iterator. All indexing is byte-accurate to keep substring extraction consistent.
- Initial scope limits character classes to literal chars and the wildcard `.`; richer predicates (Unicode categories) will be layered on later.
- Determinisation must be predictable but can favour clarity over ultimate performance for now.

## Surface API
### Types
- `struct NFABuilder` — mutable builder supporting `add_state`, `set_start`, `add_accept`, `add_transition`.
- `struct NFA` — immutable automaton with transitions, start state, and accepting states.
- `struct DFA` — deterministic automaton produced via subset construction.
- `enum NFASymbol = NFAChar char | NFAEpsilon` (additional predicates will be layered on later).
- `struct AutomataExpr` — DSL node representing composable automata expressions (literal, char sets, concat/seq, append, union, kleene/plus/optional) with helpers to produce NFAs/DFAs. Helper module exposes `literal`, `literal_string`, `any`, `range`, `dot`, `seq`, `union`, `append`, `kleene`, `plus`, `optional`, mirroring the `kleene` DSL ergonomics.
- `struct NFATransition { from: i32, to: i32, symbol: NFASymbol }`
- `struct DFATransition { from: i32, to: i32, symbol: char }`

### Key helpers
- `NFABuilder::new()`, `add_state() -> i32`, `set_start(state)`, `add_accept(state)`, `add_transition(from, to, symbol)`, `build() -> Result NFA`.
- `NFA::matches(text: String) -> bool`.
- `fn nfa_matches(nfa: NFA, text: String) -> bool`.
- `fn nfa_to_dfa(nfa: NFA) -> DFA`.
- `DFA::matches(text: String) -> bool`.
- `fn dfa_matches(dfa: DFA, text: String) -> bool`.
- Utility functions: epsilon-closure, state-set helpers (duplicate elimination, sorting), substring extraction.

### Usage sketch
```able
builder := NFABuilder::new()
s0 := builder.add_state()
s1 := builder.add_state()
builder.set_start(s0)
builder.add_accept(s1)
builder.add_transition(s0, s1, NFAChar 'a')
nfa := builder.build().unwrap()
expect(nfa.matches(String::from_builtin("a").unwrap())).to(be_truthy())
```

## Implementation Notes
- Represent state sets as sorted `Array i32` values to enable equality comparisons when determinising.
- Epsilon closure uses a stack + visited list; with current state counts the O(n²) approach is acceptable.
- DFA determinisation maintains a descriptor table mapping subsets to DFA state ids.
- Matching over `String` reuses the UTF-8 decoder already available in `text/string.able`.
- Replacement/split helpers will later depend on the DFA once regex parsing exists.

## Roadmap
1. Land `stdlib/src/text/automata.able` with NFABuilder, NFA, DFA, and execution helpers (done).
2. Ship `stdlib/src/text/automata_dsl.able` offering the compositional DSL + helpers to convert to NFAs/DFAs (done).
3. Add basic unit tests covering literal matches, `*`/`+` quantifier equivalents, determinisation correctness, and DSL ergonomics (done).
4. Extend symbol support (character classes, predicates) once the regex parser is online.
5. Optimise state-set handling (maps, caching) if performance becomes a concern after regex integration.
