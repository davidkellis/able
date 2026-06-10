# Able v12 Regex: Implementation Record and Historical Design Notes

**Status:** implemented and reconciled 2026-07-14.

The active regex contract is complete. Its sources of truth are:

1. `spec/full_spec_v12.md` §14.2 for public API and semantics.
2. The canonical external `able-stdlib` modules under `src/text/` (`regex`,
   `regex_builder`, `regex_set`, `regex_scanner`, and shared automata files)
   for implementation.
3. The shared `exec/14_02_regex_core_match_streaming` through
   `exec/14_25_regex_builder_program` fixtures for behavioral evidence in the
   Go tree-walker, bytecode VM, and compiler.

`spec/TODO_v12.md` tracks no remaining regex engine or API gap. The module is
ordinary Able code using shared automata primitives; it has no host-regex
execution path or regex-specific compiler/VM rule. The dated milestones and
proposal material below are retained as history. In particular, statements
that an option is “unimplemented,” name a “next code-bearing slice,” prescribe
Go/TypeScript parity, use the retired in-tree `stdlib/` paths, or list rollout
and follow-up work are not active v12 tasks. Any future regex work must begin
with a spec-defined behavior, canonical-source change, shared executable
coverage, and the normal cross-application performance gate.

## Current Status

- `able.spec.match_regex` delegates to `regex_is_match`; it is not a
  string-equality placeholder.
- The canonical module implements the specified deterministic tagged
  Thompson-NFA API: Unicode code-point matching, supported options,
  captures/replacements/splitting, combined-NFA `RegexSet`, incremental
  `RegexScanner`, immutable builders, and program snapshots.
- `unicode=false`, `unicode_case`, and `grapheme_mode` remain explicit
  `RegexUnsupportedFeature` results; they are never silently accepted.
- The existing regex performance record documents already-kept general
  automata improvements. It selects no further regex, compiler, VM, or
  benchmark-specific optimization.

## Historical Goals
- Provide a consistent, spec-backed `Regex` library for Able v12 with identical semantics across Go and TypeScript runtimes.
- Guarantee linear-time matching (no catastrophic backtracking) while supporting Unicode-aware patterns, lookaround, and rich replacement APIs.
- Expose layered functionality: immediate matching helpers, compiled regex handles, multi-pattern search, streaming scanners, and automata inspection.
- Keep the Able stdlib surface idiomatic: ergonomic modules, strong typing, clear error reporting, and integration with the existing testing matchers.
- Supply hooks for future tooling (bytecode export, visualization) without locking us to a specific host implementation.

## Historical Non-goals / Deferred Items
- Backreferences and constructs that require non-regular backtracking remain out of scope for the first release.
- PCRE-incompatible features (e.g., conditional expressions) will only be considered after the baseline library is stable.
- Formal specification text will land after the core API is implemented and exercised across both interpreters.
- No initial attempt to expose JIT compilation or host-specific regex engines; deterministic portable semantics take precedence.

## Historical Constraints & Principles
- Adopt RE2-style guarantees: every API must be linear in input length for fixed options and pattern.
- Default execution counts Unicode scalar values (`char`); grapheme-aware matching is opt-in via `RegexOptions.grapheme_mode`.
- Unicode correctness: pattern escapes and character classes operate on Unicode scalar values. When grapheme mode is enabled the engine segments haystacks using the standard library `Grapheme` iterator.
- Deterministic runtime: compiled regex values are immutable and can be shared across threads/procs safely.
- Error handling uses `Result`/`RegexError` with precise location data; mis-specified patterns never panic.
- Keep the regex engine entirely in Able stdlib with no host regex hooks.

## Implementation Chronology (2026-07-12)

The entries in this section are dated development snapshots. Earlier entries
are intentionally retained to explain later changes, but their interim
limitations and “next” recommendations are superseded by the current contract
above and by the specification.
- The stdlib regex module supports Unicode code-point literal tokens with
  escapes, wildcard, character classes, and quantifiers (`*`, `+`, `?`,
  `{m}`, `{m,}`, `{m,n}`), plus `match`/`find_all` and streaming `scan`.
