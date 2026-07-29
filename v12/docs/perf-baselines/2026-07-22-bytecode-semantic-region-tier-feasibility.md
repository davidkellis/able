# Bytecode semantic-region tier feasibility reconciliation

## Decision

**no-go-current-semantic-region-tier**.

Do not build another Go-level semantic-region executor, superinstruction
dispatcher, whole-function register loop, or PGO-trained CLI from the current
evidence. These are completed gates, not untried implementations.

## Current target-budget sizing model

| Application | Family | Instruction share | Required gain | Uniform-cost free-region gain | Remaining gain | Model reaches target? |
| --- | --- | ---: | ---: | ---: | ---: | --- |
| `monte_carlo_pi` | stochastic-numeric | 41.79% | 1.78x | 1.72x | 1.04x | no |
| `rms_norm` | float-array | 11.76% | 8.45x | 1.13x | 7.46x | no |
| `fixed_width_128` | wide-numeric | 6.62% | 20.61x | 1.07x | 19.24x | no |
| `future_await_race` | concurrency | 48.20% | 4.35x | 1.93x | 2.25x | no |

This is a sizing model, not a mathematical upper bound: the census records
dynamic instruction share rather than sampled wall-time share. It assumes every
instruction has equal average cost and then makes the admitted region free.
Only Monte Carlo Pi crosses its current near-target threshold in that model;
the other rows illustrate how much larger their deficits are than the observed
region coverage. The executable prototype result, not this model, closes another
Go-level region executor.

The broader current six-family architecture model independently makes all stack
transport free and still leaves at least 7.79x required.

## Completed executable gates

| Route | Result | Why it is closed |
| --- | --- | --- |
| `generic-safe-typed-region-executor` | `rejected-broad-wall-time` | The exact sequences differed, and the shared out-of-line executor regressed all three governing applications. |
| `whole-function-register-executor` | `rejected-broad-wall-time` | Allocation-neutral frames and continuation-probe removal did not make the register engine beat the ordinary VM across guards. |
| `go-cross-suite-pgo` | `rejected-deployability-and-guard` | Runtime extern plugins require the exact training profile, and matched-profile JSON still regressed. |
| `current-exact-semantic-operation` | `rejected-insufficient-shared-materiality` | The six-family semantic-work audit admits zero exact operation or operation family. |

The typed-region prototype used one generic safe-opcode executor because the
exact hot sequences differed by application. Its retained repeated means were:

| Application | Pairs | Baseline | Candidate | Change |
| --- | ---: | ---: | ---: | ---: |
| `monte_carlo_pi` | 10 | 2.749152s | 2.869319s | +4.37% |
| `rms_norm` | 5 | 4.909369s | 4.934215s | +0.51% |
| `quicksort` | 10 | 1.164481s | 1.213043s | +4.17% |

## Reconciliation conclusion

The prior recommendation used “semantic-region tier” too broadly. A second
Go dispatcher and a whole-function Go executor were already tested and rejected.
The current semantic-work audit also finds zero exact operation family material
in three unlike high-excess applications. No runtime prototype is admitted.

## Next recommendation

Complete **bytecode-native-hot-code-tier-design** as a design-and-budget tranche before writing runtime code.

Design, but do not yet implement, a genuine native hot-code tier that removes host-language dispatch rather than adding another Go-level executor.

Why: Direct opcode/layout work, the generic typed-region executor, whole-function register execution, and Go PGO have already failed broad wall-time or deployability gates. A materially different tier must cross the native-code boundary and preserve the complete v12 semantic contract.

The design must cover:

- typed primitive and boxed runtime.Value inputs/outputs.
- GC-visible roots and identity-preserving nominal values.
- calls, returns, errors, rescue/ensure, suspension, futures, and cancellation.
- extern/plugin ABI compatibility without retaining a training-profile dependency.
- deoptimization or fallback at every unproved dynamic boundary.
- portable backend policy and bounded code-cache lifecycle.

Admission gate: Prototype only if the design covers one hot function/region class in at least three unlike verifier-backed applications and an optimistic end-to-end model can close or materially reduce all three target gaps.

This remains non-WASM work and may not introduce application, benchmark, named
container, or non-primitive nominal special cases.

## Reproduction

```sh
python3 v12/bench_bytecode_semantic_region_feasibility_test.py
v12/bench_bytecode_semantic_region_feasibility --check
```
