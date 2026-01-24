# Able v11 Regex Module Design

## Background
- `match_regex` in `able.spec` is a placeholder that delegates to string equality.
- Section 5.5 of `design/stdlib-v11.md` outlines a long-term vision: an RE2-grade engine with deterministic performance and layered APIs (`Regex`, `RegexSet`, automata export, builders, streaming).
- No runtime support or parser exists today; the interpreters expose no regex primitives.

## Goals
- Provide a consistent, spec-backed `Regex` library for Able v11 with identical semantics across Go and TypeScript runtimes.
- Guarantee linear-time matching (no catastrophic backtracking) while supporting Unicode-aware patterns, lookaround, and rich replacement APIs.
- Expose layered functionality: immediate matching helpers, compiled regex handles, multi-pattern search, streaming scanners, and automata inspection.
- Keep the Able stdlib surface idiomatic: ergonomic modules, strong typing, clear error reporting, and integration with the existing testing matchers.
- Supply hooks for future tooling (bytecode export, visualization) without locking us to a specific host implementation.

## Non-goals / Deferred Items
- Backreferences and constructs that require non-regular backtracking remain out of scope for the first release.
- PCRE-incompatible features (e.g., conditional expressions) will only be considered after the baseline library is stable.
- Formal specification text will land after the core API is implemented and exercised across both interpreters.
- No initial attempt to expose JIT compilation or host-specific regex engines; deterministic portable semantics take precedence.

## Constraints & Principles
- Adopt RE2-style guarantees: every API must be linear in input length for fixed options and pattern.
- Default execution counts Unicode scalar values (`char`); grapheme-aware matching is opt-in via `RegexOptions.grapheme_mode`.
- Unicode correctness: pattern escapes and character classes operate on Unicode scalar values. When grapheme mode is enabled the engine segments haystacks using the standard library `Grapheme` iterator.
- Deterministic runtime: compiled regex values are immutable and can be shared across threads/procs safely.
- Error handling uses `Result`/`RegexError` with precise location data; mis-specified patterns never panic.
- Keep the regex engine entirely in Able stdlib with no host regex hooks.

## Current Implementation Notes (Phase 1 Partial)
- The stdlib regex module now supports literal-byte tokens with escapes and quantifiers (`*`, `+`, `?`, `{m}`, `{m,}`, `{m,n}`), plus `match`/`find_all` and streaming `scan`.
- Unsupported metacharacters still return `RegexUnsupportedFeature`; `match_regex` continues to fall back to string equality when compilation fails.
- The regex exec fixture is active again (`exec/14_02_regex_core_match_streaming`).

## Module Layout
- New package namespace: `able.text.regex`.
- Source layout:
  - `stdlib/src/text/regex.able` — user-facing API and helper types.
  - `stdlib/src/text/automata.able` / `automata_dsl.able` — reusable automata primitives + DSL leveraged by the regex engine.
  - `stdlib/src/text/regex_builder.able` — programmatic construction utilities (phase 3).
  - `stdlib/src/text/regex_set.able` — multi-pattern support (phase 2).
  - `stdlib/src/text/regex_scanner.able` — streaming interfaces (phase 2).
- No runtime-facing regex shims; the engine lives entirely in stdlib.

## Public API Surface
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
- `RegexSet.compile(patterns: Array string, options: RegexOptions) -> Result RegexSet RegexError`
- Methods: `is_match(haystack: string)`, `matches(haystack: string) -> Array i32`, `iter(haystack: string) -> RegexSetIter`
- Backed by a single combined automaton to avoid per-pattern scanning.

### Builder / Automata APIs (Phase 3)
- `RegexBuilder` exposes combinators (`literal`, `concat`, `alt`, `repeat`, `char_class`, `lookahead`, etc.).
- `RegexProgram` is a serialized representation with metadata (`states`, `transitions`, `accepting`, `groups`).
- Export helpers: `RegexProgram.to_nfa()`, `to_dfa()`, `to_graphviz()`.

### Streaming Scanner
- `RegexScanner.new(regex: Regex) -> RegexScanner`
- `RegexScanner.feed(chunk: string) -> void`
- `RegexScanner.next() -> ?Match` (resumes from internal state, enabling incremental search over streams)
- `RegexScanner.flush() -> void` finalizes partial matches at EOF.

## Pattern Semantics
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

## Character Model & Normalisation
- The Able `char` type is a Unicode scalar value (`u32`). Literal processing ensures escapes resolve to a single scalar and leaves byte order untouched.
- Grapheme handling is provided by the standard library `Grapheme` iterator built atop `String`. Regex captures expose byte spans by default; when grapheme mode is enabled, helpers also return grapheme indices.
- Normalisation is opt-in via `RegexOptions` or preprocessing helpers (`String::to_nfc()` etc.); the engine itself does not mutate haystack data.

## Implementation Architecture
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

## Testing & Tooling
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

## Rollout Plan
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

## Follow-up Tasks
- Draft spec additions describing regex syntax, option semantics, and result structures.
- Update `PLAN.md` and `spec/todo.md` to track regex milestones.
- Coordinate with interpreter teams to validate stdlib regex performance and ensure TS/Go parity without host hooks.
- Revisit `docs/testing-matchers.md` once `match_regex` is backed by the real engine.
- Investigate exposing regex support in CLI tooling (e.g., future `able test --filter` flag) after Phase 1 stabilizes.
- Coordinate with string runtime work so `String::chars()` / `String::graphemes()` are available before Phase 1; add fixtures covering combining-mark and emoji segmentation.
