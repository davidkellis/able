# Able v12 String & Text Plan

## Goals
- Treat `String` as the canonical, host-independent representation of textual data while preserving the existing `string` primitive for literals and interoperable APIs.
- Offer a coherent API surface that distinguishes byte offsets, Unicode scalars (`char`), and grapheme clusters so every interpreter + spec reference can share the same semantics.
- Provide explicit normalization, case-mapping, and slicing helpers without forcing implicit transformations or hidden allocations.
- Keep the implementation portable: the TypeScript and Go runtimes must share the same data layout (`Array u8` + cached `len_bytes`) and call the same extern shims for host-specific work.
- Document the plan so downstream work (regex, formatting, IO, diagnostics) can rely on these contracts.

## Current Status (2025-02)
- `stdlib/src/text/string.able` defines `String`, `Grapheme`, `StringBuilder`, `StringChars`, and grapheme iterators with basic UTF-8 decoding plus `split`, `concat`, and `join`.
- `StringBuilder` can push chars/bytes/Strings and emit a validated `String`, but it lacks `reserve`, `into_string`, or formatting helpers.
- Grapheme iteration currently advances one Unicode scalar at a time; full extended grapheme-cluster detection is deferred until Unicode tables land.
- Missing API surface:
  - No byte/char/grapheme length helpers beyond `len_bytes`.
  - No substring, slice, or prefix/suffix helpers; existing Able code falls back to the primitive `string` type for these operations.
  - No normalization (`to_nfc`), case conversion (`to_lower`, `casefold`), or whitespace utilities (`trim_*`, `split_whitespace`).
  - No view type for zero-copy spans; regex and parsers must clone substrings.
  - No runtime story for incremental parsing (cursor/scanner) or streaming string building.

## Target API Surface

### Construction & Inspection
- `fn len_chars(self) -> i32` and `fn len_graphemes(self) -> i32` derived via iterators with optional cached counts.
- `fn is_empty(self) -> bool` (already exists) plus new helpers `is_ascii`, `is_nfc`, `capacity`.
- `fn byte_at(self, idx: i32) -> u8 | StringIndexError` plus `fn char_at(self, idx: i32) -> char | StringIndexError`.
- `fn from_iter(iter: (Iterator char)) -> Result String` to materialize from iterables without manual builders.
- UTF-16 bridges: `fn from_utf16(Array u16) -> Result String` and `fn to_utf16(self) -> Array u16` for host IO APIs.

### Slicing & Views
- `fn slice_bytes(self, start: i32, end: i32) -> Result String` (validated UTF-8 boundaries).
- `fn slice_chars(self, start: i32, end: i32) -> Result String` using char offsets.
- Prefix helpers: `strip_prefix`, `strip_suffix`, `starts_with`, `ends_with`.
- Splitting variants: `split_once`, `rsplit_once`, `splitn(limit)`, `split_whitespace`, `lines`, `rsplit`.
- Introduce `struct StringView { string: String, start: i32, end: i32 }`:
  - Zero-copy span over a `String`.
  - Methods: `to_string()`, `bytes()`, `is_empty()`, `len_bytes()`, `chars()`, `to_builtin()`.
  - Used by regex captures, parsers, diagnostics, and formatting to avoid repeated allocation.
- `fn view_bytes(self, start: i32, end: i32) -> Result StringView` and analogous `view_chars`.
- Add `ByteSpan { start: i32, end: i32 }` helper used by iterators and regex.

### Search & Pattern Helpers
- `fn contains(self, needle: String | string | char) -> bool` (ASCII fast path + iterator fallback).
- `fn index_of(self, needle: String | string, start: i32 = 0) -> ?i32` and `fn last_index_of`.
- `fn find_char(self, predicate: char -> bool) -> ?(i32, char)` returning byte offset + char.
- `fn count(self, needle: String | string | char) -> i32`.
- `fn compare(self, other: String, locale: ?string) -> Ordering` to feed sorting/helpers (locale support optional; default binary compare).

### Transformation & Normalization
- Whitespace APIs: `trim`, `trim_start`, `trim_end`, `collapse_whitespace`, `pad_start(width, char)`, `pad_end`.
- Case operations (ASCII fast path + Unicode fallbacks):
  - `to_upper`, `to_lower`, `to_title`, `casefold`, `capitalize`, `swapcase`.
- Replace helpers: `replace(old, new)`, `replace_first`, `replace_range(span, replacement)`, `repeat(times)`.
- Normalization:
  - Enum `NormalizationForm = NFC | NFD | NFKC | NFKD`.
  - `fn normalize(self, form: NormalizationForm) -> Result String NormalizationError`.
  - `fn is_normalized(self, form) -> bool`.
- Accent/diacritic utilities once `normalize` exists (e.g., `strip_diacritics` implemented as `to_nfd` + filter).

### Builders & Streaming
- `StringBuilder` additions:
  - `fn reserve(mut self, additional: i32) -> void`, `fn capacity(self) -> i32`, `fn clear(self) -> void` (already there but document), `fn shrink_to_fit`.
  - `fn push_view(mut self, view: StringView) -> void` to avoid intermediate `String`.
  - `fn push_grapheme`, `fn push_iter(iter: (Iterator char))`.
  - `fn into_string(self) -> Result String` transferring ownership without extra clone.
  - `fn write(self) -> Writer` adapter for formatting/IO layers.
