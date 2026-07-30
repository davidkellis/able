# Performance direction model refresh retained

Date: 2026-07-30

## Decision

Retain the fail-closed performance-direction evidence refresh and no production
compiler, runtime, interpreter, VM, canonical-stdlib, language, dependency,
benchmark, or WASM change.

The default scorecard, frontier, and 23-entry invalidation ledger were current,
but the non-default architecture-selection chain was not. Several inputs still
pinned the former frontier SHA-256
`2609bdccad9679b75dc9d10e07e430c2049d65d6021530a73f34eccdcf9d716f`,
the structural strategy pinned an older closure-ledger identity, and four
contract tests asserted obsolete 128-row aggregates. The affected tools failed
before they could answer whether a new general compiler or VM direction was
admissible.

The retained correction updates those exact inputs and generated reports to
the current 130-row frontier, updates the aggregate assertions, and adds the
complete fast architecture chain to `v12/run_all_tests.sh`. Python bytecode
generation is disabled by that runner so the evidence checks do not leave
`__pycache__` artifacts in the repository.

## Current direction result

The refreshed models agree that no production candidate is admissible:

| Model | Current result |
| --- | --- |
| Combined frontier | 130 rows, 10 meets, 120 misses, 273.420211 seconds target excess |
| Residual-cost cohort | 5 unlike applications / 10 mode rows, 80.586842 seconds excess, no eligible mechanism |
| Compiled architecture budget | no eligible mechanism; even perfect removal of each selected application's largest local owner leaves at least 3.5962x required |
| Bytecode semantic regions | 5 material families, zero current semantic-region candidates |
| Bytecode native proxy | 65 common applications; compiled-equivalent substitution removes 94.42% of bytecode excess, but 29 rows still miss |
| Native hot-function gate | zero contract-eligible hot-function classes and zero selected backends |
| Foreign/JIT/direct-codegen ADR | closed under the current Go-owned runtime contract |
| Shared-runtime cutover | one bounded family, below the required three unlike families |
| Invalidation ledger | 23 current closures, zero invalidations, empty selector |

The only architecture route with enough modeled reach is a whole-engine
portable backend. The reviewed ADR still rejects every concrete form: keeping
Go-owned values causes dense semantic exits after four to seven primitive
instructions, while foreign ownership duplicates the language runtime and
requires identity-graph conversion. JIT and direct machine-code variants do
not remove that ownership boundary. These routes remain excluded.

The compiled conclusion is unchanged for a stronger reason than a stale
profile: strict applications already omit `pkg/interpreter`, primitive and
Array carriers remain native, exact residual conversions are explicit semantic
boundaries, and current profiles split among different application semantics.
No repeated unclosed generated-Go owner reaches three unlike families.

## Retained changes

- Rebased the semantic-region, native-hot-tier, compiled-budget,
  cross-engine-budget, structural-strategy, backend-ADR, and shared-runtime
  evidence inputs on current frontier and ledger identities.
- Regenerated their checked JSON/Markdown products in dependency order.
- Updated exact aggregate assertions for 130 rows, 10 meets, 120 misses,
  273.420211 seconds total excess, and the current 65-application native proxy.
- Added the frontier, ledger, residual, architecture, backend, semantic-ABI,
  and shared-runtime cutover checks to the default v12 runner.
- Set `PYTHONDONTWRITEBYTECODE=1` in the runner.

No benchmark-specific branch, named-container rule, non-primitive nominal
special case, global execution-context ABI, duplicate runtime, foreign
backend, JIT, executable-memory path, or WASM work was introduced.

## Verification and cleanup

- The 130-row scoreboard and five-sample evidence gate pass.
- Five frontier tests and the exact frontier check pass.
- Ten ledger tests pass with one intentional skip; all 23 closures are
  current and the selector is empty.
- Residual, compiled/bytecode budget, semantic-region, native-tier,
  cross-engine, structural-strategy, backend-ADR, semantic-ABI, and cutover
  tests and exact artifact checks pass.
- The complete `just bench-architecture-budget-check` passes, including the
  shared semantic ABI Go packages and generated manifest/flow/heap/binding
  checks.
- The updated default runner passes every preflight, non-compiler package, all
  34 compiler batches, and the complete bytecode fixture corpus.
- Every focused test completed well below one minute.
- The exact 142,316 KiB architecture cache, 2,400,964 KiB full-run cache,
  272 KiB direction workspace, and generated Python cache were removed.

The machine-readable companion is
`2026-07-30-performance-direction-model-refresh-retained.json`.

## Next

Audit sustained multi-feature workload depth across the 65 current
applications and admit at most one new realistic source-equivalent application
only if it fills a material duration/interaction gap.

Why: the current suite covers every feature triple, but many compiled misses
are short and all current exact owners are closed or family-specific. A
long-running realistic interaction can create a legitimate new benchmark
identity and reveal whether one compiler-added native-lowering cost repeats
outside startup noise.

What it entails: rank existing applications by sustained main-phase work,
feature interactions, target excess, and source equivalence; identify a real
gap rather than duplicating a covered family; then, only if a gap exists,
implement equivalent Able, Go, Python, and Ruby applications with a public
verifier and five-process baselines. Profile only an exact new owner that is
also material in at least two unlike existing applications, giving the
required three-family breadth.

Why it matters: without a changed production mechanism or new portable
application, the current ledger correctly prevents another speculative
optimization. A carefully selected sustained workload is the narrowest honest
way to obtain new evidence while preserving native carriers, interpreter-free
compiled graphs, and the ban on benchmark-shaped compiler rules.
