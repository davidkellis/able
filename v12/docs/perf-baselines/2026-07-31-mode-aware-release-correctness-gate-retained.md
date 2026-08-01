# Mode-aware release and correctness gate retained

Date: 2026-07-31

## Decision

Retain the general typechecker correction and the current observer
reconciliation. Select no compiler, runtime, or bytecode performance change.

The mode-aware selector emits zero bytes after verification. All 23
performance closures are current, the 132-row frontier has zero actionable
groups, and the five-node architecture evidence chain is current. No new
profile or A/B cohort is authorized by this state.

## Release-observer reconciliation

The ordinary release gate initially exposed checked observers that still
named historical reports or snapshot totals. They were advanced to current
evidence without changing their decisions:

- the compiled static-boundary census now checks the 66/66 current-default
  report, all 66 resolved dependency graphs, zero interpreter links, and the
  five broad identities already classified as output ABI or named `HashMap`
  runtime service;
- residual selected excess is 80.253579 of 274.677263 total excess, or
  29.217409%;
- the bytecode semantic-region source fingerprint is current and its
  no-prototype decision is unchanged;
- the native-hot-tier proxy reports 42 meets, 24 misses, and 12.570210525
  aggregate excess, retaining its no-prototype decision;
- the cross-engine budget reports 13 meets, 119 misses, and 274.677263 total
  excess, with no local mechanism selected; and
- the sustained workload audit still reports 24 broad multi-feature rows, 11
  sustained-native rows, one row in their intersection, and no material depth
  gap.

Four architecture evidence inputs and five checked reports were refreshed
topologically. Decision-bearing content and generated Markdown remained
unchanged.

## Reproduced correctness gap

The refreshed canonical stdlib gate failed in both execution modes before
running tests. Repeated checking of inferred generic calls converted canonical
checker types into display-oriented AST labels, then treated those synthesized
labels as explicit source arguments. That round trip lost:

- the declaration identity of imported interfaces and unions;
- the canonical target behind renamed aliases; and
- the application shape of imported generic constructors, producing forms
  such as `Array Unknown String`.

Strict callable invariance exposed a second ordering defect: callable
comparison ran before the ordinary top-level nullable lift, so a value of type
`T -> ?String` could not initialize `?(T -> ?String)`.

## Retained correction

The checker now records the exact canonical `Type` associated with every
inferred type-argument label it synthesizes. Repeated and overload-mediated
checks recover that type directly instead of reparsing a diagnostic label.
Inferred type-parameter rebinding compares canonical invariant types, including
alias expansion.

Callable equivalence is handled in target-type dispatch. Complete callable
signatures remain invariant, while the existing top-level value conversion
still permits an equivalent callable to lift into a nullable callable.

Regression coverage exercises repeated inference for interface, union,
generic, nullable, and alias types, plus positive and negative nullable
callable cases. This is a general typechecker rule; it contains no application,
benchmark, container, or non-primitive nominal special case.

No canonical stdlib file was changed. The sibling `able-stdlib` remained at
HEAD `219eff222c28406487231713753641bc49ee5b9a`; its pre-existing dirty
worktree was preserved.

## Verification

- `go test ./pkg/typechecker`: pass.
- Focused canonical `able.spec.assertions` check: pass.
- Complete canonical stdlib tree-walker suite: pass in 20 seconds.
- Complete canonical stdlib bytecode suite: pass in 15 seconds.
- `./v12/run_all_tests.sh`: pass.
  - 280 seeded exec fixtures, zero planned, and 281 fixture directories;
  - all eight tree-walker, parity, and bytecode fixture shards;
  - all 86 bounded compiler batches;
  - canonical compiler outlier: 16.657 seconds;
  - slowest compiler batch: 53.568 seconds, below the 55-second cap.
- Performance ledger: 23 current closures, zero invalidated.
- Performance frontier: 132 rows, zero actionable groups.
- Architecture evidence chain: five current nodes.
- External scoreboard: current.

All builds and test temporaries used the task workspace under disk-backed
`/var/tmp`, except the release runner's bounded 36,968 KiB extern-Go cache,
which was identified for task cleanup. The 3,273,884 KiB task workspace and
that extern cache are removed after publication.

## Next

Reconcile the block-comment contradiction in the canonical specification.

Why: §2 says block-comment syntax is TBD, while §6.9 already states that block
comments are unsupported and the parser contract implements that rule.

What it entails: prove current parser behavior, edit only the stale §2 wording,
remove the resolved TODO entry, run parser and release controls, reconcile the
reviewed spec scope and architecture chain, and rerun the selector.

Why it matters: a single lexical authority prevents false parser work and
false performance invalidations. Native-lowering work should reopen only when
current broad evidence selects a legal shared owner.

Do not begin WASM work.
