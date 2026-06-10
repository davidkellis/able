# Bytecode architecture target-budget reconciliation

## Outcome

The proposed typed-register architecture feasibility tranche was already
complete. The checked current generator reproduces its substantive evidence
and decision:

**no-go-register-representation-only**.

No VM, compiler, runtime, stdlib, benchmark, reference, language, or WASM code
changed in this reconciliation. The generated report changed only because the
current VM surface contains eight more production lines and eight more test
lines than the report's earlier provenance snapshot.

## Current checked budget

The six unlike applications execute 246,759,975 bytecode operations. Of those,
89,338,836 (36.20%) are the six modeled stack-transport operations; all 143
live opcodes remain classified exactly once, with 137 retaining semantic work.

| Application | Family | Transport share | Best representation-only speedup | Remaining target speedup |
| --- | --- | ---: | ---: | ---: |
| `fixed_width_128` | wide numeric | 31.80% | 1.47x | 14.17x |
| `distance_field` | float numeric | 41.03% | 1.70x | 7.69x |
| `concurrent_event_routing` | concurrency | 43.14% | 1.76x | 55.15x |
| `word_frequency` | text/map | 44.32% | 1.80x | 39.89x |
| `array_slice_window` | array/iterator | 32.79% | 1.49x | 14.07x |
| `reverse_complement` | byte/text | 30.09% | 1.43x | 82.15x |

The representation-only scenario is intentionally optimistic: it makes every
transport operation free while charging no translation, register-state merge,
materialization, code-size, or fallback cost. It therefore establishes an
upper bound, not a timing claim. Even that upper bound leaves every application
far outside its current Python/Ruby target.

## Reconciliation with completed implementation gates

The architecture model was followed by executable experiments, so the route
is not waiting for a prototype:

- Register-native member access reached six unlike applications and was
  neutral or slower in all six (0.00% to 14.98% slower).
- Removing essentially all added register-frame allocations still left three
  unlike applications 4.87% to 14.01% slower.
- Hoisting millions of continuation probes produced mixed results and left the
  combined Word Frequency guard 5.75% slower.
- The follow-up semantic-work amplification audit found no exact semantic
  operation materially amplified in three unlike families.
- The remaining two-family direct-Array boundary was tested with Matrix
  Multiply. Its cache was already more than 99.999% hot, its concrete storage
  descendants split by semantics, and perfect removal of the complete push
  subtree could yield only 1.046x and 1.087x in the two target-missing rows.
- A later secondary architecture search found a three-application struct field
  lookup leaf, but the generic candidate regressed all five measured guards.

Together these results close typed registers, stack transport, and their
currently known semantic descendants as the next performance tranche. A future
primary register IR remains possible only when independently measured semantic
or allocation work requires it and can clear the same broad application gate.

## Why no new timing cohort was run

This reconciliation is deterministic rather than a performance comparison.
The generator rechecks the current live opcode set, exact classification,
evidence files, semantic closure, application counts, and arithmetic target
model. The checked result differs from the committed artifact only in source
line-count provenance. Existing performance decisions already use repeated
process arithmetic means; rerunning them would measure workstation noise
without changing the architecture budget.

## Next recommendation

Build a compiled architecture and semantic-amplification target-budget audit
across unlike high-excess applications. Normalize generated-Go work by logical
application units, reconcile it with source-equivalent Go references, and
partition primitive checks, nominal representation, bridge conversion,
environment/control values, allocation, and host-library boundaries. Compute
the maximum plausible savings and remaining speedup for each shared mechanism.

This is next because the compiler still has a large target gap, while exact
leaf, control-ABI, primitive-range, aggregate-allocation, and bytecode-register
searches have all closed without a broad candidate. The audit should decide
whether the current generated representation can plausibly reach 95% of Go or
whether a larger generic compiler/runtime boundary must change before more
local experiments are justified. It must not introduce named-container,
application-specific, benchmark-specific, or WASM paths.

## Reproduction

```sh
python3 v12/bench_bytecode_architecture_budget_test.py
v12/bench_bytecode_architecture_budget \
  --json-out /tmp/able-bytecode-architecture-budget-current.json \
  --markdown-out /tmp/able-bytecode-architecture-budget-current.md
cmp v12/docs/perf-baselines/2026-07-21-bytecode-architecture-target-budget.json \
  /tmp/able-bytecode-architecture-budget-current.json
cmp v12/docs/perf-baselines/2026-07-21-bytecode-architecture-target-budget.md \
  /tmp/able-bytecode-architecture-budget-current.md
```
