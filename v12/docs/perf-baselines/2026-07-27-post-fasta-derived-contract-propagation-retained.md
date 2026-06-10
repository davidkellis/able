# Post-Fasta derived-contract propagation retained

Date: 2026-07-27

Decision: retain two corrected derived-contract expectations and add them to
their owning aggregate gates. Add no benchmark application and retain no
production performance change.

## Coverage-expansion audit

The current recommendation allowed a new broad benchmark only if it filled a
real application-coverage gap or exposed a new shared production owner. The
audit found neither:

- all 21 operation families are sufficient or intentionally local-only;
- every portable feature pair has at least two substantial applications;
- all 165 three-family interactions are represented;
- the minimum current triple depth is 11;
- the portable catalog remains 61 applications;
- the current performance frontier has zero actionable groups.

The existing weighted-triple decision already rejects adding another
application solely to duplicate a concurrency/file/callback shape. Doing so
would bias the corpus without strengthening the generality evidence.

## Invalidation found

Generating the current feature-interaction triple report exposed one stale
test expectation. The test still expected the pre-refresh frontier total of
`182.011579` seconds rather than the current exact-source total of `182.007474`
seconds established by the Fasta source-identity closure.

A complete search found one other stale consumer in the residual cost model.
Its frontier total changed to `182.007474` and its derived selected-excess
share changed from `41.481398%` to `41.482334%`. Application membership,
selected excess, mechanism status, and the no-candidate decision did not
change.

## Retained correction

- Updated `bench_feature_interaction_triples_test.py` to the current frontier
  total.
- Updated `bench_residual_cost_model_test.py` to the current frontier total and
  derived share.
- Added the triple-frontier contract test to `just bench-catalog-check`.
- Added the residual-cost-model contract test to
  `just bench-architecture-budget-check`.

The aggregate additions ensure future scorecard/frontier roll-forwards cannot
silently leave either derived consumer stale.

## Verification

- feature coverage, pair matrix, triple frontier, and operation depth: pass;
- residual cost model: five unlike applications, ten selected mode rows, and
  no eligible mechanism;
- catalog: 61 portable applications and 140 combined programs;
- current scorecard evidence: 115 selected rows with five successful
  Able/reference samples each;
- current frontier: 115 rows and zero actionable groups;
- evidence ledger: 21 current closures and zero invalidations;
- complete deterministic architecture gate, including the newly aggregated
  residual-cost-model test: pass;
- `git diff --check`: pass.

No Able/Go/Python/Ruby benchmark source, verifier, input, measurement,
selection, compiler, runtime, VM, language, canonical stdlib, dependency, or
WASM path changed.

## Next recommendation

Keep benchmark expansion and production performance mutation paused until an
authoritative invalidation appears.

Why: the present corpus already covers all governed operation and
feature-interaction surfaces, while the reconciled residual models expose no
new general owner.

What it entails: when a genuinely new application domain, retained semantic
change, correctness failure, or non-closed owner across three unlike programs
appears, refresh only the affected evidence and evaluate one general native
lowering or boundary-elimination rule with balanced repeated A/B measurements.

Why it is important: this keeps Able focused on native Go carriers and minimal
compiled/interpreted crossings without manufacturing benchmark volume or
reopening closed mechanisms. Do not begin WASM work.
