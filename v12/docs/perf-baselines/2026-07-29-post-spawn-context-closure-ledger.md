# Performance evidence invalidation ledger

## Decision

**no-invalidated-closure**.

Do not rerun a closed performance tranche; no checked evidence identity is invalidated.

The ledger contains 23 closures: 23 current and 0 invalidated.

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
| `compiled-architecture-target-budget` | compiled | extra | `closed-rejected-candidate` | **current** | — |
| `compiled-byte-output` | compiled | frontier | `closed-no-shared-leaf` | **current** | — |
| `compiled-concurrency` | compiled | frontier | `closed-rejected-candidate` | **current** | — |
| `compiled-current-control` | compiled | frontier | `closed-no-shared-leaf` | **current** | — |
| `compiled-float-numeric` | compiled | frontier | `closed-rejected-candidate` | **current** | — |
| `compiled-iterator-control` | compiled | frontier | `closed-no-shared-leaf` | **current** | — |
| `compiled-regex` | compiled | frontier | `closed-rejected-candidate` | **current** | — |
| `compiled-sudoku-quotient` | compiled | frontier | `closed-insufficient-breadth` | **current** | — |
| `compiled-target-guards` | compiled | frontier | `target-guard` | **current** | — |
| `compiled-text-map` | compiled | frontier | `closed-no-shared-leaf` | **current** | — |
| `compiled-wide-numeric` | compiled | frontier | `closed-rejected-candidate` | **current** | — |
| `cross-family-architecture-ownership` | bytecode, compiled | extra | `closed-no-shared-leaf` | **current** | — |

## Contract

A frontier closure is invalidated by a changed group definition, changed target/guard classification, changed evidence path or content, changed benchmark source, or changed relevant production scope. Extra architecture closures use the same evidence, benchmark, and scope checks without pretending to own frontier rows.

Production scopes cover the v12 spec, compiler, bytecode VM, shared runtime, and canonical external stdlib. Go tests, caches, generated artifacts, and unrelated documentation are excluded. Compiler changes invalidate compiled closures, bytecode VM changes invalidate bytecode closures, while shared semantic/runtime/stdlib changes invalidate both.

`--advance-closure ID` replaces only a named, freshly reviewed closure snapshot. It may add a newly tracked scope, but it refuses to advance when an existing required scope has drifted because replacing that global scope snapshot could falsely validate other closures that share it.

The selector does not run benchmarks. An invalidated closure authorizes a bounded evidence refresh, not an implementation candidate. Candidate admission still requires one concrete material mechanism in at least three unlike verifier-backed applications and the established guard gate.

## Next recommendation

Use this selector as the gate for future performance work. Because it currently selects nothing, do not start another optimization or profiling tranche until a checked source, semantic, stdlib, benchmark, or scorecard trigger invalidates at least one closure.

Why: all 23 closures are current. Repeating a closed profile or candidate would measure unchanged evidence, while broad aggregate labels have already failed to identify one general mechanism.

What it entails: keep `just bench-evidence-ledger-check` green during ordinary work. After an intentional production/spec/stdlib/benchmark change, run the selector, refresh only the closures it names with repeated verifier-backed arithmetic means, update their evidence hashes, and then reconsider candidate admission. If no trigger occurs, work on correctness or specified language/stdlib completeness rather than manufacturing a benchmark optimization.

## Reproduction

```sh
python3 v12/bench_performance_evidence_ledger_test.py
v12/bench_performance_evidence_ledger \
  --json-out v12/docs/perf-baselines/2026-07-21-performance-evidence-invalidation-ledger.json \
  --markdown-out v12/docs/perf-baselines/2026-07-21-performance-evidence-invalidation-ledger.md
```
