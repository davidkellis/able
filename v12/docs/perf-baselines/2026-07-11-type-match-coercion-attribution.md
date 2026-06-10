# Bytecode type-match/coercion attribution — 2026-07-11

## Scope

This tranche followed the map-lookup attribution result. It used a temporary,
opt-in probe around runtime type matching, assignment-style coercion,
typed-pattern matching, and the simple-coercion helper. Each control ran once
after the benchmark harness's normal warmup under `GOMEMLIMIT=1GiB`,
`GOGC=50`, and `GOMAXPROCS=1`. The probe was removed after collection; normal
VM execution is unchanged.

The JSON evidence is retained under
`.profiles/20260711-type-match-diagnostics/`.

## Results

| Control | Repeated target expressions | Interpretation |
| --- | --- | --- |
| `string_split_join_small` | `StringEncodingError` typed pattern / fast simple coercion: 197,508 each; `Utf8DecodeResult` typed pattern: 197,490; union match: 203,508; `u8` match: 98,748 | Text codec success/error unions dominate this control. |
| `linked_list_iterator_collect_i64_small` | `IteratorEnd` typed pattern: 62,008; `IteratorEnd` match / fast simple coercion: 62,000 each; `Self` match / coercion: about 30,012 each | Iterator protocol termination and generic `Self` handling dominate. |
| `array_map_i32_small` | `Array` match/coercion: 8 each; `Self` match: 7 | Runtime type matching is negligible here. |
| `string_builder_small` | `char` match: 208; `String` coercion: 16; `Array` match/coercion: 8 each | Small primitive checks, not a recurring nominal path. |

The source-value categories agree with the exact names: iterator collect mixes
integer/raw-integer values with Array and iterator values at the iterator
boundary, whereas split/join repeatedly processes struct-backed string-codec
union values. None of those paths recurs materially in all controls.

## Decision

No type-match/coercion optimization was attempted or retained. A fast path for
`StringEncodingError`, `Utf8DecodeResult`, `IteratorEnd`, or `Self` would
improve a single library/protocol workload rather than a broad language
semantic path. It would conflict with the performance policy of choosing
optimizations by cross-program evidence, not by one benchmark's source shape.

The temporary diagnostic file and benchmark-harness output option were removed
after collection. Focused type-match, type-coercion, typed-pattern, and
benchmark-config tests remain the regression guard for the unchanged paths.
