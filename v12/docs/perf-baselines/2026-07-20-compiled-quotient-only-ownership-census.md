# Compiled quotient-only ownership census

Date: 2026-07-20

## Decision

Keep no compiler, generated-runtime, VM, canonical-stdlib, benchmark, fixture,
or language change from this census.

Corrected Sudoku executes the exact signed Euclidean helper materially, but
three unlike quotient consumers do not. Rational Series widens its hot
normalization arithmetic to `i128`; Regex Set emits signed quotient sites for
a DFA path while this workload executes its NFA path; and K-Nucleotide calls
signed quotient only while formatting a handful of output rows. A
quotient-only generated helper or constant-divisor lowering therefore fails
the three-unlike-application admission rule before an A/B candidate.

## Inventory

Only two portable benchmark entry sources contain direct `//` expressions:

- Sudoku Masks uses two `i32` quotient-only operations by the positive constant
  three in `square_index`.
- K-Nucleotide uses one `i32` quotient-only operation by the positive constant
  1000 in `format_percent`, adjacent to a separate `% 1000` expression.

Two independently authored applications reach additional canonical stdlib
sites:

- Rational Series exercises `able.numbers.rational`. Its hot GCD loop is
  remainder-only `i128`; normalization then performs two quotient-only `i128`
  operations by the computed positive GCD.
- Regex Set imports `able.text.automata`, whose three DFA binary searches use
  quotient-only `i32` midpoint calculations by positive constant two.

Canonical `/%` paired-result sites exist in the primitive numeric interfaces,
but none is a material selected-workload owner in this cohort. Generated
source also contains cold imported functions, so static helper presence is not
treated as execution evidence.

The existing signed helper already handles the simplest general case directly:
when `a >= 0 && b > 0`, it computes native Go `/` and `%` before returning both
results and a nil control. Its remaining branches preserve zero division,
minimum-signed divided by -1, Euclidean negative remainders, and result bounds.

## Preserved timing contract

One binary per application was built and fingerprinted before any timing.
Two independent cohorts reused those exact binaries in forward/reverse order,
with five processes per application per cohort. All 40 outputs passed their
public Ruby verifiers. Every application used serial execution pinned to CPU 0,
the canonical external stdlib, catalog inputs and working directory, and a
55-second process cap.

| Application | Source SHA-256 | Binary SHA-256 |
| --- | --- | --- |
| Sudoku Masks | `88294708698dd72bd6ac6a6249633cc7fddf4274a33587930f8e932b00b199a5` | `6bfa7597447a9ce4483ba11752926cc3da71600a43747cbd098c30b1da92de1a` |
| Rational Series | `20a58261d852834d425de755ed35e7c34b0bda80945caf97a94d4ba9b2e0bf46` | `8a0164bc90ae808a1d179988a464de4ef01bdc6a4b68480919041c0eed24cc80` |
| Regex Set Audit | `124fcfb1435ab1160adff38f7d50ba442c22b4d63cce25fa06e7799a26458b9b` | `473652b1ad348981f6972774228e41ed326d96f1b07b0d6ea6af609e8019bee9` |
| K-Nucleotide | `933749cb33f84a88274e010f7459d027be839a42162f0d559eca8a1920aa8a2a` | `6dd1128b7d0ea58fc6ae1f5fd6945a4453f3a5f30bf9b704ee68f3771cecba4a` |

| Application | Cohort 1 | Cohort 2 | Pooled mean | Cohort spread |
| --- | ---: | ---: | ---: | ---: |
| Sudoku Masks | 1.810 s | 1.768 s | 1.789 s | 2.38% |
| Rational Series | 0.124 s | 0.134 s | 0.129 s | 8.06% |
| Regex Set Audit | 0.102 s | 0.104 s | 0.103 s | 1.96% |
| K-Nucleotide | 2.918 s | 2.952 s | 2.935 s | 1.17% |

