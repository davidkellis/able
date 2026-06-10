# Current strict compiled scorecard and owner closure

Date: 2026-07-26

## Decision

Retain no compiler, generated-runtime, runtime, interpreter, bytecode VM,
canonical-stdlib, benchmark, language, dependency, or WASM change from this
tranche.

All 61 coverage applications rebuilt with `--no-fallbacks`, all 305 timed
Able executions and all 305 equivalent-Go executions passed their public
verifiers, and every final application dependency graph omits
`able/interpreter-go/pkg/interpreter`.

The refreshed scorecard has 8 target passes and 53 misses. Current main-only
CPU and exact allocation evidence for the largest absolute misses plus native
parity controls does not contain one concrete owner material in three unlike
applications. The dominant work remains separated among already-native
application computation, runtime-backed generic map semantics, application
search/allocation, required concurrency services, and native Go arithmetic or
allocation. The three-unlike admission gate therefore fails before a
prototype is justified.

The compact machine-readable companion is
`2026-07-26-current-strict-compiled-scorecard-owner-closure.json`.

## Source and measurement contract

The repository was at commit
`237406eccdfb025a519d898daedadee1c8d13a7b`. The canonical external stdlib
was the dirty 70-source tree at commit
`219eff222c28406487231713753641bc49ee5b9a`; its source-tree SHA-256 was
`6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`.
No stdlib file was changed.

Fresh equivalent-Go references and strict Able rows each used five
independent verified processes. CPU affinity rotated through CPUs 4, 6, 5,
and 11 because unrelated host work made any single CPU an unstable choice.
The scorecard retains the process means rather than selecting favorable
samples. Builds used a disk-backed `/var/tmp` workspace.

The frozen compiler used for all 61 rows had SHA-256
`28697a5adf4f73918f3d83fbcddc211407dc7e539240f64e4127a4e3dd4ddcab`.
Raw evidence was intentionally kept out of the repository because the
comparison report exceeds the 1,000-line file limit:

| Evidence | SHA-256 |
| --- | --- |
| five-run Go reference report | `37970ea944f5df73f1407702c0b3e2bc58d97f6f30f018f084bade40bac1c994` |
| five-run strict comparison report | `6158777e5714c80f9629f0dcf3493da3b792ea36c10e5b83b7d7299db3c0ed9c` |
| final dependency audit | `41053c85872224e16a9d09a74b431822de332d129b877d29efc700b50e03548b` |
| canonical-stdlib source state | `d8126a7ee93174b48c3c597839bac1620f3ab7b6babe10e23eb0f95c1d3bc28b` |

## Dependency result

Every one of the 61 final graphs omits `pkg/interpreter`. Sixty graphs contain
96 packages; Base64 contains 119 because its ordinary dependency surface is
larger. No row required classification as a dynamic or runtime-service
interpreter boundary.

This verifies the immediate architectural requirement: ordinary static
compiled applications do not cross into the tree-walking interpreter.

## Refreshed scorecard

The target is `Able / Go <= 1.052632`. The corpus geometric-mean ratio is
6.2612x and the median is 7.5000x. The ratio range is 0.4629x to 217.2727x.
Positive time above the 95%-of-Go target totals 8.3233 seconds.