- Unsupported metacharacters still return `RegexUnsupportedFeature`; `match_regex` continues to fall back to string equality when compilation fails.
- Every option outside that implemented code-point subset is now explicitly rejected. In
  particular, `case_insensitive`, `multiline`, and `dot_matches_newline` may
  not be silently accepted until the Unicode/code-point engine implements
  their semantics.
- The regex exec fixture is active again (`exec/14_02_regex_core_match_streaming`).

### 2026-07-12 Phase-1 option-safety closure

The canonical external stdlib now rejects every unavailable option at compile
time, including the three fields that were previously accepted but ignored:
`case_insensitive`, `multiline`, and `dot_matches_newline`. This is a
correctness boundary, not a claim that those features are implemented.

The partial matcher now retains its internal search offsets as `i32` in
`RegexByteSpan` and constructs the specified public `Span { start: u64, end:
u64 }` only when building a `Match`. This keeps internal byte-array indexing
and public span representation separate while preserving identical tree-walker,
bytecode, and compiled results.

### 2026-07-12 Code-Point Thompson-NFA Core

The literal/escape/quantifier subset now parses Unicode scalar values and
compiles through `able.text.automata_dsl` into the shared `NFA` representation.
The matcher runs an epsilon-closure thread set without recursive backtracking,
retains the earliest start for each NFA state, and records the latest accepting
end for that start. This preserves greedy leftmost matches while keeping public
`Span` values in UTF-8 byte offsets. Shared executable coverage exercises
non-ASCII literals, bounded greedy quantifiers, non-overlapping iteration, and
zero-length progression in the tree walker, bytecode VM, and compiler.

### 2026-07-12 Wildcard and Character-Class NFA Atoms

Wildcard and character classes now compile through the same Unicode NFA path.
Classes support literals, escaped members, Unicode scalar ranges, and `[^...]`
negation without expanding ranges. Default `.` excludes newline; the
unimplemented `dot_matches_newline=true` option continues to reject rather
than silently changing that contract. The shared automata layer now models
scalar wildcards and compact class/range transitions for NFA execution while
preserving the pre-existing DSL `dot()` and range behavior used by DFA callers.

### 2026-07-12 Groups and Alternation AST

The linear token parser is now a recursive-descent expression parser with
concatenation binding tighter than alternation and quantified nested terms. It
lowers `|` and `(?:...)` directly into the existing shared Thompson-NFA DSL,
including nested group repetition. Ordinary `(...)` continues to reject as an
unimplemented capture feature: accepting it before `Match.groups` and
`named_groups` carry the specified values and byte spans would expose incorrect
public behavior. The shared `exec/14_06_regex_nfa_groups_alternation` fixture
covers group/alternation precedence, nested greedy repetition, capture
rejection, and malformed-group rejection in the tree walker, bytecode VM, and
compiled executable.

The next code-bearing slice should add absolute `^` / `$` anchors through
explicit NFA boundary predicates. `multiline=true` must remain an error until
line-boundary semantics are added. Captures, replacement, sets, and other
option flags remain explicitly unsupported until their distinct semantics and
cross-runtime fixtures are complete.

### 2026-07-12 Absolute Anchor Predicates

The shared automata NFA now represents explicit start and end predicates, and
its full-string matcher evaluates them at UTF-8 input boundaries. Regex lowers
`^` and `$` to those predicates and its leftmost search evaluates closure
transitions at each current boundary, so `^` only admits the initial thread and
`$` only admits the terminal one. Default anchors are absolute: `$` has no
before-final-newline exception. Quantifying an anchor is a syntax error, and
the DFA conversion deliberately rejects NFA programs containing boundary
predicates rather than silently changing their semantics.

The shared `exec/14_07_regex_nfa_absolute_anchors` fixture covers full Unicode
byte spans, a negative non-prefix case, grouped anchored expressions, empty
input, invalid anchor repetition, and direct shared-NFA matching across every
execution mode. `multiline=true` remains an explicit error.

The next code-bearing slice should implement `dot_matches_newline=true` by
threading `RegexOptions` into wildcard lowering. It is a contained public
option whose semantics map directly to the existing `NFAAny.include_newline`
field; captures, replacement, sets, and other flags remain unsupported until
their distinct semantics and cross-runtime fixtures are complete.

