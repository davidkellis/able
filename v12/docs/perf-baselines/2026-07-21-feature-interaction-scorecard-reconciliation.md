# Feature-interaction scorecard promotion and frontier reconciliation

Date: 2026-07-21

## Decision

Promote Concurrent Text Index, Validated Job Pipeline, Dependency Wave
Validation, and Concurrent Event Routing in both compiled and bytecode modes.
The reviewed selection now contains 83 rows: 45 compiled and 38 bytecode.
The full-status scorecard contains 90 rows, and every selected row has exactly
five successful verifier-backed Able and reference processes in its promoted
source report.

No compiler, bytecode VM, generated runtime, canonical stdlib, language,
benchmark source, nominal lowering, or WASM performance change is retained.
The eight new rows reproduce already-closed ownership families, so the rebuilt
frontier still has zero actionable groups.

## Independent promotion cohort

The promotion used fresh Go 1.26, Python 3.14, and Ruby 4.0 references and one
fresh five-process Able cohort per mode. All 40 Able and 60 reference
executions completed and verified. The earlier targeted application reports
remain independent controls rather than being substituted into the promoted
five-process cohort.

| Application | Compiled s | Able/Go | Bytecode s | Able/Python | Able/Ruby |
| --- | ---: | ---: | ---: | ---: | ---: |
| Concurrent Text Index | 0.996 | 163.28x | 0.634 | 7.51x | 6.19x |
| Validated Job Pipeline | 3.024 | 530.53x | 0.834 | 7.90x | 9.95x |
| Dependency Wave Validation | 1.246 | 296.67x | 0.420 | 8.75x | 8.17x |
| Concurrent Event Routing | 2.666 | 522.75x | 3.032 | 102.09x | 62.13x |

The scorecard evidence check accepts all 83 selected rows with five successful
Able/reference samples each. The canonical stdlib source tree remains
`6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`.

## Promotion contract

The mode-aware selection manifest now includes all four applications in both
modes. The complete refresh tool's coverage-extra partition also includes a
bounded four-application group, so a future full refresh cannot silently omit
the promoted rows. Its dry-run and focused tests pass.

The aggregate promotion deliberately reuses the 28 source reports named by
the previously promoted current cohort and appends the two fresh interaction
reports. This is the scorecard's normal explicit-source contract: no timing
row was copied, averaged, or edited by hand. The resulting current scorecard
retains seven excluded full-status probes and adds the eight selected rows.

## Frontier reconciliation

The rebuilt schema-2 frontier reports:

- 83 selected rows: 45 compiled and 38 bytecode;
- 8 snapshot target meets and 75 misses;
- 6 source-exact established guards and 2 unestablished snapshot crossings;
- 134.540 seconds above the aggregate 95%-of-reference target budget; and
- zero actionable groups.

The four compiled rows join `compiled-concurrency`. Their exact generated-main
profiles put `bridge.currentGID` / `runtime.Stack` at 95.48%-96.94%
cumulative. That expands exact breadth from seven to eleven applications but
does not invalidate the rejected fixed-context family, which materially
regressed unrelated N-Body and K-Nucleotide guards.

The four bytecode rows join `bytecode-concurrency`. Their warmed main profiles
split below the task/dispatcher parents into member lookup, call/return,
scheduler atomic, type-match, allocation, and GC work. Those generic families
have already completed broad gates, and no second material child occurs in at
least three unlike applications without a prior disposition. Exact breadth
therefore expands from seven to eleven while the group remains
`closed-no-shared-leaf`.

## Verification

- `v12/bench_selection_manifest_check --manifest v12/bench-selection-manifest.json`
- `v12/bench_refresh_external_scorecard --dry-run --tag 2026-07-21-interaction-promotion`
- `python3 v12/bench_refresh_external_scorecard_test.py`
- `python3 v12/bench_scorecard_evidence_check.py --scorecard v12/docs/perf-baselines/2026-07-21-feature-interaction-scorecard-promotion.json --selection-manifest v12/bench-selection-manifest.json --require-runs 5`
- `python3 v12/bench_performance_frontier_test.py`
- `python3 v12/bench_performance_frontier.py --check`

## Next recommendation

Start an architecture-definition tranche for a semantically complete
operand-addressed/register bytecode engine, beginning with the call and
continuation boundary rather than another boxed-stack micro-optimization.

Why: the expanded applications invalidate none of the closed local owners.
The remaining VM gap is the accumulated cost of dispatch, boxed value
transport, and continuations. Prior whole-function modeling found that a
register form could remove 30%-44% of representation instructions in eight
unlike applications, but the restricted executable prototype translated no
hot functions because `CallName` and its continuation contract were omitted.
Partial typed lanes and prelude executors then lost their savings to repeated
materialization and reconciliation. A complete vertical slice is the next
untried architecture-scale family; more helper shaves would revisit rejected
noise.

What it entails: specify one register-frame and continuation ABI that covers
cached `CallName`, native/direct dispatch, inline returns, coercion,
raise/propagation, executor yield, and resumable state; lower a bounded generic
function class end-to-end without per-op boxed-stack reconciliation; retain a
cold fallback for unsupported functions; and gate it on at least three unlike
applications plus all six established target guards. The first tranche should
be a correctness and dynamic-reach vertical slice with counters, followed by
repeated arithmetic-mean A/B timing only if real hot functions execute in the
new engine. Continue to defer WASM.
