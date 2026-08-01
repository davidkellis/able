# Architecture evidence chain automation retained

Date: 2026-07-31

## Decision

Retain the manifest-driven architecture-evidence check and refresh tooling.
The five checked reports now form one explicit topological chain. Refresh is
transactional and may update only source fingerprints; decision-bearing JSON
or checked Markdown drift fails closed and restores every file changed earlier
in the transaction.

This is release tooling only. No compiler, generated runtime, runtime,
tree-walker, bytecode VM, stdlib, language, dependency, benchmark, fixture, or
WASM production path changed.

## Retained chain

The manifest `v12/bench-architecture-evidence-chain.json` declares:

1. bytecode native hot-tier budget;
2. cross-engine structural strategy;
3. portable VM backend ADR;
4. shared-runtime semantic ABI;
5. shared-runtime closed-region cutover.

The declared dependencies exactly match references from each node's evidence
to another node's checked JSON. The loader rejects duplicates, forward or
missing dependencies, and declared/evidence dependency disagreement.

## Refresh contract

`v12/bench_architecture_evidence_chain --check` is read-only. It verifies every
pinned source fingerprint and invokes every generator's checked-artifact gate
in topological order.

`v12/bench_architecture_evidence_chain --refresh`:

- refreshes pinned evidence source hashes from their current files;
- regenerates candidate JSON and Markdown under disk-backed `/var/tmp`;
- compares JSON after masking only `sha256` leaves in records that also contain
  a source `path`;
- requires every remaining JSON field and the complete Markdown artifact to
  remain identical;
- atomically writes safe evidence and checked-JSON updates;
- reruns the complete read-only chain check; and
- restores the exact original bytes of every file written earlier in the
  transaction if any node fails.

A refresh against the current repository was a clean no-op: five nodes, zero
evidence updates, and zero checked-report updates.

The normal `run_all_tests.sh` architecture gate and the aggregate `just`
architecture recipe now use this single chain check. Dedicated per-node
commands remain available for focused diagnosis.

## Verification

- Five synthetic transaction tests pass in about 1.2 seconds:
  - current check performs no writes;
  - a two-node source change propagates in topological order;
  - manifest dependency mismatch is rejected;
  - downstream decision drift rolls back upstream updates; and
  - downstream Markdown drift rolls back upstream updates.
- All five existing architecture-report contract suites pass.
- The focused chain recipe and checked five-node production chain pass.
- The complete normal `./run_all_tests.sh` gate passes:
  - external scoreboard: 132 rows, five successful Able/reference samples per
    row;
  - performance frontier: zero actionable groups;
  - performance ledger: 23 current closures, zero invalidations;
  - parser, interpreter, fixture, parity, compiler, CLI, kernel, cleanup, and
    evidence gates all pass;
  - all 844 compiler tests pass;
  - the isolated canonical compiler test completes in 14.690 seconds; and
  - the noisiest compiler batch completes in 48.770 seconds.
- The new tool is 355 lines and its test is 276 lines.
- Removed the exact 37,906,702 bytes of idle generated extern cache and stale
  audit scratch files found by the final cleanup audit. No `/tmp/able-*`,
  `/var/tmp/able-*`, or new Python cache remains.

## Performance admission result

The post-tooling selector remains empty in both modes:

- compiled: no selected owner;
- bytecode: no selected owner.

The tooling change therefore admits no profile cohort, A/B experiment, or
production optimization. It preserves the current fallback-free compiled
architecture without inventing work from metadata churn.

## Next

Do not begin another performance or profiling tranche until an intentional
compiler, runtime, canonical-stdlib, spec, benchmark, or scorecard change
genuinely invalidates a checked closure.

When such a trigger occurs, run the architecture chain check or refresh, rerun
the mode-aware selector, and refresh only the exact closures and modes it
names. That work entails repeated verifier-backed measurements and equivalent
reference runs only after one general owner repeats across at least three
unlike applications.

This is important because the immediate objective remains native Go lowering
with minimal compiled/interpreter crossings, but the current evidence contains
no unresolved shared owner. Waiting for a real invalidation protects that
objective from benchmark-specific, named-container, cold-boundary, or
already-native experiments.
