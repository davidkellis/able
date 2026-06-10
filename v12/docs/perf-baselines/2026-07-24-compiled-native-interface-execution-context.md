# Compiled native-interface execution-context tranche — 2026-07-24

## Decision

Retain context-aware siblings for generated native-interface methods under the
existing experimental execution-context option.

The change is generic to every generated native interface. It does not name a
benchmark, stdlib container, or non-primitive nominal type. The ordinary
interface ABI remains available as a compatibility entry, and default compiler
output remains unchanged when the experimental option is disabled.

## Design

Compiled interface calls now pass their lexical execution context to an
additional generated interface method. Concrete adapters use the raw compiled
body only when the context's package-environment marker identifies the callee
package. Otherwise they use the guarded context entry, which switches to the
callee environment before running the body.

This distinction matters for spawned work. A task-local child environment is
not pointer-identical to its package environment, so the execution context
carries both:

- `env`, the current task-local environment;
- `packageEnv`, the package whose compiled body may be entered without a
  compatibility environment swap;
- `payload`, the current asynchronous task state.

Spawned children inherit the parent's `packageEnv` while retaining their own
task-local `env`. Cross-package interface calls retain the guarded entry.
Runtime adapters, dynamic boundaries, and callers using the old ABI derive a
context at the compatibility boundary.

## Repeated A/B results

Baseline and candidate executables were built independently, preserved, and
then timed under the same current runtime and canonical stdlib source. Every
timed process used `GOMEMLIMIT=1GiB`, `GOGC=50`, CPU affinity `0-3`, a
60-second process guard, and the benchmark's exact verifier.

| Application / executor | Baseline mean | Candidate mean | Reduction |
| --- | ---: | ---: | ---: |
| Packet Codecs / goroutine, pooled 10+10 | 0.838 s | 0.470 s | 43.91% |
| Audio Voices / goroutine, 5+5 | 1.404 s | 0.770 s | 45.16% |
| Scene Tiles / goroutine, 5+5 | 0.584 s | 0.324 s | 44.52% |
| Graph Visitors / goroutine, 5+5 | 0.426 s | 0.250 s | 41.31% |
| Packet Codecs / serial, 5+5 | 0.590 s | 0.360 s | 38.98% |
| Audio Voices / serial, 5+5 | 1.134 s | 0.614 s | 45.86% |

All 70 timed processes verified with zero failures or timeouts. Packet Codecs
used two independent five-process cohorts on each side because the workstation
samples were volatile:

- baseline: `0.84, 0.77, 0.89, 0.67, 0.73` and
  `0.99, 0.79, 0.94, 0.93, 0.83`;
- candidate: `0.46, 0.46, 0.44, 0.40, 0.49` and
  `0.37, 0.46, 0.54, 0.55, 0.53`.

The consistent improvement in four unlike concurrent applications and both
serial guards clears the broad-retention bar. It is not an isolated
single-benchmark result.

## Correctness coverage

Focused compiler tests cover:

- source-level interface dispatch threading the lexical context;
- unchanged default interface ABI;
- cross-package guarded bodies and package globals;
- a spawned interface call with a captured callback under the goroutine
  executor;
- nested spawn, cancellation, channels, background flush, and stdlib I/O;
- dynamic import/value boundaries and bound-method compatibility entries;
- native-interface execution with the option both enabled and disabled.

The experimental fixture parity groups and focused generated-binary tests pass.
The full repository suite is the final gate for this tranche.

## Residual profile and next boundary

A bounded Packet Codecs candidate profile still attributes 97.25% cumulative
time to `bridge.currentGID` and `runtime.Stack`. The native interface adapter
and implementation remain cumulative ancestors, but inspection of the
generated bodies shows that their context-aware path calls the raw compiled
body. The actual remaining environment swaps occur in the four captured
validator closures, which account for 20.18% to 26.61% cumulative apiece.

The next implementation candidate should therefore carry execution context
through the generated native callable ABI and captured-lambda invocation. It
must retain compatibility wrappers for dynamically stored or externally
entered callables, prove captured package-environment semantics, and guard at
least Packet Codecs, Audio Voices, Scene Tiles, Graph Visitors, and unlike
serial callback programs.

Before another production-source candidate is admitted, refresh the selected
compiled scorecard/closure evidence affected by this compiler-production
change. The evidence ledger correctly reports the compiled closures as stale;
focused A/B evidence is enough to retain this opt-in candidate, but it is not a
substitute for broad scorecard reconciliation.

No Able syntax, v12 language semantics, canonical `able-stdlib` source, bytecode
runtime, tree-walker, named-container rule, non-primitive nominal special case,
or WASM work is part of this tranche.
