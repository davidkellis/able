# Mutex Work Queue Application and Three-Program Profile Gate

Date: 2026-07-20

## Outcome

The Mutex operation-depth gap is closed with a third unlike portable
application, `mutex_work_queue`. The application, verifier, and independent
Go/Python/Ruby references are retained. The subsequent three-program compiler
and bytecode profile gate keeps no runtime or canonical-stdlib performance
change.

The new application models four workers claiming 4,096 dynamically assigned
jobs from a shared queue. It exercises an ordinary `with_lock` claim, work
outside the critical section, and an awaited `mutex.await_lock` commit whose
unlock is protected by `ensure`. Its deterministic result is:

```text
4:4096:4096:4096:2047207199:339361:2047207199
```

Tree-walker, bytecode, compiled Able, Go 1.26, Python 3.14, and Ruby 4.0 all
produce output accepted by the same verifier. The canonical Able source and
the sibling benchmark-repository source are byte-identical.

## Coverage and execution contract

`mutex_work_queue` is registered in the catalog, feature coverage,
operation-depth ledger, mode-aware selection manifest, bytecode audit, and
full-scorecard grouping. It uses the existing portable concurrency contract:

- compiled Able and Go: four logical CPUs from affinity pool `0-3`, goroutine
  executor;
- bytecode Able, Python, and Ruby: one logical CPU, affinity `0`;
- one process at a time, 1 GiB Go memory guard, and a 59-second per-process
  ceiling.

The catalog now contains 38 portable applications and 39 canonical benchmark
sources, with one intentionally diagnostic-only source. Feature coverage has
15 families, 16 normative families, and all 38 portable applications
registered. Operation depth has 21 entries: 16 sufficient, two insufficient,
and three local-only. `mutex_locking_and_ensure` is now sufficient because its
three applications are a ledger, an await journal, and a dynamic work queue
rather than three variations of one algorithm.

The reviewed selection now contains 69 application/mode rows: 38 compiled and
31 bytecode. Its SHA-256 is
`9f1d7eae4a05b7cdb4c378906ebc894648092c61f486a5db4cfd4da5aa1ee7ec`.

## Repeated baseline

The foreign references use five independent verified processes:

| Runtime | Mean | CV | Successful processes |
|---|---:|---:|---:|
| Go 1.26 | 0.0047 s | 4.77% | 5/5 |
| Python 3.14 | 0.0259 s | 10.47% | 5/5 |
| Ruby 4.0 | 0.0474 s | 5.00% | 5/5 |

Two independent five-process Able cohorts were retained so workstation noise
is averaged rather than inferred from a favorable single run:

| Mode | Cohort A mean | Cohort B mean | Pooled mean | Pooled median | Pooled CV |
|---|---:|---:|---:|---:|---:|
| compiled | 1.356 s | 1.724 s | 1.540 s | 1.440 s | 16.99% |
| bytecode | 0.412 s | 0.344 s | 0.378 s | 0.345 s | 19.87% |

All 20 Able samples verified. The pooled compiled mean is about 327.5x the Go
reference. The pooled bytecode mean is about 14.6x Python and 8.0x Ruby. This
row therefore materially misses both project targets and is useful selection
evidence despite the workstation variance.

## Profile gate

Five verifier-backed CPU-profile processes were merged for each of Mutex
Ledger, Mutex Await Journal, and Mutex Work Queue in each measured Able mode.
One separate allocation-profile process per application/mode was also
captured. Profiles are retained under `v12/interpreters/go/.profiles/` with the
prefix `20260720_mutex_`.

### Compiled

The exact common compiled owner is goroutine identity lookup:

| Application | CPU samples | `bridge.currentGID` cumulative | `runtime.Stack` cumulative |
|---|---:|---:|---:|
| Mutex Ledger | 5.54 s | 91.52% | 91.16% |
| Mutex Await Journal | 6.65 s | 93.23% | 93.08% |
| Mutex Work Queue | 10.81 s | 90.38% | 89.92% |

This is strong shared attribution, but it does not admit a new candidate. The
generic fixed execution-context ABI already replaced this lookup in an earlier
gate and caused a stable 54.7% regression in unrelated N-Body. A local Mutex
shortcut would encode benchmark/facility-specific behavior and would not
provide the task-local propagation contract needed by arbitrary Able code.
The existing rejection therefore remains binding.

