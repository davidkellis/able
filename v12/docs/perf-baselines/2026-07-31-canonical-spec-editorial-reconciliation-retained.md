# Canonical spec editorial reconciliation retained

Date: 2026-07-31

## Decision

**Retain the editorial reconciliation and rebase the reviewed `v12-spec`
performance scope.**

The changed passages now describe behavior that was already normative and
implemented:

- §11.2.1 identifies `Error` as one required core interface contract, seeded
  during bootstrap and canonically defined by `able.core.interfaces`;
- §12.2.3 states the existing `FutureError { details: String }` shape without
  calling its field TBD;
- §13.5 points named implementation selection to the function-position form
  already defined in §10.3.3; and
- the stale trailing todo lists are replaced by pointers to the forward-only
  `spec/TODO_v12.md` and `PLAN.md` authorities.

No language semantics, parser, checker, interpreter, bytecode VM, compiler,
runtime, stdlib, dependency, fixture, benchmark, frozen workspace, or WASM
behavior changed.

## Contract proof

The normative and executable surfaces agree:

- §10.3.3 defines `ImplName.method(receiver, ...)`;
- the canonical stdlib defines `Error` in `able.core.interfaces` and
  `FutureError { details: String }` in `able.concurrency.future`;
- named-implementation fixtures cover specificity, imported defaults, and
  inherent calls; and
- the concurrency fixture covers `FutureError` as an `Error` cause.

Focused parser, checker, tree-walker/bytecode parity, and strict fallback-free
compiled controls all pass. The three named-implementation compiled fixtures
run with `ABLE_COMPILER_REQUIRE_NO_FALLBACKS=1`.

## Evidence reconciliation

The canonical spec file SHA-256 moved from
`88cbd43eb7cc4cd9e22839bb6248b4fcccdcae94f35f12669ebc14220b9ebc71`
to
`e0a95d57e549a79c20db73640db2f16022ab4ab5627cf30a16eca90626c9fb35`.
Before review, the selector correctly invalidated all 23 closures for the sole
reason `scope-content-drift:v12-spec`.

The audit found no semantic or executable change, so only the shared reviewed
spec-scope snapshot was rebased. No benchmark measurement was relabeled and no
optimization candidate was admitted. The checked ledger now reports:

- 23 closures;
- 23 current;
- zero invalidated;
- no compiled selection; and
- no bytecode selection.

The manifest-driven architecture chain refreshed four source-pin evidence
records and five checked JSON reports. Every decision-bearing field and every
checked Markdown document remained identical. Attaching this durable record to
the cross-family closure then required a second topological source-pin refresh:
three evidence records and four checked JSON reports moved, again with no
decision or Markdown drift.

## Verification

The complete `./run_all_tests.sh` gate passed in 789.02 seconds at 1,279,232 KiB
peak RSS:

- 273 seeded exec fixtures, zero planned fixtures, and 274 fixture directories;
- 132 current scorecard rows and zero actionable frontier groups;
- parser corpus, tree-walker fixtures, bytecode fixtures, and cross-mode parity;
- all 844 compiler tests;
- canonical compiler outlier: 14.798 seconds;
- noisiest bounded compiler batch: 48.090 seconds; and
- checked performance-ledger and five-node architecture-chain gates.

All task builds used the disk-backed `/var/tmp` workspace.

## Next

Prepare a spec-first decision proposal for open-set `Error` exhaustiveness and
refutability.

Why: this is the unresolved language decision most directly adjacent to native
`Error` carriers, `Result` propagation, typed `match`, and rescue lowering.
The compiler cannot safely remove residual dynamic checks until the language
defines what statically exhaustive handling of an open interface means.

What it entails: inventory current checker and runtime behavior, compare the
existing `case _: Error` rule with wildcard and concrete-error cases, propose
one precise normative rule with compatibility consequences, and obtain the
language decision before changing execution code. If selected, implement it
generally across checker, tree-walker, bytecode, compiler, and fixtures, then
rerun the evidence selector.

Why it matters: a single cross-runtime exhaustiveness contract can enable sound
static control flow around native error carriers while avoiding benchmark,
container, or nominal-type special cases and unnecessary interpreter
boundaries.
