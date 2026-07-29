# Post-spawn-context scorecard and residual-owner closure

Date: 2026-07-29

## Decision

Promote the 20 source-changed compiled rows, preserve the previous evidence
for the 43 source-identical compiled rows, and retain no additional production
change.

The retained spawn-gated callable-context rule changed generated Go in exactly
20 of the 63 strict compiled applications. Five new Able and matched Go 1.26.5
processes passed the public verifier for all 29 rows in the four affected
source reports: the 20 changed rows and nine source-identical controls. The
controls showed a 10.31% geometric timing drift on a host concurrently running
unrelated workloads, so replacing their stable prior evidence would have
promoted host noise. The selective scorecard replaces only rows whose
generated source changed.

Fresh CPU and exact allocation profiles then found `bridge.ToInt` in all three
unlike residual applications. That is ordinary semantic-boundary
materialization, not a newly exposed native-carrier lowering opportunity. Its
global integer cache was already rejected because it slowed the
allocation-light TapeLang guard by 4.17%. The three-application owner gate
therefore closes without a prototype.

## Selective scorecard refresh

The accepted timing cohort used Go 1.26.5, `GOMEMLIMIT=1GiB`, `GOGC=50`, the
catalog logical CPU budgets and executor policies, and the matched affinity
pool `5,10,3,6`. Every accepted process passed its external public verifier:

- 145/145 matched Go processes;
- 145/145 Able processes;
- zero timeouts, failures, or verifier mismatches.

An initial 15-process Able generality run was discarded before comparison
because its generated application used Go 1.26.4 while the reference used Go
1.26.5. The accepted rerun forced Go 1.26.5. Rejected samples are not scorecard
evidence.

The nine byte-identical controls moved as follows relative to their retained
Able/Go ratios:

| Control | Ratio drift |
| --- | ---: |
| Await Channel Mux | -25.20% |
| Fib | +2.22% |
| Future Await Race | +5.88% |
| Manifest Normalization | +17.35% |
| Matrix Multiply | +1.06% |
| Mutex Await Journal | +6.39% |
| Mutex Work Queue | +11.17% |
| Policy Record Dispatch | +43.66% |
| Sensor Calibration | +48.20% |
| Geometric control drift | **+10.31%** |

This is a noise gate, not a claim that the changed rows regressed. The 20
newly selected row ratios have a 7.114236x geometric mean, 4.23% above their
prior formal ratios on this noisy short-process lane. The retention tranche's
more precise paired A/B evidence remains the causal measurement: it improved
the same reached surface by 11.57%.

The resulting authoritative scoreboard is:

| Measure | Before spawn gate | Selective refresh |
| --- | ---: | ---: |
| Compiled target passes | 7 / 63 | 7 / 63 |
| Compiled geometric Able/Go ratio | 4.263718x | 4.320152x |
| Compiled positive target excess | 4.750737s | 4.751789s |
| Bytecode target passes | 4 / 63 | 4 / 63 |
| Bytecode geometric comparison ratio | 12.780200x | 12.780200x |
| Bytecode positive target excess | 221.503684s | 221.503684s |

Binary Trees averages 10.2700 seconds versus Go at 11.0316 seconds, or
0.930962x Go time. It remains an established compiled target guard and
delivers 107.42% of Go throughput.

The 126-row scorecard has SHA-256
`9adbc937a7048eb69f81386920c44b604e9e817b22ecaa183989f72906ff4f3a`.
All 126 selected rows retain five successful Able and reference samples.
The regenerated frontier has zero actionable groups, and the reviewed
scope-baseline refresh leaves all 23 evidence closures current with zero
invalidations.

## Residual profile gate

Three unlike current misses were rebuilt strictly with `--no-fallbacks`,
the goroutine executor, `GOMAXPROCS=4`, the affinity pool `5,10,3,6`, and Go
1.26.5. Their binaries omit `pkg/interpreter`. Each application contributed
25 verifier-passing CPU profiles and three verifier-passing exact allocation
snapshots, for 84 successful profile executions.

