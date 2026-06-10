# Concurrent Stencil Reduction application gate — 2026-07-22

## Decision

Retain the portable `concurrent_stencil_reduction` application, its signed
numeric input, source-equivalent Go/Python/Ruby implementations, exact verifier,
catalog and feature memberships, two complete measurement cohorts, and bounded
profile evidence. Retain no compiler, generated-runtime, bytecode VM,
tree-walker, canonical-stdlib, language, dependency, or WASM change.

The workload fills the requested concurrency × control-flow × arrays/files
frontier with application-shaped numeric work. Compiled execution reproduces
the already-closed goroutine-identity owner. Bytecode execution reproduces
already-closed Array and raw-integer families. Neither profile admits a new
generic candidate under the three-unlike-application rule.

## Application contract

The program reads 32 signed integers from `samples.txt`, then sends 2,048
overlapping stencil tasks through four long-lived workers. Every task performs
six passes of 24 shared read-only Array accesses, follows negative/even/odd
control paths, and emits a nominal result. The collector checks worker and
Future totals and computes schedule-independent buckets, value, energy, and
checksum reductions:

```text
2048:2048:2048:514,476,518,540:1033771877:837288448:356086
```

Tree-walker, bytecode, compiled Able, Go 1.26, Python 3.14, and Ruby 4.0
produce that output. Its SHA-256 is
`42870ec44f0b8a860e066ec155ce13e2916bbff632d74a5c87704f7f81fa4a3b`.
The catalog passes the real input path as a program argument, selects the
goroutine executor, assigns four logical CPUs to compiled/Go and one to the
interpreter lanes, and keeps the explicit source root isolated.

The external Able source differs from the canonical source only in its package
name, preventing a loader collision when both repositories are visible. No
canonical `able-stdlib` change was needed; the existing public Array, Channel,
Future, file, and argument APIs cover the application.

## Coverage result

The application genuinely covers lexical bindings and patterns; nominal types,
generics, and unions; expressions, arrays, text, and files; control flow;
inherent methods; nullable Channel receive handling; concurrency; packages and
imports; stdlib protocols; and real program entry. It does not claim callable
or interface-dispatch coverage.

The catalog and promoted scorecard now contain 51 portable applications. The
checked 165-triple frontier remains at minimum depth three with zero depth-zero
or depth-one triples and 160 improvements over the reconstructed baseline. The
priority concurrency × control-flow × arrays/files triple increases from three
to four independent applications. After promotion, represented target excess
is `162.312737` seconds.

## Repeated measurements

The workstation showed meaningful reference-process variance, so every timed
lane received a second complete five-process cohort. All samples were retained.
All 50 timed processes passed the exact verifier, with zero failures and zero
timeouts.

| Lane | Processes | Pooled mean | Cohort A | Cohort B | Limiting ratio |
| --- | ---: | ---: | ---: | ---: | ---: |
| Able compiled | 10 | 0.232000 s | 0.2140 s | 0.2500 s | 47.718× Go |
| Go 1.26 | 10 | 0.004861871 s | 0.005063382 s | 0.004660361 s | — |
| Able bytecode | 10 | 1.832000 s | 1.8620 s | 1.8020 s | 15.469× Python / 15.999× Ruby |
| Python 3.14 | 10 | 0.118434035 s | 0.096262245 s | 0.140605824 s | — |
| Ruby 4.0 | 10 | 0.114506908 s | 0.124048875 s | 0.104964942 s | — |

The separate cohort means moved by 3.28% for bytecode, 8.29% for Go, 15.52%
for compiled, 16.67% for Ruby, and 37.44% for Python. Pooling all ten samples
rather than selecting a favorable cohort keeps that workstation volatility
visible.

## Ownership and admission

Three verified compiled main-only profiles merge to 800 ms of CPU samples.
`bridge.currentGID` and its `runtime.Stack` descendant own 88.75%
cumulatively. Channel receive and send reach that boundary while consulting the
current task payload. This is the same exact generic owner already reproduced
across more than three unlike concurrency applications. Its fixed-context
replacement improved some concurrency rows but regressed other concurrent
applications and serial guards, including N-Body materially, so the rejected
candidate is not retried.

Three warmed bytecode-main profiles average 1,620,495,041 ns/op,
114,861,512 B/op, and 2,837,824 allocs/op, merging to 4.84 seconds of CPU
samples. `execArrayReadSlot` owns 37.81% cumulatively, but 97.27% of that parent
is the application’s full Array-read value path rather than one removable
leaf. The recurring exact `bytecodeRawIntegerValueInfo` leaf is 3.10% flat;
binary execution is 19.01% cumulative and type matching 2.69%. Array identity,
raw-integer carriers, binary dispatch, and type matching have already been
tested across unlike applications and either split among different children
or failed broad guards. A stencil, Channel, task, or named-container fast path
would violate the generality rules.

## Evidence

- application gate:
  `2026-07-22-concurrent-stencil-reduction-application-gate.json`;
- two Go, Python/Ruby, and Able comparison cohorts:
  `2026-07-22-concurrent-stencil-reduction-{go-reference,interpreter-reference,application-comparison}{,-cohort-b}.{json,md}`;
- checked interaction frontier:
  `2026-07-22-concurrent-stencil-reduction-interaction-triple-frontier.{json,md}`;
- merged profiles:
  `.profiles/20260722_concurrent_stencil_reduction_{compiled_main,bytecode_runtime}_merged.cpu.pprof`;
- readable profile tables:
  `2026-07-22-concurrent-stencil-reduction-{compiled,bytecode}-profile-top.txt`.

## Verification

- exact output parity in tree-walker, bytecode, compiled, Go, Python, and Ruby;
- ten verifier-backed timed processes per compiled, bytecode, and reference
  lane;
- three verified compiled main-only profiles and three successful warmed
  bytecode-main profiles;
- focused and complete catalog, feature-coverage, operation-depth, and
  interaction-triple checks;
- all added source files remain below 1,000 lines;
- `git diff --check`.

## Next recommendation

Complete `portable-concurrent-interface-data-application-frontier`.

Why: scorecard reconciliation is now complete, all 21 closures are current,
and every existing local or backend route remains closed. The highest-ranked
minimum-depth interaction is now concurrency × arrays/files × interface
dispatch, represented by only three applications and adjacent to 36.917
seconds of current target excess. Strengthening it with an unlike workload is
the most direct way to expose a reusable semantic owner rather than retesting
the rejected currentGID or Array/raw-integer mechanisms.

What it entails: add one deterministic file-driven application whose workers
process numeric or structured Array data through ordinary user-defined
interfaces, with source-equivalent Able/Go/Python/Ruby lanes and one verifier.
Take repeated cohorts in both Able modes and profile only an exact owner that
already appears in two unlike applications. Update canonical `able-stdlib`
only for a genuine reusable API or defect, and do not begin WASM work.
