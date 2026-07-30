# Current cross-engine architecture target-budget reconciliation

## Decision

**no-go-current-cross-engine-local-mechanism**.

No local compiler or VM candidate is admitted by the current evidence. The
scorecard has no actionable implementation group, the compiler has no exact
mechanism material in three unlike families, and free bytecode transport is
far too small to close even the smallest modeled VM target gap.

## Current target budget

| Mode | Rows | Meets | Misses | Excess seconds | Share |
| --- | ---: | ---: | ---: | ---: | ---: |
| compiled | 64 | 7 | 57 | 4.794105 | 2.08% |
| bytecode | 64 | 4 | 60 | 226.057474 | 97.92% |

The full 128-row frontier has 117 misses and 230.851579 seconds above its aggregate target. Bytecode owns 97.92% of that gap.

## Optimistic architecture bounds

| Engine | Unlike applications | Favorable assumption | Best remaining requirement | Decision |
| --- | ---: | --- | ---: | --- |
| bytecode | 6 | all transport free (at most 1.80x) | 7.79x | `no-go-register-representation-only` |
| compiled | 5 | largest exact local owner free | 3.60x | `no-go-current-compiled-architecture-mechanism` |

These are upper-bound planning models, not timing claims. They omit replacement,
materialization, translation, code-size, and secondary costs. Failing under these
assumptions closes another small helper, cache, carrier, frame, or register-only
experiment as the next tranche.

## Largest remaining group budgets

| Group | Misses | Excess seconds | Disposition |
| --- | ---: | ---: | --- |
| `bytecode-portable-workload-admission` | 5 | 74.548947 | `closed-no-shared-leaf` |
| `bytecode-text-map` | 10 | 65.292211 | `closed-rejected-candidate` |
| `bytecode-regex` | 6 | 23.137895 | `closed-rejected-candidate` |
| `bytecode-concurrency` | 23 | 17.716947 | `closed-no-shared-leaf` |
| `bytecode-wide-numeric` | 3 | 17.174000 | `closed-rejected-candidate` |
| `bytecode-float-numeric` | 4 | 15.811789 | `closed-rejected-candidate` |
| `bytecode-iterator-control` | 6 | 7.028737 | `closed-no-shared-leaf` |
| `bytecode-byte-output` | 3 | 5.346947 | `closed-no-shared-leaf` |

## Next recommendation

Run **cross-engine-structural-strategy-reconciliation**.

Compare materially different remaining structural routes against the complete target budget before building more local candidates: typed bytecode specialization that removes dynamic semantic work, a lower-level portable VM backend, and further general compiled nominal-ABI simplification. Explicitly exclude completed Go-level region/register executors and the unadmitted native leaf tier.

Why: bytecode contributes most of the measured target deficit, while current
local-owner, register, semantic-region, native-reach, and carrier/consumer
gates admit no implementation candidate. Another isolated helper would repeat
closed work rather than select an architecture capable of changing the target.

Admission gate: Advance one route only when a concrete mechanism is applicable to at least three unlike verifier-backed applications, an optimistic end-to-end model removes at least 25% of target excess in every governing row, and the route is not equivalent to a completed rejection.

Required guards: Preserve bytecode JSON and PiDigits established target guards plus compiled Binary Trees, JSON, and QuickSort guards.

This reconciliation starts with checked target-budget arithmetic and existing
closure evidence. It adds runtime code only if the three-family and target-budget
gates pass. It does not begin WASM work or add named-container/benchmark rules.

## Reproduction

```sh
python3 v12/bench_bytecode_architecture_budget_test.py
python3 v12/bench_compiled_architecture_budget_test.py
python3 v12/bench_cross_engine_architecture_budget_test.py
v12/bench_cross_engine_architecture_budget --check
```
