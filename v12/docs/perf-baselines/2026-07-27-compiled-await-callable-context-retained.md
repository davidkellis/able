# Compiled await callable execution-context tranche retained

## Decision

Retain the general callable execution-context extension behind
`ExperimentalExecutionContext`. Do not enable it by default and do not refresh
the full compiled scorecard yet.

The extension carries an allocation-free `*__able_execution_context` through
native callable types, captured lambdas, bound method values, generated
runtime-callable adapters, the compiled await chain, and native Awaitable
callbacks. Static paths with a known environment now resolve `AwaitWaker` and
`AwaitRegistration` definitions through `bridge.Runtime.StructDefinitionIn`
instead of rediscovering the current goroutine.

The new callable ABI is emitted only when the experimental option is enabled
and an entry-package module contains an `await` expression. Await-free
applications retain their former callable ABI and do not emit
`__able_call_value_fast_ctx`. Default builds, dynamic compatibility entries,
captured and cross-package environment behavior, and interpreter-backed
programs remain unchanged.

This is a compiler/runtime-boundary rule, not a Channel, Mutex, Future,
Awaitable, benchmark, named-container, or non-primitive nominal special case.
No canonical stdlib, interpreter, bytecode VM, language, dependency, or WASM
source changed.

## Repeated application results

Each Able baseline, candidate, and equivalent-Go value is the mean of five
fresh, public-verifier-backed runs. Every Able build used `--no-fallbacks`;
the final dependency graphs omit `pkg/interpreter`.

| Application | Baseline Able | Candidate Able | Change | Go | Candidate / Go |
| --- | ---: | ---: | ---: | ---: | ---: |
| Await Channel Mux | 0.2880 s | 0.1240 s | -56.94% | 0.0049 s | 25.31x |
| Mutex Await Journal | 0.3480 s | 0.0360 s | -89.66% | 0.0040 s | 9.00x |
| Mutex Work Queue | 0.8560 s | 0.0420 s | -95.09% | 0.0046 s | 9.13x |

The application results justify retaining the opt-in design, but they do not
meet the product target. Compiled performance for this family is still
9.00x-25.31x slower than equivalent Go.

The four earlier await-free concurrency controls—Packet Codecs, Audio Voices,
Scene Tiles, and Graph Visitors—and unrelated N-body were regenerated under
the candidate. All five omit the new callable-context type/helper path, so the
new rule has zero execution reach there. This conservative scheduler-effect
gate was added after an ungated prototype exposed noisy and sometimes adverse
results in those controls. The pre-existing broader experimental static
context ABI remains opt-in.

The checked performance-evidence ledger conservatively selects all 12
compiled-related closures for `compiler-production` scope drift; the nine
bytecode-only closures remain current. The ledger was not silently rebased:
these three experimental rows and five zero-reach controls do not constitute a
full 63-application default-mode reach census.

## Boundary and allocation results

Exact `bridge.currentGID` overlay counts used five-run pre-tranche means and
three verified final runs:

| Application | Baseline mean | Candidate mean | Change |
| --- | ---: | ---: | ---: |
| Await Channel Mux | 23,555.0 | 8,707.0 | -63.03% |
| Mutex Await Journal | 27,879.8 | 97.7 | -99.65% |
| Mutex Work Queue | 63,014.4 | 206.0 | -99.67% |

Three main-process allocation-stat runs per side produced:

| Application | Baseline bytes | Candidate bytes | Change | Baseline objects | Candidate objects | Change |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Await Channel Mux | 10,948,504 | 10,542,160 | -3.71% | 194,060.0 | 189,484.7 | -2.36% |
| Mutex Await Journal | 8,013,376 | 5,489,608 | -31.49% | 143,647.7 | 101,028.0 | -29.67% |
| Mutex Work Queue | 16,281,541.3 | 11,205,434.7 | -31.18% | 291,272.0 | 207,095.0 | -28.90% |

The retained raw reports are:

- `2026-07-27-compiled-await-callable-context-baseline.{json,md}`
- `2026-07-27-compiled-await-callable-context-candidate.{json,md}`
- `2026-07-27-compiled-await-callable-context-go-reference.{json,md}`
- `2026-07-27-compiled-await-callable-context-current-gid-counts.tsv`
- `2026-07-27-compiled-await-callable-context-allocation-{baseline,candidate}.tsv`

## Correctness and structure gates

- Context-carrying native callable, captured lambda, bound method, native
  interface, cross-package, static spawn, nested spawn, kernel, and fixed
  runtime-helper guards pass.
- Default ABI, dynamic boundary, dynamic named/value boundary, and await-free
  reach guards pass.
- The public Mutex await contention test passes in default and experimental
  modes.
- Full experimental fixture parity passes in 50.558 seconds.
- `go test ./pkg/compiler/bridge` and `go test ./cmd/ablec` pass.
- The largest touched files remain below 1,000 lines.

## Next

First run a strict default/experimental reach census across all 63 portable
applications and resolve only the 12 selected evidence closures. Then refresh
final CPU, allocation, and exact boundary profiles across these three
applications plus unlike await-bearing controls and select the largest general
owner that repeats across at least three.

This is next because the compiler scope has changed and the callable boundary
was the measured 85%-93% owner, yet the family remains far from Go. The work
entails generating fresh strict binaries in both modes, verifying their
interpreter-free graphs and outputs, classifying actual path reach, refreshing
only reached evidence, then collecting repeated residual profiles. It is
important both to restore trustworthy closure evidence and to prevent the
remaining Await Channel Mux goroutine lookups from motivating another
one-family rule.
