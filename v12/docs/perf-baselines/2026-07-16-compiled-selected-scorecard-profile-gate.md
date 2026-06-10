# Compiled selected-program scorecard and profile gate

Date: 2026-07-16

## Decision

Keep no compiler, generated-runtime, canonical-stdlib, application, or reference
code from this tranche. The current canonical stdlib tree had drifted since the
morning promoted scorecard, so all 32 selected compiled Able applications and all
32 equivalent Go references received five new independent verifier-backed runs.
All 320 timed processes completed and every verifier accepted every output. Each
compiled Able application retained one stable output hash; the intentionally random
Go Monte Carlo reference retained five accepted hashes and is labeled
`verified-nondeterministic`.

Only 3 of 32 compiled rows currently meet the target of running no more than
`1 / 0.95` times as long as Go: QuickSort, JSON, and Base64. Monte Carlo Pi is
just outside at 1.054x. Bounded CPU and allocation profiles of the six largest
absolute misses do not identify one measured-loop owner across three unlike
programs, so no application-loop candidate was built.

## Measurement state

The reviewed compiled selection still contains 32 portable applications. The
catalog and selection checks passed before timing. Every process used canonical
`../able-stdlib/src`, `GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, a 60-second
execution cap, and the catalog's declared run directory, arguments, setup, executor,
and verifier. Workstation evidence uses five independent processes and arithmetic
means; individual samples remain in the JSON reports for median/spread analysis.

The measured canonical stdlib state is 69 files with source-tree SHA-256
`353d27460c9d1d1dcfeb8bd8a92b75de3fd6b89a8966b4cf9ac2c4a2f211ca7b` at Git
head `219eff222c28406487231713753641bc49ee5b9a`, dirty. The earlier promoted
scorecard recorded `44a1adea...`, so its compiled rows were not reused. The mixed
compiled/bytecode current scoreboard was intentionally not promoted from this
compiled-only refresh, because doing so would relabel bytecode rows measured against
the older stdlib tree.

Primary artifacts:

- `2026-07-16-compiled-selected-stdlib-source-state.json`
- `2026-07-16-compiled-selected-go-reference.{json,md}`
- `2026-07-16-compiled-selected-comparison.{json,md}`
- `2026-07-16-compiled-selected-variance.{json,md}`
- `2026-07-16-compiled-fixed-width-repeat.{json,md}`

## Refreshed ranking

The largest absolute Able-minus-Go gaps are:

| Application | Compiled Able | Go | Ratio | Absolute excess |
| --- | ---: | ---: | ---: | ---: |
| Binary Trees | 27.8140 s | 5.2735 s | 5.27x | 22.5405 s |
| Fixed Width 128 | 9.2880 s | 0.0049 s | 1895.51x | 9.2831 s |
| Sudoku Masks | 8.7580 s | 0.5292 s | 16.55x | 8.2288 s |
| K-Nucleotide | 3.5880 s | 0.0514 s | 69.81x | 3.5366 s |
| Rational Series | 2.6260 s | 0.0118 s | 222.54x | 2.6142 s |
| TapeLang Alphabet | 3.2940 s | 1.8126 s | 1.82x | 1.4814 s |
| Channel Rollup | 1.1940 s | 0.0047 s | 254.04x | 1.1893 s |
| Regex Suffix Audit | 1.2020 s | 0.0306 s | 39.28x | 1.1714 s |

The competitive edge is narrow but real:

| Application | Compiled Able | Go | Ratio | Status |
| --- | ---: | ---: | ---: | --- |
| QuickSort | 1.7400 s | 2.3459 s | 0.74x | meets |
| JSON | 0.6980 s | 1.3144 s | 0.53x | meets |
| Base64 | 2.3260 s | 2.2857 s | 1.018x | meets |
| Monte Carlo Pi | 0.2000 s | 0.1898 s | 1.054x | just misses |
| Fib | 3.4020 s | 2.9384 s | 1.158x | misses |
| PiDigits | 1.2800 s | 1.1118 s | 1.151x | misses |

The current stdlib materially improves the regex family versus the morning
scorecard: Suffix falls 53.66%, Set 43.16%, and Stream 23.08%. Word Frequency,
I-Before-E, and Lexical Rollup also improve 7.84-9.26%. Those shifts validate the
decision to refresh after stdlib drift instead of treating the old ratios as current.

## Workstation variance

Most long rows have useful five-sample spread: Binary Trees CV is 3.88%, Sudoku
1.94%, K-Nucleotide 6.21%, Rational 5.63%, and TapeLang 2.20%. Short rows such as
Array Slice Window and Document Audit have high percentage CV because the timer is
quantized at hundredths of a second; their ratio gaps are too large for the noise to
change classification.

Fixed Width 128's first five samples had 17.97% CV, caused by 11.80- and 10.22-second
outliers around three 8.04-8.20-second samples. A second independent verifier-backed
five-run batch averaged 8.1060 seconds. Across all ten samples, the mean is 8.6970
seconds, median 8.1900, range 4.03, and CV 14.03%. Its exact numeric value is noisy,
but its roughly 1,775x combined-mean Go ratio and top-three absolute ranking are not.

## Bounded profile gate

One verifier-backed whole-process CPU and allocation profile was collected for each
of the six largest unlike absolute misses. The profiler changed timing but not output;
profile timings do not replace scorecard means.

| Application | CPU/application owner | Allocation owner |
| --- | --- | --- |
| Binary Trees | allocation and GC beneath generated `make_tree` | `make_tree`, 99.58% of space |
| Fixed Width 128 | bridge/bitwise/big-integer arithmetic and GC | `math/big`, `ToUint`, AST type construction, bitwise evaluation |
| Sudoku Masks | generated `find_best_empty`, slice growth, allocation | `find_best_empty`, 98.73% of space |
| K-Nucleotide | HashMap key equality/hash, conversion, string work, GC | `ToInt`, String conversion, `ToUint`, byte arrays |
| Rational Series | big-integer scanning, direct integer bridge/evaluation | distinct GCD/absolute/build closures, `ToInt`, type construction |
| TapeLang Alphabet | generated tape execution/methods, 99.3% CPU | measured loop is nearly allocation-free |

`runtime.tryDeferToSpanScan` is at least 1% flat in five profiles, with nearby Go
allocation/GC helpers recurring as well. Allocation-owner profiles immediately split
that parent into the six different owners above. This reproduces the prior mixed
profile gate rather than revealing a new compiler semantic operation.

`bridge.ToInt`/`ToUint` occur in Fixed Width, K-Nucleotide, and Rational, but their
consumers differ: public UInt128 checked arithmetic, map/string primitive boundaries,
and nominal rational construction. The generic `ToUint` branch-layout trial and raw
value/boxing substitutions have already failed broad performance or semantic gates;
the fresh profiles do not supply a common generated source boundary that would reopen
them.

The only exact allocation symbol above 1% in three profiles is
`initBytecodeSmallIntBoxCache`, seen in K-Nucleotide, Rational, and TapeLang. It is
Go package initialization before the generated application loop, not their shared
algorithm. Unlike the application owners, however, it is product-visible in complete
compiled process time and allocates roughly 25-35 MiB even when no bytecode VM runs.
The CPU profiler starts after package initialization, so these profiles cannot measure
its startup time. It therefore warrants a separate startup-isolation gate, not an
immediate claim or an application-specific change.

## Verification and cleanup

- Portable coverage catalog: 32 applications, complete.
- Selection manifest: 32 compiled and 26 bytecode rows, valid.
- Fresh Go references: 160/160 verifier-backed processes passed.
- Fresh compiled Able: 160/160 verifier-backed processes passed.
- Fixed Width repeat: 5/5 verifier-backed processes passed.
- Profile runs: 6/6 verifier-backed processes passed and flushed CPU/heap data.
- No source or promoted-scoreboard behavior changed.
- All temporary generated binaries, profile workspaces, and comparison workspaces are
  removed after extracting this record.

## Next recommendation

Isolate and test lazy initialization of the bytecode-only fixed small-integer cache for
compiled application processes.

Why: the cache is the only exact allocation owner repeated across three unlike fresh
compiled profiles, and full application binaries currently pay its roughly 25-35 MiB
startup allocation even when they never construct a bytecode VM. This may explain a
material part of the common 60-100 ms floor in short compiled applications. It is a
generic runtime-packaging issue affecting every compiled binary, not a benchmark or
nominal-container special case. Existing profiles cannot decide it because Go package
initialization precedes CPU-profile startup.

What it entails: move cache construction behind a once-only bytecode-runtime boundary,
prove that every direct cache reader is initialized before VM execution, and retain
the current behavior after `NewBytecode` construction. Measure at least five alternating
baseline/candidate complete-process runs for three unlike short compiled applications
(for example Array Slice Window, Document Audit, and Dependency Plan), plus long
compiled controls and bytecode startup/runtime guards. Keep it only if compiled startup
and memory improve broadly without moving cost into bytecode hot paths or changing
semantics. Otherwise revert it and close the startup hypothesis. Do not begin WASM work.
