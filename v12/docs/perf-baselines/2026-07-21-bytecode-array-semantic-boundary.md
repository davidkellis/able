# Bytecode Array semantic-boundary gate

## Decision

**no-go-array-wrapper-or-storage-generalization**.

The exact canonical `Array.push` route is material in three unlike applications, but the shared wrapper is already cache-hot and its concrete storage descendants separate. No performance candidate is retained.

A new opt-in per-kind counter is retained in `BytecodeStatsSnapshot`; it adds no work when stats are disabled and makes future Array attribution exact.

## Fresh warmed-main evidence

| Application | Mean ns/op | Bytes/op | Allocs/op | Array-slot calls | Push calls | Cache hit | Array-slot CPU | Push CPU | Perfect push removal |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| `array_slice_window` | 575,390,208.5 | 14,186,504.0 | 422,243.0 | 888,266 | 288,259 | 99.9990% | 15.79% | 4.39% | 1.046x |
| `matrixmultiply` | 4,664,841,160.5 | 308,627,132.0 | 14,032,468.5 | 4,004,001 | 4,004,000 | 99.9998% | 28.85% | 26.70% | 1.364x |
| `reverse_complement` | 2,996,422,763.0 | 213,605,376.0 | 3,542,854.0 | 8,067,000 | 4,033,607 | 99.9998% | 28.26% | 8.03% | 1.087x |

Each mean is two independent fresh processes. CPU percentages come from the two merged main-only profiles, not whole-process CLI profiles. Stats counts come from separate bounded observer processes and are not timing evidence.

## Exact member-kind census

| Application | len | read_slot | write_slot | push | Fast-path misses |
| --- | ---: | ---: | ---: | ---: | ---: |
| `array_slice_window` | 312,002 | 288,004 | 1 | 288,259 | 0 |
| `matrixmultiply` | 1 | 0 | 0 | 4,004,000 | 0 |
| `reverse_complement` | 2,033,361 | 2,000,000 | 32 | 4,033,607 | 0 |
| `fixed_width_128` | 0 | 0 | 0 | 0 | 0 |
| `word_frequency` | 9,688 | 0 | 0 | 131,073 | 0 |

## Reconciliation

- **`array_slice_window`:** required independent slice backing, lease tracking, and Array size/read work.
- **`matrixmultiply`:** monomorphic f64 ArrayStore size/write/synchronization; the separate dot-loop owns 26.26% cumulative.
- **`reverse_complement`:** retained raw-u8 extraction and ArrayStoreAppendU8Promote.

The earlier `validated-array-push-shell` candidate remains rejected: The general wrapper split regressed String Split/Join 13.49% and Iterator Collect 3.42% in repeated order-balanced processes. The fresh census supplies breadth but no invalidating fact: all cached calls finish on the current fast path and there are zero fast-path misses.

Even perfect removal of the complete push subtree would yield only the Amdahl limits shown above. Matrix Multiply already meets the Python/Ruby target; the 4.6% and 8.7% ceilings for Array Slice and Reverse Complement cannot approach their required 20.94x and 117.50x speedups.

## Next recommendation

Return to the compiled residual frontier and select a shared generated-code semantic owner across at least three unlike applications.

Why: the interpreter's remaining shared Array parent is now causally closed, while compiler performance remains an independent project target. This entails normalizing generated-Go operations and allocations by logical work across unlike high-excess applications, then admitting only one exact general lowering or runtime owner present in at least three families. No WASM work is included.

## Reproduction

```sh
v12/bench_array_semantic_boundary \
  --json-out v12/docs/perf-baselines/2026-07-21-bytecode-array-semantic-boundary.json \
  --markdown-out v12/docs/perf-baselines/2026-07-21-bytecode-array-semantic-boundary.md
python3 v12/bench_array_semantic_boundary_test.py
```
