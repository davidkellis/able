# Operation-level benchmark depth contract — 2026-07-20

## Decision

Retain an operation-level extension of the feature-coverage contract and make
it part of the normal catalog check. Add no benchmark, profile, compiler, VM,
runtime, canonical-stdlib, or language change in this tranche.

The family-level contract was already complete, so this tranche does not
repeat it. The new manifest records literal Able source evidence and distinct
workload classes for performance-relevant semantic operations, validates
focused fixtures for mixed/local-only boundaries, and joins portable
operations to the current performance-frontier ownership groups.

## Result

`bench-operation-depth.json` contains 21 operations spanning every one of the
15 feature families:

- 14 have at least three unlike portable workload classes;
- four have insufficient unlike-workload breadth;
- three are intentionally local-only;
- all 18 portable or mixed operations link to current frontier evidence; and
- zero operations link sufficient breadth to an actionable frontier group.

The breadth threshold is three unlike workload classes, not merely three
benchmark names. Each portable claim names literal source markers checked
against the catalog's canonical Able target. This keeps three related API
front ends from being counted as three independent programs.

## Insufficient portable depth

| Operation | Applications | Unlike breadth | Consequence |
| --- | --- | ---: | --- |
| Wide numeric nominal methods | Fixed Width 128, Rational Series | 2 | Raw wide-integer extraction remains below the generality bar; nominal-specific compiler lowering remains prohibited. |
| Mutex locking and `ensure` | Mutex Ledger, Mutex Await Journal | 2 | A third application would still need an independently shaped synchronization contract; cancellation/policy remains local. |
| Hash-map lookup and update | K-Nucleotide, Word Frequency | 2 | The current UTF-8/map owner lacks one unlike discriminator between generic map work and key conversion. |
| Regex NFA matching | Suffix, Set, Stream audits | 1 | Three APIs share one related regex-audit workload and do not establish unlike-application breadth. |

The three local-only operations are dynamic package evaluation, user-authored
host interop, and test-reporter lifecycle. They retain focused fixtures and no
portable timing claim.

## Sufficient but closed operations

Direct recursion, primitive Array access, primitive float arithmetic, nominal
construction, higher-order callbacks, pattern/union matching, inherent method
calls, interface-typed protocols, Option/Result flow, spawn/Future lifecycle,
Channel traffic, iterator protocols, static package initialization, and real
entry/file I/O all have sufficient unlike-workload coverage.

That does not make them profile candidates. Their linked frontier groups are
target guards or are closed by no shared exact leaf, a previously rejected
generic candidate, related-algorithm evidence, or exhausted retained work. The
operation map therefore finds no missed current candidate and authorizes no
unchanged-corpus profile rerun.

## Implementation

- `bench_operation_depth_check` validates schema identity, feature-family
  closure, catalog membership, literal source evidence, workload classes,
  breadth classification, local fixtures, frontier-group identity, and that a
  linked group contains at least one claimed application.
- `bench_operation_depth_check_test.py` guards stale source markers, inflated
  unlike breadth, unknown applications and frontier groups, invalid local-only
  claims, unknown feature families, and missing family coverage.
- `bench_catalog_check` and `just bench-catalog-check` now run the new checker
  and its failure-mode tests alongside the existing family contract.

## Verification

- `v12/bench_operation_depth_check`: 21 operations, 14 sufficient, four
  insufficient, three local-only, 18 frontier-linked, zero actionable.
- `python3 v12/bench_operation_depth_check_test.py`: eight tests pass.
- `just bench-catalog-check`: 36 portable applications, 37 canonical sources,
  one diagnostic source, 79 local fixtures, 115 combined programs; both
  coverage contracts and all 13 contract tests pass.

## Next recommendation

Add one honest, verifier-backed map application with a workload unrelated to
word counting and nucleotide counting, using non-text keys and ordinary
lookup/update behavior. A deterministic inventory or event reconciliation
workload is a good shape: index records by numeric identifier, apply updates,
join a second stream, and emit a checksum. Implement the same algorithm and
operation counts in Able, Go, Python, and Ruby.

Why this operation: it is the only insufficient-depth item that is both a
widely used general-programming path and already one application short of the
three-unlike breadth rule. A numeric-key discriminator can show whether the
current shared cost is generic map protocol/storage work or merely UTF-8/key
conversion. Wide numerics risk nominal-specific work, regex remains a related
algorithm family, and the known shared compiled concurrency remedy has already
failed unrelated guards.

The next tranche should define the verifier and reference implementations,
register the application in the portable catalog and operation-depth evidence,
run tree-walker/bytecode/compiled correctness and the sub-minute corpus checks,
then collect profiles across K-Nucleotide, Word Frequency, and the new
application. Optimize only if the same concrete non-nominal leaf is material
in all three; any candidate must then pass two fresh five-run broad cohorts.
Do not begin WASM work.
