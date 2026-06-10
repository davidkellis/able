# Current strict compiled scorecard and owner closure

Date: 2026-07-28

## Decision

Retain no production change from this tranche.

All 63 current compiled applications rebuilt under the strict
fallback-free contract. All 315 Able executions and all 315 fresh
equivalent-Go executions passed their public verifiers, and every generated
dependency graph omits `able/interpreter-go/pkg/interpreter`.

Fifteen measured-main CPU profiles and three exact measured-main allocation
profiles for each of ten poorly explained misses expose no open compiler,
generated-runtime, or semantic-boundary owner that is material in at least
three unlike applications. The repeated owners are already-closed checked
arithmetic or boxing routes; same-family UTF-8/regex work; required generic
map, nominal, and concurrency semantics; or already-native application and Go
runtime work. The admission gate therefore stops before a prototype.

No compiler, generated-runtime, runtime, interpreter, bytecode VM,
canonical-stdlib, benchmark, language, dependency, non-primitive nominal, or
WASM production source changed.

## Source and measurement contract

The repository was at commit
`cbba634af6c5766e5b8aa162e8e66643654f620e` with the long-running dirty
worktree preserved. The canonical external stdlib was the dirty 70-source
tree at commit `219eff222c28406487231713753641bc49ee5b9a`; its current ledger
tree SHA-256 was
`382d256e2fb380220dcdd62a5cf83109fa72297f23d70bdd1ffe2d8daebed047`.
No stdlib file changed in this tranche.

Every Able application was freshly generated and built with:

```text
--no-fallbacks --experimental-execution-context
GOMEMLIMIT=1GiB GOGC=50
```

The frozen `ablec` binary used for all 63 rows had SHA-256
`55c7e9cb6f911406510b67a9f09ccece55cc4cb9a111cf8ceebb234adfc13871`.
The catalog supplied each row's logical CPU budget and executor policy.
Affinity rotated through the declared pool `4,6,5,11`. Able and Go each used
five independent processes per application; process means were retained
without selecting favorable samples. All large work stayed in disk-backed
`/var/tmp`.

The compact row-level evidence, including all samples, source/verifier/output
hashes, and execution contracts, is:

- `2026-07-28-current-strict-compiled-scorecard-owner-closure.tsv`
  (`77f1f29ee765d960881dfa93404139e873d8afc8fd3488c4d3d8e305d287e552`)

The intentionally unretained raw reports were larger than the repository's
1,000-line file limit:

| Evidence | SHA-256 |
| --- | --- |
| five-run Go reference JSON | `b3f5fff7392ddad7007e137637967a7ba06748b1921dd1114b7a8899a3bcc6d3` |
| five-run Go reference Markdown | `e9e62777f2dded8eeeb98a493d31666013d73e049c39c20d03c6e49b43a717f3` |
| five-run strict comparison JSON | `7d2b7f555b71387925945ab83c49bd2a9a1aca9dee3ce45bddbf985283eaa061` |
| five-run strict comparison Markdown | `3b5bfde0de31803b44442e548a139dcab7c1a309bcc819d2e2338903a39a3257` |

## Dependency result

All 63 final `go list -deps` graphs are interpreter-free. The preserved audit
is
`2026-07-28-current-strict-compiled-scorecard-owner-closure-dependency-audit.tsv`
with SHA-256
`388bfe8eb7fcd6b4177549ef2dbba80b1c9b2ab16474103a22a26231db1e6474`.

This establishes the architectural point directly: none of the measured
static programs crosses into the tree-walking interpreter. A remaining
performance gap cannot be attributed to accidental compiled/interpreted
execution.

## Refreshed scorecard

The target is `Able / Go <= 1.052632`. Nine of 63 rows pass and 54 miss. The
geometric-mean ratio is `5.2579x`, the median is `6.9231x`, and positive time
above the target totals `5.4254s`.

The prior 2026-07-26 refresh had 61 rows, 8 passes, a `6.2612x` geometric
mean, a `7.5000x` median, and `8.3233s` positive target excess. These are
descriptive current-state comparisons, not a controlled A/B claim: the
catalog grew by two rows, retained code changed between dates, and the host
is noisy.

Current target passes are:

