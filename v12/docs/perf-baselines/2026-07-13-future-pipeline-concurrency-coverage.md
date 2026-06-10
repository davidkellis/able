# Future Pipeline concurrency coverage (2026-07-13)

## Decision

Add `future_pipeline` as the second verifier-backed external concurrency
application. It is coverage infrastructure, not a runtime, compiler, or
stdlib optimization. No performance candidate is selected from this one
application alone.

## Application shape

The new application has matching Able, Go 1.26, Python 3.14, and Ruby 4.0
implementations under `../benchmarks/future-pipeline` plus the canonical Able
source under `examples/benchmarks/future_pipeline`.

- The producer sends 8,192 numeric jobs to four workers through a bounded
  channel. Each worker computes a deterministic 64-value weighted block and
  sends its result to a bounded collector channel.
- The collector waits for four completion sentinels and adds the four Future
  return values. This checks task scheduling, channel traffic, flush, and
  Future completion independently of Channel Rollup's file/String filtering
  shape.
- A cancellation probe sends a readiness signal, executes `future_yield()`,
  then a separate Future cancels it. The final status must be `Cancelled`.
- Every implementation and the Ruby verifier require exactly
  `8192:8192:1511819313164:8192:8192:1`.

The Able typechecker permits `future_yield()` only directly in an asynchronous
task, not in a named helper invoked from one. The worker helpers therefore
remain ordinary functions; the direct cancellation Future supplies the
cooperative-yield contract without inventing a special runtime path.

## Harness integration

`bench_external_catalog.sh` adds a `future-pipeline` suite, includes it in
`concurrency` and `coverage`, declares the goroutine executor, and maps its
hyphenated sibling directory. It also marks the entry source root only: the
sibling `run.able` has the same package name but is Docker runtime material,
so it must not become a second user source root during local Able builds.

## Fresh smoke baseline

The normal external harness used two verifier-backed runs with
`GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`; this short unpinned reading
is a baseline for future paired work, not a stable target scorecard.

| Mode | Able time | Go time | Python time | Ruby time | Able/Python | Able/Ruby |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| compiled | 1.0950 s | 0.0061 s | 0.1023 s | 0.1002 s | 10.70x | 10.93x |
| bytecode | 0.6500 s | 0.0061 s | 0.1023 s | 0.1002 s | 6.35x | 6.49x |
| tree-walker | 3.6200 s | 0.0061 s | 0.1023 s | 0.1002 s | 35.39x | 36.13x |

All six final-source Able measurements were verified. Direct smoke executions
of both Go interpreters, the compiled Able binary, Go, Python, Ruby, and the
verifier also produced the same exact output.

## Next recommendation

Refresh the whole `concurrency` suite—BinaryTrees, Channel Rollup, and Future
Pipeline—under the existing CPU-pinned cross-language guard, then capture
separate steady-state bytecode and compiled profiles for Channel Rollup and
Future Pipeline. This is necessary because the new benchmark establishes only
the second independent application shape; a scheduler change is eligible only
when the same material concrete helper recurs in both channel/Future programs
and remains neutral on serial text and numeric controls.
