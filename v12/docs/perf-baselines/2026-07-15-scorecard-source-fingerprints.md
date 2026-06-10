# Scorecard Canonical-Source Fingerprints — 2026-07-15

## Purpose

An external performance ratio is meaningful only for the Able program that was
actually reviewed and launched. The canonical-source contract proves that the
catalog selects `v12/examples/benchmarks`; this tranche adds content identity
and stale-report protection.

## Contract

`bench_compare_external` now records the canonical target path and SHA-256 in
each raw comparison row's `able.source` object before it launches that target.
`bench_external_scoreboard` resolves every benchmark through the authoritative
shell catalog and verifies that a recorded source path and digest still match
the canonical file. Its generated row contains:

```json
"able_source": {
  "path": "v12/examples/benchmarks/...",
  "sha256": "...",
  "provenance": "measured"
}
```

For retained source reports created before this contract, the scoreboard
computes the same canonical path and SHA-256 at rendering time and labels the
row `"provenance": "current"`. That label makes clear that the July 15
timing reports did not themselves record a source hash.

If a measured fingerprint no longer matches the canonical source, generation
fails and requires a fresh scorecard. If a legacy current fingerprint changes,
`--check` reports the checked-in artifact as stale and a bare
`just bench-scoreboard` refuses to overwrite it. The normal promotion route is
then `just bench-scorecard-refresh`, which supplies explicit fresh comparison
reports carrying measured fingerprints.

The scorecard invokes the catalog's small `targets` command rather than
duplicating benchmark-to-source paths in Python. This keeps the provenance
check and benchmark executor on one target mapping.

## Current cohort

The promoted July 15 full-coverage scorecard now contains 64 current-source
legacy fingerprints: one for each compiled or bytecode row over the 32
canonical applications. Its existing timing values, ratios, statuses, and
selected source-report cohort are unchanged.

## Verification

```sh
bash -n v12/bench_external_catalog.sh v12/bench_catalog_check v12/bench_compare_external
./v12/bench_external_catalog.sh targets fib base64 future_pipeline
python3 -m py_compile v12/bench_external_scoreboard
./v12/bench_external_scoreboard --check
./v12/bench_catalog_check
just bench-scoreboard-check
```

A synthetic verifier-backed source report with a matching `able.source`
fingerprint was also rendered successfully as `provenance: measured`, and an
in-memory changed fingerprint exercised the promotion-drift refusal. No
benchmark workload was executed.