| Application | CPU duration / samples | Main bytes, mean | Main objects, mean |
| --- | ---: | ---: | ---: |
| Await Channel Mux | 1.79s / 1.97s | 5,173,336 | 96,748.7 |
| Validated Job Pipeline | 993.92ms / 2.34s | 1,832,016 | 32,937.7 |
| Concurrent Stateful Pipeline | 335.17ms / 790ms | 10,599,960 | 192,749.0 |

The relevant exact owners are:

| Owner | Await Channel Mux | Validated Job Pipeline | Stateful Pipeline | Disposition |
| --- | ---: | ---: | ---: | --- |
| `bridge.ToInt`, objects/run | 4,098 | 4,107 | 65,554.3 | shared, but rejected global cache |
| `bridge.currentGID`, objects/run | 8,195 | 4,114 | 15.3 | material in only two |
| nominal struct construction, objects/run | — | 4,608 | 49,155.3 | material in only two; nominal special cases forbidden |

CPU ownership agrees with the exact allocation classification:

- Await Channel Mux spends 73.60% cumulative CPU in `currentGID`;
- Validated Job Pipeline spends 92.31% cumulative CPU there;
- Concurrent Stateful Pipeline spends only 2.53% there, while `ToInt`
  accounts for 21.52% cumulative CPU.

The remaining `currentGID` cost does not clear the required three-unlike-
application breadth, and broader execution-context ABI changes are already
closed. The only exact three-application leaf is `ToInt`, whose global cache
failed the TapeLang guard. Nominal struct work reaches only two applications
and may not receive named or non-primitive special cases. No admissible shared
owner remains.

## Preserved evidence and scope

Machine-readable decisions and identities are in:

- `2026-07-29-post-spawn-context-scorecard-and-owner-closure.json`;
- `2026-07-29-post-spawn-context-scorecard-refresh.json`;
- the selected, preserved-control, and full reports named
  `2026-07-29-post-spawn-context-*.json`;
- `2026-07-29-post-spawn-context-stdlib-source-state.json`.

Human-readable CPU and allocation tops are under:

- `2026-07-29-post-spawn-context-residual-profiles/`.

No compiler, runtime, interpreter, bytecode VM, canonical stdlib, language,
dependency, benchmark, fixture, nominal-lowering, or WASM change was made in
this scorecard/profile tranche. The only code addition is the general
`bench_select_comparison_rows` evidence utility and its tests.

Verification passed:

- the comparison-row selector, scorecard-evidence, frontier, and
  closure-ledger unit tests;
- the checked external scorecard, 126-row frontier, and 23-entry ledger;
- the five-sample evidence check for all 126 selected rows;
- `go test ./cmd/ablec` in 5.518 seconds;
- JSON parsing, Python byte-compilation, and whitespace checks.

The repository-wide file-length checker still reports pre-existing generated
benchmark `target/compiled` snapshots over 1,000 lines. The active compiler
sources touched by the retained spawn tranche remain below the limit; this
evidence-only tranche added files of 106 and 69 lines.

## Next recommendation

Do not begin another performance optimization until the checked evidence
selector reports an invalidated closure. Use the next tranche for bounded v12
correctness and release verification.

Why: this refresh closes the only open performance group. Repeating unchanged
profiles or forcing a rejected boxing/cache route would not move the project
toward native lowering safely.

What it entails: keep the scorecard, frontier, and closure-ledger checks green;
run the focused compiler and benchmark-evidence tests within the one-minute
test limit; then address a concrete v12 correctness, spec, or stdlib gap until
a production, benchmark, or semantic change legitimately invalidates a
performance closure.

Why it is important: a current closure ledger preserves the performance gains
and native-carrier invariants already established while ensuring the next
optimization is triggered by new evidence rather than by benchmark pressure.