| Application | Able mean | Go mean | Ratio |
| --- | ---: | ---: | ---: |
| JSON | 0.6500s | 1.5554s | 0.4179x |
| Quicksort | 1.8780s | 2.6639s | 0.7050x |
| Monte Carlo Pi | 0.1860s | 0.1984s | 0.9375x |
| I Before E | 0.0580s | 0.0602s | 0.9635x |
| Binary Trees | 10.9700s | 11.3207s | 0.9690x |
| Base64 | 2.5040s | 2.5167s | 0.9950x |
| Pidigits | 1.2200s | 1.1864s | 1.0283x |
| Fib | 3.3620s | 3.2505s | 1.0343x |
| Matrix Multiply | 1.0340s | 0.9870s | 1.0476x |

Fib is only 1.83 percentage points inside the threshold and is not promoted
as a stable target guard from this noisy refresh alone.

The largest absolute target excesses are:

| Application | Able mean | Go mean | Ratio | Target excess |
| --- | ---: | ---: | ---: | ---: |
| K-Nucleotide | 1.6020s | 0.0566s | 28.3039x | 1.5424s |
| Sudoku Masks | 1.5620s | 0.6740s | 2.3175x | 0.8525s |
| TapeLang Alphabet | 3.9660s | 3.0651s | 1.2939x | 0.7396s |
| Binary Event Log | 0.1620s | 0.0079s | 20.5063x | 0.1537s |
| Inventory Reconciliation | 0.1320s | 0.0085s | 15.5294x | 0.1231s |
| Await Channel Mux | 0.1280s | 0.0048s | 26.6667x | 0.1229s |
| Unicode Scalar Pipeline | 0.1120s | 0.0124s | 9.0323x | 0.0989s |
| Fixed Width 128 | 0.1040s | 0.0062s | 16.7742x | 0.0975s |
| Policy Record Dispatch | 0.0860s | 0.0066s | 13.0303x | 0.0791s |
| Regex Suffix Audit | 0.0780s | 0.0050s | 15.6000x | 0.0727s |

K-Nucleotide remains the closed generic-map family. The recently refreshed
Await rows retain the already-closed post-materialization owner
classification. Sudoku and TapeLang are already dominated by generated
native application bodies rather than a compiled/interpreted boundary.

## Current owner refresh

The ten selected applications received 15 independent measured-main CPU
profiles and three independent exact measured-main allocation snapshots
each. Every one of the 180 profile outputs matches the corresponding
scorecard output hash. The output audit has SHA-256
`d5e154b68f427aea757e46d68671fa5fb1f4a96a69a64c84014baf5d6e277a6a`.

Representative CPU and allocation tops are preserved under
`2026-07-28-current-strict-compiled-scorecard-owner-closure-profiles/`; its
40-file manifest has SHA-256
`3a6ae08c94ef9cb0f3263b720fb37317e04130378b65a765c02f54c9ef2d620e`.
The three-run exact allocation summary has SHA-256
`7224104387ecc52e23c6e0a2cbd7eea57c4a7e464f47dff4feebfa8f5cf168cb`.

| Application | CPU samples | Mean bytes / objects | Dominant owner and disposition |
| --- | ---: | ---: | --- |
| Binary Event Log | 1.58s | 62,449,152 / 1,073,787 | checked arithmetic plus `ToInt` and nominal record allocation; checked arithmetic and ordinary boxing are closed, record specialization is forbidden |
| Inventory Reconciliation | 1.30s | 17,037,112 / 553,060 | generic map hash/equality plus retained dynamic-`i64` and nullable materialization; named-map treatment is forbidden |
| Unicode Scalar Pipeline | 1.18s | 24,807,437 / 2,629,749 | UTF-8 decode result and rune/end union materialization; confined to the already-reviewed text traversal family |
| Fixed Width 128 | 0.85s | 35,536,440 / 2,220,987 | checked arithmetic and application nominal `UInt128` construction; nominal specialization is forbidden |
| Policy Record Dispatch | 0.64s | 20,607,808 / 447,310 | canonical regex NFA closure/move/thread/capture work |
| Validated Job Pipeline | 1.20s | 1,831,792 / 32,936 | Go lock/spin plus already-closed `currentGID`, `ToInt`, and nominal task/result materialization |
| Rational Series | 0.47s | 672 / 21 | native `Uint128.DivMod` and `Int128` arithmetic |
| Wide Integer Records | 0.57s | 10,738,269 / 640,056 | native 128-bit arithmetic plus application parsing and nominal construction |
| Regex Suffix Audit | 0.22s | 5,230,600 / 154,461 | canonical regex NFA thread/capture work; same family as Policy |
| Word Frequency | 0.11s | 3,095,536 / 56,476 | generic map and string materialization; no material UTF-8 result-union owner |

