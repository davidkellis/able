# Cache-Qualified External Scorecard

## Purpose and method

This report separates the retained extern-host cache packaging gain from
residual application execution. It is report-only and does not modify the
dirty `../benchmarks/results.json` ledger.

- Rebuilt `able-v12-base` runs `able cache prewarm` and contains 10 canonical
  Go extern-host plugins under `ABLE_EXTERN_CACHE_DIR=/able/v12/cache/extern`.
- Every Able compiled/bytecode image for Word-Frequency, Document-Audit,
  Lexical-Rollup, and Channel-Rollup was rebuilt from that base. Each direct
  Docker row below is the mean of three new containers, so it includes normal
  process loading/lowering but no first-use host-plugin compilation.
- The Go 1.26, Ruby 4.0, and Python 3.14 rows use the corresponding existing
  benchmark images and the same three-run Docker method. Ratios are meaningful
  only within a workload/mode row.
- A separate normal-process host ledger used a new, explicitly prewarmed
  persistent extern cache, `GOMEMLIMIT=1GiB`, and `GOGC=50`. Its timers are
  coarse and its source runner differs from Docker, so it is cache-state
  provenance and completion evidence, not a cross-environment comparison.

All Able images retained their expected output: Fib `1134903170`,
Word-Frequency `1937:11878177`, Document-Audit `1937:102:83257`, and both
Lexical-Rollup and Channel-Rollup `16384:4828:502100`.

## Fresh prewarmed-container rows

| Workload | Mode | Able (s) | Go (s) | Able/Go | Ruby (s) | Able/Ruby | Python (s) | Able/Python |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Word-Frequency | compiled | 0.177586 | 0.005349 | 33.20x | 0.070665 | 2.51x | 0.030224 | 5.88x |
| Word-Frequency | bytecode | 1.504245 | 0.005349 | 281.20x | 0.070665 | 21.29x | 0.030224 | 49.77x |
| Document-Audit | compiled | 0.043910 | 0.004291 | 10.23x | 0.051359 | 0.86x | 0.023583 | 1.86x |
| Document-Audit | bytecode | 0.300852 | 0.004291 | 70.10x | 0.051359 | 5.86x | 0.023583 | 12.76x |
| Lexical-Rollup | compiled | 0.059558 | 0.004461 | 13.35x | 0.062188 | 0.96x | 0.036428 | 1.64x |
| Lexical-Rollup | bytecode | 0.427349 | 0.004461 | 95.79x | 0.062188 | 6.87x | 0.036428 | 11.73x |
| Channel-Rollup | compiled, goroutine | 2.112241 | 0.005535 | 381.62x | 0.062462 | 33.82x | 0.076172 | 27.73x |
| Channel-Rollup | bytecode, goroutine | 0.463523 | 0.005535 | 83.74x | 0.062462 | 7.42x | 0.076172 | 6.09x |

The Docker bytecode rows replace the previous uncached process condition, not
the language-runtime target itself. They prove the cache change broadly across
text/map, filesystem/iterator, public pipeline, and concurrency applications.
They do not meet the Ruby/Python target: the remaining gaps are material in
every target-miss row, including the sequential controls.

## Persistent-cache host-process ledger

The isolated host cache prewarm discovered 68 visible packages (the workspace
also exposes benchmark fixture roots) and built the same 10 Go host modules.
The normal-process harness completed all three runs in each row:

| Workload | Compiled (s) | Bytecode (s) | Compiled GC | Bytecode GC |
| --- | ---: | ---: | ---: | ---: |
| Fib control | 3.5000 | 0.1333 | 6.00 | 6.33 |
| Word-Frequency | 0.1900 | 1.2767 | 8.00 | 11.00 |
| Document-Audit | 0.0667 | 0.2733 | 6.00 | 8.00 |
| Lexical-Rollup | 0.1067 | 0.4800 | 6.67 | 9.33 |
| Channel-Rollup, goroutine | 1.9067 | 0.4400 | 8.67 | 9.33 |

The reference JSON has no matching Word-Frequency, Document-Audit,
Lexical-Rollup, or Channel-Rollup entries for this source-runner condition, so
no ratio is claimed from these rows.

## No-extern control

Bytecode Fib is the no-Go-extern control. Three prewarmed-image runs averaged
`0.103477s`; three runs with `ABLE_EXTERN_CACHE_DIR` redirected to a new empty
directory averaged `0.101424s`. Both printed `1134903170`, and the empty
directory contained zero files afterward. The cache has no measurable path in
programs that do not require an extern host module.

## Decision

Retain the generic cache prewarm. Do not infer a VM speedup from its Docker
improvement, and do not target `fs`, Channel, HashMap, the loader, a named
container, or a benchmark program. The residual bytecode misses now recur in
independent sequential and concurrent applications after the shared child
build is removed.

## Next recommendation

Collect bounded bytecode-runtime profiles under the prewarmed cache for
Word-Frequency, Document-Audit, Lexical-Rollup, and Channel-Rollup, with Fib
as the no-extern control. Why: these are the residual target misses under one
cache condition, and a concrete shared VM descendant is now required before a
runtime change is credible. The work entails one profile per application,
output checks, and comparison of concrete leaves below VM dispatch; retain a
candidate only when the same generic operation repeats across at least two
independent target-miss programs and survives the other two guards.