### 2026-07-12 Dot-Matches-Newline Option

`RegexOptions.dot_matches_newline=true` is now implemented rather than
silently rejected. Option-aware AST lowering maps only `RegexAnyAtom` to the
existing `NFAAny { include_newline: true }` transition; literals and character
classes retain their normal semantics. The default remains `false`, so `.`
continues to exclude U+000A unless the caller explicitly opts in.

The shared `exec/14_08_regex_dot_matches_newline` fixture proves both default
and enabled behavior, including a grouped/anchored expression, in the tree
walker, bytecode VM, and compiled executable. Other unimplemented flags,
including `multiline`, remain compile-time errors.

The next code-bearing slice should implement `anchored=true` by wrapping the
compiled expression in the existing absolute start/end predicates. It reuses
already-tested NFA semantics and completes another public option without a
benchmark-shaped rule; captures, replacement, sets, and other flags remain
unsupported until their distinct semantics and cross-runtime fixtures are
complete.

### 2026-07-12 Anchored Option

`RegexOptions.anchored=true` is now implemented by wrapping the parsed
expression once with the same absolute start/end anchor AST atoms used for
explicit `^` and `$`. The resulting NFA requires a complete-haystack match for
all expression forms, including Unicode literals and zero-length patterns; no
search-path special case or pattern rewrite is exposed to callers.

The shared `exec/14_09_regex_anchored_option` fixture covers substring
rejection, exact and Unicode matches, and both matching and non-matching
zero-length cases in the tree walker, bytecode VM, and compiled executable.

The next code-bearing slice should implement `multiline=true` by extending the
same anchor predicates with line-boundary context. It is the remaining option
that directly composes with completed anchor semantics; captures, replacement,
sets, and case-insensitive Unicode matching remain separate larger designs.

### 2026-07-12 Multiline Option

`RegexOptions.multiline=true` now lowers explicit `^` and `$` to shared
line-start and line-end NFA predicates. A line boundary is U+000A only: `^`
matches at input start or immediately after it, and `$` matches at input end or
immediately before it. The full-string NFA matcher and regex search closure
both receive the same predecessor/successor boundary context. The
`anchored=true` wrapper deliberately continues to use absolute predicates, so
combining the options still requires an entire-haystack match.

The shared `exec/14_10_regex_multiline_option` fixture covers default versus
multiline behavior, Unicode line spans, empty lines, anchored-option
interaction, and direct shared-NFA line predicates in every execution mode.

The next code-bearing slice should implement `case_insensitive=true` only with
a specified Unicode simple-case-folding policy and shared test corpus. It must
not land as an ASCII-only or benchmark-shaped special case; captures,
replacement, sets, and other incomplete flags remain separate designs.

### 2026-07-12 Case-Insensitive Option

`RegexOptions.case_insensitive=true` now uses Unicode simple-case-fold cycles
for literal and character-class transitions. The canonical primitive advances
one scalar through its fold cycle, and shared NFA code tests literal equality
or class/range membership across that bounded cycle. This retains scalar-based
NFA execution and explicitly does not implement full case folding with
multi-scalar expansions: `ß` therefore does not match `SS`.

The shared `exec/14_11_regex_case_insensitive` fixture covers default contrast,
Kelvin-sign and final-sigma folds, ASCII ranges, negated classes, no-expansion
behavior, anchored interaction, and direct shared-NFA folding in every
execution mode. `unicode_case` remains an error because it requires a separate
policy beyond simple folding.

The next code-bearing slice should add numbered capture groups through tagged
NFA transitions. That is now the highest-value missing regular-expression API,
but it needs a dedicated leftmost-greedy tag-selection and public
`Match.groups` design; named captures, replacement, and sets remain later
layers.

### 2026-07-12 Numbered Capture Groups

Ordinary `(...)` groups now lower into the shared automata DSL as tagged NFA
start/end transitions. The regex closure carries a private capture-tag vector
with each NFA thread, clones it only when a tag changes, and keeps the
leftmost-greedy winning thread's tags with the accepted match. The public
result builder then produces one `Group` per ordinary opening parenthesis in
zero-based source order; `(?:...)` has no index, optional nonparticipating
groups have `value=nil` and `span=nil`, and a repeated group exposes its final
successful iteration.

