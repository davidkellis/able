# Compiled native-carrier census reconciliation

Date: 2026-07-29

## Decision

Retain no compiler, generated-runtime, bridge, runtime, interpreter, bytecode,
canonical-stdlib, benchmark, reference, language, dependency, or WASM
performance change.

The proposed whole-corpus native-carrier and semantic-boundary census is
already complete. The July 26 strict 61-application exact-owner census built
every application without fallbacks, proved every final dependency graph
interpreter-free, and completed 4,163 verifier-backed processes. It retained
the general lazy common-`i32` bridge cache and closed every remaining repeated
owner as already rejected, semantic, or insufficiently broad.

Subsequent compiler work added or changed the post-spawn callable-context
surface. The July 29 scorecard and owner-closure refresh measured the affected
rows, reconciled the full 63-application compiled corpus, and advanced the
invalidation ledger. The current ledger reports 23 current closures, zero
invalidations, and no selected compiled or bytecode closure.

Rebuilding and reprofiling that unchanged corpus would violate the checked
invalidation contract. This tranche therefore performs the required
identity, corpus, and external-suite coverage checks, retains no code, and
records the stopping condition.

The machine-readable companion is
`2026-07-29-compiled-native-carrier-census-reconciliation.json`.

## Current identity

- `HEAD` and `origin/master`:
  `5f84e42c43d3d514c231c7887e835895a9ba557a`.
- July 26 strict-corpus census SHA-256:
  `ec80f2e55ce44adea4146247bca881784efb061ef60e656829ff4d2389d9bbae`.
- Current closure-ledger SHA-256:
  `2d0f331971c938a7c086d8be95aa255bc9c34dfb6d7943686ce7d7debdc423b2`.
- Current scoreboard SHA-256:
  `9adbc937a7048eb69f81386920c44b604e9e817b22ecaa183989f72906ff4f3a`.
- External catalog SHA-256:
  `13c7e432b2ed15b4aac26fdaeb7f1950c28b98feb494e6007f8d84ea29efd789`.
- v12 specification SHA-256:
  `4f0405b86c122993723e8617abd6f825d9a8ff858d4c72acaf4e33469452f080`.

There is no current worktree or committed change after the July 29 closure
refresh in the compiler, compiler bridge, v12 spec, or benchmark sources.
The bounded dependency refresh and handoff documentation do not change a
compiled lowering scope. The deferred WASM paths remain outside this tranche.

## Mechanical gates

The current checked evidence reports:

| Gate | Result |
| --- | --- |
| Selected scorecard rows | 126 |
| Compiled rows | 63 |
| Verified compiled rows | 63 |
| Distinct compiled Able sources | 63 |
| Compiled target meets / misses | 7 / 56 |
| Able/reference samples per selected row | 5 / 5 |
| Performance closures | 23 |
| Current / invalidated closures | 23 / 0 |
| Actionable frontier groups | 0 |

`just bench-evidence-ledger-check`, `just bench-scoreboard-check`, and
`just bench-frontier-check` all pass. The ledger's mode-aware production
scope check is the authority for whether a prior census or profile may be
reopened.

## External-suite coverage

The `coverage` catalog contains 63 applications and maps to 63 distinct
external benchmark directories. The external repository contains 64
non-metadata benchmark directories. Its sole directory outside `coverage` is
historical `sudoku`.

That row is already exposed as the diagnostic-only `legacy-sudoku` suite. It
uses a different scan-based workload than the exact-cover Go/Python
references and cannot complete within the bounded scorecard protocol. Adding
it to the selected corpus would manufacture a non-comparable timeout rather
than broaden portable application evidence.

No new external application, reusable stdlib API, specification change,
benchmark source, scorecard contradiction, or compiler/runtime scope change
provides an invalidation trigger.

## Admission result

No owner reaches candidate admission:

- existing generated `runtime.Value`, bridge, boxing, and semantic-boundary
  families were already classified by the strict 61-row census;
- post-spawn compiler scope was explicitly refreshed into the current
  63-row scorecard and 23-entry ledger;
- no closure is invalidated;
- no new external portable application exists outside the selected corpus;
  and
- `spec/TODO_v12.md` records no compiler AOT or stdlib externalization gap.

No generated module, CPU profile, allocation profile, compiler prototype, or
A/B/Go cohort was produced. This is required restraint, not missing
verification: the authoritative selector says that unchanged evidence must
not be rerun.

## Next

Refresh the exact retained/deferred release inventory for this reconciliation
and the post-publication handoff.

Why: the no-code decision and current roadmap are now new retained evidence,
while the 34 deferred WASM files must remain outside any later consolidation.

What it entails: capture a fully expanded worktree snapshot, classify the
three handoff files plus this record and its JSON companion as retained,
prove their exact complement is the unchanged deferred WASM boundary, and
authorize no staging or commit during the inventory itself.

Why it matters: it preserves the evidence-backed stopping condition without
mixing deferred work into history. After that boundary is recorded, further
performance work requires a real invalidation trigger: a specified
language/stdlib change, a new portable external application, a relevant
compiler/runtime change, or contradictory current scorecard evidence.
