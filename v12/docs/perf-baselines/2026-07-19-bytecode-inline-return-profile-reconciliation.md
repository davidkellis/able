# Bytecode inline-return profile reconciliation

Date: 2026-07-19

Decision: close the current `finishInlineReturn` tranche with no retained
runtime change. Fresh profiles split by workload, and the one new shared
write-barrier candidate failed the broad application gate.

## Profile protocol

The restored post-frame-census binary was built as
`/tmp/able-inline-return-profile-baseline-20260719` with SHA-256
`57fe72757410edc3bc8974e0a7d5ee0be12a1bca0764f27183f887fdd8990eb1`.
Fixed Width 128, Distance Field, Option/Result Config, and Word Frequency each
ran in a separate process using its catalog input and working directory, the
canonical external stdlib, `GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`, one
allowed workstation CPU, and a 55-second cap. Every output passed its public
verifier.

Fixed Width and Distance produced 8.09 and 5.65 seconds of CPU samples.
Option/Result and Word Frequency initially produced only 0.78 and 1.42
seconds, so four additional verified profiles were collected for each and
merged. Their final merged samples total 3.78 and 7.33 seconds.

## Exact return attribution

| Child of `finishInlineReturn` | Fixed Width | Distance | Option/Result | Word Frequency |
| --- | ---: | ---: | ---: | ---: |
| complete `finishInlineReturn` | 9.15% | 8.85% | 17.99% | 13.92% |
| `popCallFrameFields` | 4.20% | 2.48% | 2.38% | 3.14% |
| program switch from return | 1.48% | 2.48% | 2.12% | 1.09% |
| inline/program return coercion | no material sample | no material sample | 5.56% plus 3.97% slotless | 5.46% |
| materialized/no-coercion check | 1.11% | 1.24% | 0.79% | 0.27% |
| slot-frame release | 0.62% | about 1.42% | 1.32% | 1.23% |
| returned-value append | 0.87% | 0.35% | 0.26% | 0.14% |

The broad parent therefore does not represent one remaining operation. Fixed
Width and Distance use the materialized/no-coercion and raw-return side;
Option/Result and Word Frequency use canonical or callable coercion/type
matching. Slot release and returned-value append repeat but are too small for
another representation experiment. Program switching repeats, but its active
lookup state was already reduced in the preceding retained tranche.

Within `popCallFrameFields`, the earlier empty-sidecar candidate did not touch
one smaller shared source: zeroing the saved seven-field active lookup bundle
on every full pop. `runtime.wbZero` or write-barrier work beneath the pop was
sampled in Fixed Width, Option/Result, and Word Frequency. The preceding census
also proved that every observed full frame carried an active lookup program,
making this a generic lifetime candidate rather than a benchmark shape.

## Deferred-clear candidate

The candidate stopped zeroing the saved active lookup bundle immediately
before the frame slice was shortened. The popped slot is inaccessible and is
overwritten by the next full-frame push. To preserve cross-run memory hygiene,
`resetForRun` explicitly cleared active lookup state across the backing frame
capacity before returning a VM to or from its pool. A focused test covered
that pool-reset cleanup.

The change did not alter the return ABI, program restoration, cache contents,
call dispatch, coercion, exceptions, scheduling, frame eligibility, nominal
lowering, or language semantics. Its binary SHA-256 was
`6ba22571ab7079dab5d14b5357146337ff35b469ce4ddb69ccf3c20e742e70b5`.

## Repeated verifier-backed A/B

Five alternating baseline/candidate pairs used the same execution contract as
the profiles. All 40 processes passed their public verifiers. Every sample,
including a 9.52-second Fixed Width candidate outlier, remains included.

| Application | Baseline mean | Candidate mean | Change | Candidate wins |
| --- | ---: | ---: | ---: | ---: |
| Fixed Width 128 | 7.992 s | 8.136 s | +1.80% | 3/5 |
| Distance Field | 5.702 s | 5.724 s | +0.39% | 3/5 |
| Option/Result Config | 0.854 s | 0.816 s | -4.45% | 3/5 |
| Word Frequency | 1.466 s | 1.468 s | +0.14% | 2/5 |

Positive changes are regressions. Fixed Width's four non-outlier candidate
samples average 7.79 seconds, but excluding the slow sample would be selective
reporting; the complete cohort is ambiguous. Distance and Word Frequency are
neutral/slightly slower with mixed pair wins. Option/Result is favorable by
mean but wins only three pairs. With two clear neutral controls and no
three-unlike-program improvement, the candidate fails the broad bar without a
longer favorable-only cohort.

Raw samples are preserved in
`2026-07-19-bytecode-inline-return-deferred-clear-ab-samples.tsv`.

## Revert and correctness

The deferred clear, pool cleanup, and focused candidate test were removed. The
rebuilt binary exactly matches the preserved baseline at SHA-256
`57fe72757410edc3bc8974e0a7d5ee0be12a1bca0764f27183f887fdd8990eb1`.
No production code, compiler, stdlib, tree-walker, fixture, language, or WASM
change remains from this tranche.

Verification after the revert:

- focused lookup-state, pool, call-frame, and return tests passed in 0.524
  seconds;
- `go test ./pkg/interpreter -run 'TestBytecode' -count=1 -timeout 60s`
  passed in 23.430 seconds;
- split parity groups 02-04, 05-08, 09-11, 12-13, and 14 passed in 10.193,
  26.078, 6.838, 12.119, and 20.854 seconds;
- every command remained below the project's one-minute test ceiling.

## Next recommendation

Close the return-boundary loop and refresh the complete mode-aware bytecode
selection scorecard after the retained active-lookup improvement. Use five
verifier-backed runs per application against the current Python and Ruby
references, then cluster the largest remaining misses by exact flat VM/runtime
leaf rather than by cumulative parent.

Why: active lookup produced a broad retained gain, but two subsequent return-
frame ideas failed averaged application gates and the fresh profiles show no
untested material child shared by three unlike programs. A suite-wide refresh
will quantify which language-feature clusters now dominate and keep the next
optimization aligned with real programs rather than repeatedly subdividing a
locally exhausted return path. Profile only the selected cluster under the
same one-process memory guardrails. Continue to defer WASM.