`exec/14_12_regex_capture_groups` exercises repetition, alternation, optional,
nested, Unicode-byte-span, empty, iterator, and scanner captures in the tree
walker, bytecode VM, and compiled executable. The work also uncovered a general bytecode defect:
a named-struct field could retain the VM's mutable raw-return scratch carrier.
Named struct construction now materializes raw scalars at that aggregate
boundary, with a direct VM regression test, so independent capture spans stay
stable rather than relying on regex-specific storage.

Named capture extensions such as `(?P<name>...)` remain explicit errors. The
next code-bearing slice should add them by attaching an optional name to the
existing numbered capture metadata and then filling `Match.named_groups` from
the final tags. This completes the other half of the specified group API
without adding a separate matching engine or a benchmark-specific path.

### 2026-07-12 Named Capture Groups

`(?P<name>...)` now shares the existing numbered group AST and tagged NFA
transitions. Parsing stores optional capture metadata alongside the source-order
ordinal; matching remains entirely name-blind and returns the same final tag
vector. The public result builder assigns `Group.name` and creates
`Match.named_groups` from those finished ordinal groups, so named and numeric
lookup cannot diverge.

Names are deliberately restricted to ASCII identifiers
`[A-Za-z_][A-Za-z0-9_]*`, and a duplicate name is an invalid pattern. This
provides deterministic map semantics while leaving `(?<name>...)`, lookaround,
and all other group extensions as explicit unsupported syntax. The shared
`exec/14_13_regex_named_capture_groups` fixture covers ordinal and named
views, repetition, nesting, alternation, optional captures, `find_all`,
scanner results, and invalid/duplicate names in every Go execution mode.

The next code-bearing slice should implement literal `Regex.replace`, using
the completed `Match` data to expand whole-match, ordinal, and named references
without touching NFA matching. It should define an explicit replacement-token
grammar and reject bad references; callback replacement and regex sets remain
separate follow-ups.

### 2026-07-12 Literal Replacement

Literal `Regex.replace` now tokenizes its replacement once, before searching,
then assembles non-overlapping output from the same NFA matches used by the
other APIs. `$0` expands the whole match, `$N` is one-based ordinal capture
notation, and `\k<name>` resolves named captures. `$$` and `\\` produce
literal markers. Invalid, dangling, malformed, or unknown references return a
`RegexCompileFailure` even when the haystack contains no match; an unmatched
optional group expands to empty.

Zero-length matches emit their replacement and then copy exactly one Unicode
scalar before resuming. This prevents an infinite loop while preserving the
input scalar (for example, replacing `a*` in `é` produces a replacement before
and after `é`). `exec/14_14_regex_literal_replace` covers named/ordinal
agreement, optional captures, no-match Unicode preservation, zero-length
progression, escapes, and invalid references in every Go execution mode.

### 2026-07-12 Functional Replacement

`ReplacementFunction` now invokes a synchronous `Match -> String` callback
once for each non-overlapping match, in search order. The callback receives the
same capture, named-capture, and span data built for every other regex API;
raised errors propagate normally. It uses the established Unicode-scalar
zero-length progression, while literal replacement remains on its tokenized
path and does not pay callback overhead.

`exec/14_15_regex_function_replace` proves capture access, ordering,
zero-length Unicode behavior, no-match non-invocation, and callback errors in
every Go execution mode. The compiled recovery was deliberately generic:
logical RHS code now remains inside its short-circuit branch, and recursive
nominal conversion emits `nil` nested structs as `runtime.NilValue`. Strict
compiled boundary audits report zero interpreter fallbacks for both literal and
functional replacement.

### 2026-07-12 Regex Split

`Regex.split` now consumes non-overlapping delimiter matches from the shared
NFA search and emits every segment between them, including empty leading,
trailing, and intermediate segments. A numeric `limit` is a maximum delimiter
match count (`0` leaves the haystack intact); `nil` is unlimited. Zero-width
matches advance only the next search position by one Unicode scalar, retaining
that scalar in the next segment and avoiding an infinite scan.

