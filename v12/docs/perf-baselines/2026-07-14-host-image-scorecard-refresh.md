# Host-image bytecode scorecard refresh — 2026-07-14

## Decision

Keep the complete-program extern host image and the idempotent prelude
registration repair. They make normal multi-package bytecode programs use one
loaded Go host image rather than independently loading one plugin per extern
package. This is host-boundary correctness and startup hygiene, not a selected
VM performance optimization.

## Method

The comparison used the existing verifier-backed Python/Ruby references,
three CPU-pinned Able processes per workload, and the persistent host-image
cache under the project temporary directory:

```text
TMPDIR="$PWD/.gotmp" ABLE_EXTERN_CACHE_DIR="$PWD/.gotmp/extern-host-image-scorecard" \
  ./bench_compare_external --benchmarks i_before_e,json,channel_rollup \
  --modes bytecode --languages python,ruby --runs 3 --timeout 90 \
  --cpu-affinity 15 \
  --reference-json docs/perf-baselines/2026-07-14-bytecode-generality-interpreter-refresh.json \
  --output-json docs/perf-baselines/2026-07-14-host-image-bytecode-scorecard.json \
  --output-md docs/perf-baselines/2026-07-14-host-image-bytecode-scorecard.md
```

All outputs passed their suite verifier. The retained machine-readable rows
are in `2026-07-14-host-image-bytecode-scorecard.json`.

| Workload | Able bytecode | Python | Able/Python | Ruby | Able/Ruby |
| --- | ---: | ---: | ---: | ---: | ---: |
| i-before-e (text/files) | 0.5200 s | 0.0923 s | 5.63x | 0.1216 s | 4.28x |
| JSON | 0.8300 s | 2.7039 s | 0.31x | 1.7460 s | 0.48x |
| Channel Rollup | 0.5767 s | unavailable | n/a | unavailable | n/a |

The first one-run attempt failed before timing because program pre-registration
and normal module evaluation appended the same Go prelude twice. That changed
the image hash, invoked the legacy fallback, and made its generated source put
an `import` after prelude declarations. `registerExternStatements` now keeps
each target's identical prelude once. The two-package image regression includes
a real `strings` prelude, re-evaluates the modules after prewarm, and calls both
same-named extern functions; it prevents both the hash drift and invalid
fallback source from returning.

## Result

The image removes the repeated-plugin loader boundary for the ordinary loaded
program path, but these three unlike workloads do not identify a shared VM
leaf or a broad ratio improvement. In particular, the text workload remains
well below the interpreter target while the JSON workload already beats both
references. No bytecode opcode, compiler lowering, benchmark source, or
stdlib-specific optimization is eligible from this refresh. The external
`able-stdlib` checkout did not change.

## Follow-up

The bounded text/file and map/iterator profiles are complete in
`2026-07-14-host-image-bytecode-miss-profiles.md`. Their concrete leaves do
not recur, so no VM candidate was selected. The next measurement priority is
to add reference-aware Python/Ruby ratios for broader coverage applications
before spending more time on an unchanged individual profile.