- Introduce `StringCursor` (aka scanner) for incremental parsing:
  - Fields: `string: String`, `offset: i32`.
  - Methods: `peek()`, `next()`, `take_while`, `consume_prefix`, `remaining_view()`.
  - Enables lexers, format parsers, and templating engines without manual byte math.

### Supporting Modules
- `able.text.case`: shared case-mapping helpers + locale-aware tables (thin wrappers over host externs).
- `able.text.unicode`: data-oriented helpers (normalization forms, grapheme break properties, canonical combining classes).
- `able.text.views`: `StringView`, `ByteSpan`, view-specific iterators.
- `able.text.builder`: writer adapters bridging `StringBuilder` to formatting protocols.
- `able.text.scanner`: `StringCursor` plus tokenization helpers.
- These modules keep `String` lean while avoiding cyclic dependencies (e.g., regex only depends on views + unicode helpers).

## Runtime & Host Requirements
- Existing externs:
  - `__able_string_from_builtin(string) -> Array u8`
  - `__able_string_to_builtin(Array u8) -> string`
  - `__able_char_from_codepoint(i32) -> char`
- New extern shims (implemented per interpreter, exposed via `runtime/string_host.go|ts`):
  - `__able_string_case_map(bytes: Array u8, mode: i32, locale: string) -> Array u8`
  - `__able_string_normalize(bytes: Array u8, form: i32) -> Result Array u8 NormalizationError`
  - `__able_string_is_normalized(bytes: Array u8, form: i32) -> bool`
  - `__able_string_grapheme_break(bytes: Array u8, offset: i32) -> i32` (returns next break in bytes)
  - `__able_string_utf16_encode(bytes: Array u8) -> Array u16` and inverse
  - `__able_string_compare(bytes_a, bytes_b, locale: ?string) -> Ordering`
- Each extern should be pure and allocation-safe so the Able layer can wrap them in `Result`.
- Both runtimes must expose shared tests verifying that host + Able implementations agree on tricky Unicode samples (combining marks, emoji, RTL, etc.).

## Testing & Fixtures
- Extend `stdlib/tests/text/` with new suites:
  - `string_lengths.test.able` for `len_chars`/`len_graphemes`.
  - `string_slice.test.able` for `slice_*`, views, and prefix helpers.
  - `string_case.test.able` covering ASCII + Unicode case conversions.
  - `string_normalize.test.able` verifying each normalization form plus error handling.
  - `string_builder.test.able` already exists; augment with reserve/into_string tests.
- Add AST fixtures under `fixtures/ast/strings/` for trimming, substring, casefold, etc., so both interpreters exercise the same code paths.
- Update the spec TODO (string entry) once the docs derived from this plan land and tests exist.

## Phase Breakdown
1. **Phase 0 (done):** landed baseline `String`, iterators, builder, `split/join/concat`.
2. **Phase 1 – Foundational helpers:**
   - Implement `len_chars`, `len_graphemes`, `starts_with`, `ends_with`, `contains`, `index_of`, byte/charmap `slice` APIs, and `StringView` with ASCII-only validation.
   - Extend `StringBuilder` with `reserve`, `into_string`, and `push_view`.
   - Port existing Able code (`io/path`, fixtures) off the primitive `string` helpers where possible.
3. **Phase 2 – Unicode-aware slicing + cursor:**
   - Finish grapheme iterator + `len_graphemes` by integrating host `__able_string_grapheme_break`.
   - Add `StringCursor`, `split_whitespace`, `lines`, `repeat`, and `replace` helpers.
4. **Phase 3 – Case + normalization:**
   - Add `able.text.case` + `able.text.unicode`.
   - Wire `to_upper/lower/title/casefold`, normalization APIs, `strip_diacritics`, and UTF-16 bridges.
   - Ensure regex + formatting rely on the new helpers instead of ad-hoc host calls.
5. **Phase 4 – Locale-aware compare + perf work:**
   - Hook `String::compare` into host locale data.
   - Explore caching (char/grapheme counts) and view reuse to reduce allocations.
   - Audit stdlib for redundant `string` usage and migrate to `String`.

Each phase should update `PLAN.md` + spec references and add parity tests before merging.

## Open Questions & Risks
- **Normalization data:** Do we embed Unicode tables in the interpreters or rely on host ICU APIs? Decision impacts wasm portability and bundle size.
- **Caching strategy:** Should `String` cache `len_chars`/`len_graphemes`? Pros: faster queries. Cons: extra memory + invalidation complexity when cloning builders.
- **Locale handling:** How many locales do we expose via `String::compare` / `to_title(locale)`? Minimal viable plan is locale-agnostic binary comparison; full locale support may slip to v12.
- **View lifetime:** `StringView` holds a `String` by value (ref-counted via shared `Array u8`). Need to confirm GC/runtime semantics keep the underlying buffer alive.
- **Mutable slices:** We deliberately avoid exposing mutable views to preserve `String` immutability. Builders remain the only mutation path.
- **Interop with primitive `string`:** Long term we may alias `string` literals to the stdlib `String` without copies; this requires interpreter work (interning, lifetime tracking) and is outside this plan’s first three phases.

This document supersedes the earlier short-form note; keep it updated as phases land so spec writers and interpreter owners can rely on a single authoritative plan.
