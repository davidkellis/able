# Concurrent event-routing feature-interaction gate — 2026-07-20

## Decision

Retain one portable file-driven application, its deterministic corpus, and its
catalog coverage. Retain no compiler, generated-runtime, bytecode-VM,
canonical-stdlib, language, nominal-lowering, or WASM performance change.

This completes the planned feature-interaction application cohort. All 55
pairwise intersections among the 11 discriminating portable/mixed feature
families now have application coverage. The exact profiles reproduce closed
compiler/runtime and VM families, so no implementation candidate passes the
broad admission rule.

## Application contract

Concurrent Event Routing reads 32 checked-in records and replays them for 128
rounds, producing 4,096 tasks for four long-lived workers. Workers split and
explicitly destructure each record, validate it as `Result EventRecord`, score
accepted records through a captured callback and five-round state machine,
recover rejected records, and return an `AcceptedRoute | RejectedRoute`
nominal state. The collector aggregates route counts through the public `Map`
interface with schedule-independent totals and checksums.

The Able source genuinely combines real program entry and file input, array
and nominal-struct destructuring, wildcard/renamed fields, nominal structs and
a payload union, inherent methods, interface dispatch, closures/callables,
Result/Error handling, nested control flow, text and Array operations, HashMap,
Channel/Future concurrency, and canonical stdlib protocols.

Able, Go 1.26, Python 3.14, and Ruby 4.0 implement the same input, parsing,
validation, routing, four-worker schedule, and aggregation. Expected output:

```text
4096:4096:3072:1024:768,768,768,768:1538066199:499321
```

The canonical and sibling Able sources are byte-identical. The sibling suite
contains the corpus, all reference sources, Docker contracts, verifier, and
README.

## Interaction coverage

Concurrent Event Routing occupies all 11 discriminating feature families. It
closes the nine remaining empty intersections: eight between the selective
lexical/binding/pattern family and concurrency, text/files, closures, nominal
types, methods, interfaces, Result handling, and stdlib protocols, plus
closures with real program entry.

Across the complete four-application interaction tranche:

| Measure | Before tranche | Current |
| --- | ---: | ---: |
| portable/mixed families | 11 | 11 |
| pairwise interactions | 55 | 55 |
| zero-coverage pairs | 29 | 0 |
| pairs improved | — | 55 |

The cumulative report is
`2026-07-20-feature-interaction-coverage-matrix.{json,md}`. Package loading is
excluded because every portable application has it and it cannot discriminate
workloads.

## Repeated baselines

All process rows used the catalog CPU/executor contract and public verifier.
Able means pool two independent five-process cohorts; reference means contain
five independent processes. No single workstation run determines a result.

| Mode | Processes | Able mean | Reference | Reference mean | Ratio |
| --- | ---: | ---: | --- | ---: | ---: |
| compiled | 10 | 2.7610 s | Go | 0.004523 s | 610.45x |
| bytecode | 10 | 2.7370 s | Ruby | 0.050209 s | 54.51x |
| bytecode | 10 | 2.7370 s | Python | 0.030993 s | 88.31x |

All 20 Able and 15 reference executions verified. The first and independent
Able cohort means were 2.878/2.644 seconds compiled and 2.746/2.728 seconds
bytecode. The spread changes no classification.

Evidence:

- `2026-07-20-concurrent-event-routing-baseline.{json,md}`
- `2026-07-20-concurrent-event-routing-independent.{json,md}`
- `2026-07-20-concurrent-event-routing-go-reference.{json,md}`
- `2026-07-20-concurrent-event-routing-interpreter-reference.{json,md}`

These remain pre-promotion measurements and do not yet alter the reviewed
scorecard or frontier.

## Exact profiles and admission result

Three verified compiled generated-main CPU profiles were merged. The warmed
bytecode runtime profile kept typechecking enabled before the timed region and
measured three complete `main` calls.

| Mode | Exact profile | Decision |
| --- | --- | --- |
| compiled main | `bridge.currentGID` 96.26% cumulative; `runtime.Stack.func1` 95.95% | Same generic goroutine-identity boundary whose fixed-context candidate regressed unrelated N-Body by 54.7%; closed. |
| bytecode main | 3,199,981,203 ns/op, 287,943,264 B/op, 2,829,868 allocs/op; `execCallOpcode` 47.20%, `GoroutineExecutor.runTask` 82.37%, `execCallMember` 21.33%, `execCallName` 18.16%, cached member lookup 14.68%, inline return 8.34%, scheduler atomic add 6.97%, typed-pattern jump 6.76%, and type matching 4.33% cumulative | Aggregate dispatcher/task frames are not candidates. Member/cache, call/return, scheduler atomic, and type-match families have completed broad gates. The generic typed-pattern metadata trial was neutral or regressive in two controls; the retained IteratorEnd shortcut already exhausts its one broadly repeated semantic category. |

The typed-pattern descendant is material here and in unlike historical text,
iterator, and byte-histogram workloads, but it is not a new candidate: the
general metadata shortcut was already tested and reverted after failing its
broad workload gate. No other concrete child is both new and independently
material in two unlike applications. No implementation experiment advanced.

## Verification

- normal typecheck;
- bytecode, tree-walker, and compiled application execution;
- exact Go, Python, and Ruby output parity;
- canonical/sibling Able byte identity;
- focused and corpus-wide catalog checks;
- feature-coverage and interaction-matrix unit tests;
- ten verifier-backed Able executions per measured mode;
- five verifier-backed executions per reference runtime;
- three verified compiled main-only profiles;
- one three-call warmed bytecode profile;
- focused Channel/Future and generic-union guards;
- source-file size checks and `git diff --check`.

## Next selection

Stop adding feature-interaction applications. Promote and reconcile Concurrent
Text Index, Validated Job Pipeline, Dependency Wave Validation, and Concurrent
Event Routing through the normal independent scorecard/frontier contract.

Why: the interaction matrix is complete, while these four real applications
show substantial compiler and bytecode product gaps. Promotion will make their
cost and evidence participate in the same reviewed selection ledger as the
rest of the suite instead of remaining isolated targeted reports.

This entails refreshing any missing second cohorts and reference evidence,
adding the four compiled/bytecode rows to the reviewed selection manifest,
rebuilding the aggregate scorecard and ownership frontier, and selecting the
largest still-unowned generic descendant shared by unlike applications. If the
frontier remains zero-actionable because all repeated owners are closed, move
directly to the next architecture-scale compiler or VM family rather than
reopening goroutine identity, typed-pattern metadata, member-cache, call/frame,
or return/type-match micro-variants. Do not begin WASM work.
