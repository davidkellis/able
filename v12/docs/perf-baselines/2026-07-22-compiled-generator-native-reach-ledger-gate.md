# Compiled generator-native reach-ledger gate — 2026-07-22

## Decision

Reject the declaration reach-ledger candidate and retain no compiler,
generated-runtime, bridge, stdlib, application, benchmark, language, bytecode,
or WASM code change.

The prerequisite was stricter than the later omission benchmark: reach
bookkeeping with omission disabled had to be effectively free. The final
bounded implementation failed that prerequisite on the small Fib compiler
path. Three alternating processes averaged 0.456667 seconds for the immutable
baseline and 0.670000 seconds for the disabled ledger, a 46.72% regression.
This fixed cost would penalize small applications even if omitted Go later
compiled faster.

Word Frequency illustrates why both scales were required. Two alternating
processes averaged 33.715 seconds baseline and 32.865 seconds with the disabled
ledger, a 2.52% apparent improvement within normal workstation variation. That
larger row does not override the repeatable small-program regression: a broad
compiler mechanism must not trade small real programs for a noisy large-row
win. The six-application omission gate was therefore not run.

## Candidate exploration

Direct reference-edge recording at every output site was not a bounded change:
the compiler currently has roughly 525 non-test function-emission statements,
with many helpers assembled through shared buffers and late specialization
passes. Instrumenting that surface would duplicate Go lexical semantics across
the generator and impose a large maintenance contract for a throughput-only
optimization.

A conservative generator-owned in-memory ledger was explored instead. It
scanned already-rendered buffers without a Go AST or second formatter, found
top-level declarations, counted identifier references across every generated
Go file, and could omit only unique private receiverless declarations from
`compiled.go`. Methods, exports, bodyless/directive declarations, `main`,
`init`, operator roots, boundary-audit accessors, cross-file references, and
recursive groups were preserved. Synthetic tests covered those roots and
dynamic, function-value, interface, extern, initialization, and concurrent
surfaces.

The bookkeeping was progressively narrowed from the standard Go scanner to a
streaming scanner, a bounded lexer, a candidate-name second pass, hashed name
matching, and finally an ASCII fast path. The final implementation still had
to inspect the multi-megabyte generated package twice and could not satisfy the
disabled-cost gate. No application, nominal type, container, stdlib API, or
benchmark name influenced reach decisions.

## Final repeated measurements

Measurements used immutable binaries in alternating order,
`ABLE_SOURCE_ROOT_ONLY=1`, the canonical external stdlib, `GOMEMLIMIT=1GiB`,
`GOGC=50`, fresh output directories, and a 55-second bound per process. All
processes completed successfully.

| Application | Baseline samples | Disabled-ledger samples | Means | Change |
| --- | ---: | ---: | ---: | ---: |
| Fib | 0.42, 0.45, 0.50 s | 0.70, 0.63, 0.68 s | 0.456667 / 0.670000 s | **+46.72%** |
| Word Frequency | 33.97, 33.46 s | 33.78, 31.95 s | 33.715 / 32.865 s | -2.52% |

Focused reach tests passed before removal. After removal, the ordinary focused
compiler semantic selection and compiler bridge suite pass. An additional
whole-`pkg/compiler` attempt under a 60-second package-level timeout did not
finish: the timeout fired while boundary-fixture batch
`12_08_blocking_io_concurrency` had been running for six seconds. It reported
no assertion failure and is not used as candidate evidence. Temporary generated
trees and immutable experiment binaries were deleted.

The machine-readable summary is
`v12/docs/perf-baselines/2026-07-22-compiled-generator-native-reach-ledger-gate.json`.

## Next recommendation

Close the generated-helper omission track unless a future compiler architecture
already exposes a declaration graph as part of normal lowering. Return to the
portable application frontier and add one real source-equivalent workload from
the shallowest high-weight interaction set, with preference for a
non-concurrency validation/transformation workload so evidence is not dominated
again by the already-closed goroutine-identity wall.

Why: both general post-render approaches have now failed end-to-end compiler
cost gates, while omitted declarations are already linker-dead and cannot
improve application runtime. The project still has 83 target misses and zero
actionable current profile groups; new honest application evidence is more
likely to expose a shared mechanism than further tuning benchmark-visible dead
source.

What it entails: rerank the current feature-interaction ledger while discounting
closed mechanisms, implement equivalent Able, Go, Python, and Ruby programs
with one verifier and deterministic input, run repeated compiled and bytecode
catalog cohorts, and profile only when an exact leaf recurs in at least two
existing unlike applications. Admit an optimization only after it clears three
unlike families and the established guards. Update canonical `able-stdlib`
only for reusable specified behavior; do not begin WASM work.
