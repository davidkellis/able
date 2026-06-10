# Compiled cross-corpus exact-leaf selection gate

Date: 2026-07-21

## Decision

Keep no compiler or generated-runtime code. Fresh preserved binaries for six
unlike compiled applications expose no new concrete compiler-controlled CPU or
allocation leaf that is material in at least three applications.

At the 1% flat CPU threshold, every exact symbol shared by three or more
applications belongs to Go GC/allocation machinery. Exact main-phase
allocation counters and caller attribution divide that GC work among matrix
backing, byte-output buffers, escaping wide nominal results, text/UTF-8/map
conversion, and regex-NFA storage. Combining those owners would optimize a Go
runtime parent rather than one Able lowering mechanism.

No compiler, generated runtime, bridge, bytecode VM, canonical stdlib,
benchmark, fixture, language, reference, scorecard, or WASM code changed.

## Preserved timing contract

All six Able binaries were built before timing began with one current compiler
binary (SHA-256
`1b826ea3848b6d124046d1bd753095ba544848ecfec5a389e5c585ba70aaf236`).
Each binary then ran in two oppositely ordered cohorts of five independent
processes. Fresh Go 1.26.4 references ran five independent processes each.
All 90 timed executions passed their public verifiers.

Processes used CPU 0, `GOMAXPROCS=1`, `GOGC=50`, a 1-GiB memory limit, and a
55-second cap. The canonical external stdlib contained 70 source files with
tree SHA-256
`6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`.
Arithmetic means follow the workstation measurement policy.

| Application | Able mean | Cohort spread | Go mean | Able / Go |
| --- | ---: | ---: | ---: | ---: |
| Matrix Multiply | 1.0950 s | 2.78% | 1.0520 s | 1.0409x |
| Mandelbrot | 0.1290 s | 11.48% | 0.0533 s | 2.4203x |
| Reverse Complement | 0.0880 s | 4.65% | 0.0165 s | 5.3333x |
| Fixed Width 128 | 0.1790 s | 3.41% | 0.0057 s | 31.4035x |
| Word Frequency | 0.1380 s | 0.00% | 0.0052 s | 26.5385x |
| Regex Set Audit | 0.1020 s | 0.00% | 0.0052 s | 19.6154x |

The target permits Able / Go up to `1 / 0.95 = 1.0526x`. Matrix Multiply
therefore meets the target in this complete current cohort; the other five do
not. This targeted cohort is decision evidence and does not by itself replace
the promoted aggregate scorecard source.

## Main-only CPU profiles

The generated phase profiler started immediately before registered Able
`main`, excluding package initialization and generated launcher/bootstrap
work. Short programs used many independent verified processes so their merged
profiles contain useful samples.

| Application | Verified profile processes | Merged main samples | Leading exact owner |
| --- | ---: | ---: | --- |
| Matrix Multiply | 3 | 3.33 s | generated `matmul`, 97.60% flat |
| Mandelbrot | 10 | 0.50 s | generated `pixel_byte`, 94.00% flat |
| Reverse Complement | 60 | 0.92 s | generated transform/wrapping plus GC and copy |
| Fixed Width 128 | 15 | 1.75 s | wide nominal construction/method entries plus GC/allocation |
| Word Frequency | 20 | 1.58 s | String split/conversion plus GC/allocation |
| Regex Set Audit | 30 | 0.90 s | NFA closure/upsert/move plus GC/allocation |

Every one of the 138 profile processes passed the same public verifier used by
the timing lane.

## Exact flat-symbol intersection

Only these exact symbols own at least 1% flat CPU in three or more
applications:

| Exact symbol | Applications | Flat-share range | Classification |
| --- | ---: | ---: | --- |
| `runtime.tryDeferToSpanScan` | 4 | 15.43%-21.74% | Go GC span selection |
| `internal/runtime/gc/scan.scanSpanPackedAVX512` | 4 | 1.71%-5.56% | Go GC scanning |
| `runtime.scanObject` | 4 | 1.14%-3.33% | Go GC scanning |
| `runtime.scanSpan` | 3 | 2.22%-5.43% | Go GC scanning |
| `runtime.findObject` | 3 | 1.27%-4.44% | Go GC metadata |
| `runtime.heapArenaOf` | 3 | 1.14%-3.26% | Go heap metadata |
| `runtime.(*spanInlineMarkBits).init` | 3 | 1.11%-2.86% | Go GC mark metadata |
| `runtime.alignDown` | 3 | 1.09%-2.22% | Go runtime helper |
| `runtime.mallocgc` | 3 | 1.27%-3.33% | Go allocator parent |
| `runtime.memclrNoHeapPointers` | 3 | 1.11%-3.26% | different allocation descendants |

