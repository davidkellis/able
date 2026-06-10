# Compiled `String.chars` native iterator — 2026-07-13

## Decision

Keep a compiler-only lowering for the canonical primitive
`String.chars() -> Iterator char` method. For a valid Go UTF-8 string whose
byte length fits Able's `i32` iterator limit, generated Go now creates the
existing `RawStringCharsIter` directly over a monomorphized `u8` slice.

Invalid or oversized values continue through the original Able body. That
body creates `validated_bytes` and raises `StringEncodingError` on invalid
UTF-8, so the lowering preserves the specified error behavior rather than
treating arbitrary Go strings as valid Able text. The matcher requires the
canonical `able.text.string` primitive method, its `string` receiver, and its
`Iterator char` return type; no `StringBuilder` or other nominal type matches.

## Shared evidence

Fresh compiled generated-main profiles used the external canonical
`able-stdlib`, `GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`. They cover
three different character-processing shapes:

- run-length encoding directly traverses text and re-encodes runs;
- Levenshtein converts text to `Array char` before dynamic programming; and
- DFA matching traverses through the canonical automata library.

Each reaches the same public primitive method and its
`validated_bytes` conversion boundary. The focused baseline attribution was
2,400,453 objects for run-length encoding (39.17% of its profile), 2,466 for
Levenshtein, and 665,648 for DFA matching (34.56%). The short Levenshtein
control has significant profile-writer noise, but its exact main-phase counter
records only 4,915 allocations, so the repeated `String.chars` path remains
material to that independent program.

The exact main-phase allocations improved in every workload:

| Workload | Baseline bytes / objects / GCs | Kept bytes / objects / GCs |
| --- | ---: | ---: |
| `run_length_encode_small` | 297,861,640 / 7,540,147 / 14 | 172,331,376 / 4,658,931 / 9 |
| `levenshtein_small` | 2,653,112 / 4,915 / 0 | 2,505,344 / 2,274 / 0 |
| `automata_dfa_small` | 90,148,336 / 2,146,605 / 6 | 56,475,808 / 1,495,471 / 4 |

Candidate artifacts are retained in `v12/interpreters/go/.profiles/` with
these prefixes:

- `20260713_string_chars_run_length_encode_small_compiled_`
- `20260713_string_chars_levenshtein_small_compiled_`
- `20260713_string_chars_automata_dfa_small_compiled_`
- `20260713_string_chars_native_run_length_encode_small_compiled_`
- `20260713_string_chars_native_levenshtein_small_compiled_`
- `20260713_string_chars_native_automata_dfa_small_compiled_`

## Timing gate

Each baseline/candidate lane was built and run three times in compiled mode
with `bench_perf --cpu-affinity 15`, the same external stdlib, and the same
memory/GC guardrails. Every paired lane has the identical stdout SHA-256 hash.

| Workload | Baseline | Kept lowering | Change |
| --- | ---: | ---: | ---: |
| `run_length_encode_small` | 1.5133 s | 0.6133 s | 59.5% faster |
| `levenshtein_small` | 0.0867 s | 0.0800 s | 7.7% faster |
| `automata_dfa_small` | 0.3600 s | 0.2300 s | 36.1% faster |

The rule removes the same redundant string-to-boxed-byte conversion from all
three distinct applications. It is a primitive String boundary, reuses the
existing iterator implementation, and retains the previous implementation for
the cases where its validation semantics matter. It therefore clears the
broad-performance gate. No canonical-stdlib source change is necessary.

## Verification

- Compiler tests cover exact primitive-method recognition, nominal exclusion,
  direct native iterator rendering, and a valid ASCII/two-byte/four-byte UTF-8
  character count.
- The existing bytecode `String.chars` invalid-UTF-8 fallback test remains
  part of the focused verification set.
- `git diff --check` passes.

The repository-wide `./run_all_tests.sh` remains blocked before Go tests by an
existing untracked `exec/12_09_nested_spawn_native_context` fixture missing
from the already-modified exec coverage index. This tranche leaves that
fixture and index untouched.

## Next recommendation

Profile the remaining canonical `RawStringCharsIter.next` adapter and
`utf8_decode` path across the same three workloads. After the accepted
construction change, those iterator-step frames still recur, particularly in
run-length encoding and DFA matching. A follow-up may be justified only if
one language-level iterator step and its error/offset behavior can be improved
across all three; it must not become a DFA- or builder-specific rule.
