# Scoreboard Unranked-Reason Transparency — 2026-07-15

## Scope

The external application scoreboard already preserved bounded Able-run status,
but its generated JSON and Markdown reported every non-rankable row only as
`unranked`. Readers had to inspect the promoted source scorecards to determine
whether the cause was an Able timeout, a partial Able result, or the absence of
a required comparison ratio.

This tranche changes only the report generator and its checked-in generated
artifacts. It does not run or alter benchmarks, references, timeout guards, the
VM, compiler, runtime, fixtures, or canonical stdlib.

## Retained report contract

`v12/bench_external_scoreboard` now emits `unranked_reason` for every
`target_status: "unranked"` row:

- `able_incomplete`, `able_timeout`, `able_failure`, or `able_unavailable`
  when the verifier-backed Able row cannot rank;
- `go_reference_unavailable`, `python_reference_unavailable`, or
  `ruby_reference_unavailable` when a verified Able row lacks a required valid
  comparison ratio.

For the bytecode target, two missing interpreter references are joined in a
stable language order. Rankable `meets` and `miss` rows carry `null` because no
unranked explanation applies. The Markdown table presents matching human
labels.

`*_reference_unavailable` intentionally makes no claim about why the foreign
source has no valid ratio. That evidence remains in the selected source
scorecard; this avoids treating a capped external reference as absent or
silently changing a guard policy.

## Current cohort check

The regenerated July 15 scoreboard classifies the compiled Sudoku row and
seven bytecode rows as `able_timeout`, K-Nucleotide as `able_incomplete`, and
bytecode Fib as `python_reference_unavailable`. Every current unranked row has
a reason and no rankable row has one.

## Verification

```sh
python3 -m py_compile v12/bench_external_scoreboard
./v12/bench_external_scoreboard --check
./v12/bench_catalog_check
just bench-scoreboard-check
```

All commands passed. These are report-integrity checks only and collect no new
performance timing evidence.