No generated `main.__able_*`, compiler bridge, shared runtime, primitive
operator, nominal constructor, collection helper, string helper, or regex
helper survives the three-application intersection. The six applications are
executing materially different user work below the common Go runtime parents.

## Exact allocation reconciliation

One separate allocation-only phase process per preserved binary supplied exact
`runtime.MemStats` deltas. All six completed and verified. Start/end pprof
subtraction was used only for attribution because profile serialization itself
allocates; the counters below are the authoritative main-phase values.

| Application | Main bytes | Main allocations | Concrete owners |
| --- | ---: | ---: | --- |
| Matrix Multiply | 32,897,352 | 8,018 | input/output matrix backing in `build_matrix` and `matmul` |
| Mandelbrot | 86,320 | 66 | pre-sized render/output storage |
| Reverse Complement | 9,314,944 | 64 | file input plus reverse/complement and wrapped-output backing |
| Fixed Width 128 | 35,536,192 | 2,220,984 | escaping loop-carried results in two wide-nominal checksums |
| Word Frequency | 31,889,000 | 726,033 | String conversion, UTF-8 decode, Array conversion, formatting, and map entries |
| Regex Set Audit | 6,942,016 | 103,780 | NFA thread sets, codepoint decoding, match flags, and capture storage |

The Fixed Width owner is the documented loop-carried nominal-result closure:
the retained generic caller-owned `_into` ABI correctly leaves these results
on the heap because their identity escapes. Word and Regex repeat
`Utf8DecodeResult`, but the coverage-wide union-payload census already found
that only this one nominal definition allocates materially; specializing it
would violate the shared nominal-lowering rule. Matrix, Mandelbrot, and Reverse
do not corroborate either family.

The GC intersection is therefore downstream of different required allocation
graphs. No candidate clears the selection rule, so candidate A/B timing and
unrelated guards are intentionally not run.

## Verification and cleanup

- 30/30 fresh Go reference timing processes verified.
- 60/60 preserved Able timing processes verified.
- 138/138 main-only CPU profile processes verified.
- 6/6 exact allocation-only processes verified.
- `go test ./pkg/profilehook -count=1 -timeout 60s` passed in 0.131s.
- Focused generated-main/compiler boundary tests passed in 0.153s.
- Temporary generated trees, binaries, profiles, timing reports, and output
  captures are removed after this record is written.

## Next recommendation

Run the previously deferred packed eager integer-cache feasibility gate,
starting with the bytecode raw-i32 cache rather than another application-main
micro-optimization.

Why: this tranche proves that the six generated mains have no new shared Able
leaf, while prior `inittrace` evidence identifies one exact 57-58 ms,
approximately 38-MiB, 707k-allocation interpreter-package initializer in every
static compiled process. Lazy removal and package isolation were correctly
rejected because reducing the initial live heap caused 35%-44% more GC cycles
and regressed allocation-heavy Binary Trees. Packing the existing eager cache
could remove per-element allocation/initialization cost while retaining real
eager storage and stable cache semantics, offering a route to improve all
short compiled programs without workload selection or GC-policy changes.

What it entails: inventory raw-i32 and boxed-integer identity, dynamic-type,
pointer, stack, and materialization assumptions; prototype only the smaller
raw-i32 cache behind centralized accessors; and preserve eager initialization.
Measure `inittrace` wall, bytes, objects, and resulting heap goal before
expanding the representation. Require focused identity/type/stack tests and
repeated order-balanced arithmetic means for short compiled applications,
allocation-heavy Binary Trees, allocation-light TapeLang, and ordinary
bytecode numeric/map workloads. Revert if GC counts rise materially or any
broad wall-time guard regresses. Do not add named-type/container/application
rules, fake ballast, `GOGC` overrides, or WASM work.