The CPU samples for the shortest applications are sparse; their stable exact
allocation snapshots provide the stronger selection signal.

## Breadth and closure reconciliation

No apparent shared label survives exact-owner and admissibility review:

- Checked multiply/add appears in Binary Event Log, Fixed Width 128, and
  Unicode Scalar Pipeline, but the broad checked-arithmetic candidates are
  already closed after mixed A/B results.
- Ordinary `bridge.ToInt` materialization is substantial in Binary Event Log
  and present in Validated Job Pipeline. The global ordinary-`i64` cache is
  already rejected because it regressed the allocation-light TapeLang guard;
  Word Frequency has only one such conversion.
- UTF-8 decode/result materialization appears in Unicode Scalar Pipeline and
  two regex applications. The primitive String decode lowering is already
  retained, while the residual result-union representation previously failed
  the generality gate outside String traversal. Two regex consumers do not
  supply a third unlike semantic family.
- Generic map hash/equality repeats across map applications, but the current
  map boundary is language-semantic and named-container compiler rules are
  prohibited.
- Nominal construction repeats only as required semantic values with
  different type-specific descendants. The shared nominal pipeline and
  positional boundary are already closed; specializing `EventRecord`,
  `UInt128`, or any other nominal is prohibited.
- Go allocation/GC is common only as an umbrella. Its application descendants
  are disjoint and cannot justify one compiler rule.
- Native 128-bit arithmetic, regex NFA execution, parsing, and concurrency
  services are separate families, not one boundary.

The largest broad rows therefore provide no new evidence to reopen accidental
interpreter/GC ballast, a broad execution-context ABI, checked arithmetic,
named containers, or non-primitive nominal lowering. The correct result is no
code.

## Verification

- 315/315 strict Able executions passed their public verifiers with no timeout
  or failure.
- 315/315 equivalent-Go executions passed; Monte Carlo Pi used its
  nondeterministic public verifier.
- 180/180 focused CPU/allocation outputs match the expected verifier-backed
  scorecard hashes.
- 63/63 final generated dependency graphs omit `pkg/interpreter`.
- The selection manifest, focused compiler/runtime tests, `cmd/ablec`, and the
  21-closure performance ledger pass.
- The compiler-production ledger hash remains
  `8d69533a81ce44f58ec8921abbd9867cbeb935aeab3ec9b39312f10aee1f7433`.
- After the compact evidence and hashes were verified, the exact disposable
  5.8 GiB `/var/tmp/able-compiled-broad-refresh-20260728-lb0k6j` workspace and
  its pointer were deleted.

## Next recommendation

Perform a semantic-work and generated-assembly equivalence audit for
TapeLang Alphabet, Sudoku Masks, and N-Body, using Matrix Multiply and
Pidigits as near-parity controls.

Why: the broad refresh proves that static applications stay entirely in
compiled Go, while the refreshed profiles find no open shared boxing or
runtime-boundary owner. The largest sustained unexplained gaps now sit inside
generated native application functions.

What it entails: compare Able source, generated Go, reference Go, and compiler
assembly; count loop iterations, calls, bounds and overflow checks,
allocations, and semantic adapters; reconcile input/output and algorithmic
work; and admit a candidate only if one proof-backed redundant generated
operation repeats in at least three unlike applications and survives balanced
verifier-backed A/B timing.

Why it is important: reaching native Go performance now depends on
distinguishing removable compiler-added work from language-required or
algorithmically different work. This audit is the shortest path to a general
proof/elision rule without reopening closed boundaries or introducing a
benchmark-specific optimization. Do not begin WASM work.
