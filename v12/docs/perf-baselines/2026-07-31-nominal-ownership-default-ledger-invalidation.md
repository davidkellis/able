# Performance evidence invalidation ledger

## Decision

**invalidated-closures-selected**.

Refresh only the selected closures before considering a candidate.

The ledger contains 23 closures: 11 current and 12 invalidated.

## Closure selector

| Closure | Modes | Origin | Disposition | Status | Reasons |
| --- | --- | --- | --- | --- | --- |
| `bytecode-byte-output` | bytecode | frontier | `closed-no-shared-leaf` | **current** | — |
| `bytecode-concurrency` | bytecode | frontier | `closed-no-shared-leaf` | **current** | — |
| `bytecode-float-numeric` | bytecode | frontier | `closed-rejected-candidate` | **current** | — |
| `bytecode-iterator-control` | bytecode | frontier | `closed-no-shared-leaf` | **current** | — |
| `bytecode-portable-workload-admission` | bytecode | frontier | `closed-no-shared-leaf` | **current** | — |
| `bytecode-regex` | bytecode | frontier | `closed-rejected-candidate` | **current** | — |
| `bytecode-register-architecture` | bytecode | extra | `closed-rejected-candidate` | **current** | — |
| `bytecode-semantic-boundary-reach` | bytecode | extra | `closed-rejected-candidate` | **current** | — |
| `bytecode-target-guards` | bytecode | frontier | `target-guard` | **current** | — |
| `bytecode-text-map` | bytecode | frontier | `closed-rejected-candidate` | **current** | — |
| `bytecode-wide-numeric` | bytecode | frontier | `closed-rejected-candidate` | **current** | — |
| `compiled-architecture-target-budget` | compiled | extra | `closed-rejected-candidate` | **invalidated** | scope-content-drift:compiler-production |
| `compiled-byte-output` | compiled | frontier | `closed-no-shared-leaf` | **invalidated** | scope-content-drift:compiler-production |
| `compiled-concurrency` | compiled | frontier | `closed-rejected-candidate` | **invalidated** | scope-content-drift:compiler-production |
| `compiled-current-control` | compiled | frontier | `closed-no-shared-leaf` | **invalidated** | scope-content-drift:compiler-production |
| `compiled-float-numeric` | compiled | frontier | `closed-rejected-candidate` | **invalidated** | scope-content-drift:compiler-production |
| `compiled-iterator-control` | compiled | frontier | `closed-no-shared-leaf` | **invalidated** | scope-content-drift:compiler-production |
| `compiled-regex` | compiled | frontier | `closed-rejected-candidate` | **invalidated** | scope-content-drift:compiler-production |
| `compiled-sudoku-quotient` | compiled | frontier | `closed-insufficient-breadth` | **invalidated** | scope-content-drift:compiler-production |
| `compiled-target-guards` | compiled | frontier | `target-guard` | **invalidated** | scope-content-drift:compiler-production |
| `compiled-text-map` | compiled | frontier | `closed-no-shared-leaf` | **invalidated** | scope-content-drift:compiler-production |
| `compiled-wide-numeric` | compiled | frontier | `closed-rejected-candidate` | **invalidated** | scope-content-drift:compiler-production |
| `cross-family-architecture-ownership` | bytecode, compiled | extra | `closed-no-shared-leaf` | **invalidated** | scope-content-drift:compiler-production |

## Contract

A frontier closure is invalidated by a changed group definition, changed target/guard classification, changed evidence path or content, changed benchmark source, or changed relevant production scope. Extra architecture closures use the same evidence, benchmark, and scope checks without pretending to own frontier rows.

Production scopes cover the v12 spec, compiler, bytecode VM, shared runtime, and canonical external stdlib. Go tests, caches, generated artifacts, and unrelated documentation are excluded. Compiler changes invalidate compiled closures, bytecode VM changes invalidate bytecode closures, while shared semantic/runtime/stdlib changes invalidate both.

`--advance-closure ID` replaces only a named, freshly reviewed closure snapshot. It may add a newly tracked scope, but it refuses to advance when an existing required scope has drifted because replacing that global scope snapshot could falsely validate other closures that share it.

The selector does not run benchmarks. An invalidated closure authorizes a bounded evidence refresh, not an implementation candidate. Candidate admission still requires one concrete material mechanism in at least three unlike verifier-backed applications and the established guard gate.

## Next recommendation

Use this selector as the gate for the next performance tranche. It currently selects 12 closures: `compiled-architecture-target-budget`, `compiled-byte-output`, `compiled-concurrency`, `compiled-current-control`, `compiled-float-numeric`, `compiled-iterator-control`, `compiled-regex`, `compiled-sudoku-quotient`, `compiled-target-guards`, `compiled-text-map`, `compiled-wide-numeric`, `cross-family-architecture-ownership`.

Why: relevant production or evidence identities changed after those closures were measured, so their prior performance dispositions are no longer sufficient evidence for the current runtime.

What it entails: refresh only the named closures with verifier-backed repeated processes and arithmetic means, update their evidence records and closure baseline, then reapply the three-unlike-family candidate gate. Do not infer that an invalidated closure already contains an eligible optimization.

## Reproduction

```sh
python3 v12/bench_performance_evidence_ledger_test.py
v12/bench_performance_evidence_ledger \
  --json-out v12/docs/perf-baselines/2026-07-21-performance-evidence-invalidation-ledger.json \
  --markdown-out v12/docs/perf-baselines/2026-07-21-performance-evidence-invalidation-ledger.md
```
