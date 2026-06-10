# Compiler struct-field reach closure

## Decision

Retain the general positional-aware named-field correction and close its
compiler-scope performance invalidation.

The 63-application census found real execution reach in only two of eleven
emitted branch families. Twenty-one rotating baseline/candidate/Go cohorts for
every reached application found no statistically distinguishable candidate
regression: all four paired 95% intervals include zero. All 63 strict graphs
remain interpreter-free, every census run passed its public verifier, and all
252 timed processes passed the same public verifiers.

This is a performance-neutral correctness closure. It does not claim that the
four concurrent applications are close to Go; they remain between 8.15x and
221.96x Go time and expose a separate, material concurrency/runtime boundary
gap.

## Strict reach census

The retained census is
`2026-07-27-compiler-struct-field-reach-census.tsv`, SHA-256
`1144bdb1e2190a2095ab21577ac96719db86820f6e4219b9ab062137dc54317c`.

- All 63 applications emitted the common Array, member, Awaitable,
  callable-member, FutureError, generated-main, and String probe sites.
- All 63 linked `array_apply`, `awaitable_shape`, `callable_member_get`,
  `future_error_details`, `main_array_format`, and `main_array_shape`;
  48 linked `array_values`.
- `callable_member_get` executed in `future_await_race`,
  `await_channel_mux`, `mutex_await_journal`, and `mutex_work_queue`.
- `awaitable_shape` executed in the latter three applications.
- String, Array, generic member, FutureError, generated-main formatting, and
  functional-update corrected reads had zero application execution reach.
- The instrumenter hard-fails if any expected common probe site is missing.
  Actual runtime hits in both reached families provide positive execution
  controls.
- Every probe binary passed its catalog setup, arguments, working directory,
  executor policy, and public verifier. Every final dependency graph omitted
  `pkg/interpreter`.

## Balanced performance disposition

The retained 21-cohort report is
`2026-07-27-compiler-struct-field-reach-balanced-21.json`, SHA-256
`fe484f24d67078d91149b58a99786a6a4855970217857f6a107c0b01e5109df3`.
Its 252 raw process records are in
`2026-07-27-compiler-struct-field-reach-balanced-21-samples.tsv`, SHA-256
`283260be52e982e10c4f943a6cdd1407feca11b7de5bfc8c6bd715114545361f`.

| Application | Baseline mean | Candidate mean | Raw mean delta | Paired mean delta | Approx. paired 95% interval | Go mean | Candidate as Go performance |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| `await_channel_mux` | 0.254228s | 0.250492s | -1.47% | -1.15% | -5.55% to +3.24% | 0.004894s | 1.95% |
| `future_await_race` | 0.034133s | 0.035581s | +4.24% | +5.52% | -1.06% to +12.09% | 0.004366s | 12.27% |
| `mutex_await_journal` | 0.391587s | 0.394607s | +0.77% | +2.26% | -5.42% to +9.95% | 0.004099s | 1.04% |
| `mutex_work_queue` | 0.945521s | 0.987128s | +4.40% | +5.66% | -3.54% to +14.86% | 0.004447s | 0.45% |

The lanes rotate baseline/candidate/Go order every cohort and use the catalog
CPU budget, affinity, executor, arguments, setup, and public verifier. The
10–20% paired standard deviations confirm that these concurrent scheduler
applications are naturally noisy; none of the observed mean deltas separates
from zero.

## Rejected inliner experiment

Go reports both original shared accessors as non-inlineable at cost 127
against an 80-unit budget. A general experiment split the normal map lookup
from the cold positional fallback, but the wrapper still cost 98 because the
fallback call itself consumes the inliner budget. Focused semantic tests
passed, but the split did not establish an inline fast path. The experiment was
removed; no speculative production code remains.

## Evidence closure

The ledger bootstrap differs from its predecessor only in the reviewed
`compiler-production` tree hash:

- old:
  `30e71f61e48405c87191288c7683da108e5de641d314ac1570f778ac4770b13d`
- new:
  `68d53484af4eba833020274b15896862b356cd091ae93818c9f7631eddab8f54`

No closure definition, evidence record, benchmark source, selection identity,
or other production scope changed. The checked ledger now reports all 21
closures current and zero invalidations. Its SHA-256 is
`0852af6b112829cdae82254a5052aea68833a797dee6064c93d28cecdb64771d`.

No additional compiler, runtime, interpreter, bytecode VM, canonical stdlib,
language, dependency, benchmark, or WASM production change was retained.

## Next

Profile the three unlike reached concurrent applications
`await_channel_mux`, `mutex_await_journal`, and `mutex_work_queue` with fresh
interpreter-free CPU, allocation, and boundary-transition counters, and compare
them to their equivalent Go programs.

Why: the field correction is neutral, but these programs deliver only
0.45%–1.95% of Go performance. That is now the largest concrete gap exposed by
this tranche and is directly aligned with the goal of keeping static Able
values in native Go carriers and avoiding compiler/runtime boundary work.

What it entails: attribute generated Go versus bridge/runtime ownership,
count `runtime.Value` materializations and dynamic calls on the hot paths, and
select a production candidate only if one exact, non-closed mechanism repeats
in all three. Any candidate must be a general compiler/runtime rule and pass
fresh verifier-backed baseline/candidate/Go A/B measurements. Retain no code
if the profiles repeat only already-closed scheduler, execution-context, or
source-semantic costs.

Why it is important: compiled Able cannot approach native Go while these broad
concurrency applications spend 51x–222x Go time. Exact cross-program ownership
is required before changing a boundary or representation, so the next change
improves unlike programs rather than encoding a benchmark or named type.
