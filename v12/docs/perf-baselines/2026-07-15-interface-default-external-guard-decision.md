# Generic interface-default external guard — 2026-07-15

## Scope

The bytecode repair that preserves the enclosing `Iterator<T>` parameters on a
dynamically materialized interface default method was a correctness change, not
a performance candidate. This guard checks the two independently verified
portable applications that actually exercise the public lazy
`filter`/`map`/`collect` pipeline:

- Document Audit;
- Lexical Rollup.

No third verifier-backed application uses this interface-default route. The
local linked-list iterator warmup remains a semantic/regression control only;
it is not a replacement for an unrelated external application.

## Measurement

The following command made five independent normal processes for each
application/mode on the shared workstation, with no CPU-affinity or
quiet-host requirement. Each successful output was checked by the canonical
Ruby verifier. `GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1` bound the
Able process only.

```sh
env GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1 \
  ./v12/bench_compare_external \
  --benchmarks document_audit,lexical_rollup \
  --modes compiled,bytecode --runs 5 --timeout 45 \
  --output-json v12/docs/perf-baselines/2026-07-15-interface-default-external-guard.json \
  --output-md v12/docs/perf-baselines/2026-07-15-interface-default-external-guard.md
```

| Application | Mode | Verified runs | Mean real time |
| --- | --- | ---: | ---: |
| Document Audit | compiled | 5/5 | 0.092 s |
| Document Audit | bytecode | 5/5 | 0.298 s |
| Lexical Rollup | compiled | 5/5 | 0.090 s |
| Lexical Rollup | bytecode | 5/5 | 0.464 s |

The checked Able source hashes exactly match the current source-aligned
scorecard. Its retained Go/Python/Ruby references therefore remain valid
context, but this guard did not rebuild foreign references and does **not**
replace or reclassify the scorecard.

## Selection result

Both applications execute correctly after the generic-default repair. These
five-run Able-only guard means use retained, rather than freshly rebuilt,
foreign references, so they are not a scorecard delta or a claim of a
performance gain/regression. They are only two applications, and the earlier
paired profile evidence already splits their remaining costs below the shared
pipeline parent: member-cache and inline-return ideas were rejected broadly,
while the concrete descendants diverge. See
`2026-07-10-external-scorecard-selection.md`.

Keep no VM, compiler, generated-runtime, canonical-stdlib, or benchmark
performance change. In particular, do not optimize `Iterator`, `filter`,
`collect`, Array, or either benchmark as named cases.

The focused generic-default metadata and iterator regressions also pass:

```sh
cd v12/interpreters/go
env GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1 \
  go test ./pkg/interpreter \
  -run 'TestIteratorDefault(MethodRetainsInterfaceGenericReturn|InterfaceMethodValueCache)$|TestBytecodeLinkedListIterator(Collect|FilterMap)BenchWarmup$' \
  -count=1 -timeout 55s
```

## Next eligible performance work

The external scorecard has no new pair or trio with a repeated, concrete,
non-nominal language-level leaf. The next performance tranche should wait for
a material cross-cutting semantic or portability change, then select at least
three unlike verifier-backed applications that newly exercise it. That is the
only route that can improve real programs without encoding a benchmark shape.
