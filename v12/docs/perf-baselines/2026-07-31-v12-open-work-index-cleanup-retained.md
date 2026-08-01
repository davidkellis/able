# v12 open-work index cleanup retained

Date: 2026-07-31

## Decision

Retain the forward-only rewrite of `spec/TODO_v12.md`.

The file claimed completed work should be removed but had grown to 607 lines,
mostly chronological regex and compiler-native performance history. That
history already exists in `LOG.md`, `v12/LOG.md`, dated performance records,
and the active/historical compiler design notes.

The rewrite is 102 lines: 96 added and 601 removed. It changes no canonical
language rule, AST, parser, checker, interpreter, compiler, runtime, stdlib,
dependency, fixture, benchmark, or WASM path.

## Retained forward state

The index now contains only:

- an explicit empty selected-implementation queue;
- genuine language-design questions that require a maintainer decision;
- stale or conflicting canonical-spec markers that need editorial
  reconciliation rather than runtime implementation;
- concise closed-surface status; and
- links to the active roadmap and implementation authorities.

The audit caught an important false lead: §13.5 still says “Named Impl
Invocation TBD”, but §10.3.3 already specifies
`ImplName.method(receiver, ...)`. The parser, checker, tree-walker, bytecode,
compiler, import surface, and fixtures already implement that contract. The
TODO therefore classifies this as canonical-spec editorial reconciliation, not
a missing language feature.

The same reconciliation class covers the stale `FutureError`, `Error`, range,
stdlib, tooling, and trailing `Todo` wording where the canonical spec or active
implementation already provides a contract.

## Verification

- Every retained authority link exists.
- `git diff --check -- spec/TODO_v12.md` passes.
- The five-node architecture evidence chain remains current.
- The performance ledger remains 23 current closures and zero invalidations.
- The 132-row frontier remains at zero actionable groups.
- The normal `./run_all_tests.sh` gate passes all 844 compiler tests and every
  parser, interpreter, fixture, parity, CLI, evidence, kernel, and cleanup
  lane. The isolated canonical compiler test takes 14.435 seconds; the
  noisiest compiler batch takes 36.566 seconds.
- Removed the exact 37,781,644 bytes of generated extern cache and Python
  bytecode created by verification.
- No `/tmp/able-*`, `/var/tmp/able-*`, or new Python cache remains.

## Next

Reconcile the stale canonical-spec markers in one editorial-only tranche,
starting with the named-implementation cross-reference and the already
canonical `FutureError` shape.

Why: these markers contradict nearby normative text and implemented
cross-runtime behavior, so they can misdirect future agents into rebuilding
completed features.

What it entails: prove each correction against the existing spec sections,
stdlib definitions, and cross-runtime fixtures; edit only contradictory
wording; run the full correctness gate; refresh the architecture chain; and
rerun the mode-aware selector because `spec/full_spec_v12.md` is a checked
performance input. Do not change runtime code unless the audit exposes a real
semantic mismatch.

Why it matters: a consistent language authority protects native Go lowering
and interpreter-free compiled applications from false invalidations while
making genuine unresolved design choices visible.
