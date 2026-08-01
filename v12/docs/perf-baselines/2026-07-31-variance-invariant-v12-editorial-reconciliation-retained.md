# Invariant-only v12 variance reconciliation retained

Date: 2026-07-31

## Decision

**Retain the already-implemented invariant-only v12 contract and rebase the
reviewed `v12-spec` performance scope.**

All v12 type parameters are invariant. V12 provides no syntax for declaring
covariance or contravariance, so no type can opt out. A valid top-level
conversion from `A` to `B` does not imply a conversion from `F A` to `F B`.
Function values require equivalent complete signatures; v12 defines neither
parameter contravariance nor result covariance.

Section 4.1.7 now states the retained option-A decision rather than describing
variance-declaration syntax as TBD. The resolved item has been removed from
`spec/TODO_v12.md`, and the active typechecker handoff now identifies the
checker and fail-closed compiler implementation as complete. The retained
variance design note now presents option A as the selected canonical rule
rather than as a pending request.

This is an editorial authority reconciliation. It changes no language
semantics, parser, AST, checker, interpreter, bytecode VM, compiler, runtime,
canonical stdlib, dependency, fixture, benchmark, frozen workspace, or WASM
behavior.

## Grammar, AST, and execution proof

The parser grammar accepts a generic parameter identifier, optional bounds,
and an optional default. The canonical `ast.GenericParameter` contains only
name, constraints, and inference state. Neither surface has a variance field
or declaration token.

An ordinary `Box<T>` control module passes `able check`. Otherwise identical
modules using `out T`, `in T`, `+T`, or `-T` are parser syntax errors. A scan
of all 834 active Able sources under fixtures, examples, canonical stdlib
source, and canonical stdlib tests found no variance declaration.

Focused controls passed for:

- ordinary generic struct, union, interface, alias, and composite parsing;
- Array, Map, Range, Iterator, Future, nullable, and user-nominal invariance;
- valid top-level integer widening and interface upcasts;
- complete callable parameter, result, arity, generic, and alias equivalence;
- coded invariant and callable-signature diagnostics;
- tree-walker and bytecode diagnostic/reconstruction fixtures; and
- strict no-fallback compiled diagnostics and explicit reconstruction.

No execution source change was needed.

## Evidence reconciliation

The canonical spec SHA-256 moved from
`ddb493eab0ea6ee2f06844ad2e0b6d0d80dc060c19c239122fa82c9b98890a9f`
to
`0965a0a48b49f5eaed9392d75b7dbf6e74965a04c2547286d39397dc0812bdcd`.
Its checked scope tree moved from
`987737c2aa5493bee012f7dfd7125ba5ad0bdd3ee19bbedaa9819e057c7881cd`
to
`5f3a8573e88d3ec612a05a429948fcb1bb63981cd7dd7f6983a83bcc71cf1095`.

Before review, all 23 performance closures were invalidated solely by
`scope-content-drift:v12-spec`. Structural comparison proved that every
closure snapshot, selection identity, input identity, benchmark source, and
non-spec production scope was unchanged. Only the reviewed spec scope was
rebased; no benchmark measurement was relabeled and no candidate was admitted.

The five-node architecture chain refreshed four source-pin evidence records
and five checked JSON reports without decision or Markdown drift. Attaching
this durable record to the shared cross-family closure required a second
topological refresh of three evidence records and four checked JSON reports,
again without decision or Markdown drift.

The final selector reports 23 current closures, zero invalidations, and no
compiled or bytecode selection. The 132-row frontier has zero actionable
groups.

## Verification and cleanup

The complete `./run_all_tests.sh` gate passed in 781.36 seconds at 1,342,220
KiB peak RSS:

- 280 seeded exec fixtures, zero planned fixtures, and 281 fixture
  directories;
- parser corpus, tree-walker fixtures, bytecode fixtures, and cross-mode
  parity;
- all 86 bounded compiler batches;
- slowest compiler batch: 47.109 seconds;
- canonical compiler outlier: 14.753 seconds; and
- checked performance-ledger and five-node architecture-chain gates.

The external canonical stdlib passed in tree-walker mode in 20 seconds and
bytecode mode in 15 seconds. All build and test work used a disk-backed
`/var/tmp` workspace. The exact 2,486,660 KiB workspace was removed after
verification; no Able task directory remains under `/tmp` or `/var/tmp`.

## Next

Reconcile the stale integer-literal configurability marker in §6.1.1.

Why: the primitive table, detailed literal-typing rules, parser/checker
behavior, and compiled carriers all use context-sensitive literals with a
fixed unconstrained default of `i32`, while one earlier sentence still calls
that default configurable/TBC.

What it entails: prove the current unsuffixed, suffixed, contextual-fit, and
overflow behavior across parser, checker, interpreter, and strict compiler;
replace only the stale configurability marker, remove the resolved TODO, run
the ordinary release gates, and rebase checked spec evidence.

Why it matters: one fixed v12 default prevents configuration-dependent
primitive types, cross-module ABI differences, and avoidable carrier
conversions while preserving explicit suffixes and contextual literal fitting.
