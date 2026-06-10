# Truthiness/cast Sudoku quotient closure refresh

Date: 2026-07-21

## Decision

The `compiled-sudoku-quotient` closure is current against the corrected
truthiness/cast semantics. Keep no compiler, generated-runtime, VM,
canonical-stdlib, benchmark, reference, language, or WASM change.

Fresh repeated timing and two current main-only profiles reproduce the signed
Euclidean quotient owner, but only in Sudoku Masks. The nine generated Sudoku
application bodies call no shared truthiness or runtime-cast helper. Even an
impossible model that removes all signed-division CPU leaves the application
2.70x short of the 95%-of-Go target. A quotient-only or constant-three
specialization remains both too narrow and quantitatively insufficient.

## Frozen contract

- The v12 spec SHA-256 is
  `4f0405b86c122993723e8617abd6f825d9a8ff858d4c72acaf4e33469452f080`.
- The Sudoku source SHA-256 remains
  `88294708698dd72bd6ac6a6249633cc7fddf4274a33587930f8e932b00b199a5`.
- Every timed process used the catalog's serial one-CPU contract, external
  setup, public Ruby verifier, canonical external stdlib, and a 55-second cap.
- Two five-process cohorts were run for both Able and Go. Every successful
  sample is retained in the arithmetic means; no workstation outlier was
  removed.

## Repeated timing

| Lane | Cohort 1 | Cohort 2 | Pooled samples | Pooled mean | Pooled CV | Range |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Able compiled | 1.980 s | 2.166 s | 10 | 2.073 s | 9.44% | 1.830-2.560 s |
| Go 1.26 | 0.634165 s | 0.640473 s | 10 | 0.637319 s | 9.18% | 0.573397-0.732245 s |

All 20 timing processes verified with zero failures and zero timeouts. The
pooled result is 3.253x Go. Reaching 95% of Go throughput would require a
3.090x speedup.

## Corrected-path reach

The normal generated package contains nine Sudoku application bodies. A
body-by-body scan finds zero calls to `__able_truthy`, `__able_cast`,
`bridge.IsTruthy`, `bridge.Cast`, or `CastValueToType`. The source's one
explicit `(byte as i32)` site lowers directly to a primitive Go conversion.
Therefore neither corrected Error truthiness nor the catchable runtime-cast
failure path can affect Sudoku's hot generated execution.

One opt-in dynamic-boundary audit process verified and recorded 100 explicit
dynamic calls, 100 residual-polymorphic calls, and 100 host-ABI calls, all from
the existing I/O/output boundary. It records zero runtime-service calls and
does not introduce a truthiness/cast helper into the generated Sudoku bodies.
The opt-in binary was diagnostic only and was removed.

## Current quotient ownership

`square_index` still emits exactly two `__able_divmod_signed` calls, one for
each positive coordinate divided by three. Two independent main-only profiles
of the same normal binary merge to 3.91 seconds of CPU samples:

| Leaf | Flat | Cumulative |
| --- | ---: | ---: |
| `__able_divmod_signed` | 0.45 s / 11.51% | 0.49 s / 12.53% |
| `__able_compiled_fn_square_index` | 0.30 s / 7.67% | 1.37 s / 35.04% |
| `__able_compiled_fn_find_best_empty` | 1.21 s / 30.95% | 3.48 s / 89.00% |

The helper already takes its direct native `/` and `%` fast branch for
non-negative dividends and positive divisors. Treating its complete 12.53%
cumulative ownership as removable yields at most a 1.143x speedup and still
leaves a 2.703x target requirement. Actual quotient-only replacement would
remove less because it must preserve division-by-zero, overflow, signed
Euclidean semantics, control propagation, and the still-required quotient.

The unlike controls remain distinct:

- Rational Series is owned by `i128` nominal division/GCD, not the signed
  `i32` helper.
- Regex Set executes the NFA path; quotient-bearing DFA midpoint sites remain
  cold.
- K-Nucleotide's percentage-formatting quotient is too rare to receive CPU
  samples beside its map and string-conversion wall.

The same concrete leaf therefore has exact material breadth one, below the
required three unlike applications. No candidate was built or timed.

## Exact artifacts

- Able timing cohorts:
  `2026-07-21-truthiness-cast-sudoku-quotient-compiled.json`
  (`f94eb33bdf9f560cc6f0bb26f4bb28ca0f1a2a6a5a7a00a7b5f69075a4cca4fc`)
  and
  `2026-07-21-truthiness-cast-sudoku-quotient-c2-compiled.json`
  (`1083c576077f15e7864f18b1272efaa13984d35c201e7f7a2037374a8d403a92`).
- Go timing cohorts:
  `2026-07-21-truthiness-cast-sudoku-quotient-go-reference.json`
  (`f2750c555e4c0f888065fb859b316355f02ebff24d5f0c29891fb962c0b540af`)
  and
  `2026-07-21-truthiness-cast-sudoku-quotient-c2-go-reference.json`
  (`99f569ca05b0afdf8183edff35eefc31c5eded8cb60e120773da75fb68cd73a0`).
- Raw dynamic boundary telemetry:
  `2026-07-21-truthiness-cast-sudoku-quotient-compiled-reach.json`
  (`26a04d46282efd83d23d03f8c1ae9c3e091d19a522032e1773bc801ba43ad371`).
- Consolidated generated-path/profile census:
  `2026-07-21-truthiness-cast-sudoku-quotient-closure-reach.json`.

Raw generated packages, binaries, and profiles are temporary observer
artifacts and are not needed to build or test Able.

## Next recommendation

Use the now-current 21-closure ledger to select a new portable application
family from the underrepresented feature-interaction frontier before another
optimization experiment.

Why: every currently measured concrete mechanism is now current and closed or
guarded, and the frontier has zero actionable groups. Retrying a closed leaf
would tune an existing benchmark. A new independently authored application can
expose whether any cost repeats in real programs beyond the existing families.

What it entails: rank uncovered or shallow high-weight feature interactions,
choose one real application shape, and add source-equivalent Able, Go, Python,
and Ruby implementations with a shared verifier and repeated-process catalog
contract. Measure compiled and bytecode modes, then admit profiling only if a
concrete leaf recurs in at least two existing unlike applications as well. Make
canonical `able-stdlib` changes only for reusable specified behavior. Do not
begin WASM work.
