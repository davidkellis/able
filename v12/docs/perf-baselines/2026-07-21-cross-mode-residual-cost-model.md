# Cross-mode residual cost model

The five highest-excess unlike applications span text/map, wide numeric, float numeric, regex/interface dispatch, and concurrency and account for more than half of the current aggregate target excess.

## Selection

The 5 unlike applications contribute 75.501 of 188.452 aggregate target-excess seconds (40.06%).

| Application | Family | Compiled s / ratio / excess | Bytecode s / ratio / excess | VM allocations | Dynamic VM ops |
| --- | --- | ---: | ---: | ---: | ---: |
| `k_nucleotide` | text-map | 2.898 / 35.82x / 2.813 | 46.530 / 31.78x / 44.989 | 1,245,836,872 B / 23,947,521 | guard-skipped |
| `fixed_width_128` | wide-numeric | 0.206 / 35.52x / 0.200 | 8.522 / 24.32x / 8.153 | 1,348,450,616 B / 31,911,044 | 60,446,436 |
| `distance_field` | float-numeric | 0.090 / 6.77x / 0.076 | 5.914 / 15.25x / 5.506 | 368,035,900 B / 26,000,120 | 78,000,165 |
| `policy_record_dispatch` | regex-dispatch | 0.226 / 25.98x / 0.217 | 7.452 / 234.34x / 7.419 | 139,533,024 B / 1,453,755 | guard-capped |
| `concurrent_event_routing` | concurrency | 2.914 / 520.36x / 2.908 | 3.256 / 96.90x / 3.221 | 289,347,356 B / 2,820,542 | 21,367,576 |

The dynamic opcode observer changes execution cost and is used only for counts. A capped or skipped row is missing evidence, never a zero. Static lowering covers all five applications.

## Fresh generated and lowering evidence

| Application | Functions / instructions | Static tracked lowering | Generic-union fast / fallback | Residual polymorphic |
| --- | ---: | --- | ---: | ---: |
| `k_nucleotide` | 17 / 652 | JumpIfNotTypedPattern=1, LoadName=19, LoadSlot=79, LoadSlotStructField=6 | 0 / 0 | 36 |
| `fixed_width_128` | 5 / 137 | LoadSlot=19 | 0 / 0 | 2 |
| `distance_field` | 1 / 58 | LoadSlot=7 | 0 / 0 | 1 |
| `policy_record_dispatch` | 16 / 623 | JumpIfNotTypedPattern=13, LoadName=59, LoadSlot=49, LoadSlotStructField=16 | 5,632 / 0 | 3,074 |
| `concurrent_event_routing` | 15 / 802 | JumpIfNotTypedPattern=10, LoadName=95, LoadSlot=30, LoadSlotStructField=16 | 11,264 / 0 | 6,146 |

## Dynamic VM operation evidence

| Application | Status / total | Leading operations | Inline-call misses |
| --- | ---: | --- | ---: |
| `k_nucleotide` | guard-skipped | The normal measured main is about 45 seconds, so stats observer overhead cannot fit the repository's sub-minute process guard. Existing bounded opcode/profile evidence remains authoritative. | n/a |
| `fixed_width_128` | measured / 60,446,436 | LoadSlot=8,283,481, LoadSlotStructField=6,816,906, Const=5,875,015, Pop=5,062,519, StoreSlotNew=4,000,006 | 0 |
| `distance_field` | measured / 78,000,165 | LoadSlot=22,000,024, Const=8,000,029, JumpIfBinaryCompareFalse=8,000,000, StoreSlotFloatRegion=6,000,000, BinaryIntAdd=4,000,008 | 0 |
| `policy_record_dispatch` | guard-capped | The stats-observed process did not complete inside 55 seconds and emitted no result. It was not extended or interpreted as zero. | n/a |
| `concurrent_event_routing` | measured / 21,367,576 | LoadSlot=4,933,960, Pop=3,007,698, StoreSlotNew=2,206,998, Const=1,227,323, Jump=1,212,825 | 0 |

