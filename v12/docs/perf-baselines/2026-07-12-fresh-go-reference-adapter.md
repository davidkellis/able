# Fresh Go reference adapter (2026-07-12)

## Change

`v12/bench_compare_external` now accepts:

```text
--go-reference-json PATH
```

The path is the existing `v12/bench_refresh_go_refs` JSON schema. Fresh Go
rows replace only Go's stored external rows. The existing `--reference-json`
continues to replace Python/Ruby rows from
`v12/bench_refresh_interpreter_refs`. Both flags may be supplied together;
language families without a fresh input retain their stored rows.

The report records both source paths separately in JSON and Markdown. This
makes the displayed Able/Go ratio use the current pinned Go reference rather
than an older corpus ledger value.

## End-to-end guard

The adapter was checked with fresh one-run CPU-2 Monte Carlo Pi references:

| Reference | Fresh real time (s) | Status |
| --- | ---: | --- |
| Go 1.26.4 | 0.2006 | verified |
| Ruby 4.0.5 | 1.7032 | verified |

A one-run compiled Able comparison (`0.2900s`, verified) supplied both fresh
reports. Its Markdown displayed Go `0.2006s` and Ruby `1.7032s`; JSON
assertions confirmed each comparison value and its reference provenance. Shell
syntax and the existing focused compiler suite also pass.

## Decision

Keep this generic measurement change. It changes no runtime, compiler
lowering, benchmark algorithm, or `able-stdlib` API. It prevents stale Go
ledger rows from silently steering performance decisions.

## Next recommendation

Use the combined fresh-reference inputs for a bounded three-family scorecard
of existing generality workloads, then select implementation work only when a
shared material miss survives against Go, Ruby, and Python controls. Why: the
adapter now makes the benchmark suite a reliable decision surface rather than
a mix of current and historical reference timings. The work entails fresh
verifier-backed references, multi-run Able rows, and profiles only after a
repeated concrete helper appears.