| Application | Able mean | Go mean | Ratio | Target |
| --- | ---: | ---: | ---: | --- |
| JSON | 0.6660s | 1.4389s | 0.4629x | pass |
| Monte Carlo Pi | 0.1420s | 0.1960s | 0.7245x | pass |
| Quicksort | 1.8580s | 2.5475s | 0.7293x | pass |
| Base64 | 2.1340s | 2.4590s | 0.8678x | pass |
| Pidigits | 1.1660s | 1.2496s | 0.9331x | pass |
| Binary Trees | 10.0780s | 10.2731s | 0.9810x | pass |
| I Before E | 0.0640s | 0.0648s | 0.9877x | pass |
| Matrix Multiply | 0.9660s | 0.9540s | 1.0126x | pass |
| Fib | 3.7580s | 3.1652s | 1.1873x | miss |
| Mandelbrot | 0.0800s | 0.0526s | 1.5209x | miss |
| Tapelang Alphabet | 3.6380s | 1.9733s | 1.8436x | miss |
| N-Body | 0.0820s | 0.0354s | 2.3164x | miss |
| Distance Field | 0.0300s | 0.0117s | 2.5641x | miss |
| Wide Integer Records | 0.0720s | 0.0252s | 2.8571x | miss |
| Reverse Complement | 0.0520s | 0.0181s | 2.8729x | miss |
| RMS Norm | 0.0300s | 0.0103s | 2.9126x | miss |
| Sudoku Masks | 1.7480s | 0.5631s | 3.1042x | miss |
| Fasta Generation | 0.0440s | 0.0133s | 3.3083x | miss |
| Rational Series | 0.0600s | 0.0128s | 4.6875x | miss |
| Dependency Plan | 0.0200s | 0.0036s | 5.5556x | miss |
| Concurrent Policy Callbacks | 0.0300s | 0.0050s | 6.0000x | miss |
| Concurrent Signal Dispatch | 0.0320s | 0.0051s | 6.2745x | miss |
| Channel Rollup | 0.0400s | 0.0063s | 6.3492x | miss |
| Array Slice Window | 0.0300s | 0.0047s | 6.3830x | miss |
| Concurrent Event Routing | 0.0340s | 0.0053s | 6.4151x | miss |
| Concurrent Text Index | 0.0400s | 0.0060s | 6.6667x | miss |
| Concurrent Document Pipeline | 0.0300s | 0.0045s | 6.6667x | miss |
| Concurrent Transform Chain | 0.0320s | 0.0048s | 6.6667x | miss |
| Future Pipeline | 0.0400s | 0.0057s | 7.0175x | miss |
| Concurrent Packet Codecs | 0.0300s | 0.0041s | 7.3171x | miss |
| Dependency Wave Validation | 0.0360s | 0.0048s | 7.5000x | miss |
| Concurrent Graph Visitors | 0.0320s | 0.0041s | 7.8049x | miss |
| Concurrent Stencil Reduction | 0.0400s | 0.0051s | 7.8431x | miss |
| Concurrent Tree Folds | 0.0300s | 0.0038s | 7.8947x | miss |
| Concurrent Audio Voices | 0.0380s | 0.0048s | 7.9167x | miss |
| Concurrent State Machines | 0.0300s | 0.0037s | 8.1081x | miss |
| Document Audit | 0.0360s | 0.0040s | 9.0000x | miss |
| Concurrent Stateful Pipeline | 0.0440s | 0.0047s | 9.3617x | miss |
| Concurrent Scene Tiles | 0.0360s | 0.0036s | 10.0000x | miss |
| Sensor Calibration | 0.0460s | 0.0046s | 10.0000x | miss |
| Word Frequency | 0.0520s | 0.0051s | 10.1961x | miss |
| Future Await Race | 0.0500s | 0.0047s | 10.6383x | miss |
| Manifest Normalization | 0.0460s | 0.0043s | 10.6977x | miss |
| Regex Stream Audit | 0.0520s | 0.0048s | 10.8333x | miss |
| Unicode Scalar Pipeline | 0.1120s | 0.0101s | 11.0891x | miss |
| Log Routing Redaction | 0.0640s | 0.0054s | 11.8519x | miss |
| Option Result Config | 0.0460s | 0.0038s | 12.1053x | miss |
| Regex Suffix Audit | 0.0600s | 0.0049s | 12.2449x | miss |
| Config Validation Extraction | 0.0480s | 0.0039s | 12.3077x | miss |
| Lexical Rollup | 0.0520s | 0.0041s | 12.6829x | miss |
| Regex Set Audit | 0.0720s | 0.0052s | 13.8462x | miss |
| Mutex Ledger | 0.0880s | 0.0056s | 15.7143x | miss |
| Binary Event Log | 0.1480s | 0.0079s | 18.7342x | miss |
| Validated Job Pipeline | 0.0760s | 0.0039s | 19.4872x | miss |
| Inventory Reconciliation | 0.1640s | 0.0082s | 20.0000x | miss |
| Policy Record Dispatch | 0.1040s | 0.0047s | 22.1277x | miss |
| Fixed Width 128 | 0.1500s | 0.0059s | 25.4237x | miss |
| K-Nucleotide | 1.5880s | 0.0587s | 27.0528x | miss |
| Await Channel Mux | 0.2480s | 0.0057s | 43.5088x | miss |
| Mutex Await Journal | 0.2960s | 0.0047s | 62.9787x | miss |
| Mutex Work Queue | 0.9560s | 0.0044s | 217.2727x | miss |

The short-row ratios are dominated by a fixed launch/service floor and should
not be interpreted as a shared lowering owner. The largest absolute target
excess instead comes from Tapelang Alphabet, K-Nucleotide, Sudoku Masks,
Mutex Work Queue, and Fib.

## Current owner refresh

Tapelang Alphabet, K-Nucleotide, Sudoku Masks, Mutex Work Queue, Fib, Pidigits,
and Binary Trees each received three new independent verified main-only CPU
profiles and three independent exact main-phase allocation-counter runs.
One exact allocation profile was also captured for each row except Binary
Trees, whose prior exact profile serialization exceeded the process guard.
Binary Trees retained lightweight counters instead.