## Cross-mode mechanism gate

| Mechanism | Modes | Unlike families | Covered app excess | Status |
| --- | --- | ---: | ---: | --- |
| End-to-end allocation pressure | bytecode, compiled | 5 | 75.501 s | `observational-only` |
| Bytecode stack/slot transport and dispatch | bytecode | 3 | 16.880 s | `closed-rejected` |
| Generated generic-union fast method dispatch | compiled | 2 | 3.125 s | `insufficient-breadth` |
| Material generated residual-polymorphic calls | compiled | 2 | 3.125 s | `insufficient-breadth` |
| Generated goroutine identity discovery | compiled | 1 | 2.908 s | `closed-rejected` |

## Reconciliation

- **End-to-end allocation pressure:** Allocation is broad but is not one mechanism: integer metadata/maps, wide nominals, float values, regex state, and concurrency environments have different semantic owners.
- **Bytecode stack/slot transport and dispatch:** LoadSlot leads all three bounded dynamic censuses, but partial register frames, operand lanes, raw carriers, and slot/call/return descendants already failed broad application gates. Source-level LoadSlot reachability alone is not new evidence for another partial rewrite.
- **Generated generic-union fast method dispatch:** Only the two structurally related dispatch applications execute it, and every observed call uses the fast path with zero fallback.
- **Material generated residual-polymorphic calls:** The material counts are 3,074 and 6,146. The other three unlike applications execute only 36, 2, and 1 calls, so this is not a three-family wall.
- **Generated goroutine identity discovery:** It is dominant only in concurrency programs, and the general execution-context candidate regressed Event Routing and Mutex Ledger despite removing much of this work.

The generated-call audit observed 16,896 fast generic-union calls and 0 fallbacks. Their complete absence from the text/map and both numeric applications prevents that boundary from explaining the selected five-family residual.

The only three-family repeated dynamic VM leader is stack/slot transport. That is architectural evidence, not a newly eligible leaf: previously attempted partial register, operand-lane, carrier, call, and return descendants did not pass broad gates.

## Decision

**no-candidate**. Do not implement another leaf optimization from this evidence; quantify whether a complete typed/register bytecode architecture can meet the remaining interpreter budget while keeping compiler work on its separate domain-owner tracks.

## Reproduction

Fresh lowering and generated-boundary acquisition:

```sh
GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1 v12/bench_bytecode_audit \
  --benchmarks k_nucleotide,fixed_width_128,distance_field,policy_record_dispatch,concurrent_event_routing \
  --output-json /tmp/able-residual-bytecode-audit.json \
  --output-md /tmp/able-residual-bytecode-audit.md
v12/bench_compiled_boundary_audit --telemetry call-path --timeout 45 \
  --benchmarks k_nucleotide,fixed_width_128,distance_field,policy_record_dispatch,concurrent_event_routing \
  --output-json /tmp/able-residual-call-path.json
v12/bench_compiled_boundary_audit --telemetry dynamic-boundary --timeout 45 \
  --benchmarks k_nucleotide,fixed_width_128,distance_field,policy_record_dispatch,concurrent_event_routing \
  --output-json /tmp/able-residual-dynamic-boundary.json
```

The VM operation counts use the existing opt-in `ABLE_BYTECODE_STATS=1` runtime-benchmark observer, one measured main per fresh process, a 1 GiB memory limit, and a 55-second outer timeout. Fixed Width, Distance Field, and Event Routing completed. Policy was capped and K-Nucleotide was skipped as recorded above.

Model generation and contract tests:

```sh
v12/bench_residual_cost_model \
  --json-out v12/docs/perf-baselines/2026-07-21-cross-mode-residual-cost-model.json \
  --markdown-out v12/docs/perf-baselines/2026-07-21-cross-mode-residual-cost-model.md
python3 v12/bench_residual_cost_model_test.py
```
