# Compiled lazy bytecode-cache gate

Date: 2026-07-16

## Decision

Reject and fully revert lazy initialization of the fixed bytecode integer-box cache.
The candidate gives a large, repeatable complete-process startup and memory win for
short compiled programs, but it slows the allocation-heavy Binary Trees application
by 7.74% across ten alternating samples per side. The regression is explained by a
35% increase in Go GC cycles, not verifier failure or output drift. Keeping it would
trade fast short benchmarks for a slower real allocation-heavy program.

No compiler, interpreter, runtime, stdlib, application, reference, or benchmark code
is retained from this tranche.

## Candidate

The former package `init` eagerly built fixed boxed integer values from `-256` through
`16,384` for twelve integer suffixes plus the extended `i32` cache through `262,143`.
The trial replaced that eager call with `sync.Once` initialization at `newInterpreter`
and `newBytecodeVM`. Cache lookup functions remained unchanged after construction, so
the experiment added no `Once` or atomic check to integer hot paths.

An isolated subprocess test proved that the candidate cache was absent after Go
package initialization, initialized before `NewBytecode` returned, and preserved
stable cached-value identity. Focused fixed/dynamic boxing, stack-snapshot, and unary
integer tests passed. Candidate code and candidate-only tests were removed after the
performance rejection.

## Alternating complete-process gate

Baseline and candidate compiled binaries were built once from the same generated
application sources. The only linked-source difference was eager versus interpreter-
boundary cache initialization. Five baseline and five candidate launches then ran in
alternating order for each application under `GOMEMLIMIT=1GiB`, `GOGC=50`, and
`GOMAXPROCS=1`. Every one of the 50 launches passed its external verifier, and each
application had one identical output hash across both variants.

| Application | Shape | Baseline mean | Candidate mean | Time delta | Baseline RSS | Candidate RSS | RSS delta |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Array Slice Window | short Array operations | 0.060 s | 0.020 s | -66.7% | 50,079 KiB | 28,356 KiB | -43.4% |
| Document Audit | short nominal/JSON work | 0.060 s | 0.020 s | -66.7% | 51,742 KiB | 29,582 KiB | -42.8% |
| Dependency Plan | short graph/collection work | 0.060 s | 0.020 s | -66.7% | 49,220 KiB | 27,526 KiB | -44.1% |
| Binary Trees | long allocation-heavy recursion | 28.424 s | 30.586 s | +7.61% | 282,903 KiB | 253,601 KiB | -10.4% |
| TapeLang Alphabet | long allocation-light dispatch | 3.386 s | 3.406 s | +0.59% | 50,442 KiB | 28,978 KiB | -42.6% |

The short-program result is exact at the timer's 0.01-second resolution: all fifteen
baseline samples were 0.06 seconds and all fifteen candidate samples were 0.02.
TapeLang is neutral within spread. Binary Trees fails the broad wall-time bar.

## Binary Trees volatility repeat

Because the first Binary Trees batch was more variable than the other rows, a second
independent set of five alternating samples per side was run with `GODEBUG=gctrace=1`.
All ten outputs again passed verification and retained the same hash.

| Variant | Second-batch mean | Range | Mean peak RSS | Mean GC cycles |
| --- | ---: | ---: | ---: | ---: |
| Eager baseline | 27.468 s | 26.34-28.22 s | 290,226 KiB | 172.8 |
| Lazy candidate | 29.630 s | 28.84-30.44 s | 232,703 KiB | 233.8 |

Across both batches, baseline averages 27.946 seconds and candidate averages 30.108
seconds: a 7.74% candidate regression. The lazy binary performs about 35.3% more GC
cycles in the diagnostic batch. The fixed cache is live heap at process start; under
the benchmark's explicit `GOGC=50`, it raises the initial heap goal and unintentionally
acts as GC ballast. Removing it lowers RSS but makes the real node-allocation loop
collect much more often.

Changing the generated application's GC percentage to compensate would override the
user/process runtime policy and turn an unrelated bytecode cache into a compiler GC
tuning mechanism. Retaining dummy ballast or selecting cache size based on a workload
would be equally unsound. The correct decision is to restore eager behavior until the
compiled runtime can be isolated from interpreter dependencies without creating a
wall-time regression.

## Verification and cleanup

- Focused cache/integer tests pass on restored source.
- Candidate initialization hooks and tests are absent.
- The original eager package initializer is restored.
- 60 alternating or repeat performance launches passed their external verifiers.
- Temporary paired binaries, generated sources, stdout captures, timing files, and
  GC traces are removed after this record.

## Next recommendation

Audit the generated compiled runtime's dependency on the monolithic interpreter
package and identify the smallest shared primitive/runtime boundary that can be
separated without changing GC policy.

Why: compiled binaries currently import interpreter bridge operations and therefore
run bytecode/tree-walker package initializers even when those subsystems are unused.
The lazy-cache experiment proves that this baggage materially costs startup and RSS,
but also proves that deleting one initializer in isolation changes GC pacing and hurts
allocation-heavy applications. A dependency-level view is needed before another
startup change can be safe.

What it entails: use generated-source imports, `go tool nm`, linker reachability,
`GODEBUG=inittrace=1`, and bounded CPU/allocation profiles for at least three unlike
compiled applications to map which interpreter/bridge helpers are actually reachable
and which initializers are collateral. Select a candidate only if one primitive or
shared runtime helper can move behind a lower-level package boundary used by compiled
code without importing VM state. Re-run alternating short startup, allocation-heavy,
allocation-light, bytecode startup, and bytecode hot-path controls. Do not change
`GOGC`, retain dummy heap ballast, begin WASM work, or add benchmark/container/type
special cases.
