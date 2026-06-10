# Compiled primitive `utf8_decode` lowering — 2026-07-13

## Decision

Keep a compiler-only lowering for the exact internal
`able.text.string.utf8_decode(Array u8, i32, i32)` primitive function. For a
non-nil byte array and an in-bounds, non-empty range, generated Go calls
`utf8.DecodeRune` over the existing raw `u8` slice. A valid decode returns the
existing `Utf8DecodeResult` payload in the existing
`StringEncodingError | Utf8DecodeResult` union.

The lowering deliberately declines malformed input (`RuneError` with a width
of one), empty/out-of-bounds ranges, and nil arrays. Those cases execute the
previous generated Able body, preserving its exact `StringEncodingError`
message and offset behavior. A valid encoded U+FFFD has a width greater than
one and therefore remains a successful decode.

This is a canonical primitive String operation, not a lowering for
`RawStringCharsIter`, `StringBuilder`, DFA, or any other nominal type. The
matcher requires the canonical package, function name, three primitive Go
parameter types, and exact union return type. No canonical-stdlib source
changed because the existing Able implementation remains the semantic
fallback.

## Shared evidence

Fresh compiled generated-main profiles used the external canonical
`able-stdlib`, `GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`. The same
three independent applications used for `String.chars` reach the shared
decoder through direct traversal, character-array conversion, and canonical
DFA matching.

Before this change, the run-length main phase attributed 1,444,544 allocation
objects to `bridge.ToUint` through the decoder and 722,278 directly to
`utf8_decode`. The candidate removes that boxed-byte bridge cost while
retaining the decoder's existing union payload allocation. The exact main
phase counters improve in every workload:

| Workload | Baseline bytes / objects / GCs | Kept bytes / objects / GCs |
| --- | ---: | ---: |
| `run_length_encode_small` | 172,331,376 / 4,658,931 / 9 | 105,158,096 / 3,214,993 / 6 |
| `levenshtein_small` | 2,505,344 / 2,274 / 0 | 2,427,328 / 1,442 / 0 |
| `automata_dfa_small` | 56,475,808 / 1,495,471 / 4 | 48,815,712 / 1,337,050 / 4 |

The corresponding profiles are retained in `v12/interpreters/go/.profiles/`
with the prefix `20260713_string_utf8_decode_native_`. In particular, DFA
`bridge.ToUint` objects fall from 158,000 to zero and its direct decoder
objects fall from 78,999 to 79,006; the remaining decoder objects are the
unchanged union result representation rather than byte coercion.

## Timing gate

All timing lanes used compiled mode, `bench_perf --cpu-affinity 15`, the same
external stdlib, and the same memory/GC guardrails. Every baseline/candidate
pair has the same stdout SHA-256 hash.

| Workload | Runs | Baseline | Kept lowering | Change |
| --- | ---: | ---: | ---: | ---: |
| `run_length_encode_small` | 3 | 0.6133 s | 0.5067 s | 17.4% faster |
| `levenshtein_small` | 9 | 0.0967 s | 0.0911 s | 5.8% faster |
| `automata_dfa_small` | 3 | 0.2367 s | 0.2267 s | 4.2% faster |

The first three-run Levenshtein sample reported 0.0867 s versus 0.1100 s, but
both had the same 0.0600 s average user CPU time. The matched nine-run rerun
resolved that short-lane wall-clock noise in favor of the candidate, while its
exact allocation count independently fell by 36.6%. The retained change thus
improves each distinct application and removes the same primitive byte
coercion rather than optimizing a benchmark-specific iterator.

## Verification

- Focused compiler tests cover exact recognition and rendering, valid ASCII,
  two-byte, and four-byte Unicode iteration, and malformed-byte error
  fallback.
- Focused bytecode String iterator and invalid-UTF-8 fallback tests pass,
  protecting the shared language semantics.
- `git diff --check` passes.

The repository-wide `./run_all_tests.sh` remains blocked before Go tests by
the existing untracked `exec/12_09_nested_spawn_native_context` fixture missing
from the already-modified exec coverage index. This tranche leaves that
fixture and index untouched.

## Next recommendation

Profile the remaining pointer-backed primitive-result union allocations across
a mixed compiled set that includes string, numeric, and map workloads. The
decoder still allocates its `Utf8DecodeResult` union payload, but any follow-up
must demonstrate a shared result-representation wall beyond string traversal.
Only then consider an escape-safe, general primitive-union lowering; otherwise
leave the representation unchanged rather than creating an iterator-specific
fast path.
