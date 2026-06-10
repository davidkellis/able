# Runtime-contract performance-evidence reconciliation

## Decision

**rebaseline-reviewed-shared-runtime-scope**.

The tree-wide runtime hash invalidated every closure, but the semanticabi package is outside that scope and its new runtime APIs have zero ordinary production call sites. The only live shape effect is one nil retained-root slice header per legacy iterator wrapper. Two independent five-process cohorts in six direct iterator/control applications produced 120 verified runs. Six rows are workstation-volatile, so these data admit neither a speedup nor a causal regression; they do admit rebasing the reviewed scope without changing any prior closure disposition.

This is an evidence rebase, not a performance win. It retains every prior closure disposition and admits no optimization candidate.

## Scope review

- Pre-reconciliation closures: 21 reviewed, all invalidated only by `scope-content-drift:runtime-production`.
- `internal/semanticabi` remains outside `runtime-production`.
- New reconstruction/snapshot API calls in ordinary production execution: 0.
- Live shape effect: IteratorValue adds one retained-root slice header. Legacy constructors leave it nil, so they add no backing allocation and preserve driver behavior; only wrapper size changes.

## Repeated-process evidence

All 120 Able processes passed their output verifiers; zero failed or timed out. Each benchmark/mode has two independent five-process cohorts. Bytecode used forward and reverse orders. `volatile` means the two cohort means differ by more than 15%.

| Benchmark | Mode | Pooled mean (s) | Cohort spread | Volatile | Reference ratio |
|---|---:|---:|---:|---:|---:|
| array_slice_window | compiled | 0.1010 | 2.00% | no | 16.03x Go |
| binary_event_log | compiled | 0.5970 | 1.69% | no | 55.28x Go |
| dependency_plan | compiled | 0.1280 | 56.00% | yes | 26.67x Go |
| document_audit | compiled | 0.1240 | 53.06% | yes | 22.96x Go |
| lexical_rollup | compiled | 0.1230 | 5.00% | no | 23.65x Go |
| option_result_config | compiled | 0.2100 | 18.75% | yes | 42.86x Go |
| array_slice_window | bytecode | 1.3670 | 80.12% | yes | 21.29x Python / 8.53x Ruby |
| binary_event_log | bytecode | 8.6630 | 8.45% | no | 23.85x Python / 17.49x Ruby |
| dependency_plan | bytecode | 0.6330 | 19.79% | yes | 17.58x Python / 5.42x Ruby |
| document_audit | bytecode | 0.4520 | 59.77% | yes | 32.29x Python / 9.70x Ruby |
| lexical_rollup | bytecode | 0.7270 | 5.95% | no | 32.03x Python / 12.08x Ruby |
| option_result_config | bytecode | 1.4200 | 9.44% | no | 37.57x Python / 12.41x Ruby |

## Interpretation

The measurements are deliberately descriptive. They are not paired before/after trials, and half the rows show substantial workstation load sensitivity. Arithmetic means across repeated processes prevent one noisy sample from governing the decision, while the spread flag prevents noisy means from being presented as an optimization result.

The reviewed runtime change does not add work to iterator stepping, calls, maps, numeric operations, return handling, or type matching. Existing iterator constructors leave the new retained-root slice nil. The new Hasher and host-driver APIs are cold outside conformance code. Consequently the scope can be rebased while every earlier performance closure remains closed on its existing rationale.

## Next recommendation

Begin one bounded `shared-value-heap-production-pilot` at a generic call/return boundary shared by both interpreters.

Why: the semantic contract is exact and the performance ledger is current again. A live generic boundary is the smallest step that can test whether the shared representation removes duplicated conversion/ownership work without committing to a foreign runtime or a wholesale migration.

What it entails: add an opt-in internal path at one broadly used call/return boundary; preserve aliases, identity, errors, effects, and exact ordinary-runtime fallback; exercise both interpreters with semantic conformance tests; then use repeated verifier-backed cohorts over unlike applications. Revert unless one shared mechanism improves broad guards. No WASM, foreign heap, executable memory, benchmark branch, named-container rule, or non-primitive nominal special case is admitted.