`exec/14_16_regex_split` covers ordinary delimiters, match limits, zero-width
Unicode/anchor delimiters, no-match, empty input, and a whole-input match in
tree-walker, bytecode, and compiled execution. Its compiled boundary audit
records no interpreter fallbacks.

`RegexSet` now compiles its patterns to one combined NFA. A fresh shared start
state has epsilon transitions to each offset source NFA, and a state-indexed
accept map preserves the matching source-pattern identity. `matches` makes one
Unicode-boundary scan, records each matched source index once, and returns
those indices in source order. `exec/14_17_regex_set` exercises captures,
overlap, zero-width patterns, empty sets, options, and compile failures through
tree-walker, bytecode, and compiled execution with a strict no-fallback audit.

`RegexScanner` is now genuinely incremental. It appends only the new chunk's
bytes and code points, retains active NFA threads and capture tags, and tracks
one pending leftmost-longest candidate. A candidate is emitted as soon as no
thread can extend it; `flush()` finalizes EOF-dependent candidates and is
idempotent. This gives `next()` a poll-style contract before EOF, avoids
re-scanning previous chunks, preserves Unicode/anchor/capture semantics, and
rejects `feed` after flush. `exec/14_18` through `exec/14_21` isolate greedy
state, absolute EOF anchors, Unicode zero-width progression, and multiline
chunk-boundary behavior; capture fixtures `exec/14_12` and `exec/14_13` cover
the same flushed scanner path.

`RegexSet.iter` now walks the existing combined NFA and returns
`RegexSetMatch` records. It selects a leftmost-longest non-overlapping span;
every pattern accepting that exact span is emitted once in source order. The
next span begins after the selected match, or one Unicode scalar after a
zero-width match. `exec/14_22` proves source-order ties, longest selection, and
overlap suppression; `exec/14_23` covers Unicode, multiline anchors, and empty
sets in tree-walker, bytecode, compiled execution, and strict no-fallback mode.

`RegexScanner` now compacts its consumed buffer after feeds and finalization.
It tracks a global byte base and prior scalar for line-boundary context, then
keeps only data referenced by an active thread, pending candidate, or capture
tag. Ready matches own their strings/groups, so consumed results release their
queue storage too. `exec/14_24` verifies absolute literal, Unicode, and capture
spans after repeated compaction across every execution mode.

### 2026-07-12 Regex Builder and Program Snapshots

`RegexBuilder` is now the programmatic counterpart to the supported textual
grammar. It builds literal, wildcard, character-class, anchor, concatenation,
alternation, bounded/unbounded repeat, ordinary capture, and named-capture
expressions directly into the existing regex expression AST. Builders are
immutable values: composition shifts the right-hand capture indices, and an
outer capture shifts its nested indices, so source-opening order remains the
same as textual regex construction. Bounds, ASCII capture-name grammar, and
duplicate names reject through the normal regex error path.

`Regex.from_builder`, `RegexBuilder.build`, and `build_with_options` compile
that AST through the same tagged Thompson-NFA and option validation as textual
patterns. Builder-created regexes use the explicit descriptive `"<builder>"`
pattern marker because no source string is reparsed or reconstructed.

`Regex.to_program` now returns a deep, caller-owned `RegexProgram` snapshot.
`nfa_snapshot` and `captured_names` return fresh data, so tooling cannot alter
an executable regex. `dfa_snapshot` is fallible and only accepts literal and
epsilon transitions; it rejects class, wildcard, fold, anchor, and capture
transitions rather than silently losing matching semantics. This is useful
automata export, not a host-regex bridge or a runtime/compiler special case.

`exec/14_25_regex_builder_program` proves composition, capture reindexing,
snapshot isolation, character classes, option propagation, supported and
unsupported DFA export, and bound rejection in the tree walker, bytecode VM,
and compiled executable. The compiled strict boundary audit reports zero
fallbacks.

