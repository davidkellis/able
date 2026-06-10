# Feature-interaction application gate — 2026-07-20

## Decision

Retain two portable multi-feature applications and the pairwise coverage-matrix
tooling. Retain no compiler, generated-runtime, bytecode-VM, canonical-stdlib,
or language performance change.

The applications expose large product gaps, but their exact profiles repeat
already-owned and already-gated mechanisms. The compiled rows reproduce the
rejected goroutine-identity boundary. The bytecode rows reproduce completed
member-cache, return/type-match, allocation, and GC families. A new candidate
would therefore be a retry of a failed broad experiment rather than evidence
from a newly discovered shared wall.

## Coverage selection

The checked-in feature manifest previously described isolated family breadth
but did not quantify intersections. `bench_feature_interaction_matrix` now
builds all pairwise intersections for portable/mixed feature families. Package
loading is excluded from the matrix because every portable application has that
property and it does not discriminate workloads.

Removing the two new applications from the current manifest reconstructs the
pre-tranche baseline without maintaining a second manifest:

| Measure | Before | Current |
| --- | ---: | ---: |
| portable/mixed families | 11 | 11 |
| pairwise interactions | 55 | 55 |
| zero-coverage pairs | 29 | 15 |
| pairs improved by this tranche | — | 32 |

The two applications close 14 previously empty pairwise intersections. The
most important new combinations are concurrency with text/files, closures,
interfaces, Option/Result errors, program entry, and stdlib protocols. This was
the two-application snapshot; two later applications completed the cohort. The
cumulative machine-readable and rendered matrix is in
`2026-07-20-feature-interaction-coverage-matrix.{json,md}`.

### Concurrent Text Index

This application reads the shared 16,384-word corpus, sends indexed words to
four Channel workers, returns nullable nominal `WordScore` values, aggregates
through the public `Map` interface, and prints a commutative checksum. It mixes
file/text processing, nullable unions, inherent methods, interface dispatch,
HashMap, Future/Channel concurrency, and a real entry argument.

Expected output in Able, Go, Python, and Ruby:

```text
16384:16384:8534:8534:8329907:12880517
```

### Validated Job Pipeline

This application sends 16,384 jobs to four workers, validates each job as
`Error | i64`, maps successful results through a closure, recovers failures,
returns nominal `JobResult` values over a Channel, and prints a commutative
checksum. It combines concurrency, generic-union methods, dynamic Error values,
closures, nominal methods, and interface/type matching.

Expected output in Able, Go, Python, and Ruby:

```text
16384:16384:16384:14608:1776:7290759586:258814
```

Canonical and sibling Able sources are byte-identical. Each application also
has Go 1.26, Python 3.14, and Ruby 4.0 implementations, Docker contracts, an
external verifier, and a README in the sibling benchmark repository.

## Repeated baselines

Every row below is an arithmetic mean of five successful verifier-backed
processes under the catalog CPU/executor contract. No single run determines a
classification.

| Application | Able mode | Able mean | Reference | Reference mean | Ratio |
| --- | --- | ---: | --- | ---: | ---: |
| Concurrent Text Index | compiled | 1.0180 s | Go | 0.0057 s | 178.60x |
| Concurrent Text Index | bytecode | 0.5760 s | Ruby | 0.0997 s | 5.78x |
| Concurrent Text Index | bytecode | 0.5760 s | Python | 0.0863 s | 6.67x |
| Validated Job Pipeline | compiled | 3.0840 s | Go | 0.0059 s | 522.71x |
| Validated Job Pipeline | bytecode | 0.7240 s | Ruby | 0.0806 s | 8.98x |
| Validated Job Pipeline | bytecode | 0.7240 s | Python | 0.0939 s | 7.71x |

All 20 Able executions and all 30 reference executions verified. These are
targeted pre-promotion rows; they are not spliced into the reviewed scorecard or
frontier until a complete promotion refresh supplies the normal independent
evidence contract.

Evidence:

- `2026-07-20-feature-interaction-baseline.{json,md}`
- `2026-07-20-feature-interaction-go-reference.{json,md}`
- `2026-07-20-feature-interaction-interpreter-reference.{json,md}`

## Exact profiles

Three whole-process CPU profiles per compiled application were merged. Three
whole-process bytecode profiles were also captured, but they are not used for
runtime ownership: source loading/typechecking dominated them and CPU profiling
greatly inflated that cold path. Main-only bytecode profiles provide the valid
runtime attribution.

| Application/mode | Exact profile result | Decision |
| --- | --- | --- |
| Concurrent Text Index compiled | `bridge.currentGID` 96.94% cumulative; `runtime.Stack` 96.61% | Same generated concurrency identity boundary already closed by the fixed-context default gate. |
| Validated Job Pipeline compiled | `bridge.currentGID` 95.97% cumulative; `runtime.Stack` 95.80% | Same owner. The general fixed-context replacement improved concurrency but regressed unrelated N-Body by 54.7%, so it is not retried. |
| Concurrent Text Index bytecode main | 369,137,107 ns/op, 33,183,402 B/op, 377,074 allocs/op over three measured calls; cached member lookup 20.91% cumulative | Existing dependency-validated member-cache family; no new text/Channel specialization. |
| Validated Job Pipeline bytecode main | 813,910,942 ns/op, 89,847,760 B/op, 1,387,899 allocs/op over three measured calls; `execCallMember` 27.57%, `finishInlineReturn` 12.35%, cached member lookup 8.64%, `matchesType` 8.23% cumulative | Existing member, return/type-match, allocation, and GC owners; no Result/Channel/JobResult specialization. |

The job-pipeline runtime profile must set `ABLE_BENCH_SKIP_TYPECHECK=0`.
Generic union payloads intentionally carry no nominal wrapper, so checked static
receiver facts distinguish `Option.map` from `Result.map`. The internal runtime
benchmark defaults to skipping typechecking; under that diagnostic-only mode a
worker fails with `Ambiguous overload for map`, and the main routine can wait for
a result that will never arrive. Enabling typechecking happens before the timed
region, matches normal execution semantics, installs 498 inference facts across
four packages, and completes the benchmark. This was a profiling-contract issue,
not a production concurrency failure, and no runtime workaround was retained.

## Scope and admission result

No benchmark-specific compiler lowering, VM opcode, nominal fast path, named
container rule, stdlib shortcut, or WASM work was introduced. The new
applications are useful precisely because they combine ordinary language
features and independently reproduce known owners. They expand the chance that
a future architecture change must help real combinations, while the existing
unlike-application guards continue to reject narrow wins.

The performance frontier remains unchanged until promotion: 75 selected rows,
8 snapshot meets, 67 misses, and zero actionable groups.
