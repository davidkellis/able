# Able v10 String & Text Plan

## Goals
- Model `String` in the standard library as a thin wrapper around `Array u8`, keeping storage portable across interpreters while exposing a rich, host-independent API.
- Clarify the relationship between bytes, Unicode scalar values (`char`), and grapheme clusters so higher-level features (regex, formatting, iteration) share consistent semantics.
- Provide normalization and segmentation helpers without forcing implicit transformations on every string operation.

## Core Types
- `struct String { bytes: Array u8, len_bytes: i32 }`
  - `len_bytes` caches the logical length; `bytes.capacity()` stores allocation metadata supplied by the runtime.
  - Constructors:
    - `String::empty()`
    - `String::from_bytes(Array u8) -> Result String StringEncodingError` (validates UTF-8)
    - `String::from_utf8(string) -> String` (bridges from the builtin `string`)
  - Accessors / views:
    - `bytes(self) -> Array u8` (read-only view)
    - `chars(self) -> (Iterator char)`
    - `graphemes(self) -> (Iterator Grapheme)`
  - Mutation helpers defer to `StringBuilder` to maintain immutability semantics.

- `struct Grapheme { start: i32, end: i32, data: Array u8 }`
  - Represents a UTF-8 slice backed by the parent string’s byte array.
  - Implements `Clone`, `Eq`, and exposes `to_string()` to materialize a standalone `String`.
  - Produced exclusively via `String::graphemes()`; callers rely on `start/end` byte indices for slicing.

- `struct StringBuilder { buffer: Array u8 }`
  - Provides efficient append semantics while deferring UTF-8 validation until `finish()`.

- Error types:
  - `StringEncodingError { message: string, offset: i32 }`
  - `NormalizationError` (deferred; surfaced when host normalization fails).

## Iteration APIs
- `String::chars()` returns an iterator that decodes UTF-8 into `char` (Unicode scalar) values.
  - Backed by host-provided decoding helpers (`__able_utf8_next_scalar`).
  - Yields `(char, next_offset)` pairs internally; public iterator returns only the scalar.

- `String::graphemes()` segments the byte buffer using Unicode text segmentation rules.
  - Relies on host externs (`__able_utf8_next_grapheme`) returning byte spans.
  - Produces `Grapheme` values referencing the original byte array without copying.

- Additional helpers:
  - `String::len_bytes() -> i32`
  - `String::len_chars() -> i32` (derived from the iterator)
  - `String::len_graphemes() -> i32`

## Normalization
- Provide explicit functions: `to_nfc`, `to_nfd`, `to_nfkc`, `to_nfkd`.
- Each accepts `self` and returns a new `String`, allocating through `Array u8`.
- Under the hood, normalization uses host externs that take UTF-8 byte arrays and return normalized copies.
- No implicit normalization is performed during construction or comparison; callers choose the form required.

## Interop with Builtin `string`
- The interpreters will continue to expose the primitive `string` type for literals and runtime values.
- `String::from_builtin(value: string)` copies the underlying bytes into `Array u8`.
- `String::to_builtin()` returns a primitive `string` (lazy caching optional).
- Long term we can alias the builtin representation to the stdlib struct; for now the struct lives alongside primitives to unlock shared algorithms.

## Array Dependency
- `String` stores data in `Array u8`. All allocations, capacity management, and copy-on-write behaviour are delegated to the runtime-provided `Array` helpers (`with_capacity`, `write_slot`, `clone`, etc.).
- Range/slice operations expose byte offsets; helper methods translate grapheme boundaries to byte indices before calling `Array::slice`.

## Host Integration
- New extern helpers:
  - `__able_string_from_builtin(string) -> Array u8`
  - `__able_string_to_builtin(Array u8) -> string`
  - `__able_char_from_codepoint(i32) -> char`
- Future work: expose normalization and grapheme-segmentation externs once Unicode tables are wired in.

## Roadmap
1. Land `stdlib/v10/src/text/string.able` with type definitions, constructors, and iterator scaffolding (done; grapheme iterator currently groups by scalar until Unicode tables land).
2. Add unit tests exercising byte length, char/grapheme iteration on ASCII + combining-mark samples (tests may be skipped until host hooks exist).
3. Update documentation/spec:
   - Tie Section 6.1.5 to this plan.
   - Describe `Grapheme` and iteration semantics in `spec/TODO.md`.
4. Coordinate with interpreter teams to export UTF-8 decoding and text segmentation externs.
5. Integrate normalization helpers and builders, then migrate regex + formatting code to depend on the new APIs.
