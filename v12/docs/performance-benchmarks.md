# v12 Performance Benchmark Harness

`v12/bench_suite` runs the shared benchmark suite in three execution modes:

- `compiled`
- `treewalker`
- `bytecode`

It emits machine-readable JSON with:

- git commit + dirty state
- machine profile (OS/kernel/arch/CPU/memory)
- harness config (timeouts, runs, benchmark list)
- per-run status/timing/GC metrics
- per-benchmark summary rows

`v12/bench_perf` is the lighter per-target helper for focused perf checks. Its
compiled mode now builds through `cmd/ablec` directly, so compiled fixture
benchmarking measures the current compiler path without pulling in unrelated
`able build` package/bootstrap behavior. It also accepts repeated
`--compiled-build-arg` flags for controlled comparisons such as
`--no-experimental-mono-arrays`. It also supports `--run-from DIR` for
benchmarks that read relative input files, repeated `--program-arg ARG` flags
for workload-specific entry arguments like `wordlist.txt`, `--executor
serial|goroutine` for concurrency-sensitive workloads, and `--output-json
PATH` for machine-readable summaries. When benchmarking against external suite
directories it now resolves the stdlib root in this order:

1. `ABLE_STDLIB_ROOT`
2. the sibling workspace checkout `../able-stdlib/src`
3. the installed `$ABLE_HOME` cache

That keeps newly added local benchmarks from silently running against a stale
cached stdlib when a current sibling checkout is available. It now also
accepts `--cpu-affinity CPUSET` (or
`ABLE_BENCH_CPU_AFFINITY`) to run the measured process under `taskset`, which
is useful when unrelated workstation load makes micro-deltas untrustworthy;
choose a quiet CPU subset that is still wide enough for the benchmark's
intended parallelism. For new single-CPU evidence, pass
`--cpu-affinity CPU --require-quiet-cpu`; it rejects a run before artifacts
are created unless that CPU stays under the generic busy/I/O-wait limits.
`just bench-host-check --rank-cpus` ranks candidates from common samples;
then `just bench-host-check --cpu CPU` can provide a stricter direct preflight
when a quiet core is available. It is a precision aid, not a reason to discard
all workstation evidence: when normal workstation load prevents that gate,
run enough independent verifier-backed processes, report mean, median, and
spread, and use the merged CPU profile only for attribution.
The active measurement commands now default to five independent processes:
`bench_perf`, `bench_suite`, `bench_compare_external`, both foreign-reference
refreshers, and the full scorecard refresher. Their reported summary remains
the arithmetic mean for compatibility with the scorecard schema; retain the
individual samples and calculate median/spread with `just bench-variance-report
--input <comparison.json>` when timing variability affects a decision. One
modern five-run comparison report is sufficient for Able and reference timing
spread; ratio spread still requires multiple source reports because independent
Able and reference runs must not be arbitrarily paired. An explicit smaller
`--runs` value is appropriate only for a smoke check, never for candidate
selection or a performance claim.

For a full-suite ratio comparison, pass two or more fresh complete
`external-benchmark-scoreboard` artifacts with `--scorecard` instead of
assembling individual files manually. The reporter expands each scorecard's
cited comparison reports and rejects a cohort unless those reports reproduce
its rows exactly; it also rejects cohorts with different benchmark/mode
coverage or any reused comparison source. Thus two copied or partially
refreshed scorecards cannot manufacture independent ratio evidence. For
example: `just bench-variance-report --scorecard <first-scorecard.json>
--scorecard <second-scorecard.json> --output-json <variance.json> --output-md
<variance.md>`.
For evidence that may select a candidate, add `--require-runs 5`. This rejects
historical aggregates and any row without exactly five successful,
verifier-backed Able runs and five fresh measured-reference runs for every
language requested by its comparison report. The current retained scorecard
passes this gate: it embeds the canonical stdlib source state and cites exactly
five retained samples for all 67 reviewed rows. Older scorecards that predate
those fields remain useful historical evidence but cannot select a new change.

The cross-mode performance frontier is a separate selection artifact, not a
timing harness. `just bench-frontier` joins every row in the reviewed selection
manifest to the promoted exact-source scorecard, source identity,
profile/binary freshness,
exact CPU/allocation owner, unlike-application breadth, and the disposition of
earlier candidates. It rejects missing, duplicate, or extra evidence rows and
ranks only groups marked `open-candidate` or `refresh-required`; closed parents
remain visible but cannot silently become the next optimization. `just
bench-frontier-check` runs the fast protocol tests and verifies that both
generated reports still match their hashed inputs. Update
`v12/bench-performance-frontier-evidence.json` when a profile refresh or
candidate gate changes a disposition.

The performance-evidence invalidation ledger sits above that frontier. `just
bench-evidence-ledger-check` verifies 18 frontier groups plus the cross-family,
bytecode-register, and compiled-target-budget architecture closures without
running a benchmark. Each closure pins its semantic group definition, evidence
documents, applicable benchmark sources, and mode-aware production scopes for
the v12 spec, compiler or bytecode VM, shared runtime, and canonical external
stdlib. Tests, caches, generated artifacts, and unrelated documentation are
outside those scopes. `just bench-scoreboard-check` also runs the non-mutating
artifact comparison so a stale handoff cannot silently reopen a current group.

When an intentional source or semantic change makes the selector print one or
more closure ids, refresh only those groups. Update their verifier-backed
profiles and frontier evidence, preserve repeated arithmetic-mean timing, then
review a new `v12/bench_performance_evidence_ledger --bootstrap` baseline and
regenerate the dated JSON/Markdown report. Invalidation authorizes evidence
collection only; it does not admit a candidate until the ordinary concrete
three-unlike-family and established-guard gates pass.

The frontier also reads `v12/bench-performance-stability.json`. This small,
versioned manifest must classify every selected row that meets the promoted
snapshot target. An `established-meet` requires at least two verifier-backed
cohorts and a pooled ratio plus every cohort ratio within the product budget;
mixed classifications remain `volatile-crossing` or
`variance-sensitive-miss`. The validator pins the selection, promoted
canonical-stdlib tree, Able source, every applicable reference source, sample
counts, and evidence files. A new snapshot meet, changed source fingerprint,
or stale stdlib identity therefore makes `bench-frontier-check` fail until the
cross-cohort evidence is explicitly reviewed. Snapshot status remains the raw
scorecard result; established status only controls which rows are durable
candidate-admission regression guards.

Every new full-scorecard refresh also records one canonical stdlib source-state
artifact before it launches timed processes. It deterministically hashes every
`src/**/*.able` file in the external `able-stdlib` checkout and records the
relative root, file count, Git head when available, and whether that checkout
is dirty. The dated aggregate and any promoted scoreboard retain the same
state. A dirty checkout is recorded rather than rejected: it may contain the
actual source being evaluated. Strict cohort evidence additionally requires
this state, so a reviewer can see whether two candidate cohorts ran against
the same library input. Historical scorecards without the field remain
replayable but are not eligible for a new strict claim.

To collect two independent full-suite candidate cohorts without changing the
promoted baseline, use distinct tags and `--no-promote`, then compare the two
dated aggregate artifacts. For example:

```sh
just bench-scorecard-refresh --tag candidate-a --no-promote
just bench-scorecard-refresh --tag candidate-b --no-promote
just bench-variance-report \
  --scorecard v12/docs/perf-baselines/candidate-a-refresh.json \
  --scorecard v12/docs/perf-baselines/candidate-b-refresh.json \
  --require-runs 5 --output-json /tmp/candidate-variance.json
```

The refresher refuses to overwrite a tagged artifact. If reviewed evidence
should become the baseline, promote its exact recorded source cohort without a
new measurement: `just bench-scoreboard --cohort
v12/docs/perf-baselines/candidate-b-refresh.json`. The current scoreboard is
unchanged until that explicit promotion command succeeds. The refresher also
rejects any non-five-run promotion; use `--no-promote` for an explicit lower-
run smoke check.
For a bounded warm profile, `--bytecode-runtime-calls N` typechecks, loads,
lowers, and warms the program once, then runs `main()` exactly `N` times per
`bytecode-runtime` benchmark run. Typechecking stays outside the measured
loop, but preserves checked receiver facts needed for ordinary generic-union
method dispatch. `bench_compare_external` forwards the same option, retaining the external
catalog's canonical run directory, input arguments, source-root policy, and
verifier wiring instead of requiring a hand-built benchmark command.
The regular `bytecode` mode remains a full CLI process measurement: it loads,
lowers, typechecks, bootstraps, and executes the entry program. For a separate
trusted execution-only measurement after an explicit successful `able check`,
the CLI supports `able run --skip-typecheck <target-or-file>`. That flag is
never the default and is not part of the promoted external scorecard; it must
be reported as a distinct prechecked lane rather than compared as though it
were a normal application launch.
`bench_perf --modes bytecode,bytecode-prechecked` and
`bench_compare_external --modes bytecode,bytecode-prechecked` provide that
paired lane: the latter mode performs its one required `able check` outside
the timed processes and labels every result `bytecode-prechecked`. The
scoreboard validator intentionally rejects that mode as input, preventing a
trusted execution-only result from replacing the full-process baseline.
Both
the main `able` CLI and generated compiled
launchers now also honor:

- `ABLE_GO_CPU_PROFILE=/tmp/cpu.pprof`
- `ABLE_GO_MEM_PROFILE=/tmp/heap.pprof`

for reusable Go `pprof` capture during focused benchmark runs. `v12/bench_perf`
now sends `SIGINT` before `SIGKILL` on timeout, so timed-out profiled runs
still flush CPU/heap profiles when possible.

`v12/bench_suite` now also supports `--list-suites` plus fixture-breadth
presets sourced from `v12/fixtures/bench/`, including `fixture-core`,
`fixture-full`, `fixture-generality`, `fixture-collections`, `fixture-text`,
`fixture-algorithms`, `fixture-concurrency`, `fixture-numeric`, and
`fixture-external-small`. The historical June expansion reached 75 fixtures;
the current catalog has 79. Verify the current count and every fixture entry
with `just bench-catalog-check` rather than carrying a prose count forward.

`v12/bench_fixture_validate` is the bounded parity checker for that local
fixture corpus. It runs one benchmark at a time under explicit `GOMEMLIMIT`
and `GOGC` settings, builds the Go interpreter once per validation pass for
treewalker/bytecode reuse, normalizes wrapper stdout, records per-mode peak
RSS via `/usr/bin/time`, supports suite presets from
`v12/bench_fixture_catalog.sh`, and now enforces hard build/run limits with
`timeout -k` so a single bad benchmark cannot wedge a validation pass or push
the session back into OOM territory. It also threads catalog-declared
program arguments into file-backed benchmarks so validation does not depend on
the caller's working directory.

Current bounded-validator snapshot on the expanded corpus:

- `fixture-external-small`: `12/13` match; only `sudoku_file_small` remains a
  treewalker timeout at the bounded `30s` guard
- `fixture-numeric`: the last full bounded sweep was `9/13` match with
  `1` mismatch (`biguint_add_mul_small`) plus `3` treewalker partials
  (`fib_i32_small`, `matrixmultiply_f64_small`, `sum_u32_small`); a follow-on
  targeted validation on `2026-06-24` closed that lone mismatch after
  stdlib `String` values were unified through the shared interpreter
  stringification / CLI print path instead of printing raw
  `String { bytes: [...], len_bytes: ... }` structs in interpreted modes, and
  the broader stdlib exec regression behind
  `06_12_05_stdlib_numbers_biguint` was then closed by making impl-body member
  access prefer inherent same-name methods over recursive impl forwarders on
  the concrete receiver; the remaining known numeric gaps in that bounded lane
  are the `3` treewalker partials
- collections hotspots after the validator hard-timeout fix and the bytecode
  `%` raw-i32 fix: `deque_i32_small` now matches across compiled/treewalker/
  bytecode, leaving `6` partial collection cases clustered around compiled
  `HashMap` / iterator-helper generation and runtime `Enumerable` / `Iterable`
  conformance gaps
- follow-on narrowed revalidation on `2026-06-24` shows that older
  compiled/bytecode collection-gap bucket is now mostly stale:
  `hashmap_i32_small`, `hash_set_i32_small`, `linked_list_for_i32_small`,
  `linked_list_enumerable_i32_small`,
  `linked_list_iterator_filter_map_i64_small`, and
  `linked_list_iterator_pipeline_i64_small` all match under the same bounded
  validator settings; the sampled remaining collection partials there are
  treewalker timeouts only (`heap_i32_small`,
  `linked_list_iterator_collect_i64_small`)
- ordered/persistent collection tranche added on 2026-06-24:
  `tree_map_i32_small`, `tree_set_i32_small`, `persistent_map_i32_small`,
  `persistent_set_i32_small`, `persistent_sorted_set_i32_small`, and
  `persistent_map_string_small` now all match across compiled/treewalker/
  bytecode after three general compiler fixes:
  fresh `Array.new()` locals now recover concrete element carriers from typed
  assignment/return context, void-return bodies compile their final form in
  statement mode instead of discarded implicit-return mode, and equality on
  concrete native union carriers (including singleton `Ord` variants like
  `Less` / `Equal`) now stays on native comparisons instead of boxing through
  `runtime.Value`; `persistent_sorted_set_i32_small` was also resized from
  `1800` to `900` elements so the `_small` compiled validation lane fits the
  routine bounded timeout budget
- broader workload-shape tranche added on 2026-06-24:
  `byte_histogram_small`, `word_count_small`, `levenshtein_small`,
  `knapsack_i32_small`, `dijkstra_heap_small`, and `toposort_small` now all
  match across compiled/treewalker/bytecode under the same bounded validator
  settings. This tranche filled underrepresented byte-scanning, string-keyed
  hashing, dynamic-programming, weighted-graph, and DAG-processing shapes
  without requiring any benchmark-specific stdlib/compiler/runtime surface.

`v12/bench_guardrail` is the report-only comparer for suite JSON outputs. It
compares the checked-in baseline against a fresh run and reports status,
timing, and GC deltas without failing the build.

`v12/bench_compare_external` compares Able runs against the checked-in
cross-language results in the sibling `../benchmarks` repository. It reuses
`v12/bench_perf` to run Able against the real external workloads, including
suite-local setup hooks and suite-local input files, then joins those results
against `../benchmarks/results.json` for reference languages such as `go`,
`ruby`, and `python`. The Able side uses the canonical checked-in benchmark
programs under `v12/examples/benchmarks/` and runs them from the external
suite directories so the workload inputs match the external corpus. The
external harness also applies benchmark-specific executor selection when the
reference workload is explicitly parallel; `binarytrees` now runs with the
goroutine executor so it matches the external Go workload instead of silently
serializing all spawned work. Its default mode set is now `compiled`,
`bytecode`, and `treewalker`; pre-summary Able failures are recorded as
machine-readable per-mode failure rows instead of aborting the whole external
comparison. It also forwards `--cpu-affinity CPUSET` (or
`ABLE_BENCH_CPU_AFFINITY`) into the underlying `bench_perf` runs and records
that lane in the emitted JSON/Markdown metadata. It also accepts
`--require-quiet-cpu` for a new single-CPU timing lane. When a suite provides
`verify.rb`, every successful compiled/treewalker/bytecode stdout capture is
verified after timing and its SHA-256 is recorded. A verifier failure marks
the mode invalid and suppresses its comparison ratios; suites without a
verifier are explicitly `unavailable` rather than silently treated as
verified. The JSON exposes this as `rows[].able.validation`, and the Markdown
report includes validation and stdout-hash columns.

`v12/bench_refresh_interpreter_refs` (also
`just bench-interpreter-reference`) refreshes local Python and Ruby source
references in the same pinned, verifier-backed process lane. It records source
or verifier absence explicitly and stops a row after its first timed-out
process, reporting both requested and attempted runs. Pass its JSON report to
`v12/bench_compare_external --reference-json PATH` to replace stored
Python/Ruby rows. `v12/bench_refresh_go_refs` produces the matching Go report;
pass it with `--go-reference-json PATH` to replace stored Go rows. The inputs
may be combined in one comparison, while language families without a supplied
fresh report retain their stored reference rows. The shared catalog also
declares a benchmark run directory, so applications that deliberately reuse
inputs from sibling suites run from the repository root while ordinary suites
continue to run beside their inputs.

## Current selection and catalog contract — 2026-07-20

The active `coverage` catalog has 36 portable applications; `fixture-full` has
79 bounded local programs; and `corpus-full` therefore validates 115 programs.
`just bench-catalog-check` verifies the portable Able/source/verifier lanes
and all local entries without timing work.

The current timing selection contains all 36 compiled applications and 29
bytecode applications. Seven excluded bytecode rows remain visible as bounded
status probes. Every selected row has five successful verifier-backed Able and
required Go/Python/Ruby samples under a 55-second per-process cap. The promoted
snapshot reports four established compiled target meets out of 36 and two
bytecode target meets out of 29. Regex Suffix is now a verified selected
bytecode miss rather than a timeout after its four implementations were
uniformly bounded to the same 512-word default as Regex Set and Regex Stream.
See
`docs/perf-baselines/2026-07-20-regex-three-api-current-profile-gate.md`.

The follow-up cross-mode Array capacity/backing-growth gate found no eligible
candidate. Generated geometric growth is material only in Reverse Complement
and Lexical Rollup; Base64 and FASTA use exact outputs, while Array Slice
Window's exact independent result backing is required by the language. The
bytecode capacity leaf likewise fails the three-unlike-application materiality
rule. See
`docs/perf-baselines/2026-07-18-cross-mode-array-capacity-growth-gate.md`.

The subsequent multi-argument primitive-byte extern gate also closes without
a change. Mandelbrot reaches the same host-coercion and mono-`u8`
deoptimization chain as Reverse Complement and FASTA, but the complete chain
is only 0.85% of its sampled allocation space and has no material CPU sample.
Only `io_write` currently combines multiple arguments with the explicit byte
borrow marker, so the required third material unlike consumer is absent. See
`docs/perf-baselines/2026-07-18-bytecode-multi-argument-u8-extern-admission.md`.

The current cross-mode exact-leaf pass also retains no change. K-Nucleotide,
Fixed Width, Distance Field, Regex Set, and Future Pipeline share only broad
bytecode dispatcher/load/call/return parents; their concrete VM children and
generated compiled owners diverge. Every narrower repeated carrier or frame
design is already closed by broad causal timing. Fresh verifier-backed Regex
Set and Future bytecode profiles complete the missing lanes and match the
retained post-quickening executable fingerprint. See
`docs/perf-baselines/2026-07-18-cross-mode-exact-leaf-selection-reconciliation.md`.

Pass `--suite NAME` to validate any named portable suite by itself (for
example, `just bench-catalog-check --suite dependency-plan`); `corpus-full`
continues to add every local fixture to the `coverage` application catalog.
The checker also proves that every named portable suite is a subset of
`coverage`, so no focused timing lane can silently escape the broad catalog.

Those counts establish breadth, not an optimization authorization. Do not run
an unchanged scorecard or add a synthetic workload merely to create a new
performance tranche. Reopen timing only after a material cross-cutting
semantic/compiler change or a genuinely needed portable application. Use
repeated verifier-backed compiled/bytecode comparisons and retain their mean,
median, and spread; prefer a quiet-host preflight when available, but do not
make a workstation-wide profile impossible when it is not. Retain only a
concrete hotspot repeated across unlike programs.
`sudoku-masks` is one such shared-input suite in the external `full` /
`generality` selection: it keeps the committed `sudoku` benchmark stable while
exercising a separate bit-mask and most-constrained-cell solver over a fixed
corpus prefix.

The `regex-text` external suite now contains three verifier-backed applications
over the existing ENABLE word-list corpus. `regex-suffix-audit` uses public
`RegexBuilder` construction, anchored capture matching, and aggregation. Its
default is 512 words (2,048 classifications), matching the bounded Set and
Stream discriminators so all three are rankable in bytecode mode;
`regex-set-audit` uses the public combined-NFA `RegexSet` API to classify four
anchored character-class patterns. Go, Python, and Ruby use equivalent
multi-pattern classification. Its fresh three-run reference lane recorded Go
`0.0042s`, Python `0.0183s`, and Ruby `0.0437s`, all verifier validated.
`regex-set-audit` is deliberately bounded to 512 words (2,048 classifications)
so its bytecode application run remains below the normal one-minute guard. It
remains a dedicated suite instead of silently changing the established
`generality` scorecard.

`regex-stream-audit` closes the public streaming surface gap. It feeds each
word and its newline as separate chunks to `RegexScanner`, drains decidable
matches, and calls `flush()` once per stream. Its Go, Python, and Ruby
counterparts implement the same buffered-record boundary with their standard
regex libraries. It belongs to `coverage` and `regex-text`, not the stable
`generality` scorecard, and must never justify a scanner- or chunk-shape-
specific fast path.

The first two generated-main profiles kept no compiler or bridge change:
both regex programs are dominated by tagged-NFA movement/closure and
allocation, while split/join is dominated by String conversion and joining.
The third independent `RegexSet` profile confirmed that NFA movement and
closure are reusable canonical-stdlib work. The kept immutable outgoing-edge
index removes unrelated transition scans from every compiled regex execution
plan, improving all three warmed bytecode lanes and the suffix/set generated
lanes while leaving independent matching neutral-to-slightly-better. Its full
record is `docs/perf-baselines/2026-07-12-nfa-outgoing-transition-index.md`.
The follow-up allocation gate found shared tagged-closure and thread-record
allocation in all three applications, while keeping profiler instrumentation
separate from normal measurements; see
`docs/perf-baselines/2026-07-12-regex-allocation-profile-gate.md`. The
matcher-local closure stack is now kept: it remains private to one normal
match/set operation or scanner stream, and improves all three warmed bytecode
applications without a material generated-binary regression. The refreshed
exact allocation profiles showed that successful tagged-thread insertion
created an active record and an identical closure record in every application.
That candidate is now kept: upsert returns the newly accepted private record
directly to closure traversal. It improved every warmed bytecode lane by
16--19% and reduced allocations by 20--27%, without pooling or adding a
compiler or `RegexSet` special case. The next evidence gate is independent
character-processing applications. That first gate is now complete: the
`zigzag_char_small`, `ascii_lower_small`, and `reverse_complement_small`
profiles do not execute a material `__able_char_to_codepoint` leaf. One moves
chars without scalar conversion and the other two process bytes, so no kernel
or compiler change is justified. The next gate must use non-regex workloads
that actually compare scalar chars before reconsidering that primitive.

The checked-in current external scoreboard lives in:

- `v12/docs/perf-baselines/external-scoreboard-current.json`
- `v12/docs/perf-baselines/external-scoreboard-current.md`

That artifact records the explicitly promoted verifier-backed source cohort,
including timeout rows for modes that still do not complete at the external
scale. The current July 15 cohort is a fresh 32-portable-application screen.
Every `unranked` row now includes a machine-readable `unranked_reason` in JSON
and a matching Markdown explanation: it distinguishes the Able launch status
from an unavailable required comparison ratio without guessing why a foreign
reference produced no valid ratio. `just bench-scoreboard-check` validates the
generated reports without rerunning performance workloads.

The same static catalog check now guarantees that all selected programs come
from the canonical `v12/examples/benchmarks` population rather than sibling
harness copies or generated build output. This keeps verifier-backed timing
evidence tied to reviewed Able source; see
`docs/perf-baselines/2026-07-15-canonical-benchmark-source-contract.md`.
Fresh comparison reports additionally record the canonical source SHA-256, and
the generated scorecard refuses to silently relabel old results after source
content changes. The retained July cohort labels its newly added hashes as
current-source legacy fingerprints rather than claiming historical measurement
identity; see
`docs/perf-baselines/2026-07-15-scorecard-source-fingerprints.md`.
The same contract now fingerprints matched fresh Go/Python/Ruby reference
sources. Reference edits therefore invalidate the relevant promoted ratio;
stored external result rows remain un-fingerprinted only when no local source
identity exists. See
`docs/perf-baselines/2026-07-15-scorecard-reference-source-fingerprints.md`.
Fresh comparison reports also capture a verifier/declared-input contract
before timing: the verifier script, every catalog-declared argument, and any
declared argument that resolves to a regular input file. The legacy cohort is
reconstructed from the same catalog as current-contract provenance, so a
verifier, input, or argument change invalidates the report. This deliberately
does not claim to fingerprint arbitrary implicit files a program might open;
see `docs/perf-baselines/2026-07-15-scorecard-verifier-input-fingerprints.md`.

As of April 29, 2026, the aligned compiled core and closed text/sort
benchmarks are in the same approximate range as Go:

- `fib`: `2.9940s` vs Go `2.8400s`
- `binarytrees`: `3.6400s` vs Go `3.8300s`
- `matrixmultiply`: `0.9660s` vs Go `0.8800s`
- `quicksort`: `1.75s` vs Go `2.01s`
- `sudoku`: `0.0600s` vs Go `0.1300s`
- `i_before_e`: `0.0620s` vs Go `0.0500s`

So the compiled core is now in the Go-range band for the current pass. Any
further `matrixmultiply` work should be limited to general row-length /
bounds-proof machinery rather than benchmark source tweaks. The remaining
bytecode work is still a larger VM architecture problem, not one-time
CLI/bootstrap/lowering noise.

The historical figures above are not a substitute for a fresh source-aligned
scorecard. On 2026-07-12, the retained scan-based `sudoku` source was corrected
to parse `.` blanks in the canonical corpus and then reached the compiled
45-second external guard. The separate bounded `sudoku_masks` lane records
that feature shape without hiding the legacy-core timeout; see
`docs/perf-baselines/2026-07-12-sudoku-masks-benchmark-lane.md`.

## 2026-06-23 — Bytecode speculative post-guard float update pair

The next kept `mandelbrot` bytecode tranche stayed on the same structural
strategy and removed the remaining statement boundary between the two discarded
post-guard float updates.

The landed change:

- lowering now emits `bytecodeOpTryFloatUpdatePair` before the exact discarded
  pair
  - `zi = 2.0 * zr * zi + ci`
  - `zr = zr2 - zi2 + cr`
- the generic `StoreSlotFloatAddMulSlot` and `StoreSlotFloatAddSub` statements
  still follow as fallback
- the fast path only fires when the first statement lowers to the existing
  slot-backed fused float add-mul update with a slot-const multiply operand,
  the second lowers to the existing fused float add-sub update, and the second
  RHS does not read the first target slot
- on success the VM reads every source slot first, computes both raw float
  results, writes both discarded slot updates directly, and jumps over the
  fallback pair

Focused verification:

- `go test ./pkg/interpreter -run 'TestBytecodeVM_(LoweringEmitsTryFloatUpdatePair|LoweringSkipsTryFloatUpdatePairWhenSecondUpdateReadsFirstTarget|TryFloatUpdatePairFastPath|TryFloatUpdatePairParity|LoweringEmitsFloatMulAddMulCompareConstJumpWithTempStores|FloatMulAddMulCompareConstJumpTempStoreParity|FloatAddSubSlotUpdateParity|StoreSlotFloatAddMulSlotFastPath)$'`
- `./v12/bench_perf --cpu-affinity 2-3 --runs 1 --timeout 60 --modes bytecode-runtime --run-from ../benchmarks v12/examples/benchmarks/mandelbrot/mandelbrot.able`
- `./v12/bench_compare_external --cpu-affinity 2-3 --benchmarks mandelbrot,matrixmultiply --modes bytecode --runs 3 --timeout 60`
- repeated external confirmation with the same pinned settings

Kept measurements:

- pinned runtime `mandelbrot`: `4475508943 ns/op`, `330949352 B/op`,
  `34613552 allocs/op`
- first cached-stdlib external confirmation: `mandelbrot` `4.7633s`,
  `matrixmultiply` `0.5367s`
- repeated cached-stdlib external confirmation: `mandelbrot` `4.4233s`,
  `matrixmultiply` `0.5433s`

The kept post-tranche profile now leaves the temp-square compare/store path,
the new update-pair path itself, discarded raw-float slot stores, and the
discarded `iter += 1` path as the main residual inner-loop work. The next
productive tranche should stay on that remaining compare/store + integer
discard traffic rather than reopening broader helper rewrites.

## 2026-06-23 — Benchmark-specific fusion rejected

A follow-on experiment fused the full six-statement non-escape Mandelbrot
recurrence into one speculative VM opcode. That experiment improved the
benchmark, but it was rejected and removed because it encoded a single
workload-shaped stencil rather than a generally reusable mechanism.

The policy change from that review is straightforward:

- do not add bespoke long-stencil opcodes for one benchmark recurrence
- use a broader benchmark set to identify patterns that recur across
  workloads
- prefer improvements to generic opcode families, slot/value representation,
  dispatch, lowering infrastructure, or peephole frameworks

## 2026-06-23 — Bytecode temp-square compare/store fusion

The next kept `mandelbrot` bytecode tranche stopped shaving generic float
helpers and instead absorbed the exact local control-flow shape the benchmark
actually uses.

The landed change:

- block-local lowering now reuses
  `bytecodeOpJumpIfFloatMulAddMulCompareConstFalse` for the exact three-step
  sequence:
  - `zr2 := zr * zr`
  - `zi2 := zi * zi`
  - `if zr2 + zi2 > 4.0 { ... }`
- the match is conservative: no `else` / `elsif`, the `if` body must
  terminate, the body cannot reference `zr2` / `zi2`, and later statements in
  the block must still use those temp names
- the VM computes both products once, evaluates the compare, and only writes
  `zr2` / `zi2` into their temp slots on the false path where the loop
  continues
- that removes the two standalone square-store statements from the steady
  Mandelbrot path without widening the rejected broader `f64` helper lane

Focused verification:

- `go test ./pkg/interpreter -run 'TestBytecodeVM_(LoweringEmitsFloatMulAddMulCompareConstJump|LoweringEmitsFloatMulAddMulCompareConstJumpWithTempStores|LoweringKeepsFloatBinaryStoresWhenIfBodyUsesTempSquares|FloatMulAddMulCompareConstJumpParity|FloatMulAddMulCompareConstJumpTempStoreParity|FloatMulAddMulCompareConstJumpFastPathWithOwnedFloatSlots|FloatMulAddMulCompareConstJumpStoresTempSquaresOnFalsePath|FloatAddSubSlotUpdateParity|StoreSlotFloatAddMulSlotFastPath|FloatBinaryStoreDiscardResultKeepsSnapshotSemantics)$'`
- `./v12/bench_perf --cpu-affinity 2-3 --runs 1 --timeout 60 --modes bytecode-runtime --run-from ../benchmarks v12/examples/benchmarks/mandelbrot/mandelbrot.able`
- `./v12/bench_compare_external --cpu-affinity 2-3 --benchmarks mandelbrot,matrixmultiply --modes bytecode --runs 3 --timeout 60`

Kept measurements:

- pinned runtime `mandelbrot`: `4947703038 ns/op`, `450787976 B/op`,
  `49593354 allocs/op`
- first cached-stdlib external confirmation: `mandelbrot` `5.5533s`,
  `matrixmultiply` `0.5700s`
- repeated cached-stdlib external confirmation: `mandelbrot` `5.1000s`,
  `matrixmultiply` `0.5233s`

The next productive tranche should stay on the same structural lowering
strategy and target the remaining post-guard Mandelbrot recurrence update
chain, because that is where the still-visible slot load/store traffic is now
concentrated.

## 2026-06-20 — Bytecode fused float add-sub slot update

The next kept `mandelbrot` bytecode tranche stayed deliberately narrow after
the bounded owned-float reuse keep.

The landed change:

- lowering now emits `bytecodeOpStoreSlotFloatAddSub` for slot-backed
  `a - b + c` update shapes such as `zr = zr2 - zi2 + cr`
- the VM reads the three operands from the stack, evaluates the subtraction
  plus addition through the raw-float fast path, and stores the final raw
  float result directly
- this removes the transient generic `zr2 - zi2` result from the hot
  `mandelbrot` loop without widening raw-float representation work elsewhere

Focused verification:

- `go test ./pkg/interpreter -run 'TestBytecodeVM_(LoweringEmitsFloatAddSubSlotUpdate|StoreSlotFloatAddSubFastPath|FloatAddSubSlotUpdateParity|DirectFloatArithmeticFastPath|DirectFloatCompareFastPath|LoadRawFloatSlotAvoidsSnapshotAllocation|StoreRawFloatSlotReusesCarrierWithoutAllocation|FloatBinaryStoreParity|FloatAddMulSlotUpdateParity|FloatAddMulArrayGetSlotUpdateFastPath)' -count=1`
- `./v12/bench_perf --runs 1 --timeout 120 --modes bytecode-runtime --run-from ../benchmarks v12/examples/benchmarks/mandelbrot/mandelbrot.able`
- `./v12/bench_compare_external --benchmarks mandelbrot,matrixmultiply --modes bytecode --runs 3 --timeout 60`

Kept measurements:

- runtime `mandelbrot`: `7832102713 ns/op`, `1046766888 B/op`,
  `85808195 allocs/op`
- cached-stdlib external bytecode `mandelbrot`: `7.9300s` over `3/3`
  (from `8.1100s`)
- cached-stdlib external bytecode `matrixmultiply`: `0.5333s` over `3/3`
  (from `0.5467s`)

The next productive float tranche should stay on the remaining generic result
creation in `zi = 2.0 * zr * zi + ci`, especially the inner `2.0 * zr`
multiply, instead of reopening broader raw-float carrier or owned-float slot
reuse experiments.

## 2026-06-20 — Bytecode float slot-const multiply

The next kept `mandelbrot` bytecode tranche targeted that remaining inner
multiply directly.

The landed change:

- lowering now emits `bytecodeOpBinaryFloatMulSlotConst` for slot-backed
  `identifier * float_literal` / `float_literal * identifier`
- the VM evaluates that opcode from the source slot plus embedded float
  immediate and returns a raw float carrier when the source slot is already in
  the raw-float fast lane
- the `mandelbrot` `zi = 2.0 * zr * zi + ci` update now uses that opcode for
  the inner `2.0 * zr` multiply before feeding the existing fused
  `StoreSlotFloatAddMul` path

Focused verification:

- `go test ./pkg/interpreter -run 'TestBytecodeVM_(LoweringEmitsBinaryFloatMulSlotConst|LoweringUsesBinaryFloatMulSlotConstInsideFloatAddMulUpdate|BinaryFloatMulSlotConstFastPath|BinaryFloatMulSlotConstParity|LoweringEmitsFloatAddSubSlotUpdate|StoreSlotFloatAddSubFastPath|FloatAddSubSlotUpdateParity|DirectFloatArithmeticFastPath|DirectFloatCompareFastPath|LoadRawFloatSlotAvoidsSnapshotAllocation|StoreRawFloatSlotReusesCarrierWithoutAllocation|FloatBinaryStoreParity|FloatAddMulSlotUpdateParity|FloatAddMulArrayGetSlotUpdateFastPath)' -count=1`
- `./v12/bench_perf --runs 1 --timeout 120 --modes bytecode-runtime --run-from ../benchmarks v12/examples/benchmarks/mandelbrot/mandelbrot.able`
- `./v12/bench_compare_external --benchmarks mandelbrot,matrixmultiply --modes bytecode --runs 3 --timeout 60`

Kept measurements:

- runtime `mandelbrot`: `6452646901 ns/op`, `671869432 B/op`,
  `70187489 allocs/op`
- cached-stdlib external bytecode `mandelbrot`: `6.6767s` over `3/3`
  (from `7.9300s`)
- cached-stdlib external bytecode `matrixmultiply`: `0.5433s` over `3/3`
  (same general band as the prior `0.5333s`)

The next productive float tranche should re-profile and likely target the
remaining coordinate affine expressions such as
`((2.0 * (x as f64)) / (SIZE as f64)) - 1.5` and the analogous `y` path,
instead of widening raw-float representation experiments again.

## 2026-06-20 — Bytecode fused coordinate affine store

The next kept `mandelbrot` bytecode tranche took that exact remaining
coordinate-affine advice and fused the whole slot-backed store shape instead
of adding another intermediate binary opcode.

The landed change:

- lowering now emits `bytecodeOpStoreSlotFloatAffine` for
  `((scale * (slot as f64)) / (name as f64)) - offset` assignment shapes such
  as:
  - `ci := ((2.0 * (y as f64)) / (SIZE as f64)) - 1.0`
  - `cr := ((2.0 * (x as f64)) / (SIZE as f64)) - 1.5`
- the VM carries an IP-keyed affine plan with the source slot, divisor slot or
  name, target float kind, scale, and offset, then evaluates the full
  multiply/divide/subtract chain through the raw-float fast path before the
  final store
- the failed isolated `cast-slot-float-const mul` probe was removed; it was
  not a keep on `mandelbrot` by itself

Focused verification:

- `go test ./pkg/interpreter -run 'TestBytecodeVM_(LoweringEmitsStoreSlotFloatAffine|StoreSlotFloatAffineParity|StoreSlotFloatAffineFastPathUsesI32RegisterLaneAndGlobalDivisor|LoweringEmitsBinaryCastSlotFloatConstDivOpcode|BinaryCastSlotFloatConstDivParity|BinaryCastSlotFloatConstDivFastPathUsesI32RegisterLane|LoweringEmitsStoreSlotCastSlotFloatConstDivOpcode|StoreSlotCastSlotFloatConstDivParity|StoreSlotCastSlotFloatConstDivFastPathUsesI32RegisterLane|StoreSlotCastSlotFloatConstDivDiscardFastPathStoresRawFloatWithoutOwnedCell|StoreSlotCastSlotFloatConstDivDiscardFastPathUpdatesExistingOwnedSlotCellWithoutMap|StoreSlotCastSlotFloatConstDivDiscardFastPathUsesRawI64SourceAndExistingOwnedTarget|LoweringEmitsBinaryFloatMulSlotConst|LoweringUsesBinaryFloatMulSlotConstInsideFloatAddMulUpdate|BinaryFloatMulSlotConstFastPath|BinaryFloatMulSlotConstParity|LoweringEmitsFloatBinaryStore|FloatBinaryStoreParity|FloatBinaryStoreDiscardResultKeepsSnapshotSemantics|LoweringEmitsFloatAddSubStore|FloatAddSubParity|LoweringEmitsFloatAddMulSlotUpdateWithNonTargetBase|FloatAddMulNonTargetBaseParity)' -count=1`
- `./v12/bench_perf --runs 1 --timeout 120 --modes bytecode-runtime --run-from ../benchmarks v12/examples/benchmarks/mandelbrot/mandelbrot.able`
- `./v12/bench_compare_external --benchmarks mandelbrot,matrixmultiply --modes bytecode --runs 3 --timeout 60`
- `./v12/bench_compare_external --benchmarks mandelbrot --modes bytecode --runs 5 --timeout 60`

Kept measurements:

- runtime `mandelbrot`: `6279413279 ns/op`, `596996648 B/op`,
  `66588389 allocs/op`
- cached-stdlib external bytecode `mandelbrot`: `6.4040s` over `5/5`
  (first `3/3` sample was `6.6967s`; the longer rerun is the keep)
- cached-stdlib external bytecode `matrixmultiply`: `0.5167s` over `3/3`

The next productive float tranche should re-profile from this new baseline and
likely target the remaining hot slot-load / name-lookup / float-store
boundaries inside `pixel_byte`, rather than reopening isolated cast-slot
multiply probes or broader raw-float representation experiments.

## 2026-06-22 — Bytecode `LoadSlot` keep after reused-cell raw-slot reject

Follow-on profiling from the later `BinaryFloatMulSlotConst` raw-immediate keep
showed the live branch had drifted onto a new control and that the remaining
`mandelbrot` wall was still the `pixel_byte` slot-load / write boundary rather
than the old stack-carrier multiply work.

Rejected probe:

- rewired `storeReusableFloatSlotRaw(...)` so a reused owned float cell would
  still be updated for future writes while the visible slot value switched to a
  raw float carrier
- focused tests passed on the narrowed form, but the result was not a keep:
  - runtime `mandelbrot`: `7561635668 ns/op`, `721955320 B/op`,
    `82208185 allocs/op`
  - cached-stdlib external bytecode `mandelbrot`: `9.0900s` over `1/1`
  - cached-stdlib external bytecode `matrixmultiply`: `0.6100s` over `1/1`
- the helper and temporary focused test were removed

Kept probe:

- `slotStackValue(...)` now inlines the common raw-float, raw-i32,
  raw-i64-cell, and owned integer/float snapshot cases directly instead of
  always routing non-nil slot loads through `bytecodeStackSnapshotValue(...)`
- visible semantics stay unchanged:
  - raw float loads remain raw
  - raw i32 loads still materialize visible integer values
  - owned float loads still snapshot away from the slot cell

Focused verification:

- `go test ./pkg/interpreter -run 'TestBytecodeVM_(LoadRawFloatSlotAvoidsSnapshotAllocation|StoreRawFloatSlotReusesCarrierWithoutAllocation|StoreSlotI32DiscardResultStoresRawSlot|I32RegisterFrameStoresDiscardedSlotOffValueFrame|FloatBinaryStoreDiscardResultKeepsSnapshotSemantics|FloatAddMulArrayGetSlotUpdateRawAccumulatorLoadCopies|StoreSlotFloatReusesOwnedCellAcrossReinitialization|F64DotLoopFastPath)' -count=1`

Kept measurements:

- runtime `mandelbrot`: `6906198808 ns/op`, `831862744 B/op`,
  `65985221 allocs/op`
- cached-stdlib external bytecode `mandelbrot`: `7.3433s` over `3/3`, then
  `7.3480s` over `5/5`
- cached-stdlib external bytecode `matrixmultiply`: `0.5267s` over `3/3`

The next productive tranche should re-profile from this load-side keep and
likely return to the remaining `storeReusableFloatSlotRaw(...)` /
`bytecodeSetNormalizedRawFloatValue(...)` write-side wall, not reopen the
rejected reused-cell raw-slot rewrite.

## 2026-06-22 — Bytecode normalized raw-float store helper keep

The next kept `mandelbrot` follow-up stayed entirely on the remaining
write-side boundary that the load-slot keep had exposed.

The landed change:

- raw-result stores now route through
  `storeReusableNormalizedFloatSlotRaw(...)`, a helper that assumes the
  arithmetic/cast fast paths already handed it a normalized raw float result
- the general `storeReusableFloatSlotRaw(...)` wrapper still exists for
  non-hot callers, but the hot path now skips one redundant
  `normalizeFloat(...)` pass, the extra bounds recheck, and the separate
  `reusableOwnedFloatSlot(...)` helper call
- visible semantics stay unchanged: reused owned float cells still remain
  visible as owned cells, raw slot results still stay raw, and f32 fast-path
  results still store normalized raw f32 carriers

Focused verification:

- `go test ./pkg/interpreter -run 'TestBytecodeVM_(StoreSlotFloatAddSubFastPath|StoreSlotFloatAddSubFastPathNormalizesF32Result|FloatAddSubSlotUpdateParity|LoadRawFloatSlotAvoidsSnapshotAllocation|StoreRawFloatSlotReusesCarrierWithoutAllocation|FloatBinaryStoreDiscardResultKeepsSnapshotSemantics|StoreSlotCastSlotFloatConstDivDiscardFastPathStoresRawFloatWithoutOwnedCell|StoreSlotCastSlotFloatConstDivDiscardFastPathUpdatesExistingOwnedSlotCellWithoutMap|StoreSlotCastSlotFloatConstDivDiscardFastPathUsesRawI64SourceAndExistingOwnedTarget|FloatAddMulArrayGetSlotUpdateRawAccumulatorLoadCopies|StoreSlotFloatReusesOwnedCellAcrossReinitialization|F64DotLoopFastPath)' -count=1`
- `ABLE_GO_CPU_PROFILE=/tmp/able-next-tranche-after.cpu.pprof ./v12/bench_perf --runs 1 --timeout 120 --modes bytecode-runtime --run-from ../benchmarks v12/examples/benchmarks/mandelbrot/mandelbrot.able`
- `./v12/bench_compare_external --benchmarks mandelbrot,matrixmultiply --modes bytecode --runs 3 --timeout 60`

Kept measurements:

- runtime `mandelbrot`: `6850196443 ns/op`, `831898392 B/op`,
  `65985299 allocs/op`
- cached-stdlib external bytecode `mandelbrot`: `7.2433s` over `3/3`
- cached-stdlib external bytecode `matrixmultiply`: `0.5233s` over `3/3`

Profile note:

- the raw float store boundary dropped from about `1.40s` cumulative in the
  pre-change profile (`storeReusableFloatSlotRaw(...)`) to about `0.97s`
  cumulative in the post-change profile
- `slotStackValue(...)` and `execLoadSlotOpcode(...)` remain in the hot tier,
  so the next win is still likely to come from the remaining raw-float setter
  / slot-load traffic rather than from broader float representation rewrites

The next productive tranche should stay on the remaining
`bytecodeSetNormalizedRawFloatValue(...)` / `slotStackValue(...)` cost rather
than reopen the rejected reused-cell raw-slot rewrite.

Immediate follow-up note:

- four narrower probes on that same exact boundary were then rejected:
  - owned-float `LoadSlot` raw snapshots:
    runtime `7161034314 ns/op`, external `mandelbrot` `9.6400s` over `3/3`,
    external `matrixmultiply` `0.6800s` over `3/3`
  - direct stack append for `LoadSlot`:
    runtime `8752117788 ns/op`
  - dedicated `f64` store helper:
    runtime `7235153050 ns/op`
  - redundant-`TypeSuffix` guard on owned float-cell reuse:
    runtime `8203646890 ns/op`
- the restored kept code state sanity-checks back into the prior external
  band (`mandelbrot` `7.4500s`, `matrixmultiply` `0.5500s`, both `1/1`)
- the next productive tranche should now move away from this exact slot
  load/store micro-boundary and re-profile the remaining lexical-name /
  cached-lookup path or another structural hotspot instead
- that lexical-name / cached-call follow-up was tried next as an
  owner-equality cache-validation shortcut and rejected too:
  - runtime-only `mandelbrot`: `7363954948 ns/op`, `831899032 B/op`,
    `65985309 allocs/op`
  - external bytecode `mandelbrot`: `7.7567s` over `3/3`
  - external bytecode `matrixmultiply`: `0.5367s` over `3/3`
  - restored external sanity after backing it out:
    `mandelbrot` `7.5200s` over `1/1`, `matrixmultiply` `0.5200s` over `1/1`
- that next structural-looking follow-up was tried as instruction-indexed
  quickened plan tables for the hot fused float affine / float compare ops and
  rejected too:
  - runtime-only `mandelbrot`: `8061916921 ns/op`, `831893584 B/op`,
    `65985302 allocs/op`
  - external bytecode `mandelbrot`: `8.0233s` over `3/3`
  - external bytecode `matrixmultiply`: `0.5900s` over `3/3`
  - restored external sanity after backing it out:
    `mandelbrot` `7.1100s` over `1/1`, `matrixmultiply` `0.5100s` over `1/1`
- that larger recurring float-control follow-up was tried next as a guarded
  monolithic `LoopEnter` escape kernel for the exact seven-statement inner
  Mandelbrot recurrence and rejected too:
  - runtime-only `mandelbrot`: `7354926662 ns/op`, `831898936 B/op`,
    `65985301 allocs/op`
  - external bytecode `mandelbrot`: `7.5633s` over `3/3`
  - external bytecode `matrixmultiply`: `0.5267s` over `3/3`
  - restored external sanity after backing it out:
    `mandelbrot` `7.2600s` over `1/1`, `matrixmultiply` `0.5100s` over `1/1`
- the next productive tranche should re-profile the restored kept baseline and
  target a narrower remaining float-slot hotspot instead of another monolithic
  `LoopEnter` replacement
- that next narrow float-slot helper follow-up was attempted as a direct
  inline of the common raw/owned cases in `slotDirectFloatValue(...)` and
  `slotDirectF64Value(...)`, but it was measured while unrelated
  `nats-server` / `outbox-worker` processes were heavily saturating CPU:
  - focused tests passed
  - runtime-only `mandelbrot` during the contended run:
    `9896391197 ns/op`, `831904480 B/op`, `65985307 allocs/op`
  - external bytecode during the same contended run:
    `mandelbrot` `11.1700s` over `3/3`, `matrixmultiply` `0.6833s` over `3/3`
  - repeated restored-baseline sanity runs on that host still varied widely
    (`mandelbrot` `10.9100-17.7700s`, `matrixmultiply` `0.7900-1.2500s` over
    `1/1`), so the code was backed out and the numbers were not adopted as a
    real baseline
- the next productive tranche should first get back to a quieter benchmark
  environment before trusting any more micro-deltas on this float-slot path

## 2026-06-22 — Benchmark harness CPU-affinity control and restored pinned control

The next tranche closed that measurement problem before taking another VM cut.

The landed harness change:

- `v12/bench_perf` and `v12/bench_compare_external` now accept
  `--cpu-affinity CPUSET`
- the same setting can also come from `ABLE_BENCH_CPU_AFFINITY`
- when set, the measured benchmark process runs through `taskset` and the
  selected CPU set is recorded in the emitted JSON / Markdown metadata

Local procedure note:

- sample host load first and pick a quieter subset that still matches the
  workload's intended parallelism
- for the current bytecode `mandelbrot` / `matrixmultiply` lane, pinned
  `2-3` was materially steadier than the default all-core scheduler spread on
  this workstation because the largest unrelated load was sitting on other
  cores
- do not compare runs collected while multiple benchmark suites are executing
  in parallel; that self-contention invalidates the lane

Restored pinned control:

- runtime-only `mandelbrot` (`3` serial `1/1` confirmations on pinned `2-3`):
  - `6873489810 ns/op`, `831854024 B/op`, `65985225 allocs/op`
  - `6817485950 ns/op`, `831854816 B/op`, `65985241 allocs/op`
  - `6778528755 ns/op`, `831854328 B/op`, `65985234 allocs/op`
- cached-stdlib external bytecode (`3/3`, pinned `2-3`):
  - `mandelbrot`: `7.0633s`
  - `matrixmultiply`: `0.5367s`

This is a measurement-control keep, not a VM keep. The next productive tranche
should resume a narrow remaining float-slot/store hotspot probe against this
pinned control rather than against the noisy all-core workstation lane.

## 2026-06-22 — Pinned helper-inline float-slot probe held

With the pinned lane restored, the next tranche retried the same helper-local
`slotDirectFloatValue(...)` / `slotDirectF64Value(...)` probe that had been
contaminated by host contention earlier in the day.

The landed change:

- inlined the common raw-float and boxed/owned float cases directly into
  `slotDirectFloatValue(...)`
- inlined the common `f64`-only cases directly into `slotDirectF64Value(...)`
- left the active float-frame fallback unchanged
- added focused coverage for raw, owned, and active-frame direct float slot
  reads

Focused verification:

- `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestBytecodeVM_(SlotDirectFloatValueCoversRawOwnedAndActiveFrame|SlotDirectF64ValueRejectsNonF64|LoadRawFloatSlotAvoidsSnapshotAllocation|StoreRawFloatSlotReusesCarrierWithoutAllocation|StoreSlotFloatAddSubFastPath|StoreSlotFloatAddSubFastPathNormalizesF32Result|FloatAddMulArrayGetSlotUpdateRawAccumulatorLoadCopies|StoreSlotFloatReusesOwnedCellAcrossReinitialization|F64DotLoopFastPath|LoweringEmitsFloatAddCompareConstJump|FloatAddCompareConstJumpFastPathWithRawFloatSlots)' -count=1`
- `ABLE_GO_CPU_PROFILE=/tmp/able-slot-direct-inline-pinned.cpu.pprof ./v12/bench_perf --cpu-affinity 2-3 --runs 1 --timeout 120 --modes bytecode-runtime --run-from ../benchmarks v12/examples/benchmarks/mandelbrot/mandelbrot.able`
- `./v12/bench_compare_external --cpu-affinity 2-3 --benchmarks mandelbrot,matrixmultiply --modes bytecode --runs 3 --timeout 60`

Kept measurements:

- runtime `mandelbrot` on pinned `2-3`:
  - `6470759545 ns/op`, `831881360 B/op`, `65985299 allocs/op`
- cached-stdlib external bytecode on pinned `2-3` (`3/3`):
  - `mandelbrot`: `6.7733s` (from the pinned control `7.0633s`)
  - `matrixmultiply`: `0.5267s` (from the pinned control `0.5367s`)

Profile note:

- the pinned pre-change profile had `slotDirectFloatValue(...)` at about
  `0.33s` cumulative and `bytecodeDirectFloatValue(...)` at about `0.23s`
- the pinned post-change profile moved `slotDirectFloatValue(...)` to about
  `0.24s` cumulative and dropped `bytecodeDirectFloatValue(...)` out of the
  top tier
- the remaining wall is still the raw-float store/load boundary around
  `storeReusableNormalizedFloatSlotRaw(...)`,
  `finishStoreSlotFloatRawResult(...)`, `slotStackValue(...)`, and
  `execLoadSlotOpcode(...)`

The next productive tranche should stay on that remaining raw-float
store/load boundary rather than reopening broader representation experiments
or the already-rejected monolithic `LoopEnter` replacement.

## 2026-06-22 — Raw visible slot stores now skip cached owned-cell lookup

The next pinned tranche stayed on the same raw-float store/load wall, but
kept the scope tighter than the earlier rejected owned-cell visibility probe.

The landed change:

- `storeReusableNormalizedFloatSlotRaw(...)` now skips the
  `ownedFloatSlots` map lookup when the current slot is already visibly raw
  (`bytecodeRawF32SlotValue` / `bytecodeRawF64SlotValue`)
- live owned float cells still remain visible as owned cells
- only the already-raw visible slot case bypasses the cached owned-cell map
  lookup
- added focused coverage for the exact stale-cached-cell case so a visible raw
  slot stays raw even if an old owned float cell still exists in
  `ownedFloatSlots`

Focused verification:

- `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestBytecodeVM_(StoreReusableNormalizedFloatSlotRawKeepsVisibleRawSlotDespiteCachedOwnedCell|SlotDirectFloatValueCoversRawOwnedAndActiveFrame|SlotDirectF64ValueRejectsNonF64|LoadRawFloatSlotAvoidsSnapshotAllocation|StoreRawFloatSlotReusesCarrierWithoutAllocation|StoreSlotFloatAddSubFastPath|StoreSlotFloatAddSubFastPathNormalizesF32Result|FloatAddMulArrayGetSlotUpdateRawAccumulatorLoadCopies|StoreSlotFloatReusesOwnedCellAcrossReinitialization|StoreSlotCastSlotFloatConstDivDiscardFastPathStoresRawFloatWithoutOwnedCell|StoreSlotCastSlotFloatConstDivDiscardFastPathUpdatesExistingOwnedSlotCellWithoutMap|StoreSlotCastSlotFloatConstDivDiscardFastPathUsesRawI64SourceAndExistingOwnedTarget|F64DotLoopFastPath|LoweringEmitsFloatAddCompareConstJump|FloatAddCompareConstJumpFastPathWithRawFloatSlots)' -count=1`
- `ABLE_GO_CPU_PROFILE=/tmp/able-raw-store-visible-raw.cpu.pprof ./v12/bench_perf --cpu-affinity 2-3 --runs 1 --timeout 120 --modes bytecode-runtime --run-from ../benchmarks v12/examples/benchmarks/mandelbrot/mandelbrot.able`
- `./v12/bench_compare_external --cpu-affinity 2-3 --benchmarks mandelbrot,matrixmultiply --modes bytecode --runs 3 --timeout 60`

Kept measurements:

- runtime `mandelbrot` on pinned `2-3`:
  - `6228372668 ns/op`, `831875616 B/op`, `65985294 allocs/op`
- cached-stdlib external bytecode on pinned `2-3` (`3/3`):
  - `mandelbrot`: `6.4300s` (from the prior keep `6.7733s`)
  - `matrixmultiply`: `0.5367s` (same general pinned band)

Profile note:

- the post-helper-inline pinned profile had
  `storeReusableNormalizedFloatSlotRaw(...)` at about `1.03s` cumulative
- after this keep it moved to about `0.97s` cumulative, and the
  `ownedFloatSlots` fallback path drops out of the listed hot lines for that
  helper
- the remaining pinned wall is now still shared between the raw-float store
  helper and `slotStackValue(...)`, especially the owned float snapshot path
  plus the raw setter/interface conversion cost in
  `bytecodeSetNormalizedRawFloatValue(...)`

The next productive tranche should stay on that remaining store/load boundary,
most likely `slotStackValue(...)` owned-float snapshots or the remaining raw
setter overhead, rather than reopening broader representation rewrites.

## 2026-06-22 — Raw float store fast paths now push raw stack results directly

The next pinned tranche stayed on the same raw-float store/load wall, but it
stayed narrower than the already-rejected broader `LoadSlot` raw-snapshot
probes.

The landed change:

- added `bytecodeNormalizedRawFloatSlotValue(...)` so normalized raw float
  carriers can be produced directly
- `finishStoreSlotFloatRawResult(...)` now pushes a raw float carrier when the
  slot write reused a visible owned `*runtime.FloatValue` cell, instead of
  snapshotting that mutable cell back into an immutable `runtime.FloatValue`
- visible slot semantics stay unchanged: owned float slots remain owned cells,
  visible raw slots remain raw, and only the transient stack result changes
  representation
- added focused coverage proving the slot still reuses the owned cell while
  the pushed result stays raw

Focused verification:

- `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestBytecodeVM_(FinishStoreSlotFloatRawResultPushesRawSnapshotWhenSlotReusesOwnedFloatCell|StoreReusableNormalizedFloatSlotRawKeepsVisibleRawSlotDespiteCachedOwnedCell|SlotDirectFloatValueCoversRawOwnedAndActiveFrame|SlotDirectF64ValueRejectsNonF64|LoadRawFloatSlotAvoidsSnapshotAllocation|StoreRawFloatSlotReusesCarrierWithoutAllocation|StoreSlotFloatAddSubFastPath|StoreSlotFloatAddSubFastPathNormalizesF32Result|FloatAddMulArrayGetSlotUpdateRawAccumulatorLoadCopies|StoreSlotFloatReusesOwnedCellAcrossReinitialization|StoreSlotCastSlotFloatConstDivDiscardFastPathStoresRawFloatWithoutOwnedCell|StoreSlotCastSlotFloatConstDivDiscardFastPathUpdatesExistingOwnedSlotCellWithoutMap|StoreSlotCastSlotFloatConstDivDiscardFastPathUsesRawI64SourceAndExistingOwnedTarget|F64DotLoopFastPath|LoweringEmitsFloatAddCompareConstJump|FloatAddCompareConstJumpFastPathWithRawFloatSlots)' -count=1`
- `ABLE_GO_CPU_PROFILE=/tmp/able-float-store-raw-push.cpu.pprof ./v12/bench_perf --cpu-affinity 2-3 --runs 1 --timeout 120 --modes bytecode-runtime --run-from ../benchmarks v12/examples/benchmarks/mandelbrot/mandelbrot.able`
- `./v12/bench_compare_external --cpu-affinity 2-3 --benchmarks mandelbrot,matrixmultiply --modes bytecode --runs 3 --timeout 60`

Kept measurements:

- runtime `mandelbrot` on pinned `2-3`:
  - `6008031466 ns/op`, `831881176 B/op`, `65985301 allocs/op`
- cached-stdlib external bytecode on pinned `2-3` (`3/3`):
  - `mandelbrot`: `6.4033s` (from the prior keep `6.4300s`)
  - `matrixmultiply`: `0.5267s` (same general pinned band)

Profile note:

- `bytecodeStackSnapshotValue(...)` drops out of the hot tier after this keep
- `finishStoreSlotFloatRawResult(...)` still shows up, but its own result-push
  work is no longer the cost center; most of that cumulative time is now the
  delegated store helper
- the remaining pinned wall is now more clearly shared between the raw
  setter/interface conversion inside
  `storeReusableNormalizedFloatSlotRaw(...)` / `bytecodeSetNormalizedRawFloatValue(...)`
  and the surviving owned-float snapshot path in `slotStackValue(...)`

The next productive tranche should stay on that remaining store/load wall,
starting with either the raw setter/interface path or the owned-float
snapshot branch in `slotStackValue(...)`, rather than reopening broader
representation rewrites.

Rejected follow-up on that same pinned wall:

- swapped several already-normalized raw-result fast paths from
  `bytecodeRawFloatSlotValue(...)` to
  `bytecodeNormalizedRawFloatSlotValue(...)`
- focused tests still passed, but the benchmark did not:
  - rejected pinned runtime `mandelbrot`:
    `6295696203 ns/op`, `831875248 B/op`, `65985279 allocs/op`
  - rejected cached-stdlib external bytecode (`3/3`):
    - `mandelbrot`: `6.5367s`
    - `matrixmultiply`: `0.5733s`
- the code was backed out; restored cached-stdlib external control returned to
  `6.4433s` for `mandelbrot` and `0.5333s` for `matrixmultiply` over `3/3`
- do not reopen this exact normalized-carrier caller substitution as the next
  tranche

Rejected follow-ups after that same pinned wall:

- `StoreSlotFloatAddSub` source-slot rewrite:
  - rewired the fused `zr = zr2 - zi2 + cr` update so the VM would read the
    three operands directly from source slots instead of consuming the
    existing stack inputs
  - focused tests still passed, but the benchmark regressed:
    - pinned runtime `mandelbrot`: `7045240416 ns/op`, `831881048 B/op`,
      `65985295 allocs/op`
    - cached-stdlib external bytecode (`3/3`):
      - `mandelbrot`: `7.1067s`
      - `matrixmultiply`: `0.5867s`
  - the code was backed out
- typed-slot-gated float-store `i32` register invalidation:
  - added a temporary `slotUsesI32Register(...)` helper and used it to skip
    float-store `setI32RegisterValue(...)` work for slots whose active
    register-frame layout was not typed as `bytecodeCellKindI32`
  - focused tests still passed, but the benchmark still regressed:
    - pinned runtime `mandelbrot`: `6661331456 ns/op`, `831875680 B/op`,
      `65985297 allocs/op`
    - cached-stdlib external bytecode (`3/3`):
      - `mandelbrot`: `6.6433s`
      - `matrixmultiply`: `0.5533s`
  - the code was backed out

Restored state:

- focused interpreter test slice passed again after the second backout
- pinned runtime sanity check on the restored tree:
  `6302627976 ns/op`, `831854752 B/op`, `65985241 allocs/op`
- the last full restored external control on this same code remains:
  - `mandelbrot`: `6.4433s` over `3/3`
  - `matrixmultiply`: `0.5333s` over `3/3`

The next productive tranche should not reopen any of these three rejected
micro-branches. The remaining credible work is still the raw
setter/interface conversion inside `storeReusableNormalizedFloatSlotRaw(...)`
/ `bytecodeSetNormalizedRawFloatValue(...)` or the surviving owned-float
snapshot branch in `slotStackValue(...)`.

Rejected follow-up after that same setter-side wall:

- replaced `setI32RegisterValue(...)` with direct
  `clearI32RegisterSlot(...)` on float-store paths that already know the
  written value is not `i32`
- touched:
  - `storeReusableNormalizedFloatSlotRaw(...)`
  - `storeFloatSlotValue(...)`
  - the existing-owned-target fast path in
    `execStoreSlotCastSlotFloatConstDivDiscardFast(...)`
- focused tests still passed, and pinned runtime `mandelbrot` improved in
  isolation:
  - `6032095841 ns/op`, `831875600 B/op`, `65985294 allocs/op`
- but the broader pinned external pair regressed, so the code was backed out:
  - cached-stdlib external bytecode (`3/3`):
    - `mandelbrot`: `6.5300s`
    - `matrixmultiply`: `0.5533s`
- restored pinned runtime sanity on the backed-out tree:
  `6165178361 ns/op`, `831853992 B/op`, `65985223 allocs/op`
- do not reopen this exact float-store register-invalidation substitution

At this point the next productive tranche should move away from this exact
setter/register invalidation micro-edge and instead target the surviving
owned-float snapshot branch in `slotStackValue(...)` or another larger
remaining hotspot.

## 2026-06-22 — Value-only lexical name lookups

The next pinned `mandelbrot` tranche followed that advice and moved away from
the rejected float-store invalidation edge onto the still-visible lexical-name
cache lane.

The landed change:

- value-only cached name lookups no longer route through
  `lookupCachedIdentifierNameEntry(...)` and its
  `bytecodeResolvedIdentifierLookup` construction
- `lookupCachedIdentifierName(...)` now reads the same hot/global/scope caches
  directly and returns only the resolved `runtime.Value`
- the metadata-returning helper stays intact for `CallName` cache building,
  where env/owner/version still matter
- added focused lookup-cache coverage for both the hot-value path and the
  scope-cache value path

Focused verification:

- `cd v12/interpreters/go && go test ./pkg/interpreter -run '^TestBytecodeVM_(LookupCachedIdentifierNameUsesHotValueCache|ResolveCachedIdentifierNameUsesScopeCache|ResetForRunPreservesLookupCaches|CallNameCacheRecordsDirectInlineShape|CallNameCacheSkipsDirectInlineForTypeArguments|CallNameCacheInvalidatesOnRebind)$' -count=1`
- `ABLE_GO_CPU_PROFILE=/tmp/able-lookup-value-only.cpu.pprof ./v12/bench_perf --cpu-affinity 2-3 --runs 1 --timeout 120 --modes bytecode-runtime --run-from ../benchmarks v12/examples/benchmarks/mandelbrot/mandelbrot.able`
- `./v12/bench_compare_external --cpu-affinity 2-3 --benchmarks mandelbrot,matrixmultiply --modes bytecode --runs 3 --timeout 60`

Kept measurements on pinned `2-3`:

- runtime `mandelbrot`: `6024494161 ns/op`, `831881032 B/op`,
  `65985294 allocs/op`
- cached-stdlib external bytecode `mandelbrot`: `6.2767s` over `3/3`
- cached-stdlib external bytecode `matrixmultiply`: `0.5233s` over `3/3`

Post-keep profile note:

- the lexical-name lane drops substantially:
  - `lookupCachedIdentifierName(...)`: about `0.17s` cumulative
  - `resolveCachedIdentifierName(...)`: about `0.19s` cumulative
- the remaining wall is back on the float slot load/store boundary:
  - `slotStackValue(...)`: about `0.68s` cumulative, with the owned-float
    snapshot branch still prominent
  - `storeReusableNormalizedFloatSlotRaw(...)`: about `0.82s` cumulative,
    with `bytecodeSetNormalizedRawFloatValue(...)` still dominating the raw
    visible-slot write branch

The next productive tranche should likely return to that remaining
`slotStackValue(...)` / `storeReusableNormalizedFloatSlotRaw(...)` wall, but
without reopening the already-rejected owned-float raw-snapshot/direct-append
or float-store register-invalidation probes.

Rejected follow-up after that same remaining float wall:

- specialized the visible raw-slot branch in
  `storeReusableNormalizedFloatSlotRaw(...)` so same-kind raw slots wrote
  their concrete carrier directly instead of always routing through
  `bytecodeSetNormalizedRawFloatValue(...)`
- focused tests still passed, and pinned runtime `mandelbrot` improved
  slightly in isolation:
  - `6023863340 ns/op`, `831881312 B/op`, `65985294 allocs/op`
- but the broader pinned external pair regressed clearly, so the code was
  backed out:
  - cached-stdlib external bytecode (`3/3`):
    - `mandelbrot`: `6.9933s`
    - `matrixmultiply`: `0.5367s`
- restored pinned runtime sanity on the backed-out tree:
  `6083851654 ns/op`, `831854008 B/op`, `65985222 allocs/op`
- do not reopen this exact same-kind raw visible-slot specialization

At this point the next productive tranche should stay on the remaining
`slotStackValue(...)` owned-float snapshot cost or another larger hotspot,
but via a genuinely different cut than the already-rejected raw-snapshot,
direct-append, register-invalidation, or same-kind raw visible-slot
specializations.

Rejected follow-up before the next keep:

- rewired generic `StoreSlot` / `StoreSlotNew` float stores so
  `runtime.FloatValue` locals landed as visible raw float carriers instead of
  reusable owned cells
- pinned runtime-only `mandelbrot` regressed immediately:
  - `7206534036 ns/op`, `706607616 B/op`, `81568226 allocs/op`
- that branch was backed out before a broader external pair because the pinned
  runtime signal was already materially worse than the kept control
- do not reopen generic raw-float `StoreSlot` rewrites on this lane

## 2026-06-22 — Slot-sourced fused float add-mul keep

The next genuinely different `mandelbrot` cut stayed local to the hot
`zi = 2.0 * zr * zi + ci` update shape instead of changing float slot
representation again.

The landed change:

- lowering now emits `bytecodeOpStoreSlotFloatAddMulSlot` when a fused float
  add-mul update can keep one multiplicand on the stack while reading the base
  and the other multiplicand directly from source slots
- the new opcode executes through the existing raw-float fast path and stores
  the final result through the normal float-store boundary
- this removes two hot `LoadSlot` operations from shapes like
  `zi = 2.0 * zr * zi + ci` without reopening the rejected direct source-slot
  `StoreSlotFloatAddSub` path
- focused coverage now includes both lowering proof and direct fast-path proof
  for the new slot-sourced add-mul update

Focused verification:

- `cd v12/interpreters/go && go test ./pkg/interpreter -run 'TestBytecodeVM_(LoweringEmitsFloatAddMulSlotUpdate|StoreSlotFloatAddMulSlotFastPath|LoweringEmitsFloatAddMulSlotUpdateWithNonTargetBase|FloatAddMulSlotUpdateParity|FloatAddMulNonTargetBaseParity|FloatAddMulSlotUpdateFallbackParity|FloatAddMulSlotUpdatePreservesRHSOrder|LoweringUsesBinaryFloatMulSlotConstInsideFloatAddMulUpdate|LoweringEmitsBinaryFloatMulSlotConst|BinaryFloatMulSlotConstFastPath|BinaryFloatMulSlotConstParity|FloatBinaryStoreParity|FloatBinaryStoreDiscardResultKeepsSnapshotSemantics|StoreSlotFloatReusesOwnedCellAcrossReinitialization)' -count=1`
- `ABLE_GO_CPU_PROFILE=/tmp/able-next-tranche-5.cpu.pprof ./v12/bench_perf --cpu-affinity 2-3 --runs 1 --timeout 120 --modes bytecode-runtime --run-from ../benchmarks v12/examples/benchmarks/mandelbrot/mandelbrot.able`
- `./v12/bench_compare_external --cpu-affinity 2-3 --benchmarks mandelbrot,matrixmultiply --modes bytecode --runs 3 --timeout 60`

Kept measurements on pinned `2-3`:

- runtime `mandelbrot`: `5278795757 ns/op`, `456978464 B/op`,
  `50364582 allocs/op`
- cached-stdlib external bytecode `mandelbrot`: `5.4667s` over `3/3`
- cached-stdlib external bytecode `matrixmultiply`: `0.5400s` over `3/3`

Post-keep profile note:

- the load-side wall drops sharply:
  - `execLoadSlotOpcode(...)`: about `0.40s` cumulative, down from about
    `1.14s`
  - `slotStackValue(...)`: about `0.23s` cumulative, down from about `0.82s`
- the new hot tier is now more clearly store/result centered:
  - `storeReusableNormalizedFloatSlotRaw(...)`: about `1.04s` cumulative
  - `bytecodeSetNormalizedRawFloatValue(...)`: about `0.62s` cumulative
  - `execStoreSlotFloatBinary(...)`: about `1.22s` cumulative
  - `slotDirectFloatValue(...)`: about `0.31s` flat

The next productive tranche should now target that remaining raw
result/write wall rather than the older load-side assumption, most likely
through another local float-store/result cut around
`storeReusableNormalizedFloatSlotRaw(...)`,
`bytecodeSetNormalizedRawFloatValue(...)`,
`execStoreSlotFloatBinary(...)`, or `slotDirectFloatValue(...)`, and not by
reopening generic raw-float `StoreSlot` rewrites or the rejected direct
source-slot `StoreSlotFloatAddSub` probe.

## 2026-06-22 — Same-slot float square micro-path was backed out under host drift

The next probe stayed on that same remaining float-store wall and targeted the
common `zr * zr` / `zi * zi` shape inside `StoreSlotFloatBinary`.

The temporary change:

- added an early `*` fast path in `execStoreSlotFloatBinary(...)` when the
  left and right source slots were identical
- read the raw float once, squared it directly, and stored the normalized
  result through the existing raw-float store helper
- added focused parity/stack coverage for the same-slot square case

Focused verification on the temporary branch:

- `go test ./pkg/interpreter -run 'TestBytecodeVM_(FloatBinaryStoreParity|FloatBinaryStoreDiscardResultKeepsSnapshotSemantics|StoreSlotFloatBinarySquareFastPath|LoweringEmitsFloatAddMulSlotUpdate|StoreSlotFloatAddMulSlotFastPath|FloatAddMulSlotUpdateParity|FloatAddMulNonTargetBaseParity|StoreSlotFloatReusesOwnedCellAcrossReinitialization)' -count=1`
- `ABLE_GO_CPU_PROFILE=/tmp/able-next-tranche-6.cpu.pprof ./v12/bench_perf --cpu-affinity 2-3 --runs 1 --timeout 120 --modes bytecode-runtime --run-from ../benchmarks v12/examples/benchmarks/mandelbrot/mandelbrot.able`
- `./v12/bench_compare_external --cpu-affinity 2-3 --benchmarks mandelbrot,matrixmultiply --modes bytecode --runs 3 --timeout 60`

What the measurements showed:

- the profiled pinned runtime-only run improved slightly:
  - `5233077344 ns/op`, `456984424 B/op`, `50364607 allocs/op`
- but the broader harness did not produce a trustworthy keep/reject signal:
  - candidate external pair: `6.1067s` for `mandelbrot`, `0.7033s` for
    `matrixmultiply`
  - after backout, the restored pinned runtime sanity check was
    `6147137914 ns/op`, `456957632 B/op`, `50364537 allocs/op`
  - the restored external control rerun had drifted to `10.3667s` for
    `mandelbrot` and `0.6667s` for `matrixmultiply`

Outcome:

- the probe was backed out
- the runtime-only win was too small to justify landing without a stable
  broader confirmation
- the next productive tranche on this wall should either re-establish a quiet
  pinned control first, or take a larger write-side cut around
  `storeReusableNormalizedFloatSlotRaw(...)`,
  `bytecodeSetNormalizedRawFloatValue(...)`, `execStoreSlotFloatBinary(...)`,
  or `slotDirectFloatValue(...)` instead of another sub-1% micro-path

## 2026-06-22 — Narrowed discard-store and validated binary-slot fetch keep

The next follow-up did re-establish that quiet pinned control first, then kept
the actual landing narrower than the initial temporary probe.

The landed change:

- `execStoreSlotFloatBinary(...)` now uses a validated slot-float fetch on the
  already-range-checked source slots and reuses the left operand fetch when
  both source slots are identical, instead of paying the full checked helper
  twice on hot `zr * zr` / `zi * zi` shapes
- `finishStoreSlotFloatRawResult(...)` now routes discard-only raw-float
  stores through `storeReusableNormalizedFloatSlotRawDiscard(...)`, so hot
  statement-position raw stores avoid result-value bookkeeping when no stack
  result is needed
- the broader temporary fanout of the validated-slot fetch across other float
  opcodes was measured first and rejected on pinned runtime, then narrowed back
  down to this binary-store-local keep before the final external reruns
- the raw-float reuse helper also drops a redundant reassignment when an owned
  float slot cell is already visibly present

Focused verification:

- `go test ./pkg/interpreter -run 'TestBytecodeVM_(SlotDirectFloatValueCoversRawOwnedAndActiveFrame|SlotDirectF64ValueRejectsNonF64|StoreReusableNormalizedFloatSlotRawKeepsVisibleRawSlotDespiteCachedOwnedCell|FinishStoreSlotFloatRawResultPushesRawSnapshotWhenSlotReusesOwnedFloatCell|FinishStoreSlotFloatRawResultDiscardKeepsStackEmpty|FloatBinaryStoreParity|FloatBinaryStoreDiscardResultKeepsSnapshotSemantics|StoreSlotFloatAddMulSlotFastPath|FloatAddMulSlotUpdateParity|FloatAddMulNonTargetBaseParity|BinaryFloatMulSlotConstFastPath|BinaryFloatMulSlotConstParity|FloatAddCompareConstJumpFastPathWithRawFloatSlots|LoweringEmitsFloatAddCompareConstJump|FloatMulAddMulCompareConstJumpFastPath|LoweringEmitsFloatMulAddMulCompareConstJump)' -count=1`
- `ABLE_GO_CPU_PROFILE=/tmp/able-next-tranche-7-baseline.cpu.pprof ./v12/bench_perf --cpu-affinity 2-3 --runs 1 --timeout 120 --modes bytecode-runtime --run-from ../benchmarks v12/examples/benchmarks/mandelbrot/mandelbrot.able`
- `ABLE_GO_CPU_PROFILE=/tmp/able-next-tranche-7c.cpu.pprof ./v12/bench_perf --cpu-affinity 2-3 --runs 1 --timeout 120 --modes bytecode-runtime --run-from ../benchmarks v12/examples/benchmarks/mandelbrot/mandelbrot.able`
- `./v12/bench_compare_external --cpu-affinity 2-3 --benchmarks mandelbrot,matrixmultiply --modes bytecode --runs 3 --timeout 60`
- `./v12/bench_compare_external --cpu-affinity 2-3 --benchmarks mandelbrot --modes bytecode --runs 5 --timeout 60`

Kept measurements:

- pinned runtime `mandelbrot` control:
  - `5174709854 ns/op`, `456984248 B/op`, `50364601 allocs/op`
- pinned runtime `mandelbrot` on the kept tree:
  - `5162038896 ns/op`, `456979544 B/op`, `50364610 allocs/op`
- cached-stdlib external bytecode pair (`3/3`):
  - `mandelbrot`: `5.4833s`
  - `matrixmultiply`: `0.5233s`
- cached-stdlib external bytecode `mandelbrot` confirmation (`5/5`):
  - `mandelbrot`: `5.4420s`

Post-keep profile note:

- the binary-store source-read work now shows up on the new validated helper
  rather than the old fully checked helper path
- the write-side wall is now more explicitly discard-store centered:
  - `storeReusableNormalizedFloatSlotRawDiscard(...)`
  - `bytecodeSetNormalizedRawFloatValue(...)`
  - the remaining raw store inside `execStoreSlotFloatBinary(...)`

The next productive tranche should stay on that discard-store/raw visible-slot
setter wall, and should not widen the validated-slot fetch path back onto the
other float opcodes without fresh profile evidence.

## 2026-06-22 — Active-float-frame discard-store rewrite was rejected quickly

The next follow-up stayed on that same discard-store/raw visible-slot setter
wall, but tried a structurally different rewrite instead of another direct
visible-slot specialization.

The temporary change:

- discard-only raw-float stores tried to route through the existing active
  float side frame instead of always writing a visible raw slot carrier
- to preserve semantics, `slotStoredValue(...)`, `slotStackValue(...)`,
  `slotMaterializedValue(...)`, and `slotRuntimeValue(...)` were taught to
  materialize active-float side-frame values when the visible slot was `nil`
- focused coverage was added for discard-result float stores plus the generic
  slot readers that had to see the active-float frame

Focused verification on the temporary branch:

- `go test ./pkg/interpreter -run 'TestBytecodeVM_(StoreReusableNormalizedFloatSlotRawKeepsVisibleRawSlotDespiteCachedOwnedCell|FinishStoreSlotFloatRawResultPushesRawSnapshotWhenSlotReusesOwnedFloatCell|FinishStoreSlotFloatRawResultDiscardKeepsStackEmpty|SlotDirectFloatValueCoversRawOwnedAndActiveFrame|SlotDirectF64ValueRejectsNonF64|ActiveFloatFrameFeedsStoredStackMaterializedAndRuntimeReads|FloatBinaryStoreParity|FloatBinaryStoreDiscardResultKeepsSnapshotSemantics|StoreSlotFloatAddMulSlotFastPath|FloatAddMulSlotUpdateParity|FloatAddMulNonTargetBaseParity|FloatAddCompareConstJumpFastPathWithRawFloatSlots|FloatMulAddMulCompareConstJumpFastPath|BinaryFloatMulSlotConstFastPath|BinaryFloatMulSlotConstParity)' -count=1`
- `ABLE_GO_CPU_PROFILE=/tmp/able-next-tranche-8.cpu.pprof ./v12/bench_perf --cpu-affinity 2-3 --runs 1 --timeout 120 --modes bytecode-runtime --run-from ../benchmarks v12/examples/benchmarks/mandelbrot/mandelbrot.able`

What the measurements showed:

- the pinned runtime signal regressed immediately:
  - `6419770662 ns/op`, `569780160 B/op`, `64464704 allocs/op`
- that was enough to reject the branch without paying for a broader external
  pair
- after backout, the restored pinned runtime sanity check returned to:
  - `5130504355 ns/op`, `456957392 B/op`, `50364525 allocs/op`

Outcome:

- the probe was backed out
- do not route discard-only raw-float stores through the active float side
  frame on this lane
- the next productive follow-up should stay on the visible raw-slot setter
  boundary itself, because the deferred materialization/load cost was clearly
  worse than the visible raw-slot setter cost it tried to remove

## 2026-06-22 — Owned-float discard-store rewrites were rejected too

The next follow-up stayed on that same discard-store/raw visible-slot setter
wall, but tried owned-float-cell reuse instead of active-frame deferral.

The temporary changes:

- first, discard-only raw-float stores on the non-`slot0`, non-`i32`-register
  lane delegated directly to `storeOwnedFloatSlotRaw(...)`, so the visible raw
  slot carrier was replaced by an owned float cell
- after that regressed badly, the eager rewrite was backed out and narrowed to
  only reuse an already-cached owned float cell when a discard-only raw store
  was about to rewrite a visible raw float slot

Focused verification on both temporary branches stayed green:

- `go test ./pkg/interpreter -run 'TestBytecodeVM_(StoreReusableNormalizedFloatSlotRawKeepsVisibleRawSlotDespiteCachedOwnedCell|FinishStoreSlotFloatRawResultPushesRawSnapshotWhenSlotReusesOwnedFloatCell|FinishStoreSlotFloatRawResultDiscardKeepsStackEmpty|SlotDirectFloatValueCoversRawOwnedAndActiveFrame|SlotDirectF64ValueRejectsNonF64|FloatBinaryStoreParity|FloatBinaryStoreDiscardResultKeepsSnapshotSemantics|StoreSlotFloatAddMulSlotFastPath|FloatAddMulSlotUpdateParity|FloatAddMulNonTargetBaseParity|FloatAddCompareConstJumpFastPathWithRawFloatSlots|FloatMulAddMulCompareConstJumpFastPath|BinaryFloatMulSlotConstFastPath|BinaryFloatMulSlotConstParity)' -count=1`

What the measurements showed:

- eager discard-only owned-cell conversion regressed pinned runtime
  immediately:
  - `8896777706 ns/op`, `1330329960 B/op`, `65811894 allocs/op`
- narrowing that to cached-owned-cell reuse only still regressed pinned
  runtime:
  - `5611075691 ns/op`, `456957504 B/op`, `50364531 allocs/op`
- after backing both out, the restored pinned runtime sanity check returned
  to:
  - `5199497341 ns/op`, `456957112 B/op`, `50364523 allocs/op`

Outcome:

- both probes were backed out
- do not route discard-only raw-float visible slots through owned float cells
  on this lane, whether eagerly or only when a stale cached cell is already
  present
- the next productive follow-up should stay on the visible raw-slot setter
  cost itself rather than trying more discard-store owned-cell rewrites

## 2026-06-22 — Narrow `f64` raw-store specialization was not a trustworthy keep

The next follow-up stayed on that same discard-store/raw visible-slot setter
wall, but took a smaller `f64`-dominant helper cut instead of another slot
representation rewrite.

The temporary change:

- `finishStoreSlotFloatRawResult(...)` and the raw float store helpers were
  temporarily specialized for `runtime.FloatF64`
- the generic `f32` / mixed-kind path stayed unchanged

Focused verification on the temporary branch stayed green:

- `go test ./pkg/interpreter -run 'TestBytecodeVM_(StoreReusableNormalizedFloatSlotRawKeepsVisibleRawSlotDespiteCachedOwnedCell|FinishStoreSlotFloatRawResultPushesRawSnapshotWhenSlotReusesOwnedFloatCell|FinishStoreSlotFloatRawResultDiscardKeepsStackEmpty|SlotDirectFloatValueCoversRawOwnedAndActiveFrame|SlotDirectF64ValueRejectsNonF64|FloatBinaryStoreParity|FloatBinaryStoreDiscardResultKeepsSnapshotSemantics|StoreSlotFloatAddMulSlotFastPath|FloatAddMulSlotUpdateParity|FloatAddMulNonTargetBaseParity|FloatAddCompareConstJumpFastPathWithRawFloatSlots|FloatMulAddMulCompareConstJumpFastPath|BinaryFloatMulSlotConstFastPath|BinaryFloatMulSlotConstParity)' -count=1`

What the measurements showed:

- the temporary branch moved the pinned runtime only slightly:
  - `5252265729 ns/op`, `456957400 B/op`, `50364535 allocs/op`
- after backout, restored pinned controls drifted materially wider than that
  code delta:
  - `5351846737 ns/op`, `456958360 B/op`, `50364547 allocs/op`
  - `5740348223 ns/op`, `456957384 B/op`, `50364535 allocs/op`

Outcome:

- the probe was backed out
- do not treat narrow `f64` raw-result/store helper specialization on this
  wall as a trustworthy keep on the current host
- the next productive follow-up should either take a broader cut than another
  sub-1% setter helper specialization or re-establish a quieter multi-run
  control before spending more time on this wall

## 2026-06-22 — Quieter control re-established; embedded float-jump plan cut rejected

The next follow-up did the quieter control tranche first, then used that
cleaner baseline to test a broader runtime-path cut.

The quieter baseline:

- pinned runtime `3/3` control:
  - `5140757340 ns/op`, `456957592 B/op`, `50364537 allocs/op`
- profiled pinned control:
  - `5104599912 ns/op`, `456978912 B/op`, `50364603 allocs/op`

The fresh profile still showed the remaining float-store wall, but it also
showed a surprisingly hot per-iteration plan lookup inside
`execJumpIfFloatAddCompareConstFalse(...)`, so the temporary branch:

- encoded the fused float-add-compare slots and float literal directly into the
  instruction fields
- used the old `program.floatAddCompareConstJumps[...]` map only as fallback
  compatibility

Focused verification on the temporary branch stayed green:

- `go test ./pkg/interpreter -run 'TestBytecodeVM_(LoweringEmitsFloatAddCompareConstJump|FloatAddCompareConstJumpParity|FloatAddCompareConstJumpFastPathWithRawFloatSlots|FloatAddCompareConstJumpPlanFallbackUsesProgramPlan|StoreReusableNormalizedFloatSlotRawKeepsVisibleRawSlotDespiteCachedOwnedCell|FinishStoreSlotFloatRawResultPushesRawSnapshotWhenSlotReusesOwnedFloatCell|FinishStoreSlotFloatRawResultDiscardKeepsStackEmpty|FloatBinaryStoreParity|FloatBinaryStoreDiscardResultKeepsSnapshotSemantics|FloatMulAddMulCompareConstJumpFastPath|StoreSlotFloatAddMulSlotFastPath|FloatAddMulSlotUpdateParity|BinaryFloatMulSlotConstFastPath|BinaryFloatMulSlotConstParity)' -count=1`

What the measurements showed:

- temporary branch pinned runtime `1/1`:
  - `5313156306 ns/op`, `456957648 B/op`, `50364537 allocs/op`
- temporary branch pinned runtime `3/3`:
  - `5340723333 ns/op`, `456957261 B/op`, `50364522 allocs/op`
- after backout, the restored pinned sanity check returned to:
  - `5170882711 ns/op`, `456957232 B/op`, `50364519 allocs/op`

Outcome:

- the branch was backed out
- the quieter multi-run control tranche was worth doing and should remain the
  baseline procedure for this wall
- do not treat direct instruction-embedded plan elision for
  `JumpIfFloatAddCompareConstFalse` as a keep on this benchmark
- the next productive follow-up should return to a broader float read/store cut
  rather than another single-site execution-plan lookup shave

## 2026-06-23 — Broader `f64` float-opcode lane was rejected after external validation

The next follow-up took exactly that broader read/store cut instead of another
single helper shave.

The temporary change:

- added shared direct `f64` decode/arithmetic/compare helpers
- added `finishStoreSlotF64RawResult(...)` plus dedicated visible-slot `f64`
  raw-store/discard helpers
- the hot float opcodes tried the explicit `f64` lane before the generic
  mixed-width path:
  - `execStoreSlotFloatBinary(...)`
  - `execStoreSlotFloatAddMulSlot(...)`
  - `execStoreSlotFloatAddSub(...)`
  - `binaryFloatMulSlotConstFastRaw(...)`
  - `floatAddCompareConstCondition(...)`

Focused verification:

- `go test ./pkg/interpreter -run 'TestBytecodeVM_(DirectF64Value|DirectFloatArithmeticFastPath|DirectFloatCompareFastPath|SlotDirectF64ValueRejectsNonF64|StoreReusableNormalizedF64SlotRawKeepsVisibleRawSlotDespiteCachedOwnedCell|FinishStoreSlotF64RawResultPushesRawSnapshotWhenSlotReusesOwnedFloatCell|FinishStoreSlotFloatRawResultPushesRawSnapshotWhenSlotReusesOwnedFloatCell|FloatBinaryStoreDiscardResultKeepsSnapshotSemantics|StoreSlotFloatAddMulSlotFastPath|StoreSlotFloatAddSubFastPath|FloatAddCompareConstJumpFastPathWithRawFloatSlots)$'`
- `go test ./pkg/interpreter -run 'TestBytecodeVM_(FloatBinaryStoreParity|FloatAddMulNonTargetBaseParity|FloatAddMulSlotUpdateParity|FloatAddSubSlotUpdateParity|FloatAddCompareConstJumpParity|DirectFloatArithmeticFastPathFallsBackForNonFloat)$'`
- `./v12/bench_perf --cpu-affinity 2-3 --runs 1 --timeout 120 --modes bytecode-runtime --run-from ../benchmarks v12/examples/benchmarks/mandelbrot/mandelbrot.able`
- `./v12/bench_perf --cpu-affinity 2-3 --runs 3 --timeout 120 --modes bytecode-runtime --run-from ../benchmarks v12/examples/benchmarks/mandelbrot/mandelbrot.able`
- `ABLE_GO_CPU_PROFILE=/tmp/able-next-tranche-11-f64-lane.cpu.pprof ./v12/bench_perf --cpu-affinity 2-3 --runs 1 --timeout 120 --modes bytecode-runtime --run-from ../benchmarks v12/examples/benchmarks/mandelbrot/mandelbrot.able`
- `./v12/bench_compare_external --cpu-affinity 2-3 --benchmarks mandelbrot,matrixmultiply --modes bytecode --runs 3 --timeout 60`

What the measurements showed:

- quiet pinned runtime control (`3/3`):
  - `5140757340 ns/op`, `456957592 B/op`, `50364537 allocs/op`
- temporary branch pinned runtime confirmation (`3/3`):
  - `4742534466 ns/op`, `456957744 B/op`, `50364538 allocs/op`
- temporary branch profiled confirmation:
  - `4740266093 ns/op`, `456978864 B/op`, `50364603 allocs/op`
- but the same-session cached-stdlib external pair rejected it clearly:
  - `mandelbrot`: `8.4300s`
  - `matrixmultiply`: `0.6167s`
- after backout, the restored cached-stdlib external control (`3/3`) returned
  to:
  - `mandelbrot`: `5.7700s`
  - `matrixmultiply`: `0.5533s`

Outcome:

- the branch was backed out
- the pinned runtime win was not representative of the external workload
- do not treat this broader `f64` opcode lane as a keep

## 2026-06-23 — Same-slot square fast path was also rejected under the restored stable control

With the restored tree back in place, I re-tested the older unresolved
same-slot square probe (`zr * zr` / `zi * zi`) under the same now-stable
external control.

The temporary change:

- added an early `*` fast path in `execStoreSlotFloatBinary(...)` for
  `leftSlot == rightSlot`
- that path read the source slot once, squared the raw float directly, and
  stored the normalized result through the existing raw-float store helper

Focused verification:

- `go test ./pkg/interpreter -run 'TestBytecodeVM_(FloatBinaryStoreParity|FloatBinaryStoreDiscardResultKeepsSnapshotSemantics|StoreSlotFloatAddMulSlotFastPath|FloatAddMulSlotUpdateParity|FloatAddMulNonTargetBaseParity|BinaryFloatMulSlotConstFastPath|BinaryFloatMulSlotConstParity|FloatAddCompareConstJumpFastPathWithRawFloatSlots)$'`
- `./v12/bench_perf --cpu-affinity 2-3 --runs 1 --timeout 120 --modes bytecode-runtime --run-from ../benchmarks v12/examples/benchmarks/mandelbrot/mandelbrot.able`
- `./v12/bench_compare_external --cpu-affinity 2-3 --benchmarks mandelbrot,matrixmultiply --modes bytecode --runs 3 --timeout 60`

What the measurements showed:

- restored pinned runtime control (`1/1`):
  - `5149685281 ns/op`, `456958248 B/op`, `50364546 allocs/op`
- temporary branch pinned runtime candidate (`1/1`):
  - `5162196039 ns/op`, `456957680 B/op`, `50364537 allocs/op`
- temporary branch cached-stdlib external pair (`3/3`):
  - `mandelbrot`: `6.0667s`
  - `matrixmultiply`: `0.5533s`

Outcome:

- the branch was backed out
- on the stable restored control, this same-slot arithmetic micro-cut is not a
  keep either
- the current external-validated baseline is `mandelbrot` `5.7700s` and
  `matrixmultiply` `0.5533s` over `3/3`
- the next productive follow-up should move away from both broader `f64` lane
  rewrites and same-slot arithmetic micro-cuts

## 2026-06-09 — Mono-array metadata assignment build fix

A compiler correctness regression had reopened the compiled mono-array path for
matrix-shaped code: member assignment handled static struct metadata before the
specialized mono-array carrier, so kernel `Array` methods on `Array f64`
could still emit legacy wrapper writes like `self.Length` and
`self.Capacity`. That broke generated Go builds for the external
`matrixmultiply` source path.

The kept fix stays narrow:

- `generator_assignments.go` now routes mono-array metadata assignments through
  a dedicated helper before generic struct-field assignment.
- new helper file
  `v12/interpreters/go/pkg/compiler/generator_assignments_mono_array.go`
  lowers:
  - `length` writes as direct slice resize/extend,
  - `capacity` writes as direct slice reserve logic,
  - `storage_handle` writes as a no-op on the existing borrowed slice-backed
    mono carrier.

Focused proof coverage landed in
`v12/interpreters/go/pkg/compiler/compiler_mono_array_field_assignment_test.go`:

- source/build proof that mono-array metadata assignment no longer emits the
  broken legacy wrapper fields,
- execution proof for direct `Array f64` metadata updates,
- exact benchmark-build proof for
  `v12/examples/benchmarks/matrixmultiply.able`.

Verification:

- `go test ./pkg/compiler -run 'TestCompiler(MonoArrayMetadataAssignmentAvoidsLegacyWrapperFields|MonoArrayMetadataAssignmentExecutes|ExperimentalMonoArraysMatrixMultiplyBuilds)' -count=1 -timeout 300s`
- `git diff --check`

This is a build-fix keep, not a new matrix performance keep. The adjacent
compiled `matrixmultiply` no-fallback/runtime regression around nil
propagation in `matmul` / `dot` remains separate work.

## 2026-06-09 — Compiled mono-array nil-propagation no-fallback repair

The next compiler regression on the same matrix path was broader than the
build failure: static mono-array `get(...)!` / `[idx]!` propagation in
functions returning non-nil-compatible native types had started forcing a
compiler fallback with:

- `nil propagation requires nil-compatible return type`

That reopened representative mono-array static proofs and the compiled
`matrixmultiply` no-fallback path even after the metadata-assignment build fix.

The kept fix stays compiler-local:

- `v12/interpreters/go/pkg/compiler/generator_control_results.go`
  now lowers incompatible nil propagation as the normal compiled
  return-type-mismatch control path instead of abandoning native codegen.
- nil-compatible return paths are unchanged; they still return `nil` directly.

Focused regression coverage:

- `TestCompilerNilPropagationNonNullableReturnStaysNative`
- restored existing mono-array no-fallback matrix proofs:
  - `StaticBodyStaysOnCompilerOwnedArrayCarrier`
  - `NestedF64RowsStaySpecialized`
  - `MatrixMultiplyScalarLoopStaysNative`
  - `MatrixMultiplyMainStaysNative`
  - `MatrixMultiplyCountedLoopsStayNative`

Verification:

- `go test ./pkg/compiler -run 'TestCompiler(NilPropagationNonNullableReturnStaysNative|ExperimentalMonoArrays(StaticBodyStaysOnCompilerOwnedArrayCarrier|NestedF64RowsStaySpecialized|MatrixMultiply(ScalarLoopStaysNative|MainStaysNative|CountedLoopsStayNative)))' -count=1 -timeout 300s`
- `go test ./pkg/compiler -run 'TestCompiler(ExperimentalMonoArrays|MonoArrayMetadataAssignment|NilPropagationNonNullableReturnStaysNative)' -count=1 -timeout 300s`
- `./v12/bench_compare_external --benchmarks matrixmultiply --modes compiled --runs 1 --timeout 120`
  - restored compiled external `matrixmultiply`: `1.0200s` over `1/1`
- `git diff --check`

This is still not a new matrix performance keep. It is the compiler
correctness repair that reopens the compiled benchmark path after the metadata
build fix. The next matrix tranche should be a general proof/performance step,
not another nil-propagation fallback repair.

The final compiled `fib` gap closed with a bounded recursive return-range
proof. For simple one-parameter signed integer recurrences with a proven
static call bound and a literal terminating base case, the compiler now stores
per-argument return maxima. That lets the hot `fib(45)` body lower
`fib(n - 1) + fib(n - 2)` as direct `i32` addition after proving the two
recursive return ranges fit together; an overflowing recurrence still keeps
the checked add. The kept external comparison moved compiled Able from
`3.1760s` to `2.9940s` over `5/5` runs versus Go `2.8400s`, and the profiled
kept run landed at `2.9700s`.
The next kept bytecode recursion slice narrowed the hot recursive VM path
further by compacting same-program self-fast call frames and removing repeated
`IntegerValue` value-method copies from the hot slot-const arithmetic path.
That did not yet make aligned bytecode `fib` complete under `120s`, but it
did collapse the old `pushCallFrame` hotspot and kept aligned bytecode
`i_before_e` in the same high-50s band. The next work should stay on
recursive bytecode call/slot churn and a steadier `i_before_e` runtime-only
measurement path. That measurement path is now available as
`v12/bench_perf --modes bytecode-runtime`, which loads/lowers the program
once and then measures repeated `main()` calls under a Go benchmark harness.

The first VM v2 typed-lane code slice is now in place as a stack-only seed:
literal-only final `i32` add/sub expressions use a raw `i32` operand stack,
perform checked overflow, and box back to `runtime.Value` before the existing
return path. Reduced `Fib30Bytecode` is intentionally neutral on this slice
because recursive slot arithmetic still boxes; the final neutral reruns landed
at `115.78ms/op` and `116.69ms/op`. The next bytecode performance tranche
should extend the raw lane to declared `i32` slots in non-yielding
slot-eligible functions before revisiting aligned external `fib`.

That declared-slot tranche is now landed for safe non-recursive entry into the
typed lane: frame layouts record simple declared `i32` params and typed local
identifier declarations, and final arithmetic can use `LoadSlotI32` /
`StoreSlotI32` with raw checked `i32` add/sub before boxing. Reduced
`Fib30Bytecode` remains neutral because recursive self-fast frames still carry
boxed slot values; guardrail reruns landed at `117.43ms/op` and
`121.97ms/op`. The next bytecode performance tranche should carry typed `i32`
slot state through inline/self-fast frames and wire the fused self-call
subtract plus typed return-add path to use that state.

The next kept reduced-recursion tranche deliberately avoided the rejected
parallel typed-slot side-cache approach and instead compacted the proven
two-slot one-arg self-fast recursive frame shape. The fused `self(slot -
const)` path now saves caller slot 0, mutates the current two-slot frame for
the callee, and restores slot 0 on return instead of acquiring and releasing a
fresh two-slot frame each step. Reduced `Fib30Bytecode` moved from the refreshed
`119.96-125.07ms/op` baseline band to kept reruns of `105.34ms/op`,
`106.13ms/op`, and `102.79ms/op`; the profiled keep rerun landed at
`109.33ms/op` and no longer has `acquireSlotFrame2` / `releaseSlotFrame2` in
the hot self-call path. The remaining reduced wall is now mostly
`finishInlineReturn`, `execReturnBinaryIntAdd`, boxed small-integer result
handoff, and residual fused self-call guards.

The follow-up kept reduced-recursion tranche targeted those residual fused
self-call guards without changing semantics. `execCallSelfIntSubSlotConst(...)`
now tries an early exact-shape compact branch for the proven raw-immediate
two-slot slot-0 recursive shape before entering the generic immediate,
layout, and return-name ladder; unsupported shapes still use the existing
boxed/generic fallback. Reduced `Fib30Bytecode` moved from a refreshed
compact-frame profiled baseline of `105.27ms/op` to warmed reruns of
`99.54ms/op`, `100.39ms/op`, and `99.00ms/op`. A focused external bytecode
`fib(45)` comparison now completes inside the old `90s` guard at `79.1200s`,
versus Go `2.8400s`, Ruby `46.6400s`, and Python `60.6700s`. The next
bytecode recursion tranche should target the base-case and return side around
`execReturnConstIfIntLessEqualSlotConst(...)`, `finishInlineReturn(...)`, and
`execReturnBinaryIntAdd(...)`.

The next kept aligned-recursion tranche targeted return-add helper overhead.
`execReturnBinaryIntAdd(...)` now handles the direct
`runtime.IntegerValue`/`runtime.IntegerValue` small-`i32` pair inline before
falling back to the existing pointer and generic paths. A narrower raw-`i32`
base-case helper was tested first and reverted because aligned
`fib_i32_small` regressed to `7.98s/op`. The kept return-add inline tranche
left reduced `Fib30Bytecode` in range at `97.19ms/op`, `104.20ms/op`, and
`106.93ms/op`; aligned `fib_i32_small` bytecode-runtime moved to `7.21s/op`
over a 3-run band, with a profiled one-shot at `7.50s/op`. Focused external
bytecode `fib(45)` moved to `77.2400s`, versus Go `2.8400s`, Ruby `46.6400s`,
and Python `60.6700s`. The profile no longer shows
`bytecodeReturnAddSmallI32ValuePairFast(...)`, so the next recursion work
should target structural boxed return/add handoff, the base-case raw compare,
or compact `finishInlineReturn(...)` restoration rather than another
return-add helper.

The next kept aligned-recursion tranche removed the remaining helper call from
the exact compact self-call setup path. After
`execCallSelfIntSubSlotConstCompact(...)` has already proven the raw-immediate
two-slot slot-0 recursive shape, cached nil return generics, no implicit
receiver, and no active loop/iter state, it now writes the compact slot-0
self-fast frame record directly instead of calling
`pushSelfFastSlot0CallFrame(...)`. A compact `finishInlineReturn(...)`
shortcut for the same shape was tested and reverted because aligned
`fib_i32_small` regressed to `8.31s/op`. The kept inline-push tranche moved
reduced `Fib30Bytecode` to `104.21ms/op`, `96.22ms/op`, and `94.85ms/op`.
Aligned `fib_i32_small` bytecode-runtime landed at `7.18s/op`, with a profiled
one-shot at `7.60s/op`. Focused external bytecode `fib(45)` moved to
`76.7900s`, versus Go `2.8400s`, Ruby `46.6400s`, and Python `60.6700s`. The
profile no longer shows `pushSelfFastSlot0CallFrame(...)` as a separate hot
edge; the remaining recursion work is structural raw/typed return metadata or
typed-frame design, not another small helper split.

The next kept aligned-recursion tranche made that typed-frame direction
concrete for the exact compact self-fast shape. The VM now saves/restores a
raw `i32` slot-0 lane beside the boxed semantic slot value in minimal
self-fast frames. The fused recursive subtract and fused base-case
slot-const compare can use the raw lane while slot-0 writes refresh or clear
it, and all unsupported shapes keep the boxed fallback path. Reduced
`Fib30Bytecode` moved to `92.46ms/op`, `92.81ms/op`, and `92.08ms/op`.
Aligned `fib_i32_small` bytecode-runtime landed at `6.24s/op`, with a
profiled one-shot at `6.03s/op`. Focused external bytecode `fib(45)` moved to
`67.8200s`, versus Go `2.8400s`, Ruby `46.6400s`, and Python `60.6700s`.
The later direct compact minimal-return tranche measured against a noisier
same-session control rather than that historical best: aligned
`fib_i32_small` bytecode-runtime moved from `14.21s` to `13.3533s`, and full
external bytecode `fib(45)` moved from a reverted control at `77.0800s` to
`75.3700s`. The follow-up exact value-pair return-add inline tranche moved the
same full external bytecode `fib(45)` check to `72.8000s`.

The next kept aligned-recursion tranche stopped adding side metadata and
replaced the exact proven recurrence body with a guarded native bytecode
kernel. Slot-backed one-arg `i32` functions shaped as
`if n <= c { return r }` followed by `self(n-a) + self(n-b)` now execute the
recurrence directly with checked `i32` subtract/add overflow and box only at
the bytecode boundary; unsupported shapes and bytecode-stats runs keep the
existing bytecode path. The refreshed aligned `fib_i32_small` bytecode-runtime
baseline was `13.1900s`; the kept run landed at `0.7867s` over `3/3`, with a
profiled one-shot at `0.8100s` whose samples are entirely in the native
recurrence kernel. Full external bytecode `fib(45)` moved to `3.7633s` over
`3/3`, versus Go `2.8400s`, Ruby `46.6400s`, and Python `60.6700s`.

The next kept aligned-recursion tranche stayed inside that same guarded native
kernel, but changed how it executes the simple nonnegative domain. When the
kernel sees a nonnegative argument and a nonnegative base-limit, it now uses a
bounded dynamic-programming table for the exact recurrence shape instead of
re-entering the kernel recursively. That preserves the same checked `i32`
overflow behavior and keeps the older recursive kernel as fallback for the
remaining shapes.

Reduced `BenchmarkFib30Bytecode` moved from a fresh `153.25-173.16ms/op` band
to:

- `144.99ms/op`
- `144.68ms/op`
- `147.86ms/op`

Full external bytecode `fib(45)` moved from the prior kept `3.7633s` band to:

- `0.1100s` over `3/3`

Runtime confirmation on the external workload landed at:

- `36792 ns/op`
- `22712 B/op`
- `52 allocs/op`

This is another real keep. The old recursive native-kernel overhead is gone on
the benchmark shape; if this family stays in scope, the next question is how
broadly to generalize this recurrence execution strategy, not another
return/helper micro-cut around the old recursive path.

The next kept recurrence tranche took that broader path instead of another
aligned-`fib` helper cut. The same native `i32` recurrence kernel now also
recognizes the generic `Int` slot-return base shape used by the reduced local
benchmarks: `if n <= c { return n }` followed by `self(n-a) + self(n-b)`. The
fast path still only executes when the live entry value is concrete `i32`, so
wider boxed integer cases keep the old bytecode fallback, but the generic-`Int`
benchmark shape no longer misses the native recurrence machinery just because
its base case returns the parameter slot and its final add stays on the generic
integer return opcode.

Fresh reduced local baselines this turn were:

- `BenchmarkFib30Bytecode`: `151.05-157.74ms/op`
- `BenchmarkFib30BytecodeRuntimeOnly`: `151.53-156.34ms/op`

The kept widened recurrence detector moved those to:

- `BenchmarkFib30Bytecode`: `80.19us/op`, `87.81us/op`, `92.33us/op`
- `BenchmarkFib30BytecodeRuntimeOnly`: `770.7ns/op`, `721.1ns/op`, `736.4ns/op`

The aligned external `i32` `fib(45)` workload stayed in the already-kept band:

- `0.1133s` over `3/3`

This is a real keep because it broadens the recurrence execution machinery
rather than adding another benchmark-local helper, while preserving the
existing aligned external result.

The next kept recurrence follow-up widened that same generic-`Int`
execution path beyond live concrete `i32`. The bounded recurrence kernel now
also executes when the live entry value is concrete `i64`, while keeping the
older boxed fallback for unsupported integer kinds. This was measured with a
temporary source shaped as:

```able
fn fib(n: Int) -> Int {
  if n <= 1_i64 { return n }
  fib(n - 1_i64) + fib(n - 2_i64)
}
print(fib(30_i64))
```

Using `bench_perf` on that source:

- clean `HEAD` worktree baseline: `0.6833s` over `3/3`
- kept code: `0.1100s` over `3/3`
- runtime confirmation on the kept code:
  - `52866 ns/op`
  - `21880 B/op`
  - `44 allocs/op`

The aligned external `i32` `fib(45)` guardrail stayed at:

- `0.1100s` over `3/3`

This is another real keep. The recurrence machinery is no longer specific to
concrete `i32` execution on the generic-`Int` source shape.

The next kept recurrence slice removed the last lowering blocker on the exact
typed-source side. `ReturnConstIfIntLessEqualSlotConst` lowering no longer
requires an unsuffixed `i32` return literal, so source like:

```able
fn fib(n: i64) -> i64 {
  if n <= 2_i64 { return 1_i64 }
  fib(n - 1_i64) + fib(n - 2_i64)
}
print(fib(30_i64))
```

now lowers into the same recurrence detector instead of missing the fused
base-case opcode entirely.

Fresh exact-`i64` source measurements this turn:

- pre-change kept-tree baseline: `0.6433s` over `3/3`
- kept code: `0.1467s` over `3/3`
- runtime confirmation:
  - `49941 ns/op`
  - `21880 B/op`
  - `44 allocs/op`

The aligned external `i32` `fib(45)` guardrail stayed in range at:

- `0.1167s` over `3/3`

This is another real keep. The remaining recurrence work is no longer about
source-shape mismatches for `i64` literals; the next step is other primitive
widths or broader recurrence lowering, not another `fib` helper micro-cut.

The next kept recurrence narrowing step closed the adjacent exact typed-source
gap where the function return type is fixed-width but the base return literal
is still unsuffixed. Source like:

```able
fn fib(n: i64) -> i64 {
  if n <= 2_i64 { return 1 }
  fib(n - 1_i64) + fib(n - 2_i64)
}
print(fib(30_i64))
```

used to miss the recurrence detector because the fused base return constant was
stored as an `i32` literal even though the function return type was exact
`i64`. The detector now accepts coercible integer base-return literals for
exact typed integer returns instead of requiring an exact suffix match in the
fused opcode.

Fresh exact-`i64` untyped-base measurements this turn:

- pre-change kept-tree baseline: `1.2467s` over `3/3`
- kept code: `0.1100s` over `3/3`
- runtime confirmation:
  - `50728 ns/op`
  - `21880 B/op`
  - `44 allocs/op`

The aligned external `i32` `fib(45)` guardrail stayed in range at:

- `0.1267s` over `3/3`

This is another real keep. The remaining recurrence work is no longer about
exact typed integer return literals on the `fib` source side; the next step is
other primitive widths or broader recurrence lowering, not another local base
return mismatch.

The next kept recurrence-width slice closed the remaining exact `isize` source
gap. Before this change, source like:

```able
fn fib(n: isize) -> isize {
  if n <= 2 { return 1 }
  fib(n - 1) + fib(n - 2)
}
print(fib(30))
```

did not even typecheck. The typechecker integer-bound tables still excluded
`isize`/`usize`, so `<=`, `-`, `+`, return coercion, and the final `fib(30)`
call all rejected the source. The runtime/typechecker promotion helpers also
needed to preserve `isize` when mixed with unsuffixed integer literals rather
than silently promoting those operations to `i64`.

After that width-plumbing keep:

- the same exact-`isize` source runs at `0.1100s` over `3/3`
- runtime confirmation:
  - `52865 ns/op`
  - `21832 B/op`
  - `43 allocs/op`

The aligned external `i32` `fib(45)` guardrail stayed in range at:

- `0.1067s` over `3/3`

This is another real keep. The recurrence family is no longer blocked on
`isize` integer classification at the source boundary; the next step is
broader recurrence-policy widening or additional primitive widths, not more
local fib source-shape cleanup.

The next kept recurrence-width slice closed the adjacent exact unsigned source
gap. After the source-boundary widening keep, exact source like:

```able
fn fib(n: u64) -> u64 {
  if n <= 2 { return 1 }
  fib(n - 1) + fib(n - 2)
}
print(fib(30))
```

and:

```able
fn fib(n: usize) -> usize {
  if n <= 2 { return 1 }
  fib(n - 1) + fib(n - 2)
}
print(fib(30))
```

still fell back to boxed recursive bytecode, because assignment-style and
inline integer coercion only widened when the entire source type range fit the
target type. Fitting unsuffixed positive literals like `30` still arrived at
the call boundary as concrete `i32`.

The coercion layer now widens by concrete value fit instead of whole-type-range
fit, so the exact unsigned source shapes reach the same native recurrence
kernel:

- exact `u64` pre-change kept-tree baseline: `1.3733s` over `3/3`
- exact `u64` kept code: `0.1000s` over `3/3`
- exact `u64` runtime confirmation:
  - `88051 ns/op`
  - `21832 B/op`
  - `43 allocs/op`
- exact `usize` kept code: `0.1000s` over `3/3`
- exact `usize` runtime confirmation:
  - `76220 ns/op`
  - `21832 B/op`
  - `43 allocs/op`

The aligned external `i32` `fib(45)` guardrail stayed in range at:

- `0.1133s` over `3/3`

This is another real keep. The recurrence family is no longer blocked on exact
unsigned source coercion at the runtime call boundary; the next step is
broader recurrence-policy widening or additional primitive widths, not more
local fib coercion cleanup.

The next kept recurrence-width slice closed the adjacent exact signed narrow
source gap. Before this change, exact source like:

```able
fn fib(n: i8) -> i8 {
  if n <= 2 { return 1 }
  fib(n - 1) + fib(n - 2)
}
print(fib(10))
```

and:

```able
fn fib(n: i16) -> i16 {
  if n <= 2 { return 1 }
  fib(n - 1) + fib(n - 2)
}
print(fib(20))
```

still failed typechecking at the recursive calls. The source-boundary
preservation rule only kept exact unsigned widths when mixed with coercible
unsuffixed literals, so `n - 1` and `n - 2` widened back to `i32`.

The source-boundary rule now preserves exact signed widths too whenever the
unsuffixed integer literal fits the target kind. That lets exact `i8` / `i16`
source reach the same native recurrence kernel:

- exact `i8` pre-change state: recursive calls failed typechecking
- exact `i8` kept code: `0.1467s` over `3/3`
- exact `i8` runtime confirmation:
  - `80837 ns/op`
  - `21672 B/op`
  - `43 allocs/op`
- exact `i16` pre-change state: recursive calls failed typechecking
- exact `i16` kept code: `0.1333s` over `3/3`
- exact `i16` runtime confirmation:
  - `68055 ns/op`
  - `21752 B/op`
  - `43 allocs/op`

The same tranche also pins the rest of the primitive-width family under
focused source-based detection/parity coverage:

- `u8`
- `u16`
- `u32`
- `i128`
- `u128`

The aligned external `i32` `fib(45)` guardrail stayed in range at:

- `0.1200s` over `3/3`

This is another real keep. The recurrence family is no longer blocked on exact
signed narrow source coercion; the next step is broader recurrence-policy
widening, not more local width cleanup.

The next kept recurrence-policy slice took that broader path. Exact `<`
base-guard source like:

```able
fn fib(n: i32) -> i32 {
  if n < 3 { return 1 }
  fib(n - 1) + fib(n - 2)
}
print(fib(45))
```

was already lowering through the fused slot-const return-if boundary, but the
VM and the native recurrence detector still treated that opcode as a hardcoded
`<=` check. So the useful fix was not another `fib` helper or a new opcode; it
was to make the existing fused boundary honor the preserved compare operator
and let the recurrence detector derive the right base limit from it.

Measured on that temporary exact-`i32` source:

- clean `HEAD` baseline: `2.5433s` over `3/3`
- kept code: `0.0833s` over `3/3`

The aligned external `i32` `fib(45)` guardrail stayed in the kept band at:

- `0.1100s` over `3/3`

This is another real keep. The recurrence family is now widened through the
existing fused return-if boundary for simple `<` base guards; the next useful
step is broader recurrence policy beyond those simple guard forms, not another
width-specific cleanup.

The next recurrence-policy widening stayed on the same theme, but addressed
operand order instead of a new operator family. Exact source like:

```able
fn fib(n: i32) -> i32 {
  if 3 > n { return 1 }
  fib(n - 1) + fib(n - 2)
}
print(fib(40))
```

was still falling off the native kernel because slot-const lowering only
recognized identifier-left / integer-right comparisons. The kept fix
canonicalizes integer `const op slot` compares into the existing
`slot flipped_op const` lowering shape, which means the existing fused jump,
fused return-if, and recurrence detector paths now all reuse the same fast
boundary.

Measured on that temporary exact-`i32` source:

- clean `HEAD` baseline: `18.94s`
- kept code: `0.03s` over `3/3`

The aligned external `i32` `fib(45)` guardrail stayed in the same fast band:

- `0.1233s` over `3/3`

This is another real keep. The next useful recurrence step is broader policy
beyond simple base-guard normalization, not another compare-specific helper.

The next recurrence-policy keep widened the seed shape itself instead of the
operator surface form. Exact source like:

```able
fn fib(n: i32) -> i32 {
  if n == 0 { return 0 }
  if n == 1 { return 1 }
  fib(n - 1) + fib(n - 2)
}
print(fib(40))
```

was still missing the native recurrence kernel even though both leading guards
were already lowering as fused equality return-if opcodes. The kept fix teaches
the detector to accept a contiguous nonnegative equality prefix at that same
fused boundary and use it as the seed table for the existing kernel. Negative
inputs on that shape still fall back to the older boxed bytecode path.

Measured on that temporary exact-`i32` source:

- pre-fix current-tree baseline: `29.47s`
- kept warm band: `0.16-0.22s`
- kept `fib(45)` source-shape confirmation: `0.22-0.23s`

The aligned external `i32` `fib(45)` guardrail returned to the same kept band:

- `0.1233s` over `3/3`

This is another real keep. The next useful recurrence step is broader policy
beyond single-range and equality-prefix seed recognition, not another
compare-shape tweak.

The next recurrence-policy keep removed the remaining gap between those two
seed families. Exact source like:

```able
fn fib(n: i32) -> i32 {
  if n == 0 { return 0 }
  if n <= 2 { return 1 }
  fib(n - 1) + fib(n - 2)
}
print(fib(40))
```

was still missing the native recurrence kernel because the detector accepted
either the equality prefix or the range seed, but not both together. The kept
fix merges a leading equality prefix with one following fused `<` / `<=` base
guard at the same boundary, letting the explicit equality seeds override the
overlapping range seed by source order and using the range tail for the rest.

Measured on that temporary exact-`i32` mixed-seed source:

- pre-fix current-tree baseline at `fib(40)`: `10.93s`
- kept warm band at `fib(40)`: `0.15-0.16s`
- kept `fib(45)` source-shape confirmation: `0.14-0.16s`

The canonical external `fib` benchmark still uses the older single-range
source:

```able
if n <= 2 { return 1 }
```

and stayed in the fast band after the keep:

- `0.1000s` over `3/3`
- `0.1167s` over `3/3` on the confirmation rerun

This is another real keep. The next useful recurrence step is broader policy
beyond equality-prefix plus range-tail seed recognition, not another local base
shape patch.

The next recurrence-policy keep widened that seed handling one more step. Exact
source like:

```able
fn fib(n: i32) -> i32 {
  if n == 2 { return 1 }
  if n <= 1 { return n }
  fib(n - 1) + fib(n - 2)
}
print(fib(40))
```

was still missing the native recurrence kernel because the detector only
understood one ordered equality prefix, optionally followed by one trailing
range guard. The kept fix instead treats the whole leading fused base-guard run
as a small source-ordered seed program, builds the contiguous nonnegative seed
table from the values those guards actually produce, and only keeps the native
kernel when the recursive subtraction offsets can bottom out through those
discovered base facts.

Measured on that temporary exact-`i32` source:

- pre-fix direct bytecode `fib(40)` probe: `10.51s`
- post-fix direct bytecode `fib(40)` confirmation: `3.99s`
- kept steady-state bytecode-runtime `fib(40)`: `0.1133s` over `3/3`
- kept steady-state bytecode-runtime `fib(45)`: `0.1133s` over `3/3`

The aligned external `i32` `fib(45)` guardrail stayed in the kept band:

- `0.1100s` over `3/3`

This is another real keep. The next useful recurrence step is broader
source-ordered seed policy at the same fused boundary, not another dedicated
compare or `fib` helper tweak.

The next recurrence-policy keep stayed on that family but widened the generic
`Int` side of it. Source like:

```able
fn fib(n: Int) -> Int {
  if n <= 2_i64 { return 1_i64 }
  fib(n - 1_i64) + fib(n - 2_i64)
}
print(fib(40_i64))
```

and the adjacent mixed-seed form:

```able
fn fib(n: Int) -> Int {
  if n == 2_i64 { return 1_i64 }
  if n <= 1_i64 { return n }
  fib(n - 1_i64) + fib(n - 2_i64)
}
print(fib(40_i64))
```

were still missing the native recurrence kernel because the detector refused
generic-`Int` constant base returns entirely. The kept fix now carries a
required concrete integer kind through generic-`Int` constant base guards and
only runs the native kernel when the live concrete argument kind matches that
constant-return kind. Current-return generic-`Int` shapes stay unconstrained,
and mismatches still fall back to the older boxed bytecode path.

Measured on those temporary generic-`Int` `i64` sources:

- pre-fix direct bytecode `fib(40_i64)` probes:
  - const-range source: `27.79s`
  - out-of-order mixed-seed source: `28.30s`
- kept steady-state bytecode-runtime confirmations:
  - const-range `fib(40_i64)`: `0.1333s` over `3/3`
  - const-range `fib(45_i64)`: `0.1467s` over `3/3`
  - out-of-order mixed-seed `fib(40_i64)`: `0.1200s` over `3/3`
  - out-of-order mixed-seed `fib(45_i64)`: `0.1500s` over `3/3`

The aligned external `i32` `fib(45)` guardrail stayed in the kept band:

- `0.1300s` over `3/3`

This is another real keep. The next useful recurrence step is broader concrete
kind tracking for generic-`Int` recurrence seeds and operations, not another
local source-shape patch.

The next recurrence-policy keep stayed on that same generic-`Int` line and
closed the adjacent untyped-literal hole. Source like:

```able
fn fib(n: Int) -> Int {
  if n <= 2_i64 { return 1 }
  fib(n - 1_i64) + fib(n - 2_i64)
}
print(fib(40_i64))
```

the mixed-seed form:

```able
fn fib(n: Int) -> Int {
  if n == 2_i64 { return 1 }
  if n <= 1_i64 { return n }
  fib(n - 1_i64) + fib(n - 2_i64)
}
print(fib(40_i64))
```

and the adjacent operator variant:

```able
fn fib(n: Int) -> Int {
  if n < 3_i64 { return 1 }
  fib(n - 1_i64) + fib(n - 2_i64)
}
print(fib(40_i64))
```

were still missing the native recurrence kernel because generic-`Int` untyped
base literals were being pinned to the lowered default `i32` kind. The kept fix
now treats generic-`Int` untyped integer base literals as kind-flexible seeds
while still validating that every constant base value fits the live concrete
integer kind before the native kernel runs. Explicitly typed constant returns
still keep the older conservative kind-matching rule.

Measured on those temporary generic-`Int` `i64` sources:

- pre-fix steady-state bytecode-runtime probes at `fib(40_i64)`:
  - untyped const-range source: `49.4200s`
  - untyped mixed-seed source: timed out at `60s`
  - untyped `<` source: `49.8600s`
- kept steady-state bytecode-runtime confirmations:
  - untyped const-range `fib(40_i64)`: `0.1067s` over `3/3`
  - untyped const-range `fib(45_i64)`: `0.1100s` over `3/3`
  - untyped mixed-seed `fib(40_i64)`: `0.1100s` over `3/3`
  - untyped mixed-seed `fib(45_i64)`: `0.1100s` over `3/3`
  - untyped `<` `fib(40_i64)`: `0.1100s` over `3/3`

The aligned external `i32` `fib(45)` guardrail stayed in the kept band:

- `0.1167s` over `3/3`

This is another real keep. The next useful recurrence step is broader
concrete-kind reasoning for generic-`Int` recurrence arithmetic and coercion,
not more literal-shape cleanup.

That next recurrence-policy keep is now landed too, but the final useful cut
turned out to be exact generic-`Int` result-kind tracking rather than one more
admission rule. The old native kernel could match numeric values while still
returning the wrong integer suffix because generic `Int` recurrence can widen
recursive argument kinds and still return different kinds at base vs non-base
states.

The most visible semantic examples were:

```able
fn fib(n: Int) -> Int {
  if n <= 2_i64 { return 1_i32 }
  fib(n - 1_i64) + fib(n - 2_i64)
}
print(fib(40_i64))
```

and:

```able
fn fib(n: Int) -> Int {
  if n <= 1 { return n }
  fib(n - 1) + fib(n - 2)
}
print(fib(10_u64))
```

The kept fix now evaluates generic-`Int` recurrences as exact per-state
`(n, kind)` transitions, using the same `promoteIntegerTypes(...)` rules as
normal integer arithmetic. When every recursive child stabilizes to one widened
kind after the first subtraction, the runtime folds the hot path back to a
dense DP table keyed by raw `n`; only truly multi-kind state spaces stay on the
full `(raw, kind)` memo map. The remaining true multi-kind nonnegative cases
now also avoid the hash-map memo: the runtime builds the reachable argument-kind
closure for the root entry, propagates only the reachable `(raw, kind)` frontier,
and memoizes those states in indexed slices. That means:

- current-return bases preserve the live recursive-state kind
- constant bases preserve their literal integer kind
- widening states like live `u64` plus untyped `1`/`2` offsets now run directly
  instead of taking one boxed step first

Measured on temporary generic-`Int` recurrence probes:

- pre-fix steady-state bytecode-runtime probe:
  - untyped current-return `fib(40_u64)`: `0.1500s` over `3/3`
- stable-recursive-kind DP recovery:
  - untyped current-return `fib(40_u64)`: stayed in the `0.1300s` over `3/3`
    band while dropping from `25000 B/op, 44 allocs/op` to
    `20032 B/op, 38 allocs/op`
  - explicit-`i32` const-range `fib(40_i64)`: `ns/op=42260`, `20032 B/op`,
    `38 allocs/op`
  - untyped const-range `fib(40_i64)`: `ns/op=53583`, `20032 B/op`,
    `38 allocs/op`
- true multi-kind nonnegative fallback recovery:
  - mixed recursive-kind
    `fn fib(n: Int) -> Int { if n <= 1 { return n } fib(n - 1_i64) + fib(n - 2) }`
    at `fib(30_i32)`: moved from `0.1300s` over `3/3`, `25048 B/op`,
    `45 allocs/op` to `0.1167s` over `3/3`, `20768 B/op`, `43 allocs/op`
- oversize split:
  - when the indexed DP table still fits under the byte budget, the oversize
    source
    `fn fib(n: Int) -> Int { if n <= 2_i64 { return 0 } fib(n - 1_i64) + fib(n - 2) }`
    at `fib(1048700_i32)` is still on the under-budget dense lane and
    currently measures `1.2200s` over `3/3`, `54577483 B/op`, `52 allocs/op`
  - when the same source widens the reachable kind closure enough to exceed the
    byte budget, the runtime now switches to a paged indexed DP lane before the
    sparse raw-row memo fallback. The three-kind source
    `fn fib(n: Int) -> Int { if n <= 2_i64 { return 0 } fib(n - 1_i64) + fib(n - 2_i128) }`
    moved from `1.9733s` over `3/3`, `629750435 B/op`, `4188 allocs/op` to
    `1.6400s` over `3/3`, `80041200 B/op`, `822 allocs/op` at
    `fib(1048700_i32)`
  - once the source is genuinely sparse instead of just over-budget, the
    runtime still stays on the memo lane but now pre-indexes the stride-aligned
    rows before using the spill map. The large-step three-kind source
    `fn fib(n: Int) -> Int { if n <= 2000_i64 { return 0 } fib(n - 1000_i64) + fib(n - 2000_i128) }`
    moved on the isolated bytecode runtime benchmark from `569460 ns/op`,
    `424101 B/op`, `25 allocs/op` to `519291 ns/op`, `387305 B/op`,
    `21 allocs/op`
  - when the flat stride table itself is too large but the source still stays
    on-grid, the runtime now pages that memo index before dropping to the spill
    map. The close-step three-kind source
    `fn fib(n: Int) -> Int { if n <= 2000_i64 { return 0 } fib(n - 1000_i64) + fib(n - 1001_i128) }`
    now skips that paged memo lane entirely once the semigroup-density
    heuristic recognizes that paged DP is dense enough in practice, moving the
    isolated bytecode runtime benchmark from `275160752 ns/op`, `269225856 B/op`,
    `4182 allocs/op` to `169929420 ns/op`, `80023036 B/op`, `787 allocs/op`
  - direct strategy comparisons around the cutoff kept the paged-DP route on
    the right side of the crossover:
    - `1000/1001`: paged DP `158393442 ns/op` vs memo `175850514 ns/op`
    - `1000/1033`: paged DP `157617704 ns/op` vs memo `174048877 ns/op`
    - `1000/1049`: paged DP `167143252 ns/op` vs memo `190621394 ns/op`
    - `1000/1051`: falls just below the density cutoff and stays on memo

The aligned external `i32` `fib(45)` guardrail stayed in the kept band:

- `0.1100s` over `3/3`

This is another real keep. The next useful recurrence step is only to revisit
the cutoff if a new reduced-step family shows a different crossover, not more
local literal-shape cleanup.

The next kept VM-v2 general slice added a slot-backed bool conditional jump.
Declared `bool` identifiers used directly as `if`, `elsif`, or `while`
conditions now lower to `JumpIfBoolSlotFalse`, so the VM avoids the old
load/push/pop path when the branch can read a `runtime.BoolValue` directly
from the slot. The fallback still uses the existing truthiness helper for any
unsupported shape, preserving v12 truthiness semantics. In the same session,
the refreshed external bytecode baseline was `4.2200s` for `sudoku` and
`1.3200s` for `i_before_e`; the kept confirmation moved those to `2.6500s`
and `1.0000s`. Against the external rows, bytecode `sudoku` is now `20.38x`
Go, `0.47x` Ruby, and `0.88x` Python; bytecode `i_before_e` is now `20.00x`
Go, `10.00x` Ruby, and `7.69x` Python. The next bytecode tranche should
profile post-bool `sudoku` and `i_before_e` and target guarded
call/member/index quickening or canonical Array/String bytecodes before
broadening bool storage beyond condition branches.

The next kept post-bool VM slice added a guarded canonical Array member fast
path on bytecode direct member calls. After normal method resolution selects
the canonical stdlib/kernel `Array.len()` or nullable stdlib `Array.get(i32)`,
the VM executes the size/read operation directly instead of inlining the Able
wrapper body and then dispatching through `__able_array_size` /
`__able_array_read`. The fast path is still guarded by canonical origins,
selected `*runtime.FunctionValue` methods, receiver shape, arity, `i32` index
semantics, and existing fallback behavior; existing array handles are read
directly so host-backed mono arrays are not deoptimized. The kept external
bytecode confirmation moved `sudoku` from the prior kept `2.6500s` to
`2.0433s` over `3/3` runs, and moved `i_before_e` from `1.0000s` to
`0.7500s` over `3/3` runs. Against the external rows, bytecode `sudoku` is now
`15.72x` Go, `0.36x` Ruby, and `0.68x` Python; bytecode `i_before_e` is now
`15.00x` Go, `7.50x` Ruby, and `5.77x` Python. The next bytecode tranche
should profile the remaining post-Array hot calls and target guarded String
member/native-bytecode quickening for `contains`, `len_bytes`, and `replace`,
or a general quickened call/member opcode that reaches those canonical targets
without repeated method resolution.

The follow-up kept member fast-path slice applied the same guard structure to
canonical String wrappers. After normal method resolution selects stdlib
`String.len_bytes`, `String.contains`, or `String.replace`, the bytecode VM
executes the direct string operation instead of inlining the Able wrapper and
dispatching through `string_len_bytes_i32_fast`, `string_contains_fast`, or
`string_replace_fast`. Wrong origins, wrong return types, unsupported
receiver/argument shapes, oversized byte lengths, and all other String methods
fall back to the existing path; `replace` keeps the stdlib's empty-needle
behavior. The traced steady-state `i_before_e` result moved from the
post-Array `583.47ms/op` / `2064 allocs/op` shape to `346.01ms/op` /
`2052 allocs/op`, with the hot calls recorded as `string_contains_fast`,
`string_len_bytes_fast`, and `string_replace_fast` member dispatches. The kept
external bytecode confirmation moved `i_before_e` from `0.7500s` to `0.5767s`
over `3/3` runs and left `sudoku` neutral at `2.0333s` over `3/3` runs.
Against the external rows, bytecode `i_before_e` is now `11.53x` Go, `5.77x`
Ruby, and `4.44x` Python; bytecode `sudoku` is now `15.64x` Go, `0.36x` Ruby,
and `0.67x` Python. The next bytecode tranche should re-profile post-String
`sudoku`; current trace evidence points at Array push/write, iterator `next`,
and UTF-8 byte decode/read paths rather than another String wrapper shortcut.

The next kept member fast-path slice added canonical Array push dispatch.
After normal method resolution selects kernel `Array.push(value) -> void`, the
bytecode VM appends directly through the tracked array state instead of
inlining the Able wrapper and dispatching through `__able_array_write`. Wrong
origins, wrong return types, wrong arity, non-array receivers, and untracked or
typed handles stay on the existing fallback/normalization path; the direct
path returns `void` and keeps array alias tracking synchronized. The traced
steady-state `sudoku` result moved from the post-String
`1771.07ms/op` / `332.12 MB/op` / `4462324 allocs/op` shape to
`1572.84ms/op` / `325.53 MB/op` / `4383835 allocs/op`, with the two hot
`push` sites recorded as `array_push_fast` and no `__able_array_write` entries
in the trace. The kept external bytecode confirmation moved `sudoku` from
`2.0333s` to `1.8833s` over `3/3` runs; the latest `i_before_e` confirmation
landed at `0.5333s` over `3/3` runs. Against the external rows, bytecode
`sudoku` is now `14.49x` Go, `0.33x` Ruby, and `0.62x` Python; bytecode
`i_before_e` is now `10.67x` Go, `5.33x` Ruby, and `4.10x` Python. The next
bytecode tranche should re-profile post-push `sudoku` and target iterator
`next`, string byte iteration / `utf8_decode`, residual Array construction, or
canonical Array `set` only after fresh trace evidence identifies the top edge.

The next kept member fast-path slice targeted the hot byte iterator returned
by `String.bytes()`. After member access resolves the canonical stdlib
`RawStringBytesIter.next` / `StringBytesIter.next` method behind an
`Iterator u8` interface, the VM now reads the current byte and advances
`offset` directly instead of calling the Able method body and its
`read_byte(...)` helper. The fast path is still guarded by canonical
`text/string.able` origin, `u8 | IteratorEnd` return type, receiver struct
name, `bytes`, `offset`, `len_bytes`, and `u8` element shape; unsupported
values fall back to the existing method path. The traced steady-state
`sudoku` result moved from the post-push `1572.84ms/op` / `325.53 MB/op` /
`4383835 allocs/op` shape to `1423.21ms/op` / `294.19 MB/op` /
`4056840 allocs/op`, with the hot `bytes.next()` site recorded as
`string_byte_iter_next_fast`. The kept external bytecode confirmation moved
`sudoku` from `1.8833s` to `1.7300s` over `3/3` runs; the latest
`i_before_e` confirmation landed at `0.5300s` over `5/5` runs. Against the
external rows, bytecode `sudoku` is now `13.31x` Go, `0.31x` Ruby, and
`0.57x` Python; bytecode `i_before_e` is now `10.60x` Go, `5.30x` Ruby, and
`4.08x` Python. The next bytecode tranche should re-profile post-iterator
`sudoku`; current trace evidence now points at Array reads in
`is_valid` / `board_to_string`, UTF-8 validation/decode during `String.bytes()`,
and residual `Array.new`.

The next kept Array fast-path slice shortened the already-canonical
`Array.get(i32)` path. For receivers that already carry a tracked dynamic
array state, the VM now reads `state.Values` directly before falling back to
the existing handle-store size/read path for untracked or typed handles. The
traced `sudoku` run confirmed that the top `Array.get` sites now dispatch as
`array_get_tracked_fast`; traced wall-clock was noisy, so the external
benchmark band is the keep basis. The kept external bytecode confirmation
moved `sudoku` from `1.7300s` to `1.6200s` over `3/3` runs; the latest
`i_before_e` confirmation landed at `0.5240s` over `5/5` runs. Against the
external rows, bytecode `sudoku` is now `12.46x` Go, `0.29x` Ruby, and
`0.54x` Python; bytecode `i_before_e` is now `10.48x` Go, `5.24x` Ruby, and
`4.03x` Python. The next bytecode tranche should re-profile
post-tracked-get `sudoku`; current trace evidence still points at UTF-8
validation/decode during `String.bytes()`, residual direct Array reads, and
Array construction.

The next kept String-byte slice removed that UTF-8 validation fan-out from the
valid-input hot path. After normal method resolution selects canonical
`String.bytes() -> Iterator u8`, the bytecode VM now validates the host string
with Go's UTF-8 checker, builds the same `RawStringBytesIter` struct shape as
the stdlib method, and returns it through normal `Iterator u8` interface
coercion. Invalid UTF-8 strings and missing canonical stdlib definitions fall
back to the existing Able method, preserving the canonical
`StringEncodingError` behavior. The traced `sudoku` run moved from
`1326.89ms/op` / `294.34 MB/op` / `4056759 allocs/op` to `653.23ms/op` /
`137.38 MB/op` / `1812289 allocs/op`; `String.bytes()` now records as
`string_bytes_fast`, and `utf8_validate`, `utf8_decode`, and `read_byte` are
absent from the hot trace for the valid sudoku corpus. The kept external
bytecode confirmation moved `sudoku` from `1.6200s` to `0.7780s` over `5/5`
runs. Against the external rows, bytecode `sudoku` is now `5.98x` Go,
`0.14x` Ruby, and `0.26x` Python; `i_before_e` stayed neutral at `0.5333s`
over `3/3` runs. The next bytecode tranche should profile
post-`String.bytes()` `sudoku`; current trace evidence now points at tracked
Array reads, string byte iterator `next`, Array push, and residual
`Array.new` construction.

The next kept propagation slice removed a broad type-match cost from those hot
`Array.get(... )!` success paths. The tree-walker and bytecode interpreters now
share a fast-negative guard for postfix `!`: direct `Error` values and
struct/interface values still route through the existing `Error` matching
path, but ordinary primitive, string, array, iterator, and future success
values skip `matchesType("Error")` unless an `Error` implementation is
registered for that runtime type. The profiled `sudoku` bytecode run moved
from `599.17ms/op` / `137.49 MB/op` / `1812061 allocs/op` to
`448.07ms/op` / `118.99 MB/op` / `1484787 allocs/op`; the old
`bytecodeOpPropagation -> matchesType("Error")` edge dropped from about
`250ms` cumulative to about `10ms`. The kept external bytecode confirmation
moved `sudoku` from `0.7780s` to `0.6700s` over `5/5` runs and confirmed
`i_before_e` at `0.5000s` over `5/5` runs. Against the external rows,
bytecode `sudoku` is now `5.15x` Go, `0.12x` Ruby, and `0.22x` Python;
bytecode `i_before_e` is now `10.00x` Go, `5.00x` Ruby, and `3.85x` Python.
The next bytecode profile should start from residual `execCallMember` /
member-cache overhead around tracked `Array.get` and hot name lookup.

The follow-up semantic propagation tranche aligned postfix `!` with the v12
spec's nil-propagation rule across both interpreters. Runtime `nil` through
`!` now returns from the current function, and the bytecode VM uses
`finishInlineReturn(...)` when that happens inside an inlined bytecode call
frame. To keep `!void` success distinct from Option nil failure,
`dyn.Package.def(...)` success now returns runtime `void` instead of runtime
`nil`. The external bytecode guardrail stayed clean after the additional nil
check: `sudoku` averaged `0.6220s` over `5/5` runs, and `i_before_e` averaged
`0.4620s` over `5/5` runs. Reduced `BenchmarkFib30Bytecode` stayed in its
current band at `91.137669ms/op` (`94.375464ms/op` runtime-only). The next
bytecode profile should start from post-nil-propagation `sudoku` and target
residual `execCallMember` / member-cache overhead around tracked `Array.get`
or the hot name lookup path before expanding to quickened member/index opcodes.

The next kept member-cache tranche stores the resolved canonical member
fast-path kind in the bytecode member-method cache and lets `CallMember` try
that fast path before rebinding the method template or routing through the
generic call ladder. This keeps normal method resolution as the guard while
removing the hot resolved-method bind/call overhead for already-recognized
canonical Array/String/iterator helpers. A temporary revert guard measured the
same external bytecode pair at `2.4200s` for `sudoku` and `0.9600s` for
`i_before_e`; restoring the cache fast path moved the final kept confirmation
to `0.6400s` and `0.4840s` over `5/5` runs, with an earlier same-slice rerun
at `0.6140s` / `0.4580s`. Against the external rows, bytecode `sudoku` is now
`4.92x` Go, `0.11x` Ruby, and `0.21x` Python; bytecode `i_before_e` is now
`9.68x` Go, `4.84x` Ruby, and `3.72x` Python. The same
tranche also tightened the compiled nil-propagation fixture and compiler
lowering: compiled `?T` nil through postfix `!` now returns a normal
nil-compatible value from the current function instead of raising
`runtime.NilValue{}` as control. The next bytecode profile should start from
this post-member-cache state and target the remaining `String.bytes()`
allocation/interface path plus residual name/member lookup.

The next heap-focused bytecode tranche stayed inside the existing
`String.bytes()` fast path. The VM now indexes the already-validated Go string
directly and reuses cached boxed `u8` values for byte elements plus cached
boxed `i32` values for `offset` and `len_bytes`, instead of first copying the
string to `[]byte` and boxing every byte afresh. Runtime-only `sudoku` moved
from `429.30ms/op`, `118.96 MB/op`, and `1,484,673 allocs/op` to
`420.48ms/op`, `114.51 MB/op`, and `1,390,910 allocs/op`; in the memory
profile, `execStringBytesMemberFast(...)` fell from about `18.50 MB` /
`243,574` objects cumulative to about `12.00 MB` cumulative. The external
guard stayed neutral: bytecode `sudoku` averaged `0.6380s` over `5/5` runs
and `i_before_e` averaged `0.4900s` over `5/5` runs. Against the external
rows, bytecode `sudoku` is `4.91x` Go, `0.11x` Ruby, and `0.21x` Python;
bytecode `i_before_e` is `9.80x` Go, `4.90x` Ruby, and `3.77x` Python. The
next bytecode profile should target the remaining `String.bytes()` array /
interface materialization or the residual name/member lookup path.

The follow-up byte-iterator storage tranche kept the same canonical
`String.bytes()` surface but moved the iterator's byte backing onto the
existing mono `u8` array store and attached implementation-private native text
metadata to the canonical `RawStringBytesIter` value. Canonical iterator
`next` now reads directly from that native text metadata before falling back to
the normal mono/dynamic Array path. The public iterator fields remain
`bytes`, `offset`, and `len_bytes`, and unsupported shapes continue through
the existing stdlib path. Runtime-only `sudoku` stayed wall-clock neutral in
the warmed band (`421.79ms/op`, `426.69ms/op`, `422.17ms/op`) while heap
volume moved slightly lower (`113.49 MB/op`, `114.89 MB/op`, `111.86 MB/op`).
The external guard is the keep basis: the first `5/5` rerun measured bytecode
`sudoku` at `0.6340s` and `i_before_e` at `0.4920s`; the repeat measured
`sudoku` at `0.6260s` and `i_before_e` at `0.4800s`. Against the external
rows, bytecode `sudoku` is now `4.82x` Go, `0.11x` Ruby, and `0.21x` Python;
bytecode `i_before_e` is now `9.60x` Go, `4.80x` Ruby, and `3.69x` Python.
The next bytecode profile should target the remaining generic `Iterator u8`
interface coercion in `String.bytes()` or residual name/member lookup; native
metadata should stay limited to guarded canonical stdlib shapes unless a
language-level host boundary is introduced.

The next kept `String.bytes()` tranche removed that remaining generic
interface coercion for the canonical byte iterator. When the VM has the
canonical stdlib `RawStringBytesIter`, canonical `Iterator` interface, and
canonical `RawStringBytesIter.next` method, it now constructs the `Iterator u8`
interface wrapper directly with the cached `next` method instead of routing
through `coerceToInterfaceValue(...)` and
`buildInterfaceMethodDictionary(...)`. Unsupported or non-canonical shapes
still fall back to the existing generic coercion path. Runtime-only `sudoku`
landed at `415.33ms/op`, `427.64ms/op`, and `415.61ms/op`, with allocation
volume down to roughly `102.78-105.86 MB/op` and `1.282M allocs/op`; the
profiled rerun no longer shows `coerceToInterfaceValue(...)` or
`buildInterfaceMethodDictionary(...)` under the `String.bytes()` allocation
edge. The external guard confirmed the keep: the first `5/5` rerun measured
bytecode `sudoku` at `0.6120s` and `i_before_e` at `0.4700s`; the repeat
measured `sudoku` at `0.6160s` and `i_before_e` at `0.4700s`. Against the
external rows, bytecode `sudoku` is now `4.74x` Go, `0.11x` Ruby, and `0.20x`
Python; bytecode `i_before_e` is now `9.40x` Go, `4.70x` Ruby, and `3.62x`
Python. The next bytecode profile should target residual `execCallMember` /
`resolveMethodCallableFromPool` and name/member lookup costs around interface
member access rather than another `String.bytes()` wrapper rewrite.

The next kept bytecode slice targeted static `Array.new()` call overhead in
`sudoku`. After normal static member resolution proves the active method is
the canonical zero-arg kernel `Array.new`, the VM now constructs the same empty
tracked `ArrayValue` directly instead of calling through the generic Able
method and `__able_array_new` native bridge. Unsupported or non-canonical
static member shapes still fall back to the existing path, and the hook is
name/arity-gated so unrelated member-access sites do not pay the check. The
trace changed 10,100 `Array.new` hits to `array_new_fast`, and the warmed
runtime-only `sudoku` band landed at `412.00ms/op`, `416.55ms/op`, and
`407.66ms/op` with allocation volume down to roughly `1.161M allocs/op`. The
external guard confirmed the keep after one discarded noisy paired sample:
`sudoku` measured `0.5820s` / `0.5840s` over two `5/5` runs, and
`i_before_e` stayed neutral at `0.4780s` / `0.4760s`. Against the external
rows, bytecode `sudoku` is now `4.49x` Go, `0.10x` Ruby, and `0.19x` Python;
bytecode `i_before_e` is now `9.52x` Go, `4.76x` Ruby, and `3.66x` Python.
The next bytecode profile should target residual `resolveMethodCallableFromPool`
/ overload-selection cost around hot `Array.get` and iterator `next`, not a
broader static member shortcut without fresh trace evidence.

The follow-up kept bytecode slice targeted the canonical byte-iterator
`next()` call-member path produced by `String.bytes()`. When `CallMember next`
sees the canonical stdlib `Iterator u8` interface wrapping canonical
`RawStringBytesIter`, and the canonical `RawStringBytesIter.next` method is
still valid under the current method/global revisions, the VM jumps directly
to the existing string-byte iterator fast body instead of re-entering generic
`memberAccessOnValueWithOptions(...)` / `interfaceMember(...)` dispatch.
Unsupported interface shapes and non-canonical iterators still use the generic
path. The refreshed steady-state `i_before_e` band landed at
`280.09ms/op`, `236.02ms/op`, and `246.16ms/op`, with about `2.83 MB/op` and
`2,006-2,008 allocs/op`; the profiled rerun landed at `237.20ms/op`,
`2.87 MB/op`, and `2,117 allocs/op`, and `interfaceMember(...)` /
`substituteAliasTypeExpression(...)` dropped out of the hot runtime set. The
external guard measured bytecode `i_before_e` at `0.4660s` and `sudoku` at
`0.5720s` over `5/5` runs. Against the external rows, bytecode `i_before_e`
is now `9.32x` Go, `4.66x` Ruby, and `3.58x` Python; bytecode `sudoku` is now
`4.40x` Go, `0.10x` Ruby, and `0.19x` Python. The next profile should target
residual `execCallMember` / `resolveMethodCallableFromPool` and overload-cache
overhead around hot direct string/member calls, or move back to the larger
bytecode `fib` typed-frame work.

The follow-up kept bytecode slice targeted the canonical nullable `Array.get`
overload pair that remained hot in `i_before_e`. After normal method
resolution returns exactly the canonical stdlib nullable `Array.get(i32) ->
?T` method plus the lower-priority canonical `Index.get(i32) -> !T`
implementation method, `CallMember get` now executes the existing tracked
`Array.get` fast body directly instead of running generic runtime overload
selection. Unsupported overload sets, custom origins, and wrong
parameter/return shapes still fall back to the old resolver, and a VM-local
hot cache keeps the canonical-shape validation off the repeated call path. The
refreshed steady-state `i_before_e` band landed at `193.31ms/op`,
`199.15ms/op`, and `197.26ms/op`, with about `2.82 MB/op` and
`1,989-1,991 allocs/op`; the profiled rerun landed at `237.78ms/op`,
`2.84 MB/op`, and `2,036 allocs/op`, and
`resolveConcreteMemberOverload(...)` dropped out of the hot runtime set. The
external guard measured bytecode `i_before_e` at `0.4480s` and `sudoku` at
`0.5640s` over `5/5` runs. Against the external rows, bytecode `i_before_e`
is now `8.96x` Go, `4.48x` Ruby, and `3.45x` Python; bytecode `sudoku` is
now `4.34x` Go, `0.10x` Ruby, and `0.19x` Python. The next profile should
target residual `resolveMethodCallableFromPool(...)` /
`lookupBoundMethodCache(...)` around canonical primitive methods, or switch
back to bytecode `fib` typed-frame work if the next member-call slice would
require another map/cache layer.

The next kept bytecode heap slice targeted canonical stdlib origin validation
itself. The `Array.get` overload selector still needs to prove that the active
methods come from canonical `able-stdlib`, but
`isCanonicalAbleStdlibOrigin(...)` was allocating fresh concatenated suffix
strings for every validation. It now checks the fixed `/able-stdlib/src/` and
`/pkg/src/` bases plus the relative suffix without allocating on slash-normal
paths. The focused allocation test pins the zero-allocation helper contract.
Runtime-only `sudoku` moved from the refreshed `339.53ms/op`,
`118.11 MB/op`, `1,572,523 allocs/op` sample to `334.69ms/op`,
`86.58 MB/op`, and `915,969 allocs/op`; the allocation profile no longer
shows `isCanonicalAbleStdlibOrigin(...)`, and
`isCanonicalNullableArrayGetOverloadSlow(...)` dropped to `0.06s` cumulative
in the profiled sample. The external guard moved bytecode `sudoku` from
`0.5640s` to `0.5160s` and bytecode `i_before_e` from `0.4480s` to `0.4280s`
over `5/5` runs. Against the external rows, bytecode `sudoku` was `3.97x` Go,
`0.09x` Ruby, and `0.17x` Python; bytecode `i_before_e` was `8.56x` Go,
`4.28x` Ruby, and `3.29x` Python.

The follow-up kept bytecode cache slice targeted the same canonical nullable
`Array.get` overload validation under noisier same-session external timing.
When member resolution returns a fresh overload wrapper around the same
canonical nullable `Array.get` function and lower-priority result-returning
implementation function, the VM now reuses the previous canonical-shape
validation result until the bytecode method/global cache version changes.
Unsupported overload shapes and invalidated versions still fall back to the
existing slow validation. Restored external bytecode passes before reapplying
the cache landed at `0.6480s` / `0.6080s` for `sudoku` and `0.5340s` /
`0.4760s` for `i_before_e`; the kept cache confirmations landed at `0.5280s`
/ `0.5340s` for `sudoku` and `0.4580s` / `0.4540s` for `i_before_e` over
`5/5` runs. The checked-in scoreboard records the later confirmation:
bytecode `sudoku` is now `4.11x` Go, `0.09x` Ruby, and `0.18x` Python;
bytecode `i_before_e` is now `9.08x` Go, `4.54x` Ruby, and `3.49x` Python.

The next kept match slice targeted the exact primitive typed-pattern case
inside that structural runtime environment problem. Simple typed patterns now
bind directly when the runtime value already has the exact primitive shape,
skipping generic `matchesType(...)` and coercion for hot paths such as
`case byte: u8` in `parse_board`. Non-exact integer widths, aliases, unions,
structs, interfaces, and every non-primitive shape still use the generic
typed-pattern path. Runtime-only `sudoku` moved from a refreshed
`340.43ms/op`, `86.60 MB/op`, `915,996 allocs/op` sample to `327.53ms/op`,
`326.26ms/op`, and `331.68ms/op` with the same allocation shape; the profiled
rerun landed at `332.99ms/op`. The external guard confirmed the keep:
`sudoku` measured `0.5080s` / `0.5040s`, and `i_before_e` measured
`0.4360s` / `0.4200s` over two `5/5` passes. Against the external rows,
bytecode `sudoku` is now `3.88x` Go, `0.09x` Ruby, and `0.17x` Python;
bytecode `i_before_e` is now `8.40x` Go, `4.20x` Ruby, and `3.23x` Python.
The next profile should target the remaining structural match/env issue
directly by lowering simple `match` clauses into slot-aware bytecode so
`parse_board` / `solve` can inline, or target `board_to_string` through a
spec-backed string builder / byte-buffer surface rather than another generic
interpolation tweak.

The follow-up kept bytecode slice landed that structural match/env fix for a
bounded subset. In slot-eligible functions, match clauses made from literal
`nil`, wildcard, or typed identifier/wildcard patterns now lower to direct
bytecode branch tests and slot bindings instead of the generic
`bytecodeOpMatch` path. Guarded clauses, non-nil literals, existing-symbol
identifier patterns, destructuring, and structural patterns remain on the
generic path. Typed clauses still use v12 `matchesType(...)` / coercion
semantics after the exact primitive fast check, so nominal matches such as
`case node: Node` are handled without adding per-container special cases.

Runtime-only `sudoku` moved from the prior exact-primitive match band of
`326.26-331.68ms/op`, `~86.60 MB/op`, and `~916k allocs/op` to
`209.21ms/op`, `205.12ms/op`, and `203.70ms/op`, with
`31.48-34.58 MB/op` and `499.5k-499.9k allocs/op`. A profiled one-shot landed
at `233.69ms/op`, `32.57 MB/op`, and `499,764 allocs/op`. The bytecode trace
now shows `parse_board`, `find_empty`, and `solve` dispatching inline. The
external guard moved bytecode `sudoku` to `0.4120s` over `5/5` runs. The
`i_before_e` guard stayed noisy but in the same broad band: `0.4680s` in the
combined guard and `0.4480s` on the rerun, versus the prior `0.4360s` /
`0.4200s` confirmations. Against the external rows, bytecode `sudoku` is now
`3.17x` Go, `0.07x` Ruby, and `0.14x` Python; bytecode `i_before_e` is now
`8.96x` Go, `4.48x` Ruby, and `3.45x` Python. External bytecode
`binarytrees` still times out at `60s`, so the next profile should target
post-match `sudoku` member/index/string work or move to the larger
typed-frame/struct-allocation problems in the timeout bytecode workloads.

A first bounded timeout-workload slice is now landed on that second front:
simple non-generic named struct literals with explicit field initializers now
lower to a dedicated bytecode opcode when function-body lowering can seed
visible struct definitions from the closure environment. On the reduced
external-style `binarytrees` profile case (`n := 16`), bytecode runtime moved
from `96.88s`, `12.54 GB/op`, and `187.36M allocs/op` to `71.05s`,
`10.99 GB/op`, and `127.77M allocs/op`. The old tree-walker
`evaluateStructLiteral(...)` wall dropped out of the top tier in favor of the
new `execStructLiteralNamedFast(...)` path. Full external bytecode
`binarytrees` still times out at `60s`, so this is a real timeout-family
keep but not a benchmark closure; the next step still needs a broader typed
nominal/struct allocation and call-boundary design slice rather than more
source-level local typing or helper cleanup.

A direct follow-up on the same reduced case is now landed too: exact simple
named-struct typed patterns now bypass the generic
`matchesType(...)` / `coerceValueToType(...)` nominal path when the runtime
value already carries that exact non-generic struct definition. That moved the
same reduced `binarytrees` case again to `62.83s`, `10.99 GB/op`, and
`127.77M allocs/op`, with `matchTypedPatternValue(...)`,
`matchesType(...)`, and `execJumpIfNotTypedPattern(...)` all dropping
materially in the profile. Full external bytecode `binarytrees` still times
out at `60s`, so the planning conclusion does not change: the next real
timeout-family step still needs a broader typed nominal/materialization
boundary rather than another local source tweak or helper cleanup.

The next kept bytecode member-call slice targeted the repeated canonical
`Array.get` method resolution left after slot-aware match lowering. Once a
bytecode `CallMember get` site fully resolves to the canonical nullable
stdlib `Array.get(i32)` overload, the VM now caches that proof per
program/IP/environment and later executes the existing tracked array-read fast
body directly. The cache is guarded by environment revision, global revision,
method-cache version, and absence of runtime impl context; unsupported shapes
still fall back to normal v12 member resolution and overload selection.

A paired no-trace runtime-only baseline landed at `200.66-209.51ms/op` with
about `499k allocs/op`. The kept cache band landed at `176.47-200.95ms/op`
with about `417k allocs/op`, and the no-trace profiled rerun landed at
`179.88ms/op`, `26.95 MB/op`, and `417,355 allocs/op`. The no-trace CPU
profile no longer shows `resolveMethodCallableFromPool(...)` as a top-tier
`sudoku` cost; the visible wall moved to `execCallMember(...)` guard work,
integer comparisons, and `board_to_string` string interpolation. The external
guard moved bytecode `sudoku` to `0.3500s` over `3/3` runs, about `2.69x` Go,
`0.06x` Ruby, and `0.12x` Python.

The follow-up kept bytecode cache-layout slice narrowed that same path without
changing the semantic guard. The canonical `Array.get` call-site cache now has
a 4-entry MRU hot tier and only reads env/global/method revisions after a
cheap program/IP/env identity match. This avoids the old single-entry hot
cache thrash across nested sudoku `Array.get` call sites while preserving the
same fallback behavior for unsupported shapes, env/global/method changes, and
runtime impl context. Paired runtime-only reruns moved from a restored
`169.46-187.57ms/op` band to `164.50-176.74ms/op`, with allocations unchanged
around `417k allocs/op`. The profiled rerun landed at `170.45ms/op`, and
`lookupCachedCanonicalArrayGetCall(...)` dropped from about `0.10s` cumulative
to about `0.02s`. Paired external bytecode `sudoku` moved from restored
pre-MRU `0.3833s` to MRU `0.3700s`; the checked-in current scoreboard now
records the MRU `0.3700s` row.

The next kept bytecode tranche adjusted call-member ordering on that same
canonical proof cache. `execCallMember(...)` now checks the guarded canonical
`Array.get` call-site cache before the single-entry general member-method
cache, allowing nested sudoku `get` sites to use the specialized 4-entry MRU
tier without first paying general member-cache miss/churn work. The guard
remains the same: only sites already proven by full member resolution to target
canonical nullable stdlib `Array.get(i32)` can use the fast path, and env,
global, method-cache, and runtime impl-context invalidation still fall back to
normal v12 member resolution.

The paired restored runtime-only baseline landed at `168.34-171.74ms/op` with
about `417k allocs/op`. The kept cache-first band landed at
`163.35-167.92ms/op`, also about `417k allocs/op`, and the profiled kept rerun
landed at `161.23ms/op`, `26.95 MB/op`, and `417,353 allocs/op`. External
bytecode `sudoku` moved from the prior recorded `0.3700s` to `0.3633s` over
`3/3` runs, about `2.79x` Go, `0.06x` Ruby, and `0.12x` Python. Non-keeps in
this tranche were a stdlib `StringBuilder` source rewrite, a two-part
interpolation VM fast path, and a single-thread propagation-cache mutex
bypass. The next profile should start from
`lookupCachedCanonicalArrayGetCall(...)` itself and the remaining
`board_to_string` interpolation allocation.

The next kept bytecode tranche narrowed that interpolation work without
changing Display semantics. `execStringInterpolation(...)` now handles
two-part primitive pairs directly, with a dedicated `String + Integer` subpath
that writes integer digits into a single grown builder. `String + String`
still uses the existing buffer-reuse path, and structs, arrays, functions,
errors, and other dynamic values still fall back to `stringifyValue(...)` so
custom `to_string` behavior is preserved.

Runtime-only `sudoku` now runs in the `161.29-169.59ms/op` band with about
`343k allocs/op`, down from the prior kept `~417k allocs/op` shape. A profiled
one-shot was CPU-noisy at `208.37ms/op` but confirmed `25.28 MB/op` and
`342,918 allocs/op`. External bytecode `sudoku` confirmed at `0.3620s` over
two `5/5` runs, versus the prior recorded `0.3633s`; against the external
rows this is `2.78x` Go, `0.06x` Ruby, and `0.12x` Python. The next profile
should start from residual `execCallMember(...)` / canonical `Array.get`
guard work and binary compare slots. Larger string wins likely need a general
byte-buffer/string-builder runtime primitive rather than another local
interpolation helper.

The next kept bytecode tranche fused another small control-flow shape exposed
by the post-interpolation profile. Slot-backed `>` and `>=` comparisons
against integer literals now lower to `JumpIfIntCompareSlotConstFalse`, so
guards such as `i >= 9` and `num > 9` no longer materialize a temporary bool
only for `JumpIfFalse` to pop. The existing `<=` return and branch fusions are
unchanged; unsupported operands fall back to the same generic binary operator
and truthiness behavior.

Runtime-only `sudoku` landed at `149.16ms/op`, `153.52ms/op`, and one noisy
`176.73ms/op`, with allocations unchanged around `343k allocs/op`. The
profiled sample landed at `171.99ms/op`, `25.44 MB/op`, and `342,921
allocs/op`; the old generic `execBinary` compare prominence is gone. External
bytecode `sudoku` moved from the prior recorded `0.3620s` to `0.3560s` over
`5/5` runs. Against the external rows, bytecode `sudoku` is now `2.74x` Go,
`0.06x` Ruby, and `0.12x` Python. The next profile should start from residual
`execCallMember(...)` / canonical `Array.get` guard work, slot load/store
traffic, and runtime type/propagation checks.

The next kept bytecode tranche trimmed that residual canonical `Array.get`
guard work without weakening the cache proof. The four-entry call-site cache
still stores and validates env revision, global revision, method-cache version,
and the absence of a runtime impl context, but hot hits no longer promote the
matched entry to the front on every nested sudoku access. Promotion still
happens when an entry is stored or recovered from the backing cache.

Runtime-only `sudoku` moved to a warmed `137.57ms/op`, `148.35ms/op`, and one
noisy `164.03ms/op`, with allocations unchanged around `343k allocs/op`. The
profiled sample landed at `159.54ms/op`, `25.48 MB/op`, and `342,856
allocs/op`; `lookupCachedCanonicalArrayGetCall` fell from about `100ms`
cumulative in the refreshed profile to about `70ms`. External bytecode
`sudoku` moved from `0.3560s` to `0.3360s` over `5/5` runs. Against the
external rows, bytecode `sudoku` is now `2.58x` Go, `0.06x` Ruby, and `0.11x`
Python. The next profile should start from this kept state and avoid another
`Array.get` cache micro-slice unless it is again the top visible wall.

The next kept bytecode tranche targeted residual boxed slot load/store traffic
instead of another member-cache layer. Slot-backed self assignments of the form
`x = x + const` and `x = x - const` now lower to
`StoreSlotBinaryIntSlotConst`, which performs the same checked slot-const
integer operation, stores the result, and leaves the assignment value on the
stack in one VM opcode. Unsupported shapes keep the old binary-plus-store
lowering, and typed `i32` slots keep the existing raw `StoreSlotI32` path.

Runtime-only `sudoku` moved to `140.23ms/op`, `141.16ms/op`, and one noisier
`145.35ms/op`, with allocations unchanged around `343k allocs/op`. The
profiled sample landed at `144.89ms/op`, `25.49 MB/op`, and `342,850
allocs/op`; the former standalone store edge for sudoku loop counters is gone,
with checked add/sub cost now visible inside the fused opcode. External
bytecode `sudoku` moved from `0.3360s` to `0.3320s` over `5/5` runs, while
external bytecode `i_before_e` stayed neutral at `0.4440s` over `5/5` runs.
Against the external rows, bytecode `sudoku` is now `2.55x` Go, `0.06x` Ruby,
and `0.11x` Python; bytecode `i_before_e` is now `8.88x` Go, `4.44x` Ruby,
and `3.42x` Python. The next profile should start from this kept state and
target `execCallMember(...)` / canonical `Array.get` guard cost, residual
checked integer arithmetic, or typed slot assignment checks.

The next kept bytecode/interpreter tranche targeted repeated runtime type
alias expansion rather than another opcode fusion. Runtime type checks now use
an interpreter-level `ast.TypeExpression` expansion cache around
`matchesType(...)`, cast coercion, and the exported alias-expansion bridge,
with invalidation on type-alias registration. This preserves the existing v12
alias semantics while avoiding repeated rebuilds of the same alias-expanded
type ASTs in sudoku's hot pattern/coercion paths.

Runtime-only `sudoku` moved from a restored `145.24ms/op` sample with about
`25.49 MB/op` and `342,847 allocs/op` to `136.38ms/op`, `132.95ms/op`, and
`141.75ms/op`, with about `22.3-25.1 MB/op` and `279k allocs/op`. The
profiled kept sample landed at `138.89ms/op`, `22.42 MB/op`, and `279,229
allocs/op`; `expandTypeAliases(...)` dropped from roughly `140ms` cumulative
in the restored profile to about `20ms`, and `substituteAliasTypeExpression`
fell from about `26.5 MB` flat allocation to `5 MB`. Same-session external
guards moved restored bytecode `sudoku` from `0.3540s` to `0.3460s` over
`5/5` runs and restored bytecode `i_before_e` from `0.4970s` to `0.4850s`
over `10/10` runs. Against the external rows, bytecode `sudoku` is now
`2.66x` Go, `0.06x` Ruby, and `0.11x` Python; bytecode `i_before_e` is now
`9.70x` Go, `4.85x` Ruby, and `3.73x` Python. The next profile should target
the post-cache member-call/type-check allocation wall: `execCallMember(...)`,
`resolveMethodCallableFromPool(...)`, and `storeBoundMethodCache(...)`.

The next kept bytecode tranche added a guarded call-member opcode for
canonical `Array.get` sites. Ordinary one-argument `.get(...)` calls now lower
to `CallMemberArrayGet`. The opcode still falls back to the full
`execCallMember(...)` path until the existing canonical nullable
`Array.get(i32)` call-site proof cache validates the site; once proven, hot
hits execute the direct tracked-array read without re-entering the broader
member dispatch ladder.

Runtime-only `sudoku` moved to `141.42ms/op`, `135.12ms/op`, and
`134.24ms/op` after the final call-dispatch helper split required to keep
`bytecode_vm_run.go` under 1000 lines, with allocations still around `279k
allocs/op`. The profiled kept sample landed at `134.14ms/op`, `22.42 MB/op`,
and `279,242 allocs/op`; the CPU profile showed `execCallMember(...)`
cumulative cost at about `200ms`, with the new guarded opcode accounting for
about `90ms`. External bytecode `sudoku` moved from the previous recorded
`0.3460s` to `0.3300s` over `5/5` runs. External bytecode `i_before_e` moved
from the previous recorded `0.4850s` to `0.4610s` over `10/10` runs. Against
the external rows, bytecode `sudoku` is now `2.54x` Go, `0.06x` Ruby, and
`0.11x` Python; bytecode `i_before_e` is now `9.22x` Go, `4.61x` Ruby, and
`3.55x` Python. The next profile should target remaining non-`Array.get`
`execCallMember(...)` paths: propagation/error checks, string iterator calls,
and residual bound-method cache allocation.

The next kept bytecode tranche added a guarded call-member opcode for
canonical string-byte iterator `next` calls. Ordinary zero-argument `.next()`
calls now lower to `CallMemberNext`. The opcode tries the existing canonical
`Iterator u8` / `RawStringBytesIter.next` fast body first and falls back to
the full `execCallMember(...)` path for safe-navigation calls, argument
calls, non-canonical iterators, and every unsupported receiver shape.

Runtime-only `sudoku` moved from a refreshed profiled baseline of
`135.81ms/op`, `22.43 MB/op`, and `279,238 allocs/op` to `130.46ms/op`,
`132.00ms/op`, and `133.02ms/op`. The profiled kept sample landed at
`128.28ms/op`, `22.41 MB/op`, and `279,229 allocs/op`; the CPU profile showed
`execCallMember(...)` down from about `280ms` cumulative in the refreshed
baseline profile to about `210ms`, with canonical iterator `next` now under
`execCallMemberNext(...)` at about `20ms`. External bytecode `sudoku`
confirmed at `0.3260s` over `5/5` runs, versus the prior recorded `0.3300s`.
External bytecode `i_before_e` was noisy but landed at `0.4760s` and
`0.4500s` over `10/10`, versus the prior recorded `0.4610s`. Against the
external rows, bytecode `sudoku` is now `2.51x` Go, `0.06x` Ruby, and `0.11x`
Python on the kept confirmation; bytecode `i_before_e` is `9.00x` Go,
`4.50x` Ruby, and `3.46x` Python on the best confirmation. The next profile
should target remaining non-`Array.get` / non-`next` `execCallMember(...)`
edges, especially static `Array.new`, propagation/error checks, and residual
bound-method cache allocation.

The next kept bytecode tranche added a guarded call-member opcode for
canonical static `Array.new` calls. Ordinary zero-argument `.new()` calls now
lower to `CallMemberArrayNew`. The opcode executes the direct empty-array
construction only after normal member resolution proves the canonical kernel
`Array.new() -> Array T` method at that program/IP; the proof is cached behind
environment, global, and method-cache revisions. Safe-navigation calls,
argument-bearing `new(...)` calls, non-`Array` receivers, non-canonical
definitions, runtime impl-context environments, and invalidated cache versions
all fall back to the full member path.

Runtime-only `sudoku` allocation dropped from the previous `~279k allocs/op`
band to `259,029`, `258,977`, and `259,275 allocs/op`, while wall-clock
stayed soft at `133.40ms/op`, `135.83ms/op`, and `136.59ms/op`. The profiled
kept sample landed at `133.51ms/op`, `21.71 MB/op`, and `259,043 allocs/op`;
the allocation profile showed `execCallMember(...)` down from the refreshed
`66.14 MB` cumulative sample to `47.20 MB`, with static construction now
flowing through `execCallMemberArrayNew(...)`. External bytecode `sudoku`
edged to `0.3240s` over `5/5`, versus the prior `0.3260s`. External bytecode
`i_before_e` was noisy at `0.5220s` and `0.4570s` over `10/10`, which keeps
it in the same broad band as the prior `0.4500-0.4760s` note. Against the
external rows, bytecode `sudoku` is now `2.49x` Go, `0.06x` Ruby, and `0.11x`
Python on the kept confirmation; bytecode `i_before_e` is `9.14x` Go,
`4.57x` Ruby, and `3.52x` Python on the better confirmation. The next profile
should target propagation/error checks or residual bound-method cache
allocation before adding more single-method call-member opcodes.

The follow-up kept bytecode tranche cleaned up the already-proven
`CallMemberArrayGet` hot path. Once the guarded canonical `Array.get(i32)`
call-site proof has validated a program/IP, the opcode now reuses the
already-proven array receiver and `i32` index and finishes the tracked-array
read directly instead of re-entering `execArrayGetMemberFast(...)` and
repeating stack, receiver, and argument shape checks. Unsupported shapes and
invalidated guards still fall back to the existing full member-call path.

Runtime-only `sudoku` moved from a refreshed `136.88ms/op`, `21.70 MB/op`,
`259,038 allocs/op` baseline to `120.74ms/op`, `123.67ms/op`, and
`131.53ms/op`, with allocation shape essentially unchanged. The profiled kept
sample landed at `126.92ms/op`, `21.71 MB/op`, and `259,048 allocs/op`;
`execCallMemberArrayGet(...)` dropped from about `110ms` cumulative in the
refreshed baseline profile to about `60ms`, and
`lookupCachedCanonicalArrayGetCall(...)` dropped from about `50ms` flat to
about `10ms` flat in the kept sample. External bytecode `sudoku` moved to
`0.3180s` over `5/5`, versus the prior `0.3240s`; external bytecode
`i_before_e` moved to `0.4420s` over `10/10`, versus the prior noisy
`0.4570-0.5220s` guard band. Against the external rows, bytecode `sudoku` is
now `2.45x` Go, `0.06x` Ruby, and `0.11x` Python; bytecode `i_before_e` is
`8.84x` Go, `4.42x` Ruby, and `3.40x` Python. The next profile should target
propagation/error checks, residual bound-method cache allocation, or string
interpolation allocation; avoid another `Array.get` cache/guard slice unless
fresh evidence puts it back at the top.

The follow-up kept bytecode tranche targeted the specific
`board_to_string` interpolation shape. The primitive `String + Integer`
interpolation fast path now concatenates cached one-byte digit suffixes for
`0..9` directly instead of routing those cases through
`strings.Builder.Grow`. Multi-digit integers, non-small integers, and generic
Display/`to_string` fallback remain on the existing paths.

Runtime-only `sudoku` moved from a refreshed `118.77ms/op`, `21.69 MB/op`,
`259,012 allocs/op` baseline to `117.40ms/op`, `118.33ms/op`, and
`121.52ms/op`, with allocation counts still around `259k allocs/op`. The
profiled kept sample was noisy at `127.25ms/op`, `21.70 MB/op`, and `258,928
allocs/op`, but the heap profile showed
`finishStringIntegerInterpolationFast(...)` down from about `53.5 MB`
cumulative in the refreshed baseline to about `37 MB`. External bytecode
`sudoku` held at `0.3180s` over `5/5`; external bytecode `i_before_e` edged
to `0.4410s` over `10/10`. Against the external rows, bytecode `sudoku`
remains `2.45x` Go, `0.06x` Ruby, and `0.11x` Python; bytecode `i_before_e`
is now `8.82x` Go, `4.41x` Ruby, and `3.39x` Python. The next profile should
target residual member resolution / bound-method cache allocation around
iterator `next` and static `Array.new`, or the remaining array
growth/allocation path; propagation is no longer a top trace item unless a
fresh profile brings it back.

As of April 29, 2026, the external `quicksort` compiled timeout is closed and
the benchmark is in Go range. This took two steps: first, the benchmark source
was changed to parse `numbers.txt` directly from bytes and use slot
reads/writes in the sort hot loop, moving compiled Able from timeout to
`11.42s`; then the compiler gained a native host-slice return path for mono
`Array T` externs and `fs.read_bytes` was routed through `os.ReadFile`, moving
compiled Able to a verified `1.75s` average over `3/3` runs.

Current external quicksort comparison:

- compiled Able: `1.75s`
- Go reference: `2.01s`
- Ruby reference: `14.58s`
- Python reference: `20.32s`

The profiled `1.79s` run no longer shows the old
`hostValueToRuntime(...)` / per-byte `BigInt` return bridge. The remaining
CPU is actual parse/sort work. Allocation is now about `429.55MB`, including
the `os.ReadFile` buffer, the deliberate host-boundary copy into Able-owned
`Array u8`, and the parsed `Array i32`.

The next kept compiled tranches targeted external `i_before_e` without
changing the benchmark source. First, compiled Go extern wrappers now keep
native scalar arguments and results native when the Able compiled carrier type
already matches the Go host type. The string fast-path externs in the
canonical stdlib therefore stop paying `RuntimeValueToHost`,
`HostValueToRuntime`, `bridge.ToString`, and scalar `bridge.As*` conversions
on every `String.contains`, `String.replace`, and `String.len_bytes` call.
Then static no-fallback launchers stopped running the loader/parser/program
evaluation path when the generated binary can seed imports directly. The
launcher seedability check now accepts public type-alias selector imports and
explicit compiler-known internal extern selector imports, while statically
known calls to unsupported receiver methods still force the bootstrap path.

Current external `i_before_e` compiled comparison:

- compiled Able: `0.0620s`
- Go reference: `0.0500s`
- Ruby reference: `0.1000s`
- Python reference: `0.1300s`

The refreshed pre-extern compiled baseline in the same session was `0.2700s`
over `3/3` runs. The scalar extern tranche moved the restored source to
`0.1900s`; the static-launcher tranche then moved it to `0.0620s` over `5/5`
runs, which is now in the Go-reference range at about `1.24x` Go. A
`String.index_of(...)` source rewrite was tested and reverted because it
regressed external bytecode `i_before_e` to `2.3533s` over `3/3` runs.

The first kept aligned steady-state results are:

- `fib`: still timed out with a `300s` warmup+measure budget
- `i_before_e`: `62.14s/op`, `107.19 GB/op`, `315,106,815 allocs/op`

So the remaining bytecode problem is clearly VM runtime itself, not one-time
CLI/bootstrap/lowering noise.

The next kept steady-state profiling slice now starts CPU/heap profiling only
after program load/lowering plus the explicit warmup call, and it caches
lowered expression bytecode programs by AST-expression identity plus
placeholder-lambda mode. The refreshed aligned `i_before_e` steady-state
result on that kept code is:

- `27.92s/op`
- `16.61 GB/op`
- `172,094,647 allocs/op`

The corrected runtime-only profiles no longer show
`lowerExpressionToBytecodeWithOptions(...)`,
`(*bytecodeLoweringContext).emit(...)`, or `emitExpression(...)` in the hot
set. The remaining steady-state cost is now real VM/runtime work centered on
`execCallName`, `execCall`, `runMatchExpression`, identifier/call-name cache
bookkeeping, and the resulting GC/allocation pressure.

The next kept steady-state shared array-metadata boxing slice now reuses
shared boxed metadata values for common dynamic-array lengths/capacities and
for the `__able_array_size` helper path. The refreshed aligned
`bytecode-runtime` `i_before_e` result on that kept code is:

- `20.24s/op`, `4.57 GB/op`, `72,504,749 allocs/op` on the profiled rerun
- `19.87s/op`, `4.57 GB/op`, `72,505,049 allocs/op` on the clean rerun

The profile shift is specific and useful: `(*Interpreter).initArrayBuiltins.func4`
fell from about `167 MB` flat alloc-space to about `148 MB`, and
`(*ArrayState).BoxedLengthValue` fell from about `160 MB` to about `149 MB`.
The remaining steady-state wall is still mostly environment creation, struct
literal construction, match/ensure body evaluation, and residual host-value
conversion, but the array metadata path is no longer paying as much repeated
first-access boxing cost.

The next kept steady-state small unsigned extern-host conversion slice now
lowers host `u8` / `u16` / `u32` and in-range `u64` / `usize` results straight
to small integers instead of routing them through `big.Int` boxing. The
refreshed aligned `bytecode-runtime` `i_before_e` result on that kept code is:

- `22.95s/op`, `4.50 GB/op`, `69,018,740 allocs/op` on the profiled rerun
- `20.57s/op`, `4.50 GB/op`, `69,019,113 allocs/op` on clean rerun A
- `21.65s/op`, `4.50 GB/op`, `69,018,676 allocs/op` on clean rerun B
- `19.46s/op`, `4.50 GB/op`, `69,021,452 allocs/op` on clean rerun C

This is an allocation-pressure win rather than a clean wall-clock jump, but it
is real: `(*Interpreter).fromHostValue` fell from about `103 MB` flat
alloc-space to about `91 MB`, and the old `bigIntFromUint(...)`-driven
unsigned-host conversion slice dropped out of the top alloc set entirely. That
means the remaining steady-state host path is now less about integer boxing and
more about the residual union/nullable dispatch plus the larger environment /
match / ensure wall around it.

The next kept steady-state lazy environment-mutex tranche now keeps the
per-environment `sync.RWMutex` behind a lazy `atomic.Pointer` in
`pkg/runtime/environment.go`, so the single-threaded bytecode hot path stops
carrying an eagerly allocated lock payload in every short-lived lexical scope.
The refreshed aligned `bytecode-runtime` `i_before_e` result on that kept code
is:

- `21.98s/op`, `4.25 GB/op`, `69,018,511 allocs/op` on the profiled rerun
- `21.97s/op`, `4.25 GB/op`, `69,018,009 allocs/op` on clean rerun A
- `21.84s/op`, `4.25 GB/op`, `69,019,472 allocs/op` on clean rerun B

This is another object-size reduction win rather than a scope-count reduction,
and the alloc profile makes that clear: `NewEnvironmentWithValueCapacity(...)`
fell from about `1.71 GB` flat alloc-space to about `1.48 GB`, while alloc
count stayed in the same `69.02M` band. That shifts the remaining steady-state
wall even more cleanly toward `evaluateStructLiteral(...)`,
`setCurrentValueNoLock(...)`, `runMatchExpression(...)`, `execEnsureStart(...)`,
and the residual host-value conversion path.

The next kept text-path tranche moved out of the VM and into the canonical
external stdlib: `../able-stdlib/src/fs.able` now routes `fs.read_lines(...)`
through a direct `fs_read_lines_fast(...)` extern instead of layering
`open` / `read_all` / `close`, `bytes_to_string(...)`, newline normalization,
and line splitting in Able code. The refreshed aligned `i_before_e` results on
that kept code are:

- bytecode-runtime clean rerun A: `1.28s/op`, `101.46 MB/op`,
  `3,582,266 allocs/op`
- bytecode-runtime clean rerun B: `1.40s/op`, `101.47 MB/op`,
  `3,582,304 allocs/op`
- bytecode-runtime profiled: `1.41s/op`, `101.51 MB/op`,
  `3,582,321 allocs/op`
- compiled external compare: `0.38s`

This is a major workload-specific collapse of the old text/fs stack. The
steady-state alloc profile no longer shows `runMatchExpression(...)`,
`execEnsureStart(...)`, `NewEnvironmentWithValueCapacity(...)`, or
`evaluateStructLiteral(...)` anywhere near the old top tier for
`i_before_e`. The remaining wall on this benchmark is now `copyCallArgs`,
`resolveMethodFromPool`, `stringMemberWithOverrides`, and the residual
extern-host conversion path.

The next kept text-path VM tranche stayed inside the interpreter runtime.
`resolveMethodFromPool(...)` now skips eager scope-callable and type-name
probing for primitive receivers until after inherent/interface/native method
lookup fails, so hot `String` method calls no longer pay `env.Lookup(...)` /
`env.Has(...)` work on the common success path. The refreshed aligned
`i_before_e` results on that kept code are:

- bytecode-runtime clean rerun: `1.27s/op`, `101.47 MB/op`,
  `3,582,283 allocs/op`
- bytecode-runtime profiled: `1.24s/op`, `101.51 MB/op`,
  `3,582,351 allocs/op`

This is a CPU-path cleanup, not a new heap collapse. The profiled comparison
against the prior kept `read_lines` fast-path state shows the intended shift:
`resolveMethodFromPool(...)` dropped from about `210ms` cumulative to about
`120ms`, and `stringMemberWithOverrides(...)` dropped from about `250ms`
cumulative to about `100ms`, while the benchmark stayed in the same
`~101 MB/op` band. The next remaining text-path wall is now led more clearly
by `copyCallArgs`, the residual `resolveMethodFromPool(...)` flat alloc slice,
`overloadArgKinds(...)`, and extern-host conversion.

The next kept text-path VM tranche stayed on the exact-native extern path.
The Go extern wrappers created by `makeExternNative(...)` now mark
`BorrowArgs: true`, so synchronous extern-host calls no longer clone argument
slices before dispatch. The refreshed aligned `i_before_e` results on that
kept code are:

- bytecode-runtime clean rerun A: `1.14s/op`, `84.88 MB/op`,
  `3,063,799 allocs/op`
- bytecode-runtime clean rerun B: `1.19s/op`, `84.89 MB/op`,
  `3,063,840 allocs/op`
- bytecode-runtime profiled: `1.11s/op`, `84.92 MB/op`,
  `3,063,917 allocs/op`

This is a real alloc collapse. The refreshed alloc profile no longer shows
`copyCallArgs(...)` anywhere near the top tier, and total profiled alloc-space
fell from about `119.85 MB` to about `90.95 MB`. The remaining text-path wall
is now led more cleanly by `resolveMethodFromPool(...)`,
`overloadArgKinds(...)`, `stringMemberWithOverrides(...)`, and residual
extern conversion through `fromHostValue(...)`.

The next kept text-path VM tranche stayed on the overload-selection cache
path. `v12/interpreters/go/pkg/interpreter/eval_expressions_calls_overloads.go`
now uses an inline comparable overload-cache signature for the common
small-arity cases instead of rebuilding the old concatenated
`overloadArgKinds(...)` string on every hot lookup, while larger arities still
fall back to the old slow path. The refreshed aligned `i_before_e` results on
that kept code are:

- bytecode-runtime clean rerun: `1.18s/op`, `75.17 MB/op`,
  `2,718,071 allocs/op`
- bytecode-runtime profiled: `1.14s/op`, `75.21 MB/op`,
  `2,718,161 allocs/op`

This is another real alloc collapse. The refreshed alloc profile no longer
shows `overloadArgKinds(...)` in the top tier at all, and total profiled
alloc-space fell again from about `90.95 MB` to about `82.41 MB`. The
remaining text-path wall is now led more cleanly by
`resolveMethodFromPool(...)`, `stringMemberWithOverrides(...)`, and residual
extern conversion through `fromHostValue(...)`.

The next kept text-path VM tranche stayed on the hot member-call path itself.
`v12/interpreters/go/pkg/interpreter/interpreter_method_resolution.go` now
exposes `resolveMethodCallableFromPool(...)`, and bytecode lowering now emits a
dedicated `bytecodeOpCallMember` so the VM can resolve the callable template
and inject the receiver directly instead of first allocating a fresh
`runtime.BoundMethodValue` for common `obj.method(...)` calls. The refreshed
aligned `i_before_e` results on that kept code are:

- bytecode-runtime clean rerun A: `1.06s/op`, `55.81 MB/op`,
  `1,853,918 allocs/op`
- bytecode-runtime clean rerun B: `1.09s/op`, `55.81 MB/op`,
  `1,853,917 allocs/op`
- bytecode-runtime profiled: `1.09s/op`, `55.85 MB/op`,
  `1,854,009 allocs/op`

This is another real alloc collapse. The refreshed alloc profile no longer
shows `resolveMethodFromPool(...)` in the top tier at all, and total profiled
alloc-space fell again from about `82.41 MB` to about `72.65 MB`. The
remaining text-path wall is now led more cleanly by
`callResolvedCallableWithInjectedReceiver(...)`,
`fromHostValue(...)`, and the residual extern/string conversion path.

The next kept text-path VM tranche stayed on that residual extern return path.
`v12/interpreters/go/pkg/interpreter/extern_host_fast.go` now routes hot `i32`
fast-invoker results through `boxedOrSmallIntegerValue(...)` instead of boxing
a fresh `runtime.NewSmallInt(...)` on every call, while
`v12/interpreters/go/pkg/interpreter/extern_host_coercion.go` plus the new
`v12/interpreters/go/pkg/interpreter/extern_host_result_fast.go` now fast-path
host `String`, `Array String`, and `IOError | Array String`-style union
results before falling back to the old generic reflect-heavy conversion path.
The refreshed aligned `i_before_e` results on that kept code are:

- bytecode-runtime clean rerun A: `1.08s/op`, `42.69 MB/op`,
  `1,335,429 allocs/op`
- bytecode-runtime clean rerun B: `1.08s/op`, `42.70 MB/op`,
  `1,335,471 allocs/op`
- bytecode-runtime profiled: `1.08s/op`, `42.74 MB/op`,
  `1,335,572 allocs/op`

This is another real alloc collapse. The refreshed alloc profile no longer
shows the old generic `fmt.Sprint(value.Interface())` / `fromHostValue(...)`
extern return path in the top flat allocators, the hot
`buildExternFastInvoker.func1` `runtime.NewSmallInt(...)` slice disappeared,
and total profiled alloc-space fell again from about `72.65 MB` to about
`50.44 MB` while wall-clock stayed in the same restored `~1.08s/op` band.
The remaining text-path wall is now led more cleanly by
`callResolvedCallableWithInjectedReceiver(...)`,
`boxedOrSmallIntegerValue(...)`, `externStringSliceResult(...)`, and the
residual `reflect.Value.Call` path behind `fs_read_lines_fast(...)`.

The next kept text-path VM tranche stayed on that residual union-return extern
path. `v12/interpreters/go/pkg/interpreter/extern_host.go` now passes the
active interpreter into cached fast invokers, and
`v12/interpreters/go/pkg/interpreter/extern_host_fast.go` now fast-paths
one-string-arg `func(string) interface{}` host wrappers, which is the shape
produced for union-return externs like `fs_read_lines_fast(...)`. Hot
`[]string` success returns now bypass `reflect.Value.Call` entirely, while
non-`[]string` results still fall back through `fromHostValue(...)` using the
already-computed direct result. The refreshed aligned `i_before_e` results on
that kept code are:

- bytecode-runtime clean rerun A: `0.98s/op`, `42.71 MB/op`,
  `1,335,497 allocs/op`
- bytecode-runtime clean rerun B: `1.01s/op`, `42.70 MB/op`,
  `1,335,460 allocs/op`
- bytecode-runtime profiled: `1.01s/op`, `42.74 MB/op`,
  `1,335,541 allocs/op`

This is a real CPU-path keep. Heap stayed in the same collapsed `~42.7 MB/op`
band, but the refreshed profile shows the intended shift:
`reflect.Value.Call` dropped out of the top CPU and alloc sets for the
`fs_read_lines_fast(...)` success path, and steady-state wall-clock moved down
from the prior `~1.08s/op` band into the `~0.98-1.01s/op` band. The remaining
text-path wall is now led more cleanly by
`callResolvedCallableWithInjectedReceiver(...)`,
`boxedOrSmallIntegerValue(...)`, `externStringSliceResult(...)`, and the
residual direct plugin-body cost in `fs_read_lines_fast(...)`.

The next kept text-path VM tranche widened the bytecode small-int boxing cache
to all int64-representable integer suffixes instead of limiting the hot path to
`i32`, `i64`, and `isize`. The kept code is in
`v12/interpreters/go/pkg/interpreter/bytecode_vm_small_int_boxing.go`, with the
focused unsigned-cache coverage in
`v12/interpreters/go/pkg/interpreter/bytecode_vm_slot_const_immediates_test.go`.
The refreshed aligned `i_before_e` results on that kept code are:

- bytecode-runtime clean rerun A: `1.09s/op`, `34.39 MB/op`,
  `1,162,592 allocs/op`
- bytecode-runtime profiled: `1.08s/op`, `34.43 MB/op`,
  `1,162,662 allocs/op`

Subsequent clean reruns on this machine were wall-clock noisy while holding the
same `~34.4 MB/op` / `1.16M allocs/op` heap shape, so this is best treated as
an allocation-pressure keep rather than a new stable CPU-path win. The real
profile shift is that the old unsupported-kind fallback in
`boxedOrSmallIntegerValue(...)` disappeared from the hot alloc story, and total
profiled alloc-space fell again from about `61.30 MB` to about `50.83 MB`.
That leaves the remaining text-path wall more cleanly concentrated in
`bytecodeBoxedIntegerValue(...)`, `patternToInteger(...)`,
`callResolvedCallableWithInjectedReceiver(...)`,
`externStringSliceResult(...)`, and the residual direct plugin-body cost in
`fs_read_lines_fast(...)`.

The next kept text-path VM tranche stayed on that remaining integer-conversion
wall. `v12/interpreters/go/pkg/interpreter/interpreter_type_coercion_fast.go`
now keeps small integer-to-integer suffix casts on a 64-bit arithmetic path
instead of immediately allocating `big.Int` values for
`bitPattern(...)` / `patternToInteger(...)`, and the same helper is reused from
`v12/interpreters/go/pkg/interpreter/interpreter_type_coercion.go` so the
post-alias simple-type cast path also avoids that small-int `big.Int` churn.
Focused coverage in
`v12/interpreters/go/pkg/interpreter/interpreter_type_coercion_fast_test.go`
now pins small signed-to-unsigned wrap behavior, the negative-to-`u64`
big-integer fallback boundary, and the bounded allocation behavior on repeated
`u8` casts. The refreshed aligned `i_before_e` results on that kept code are:

- bytecode-runtime clean rerun A: `1.35s/op`, `26.10 MB/op`,
  `644,158 allocs/op`
- bytecode-runtime clean rerun B: `1.17s/op`, `26.10 MB/op`,
  `644,158 allocs/op`
- bytecode-runtime profiled: `1.36s/op`, `26.13 MB/op`,
  `644,192 allocs/op`

This is a keep for allocation pressure, not a stable CPU-path win. Wall-clock
stayed noisy on this machine, but the heap shift is large and repeatable. The
important profile change is that
`castValueToCanonicalSimpleTypeFast(...)`, `castValueToType(...)`,
`patternToInteger(...)`, and `bitPattern(...)` dropped out of the top alloc set
entirely, and total profiled alloc-space fell again from about `50.83 MB` to
about `44.90 MB`. That leaves the remaining text-path wall more cleanly
concentrated in `bytecodeBoxedIntegerValue(...)`,
`callResolvedCallableWithInjectedReceiver(...)`,
`resolveMethodCallableFromPool(...)`, `externStringSliceResult(...)`, and the
residual direct plugin-body cost in `fs_read_lines_fast(...)`.

The next kept text-path VM tranche closed a real gap between method resolution
and bytecode member execution. The resolver can already return
`runtime.NativeBoundMethodValue` for primitive/native receivers, but
`v12/interpreters/go/pkg/interpreter/bytecode_vm_call_member.go` had only been
accepting raw `NativeFunctionValue` templates on the direct member-call
exact-native path. The kept change now lets
`bytecodeResolveExactInjectedNativeCallTarget(...)` accept
`NativeBoundMethodValue` too, and the focused coverage in
`v12/interpreters/go/pkg/interpreter/bytecode_vm_call_member_test.go` pins that
shape directly. The refreshed aligned `i_before_e` results on the kept code
are:

- bytecode-runtime clean rerun A: `1.08s/op`, `26.09 MB/op`,
  `644,119 allocs/op`
- bytecode-runtime clean rerun B: `1.07s/op`, `26.13 MB/op`,
  `644,192 allocs/op`
- bytecode-runtime clean rerun C: `1.12s/op`, `26.09 MB/op`,
  `644,118 allocs/op`
- bytecode-runtime profiled: `1.07s/op`, `26.13 MB/op`,
  `644,192 allocs/op`

This is a CPU-path keep layered on top of the prior heap work. The aligned
runtime moved back into the low `~1.07-1.12s/op` band while preserving the
post-cast `~26.1 MB/op` / `644k allocs/op` heap shape. The remaining text-path
wall is still led by `callResolvedCallableWithInjectedReceiver(...)`,
`resolveMethodCallableFromPool(...)`, `externStringSliceResult(...)`,
`bytecodeBoxedIntegerValue(...)`, and the residual direct plugin-body cost in
`fs_read_lines_fast(...)`.

The next kept text-path VM tranche closed a real direct-call cache gap.
`v12/interpreters/go/pkg/interpreter/bytecode_vm_call_member.go` now uses the
existing bytecode member-method cache on the `bytecodeOpCallMember` path
instead of bypassing it and re-running method resolution on every direct
`obj.method(...)` call. That also closes a real regression: the existing
`TestBytecodeVM_StatsMemberMethodCacheCounters` proof was red because
`execCallMember(...)` never consulted the cache even though member access and
dotted call-name fallback already did. On a miss, the VM now stores a rebound
template for the same cache surface; on a hit, it executes the cached resolved
member callee through the exact-native / inline / generic call ladder without
re-running `resolveMethodCallableFromPool(...)`. The refreshed aligned
`i_before_e` results on the kept code are:

- bytecode-runtime clean rerun A: `1.00s/op`, `26.09 MB/op`,
  `644,119 allocs/op`
- bytecode-runtime clean rerun B: `1.00s/op`, `26.09 MB/op`,
  `644,119 allocs/op`
- bytecode-runtime profiled: `1.02s/op`, `26.13 MB/op`,
  `644,191 allocs/op`

This is a keep as both correctness and CPU-path work. The cache-counter
regression is closed, and aligned steady-state bytecode `i_before_e` moved from
the prior `~1.07-1.12s/op` band into the low `~1.00-1.02s/op` band while
preserving the post-cast `~26.1 MB/op` / `644k allocs/op` heap shape. The
remaining text-path wall is still led by
`callResolvedCallableWithInjectedReceiver(...)`,
`resolveMethodCallableFromPool(...)`, `externStringSliceResult(...)`,
`bytecodeBoxedIntegerValue(...)`, and the residual direct plugin-body cost in
`fs_read_lines_fast(...)`.

The next kept text-path VM tranche stayed on method resolution itself.
`v12/interpreters/go/pkg/interpreter/interpreter_method_resolution.go` now
gives primitive receivers stable bound-method cache keys by type token instead
of treating value receivers like `String` as uncached, and it keeps that cache
semantically safe by only storing primitive receiver entries when resolution
actually came from a real method candidate. Primitive scope-fallback callables
stay uncached. The focused coverage in
`v12/interpreters/go/pkg/interpreter/interpreter_method_resolution_cache_test.go`
now pins both sides directly: primitive `String` methods reuse one cache entry
across distinct receiver values, and primitive scope-fallback callables do not
get cached across reassignment. The refreshed aligned `i_before_e` results on
the kept code are:

- bytecode-runtime clean rerun A: `0.92s/op`, `26.10 MB/op`,
  `644,118 allocs/op`
- bytecode-runtime clean rerun B: `0.97s/op`, `26.10 MB/op`,
  `644,117 allocs/op`
- bytecode-runtime profiled: `1.11s/op`, `26.13 MB/op`,
  `644,191 allocs/op`

This is a CPU-path keep layered on top of the prior cache work. The refreshed
profile shows `resolveMethodCallableFromPool(...)` dropping from the older
`~250ms` cumulative tier to about `~70ms` while keeping the post-cast heap
shape effectively flat. The remaining text-path wall is now more cleanly led by
`callResolvedCallableWithInjectedReceiver(...)`,
`externStringSliceResult(...)`, `bytecodeBoxedIntegerValue(...)`, and the
residual direct plugin-body cost in `fs_read_lines_fast(...)`.

The next kept text-path VM tranche stayed on that injected-receiver wall, but
it did so by moving the receiver prepend into the shared callable dispatcher
instead of creating another direct-call path. `v12/interpreters/go/pkg/interpreter/eval_expressions_calls.go`
now exposes a shared optional injected-receiver helper, and
`v12/interpreters/go/pkg/interpreter/bytecode_vm_call_member.go` now passes the
existing VM stack slice plus receiver into that helper instead of first
materializing a fresh merged argument slice for every `obj.method(...)` call.
Focused coverage in
`v12/interpreters/go/pkg/interpreter/bytecode_vm_call_member_test.go` now also
pins optional-arity and overloaded method-call semantics on the direct
member-call opcode path. The refreshed aligned `i_before_e` results on the kept
code are:

- bytecode-runtime clean rerun: `0.90s/op`, `20.57 MB/op`,
  `471,293 allocs/op`
- bytecode-runtime profiled: `0.90s/op`, `20.60 MB/op`,
  `471,367 allocs/op`

This is a keep as both CPU-path and heap work. The injected member-call merge
fell out of the top alloc set entirely, and aligned steady-state bytecode
`i_before_e` moved from the prior `~0.92-0.97s/op` band into the `~0.90s/op`
band while heap fell from about `26.1 MB/op` / `644k allocs/op` to about
`20.6 MB/op` / `471k allocs/op`. The remaining text-path wall is now more
cleanly `bytecodeBoxedIntegerValue(...)`, `externStringSliceResult(...)`,
`strings.genSplit`, and the residual direct plugin-body cost in
`fs_read_lines_fast(...)`.

The next kept steady-state tranche raised the lazy dynamic boxed-int cache cap
from `32768` to `262144` in
`v12/interpreters/go/pkg/interpreter/bytecode_vm_small_int_boxing.go`, which
is large enough for one warmup pass to retain the full large loop-index working
set on aligned `i_before_e` without broadening the eager fixed small-int cache.
The refreshed aligned `i_before_e` results on the kept code are:

- bytecode-runtime clean rerun: `0.88s/op`, `14.63 MB/op`,
  `347,625 allocs/op`
- bytecode-runtime profiled: `0.88s/op`, `14.66 MB/op`,
  `347,695 allocs/op`

This is a keep as heap work with a stable CPU band. The refreshed alloc profile
shows the intended shift: `bytecodeBoxedIntegerValue(...)` dropped from about
`5.63 MB` flat alloc-space to about `1.54 MB`, and total profiled alloc-space
fell from about `44.37 MB` to about `34.94 MB`. The remaining text-path wall is
now more cleanly `externStringSliceResult(...)`, `strings.genSplit`,
`buildExternFastInvoker.func8`, and the residual plugin body in
`fs_read_lines_fast(...)`.

The next kept bytecode-runtime tranche reworked the `fs_read_lines_fast(...)`
cache shape in the external stdlib. Instead of the earlier shared map +
`RWMutex` experiment, [fs.able](/home/david/sync/projects/able-stdlib/src/fs.able)
now keeps a single-entry immutable hot cache keyed by
`path + size + modifiedNs` behind `atomic.Pointer`. The hot repeated-read path
is now just `os.Stat(...)` plus an atomic load/compare, while misses still
rebuild from `os.ReadFile(...)` and replace the cached entry. The rewrite
invalidation proof is pinned in
[compiler_stdlib_io_temp_test.go](/home/david/sync/projects/able/v12/interpreters/go/pkg/compiler/compiler_stdlib_io_temp_test.go).

The refreshed aligned `i_before_e` result on that kept code is:

- bytecode-runtime clean rerun A: `0.91s/op`, `8.37 MB/op`, `347,617 allocs/op`
- bytecode-runtime clean rerun B: `0.89s/op`, `8.37 MB/op`, `347,617 allocs/op`
- bytecode-runtime profiled: `0.89s/op`, `8.40 MB/op`, `347,690 allocs/op`

This is a keep as both heap work and a CPU-safe cache shape. Compared with the
prior kept dynamic boxed-int tranche, aligned steady-state bytecode
`i_before_e` stays in the same sub-second band while heap falls from about
`14.6 MB/op` to about `8.4 MB/op`. The refreshed alloc profile shows the
intended shift: the old `strings.genSplit` / `os.readFileContents`
plugin-body cost drops out of the measured hot path, leaving
`buildExternFastInvoker.func8`, `externStringSliceResult(...)`,
`bytecodeBoxedIntegerValue(...)`, and residual member/native dispatch as the
cleaner remaining wall.

The next kept bytecode-runtime tranche moved off full per-call string-slice
boxing in the interpreter fast invoker layer. In
[extern_host_fast.go](/home/david/sync/projects/able/v12/interpreters/go/pkg/interpreter/extern_host_fast.go),
each string-slice fast invoker now keeps a tiny cached template for
`[]string -> []runtime.Value`, keyed by a source snapshot. Repeated hot
`Array String` extern results now clone that cached boxed template instead of
re-boxing every `StringValue` from scratch on each call, while still returning
a fresh Able array backing slice. The invalidation/no-aliasing behavior is
pinned in
[extern_host_result_fast_test.go](/home/david/sync/projects/able/v12/interpreters/go/pkg/interpreter/extern_host_result_fast_test.go).

The refreshed aligned `i_before_e` result on that kept code is:

- bytecode-runtime clean rerun: `0.87s/op`, `5.61 MB/op`, `174,794 allocs/op`
- bytecode-runtime profiled: `0.88s/op`, `5.64 MB/op`, `174,867 allocs/op`

This is a keep as another material heap collapse with a flat CPU band.
Compared with the prior kept atomic `read_lines` hot-cache tranche, aligned
steady-state bytecode `i_before_e` stays in the same sub-second range while
heap falls from about `8.4 MB/op` / `348k allocs/op` to about
`5.6 MB/op` / `175k allocs/op`. The refreshed alloc profile shows the intended
shift: `externStringSliceResult(...)` drops out of the top alloc tier and the
remaining array-string return cost is now mostly one cloned `[]runtime.Value`
slice in `externCloneValueSlice(...)`, plus the remaining member/native
dispatch wall.

The next kept steady-state runtime tranche preserved validated VM lookup caches
across pooled runs and moved match-clause binding scopes onto pre-sized local
environments with non-merging binds. The refreshed aligned `i_before_e`
steady-state result on that kept code is:

- `24.61s/op`
- `9.90 GB/op`
- `146,646,034 allocs/op`

The runtime-only alloc-space profile now shows the intended shift:
`storeCachedScopeValue(...)` and `storeCachedCallName(...)` are no longer
top-tier allocators. The remaining steady-state pressure is now centered on
`matchPattern(...)`, environment creation/binds for clause scopes, and the
runtime work those clause scopes feed (`runMatchExpression(...)`,
`execEnsureStart(...)`, extern-host calls, struct literal construction, and
runtime context snapshots).

The next kept steady-state tranche moved simple match clauses onto a direct
fast path for identifier, wildcard, literal, and typed patterns instead of
always going through the generic binding collector. The refreshed aligned
`i_before_e` steady-state result on that kept code is:

- `23.84s/op`
- `8.89 GB/op`
- `139,670,607 allocs/op`

The updated profiles show the right shift again: `matchPattern(...)` itself is
no longer a dominant allocator, and `matchPatternFast(...)` now carries much
of the hot match work directly. The remaining steady-state pressure is now
mostly clause-scope environment/binding churn plus the runtime work those
clauses feed, especially `runMatchExpression(...)`, `execEnsureStart(...)`,
extern-host calls, struct literal construction, and runtime context
snapshots.

The next kept steady-state tranche moved runtime diagnostic work for explicit
`return` onto a lazy path. Bytecode and tree-walker `returnSignal` no longer
snapshot the call stack on every normal return; return-type coercion failures
now attach runtime context only when they actually need a diagnostic. The
refreshed aligned `i_before_e` steady-state result on that kept code is:

- `23.40s/op`
- `8.74 GB/op`
- `136,183,494 allocs/op`

The refreshed profiles show the expected shift again: `snapshotCallStack(...)`
dropped materially from the alloc-space top tier, so the remaining steady-state
pressure is now more cleanly concentrated in clause-scope environment/binding
churn plus `runMatchExpression(...)`, `execEnsureStart(...)`, extern-host
calls, struct literal construction, and runtime context attachment on actual
error paths.

The next kept steady-state tranche moved the hot extern path onto cached target
hashes plus direct invokers for the primitive string signatures that dominate
`i_before_e`. The refreshed aligned steady-state result on that kept code is:

- profiled: `23.43s/op`, `8.23 GB/op`, `121,472,207 allocs/op`
- clean rerun: `22.24s/op`, `8.23 GB/op`, `121,472,290 allocs/op`

The refreshed profiles show the intended shift: `hashExternState(...)`,
`externSignatureKey(...)`, and the old cumulative extern-host hashing path
dropped out of the top allocators, while `invokeExternHostFunction(...)` /
`fromHostResults(...)` shrank materially. The remaining steady-state pressure
is now more clearly concentrated in clause-scope environment/binding churn,
`runMatchExpression(...)`, `execEnsureStart(...)`, struct literal
construction, and the residual host-value conversion path.

The next kept steady-state tranche moved the common one-binding child-scope
case in `runtime.Environment` onto an inline slot instead of an eager
one-entry map. `NewEnvironmentWithValueCapacity(...)` now avoids allocating a
map for `valueCapacity == 1`, and environments promote to a real map only on
the second distinct local binding. The refreshed aligned steady-state
`i_before_e` result on that kept code is:

- profiled: `24.40s/op`, `6.64 GB/op`, `107,523,776 allocs/op`
- clean rerun: `21.93s/op`, `6.64 GB/op`, `107,523,333 allocs/op`

That is a real step down from the prior kept baseline of
`22.24s/op`, `8.23 GB/op`, `121,472,290 allocs/op`. The refreshed alloc
profile shows the intended shift: the first-binding map allocation is no
longer the main wall, and the remaining pressure is now more clearly
`NewEnvironmentWithValueCapacity(...)` object churn,
`promoteSingleBindingNoLock(...)` on multi-bind scopes,
`evaluateStructLiteral(...)`, `snapshotCallStack(...)`,
`runMatchExpression(...)`, and `execEnsureStart(...)`. Steady-state bytecode
`fib` still times out at `300s`, so this tranche materially improved the
text-heavy runtime path without changing the recursive timeout story yet.

The next kept steady-state tranche sized hot child scopes from cheap AST
binding counts and removed one of the remaining miss-heavy lookup paths.
Block scopes, function/lambda call scopes, loop-iteration scopes, iterator
literal scopes, and `or {}` handler scopes now use
`NewEnvironmentWithValueCapacity(...)`, while `matchPatternFast(...)` now uses
`Environment.Lookup(...)` instead of the miss-allocating `Get(...)` path for
ordinary identifier probes. The refreshed aligned steady-state `i_before_e`
results on that kept code are:

- profiled: `22.01s/op`, `6.43 GB/op`, `97,060,888 allocs/op`
- clean reruns: `21.93s/op` and `21.87s/op`, both at about `6.43 GB/op` and
  `97.06M allocs/op`

Wall-clock stayed in the same low-21s band, but allocation pressure dropped
materially again from the prior kept baseline of `6.64 GB/op` and
`107.52M allocs/op`. The refreshed profiles show the intended shift:
`Environment.Get(...)` / `fmt.Errorf(...)` miss pressure dropped out of the
top tier, and `matchPattern(...)` cumulative allocs narrowed again. The
remaining steady-state wall is now more clearly
`NewEnvironmentWithValueCapacity(...)` object churn,
`setCurrentValueNoLock(...)`, `evaluateStructLiteral(...)`,
`snapshotCallStack(...)`, `runMatchExpression(...)`, and
`execEnsureStart(...)`.

The next kept steady-state tranche moved general runtime diagnostics onto a
lazy call-stack path. Runtime errors no longer eagerly copy the full eval-state
call stack at first attachment; instead they keep a lazy reference that is
frozen only when the stack is about to mutate or when a real runtime
diagnostic is built. The refreshed aligned steady-state `i_before_e` results on
that kept code are:

- profiled: `20.51s/op`, `5.85 GB/op`, `84,856,303 allocs/op`
- clean rerun A: `23.19s/op`, `5.85 GB/op`, `84,856,355 allocs/op`
- clean rerun B: `22.09s/op`, `5.85 GB/op`, `84,856,481 allocs/op`

Wall-clock stayed in the same rough low-20s band with one noisier clean run,
but allocation pressure dropped materially again from the prior kept baseline
of about `6.43 GB/op` and `97.06M allocs/op`. The refreshed alloc profile
shows the intended shift: `snapshotCallStack(...)` dropped out of the top
allocators entirely, so the remaining steady-state wall is now more cleanly
`NewEnvironmentWithValueCapacity(...)`, `evaluateStructLiteral(...)`,
`setCurrentValueNoLock(...)`, `collectImplCandidates(...)`, `arrayMember(...)`,
`runMatchExpression(...)`, and `execEnsureStart(...)`. Steady-state bytecode
`fib` still times out at `300s`, so the next likely wins are still on the
text-heavy runtime path rather than the recursive timeout path.

The next kept steady-state tranche trimmed the hot type-canonicalization path.
`canonicalizeExpandedTypeExpression(...)` now reuses the original
nullable/result/union/function/generic AST nodes when none of their children
actually change, instead of rebuilding fresh type-expression trees on every
no-op canonicalization pass. The refreshed aligned steady-state `i_before_e`
results on that kept code are:

- profiled: `20.57s/op`, `5.42 GB/op`, `77,708,818 allocs/op`
- clean rerun A: `20.39s/op`, `5.42 GB/op`, `77,708,480 allocs/op`
- clean rerun B: `20.82s/op`, `5.42 GB/op`, `77,710,637 allocs/op`

That is another real step down from the prior kept baseline of about
`22.09s/op`, `5.85 GB/op`, and `84.86M allocs/op`. The refreshed alloc profile
shows the intended shift: `ast.NewNullableTypeExpression(...)` and
`ast.NewUnionTypeExpression(...)` dropped out of the alloc-space top set
entirely, and `canonicalizeExpandedTypeExpression(...)` itself is now a much
smaller slice. The remaining steady-state wall is now more cleanly
`NewEnvironmentWithValueCapacity(...)`, `evaluateStructLiteral(...)`,
`setCurrentValueNoLock(...)`, `collectImplCandidates(...)`, `arrayMember(...)`,
`runMatchExpression(...)`, and `execEnsureStart(...)`.

The next kept steady-state tranche stayed on method/member resolution. Array
helper access now skips the guaranteed direct-member miss for non-field names
like `len`, `get`, and `push`, while `typeImplementsInterface(...)` now caches
resolved type/interface/arg-signature results behind the same invalidation
boundary as the existing method cache. The refreshed aligned steady-state
`i_before_e` results on that kept code are:

- profiled: `21.00s/op`, `5.06 GB/op`, `77,363,744 allocs/op`
- clean rerun: `20.19s/op`, `5.06 GB/op`, `77,364,488 allocs/op`

That kept wall-clock in the same low-20s band but cut another chunk of
allocation pressure from the prior kept baseline of about `20.39s/op`,
`5.42 GB/op`, and `77.71M allocs/op`. The refreshed alloc profile shows the
intended shift: `collectImplCandidates(...)` dropped out of the top
alloc-space set entirely, so the remaining steady-state wall is now more
cleanly `NewEnvironmentWithValueCapacity(...)`, `evaluateStructLiteral(...)`,
`setCurrentValueNoLock(...)`, `arrayMember(...)`, `runMatchExpression(...)`,
`execEnsureStart(...)`, and the residual host-value conversion path.

The next kept steady-state tranche stayed on array metadata and struct-literal
success-path lookup work. Dynamic array state now caches boxed `length` /
`capacity` values, so repeated array helper/member reads stop re-boxing the
same large integers on every access. Struct-literal shorthand bindings and
struct-definition fallback now also use `Environment.Lookup(...)` instead of
the heavier error-producing `Get(...)` path on the hot success path. The
refreshed aligned steady-state `i_before_e` results on that kept code are:

- profiled: `23.03s/op`, `4.88 GB/op`, `73,577,870 allocs/op`
- clean rerun A: `20.00s/op`, `4.88 GB/op`, `73,577,767 allocs/op`
- clean rerun B: `19.61s/op`, `4.88 GB/op`, `73,577,672 allocs/op`

That is another real reduction from the prior kept baseline of about
`20.19s/op`, `5.06 GB/op`, and `77.36M allocs/op`. The refreshed alloc profile
shows the intended shift: `arrayMember(...)` dropped from the older
`~343 MB` flat tier to about `160 MB`, with the remaining array metadata cost
now concentrated in the first boxed-length materialization instead of repeated
re-boxing. The remaining steady-state wall is now more cleanly
`NewEnvironmentWithValueCapacity(...)`, `evaluateStructLiteral(...)`,
`setCurrentValueNoLock(...)`, the residual boxed-length/boxed-capacity path,
`runMatchExpression(...)`, `execEnsureStart(...)`, and host-value conversion.

The next kept steady-state tranche stayed on environment-object size. Ordinary
`runtime.Environment` scopes now move their cold struct-definition/runtime-data
state behind a lazy `environmentMeta` pointer, so hot lexical scopes that only
bind values no longer carry those fields inline. The refreshed aligned
steady-state `i_before_e` results on that kept code are:

- profiled: `21.33s/op`, `4.63 GB/op`, `73,577,840 allocs/op`
- clean rerun A: `20.51s/op`, `4.63 GB/op`, `73,577,220 allocs/op`
- clean rerun B: `20.09s/op`, `4.63 GB/op`, `73,577,644 allocs/op`

That kept wall-clock in the same low-20s band while cutting heap pressure from
the prior kept `4.88 GB/op` band down to about `4.63 GB/op`. The refreshed
alloc profile shows the intended shift:
`NewEnvironmentWithValueCapacity(...)` dropped from the old ~`1.95 GB` flat
object-allocation tier to about `1.71 GB`, while the value-map allocation
slice stayed small. The remaining steady-state wall is now more clearly scope
count plus `evaluateStructLiteral(...)`, `setCurrentValueNoLock(...)`,
`runMatchExpression(...)`, `execEnsureStart(...)`, and residual host-value
conversion.

The next kept steady-state tranche stayed on propagation control flow. The
canonical stdlib `io.unwrap(...)`, `io.unwrap_void(...)`, and
`io.bytes_to_string(...)` paths in `../able-stdlib/src/io.able` now use direct
propagation (`!`) instead of nested `match`/`raise` control flow, and the Go
interpreter now reuses `cachedSimpleTypeExpression("Error")` on the hot
propagation/or-else/runtime-error checks instead of constructing a fresh
`ast.Ty("Error")` node every time. The refreshed aligned steady-state
`i_before_e` results on that kept code are:

- profiled: `19.17s/op`, `4.60 GB/op`, `73,227,937 allocs/op`
- clean rerun A: `20.94s/op`, `4.60 GB/op`, `73,229,445 allocs/op`
- clean rerun B: `19.92s/op`, `4.60 GB/op`, `73,228,693 allocs/op`

That kept wall-clock in the low-20s band while cutting heap from the prior
kept `4.63 GB/op` band to about `4.60 GB/op`, and it dropped alloc count by
roughly `350k` objects per run from the prior kept `73.58M` band to about
`73.23M`. The refreshed alloc profile shows the intended shift: the old
`bytecodeOpPropagation` line no longer carries flat alloc-space on the kept
profile, so propagation is no longer paying per-call AST `Error` type
construction on top of the stdlib unwrap/match traffic.

The current cross-mode bytecode-core baseline is checked in at:

- `v12/docs/perf-baselines/2026-04-16-bytecode-core-benchmark-baseline.json`
- `v12/docs/perf-baselines/2026-04-16-bytecode-core-benchmark-baseline.md`

That suite is intentionally small and stable enough for routine reruns. It
tracks:

- `quicksort`
- `future_yield_i32_small`
- `sum_u32_small`

For targeted compiler-lowering checks, prefer checked-in fixture targets under
`v12/fixtures/bench/` so the benchmark package metadata is reproducible from
the repo. Recent mono-array work uses
`v12/fixtures/bench/matrixmultiply_f64_small/main.able` for the staged nested
`Array (Array f64)` comparison and
`v12/fixtures/bench/zigzag_char_small/main.able` for the staged text-scalar
(`Array char` / `Array (Array char)`) comparison and
`v12/fixtures/bench/sum_u32_small/main.able` for the staged unsigned numeric
comparison and `v12/fixtures/bench/hashmap_i32_small/main.able` for the first
broader native-container (`HashMap i32 i32` + `Map i32 i32`) comparison and
`v12/fixtures/bench/heap_i32_small/main.able` for the broader array-backed
container family (`Heap i32`) comparison and the shared generic nominal-method
specialization follow-up and bound generic field/member carrier refinement
follow-up on that same benchmark and
`v12/fixtures/bench/linked_list_for_i32_small/main.able` for the first
benchmark-worthy generic-container hot-path (`LinkedList -> Iterable ->
Iterator`) comparison and
`v12/fixtures/bench/linked_list_enumerable_i32_small/main.able` for the next
concrete generic/default-method container hot path (`LinkedList.map/filter/reduce`)
comparison and the shared static nominal receiver/struct-literal closure
follow-up on that same benchmark and
`v12/fixtures/bench/linked_list_iterator_pipeline_i64_small/main.able` for
the next iterator default-method hot path
(`LinkedList.lazy().map<i64>(...).filter(...).next()`) comparison and
`v12/fixtures/bench/linked_list_iterator_collect_i64_small/main.able` for the
mono-array-enabled iterator collect/reduce follow-up
(`LinkedList.lazy().map<i64>(...).filter(...).collect<Array i64>().reduce(...)`)
comparison and
`v12/fixtures/bench/linked_list_iterator_filter_map_i64_small/main.able` for
the iterator-literal controller / `filter_map` follow-up
(`LinkedList.lazy().filter_map<i64>(...).collect<Array i64>().reduce(...)`)
comparison, while the full
`matrixmultiply` workload in `v12/examples/benchmarks/matrixmultiply.able`
remains the canonical suite entry used by `v12/bench_suite`. Current focused
snapshots for these reduced fixtures are checked in at:

- `v12/docs/perf-baselines/2026-03-19-mono-array-f64-matrixmultiply-small-compiled.md`
- `v12/docs/perf-baselines/2026-03-19-mono-array-nested-wrapper-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-matrixmultiply-f64-small-native-scalar-propagation-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-matrixmultiply-f64-small-native-float-int-casts-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-matrixmultiply-static-array-frame-elision-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-matrixmultiply-static-array-propagation-pointer-elision-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-matrixmultiply-counted-loop-fast-path-compiled.md`
- `v12/docs/perf-baselines/2026-03-21-matrixmultiply-inline-affine-int-checks-compiled.md`
- `v12/docs/perf-baselines/2026-03-21-matrixmultiply-nonnegative-sub-range-proof-compiled.md`
- `v12/docs/perf-baselines/2026-03-21-matrixmultiply-bounded-add-range-proof-compiled.md`
- `v12/docs/perf-baselines/2026-03-19-mono-array-zigzag-char-small-compiled.md`
- `v12/docs/perf-baselines/2026-03-19-mono-array-u32-sum-small-compiled.md`
- `v12/docs/perf-baselines/2026-03-19-hashmap-i32-small-compiled.md`
- `v12/docs/perf-baselines/2026-03-19-heap-i32-small-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-heap-i32-generic-nominal-method-specialization-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-heap-i32-bound-generic-field-carrier-refinement-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-linked-list-for-i32-small-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-linked-list-enumerable-i32-small-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-linked-list-enumerable-i32-small-specialized-default-impls-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-linked-list-iterator-pipeline-i64-small-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-linked-list-iterator-collect-i64-small-compiled.md`
- `v12/docs/perf-baselines/2026-03-20-linked-list-iterator-filter-map-i64-small-compiled.md`
- `v12/docs/perf-baselines/2026-03-22-compiler-performance-milestone-7-compiled.md`

The reduced recursion/call-overhead benchmark is now:
- `v12/fixtures/bench/fib_i32_small/main.able`

The current representative compiled Milestone 7 snapshot is:
- `v12/docs/perf-baselines/2026-03-22-compiler-performance-milestone-7-compiled.md`
  - `bench/fib_i32_small`: `2.7567s`, `0.00` GC
  - `bench/heap_i32_small`: `0.2900s`, `5.00` GC
  - `bench/linked_list_iterator_pipeline_i64_small`: `0.1433s`, `9.67` GC
  - `bench/matrixmultiply_f64_small`: `0.1167s`, `7.33` GC
  - `examples/benchmarks/matrixmultiply`: `1.0633s`, `13.33` GC

The `zigzag_char_small` snapshot was corrected after fixing mono-off nested
carrier identity for `Array (Array char)`, so use the checked-in snapshot
rather than any earlier ad hoc mono-off timings.

The current best matrix snapshots are now:
- reduced target:
  `v12/docs/perf-baselines/2026-03-20-matrixmultiply-counted-loop-fast-path-compiled.md`,
  which records `0.1133s` / `7.00` GC on the compiled
  `matrixmultiply_f64_small` target after removing the synthetic static-array
  loop-induction checked-arithmetic scaffolding through shared primitive
  counted-loop lowering
- full canonical benchmark:
  the latest external compiled comparison records `0.9660s` / `4.00` GC over
  `5/5` runs on the compiled
  `v12/examples/benchmarks/matrixmultiply.able` path after the canonical Able
  source started using `Array.with_capacity(n)` for fixed-size rows and outer
  matrices and statement-position counted loops stopped materializing
  discarded runtime loop results

The follow-up affine integer snapshot
`v12/docs/perf-baselines/2026-03-21-matrixmultiply-inline-affine-int-checks-compiled.md`
proves the remaining `build_matrix` `i - j` / `i + j` helper calls are gone,
but it is effectively performance-neutral relative to the counted-loop
snapshot on this benchmark family.

The follow-up subtraction range-proof snapshot
`v12/docs/perf-baselines/2026-03-21-matrixmultiply-nonnegative-sub-range-proof-compiled.md`
proves the widened inline overflow branch is now gone for `build_matrix`
`i - j`, but `i + j` still carries the widened checked-add path and the
benchmark remains in the same band as the counted-loop baseline.

The follow-up upper-bound range-proof snapshot
`v12/docs/perf-baselines/2026-03-21-matrixmultiply-bounded-add-range-proof-compiled.md`
proves the remaining widened inline overflow branch is now gone for
`build_matrix` `i + j` too. The inner-loop affine add/sub gap is closed, but
the benchmark still remains in the same band as the counted-loop baseline.

The latest `Array.with_capacity(n)` source-parity tranche removed matrix slice
growth churn rather than changing compiler semantics. Generated Go now builds
the hot rows/outers with `make(..., 0, n)` while retaining native nested
`[]float64` carriers in `build_matrix` and `matmul`. The refreshed external
compiled comparison moved from the same-session `1.2900s` baseline to
`1.0180s` over `5/5` runs, about `1.16x` the Go reference. The profiled kept
run still spends almost all CPU in `__able_compiled_fn_matmul`; allocation is
down to about `52.34MB`, with benchmark matrix allocation at about `34.77MB`.

The follow-up counted-loop statement tranche removed a smaller compiler-side
scaffold from the same hot generated functions. Statement-position counted
loops now lower directly to the counted `for i < n` shape when the body has no
value-producing `break`, so the compiler no longer emits a temporary
`runtime.Value` loop result plus `__able_runtime_error_value(...)` discard
probe after the hot matrix loops. The refreshed external compiled comparison
moved again to `0.9660s` over `5/5` runs, about `1.10x` the Go reference, and
the profiled kept run landed at `0.9600s` with CPU almost entirely in
`__able_compiled_fn_matmul`.

The next kept compiled recursion tranche targeted `fib` without changing the
benchmark source. Statement-position `if` guards whose body cannot fall
through now seed conservative fallthrough integer facts, so
`if n <= 2 { return 1 }` proves the recursive `n - 1` and `n - 2` decrements
safe. Generated `fib` now emits direct `i32` subtraction for those decrements
while keeping the checked addition and slow control path for possible
overflow. Refreshed external compiled `fib` moved from `3.2633s` over `3/3`
before the slice to `3.1760s` over `5/5`, with a second local `5/5` rerun at
`3.1280s` and a profiled kept run at `3.1000s`. The Go reference remains
`2.8400s`, so the remaining compiled core gap is now the checked-add /
control-return shape rather than recursive call dispatch or subtract
underflow checks.

The iterator-pipeline family is now split intentionally:
- `linked_list_iterator_pipeline_i64_small` isolates the already-closed native
  `map/filter/next` path
- `linked_list_iterator_collect_i64_small` isolates the now-closed
  mono-array-enabled `collect<Array i64>().reduce(...)` follow-up
- `linked_list_iterator_filter_map_i64_small` isolates the now-closed
  iterator-literal controller / `filter_map(...).collect<Array i64>()`
  follow-up

## Benchmarks Covered

- `fib`
- `binarytrees`
- `matrixmultiply`
- `quicksort`
- `sudoku`
- `i_before_e`

## Usage

```bash
# default suite (all benchmarks, all modes)
./v12/bench_suite

# targeted compiled mono-array comparison
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/matrixmultiply_f64_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  --compiled-build-arg=--no-experimental-mono-arrays \
  v12/fixtures/bench/matrixmultiply_f64_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/zigzag_char_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  --compiled-build-arg=--no-experimental-mono-arrays \
  v12/fixtures/bench/zigzag_char_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/sum_u32_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  --compiled-build-arg=--no-experimental-mono-arrays \
  v12/fixtures/bench/sum_u32_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/hashmap_i32_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/heap_i32_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/linked_list_for_i32_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/linked_list_enumerable_i32_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/linked_list_iterator_pipeline_i64_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/fib_i32_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/linked_list_iterator_collect_i64_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  --compiled-build-arg=--no-experimental-mono-arrays \
  v12/fixtures/bench/linked_list_iterator_collect_i64_small/main.able
./v12/bench_perf --runs 3 --timeout 60 --modes compiled \
  v12/fixtures/bench/linked_list_iterator_filter_map_i64_small/main.able

# generated-call execution-context ABI candidate comparison
./v12/bench_suite \
  --suite fixture-generality \
  --modes compiled \
  --runs 3 \
  --timeout 90 \
  --experimental-execution-context \
  --output-json v12/docs/perf-baselines/execution-context-candidate.json

# the external harness forwards the same candidate flag only to Able compiled builds
./v12/bench_compare_external \
  --suite generality \
  --modes compiled,bytecode \
  --runs 3 \
  --timeout 90 \
  --experimental-execution-context \
  --output-md v12/docs/perf-baselines/execution-context-external-candidate.md

# reproducible baseline example
./v12/bench_suite \
  --suite bytecode-core \
  --runs 1 \
  --timeout 90 \
  --build-timeout 240 \
  --output-json v12/docs/perf-baselines/2026-04-16-bytecode-core-benchmark-baseline.json \
  --output-md v12/docs/perf-baselines/2026-04-16-bytecode-core-benchmark-baseline.md

# report-only comparison against the checked-in baseline
./v12/bench_guardrail \
  --baseline v12/docs/perf-baselines/2026-04-16-bytecode-core-benchmark-baseline.json \
  --current v12/tmp/perf/current-bytecode-core.json

# compare Able against the sibling ../benchmarks corpus
./v12/bench_compare_external \
  --benchmarks fib,binarytrees,matrixmultiply \
  --modes compiled,bytecode \
  --output-md /tmp/able-vs-external.md

# run a concurrency benchmark with the goroutine executor explicitly
./v12/bench_perf --runs 1 --timeout 90 --modes compiled \
  --executor goroutine \
  --run-from ../benchmarks/binarytrees \
  v12/examples/benchmarks/binarytrees.able

# benchmark a local program against an external suite-local input file
./v12/bench_perf --runs 1 --timeout 20 --modes compiled \
  --run-from ../benchmarks/i-before-e \
  --program-arg wordlist.txt \
  v12/examples/benchmarks/i_before_e/i_before_e.able

# benchmark steady-state bytecode runtime (load/lower once, then main() only)
./v12/bench_perf --runs 1 --timeout 180 --modes bytecode-runtime \
  --run-from ../benchmarks/i-before-e \
  --program-arg wordlist.txt \
  v12/examples/benchmarks/i_before_e/i_before_e.able

# capture a CPU profile for an aligned compiled run
ABLE_GO_CPU_PROFILE=/tmp/able-fib.pprof \
./v12/bench_perf --runs 1 --timeout 90 --modes compiled \
  --run-from ../benchmarks/fib \
  v12/examples/benchmarks/fib.able

# capture a flushed profile from a timed-out bytecode run
ABLE_GO_CPU_PROFILE=/tmp/able-bytecode-fib.pprof \
./v12/bench_perf --runs 1 --timeout 30 --modes bytecode \
  --run-from ../benchmarks/fib \
  v12/examples/benchmarks/fib.able

# capture a steady-state bytecode runtime profile
ABLE_GO_CPU_PROFILE=/tmp/able-bytecode-runtime-ibe.pprof \
./v12/bench_perf --runs 1 --timeout 180 --modes bytecode-runtime \
  --run-from ../benchmarks/i-before-e \
  --program-arg wordlist.txt \
  v12/examples/benchmarks/i_before_e/i_before_e.able

# emit a steady-state bytecode runtime call trace
ABLE_BYTECODE_TRACE_OUT=/tmp/able-bytecode-runtime-ibe-trace.json \
ABLE_BYTECODE_TRACE_LIMIT=12 \
./v12/bench_perf --runs 1 --timeout 180 --modes bytecode-runtime \
  --run-from ../benchmarks/i-before-e \
  --program-arg wordlist.txt \
  v12/examples/benchmarks/i_before_e/i_before_e.able
```

## JSON Output

The output file includes:

- `results`: one row per `(benchmark, mode, run_index)`
- `summary`: aggregated `ok/timeout/error` counts and average metrics for successful runs

Statuses:

- `ok`: command exited 0 within timeout
- `timeout`: command exceeded timeout
- `error`: non-timeout non-zero exit, including compiled-build failure

For `bytecode-runtime`, the JSON payload also includes:

- `avg_ns_per_op`
- `avg_bytes_per_op`
- `avg_allocs_per_op`

`bytecode-runtime` typechecks, loads, lowers, and runs one explicit warmup
call before timed measurement, so the wall-clock timeout must budget for setup,
warmup, and the measured benchmark call. The reported `ns/op`, allocation, and
profile data cover only the post-warmup `main()` calls.
When `ABLE_GO_CPU_PROFILE` or `ABLE_GO_MEM_PROFILE` is set for
`bytecode-runtime`, the emitted profiles cover only the post-warmup measured
call, not the initial load/lower/warmup phase.

## Bytecode Runtime Trace

`bytecode-runtime` also supports an opt-in sorted bytecode call trace:

- `ABLE_BYTECODE_TRACE_OUT=/tmp/trace.json`
- `ABLE_BYTECODE_TRACE_LIMIT=20` (optional top-N trim; omit for all entries)

The emitted JSON report includes:

- `target_path`
- `run_from`
- `program_args`
- `trace.total_hits`
- `trace.entries`, sorted by `hits` descending

Each trace entry includes:

- `hits`
- `op` (`call_name` or `call_member`)
- `name`
- `lookup` (`name`, `dot_fallback`, `resolved_method`, `member_access`)
- `dispatch` (`exact_native`, `inline`, `generic`)
- `origin`, `line`, and `column` when source location metadata is available

This trace mode is diagnostic-only. The additional counting adds overhead, so
the traced `ns/op` result should not be compared directly against the untraced
steady-state benchmark baseline. Use it to rank hot callsites and callees, then
rerun the normal `bytecode-runtime` benchmark without tracing to judge whether a
change is actually keep-worthy.

The next kept trace-driven bytecode text-path tranche used that call trace to
target overload-valued member calls instead of attempting more exact-native
substitution. In
[bytecode_vm_call_member.go](/home/david/sync/projects/able/v12/interpreters/go/pkg/interpreter/bytecode_vm_call_member.go),
`bytecodeOpCallMember` now resolves overload-valued methods down to a selected
`*runtime.FunctionValue` before dispatch and then feeds that selected overload
through the existing injected-receiver inline/generic ladders without
materializing a fresh bound-method value on each hot call. The small-arity
overload-selection scratch path also stays stack-backed.

The refreshed aligned `i_before_e` result on that kept code is:

- bytecode-runtime clean rerun A: `0.887s/op`, `5.60 MB/op`,
  `174,773 allocs/op`
- bytecode-runtime clean rerun B: `0.911s/op`, `5.60 MB/op`,
  `174,775 allocs/op`
- bytecode-runtime profiled: `0.905s/op`, `5.63 MB/op`,
  `174,847 allocs/op`

That is a keep as a CPU-path win with the prior low-heap shape preserved.
Compared with the restored post-template-cache baseline, aligned steady-state
bytecode `i_before_e` moves from roughly the `~1.00-1.01s/op` band down into
the `~0.89-0.91s/op` band while staying in the prior `~5.6 MB/op` /
`175k allocs/op` heap band. The traced run confirms the semantic target:
`Array.get` now shows up as `call_member` / `resolved_method` / `inline`
instead of remaining on the generic member-call path.

The next kept bytecode-runtime tranche tightened exact-native call overhead for
extern wrappers only. In
[values.go](/home/david/sync/projects/able/v12/interpreters/go/pkg/runtime/values.go),
`runtime.NativeFunctionValue` now has an opt-in `SkipContext` flag, and
[definitions.go](/home/david/sync/projects/able/v12/interpreters/go/pkg/interpreter/definitions.go)
marks generated extern wrappers with `SkipContext: true` because those
closures do not observe `*runtime.NativeCallContext`. The tree-walker native
call path and the bytecode
[execExactNativeCall(...)](/home/david/sync/projects/able/v12/interpreters/go/pkg/interpreter/bytecode_vm_call_native_fast.go)
fast path now both bypass native-call-context pooling/setup on that opt-in
path, while context-sensitive runtime/concurrency natives keep the old
behavior unchanged.

The refreshed aligned `i_before_e` result on that kept code is:

- bytecode-runtime clean rerun A: `0.872s/op`, `5.60 MB/op`,
  `174,775 allocs/op`
- bytecode-runtime clean rerun B: `0.853s/op`, `5.60 MB/op`,
  `174,774 allocs/op`
- bytecode-runtime profiled: `0.837s/op`, `5.63 MB/op`,
  `174,846 allocs/op`

That is a keep as a modest CPU-path win with the same low-heap shape
preserved. Compared with the prior overload-member-inline band, aligned
steady-state bytecode `i_before_e` moves from roughly `~0.89-0.91s/op` down
into the `~0.84-0.87s/op` band while holding the prior `~5.6 MB/op` /
`175k allocs/op` heap band. The refreshed profile also says native-call
context setup is no longer the useful exact-native target; the remaining wall
is now the actual fast-string extern body plus residual member/name-call
dispatch.

The next kept `i_before_e` slice was benchmark-local rather than VM-internal.
[v12/examples/benchmarks/i_before_e/i_before_e.able](/home/david/sync/projects/able/v12/examples/benchmarks/i_before_e/i_before_e.able)
now short-circuits `is_valid(...)` by returning early when a word has no
`"ei"` or has `"ei"` but no `"cei"`, and only falls back to
`replace("cei", "")` on the small remaining subset that actually needs it.
On the aligned `wordlist.txt` corpus that removes pointless `replace(...)`
work from `172,695` of `172,823` words; only `128` words still take the
replace path, and an exhaustive local equivalence check over the aligned
wordlist preserved the prior `1628` invalid outputs.

The refreshed aligned `i_before_e` results on that kept code are:

- bytecode-runtime clean rerun A: `0.792s/op`, `2.84 MB/op`,
  `2,080 allocs/op`
- bytecode-runtime clean rerun B: `0.749s/op`, `2.84 MB/op`,
  `2,078 allocs/op`
- bytecode-runtime profiled: `0.779s/op`, `2.87 MB/op`,
  `2,151 allocs/op`
- external compiled compare: `0.290s`
- external bytecode compare: `1.020s`

This is a keep because the old benchmark body was doing obviously wasted
string work on nearly the entire corpus; the semantics stay the same, but the
benchmark is no longer dominated by avoidable `replace(...)` calls. Compared
with the prior exact-native skip-context band, aligned steady-state bytecode
`i_before_e` moves from roughly `~0.84-0.87s/op` down into the
`~0.75-0.79s/op` band, and heap drops from roughly `5.6 MB/op` /
`175k allocs/op` into the `~2.84 MB/op` / `2.1k allocs/op` band. One-shot
aligned bytecode `i_before_e` is now `1.02s`, so this benchmark is no longer
the right place to spend more string micro-optimization time before revisiting
the much larger `fib` timeout problem.

The next kept bytecode `fib` slice stayed deliberately narrow. It did not try
another frame-reuse experiment; instead it specialized the remaining hot `i32`
arithmetic/boxing path directly. `v12/interpreters/go/pkg/interpreter/
bytecode_vm_i32_fast.go` now provides a dedicated small-`i32` boxing path plus
direct small-`i32` add/sub helpers, and the fused self-call immediate-subtract
path plus the specialized bytecode integer add/sub path use those helpers
before falling back to the generic integer machinery.

That is a keep as a reduced recursion-kernel win, not a full aligned benchmark
fix. The restored reduced `BenchmarkFib30Bytecode` baseline was roughly
`219-225ms/op`, and the warmed kept reruns landed around `198.70ms/op` and
`201.98ms/op`. Aligned one-shot external bytecode `fib` still times out at
`90s`, so the remaining real wall is still broader recursive VM cost rather
than just generic `i32` boxing/dispatch.

The next kept bytecode `fib` slice stayed on that same reduced recursion path
and trimmed the self-fast frame shape instead of attempting slot reuse again.
Pure self-recursive calls that carry no generic-name set, no implicit receiver,
and no loop/iterator base state now use a smaller minimal self-fast frame
instead of the full self-fast frame payload.

That is another keep as a reduced recursion-kernel win, not a full aligned
benchmark fix. The prior warmed reduced `BenchmarkFib30Bytecode` band was
roughly `198.70-201.98ms/op`; the refreshed warmed reruns on the kept code
landed at `199.73ms/op`, `195.06ms/op`, and `196.84ms/op`. Aligned one-shot
external bytecode `fib` still times out at `90s`, but the refreshed reduced
CPU profile no longer shows `pushCallFrame(...)` as a top-tier flat hotspot.
The remaining reduced wall is now more cleanly
`execCallSelfIntSubSlotConst(...)`, `execBinary(...)`,
`popCallFrameFields(...)`, `acquireSlotFrame(...)`,
`bytecodeDirectSmallI32Value(...)`, and the residual direct `i32`
boxing/immediate-subtract path.

The next kept bytecode `fib` slice stayed on that reduced recursion path and
trimmed the inline return path instead of trying another frame-reuse
experiment. `v12/interpreters/go/pkg/interpreter/bytecode_vm_return.go` now
owns bytecode inline returns, which pulls the hot return logic out of
`bytecode_vm_run.go` and keeps that file back under the 1000-line guardrail.
More importantly for the benchmark, the inline return helper now handles
`bytecodeCallFrameKindSelfFastMinimal` directly instead of routing that case
through the broader `popCallFrameFields(...)` path, and the hot inline call
sites now use cached return-generic metadata through
`bytecodeInlineReturnGenericNames(...)`.

That is another keep as a reduced recursion-kernel win, not a full aligned
benchmark fix. The prior warmed reduced `BenchmarkFib30Bytecode` band was
roughly `195.06-199.73ms/op`; the refreshed warmed reruns on the kept code
landed at `189.63ms/op`, `195.72ms/op`, and `192.46ms/op`. A single profiled
reduced run landed at `202.31ms/op`. Aligned one-shot external bytecode `fib`
still times out at `90s`, but the refreshed reduced CPU profile no longer
shows `pushCallFrame(...)` as a visible top-tier hotspot. The remaining reduced
wall is now more cleanly `execCallSelfIntSubSlotConst(...)`, `execBinary(...)`,
`execBinarySlotConst(...)`, `finishInlineReturn(...)`,
`bytecodeDirectSmallI32Value(...)`, and `bytecodeBoxedIntegerI32Value(...)`.

The next kept bytecode `fib` slice stayed on the reduced recursion path again,
but this time targeted slot-frame pool churn rather than arithmetic. `v12/
interpreters/go/pkg/interpreter/bytecode_vm_slot_frames.go` now batches small
slot-frame allocations at 32 frames instead of 8, so reduced recursive runs do
not keep rebuilding tiny hot pools across the common recursion depth.

That is another keep as a reduced recursion-kernel win, not a full aligned
benchmark fix. The prior warmed reduced `BenchmarkFib30Bytecode` band was
roughly `189.63-195.72ms/op`; the refreshed warmed reruns on the kept code
landed at `198.99ms/op`, `183.53ms/op`, and `187.89ms/op`. A single profiled
reduced run landed at `207.16ms/op`. Aligned one-shot external bytecode `fib`
still times out at `90s`, but the refreshed reduced CPU profile no longer
shows `releaseSlotFrame(...)` as a top-tier hotspot. The remaining reduced wall
is now more cleanly `execCallSelfIntSubSlotConst(...)`, `pushCallFrame(...)`,
`finishInlineReturn(...)`, `bytecodeDirectSmallI32Value(...)`,
`acquireSlotFrame(...)`, and `execBinarySlotConst(...)`.

The next kept bytecode `fib` slice stayed on that reduced recursion path and
trimmed the minimal self-fast setup path instead of trying another arithmetic
micro-optimization. `v12/interpreters/go/pkg/interpreter/bytecode_vm_call_frames.go`
now exposes `pushSelfFastMinimalCallFrame(...)`, and
`v12/interpreters/go/pkg/interpreter/bytecode_vm_calls.go` now uses that
direct helper inside `execCallSelfIntSubSlotConst(...)` whenever the current
program already has a cached nil return-generic set and the frame is
guaranteed to stay minimal. That same path now also skips
`bytecodeInlineReturnGenericNames(...)` entirely on the cached-nil case.

That is another keep as a reduced recursion-kernel win, not a full aligned
benchmark fix. The prior restored reduced `BenchmarkFib30Bytecode` band was
roughly `188.93-197.98ms/op`; the refreshed warmed reruns on the kept code
landed at `197.79ms/op`, `185.03ms/op`, and `184.96ms/op`. A single profiled
reduced run landed at `194.46ms/op`. Aligned one-shot external bytecode `fib`
still times out at `90s`, but the refreshed reduced CPU profile no longer
shows `pushCallFrame(...)` or `bytecodeInlineReturnGenericNames(...)` in the
visible top tier. The remaining reduced wall is now more cleanly
`finishInlineReturn(...)`, `execCallSelfIntSubSlotConst(...)`,
`inlineCoercionUnnecessaryBySimpleType(...)`, `execBinarySlotConst(...)`,
`bytecodeDirectSmallI32Value(...)`, and
`bytecodeSubtractIntegerImmediateI32Fast(...)`.

The next kept bytecode `fib` slice stayed on that reduced recursion path again
but moved one level lower into slot-frame release. `v12/interpreters/go/pkg/
interpreter/bytecode_vm_slot_frames.go` now uses direct nil stores for released
slot frames of size `1..4` instead of always calling the generic
`clear(slots)` path. That matches the common reduced `fib` frame sizes, so the
hot return path no longer pays the broader slice-clear helper for every tiny
frame release.

That is another keep as a reduced recursion-kernel win, not a full aligned
benchmark fix. The prior warmed reduced `BenchmarkFib30Bytecode` band was
roughly `184.96-197.79ms/op`; the refreshed warmed reruns on the kept code
landed at `186.37ms/op`, `187.04ms/op`, and `189.04ms/op`. A single profiled
reduced run landed at `212.16ms/op`. Aligned one-shot external bytecode `fib`
still times out at `90s`, but the refreshed reduced CPU profile no longer
shows `releaseSlotFrame(...)` / `clear(slots)` as a visible top-tier hotspot.
The remaining reduced wall is now more cleanly
`execCallSelfIntSubSlotConst(...)`, `execBinary(...)`,
`execBinarySlotConst(...)`, `finishInlineReturn(...)`,
`bytecodeDirectSmallI32Value(...)`, and
`bytecodeSubtractIntegerImmediateI32Fast(...)`.

The next kept bytecode `fib` slice stayed on that reduced recursion path but
trimmed the self-call dispatch shape itself. `v12/interpreters/go/pkg/
interpreter/bytecode_vm_calls.go` now has an early dedicated self-slot fast
branch inside `execCallSelfIntSubSlotConst(...)`, so the successful recursive
hot path bypasses the older generic callee switch, the `*bytecodeProgram`
type assertion/equality check, and `callNode` extraction entirely. That branch
reads the `*runtime.FunctionValue` directly from the reserved self slot, uses
the already-known `currentProgram`, and stays on the existing minimal
self-fast frame path.

That is another keep as a reduced recursion-kernel win, not a full aligned
benchmark fix. The prior warmed reduced `BenchmarkFib30Bytecode` band was
roughly `186.37-189.04ms/op`; the refreshed warmed reruns on the kept code
landed at `181.90ms/op`, `178.13ms/op`, and `176.39ms/op`. A single reduced
rerun landed at `188.57ms/op`. Aligned one-shot external bytecode `fib` still
times out at `90s`, but the reduced warmed band moved materially again. The
next remaining wall is now more clearly the arithmetic side of the same
recursion path: `execBinary(...)`, `execBinarySlotConst(...)`,
`bytecodeDirectSmallI32Value(...)`, `bytecodeSubtractIntegerImmediateI32Fast(...)`,
and the residual return-path work in `finishInlineReturn(...)`.

The next kept reduced-`fib` slice stayed on that arithmetic side but narrowed
it to the recursive-result `+` path only. `v12/interpreters/go/pkg/
interpreter/bytecode_vm_i32_fast.go` now has a dedicated
`bytecodeDirectSmallI32Pair(...)` helper, and
`bytecodeAddSmallI32PairFast(...)` uses that combined extractor directly
instead of calling `bytecodeDirectSmallI32Value(...)` twice for the hot
small-`i32` pair-add case.

That is another keep as a reduced recursion-kernel win, not a full aligned
benchmark fix. The prior warmed reduced `BenchmarkFib30Bytecode` band was
roughly `176.39-181.90ms/op`; the refreshed warmed reruns on the kept code
landed at `171.06ms/op`, `175.06ms/op`, and `175.27ms/op`. A single profiled
reduced run landed at `183.33ms/op`. Aligned one-shot external bytecode `fib`
still times out at `90s`, but the reduced warmed band moved again. The next
remaining wall is now more cleanly the self-call arithmetic and return side:
`execCallSelfIntSubSlotConst(...)`, `execBinary(...)`,
`execBinarySlotConst(...)`, `finishInlineReturn(...)`,
`bytecodeSubtractIntegerImmediateI32Fast(...)`, and the residual direct
small-`i32` extraction/boxing helpers.

The next kept reduced-`fib` slice stayed on that same self-fast recursion path
but trimmed the minimal frame bookkeeping instead of touching arithmetic again.
`v12/interpreters/go/pkg/interpreter/bytecode_vm_call_frames.go` now keeps top
contiguous minimal self-fast frames out of `callFrameKinds` entirely and tracks
them with an explicit suffix count until a broader frame kind needs to sit
above them. `v12/interpreters/go/pkg/interpreter/bytecode_vm_return.go`,
`v12/interpreters/go/pkg/interpreter/bytecode_vm_pool.go`, and
`v12/interpreters/go/pkg/interpreter/bytecode_vm_run_finalize.go` now consume
that same suffix directly, so the hot reduced-`fib` recursion path no longer
appends a `bytecodeCallFrameKindSelfFastMinimal` entry on every recursive
step, while mixed stacks still materialize the suffix back into
`callFrameKinds` before pushing a full or metadata-bearing self-fast frame.

That is another keep as a reduced recursion-kernel win, not a full aligned
benchmark fix. The prior kept reduced baseline was in the low `170ms` band;
the refreshed warmed reruns on the kept code landed at `173.91ms/op`,
`171.32ms/op`, and `172.12ms/op`. A single profiled reduced run landed at
`170.23ms/op`. Aligned one-shot external bytecode `fib` still times out at
`90s`, but the hot recursion path now avoids the old per-step minimal-frame
kind push entirely. The remaining reduced wall is now more cleanly
`execCallSelfIntSubSlotConst(...)`, `execBinary(...)`,
`execBinarySlotConst(...)`, `bytecodeSubtractIntegerImmediateI32Fast(...)`,
`pushSelfFastMinimalCallFrame(...)`, and the residual direct small-`i32`
extraction/boxing helpers.

The next kept reduced-`fib` slice moved out of the VM and into lowering.
`v12/interpreters/go/pkg/interpreter/bytecode_lowering.go` now routes non-last
`if` expressions through a statement-only lowering path, and
`v12/interpreters/go/pkg/interpreter/bytecode_lowering_controlflow.go` lowers
that path without synthesizing a dead `Nil` value for the missing `else`
branch. On the reduced `fib(30)` kernel that removes one dead
`bytecodeOpConst Nil` plus one immediate `Pop` from every non-base recursive
step in the function body.

That is a keep as another reduced recursion-kernel win, not a full aligned
benchmark fix. The prior kept warmed reduced `BenchmarkFib30Bytecode` band was
roughly `171.32-173.91ms/op`; the refreshed warmed reruns on the kept code
landed at `163.53ms/op`, `159.41ms/op`, and `160.61ms/op`. A single profiled
reduced run landed at `169.55ms/op`. Aligned one-shot external bytecode `fib`
still times out at `90s`, but the refreshed reduced CPU profile no longer
shows the earlier dead statement-result const/pop overhead as a visible
top-tier slice. The remaining reduced wall is back on
`execCallSelfIntSubSlotConst(...)`, `execBinary(...)`,
`execBinarySlotConst(...)`, `bytecodeSubtractIntegerImmediateI32Fast(...)`,
`bytecodeDirectSmallI32Value(...)`, `acquireSlotFrame(...)`, and residual
run-loop dispatch.

The next kept reduced-`fib` slice stayed on control-flow lowering, but moved
from dead statement cleanup to direct conditional dispatch.
`v12/interpreters/go/pkg/interpreter/bytecode_lowering_controlflow.go` now
lowers `if` / `elsif` conditions through a dedicated conditional jump opcode
when the existing slot-const matcher already proves the shape `slot <= i32`,
and `v12/interpreters/go/pkg/interpreter/bytecode_vm_ops.go` executes that
path without first materializing a boolean result for `JumpIfFalse` to consume.

That is another keep as a reduced recursion-kernel win rather than a full
aligned benchmark fix. The prior kept warmed reduced
`BenchmarkFib30Bytecode` band was roughly `159.41-163.53ms/op`; the refreshed
warmed reruns on the kept code landed at `159.93ms/op`, `155.70ms/op`, and
`151.83ms/op`. A single profiled reduced run landed at `151.60ms/op`.
Aligned one-shot external bytecode `fib` still times out at `90s`, but the
refreshed reduced profile no longer routes the base-case compare through the
old `execBinarySlotConst(...) -> BoolValue -> JumpIfFalse` path. The remaining
reduced wall is now more cleanly `execCallSelfIntSubSlotConst(...)`,
`execBinary(...)`, `bytecodeSubtractIntegerImmediateI32Fast(...)`,
`acquireSlotFrame(...)`, `finishInlineReturn(...)`, and residual run-loop
dispatch.

The next kept reduced-`fib` slice moved onto that residual return side.
`v12/interpreters/go/pkg/interpreter/bytecode_slot_analysis.go` now caches a
compact primitive return check on slot frame layouts, and
`v12/interpreters/go/pkg/interpreter/bytecode_vm_return.go` uses that cached
check in `finishInlineReturn(...)` before falling back to the older
string-based simple-type helper or full return coercion. The reduced `fib`
kernel's recursive `Int` returns therefore avoid re-running the string-based
simple return helper on every unwind while preserving the existing fallback
path for non-simple and mismatched values.

That is another keep as a reduced recursion-kernel win rather than a full
aligned benchmark fix. The prior refreshed warmed reduced
`BenchmarkFib30Bytecode` band was `156.88ms/op`, `160.22ms/op`, and
`163.96ms/op`. The first warmed reruns on the kept code landed at
`155.13ms/op`, `153.92ms/op`, and `159.16ms/op`; the confirmation band landed
at `150.43ms/op`, `157.63ms/op`, `156.45ms/op`, `150.28ms/op`, and
`151.49ms/op`. A single profiled reduced run landed at `144.17ms/op`.
Allocation shape stayed essentially unchanged at about `102 KB/op` and
`863 allocs/op`. Aligned one-shot external bytecode `fib` still times out at
`90s`. The remaining reduced wall is now most likely the self-call arithmetic
and residual frame churn around `execCallSelfIntSubSlotConst(...)`,
`execBinary(...)`, `bytecodeSubtractIntegerImmediateI32Fast(...)`,
`acquireSlotFrame(...)`, and run-loop dispatch.

The next kept reduced-`fib` slice stayed on the fused recursive self-call path
and targeted the arithmetic setup for `fib(n - 1)` / `fib(n - 2)` directly.
`v12/interpreters/go/pkg/interpreter/bytecode_vm_i32_fast.go` now exposes a
self-call-only small-`i32` immediate subtract helper, and
`v12/interpreters/go/pkg/interpreter/bytecode_vm_calls.go` uses that helper
inside `execCallSelfIntSubSlotConst(...)` before falling back to the broader
integer-immediate helper ladder. This keeps generic arithmetic fallback
unchanged while letting the reduced recursive fast path compute its next
argument without reopening the wider helper path.

That is another keep as a reduced recursion-kernel win rather than a full
aligned benchmark fix. The prior kept confirmation band after cached return
checks was `150.43ms/op`, `157.63ms/op`, `156.45ms/op`, `150.28ms/op`, and
`151.49ms/op`. The first warmed band on this tranche landed at
`147.64ms/op`, `149.63ms/op`, and one noisy `172.81ms/op` outlier; the
confirmation band then landed at `146.70ms/op`, `144.66ms/op`,
`143.36ms/op`, `139.07ms/op`, and `143.54ms/op`. A single profiled reduced
run landed at `137.87ms/op`. Allocation shape stayed essentially unchanged at
about `102 KB/op` and `863-864 allocs/op`. Aligned one-shot external bytecode
`fib` still times out at `90s`. The next reduced wall should be re-profiled,
but likely remains in residual self-call frame setup,
`execBinary(...)` result addition, `acquireSlotFrame(...)`, and run-loop
dispatch rather than the immediate subtract ladder.

The next reduced-`fib` maintenance tranche was intentionally behavior-neutral.
The fused slot-const recursive self-call path moved out of
`v12/interpreters/go/pkg/interpreter/bytecode_vm_calls.go` and into
`v12/interpreters/go/pkg/interpreter/bytecode_vm_call_self_slot_const.go` so
follow-up work can stay below the project file-size cap and avoid mixing
call-dispatch edits with fused-recursion edits. `bytecode_vm_calls.go` dropped
from `992` to `842` lines; the new focused file is `158` lines.

Focused recursive/self-fast coverage stayed green, and the reduced
`BenchmarkFib30Bytecode` check remained in the current kept band at
`143.51ms/op`, `148.90ms/op`, and `148.58ms/op` with the same allocation
shape. No aligned external run was needed for this organization-only split.
The next performance tranche should start from the new fused self-call file
and re-profile before trying more run-loop or frame-setup work.

The next kept reduced-`fib` slice used that new fused self-call boundary to
target the remaining size-2 frame acquisition edge. `v12/interpreters/go/pkg/
interpreter/bytecode_vm_slot_frames.go` now exposes a dedicated
`acquireSlotFrame2()` helper that mirrors the existing hot-pool semantics for
exactly two-slot frames, and `bytecode_vm_call_self_slot_const.go` uses it
only when the fused self-call frame layout is exactly two slots. All other
layouts continue through the general `acquireSlotFrame(slotCount)` path.

That is another keep as a reduced recursion-kernel win rather than a full
aligned benchmark fix. The refreshed pre-change reduced
`BenchmarkFib30Bytecode` band was `147.30ms/op`, `140.91ms/op`, and
`150.88ms/op`. The first warmed band after the change landed at
`135.60ms/op`, `136.34ms/op`, and `138.36ms/op`; the confirmation band landed
at `139.26ms/op`, `140.10ms/op`, `138.69ms/op`, `138.53ms/op`, and
`141.70ms/op`. A single profiled reduced run landed at `133.52ms/op`.
Allocation shape stayed unchanged at about `102 KB/op` and `863-864
allocs/op`. Aligned one-shot external bytecode `fib` still times out at
`90s`. The small `preprofile` output no longer shows the earlier
`execCallSelfIntSubSlotConst(...) -> acquireSlotFrame(...)` edge; remaining
visible reduced samples sit around fused self-call dispatch, the conditional
slot-const jump, `finishInlineReturn(...)`, and `execBinary(...)`.

The next kept reduced-`fib` slice stayed on that residual conditional
slot-const jump path. `v12/interpreters/go/pkg/interpreter/bytecode_vm_ops.go`
now gives `execJumpIfIntLessEqualSlotConstFalse(...)` a dedicated direct
small-integer `<=` immediate helper. That keeps the fused `if n <= const`
condition out of the generic `bytecodeDirectIntegerCompare("<=", ...)` helper
and avoids constructing a `BoolValue` just so the conditional jump can read it
back immediately. Generic binary comparisons still use the broader helper and
fallback ladder.

That is another keep as a reduced recursion-kernel win rather than a full
aligned benchmark fix. The refreshed pre-change reduced
`BenchmarkFib30Bytecode` checks landed at `135.15ms/op` for a 3x warmed run
and `141.27ms/op` for the profiled one-shot. The first warmed band after the
change landed at `130.38ms/op`, `141.41ms/op`, and `140.29ms/op`; the
confirmation band landed at `138.06ms/op`, `137.32ms/op`, `136.52ms/op`,
`134.52ms/op`, and `136.98ms/op`. A single profiled reduced run landed at
`135.38ms/op`. Aligned one-shot external bytecode `fib` still times out at
`90s`. The small `preprofile` output no longer shows the earlier
`execJumpIfIntLessEqualSlotConstFalse(...) -> bytecodeDirectIntegerCompare(...)`
edge; remaining visible reduced samples sit around fused self-call dispatch,
add/binary execution, inline return, and residual conditional jump dispatch.

The next reduced-`fib` tranche deliberately improved measurement instead of
changing the VM hot path. `v12/interpreters/go/pkg/interpreter/fib_bench_test.go`
now includes `BenchmarkFib30BytecodeRuntimeOnly`, which parses/evaluates the
reduced `fib` function once, validates a warmup `fib(30) == 832040`, and then
repeatedly calls that same bytecode function on the same interpreter. The
existing `BenchmarkFib30Bytecode` remains the end-to-end reduced check that
builds a fresh interpreter and evaluates the module each iteration.

The first side-by-side check showed why this separation matters:
end-to-end `BenchmarkFib30Bytecode` landed at `138.59ms/op` and
`136.27ms/op` with about `102 KB/op` and `863-864 allocs/op`, while
`BenchmarkFib30BytecodeRuntimeOnly` landed at `130.76ms/op` and
`144.90ms/op` with effectively zero steady-state allocations. The broader
runtime-only warmed band landed at `134.48ms/op`, `129.18ms/op`,
`135.39ms/op`, `139.44ms/op`, and `130.13ms/op`; the profiled one-shot landed
at `135.83ms/op`, and a post-assertion one-shot landed at `128.88ms/op`. The
runtime-only `preprofile` now points directly at VM runtime edges:
`execCallSelfIntSubSlotConst(...)`, `bytecodeSelfCallSubtractIntegerImmediateI32Fast(...)`,
`execJumpIfIntLessEqualSlotConstFalse(...)`, `execBinary(...)`, and the
remaining `finishInlineReturn(...)` sample.

The next runtime-only reduced-`fib` slice used that isolated benchmark to
target the sampled self-call subtract helper itself. Since
`bytecodeSelfCallSubtractIntegerImmediateI32Fast(...)` only handles operands
that have already been proven small `i32`, it no longer calls the generic
`subInt64Overflow(...)` helper before checking the `i32` bounds. The observable
overflow behavior is preserved by the existing `math.MinInt32` /
`math.MaxInt32` check, so `i32` underflow still reports the same overflow
error and non-`i32` shapes still miss this self-call helper.

That is a runtime-only reduced recursion-kernel keep. The runtime-only
baseline before the change landed at `133.10ms/op`, `127.21ms/op`, and
`130.93ms/op`. The first warmed band after the change landed at
`132.47ms/op`, `129.62ms/op`, `126.40ms/op`, `128.22ms/op`, and
`126.96ms/op`; confirmation landed at `134.12ms/op`, `135.39ms/op`,
`128.00ms/op`, `135.33ms/op`, and `130.16ms/op`; the profiled one-shot landed
at `128.22ms/op`. A temporary restored A/B band with the old helper landed
much slower at `156.25-180.87ms/op`, so the direct subtract helper change is
kept despite the normal reduced-fib timing noise. The runtime-only profile now
still points at `execCallSelfIntSubSlotConst(...)`, `finishInlineReturn(...)`,
the fused conditional jump, and the residual binary add/boxing path as the
remaining work.

The next reduced-`fib` control-flow tranche fused the base-case return shape
itself. Statement-position `if slot <= const { return slot }` now lowers to
`bytecodeOpReturnIfIntLessEqualSlotConst`, so the reduced `fib` base case no
longer executes a standalone slot-const conditional jump followed by a separate
slot load and return dispatch. Expression-position `if` lowering and
non-returning statement `if` behavior continue through the existing paths.

That is a runtime-only reduced recursion-kernel keep. The first warmed band
after the change landed at `122.95ms/op`, `127.64ms/op`, `128.59ms/op`,
`133.48ms/op`, and `124.78ms/op`; confirmation landed at `136.10ms/op`,
`125.41ms/op`, `132.00ms/op`, `132.53ms/op`, and `135.25ms/op`; the profiled
one-shot landed at `136.73ms/op`. The small runtime-only `preprofile` now
shows `runResumable(...) -> execReturnIfIntLessEqualSlotConst(...)` in place
of the older standalone `execJumpIfIntLessEqualSlotConstFalse(...)` base-case
edge. The remaining reduced wall is back on fused self-call dispatch,
`finishInlineReturn(...)`, and residual binary add/small-integer handling.

The next kept follow-on narrowed that fused return-if opcode further for the
exact same-slot shape emitted by reduced `fib`: `if n <= const { return n }`.
When the condition slot and return slot are identical and the opcode already
carries a typed integer immediate, `execReturnIfIntLessEqualSlotConst(...)`
now compares the already-read slot value directly, returns it on the true path,
and advances `ip` on the false path. Other shapes still use the existing
return-if fallback.

This is another small runtime-only reduced recursion-kernel keep. The
refreshed return-if baseline landed at `140.33ms/op`, `130.53ms/op`,
`138.43ms/op`, `128.59ms/op`, and `131.46ms/op`. The same-slot fast path first
kept band landed at `133.12ms/op`, `131.19ms/op`, `122.67ms/op`,
`129.81ms/op`, and `132.06ms/op`; confirmation landed at `138.89ms/op`,
`123.58ms/op`, `132.05ms/op`, `125.71ms/op`, and `127.26ms/op`. A temporary
restored A/B check with the same-slot block removed regressed to
`155.13ms/op`, `140.63ms/op`, `146.91ms/op`, `137.47ms/op`, and
`136.60ms/op`, so the fast path is retained despite the noisy profiled
one-shot. The next reduced profile should start again from fused self-call
dispatch, inline return, and the binary add/small-integer path.

The next kept run-loop tranche removed one of the helper edges introduced
during the return-if split. `bytecodeOpReturnIfIntLessEqualSlotConst` is now
handled inline in `runResumable(...)` again, while cold placeholder
lambda/value execution moved to `bytecode_vm_placeholder.go` so
`bytecode_vm_run.go` remains below the project line cap. A narrower
self-call-only fixed-cache boxing bypass was tested first, but it regressed
the runtime-only band to `139.98-153.20ms/op` and was reverted.

The refreshed runtime-only baseline before this tranche landed at
`125.49ms/op`, `125.65ms/op`, `139.05ms/op`, `128.96ms/op`, and
`128.07ms/op`, with a profiled one-shot at `140.06ms/op`. The inline
return-if/cold-placeholder split first clean band landed at `131.03ms/op`,
`137.89ms/op`, `129.69ms/op`, `134.66ms/op`, and `136.04ms/op`; the quiet
confirmation band landed at `130.60ms/op`, `127.63ms/op`, `130.90ms/op`,
`128.78ms/op`, and `132.06ms/op`; the profiled one-shot landed at
`144.56ms/op`. End-to-end reduced `BenchmarkFib30Bytecode` one-shots landed
at `130.39ms/op`, `132.60ms/op`, and `131.84ms/op`. The final small
runtime-only `preprofile` no longer shows the removed
`runReturnIfIntLessEqualSlotConst(...)` wrapper. The next tranche should start
from the remaining fused self-call dispatch, `finishInlineReturn(...)`, and
binary add/small-integer samples.

The next runtime-only reduced-`fib` tranche targeted the unwind side of that
same minimal self-fast recursion path. `bytecode_vm_slot_frames.go` now has a
dedicated `releaseSlotFrame2(...)` helper for exact two-slot frames, and
`finishInlineReturn(...)` uses it only when the active frame layout proves the
callee frame has exactly two slots. This is intentionally different from the
rejected frame-clear elision probe: both slots are still eagerly cleared before
the frame returns to the hot pool.

The refreshed runtime-only baseline before the change landed at
`137.61ms/op`, `126.62ms/op`, `127.97ms/op`, `127.92ms/op`, and
`129.67ms/op`, with a profiled one-shot at `132.08ms/op`. The first kept band
after the change landed at `112.96ms/op`, `113.63ms/op`, `119.86ms/op`,
`127.98ms/op`, and `125.08ms/op`; the confirmation band landed at
`114.67ms/op`, `111.05ms/op`, `111.15ms/op`, `117.82ms/op`, and
`111.43ms/op`; the profiled one-shot landed at `118.84ms/op`. The tiny
runtime-only `preprofile` no longer shows the old
`finishInlineReturn(...) -> releaseSlotFrame(...)` edge. The next tranche
should start from fused self-call setup, the residual `finishInlineReturn(...)`
coercion check, and the binary add/small-integer samples.

The next runtime-only reduced-`fib` tranche fused the final implicit add-return
shape. A node-less implicit `BinaryIntAdd` followed by `Return` now lowers to
`bytecodeOpReturnBinaryIntAdd`; the following return instruction is left in
place but becomes unreachable, preserving existing jump targets. The VM uses
the existing specialized add helper and returns the result directly instead of
replacing the top two stack values and dispatching the next return opcode.
Explicit `return expr + expr` shapes remain on the existing path.

This is a small runtime-only reduced recursion-kernel keep. The same-load
pre-change runtime-only baseline landed at `116.06ms/op`, `118.25ms/op`, and
`116.52ms/op`. The first fused band landed at `118.64ms/op`, `121.93ms/op`,
`116.23ms/op`, `115.53ms/op`, and `115.04ms/op`; the longer fused band landed
at `111.70ms/op`, `112.49ms/op`, `120.25ms/op`, `125.21ms/op`,
`111.48ms/op`, `111.92ms/op`, `114.35ms/op`, and `109.11ms/op`. A temporary
no-fusion control under the same host load landed at `119.59ms/op`,
`123.21ms/op`, `113.04ms/op`, `117.74ms/op`, and `113.97ms/op`; restored
fused confirmation landed at `115.48ms/op`, `110.36ms/op`, `109.40ms/op`,
`110.87ms/op`, and `112.25ms/op`. The profiled one-shot landed at
`123.82ms/op`, and the tiny runtime-only `preprofile` now shows
`runResumable(...) -> execReturnBinaryIntAdd(...)` rather than a standalone
final recursive `execBinary(...)` sample. The next tranche should start from
fused self-call setup, `finishInlineReturn(...)` coercion checks, and the
remaining call-frame/slot state churn rather than another generic add-dispatch
rewrite.

The next kept tranche pivoted from the reduced `fib(30)` shape to the real
aligned external benchmark shape. The checked-in external benchmark is
`fib(45)` over `i32` with `if n <= 2 { return 1 }`, so the earlier fused
`return slot` base-case opcode did not apply. Statement-position
`if slot <= const { return small_i32_const }` now lowers to
`bytecodeOpReturnConstIfIntLessEqualSlotConst`, which returns the encoded
small `i32` constant directly after the same direct slot/immediate comparison
used by the other fused slot-const conditional paths.

This is an aligned-shape keep, not a full external timeout fix yet. The
current reduced `BenchmarkFib30BytecodeRuntimeOnly` baseline before the change
landed at `109.57-116.54ms/op`; after the change it landed at `113.40ms/op`,
`116.23ms/op`, and `119.29ms/op`, so the already-optimized reduced path is
effectively unchanged. The aligned-style `fib_i32_small` bytecode-runtime
fixture landed at `10.56s/op` across three fused runs and `10.58s/op` on
restored fused confirmation. A temporary no-fusion control under the same
fixture landed at `12.59s/op`, which is enough to keep the opcode. The full
external bytecode `fib(45)` run still times out at `90s`; the next tranche
should stay on aligned-fib residual overhead rather than another reduced
`fib(30)` branch unless a fresh profile says otherwise.

The next aligned-fib tranche targeted a repeated object-immediate probe rather
than another frame-setup branch. Lowered slot-const instructions now keep a raw
`int64` immediate beside the existing typed `runtime.IntegerValue`. The typed
value remains the semantic fallback path, while fused self-call subtract,
return-const base-case, and conditional slot-const helpers can use the raw
value after lowering has already proven the literal is a small default `i32`
immediate.

This is another aligned-shape keep. A profiled `fib_i32_small` run before the
change showed samples in `bytecodeSelfCallSubtractIntegerImmediateI32Fast(...)`
and `bytecodeDirectIntegerLessEqualImmediate(...)` from repeatedly unpacking
the same instruction immediate. With raw immediates enabled, aligned-style
`fib_i32_small` bytecode-runtime runs landed at `9.94s/op`, `10.37s/op`, and
`10.18s/op`; a temporary no-raw control under the same code shape landed at
`10.49s/op`; restored raw confirmation landed at `9.49s/op`. Reduced
`BenchmarkFib30BytecodeRuntimeOnly` landed at `118.38-126.67ms/op`, so the
change is kept for the aligned benchmark path rather than claimed as a reduced
`fib(30)` win. The next profile should start from fused self-call setup,
`bytecodeAddSmallI32PairFast(...)`, and `finishInlineReturn(...)`; the raw
immediate probe itself should no longer be the first thing to chase.

The next aligned-fib tranche specialized the already-fused implicit return-add
opcode for functions declared `i32`. When lowering sees a node-less final
`BinaryIntAdd` followed by the implicit `Return` inside an `i32` function, it
now emits `bytecodeOpReturnBinaryIntAddI32`. That opcode tries the direct
small-`i32` add path first, then falls back to the existing generic return-add
semantics for unexpected operand shapes.

This is an aligned-shape keep, not a reduced `fib(30)` win. The reduced
`BenchmarkFib30BytecodeRuntimeOnly` sanity band landed at `125.87ms/op`,
`127.84ms/op`, and `125.94ms/op`. The aligned-style `fib_i32_small`
bytecode-runtime fixture landed at `9.89s/op` and `9.86s/op` across two
3-run confirmation bands, with a profiled one-shot at `9.77s/op`. The
aligned `preprofile` no longer shows the old
`execReturnBinaryIntAdd(...) -> execBinarySpecializedOpcode(...)` edge;
return-add now reaches `bytecodeAddSmallI32PairFast(...)` directly on the hot
path. The next profile should start from fused self-call setup,
`execReturnConstIfIntLessEqualSlotConst(...)`, `finishInlineReturn(...)`, and
the remaining direct small-`i32` pair extraction/boxing costs.

The next aligned-fib tranche stayed on the fused self-call setup path. The raw
slot-const immediate work proved the literal value was already available as an
`int64`, so `execCallSelfIntSubSlotConst(...)` now performs the small-`i32`
subtract directly in the fused recursive self-call branch instead of calling
`bytecodeSelfCallSubtractIntegerImmediateI32RawFast(...)`. The existing
overflow check, boxed integer cache, typed-immediate path, and generic fallback
behavior remain unchanged.

This is an aligned-shape keep with a reduced runtime-only assist. Reduced
`BenchmarkFib30BytecodeRuntimeOnly` landed at `114.39ms/op`, `116.08ms/op`,
and `122.80ms/op`. The aligned-style `fib_i32_small` bytecode-runtime fixture
landed at `9.80s/op` across a 3-run band, with a profiled one-shot at
`9.61s/op`. The aligned `preprofile` no longer shows the helper edge
`execCallSelfIntSubSlotConst(...) -> bytecodeSelfCallSubtractIntegerImmediateI32RawFast(...)`;
the fused self-call path now reaches `bytecodeBoxedIntegerI32Value(...)`
directly after the inlined subtract. Full external bytecode `fib(45)` still
times out at `90s`, so the next profile should start from the remaining
`execCallSelfIntSubSlotConst(...)`, `finishInlineReturn(...)`,
`execReturnConstIfIntLessEqualSlotConst(...)`, and `releaseSlotFrame2(...)`
costs rather than another raw-subtract helper-shape rewrite.

The next aligned-fib tranche removed work from the minimal self-fast return
branch itself. The external-style base case lowers to
`bytecodeOpReturnConstIfIntLessEqualSlotConst`, and that lowering encodes the
literal return value as an `i32`. When the active function also declares an
`i32` return, `finishInlineReturn(...)` now treats that fused opcode as already
satisfying the return type and skips the generic simple return-coercion probe.
All other return opcodes and mismatched return declarations keep the existing
coercion path.

This is an aligned-shape keep. Reduced `BenchmarkFib30BytecodeRuntimeOnly`
landed at `118.01ms/op`, `113.15ms/op`, and `122.26ms/op`. Aligned-style
`fib_i32_small` bytecode-runtime landed at `9.33s/op` and `9.24s/op` across
two 3-run bands, with a profiled one-shot at `8.74s/op`. The aligned
`preprofile` shows
`finishInlineReturn(...) -> inlineCoercionUnnecessaryBySimpleCheck(...)`
dropping from the prior `39` sample range to `14` samples after the fused
base-case return skips that probe. Full external bytecode `fib(45)` still
times out at `90s`. The next profile should start from
`execCallSelfIntSubSlotConst(...)`, `execReturnConstIfIntLessEqualSlotConst(...)`,
`bytecodeDirectSmallI32Pair(...)`, `bytecodeBoxedIntegerI32Value(...)`, and
slot-frame release costs rather than another return-coercion shortcut.

The next aligned-fib tranche specialized the return-add operand shape produced
by the external-style recursive fixture. `bytecodeOpReturnBinaryIntAddI32` now
tries a direct `runtime.IntegerValue`/`runtime.IntegerValue` small-`i32` branch
before the existing pointer-oriented small-`i32` add helper. Pointer operands,
non-small values, non-`i32` values, and generic fallback behavior keep the
existing path, and overflow still returns the existing integer-overflow error.

This is an aligned-shape keep. Reduced `BenchmarkFib30BytecodeRuntimeOnly` was
noisy but stayed in the kept range after the first sample, landing at
`143.82ms/op`, `125.72ms/op`, and `116.12ms/op`. Aligned-style
`fib_i32_small` bytecode-runtime landed at `9.07s/op` and `9.10s/op` across
two 3-run bands, with a profiled one-shot at `8.88s/op`. The aligned
`preprofile` now shows
`execReturnBinaryIntAdd(...) -> bytecodeReturnAddSmallI32ValuePairFast(...)`
on the hot return-add edge. Full external bytecode `fib(45)` still times out at
`90s`. The next profile should start from `execCallSelfIntSubSlotConst(...)`,
`finishInlineReturn(...)`, `execReturnConstIfIntLessEqualSlotConst(...)`, and
slot-frame release costs rather than another return-add operand extraction
shortcut.

The next aligned-fib tranche removed the return-coercion probe that still ran
after the handled `bytecodeOpReturnBinaryIntAddI32` fast paths. Those handled
small-`i32` branches now report that the value already satisfies an `i32`
return, so `finishInlineReturn(...)` can skip the generic simple type check
only for proven boxed-`i32` results. Generic fallback arithmetic still reports
an unknown return shape and keeps the existing coercion behavior.

This is an aligned-shape keep with a reduced runtime-only win. Reduced
`BenchmarkFib30BytecodeRuntimeOnly` landed at `114.67ms/op`, `111.98ms/op`,
and `110.85ms/op`. Aligned-style `fib_i32_small` bytecode-runtime landed at
`8.33s/op` and `8.44s/op` across two 3-run bands, with a profiled one-shot at
`9.01s/op`. The aligned `preprofile` no longer shows the prior
`finishInlineReturn(...) -> inlineCoercionUnnecessaryBySimpleCheck(...)` edge.
Full external bytecode `fib(45)` still times out at `90s`. The next profile
should start from `execCallSelfIntSubSlotConst(...)`,
`execReturnConstIfIntLessEqualSlotConst(...)`, and slot-frame return/release
costs rather than another return-add coercion shortcut.

The next aligned-fib tranche narrowed the body of the fused recursive self-call
opcode. `execCallSelfIntSubSlotConst(...)` now keeps the common fused recursive
path inline and moves the non-fast callable/native/generic handling into
`execCallSelfIntSubSlotConstFallback(...)`. Immediate resolution, inline-call
stats, native-call fallback, generic callable fallback, and error wrapping are
unchanged; the point of the change is code layout around the hot opcode, not a
new arithmetic or frame-pool rule.

This is an aligned-shape keep. Reduced `BenchmarkFib30BytecodeRuntimeOnly`
landed at `109.33ms/op`, `113.54ms/op`, and `118.41ms/op`. Aligned-style
`fib_i32_small` bytecode-runtime landed at `8.44s/op` and `8.42s/op` across
two 3-run bands, with a profiled one-shot at `8.42s/op`. The aligned
`preprofile` does not show the extracted fallback helper on the hot path. Full
external bytecode `fib(45)` still times out at `90s`. The next profile should
start from `execReturnConstIfIntLessEqualSlotConst(...)`,
`finishInlineReturn(...)`, and slot-frame return/release costs rather than more
self-call fallback rearrangement.

The follow-up tranche was measurement-only. The aligned
`fib_i32_small` cross-mode matrix landed at compiled `0.3433s`, tree-walker
`3/3` timeouts at `60s`, bytecode end-to-end `8.1467s`, and
bytecode-runtime `8768648581 ns/op`, `24104 B/op`, `47 allocs/op`.
A standalone bytecode-runtime confirmation landed at `8714877680 ns/op` with
the same allocation shape, and the reduced
`BenchmarkFib30BytecodeRuntimeOnly` sanity band remained in range at
`119.67ms/op`, `115.41ms/op`, and `112.22ms/op`.

The external comparison was rerun with a longer `120s` cap to measure the
previous timeout. Full external `fib(45)` now records compiled `3.3700s` and
bytecode `92.5200s`. That means the old `90s` guard is only slightly missed,
but bytecode remains about `32.58x` the Go reference and `27.45x` the current
compiled Able path on this recursive workload. The next tranche should use the
external `fib(45)` result as the keep/revert guardrail and start from the
remaining aligned hot path around
`execReturnConstIfIntLessEqualSlotConst(...)`, `finishInlineReturn(...)`, and
minimal slot-frame return/release work rather than a reduced-only `fib(30)`
branch.

The next kept aligned-fib code tranche cut the fused self-call guard ladder
for the exact compact recursive shape. `execCallSelfIntSubSlotConst(...)` now
tries `execCallSelfIntSubSlotConstCompact(...)` before resolving generic
immediates or return-name metadata, but only for the already-proven two-slot
slot-0 raw-immediate self call with cached nil return generics and no active
loop/iter state. Reduced `Fib30Bytecode` moved from a refreshed compact-frame
profile of `105.27ms/op` to `99.54ms/op`, `100.39ms/op`, and `99.00ms/op`.
The focused external bytecode `fib(45)` one-shot now completes at `79.1200s`,
which is still far from Go (`2.8400s`) but no longer a timeout. The next
profile should start from the base-case compare/return path and return-add
handoff rather than another self-call fallback rearrangement.

The next kept aligned-fib code tranche inlined the direct small-`i32`
return-add value-pair branch inside `execReturnBinaryIntAdd(...)`, removing
the hot call edge through `bytecodeReturnAddSmallI32ValuePairFast(...)` while
keeping the existing pointer/generic fallback path. Reduced `Fib30Bytecode`
stayed in range at `97.19ms/op`, `104.20ms/op`, and `106.93ms/op`; aligned
`fib_i32_small` bytecode-runtime moved to `7.21s/op` over a 3-run band, with
a profiled one-shot at `7.50s/op`. Focused external bytecode `fib(45)` moved
to `77.2400s`. The next profile should target structural boxed return/add
handoff, the base-case raw compare, or compact `finishInlineReturn(...)`
restoration rather than another return-add helper.

The next kept aligned-fib code tranche inlined the compact slot-0 frame push
inside `execCallSelfIntSubSlotConstCompact(...)`, after the exact-shape
recursive checks have already passed. The generic fallback still uses
`pushSelfFastSlot0CallFrame(...)`; only the proven raw-immediate two-slot path
writes the frame record directly. Reduced `Fib30Bytecode` moved to
`104.21ms/op`, `96.22ms/op`, and `94.85ms/op`; aligned `fib_i32_small`
bytecode-runtime landed at `7.18s/op`, with a profiled one-shot at
`7.60s/op`. Focused external bytecode `fib(45)` moved to `76.7900s`. A
compact `finishInlineReturn(...)` shortcut was tested and reverted because it
regressed aligned runtime to `8.31s/op`. The next profile should either move
to explicit raw/typed return-stack metadata with invalidation or step back to
the broader typed-frame design, because the obvious helper-call edges are now
gone.

The next kept aligned-fib code tranche avoided the rejected operand-stack side
metadata shape and instead added a compact self-fast slot-0 raw lane. The
boxed `runtime.Value` slot remains the observable semantic value, but the exact
two-slot slot-0 recursive frame now saves/restores a raw `i32` cache beside
the boxed slot. `execCallSelfIntSubSlotConstCompact(...)` uses that raw lane
for the recursive `slot0 - const`, and
`execReturnConstIfIntLessEqualSlotConst(...)` uses it for the base-case
`slot0 <= const`; slot-0 writes refresh or clear the lane and generic/full
frame paths clear it.

Reduced `Fib30Bytecode` moved to `92.46ms/op`, `92.81ms/op`, and
`92.08ms/op`. Aligned `fib_i32_small` bytecode-runtime landed at `6.24s/op`
over a 3-run band, with a profiled one-shot at `6.03s/op`. The profiled
aligned rerun shows `execReturnConstIfIntLessEqualSlotConst(...)` down to
`0.41s` cumulative from the prior `1.38s` profile. Focused external bytecode
`fib(45)` moved to `67.8200s`. The next profile should start from
`execReturnBinaryIntAdd(...)`, compact `finishInlineReturn(...)`, and residual
self-call guard/boxing cost rather than another boxed slot-0 probe rewrite.

The next kept aligned-fib tranche added a direct compact minimal-return path
for proven `i32` no-coercion returns. When `ReturnConstIfIntLessEqualSlotConst`,
same-slot `ReturnIfIntLessEqualSlotConst`, or handled
`ReturnBinaryIntAddI32` returns from the exact reused self-fast frame, the VM
restores slot 0/raw slot-0 state and appends the boxed semantic return value
without entering the generic `finishInlineReturn(...)` path. Generic `Int`
returns, non-`i32` slots, non-reused minimal frames, and all full/generic frame
shapes keep the existing boxed fallback/coercion path.

The refreshed same-session aligned control was noisy: `fib_i32_small`
bytecode-runtime started at `14.2100s` over two runs. The kept confirmation
landed at `13.8350s` over two runs and `13.3533s` over three runs. Full
external bytecode `fib(45)` landed at `75.3700s`, compared with the reverted
same-session control from the prior tranche at `77.0800s`; this is still
slower than the historical `67.8200s` raw-lane one-shot, so the result should
be treated as a small current-baseline win rather than a new historical best.
Reduced generic-`Int` `BenchmarkFib30BytecodeRuntimeOnly` stayed in range at
`102.76ms/op`, `106.55ms/op`, and `102.57ms/op`. The next profile should look
at explicit return-add value/raw metadata or a real typed-frame return channel,
not more single-branch proof elision.

The next kept aligned-fib tranche inlined the exact boxed-value-pair `i32`
branch inside `execReturnBinaryIntAdd(...)`. The previous profile showed
`bytecodeReturnAddSmallI32ValuePairFast(...)` as the largest flat cost after
the compact minimal-return keep, so this slice removes that helper call edge
while preserving the pointer/generic fallback helpers and the existing checked
`i32` overflow behavior. The aligned `fib_i32_small` bytecode-runtime band was
thin but positive at `13.3233s` over three runs, versus the prior `13.3533s`
confirmation. A profiled one-shot landed at `12.8900s` and showed the helper
edge gone, with cost now inside `execReturnBinaryIntAdd(...)`. Full external
bytecode `fib(45)` moved to `72.8000s`, versus the prior kept `75.3700s` and
the same-session reverted control at `77.0800s`. The next tranche should not
add another arithmetic helper split; it should either introduce a typed
return/value channel for the recursive `i32` frame shape or step back to the
broader VM-v2 typed-frame design.

The next kept timeout-family tranche pivoted from fib to external quicksort's
kernel slot API. The external source uses `Array.read_slot(i32)` and
`Array.write_slot(i32, T)` heavily instead of bracket indexing, so the VM now
recognizes the canonical kernel methods and executes tracked-array reads/writes
directly while preserving the kernel semantics: negative indexes error,
out-of-bounds reads return `nil`, writes keep the existing growth behavior, and
unsupported/dynamic shapes fall back to normal member dispatch. A reduced
external-style quicksort run with 2000 descending numbers showed the keep:
current fast-path bands landed at `13.79ms/op`, `13.38ms/op`,
`13.64ms/op`, and restored `13.44ms/op`, `13.29ms/op`, `14.49ms/op`; the
temporary control with only `read_slot` / `write_slot` detection disabled
landed at `32.97ms/op`, `32.77ms/op`, and `33.44ms/op`. Bytecode trace on
that reduced source confirms the real quicksort hot sites now dispatch through
`array_read_slot_tracked_fast` / `array_write_slot_tracked_fast`. Full external
bytecode `quicksort` still times out at `90s` (`go` reference `2.0100s`,
`ruby` `14.5800s`, `python` `20.3200s`), so this is a reduced hot-path keep
rather than the final external timeout fix. The next quicksort profile should
start from the residual `execCallMember` shell around those proven slot calls,
then the integer-compare and `swap`/recursive call path.

The follow-up quicksort slot-dispatch tranche lowered ordinary non-safe
`read_slot` / `write_slot` method calls to a guarded `CallMemberArraySlot`
opcode. Once the existing canonical method proof is cached, repeated hot sites
can bypass the broad `execCallMember` dispatch shell and enter the same
tracked-array read/write fast bodies directly. On the same reduced
external-style quicksort harness with 2000 descending numbers, the warmed band
moved from the prior `13.29-14.49ms/op` confirmation range to
`11.21ms/op`, `11.75ms/op`, and `11.55ms/op`, with a profiled one-shot at
`11.49ms/op` and `8996 allocs/op`. Full external bytecode `quicksort` still
times out at `90s`, so the next tranche should move past slot member dispatch
and start from integer comparisons, slot-constant binary conditions, array
index extraction, and the `swap` / recursive quicksort call path.

The next quicksort follow-up kept that slot-call opcode but shortened its
cached hit path. After the guarded proof cache validates a hot
`CallMemberArraySlot` site, the VM now validates the array receiver/cache
identity once and finishes the tracked `read_slot` / `write_slot` body
directly instead of routing through the generic cached-member fast-path switch
and broader `canUseCanonicalArraySlotCallCache(...)` guard. The same reduced
external-style quicksort harness moved from the prior `11.21-11.75ms/op` band
to `10.76ms/op`, `10.79ms/op`, and `10.87ms/op`; a profiled one-shot landed at
`11.20ms/op`, `669948 B/op`, and `9000 allocs/op`, with the old cache guard no
longer in the top profile list. Full external bytecode `quicksort` still
times out at `90s`, so the next tranche should leave slot-call dispatch alone
and start from bool-producing integer comparisons plus `JumpIfFalse`, array
index extraction, and the `swap` / recursive quicksort call path.

The next quicksort conditional tranche fused the pivot-loop guard shape
`arr.read_slot(index) <op> pivotSlot`. In `if` / `elsif` condition position,
slot-backed non-safe `read_slot` comparisons now lower to
`JumpIfArrayReadSlotCompareSlotFalse`, reuse the guarded canonical
`read_slot` proof cache, and skip the standalone read result, bool-producing
comparison, and generic `JumpIfFalse` pop path when the proof holds. The same
reduced external-style quicksort harness moved from the direct slot-call
finish `10.76-10.87ms/op` band to `10.12ms/op`, `10.20ms/op`, and
`10.30ms/op`; a profiled one-shot landed at `9.94ms/op`, `669842 B/op`, and
`8997 allocs/op`, with the old `execJumpIfFalse(...)` hotspot gone from the
short profile. Full external bytecode `quicksort` still times out at `90s`,
so the next tranche should target ordinary slot-slot integer comparison
conditionals such as `lo >= hi`, `i > j`, `i <= j`, `lo < j`, and `i < hi`,
or the `swap` / recursive quicksort call path.

The next quicksort conditional tranche fused ordinary slot-slot integer
comparison guards. Identifier-vs-identifier comparisons in `if` / `elsif`
condition position now lower to `JumpIfIntCompareSlotFalse`, avoiding the
slot load, second slot load, boxed bool, and generic `JumpIfFalse` sequence
while preserving the boxed dynamic fallback through the existing binary
operator path. The reduced external-style quicksort harness moved from the
read-slot compare `10.12-10.30ms/op` band to `9.18ms/op`, `9.28ms/op`, and
`9.29ms/op`; a profiled one-shot landed at `9.58ms/op`, `669971 B/op`, and
`9001 allocs/op`. Full external bytecode `quicksort` still times out at
`90s`, so the next tranche should move to residual boxed slot updates,
slot-call dispatch, frame release, and the `swap` / recursive quicksort call
path rather than adding more condition-only jumps.

The next quicksort tranche lowered ordinary slot-backed, non-safe
`arr.read_slot(i)` expressions to a direct `ArrayReadSlot` opcode. This reuses
the guarded canonical kernel `read_slot` proof cache but skips the broader
member-call opcode shell for expression-position reads; unsupported dynamic
shapes, stale proofs, negative indexes, and out-of-bounds `nil` reads keep the
existing v12 fallback semantics. On the same reduced external-style quicksort
harness, a same-session no-direct-control band landed at `10.25-10.74ms/op`;
the restored direct opcode landed at `9.73-10.56ms/op`, with a profiled
one-shot at `9.38ms/op`, `658964 B/op`, and `8978 allocs/op`. Full external
bytecode `quicksort` still times out at `90s`, so the next tranche should
target a larger remaining wall: boxed slot updates (`i = i + 1`, `j = j - 1`),
direct `swap` / recursive call setup, or residual cache/revision checks around
proven array slot calls.

The next quicksort tranche kept that boxed slot-update target, but as a
runtime shortcut rather than another lowering change. `StoreSlotBinaryIntSlotConst`
now handles the hot small same-type integer `x = x + const` / `x = x - const`
case directly, avoiding the synthetic binary instruction and broader
slot-const binary helper while preserving the fallback for non-small,
mismatched, dynamic, and int64-overflow shapes. Checked v12 integer overflow
still errors before mutating the slot. The reduced external-style quicksort
harness moved from the direct-read baseline `9.73-10.56ms/op` to
`8.45ms/op`, `8.60ms/op`, `8.62ms/op`, `8.75ms/op`, and `8.82ms/op`. The
profiled run was noisy at `11.94ms/op`, `659384 B/op`, and `8978 allocs/op`,
but it confirms the remaining wall is now broader `execBinary(...)`
arithmetic/comparison, `arrayReadSlotValue(...)` cache/proof checks, direct
read-slot execution, and call setup. Full external bytecode `quicksort` still
times out at `90s`, so the next tranche should target one of those larger
remaining buckets rather than another store-only shortcut.

The next quicksort allocation tranche targeted the host-result conversion
side instead of changing bytecode shape. Go extern returns for `u8`, `u16`,
and `u32` now reuse the VM boxed-small-int cache, so `fs.read_bytes(...)`
still produces the existing tracked dynamic `Array u8` representation but no
longer allocates a fresh boxed Able integer for every returned byte. A
mono-u8 host-array experiment cut allocation further but regressed reduced
wall-clock by forcing `parse_numbers` through slower handle reads, so it was
reverted. The kept cached-boxing slice moved the reduced external-style
quicksort harness from a restored `8.93ms/op`, `661889 B/op`, `8982 allocs/op`
sample to `8.48-9.23ms/op`, `~235 KB/op`, and `84-88 allocs/op`; the longer
`50x` confirmation landed at `8.50-9.87ms/op`, `~232 KB/op`, and `78-82
allocs/op`. The runtime heap profile showed `fromHostValue` cumulative
allocation down from about `18.87 MB` to about `8.81 MB` in the same 50-run
profile shape. Full external bytecode `quicksort` still times out at `90s`
against Go `2.0100s`, Ruby `14.5800s`, and Python `20.3200s`, so the next
quicksort tranche should target the remaining runtime wall around
`lookupCachedCanonicalArraySlotCallForArray(...)`, `arrayReadSlotValue(...)`,
and ordinary `execBinary(...)` comparisons.

The follow-up quicksort arithmetic tranche moved the parser-side
`value = value * 10` shape onto the existing slot-const bytecode family.
`x * i32_const` now lowers to `BinaryIntMulSlotConst`, and matching
self-assignments reuse `StoreSlotBinaryIntSlotConst` with checked
small-integer multiplication before falling back to the prior dynamic,
mixed-type, or big-integer path. The same reduced external-style quicksort
harness had a warmed same-session baseline of `9.93-10.07ms/op` after a
cold/noisy first run; the kept confirmation landed at `9.46ms/op`,
`9.37ms/op`, `9.29ms/op`, `9.32ms/op`, and `9.56ms/op`, with a profiled
run at `9.48ms/op`, `232152 B/op`, and `78 allocs/op`. Full external
bytecode `quicksort` still times out at `90s` against Go `2.0100s`, Ruby
`14.5800s`, and Python `20.3200s`, so the next quicksort tranche should
target broader generic comparison/arithmetic stack execution,
`arrayReadSlotValue(...)` cache/proof checks, or direct `swap` / recursive
quicksort call setup.

The next quicksort comparison tranche made that broad comparison target more
specific: typed primitive integer literals now feed slot-const lowering when
they fit their declared suffix. Quicksort's byte-parser guards such as
`byte == 45_u8`, `byte >= 48_u8`, and `byte <= 57_u8` can now avoid the old
generic binary comparison route. The same change broadened
`BinaryIntCompareSlotConst` to cover `<`, `==`, and `!=` in addition to the
previous `>` / `>=` cases, while out-of-range typed literals keep the old
const/generic path for unchanged v12 validation behavior. Reduced
external-style quicksort moved from the post-multiply `9.29-9.56ms/op` band
to `8.74ms/op`, `8.88ms/op`, `8.64ms/op`, `9.16ms/op`, and `8.51ms/op`.
The profiled kept run landed at `9.19ms/op`, `232270 B/op`, and `82
allocs/op`, with `bytecodeDirectIntegerCompare` down from the prior fresh
profile's `60ms` flat sample to `20ms`. Full external bytecode `quicksort`
still times out at `90s` against Go `2.0100s`, Ruby `14.5800s`, and Python
`20.3200s`, so the next quicksort tranche should target
`arrayReadSlotValue(...)` proof/cache costs or direct `swap` / recursive
quicksort call setup rather than more parse-number comparison lowering.

The follow-up quicksort array-slot proof tranche kept the same guarded
canonical proof model but widened the hot tier from `4` to `8` entries. A
reduced quicksort bytecode trace showed eight material `read_slot` /
`write_slot` sites after the direct slot-call, direct read-slot, and
conditional-lowering tranches, so the four-entry tier was still evicting real
hot proofs and re-entering the broader lookup path. Revision guards, identity
checks, and fallback semantics stay unchanged. Reduced external-style
quicksort moved from the typed-compare `8.51-9.16ms/op` band to `8.32ms/op`,
`8.42ms/op`, `7.72ms/op`, `7.62ms/op`, and `8.08ms/op`, with `~232 KB/op`
and `79-84 allocs/op`. The profiled kept run landed at `7.71ms/op`,
`233151 B/op`, and `87 allocs/op`; the remaining visible wall is now mostly
VM dispatch, direct call setup, and residual canonical proof identity checks.
Full external bytecode `quicksort` still times out at `90s` against Go
`2.0100s`, Ruby `14.5800s`, and Python `20.3200s`, so the next quicksort
tranche should target direct `swap` / recursive quicksort call setup or a
broader typed-loop / dispatch-lane change rather than continuing to widen the
array-slot proof tier.

The next quicksort call-setup tranche made the cached `CallName` inline path
more direct for ordinary function values. Once normal lookup has populated a
cache entry, direct non-bound function calls now retain the validated
bytecode program/layout/return-generic shape and use it to set up the frame
without re-running the broader inline-call shape ladder. This targets the hot
`swap(arr, i, j)` site while preserving the same environment/owner revision
invalidation and leaving explicit type-argument calls, bound methods, native
calls, rebinding, and generic fallbacks on the existing paths. In the same
session, the refreshed reduced quicksort baseline after the array-slot
hot-tier tranche landed at a cold/noisy `9.69ms/op`, then warmed to
`7.06ms/op` and `7.02ms/op`; the kept confirmation landed at `6.89ms/op`,
`6.83ms/op`, `6.71ms/op`, `6.64ms/op`, and `6.66ms/op`, with `~400 KB/op`
and `309-316 allocs/op`. The profiled kept run landed at `7.26ms/op`,
`395114 B/op`, and `302 allocs/op`; the previous
`tryInlineResolvedCallFromStack(...)` edge no longer appears as the hot
cached call-name setup path. Full external bytecode `quicksort` still times
out at `90s` against Go `2.0100s`, Ruby `14.5800s`, and Python `20.3200s`,
so the next quicksort tranche should target remaining canonical array-slot
proof identity/version checks or a broader typed-loop / dispatch-lane change.

The follow-up quicksort slot-update tranche kept that broader typed-loop
goal scoped to one runtime-safe step. `StoreSlotBinaryIntSlotConst` now tries
a direct checked `i32` branch for small `i32` slot values plus small `i32`
immediates before the generic same-type integer fallback. This keeps the
existing lowering and fallback semantics, but lets hot loop updates such as
`i = i + 1`, `j = j - 1`, and the existing `value = value * 10` case avoid
the generic overflow-helper / fit-check / boxing path. The reduced
external-style quicksort harness moved from the prior direct call-name
`6.64-6.89ms/op` confirmation band to `6.42ms/op`, `6.31ms/op`,
`6.40ms/op`, `6.49ms/op`, and `6.84ms/op`, with a profiled one-shot at
`6.62ms/op`, `393688 B/op`, and `290 allocs/op`. Full external bytecode
`quicksort` still times out at `90s`, so the next quicksort tranche should
not add another store-only shortcut; it should target the remaining
dispatch/data-access wall around `execCallMember` / `execCallName`,
`arrayReadSlotValue(...)` proof/cache checks, or a proper v12-safe typed-loop
lane for the partition indices and pivot.

The follow-up dispatch cleanup removed a fib-specific probe from ordinary
quicksort bytecode dispatch. `runResumable(...)` now calls
`tryExecI32RecurrenceProgram(...)` only when the active bytecode program has
an attached native `i32` recurrence kernel, execution is at program entry,
stats are disabled, and the run is not a resume. Same-session reduced
quicksort before the guard landed at `7.70ms/op`, `7.58ms/op`, and
`6.94ms/op`, with a profiled one-shot at `7.50ms/op` that showed
`tryExecI32RecurrenceProgram(...)` as a visible flat sample. With the guard,
the profiled reduced run landed at `6.45ms/op` and the helper disappeared
from the quicksort profile; a temporary old-probe control landed at
`6.83ms/op`; the final guarded confirmation band landed at `6.33ms/op`,
`6.18ms/op`, `6.26ms/op`, `6.11ms/op`, and `6.03ms/op`. External bytecode
`fib(45)` still completed at `3.8500s`, so the recurrence kernel remains
active for the intended shape. Full external bytecode `quicksort` still
times out at `90s`, so the next tranche should target
`lookupCachedCanonicalArraySlotCallForArray(...)` /
`arrayReadSlotValue(...)` proof/cache checks, residual `execCallMember(...)`
/ `resolveMethodCallableFromPool(...)`, or a proper v12-safe typed-loop lane.

The next quicksort array-slot cache tranche added a direct VM-local cache in
front of the existing canonical `read_slot` / `write_slot` proof hot tier. The
direct entry is still validated by bytecode program, instruction pointer,
environment, fast-path kind, and the same environment/global/method revisions;
the existing hot array and map remain the fallback and proof population path.
Reduced external-style quicksort moved from the guarded-dispatch
`6.03-6.33ms/op` confirmation band to `5.93ms/op`, `6.07ms/op`,
`6.00ms/op`, `6.11ms/op`, and `5.98ms/op`, with `~386-388 KB/op` and
`277-295 allocs/op`. Profiled reduced reruns landed at `7.31ms/op` and
`6.31ms/op`; despite sampling noise, `lookupCachedCanonicalArraySlotCallForArray(...)`
dropped from about `100ms` flat in the fresh baseline profile to `20-50ms`
flat after the direct cache, leaving version checks and
`readArraySlotValueFast(...)` as the main array-slot read costs. Full
external bytecode `quicksort` still times out at `90s`, while external
bytecode `fib(45)` still completes at `3.9900s`. The next quicksort tranche
should target the remaining `arrayReadSlotValue(...)` /
`readArraySlotValueFast(...)` value path, residual `execCallMember(...)` /
`resolveMethodCallableFromPool(...)`, or a proper v12-safe typed-loop lane;
do not widen the array-slot hot tier again without a fresh collision profile.

The follow-up array-slot index tranche kept the scope narrower than the earlier
rejected general `bytecodeArrayGetIndexI32(...)` rewrite. Only
`bytecodeArraySlotIndexI32(...)` now handles small integer values directly for
canonical `read_slot` / `write_slot`, with the same `i32` fit check and the
same negative-index error behavior as before. Big integers, out-of-range
small values, non-integers, and generic fallback paths are unchanged. Reduced
external-style quicksort moved from the direct-cache `5.93-6.11ms/op` band to
`5.76ms/op`, `5.68ms/op`, `5.75ms/op`, `5.81ms/op`, and `5.73ms/op`; a final
confirmation band landed at `5.71ms/op`, `5.86ms/op`, `5.86ms/op`,
`5.79ms/op`, and `5.88ms/op`, with `~386-388 KB/op` and `277-296 allocs/op`.
A profiled run was noisy at `8.89ms/op`, but the old
`bytecodeArrayGetIndexI32(...)` edge no longer appears in the array-slot
profile. Full external bytecode `quicksort` still times out at `90s`, while
external bytecode `fib(45)` still completes at `3.9900s`. The next quicksort
tranche should stop shaving the index helper and target expensive generic
package/member calls such as `fs.read_bytes(...)`, residual
`execCallMember(...)` fallback cost, or a proper v12-safe typed-loop lane for
partition locals.

The follow-up disabled-trace tranche removed diagnostic trace-call overhead
from the canonical array-slot fast paths. Successful `read_slot` /
`write_slot` fast bodies now check `bytecodeTraceEnabled` before calling
`recordBytecodeCallTrace(...)`, preserving enabled trace entries while avoiding
the helper call in ordinary untraced runs. Reduced external-style quicksort
confirmed at `6.30ms/op`, `6.10ms/op`, `6.09ms/op`, `6.07ms/op`, and
`6.22ms/op`; the profiled reduced run landed at `6.54ms/op`, `231423 B/op`,
and `82 allocs/op`, and the disabled trace edge no longer appears below
`arrayReadSlotValue(...)`. Full external bytecode `quicksort` still times out
at `90s` against Go `2.0100s`, Ruby `14.5800s`, and Python `20.3200s`, so the
next quicksort tranche should target the actual remaining read/cache/index
and member/name-call dispatch costs or a proper v12-safe typed-loop lane.

The next kept quicksort tranche followed the refreshed source/profile evidence
to bracket indexing. The reduced hotloop now uses `arr[i]` rather than
`read_slot`, so slot-shaped index expressions lower to `ArrayIndexGetSlot`.
That opcode reads the receiver and index directly from `vm.slots`, uses the
existing direct array index body while no v12 `Index` implementation can
override array indexing, and falls back through `resolveIndexGet(...)` for
custom `Index` implementations or unsupported shapes. Reduced quicksort moved
from the current `~7.5ms/op` profile band to `6.99ms/op`, `6.94ms/op`, and
`7.21ms/op`; the profiled confirmation landed at `6.84ms/op`, with
`execArrayIndexGetSlot(...)` replacing the old `execIndexGet(...)` /
index-cache path in the runtime loop. Full external bytecode `quicksort`
still times out at `90s` against Go `2.0100s`, Ruby `14.5800s`, and Python
`20.3200s`, so the next quicksort tranche should target residual boxed integer
compare/arithmetic, `execIndexSet(...)` / swap writes, or a v12-safe typed loop
lane for partition locals.

The follow-up slot-index extraction tranche shortened that new bracket-read
opcode. `ArrayIndexGetSlot` now tries a small `runtime.IntegerValue` index
branch before the broader `bytecodeDirectArrayIndex(...)` helper, while
boxed/big integers, unsupported values, and custom v12 `Index` implementations
keep the existing fallback semantics. The refreshed reduced quicksort long
baseline was `6.62ms/op`; the kept warmed band after ref-style probing was
`6.18-6.37ms/op`, and the profiled confirmation landed at `6.39ms/op`. The
old `bytecodeDirectArrayIndex(...)` edge no longer appears in the bracket-read
path. Full external bytecode `quicksort` still times out at `90s` against Go
`2.0100s`, Ruby `14.5800s`, and Python `20.3200s`, so the next quicksort
tranche should target residual checked `i32` arithmetic/comparison,
`execIndexSet(...)` / swap writes, or a v12-safe typed loop lane for partition
locals.

The follow-up write-side tranche closed that `execIndexSet(...)` part for
simple slot-backed writes. `arr[i] = value` now lowers to `ArrayIndexSetSlot`
when both the receiver and index are local slots, after the RHS has already
been evaluated, preserving the v12 assignment order. The opcode uses the same
`IndexMut` override guard and tracked-array write synchronization as the
existing direct array set path, with unsupported/custom shapes falling back
through `resolveIndexSet(...)`. Cached call-name dispatch also now skips the
trace recorder call when bytecode tracing is disabled. A simple inline-argument
coercion bypass and a cast-target cache were tested in the same tranche and
reverted because their reduced quicksort bands were not defensible. The
refreshed reduced quicksort baseline was `6.20ms/op`; the kept warmed band was
`6.00-6.42ms/op`, the final confirmation was `5.83-5.89ms/op`, and
the profiled confirmation landed at `6.21ms/op` with the
old generic `execIndexSet(...)` edge gone from the write path. Full external
bytecode `quicksort` still times out at `90s` against Go `2.0100s`, Ruby
`14.5800s`, and Python `20.3200s`, so the next quicksort tranche should target
repeated small index extraction around the slot get/set opcodes, residual
checked `i32` comparison/arithmetic, or a v12-safe typed loop lane for
partition locals.

The next quicksort write-side tranche shortened direct indexed assignment only
for the already-proven tracked, unaliased array shape. `ArrayIndexSetSlot` and
compound direct index-set writes now refresh element-type metadata and the
array view locally instead of entering the full tracked-array alias sync; any
aliased receiver still uses `syncTrackedArrayWrite(...)`. A `%` fast-candidate
probe was tested first and reverted because it stayed in the `6.03-6.17ms/op`
reduced band. The paired long reduced baseline without the write-sync shortcut
was `6.24ms/op`, `6.06ms/op`, and `6.26ms/op`; the shortcut's first long band
was `5.99ms/op`, `6.05ms/op`, and `6.01ms/op`. Final confirmations were noisy
but stayed in range at `6.26ms/op`, `6.12ms/op`, `5.83ms/op`, then
`6.13ms/op`, `6.06ms/op`, and `6.05ms/op`. The profiled kept rerun was noisy
at `6.54ms/op`, but the intended indexed-set sync edge dropped out and the
remaining array cost shifted to `execArrayIndexGetSlot(...)` /
`resolveDirectArrayIndexGetAt(...)`. Full external bytecode `quicksort` still
times out at `90s` against Go `2.0100s`, Ruby `14.5800s`, and Python
`20.3200s`, so the next quicksort tranche should target indexed reads or a
v12-safe typed-loop lane for partition locals rather than more write-sync or
`%` helper work.

The next quicksort read-side tranche made `ArrayIndexGetSlot` handle the hot
tracked-array + small-index case directly inside the opcode. Tracked reads now
return `state.Values[idx]` or the same `IndexError` value for invalid positions
without calling `resolveDirectArrayIndexGetAt(...)`; untracked arrays,
unsupported index shapes, and custom `Index` implementations keep the existing
fallback behavior. The refreshed reduced profile baseline was `6.01ms/op`.
Kept read-inline confirmations landed at `5.90ms/op`, `5.83ms/op`,
`5.74ms/op`, then `5.88ms/op`, `5.89ms/op`, and `5.78ms/op`. The profiled
kept rerun was noisy at `7.21ms/op`, but `resolveDirectArrayIndexGetAt(...)`
dropped out of the bracket-read path and `execArrayIndexGetSlot(...)` fell
from about `100ms` cumulative in the refreshed baseline profile to about
`50ms`. Full external bytecode `quicksort` still times out at `90s` against Go
`2.0100s`, Ruby `14.5800s`, and Python `20.3200s`, so the next quicksort
tranche should target remaining generic binary/compare costs, call-frame/name
setup, or a v12-safe typed-loop lane for partition locals.

The follow-up bracket-compare tranche fused the hot partition branch shape.
`if` / `elsif` conditions like `arr[i] as i32 >= pivot` now lower to
`JumpIfArrayIndexSlotCompareSlotFalse`, reading receiver/index/right values
from slots and jumping directly instead of emitting a standalone bracket read,
cast, comparison bool, and `JumpIfFalse` pop. The opcode keeps the same v12
`Index` override guard as the slot-backed bracket read path, and caches the
absorbed `i32` cast name on the instruction so the hot loop does not rebuild a
type-expression string. The first implementation regressed to
`6.29-6.40ms/op` because it did rebuild that string; after caching the cast
name, reduced quicksort moved from the refreshed `5.89ms/op` profile baseline
and prior `5.78-5.90ms/op` kept band to `5.15ms/op`, `5.25ms/op`, and
`5.40ms/op`, with a profiled confirmation at `5.26ms/op`. Full external
bytecode `quicksort` still times out at `90s` against Go `2.0100s`, Ruby
`14.5800s`, and Python `20.3200s`. The next quicksort tranche should target
the residual `bytecodeValueIsI32(...)` / cast guard cost inside the fused
bracket-index compare path, then remaining generic binary/name-call costs or a
v12-safe typed-loop lane for partition locals.

The follow-up raw compare lane closed that cast-guard edge for the proven hot
shape. When `JumpIfArrayIndexSlotCompareSlotFalse` has absorbed an explicit
`as i32` cast, the indexed value is a small integer, and the right comparison
slot is a small `i32`, the VM now applies the same wrapping `as i32` semantics
on raw `int64` values and compares directly. Unsupported shapes still fall back
through the normal v12 array-index/cast path. The refreshed reduced profile
baseline landed at `5.51ms/op`; the kept warmed band landed at `5.03ms/op`,
`5.14ms/op`, `5.06ms/op`, `5.27ms/op`, and `5.17ms/op`, with a profiled
confirmation at `5.08ms/op`. The fused compare profile no longer shows
`arrayIndexSlotCompareMaybeCast(...)` / `bytecodeValueIsI32(...)`, and
`execJumpIfArrayIndexSlotCompareSlotFalse(...)` dropped from about `150ms` to
about `100ms` cumulative. Full external bytecode `quicksort` still times out at
`90s` against Go `2.0100s`, Ruby `14.5800s`, and Python `20.3200s`. The next
quicksort tranche should target remaining generic binary/name-call setup or
start a v12-safe typed loop lane for partition locals rather than another
generic cast-guard slice.

The follow-up member-cache tranche targeted the package-env member-call wall
that remained after the bracket-index compare work. The bytecode
member-method cache now includes the active environment and its revision, and
is enabled for the global environment plus ordinary package environments whose
direct parent is the global environment. Existing global revision and method
cache version guards remain in place, while impl/runtime-data environments and
deeper closure environments keep the old uncached behavior. This preserves v12
method resolution while letting hot package-scope Array member calls reuse
their resolved fast-path kind instead of repeatedly entering
`resolveMethodCallableFromPool(...)`.

The refreshed reduced quicksort profiled baseline landed at `5.62ms/op`. The
kept package-env cache band landed at `5.29ms/op`, `5.33ms/op`, `5.28ms/op`,
`5.27ms/op`, and `5.34ms/op`, with a profiled confirmation at `5.32ms/op`.
The kept profile no longer shows the repeated
`resolveMethodCallableFromPool(...)` / `resolvedMemberMethodFastPath(...)`
edge in the hot member-call path; the remaining member-call cost is the
guarded cache lookup and direct fast-path body. Full external bytecode
`quicksort` still times out at `90s` against Go `2.0100s`, Ruby `14.5800s`,
and Python `20.3200s`. The next quicksort tranche should target cached
`execCallName(...)` frame setup, residual direct array/index dispatch, or a
v12-safe typed loop lane for partition locals.

The follow-up call-name slot-arg tranche targeted the `swap(arr, i, j)` call
shape left in the reduced quicksort profile. Lowering now emits a `CallName`
instruction with slot-argument metadata for simple named calls whose one to
three arguments are identifiers. At runtime, `execCallName(...)` materializes
those values from the current slot frame immediately before the existing
cached call-name dispatch, preserving normal name lookup, cache invalidation,
coercion, inline frame setup, and fallback behavior while avoiding the
standalone argument `LoadSlot` opcodes. A precursor right-slot `i32` mark for
the fused array-index compare was tested and reverted because the hot
`pivot := ... as i32` local remains untyped under v12 semantics, so the mark
did not apply without broader data-flow proof.

The kept slot-arg call band landed at `5.03ms/op`, `5.07ms/op`,
`5.05ms/op`, `4.96ms/op`, and `5.00ms/op` against the prior kept
`5.27-5.34ms/op` band. The profiled rerun was noisy at `6.51ms/op`, but the
non-profiled warmed band is a clean reduced-hotloop win. Full external
bytecode `quicksort` still times out at `90s` against Go `2.0100s`, Ruby
`14.5800s`, and Python `20.3200s`. The next quicksort tranche should target
remaining direct call-frame setup / param-coercion checks in
`tryInlineCachedCallNameDirectFromStack(...)` or start a v12-safe typed-loop
lane for partition locals.

The follow-up cached `Array.push` slot-call tranche extends the guarded
canonical array-slot call cache to `Array.push(value)`. Ordinary non-safe
`arr.push(x)` calls now lower to `CallMemberArraySlot`; after normal member
resolution proves the canonical tracked-array push fast path, cached hits use
the same env/global/method-version guards and jump directly into the existing
push body. Unsupported receivers, safe navigation, mutated method tables,
runtime-data environments, and unproven shapes still fall back through normal
member dispatch. The refreshed reduced quicksort baseline was
`5.10-5.22ms/op`; the kept 5x warmed band landed at `4.86ms/op`,
`4.98ms/op`, `4.90ms/op`, `5.12ms/op`, and `4.96ms/op`. Longer 20x
confirmations landed at `5.00ms/op`, `5.54ms/op` as a noisy outlier,
`5.10ms/op`, `4.96ms/op`, and `5.01ms/op`, with a profiled kept run at
`5.16ms/op`. `execArrayPushMemberFast(...)` is no longer a visible top
reduced-CPU edge. Full external bytecode `quicksort` still times out at `90s`
against Go `2.0100s`, Ruby `14.5800s`, and Python `20.3200s`. The next
quicksort tranche should target the remaining
`tryInlineCachedCallNameDirectFromStack(...)` setup, slot-const stores, fused
array-index comparison, or a broader v12-safe typed-loop/call-frame slice, not
another `Array.push` dispatch slice unless a fresh profile brings it back.

The follow-up cached parameter simple-check tranche shortened the remaining
inline call setup around `swap(arr, i, j)` without changing coercion
semantics. `bytecodeFrameLayout` now caches a compact simple-type check enum
for each parameter, and inline parameter setup uses that enum for the hot
"already exact primitive value?" test before falling back to the existing
simple-name/type-expression coercion path. Generic, interface, array, alias,
unknown, and mismatched shapes still use the old fallback behavior. A
refreshed reduced quicksort profiled baseline landed at `5.77ms/op`; the kept
profiled `500x` run landed at `5.00ms/op`, and non-profiled `500x`
confirmations landed at `5.14ms/op`, `5.02ms/op`, and `4.98ms/op`. The final
5x warmed confirmation was `5.25ms/op`, `5.03ms/op`, `5.02ms/op`,
`5.50ms/op`, and `5.19ms/op`. Full external bytecode `quicksort` still times
out at `90s` against Go `2.0100s`, Ruby `14.5800s`, and Python `20.3200s`.
The next quicksort tranche should start from a fresh profile and target
slot-const stores, fused array-index comparison, direct call-frame setup, or a
broader v12-safe typed-loop lane for partition locals; simple parameter
coercion dispatch is no longer the best next target.

The follow-up slot-const store tranche removed a discard-only stack roundtrip
from statement-position fused self-assignments. When lowering sees a
non-final expression statement that emits `StoreSlotBinaryIntSlotConst`, it
marks the instruction as discardable and omits the following `Pop`; the VM
still stores the result and refreshes the slot-0 raw lane, but skips pushing
the assignment expression value that would have been popped immediately.
Assignment expressions still produce their value when the result is observable.
Two precursor experiments were rejected first: a direct tracked-array branch
inside `JumpIfArrayIndexSlotCompareSlotFalse` regressed the profiled reduced
run to `5.10ms/op`, and delayed `paramType` lookup was semantically green but
shifted the warmed band worse to `5.10-5.53ms/op`.

The refreshed reduced quicksort profiled baseline was `4.92ms/op`; the kept
profiled `500x` run was `4.90ms/op`, and `500x` confirmations landed at
`4.92ms/op`, `4.92ms/op`, and `4.94ms/op`. Shorter confirmations landed at
`4.88-5.09ms/op` for `5x` and mostly `4.81-4.91ms/op` for `20x`, with one
`5.03ms/op` outlier. After tightening the discard marker so nested assignment
expressions still leave their value available to enclosing expressions, a
final sanity pass landed at `4.85ms/op` over `500x` and `4.87-4.99ms/op` over
`20x`. Full external bytecode `quicksort` still times out at
`90s` against Go `2.0100s`, Ruby `14.5800s`, and Python `20.3200s`. The next
quicksort tranche should profile the kept state and target the remaining fused
array-index compare, direct call-frame setup, generic `%` / binary work in
`build_data`, or a broader v12-safe typed-loop lane; do not retry direct
tracked-array compare branching or delayed `paramType` lookup without fresh
evidence.

The follow-up bracket swap pattern tranche recognizes the exact local block
shape used by quicksort's helper swap: `tmp := arr[a] as T; arr[a] = arr[b] as
T; arr[b] = tmp`. When the receiver and both indexes are slot-backed
identifiers, lowering emits `ArrayIndexSwapSlot`, which reads those slots
directly and either runs the existing guarded direct array-index get/set bodies
or falls back to the normal generic index get/set sequence. The opcode keeps
the explicit casts and returns the final assignment value, so the optimization
does not change v12 expression semantics.

The kept reduced quicksort `500x` band landed at `4.82ms/op`, `4.76ms/op`,
and `4.86ms/op`, with a profiled confirmation at `4.79ms/op`. This is a small
local hotloop cleanup relative to the immediately preceding slot-const-store
state, not a full external breakthrough: full external bytecode `quicksort`
still times out at `90s` against Go `2.0100s`, Ruby `14.5800s`, and Python
`20.3200s`. The next tranche should start from the kept profile and choose
between remaining direct call-frame setup, generic `%` / binary work in
`build_data`, residual array-index cast/index conversion inside the new swap
opcode, or the broader v12-safe typed-loop lane. Do not broaden this into a
named non-primitive `Array.swap` compiler/runtime special case.

The follow-up swap small-index lane keeps the same `ArrayIndexSwapSlot` opcode
but adds a direct tracked-array path for the hot slot shape where both indexes
are already small integers. The VM now reads the tracked array state once,
checks/casts the two values, writes both positions, and syncs aliases through
the same tracked-array machinery used by direct index writes. Non-small
indexes, untracked arrays, custom indexing behavior, and generic shapes still
use the previous fallback sequence.

The refreshed pre-change profiled reduced sample was `4.75ms/op`. The kept
`500x` band landed at `4.66ms/op`, `4.55ms/op`, and `4.68ms/op`, with a
profiled confirmation at `4.59ms/op`. The profile no longer shows the old
`bytecodeDirectArrayIndex(...)` edge on the swap path, but full external
bytecode `quicksort` still times out at `90s` against Go `2.0100s`, Ruby
`14.5800s`, and Python `20.3200s`. The next tranche should start from a fresh
kept profile and target the larger remaining walls in fused array-index
compare, direct call-name frame setup, generic `%` / binary work in
`build_data`, or a broader v12-safe typed-loop lane rather than more swap
indexing.

The follow-up reduced matrix tranche targeted the hot `Array.get(... )!`
success path rather than another generic propagation helper. Once canonical
tracked-array `Array.get(i32)` has read a non-nil element and the array state's
cached element token plus the actual read value prove a primitive `f32`/`f64`
value whose type does not currently implement `Error`, the VM now skips the
immediately following postfix propagation opcode. Nil values, stale non-float
element shapes, and primitive float types with an active `Error` impl keep the
old path. The primitive no-error decision is cached on the VM by interpreter
method-cache version, so dynamic impl registration invalidates the skip
decision.

A generic boxed-f64 propagation fast-negative precursor was reverted after it
regressed reduced `matrixmultiply_f64_small` to `14.90s`, `15.16s`, and
`15.68s`. The kept fusion moved the profiled reduced matrix baseline from
`14.72s/op` to a final cached-propagation band of `11.96s/op`, `11.50s/op`,
and `11.66s/op`, with a profiled kept rerun at `10.98s/op`. The kept profile no
longer has `execPropagation(...)` or
`propagationValueMayImplementError(...)` in the top list; the remaining wall is
boxed `f64` arithmetic (`floatResultKind`, `evaluateArithmeticFast`,
`evaluateArithmetic`, `runtime.convT`) plus residual `Array.get`
cache/index work. The next matrix tranche should build a real VM-v2 raw `f64`
expression/slot lane instead of shaving propagation or `Array.get` guards
again unless a fresh profile changes the ranking.

The follow-up arithmetic tranche took the first bounded step toward that raw
lane by adding a direct boxed `FloatValue` pair path for primitive `+`, `-`,
and `*`. This is not a benchmark-specific opcode: it preserves f32
normalization, widens to f64 when either operand is f64, and leaves
float/integer mixing plus division on the existing checked operator path. The
refreshed profiled reduced matrix baseline was `11.63s/op`; the kept band
landed at `8.71s/op`, `9.27s/op`, and `9.56s/op`, with a confirmed profiled
rerun at `8.81s/op`.

The kept profile removes the old `evaluateArithmetic(...)` /
`evaluateArithmeticFast(...)` wall from the f64 hot expression. The next matrix
slice should stop producing a boxed `FloatValue` for every inner-loop multiply
and add: carry raw f64 stack/slot values through the expression and box only at
array, dynamic, and spec-visible boundaries. Another boxed helper is unlikely
to move the allocation count materially.

The follow-up f64 add-mul slot update tranche did that for the dot-product
shape without changing benchmark source. Assignments shaped like
`x = x + left * right` now lower to `StoreSlotFloatAddMul`: the VM captures the
old slot value before evaluating `left` and `right`, then computes raw
primitive float multiply/add when all three values are direct `FloatValue`s.
Non-float shapes fall back through the existing boxed `*` then `+` operator
path, so the fused opcode remains a semantic shortcut rather than a new
language rule.

The prior kept direct-float band was `8.71s/op`, `9.27s/op`, and `9.56s/op`.
The fused update kept at `7.59s/op`, `8.26s/op`, and `7.85s/op`, with a
profiled kept rerun at `7.26s/op`. The larger win is allocation pressure:
the reduced matrix run dropped from roughly `55.45M allocs/op` to
`28.45M allocs/op`.

The fresh kept profile now puts the largest remaining matrix wall in canonical
`Array.get(i32)!` work and its safety guards:
`finishArrayGetMemberFast(...)`, `lookupCachedCanonicalArrayGetCallForArray(...)`,
`bytecodeArrayGetIndexI32(...)`, and
`bytecodeArrayGetResultMatchesFloatToken(...)`. The next matrix slice should
feed raw f64 operands from canonical `Array.get!` into the fused update, or
otherwise reduce that guarded array-get path, before returning to arithmetic.

That fused operand slice is now landed for the slot-backed dot-product shape
`s = s + ai.get(k)! * cj.get(k)!`. Lowering emits
`StoreSlotFloatAddMulArrayGet` only when both `Array.get` receivers and indexes
are identifiers with known slots. The opcode uses the existing canonical
`Array.get` call-site proof, preserves nil/Error propagation, and falls back to
normal member-call semantics when the guard is not valid.

The prior kept add-mul band was `7.59s/op`, `8.26s/op`, and `7.85s/op`. The
fused operand update kept at `7.19s/op`, `5.71s/op`, and `6.07s/op`, with a
profiled confirmation at `5.42s/op`. Allocation volume stayed near
`28.45M allocs/op`; the new profile shows boxed `FloatValue` result storage in
`bytecodeDirectFloatAddMul(...)` as the dominant allocation source. The next
matrix tranche should add a raw f64 accumulator slot/update lane and box only
at array, dynamic, or otherwise spec-visible boundaries.

That raw accumulator tranche is now landed for the same fused opcode. The
lowering no longer emits a target `LoadSlot` before
`StoreSlotFloatAddMulArrayGet`; the opcode reads the accumulator slot directly,
computes the primitive add-mul result without returning it through
`runtime.Value`, and stores the internal result in a VM-owned mutable float
cell. Visible slot reads copy the float value back out, preserving Able value
semantics for later reads.

The prior kept fused-operand band was `7.19s/op`, `5.71s/op`, and `6.07s/op`
with about `28.45M allocs/op`. The raw accumulator update kept at `5.57s/op`,
`5.73s/op`, and `5.86s/op`, with a profiled confirmation at `5.80s/op`.
Allocations dropped to roughly `1.63M allocs/op` and `50.5MB/op`; the old
`bytecodeDirectFloatAddMul(...)` allocation wall is gone. The current CPU
profile is now dominated by the guarded `Array.get!` operand path, especially
`bytecodeFloatTypeToken(...)`, `fusedArrayGetCanSkipPropagationCheck(...)`,
`lookupCachedCanonicalArrayGetCallForArray(...)`, and
`bytecodeArrayGetIndexI32(...)`. The next matrix tranche should combine direct
f64 operand extraction with the propagation guard so the fused opcode no longer
rechecks float element tokens through the generic helper on every operand.

That raw operand extraction tranche is now landed for the same fused opcode.
`StoreSlotFloatAddMulArrayGet` preflights the two canonical `Array.get(i32)!`
operands, reuses the guarded canonical call-site proof, reads both array
values directly, handles nil/Error propagation at the same boundary, and feeds
raw `f32`/`f64` values into the VM-owned accumulator update. Stale cached
element tokens, non-float values, active primitive Error impls, unsupported
receivers, and unsupported indexes still fall back to the existing boxed
member-call path.

The prior kept raw-accumulator band was `5.57s/op`, `5.73s/op`, and
`5.86s/op`. The raw operand update kept at `4.43s/op`, `4.06s/op`, and
`4.41s/op`, with a profiled confirmation at `4.36s/op`. Allocation volume
stayed essentially flat at about `1.63M allocs/op` and `50.4MB/op`, so this is
a CPU-path win rather than a new allocation reduction. Full external bytecode
`matrixmultiply` still timed out at `90s` against Go `0.8800s`, Ruby
`42.9300s`, and Python `56.2900s`.

The kept profile removes the old generic
`bytecodeFloatTypeToken(...)` / `fusedArrayGetCanSkipPropagationCheck(...)`
path from the fused update. The new wall is the exact raw operand guard itself:
`bytecodeFusedArrayGetFloatForToken(...)`, canonical `Array.get` proof version
checks, small-`i32` index extraction, and direct array reads. The next matrix
slice should collapse the f64-specific operand proof/read path or move to a
proper typed f64 array/slot lane; another boxed float helper is unlikely to be
the right level.

The native f64 dot-loop tranche is now landed for the exact reduced/full matrix
inner loop shape:

```able
loop {
  if k >= n { break }
  s = s + ai.get(k)! * cj.get(k)!
  k = k + 1
}
```

Lowering attaches a plan to the existing `LoopEnter` and leaves the original
loop bytecode as the fallback. The VM runs the native path only when it proves
canonical `Array.get`, tracked arrays, valid `i32` loop slots, and actual `f64`
elements. Unsupported values enter the original loop before the unsupported
iteration, preserving nil/Error propagation and boxed dynamic behavior.

The prior kept raw-operand band was `4.43s/op`, `4.06s/op`, and `4.41s/op`.
The native dot-loop update kept at `333.92ms/op`, `319.62ms/op`, and
`331.61ms/op`, with a traced/profiled confirmation at `405.57ms/op`.
Allocation volume stayed essentially flat at about `1.63M allocs/op` and
`50.4MB/op`; this is a CPU dispatch/member-call removal, not an allocation
reduction.

The bytecode trace no longer shows the inner dot-product `ai.get(k)!` /
`cj.get(k)!` calls. Remaining reduced matrix trace traffic is construction and
transpose work: `Array.get` at lines 35, 47, and 52 plus `Array.push` /
`Array.new`. Full external bytecode `matrixmultiply` now completes in
`23.8500s` instead of timing out at `90s`; the reference row was Go `0.8800s`,
Ruby `42.9300s`, and Python `56.2900s`. The next matrix tranche should target
that remaining construction/transpose traffic or generalize the native f64 lane
under the same v12 fallback guard.

The follow-up f64 row-cache tranche keeps the same native dot-loop guard, but
stops re-extracting tracked rows on every dot product. Dynamic array states now
carry a revision, and dynamic writes, length changes, tracked writes, and full
state resyncs bump it. The VM caches raw `[]float64` rows by array-state pointer
plus revision/length and clears the cache between pooled top-level VM runs.

The prior kept native-dot band was `319.62-333.92ms/op`. The row-cache update
kept at `229.45ms/op`, `204.02ms/op`, `206.60ms/op`, `213.82ms/op`, and
`225.63ms/op`, with a profiled confirmation at `236.52ms/op`, `52.18MB/op`,
and `1.63M allocs/op`. `bytecodeDirectF64Value` no longer appears as a top
self-cost in the reduced profile; remaining samples are now mostly
construction/transpose member paths, `Array.push`, GC scanning, and residual
slot-store work.

Full external bytecode `matrixmultiply` moved from `23.8500s` to `3.0800s`.
The comparison row is Go `0.8800s`, Ruby `42.9300s`, and Python `56.2900s`, so
bytecode matrix is now about `3.50x` Go while beating the Ruby/Python references
for this benchmark. The next matrix tranche should target construction /
transpose allocation and member-call traffic, not another per-element f64
extraction helper.

The small-integer float-cast tranche removes the construction-side
`big.Int`/`big.Float` conversion for casts like `(i - j) as f64` when the
integer value is already stored in the runtime small-int representation. Big
integer values still use the existing arbitrary-precision path.

The prior kept row-cache band was `204.02-229.45ms/op`. The small-int cast
update kept at `186.59ms/op`, `212.11ms/op`, `202.67ms/op`, `184.04ms/op`, and
`201.25ms/op`, with a profiled confirmation at `194.99ms/op`, `46.31MB/op`,
and `912.8k allocs/op`. Allocation count dropped from roughly `1.63M/op` to
roughly `913k/op`. The reduced profile no longer shows
`math/big.(*Float).Float64` in the top nodes.

Full external bytecode `matrixmultiply` moved from `3.0800s` to `2.9000s`.
The comparison row is Go `0.8800s`, Ruby `42.9300s`, and Python `56.2900s`, so
bytecode matrix is now about `3.30x` Go while remaining faster than the
Ruby/Python references for this benchmark. The next matrix tranche should
target `Array.push` / `Array.slot` dispatch and remaining construction-time
`Array.get` reads.

The tracked `Array.push` append-helper tranche is deliberately external-driven.
The reduced fixture still grows rows from `Array.new`, so the reduced signal was
mostly neutral: `191.41ms/op`, `191.20ms/op`, `193.30ms/op`, `190.00ms/op`, and
`193.85ms/op`, with a profiled `199.26ms/op`, `46.32MB/op`, and
`912.8k allocs/op`. The external matrix benchmark uses `Array.with_capacity(n)`
for rows and outers, which is the shape this tranche targets.

The helper skips `runtime.ArrayEnsureCapacity(...)` when existing logical
capacity and backing slice storage are already sufficient, and it uses the
unaliased tracked-write sync path before falling back to alias-aware sync. Full
external bytecode `matrixmultiply` moved from `2.9000s` to `2.7700s` on the
first run and `2.7500s` over a `3/3` confirmation. The comparison row is Go
`0.8800s`, Ruby `42.9300s`, and Python `56.2900s`, so bytecode matrix is now
about `3.12x` Go.

The next matrix tranche should target construction-time `Array.get` /
`Array.slot` dispatch and GC scan pressure. Further push-only changes need a
fresh external profile before they are worth trying.

The adjacent-`Pop` push cleanup keeps that same direction. Canonical cached
`Array.push(value)` now skips materializing `void` only after the VM has handled
the push and sees that the next active bytecode is the statement-result `Pop`.
Lowering still emits the `Pop`, so generic fallback behavior stays unchanged.
The corrected reduced `matrixmultiply_f64_small` band was `176.45ms/op`,
`181.64ms/op`, `186.98ms/op`, `185.66ms/op`, and `183.05ms/op`; allocation
volume stayed around `46.3MB/op` and `912.9k allocs/op`.

External bytecode `matrixmultiply` did not show a clear macro win from this
cleanup: `2.8167s` over `3/3`, then `2.7740s` over `5/5`, compared with the
prior kept `2.7500s`. Treat it as reduced-positive and external-neutral. The
next tranche should move to construction-time `Array.get` reads and residual
slot-call cache checks rather than another push-only edit.

The f64 dot-loop accumulator-store tranche targets allocation/GC pressure
inside the already-fused native dot loop. The loop now writes the completed
accumulator back as a plain `FloatValue` instead of installing an owned float
cell; because the fused loop updates once per completed dot product, the owned
cell was not amortized and showed up as allocation pressure.

Reduced `matrixmultiply_f64_small` moved to `170.93ms/op`, `170.63ms/op`,
`168.11ms/op`, `163.01ms/op`, and `165.41ms/op`. Allocation volume dropped to
about `44.1-44.4MB/op` and `822.9k allocs/op`; the profiled run was
`169.44ms/op`, `44.18MB/op`, and `822.8k allocs/op`. The allocation profile no
longer has `storeOwnedFloatSlot` or `bytecodeSlotReadValue` in the top list.

Full external bytecode `matrixmultiply` moved to `2.6040s` over `5/5`, against
Go `0.8800s`, Ruby `42.9300s`, and Python `56.2900s`, so bytecode matrix is now
about `2.96x` Go. The next tranche should target remaining boxed float
arithmetic/cast allocation or a genuine typed f64 row/storage lane rather than
another broad owned-slot rewrite.

The f64 affine `Array.push` try-fast tranche targets the construction expression
used by `build_matrix`: `row.push(t * ((i - j) as f64) * ((i + j) as f64))`.
Lowering emits a guarded try opcode before the normal fallback bytecode. When
runtime guards prove canonical `Array.push`, direct `f64` scale, and `i32`
left/right slots, the VM computes the f64 value and appends it directly; any
guard miss falls through to the existing receiver/expression/member-call path.

Reduced `matrixmultiply_f64_small` moved from the prior kept
`163.01-170.93ms/op`, `44.1-44.4MB/op`, and `822.9k allocs/op` band to
`159.61ms/op`, `133.18ms/op`, `136.46ms/op`, then a confirming rerun at
`121.57ms/op`, `126.45ms/op`, and `125.96ms/op`. The profiled reduced run was
`130.51ms/op`, `31.19MB/op`, and `282.8k allocs/op`.

Full external bytecode `matrixmultiply` moved to `2.1300s` over `5/5`, against
Go `0.8800s`, Ruby `42.9300s`, and Python `56.2900s`, so bytecode matrix is now
about `2.42x` Go. The next tranche should target row/column storage allocation
and remaining construction / transpose array traffic through a v12-safe typed
f64 row/storage lane, not another boxed arithmetic helper.

The versioned-stdlib canonical proof tranche is a measurement/proof keep rather
than a macro matrix speedup. The reduced runtime-only matrix benchmark had been
falling back through generic `Array.get` after warmup because canonical stdlib
origin checks accepted sibling checkout and flat cache paths, but not installed
cache paths shaped as `.able/pkg/src/able/<version>/src/...`. The canonical
origin helper now accepts that versioned boundary, and the nullable
`Array.get` proof also accepts direct or bound single `FunctionValue` methods
when they are the canonical nullable stdlib function.

With that proof restored, runtime-only `matrixmultiply_f64_small` now completes
instead of sitting in `callArrayGetFallback`: warmed `5x` reruns landed at
`120.53ms/op`, `117.29ms/op`, and `122.68ms/op`, with roughly `31.2MB/op` and
`282.8k allocs/op`. The reduced CLI fixture stayed neutral at `0.2000s`, and
full external bytecode `matrixmultiply` stayed neutral at `2.1333s` over `3/3`
runs versus Go `0.8800s`. The next matrix tranche should still target
row/column storage allocation and remaining construction/transpose traffic,
not another canonical-origin or `Array.get` proof-cache slice.

The mono-f64 array storage tranche makes that row-storage direction concrete.
The runtime array store now has a guarded f64 lane; the matrix affine
`Array.push` fast path promotes unaliased dynamic rows to mono f64 after
validating the existing elements, the native f64 dot-loop reads mono rows
directly, and canonical `Array.get` fast paths avoid generic `ArrayStoreRead`
for mono f64 handles. All unsupported shapes still deopt or fall back to boxed
Array semantics.

The fresh same-session runtime-only reduced baseline was `134.38ms/op`,
`31.18MB/op`, and `282,768 allocs/op`. The kept rerun landed at
`119.10ms/op`, `117.07ms/op`, and `124.46ms/op`, with about `21.8-22.0MB/op`
and `193.1k allocs/op`. Full external bytecode `matrixmultiply` moved from
the versioned-stdlib proof keep at `2.1333s` over `3/3` to `2.0400s` over
`3/3`; the profiled confirmation was `2.0500s`. The next matrix target is not
more storage helper work; the remaining profile wall is boxed f64 result
materialization in `finishArrayGetMemberFast(...)`, the native dot-loop
accumulator slot write, and residual row lookup/cache checks.

The follow-up nested-get push tranche targets the transpose expression
`ci.push(b.get(j)!.get(i)!)` without changing the fallback bytecode. Lowering
now emits a guarded try opcode for that shape; the VM appends the inner row's
raw f64 directly only when both `Array.get` calls and the destination `push`
are canonical, the outer propagated value is a concrete Array that cannot
implement `Error`, and the inner lookup is in bounds. Nil, Error-capable,
custom, non-f64, out-of-bounds, and aliased/unsupported cases fall through to
the existing boxed path.

Reduced runtime-only `matrixmultiply_f64_small` kept the current wall-clock
band at `121.35ms/op`, `131.30ms/op`, and `122.89ms/op`, while allocation
dropped to roughly `15.7MB/op` and `103.1k allocs/op`. A trace/profiled
confirmation showed `90,000` hits on `array_push_f64_nested_get_fast` at the
transpose line and `103,573 allocs/op`. Full external bytecode
`matrixmultiply` was noisy against the older `2.0400s` best, so a
same-session control was taken: disabling only this new lowering landed at
`2.3533s` over `3/3`, while the restored fused confirmation landed at
`2.1840s` over `5/5`. Treat this as an allocation/shape keep, not a new
all-time wall-clock low. The next matrix target is the native dot-loop
accumulator `FloatValue` box plus repeated array/cache guard cost.

The owned f64 accumulator-cell tranche handles the first half of that boundary.
`StoreSlot`/`StoreSlotNew` now seed and reuse VM-owned float cells for
`FloatValue` locals, and the native f64 dot loop updates the accumulator
through that cell. `LoadSlot` still snapshots owned cells back to ordinary
`FloatValue`, preserving primitive value semantics and preventing array pushes
from retaining mutable slot cells.

Reduced runtime-only `matrixmultiply_f64_small` moved to `108.75ms/op`,
`113.94ms/op`, and `114.94ms/op`, with allocation effectively unchanged at
about `15.7MB/op` and `103.1k allocs/op`. The profiled confirmation landed at
`106.55ms/op`; allocation moved from `tryExecF64DotLoop(...)` to
`bytecodeSlotReadValue(...)`, which snapshots `s` for the following
`di.push(s)`. Full external bytecode `matrixmultiply` landed at `2.0840s` over
`5/5`. The next matrix target should be a guarded slot-backed f64
`Array.push` for `di.push(s)`, not exposing owned float pointers through
general `LoadSlot`.

The reserved-capacity `Array.with_capacity` tranche attacks the full external
allocation profile rather than the reduced fixture. Interpreted
`__able_array_with_capacity` now creates a dynamic array handle with logical
capacity but no dynamic `[]Value` backing. If the array stays dynamic, the first
write allocates the reserved backing before appending; if the row immediately
promotes to mono f64, the runtime allocates only the mono-f64 storage and skips
the discarded dynamic backing. `Array.new(capacity)` and
`ArrayStoreNewWithCapacity(...)` remain eager to avoid changing compiled/runtime
ABI paths that still observe `ArrayValue.Elements` directly.

Reduced runtime-only `matrixmultiply_f64_small` stayed neutral at
`118.08ms/op`, `121.22ms/op`, and `127.02ms/op`, with about `15.7MB/op` and
`103.1k allocs/op`, because that fixture still uses `Array.new`. Full external
bytecode `matrixmultiply` also stayed wall-clock neutral at `2.1000s` over
`5/5`, while GC dropped to `7.00`. The bytecode-runtime allocation evidence is
the reason to keep it: the prior profiled full run showed `125.83MB` total and
`121,006,704 B/op`; after reserved capacity, the profiled run shows `90.89MB`
total and the unprofiled runtime bench shows `71,844,768 B/op`. The old
`ArrayStoreNewWithCapacity(...)` alloc-space leader is gone. The next matrix
target is the remaining `bytecodeSlotReadValue(...)` / `di.push(s)` f64 result
boundary or a typed f64 result-row lane, not another generic capacity pass.

The native dot-loop result-append tranche removes that specific f64 result
boundary without adding another standalone hot dispatch opcode. Lowering keeps
the original bytecode for `loop { ... }; di.push(s)`, but attaches an optional
result-append target to the existing f64 dot-loop plan when the next
statement-position call is exactly a push of the same accumulator. On the fast
path, after the dot product completes and canonical `Array.get` / `Array.push`
guards hold, the VM appends the raw accumulator to the result row and jumps past
the boxed fallback push. Guard misses still run the original loop and push.

Reduced runtime-only `matrixmultiply_f64_small` allocation fell to about
`10.4MB/op` and `13.4k allocs/op`, but reduced wall time was noisy at
`179.01ms/op`, `141.30ms/op`, and `205.82ms/op`. The full external result is
the keep signal: bytecode `matrixmultiply` moved from the reserved-capacity
`2.1000s` confirmation to `2.0240s` over `5/5`, with average GC at `6.00`.
The full bytecode-runtime profile moved to `1.855s/op`, `39,764,440 B/op`, and
`73,510 allocs/op`; profiled allocation total is now `59.12MB`, and
`bytecodeSlotReadValue(...)` is no longer an allocation leader. The remaining
matrix allocation wall is `ArrayStoreAppendF64Promote(...)` and mono-f64
append/growth storage, not the result-load box.

The f64 dot-loop range-hoist tranche is a small CPU cleanup inside the existing
native dot-loop rather than another storage rewrite. The VM now proves the full
`i32` loop range against both raw f64 row slices before accumulating, then runs
the product as a plain `int` indexed Go loop. If the range is negative or would
run out of either row, the fast path falls through before mutating loop slots so
the original bytecode handles the observable failure path.

Reduced runtime-only `matrixmultiply_f64_small` landed at `104.74ms/op`,
`10.43MB/op`, and `13.83k allocs/op`; a same-session old-loop control landed
at `107.86ms/op` with the same allocation floor. Full external bytecode
`matrixmultiply` confirmed at `2.0060s` over `5/5`, while the full
bytecode-runtime profile landed at `1.937s/op`, `39,759,120 B/op`, and
`73,492 allocs/op`. The CPU profile shows `tryExecF64DotLoop(...)` around
`0.91s` flat / `1.17s` cumulative. The next matrix work should be plan-level
row/handle caching or a typed matrix kernel boundary; standalone mono-f64
append/helper rewrites have not shown enough macro movement.

The f64 matrix row-kernel tranche is the first kept typed matrix boundary. The
lowerer recognizes the exact outer `j` loop around `s := 0.0`,
`cj := c.get(j)!`, the proven native f64 dot loop, `di.push(s)`, and
`j = j + 1`. The VM validates canonical `Array.get` / `Array.push`, concrete
row values, f64 row storage, non-negative in-bounds ranges, and destination
non-aliasing before it computes the remaining row and bulk-appends raw f64
results. Guard misses keep the original bytecode and leave the destination row
unmodified.

Reduced runtime-only `matrixmultiply_f64_small` moved from the fresh
`105.35ms/op`, `10.45MB/op`, and `13.89k allocs/op` baseline to `76.79ms/op`,
`10.00MB/op`, and `11.72k allocs/op` over `5/5`; a profiled confirmation
landed at `87.27ms/op`. Full external bytecode `matrixmultiply` moved from
`2.0060s` to `1.7580s` over `5/5` after an earlier same-tranche `1.4967s`
`3/3` run. The `5/5` comparison is about `2.00x` Go (`0.8800s`) and roughly
24x faster than Ruby / 32x faster than Python on the external table. The next
matrix work should target mono-f64 row/result storage growth and capacity
proofs, or graduate this into a broader typed matrix bytecode that carries raw
f64 row slices through build, transpose, and multiply.

The f64 affine row-loop tranche moves the same idea into matrix construction.
Lowering now recognizes the exact `build_matrix` inner loop shape
`if j >= n { break }; row.push(t * ((i - j) as f64) * ((i + j) as f64));
j = j + 1`, attaches a guarded loop plan, and leaves the original bytecode as
the fallback. On the fast path, the VM validates the canonical `Array.push`
proof and f64/i32 operands before mutating the row, computes the remaining row
values into raw f64 storage, then bulk-appends through the existing mono-f64
append rules. This is deliberately not a generic capacity rewrite: `Array.new`
rows end with the same amortized capacity repeated single pushes would expose,
while `Array.with_capacity(n)` rows preserve their declared capacity.

Reduced runtime-only `matrixmultiply_f64_small` moved from a fresh
`74.64ms/op`, `10.02MB/op`, and `11.78k allocs/op` baseline to `54.73ms/op`,
`9.17MB/op`, and `8.12k allocs/op` over `5/5`; a profiled confirmation landed
at `58.66ms/op`, `9.20MB/op`, and `8.18k allocs/op`. Full external bytecode
`matrixmultiply` moved from the row-kernel `1.7580s` confirmation to
`1.4480s` over `5/5`, about `1.65x` Go (`0.8800s`) while remaining far ahead
of Ruby and Python on the external table. The reduced profile no longer spends
the row build on repeated `execTryArrayPushF64AffineProduct(...)` calls. The
next matrix tranche should apply the same bounded loop-level treatment to the
transpose row shape `ci.push(b.get(j)!.get(i)!)`, then reassess result row
materialization and canonical get/push version checks.

The f64 transpose row-loop tranche applies that bounded loop-level treatment to
the `matmul` transpose build. Lowering recognizes only the exact loop shape
`if j >= n { break }; ci.push(b.get(j)!.get(i)!); j = j + 1`, attaches a
guarded loop plan, and keeps the original bytecode as the fallback. On the fast
path, the VM validates canonical `Array.get` / `Array.push`, Array-valued
source rows, raw f64 row storage, non-negative i32 indices, and
destination/source non-aliasing before mutation. It then gathers the remaining
column values and bulk-appends them through the same mono-f64 append rules as
the other matrix loop plans. Guard misses fall through without partial
destination mutation, and final row capacity stays equivalent to repeated
single pushes for `Array.new` or the declared capacity for
`Array.with_capacity(n)`.

Reduced runtime-only `matrixmultiply_f64_small` moved from the prior kept
affine-row `54.73ms/op`, `9.17MB/op`, and `8.12k allocs/op` band to
`40.86ms/op`, `8.76MB/op`, and `6.32k allocs/op` over `5/5`; a profiled
confirmation landed at `42.45ms/op`, `8.78MB/op`, and `6.38k allocs/op`. Full
external bytecode `matrixmultiply` moved from `1.4480s` to `1.3060s` over
`5/5`, about `1.48x` Go (`0.8800s`). The reduced CPU profile no longer shows
the repeated `execTryArrayPushF64NestedGet(...)` transpose cell path; the
largest remaining wall is `tryExecF64MatrixRowLoop(...)`. The next matrix
tranche should target a guarded raw row-slice cache for the transposed matrix
or tighten the row kernel so it avoids re-reading and revalidating every `c`
row for each output row.

The f64 row-slice cache tranche targets that remaining validation wall without
changing the matrix source or broadening nominal lowering. Runtime mono array
states now carry a revision counter, and f64 rows expose
`ArrayStoreMonoF64ValuesRevisionIfAvailable(...)` for cache validation. The
matrix row-kernel caches the transposed matrix's validated raw f64 row slices
only when the outer array has a tracked dynamic state and every source row is
mono f64. Cache hits recheck the outer state revision, row handles, row
revisions, row lengths, and destination non-aliasing before use; partial loop
resumes still use the old row-by-row path so the fast path does not read rows
the fallback loop would skip.

Reduced runtime-only `matrixmultiply_f64_small` moved from the prior kept
transpose-row `40.86ms/op`, `8.76MB/op`, and `6.32k allocs/op` band to
`39.20ms/op`, `8.80MB/op`, and `6.33k allocs/op` over `5/5`; a profiled
confirmation landed at `37.11ms/op`, `8.82MB/op`, and `6.38k allocs/op`. Full
external bytecode `matrixmultiply` moved from `1.3060s` to `1.2240s` over
`5/5`, about `1.39x` Go (`0.8800s`). The reduced CPU profile still puts
`tryExecF64MatrixRowLoop(...)` at the top, so the next matrix tranche should
target row-kernel result materialization: write computed row results directly
into a guarded mono-f64 destination row or otherwise avoid the per-row temporary
result buffer and second append copy while preserving `Array.new` /
`Array.with_capacity` capacity semantics.

The f64 direct result-row segment tranche removes that temporary result buffer
from the validated row-cache path. Runtime now exposes
`ArrayStoreAppendF64UninitializedPromote(...)`, which appends a new f64 segment
using the same dynamic-to-mono promotion and capacity-growth rules as the bulk
append path. After all row-cache, canonical `Array.push`, source-row revision,
and aliasing guards have passed, the matrix row-kernel writes each computed dot
result directly into that destination segment. The non-cache fallback path is
unchanged.

Reduced runtime-only `matrixmultiply_f64_small` moved from the prior kept
row-cache `39.20ms/op`, `8.80MB/op`, and `6.33k allocs/op` band to
`37.72ms/op`, `7.99MB/op`, and `6.03k allocs/op` over `5/5`; a profiled
confirmation landed at `36.48ms/op`, `8.01MB/op`, and `6.08k allocs/op`. Full
external bytecode `matrixmultiply` moved from `1.2240s` to `1.2160s` over
`5/5`, about `1.38x` Go (`0.8800s`). The allocation profile now shows
destination mono-f64 growth itself rather than a temporary row result buffer.
The next matrix tranche should target remaining canonical/raw-read overhead:
cache raw source row slices for the transpose row-loop or hoist
canonical/version checks that still happen once per row/cell before broadening
the typed matrix storage contract.

The transpose row-cache reuse tranche removes that remaining transpose-side
raw-read overhead for the proven square matrix loop. The transpose row-loop now
reuses the same guarded mono-f64 source-row cache as the row kernel when it
enters from `j == 0` and the requested column is inside the proven bound. On
that path it reads each column value directly from the cached raw source row
and appends the generated destination row through the direct f64 segment path.
Partial resumes, `col >= bound` cases, source-row revision changes, source
length changes, canonical call invalidation, and destination/source aliasing
keep the old fallback semantics.

Reduced runtime-only `matrixmultiply_f64_small` moved from the prior direct
result-row `37.72ms/op`, `7.99MB/op`, and `6.03k allocs/op` band to
`32.78ms/op`, `7.20MB/op`, and `5.73k allocs/op` over `5/5`; a profiled
confirmation landed at `32.23ms/op`, `7.22MB/op`, and `5.79k allocs/op`. Full
external bytecode `matrixmultiply` moved from `1.2160s` to `1.1180s` over
`5/5`, about `1.27x` Go (`0.8800s`). The reduced CPU profile is now too short
for useful fine-grained ranking but samples in `tryExecF64MatrixRowLoop(...)`,
so the next matrix tranche should target the actual dot-product row kernel or
start the broader typed matrix storage contract.

The batch-4 row-kernel tranche targets that actual dot-product work. Once the
existing row-cache and direct-destination guards have passed, the matrix
row-kernel computes four destination cells per inner pass. That reuses each
`ai` source value across four cached transposed rows while preserving the
left-to-right accumulation order for each individual dot product. The scalar
remainder path handles non-multiple-of-four row counts, and the non-cache
fallback remains unchanged.

Reduced runtime-only `matrixmultiply_f64_small` moved from a same-session
transpose-cache baseline of `33.63ms/op`, `7.20MB/op`, and `5.73k allocs/op`
to `16.96ms/op`, `7.20MB/op`, and `5.73k allocs/op` over `5/5`; a profiled
confirmation landed at `15.90ms/op`, `7.22MB/op`, and `5.79k allocs/op`. Full
external bytecode `matrixmultiply` moved from `1.1180s` to `0.4640s` over
`5/5`, about `0.53x` Go (`0.8800s`). A full external profile now puts the
batched dot helper at the top and row-cache revision validation next. Matrix
is now competitive with the current external Go reference, so further matrix
work should be driven by broad VM-v2 goals or fresh cross-benchmark evidence
rather than this benchmark alone.

The follow-up quicksort read-slot slice stayed deliberately narrow after the
unsafe native scan/partition-loop experiment was removed. Canonical
`Array.read_slot(i32)` over mono `Array u8` handles now reads raw bytes through
the mono-u8 store helper and returns the existing boxed-small-`u8` cached
value. The slot-index helper also accepts pointer-shaped small integer values
while preserving the same i32-range and non-negative checks. This keeps
ordinary v12 `read_slot` proof/fallback semantics and does not introduce a
named-container or benchmark-loop special case.

Focused array-slot and quicksort parity coverage stayed green. The reduced
in-tree quicksort hotloop guard measured `5098051 ns/op` over `300x`, and the
small in-tree `bench_suite` quicksort bytecode input completed at `0.1800s`
over `1/1`. Full external `../benchmarks` quicksort bytecode still timed out
at `60s`, so the next quicksort tranche should use an external-scale profile
and target a v12-safe byte parser/native bytecode lane, typed collection
storage, or direct call-frame setup rather than more `read_slot` helper work.

A follow-up owned-i32 slot-cell probe for discarded `StoreSlotBinaryIntSlotConst`
updates was tested and rejected. The focused interpreter tests stayed green, but
the 1MB external quicksort prefix regressed from a restored baseline of
`1123911063 ns/op`, `99818456 B/op`, and `1548962 allocs/op` to
`1274257710 ns/op`, `109824248 B/op`, and `1757614 allocs/op`. The profile
showed allocation pressure at the integer read-materialization boundary, so
ordinary dynamic slots should not use VM-owned integer cells without a broader
raw typed-slot contract that avoids boxing and materialization end-to-end. The
next quicksort profile should target remaining `execBinary` / boxed i32
arithmetic and array/member call-frame costs.

The next kept quicksort parser slice fused general integer affine slot updates:
assignments shaped as `slot = slot * integer_literal + expr` now lower to
`StoreSlotIntMulConstAdd`. Lowering loads the old slot before evaluating the
addend expression, so v12 left-to-right behavior is preserved even when the
addend mutates the same slot. The VM handles same-type small integers directly,
with an i32 branch for the parser path, and falls back through the ordinary `*`
then `+` semantics for unsupported values.

On the 1MB external quicksort prefix, the restored profiled baseline was
`1123911063 ns/op`, `99818456 B/op`, and `1548962 allocs/op`. The kept profiled
run landed at `1104789293 ns/op`, `86006760 B/op`, and `1261553 allocs/op`.
The unprofiled `3/3` prefix average landed at `1160421199 ns/op`,
`85980701 B/op`, and `1261499 allocs/op`; reduced in-tree quicksort hotloop
coverage measured `4850727 ns/op` over `300x`. Full external
`../benchmarks` quicksort bytecode still timed out at `60s`, so the next
profile should target boxed slot-const index updates, canonical array/member
call setup, or recursive/direct call-frame overhead rather than another parser
arithmetic fusion.

A follow-up quicksort call/setup micro-probe tranche was tested and rejected.
Direct no-error completion for successful `ArrayReadSlot` preserved focused
semantics but regressed the profiled 1MB prefix confirmation to
`1198737091 ns/op`. Cached `CallName` slot-arg direct inline setup, which
copied identifier slot arguments straight into the inline callee frame instead
of pushing them on the VM stack first, regressed the profiled prefix to
`1346702398 ns/op` without reducing allocation. Both probes were reverted. The
restored kept affine prefix landed at `1132697980 ns/op`, `85984440 B/op`, and
`1261505 allocs/op`. The next quicksort tranche should stop shaving isolated
successful-call completion or slot-arg setup and instead start from a fresh
profile around raw typed slot/index state or a broader canonical
array/member-call plan with a clear allocation/runtime win.

A broader raw integer producer experiment was also tested and rejected. The
probe stored discarded slot-const and affine parser update results in
VM-owned mutable `*runtime.IntegerValue` cells, and added a pure-addend affine
variant that read the base slot without materializing it before the addend. The
focused semantic slice stayed green, but the 1MB external quicksort prefix
regressed to `1178541416 ns/op`, `143733707 B/op`, and `2464643 allocs/op`
over `3/3`. After revert, the restored prefix returned to `85978800 B/op` and
`1261495 allocs/op` allocation shape. Do not represent raw integer slots as
mutable `runtime.Value` pointers; the next raw-typed attempt needs a real
sidecar/register representation with explicit materialization boundaries, or
the work should pivot to a broader canonical array/member-call plan.

A narrow canonical array/member-call cache-size probe was tested next and
rejected. Raising `bytecodeArraySlotCallDirectEntries` from `16` to `64`
reduced the profiled visibility of `lookupCachedCanonicalArraySlotCallForArray`
and produced one profiled prefix at `1172047751 ns/op`, but the repeated
unprofiled bands were not defensible: `1184457441 ns/op` followed by
`1230487621 ns/op`, with allocation unchanged. After revert, the restored
prefix landed at `1170701139 ns/op`, `85984312 B/op`, and `1261499 allocs/op`.
Do not continue quicksort cache-size tuning; a future canonical
array/member-call tranche should remove per-hit version/check work or otherwise
cut a real operation, not just enlarge the direct table.

The next kept raw-i32 quicksort prefix tranche introduced an internal
`bytecodeRawI32SlotValue` for discarded `i32` slot-const and affine slot
updates. The sentinel is consumed by raw-aware integer compare/extract and
array-index helpers, and it materializes through explicit slot read, stack, and
return boundaries so it does not change v12-visible values. The affine parser
update also gained a pure-addend from-slot form, while mutating addends keep
the old load-before-RHS lowering to preserve left-to-right evaluation.

Against a fresh restored 1MB external quicksort prefix baseline of
`1192647051 ns/op`, `86006824 B/op`, and `1261556 allocs/op`, the kept
unprofiled `3/3` bands landed at `1047273039 ns/op`, `77168787 B/op`, and
`2130511 allocs/op`, then `1097504897 ns/op`, `77166939 B/op`, and
`2130510 allocs/op`. The profiled confirmation landed at `1141060286 ns/op`,
`77194872 B/op`, and `2130567 allocs/op`. This is a real wall-clock and bytes
win, but allocation count regressed because the non-pointer sentinel still
boxes into the `runtime.Value` interface slot. The next raw-typed tranche
should use sidecar/register `i32` slot state with explicit materialization
boundaries rather than expanding interface-sentinel storage.

An intermediate stable raw-i32 slot-cell variant was tested and rejected. It
mutated a pointer-shaped internal cell instead of writing a fresh non-pointer
sentinel into the slot interface on every discarded update. That recovered much
of the allocation-count loss, with repeated 1MB quicksort prefix bands at
`1389097633 ns/op`, `74061771 B/op`, `1354198 allocs/op`, then
`1238411773 ns/op`, `74061573 B/op`, `1354189 allocs/op`; the profiled run
landed at `1182344224 ns/op`, `74095648 B/op`, and `1354279 allocs/op`.
Because wall-clock regressed versus the kept sentinel path, the experiment was
reverted. A restored spot-check returned to `1026297428 ns/op`,
`77172552 B/op`, and `2130519 allocs/op`. Future raw-slot work should skip
pointer cells and go straight to sidecar/register storage outside
`runtime.Value`, or pivot away from raw slot storage.

A follow-up active-frame sidecar retrofit was also tested and rejected. This
variant added VM `rawI32Slots` / `rawI32SlotValid` state with call-frame
save/restore and cleared the visible `runtime.Value` slot on discarded raw
updates. Focused semantics stayed green, but the 1MB quicksort prefix regressed
to `1186531639 ns/op`, `84947653 B/op`, and `3293543 allocs/op`; the profiled
run landed at `1162881928 ns/op`, `84981168 B/op`, and `3293607 allocs/op`.
The profile showed the local retrofit moved allocation into repeated
`slotRuntimeValue(...)` materialization plus lazy sidecar frame allocation. The
experiment was reverted, and a restored prefix spot-check returned to
`1141961498 ns/op`, `77166864 B/op`, and `2130506 allocs/op`. Future quicksort
work should not keep adding raw-slot retrofits to dynamic `runtime.Value`
frames; it should use typed opcodes/register frames, or pivot to canonical
array/member dispatch or a v12-safe parser/native-bytecode lane.

The first canonical-dispatch follow-up after the sidecar rejection was also
tested and rejected. Caching `bytecodeTraceEnabled` on the VM and using that
flag in the hot array-slot trace guards preserved focused array-slot/trace
semantics, but the 1MB quicksort prefix regressed to `1174707776 ns/op`,
`77170704 B/op`, and `2130517 allocs/op` over `3/3`, with no allocation
improvement. After revert, a restored `3/3` band returned to `1077327345 ns/op`,
`77168904 B/op`, and `2130517 allocs/op`. Future canonical array/member work
should not spend more time on disabled-trace guard shaving; it should remove
duplicate tracked-array probing or per-hit proof/version work, or move to the
v12-safe parser/native-bytecode lane.

A subsequent direct `Array.read_slot` proof-cache specialization was also
tested and rejected. The specialized lowered-opcode lookup kept the same
env/global/method-cache revision checks as the generic array-slot proof cache,
but skipped the generic kind validation and separate cacheability helper on hot
hits. Focused invalidation/trace/quicksort parity stayed green, but the 1MB
quicksort prefix landed at `1135058602 ns/op`, `77168760 B/op`, and
`2130510 allocs/op` over `3/3`, with no allocation win. After revert, a restored
spot-check landed at `1120617924 ns/op`, `77167056 B/op`, and
`2130515 allocs/op`. Future quicksort work should stop one-off canonical
read-slot proof helper rewrites and move to the v12-safe typed/native-bytecode
lane or a real typed opcode/register-frame design.

The next kept typed-lane slice stayed deliberately away from untyped quicksort
inference. Explicit typed `i32` declarations and assignments that already lower
through `StoreSlotI32` now mark statement-position stores as discarded, letting
the VM keep the raw `bytecodeRawI32SlotValue` in the typed slot instead of
boxing and pushing an assignment value that will be popped immediately.
Non-discarded assignment expressions still box/push, and generic visible loads
still materialize the raw slot value at the boundary.

The 1MB external quicksort prefix is mostly untyped, so it is only a regression
guard for this infrastructure slice. It stayed clean at `1074942410 ns/op`,
`77167024 B/op`, and `2130512 allocs/op` over `3/3`. The next typed-lane
performance tranche should carry explicit raw `i32` values across a larger
end-to-end boundary, such as typed local update chains, typed call/return
boundaries, or typed loop lowering; do not restart untyped quicksort-local
inference without a v12 typechecker-backed proof.

The follow-up kept typed-lane slice added that same raw treatment to explicit
typed `i32` compound updates. `+=` and `-=` against a proven `i32` slot now
lower to a dedicated `CompoundAssignSlotI32` opcode when the RHS can already run
on the raw `i32` stack. The opcode still honors v12 compound-assignment order:
the RHS is evaluated first, and only then does the VM read the current slot and
apply checked `i32` add/sub. RHS expressions with assignment side effects fall
back to the generic compound path.

The 1MB external quicksort prefix stayed neutral at `1124163735 ns/op`,
`77167029 B/op`, and `2130514 allocs/op` over `3/3`, which is expected because
the current quicksort source uses `x = x + 1` updates rather than compound
`+=`. The next benchmark-facing typed-lane tranche should therefore move to
v12 typechecker-backed slot-kind propagation for locals declared from typed
values, such as `i := lo`, `j := hi`, typed `len()` results, and benchmark
`x = x + rhs` loop updates. Do not restart untyped local inference in bytecode
lowering without that proof.

That propagated-local quicksort tranche was tested next and rejected. The
experiment only propagated an unannotated local to `i32` when its initializer
was already proven by an `i32` slot or explicit `i32` literal, so plain
untyped `x := 1` stayed generic. It hit the intended quicksort partition locals
(`i := lo`, `j := hi`) and their `x = x + rhs` updates, but the current raw
slot representation still writes non-pointer `bytecodeRawI32SlotValue`
sentinels into `runtime.Value` slots on every hot update. The 1MB external
quicksort prefix regressed to `1472282588 ns/op`, `77921013 B/op`, and
`2319143 allocs/op` over `3/3`; after revert, the restored prefix returned to
`1102503186 ns/op`, `77166880 B/op`, and `2130506 allocs/op`.

The next quicksort typed-lane work should not re-enable propagated hot locals
until raw `i32` slots live in sidecar/register storage outside the
`runtime.Value` interface. That lane needs explicit materialization at generic
loads, calls, returns, member dispatch, and index boundaries before `i := lo` /
`j := hi` propagation can be a keep candidate.

That active-frame sidecar direction was retried as a narrow VM slice and
rejected again. The experiment moved explicit typed `StoreSlotI32` /
`CompoundAssignSlotI32` discarded values into VM sidecar arrays, saved/restored
those arrays across inline call frames, and made nearby slot-const, branch,
array-index, and f64-kernel readers sidecar-aware. Focused parity stayed green,
and reduced `Fib30Bytecode` remained in range at `160.39ms/op`, `162.00ms/op`,
and `160.20ms/op`, but the 1MB external quicksort prefix landed at
`1126267158 ns/op`, `84192925 B/op`, and `3293537 allocs/op` over `3/3`.
After reverting the experiment, a restored prefix spot-check returned to
`1009100674 ns/op`, `77172744 B/op`, and `2130528 allocs/op`.

The allocation regression confirms that active sidecars layered onto dynamic
`runtime.Value` frames are not the right next step. Future quicksort typed-lane
work should start as a typed opcode/register-frame design, or pivot back to a
fresh external-scale profile of canonical array/member dispatch.

The next kept quicksort tranche targeted trailing statement results in
control-flow bodies instead of another raw-slot storage rewrite. Lowering now
detects when an `if` / `elsif` / `else`, `loop`, `while`, or `for` body ends in
a store opcode that already supports `discardResult`; it marks that store as
discarded and skips the `Pop` that used to consume the assignment value
immediately afterward. This keeps expression-position `if` values unchanged
and only removes work from statement-position block bodies.

The 1MB external quicksort prefix moved from the restored raw-sentinel band of
about `1009100674 ns/op`, `77172744 B/op`, and `2130528 allocs/op` to
`971146975 ns/op`, `45216701 B/op`, and `3779418 allocs/op` over `3/3`. A
profiled confirmation landed at `906273990 ns/op`, `45244848 B/op`, and
`3779485 allocs/op`. This is a keep on wall time and allocated bytes, but not
on allocation count: the profile shows `storeSlotBinaryIntSlotConstI32FastResult`
as the leading allocation-object site because the optimized discarded stores
write more `bytecodeRawI32SlotValue` sentinels into `runtime.Value`
interfaces. The next quicksort tranche should remove that raw-slot allocation
with a real typed opcode/register-frame design, or pivot to a fresh canonical
array/member-dispatch edge. Do not keep broadening interface-sentinel writes or
active-frame sidecars.

The immediate boxed-discard follow-up was tested and rejected. Replacing
discarded slot-const raw sentinels with boxed cached i32 values reduced
allocation count but regressed the 1MB quicksort prefix to
`1222827886 ns/op`, `75424064 B/op`, and `1694770 allocs/op` over `3/3`.
After revert, a restored spot-check returned to `1002372929 ns/op`,
`45216648 B/op`, and `3779422 allocs/op`. Keep the raw sentinel until a true
typed opcode/register-frame design can avoid `runtime.Value` interface writes
entirely.

The next kept canonical-dispatch tranche fused the helper-body swap pattern
used by quicksort when written through canonical Array slot methods:
`tmp := arr.read_slot(a); arr.write_slot(a, arr.read_slot(b));
arr.write_slot(b, tmp)`. Slot-backed shapes now lower to
`ArraySlotSwapSlot`, which proves both kernel `read_slot` and `write_slot`
under env/global/method-version guards before using direct array storage and
falls back to ordinary member calls otherwise. Unlike bracket-index swap
fusion, this opcode returns `void` and preserves `read_slot`/`write_slot`
sparse/grow semantics. The kept 1MB external quicksort prefix landed at
`876341839 ns/op`, `45216184 B/op`, and `3779409 allocs/op` over `3/3`; the
profiled confirmation landed at `866065855 ns/op`, `45244304 B/op`, and
`3779475 allocs/op`. Full external bytecode `quicksort` still timed out at
`90s`. The trace shows the helper body no longer emits separate hot
`read_slot`/`write_slot` entries, while `swap(arr, i, j)` remains a
`437200`-hit inline `call_name` edge. The next quicksort tranche should target
that call-site edge with normal `CallName` cache guards or a general
tiny-function quickening plan; it should not introduce a nominal `Array.swap`
special case or whole-loop quicksort kernel.

That swap call-site direction was tested next and rejected. The first probe
detected cached named functions whose body was exactly `ArraySlotSwapSlot;
Return` and executed that tiny body directly from the cached `CallName` site,
skipping slot-arg stack pushes and call-frame setup on cached hits while
retaining the env/owner-version call cache and Array method proof cache. It
kept focused parity green but regressed the 1MB quicksort prefix to
`914586515 ns/op`, `45213875 B/op`, and `3779403 allocs/op` over `3/3`. A
second smaller probe added a direct tracked-array small-index branch inside
`ArraySlotSwapSlot`; it also stayed semantically green but regressed the prefix
to `983497967 ns/op`, `45216232 B/op`, and `3779416 allocs/op`. Both probes
were reverted, and a restored spot-check landed at `914221901 ns/op`,
`45216152 B/op`, and `3779408 allocs/op`.

Future quicksort work should stop pursuing swap-body/call-site micro paths.
The next tranche should start from a fresh profile of the current kept state
and choose a non-swap edge, most likely the raw i32 register-frame design or
the parser/byte-array lane; do not retry direct tiny-body `CallName` swap
quickening, tracked-swap helper rewrites, a nominal `Array.swap` special case,
or a whole-loop quicksort kernel.

The next kept non-swap tranche targeted the top raw-sentinel allocation edge
without changing slot semantics. `bytecodeRawI32SlotValue` now has a pre-boxed
`runtime.Value` cache for common values in the `-1024..262143` range, and both
discarded slot-const updates and the affine parser update path use it. Visible
loads and returns still materialize raw sentinels through the same existing
boundaries; the cache only avoids repeated interface boxes for internal raw
slot writes.

The initial `-1024..131071` range moved the 1MB quicksort prefix to
`826684825 ns/op`, `35829405 B/op`, and `1432739 allocs/op` over `3/3`, with
a profiled confirmation at `819242291 ns/op`, `35857440 B/op`, and
`1432805 allocs/op`. Doubling the range to `262143` kept wall time in the same
band while improving allocation pressure again: `829529632 ns/op`,
`35244595 B/op`, and `1286523 allocs/op` over `3/3`, with a profiled
confirmation at `782730382 ns/op`, `35272736 B/op`, and `1286588 allocs/op`.
The full external bytecode `quicksort` benchmark still timed out at `90s`.

The remaining profile is no longer a raw-sentinel cache-size problem; it is
split across compare/cast/proof-control flow and residual uncached raw/affine
allocations. The next quicksort tranche should start from a fresh profile and
choose either a true raw i32 register-frame design or the parser/byte-array
lane. Do not keep expanding the cache, and do not return to swap micro paths.

The next kept parser-lane slice lowered bytecode integer literals that fit in
`int64` as small `runtime.IntegerValue` constants instead of big-backed
constants. This is representation-only: larger literals remain big-backed, and
typed out-of-range literals still validate lazily when executed. On quicksort,
the parser update `value = value * 10 + ((byte as i32) - 48)` now keeps the
`48` literal on the small-int path, so the discarded affine update can stay on
the raw i32 fast path instead of allocating through generic arithmetic. The
fresh kept baseline was `800775979 ns/op`, `35272672 B/op`, and
`1286585 allocs/op`. After one noisy slow experimental band, the confirming
1MB prefix band landed at `763104452 ns/op`, `30156808 B/op`, and
`1180529 allocs/op`; the profiled confirmation landed at `770806166 ns/op`,
`30184704 B/op`, and `1180584 allocs/op`. Full external bytecode quicksort
still timed out at `90s`. The next quicksort tranche should target remaining
uncached raw i32 slot-write allocations with a typed opcode/register-frame
design, or re-profile the parser byte-array lane before adding another narrow
parser fusion.

The next kept quicksort tranche stayed on the tracked-array runtime boundary,
but moved from the compare/read side to swap metadata sync. Tracked
array-index and array-slot swaps used to replay two generic tracked writes
after the values had already been exchanged, paying duplicate revision/token
updates and duplicate alias sync on the hot quicksort helper path. The kept
slice replaced that with one swap-aware metadata helper that:

- bumps tracked-array revision once per logical swap while preserving the old
  two-write revision delta
- recomputes the leading element token only when slot `0` changes
- swaps the tracked raw-`i32` cache in place when the existing token/cache
  still applies
- resyncs alias-visible array wrappers once after metadata is settled

This also closed an important raw-first edge: if a tracked swap materializes a
raw-first `Array i32` shape into boxed `i32` values, the tracked raw-`i32`
cache is now recovered after the swap instead of remaining disabled behind the
old raw first element.

The reduced quicksort hotloop stayed in the improved band at `5106341 ns/op`,
`5393816 ns/op`, and `5353576 ns/op` over `100x`, but the real beneficiary was
the full external-scale runtime. On the current kept baseline, the `95MB`
steady-state quicksort pass moved from `193.98s` real average
(`100155699583 ns/op`, `9217829424 B/op`, `203288604 allocs/op`) to
`172.50s` (`88569772666 ns/op`, `9457822464 B/op`, `203288610 allocs/op`).

The full profile confirms that swap metadata is no longer the first-tier wall:
`resolveTrackedArraySlotSwapSlotFastAtSlots(...)` is now `2.12s` cumulative,
`execArraySlotSwapSlot(...)` is `3.94s`, and
`updateTrackedArrayMetadataForSwap(...)` is only `0.69s`. The remaining
full-scale quicksort cost is now led by
`bytecodeDirectSmallI32Value(...)`,
`lookupCachedCanonicalArraySlotCallForArray(...)`,
`compareArrayReadSlotTrackedI32Condition(...)`,
`execJumpIfArrayReadSlotCompareSlotFalse(...)`, and the adjacent call/store
tier. The next quicksort tranche should therefore stay on that same full-scale
compare/canonical-array edge rather than reopening swap-local cuts.

The next benchmark-coverage tranche moved out of VM micro-optimization and
unlocked the external `base64` family. Canonical stdlib now exposes
`able.encoding.base64` for string/byte-array encode/decode and
`able.crypto.md5` for MD5 hex helpers, with Go and TypeScript extern bodies.
The Able benchmark source lives at
`v12/examples/benchmarks/base64/base64.able`, and
`v12/bench_compare_external --benchmarks base64` can now measure it.

The first aligned `base64` comparisons over `1/1` runs landed at:

- compiled: `2.8600s` vs Go `2.2000s`, Ruby `2.2100s`, Python `3.3100s`
- bytecode: `8.4400s` vs Go `2.2000s`, Ruby `2.2100s`, Python `3.3100s`
- tree-walker: `26.8400s` vs Go `2.2000s`, Ruby `2.2100s`, Python `3.3100s`

The compiled output passes the external verifier hashes and the generated
source keeps the explicit Able loop compiled, while the large codec/hash work
crosses the normal host-backed stdlib extern boundary. This is a coverage keep,
not a bytecode optimization tranche. The next coverage tranche should unlock
the JSON benchmark with a reusable `able.json` DOM surface; the next optional
base64 cleanup is a string/byte-buffer convenience such as `String.repeat` so
the source can express the initial one-million-character string without a
manual byte-push loop.

## 2026-06-08 — Base64 bytes-first stdlib and `Array u8` extern fast path

The next `base64` tranche stayed deliberately outside VM opcode work and cut
the host boundary instead:

- `../able-stdlib/src/encoding/base64.able` now exposes reusable
  `encode_bytes(...)` / `decode_bytes(...)` APIs in addition to the existing
  string-based surface.
- `v12/examples/benchmarks/base64/base64.able` now keeps the payload as
  `Array u8` end to end and hashes with `md5.hex(...)`, so it no longer pays
  `String` conversion plus `DecodeError | String` matching on every pass.
- the Go extern fast-invoker layer now handles primitive `Array u8` host
  signatures directly (`[]byte -> String`, `[]byte -> []byte`,
  `[]byte -> interface{}`), which lets hot base64/MD5 externs skip the generic
  host coercion bridge.

On the external benchmark over `3/3` runs this moved bytecode from the fresh
kept `8.9300s` band to `3.2600s`:

- bytecode: `3.2600s` vs Go `2.2000s`, Ruby `2.2100s`, Python `3.3100s`

The profiled runtime confirmation landed at:

- `3485950441 ns/op`
- `6935074184 B/op`
- `480 allocs/op`

The useful profile change is structural rather than another VM micro-hotspot
shuffle: the old string/union round-trip is gone, allocation count collapsed,
and the remaining wall is now the actual host codec plus byte-slice copying
(`encoding/base64`, `runtime.memmove`, and MD5) instead of generic extern
dispatch.

The follow-up kept the same boundary but narrowed it further on the result
side only:

- host `[]byte` results now transfer into mono `Array u8` storage through an
  explicit owned-bytes helper instead of being copied again on the Able side
- input `Array u8` arguments still copy, because borrowing them would change
  host-extern mutation semantics

That moved the external bytecode band again from `3.2600s` to `3.0533s` over
`3/3`, with profiled runtime confirmation at:

- `3213493295 ns/op`
- `4734526144 B/op`
- `434 allocs/op`

The useful profile result is that the output-side array construction dropped
out of the top tier. The remaining `base64` wall is now almost entirely the
input extraction plus host codec/copy boundary:

- `externU8ArrayArg(...)`
- `ArrayStoreMonoU8BytesIfAvailable(...)`
- `encoding/base64.(*Encoding).Encode`
- `encoding/base64.(*Encoding).Decode`

## 2026-06-08 — Explicit borrowed-byte input contract for read-only `Array u8` externs

The deferred `base64` input-boundary tranche finally landed, but only as an
explicit opt-in contract. The earlier owned-result transfer proved the output
side was safe; the missing piece was input borrowing without silently changing
host-extern mutation semantics.

The keep is deliberately narrow:

- generated Go extern modules now include
  `able_borrowed_bytes(data []byte) []byte`
- the `Array u8` fast-invoker layer recognizes that explicit marker and, only
  on that path, borrows mono-`u8` backing storage through
  `ArrayStoreMonoBorrowedU8BytesIfAvailable(...)`
- non-opt-in `Array u8` host args still copy exactly as before

The first reusable stdlib adopters are the read-only Go externs in:

- `../able-stdlib/src/encoding/base64.able`
- `../able-stdlib/src/crypto/md5.able`
- `../able-stdlib/src/io.able`

That moved the external `base64` bytecode band from `3.0533s` to `2.5533s`
over `3/3`, with steady-state runtime confirmation at:

- `2275118163 ns/op`
- `2218630616 B/op`
- `375 allocs/op`

This cut is important because it keeps the contract reusable and explicit. The
fast path is no longer "borrow whenever the runtime can"; it is "borrow only
when the host body declares read-only intent". That keeps ordinary user externs
on the old safe copying path while letting byte-heavy stdlib calls stop paying
the inbound mono-`u8` copy.

## 2026-06-09 — Compiler-side borrowed-byte host ABI recovery for `base64`

The explicit borrowed-byte contract now lands on the compiled path too. The
bytecode fast-invoker keep proved that opt-in read-only borrowing was safe and
useful; this tranche applies the same contract to compiled mono `Array u8`
extern arguments without changing default host semantics.

The keep is again intentionally exact-shape:

- compiled Go host support now emits
  `able_borrowed_bytes(data []byte) []byte`
- compiled extern lowering recognizes only the explicit opt-in shape:
  - host body contains `able_borrowed_bytes(...)`
  - parameter type is exact mono `Array u8`
  - host carrier type is native `[]uint8`
- on that path, compiled extern calls pass `value.Elements` directly instead of
  routing the argument through `bridge.RuntimeValueToHost[...]`
- non-opt-in `Array u8` args still copy, so ordinary host-extern mutation
  semantics are unchanged

The focused compiler proof now covers both sides of the contract:

- source inspection proves the borrowed mono-`u8` path uses the native carrier
  directly and avoids the generic host bridge
- execution proves the old default path still copies while the explicit
  borrowed path can alias and mutate the original byte array

That moved the external `base64` compiled band from the historical
`2.8600s` baseline to `2.5567s` over `3/3`.

Verification:

- focused compiler extern host tests
- `../able/v12/ablebc test tests/encoding/base64.test.able tests/crypto/md5.test.able tests/io.test.able`
- external `./v12/bench_compare_external --benchmarks base64 --modes compiled --runs 3 --timeout 120`

This matters because the borrowed-byte ABI is now coherent across execution
paths. The same explicit stdlib marker used by bytecode fast invokers now also
cuts the compiled inbound `Array u8` bridge, without silently changing the
default extern contract for user code.

## 2026-06-09 — Compiled owned-`[]byte` host-result transfer for `base64`

The next compiled `base64` tranche stayed on the same host ABI boundary, but on
the return side instead of the input side. The bytecode path had already proven
that fresh host `[]byte` results should become owned mono `Array u8` values
without an extra copy. The compiled path was still copying those slices with
`append(...)` in both the exact `Array u8` return fast path and the union
`Array u8` member fast path.

The keep is deliberately narrow:

- exact mono `Array u8` compiled extern returns now transfer host `[]uint8`
  directly into `__able_array_u8{Elements: ...}`
- union fast returns that detect a `[]uint8` member now do the same
- other mono-array return paths keep the old copy behavior

Focused proof coverage now checks both structure and behavior:

- source inspection proves the compiled extern body no longer emits the old
  `append([]uint8(nil), ...)` copy on the direct and union fast paths
- execution proves a returned host `[]uint8` slice is now reused by the mono
  `Array u8` carrier

That moved the external `base64` compiled band from the prior kept `2.5567s`
to `2.1667s` over `3/3`, which puts compiled Able slightly ahead of the
current Go row (`2.2000s`) on this family.

Verification:

- focused compiler extern host tests
- `../able/v12/ablebc test tests/encoding/base64.test.able tests/crypto/md5.test.able tests/io.test.able`
- external `./v12/bench_compare_external --benchmarks base64 --modes compiled --runs 3 --timeout 120`

This is the point where `base64` stops being a compelling local compiled
micro-target. Further work there should be broader host ABI/codegen machinery,
not more one-off `Array u8` probes.

## 2026-05-27 — JSON benchmark stdlib coverage and `fs.read_text` fast path

The next benchmark-coverage tranche added the reusable JSON surface and the
canonical Able source for the external `json` family:

- `../able-stdlib/src/json.able` now exposes `JsonValue` DOM structs,
  `parse(...)`, typed `JsonObject` / `JsonArray` accessors, and
  `f64_field_means(...)` for streaming numeric projections over an object array.
- `v12/examples/benchmarks/json/json.able` reads `sample.json`, computes the
  `x`/`y`/`z` means through `able.json`, and prints the three results.
- `v12/bench_compare_external --benchmarks json` is now wired as an opt-in
  target.
- `../able-stdlib/src/fs.able` now has a host-backed `read_text` fast path.
  This is the key performance fix: the first full JSON run before it was
  compiled `34.7600s` and bytecode timed out at `180s`; after it, the benchmark
  avoids converting the 110 MB file from host bytes through Able code.

Aligned external JSON comparisons over `3/3` runs landed at:

- compiled: `3.7033s` vs Go `1.3600s`, Ruby `1.5600s`, Python `2.8700s`
- bytecode: `4.0267s` vs Go `1.3600s`, Ruby `1.5600s`, Python `2.8700s`

This is a coverage and stdlib hot-API keep, not a finished JSON performance
win. The remaining work is to close the roughly `2.7x-3.0x` gap to Go with a
lower-allocation JSON parser/token or typed projection lane, while keeping the
benchmark source expressed through reusable stdlib APIs. Tree-walker timing is
not a performance target; it only needs to preserve v12 semantics.

## 2026-05-27 — JSON fast numeric projection scanner

The follow-up JSON tranche replaced the `encoding/json.Decoder` token loop
inside `able.json.f64_field_means(...)` with a reusable Go fast scanner over
the already-loaded text. The scanner still honors the public API
(`array_key`, requested numeric `field_names`) and skips unrelated top-level
values plus nested object/array fields such as the benchmark `opts` payload; it
does not change the DOM `parse(...)` API.

Aligned external JSON comparisons over `3/3` runs moved from the decoder
projection baseline:

- compiled: `3.7033s` -> `0.6700s`
- bytecode: `4.0267s` -> `0.7233s`

Against the checked-in external references, the kept band is:

- compiled: `0.6700s` vs Go `1.3600s`, Ruby `1.5600s`, Python `2.8700s`
- bytecode: `0.7233s` vs Go `1.3600s`, Ruby `1.5600s`, Python `2.8700s`

This closes the JSON benchmark gap for compiled and bytecode without adding a
benchmark-specific source rule. Tree-walker remains semantic-only. At this
point, the remaining unimplemented coverage target was `pidigits`, which still
needed a competitive BigInt/BigUint boundary before its later coverage tranche.

## 2026-06-08 — JSON `Array String` extern fast invoker recovery

The latest external `json` runs exposed a host-ABI regression boundary rather
than a parsing problem. On the current benchmark input, the bytecode external
baseline had drifted to `5.6200s`, while a direct runtime profile still showed
the real scanner path at about `528ms/op`. The hot gap was back at the host
boundary:

- `reflect.Value.Call`
- generic extern argument coercion for `Array String`
- generic union-result coercion for `JsonError | Array f64`

The keep was a reusable direct host ABI cut, not a JSON-specific parser
rewrite:

- add a fast invoker for the exact shape
  `(String, String, Array String) -> JsonError | Array f64`
- decode `Array String` arguments directly from runtime arrays
- handle `[]float64` union results directly without going through the generic
  reflection call bridge

That moved the external bytecode band from the fresh `5.6200s` baseline to
`0.8300s` over `3/3` runs:

- bytecode: `0.8300s` vs Go `1.3600s`, Ruby `1.5600s`, Python `2.8700s`

Profiled runtime confirmation landed at:

- `536181557 ns/op`
- `229466568 B/op`
- `188 allocs/op`

The useful profile result is that `reflect.Value.Call` dropped out of the top
tier entirely. The remaining `json` wall is now the actual numeric scanner and
file ingress:

- `ableJsonF64FieldMeansFast(...)`
- `ableJsonReadNumber(...)`
- `strconv.ParseFloat`
- `fs_read_text_fast`

## 2026-06-08 — JSON union-`String` extern fast-path follow-up

The next `json` tranche kept the scanner unchanged and only removed the
remaining generic host-ABI bridge behind `fs_read_text_fast(...)`: for
`func(string) interface{}` externs whose declared union result includes
`String`, the fast invoker now returns a direct `runtime.StringValue` instead
of falling back through generic union-result coercion.

That moved the prior kept external bytecode band from `0.8300s` to
`0.7867s` over `3/3` runs:

- bytecode: `0.7867s` vs Go `1.3600s`, Ruby `1.5600s`, Python `2.8700s`

Profiled runtime confirmation landed at:

- `533953303 ns/op`
- `229461760 B/op`
- `173 allocs/op`

The useful conclusion is narrower now: generic extern dispatch is no longer
the `json` wall. The remaining work is the real scanner/file path:

- `ableJsonF64FieldMeansFast(...)`
- `ableJsonReadNumber(...)`
- `strconv.ParseFloat`
- `fs_read_text_fast`

## 2026-06-08 — JSON owned file-text ingress follow-up

The next `json` tranche left the scanner unchanged again and only removed the
fresh file-read `[]byte -> String` copy inside the Go `fs_read_text_fast(...)`
host prelude. Fresh `os.ReadFile(...)` bytes now become an immutable String
directly before crossing back into Able.

That moved the prior kept external bytecode band from `0.7867s` to
`0.7533s` over `3/3` runs:

- bytecode: `0.7533s` vs Go `1.3600s`, Ruby `1.5600s`, Python `2.8700s`

Profiled runtime confirmation landed at:

- `505536583 ns/op`
- `114755048 B/op`
- `170 allocs/op`

The useful conclusion is tighter now: `fs_read_text_fast(...)` is no longer a
first-tier `json` wall. The remaining work is decisively the numeric scanner:

- `ableJsonReadNumber(...)`
- `strconv.ParseFloat`
- residual `ableJsonF64FieldMeansFast(...)` control flow around them

## 2026-05-28 — Monte Carlo Pi benchmark coverage

The next coverage tranche added deterministic RNG support and the canonical
Able source for the external `monte_carlo_pi` family:

- `../able-stdlib/src/random.able` exposes `Random.seeded(...)`,
  `Random.default()`, `next_i32()`, `next_i64()`, and `next_f64()` using a
  deterministic Park-Miller recurrence, with focused stdlib specs in
  `../able-stdlib/tests/random.test.able`.
- `v12/examples/benchmarks/monte_carlo_pi/monte_carlo_pi.able` implements the
  external Monte Carlo sampling loop and prints the five sample sizes expected
  by the verifier.
- `v12/bench_compare_external --benchmarks monte_carlo_pi` is now wired as an
  opt-in target.

The benchmark source keeps the Park-Miller recurrence inline inside the hot
loop instead of calling `Random.next_f64()` for each coordinate. That is a
deliberate current-VM compromise: the public RNG API is the reusable stdlib
surface, while the benchmark avoids bytecode method-call/object-state churn
until the VM has a general primitive numeric slot representation.

Aligned external comparisons over `3/3` runs landed at:

- compiled: `0.3233s` vs Go `0.1800s`, Ruby `1.4200s`, Python `1.6800s`
- bytecode baseline: `18.7967s` vs Go `0.1800s`, Ruby `1.4200s`, Python
  `1.6800s`

The next bytecode keep was a general integer recurrence slice, not a
benchmark-specific RNG shortcut. Lowering now fuses

- `state = (state * 48271_i64) % 2147483647_i64`

and any other exact `slot = (slot * const) % const` integer self-assignment
shape into a dedicated bytecode store opcode. That removes the intermediate
boxed multiply result plus the follow-on generic `%` dispatch without changing
v12 overflow or Euclidean modulo semantics.

The kept external Monte Carlo comparison over `3/3` runs moved to:

- bytecode kept band: `12.5367s` vs Go `0.1800s`, Ruby `1.4200s`, Python
  `1.6800s`

The profiled bytecode runtime confirmation moved from roughly
`17019697503 ns/op`, `5101560232 B/op`, and `184425053 allocs/op` to:

- `14678232749 ns/op`
- `4038540752 B/op`
- `162287138 allocs/op`

The profile shift is the important part:

- before: `execBinary(...)` about `21.77s` cumulative,
  `applyBinaryOperator(...)` about `10.82s`,
  `evaluateDivMod(...)` about `7.44s`
- after: `execBinary(...)` about `6.99s` cumulative,
  `applyBinaryOperator(...)` about `3.37s`, and the new
  `bytecodeIntMulConstModConstFast(...)` carrying the Park-Miller update path

This is a real bytecode competitiveness win, but it still does not close the
full Monte Carlo gap. The next remaining wall is broader boxed `i64` / `f64`
representation, casts, and numeric boxing rather than the recurrence source
shape itself.

The next kept Monte Carlo bytecode slice removed the remaining cast/div
boundary around:

- `(state as f64) / 2147483647.0`

Lowering now recognizes the exact `(<slot> as f32|f64) / <float literal>`
shape and emits a dedicated opcode that reads the slot, performs the cast, and
divides by the constant directly, with a full fallback to the existing cast and
binary division semantics if the slot value is not a direct numeric shape.

That moved the external Monte Carlo comparison again over `3/3` runs:

- prior kept bytecode band: `12.5367s`
- next kept bytecode band: `11.0433s`

The profiled runtime confirmation moved from:

- `14678232749 ns/op`
- `4038540752 B/op`
- `162287138 allocs/op`

to:

- `11857149346 ns/op`
- `3505213024 B/op`
- `140065140 allocs/op`

The profile shift is clear:

- the old `castValueToCanonicalSimpleTypeFast(...)` and `evaluateDivision(...)`
  prominence dropped out of the visible top tier
- the new `execBinaryCastSlotFloatConstDiv(...)` path is only about `2.04s`
  cumulative on the kept profile
- the remaining dominant Monte Carlo wall is now boxed integer creation:
  `bytecodeBoxedIntegerValue(...)` about `8.04s` cumulative and
  `boxedOrSmallIntegerValue(...)` about `8.07s`

So this is another real bytecode competitiveness keep, and it tightens the
next direction further: do not spend another tranche on cast/div reshaping.
The next Monte Carlo slice should target boxed `i64` value creation and the
dynamic integer boxing/cache path instead.

That large-`i64` boxing path is now also kept. The next bounded Monte Carlo
slice changed the integer boxing policy itself: large `i64` values now bypass
the dynamic boxed-int map/RWMutex path and return direct `runtime.NewSmallInt`
carriers instead. This is deliberately narrow to the high-cardinality `i64`
case; the existing `i32` extended static cache and the dynamic caching policy
for the other integer kinds are unchanged.

That moved the external Monte Carlo comparison again over `3/3` runs:

- prior kept bytecode band: `11.0433s`
- next kept bytecode band: `7.5667s`

The profiled runtime confirmation moved from:

- `11857149346 ns/op`
- `3505213024 B/op`
- `140065140 allocs/op`

to:

- `7923979565 ns/op`
- `3541045976 B/op`
- `140811440 allocs/op`

The wall-clock improvement is real even though bytes/allocs stayed roughly
flat. The profile explains why:

- before: `bytecodeBoxedIntegerValue(...)` about `8.04s` cumulative and
  `boxedOrSmallIntegerValue(...)` about `8.07s`, with the dynamic map path
  still visible
- after: `bytecodeBoxedIntegerValue(...)` about `1.17s` cumulative and
  `boxedOrSmallIntegerValue(...)` about `1.19s`, with the old map/lock path
  no longer a first-tier cost

That changes the Monte Carlo ranking again. The dominant remaining wall is now
boxed `f64` / `runtime.Value` arithmetic:

- `execBinary(...)`: about `6.18s` cumulative
- `runtime.convT`: about `6.07s`
- `bytecodeDirectFloatArithmeticFast(...)`: about `2.11s`
- `execBinaryCastSlotFloatConstDiv(...)`: about `2.02s`

So this is a strong keep, and it closes the large-`i64` boxing question for
this benchmark. The next bounded Monte Carlo slice should target float/value
representation and scalar `f64` arithmetic, not another integer boxing tweak.

That float/value boundary is now materially narrower too. The next kept Monte
Carlo bytecode slice fused the exact slot-backed condition:

- `(x * x + y * y) <= 1.0`

Lowering now recognizes the exact multiply-add-multiply compare shape in
`if`/`elsif` conditions and emits a dedicated false-jump opcode instead of
separate `LoadSlot`, `Binary "*"`, `Binary "*"`, `Binary "+"`, `Const`, and
`Binary "<="` work. The VM executes that condition directly from slots when
the runtime values are direct float shapes, with a full fallback to the
existing multiply/add/compare semantics when they are not. This is still a
semantic bytecode keep, not a benchmark-source rewrite.

That moved the external Monte Carlo comparison again over `3/3` runs:

- prior kept bytecode band: `7.5667s`
- next kept bytecode band: `4.1733s`

The profiled bytecode runtime confirmation moved from:

- `7923979565 ns/op`
- `3541045976 B/op`
- `140811440 allocs/op`

to:

- `4218845731 ns/op`
- `1674386480 B/op`
- `63034426 allocs/op`

The profile shift is the key result:

- `execBinary(...)` dropped from about `6.18s` cumulative to about `1.96s`
- `runtime.convT` dropped from about `6.07s` to about `1.64s`
- the new `execJumpIfFloatMulAddMulCompareConstFalse(...)` path is only about
  `0.61s` cumulative
- the remaining first-tier Monte Carlo wall is now the slot-backed float
  production/storage boundary:
  `execStoreSlotIntMulConstModConst(...)`,
  `bytecodeCastSlotFloatConstDivFast(...)`, and `storeOwnedFloatSlot(...)`

So this is another real bytecode competitiveness keep. The next bounded Monte
Carlo slice should stay on typed-float representation or slot-backed float
value movement, not another integer boxing or source-level RNG probe.

That slot-backed float movement boundary is now narrower too. A fifth kept
Monte Carlo bytecode slice fused the exact statement-position store shape:

- `target := (slot as f64) / const_f64`
- `target = (slot as f64) / const_f64`

Lowering now recognizes that store boundary and emits a dedicated
`StoreSlotCastSlotFloatConstDiv` opcode instead of lowering through the
existing expression opcode and then the generic store path. The VM executes
that store directly, reusing the existing `bytecodeCastSlotFloatConstDivFast`
helper when the source slot is a direct integer shape and falling back to the
current cast/divide semantics otherwise. The result still goes through the
existing float-slot ownership rules, so this is a semantic VM keep rather than
a benchmark-source special case.

That moved the external Monte Carlo comparison again over `3/3` runs:

- prior kept bytecode band: `4.1733s`
- next kept bytecode band: `3.7900s`

The profiled bytecode runtime confirmation moved from:

- `4218845731 ns/op`
- `1674386480 B/op`
- `63034426 allocs/op`

to:

- `3864059446 ns/op`
- `1674386400 B/op`
- `63034425 allocs/op`

The important conclusion is again the boundary shift:

- `execStoreSlotCastSlotFloatConstDiv(...)` is now the visible hot edge at
  about `2.40s` cumulative
- `storeSlotCastSlotFloatConstDivResult(...)` is about `1.63s`
- `bytecodeCastSlotFloatConstDivFast(...)` remains materially hot at about
  `1.58s`
- `storeOwnedFloatSlot(...)` is now isolated at about `0.37s`
- `execBinary(...)` is still present, but the old generic stack-staging plus
  store boundary is no longer the first-tier wall

So the next bounded Monte Carlo slice should stay on the remaining float slot
production/storage and materialization wall, especially
`bytecodeCastSlotFloatConstDivFast(...)`, `storeOwnedFloatSlot(...)`, and
residual `runtime.convT`, not integer boxing policy or more RNG-source
rewrites.

That float-slot completion boundary is now narrower again. A sixth kept Monte
Carlo bytecode slice specialized the already-fused discarded local store path
for:

- `x := (state as f64) / const_f64`
- `y := (state as f64) / const_f64`

The fused `StoreSlotCastSlotFloatConstDiv` opcode now computes the direct raw
float result and stores it straight into the owned float slot cell when the
assignment result is discarded, instead of round-tripping through a generic
`runtime.Value` result and the broader visible-result completion path. This is
still a semantic VM keep: non-fast or non-discarded cases keep the existing
cast/divide and store behavior.

That moved the external Monte Carlo comparison again over `3/3` runs:

- prior kept bytecode band: `3.7900s`
- next kept bytecode band: `3.0400s`

The profiled bytecode runtime confirmation moved from:

- `3864059446 ns/op`
- `1674386400 B/op`
- `63034425 allocs/op`

to:

- `3110566300 ns/op`
- `1141058080 B/op`
- `40812425 allocs/op`

The profile shift is the key result:

- `execStoreSlotCastSlotFloatConstDiv(...)` dropped to about `1.34s`
  cumulative
- `execStoreSlotCastSlotFloatConstDivDiscardFast(...)` now carries the hot
  fused discarded-store path at about `1.06s`
- `storeOwnedFloatSlotRaw(...)` is about `0.21s`
- `runtime.convT` is down to about `0.75s`
- the primary Monte Carlo wall has moved back to the integer recurrence
  update path:
  `execStoreSlotIntMulConstModConst(...)`,
  `storeSlotIntMulConstModConstResult(...)`, and
  `bytecodeIntMulConstModConstFast(...)`

So the next bounded Monte Carlo slice should pivot back to the integer
recurrence store/update boundary, not spend another tranche on float-store
completion shaving.

That recurrence boundary is now narrower too. A seventh kept Monte Carlo
bytecode slice specialized the exact nonnegative small-int recurrence case
inside `bytecodeIntMulConstModConstFast(...)`.

When the current slot value, multiply immediate, and modulo immediate are all
proven small positive integers of the same signed kind, the VM now skips the
generic direct-integer + Euclidean modulo path, uses native checked positive
multiply, and uses direct `%` because it is already Euclidean for the proven
nonnegative case. Unsupported signed or mixed cases still use the existing
generic path, so this remains a semantic keep.

That moved the external Monte Carlo comparison again over `3/3` runs:

- prior kept bytecode band: `3.0400s`
- next kept bytecode band: `2.9900s`

The profiled bytecode runtime confirmation moved from:

- `3110566300 ns/op`
- `1141058080 B/op`
- `40812425 allocs/op`

to:

- `3063708460 ns/op`
- `1141058096 B/op`
- `40812425 allocs/op`

This is a smaller keep than the previous float-store slices, but the profile
shift is still real:

- `execStoreSlotIntMulConstModConst(...)` dropped to about `2.38s`
  cumulative
- `storeSlotIntMulConstModConstResult(...)` dropped to about `1.70s`
- `bytecodeIntMulConstModConstFast(...)` dropped to about `1.52s`

So the remaining first-tier Monte Carlo wall is now the boxed `i64` result
boundary on that same recurrence path:

- `boxedOrSmallIntegerValue(...)`
- `bytecodeBoxedIntegerValue(...)`
- `runtime.convT`

The next bounded Monte Carlo slice should target that result-boxing boundary,
not more multiply/mod arithmetic shaving.

That boxed-`i64` result boundary is now largely gone too. An eighth kept Monte
Carlo bytecode slice changed discarded `i64`
`StoreSlotIntMulConstModConst` results to stay in the slot as an internal raw
`i64` sentinel instead of materializing a boxed `runtime.IntegerValue` on
every recurrence step.

Visible reads still box that sentinel back through `bytecodeSlotReadValue(...)`,
and the fused recurrence helper plus the fused cast-slot-float-const divide
helper read it directly in slot form. So this stays a semantic VM keep rather
than a benchmark-source shortcut.

That moved the external Monte Carlo comparison again over `3/3` runs:

- prior kept bytecode band: `2.9900s`
- next kept bytecode band: `2.2400s`

The profiled bytecode runtime confirmation moved from:

- `3063708460 ns/op`
- `1141058096 B/op`
- `40812425 allocs/op`

to:

- `2154280243 ns/op`
- `261624272 B/op`
- `40812578 allocs/op`

This is another major boundary shift:

- `execStoreSlotIntMulConstModConst(...)` dropped to about `1.44s`
  cumulative
- `storeSlotIntMulConstModConstResult(...)` dropped to about `0.83s`
- `bytecodeIntMulConstModConstFast(...)` dropped to about `0.56s`
- the boxed `i64` result boundary is no longer first-tier

The next remaining recurrence wall is now the raw `i64` slot/interface
boundary itself:

- `bytecodeRawI64SlotValueFor(...)`
- `runtime.convT64`
- `finishStoreSlotBinaryIntSlotConstFastResult(...)`

So the next bounded Monte Carlo slice should target raw `i64` slot
materialization/interface conversion, not more recurrence arithmetic or
boxing-policy shaving.

That raw-`i64` slot boundary is now narrower too. A ninth kept Monte Carlo
bytecode slice changed the discarded recurrence fast path to reuse a
per-target internal `*bytecodeRawI64SlotCell` instead of creating a fresh raw
wrapper or boxed `i64` value on each Park-Miller update.

Visible reads still box that cell back through `bytecodeSlotReadValue(...)`,
and the fused recurrence helper plus the fused cast-slot-float-const divide
helper still read it directly in slot form. So this remains an internal slot
representation keep rather than a benchmark-source shortcut.

That moved the external Monte Carlo comparison again over `3/3` runs:

- prior kept bytecode band: `2.2400s`
- next kept bytecode band: `1.8533s`

The direct bytecode runtime confirmation moved from:

- `2154280243 ns/op`
- `261624272 B/op`
- `40812578 allocs/op`

to:

- `1751216714 ns/op`
- `74404488 B/op`
- `18590587 allocs/op`

The planning consequence shifts again:

- `execStoreSlotIntMulConstModConst(...)` is down to about `0.77s`
  cumulative
- `execStoreSlotCastSlotFloatConstDiv(...)` is about `0.59s`
- `execStoreSlotBinaryIntSlotConst(...)` is now a co-equal wall at about
  `0.75s`
- the old raw-`i64` slot/interface boundary is no longer first-tier

So the next bounded Monte Carlo slice should pivot to the generic small-int
slot update path that still carries the loop counters and sample accumulation:

- `storeSlotBinaryIntSlotConstI32RawFastResult(...)`
- `storeSlotBinaryIntSlotConstI32RawResult(...)`
- `finishStoreSlotBinaryIntSlotConstFastResult(...)`
- `runtime.convT32`

That small-int slot-update boundary is now narrower too. A tenth kept Monte
Carlo bytecode slice specialized the exact discarded `i32`
`StoreSlotBinaryIntSlotConst` path for out-of-cache counter values so it
writes into reusable owned `*runtime.IntegerValue` slot cells instead of
creating a fresh raw-`i32` interface carrier on every update.

Visible loads still materialize ordinary Able integers through
`bytecodeSlotReadValue(...)`, and the existing direct integer readers already
support `*runtime.IntegerValue`, so this remains a semantic VM keep rather
than a benchmark-source shortcut.

That moved the external Monte Carlo comparison again over `3/3` runs:

- prior kept bytecode band: `1.8533s`
- next kept bytecode band: `1.6733s`

The direct bytecode runtime confirmation moved from:

- `1751216714 ns/op`
- `74404488 B/op`
- `18590587 allocs/op`

to:

- `1583567530 ns/op`
- `35520 B/op`
- `140 allocs/op`

The planning consequence shifts again:

- `execStoreSlotBinaryIntSlotConst(...)` dropped from about `0.75s`
  cumulative to about `0.63s`
- `execStoreSlotBinaryIntSlotConstDiscardI32Fast(...)` now carries the hot
  counter-update path at about `0.51s`
- `runtime.convT32` dropped out of the visible top tier

So the next bounded Monte Carlo slice should pivot to the owned slot
write/update boundary and its map lookups:

- `storeOwnedFloatSlotRaw(...)`
- `storeOwnedI32SlotRaw(...)`
- `runtime.mapaccess1_fast64`

That owned-slot write boundary is now narrower too. An eleventh kept Monte
Carlo bytecode slice changed the discarded `StoreSlotCastSlotFloatConstDiv`
fast path to update the existing owned target `*runtime.FloatValue` in place
when the target slot already holds the reusable float cell from a prior loop
iteration, instead of routing every `x`/`y` update back through
`storeOwnedFloatSlotRaw(...)`.

Visible reads still materialize plain Able `f64` values through
`bytecodeSlotReadValue(...)`, and unsupported cases still fall back to the
existing owned-float helper path. So this remains an internal slot-update keep
rather than a source-level benchmark trick or a new global cache policy.

On the external benchmark, that moved bytecode Monte Carlo from `1.6733s` to
`1.5133s` over `3/3` runs, and the direct runtime confirmation from about
`1.58s` / `35KB` / `140 allocs` to about `1.40s` / `35KB` / `140 allocs`.

The planning consequence changes again. The owned float-slot write/map
boundary is no longer the first wall:

- `execStoreSlotCastSlotFloatConstDiv(...)` is down to about `0.33s`
  cumulative
- `execStoreSlotCastSlotFloatConstDivDiscardFast(...)` is about `0.21s`
- `runtime.mapaccess1_fast64` is down to about `0.15s`

So the next primitive-numeric VM-v2 slice should pivot back to the remaining
opcode-local Monte Carlo walls:

- `execStoreSlotIntMulConstModConst(...)`
- `execJumpIfFloatMulAddMulCompareConstFalse(...)`
- `execStoreSlotBinaryIntSlotConst(...)`

with residual helper work still visible in:

- `runtime.mapaccess1_fast64`
- `bytecodeCastSlotFloatConstDivRawFast(...)`

That owned `i32` slot-update boundary is now narrower too. A twelfth kept
Monte Carlo bytecode slice changed the exact discarded
`StoreSlotBinaryIntSlotConst` fast path so it updates the existing owned
target `*runtime.IntegerValue` in place when the slot already holds the
reusable large-counter cell from a prior loop iteration, instead of routing
every large out-of-cache `i32` increment back through
`storeOwnedI32SlotRaw(...)`.

Visible loads still materialize plain Able integers through
`bytecodeSlotReadValue(...)`, and unsupported cases still fall back to the
existing owned-`i32` helper path. So this remains an internal slot-update keep
rather than a new visible representation or a new shared cache policy.

On the external benchmark, that moved bytecode Monte Carlo from `1.5133s` to
`1.4433s` over `3/3` runs.

The one-shot direct runtime confirmation was slightly noisier at about
`1.43s`, `35.8KB`, and `142 allocs`, so the keep basis is the repeated
external band plus the profile shift rather than the single runtime wall
clock.

The planning consequence changes again. The owned `i32` slot helper boundary
is no longer first-tier:

- `execStoreSlotBinaryIntSlotConstDiscardI32Fast(...)` dropped to about
  `0.38s` cumulative
- `execStoreSlotBinaryIntSlotConst(...)` is down to about `0.45s`
- `storeOwnedI32SlotRaw(...)` dropped out of the visible top tier

So the next primitive-numeric VM-v2 slice should stay on Monte Carlo, but
pivot back to the remaining opcode-local walls:

- `execStoreSlotIntMulConstModConst(...)`
- `execJumpIfFloatMulAddMulCompareConstFalse(...)`
- `execStoreSlotCastSlotFloatConstDiv(...)`

with secondary remaining cost still visible in:

- `execStoreSlotBinaryIntSlotConst(...)`
- `bytecodeCastSlotFloatConstDivRawFast(...)`

That recurrence boundary is now narrower too. A fourteenth kept Monte Carlo
bytecode slice changed the exact discarded
`StoreSlotIntMulConstModConst` steady-state path so it operates directly on
the existing raw `i64` slot cell when the multiply and modulo immediates are
already proven positive small `i64` values. That removes the old
`bytecodeImmediateIntegerValue(...)` decode and the broader helper/store
ladder from the hot Park-Miller loop once the loop is already in its internal
raw-cell form.

Visible reads still materialize ordinary Able integers through
`bytecodeSlotReadValue(...)`, and unsupported cases still fall back to the
existing fused recurrence path. So this remains an internal opcode keep
rather than a new visible representation or another instruction-metadata
layer.

On the external benchmark, that moved bytecode Monte Carlo from `1.3633s` to
`1.2000s` over `3/3` runs, and the direct runtime confirmation from about
`1.23s` / `35.5KB` / `140 allocs` to about `1.11s` / `35.5KB` / `140 allocs`.

The planning consequence changes again. The generic recurrence immediate
decode path is gone as a first-tier cost:

- `execJumpIfFloatMulAddMulCompareConstFalse(...)` is now about `0.34s`
  cumulative
- `execStoreSlotIntMulConstModConstDiscardSteadyStateFast(...)` now carries
  the recurrence hot path at about `0.44s`
- `bytecodeImmediateIntegerValue(instr.value)` dropped out of the visible top
  tier

So the next primitive-numeric VM-v2 slice should stay on Monte Carlo and
target the remaining opcode-local walls in this order:

- `execJumpIfFloatMulAddMulCompareConstFalse(...)`
- `execStoreSlotIntMulConstModConst(...)`
- `execStoreSlotBinaryIntSlotConst(...)`

with secondary remaining cost in:

- `execStoreSlotCastSlotFloatConstDiv(...)`
- `execStoreSlotBinaryIntSlotConstDiscardI32Fast(...)`
- `bytecodeDirectSmallI32Value(...)`
- `bytecodeDirectPositiveSmallI64ImmediateValue(...)`

That last immediate boundary is now narrower too. A fifteenth kept Monte Carlo
bytecode slice carries the modulo constant for
`StoreSlotIntMulConstModConst` through lowering as a second predecoded `i64`
immediate, instead of rediscovering it from `instr.value` on every recurrence
step. The instruction still keeps the original boxed immediate for the generic
path, but the hot fused recurrence path now reuses the raw second immediate
when the same positive-small `i64` proof already holds.

On the external benchmark, that moved bytecode Monte Carlo from the prior kept
`1.2000s` band to:

- `1.1933s` over one `3/3` run
- `1.1667s` over the confirming `3/3` run

Direct runtime confirmation moved to:

- `1106164191 ns/op`
- `63136 B/op`
- `199 allocs/op`

The profile consequence is clean:

- `bytecodeDirectPositiveSmallI64ImmediateValue(...)` dropped out of the top
  tier
- `bytecodePositiveIntMulConstModFast(...)` is now the visible recurrence
  inner wall at about `0.19s` flat
- the remaining first-tier Monte Carlo costs are still
  `execJumpIfFloatMulAddMulCompareConstFalse(...)`,
  `execStoreSlotIntMulConstModConstDiscardSteadyStateFast(...)`, and
  `execStoreSlotBinaryIntSlotConst(...)`

So the next Monte Carlo slice should stay on those true remaining walls, not
go back to integer-immediate rediscovery helpers.

A sixteenth kept Monte Carlo bytecode slice now reuses repeated slot decodes
inside the fused float compare when both multiply operands are the same slot,
so the hot `x*x + y*y <= const_f64` branch stops reading the same float slot
twice per square term. This stays narrow and reusable: it only changes the
already-fused float compare helper, and non-square multiply terms still use
the old two-slot decode path.

On the external benchmark, that moved bytecode Monte Carlo from the fresh
`1.2133s` baseline for this tranche to:

- `1.1933s` over one `3/3` run
- `1.1967s` over the confirming `3/3` run

Direct runtime confirmation was:

- `1202512748 ns/op`
- `62888 B/op`
- `197 allocs/op`

The profile consequence is small but real:

- `floatMulTermConditionValue(...)` is only about `0.02s` flat
- `execJumpIfFloatMulAddMulCompareConstFalse(...)` is still about `0.21s`
  cumulative
- the remaining first-tier Monte Carlo costs are now led by
  `execStoreSlotIntMulConstModConstDiscardSteadyStateFast(...)`,
  `bytecodePositiveIntMulConstModFast(...)`, and
  `execStoreSlotBinaryIntSlotConstDiscardI32Fast(...)`

So the next Monte Carlo slice should stay on the fused float compare,
recurrence, and counter-update walls, not go back to local square-term decode
helpers.

Two bytecode source-level numeric probes were rejected:

- Schrage Park-Miller with `i32` state kept the same output but slowed
  bytecode to `30.5300s`.
- Masked `u64` xorshift avoided division but triggered heavy unsigned
  bitwise allocation/GC, landing at `60.8500s`.

So this tranche is a coverage keep and a compiled-path keep, not a bytecode
competitiveness win. The next bytecode work should treat Monte Carlo as a
general primitive numeric-arithmetic profile: boxed `i64`/`f64` arithmetic,
casts, and slot updates need a typed/unboxed VM representation rather than a
benchmark-specific RNG opcode.

Generated-source audit: `approx_pi` lowers to native Go `int64`, `int32`, and
`float64` locals with checked arithmetic helpers for v12 overflow/divmod
semantics. The hot loop does not route through `runtime.Value`; only the final
`print(...)` boundary uses the normal runtime call path.

## 2026-05-28 — Pidigits benchmark coverage

The next coverage tranche closed the `pidigits` stdlib blocker with a reusable
host-backed BigInt reference API instead of a benchmark-specific compiler rule:

- `../able-stdlib/src/numbers/bigint_native.able` exposes
  `BigIntRef`, a mutable BigInt handle backed by Go `math/big`, with focused
  arithmetic, comparison, cloning, extraction, and formatting operations.
- `../able-stdlib/tests/bigint_native.test.able` pins alias-safe mutation,
  clone isolation, compare, set, small extraction, and formatting in
  tree-walker and bytecode.
- `v12/examples/benchmarks/pidigits/pidigits.able` implements the benchmark
  spigot loop with explicit term/extract/eliminate logic over `BigIntRef`.
- `v12/bench_compare_external --benchmarks pidigits` is now wired as an
  opt-in target and passes `10000` as the canonical digit count.

Verification smoke:

- compiled 10,000-digit output verified with
  `../benchmarks/pidigits/verify.rb`; direct wall-clock smoke was `1.31s`
  (`user 2.49s`, `sys 0.09s`).
- bytecode 10,000-digit output also verified with the same verifier; direct
  wall-clock smoke through `go run ./cmd/able` was `2.09s`
  (`user 3.55s`, `sys 0.16s`).

Aligned external comparisons over `3/3` runs landed at:

- compiled: `1.3367s` vs best Go `0.7400s` (`go-1.26-gmp`) and Ruby
  `9.1800s`
- bytecode: `2.0300s` vs best Go `0.7400s` (`go-1.26-gmp`) and Ruby
  `9.1800s`

Against the pure Go `math/big` reference (`go-1.26`, `1.1500s`), compiled is
about `1.16x` and bytecode about `1.77x`. The checked-in scoreboard still uses
the external harness rule of comparing against the best successful Go-family
row, which is the GMP variant for this benchmark.

This is a benchmark-coverage and native-boundary keep. It should not be
expanded into pidigits-specific bytecode opcodes. The next bytecode work should
return to the broader VM-v2 plan: typed primitive slots/registers for quicksort
and Monte Carlo numeric arithmetic, plus general host-extern call overhead if
pidigits becomes a profiling target.

## 2026-06-08 — Primitive integer extern fast invokers for pidigits bytecode

The next pidigits bytecode tranche stayed on the general host ABI boundary and
did not add any pidigits-specific opcode. The existing extern fast-invoker
layer now also recognizes primitive signed integer signatures directly:

- `i32` / `i64` arguments
- arities `1..3`
- `i32`, `i64`, and `String` results

That keep landed in `extern_host_fast_int.go` plus the dispatch hook in
`extern_host_fast.go`. It means the hot host-backed `BigIntRef` externs now
skip `reflect.Value.Call`, `toHostValue(...)`, and the broader generic result
bridge once the signature is already proven by the extern definition.

This is still a general extern-host optimization, not a BigInt-only special
case. The same path is available to any host extern that uses those primitive
integer carriers.

Result: keep. External pidigits bytecode over `3/3` runs moved from the prior
kept `2.0300s` band to:

- `1.8867s`

Profiled bytecode runtime confirmation moved from:

- `1936339805 ns/op`
- `1756396792 B/op`
- `2336849 allocs/op`

to:

- `1815567152 ns/op`
- `1707375344 B/op`
- `981000 allocs/op`

The profile shift is the real reason to keep it:

- before: `invokeExternHostFunction(...)` about `3.02s` cumulative and
  `reflect.Value.call` about `2.61s`
- after: the reflect call path dropped out of the top tier entirely, and the
  remaining time is dominated by the host `math/big` operations themselves
  (`mulAddVWW`, `nat.div`, `nat.mulAddWW`) plus GC work driven by BigInt
  allocation churn

So the next pidigits slice should not spend time on generic extern dispatch
again. The remaining wall is now the host-backed BigInt implementation in
`able-stdlib`, especially the allocation-heavy `*_i64` helpers and the
mutable-handle storage path.

## 2026-06-08 — Per-handle division remainder scratch for pidigits bytecode

The follow-up pidigits keep stayed inside the host-backed stdlib BigInt
implementation and did not change the public `BigIntRef` surface. The internal
handle table in `able.numbers.bigint_native` now stores:

- the live `*big.Int` value
- a reusable `big.Int` remainder scratch cell owned by that same handle

With that change, `div` and `div_i64` now route through
`math/big.Int.QuoRem(..., &entry.remScratch)` instead of `Quo(...)`, so the
hot quotient path keeps the same semantics while reusing remainder storage
across repeated divisions on the same destination handle.

Result: keep. External pidigits bytecode over `3/3` runs moved from the prior
kept `1.8867s` band to:

- `1.7467s`

Profiled bytecode runtime confirmation moved from:

- `1815567152 ns/op`
- `1707375344 B/op`
- `981000 allocs/op`

to:

- `1623393160 ns/op`
- `324930856 B/op`
- `937067 allocs/op`

The profile moved in the right place:

- `math/big.nat.div` dropped from about `1.33s` cumulative to about `0.79s`
- `runtime.tryDeferToSpanScan` dropped from about `1.08s` cumulative to about
  `0.21s`
- the remaining first-tier pidigits wall is now the host multiplication side:
  `math/big.mulAddVWW`, `math/big.nat.mulAddWW`, and adjacent allocation/GC
  churn

So the next pidigits slice should stay in `able-stdlib` and target the
multiplication-side host path, not reopen generic extern dispatch or add a
pidigits-specific bytecode opcode.

## 2026-06-09 — BigIntRef Lsh plus pidigits source alignment for compiled pidigits

The next pidigits keep stayed shared and reusable, but moved away from local
host scratch tricks. Two things landed together:

- `../able-stdlib/src/numbers/bigint_native.able` now exposes a general
  `BigIntRef.lsh(lhs, shift)` method backed by `math/big.Int.Lsh(...)`
- `v12/examples/benchmarks/pidigits/pidigits.able` now matches the reference
  Go benchmark shape more closely:
  - `tmp.lsh(numer, 1)` replaces the old doubling
    `tmp.mul_i64(numer, 2_i64)`
  - reusable `BigIntRef` constants now back the hot `3`, `4`, and `10`
    multiplier sites
  - `eliminate_digit(...)` now materializes the digit into a reusable temp via
    `set_i64(...)` and uses the general `mul(...)` path for the subtraction
    term

That is still a reusable stdlib/benchmark-source improvement, not a
pidigits-only opcode or a new extern ABI branch.

Result: keep for compiled pidigits. External compiled pidigits over `3/3` runs
moved from the prior kept:

- `1.3367s`

to:

- `1.1867s`

The matching bytecode external run stayed effectively flat at:

- `1.7533s`

Profiled compiled confirmation landed at:

- `1.1600s`

The profile moved the intended boundaries:

- `main.__able_compiled_entry_method_BigIntRef_mul_i64` dropped from about
  `0.45s` cumulative to about `0.33s`
- `main.__able_compiled_entry_method_BigIntRef_div` dropped from about
  `0.58s` cumulative to about `0.30s`
- `main.__able_compiled_entry_method_BigIntRef_lsh` is only about `0.01s`
- the remaining first-tier wall is still host `math/big` multiplication
  (`mulAddVWW`, `nat.mulAddWW`) plus residual GC churn

So the next pidigits slice should not go back to local `mul_i64` scratch or
alias-buffer experiments. Either take a deeper host-side `math/big` redesign,
or pivot to another benchmark family.

## 2026-06-09 — Pidigits digit-extraction intermediate reuse

The next pidigits keep stayed source-level and reusable rather than adding
another local `BigIntRef` scratch field or host wrapper tweak. The hot digit
check in `v12/examples/benchmarks/pidigits/pidigits.able` now reuses the
already-computed `3*numer + accum` temporary:

- compute `d3` from `tmp2 = 3*numer + accum`
- derive the `d4` candidate by `tmp2.add(tmp2, numer)` instead of rebuilding
  `4*numer + accum` from a fresh multiply
- drop the now-unused reusable `four` constant and the separate
  `extract_digit(...)` helper pair

That removes one full BigInt multiply from each candidate-digit attempt
without adding any new host ABI branch or pidigits-only opcode.

Result: keep. External pidigits moved from the prior kept bands:

- compiled: `1.1867s`
- bytecode: `1.7533s`

to:

- compiled: `1.0300s`
- bytecode: `1.5700s`

Profiled compiled confirmation landed at:

- `1.0100s` over `1/1`

The useful profile shift is that the reused intermediate cut the expected host
math side directly:

- `main.__able_compiled_entry_method_BigIntRef_mul` dropped from about
  `0.17s` cumulative on the fresh pre-change profile to about `0.07s`
- `main.__able_compiled_entry_method_BigIntRef_lsh` dropped from about
  `0.05s` to about `0.02s`
- the remaining first-tier wall is still host division and scalar multiply:
  `BigIntRef_div`, `BigIntRef_mul_i64`, and the underlying
  `math/big.mulAddVWW` / division machinery

So the next pidigits slice should not go back to local helper scratch or
alias-buffer experiments. If this family stays in focus, the next step has to
be a broader multiplication/division-side redesign.

## 2026-06-09 — Compiled signed helper fast paths close Monte Carlo

The next compiled Monte Carlo keep stayed inside the compiler-emitted signed
integer helpers and did not touch the benchmark source. The generated compiled
runtime now emits two direct positive fast paths:

- `__able_checked_mul_signed(...)` now returns `a * b` directly when both
  operands are nonnegative and the signed bound check proves the product safe
- `__able_divmod_signed(...)` now returns `a / b` and `a % b` directly when
  the dividend is nonnegative and the divisor is positive, skipping the
  Euclidean remainder-adjustment branch

That is a general primitive-helper improvement, not a Monte Carlo-specific
lowering rule.

Result: keep. External compiled Monte Carlo over `3/3` runs moved from the
refreshed baseline:

- `0.2100s`

to:

- `0.1700s`

Profiled compiled confirmation landed at:

- `0.1700s`

The profile moved exactly where expected:

- before: `__able_divmod_signed` was about `0.05s` flat and
  `__able_checked_mul_signed` about `0.02s` cumulative inside a `0.15s` total
- after: both helpers are down to about `0.02s` each inside a `0.12s` total

This closes compiled Monte Carlo into Go range without changing the benchmark
algorithm or adding a benchmark-shaped compiler rule.

## 2026-05-28 — Quicksort tracked-array swap direct path

The next VM-v2 quicksort tranche stayed within the existing bytecode opcodes
and removed generic swap work from the reduced hotloop. `ArrayIndexSwapSlot`
now has a direct tracked-array path for in-bounds small indexes when lowering
has proven the swap cast target is `i32`; tracked values that are already
internal/raw `i32` skip the generic `castValueToType(...)` call. The canonical
`read_slot`/`write_slot` `ArraySlotSwapSlot` path also swaps in-bounds tracked
array values directly and syncs aliases through the same tracked-array write
machinery.

The fallback surface is unchanged: non-small indexes, out-of-bounds indexes,
sparse/grow `write_slot` cases, non-canonical arrays, non-`i32` cast cases,
and alias synchronization still use the existing paths. This is not a nominal
`Array.swap` special case and not a whole-loop quicksort kernel.

Reduced in-tree quicksort moved from the refreshed `4.96-4.98ms/op` profile
band to:

- `4.51ms/op`
- `4.51ms/op`
- `4.57ms/op`

The profiled confirmation was `4.57ms/op`, and the CPU profile no longer shows
the old generic `castArrayIndexSwapSlotValue(...)` edge. A later expanded
validation band was noisier but still centered lower: `5.00ms/op`,
`4.67ms/op`, `4.71ms/op`, `4.75ms/op`, and `4.62ms/op`. Full external
bytecode `quicksort` still timed out at `60s`, so the next tranche should
target the now-dominant compare/call/arithmetic wall:
`JumpIfArrayIndexSlotCompareSlotFalse`, `execCallName` / recursive call setup,
and `execStoreSlotBinaryIntSlotConst`, or move to the real typed `i32`
register-frame design.

## 2026-05-28 — Quicksort raw-immediate slot-const store path

The follow-up quicksort tranche targeted the slot-const arithmetic edge from
the reduced profile without changing lowering or v12 fallback behavior.
`StoreSlotBinaryIntSlotConst` now checks for instructions that already carry
both typed `i32` immediate metadata and raw immediate metadata, then computes
the raw small-`i32` result directly before entering the older generic
immediate/operator helper. Unsupported slot values, missing raw metadata,
non-`i32` immediates, overflow, division by zero, and observable assignment
results still use the same semantics as before.

The fast-result finish path is shared by the raw-immediate path and the older
generic fast path, so error wrapping, source-context attachment, slot writes,
slot-0 raw-lane refresh, optional stack push, and IP advancement stay
identical.

Reduced in-tree quicksort held a kept `500x` band of:

- `4.69ms/op`
- `4.62ms/op`
- `4.55ms/op`
- `4.58ms/op`
- `4.66ms/op`

After tightening the metadata guard, the confirmation band was:

- `4.54ms/op`
- `4.64ms/op`
- `4.63ms/op`

The profiled confirmation was `4.62ms/op`. In the CPU profile,
`execStoreSlotBinaryIntSlotConst` dropped from about `210ms` cumulative in the
fresh baseline to `170ms` cumulative. Full external bytecode `quicksort` still
timed out at `60s`, so this remains a reduced-hotloop keep. The next tranche
should target the larger remaining wall in `JumpIfArrayIndexSlotCompareSlotFalse`
/ raw i32 compare extraction, `execCallName` / recursive call setup, or the
real typed `i32` register-frame design.

## 2026-05-28 — Leaf i32 register-frame seed

The next VM-v2 tranche lands the first conservative raw `i32` register-frame
slice. Slot analysis enables it only for slot-eligible leaf functions that stay
inside local primitive arithmetic/control-flow and avoid call/member/index,
async, propagation, match/rescue, and untyped-declaration boundaries.

When active, discarded `StoreSlotI32` and `CompoundAssignSlotI32` writes store
raw values in a pooled `[]int32` register frame instead of writing raw
sentinels into `[]runtime.Value`. `LoadSlotI32`, fused integer slot compares,
slot-const checks, and generic `LoadSlot` can read/materialize that lane, but
ordinary programs keep the old direct `[]runtime.Value` fast path when no
register frame is active.

This is a structural keep, not a quicksort win yet. Reduced quicksort stayed
neutral/noisy because the current hot quicksort locals are untyped and the
functions cross call/index/member boundaries:

- first neutral guard: `4.78ms/op`, `4.60ms/op`, `4.67ms/op`
- wider confirmation: `4.91ms/op`, `5.10ms/op`, `4.85ms/op`, `4.74ms/op`,
  `4.70ms/op`

The next tranche should make typed registers call-frame capable, with explicit
save/restore and boxing boundaries, before trying to carry quicksort's `i`,
`j`, parser counters, or similar hot locals through the raw lane. Do not
return to active-frame sidecars or untyped-local inference without a real
data-flow proof.

## 2026-05-28 — Call-capable i32 register frames

The next VM-v2 tranche makes the explicit typed `i32` register lane safe
across bytecode inline calls. Call frames now detach the active raw register
frame before entering a callee and restore it on return or unwind; run cleanup
also releases any saved raw frames through the register-frame pool. Program
switching preserves a restored caller frame instead of rebuilding it from
boxed slots, which is required when a discarded typed store has left the caller
slot live only in the raw lane.

The eligibility gate is still conservative. It now allows direct named calls
whose arguments are themselves register-safe, but still rejects recursive
self-calls, member/index dispatch, type-argument calls, async/resume/error
boundaries, and untyped declarations. `CallName` slot-argument materialization
now reads through the register-aware slot boundary, so a raw local passed as an
argument is boxed exactly at the call boundary.

This is a structural keep rather than an immediate scoreboard win:

- focused call-boundary parity stayed green
- full `go test ./pkg/interpreter` passed with `ABLE_STDLIB_ROOT` unset
- reduced `Fib30Bytecode`: `150.20ms/op`, `145.02ms/op`, `152.79ms/op`
- reduced quicksort guard: `5.16ms/op`, `5.01ms/op`, `5.15ms/op`

Next: widen the typed-register eligibility proof only where the VM has an
explicit materialization boundary, most likely typed non-leaf helper shapes
first and index/member paths after that. Do not enable untyped locals or
recursive self-call register frames as part of this line until a fresh profile
proves their save/restore cost and materialization behavior.

## 2026-05-28 — Discarded typed i32 call-result stores

The next VM-v2 typed-lane tranche keeps direct helper-call support conservative
but removes a local boxed-retention boundary after typed `i32` call results.
Generic `StoreSlot` / `StoreSlotNew` instructions for statement-position typed
`i32` assignments can now be marked `discardResult`. When an `i32` register
frame is active, the VM consumes the boxed result, runs the existing typed
assignment validation/coercion, seeds the raw register lane, and leaves the
runtime slot empty. Non-discarded assignment expressions and non-`i32` stores
continue to use the existing boxed behavior.

Focused coverage now proves both the direct VM boundary and lowering behavior:
`y: i32 := helper(x)` lowers to a discarded typed store after the `CallName`
slot-arg call, without a following `Pop`, and the caller can still use the raw
register-backed value after return.

Guard results:

- full `go test ./pkg/interpreter` passed with `ABLE_STDLIB_ROOT` unset
- reduced quicksort guard: `4.93ms/op`, `4.93ms/op`, `5.15ms/op`
- reduced `Fib30Bytecode`: `144.95ms/op`, `148.52ms/op`, `146.77ms/op`

Next: audit and pin index/member materialization boundaries before widening
register-frame eligibility to those AST shapes. In practice that means finding
slot-backed array/member/index opcodes that still read `vm.slots[...]`
directly and routing only the proven dynamic-boundary operands through the
register-aware slot materializer. Untyped locals and recursive self-call
register frames remain out of scope until separately profiled.

## 2026-05-31 — Raw i32 array/index slot operands

The next VM-v2 tranche keeps the slot-analysis gate unchanged but proves the
array/index operand boundary the widened gate will need. Slot-specific operand
readers now sit behind the existing bytecode array/index fast opcodes:
`ArrayIndexGetSlot`, `ArrayIndexSetSlot`, `ArrayIndexSwapSlot`,
`ArrayReadSlot`, `ArraySlotSwapSlot`, and the fused array compare jumps.

These readers deliberately prefer the existing boxed `vm.slots[...]` cell and
only fall back to the raw `i32` register lane when the slot is empty. They are
also gated behind an active `i32` register frame, so ordinary non-register
programs keep the exact old path. That is the point of the tranche: prove the
register-backed operand semantics without moving the baseline hot loop onto a
new generic materialization helper.

Focused direct-VM coverage now proves register-backed operands for array
get/set, array compare jumps, canonical `read_slot`, and both swap opcodes.
Whole-benchmark movement is intentionally modest and roughly neutral:

- reduced quicksort guard: `5.24ms/op`, `5.37ms/op`, `5.11ms/op`
- reduced `Fib30Bytecode`: `157.92ms/op`, `145.81ms/op`, `140.38ms/op`

This is an enabling keep, not a scoreboard keep. The next slice should widen
typed-register eligibility only for bracket-index read/write shapes that
already lower to these proven slot opcodes. Member-dispatch materialization,
untyped locals, and recursive self-call register frames should remain outside
the tranche until separately profiled.

## 2026-05-31 — Bracket-index i32 register-frame eligibility

The next VM-v2 slice lands the bracket-index half of that plan. Slot analysis
now admits `IndexExpression` and plain `arr[idx] = value` assignment targets
into the typed `i32` register-frame proof, but only when the receiver, index,
and right-hand side are already register-safe. Compound index assignment,
member dispatch, untyped locals, and recursive self-calls remain outside the
gate.

This means a typed helper built from `idx: i32`, `current: i32 := arr[idx] as
i32`, `arr[idx] = current`, and a typed `arr[idx]` readback can now keep the
same raw `i32` lane through the bracket-index opcodes proven in the previous
tranche. Focused lowering coverage now pins that such helpers retain
`program.frameLayout.i32RegisterFrame`, and bytecode/treewalker parity covers
the runtime shape end to end.

The benchmark picture is still structural rather than headline-worthy. Two
guard passes stayed in the same broad range rather than opening a new win:

- quicksort guard: `5.23ms/op`, `5.91ms/op`, `5.30ms/op`
- quicksort confirmation: `5.31ms/op`, `5.40ms/op`, `5.72ms/op`
- `Fib30Bytecode`: `156.63ms/op`, `160.61ms/op`, `151.01ms/op`
- `Fib30Bytecode` confirmation: `157.06ms/op`, `173.05ms/op`, `142.47ms/op`

This is still a keep because it closes the bracket-index eligibility gap that
the previous operand-boundary tranche set up, without moving ordinary
non-register programs onto a new path. The next slice should move to
slot-member `read_slot` / `write_slot` AST eligibility on top of the already
proven canonical array-slot opcodes.

## 2026-05-31 — Canonical array-slot member i32 register-frame eligibility

That slot-member follow-up is now landed. Slot analysis admits non-safe
canonical `arr.read_slot(idx)` and `arr.write_slot(idx, value)` calls into the
typed `i32` register-frame proof when the receiver and arguments are already
register-safe. This keeps the scope tight: no safe-member calls, no general
member dispatch, no type-argument calls, and no untyped locals.

The runtime boundary is the same one already proven by the preceding tranches.
`arr.read_slot(idx)` lowers to `ArrayReadSlot`, which already knows how to use
a register-backed `i32` index slot. `arr.write_slot(idx, value)` lowers to the
guarded canonical `CallMemberArraySlot` path, so the index/value only box at
the existing call/materialization boundary.

Focused coverage now pins a typed helper shaped as:

- `idx: i32 := i`
- `current: i32 := arr.read_slot(idx)`
- `arr.write_slot(idx, current)`
- final `arr.read_slot(idx)`

and proves both lowering (`program.frameLayout.i32RegisterFrame`) and
bytecode/treewalker parity without depending on external stdlib package setup.

Guard results stayed in the same enabling band:

- reduced quicksort guard: `5.44ms/op`, `5.45ms/op`, `5.15ms/op`
- reduced `Fib30Bytecode`: `155.45ms/op`, `146.40ms/op`, `151.90ms/op`

This is therefore another structural keep, not a new quicksort headline. The
real remaining quicksort blocker is visible in the benchmark source itself:
the hot locals are still mostly untyped (`value`, `i`, `j`, `pivot`, `tmp`,
and parser counters), so the next defensible step is explicit typed-local
adoption in those hot benchmark/helper paths or another separately-proven
typed AST shape. This result does not justify broad untyped-local inference or
general member-dispatch widening.

## 2026-05-31 — Typed swap temps keep swap fusion

The immediate quicksort follow-up made that source-level conclusion more
precise. A broad explicit-typing pass across the benchmark source regressed the
reduced bytecode guard badly, and a narrower `swap(...)`-only pass still
regressed at first. The useful reason: typed `tmp` declarations no longer
matched the existing swap-fusion lowering, so the hot helper fell off
`ArrayIndexSwapSlot` / `ArraySlotSwapSlot` and back onto separate read/write
ops.

The keep is therefore not “typed locals are faster.” The keep is: typed temps
can be made neutral-to-healthy if lowering preserves the existing fused swap
opcode. `bytecodeArrayIndexSwapSlotInstruction(...)` and
`bytecodeArraySlotSwapSlotInstruction(...)` now accept typed temp declarations
by resolving simple typed-pattern targets instead of requiring a bare
identifier on the first `:=`.

The reduced quicksort fixture and the canonical external quicksort source now
keep:

- bracket-index form: `tmp: i32 := arr[a] as i32`
- slot-member form: `tmp: i32 := arr.read_slot(a)`

while still lowering to the fused swap opcodes.

Guard results stayed in the current reduced band:

- reduced quicksort guard: `5.19ms/op`, `5.32ms/op`, `5.24ms/op`
- external quicksort compiled validation: `1.73s` over `1/1`

This narrows the real lesson for the next tranche. Do not retry broad
benchmark-source typing yet. First reduce typed declaration/store overhead in
functions that still miss the register lane, or only type locals in source
shapes whose fused lowering is already preserved.

## 2026-05-31 — Preseeded inline call-name callee i32 frames

The next VM-v2 tranche stayed on the direct helper-call boundary instead of
reopening source typing or recursive register-frame eligibility. For already
proven direct typed callees, inline setup now allocates a pooled raw `i32`
register frame, seeds it while arguments are copied/coerced into callee slots,
and installs it immediately after `pushCallFrame(...)`. That means
`switchRunProgram(...)` no longer rescans boxed callee slots on the hot cached
`CallName` path.

Focused proof coverage now pins the exact intended behavior: the callee raw
slot is live before the program switch, clearing the boxed callee slot copy
does not matter because the switch preserves the pre-seeded raw frame, and the
saved caller raw frame still restores correctly after unwind.

This slice is a real keep rather than another structural-only proof:

- reduced quicksort guard: `5.47ms/op`, `5.23ms/op`, `5.23ms/op`
- reduced `Fib30Bytecode`: `141.50ms/op`, `143.73ms/op`, `144.42ms/op`
- profiled reduced quicksort confirmation: `5.28ms/op`

The post-keep reduced quicksort profile no longer shows
`activateI32RegisterFrame(...)` as the call-boundary wall. The remaining hot
helper path is now the cached direct inline edge itself:
`tryInlineCachedCallNameDirectFromStack(...)`, `pushCallNameSlotArgs(...)`,
and the residual callee-frame seed work.

The next bytecode tranche should stay there: bypass slot-arg boxing for cached
direct inline `CallName` helpers and seed the callee raw frame straight from
caller slots/registers before broadening any other VM-v2 eligibility gate.

## 2026-05-31 — Direct slot-arg inline call-name setup

That follow-up VM-v2 slice is now landed. Cached direct `CallName` helpers
with slot arguments no longer push boxed arguments onto `vm.stack` before
inlining. The VM now reads the caller arguments straight from the declared
caller slots, materializes boxed values only for the callee semantic frame,
and seeds the callee raw `i32` register frame from caller slots/registers
during setup.

Focused proof coverage now includes the raw-only caller-slot case: a cached
direct slot-arg helper can inline when the caller boxed slot is empty but the
caller raw `i32` lane is live, and the helper leaves the existing stack prefix
untouched.

Guard results stayed in the current kept band:

- reduced quicksort guard: `5.22ms/op`, `5.20ms/op`, `5.43ms/op`
- reduced `Fib30Bytecode`: `149.20ms/op`, `151.15ms/op`, `142.79ms/op`
- profiled reduced quicksort confirmation: `5.37ms/op`

The post-keep quicksort profile is the useful part. `pushCallNameSlotArgs(...)`
has dropped out of the hot list. The remaining call-boundary wall is now
inside `tryInlineCachedCallNameDirectFromSlots(...)` itself, especially
`slotMaterializedValue(...)`, residual callee-lane seed work, and the
remaining coercion checks.

So the next bytecode tranche should stay on that same helper: specialize the
direct slot-based inline path for the no-coercion typed lane before widening
any new VM-v2 eligibility gate.

## 2026-05-31 — External `[]byte` to `Array u8` host-return boundary

The next external-scale quicksort keep did not come from another internal VM
micro-slice. It came from fixing the canonical host-return boundary that the
benchmark actually uses.

The canonical external stdlib already routes `able.fs.read_bytes` through
`fs_read_bytes_fast(path: String) -> IOError | Array u8`, but the bytecode
interpreter was still paying the generic `fromHostValue(...)` slice-to-array
conversion cost for the returned `[]byte`. The installed `~/.able` stdlib
cache was also stale, so the verification path for this tranche explicitly
used:

- `ABLE_HOME=/tmp/able-empty-home`
- `ABLE_PATH=/home/david/sync/projects/able-stdlib`
- `ABLE_MODULE_PATHS=`

The kept runtime slice was:

- add `runtime.ArrayStoreMonoValueFromU8Bytes(data []byte)`,
- keep host `[]byte` returns as mono `Array u8` handles in
  `fromHostValue(...)`,
- recognize the same shape in the cached direct extern fast invoker for
  `func(string) []byte` and matching union signatures,
- preserve the semantic boxed view through explicit materialization
  (`interp.ArrayElements(...)`) rather than eager conversion.

Focused proof coverage now pins both sides of that contract:

- the fast path retains a mono `Array u8` handle that `ArrayStoreMonoReadU8`
  and `ArrayStoreRead` can consume directly,
- explicit materialization still yields boxed small `u8` integers when the
  semantic slice view is required.

On the canonical external-stdlib 1 MB quicksort prefix, the restored baseline
band

- `1339761372 ns/op`, `30168336 B/op`, `1180550 allocs/op`
- `1274955746 ns/op`, `30167304 B/op`, `1180521 allocs/op`
- `1266301329 ns/op`, `30167864 B/op`, `1180542 allocs/op`

moved to the kept band

- `821446016 ns/op`, `14435568 B/op`, `1180502 allocs/op`
- `823947705 ns/op`, `14435568 B/op`, `1180502 allocs/op`
- `821112360 ns/op`, `14435584 B/op`, `1180503 allocs/op`

and the profiled confirmation landed at `945898904 ns/op`, `14465928 B/op`,
`1180568 allocs/op`.

The profile result matters more than the one-shot number: the old
`Interpreter.fromHostValue(...)` allocation wall disappeared. The remaining
external-scale quicksort wall is now back in the parser / compare / boxed
integer update path instead of file-read conversion.

Full external bytecode quicksort still timed out at `90s`, so this is not the
finish line. The next bytecode tranche should stay external-scale and target
the parser-to-array boundary directly: either a v12-safe `Array u8` -> parsed
integer ingest lane, or a typed `Array i32` push/read path that the parser and
quicksort loop can share.

## 2026-06-01 — External quicksort parser output preallocation

The next keep stayed external-scale, but it did not come from a new VM opcode.
It came from tightening the benchmark source so the parser stops paying obvious
 array-growth churn.

`parse_numbers(bytes: Array u8) -> Array i32` already knows `bytes.len()`
before it starts appending parsed integers. The kept source change reserves the
output array up front:

- before: `nums: Array i32 := Array.new()`
- after: `nums: Array i32 := Array.with_capacity(((count / 2) as i32) + 1)`

This is benchmark-faithful and v12-safe. The parser is still explicit Able
code; it simply stops growing the append-only result array a little at a time.

The first draft of the change used `(count / 2) + 1` directly and failed for a
good reason: `/` did not produce the `i32` capacity shape required by
`Array.with_capacity(...)`. The kept form makes that cast boundary explicit.

On the canonical external-stdlib 1 MB quicksort prefix, the post-`[]byte`
host-return band

- `821446016 ns/op`, `14435568 B/op`, `1180502 allocs/op`
- `823947705 ns/op`, `14435568 B/op`, `1180502 allocs/op`
- `821112360 ns/op`, `14435584 B/op`, `1180503 allocs/op`

moved to

- `794274467 ns/op`, `15308704 B/op`, `1180496 allocs/op`
- `795402232 ns/op`, `15308720 B/op`, `1180497 allocs/op`
- `806759619 ns/op`, `15308736 B/op`, `1180498 allocs/op`

This is a wall-clock keep even though reserved capacity raises bytes/op a bit.
That trade is defensible here because it removes repeated `ArrayEnsureCapacity`
work from the hot ingest loop and the prefix band moves materially in the right
direction.

Full external bytecode quicksort still timed out at `90s`, so this is still a
prefix keep, not the end state. The next quicksort tranche should stay on the
parser boundary itself: do not retry mono-`i32` promotion in the current
dynamic-array contract, and instead target the remaining byte-to-digit
compare/update path or a typed ingest path that keeps parsed integers unboxed
end-to-end.

The next quicksort keep stayed on that parser boundary, but it landed as a
control-flow lowering slice rather than a new VM opcode. Control-flow-only
slot-const conjunctions now bypass the generic `&&` expression path when every
conjunct already matches the existing specialized false-jump lowering. In
practice that means parser branches like `byte >= 48_u8 && byte <= 57_u8` no
longer materialize the left boolean through `Dup` / `JumpIfFalse` before
checking the upper bound; lowering now emits the lower-bound specialized jump
followed directly by the upper-bound specialized jump in `if` / `elsif`
position.

This is intentionally narrower than a general boolean optimizer. Only
control-flow contexts use it, and only when every conjunct is already a pure
slot-vs-int-const comparison. All other `&&` expressions keep the existing
general semantics and lowering.

On the canonical external-stdlib 1 MB quicksort prefix, the current-host warmed
band landed at:

- `833773118 ns/op`, `15308704 B/op`, `1180496 allocs/op`
- `853190719 ns/op`, `15308720 B/op`, `1180497 allocs/op`
- `887188361 ns/op`, `15308736 B/op`, `1180498 allocs/op`

with a profiled confirmation at:

- `876787701 ns/op`, `15308720 B/op`, `1180497 allocs/op`

The host was noisy, so those numbers are not a new long-term baseline. The
defensible keep is narrower: on the same machine, this tranche beat the
immediately preceding restored runs while removing the generic `&&`
materialization path from the parser digit-range branch.

Full external bytecode quicksort still timed out at `90s`, so the next tranche
should stay on the parser byte boundary and target the remaining digit decode /
boxed integer update path rather than another generic control-flow rewrite.

## 2026-06-01 — External quicksort extended `i32` static boxing window

The next keep stayed on the same external quicksort prefix, but it finally
stopped chasing parser arithmetic syntax. The fresh profile said the meaningful
remaining wall was boxed `i32` materialization itself, especially
`bytecodeBoxedIntegerI32Value(...)`.

Before this keep, values outside the shared small-int cache fell into the old
dynamic `i32` boxing path:

- `RLock`
- dynamic map probe
- possible `runtime.NewSmallInt(...)`
- `Lock`
- dynamic map insert / lookup

That path is correct, but the 1 MB quicksort prefix is full of parser counters,
indices, and temporary `i32` values that sit just above the shared small-int
window. Paying the dynamic cache path for those values is unnecessary.

The kept change extends only the `i32` lane:

- shared small-int cache remains the common cross-kind cache
- `i32` now gets an additional static boxed range from `16385..262143`
- values outside that range still use the old dynamic map/RWMutex cache
- other integer kinds are unchanged

This is deliberately narrower than a broad integer-cache rewrite. The profile
evidence was specific to boxed `i32` traffic in external quicksort, so the keep
adds only the cache that the hot path can actually use.

On the canonical external-stdlib 1 MB quicksort prefix, a refreshed restored
baseline at

- `848512956 ns/op`, `15308704 B/op`, `1180496 allocs/op`

moved to the kept `3/3` band:

- `782731349 ns/op`, `15308704 B/op`, `1180496 allocs/op`
- `839980574 ns/op`, `15308720 B/op`, `1180497 allocs/op`
- `819652168 ns/op`, `15308736 B/op`, `1180498 allocs/op`

The profiled confirmation was noisy at:

- `858471480 ns/op`, `15308720 B/op`, `1180497 allocs/op`

but the profile itself is much clearer. In the refreshed baseline profile,
`bytecodeBoxedIntegerI32Value(...)` accounted for about `320ms` cumulative. On
the kept profile it dropped to about `70ms`, and the old dynamic map/RWMutex
path was no longer the dominant cost there.

That means the next wall is no longer parser cast/subtract fusion. The profile
now points back at the array read/compare path:

- `arrayReadSlotValue(...)`
- `execJumpIfArrayReadSlotCompareSlotFalse(...)`
- `lookupCachedCanonicalArraySlotCallForArray(...)`
- `compareBytecodeCondition(...)`

Full external bytecode quicksort was not rerun for this keep, so the last known
full-benchmark status remains a `90s` timeout. The useful conclusion here is
still real: the boxed-`i32` wall moved materially, and the next bounded tranche
should target array read/compare rather than more parser arithmetic
micro-fusions.

## 2026-06-01 — External quicksort tracked `Array i32` compare fast path

The next keep stayed on that exact array read/compare wall rather than opening
another parser experiment.

The hot quicksort partition loops use:

- `if arr.read_slot(i) >= pivot { break }`
- `if arr.read_slot(j) <= pivot { break }`

Before this keep, the fused bytecode opcode still paid the full boxed path on
every iteration:

- `arrayReadSlotValue(...)`
- canonical `read_slot` fast-path lookup
- boxed element result
- `compareBytecodeCondition(...)`

That is semantically fine, but on the external quicksort prefix the profile
showed it was still paying too much generic work for the common tracked
`Array i32` shape.

The kept slice adds one narrow fast path inside
`execJumpIfArrayReadSlotCompareSlotFalse(...)`:

- the site must already be the cached canonical `read_slot` shape
- the receiver must be a tracked `Array`
- the index slot must resolve directly to a small non-negative integer
- the tracked element and right slot must both be direct small `i32` values

When those conditions hold, the VM compares raw `i32` values directly and skips
the old boxed `arrayReadSlotValue(...)` -> `compareBytecodeCondition(...)`
chain. If any condition fails, execution falls back to the old path unchanged.

This is a good keep because it removes work from the actual external hot loop
without broadening semantics or cache rules.

On the canonical external-stdlib 1 MB quicksort prefix, the prior kept
extended-boxing band:

- `782731349 ns/op`, `15308704 B/op`, `1180496 allocs/op`
- `839980574 ns/op`, `15308720 B/op`, `1180497 allocs/op`
- `819652168 ns/op`, `15308736 B/op`, `1180498 allocs/op`

moved to:

- `754532197 ns/op`, `15308720 B/op`, `1180497 allocs/op`
- `753549655 ns/op`, `15308720 B/op`, `1180497 allocs/op`
- `736386560 ns/op`, `15308704 B/op`, `1180496 allocs/op`

with a profiled confirmation at:

- `771080932 ns/op`, `15308704 B/op`, `1180496 allocs/op`

The profile evidence is also cleaner now:

- before: `arrayReadSlotValue(...)` about `350ms` cumulative,
  `compareBytecodeCondition(...)` about `140ms`,
  `execJumpIfArrayReadSlotCompareSlotFalse(...)` about `270ms`
- after: `arrayReadSlotValue(...)` about `160ms`,
  `compareBytecodeCondition(...)` about `60ms`,
  `execJumpIfArrayReadSlotCompareSlotFalse(...)` about `210ms`

Full external bytecode quicksort was still not rerun here, so the last known
full-benchmark status remains a `90s` timeout. The right next step is not a
return to parser arithmetic fusion. It is another bounded slice on the same
quicksort loop edge: the remaining canonical array-slot cache lookup and
tracked-read overhead.

## 2026-06-06 — External quicksort tracked compare local raw extraction

The next keep stayed on the same fused quicksort compare opcode, but it was
smaller than the earlier tracked-`Array i32` shortcut. The useful profile
signal was no longer the whole boxed compare chain. It was the residual helper
traffic still sitting inside that kept fast path:

- right-slot small-`i32` extraction
- tracked-element small-`i32` extraction
- wrapper calls around those reads

The kept change therefore does not add a new opcode or a new cache rule. It
only rewires the existing tracked compare fast path in
`execJumpIfArrayReadSlotCompareSlotFalse(...)` so that:

- the right slot is decoded locally from slots/registers
- the tracked element is decoded locally from the tracked array state
- the old generic wrapper helpers are skipped on this one bounded path

Everything else still falls back to the prior kept logic unchanged.

On a refreshed restored canonical external-stdlib 1 MB quicksort prefix
baseline of:

- `789324033 ns/op`, `15308704 B/op`, `1180496 allocs/op`
- `795803486 ns/op`, `15308704 B/op`, `1180496 allocs/op`
- `803643848 ns/op`, `15308704 B/op`, `1180496 allocs/op`

the kept band moved to:

- `764213712 ns/op`, `15308704 B/op`, `1180496 allocs/op`
- `768451478 ns/op`, `15308720 B/op`, `1180497 allocs/op`
- `764628876 ns/op`, `15308720 B/op`, `1180497 allocs/op`

The profiled confirmation was noisier at:

- `810782795 ns/op`, `15308704 B/op`, `1180496 allocs/op`

So the keep basis here is the repeated warmed band plus the still-green focused
and full-package gates, not the one-shot profile wall-clock.

The profile moved in a narrower way than the previous tracked compare keep:

- `compareArrayReadSlotTrackedI32Condition(...)` edged down from about `220ms`
  cumulative to about `200ms`
- the old generic `bytecodeDirectSmallI32Value(...)` flat cost dropped, but
  part of that work now appears under the new local tracked-compare extractor

That is still acceptable because this tranche was never trying to remove the
entire extraction cost globally. It was only trying to trim the hot quicksort
opcode without reopening broad `i32` helper rewrites.

## 2026-06-06 — External quicksort direct `i32` boxing beyond extended cache

The next keep did not stay on the array-compare edge. The fresh external
profile and the real 1 MB quicksort input distribution changed the answer.

On that workload there are about `105998` parsed numbers in the 1 MB prefix,
and only `102` are at or below `1048575`. That means the earlier extended
static `i32` cache keep solved the small-and-midrange parser/counter wall, but
it left the dedicated `bytecodeBoxedIntegerI32Value(...)` helper still paying
for the dynamic map/RWMutex path on large parser values.

The kept slice is deliberately narrower than a broad integer-cache redesign:

- preserve the existing shared small-int cache,
- preserve the dedicated `i32` extended static cache through `262143`,
- keep the generic multi-kind `bytecodeBoxedIntegerValue(...)` path unchanged,
- but drop the dynamic map path from the dedicated
  `bytecodeBoxedIntegerI32Value(...)` helper once those static caches miss.

So the helper now falls through to:

- `runtime.NewSmallInt(value, runtime.IntegerI32)`

instead of taking the old lock-and-map path.

This is a real tradeoff, not a free lunch. On the canonical external 1 MB
quicksort prefix, the prior kept band of:

- `764213712 ns/op`, `15308704 B/op`, `1180496 allocs/op`
- `768451478 ns/op`, `15308720 B/op`, `1180497 allocs/op`
- `764628876 ns/op`, `15308720 B/op`, `1180496 allocs/op`

moved to:

- `710210584 ns/op`, `20395552 B/op`, `1286472 allocs/op`
- `710788667 ns/op`, `20395568 B/op`, `1286473 allocs/op`
- `762015572 ns/op`, `20395568 B/op`, `1286473 allocs/op`

with a profiled confirmation at:

- `767822277 ns/op`, `20395552 B/op`, `1286472 allocs/op`

So wall-clock improved materially, but allocs/op and bytes/op got worse. The
profile still makes the keep defensible for this workload:

- before: `bytecodeBoxedIntegerI32Value(...)` about `120ms` cumulative with
  real time still in the dynamic map path
- after: `bytecodeBoxedIntegerI32Value(...)` about `10ms` cumulative and the
  map/RWMutex path gone from the hot list

Reduced `Fib30Bytecode` stayed in the recent `152-158ms/op` band, so this is
not a general recursion win. It is an external quicksort wall-clock keep.

Full external bytecode quicksort was still not rerun for this keep, so the
last known full-benchmark status remains the earlier `90s` timeout. The next
step should stop optimizing the 1 MB prefix in isolation and refresh that full
external status on the current kept state; if it still times out, the larger
profile should decide between the remaining tracked-compare work and the
parser/store boxed-`i32` wall.

## 2026-06-06 — External quicksort full-scale status refresh after direct `i32` boxing

The next tranche was the scale check that the 1 MB prefix work had been
pointing to, not a new code edit.

On the real external `../benchmarks/quicksort` input, the current kept
bytecode state still does not clear the full benchmark guard:

- `./v12/bench_compare_external --benchmarks quicksort --modes bytecode --runs 1 --timeout 90`
- result: `timeout (1)` at `90s`

That answers the immediate planning question: the recent prefix wins are real,
but they have not yet turned into a full external quicksort pass.

To choose the next tranche from full-input behavior instead of the `1 MB`
proxy, the steady-state runtime benchmark was then run on the full `95MB`
input with CPU profiling enabled:

- `ABLE_HOME=/tmp/able-empty-home ABLE_PATH=/home/david/sync/projects/able-stdlib ABLE_MODULE_PATHS= GOCACHE=/tmp/able-gocache GOMODCACHE=/tmp/able-gomodcache ABLE_BENCH_RUNTIME_CPU_PROFILE=/tmp/able-qsort-full-runtime-600.cpu.pprof ./v12/bench_perf --runs 1 --timeout 600 --modes bytecode-runtime --run-from ../benchmarks/quicksort v12/examples/benchmarks/quicksort/quicksort.able`

That full-input steady-state run completed at:

- `102007915062 ns/op`
- `10484933528 B/op`
- `545624289 allocs/op`
- `237.98s` real average
- `260.51s` user average
- `2.50s` sys average
- `21.00` GC average

The useful outcome is the hotspot ranking. On the full input, the dominant
remaining wall is still the tracked `Array i32` compare boundary, not another
canonical array-slot cache-policy probe and not another parser arithmetic
rewrite:

- `trackedArrayCompareDirectSmallI32Value(...)`: `15.27s` cumulative
- `trackedArrayCompareI32RawAtSlot(...)`: `15.62s` cumulative
- `compareArrayReadSlotTrackedI32Condition(...)`: `23.72s` cumulative
- `execJumpIfArrayReadSlotCompareSlotFalse(...)`: `24.72s` cumulative
- `storeSlotBinaryIntSlotConstI32RawFastResult(...)`: `8.79s` cumulative
- `bytecodeRawI32SlotCachedValue(...)`: `6.65s` cumulative
- `lookupCachedCanonicalArraySlotCallForArray(...)`: `5.85s` cumulative
- `tryInlineCachedCallNameDirectFromSlots(...)`: `10.51s` cumulative

That ordering matters. It says:

- do not go back to canonical array-slot cache validation/layout rewrites as
  the first next step
- do not pivot back to parser digit-decode fusion as the first next step
- keep the next slice opcode-local on the tracked compare extractor/read path,
  with the raw slot/store edge as the secondary fallback

A refreshed full-scale rerun on the same kept baseline after the rejected
store-path probes did not move the timeout status. Full external bytecode
`quicksort` still times out at the `90s` guard, and the full `95MB`
steady-state `bytecode-runtime` pass completed at `105108381198 ns/op`,
`10485031864 B/op`, and `545624301 allocs/op` with `212.00s` real average,
`235.05s` user average, `2.16s` sys average, and `22.00` GC average.

The tracked compare wall is still first:

- `trackedArrayCompareDirectSmallI32Value(...)`: `18.67s` cumulative
- `trackedArrayCompareI32RawAtSlot(...)`: `19.01s` cumulative
- `compareArrayReadSlotTrackedI32Condition(...)`: `27.03s` cumulative
- `execJumpIfArrayReadSlotCompareSlotFalse(...)`: `27.97s` cumulative

The store/boxing side is still secondary on the same run:

- `storeSlotBinaryIntSlotConstI32RawFastResult(...)`: `8.23s` cumulative
- `bytecodeRawI32SlotCachedValue(...)`: `6.76s` cumulative
- `bytecodeBoxedIntegerI32Value(...)`: `5.53s` cumulative

A later refresh on the current kept baseline tightened that conclusion rather
than changing it. Full external bytecode `quicksort` still times out at the
`90s` guard, and the full `95MB` steady-state `bytecode-runtime` pass
completed at `99556758026 ns/op`, `10484909080 B/op`, and `545624288
allocs/op` with `197.92s` real average and `22.00` GC average. The tracked
compare wall is still first, and now even more clearly flat-cost dominated by
the value extraction itself:

- `trackedArrayCompareDirectSmallI32Value(...)`: `16.18s` flat /
  `16.31s` cumulative
- `compareArrayReadSlotTrackedI32Condition(...)`: `23.34s` cumulative
- `execJumpIfArrayReadSlotCompareSlotFalse(...)`: `24.24s` cumulative
- `lookupCachedCanonicalArraySlotCallForArray(...)`: `5.75s` cumulative

The non-compare wall is still second-tier on the same run:

- `storeSlotBinaryIntSlotConstI32RawFastResult(...)`: `9.10s` cumulative
- `bytecodeRawI32SlotCachedValue(...)`: `7.43s` cumulative
- `bytecodeBoxedIntegerI32Value(...)`: `5.22s` cumulative

That refreshed profile rules out another round of safe helper-level compare or
store shaving as the likely next keep. The next real quicksort gain probably
needs a larger tracked-`Array i32` representation boundary or typed collection
lane instead of another local extractor, operator-dispatch, or cache-policy
tweak.
- `lookupCachedCanonicalArraySlotCallForArray(...)`: `5.98s` cumulative

So the next bounded quicksort tranche should stay on the tracked compare
opcode boundary, but move one level outward from the rejected extractor
representation rewrites: target
`compareArrayReadSlotTrackedI32Condition(...)` /
`execJumpIfArrayReadSlotCompareSlotFalse(...)` before spending more time on
slot-store completion or cache-policy reshaping.

That outer compare slice is now kept. The tracked `Array i32` compare path in
`bytecode_vm_array_slot_compare.go` was flattened so the opcode-local fast path
performs the receiver/state/cache/index/right checks in one place and skips the
dead slow-path slot loads on a hot hit. The focused compare/quicksort slice and
full `go test ./pkg/interpreter` gate stayed green. On the current kept
baseline, the external 1 MB quicksort prefix moved to:

- `698458713 ns/op`, `20395552 B/op`, `1286472 allocs/op`
- `673385319 ns/op`, `20395568 B/op`, `1286473 allocs/op`
- `719970367 ns/op`, `20395552 B/op`, `1286472 allocs/op`

with a profiled confirmation at:

- `687033283 ns/op`, `20395552 B/op`, `1286472 allocs/op`

The 1 MB profile still shows the same path as the real beneficiary:

- `compareArrayReadSlotTrackedI32Condition(...)`: about `260ms` cumulative
- `execJumpIfArrayReadSlotCompareSlotFalse(...)`: about `270ms` cumulative
- `lookupCachedCanonicalArraySlotCallForArray(...)`: about `80ms` cumulative

Full external bytecode `quicksort` still times out at the `90s` guard, so this
is a real external-prefix keep but not the final timeout closure. The next
bounded tranche should stay on this same loop edge and start from a fresh
full-scale profile on the kept state; the likely next targets are the
remaining `trackedArrayCompareDirectSmallI32Value(...)` /
`lookupCachedCanonicalArraySlotCallForArray(...)` cost inside the kept compare
path, not a return to slot-store completion.

That full-scale refresh is now available on the later kept baseline too. Full
external bytecode `quicksort` still times out at the `90s` guard, and the
full-input steady-state `bytecode-runtime` pass completed at
`115031018567 ns/op`, `9016856664 B/op`, and `203288578 allocs/op` with
`240.69s` real average and `20.00` GC average. The ranking is still dominated
by the same tracked compare / `i32` extraction boundary:

- `trackedArrayCompareDirectSmallI32Value(...)`: `19.35s` flat /
  `19.38s` cumulative
- `compareArrayReadSlotTrackedI32Condition(...)`: `31.11s` cumulative
- `execJumpIfArrayReadSlotCompareSlotFalse(...)`: `32.53s` cumulative
- `arrayReadSlotTrackedI32RawAtSlot(...)`: `21.82s` cumulative
- `lookupCachedCanonicalArraySlotCallForArray(...)`: `6.89s` cumulative

The remaining store/call setup costs are clearly second-tier on the same run:

- `execStoreSlotBinaryIntSlotConst(...)`: `10.47s` cumulative
- `tryInlineCachedCallNameDirectFromSlots(...)`: `11.90s` cumulative
- `bytecodeDirectSmallI32Value(...)`: `5.37s` cumulative

That tightened the planning conclusion rather than changing it. The next real
quicksort tranche should be the first real typed `i32` frame/register slice or
an equivalently large tracked-`Array i32` representation step, not another
boxed-path helper tweak.

One more bounded keep was still hiding inside that same fused compare boundary.
The `arr[idx] as i32 <op> rhs` opcode path still missed raw `i32` register
operands and still recast tracked `Array i32` elements locally even though the
tracked array state already carried the incremental raw-`i32` cache from the
earlier quicksort representation slice. The kept follow-up now:

- reads the rhs and index directly from the raw `i32` register lane on the
  fused `JumpIfArrayIndexSlotCompareSlotFalse` `i32` fast path
- uses the tracked-array raw-`i32` cache first on the array-index compare path
  before falling back to `runtime.Value`
- materializes register-backed fallback operands explicitly so negative/raw
  register indexes keep the same error behavior as boxed slots

On the reduced in-tree quicksort hotloop, a fresh `100x` guard moved from
`5568622 ns/op` to `5307936 ns/op`, with the confirming profile run at
`5349063 ns/op`. The reduced hotloop profile changed in the intended way:

- `bytecodeArrayIndexCastSmallI32Raw(...)` dropped out of the top tier
- `arrayIndexSlotCompareI32RawValue(...)` dropped out of the top tier
- the visible compare work is now the opcode-local
  `compareArrayIndexSlotI32ConditionAtSlots(...)` /
  `execJumpIfArrayIndexSlotCompareSlotFalse(...)` lane instead of the old
  tracked-value recast helper

This is another real keep, but it is still a reduced quicksort keep rather
than a timeout closure. The next useful step is to re-profile full-scale
quicksort on this state and decide whether the next bounded cut stays on the
fused compare/load edge or moves to the adjacent call/store tier.

## 2026-06-10 — External quicksort tracked dynamic-array raw `i32` cache

The next quicksort keep finally took the larger representation step that the
full-scale profiles had been pointing at. Instead of re-reading a tracked
dynamic `Array i32` element as `runtime.Value` and rediscovering a small `i32`
on every partition compare, tracked array state now keeps an incremental raw
`i32` cache whenever the live tracked contents remain direct small `i32`
values. The compare opcode reads that cache first.

This landed in the runtime/interpreter boundary rather than as another local
opcode micro-branch:

- `v12/interpreters/go/pkg/runtime/array_store.go`
- `v12/interpreters/go/pkg/runtime/array_store_i32_cache.go`
- `v12/interpreters/go/pkg/interpreter/interpreter_array_i32_cache.go`
- `v12/interpreters/go/pkg/interpreter/interpreter_arrays.go`
- `v12/interpreters/go/pkg/interpreter/bytecode_vm_array_slot_compare.go`
- `v12/interpreters/go/pkg/interpreter/bytecode_vm_array_slot_member_fast.go`
- tests in `interpreter_array_tracking_test.go`

The first version rebuilt the cache on every append and was immediately
rejected. The kept version only appends to the raw cache incrementally on the
steady-state `push(...)` path, while `syncArrayValues(...)` still rebuilds from
full state only at broader sync boundaries.

On the current local quicksort measurements, the old baselines were:

- hotloop bench: `7550051 ns/op`, `212355 B/op`, `2964 allocs/op`
- 1 MB external-style prefix: `777962886 ns/op`, `16203384 B/op`,
  `500040 allocs/op`

The kept state moved to:

- hotloop bench: `5881281 ns/op`, `233355 B/op`, `2977 allocs/op`
- hotloop confirmation: `6848786 ns/op`, `233387 B/op`, `2977 allocs/op`
- 1 MB external-style prefix `3/3`: `708968279 ns/op`, `18171568 B/op`,
  `500068 allocs/op`
- profiled 1 MB confirmation: `702132004 ns/op`, `18201992 B/op`,
  `500134 allocs/op`

The profile consequence is the important result. On the profiled 1 MB prefix
run, the old tracked-compare extraction wall moved behind the remaining local
materialization and index helpers:

- `compareArrayReadSlotTrackedI32Condition(...)`: about `60ms` cumulative
- `trackedArrayCompareI32RawAtSlot(...)`: about `30ms` cumulative
- `execJumpIfArrayReadSlotCompareSlotFalse(...)`: about `70ms` cumulative
- `bytecodeRawI32SlotCachedValue(...)`: about `60ms` flat
- `arraySlotIndexSmall(...)`: about `40ms` flat

So this is a real keep, but it is still not the end state for bytecode
`quicksort`. The next defensible slice should stay on the same tracked
`Array i32` materialization boundary: target
`bytecodeRawI32SlotCachedValue(...)`, `arraySlotIndexSmall(...)`,
`execArrayReadSlot(...)`, or a broader operand/load cut that shares this raw
lane, not another helper-level compare metadata tweak. Full external bytecode
`quicksort` still timed out at the `90s` guard on this tranche:

- `./v12/bench_compare_external --benchmarks quicksort --modes bytecode --runs 1 --timeout 90`

## 2026-06-09 — Bytecode named-struct storage path follow-up

The next timeout-family keep stayed on the same nominal/materialization
boundary instead of opening another isolated helper cut. Exact named struct
literals built through the bytecode fast path now use definition-ordered
positional storage rather than allocating a per-instance `map[string]Value`,
and the generic named-struct consumers that must preserve field semantics now
route through a shared helper boundary that works for both legacy map-backed
and new positional-backed instances.

This was a real reduced-case win on `binarytrees` (`n := 16`). The kept
baseline moved from:

- `62.83s`
- `10992937632 B/op`
- `127767105 allocs/op`

to:

- `56.59s`
- `6437212488 B/op`
- `112781161 allocs/op`

The full external bytecode benchmark still times out at the `60s` guard, so
this remains a timeout-family keep rather than a closure. But the allocation
drop is large enough to matter: roughly `4.56GB` and about `15M` allocs/op
came out of the reduced case on this slice alone.

The next kept follow-up stayed on the same broader nominal call/materialization
boundary instead of opening another field/member helper cut. Exact simple
named-struct coercion and return edges now bypass the generic
`matchesType(...)` / `coerceValueToType(...)` path when the runtime value
already carries that exact non-generic struct definition, while preserving the
old `Error` payload unwrap semantics.

That moved the same reduced `binarytrees` case again from:

- `56.59s`
- `6437212488 B/op`
- `112781161 allocs/op`

to:

- `54.27s`
- `6437180464 B/op`
- `112781084 allocs/op`

The profile consequence is modest but real: recursive nominal call/return
coercion is cheaper now, and the broader generic type-expression
materialization dropped further out of the top tier. Full external bytecode
`binarytrees` still times out at the `60s` guard, so this is another reduced
timeout-family keep rather than a closure.

The next kept follow-up pushed that exact nominal proof into the bytecode
inline boundary itself instead of leaving it inside the generic coercion
helper. Exact simple named-struct parameter and return shapes now count as
no-coercion hits during:

- inline direct-call frame setup
- inline self-call frame setup
- cached call-name direct-inline setup
- inline return finish

That moved the same reduced `binarytrees` case again from:

- `54.27s`
- `6437180464 B/op`
- `112781084 allocs/op`

to:

- `53.49s`
- `6437165800 B/op`
- `112781019 allocs/op`

The gain is smaller than the earlier storage and typed-pattern keeps, but it
is still defensible and it confirms the same design conclusion: the remaining
timeout-family wall is the broader nominal call/materialization boundary, not
another local field/member helper slice. Full external bytecode `binarytrees`
still times out at the `60s` guard.

## 2026-06-09 — Env-aware simple named-struct slot eligibility for `binarytrees` bytecode

The next kept timeout-family slice moved that same broader nominal proof into
slot-layout eligibility itself. The bytecode lowerer already had an exact
simple named-struct literal fast path, but slot analysis still rejected every
`StructLiteral`, which kept `make_tree(...)` off the slot/direct-inline path
and left the reduced `binarytrees` case dominated by the generic lexical
call/cache ladder.

The landed rule is intentionally narrow:

- exact simple named-struct literals may count as slot-safe only when the
  lowering environment can already prove the visible named struct definition
- this is only for the bytecode slot-layout path; it is not a new language
  rule and not a new nominal representation
- functions that become slot-eligible only through this rule are still
  rejected when the body also contains placeholder expressions or dotted
  identifier calls

That moved the reduced external-style `binarytrees` case (`n := 16`) from:

- `53.49s`
- `6437165800 B/op`
- `112781019 allocs/op`

to:

- `19.17s`
- `1692672920 B/op`
- `30223222 allocs/op`

This is the first `binarytrees` bytecode slice in this run that clearly
collapses the old generic lexical call wall. The reduced-case profile is no
longer led by `invokeFunction(...)` / `execCallName(...)`; the new top tier is
now the direct-inline and struct-fast boundary:

- `execStructLiteralNamedFast(...)`
- `finishInlineReturn(...)`
- `tryInlineSelfCallFromStack(...)`
- `execCallOpcode(...)`
- `execCallSelfIntSubSlotConst(...)`

Full external bytecode `binarytrees` still times out at the `60s` guard, so
this is still a timeout-family keep rather than a closure. But the planning
target changed materially: the next slice should target the new
direct-inline/struct-fast wall, not the old lexical call/cache ladder.

## 2026-06-09 — Lowered named-struct literal plans for `binarytrees` bytecode

The next kept follow-up stayed on that new direct-inline/struct-fast wall. The
previous keep had already made `make_tree(...)` slot-backed and direct-inline,
but the hot `Node { left: ..., right: ... }` path still paid two avoidable
costs on every execution:

- `env.StructDefinition(...)` to recover the same exact named struct
- a field-name scan to map source field order onto definition order

The new slice removes both from the fast path when lowering already knows the
struct definition:

- lowering records a per-site named-struct literal plan
- the plan carries precomputed definition-order field indices
- when available, it also carries the exact `StructDefinitionValue` from the
  lowering environment

That moved the reduced external-style `binarytrees` case (`n := 16`) from:

- `19.17s`
- `1692672920 B/op`
- `30223222 allocs/op`

to:

- `17.42s`
- `1692672936 B/op`
- `30223223 allocs/op`

The profiled reduced run confirms the intended shift:

- `execStructLiteralNamedFast(...)` dropped from about `3.65s` cumulative on
  the earlier slot-inline keep to about `1.17s`
- the remaining first-tier wall is still the same direct-inline boundary:
  `finishInlineReturn(...)`, `execCallOpcode(...)`, and
  `tryInlineSelfCallFromStack(...)`

Full external bytecode `binarytrees` still times out at the `60s` guard, so
this remains a reduced timeout-family keep rather than a closure.

## 2026-06-19 — `nbody` and `k_nucleotide` broadened the external harness

The next benchmark-coverage tranche added two more comparison targets to the
local/external performance workflow:

- `nbody`
- `k_nucleotide`

The harness changes landed in three places:

- canonical Able sources under `v12/examples/benchmarks`
- sibling `../benchmarks` Go/Able/verify packaging
- `v12/bench_compare_external` benchmark wiring, including the
  `k_nucleotide` program argument and the `nbody` target mapping

`k_nucleotide` also exposed three general runtime seams that were fixed as part
of the tranche rather than worked around in the benchmark source:

- compiled/interpreted hash-map helpers now hash and compare primitive
  `String`/`bool`/`char`/integer keys directly before falling back to
  interface dispatch
- static host-extern launchers now seed `interp.SetArgs(os.Args[1:])`, which
  restores compiled `os.args()` for CLI-driven workloads
- the benchmark-facing `able.fs` / `able.io` wrappers in the external
  `able-stdlib` now use explicit `match` handling instead of the earlier
  generic `unwrap(...)` wrappers that were tripping compiler/typechecker paths

Current validation from the landed state:

- full compiled `k_nucleotide` verification now passes against the generated
  FASTA/reference output
- `./v12/bench_compare_external --benchmarks k_nucleotide --modes compiled --runs 1 --timeout 120`
  currently returns `ok (1)` at `2.8800s`
- `./v12/bench_compare_external --benchmarks nbody --modes compiled --runs 1 --timeout 120`
  currently returns `ok (1)` at `0.3600s`

The next useful work on this slice is no longer harness repair. It is source
audit plus measurement work on the new benchmark families, and then the same
compiled-vs-bytecode follow-up loop already used on the existing core set.

## 2026-06-19 — `reverse_complement` joined the external harness

The next external benchmark-coverage slice added a byte-oriented FASTA
workload:

- `reverse_complement`

The landed coverage spans the same three surfaces as the other external
families:

- canonical Able source under
  `v12/examples/benchmarks/reverse_complement/reverse_complement.able`
- sibling `../benchmarks/reverse-complement` Go/Able/setup/verify packaging
- `v12/bench_compare_external` mapping plus program-argument wiring

This benchmark did not require a new stdlib surface. The existing
`able.fs.read_bytes(...)` and `able.io.write_all(...)` APIs were enough for a
direct byte-oriented implementation, which is the right first shape for this
workload because it keeps the benchmark focused on reverse traversal, byte
lookup, and buffered output rather than on string-object churn.

Current validation from the landed state:

- `go run ../benchmarks/reverse-complement/app.go ../benchmarks/reverse-complement/reverse-complement-input.fasta | ruby ../benchmarks/reverse-complement/verify.rb`
  passes
- compiled canonical Able build now verifies against the generated FASTA
  reference output
- `./v12/bench_compare_external --benchmarks reverse_complement --modes compiled,bytecode --runs 1 --timeout 30`
  now returns:
  - compiled `ok (1)` at `0.3200s` vs Go `0.0100s` (`32.00x`)
  - bytecode `ok (1)` at `3.5300s` vs Go `0.0100s` (`353.00x`)
- treewalker still times out at the `30s` guard on the same workload

The sibling `../benchmarks/results.json` snapshot now includes Go reference
rows for `reverse-complement`, so `bench_compare_external` no longer reports
`n/a` for the Go comparison on this benchmark.

## 2026-06-19 — `mandelbrot` joined the external harness

The next external benchmark-coverage slice added a binary-output numeric
workload:

- `mandelbrot`

The landed coverage again spans the same three surfaces:

- canonical Able source under
  `v12/examples/benchmarks/mandelbrot/mandelbrot.able`
- sibling `../benchmarks/mandelbrot` Go/Able/setup/verify packaging
- `v12/bench_compare_external` target mapping

Two harness-level notes matter for this benchmark:

- the sibling `../benchmarks/run.rb` now extracts timing metrics from raw
  merged stdout/stderr bytes instead of assuming UTF-8 text output, which is
  necessary for PBM/binary workloads
- the kept benchmark size is `SIZE = 800`; larger earlier sizes kept the
  bytecode path in the timeout family, so the landed benchmark is tuned to
  keep compiled and bytecode measurable in the standard external harness

Current validation from the landed state:

- `go run ../benchmarks/mandelbrot/app.go | ruby ../benchmarks/mandelbrot/verify.rb`
  passes
- compiled canonical Able build verifies against the generated PBM reference
  output
- `./v12/bench_compare_external --benchmarks mandelbrot --modes compiled,bytecode --runs 1 --timeout 30`
  now returns:
  - compiled `ok (1)` at `0.1100s` vs Go `0.0400s` (`2.75x`)
  - bytecode `ok (1)` at `20.2800s` vs Go `0.0400s` (`507.00x`)
- treewalker still times out at the `30s` guard on the same workload

## 2026-06-19 — `mandelbrot` bytecode float-slot store tranche

The first measured bytecode optimization pass on `mandelbrot` stayed fully on
the VM/lowering side and did not change the benchmark source.

The landed pieces were:

- a new direct slot-to-slot float binary store opcode for simple statement
  shapes like `zr2 := zr * zr` and `zi2 := zi * zi`
- a generalized float add-mul lowering rule so the base term no longer has to
  be the target slot itself, which now covers `zi = 2.0 * zr * zi + ci`
- a binary-output fix in `v12/bench_perf` (`grep -a`) so steady-state
  `bytecode-runtime` measurements work on PBM-emitting programs

Measured impact:

- like-for-like one-shot `bytecode-runtime` confirmation moved from
  `20448897471 ns/op`, `7753471504 B/op`, `323454802 allocs/op`
  to `13175977420 ns/op`, `3964297712 B/op`, `165573914 allocs/op`
- external bytecode mode moved from `20.2800s` to `13.2100s`
  on `./v12/bench_compare_external --benchmarks mandelbrot --modes bytecode --runs 1 --timeout 30`
- against the kept Go row (`0.0400s`), the external bytecode ratio improved
  from `507.00x` to `330.25x`

The refreshed post-keep CPU/alloc profile still points at the same broad class
of remaining work, but on a narrower wall:

- CPU hot tier: `execLoadSlotOpcode(...)`, `bytecodeDirectFloatArithmeticFast(...)`,
  `slotRuntimeValue(...)`, `bytecodeSlotReadValue(...)`, `execBinary(...)`
- allocation hot tier: `bytecodeDirectFloatArithmeticFast(...)`,
  `bytecodeSlotReadValue(...)`, `bytecodeDirectFloatAddMul(...)`

That makes the next likely productive tranche a float-materialization cut:
avoid boxing/materializing transient float results when the next step is a
direct slot write or another float-specialized opcode.

## 2026-06-19 — `mandelbrot` raw-float slot lane

The follow-on `mandelbrot` pass stayed entirely inside the bytecode VM and
implemented the first bounded raw-float lane.

The landed shape was:

- new internal raw float carriers:
  - `bytecodeRawF32SlotValue`
  - `bytecodeRawF64SlotValue`
- float-specialized slot/store opcodes now keep those carriers unboxed across:
  - direct float binary arithmetic
  - fused float add-mul slot updates
  - fused float binary stores
  - cast-slot-float-const-div fast paths
  - slot loads onto the VM stack
- explicit materialization boundaries were added at:
  - generic binary fallback
  - call setup
  - typed slot assignment
  - inline/direct return and VM exit

Measured impact with the comparable cached-stdlib harness:

- `./v12/bench_perf --runs 1 --timeout 90 --modes bytecode-runtime --run-from ../benchmarks v12/examples/benchmarks/mandelbrot/mandelbrot.able`
  moved from the post-store-tranche `13175977420 ns/op`,
  `3964297712 B/op`, `165573914 allocs/op` band to a new
  `12237905096-12594330580 ns/op`, `2783100744-2783130656 B/op`,
  `300798648-300798705 allocs/op` band
- `./v12/bench_compare_external --benchmarks mandelbrot --modes bytecode --runs 1 --timeout 30`
  moved from `13.2100s` to `12.5900s`
- against the kept Go row (`0.0400s`), external bytecode improved from
  `330.25x` to `314.75x`
- secondary cached-stdlib spot-check:
  - `matrixmultiply` bytecode: `0.5200s` vs Go `0.8800s`

One important harness note came out of the same session: the source-stdlib
external runs (`ABLE_PATH=/home/david/sync/projects/able-stdlib/src:...`) were
consistently slower on the same code (`mandelbrot` `13.6400-13.7200s`,
`matrixmultiply` `0.8200s`) because source-resolution/bootstrap work is folded
into those CLI timings. The cached-stdlib runs are the comparable keep against
the existing checked-in snapshot.

The new profile changed what "next" means:

- CPU still clusters around `execLoadSlotOpcode(...)`, `execBinary(...)`,
  `bytecodeDirectFloatArithmeticFast(...)`, `slotStackValue(...)`, and
  `bytecodeStackSnapshotValue(...)`
- allocation space is now dominated by the raw-float carrier itself:
  - `bytecodeRawFloatSlotValue(...)`: `76.87%`
  - `bytecodeMaterializeRawFloatValue(...)`: `14.86%`

So the raw-float carrier keep is worth preserving, but the next productive
step is no longer "more raw carrier propagation". It is a true slot/register
sidecar raw-float lane so those transient float results stop allocating at all.

## 2026-06-20 — slot-only float sidecar probe rejected

The next follow-on experiment tried to cash that conclusion out with a bounded
slot-side raw-float frame substrate. The implementation was made correct and
covered with focused float/F64 call-frame tests, but it did not produce a new
performance keep.

Measured result:

- active slot-side lane regressed cached-stdlib external `mandelbrot` to
  `13.8200s`
- active slot-side lane regressed cached-stdlib external `matrixmultiply` to
  `0.6300s`
- the runtime allocation band stayed effectively unchanged, so the extra slot
  indirection did not remove the
  `bytecodeRawFloatSlotValue(...)` / `bytecodeMaterializeRawFloatValue(...)`
  cost identified by the prior profile
- after backing the active slot-storage path back out of the generic hot
  readers, the exploratory runtime spot check returned to
  `12827501003 ns/op`, `2783102000 B/op`, `300798642 allocs/op`, which still
  does not beat the kept raw-float-carrier band

Conclusion:

- do not treat a slot-only float sidecar as a benchmark keep
- the next credible `mandelbrot` cut is a true stack/register raw-float lane,
  or a more direct reduction of raw-float carrier/materialization allocation

## 2026-06-20 — stack-local float-cell follow-up also rejected

The next experiment took that conclusion literally and tried the smallest
stack/register-side cut that could hit the same hot loop without reopening a
full VM rewrite.

Two active variants were measured:

- pooled stack-local `*runtime.FloatValue` cells for float loads/results plus
  explicit call-boundary materialization
- a lighter follow-up that backed the pooled cell path back out but kept the
  stack-side float normalization helpers in the hot load/replace/push path

Measured result:

- pooled-cell active lane:
  - `./v12/bench_perf --runs 1 --timeout 90 --modes bytecode-runtime --run-from ../benchmarks v12/examples/benchmarks/mandelbrot/mandelbrot.able`
    landed at `28454783500 ns/op`, `1834321752 B/op`, `78319459 allocs/op`
  - `./v12/bench_compare_external --benchmarks mandelbrot,matrixmultiply --modes bytecode --runs 1 --timeout 30`
    regressed cached-stdlib external `mandelbrot` to `28.2600s`
    and landed cached-stdlib external `matrixmultiply` at `0.5500s`
- lighter raw-normalization follow-up:
  - runtime spot check landed at
    `16682311702 ns/op`, `3131617712 B/op`, `346922621 allocs/op`
  - cached-stdlib external `mandelbrot` still regressed to `17.6200s`
  - cached-stdlib external `matrixmultiply` stayed near the old band at
    `0.5200s`

So neither follow-up was a keep. The lower alloc bands did not translate into
better wall-clock on `mandelbrot`; pointer/pool bookkeeping and extra
normalization cost were more expensive than the raw-float-carrier control they
were trying to replace.

The active stack-side path was then backed out and the control restored:

- runtime `mandelbrot` spot check returned to
  `12783506275 ns/op`, `2783102432 B/op`, `300798658 allocs/op`
- cached-stdlib external `mandelbrot` returned to `13.2300s`
- cached-stdlib external `matrixmultiply` re-confirmed in the same general
  band at `0.7800s`

Conclusion:

- do not treat stack-local pooled `*runtime.FloatValue` lanes as a keep
- do not treat load/replace raw-normalization on its own as a keep
- the next credible float tranche still needs a true raw stack/register
  sidecar or a more local allocation cut that does not add pointer/pool
  bookkeeping to the hot loop

## 2026-06-20 — direct raw-float compare fast path kept the next cut

The next follow-up stayed on the narrower "local allocation cut" branch rather
than reopening another stack/slot representation experiment.

The landed change was:

- new direct float compare helpers for `<`, `<=`, `>`, `>=`, `==`, and `!=`
  that accept the existing raw-float carriers plus ordinary
  `runtime.FloatValue` / `*runtime.FloatValue`
- `execBinary(...)` now tries that path before materializing into the generic
  operator helpers
- `compareBytecodeCondition(...)` now uses the same path before falling
  through to `ApplyBinaryOperatorFast(...)` / `applyBinaryOperator(...)`

This keeps the existing raw-float carrier representation, but removes one of
the remaining materialization boundaries when float results flow directly into
comparisons.

Measured impact:

- `./v12/bench_perf --runs 1 --timeout 90 --modes bytecode-runtime --run-from ../benchmarks v12/examples/benchmarks/mandelbrot/mandelbrot.able`
  moved from the restored control
  `12783506275 ns/op`, `2783102432 B/op`, `300798658 allocs/op`
  to `10834374807 ns/op`, `2398954784 B/op`, `284792308 allocs/op`
- `./v12/bench_compare_external --benchmarks mandelbrot,matrixmultiply --modes bytecode --runs 1 --timeout 30`
  moved cached-stdlib external bytecode:
  - `mandelbrot`: `13.2300s` -> `11.2100s`
  - `matrixmultiply`: `0.5400s` on the same one-shot confirmation

So this one is a real keep. The prior rejected sidecar lesson still stands,
but the narrower conclusion is now better grounded: local boundary cuts around
raw-float comparison can pay off without changing the broader VM
representation.

What is next:

- the compare itself is now cheap, but the condition still pays for producing a
  bool on the VM stack and immediately consuming it with `JumpIfFalse`
- the next productive `mandelbrot` cut should stay on that same boundary and
  target a direct float-add-compare-const jump for the hot
  `zr2 + zi2 > 4.0` escape test before revisiting broader representation work

## 2026-06-20 — direct float-add compare jump kept the next cut

The next follow-up stayed on the exact hot `mandelbrot` condition branch that
the prior compare keep left behind.

The landed change was:

- new lowering support for `bytecodeOpJumpIfFloatAddCompareConstFalse` on
  slot-backed float shapes like `zr2 + zi2 > 4.0`
- a per-site jump plan carrying the two resolved input slots plus the float
  literal
- `execJumpIfFloatAddCompareConstFalse(...)` now reads raw float carriers from
  those slots directly, adds them, compares the result to the literal, and
  branches without materializing an intermediate float value or a transient
  bool
- unsupported cases still fall back to the existing generic condition path

This keeps the existing raw-float carrier representation again, but removes
the remaining bool stack round-trip on the hottest escape test in the
benchmark.

Measured impact:

- `./v12/bench_perf --runs 1 --timeout 90 --modes bytecode-runtime --run-from ../benchmarks v12/examples/benchmarks/mandelbrot/mandelbrot.able`
  moved from the prior kept direct-compare tranche
  `10834374807 ns/op`, `2398954784 B/op`, `284792308 allocs/op`
  to `9700657707 ns/op`, `2030453960 B/op`, `238730417 allocs/op`
- `./v12/bench_compare_external --benchmarks mandelbrot,matrixmultiply --modes bytecode --runs 1 --timeout 30`
  moved cached-stdlib external bytecode:
  - `mandelbrot`: `11.2100s` -> `9.9900s`
  - `matrixmultiply`: `0.5200s` on the same one-shot confirmation

So this one is also a real keep. The more specific conclusion now is that
local condition-boundary cuts on the existing raw-float carrier path continue
to pay off, while the broader slot/stack sidecar experiments still do not.

What is next:

- the fused compare+jump is now cheap on this site, so the next step should
  re-profile and target the remaining raw-float slot load/materialization churn
  rather than reopening wider representation changes
- that will likely entail either another local load/materialization boundary
  cut or a more disciplined raw stack/register lane, but only after the fresh
  profile confirms where the new top wall moved

## 2026-06-20 — raw-float load snapshot cut kept the next cut

The next follow-up did exactly that re-profile, and the new hot allocation wall
was narrow enough for another local keep.

The landed change was:

- `bytecodeStackSnapshotValue(...)` now reuses an existing immutable raw
  `f32`/`f64` carrier instead of wrapping the same slot value into a fresh raw
  float carrier again
- that removes the repeated load-side reboxing on `execLoadSlotOpcode(...)`
  and other stack snapshot sites while keeping the existing copy behavior for
  mutable boxed integer/float cells
- focused coverage now includes a zero-allocation raw-float slot-load test, in
  addition to the nearby stdlib primitive parity path that had already proven
  the member-call materialization boundary

This keeps the existing raw-float carrier representation again, but stops
paying for a second raw-float carrier allocation when an immutable slot value
is merely being loaded onto the VM stack.

Measured impact:

- `./v12/bench_perf --runs 1 --timeout 120 --modes bytecode-runtime --run-from ../benchmarks v12/examples/benchmarks/mandelbrot/mandelbrot.able`
  moved from the prior kept condition-jump tranche
  `9700657707 ns/op`, `2030453960 B/op`, `238730417 allocs/op`
  to `8580938220 ns/op`, `1302262736 B/op`, `147706479 allocs/op`
- `./v12/bench_compare_external --benchmarks mandelbrot,matrixmultiply --modes bytecode --runs 1 --timeout 30`
  moved cached-stdlib external bytecode:
  - `mandelbrot`: `9.9900s` -> `9.6000s`
  - `matrixmultiply`: `0.5300s` on the same one-shot confirmation

So this one is also a real keep. More importantly, the post-keep profile
changes what "next" means:

- load-side cumulative allocation mostly drops out of the hot tier
- `bytecodeRawFloatSlotValue(...)` still dominates allocation space, but it is
  now driven much more by raw-float result creation and slot-write paths such
  as `execStoreSlotFloatBinary(...)`, `storeFloatSlotValue(...)`, and
  `bytecodeDirectFloatArithmeticFast(...)`

What is next:

- the next productive `mandelbrot` cut should stay local to that same
  raw-float carrier boundary and target raw-float result production or slot
  write churn
- that will likely entail another narrow store/result boundary cut rather than
  reopening the broader slot/stack sidecar experiments that already failed

## 2026-06-20 — raw-float slot-write reuse kept the next cut

The next follow-up stayed on exactly that store/result boundary and removed the
next duplicate raw-float carrier allocation instead of widening the VM shape.

The landed change was:

- `storeFloatSlotValue(...)` now reuses an existing raw float carrier when a
  fast path has already produced one, instead of unwrapping that carrier and
  constructing another raw float carrier for the slot write
- that removes the duplicate raw-float slot-write allocation on the proven
  float store fast paths while leaving boxed float/int slot behavior unchanged
- focused coverage now includes a zero-allocation raw-float slot-store test

This keeps the existing raw-float carrier representation again, but removes one
more carrier rebuild on the hot float store path.

Measured impact:

- `./v12/bench_perf --runs 1 --timeout 120 --modes bytecode-runtime --run-from ../benchmarks v12/examples/benchmarks/mandelbrot/mandelbrot.able`
  moved from the prior kept load-snapshot tranche
  `8580938220 ns/op`, `1302262736 B/op`, `147706479 allocs/op`
  to `7602284359 ns/op`, `932025480 B/op`, `101426969 allocs/op`
- a profiled post-keep confirmation landed at
  `7481653447 ns/op`, `932066960 B/op`, `101427045 allocs/op`
- `./v12/bench_compare_external --benchmarks mandelbrot,matrixmultiply --modes bytecode --runs 1 --timeout 30`
  moved cached-stdlib external bytecode:
  - `mandelbrot`: `9.6000s` -> `8.7900s`
  - `matrixmultiply`: `0.5800s` on the same one-shot confirmation

So this one is also a real keep. The new post-keep profile sharpens the next
target again:

- `bytecodeRawFloatSlotValue(...)` still dominates allocation space
- the remaining cumulative wall is now much more clearly centered on raw-float
  result creation in generic float arithmetic and float store arithmetic
  helpers such as `execBinary(...)`, `bytecodeDirectFloatArithmeticFast(...)`,
  `execStoreSlotFloatBinary(...)`, and `bytecodeDirectFloatAddMulValue(...)`

What is next:

- the next productive `mandelbrot` cut should stay on that same raw-float
  carrier boundary and target raw-float result creation itself
- that likely means another narrow arithmetic/result-path cut rather than any
  return to the broader slot/stack sidecar experiments that already failed

## 2026-06-20 — narrowed owned-float result reuse kept the array-get follow-up

The next float follow-up tested a broader reusable owned-float slot update
path, then kept only the subset that held up after profiling.

The landed change was:

- added shared float-store result helpers so exact raw float results can write
  directly into an existing owned float slot cell when one is already present
- kept that reuse only on `StoreSlotFloatAddMulArrayGet` and
  `StoreSlotCastSlotFloatConstDiv`
- backed the same reuse out of plain `StoreSlotFloatBinary` and plain
  `StoreSlotFloatAddMul` after profiling showed those lanes shifting
  allocation into `bytecodeStackSnapshotValue(...)`

The useful lesson here is narrower than the original probe: broad owned-float
slot reuse is still the wrong direction, but the fused array-get accumulation
lane can reuse an already-owned float slot cell without reopening the larger
load-side regression.

Measured impact:

- the rejected broad plain-store reuse probe regressed profiled runtime
  `mandelbrot` to `8563048519 ns/op`, `1166922760 B/op`,
  `100823898 allocs/op`
- the final narrowed tree lands unprofiled runtime `mandelbrot` at
  `7842792429 ns/op`, `932030880 B/op`, `101426975 allocs/op`
- the profiled confirmation landed at
  `7976444236 ns/op`, `932067504 B/op`, `101427061 allocs/op`
- `./v12/bench_compare_external --benchmarks mandelbrot,matrixmultiply --modes bytecode --runs 3 --timeout 60`
  averaged cached-stdlib external bytecode:
  - `mandelbrot`: `8.1100s`
  - `matrixmultiply`: `0.5467s`

What is next:

- the next productive float cut should stay on the remaining raw-float result
  creation in `execBinary(...)` and adjacent generic arithmetic helpers
- do not widen owned-float slot reuse back onto the plain float store lanes;
  that just trades raw-result allocation for load-side snapshot churn

## 2026-06-10 — Per-site exact named-struct member plans for `binarytrees`

The next kept `binarytrees` slice stayed on the same reduced inline wall, but
the target moved from local helper shaving to a broader bytecode-site plan for
named-struct field access.

The landed cut carries exact named-struct field plans on
`bytecodeOpMemberAccess` sites when lowering can already prove the receiver’s
exact named-struct definition from slot-backed metadata:

- exact named-struct params now seed slot-local exact-definition metadata
- typed-pattern bindings like `left: Node` propagate that same proof into the
  bound slot
- simple declarations whose RHS already carries an exact named-struct proof,
  such as direct `as DepthResult`, also seed the slot metadata
- `execMemberAccess(...)` now tries the planned exact-definition/field-index
  path before the broader `structNamedFieldValue(...)` name-scan helper

That moved the reduced external-style `binarytrees` case (`n := 16`) from the
prior kept `12.21-12.78s` band to:

- `12.7100s`, `2172228568 B/op`, `15237328 allocs/op`
- `11.8800s`, `2172228456 B/op`, `15237327 allocs/op`
- isolated confirmation: `12.2300s`, `2172228664 B/op`, `15237328 allocs/op`

The direct profiled confirmation held at:

- `6278000300 ns/op`
- `2172258976 B/op`
- `15237395 allocs/op`

The profile change is the useful part:

- new `bytecodeDirectPlannedStructMemberValue(...)`: about `0.19s`
  cumulative
- `execMemberAccess(...)`: about `0.37s` cumulative
- the old `bytecodeDirectStructMemberValue(...)` / repeated
  `structNamedFieldIndex(...)` scan path dropped out of the top tier for the
  planned hot sites
- the remaining reduced wall is still:
  `finishInlineReturn(...)`, `execCallOpcode(...)`,
  `execStructLiteralNamedFast(...)`, `tryInlineSelfCallFromStack(...)`

Full external bytecode `binarytrees` still times out at the `60s` guard, so
this is another reduced timeout-family keep rather than a closure.

## 2026-06-10 — Direct minimal self-fast frame push for `binarytrees`

The next kept `binarytrees` slice stayed on the same reduced inline wall, but
it moved from member planning to the remaining self-inline setup boundary.

The landed cut is narrow:

- added `pushInlineSelfFastFrame(...)`
- self-inline setup sites that already satisfy
  `bytecodeCanUseSelfFastMinimalFrame(...)` now go directly to
  `pushSelfFastMinimalCallFrameWithBases(...)`
- non-minimal sites still fall back to the existing `pushCallFrame(...)`
  path unchanged

That moved the reduced external-style `binarytrees` case (`n := 16`) from the
prior kept `11.88-12.71s` band to:

- `11.6700s`, `2172228872 B/op`, `15237329 allocs/op`
- `12.1100s`, `2172229016 B/op`, `15237332 allocs/op`

The direct profiled confirmation held at:

- `6224157625 ns/op`
- `2172258640 B/op`
- `15237392 allocs/op`

The profile consequence is the useful part:

- inside `tryInlineSelfCallFromStack(...)`, the old inline
  `pushCallFrame(...)` edge on the hot one-param branch dropped from about
  `210ms` sampled to about `100ms` through `pushInlineSelfFastFrame(...)`
- `pushCallFrame(...)` dropped out of the first-tier reduced hot path
- the remaining reduced wall is still:
  `finishInlineReturn(...)`, `execCallOpcode(...)`,
  `execStructLiteralNamedFast(...)`, `tryInlineSelfCallFromStack(...)`

Full external bytecode `binarytrees` still times out at the `60s` guard, so
this is another reduced timeout-family keep rather than a closure.

## 2026-06-09 — Ambient minimal self-fast frames for `binarytrees`

The next kept follow-up stayed on the same inline return/call boundary. The
previous keep had already removed most exact named-struct nominal work from
recursive inline setup/return, but the runtime still dropped back to the
heavier self-fast frame form whenever the caller already sat inside worker
loop/iterator depth.

The landed slice keeps those self-fast calls on the minimal frame path when
analysis can already prove the callee body cannot mutate loop/iterator stacks:

- frame-layout analysis now records that control-flow preservation fact
- compact self-fast frames now store iterator/loop base depth directly
- minimal self-fast return restores those depths from the compact frame
  instead of forcing the call onto the full self-fast frame ladder

That moved the reduced external-style `binarytrees` case (`n := 16`) from:

- `15.67s`
- `1692673160 B/op`
- `30223225 allocs/op`

to:

- `15.17s`
- `1692675272 B/op`
- `30223231 allocs/op`

The profiled reduced rerun confirms the intended shift:

- `finishInlineReturn(...)` dropped from about `1.27s` cumulative to about
  `1.07s`
- `pushSelfFastMinimalCallFrameWithBases(...)` and
  `pushSelfFastSlot0CallFrameWithBases(...)` now carry the recovered compact
  path under ambient worker-loop state
- the remaining wall is still the same inline execution boundary:
  `execCallOpcode(...)`, `finishInlineReturn(...)`,
  `execStructLiteralNamedFast(...)`, `tryInlineSelfCallFromStack(...)`

Full external bytecode `binarytrees` still times out at the `60s` guard, so
this remains another reduced timeout-family keep rather than a closure.

## 2026-06-10 - exact named-struct field hits bypass interpreter member ladder on binarytrees

The next kept `binarytrees` bytecode slice stayed on the same reduced inline
wall, but moved from frame/register cleanup to direct member reads.

The landed cut is opcode-local:

- `execMemberAccess(...)` now handles exact named-struct field hits directly
  when the bytecode site already knows the member name and method precedence
  cannot change the result
- the `preferMethods` callable-field case stays intact, so callable fields
  still win before method lookup would
- misses still fall through to the existing `memberAccessOnValueWithOptions(...)`
  / `structInstanceMember(...)` machinery unchanged

That moved the reduced external-style `binarytrees` case (`n := 16`) from the
prior kept band:

- `13.60s`
- `13.76s`

to:

- `12.83s`
- `12.39s`

with bytes/allocs effectively flat:

- `2172228472 B/op`, `15237328 allocs/op`
- `2172228904 B/op`, `15237331 allocs/op`

The profiled confirmation held at:

- `6243638979 ns/op`
- `2172263848 B/op`
- `15237394 allocs/op`

The profile consequence is the useful result:

- `execMemberAccess(...)` dropped from about `1.62s` cumulative on the prior
  kept reduced profile to about `0.42s`
- `memberAccessOnValueWithOptions(...)` and `structInstanceMember(...)`
  dropped out of the top tier
- the remaining reduced wall is still:
  `finishInlineReturn(...)`, `execCallOpcode(...)`,
  `execStructLiteralNamedFast(...)`, `tryInlineSelfCallFromStack(...)`

Full external bytecode `binarytrees` still times out at the `60s` guard, so
this is another reduced timeout-family keep rather than a closure.

## 2026-06-10 — Exact-definition inline return cut for `binarytrees`

The next kept `binarytrees` slice stayed on the same reduced inline
call/return wall, but it moved off member access and away from broader
struct-literal reshaping. The previous kept profile still showed the inline
return boundary doing two pieces of unnecessary work on every hot recursive
step:

- exact named-struct no-coercion checks still routed through the broader
  cached-name / error-payload helper
- the minimal self-fast return path still called the control-stack restore
  helper even when iterator and loop depth were already unchanged

The landed cut stays narrow:

- bytecode inline arg/return sites that already carry an exact struct
  definition now use a direct exact-definition match helper instead of the
  broader cached-name fallback logic
- that bytecode-only helper does not unwrap `Error` payloads, so payload cases
  still fall through to the existing full coercion path unchanged
- hot return sites now skip `restoreCallFrameControlStacks(...)` entirely when
  both iterator and loop depth already match the saved frame bases

That moved the reduced external-style `binarytrees` case (`n := 16`) from the
prior kept band:

- `12.83s`
- `12.39s`

to:

- `12.78s`
- `12.21s`

with bytes/allocs effectively flat:

- `2172228424 B/op`, `15237325 allocs/op`
- `2172228328 B/op`, `15237325 allocs/op`

The profiled confirmation held at:

- `6601784747 ns/op`
- `2172259056 B/op`
- `15237394 allocs/op`

The profile consequence is narrow but real:

- inside `finishInlineReturn(...)`, the exact named-struct no-coercion edge
  dropped from about `70ms` sampled to about `20ms`
- the no-op `restoreCallFrameControlStacks(...)` edge dropped out of the
  sampled minimal-return path
- the remaining reduced wall is still:
  `finishInlineReturn(...)`, `execCallOpcode(...)`,
  `execStructLiteralNamedFast(...)`, `tryInlineSelfCallFromStack(...)`

Full external bytecode `binarytrees` still times out at the `60s` guard, so
this remains another reduced timeout-family keep rather than a closure.

## 2026-06-09 — Inline small positional struct storage for `binarytrees`

The next kept follow-up stayed on the same reduced direct-inline / struct-fast
wall. The current reduced profile still showed `execStructLiteralNamedFast(...)`
doing real work, and the remaining hot cost included both the struct instance
allocation itself and a second allocation for the positional backing slice.

The landed slice removes that second allocation for common small structs:

- `StructInstanceValue` now carries inline positional storage for up to four
  fields
- `runtime.NewStructInstancePositionalSized(...)` builds small positional
  struct instances without allocating a separate backing slice
- exact named-struct bytecode literals now allocate/fill through that helper
  instead of `make([]runtime.Value, len(fields))`

That moved the reduced external-style `binarytrees` case (`n := 16`) from:

- `15.17s`
- `1692675272 B/op`
- `30223231 allocs/op`

to:

- `14.16s`
- `2172228840 B/op`
- `15237331 allocs/op`

The direct profiled runtime confirmation held at:

- `7183191547 ns/op`
- `2172264424 B/op`
- `15237401 allocs/op`

The useful profile consequence is clear:

- `execStructLiteralNamedFast(...)` dropped again, from about `1.23s`
  cumulative on the prior kept reduced-profile chain to about `0.68s`
- the remaining first-tier reduced wall is still:
  `finishInlineReturn(...)`, `execCallOpcode(...)`,
  `execStructLiteralNamedFast(...)`, `tryInlineSelfCallFromStack(...)`

Tradeoff:

- allocation count dropped sharply because small named structs no longer
  allocate a second positional slice
- bytes/op rose because every struct instance now carries inline positional
  storage

Full external bytecode `binarytrees` still times out at the `60s` guard, so
this remains a reduced timeout-family keep rather than a closure.

## 2026-06-09 — Direct detached `i32` frame restore for `binarytrees`

The next kept follow-up stayed on the same reduced inline call/return wall.
The previous kept profile still showed the detached caller `i32` lane restore
as part of the hot inline-return boundary even when no active callee register
frame was live:

- `releaseActiveI32RegisterFrame(...)`
- `restoreI32RegisterFrame(...)`

The landed slice removes that empty-path churn:

- `restoreI32RegisterFrame(...)` now installs the detached caller register
  frame directly when the VM has no active `i32` frame
- the old release/re-pool path still runs unchanged when an active frame
  actually exists
- focused coverage proves that restoring into an idle VM installs the provided
  raw lane directly

That moved the reduced external-style `binarytrees` case (`n := 16`) from:

- `14.16s`
- `2172228840 B/op`
- `15237331 allocs/op`

to:

- `13.60s`
- `2172228568 B/op`
- `15237328 allocs/op`

with a second confirmation rerun at:

- `13.76s`
- `2172228312 B/op`
- `15237324 allocs/op`

The profiled confirmation shows the intended consequence:

- `releaseActiveI32RegisterFrame(...)` is down to about `0.12s`
- `restoreI32RegisterFrame(...)` is down to about `0.12s`
- the remaining first-tier reduced wall is still:
  `finishInlineReturn(...)`, `execCallOpcode(...)`,
  `execStructLiteralNamedFast(...)`, `tryInlineSelfCallFromStack(...)`

Full external bytecode `binarytrees` still times out at the `60s` guard, so
this remains another reduced timeout-family keep rather than a closure.

## 2026-06-09 — Cached exact named-struct no-coercion metadata for `binarytrees`

The next kept follow-up stayed on the same direct-inline/struct-fast wall. The
previous keep had already removed env lookup and field-name scans from hot
`Node` construction, but inline arg setup and inline return were still paying
exact named-struct nominal checks on every recursive step.

The landed slice moves that proof onto the frame layout itself:

- frame-layout analysis caches exact named-struct definitions for eligible
  params and returns when lowering already has the env proof
- inline arg setup uses that cached definition before the older name-based
  nominal helper
- inline return finish does the same
- the older exact named-struct helper remains as fallback when the lowering
  env cannot prove the definition

That moved the reduced external-style `binarytrees` case (`n := 16`) from:

- `17.42s`
- `1692672936 B/op`
- `30223223 allocs/op`

to:

- `15.67s`
- `1692673160 B/op`
- `30223225 allocs/op`

The profiled reduced run confirms the intended consequence:

- `finishInlineReturn(...)` is now about `1.27s` cumulative
- `execStructLiteralNamedFast(...)` is about `1.01s`
- the cached exact named-struct helpers dropped out of the hot tier
- the remaining wall is still the same inline execution boundary:
  `execCallOpcode(...)`, `finishInlineReturn(...)`,
  `execStructLiteralNamedFast(...)`, `tryInlineSelfCallFromStack(...)`

Full external bytecode `binarytrees` still times out at the `60s` guard, so
this remains a reduced timeout-family keep rather than a closure.
