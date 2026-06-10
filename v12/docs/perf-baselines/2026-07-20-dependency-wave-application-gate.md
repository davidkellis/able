# Dependency-wave feature-interaction gate — 2026-07-20

## Decision

Retain one portable graph/state-machine application and its catalog coverage.
Retain no compiler, generated-runtime, bytecode-VM, canonical-stdlib, language,
nominal-lowering, or WASM performance change.

The application closes the six intended control-flow interaction gaps and
exposes large product misses, but its exact profiles reproduce owners that have
already completed broad candidate gates. Building another candidate here would
retry rejected scheduler, member-cache, return, or type-match work without new
cross-application evidence.

## Application contract

Dependency Wave Validation processes 4,096 nodes in 32 dependency waves with
four long-lived workers. Each wave waits for all predecessor results before it
constructs the next wave. Workers validate nominal tasks as `Result i64`, map
successful values through a closure and six-round state transition, recover
invalid nodes, and return an `Accepted | Recovered` nominal state. The main
routine stores results in a `HashMap` accessed through the public `Map`
interface and aggregates only commutative values.

This combines nested control flow, nominal structs and a payload union,
inherent methods, interface dispatch, closures, Result/Error handling, HashMap,
Channel/Future concurrency, and canonical stdlib protocols. Able, Go 1.26,
Python 3.14, and Ruby 4.0 implement the same graph, validation rules, state
transitions, four-worker schedule, and aggregation.

Expected output in all four implementations:

```text
4096:32:4096:4096:4054:42:2053151855:383298
```

The canonical and sibling Able sources are byte-identical. The sibling suite
also supplies Docker contracts, an external verifier, and a README.

## Interaction coverage

The application closes every previously empty intersection between control
flow and concurrency, closures, inherent methods, interfaces, Option/Result,
stdlib protocols, and nominal types. Across the full three-application feature
interaction tranche:

| Measure | Before tranche | Current |
| --- | ---: | ---: |
| portable/mixed families | 11 | 11 |
| pairwise interactions | 55 | 55 |
| zero-coverage pairs | 29 | 9 |
| pairs improved | — | 39 |

All six gaps closed by this application are control-flow intersections. At
this tranche boundary, nine empty pairs remained, dominated by the deliberately
selective lexical/binding/pattern family plus closures with real program entry.
Concurrent Event Routing subsequently closed them; the linked cumulative
report now records zero empty pairs:
`2026-07-20-feature-interaction-coverage-matrix.{json,md}`.

## Repeated baselines

All process rows used the catalog CPU/executor contract and external verifier.
The Able figures pool two independent five-process cohorts so workstation
variation cannot decide the result. Reference figures are arithmetic means of
five independent processes.

| Mode | Processes | Able mean | Reference | Reference mean | Ratio |
| --- | ---: | ---: | --- | ---: | ---: |
| compiled | 10 | 1.3010 s | Go | 0.004285 s | 303.62x |
| bytecode | 10 | 0.4000 s | Ruby | 0.050636 s | 7.90x |
| bytecode | 10 | 0.4000 s | Python | 0.034637 s | 11.55x |

Every one of the 20 Able executions and 15 reference executions verified. The
first and independent Able cohort means were 1.270/1.332 seconds compiled and
0.426/0.374 seconds bytecode. Their variation changes no classification.

Evidence:

- `2026-07-20-dependency-wave-baseline.{json,md}`
- `2026-07-20-dependency-wave-independent.{json,md}`
- `2026-07-20-dependency-wave-go-reference.{json,md}`
- `2026-07-20-dependency-wave-interpreter-reference.{json,md}`

These remain targeted pre-promotion rows and do not alter the reviewed
scorecard or performance frontier.

## Exact profiles and admission gate

Three verified generated-main CPU profiles were merged for compiled execution.
The warmed bytecode runtime benchmark kept typechecking enabled before the
timed region and profiled three measured `main` calls.

| Mode | Exact result | Admission decision |
| --- | --- | --- |
| compiled main | `bridge.currentGID` 95.48% cumulative; the descendants are `runtime.Stack`, traceback, and print-lock work | Same generic concurrency identity boundary whose fixed-context replacement improved concurrent applications but regressed unrelated N-Body by 54.7%; closed. |
| bytecode main | 328,966,168 ns/op, 23,282,352 B/op, 382,113 allocs/op; `execCallOpcode` 36.73%, `GoroutineExecutor.runTask` 29.59%, `execCallMember` 28.57%, `finishInlineReturn` 14.29%, cached member lookup 10.20%, atomic `Int32.Add` 9.18%, and `matchesType` 5.10% cumulative | Dispatcher/task parents are not candidates. Atomic bookkeeping was already reconciled across 13 applications with divergent Array, executor/channel, and GC ownership; member, return, and type-match families have completed broad gates. |

No concrete new child is both material here and independently material in two
unlike existing applications. The candidate-admission rule therefore rejects a
runtime implementation experiment in this tranche.

## Verification

- normal Able typecheck;
- bytecode, tree-walker, and compiled executions against the expected output;
- Go, Python, and Ruby executions against the external verifier;
- canonical/sibling Able byte identity;
- focused and corpus-wide catalog checks;
- feature coverage and interaction-matrix unit tests;
- ten verifier-backed Able processes per measured mode;
- five verifier-backed processes per reference runtime;
- three compiled main-only profiles and one three-call warmed bytecode profile;
- `git diff --check`.

## Next selection

Add one final pattern-heavy portable application that combines real program
entry/file input with destructuring, nominal states, a closure callback,
interface dispatch, Result recovery, and bounded worker concurrency.

Why: all nine remaining empty interactions are either lexical/binding/pattern
intersections or closures with program entry. One genuine record-routing or
configuration application can close that cluster without adding another
isolated microbenchmark. Completing the matrix gives future performance gates
a compact application cohort that exercises both individual features and their
important combinations.

This entails a checked-in deterministic input corpus, source-equivalent
Able/Go/Python/Ruby implementations, external verification, repeated arithmetic
means, and exact profiles. Admit code only if a new generic child repeats in at
least two unlike applications. After that application, stop expanding the
interaction cohort and refresh the independent pre-promotion scorecard/frontier
for all three interaction applications before selecting the next shared
compiler or VM wall. Do not begin WASM work.
