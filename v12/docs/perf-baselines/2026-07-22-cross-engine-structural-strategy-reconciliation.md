# Cross-engine structural-strategy reconciliation

## Decision

**portable-backend-performance-capable-but-no-concrete-mechanism**.

No implementation prototype is admitted. The portable lower-level backend is
the only route with enough optimistic end-to-end reach, but it is still a route
class rather than a concrete mechanism. Timing reach does not answer who owns
boxed values and GC roots, how effects and suspension cross the boundary, or how
the backend is built and distributed.

The governing frontier has 132 selected rows,
119 misses, and 274.677263
seconds of target excess. Bytecode contributes 268.726526
seconds. All 23 completed performance closures remain current.

## Route comparison

| Route | Optimistic material rows | Implementation / deployment cost | Semantic risk | Decision |
| --- | ---: | --- | --- | --- |
| Typed bytecode specialization | 3/6 | high / low | medium | `reject-current-route` |
| Portable lower-level backend | 5/5 phase-one; 6/6 full | very high / high | very high | `performance-capable-design-only` |
| Compiled nominal ABI | max breadth 2 | high / low | medium | `reject-current-route` |

### Typed bytecode semantic specialization

This deliberately favorable model makes every dynamically observed instruction
with a checked primitive proof free at uniform instruction cost. It is not a
wall-time claim. Failing this model is strong evidence that another proof flag,
slot hint, or Go-level specialized dispatch loop is too small.

| Application | Family | Proven instruction share | Excess reduction | Gate |
| --- | --- | ---: | ---: | --- |
| `fixed_width_128` | wide-numeric | 33.25% | 34.64% | pass |
| `distance_field` | float-numeric | 41.03% | 43.39% | pass |
| `concurrent_event_routing` | concurrency-text | 8.68% | 8.78% | fail |
| `word_frequency` | text-map | 8.90% | 9.06% | fail |
| `array_slice_window` | array-iterator | 29.35% | 30.73% | pass |
| `reverse_complement` | byte-text | 23.27% | 23.42% | fail |

Only three of six rows clear 25%, so the route fails the every-governing-row
gate. It also substantially overlaps the rejected typed-region, register-loop,
integer-slot-hint, and carrier-specialization experiments.

### Portable lower-level VM backend

The proxy substitutes the current semantics-preserving compiled application time
for bytecode time. This is optimistic: it charges no translation, boundary,
compile-latency, code-cache, or foreign-runtime cost. The phase-one scope excludes
concurrency execution because the current compiled proxy is itself dominated by
the closed goroutine-identity wall there; the ordinary VM must remain the exact
fallback for unsupported activations.

| Application | Family | Phase one | Excess reduction | Gate |
| --- | --- | --- | ---: | --- |
| `fixed_width_128` | wide-numeric | yes | 100.00% | pass |
| `distance_field` | float-numeric | yes | 100.00% | pass |
| `concurrent_event_routing` | concurrency-text | fallback | 100.00% | pass |
| `word_frequency` | text-map | yes | 99.56% | pass |
| `array_slice_window` | array-iterator | yes | 98.98% | pass |
| `reverse_complement` | byte-text | yes | 99.68% | pass |

All 5 unlike phase-one rows clear the
materiality bar, and five of six full-corpus rows do. This makes the route worth
an architecture decision, not yet an implementation. A per-opcode foreign
callback would merely exchange Go dispatch for ABI overhead, while a full
foreign runtime risks semantic duplication. The next decision must resolve that
ownership boundary explicitly.

### Further compiled general nominal-ABI simplification

The current five-application compiled budget has no eligible mechanism. Its best
shared exact mechanism is material in only 2 unlike families.
The apparent three-application typed-boundary intersection is composed of
allocation/type-match work in the serial row and the already-closed goroutine
identity wall in both concurrent rows. A shared counter name is not a shared
optimization mechanism, so no nominal ABI candidate is admitted.

## Semantic and deployment obligations

- **values-identity-and-gc**: Preserve primitive values, boxed values, nominal identity, aliases, mutation, and liveness without retaining movable Go pointers in foreign memory.
- **dispatch-and-nominal-semantics**: Preserve overload, interface, generic-union, inherent-method, dynamic member, and general non-primitive nominal translation rules without named-type exceptions.
- **errors-and-structured-unwind**: Preserve source-attached errors, raise/rescue/ensure/rethrow, propagation, cleanup, break/continue, and return coercion.
- **concurrency-and-suspension**: Preserve spawn, Future, await, cancellation, yielding, serial scheduling, goroutine scheduling, and bounded fairness polls.
- **externs-and-host-interop**: Preserve extern calls, host values, package/runtime state, and callback reentrancy without depending on an unstable Go ABI.
- **diagnostics-and-exact-resume**: Retain original instruction and source identity for failures, side exits, fallback, and reproducible diagnostics.
- **portable-fallback-and-distribution**: Define supported platforms, dependency and license policy, executable-memory policy, compile latency, bounded caches, and exact ordinary-VM fallback.

## Next recommendation

Complete **portable-vm-backend-abi-dependency-adr**.

Resolve whether a non-WASM portable lower-level engine can own whole bytecode activations across boxed and effectful boundaries without per-opcode foreign callbacks or a semantically divergent second runtime. Compare whole-engine C-ABI, portable JIT-library, and direct-codegen classes; select the value/root ABI, effect/resume protocol, toolchain and distribution policy, or close the route.

Why: this reconciliation shows that local specialization and another nominal ABI
refinement cannot be large enough across the governing programs, while a whole-
engine route can. The remaining uncertainty is architectural feasibility, not
another profile sample or helper-level hotspot.

Prototype admission: Admit an executable prototype only after one backend class, ABI, ownership model, supported-platform policy, and conformance/fallback plan are concrete. The retained timing model must cover at least three of the five serial governing families at 25% target-excess reduction each, and JSON plus PiDigits must retain exact ordinary-VM fallback guards.

Exclusions: No WASM, benchmark or application branches, named-container rules, non-primitive nominal fast paths, per-opcode cgo callbacks, or retry of Go-level region/register executors.

## Reproduction

```sh
python3 v12/bench_cross_engine_structural_strategy_test.py
v12/bench_cross_engine_structural_strategy --check
just bench-evidence-ledger --check
```
