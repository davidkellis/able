# Compiled spawn-gated callable context retained

Date: 2026-07-29

## Decision

Retain activation of the existing callable execution-context ABI when any
statically loaded Able module contains `spawn` or `await`.

This is a general scheduler-syntax rule. It does not introduce another ABI,
special-case an application or non-primitive nominal type, change the runtime,
or alter language semantics. Spawn-free and await-free programs retain their
previous generated source; dynamic and host compatibility entries remain.

## Implementation

- Loaded-graph scheduler-context detection now recognizes both
  `ast.SpawnExpression` and `ast.AwaitExpression`.
- Detection remains independent of compiler options and entry-package
  placement, so imported spawn bodies select the same ABI.
- Context-aware native awaitable helpers are emitted whenever callable
  context is active, including spawn-only programs with no explicit source
  `await`.
- Imported-spawn and generated-source guards cover activation and the
  spawn-only helper dependency.

No runtime, interpreter, bytecode VM, canonical stdlib, language, dependency,
fixture, benchmark, non-primitive nominal, or WASM change was made.

## Frozen compiler and full-catalog gate

The baseline and candidate were built with Go 1.26.5:

- baseline SHA-256:
  `059d453b83ba8ab55f25013995cbfca2916b4f0aed3290ec84552ff680a092bb`;
- candidate SHA-256:
  `2d20c8724ca170756ca0196f4c88f5a6a130d2635fea0f358ecaa767617ea193`.

The complete 63-application strict census passed:

- 63/63 baseline and 63/63 candidate executions;
- 126/126 public-verifier checks;
- 126/126 final dependency graphs without `pkg/interpreter`;
- identical stdout for every baseline/candidate pair;
- unchanged application source identities;
- exactly 20 candidate generated sources newly selected callable context; and
- 43 generated sources were byte-identical: 39 spawn-free controls plus the
  four rows already selected by the retained await gate.

The 20 newly reached applications were Binary Trees, Channel Rollup, Future
Pipeline, Mutex Ledger, Concurrent Text Index, Validated Job Pipeline,
Dependency Wave Validation, Concurrent Event Routing, Concurrent Document
Pipeline, Concurrent Stencil Reduction, Concurrent Signal Dispatch,
Concurrent Transform Chain, Concurrent Policy Callbacks, Concurrent Graph
Visitors, Concurrent Audio Voices, Concurrent Packet Codecs, Concurrent
Scene Tiles, Concurrent Tree Folds, Concurrent State Machines, and Concurrent
Stateful Pipeline.

The baseline and candidate census TSV SHA-256 values are
`6885f8d08333491ef78c6932434e125a253dbc0476a80c913838f23a5b544b50`
and
`d3289f9d6806cc28964da0b479d895810943e6c157dc3adc08df99e6e21684ce`.
Binary identities are not compared across separate build roots; generated
source identity is the governing zero-reach check.

## Repeated timing

All 20 reached applications ran in seven order-rotated
baseline/candidate/Go cohorts pinned to CPUs 0-3 with `GOMAXPROCS=4`, the
goroutine executor, `GOGC=50`, and `GOMEMLIMIT=1GiB`. Every process passed its
public verifier. The raw 420-process TSV SHA-256 is
`52339e6ec4d3dcd9f743e491c7388990769c702b00cd019fdba56007556dc1ae`.

Because the host was naturally noisy, Binary Trees and eight initially
regressing short rows were extended to fifteen paired samples. Their extension
TSV SHA-256 values are
`fe7a8a2082ac34e8a62ccb39dcd2aecee5b6a60194cd5c9efb3d4e70878d85f1`
and
`3189ab6cf6645510c48c067dea882b7d8ef163c38210952a02be436b89d25368`.

The final per-application paired geometric ratios were:

| Application | Pairs | Candidate / baseline | Result |
| --- | ---: | ---: | ---: |
| Binary Trees | 15 | 0.988248 | -1.18% |
| Channel Rollup | 15 | 1.044373 | +4.44% |
| Future Pipeline | 7 | 0.971853 | -2.81% |
| Mutex Ledger | 7 | 0.151163 | -84.88% |
| Concurrent Text Index | 7 | 0.927981 | -7.20% |
| Validated Job Pipeline | 15 | 1.040388 | +4.04% |
| Dependency Wave Validation | 7 | 0.923312 | -7.67% |
| Concurrent Event Routing | 7 | 0.962197 | -3.78% |
| Concurrent Document Pipeline | 7 | 0.936014 | -6.40% |
| Concurrent Stencil Reduction | 7 | 0.753739 | -24.63% |
| Concurrent Signal Dispatch | 7 | 0.986201 | -1.38% |
| Concurrent Transform Chain | 15 | 1.014250 | +1.43% |
| Concurrent Policy Callbacks | 15 | 1.036511 | +3.65% |
| Concurrent Graph Visitors | 7 | 0.966410 | -3.36% |
| Concurrent Audio Voices | 15 | 1.066582 | +6.66% |
| Concurrent Packet Codecs | 15 | 0.984217 | -1.58% |
| Concurrent Scene Tiles | 7 | 0.878503 | -12.15% |
| Concurrent Tree Folds | 15 | 1.009033 | +0.90% |
| Concurrent State Machines | 7 | 0.976392 | -2.36% |
| Concurrent Stateful Pipeline | 15 | 1.022718 | +2.27% |