Compiled allocation profiles do not expose a second shared application wall.
Process initialization dominates Ledger and Journal. Work Queue additionally
attributes 7.31% of allocation objects to `currentGID`, 6.49% to generic
call-value transport, and 2.44% to await-lock construction, but these leaves
are not material in all three applications. The eager small-integer cache
accounts for 66.57%-86.27% of sampled objects and remains closed after its
previous lazy, sizing, packing, and transport gates failed broad guards.

### Bytecode

The bytecode profiles have no exact removable Mutex leaf shared by all three:

| Application | CPU samples | Broad owner | Mutex/call evidence |
|---|---:|---|---|
| Mutex Ledger | 1.29 s | `runResumable` 69.77%; `execBinary` 16.28% | ordinary-lock builtin 2.33%; call opcode 20.16% |
| Mutex Await Journal | 0.56 s | `runResumable` 60.71%; `execBinary` 16.07% | await callback 16.07%; call opcode 7.14% |
| Mutex Work Queue | 1.34 s | `runResumable` 67.16% | call opcode 23.88%; call-name 15.67%; await callback 8.21%; binary 8.21% |

Ledger's ordinary locking and the two awaited-lock applications deliberately
take different semantic paths. Their only common parents are the broad VM run,
call, arithmetic, lookup, and frame machinery already tested across much wider
application groups. None identifies a new Mutex optimization, and reopening a
closed broad VM parent from these profiles would optimize profile aggregation
rather than a shared language operation.

Bytecode allocation profiles are likewise dominated by process startup: the
eager small-integer cache accounts for 70.93%-93.62% of sampled objects, with
parser/typechecker startup visible where the application body is short. There
is no common application allocation leaf.

## Tooling robustness retained

The repository cleanup command intentionally removes `v12/tmp`. The default
scratch paths in `bench_refresh_go_refs`, `bench_refresh_interpreter_refs`, and
`bench_compare_preserved_compiled` now recreate that directory before calling
`mktemp`, matching the already-correct external comparison and suite runners.
This is a tooling fix only; it allows a benchmark refresh immediately after
cleanup without manual directory recreation.

## Full scorecard reconciliation

One fresh full-corpus cohort promoted the expanded selection without row
splicing. It contains 76 application/mode status rows and all 69 reviewed rows
(38 compiled and 31 bytecode) have exactly five successful Able and matched
reference samples. Every selected Able sample verified; timeouts occur only in
the explicitly excluded full-status rows.

The current selected scorecard has six compiled target meets—Binary Trees,
QuickSort, Base64, JSON, Monte Carlo Pi, and PiDigits—and two bytecode target
meets—JSON and PiDigits. The other 61 selected rows miss. Mutex Work Queue is
1.426 s compiled versus 0.0046 s Go (310.00x) and 0.350 s bytecode versus
0.0384 s Python / 0.0448 s Ruby (9.11x / 7.81x). Both Able values agree with
the two independent preselection cohorts.

The regenerated 69-row cross-mode frontier reports 107.901 seconds of
aggregate time above the per-row target budgets and zero actionable groups.
The new compiled Mutex row joins the already-rejected execution-context group;
the bytecode row joins the current exact concurrency group with no common
removable leaf. Scorecard, evidence, selection, catalog, and frontier checks
all pass.

## Decision

Keep the benchmark, references, verifier, catalog/coverage registration,
selection row, profiles, and scratch-root fixes. Keep no compiler, VM, runtime,
or stdlib performance candidate from this gate. The three-program evidence is
general enough to confirm the compiled owner and broad enough to reject a
Mutex-specific response; it also shows that bytecode has no comparable exact
Mutex leaf.

## Next recommendation

Add a third unlike portable wide-numeric application and then profile
`fixed_width_128`, `rational_series`, and that application together. Wide
numeric nominal methods remain one of only two insufficient portable operation
depths and currently have only two consumers, so another implementation trial
would still risk learning one workload's shape.

The application should use wide signed/unsigned arithmetic in a task unlike a
checked-wrapper exercise or rational series—for example deterministic wide
integer parsing, accumulation, and checksum/range validation. It requires
equivalent Go/Python/Ruby programs, one shared verifier, source-equivalence and
catalog/coverage/operation-depth/selection registration, and two repeated
five-process Able cohorts before profiling. Advance only a primitive-wide or
shared nominal-lowering improvement whose exact leaf is material across all
three and which clears unrelated guards. Do not introduce an `Int128`,
`UInt128`, rational, container, or benchmark-named compiler rule.
