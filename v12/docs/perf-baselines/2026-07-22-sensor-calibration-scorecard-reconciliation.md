# Sensor Calibration scorecard reconciliation — 2026-07-22

## Decision

Promote `sensor_calibration` into the current compiled and bytecode scorecards,
regenerate the complete cross-mode performance frontier, and regenerate the
cross-engine architecture target budget. Retain no compiler, generated-runtime,
bytecode VM, tree-walker, canonical-stdlib, benchmark, language, or WASM code
change from this reconciliation.

The current reports now cover 50 portable applications, 93 selected repeated-
measurement rows, and 100 full-status rows. The two new selected rows are clear
target misses and repeat only closed generic owner families, so the architecture
decision remains `no-go-current-cross-engine-local-mechanism`.

## Promotion evidence

The promotion cohort uses the same source, verifier, input, execution contract,
and one-CPU serial policy as the application gate. The host CPU pool is `0-3`,
from which the catalog resolves CPU `0` for every Sensor Calibration process.
The 1 GiB `GOMEMLIMIT` and `GOGC=50` guard are preserved.

Every lane received five independent processes. All 25 processes passed, with
zero failures and zero timeouts; no sample was removed as workstation noise.

| Lane | Samples | Mean | Current ratio |
| --- | --- | ---: | ---: |
| Able compiled | `0.30, 0.25, 0.25, 0.24, 0.24` | 0.2560 s | 50.196× Go |
| Go 1.26 | `0.005696551, 0.004514813, 0.005306765, 0.005150093, 0.004684653` | 0.0051 s | — |
| Able bytecode | `3.33, 4.64, 5.12, 2.95, 2.93` | 3.7940 s | 114.622× Python / 51.831× Ruby |
| Python 3.14 | `0.028099432, 0.041965373, 0.043484423, 0.026226302, 0.025862734` | 0.0331 s | — |
| Ruby 4.0 | `0.074560789, 0.073700014, 0.070819424, 0.069513375, 0.077506164` | 0.0732 s | — |

Bytecode and Python were visibly volatile, so the decision is also checked
against the independent five-process application-gate cohort. Pooling the two
cohort means gives 3.3210 seconds for bytecode versus 0.03115 seconds for
Python and 0.07630 seconds for Ruby. The limiting ratios remain approximately
106.61× Python and 43.53× Ruby. The compiled pooled means are 0.2870 seconds
Able versus approximately 0.0049 seconds Go. Volatility cannot change either
target classification.

## Scorecard reconciliation

The promoted aggregate reuses the three explicit current source scorecards and
adds the two new promotion reports. Before merging, the canonical stdlib state
was remeasured and matched the current scoreboard exactly: 70 source files,
source-tree SHA-256
`6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`,
git head `219eff222c28406487231713753641bc49ee5b9a`, with the same recorded dirty
state. The evidence check requires five successful Able and reference samples
for all 93 selected rows and permits the seven intentional one-process excluded
bytecode status probes.

The current scorecard has:

- 50 compiled rows: five meets and 45 misses;
- 43 selected bytecode rows: three meets and 40 misses;
- seven additional unranked bytecode status rows;
- selection SHA-256
  `c9cf59cb991f0916470508777a4636aab81c56064c19747e1258a43492e02738`.

The compiled sum-of-row-means diagnostic is 49.7820 seconds Able versus
28.1489 seconds Go, or 1.7685×, with 23.618632 seconds of target excess. The
bytecode diagnostic is 151.5800 seconds Able versus 17.2499 seconds for each
row's faster Python/Ruby reference, or 8.7873×, with 136.724842 seconds of
target excess. These sums weight applications equally as coverage diagnostics;
they are not a prediction for one typical application.

## Frontier and architecture result

The regenerated frontier has 93 rows, eight snapshot meets, 85 misses,
160.343474 seconds of target excess, and zero actionable groups. Sensor
Calibration joins `compiled-text-map` as a closed-no-shared-leaf row and
`bytecode-text-map` as a closed-rejected-candidate row. The latter is now the
largest group budget at 64.822211 excess seconds, but its new profile repeats
the same rejected raw-integer, stack, type-match, string, and Array mechanisms.

Bytecode owns 85.269976% of the current total deficit. The per-engine optimistic
bounds remain unchanged because the new row admits no mechanism: making all
modeled bytecode stack transport free still leaves at least a 7× requirement,
and making the largest attributed compiled local owner free still leaves more
than a 3× requirement. The cross-engine decision therefore remains:

```text
no-go-current-cross-engine-local-mechanism
```

The architecture report no longer recommends the already-completed semantic-
region feasibility tranche. Its next gate is a structural-strategy
reconciliation across materially different routes, with no implementation
admitted until a route applies to at least three unlike applications and an
optimistic end-to-end model removes at least 25% of target excess in every
governing row.

The duplicate current frontier used by the closure and residual-cost tooling,
both per-engine architecture budgets, semantic-region feasibility report, and
native-hot-tier design budget were regenerated from the same 93-row source.
The performance closure ledger was then rebaselined only after those later
current-source gates were present; all 21 closures are current with zero
invalidations. The native compiled-proxy model now covers 43 common rows: 11
meet and 32 still miss even if bytecode is replaced wholesale by current
compiled execution.

## Verification

- 25/25 fresh Sensor Calibration promotion/reference processes verified;
- 93 selected rows and 100 full-status rows reconcile under the new manifest;
- five successful Able/reference samples are present for every selected row;
- current scoreboard regeneration and exact checked-artifact verification pass;
- performance-frontier regeneration and exact checked-artifact verification pass;
- cross-engine architecture generator tests and checked-artifact verification pass;
- all 21 performance closures are current with zero invalidations;
- catalog, selection, coverage, operation-depth, and interaction contracts pass;
- JSON, shell, Python, whitespace, and diff checks pass;
- temporary compiler, benchmark, and stdlib-state workspaces are removed.

## Next recommendation

Complete `cross-engine-structural-strategy-reconciliation` before another
performance implementation tranche.

Why: the fully current 93-row frontier has no actionable exact owner, and all
current small-helper, cache, carrier, frame, register, Go-level semantic-region,
native-leaf-reach, generated-helper, and nominal-allocation routes are either
insufficiently broad or have failed unlike-program guards. Continuing to tune
one of them would repeat a closed experiment while bytecode remains responsible
for 85.27% of the target deficit.

What it entails: compare three genuinely different general routes—typed
bytecode specialization that removes dynamic semantic work, a lower-level
portable VM backend, and further shared compiled nominal-ABI simplification—
against the current target budget and completed closure ledger. Map each route
to exact semantics, implementation/deployment cost, and at least three unlike
applications; model end-to-end target-excess reduction; then select at most one
prototype only if it clears the 25% per-row reduction gate and preserves the
five established target guards. Do not begin WASM work.