Thirteen of 20 applications improve. The cross-application paired geometric
ratio is 0.884258, an 11.57% reduction. Excluding Mutex Ledger, the remaining
19 still have a 0.970408 ratio, a 2.96% reduction.

The seven slower rows are all short applications and range from 0.90% to
6.66% by paired geometric ratio. They are counterexamples, not grounds for a
benchmark-specific compiler predicate. A narrower gate based on application
identity, nominal type, family, or runtime boundary counts would violate the
generality requirement.

Binary Trees received the strongest guard because it was already at native
performance. Its fifteen-run means were 17.077888 seconds baseline,
16.624459 seconds candidate, and 16.246988 seconds Go. The paired
candidate/Go geometric ratio was 1.023852, so the retained candidate delivers
97.67% of equivalent Go performance and remains above the 95% target. Its
paired A/B evidence is mixed enough to treat the small change as neutral.

## Allocation and causal boundary evidence

Five alternating main-phase allocation processes per lane covered all 20
reached applications, for 200 verified processes. The raw and summary
SHA-256 values are
`4cd74d31f2220b6bb83fef54b258612e7c80f801696806802584daa8b6438db3`
and
`8cce948838ebbcbc3faf921e9f88db99b9e7020c4f868b205ece8c43add90384`.

Mutex Ledger allocated 56.17% fewer bytes and 49.39% fewer objects. Future
Pipeline and Dependency Wave reduced both measures by roughly 0.4%-0.5%.
Binary Trees was effectively unchanged. Every other row stayed within
approximately -0.04% to +0.12% bytes and -0.04% to +0.06% objects. The rule
therefore has no material broad allocation penalty.

A diagnostic-only bridge overlay counted exact `currentGID` calls in five
alternating processes per lane across all 20 applications, for another 200
verified processes. The raw and summary SHA-256 values are
`014918962e98160c2501ef5873725558ce99b74e6e7fd1207afa10deeeb30255`
and
`58e21a65e0953a2d3e64553e392795fc3187ba54a60828bfb569b4e0292b5bc5`.

Nineteen of 20 applications reduced exact calls; thirteen reduced them by at
least 36%. Examples include Channel Rollup 82.2 to 7 (-91.48%), Future
Pipeline 790.4 to 64.6 (-91.83%), Mutex Ledger 8,227.8 to 15 (-99.82%),
Dependency Wave 658.0 to 14.8 (-97.75%), and Concurrent Stateful Pipeline
82.4 to 15 (-81.80%). Validated Job Pipeline is the retained counterexample:
4,122.0 to 4,114.0 (-0.19%), showing most of its remaining boundary calls
belong to a compatibility or scheduler surface outside this activation rule.

## Verification

Bounded commands, each below one minute, passed:

- default, imported, spawn-only, await, force-on, and no-context source-shape
  guards;
- nested spawn and imported/package-environment behavior;
- captured receiver, captured interface, callable, Future, Channel, Mutex,
  cancellation, native-waker, stale-waker, and both-executor parity;
- strict Vector and closure-owned kernel parity;
- four partitioned concurrency fixture groups;
- goroutine await/Future parity;
- `go test ./cmd/ablec ./pkg/compiler/bridge -count=1 -timeout 60s`; and
- formatting, source-size, and whitespace checks.

After recording the summaries and SHA-256 identities, the disposable 16 GiB
tranche workspace under `/var/tmp` and an inactive 11 MiB Able extern-Go cache
under `/tmp` were deleted. No unrelated temporary directory was touched.

## Next

Selectively refresh and promote the authoritative 63-application compiled
scorecard and its derived frontier.

Why: the checked scorecard and closure ledger predate spawn-gated activation.
They still rank residuals using the old boundary costs and cannot safely
select the next compiler owner.

What it entails: preserve the 43 source-identical rows, collect or promote
stable matched Go 1.26.5 evidence for the 20 changed rows, reject noisy
cohorts using source-identical controls, regenerate the 126-row scoreboard,
frontier, and closure ledger, and then profile the largest remaining exact
owner shared by at least three unlike misses.

Why it is important: this tranche removed the intended boundary but most
short concurrency applications remain well behind Go. A current scorecard
distinguishes residual scheduler/runtime work from costs already eliminated,
so the next optimization remains evidence-led and general.
