# Concurrent document-pipeline application gate — 2026-07-21

## Decision

Retain one portable file-driven application, its Go/Python/Ruby references,
public verifier, catalog contract, feature memberships, selected scorecard rows,
and checked frontier evidence. Retain no compiler, generated-runtime, bytecode
VM, canonical-stdlib, language, or WASM performance change.

The application closes the last two depth-one portable feature triples. Its
compiled profile reproduces the already-closed goroutine-identity owner, and
the bytecode row supplies no new exact leaf that meets the three-unlike-family
admission rule. The 87-row performance frontier therefore remains zero
actionable.

## Application contract

Concurrent Document Pipeline reads 32 document-status lines and replays them
for 32 rounds through four long-lived workers. Each worker captures a salt in a
scoring callback, classifies text into four buckets, runs a four-step transform,
and returns a nominal score. The collector emits schedule-independent counts,
total, and checksum:

```text
1024:1024:192,224,224,384:509294131:788118
```

Able, Go 1.26, Python 3.14, and Ruby 4.0 implement the same file input, worker
topology, callback, state machine, and aggregation. Every runtime output passes
the same Ruby verifier. The sibling Able Docker source uses a distinct package
name because the current loader sees canonical and sibling roots together; its
executable program body is otherwise identical to the canonical source.

The 32-round scale keeps tree-walker, bytecode, compiled, and reference
correctness processes below the one-minute project cap. The earlier 128-round
smoke was reduced before measurement after the tree-walker approached its
bounded timeout; no 128-round timing participates in the evidence.

## Feature-interaction result

The workload genuinely covers file/text/Array expressions, nominal task and
score structs, an inherent method, a captured callable, nested control flow,
typed/nil Channel matching, spawn/Future concurrency, stdlib protocols, and a
real entry argument.

The weighted triple frontier moves from minimum depth one to minimum depth two:

| Measure | Reconstructed baseline | Current |
| --- | ---: | ---: |
| three-family interactions | 165 | 165 |
| zero-depth triples | 0 | 0 |
| minimum depth | 1 | 2 |
| depth-one triples | 8 | 0 |
| improved triples | — | 133 |

The two intended gaps now each have two unlike applications:

- concurrency × expressions/files × closures/callables;
- concurrency × closures/callables × program entry.

## Repeated measurements

All successful samples are retained. Able has three independent five-process
cohorts per mode: the initial cohort, a matched volatility cohort, and the
mode-specific promotion cohort. References have two independent five-process
cohorts. Every one of 60 measured processes verified, with zero failures and
zero timeouts.

| Lane | Processes | Pooled mean | CV | Limiting ratio |
| --- | ---: | ---: | ---: | ---: |
| Able compiled | 15 | 0.254667 s | 14.91% | 65.34x Go |
| Go | 10 | 0.00389765 s | 13.97% | — |
| Able bytecode | 15 | 0.298000 s | 18.61% | 14.30x Python / 7.29x Ruby |
| Python | 10 | 0.0208439 s | 16.76% | — |
| Ruby | 10 | 0.0408937 s | 4.37% | — |

The checked scoreboard uses the normal mode-specific five-process promotion
reports: compiled 0.252 seconds versus Go 0.0037 seconds (68.108x), and
bytecode 0.348 seconds versus Python 0.0223 seconds (15.605x) and Ruby 0.0419
seconds (8.305x). The pooled values above are the better workstation estimate;
the different promotion snapshot does not change either miss classification.

## Ownership and admission

The compiled profile was admitted because the exact `bridge.currentGID` /
`runtime.Stack` mechanism already recurs in multiple unlike concurrency
applications. Three verified current main-only profiles merge to 460 ms of CPU
samples and place `bridge.currentGID` at 89.13% cumulative. That independently
reproduces the closed compiled-concurrency owner; the previously tested generic
fixed-context alternative improved concurrency but materially regressed an
unrelated N-Body guard, so it is not retried.

No bytecode profile was admitted. The new source combines the already-profiled
Channel/Future, call, member, return, and typed-match shapes, but timing alone
does not identify a new concrete VM child. Reprofiling aggregate dispatch or
allocation would reopen completed families without a new exact-leaf
hypothesis.

## Verification

- normal Able typecheck;
- tree-walker, bytecode, and compiled output parity;
- exact Go, Python, and Ruby output parity through one verifier;
- 15 verifier-backed Able processes per selected mode and ten per reference;
- three verified compiled main-only profiles;
- complete 47-application catalog and 126-program combined corpus checks;
- feature-coverage, pair-interaction, triple-frontier, selection, scoreboard,
  performance-frontier, and evidence-ledger checks;
- source files remain below 1,000 lines;
- `git diff --check`.

## Next recommendation

Select one different portable application from the new depth-two frontier that
combines expressions/files, closures/callables, and Option/Result handling.

Why: this is now the highest-weight shallow interaction, represented only by
Concurrent Event Routing and Policy Record Dispatch. The present tranche
confirmed that another concurrency-only variant lands on the closed
goroutine-identity wall; a non-routing validation/transformation application is
more likely to expose a broadly reusable error/callback boundary without
training on one benchmark.

What it entails: define a real source-equivalent validation or transformation
workload with Able, Go, Python, and Ruby implementations and one verifier; add
it only if every runtime stays within the bounded process guard; take repeated
compiled/bytecode/reference cohorts; and profile only an exact compiler or VM
leaf already present in two unlike existing applications. Update canonical
`able-stdlib` only if the workload reveals reusable specified library behavior.
Do not begin WASM work.
