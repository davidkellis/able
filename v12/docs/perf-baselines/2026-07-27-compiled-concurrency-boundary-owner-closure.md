# Compiled concurrency boundary-owner closure — 2026-07-27

## Decision

Retain no compiler, generated-runtime, bridge, interpreter, bytecode,
canonical-stdlib, language, dependency, benchmark-source, or WASM production
change.

Fresh interpreter-free profiles of `await_channel_mux`,
`mutex_await_journal`, and `mutex_work_queue` identify one exact shared owner:
runtime-callable and native Awaitable paths recover the active environment
through `bridge.currentGID`, which calls `runtime.Stack`. That path consumes
85.47%-91.36% of merged main CPU samples and is the only material CPU owner
repeated by all three applications.

This is not a new admissible candidate. The retained native-interface
execution-context experiment deliberately stops at runtime callables and
captured lambdas. Removing this lookup requires the broader callable
execution-context ABI explicitly excluded by this tranche. A Channel, Mutex,
Future, Awaitable, or application-specific shortcut would violate the general
lowering rule. Because no exact shared non-closed owner remains, no production
A/B candidate was advanced.

## Protocol and verification

All Able binaries were emitted with `--no-fallbacks`; final dependency graphs
omit `pkg/interpreter`. Runs used the goroutine executor, CPU affinity 0-3,
`GOMAXPROCS=4`, `GOMEMLIMIT=1GiB`, and `GOGC=50`.

The final evidence contains 120 successful verifier-backed processes:

- 15 ordinary compiled CPU-profile launches, nine main-allocation-stat
  launches, and three sampled allocation-profile launches;
- 15 exact `currentGID` counter-overlay launches;
- 15 typed-boundary and 15 dynamic-boundary counter launches;
- 15 equivalent-Go CPU-profile launches and nine single-main Go allocation
  launches;
- 15 CPU and nine allocation launches under the existing opt-in fixed-context
  diagnostic.

There were zero verifier failures and zero timeouts. Each equivalent-Go CPU
launch repeated the same benchmark main 100 times to move the 4-5 ms
references above profiler sampling noise; only the final result was printed
and verified. Allocation comparisons use one benchmark main per process.

## Shared CPU owner

| Application | Merged Able samples | `currentGID` cumulative | `runtime.Stack` cumulative | `__able_call_value_fast` cumulative |
| --- | ---: | ---: | ---: | ---: |
| await-channel-mux | 1.72 s | 85.47% | 85.47% | 63.37% |
| mutex-await-journal | 4.82 s | 91.08% | 90.25% | 88.38% |
| mutex-work-queue | 9.38 s | 91.36% | 90.94% | 83.48% |

The equivalent-Go profiles contain ordinary application work and Go
channel/Mutex scheduling. Their leading flat application owners are
`laneScore` at 13.79%, `entryScore` at 55.56%, and `jobScore` at 20.45%;
none contains an environment-recovery analogue.

An exact overlay incremented an atomic counter in the copied
`bridge.currentGID` implementation without changing production sources:

| Application | Five verified counts | Mean | Population CV |
| --- | --- | ---: | ---: |
| await-channel-mux | 23,555 × 5 | 23,555.0 | 0.00% |
| mutex-await-journal | 28,490; 29,160; 27,533; 28,205; 26,011 | 27,879.8 | 4.29% |
| mutex-work-queue | 63,552; 62,451; 62,527; 62,801; 63,741 | 63,014.4 | 0.94% |

The frequency and low dispersion confirm that `runtime.Stack` is application
work induced by the generated boundary, not accidental bootstrap or sampling
ballast.

## Allocation and transition evidence

| Application | Able main objects | Go main objects | Object amplification | `currentGID` sampled objects |
| --- | ---: | ---: | ---: | ---: |
| await-channel-mux | 194,070.33 | 4,114.00 | 47.17× | 13.39% |
| mutex-await-journal | 125,130.33 | 25.67 | 4,875.21× | 17.01% |
| mutex-work-queue | 330,010.33 | 26.67 | 12,375.39× | 22.73% |

The ordinary compiled main means are 10.95 MB, 7.10 MB, and 18.23 MB
allocated. The equivalent Go means are 314.49 KB, 3.51 KB, and 3.52 KB.
Allocation profiles also contain await registration/wakers, native wrapper
structs, AST type nodes, and runtime-callable values. Those sites are either
not shared across all three CPU profiles or sit beneath the same closed
environment boundary; none supplies a separate cross-application CPU
candidate.

Five typed-boundary runs per application confirm repeated
Array-to-runtime, interface round trips, callable-to-runtime conversion, and
error-control conversion. Five dynamic-boundary runs are stable at
1,025/2,049/4,097 residual-polymorphic calls and
2,560/2,052/4,100 runtime-service calls. These are reach counters, not CPU
attribution. The CPU profiles identify environment recovery below the
runtime-callable path as the concrete cost.

## Diagnostic excision result

The existing opt-in execution-context mode is not a candidate A/B in this
tranche; it was used only to test whether static-call context propagation
would expose a different residual owner. It did not:

| Application | Default `currentGID` | Fixed-context `currentGID` |
| --- | ---: | ---: |
| await-channel-mux | 85.47% | 87.23% |
| mutex-await-journal | 91.08% | 92.58% |
| mutex-work-queue | 91.36% | 92.94% |

Generated-source inspection explains the result:
`__able_call_value_fast` still obtains `__able_runtime.Env()`, and generated
native callable adapters still enter through a runtime `NativeCallContext`.
The option carries explicit context through static calls and native-interface
methods, but not through runtime-callable/captured-lambda invocation.

The next implementation needed to remove the owner would therefore be a
generic native-callable execution-context ABI with compatibility entries,
captured package-environment proof, cross-package guards, and a broad serial
regression gate. That is the same broad ABI family excluded here and cannot be
smuggled in as a local helper optimization.

## Retained evidence and tooling

- `2026-07-27-compiled-concurrency-boundary-owner-closure.json`
- `2026-07-27-compiled-concurrency-boundary-counts.jsonl`
- `2026-07-27-compiled-concurrency-current-gid-counts.tsv`
- `2026-07-27-compiled-concurrency-boundary-owner-profiles/`
- `2026-07-27-compiled-concurrency-{typed,dynamic}-boundary-audit.{json,md}`
- `bench_compiler_concurrency_boundary_instrument.py`
- `bench_go_reference_phase_profile.py`

The typed-boundary audit renderer now accepts the current nested
`telemetry.categories` payload; this corrects a tooling/report mismatch only.
The two new scripts generate diagnostic overlays or temporary Go-reference
wrappers and do not alter emitted production programs.

Cleanup removed the exact 1.9 GiB disk-backed tranche workspace and the
generated Python cache. No RAM-backed `/tmp/able-*` directory remains.

## Next recommendation

Keep the production pause for this owner unless the maintainer explicitly
authorizes a new broad callable execution-context design.

Why: the dominant compiled cost is real and directly prevents native-Go
performance, but every local route either repeats a rejected ABI experiment or
special-cases concurrency/kernel nominal shapes. What authorization would
entail: first specify an allocation-free context carrier for native callables
and captured lambdas; prove captured and cross-package environment semantics;
retain dynamic compatibility entries; and gate the prototype across these
three applications, packet codecs, audio voices, scene tiles, graph visitors,
and unrelated serial programs including N-body. Why it is important: only
removing the 85%-93% environment-recovery wall can materially change this
compiled family, and the previous default-context attempt regressed N-body by
54.7%, so correctness and broad performance must be designed together before
more production code is written.
