# Bytecode architecture feasibility and target budget

## Decision

**no-go-register-representation-only**.

Do not implement a complete typed/register representation as the next performance tranche. Measure semantic-work amplification per logical workload unit and select a shared language/runtime boundary that reduces semantic operations or allocations in at least three unlike applications.

This decision is scoped to a register representation by itself. It does not forbid future primary-VM redesign when separately justified semantic-operation or allocation work requires it.

## Target-budget model

| Application | Family | Current / target s | Operations | Transport share | Required speedup | Uniform-cost transport speedup | Remaining speedup | Target ns / semantic op |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| `fixed_width_128` | wide-numeric | 8.522 / 0.369 | 60,446,436 | 31.80% | 23.10x | 1.47x | 15.76x | 8.947 |
| `distance_field` | float-numeric | 5.914 / 0.408 | 78,000,165 | 41.03% | 14.49x | 1.70x | 8.55x | 8.872 |
| `concurrent_event_routing` | concurrency | 3.256 / 0.035 | 21,367,576 | 43.14% | 92.06x | 1.76x | 52.35x | 2.911 |
| `word_frequency` | text-map | 1.414 / 0.025 | 17,591,889 | 44.32% | 57.41x | 1.80x | 31.96x | 2.515 |
| `array_slice_window` | array-iterator | 0.744 / 0.064 | 8,751,239 | 32.79% | 11.59x | 1.49x | 7.79x | 10.916 |
| `reverse_complement` | byte-text | 3.758 / 0.040 | 60,602,670 | 30.09% | 94.20x | 1.43x | 65.86x | 0.942 |

Across the six applications, 89,338,836 of 246,759,975 operations are modeled transport (36.20%). The largest equal-cost planning gain is 1.80x, while the smallest remaining target requirement is still 7.79x.

The equal-cost calculation is deliberately favorable to transport removal: it charges each removed load/constant/pop the same average cost as a semantic operation, makes removal free, and adds no translation, register, merge, materialization, or code-size cost. It is a planning scenario, not a timing claim. The target nanoseconds show why representation work alone cannot close the budget: the surviving calls, branches, stores, collection operations, errors, and concurrency semantics would need to average 0.942-10.916 ns each.

## Live opcode closure

All 144 live opcodes are classified exactly once. Only 6 are removable representation dispatches; 138 retain observable or required semantic work.

| Category | Role | Opcodes |
| --- | --- | ---: |
| `transport` | removable-representation-dispatch | 6 |
| `scalar-numeric-conversion` | required-semantic-work | 33 |
| `binding-environment-storage` | required-semantic-work | 18 |
| `nominal-collection-text` | required-semantic-work | 17 |
| `calls-callables-dynamic-dispatch` | required-semantic-work | 14 |
| `control-return-scope-error` | required-semantic-work | 38 |
| `iteration-concurrency-suspension` | required-semantic-work | 8 |
| `definitions-imports` | required-semantic-work | 10 |

## Complete semantic closure

| Area | Required contract | Current evidence |
| --- | --- | --- |
| `cfg-register-state` | Translate complete control-flow graphs with compatible boxed/raw stacks, slot versions, liveness, and merge values before effects. | The prototype proved simple loops and returns, but current censuses retain merge_slot_versions and merge_value_identity failures. |
| `value-representation` | Represent runtime.Value plus raw i32, i64, arbitrary integers, and f32/f64 lanes, with materialization at every uncertain boundary. | Current VM sidecars and raw carriers are authoritative; isolated carrier and operand-lane candidates have mixed or negative broad results. |
| `calls-continuations` | Preserve native, inline, generic, member, static, self, and dynamic calls plus return coercion, lookup caches, suspension, and caller restoration. | The prototype proved CallName suspend/resume, but material reach existed in only two unlike applications before later semantic slices. |
| `errors-unwind-scopes` | Preserve raise, rescue, ensure, rethrow, propagation, break/continue, transient scopes, diagnostics, and effect ordering. | No complete register executor covered this closure; the ordinary VM remains authoritative. |
| `iteration-concurrency` | Preserve iterator cleanup, yield/resume, spawn, Future await/cancellation, executor flushing, and goroutine/serial state transfer. | Only the call continuation boundary was prototyped; generators and the full concurrency closure stayed on the ordinary VM. |
| `nominal-identity-storage` | Preserve struct, union, Array, map, member/index, ownership, mutation, aliasing, and generic nominal semantics without named-container rules. | Register-native MemberAccess reached six applications but regressed the broad wall-time gate; allocation and storage owners remain application-dependent. |
| `dynamic-environment-definitions` | Preserve lexical/global lookup, implicit receivers, lambdas, definitions, implementations, externs, imports, dynamic imports, and cache invalidation. | These operations remained semantic/materialization boundaries in the feasibility model. |
| `source-errors-debugging` | Retain source nodes, precise Able errors, overflow/type diagnostics, breakpoints, and public materialized results. | The ordinary instruction and runtime.Value contracts remain the only complete implementation. |

The current VM surface contains 133 production `bytecode_vm*.go` files / 32,816 lines and 119 test files / 36,191 lines. A complete parallel executor would therefore duplicate a large semantic authority unless it replaced the primary dispatcher incrementally; the prior cold-fallback executor never reached that point.

## Empirical architecture gates

| Gate | Applications | Result | Observed range |
| --- | ---: | --- | ---: |
| `register-member-access` | 6 | `rejected` | +0.00% to +14.98% |
| `allocation-neutral-register-frames` | 3 | `rejected` | +4.87% to +14.01% |
| `continuation-probe-hoist` | 3 | `rejected` | -4.34% to +5.75% |

Reasons:

- **register-member-access:** The executable register slice reached real semantic work but was neutral or slower in all six applications.
- **allocation-neutral-register-frames:** Removing essentially all added register-frame allocations left all three applications slower.
- **continuation-probe-hoist:** The hoist removed millions of misses, but the combined Word Frequency guard remained 5.75% slower than the ordinary VM.

## Reproduction

The six opcode censuses use `ABLE_BYTECODE_STATS=1` with one measured main per fresh process, `GOMAXPROCS=1`, `GOGC=50`, a 1 GiB memory limit, and a 55-second timeout. Counts are deterministic and are not timing evidence. The target side is the current five-process arithmetic-mean frontier.

```sh
v12/bench_bytecode_architecture_budget \
  --json-out /tmp/able-bytecode-architecture-budget.json \
  --markdown-out /tmp/able-bytecode-architecture-budget.md
python3 v12/bench_bytecode_architecture_budget_test.py
```