The verifier-backed cross-language benchmark corpus now contains
`regex-suffix-audit` (public builder composition, captures, file input, and
aggregation) and `regex-set-audit` (public combined-NFA set classification).
The latter has Go/Python/Ruby reference implementations and a bounded default
that preserves ordinary application control flow while keeping bytecode runs
within the normal guard. It confirmed the compiled profile gate: generic
`regex_nfa_move`, `regex_nfa_add_closure`, and their allocations recur across
suffix audit, independent `is_match`, and set classification. The next
candidate was kept: an immutable `RegexNFAIndex`, owned by each compiled
`Regex` or `RegexSet`, groups stable transition positions by source state.
Movement and closure now visit only the active state's edges while retaining
the previous traversal order. Public `NFA` snapshots remain mutable and are
not given this cache, so exported automata semantics do not change. The
cross-application gate improved every bytecode lane, materially improved
suffix/set generated profiles, and left independent matching neutral. The next
allocation-space gate is also complete: the concrete common descendants are
`regex_nfa_add_closure`'s temporary work stack and
`regex_nfa_upsert_thread`'s tagged thread records. The matcher-local reusable
closure stack is now kept: normal matches and RegexSet operations own one
scratch stack, and a scanner owns one for its stream. It drains at each
closure boundary, leaving active thread/capture state isolated. It improved
all three public bytecode workloads while leaving generated applications
neutral or better. The post-change exact allocation gate found a more direct
shared ownership issue than next-state buffers: successful upsert materialized
one active `RegexNFAThread` and closure immediately materialized the same
immutable record again. That candidate is now kept: upsert returns the exact
newly accepted private record and closure pushes it directly. It improved all
three tagged-NFA bytecode applications by 16--19% and reduced allocations by
20--27%, while generated application checks remained correct. The next
investigation is deliberately outside regex. The first three controls
(`zigzag_char_small`, `ascii_lower_small`, and `reverse_complement_small`) did
not execute a material `__able_char_to_codepoint` path: one transports chars,
while the latter two process bytes. They therefore justify no AOT change.
The second scalar-character gate (`run_length_encode_small`,
`levenshtein_small`, and `automata_dfa_small`) also rejects that primitive
candidate: only dynamic DFA predicates make the helper material, while typed
Levenshtein comparisons lower directly and run-length encoding is dominated by
UTF-8 output conversion. No AOT lowering, `RegexSet` branch, named container
rule, corpus shortcut, or pattern-specific recognition is justified. The next
independent gate was the public `String.bytes` conversion boundary. It is now
kept as a canonical primitive lowering after ASCII conversion, byte histogram,
and MD5 controls all shared the same conversion family and improved together.
The lowering retains the existing raw byte iterator and applies only to the
primitive `String` API; it does not create a regex- or container-specific
branch. The next independent evidence gate is shared `StringBuilder` work.

## Historical Module Layout
- New package namespace: `able.text.regex`.
- Source layout:
  - `stdlib/src/text/regex.able` — user-facing API and helper types.
  - `stdlib/src/text/automata.able` / `automata_dsl.able` — reusable automata primitives + DSL leveraged by the regex engine.
- `stdlib/src/text/regex_builder.able` — programmatic construction utilities.
- `stdlib/src/text/regex_program.able` — immutable compiled-program snapshots.
  - `stdlib/src/text/regex_set.able` — multi-pattern support (phase 2).
  - `stdlib/src/text/regex_scanner.able` — streaming interfaces (phase 2).
- No runtime-facing regex shims; the engine lives entirely in stdlib.

## Historical Public API Sketch
### Core Types
- `struct Regex { pattern: String, options: RegexOptions, program: RegexHandle }`
- `struct RegexOptions { case_insensitive: bool, multiline: bool, dot_matches_newline: bool, unicode: bool, anchored: bool, unicode_case: bool, grapheme_mode: bool }`
- `enum RegexError = InvalidPattern { message: string, span: Span } | UnsupportedFeature { message: string, hint: ?string } | CompileFailure { message: string }`
- `struct Match { matched: String, span: Span, groups: Array Group, named_groups: Map string Group }`
- `struct Group { name: ?string, value: ?String, span: ?Span }`
- `struct Span { start: i32, end: i32 }`