These cohort means are ownership controls rather than a refreshed reference
scorecard. The current strict scorecard still classifies all four compiled rows
as misses: Sudoku 3.32x Go, Rational Series 12.66x, Regex Set 21.89x, and
K-Nucleotide 62.92x.

## Main-only CPU attribution

Short workloads received more independent profile launches so the merged
profiles had useful resolution: three Sudoku, ten Rational, fifteen Regex, and
three K-Nucleotide launches. All 31 profile outputs verified. Profiles were
merged only within the same fingerprinted executable.

| Application | CPU samples | Signed helper | Material concrete ownership |
| --- | ---: | ---: | --- |
| Sudoku Masks | 5.24 s | 11.45% flat / 13.17% cumulative | `find_best_empty` 86.07% cumulative; `square_index` 33.59% |
| Rational Series | 530 ms | zero samples | `Uint128.DivMod` 22.64% flat; `Int128.DivMod` 33.96% cumulative; Rational GCD 37.74% cumulative |
| Regex Set Audit | 400 ms | zero samples | NFA closure 35.00% cumulative; NFA move 27.50%; set matching 80.00% |
| K-Nucleotide | 8.50 s | zero samples | `runtime.convT` 39.76% cumulative; `runtime.mallocgc` 35.65%; primitive HashMap equality 10.47% |

The exact leaf repeats in only one application. Rational's wide division is a
different primitive representation and operation mix. Regex's generated DFA
quotient sites are cold because the selected Regex Set workload is NFA-owned.
K-Nucleotide's quotient/remainder formatting sites execute too rarely to
receive a sample beside its map/conversion wall.

## Candidate rejection

A quotient-only helper could avoid returning or computing the unused
remainder, and a proven positive constant divisor could avoid the general
negative/overflow path. Neither observation authorizes production work here:

- only Sudoku has material CPU in the exact helper;
- the helper already has a positive-input fast path;
- current lowering has no general range proof for Sudoku coordinates, Regex
  binary-search differences, or K-Nucleotide's rounded percentage;
- implementing proof machinery for one material application would be a
  benchmark-shaped trade; and
- specializing Rational's nominal type or Regex's automata representation is
  prohibited and would not address the same primitive leaf.

No candidate was written, so alternating candidate/control timing was neither
needed nor represented as evidence.

## Verification and cleanup

The following bounded checks pass:

```text
ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_division_ops go test ./pkg/compiler -run '^TestCompilerExecFixtures$|^TestCompilerDivModConcreteCarrierStaysNative$' -count=1 -timeout 50s
just bench-catalog-check
just bench-selection-check
just bench-scoreboard-check
just bench-scorecard-evidence-check --scorecard v12/docs/perf-baselines/2026-07-20-source-equivalence-scorecard.json --selection-manifest v12/bench-selection-manifest.json --require-runs 5
```

The strict evidence check retains 65 selected rows, 72 full-status rows, and
five successful Able/reference samples for every selected row. No WASM work
was performed.

## Next recommendation

Build a current cross-mode performance-frontier ledger before selecting
another implementation candidate.

Why: the quotient census closes the latest concrete compiled hypothesis, while
many tempting broad parents—map lookup, primitive conversion, generated calls,
nullable carriers, float boxing, and VM call/return work—already have retained
or rejected gates. Selecting only by the largest ratio now risks repeating a
closed experiment or optimizing a benchmark-specific descendant.

What it entails: join all 65 selected rows with their current target ratio,
absolute wall time, source/binary freshness, dominant exact CPU/allocation
leaves, breadth across unlike applications, and the disposition of prior
candidates. Rank only unclosed leaves by target gap, material wall ownership,
and application breadth. Refresh profiles solely for stale fingerprints among
the leading rows, then choose the first compiler or bytecode leaf that is both
material and repeated in at least three unlike applications. This makes the
next code tranche evidence-driven without reopening closed Map, Array, regex,
numeric-carrier, call-frame, or workload-specific designs.
