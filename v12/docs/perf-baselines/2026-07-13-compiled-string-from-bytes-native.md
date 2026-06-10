# Compiled `String.from_bytes` primitive lowering — 2026-07-13

## Decision

Keep a compiler-only lowering for the canonical primitive
`String.from_bytes_unchecked(bytes: Array u8) -> String` implementation. For
a non-nil monomorphized `Array u8`, generated Go now returns
`string(bytes.Elements)` directly. A nil input falls through to the existing
Able implementation, preserving its empty-string result.

The public `String.from_bytes` implementation is unchanged: it still calls
`utf8_validate` and returns its existing `Result String` error for invalid
UTF-8 before it reaches the unchecked construction step. The rule is matched
only by the canonical `able.text.string` primitive `String` static method, its
`*__able_array_u8` parameter, and `string` result. It excludes every nominal
type, including `StringBuilder`.

## Evidence

The prior bounded profiles found the identical construction path in three
independent applications:

`StringBuilder.finish` -> `String.from_bytes` -> UTF-8 validation ->
`String.from_bytes_unchecked` -> generic runtime-value bridge -> Go string.

The final step previously boxed every byte through `__able_array_u8_to` before
walking those boxed values again to build a Go string. The retained candidate
profiles remove only that repeated representation conversion. Focused
`StringBuilder.finish` allocation objects fell in every workload:

| Workload | Baseline | Kept lowering | Change |
| --- | ---: | ---: | ---: |
| `string_builder_small` | 694,235 | 416,533 | -40.0% |
| `run_length_encode_small` | 824,338 | 494,544 | -40.0% |
| `byte_histogram_small` | 84,915 | 50,947 | -40.0% |

Generated-main phase allocation totals also improved in every lane:

| Workload | Baseline bytes / objects | Kept bytes / objects |
| --- | ---: | ---: |
| `string_builder_small` | 93,786,352 / 1,941,096 | 67,752,824 / 1,662,989 |
| `run_length_encode_small` | 330,158,808 / 7,870,476 | 297,881,008 / 7,540,241 |
| `byte_histogram_small` | 14,396,528 / 239,304 | 12,066,408 / 205,328 |

The retained candidate allocation artifacts are in
`v12/interpreters/go/.profiles/` with these prefixes:

- `20260713_string_from_bytes_native_string_builder_small_compiled_`
- `20260713_string_from_bytes_native_run_length_encode_small_compiled_`
- `20260713_string_from_bytes_native_byte_histogram_small_compiled_`

## Timing gate

The final comparison used the canonical external `able-stdlib`,
`GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`. Each lane was built and run
three times in compiled mode with `bench_perf --cpu-affinity 15`; each
baseline/candidate pair produced the same stdout SHA-256 hash.

| Workload | Baseline | Kept lowering | Change |
| --- | ---: | ---: | ---: |
| `string_builder_small` | 0.2967 s | 0.2467 s | 16.9% faster |
| `run_length_encode_small` | 0.8633 s | 0.8100 s | 6.2% faster |
| `byte_histogram_small` | 0.1133 s | 0.1033 s | 8.8% faster |

Because the same primitive boundary is material in all three applications,
each lane improves, and UTF-8 validation behavior is separately exercised,
the lowering clears the broad-performance gate. No canonical-stdlib change is
needed: the language implementation remains the semantic authority and the
compiler only removes redundant Go-level representation work.

## Verification

- Focused compiler tests cover exact primitive-method recognition, nominal
  exclusion, static no-fallback lowering, and valid/invalid `String.from_bytes`
  results.
- Focused bytecode String byte-iterator tests pass unchanged.
- `git diff --check` passes.
- The repository-wide `./run_all_tests.sh` invocation stopped before Go tests:
  an existing untracked `exec/12_09_nested_spawn_native_context` fixture is
  absent from the already-modified exec coverage index. This tranche does not
  change that fixture or index.

## Next recommendation

Profile public `String.chars` and its UTF-8 decode/byte-bridge path across
three independent character-traversal applications. It remains the largest
distinct allocation owner in run-length encoding after this boundary is
removed, but it should receive a candidate only if the same caller and leaf
repeat across non-builder workloads. That keeps the next step focused on a
primitive language operation rather than a benchmark or container shape.