### Constructors & Helpers
- `Regex.compile(pattern: String, options: RegexOptions = RegexOptions.default()) -> Result Regex RegexError`
- `Regex.from_builder(builder: RegexBuilder, options: RegexOptions) -> Result Regex RegexError`
- `Regex::is_match(haystack: String) -> bool`
- `Regex::match(haystack: String) -> ?Match`
- `Regex::find_all(haystack: string) -> RegexIter`
- `Regex::replace(haystack: String, replacement: Replacement) -> String`
  - `Replacement` is either `Replacement::Literal(string)` or `Replacement::Function(fn(match: Match) -> string)`
- `Regex::split(haystack: String, limit: ?i32 = nil) -> Array String`
- `Regex::scan(haystack: string) -> RegexScanner` (lazy iteration with resumable state)
- `Regex::to_program() -> RegexProgram` (for introspection/export)

### RegexSet
- `RegexSet.compile(patterns: Array String) -> RegexSet | Error`
- `RegexSet.compile_with_options(patterns: Array String, options: RegexOptions) -> RegexSet | Error`
- Methods: `is_match(haystack: String)`, `matches(haystack: String) -> Array u64`
- Backed by a single combined automaton and source-pattern accept map to avoid
  per-pattern scanning; output indices are unique and source ordered.
- `iter(haystack: String) -> RegexSetIter` yields leftmost-longest,
  non-overlapping `RegexSetMatch` values; same-span identities are source
  ordered.

### Builder / Automata APIs
- `RegexBuilder` exposes the supported combinators: `empty`, `literal`,
  `any_scalar`, `char_class`, anchors, `concat`, `alternate`, repetition,
  `capture_group`, and `named_capture`. Unsupported grammar features such as
  lookaround are not represented by a partial builder API.
- `RegexProgram` is a deep snapshot with an NFA and deterministic capture
  metadata. `nfa_snapshot()` and `captured_names()` always return fresh data;
  `dfa_snapshot()` returns an error unless the NFA contains only literal and
  epsilon transitions. Graphviz formatting is ordinary tooling layered on the
  exported NFA, not a required runtime API.

### Streaming Scanner
- `Regex.scan(haystack: String) -> RegexScanner` starts an open input stream.
- `RegexScanner.feed(chunk: String) -> void` appends one chunk.
- `RegexScanner.next() -> Match | IteratorEnd` polls finalized matches; before
  `flush`, an empty result can be temporary.
- `RegexScanner.flush() -> void` idempotently finalizes partial matches at EOF.

## Historical Pattern-Semantics Sketch
- Code-point execution is the default: atoms, quantifiers, and spans advance in Unicode scalar units (`char`). When `grapheme_mode` is enabled the engine counts grapheme clusters returned by `String::graphemes()`, ensuring dot/quantifiers align with user-perceived characters.
- Syntax aligns with a conservative RE2 subset:
  - Literals, escaped characters, wildcard `.`
  - Character classes `[abc]`, ranges, negated classes `[^...]`, POSIX style `[:alpha:]`, Unicode classes `\p{}` / `\P{}`
  - Quantifiers `*`, `+`, `?`, `{m}`, `{m,}`, `{m,n}` with lazy variants `*?`, etc.
  - Anchors `^`, `$`, `\A`, `\z`, word boundaries `\b`, `\B`
  - Alternation `|`
  - Grouping `(...)` with numbered and named captures `(?P<name>...)`
  - Non-capturing groups `(?:...)`
  - Lookahead/lookbehind `(?=...)`, `(?!...)`, `(?<=...)`, `(?<!...)` (initially limited to fixed-width when lookbehind)
  - Inline flags `(?i)`, `(?-i)`, `(?i:...)` consistent with `RegexOptions`
- Backtracking is simulated by Thompson NFA or tagged DFA; catastrophic blowups are prohibited by construction.
- Unicode escapes support `\u{...}` and `\x{...}`; default mode treats the pattern as UTF-8 aware and works directly with scalar values. Grapheme mode matches clusters whose constituent scalars satisfy the pattern.
- Replacement backreferences use `$0`/`$1` or `${name}` syntax; unsupported references cause compile errors.

