# Canonical Benchmark Source Contract — 2026-07-15

## Purpose

The external benchmark harness must measure the same Able applications that
the project reviews and tests. `v12/examples/benchmarks` is the canonical Able
source location; the sibling `../benchmarks` repository supplies foreign
references, inputs, verifiers, run directories, and optional packaging only.

Without a mechanical check, a catalog entry could silently point to a
benchmark-local copy or a generated file. That would let measurement behavior
drift from the reviewed source and undermine the broad-performance policy.

## Retained contract

`v12/bench_catalog_check` now verifies, without building or timing a program,
that:

- every selected portable catalog target resolves beneath
  `v12/examples/benchmarks`;
- no two selected entries share a canonical Able source;
- the complete `coverage` catalog has one entry for every canonical
  `*.able` benchmark source and no extra entry.

Generated benchmark-local directories named `target` are deliberately pruned
from the source inventory. They are compiler execution residue, not benchmark
source; their presence must not alter the catalog contract.

The check does not require sibling `run.able` files to match canonical source.
Those files are external harness packaging and are not selected by
`bench_compare_external`, which receives its program path exclusively through
`bench_external_target(...)`.

## Result

The current full catalog has 32 portable applications and 32 distinct
canonical Able sources. The normal corpus check now prints that source count
beside its 77 bounded local fixtures.

## Verification

```sh
bash -n v12/bench_catalog_check
./v12/bench_catalog_check
./v12/bench_catalog_check --suite core
```

All commands passed. This is a static provenance check only: it changes no
benchmark source, reference, VM, compiler, runtime, stdlib, or timing result.
