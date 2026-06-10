# Static compiled metadata-retention gate — 2026-07-14

## Scope

This gate evaluates the compiler-wide static/dynamic metadata-retention rule:
no-bootstrap executables omit package-interface `FunctionSignature.DefaultImpl`
AST bodies, while bootstrap and library output retain existing metadata. It is
not a named-container, iterator, text, file, or benchmark-specific lowering.

## Semantic and generated-source guards

- `TestCompilerStaticMainOmitsPackageInterfaceDefaultASTMetadata` proves a
  static launcher keeps its signature but emits no decoded package default AST.
- `TestCompilerBootstrapMainRetainsPackageInterfaceDefaultASTMetadata` proves
  a fallback launcher keeps the decoded metadata.
- The no-bootstrap `10_15_interface_default_generic_method` fixture and the
  focused default-interface/local-metadata compiler tests passed.
- A fresh canonical-stdlib Word Frequency build selected
  `RegisterIn(nil, entryEnv)` and had zero `DecodeNodeJSON` calls across its
  generated `compiled*.go` and `main.go` files. The earlier source audit found
  39 package-default calls. The build completed in 23.33 seconds; that is a
  non-regression completion check, not a portable timing comparison.

## Verified coverage scorecard

Command:

```sh
./v12/bench_compare_external \
  --benchmarks word_frequency,document_audit,lexical_rollup \
  --modes compiled --languages go --runs 3 --timeout 90 --cpu-affinity 15 \
  --go-reference-json v12/docs/perf-baselines/2026-07-14-compiled-coverage-go-refs.json
```

| Benchmark | Validation | Candidate | Previous scorecard | Change |
| --- | --- | ---: | ---: | ---: |
| Word Frequency | 3/3 verified | 0.2067s | 0.2100s | -1.6% |
| Document Audit | 3/3 verified | 0.0733s | 0.0800s | -8.4% |
| Lexical Rollup | 3/3 verified | 0.1000s | 0.1200s | -16.7% |

The candidate introduced no coverage regression. Each output SHA-256 matched
its existing Ruby verifier.

## Verified generality scorecard

Command:

```sh
./v12/bench_compare_external \
  --suite generality --modes compiled --languages go --runs 1 --timeout 90 \
  --cpu-affinity 15 \
  --go-reference-json v12/docs/perf-baselines/2026-07-14-compiled-generality-go-refresh.json
```

All 15 rows that were previously supported completed and passed their Ruby
verifiers: `fib`, `binarytrees`, `matrixmultiply`, `quicksort`, `sudoku_masks`,
`i_before_e`, `base64`, `json`, `monte_carlo_pi`, `pidigits`, `mandelbrot`,
`reverse_complement`, `k_nucleotide`, `nbody`, and `tapelang_alphabet`.
`sudoku` retained its existing 90-second timeout. There were no new failures or
broad result regressions. The one-run evidence is a generality guard, not a
precision performance claim; the three-run coverage table is the primary
measurement.

## Decision

Retain the change. It removes static-only unreachable metadata while preserving
the interpreter path and verified application behavior. The next tranche must
return to profile evidence and seek a repeated leaf cost rather than extending
this rule to local declarations or a nominal library type.