## Historical Character Model & Normalisation Notes
- The Able `char` type is a Unicode scalar value (`u32`). Literal processing ensures escapes resolve to a single scalar and leaves byte order untouched.
- Grapheme handling is provided by the standard library `Grapheme` iterator built atop `String`. Regex captures expose byte spans by default; when grapheme mode is enabled, helpers also return grapheme indices.
- Normalisation is opt-in via `RegexOptions` or preprocessing helpers (`String::to_nfc()` etc.); the engine itself does not mutate haystack data.

## Historical Implementation Architecture
1. **Front-end Parser (Able)**  
   - Recursive-descent parser producing `RegexAst`.  
   - Validates syntax, emits errors with spans.
   - Expands inline flags into scoped option stacks.

2. **IR & Compilation (Able)**  
   - Convert AST to a Thompson NFA with tagged transitions for capture groups and lookaround boundaries.  
   - Optionally determinize to a DFA or lazily run the NFA using thread sets.  
   - Emit a canonical `RegexProgram` structure (states, transitions, epsilon closures, tags).  
   - Store `RegexProgram` as stdlib data for interpretation and reuse.

3. **Execution Engine (Stdlib)**  
   - Regex execution remains in Able code using the automata primitives; no host extern hooks.

4. **Caching & Thread Safety**  
- `RegexHandle` wraps compiled program state and is shareable across threads/tasks.  
   - Memoize compiled handles per pattern/options combination to avoid duplicate compilation.

5. **RegexSet Implementation**  
   - Combine patterns into a single automaton by building a unified start state with tagged accept states.

6. **Streaming Scanner**  
- Stdlib manages incremental matching state, chunk boundaries, and `future_yield`-friendly yielding.

## Historical Testing & Tooling Plan
- Add Able-level unit tests covering:
  - Literal/quantifier/class behaviour, Unicode escapes, anchors, lookaround basics.
  - Replacement semantics (literal vs function).
  - `RegexSet` multi-match scenarios.
  - Streaming scanner with chunked input and overlapping matches.
- Extend fixtures:
  - Shared AST fixtures invoking regex APIs to ensure Go/TS parity.
  - Golden suites containing patterns and expected outputs compiled from the spec.
- Integrate existing `match_regex` matcher to call `Regex.compile(pattern).is_match(value)` and add new matcher tests.
- Provide fuzz harness hooks (host-level) to catch panic cases; wire into future CI when sandbox permits.

## Historical Rollout Plan
1. **Phase 0** — scaffolding
   - Land module skeleton, option structs, error enums, and `Regex.compile` returning `UnsupportedFeature`.
   - Add string/iteration design notes and spec TODOs covering byte spans, `char` vs `Grapheme`, and segmentation helpers.
   - Write spec outline in `spec/todo.md` describing required semantics.

2. **Phase 1** — core engine
   - Implement parser + NFA compiler for literals, classes, quantifiers, anchors.
   - Provide `is_match`, `match`, `find_all`, `replace` (literal replacement only).
   - Update testing matcher to use the real engine; add high-signal unit tests.

3. **Phase 2** — advanced features
   - Add lookaround, named groups, functional replacement, `RegexSet`, streaming scanner.
   - Introduce automata export (`to_program`, `to_nfa`, `to_dfa`).

4. **Phase 3** — tooling & builders
   - Publish programmatic builder APIs, Graphviz export, bytecode serialization for caching.
   - Document integration patterns and performance characteristics.

Each phase should land with synchronized Go/TypeScript implementations, updated fixtures, and spec text.

## Historical Follow-up Tasks
- Draft spec additions describing regex syntax, option semantics, and result structures.
- Update `PLAN.md` and `spec/todo.md` to track regex milestones.
- Coordinate with interpreter teams to validate stdlib regex performance and ensure TS/Go parity without host hooks.
- Revisit `docs/testing-matchers.md` once `match_regex` is backed by the real engine.
- Investigate exposing regex support in CLI tooling (e.g., future `able test --filter` flag) after Phase 1 stabilizes.
- Coordinate with string runtime work so `String::chars()` / `String::graphemes()` are available before Phase 1; add fixtures covering combining-mark and emoji segmentation.
