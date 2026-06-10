# Weighted feature-interaction triple audit

Date: 2026-07-21

## Decision

Keep the report-only weighted triple frontier and recognize Concurrent Text
Index's already-executed lexical/pattern, control-flow, and Option semantics.
Add no new benchmark application and keep no compiler, bytecode-VM, runtime,
canonical-stdlib, language, or WASM performance change.

All 165 three-family combinations are already represented. The membership
reconciliation reduces depth-one triples from eight to two; both remaining
triples have one substantial source-equivalent, verifier-backed application.
Uniform depth two is not an admission requirement for triples because it would
reward combinatorial benchmark duplication rather than new language coverage.

## Ranking contract

`bench_feature_interaction_triples` joins three checked inputs:

- the portable feature-coverage manifest;
- reviewed per-family semantic priorities on a 1–3 scale;
- the current evidence-backed performance frontier.

Coverage depth is always the primary sort key. Equal-depth triples are ranked
by descending semantic weight, then by target-excess seconds from applications
that cover at least two of the three families. This makes current performance
pressure useful without allowing one slow application to outrank a genuine
coverage gap. The weights affect audit order only; they never relax semantic,
source-equivalence, verifier, breadth, or regression requirements.

The machine-readable report records exact SHA-256 identities for all three
inputs, the reviewed reasons for every weight, all adjacent applications, and
all 165 intersections. Evidence:
`2026-07-21-feature-interaction-triple-frontier.{json,md}`.

## Existing-application audit

Before this reconciliation, the suite had no empty triple and eight depth-one
triples. All eight depended on Concurrent Event Routing and involved
concurrency plus some combination of text/files, lexical patterns, control
flow, Option/Result, callables, or real program entry.

Concurrent Text Index was conservatively under-annotated. Its default contract
reads and processes 16,384 words with four workers. Per verified run it
executes:

- 16,388 typed-or-nil Channel receive selections (one per word plus four
  closed-channel worker exits);
- 16,384 nullable `classify(...)` matches;
- 8,542 nullable public-Map lookup matches (8,534 selected words plus eight
  final bucket reads).

That is 41,314 explicit hot-path Option/pattern selections. The producer,
workers, classifier, aggregation loop, and bucket loop also execute substantive
branching and iteration. These are material lexical/pattern, Option, and
control-flow semantics, not incidental declarations or dead syntax.

The canonical and sibling Able sources remain byte-identical with SHA-256
`5a2429f00e30e21af242850073e6e8c24d4022b8cad9bc7aee5e6e7860999eef`.
The current scorecard already supplies source-exact verifier evidence:

| Mode/reference | Able processes | Able mean | Reference mean | Ratio |
| --- | ---: | ---: | ---: | ---: |
| compiled / Go | 5 verified | 0.8760 s | 0.0062 s | 141.29x |
| bytecode / Python | 5 verified | 0.6340 s | 0.0844 s | 7.51x |
| bytecode / Ruby | 5 verified | 0.6340 s | 0.1024 s | 6.19x |

All Able and applicable reference processes have zero failures and timeouts.
No source, verifier, input, selection, or timing row changed in this tranche.

## Triple-frontier result

The exact baseline removes only Concurrent Text Index's three newly recognized
memberships. Across 11 discriminating portable/mixed families:

| Measure | Before | Current |
| --- | ---: | ---: |
| three-family interactions | 165 | 165 |
| zero-depth triples | 0 | 0 |
| depth-one triples | 8 | 2 |
| triples strengthened | — | 85 |

The six promoted depth-one triples now have Concurrent Text Index and
Concurrent Event Routing as independent guards. The only remaining depth-one
triples are:

1. `concurrency × expressions_arrays_text_files × functions_closures_callables`
2. `concurrency × functions_closures_callables × program_entry`

Concurrent Event Routing is a credible guard for both: it reads a real input,
runs 4,096 tasks through four Channel workers, creates a capturing scoring
closure per worker, invokes that callable through the routing path, and emits a
deterministic externally verified result. Concurrent Text supplies the
text/file, concurrency, and entry combination but intentionally has no
first-class callable; Validated Job and Dependency Wave supply callables and
concurrency but intentionally have no real file/argument entry. There is no
misclassified existing application to promote.

Adding another application solely to duplicate those two nonempty triples
would bias the corpus toward one concurrency/file/callback shape. The audit
therefore stops without a new benchmark.

## Performance disposition

The scorecard remains 46 compiled and 39 bytecode selections, with 92
full-status rows. The frontier remains 77 misses, 143.927 target-excess
seconds, and zero actionable groups.

Concurrent Text's existing exact compiled profile reproduces the closed
goroutine-identity boundary. Its exact bytecode profile reproduces closed
member-cache/call/return owners. Recognizing already-executed features exposes
no new concrete leaf and invalidates no prior candidate gate, so performance
implementation work is not admissible in this tranche.

## Verification

- triple-frontier generation and positive/error-path unit tests;
- pairwise matrix and feature-coverage contract tests;
- current coverage, selection, scorecard, and frontier replay;
- canonical/sibling source identity and current verifier/source hashes;
- full bytecode regression family;
- JSON, Python syntax, executable, file-size, whitespace, and temporary-file
  checks.

## Next recommendation

Pivot from benchmark-coverage expansion to a cross-mode residual-cost-model
audit across the highest-excess, unlike application families.

Why: pairwise coverage has minimum depth two and every three-family
interaction is present. The remaining shallow triples are credible rather than
missing, while symbol-level CPU/allocation intersections have exhausted their
current actionable leaves. More applications or another leaf-cache experiment
would add evidence volume without identifying a broad optimization.

What it entails: select high-excess representatives from unlike text/map,
numeric, regex, and concurrency families; reconcile generated-code shapes,
bytecode opcode/semantic-operation counts, allocations, and reference-runtime
work at the application level; then estimate which shared architectural cost
could close meaningful portions of multiple target gaps. Admit a candidate
only when the same concrete compiler or VM mechanism explains material excess
in at least three unlike programs and the predicted saving is large enough to
matter end to end. Continue to exclude named-container, benchmark-specific,
and WASM work.
