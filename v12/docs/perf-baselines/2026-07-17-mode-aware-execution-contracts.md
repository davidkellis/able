# Mode-aware benchmark execution contracts

## Outcome

CPU and executor budgets are now first-class row metadata throughout fresh
Go/Python/Ruby reference collection, Able comparison, and scoreboard
aggregation. This tranche changes measurement tooling only. It retains no
compiler, VM, runtime, stdlib, benchmark, verifier, or reference-source change
and makes no new product-performance claim.

## Contract

The catalog assigns every benchmark/mode pair a logical CPU budget and every
benchmark an executor policy:

- ordinary compiled applications use one logical CPU and the serial policy;
- Binary Trees and the six concurrency applications use four logical CPUs in
  compiled mode and the goroutine policy;
- bytecode, prechecked bytecode, warmed bytecode-runtime, and tree-walker modes
  use one logical CPU;
- concurrency applications retain the goroutine executor policy in interpreter
  modes while remaining confined to that one-CPU interpreter budget.

`--cpu-affinity` now names an ordered CPU pool. Each runner selects the first N
CPUs required by the row, pins the process to that resolved set, and sets
`GOMAXPROCS` to N for Go/Able processes. Without an explicit pool, the runner
uses the calling process's allowed CPU set.

Each fresh row records:

```json
{
  "execution_contract": {
    "mode": "compiled",
    "logical_cpu_budget": 4,
    "cpu_affinity": "0,1,2,3",
    "executor_policy": "goroutine"
  }
}
```

Reference and Able runtime names may differ, but their CPU budget, resolved
affinity, and executor policy must agree before a ratio is created. Fresh
contract mismatches abort comparison. Scoreboard validation repeats that check
and preserves contracts in compact rows. Legacy reports without this metadata
remain readable; they cannot silently erase metadata from new rows.

The full refresh no longer exports one global `GOMAXPROCS=1`. Its child
runners derive `GOMAXPROCS` from each row, allowing one refresh to contain fair
four-CPU compiled concurrency comparisons and serial interpreter comparisons.

## Verification

- `python3 v12/bench_execution_contract_test.py` — 7 tests pass, covering
  serial/default, parallel, concurrency, CPU-pool resolution, normalization,
  compatible cross-runtime mode labels, and mismatch rejection.
- `python3 v12/bench_refresh_external_scorecard_test.py` — 2 dry-run partition
  tests pass.
- `python3 v12/bench_scorecard_selection_test.py` — 6 selection/aggregation
  tests pass.
- Shell syntax checks pass for the catalog, all three measurement runners, and
  the grouped refresh.
- A missing-source interpreter report smoke parsed its expanded row and
  retained a one-CPU bytecode contract.
- A one-second Python launch smoke exercised the resolved `taskset` path and
  retained its timeout as unavailable status with the same row contract.
- A bounded verifier-backed `fib` smoke carried a fresh Go contract through a
  compiled Able comparison with matching single-CPU serial metadata.

The smoke timings are integration evidence only: each side had one sample and
must not be used for selection or a performance claim.

## Next gate

Run the complete grouped scorecard refresh with five independent samples for
every reviewed selected row and one bounded status probe for excluded rows.
This is next because the old promoted scoreboard used one global CPU contract
and therefore cannot accurately classify intentionally parallel compiled
applications alongside serial interpreter comparisons. The refresh entails
new matched Go/Python/Ruby references and Able compiled/bytecode rows under the
catalog contract, followed by aggregation, strict variance checks, and
promotion only if all source, verifier, stdlib-state, sample-count, and
execution-contract checks pass.
