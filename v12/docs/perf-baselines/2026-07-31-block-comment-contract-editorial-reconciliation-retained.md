# Block-comment contract editorial reconciliation retained

Date: 2026-07-31

## Decision

**Retain the existing line-comment-only language contract and rebase the
reviewed `v12-spec` performance scope.**

Able line comments begin with `##` and continue through the end of the line.
Block comments are not supported. Section 2 now states that same rule instead
of describing block-comment syntax as TBD, and the resolved marker has been
removed from `spec/TODO_v12.md`.

This is an editorial authority correction. It changes no language semantics,
parser, AST, checker, interpreter, bytecode VM, compiler, runtime, canonical
stdlib, dependency, fixture, benchmark, frozen workspace, or WASM behavior.

## Contract proof

The normative and executable surfaces agree:

- §2 and §6.9 both state the line-comment-only rule;
- the tree-sitter grammar recognizes only `##` followed by non-newline text;
- the existing focused Go parser controls cover comments at module and
  composite-list, struct-literal, and struct-pattern positions;
- a comment-free control module passes `able check`; and
- an otherwise identical module containing `/* ... */` is rejected by the
  parser.

No parser or AST source change was needed.

## Evidence reconciliation

The canonical spec file SHA-256 moved from
`709468672e761782a0aba6b69e9576e421f707d9315452d5790ddabef71097f7`
to
`ddb493eab0ea6ee2f06844ad2e0b6d0d80dc060c19c239122fa82c9b98890a9f`.
Its checked scope tree moved from
`e49c125c826dfafd1a9e52f550f1bcc46d5bd3854d0331428ca050de6a103ee1`
to
`987737c2aa5493bee012f7dfd7125ba5ad0bdd3ee19bbedaa9819e057c7881cd`.

Before review, the mode-aware selector correctly invalidated all 23 closures
for the sole reason `scope-content-drift:v12-spec`. Every closure definition,
row identity, evidence set, benchmark source, mode, and disposition remained
unchanged. Only the shared reviewed spec snapshot was rebased; no measurement
was relabeled and no performance candidate was admitted.

The five-node architecture chain refreshed four source-pin evidence records
and five checked JSON reports. Every decision-bearing field and every checked
Markdown document remained identical. Attaching this durable record to the
shared cross-family closure then required a second topological refresh of
three evidence records and four checked JSON reports, again without decision
or Markdown drift.

The final selector reports 23 current closures, zero invalidations, and no
compiled or bytecode selection. The 132-row frontier has zero actionable
groups.

## Verification

Focused comment parser controls passed. The complete `./run_all_tests.sh` gate
then passed in 686.23 seconds at 1,287,204 KiB peak RSS:

- 280 seeded exec fixtures, zero planned fixtures, and 281 fixture
  directories;
- parser corpus, tree-walker fixtures, bytecode fixtures, and cross-mode
  parity;
- all 86 bounded compiler batches;
- slowest compiler batch: 49.427 seconds;
- canonical compiler outlier: 14.300 seconds; and
- checked performance-ledger and five-node architecture-chain gates.

An initial run used an intentionally tighter 55-second cumulative compiler
batch cap and timed out in batch 50. The named test had been executing for
only three seconds when the package-level alarm fired. Three isolated
fresh-generated-cache processes passed in 22.617, 22.138, and 23.237 seconds,
a 22.664-second arithmetic mean. The complete clean rerun passed batch 50 in
49.427 seconds under the required one-minute cap. No individual test exceeded
one minute.

The external canonical stdlib also passed in tree-walker mode in 20 seconds
and bytecode mode in 15 seconds. All build and test work used the disk-backed
`/var/tmp` task workspace. After publication, the exact 2,845,840 KiB
workspace was removed; no Able task directory remains under `/tmp` or
`/var/tmp`.

## Next

Canonicalize the already-selected invariant-only v12 variance boundary by
replacing the remaining variance-declaration-syntax TBD with an explicit v12
deferral and reconciling the stale open-work marker.

Why: the retained variance decision and checker already enforce invariance,
but §4.1.7 and the forward indexes still make the syntax look undecided. That
authority mismatch can invite an implicit covariance feature that Go carriers
cannot preserve without copying, boxing, or changing identity.

What it entails: prove the current parser/AST has no variance declaration,
reconcile §4.1.7, `spec/TODO_v12.md`, and the active typechecker handoff with
the retained option-A decision, run the invariant Array/nominal/callable
controls and ordinary release gates, then rebase checked spec evidence without
changing execution code.

Why it matters: an explicit invariant-only v12 contract protects exact native
Go instantiations and prevents recursive nominal/container conversions from
reintroducing `runtime.Value` adapters or compiled/interpreted boundaries.
