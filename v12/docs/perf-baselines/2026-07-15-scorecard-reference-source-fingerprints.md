# Scorecard Reference-Source Fingerprints — 2026-07-15

## Purpose

An Able/reference ratio is not reproducible if either side can change while an
old report remains promoted. The preceding Able-source contract records the
selected Able source. This tranche applies the same identity rule to the
matched Go, Python, and Ruby reference programs.

## Contract

Fresh reference refreshes now record `source_sha256` beside each reference
source path:

- `bench_refresh_go_refs` fingerprints the selected Go `app.go` before its
  verified reference launch;
- `bench_refresh_interpreter_refs` fingerprints the selected Python `run.py`
  or Ruby `run.rb` before its verified reference launch.

When `bench_compare_external` selects a fresh reference row, it verifies that
digest against the live source and writes a `reference_source` path/SHA-256
object into the resulting Able comparison row. A changed fresh reference source
therefore rejects comparison generation and requires a reference refresh.

The scoreboard preserves that object under each comparison's `source` field
and validates it again before promotion. For retained comparison reports that
predate source hashing, it follows their named Go/interpreter refresh report,
hashes the live selected reference source, and labels the output
`"provenance": "current"`. A later edit makes `--check` stale and a bare
`just bench-scoreboard` refuses to rewrite the report; the normal remedy is a
fresh `just bench-scorecard-refresh`.

Stored external `results.json` references remain intentionally un-fingerprinted
because they do not identify a local source file. They are not used by the
promoted July 15 full-coverage cohort, whose Go/Python/Ruby comparisons all
come from named fresh reference reports.

## Current cohort

The current scorecard retains its original timing values, target statuses, and
reference implementations. It now carries 91 current-source legacy reference
fingerprints—the available matched comparisons across its 64 Able rows. Every
digest was checked against its sibling Go/Python/Ruby source file.

## Verification

```sh
bash -n v12/bench_refresh_go_refs v12/bench_refresh_interpreter_refs \
  v12/bench_compare_external v12/bench_external_scoreboard
python3 -m py_compile v12/bench_external_scoreboard
./v12/bench_external_scoreboard --check
./v12/bench_catalog_check
just bench-scoreboard-check
```

A synthetic verifier-backed comparison with measured Able and Go source
fingerprints rendered as `provenance: measured`. An in-memory changed Go
reference fingerprint exercised the promotion-drift refusal. No timing
workload was run.