| Application | Merged CPU | Dominant exact CPU owner |
| --- | ---: | --- |
| Tapelang Alphabet | 14.36s | `execute` 65.32% flat; `Tape.inc` 26.67%; `Tape.get` 6.27% |
| K-Nucleotide | 4.70s | map key equality 17.66%; map hash 6.81%; entry search 4.89%; allocation/GC |
| Sudoku Masks | 5.34s | `find_best_empty` 32.58%; checked multiply 13.30%; `bit_count` 13.11%; signed divmod 12.73% |
| Mutex Work Queue | 5.24s | runtime lock/spin/futex work beneath the already-closed goroutine-identity service |
| Fib | 10.76s | generated native `fib` 99.72% |
| Pidigits | 3.22s | native `math/big`, led by `mulAddVWW` 43.17% and shifts |
| Binary Trees | 118.36s | native allocation/GC beneath generated `make_tree` |

The new evidence reproduces the earlier owner separation. In particular, Fib
is already a native recursive Go function with no application allocation,
Pidigits is already native `math/big`, and Binary Trees is already within the
Go target while doing native allocation and collection.

## Exact allocation result

| Application | Mean allocated bytes | Mean allocations | Mean GC |
| --- | ---: | ---: | ---: |
| Tapelang Alphabet | 6,160 | 48 | 0.00 |
| K-Nucleotide | 614,185,136 | 16,232,474 | 283.33 |
| Sudoku Masks | 156,390,125 | 7,802,766 | 132.67 |
| Mutex Work Queue | 16,598,485 | 297,440 | 12.67 |
| Fib | 144 | 6 | 0.00 |
| Pidigits | 298,930,045 | 25,222 | 274.00 |
| Binary Trees | 9,820,314,523 | 613,767,018 | 205.67 |

Exact application allocation attribution is likewise disjoint:

- Tapelang and Fib have no material application allocation.
- K-Nucleotide allocates primarily in primitive conversions at its
  runtime-backed generic map boundary.
- Sudoku allocates almost entirely in its generated `find_best_empty`
  application body.
- Mutex Work Queue divides allocation among the goroutine-identity,
  callable, awaitable, and waker services.
- Pidigits allocates native `math/big` storage.
- Binary Trees performs the expected native tree allocation.

No exact CPU or allocation leaf is material in three unlike target misses.
Grouping these semantically different descendants under labels such as
“generated runtime,” “allocation,” or “boxing” would not satisfy the
generality rule.

## Closed-route reconciliation

The result is consistent with the current closure ledger:

- strict applications already omit accidental interpreter and GC ballast;
- fixed launch and generated registration cost has been measured and closed;
- checked arithmetic prototypes had mixed broad A/B results;
- a broad execution-context ABI for goroutine identity regressed controls;
- the corpus-wide compiled owner census already retained the only admissible
  shared primitive bridge leaf, the lazy common-`i32` cache;
- the seven-program runtime-value boundary census found only a required
  escaping semantic nominal value at an explicit boundary;
- named-container and non-primitive nominal special cases remain forbidden.

There is therefore no admissible current-profile candidate to advance. The
correct result is no code.

## Verification

The complete `./run_all_tests.sh` handoff passed:

- every coverage, external-scoreboard, selection, and threshold-control
  contract;
- every non-compiler Go package;
- all 32 bounded compiler batches; and
- the final bytecode fixture corpus in 83.905 seconds.

The known large compiler aggregates included batch 19 at 180.092 seconds,
batch 28 at 73.809 seconds, and batch 29 at 85.040 seconds. They are existing
multi-test sharding debt; this tranche adds no test or production source file.

## Next recommendation

Audit semantic-work equivalence for stable compiled misses that are already
pure native Go: begin with Tapelang Alphabet, Fib, and Sudoku Masks, with
Matrix Multiply and Pidigits as parity controls.

Why: interpreter-free graphs and native carriers are now proven, and the
largest current profiles do not share a local compiler/runtime owner. The
remaining question is whether generated Go performs redundant compiler-added
work or whether the Able and handwritten Go benchmark implementations perform
different algorithmic or language-required work.

What it entails: compare generated Go with reference Go at source and assembly
level; count loop iterations, calls, bounds/overflow checks, allocations, and
semantic adapters; reconcile input and output contracts; and admit a change
only if one proof-backed redundant compiler operation repeats in at least
three unlike applications. Do not retry named-container, checked-arithmetic,
fixed-context ABI, or benchmark-specific routes without new evidence.

Why it is important: profile optimization has reached native application
functions rather than a shared boundary. Establishing work equivalence is now
the shortest path to knowing whether Go parity requires a general compiler
proof/elision rule or correction of non-equivalent benchmark work. It prevents
optimizing irreducibly different programs while preserving the goal of native
Go performance. Do not begin WASM work.
