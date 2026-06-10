# Post-i64 compiled frontier reconciliation

## Outcome

The retained dynamic-`i64` compiler-boundary optimization now has a complete
current compiled product scorecard. All 45 selected compiled applications ran
five times, as did all 45 freshly built Go references: 450 timed processes,
all successful and verifier-backed. The authoritative coverage catalog chose
each row's CPU budget and executor; Able ran with `GOMEMLIMIT=1GiB` and
`GOGC=50`, and every individual timeout remained below one minute.

Five compiled snapshot rows meet the 95%-of-Go target:

- `fib` 1.029x, `binarytrees` 0.909x, `quicksort` 0.706x, `json` 0.538x,
  and `monte_carlo_pi` 0.994x.
- Independent five-run threshold cohorts establish `fib` at 1.029x pooled.
  Binary Trees, QuickSort, and JSON retain their prior established status.
- Monte Carlo remains a non-established volatile meet because older
  current-source cohorts crossed the target.
- `base64` crossed from 1.216x to 0.940x in the independent cohort and has a
  pooled 1.074x miss. It is no longer labeled a compiled target guard.
- `matrixmultiply` missed both current cohorts at 1.129x and 1.340x.

The promoted cross-mode scorecard contains 83 selected rows: 5/45 compiled
and 3/38 bytecode snapshot meets. Cross-cohort stability establishes four
compiled and two bytecode guards; Monte Carlo compiled and Base64 bytecode are
snapshot meets but remain non-established. The other 75 rows miss, totaling
135.228 seconds above their per-row 95%-of-reference budgets.

## Selection result

The rebuilt ownership ledger has zero actionable groups. Every current group
is either a protected target guard, lacks one exact concrete leaf shared by at
least three unlike programs, has insufficient breadth, or has already rejected
its general candidate under broad guards. The full refresh therefore does not
justify another arbitrary profile or a benchmark-specific optimization. No
compiler, VM, stdlib, language, workload, reference, or WASM code changed in
this tranche.

The evidence membership was also reconciled: `fib/compiled` moved into the
compiled target guards, while `base64/compiled` moved into the byte-output
family. This keeps group recommendations consistent with the independent
stability evidence rather than merely satisfying the frontier validator.

## Verification

- Strict scorecard evidence: 83 selected rows and 90 full-status rows, with
  exactly five successful Able/reference samples for every selected row.
- `bench_external_scoreboard --check`: pass.
- `bench_performance_frontier.py --check`: pass.
- Performance-frontier tests: 5 pass.
- Scorecard-refresh tests: 2 pass.
- Feature coverage tests: 5 pass.
- Feature-interaction matrix test: pass.
- Operation-depth tests: 8 pass.

## Next recommendation

Refresh the weakest remaining compiled ownership evidence: the
`compiled-iterator-control` group is still marked `pre-current-binary` across
Array Slice Window, Dependency Plan, Document Audit, Lexical Rollup, and
Option/Result Config. Build current binaries first, then collect bounded clean
generated-main CPU/allocation profiles and intersect exact concrete owners.
Advance only an owner material in at least three unlike applications and not
already closed by the Array-growth, iterator, nullable, union, or map/cache
gates. This is the smallest evidence refresh that can legitimately reopen an
optimization family; profiling a currently exact closed group would only
repeat evidence.
