# Bytecode Cross-Suite Go PGO Rejection

Date: 2026-07-18

## Decision

Keep no bytecode VM, compiler, runtime, build-tooling, canonical-stdlib,
workload, fixture, or language change from this tranche.

A merged Go PGO profile trained on six unlike selected bytecode applications
produced material improvements on most of a disjoint nine-application
validation set. It nevertheless failed both deployability and broad
performance requirements:

1. Go extern plugins must be rebuilt with the exact same PGO profile as the
   CLI or Go rejects their package ABI.
2. The matching plugin build requires the training profile to remain present
   and propagated into every runtime plugin build.
3. Corrected, profile-matched runs still regress unseen JSON by 5.61% across
   the final three warm pairs, losing all three. The complete five-pair JSON
   mean, including cold plugin builds, regresses 58.31%.

The candidate therefore did not advance to the full selected-bytecode
scorecard or permanent tooling integration. No WASM work was performed.

## Build and training contract

The baseline was built explicitly with `-pgo=off` under Go 1.26.4 and exactly
reproduced the existing CLI SHA-256:
`b406e483ee6f05bc7c9d3681ffe8da64de297c3b3a2bec7b63310d122cfe63e0`.

One verifier-backed CPU profile was collected on CPU 0 from each training
application under `GOMAXPROCS=1`, `GOMEMLIMIT=1GiB`, `GOGC=50`, the catalog's
executor/input contract, and canonical external `able-stdlib`:

- Word Frequency: text and maps;
- Mandelbrot: floating-point numeric loops;
- Option/Result Configuration: union matching and errors;
- Mutex Ledger: concurrency and synchronization;
- Unicode Scalar Pipeline: Unicode and iteration;
- Array Slice Window: arrays and collection operations.

All six executions verified. `go tool pprof -proto` merged 12.70 seconds of
CPU samples into a 46,566-byte profile with SHA-256
`bad72bbe14e41d0c45cd7957a4bf0be62da1af346d92f154c7144895090ee48e`.
The primary shared sample was `runResumable` at 10.87% flat and 90.79%
cumulative; its children included stack/slot transport, binary operations,
float loops, integer extraction, allocation, and map access. The exact
training runs are retained in `2026-07-18-bytecode-pgo-training-runs.tsv`.

The PGO build was otherwise source-identical:

| Item | Baseline | PGO |
| --- | ---: | ---: |
| CLI SHA-256 | `b406e483ee6f05bc7c9d3681ffe8da64de297c3b3a2bec7b63310d122cfe63e0` | `786f9d92996d542b7824adad738ddc4d66b21fdb810b6d87532589a19ba5a08e` |
| Binary size | 45,813,432 B | 46,159,688 B |
| `runResumable` size | 35,909 B | 42,053 B |

PGO therefore changed code generation materially: the binary grew 346,256
bytes and the primary dispatcher grew 6,144 bytes.

## Disjoint validation and plugin ABI discovery

The nine validation applications were excluded from training: Base64,
Dependency Plan, Distance Field, Future Pipeline, I-Before-E, JSON, Rational
Series, Reverse Complement, and Regex Stream Audit. Together they cover text,
bytes, graphs/collections, numeric arrays, concurrency, matching, wide
integers, sequence processing, and canonical regex execution.

The first five-pair gate used an ordinary environment for both binaries.
Dependency Plan, Distance Field, Future Pipeline, and Rational Series completed
for both variants. Every PGO execution of Base64, I-Before-E, JSON, Reverse
Complement, and Regex Stream failed before Able main execution with Go's
plugin error:

`plugin was built with a different version of package internal/runtime/maps`

Those five applications load Go extern plugins. The baseline side and all
plugin-free PGO rows verified, yielding 65 verified processes and 25 explicit
PGO errors. The complete failed-admission ledger is retained in
`2026-07-18-bytecode-pgo-unmatched-plugin-validation.tsv`; its error timings
are not performance samples.

The compatibility diagnosis was then tested, not assumed. PGO launches set
`GOFLAGS=-pgo=<exact merged profile>`, causing runtime-built extern plugins and
their dependencies to use the identical profile as the CLI. All 90 corrected
processes completed and verified. This proves the semantic/plugin boundary can
work, but also proves that a PGO CLI is not a standalone artifact: the exact
profile becomes a runtime plugin-build dependency.

## Corrected repeated gate

Every application received five order-balanced baseline/PGO pairs on CPU 0.
All samples and workstation/plugin-build outliers remain in the arithmetic
means. The exact ledger is retained in
`2026-07-18-bytecode-pgo-matched-validation.tsv`.

| Validation application | Samples/variant | Baseline mean | PGO mean | Change |
| --- | ---: | ---: | ---: | ---: |
| Base64 | 5 | 3.188 s | 5.321 s | +66.92% |
| Dependency Plan | 5 | 0.488 s | 0.474 s | -2.82% |
| Distance Field | 5 | 5.848 s | 5.605 s | -4.14% |
| Future Pipeline | 5 | 0.411 s | 0.385 s | -6.46% |
| I-Before-E | 5 | 0.544 s | 0.628 s | +15.41% |
| JSON | 5 | 0.821 s | 1.299 s | +58.31% |
| Rational Series | 5 | 4.000 s | 3.772 s | -5.70% |
| Reverse Complement | 5 | 6.953 s | 6.729 s | -3.22% |
| Regex Stream Audit | 5 | 3.801 s | 3.602 s | -5.23% |

The large Base64, I-Before-E, and JSON means include cold PGO plugin-build
work. Those samples are real integration costs and are not discarded. To
separate that cost from warm code generation, pairs three through five were
also summarized without using them as replacement selection evidence:

| Extern application, pairs 3-5 | Baseline mean | PGO mean | Change |
| --- | ---: | ---: | ---: |
| Base64 | 3.252 s | 3.136 s | -3.56% |
| I-Before-E | 0.567 s | 0.511 s | -9.97% |
| JSON | 0.816 s | 0.862 s | +5.61% |

Thus the candidate fails even after warmup: unseen JSON consistently regresses
while the other families improve. Retuning the training mix to JSON or the
validation set would violate the disjoint generality gate and risk producing a
benchmark-trained runtime.

## Restoration and scope

- No source or build-tooling change was made.
- The baseline remains the ordinary `-pgo=off`/no-`default.pgo` build.
- No profile was installed into the repository or Go plugin environment.
- The canonical external stdlib was unchanged.
- Raw profiles, binaries, plugin outputs, and runners are cleanup-only; the
  compact training and validation ledgers are retained.

## Next recommendation

Return to the compiled-performance lane with a coverage-wide generated-helper
execution census across all selected compiled applications.

Why: the bytecode dispatcher has now rejected direct layout changes, a large
two-tier reduction, and cross-suite host PGO. Meanwhile only 7 of 35 compiled
rows meet the 95%-of-Go goal, but recent top profiles divide below broad parent
symbols. The next useful evidence is an exact matrix of which shared generated
runtime/compiler helpers execute materially across unlike applications.

What it entails: add temporary opt-in counters at general generated helper
families—dynamic calls, boxed primitive operators, conversions, interface
dispatch, checked primitive boundaries, and shared nominal encoding—then run
the selected compiled suite once under the usual verifier and sub-minute
guardrails. Profile only an exact helper that is material in at least three
unlike applications, and admit a candidate only if it improves separate
guards without a named-container or application-specific lowering. Remove all
counters afterward and update canonical `able-stdlib` only if the census
demonstrates a reusable source/API boundary. Continue to defer WASM.
