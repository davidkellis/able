# Complete scorecard cohort variance — 2026-07-15

## Scope

The five-run retention work made one focused comparison auditable, but the
promoted external application result is a complete scorecard assembled from
many bounded comparison reports. Selecting only a few source files by hand
would weaken the broad-program rule and could accidentally compare two
scorecards that share the same underlying measurements.

`bench_variance_report` now accepts repeated `--scorecard` arguments for
complete `external-benchmark-scoreboard` JSON artifacts. It reads every cited
comparison source, uses the retained per-run Able and foreign-reference
samples when present, and preserves source-level ratios as unpaired values.

## Cohort safeguards

Each supplied scorecard must:

- cite non-empty comparison sources that reproduce its own benchmark/mode rows
  exactly;
- have precisely the same benchmark/mode coverage as every other supplied
  cohort; and
- share no underlying comparison report with another supplied cohort.

Each new refresh also writes one canonical stdlib source-state artifact before
timing begins. It hashes all `src/**/*.able` files in the external
`able-stdlib` checkout, and records its root, file count, Git head when
available, and dirty state. The aggregate carries that immutable record through
`--cohort` replay and promotion. A dirty checkout is valid evidence—it is the
actual library input—not an automatic failure.

At least two cohorts are required for this mode. The last rule means a copied
scorecard, a later aggregate that reuses an earlier comparison, or a partial
refresh cannot be presented as independent ratio evidence. The reader remains
report-only: it neither launches a benchmark nor creates a performance gate.

For candidate-selection evidence, add `--require-runs 5`. This opt-in check
requires every cited Able row and every requested fresh-reference language to
have exactly five successful timed samples, runs numbered one through five,
verified output hashes, a measured reference-source fingerprint, and a usable
ratio. It intentionally rejects legacy one- and three-run source reports;
they remain readable without the option as historical context, not as a new
five-run performance claim. Strict cohort evidence also requires the canonical
stdlib source state, so a newly claimed candidate cannot omit a material
runtime-library input.

## Use

After a material, cross-cutting candidate has earned fresh full-suite runs,
compare the two independently refreshed scorecards directly:

```sh
just bench-variance-report \
  --scorecard v12/docs/perf-baselines/<first-full-scorecard>.json \
  --scorecard v12/docs/perf-baselines/<second-full-scorecard>.json \
  --selection-manifest v12/bench-selection-manifest.json \
  --require-runs 5 \
  --output-json /tmp/able-full-suite-variance.json \
  --output-md /tmp/able-full-suite-variance.md
```

The generated report records the complete scorecard path and every expanded
comparison source, so a review can trace timing and ratio samples back to the
verifier-backed runs.

Follow-up: strict complete-scorecard variance now requires the explicit
mode-aware selection manifest. Each aggregate still retains every full-status
row, including timeouts, while the successful-run gate applies to the same
reviewed selected rows in both cohorts. The current contract and re-admission
rules are recorded in
`2026-07-15-mode-aware-scorecard-selection-manifest.md`.

To create those two independent cohorts without relabeling an experiment as
the current baseline, choose distinct tags and use `--no-promote` on both
refreshes:

```sh
just bench-scorecard-refresh --tag candidate-a --no-promote
just bench-scorecard-refresh --tag candidate-b --no-promote
```

Each writes its own `<tag>-refresh.json` aggregate and refuses to overwrite an
existing path. If one reviewed aggregate should become current, promote its
exact source manifest without rerunning work:

```sh
just bench-scoreboard --cohort v12/docs/perf-baselines/candidate-b-refresh.json
```

The `--cohort` option rebuilds the scoreboard from that aggregate's listed
comparison reports; it cannot be mixed with hand-selected `--input` paths. It
also preserves the aggregate's canonical stdlib source state. The refresher
itself rejects promotion unless `--runs` is exactly five, so a lower-run smoke
command must include `--no-promote`.

## Verification and decision

A temporary two-cohort harness check with disjoint comparison files confirmed
that the report contains two source-level ratio samples for every row. A
second check with the current and reconciled scorecards correctly failed
because they cite the same comparison sources. A retained modern single
comparison still reports five Able and five foreign-reference timing samples
through `--input`.

The strict five-run check passes for that retained modern comparison and for a
temporary disjoint two-cohort harness fixture. The current scorecard correctly
fails first because it lacks a canonical stdlib source state, and its cited
reports also contain fewer than five retained runs; no historical timing was
reclassified or rerun to make it pass.

Dry-run lifecycle checks confirm that two distinct tags produce 31 disjoint
JSON artifact paths with the five-run default, `--no-promote` omits only the
current-scoreboard write, and normal refresh promotion remains the default. A
cohort replay of the current aggregate reproduces its checked-in JSON and
Markdown byte-for-byte without benchmark execution.

The current historical aggregate correctly remains source-state legacy. A
temporary state-aware aggregate replay retained the recorded 69-file external
stdlib fingerprint byte-for-byte, and a dry-run refresh showed the source-state
capture occurs before all reference and Able timing commands without creating
artifacts.

Keep no compiler, bytecode VM, canonical-stdlib, workload, target, timeout,
or scorecard performance change. This is evidence plumbing only. The next
measurement decision still requires a material cross-cutting change and fresh
five-run, verifier-backed cohorts before a profile can admit a shared runtime
or compiler candidate.
