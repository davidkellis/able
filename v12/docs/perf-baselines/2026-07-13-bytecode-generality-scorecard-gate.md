# Bytecode generality scorecard and CPU-miss gate — 2026-07-13

## Decision

Keep no bytecode VM, compiler, runtime, or canonical-stdlib performance change.
The fresh Python/Ruby scorecard confirms material interpreter gaps, but the two
runnable target-miss families either share an already-rejected raw-float lane
or have distinct concrete VM leaves. No new general VM operation is established.
Do not add a text, byte-array, codec, Mandelbrot, or benchmark-shaped fast path.

## Fresh scorecard

Fresh local Python and Ruby implementations each used one CPU-15-pinned,
verifier-backed process with `GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and
a 45-second cap. The matching Able bytecode process used the same pin, cap,
canonical external stdlib, and verifier. A one-process Able row is a status
screen, not a variance claim. References stop after their first timeout, and
timeouts remain status-only rather than becoming ratios.

| Benchmark | Bytecode (s) | Bytecode/Python | Bytecode/Ruby | Status |
| --- | ---: | ---: | ---: | --- |
| Fib | 0.1600 | n/a | n/a | bytecode verified; references cap-bound |
| BinaryTrees | timeout | n/a | n/a | cap-bound |
| MatrixMultiply | 4.4800 | n/a | n/a | bytecode verified; references cap-bound |
| QuickSort | timeout | n/a | n/a | bytecode cap-bound |
| Sudoku | timeout | n/a | n/a | bytecode cap-bound |
| Sudoku Masks | timeout | n/a | n/a | bytecode cap-bound |
| I-Before-E | 0.5500 | 5.40x | 4.20x | verified miss |
| Base64 | 4.3500 | 1.10x | 1.74x | verified miss |
| JSON | 1.2000 | 0.46x | 0.69x | verified control |
| Monte Carlo Pi | 2.6800 | 1.74x | 1.61x | verifier-accepted nondeterministic miss |
| PiDigits | 2.7100 | 0.67x | 0.26x | verified control |
| Mandelbrot | 6.7900 | 5.22x | 3.48x | verified miss |
| Reverse Complement | 6.7500 | 160.71x | 80.26x | verified miss |
| K-Nucleotide | timeout | n/a | n/a | bytecode cap-bound |
| N-body | timeout | n/a | n/a | bytecode cap-bound |
| TapeLang Alphabet | timeout | n/a | n/a | cap-bound |

Of the seven completed Python/Ruby pairs, JSON and PiDigits meet the
95%-of-each-interpreter target (ratio at most `1.0526x`). The fresh reference
and comparison artifacts are
`2026-07-13-bytecode-generality-interpreter-refresh.*` and
`2026-07-13-bytecode-generality-scorecard.*`.

## CPU-only repeated-miss gate

Only two application pairs qualify for a current bounded VM check: Monte Carlo
Pi/Mandelbrot are independently written numeric misses, while I-Before-E and
Reverse Complement are independently written text/byte misses. Their normal
scorecard executions first passed canonical verifiers. The warmed runtime
benchmark then loaded and warmed each program before its CPU profiler began.

| Workload | Steady-state CPU sample | Material current-source work | Result |
| --- | ---: | --- | --- |
| I-Before-E | 290 ms | cached member lookup and member-call dispatch | does not recur in Reverse Complement |
| Reverse Complement | 7.70 s | Array-slot calls/push, integer boxing, raw integer metadata | does not recur in I-Before-E |
| Monte Carlo Pi | 2.56 s | normalized raw-float stores and fused float jump | repeats the already-rejected float lane |
| Mandelbrot | 7.18 s | raw-float stores/jumps, binary execution, stack/GC work | repeats the already-rejected float lane |

The numeric pair repeats a language-level float/control-flow family, but the
current implementation already contains its generic raw-float lanes and prior
generic representation, store, and quickening experiments failed broad guards.
The text/byte pair does not share a concrete child below `runResumable`.
Base64 and JSON remain independent codec/parser controls; neither validates a
generic VM change for the two misses.

The retained CPU profiles are
`v12/interpreters/go/.profiles/20260713_bytecode_generality_miss_*.cpu.pprof`.

## Why no candidate is justified

A raw-float rewrite would revisit a repeatedly rejected generic candidate
without new cross-workload evidence. A member-cache change would be selected by
I-Before-E alone, and an Array/boxing change by Reverse Complement alone; both
have already failed broader application guards. The scorecard also does not
justify turning timeout rows into ratios or extending them only to manufacture
a profiling target.

No canonical `able-stdlib` change is needed.

## Verification

- Every completed Python/Ruby reference and bytecode scorecard row passed its
  canonical verifier; timeouts are explicitly retained as such.
- All four warmed CPU profile launches completed. Their public-output
  correctness is established by the preceding verified normal process.
- No source behavior changed, so no new semantic test is required.
- `git diff --check` passes.

The repository-wide `./run_all_tests.sh` remains blocked before Go tests by
the existing untracked `exec/12_09_nested_spawn_native_context` fixture missing
from the already-modified exec coverage index. This tranche leaves that
fixture and index untouched.

## Next recommendation

Add a second application-level concurrency benchmark with Go, Python, Ruby,
and Able implementations plus a deterministic verifier. It should exercise a
bounded producer/worker/collector pipeline with `spawn`, Future completion,
yield/flush, and cancellation semantics, but use a different data shape from
Channel Rollup. Current external coverage has only Channel Rollup as a
cross-language concurrency application, so it cannot prove a scheduler
optimization is general. Once both applications have fresh pinned scorecards,
profile a scheduler/helper only if the same concrete descendant repeats in
both, with serial text/numeric controls guarding against global regressions.
