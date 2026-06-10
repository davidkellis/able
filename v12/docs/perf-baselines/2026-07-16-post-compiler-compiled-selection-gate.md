# Post-compiler compiled selection gate

Date: 2026-07-16

## Decision

Retain two benchmark-infrastructure correctness fixes and no compiler,
generated-runtime, interpreter, stdlib, application, verifier, or reference
implementation change. The feature-coverage assertion now expects the current
34 portable applications, and the complete scorecard refresh partitions now
include Distance Field and RMS Norm in their own bounded generality group.

Fresh verifier-backed measurements show that 6 of 34 compiled Able applications
currently meet the project target of taking no more than `1 / 0.95` times their
equivalent Go implementation: Base64, JSON, Matrix Multiply, Monte Carlo Pi,
PiDigits, and QuickSort. The largest remaining absolute misses still divide
below generic Go allocation/GC parents into different Able operations, so this
tranche does not construct a workload-specific candidate.

## Contract repair

`bench_catalog_check` correctly reported 34 portable applications after
Distance Field and RMS Norm landed, but its fast test still asserted 32. The
full refresh script's dynamically read generality suite also contained both new
programs, while its bounded group list omitted them. That omission made the
manifest-driven dry-run test prove that four required benchmark/mode rows could
never run in a complete refresh.

The retained repairs are:

- update the coverage test's expected portable count from 32 to 34;
- add `distance_field,rms_norm` to the refresh's generality groups and update
  its help text from 15 to 17 generality applications.

The catalog, feature-coverage tests, reviewed 62-row mode selection, refresh
partition tests, and checked-in scoreboard validation all pass after the fix.

## Measurement method

The reviewed compiled selection contains all 34 portable applications. Fresh
Go references and current compiled Able binaries each received five independent
processes, for 170 Go and 170 Able executions. Every successful stdout was
checked by the benchmark's public verifier. All processes used `GOMAXPROCS=1`;
compiled Able used the standard `GOMEMLIMIT=1GiB` and `GOGC=50` guard. The
canonical stdlib source state is 69 files at tree SHA-256
`f37de0ac91abf02ab7c2af47e66cc06c9a37b9e32d618f4b12aee6ff11587f1d`
and Git head `219eff222c28406487231713753641bc49ee5b9a`, dirty.

All 340 executions passed their verifiers with no timeout or failure. Individual
samples, output hashes, reference provenance, ratios, and sample-spread metrics
are retained in:

- `2026-07-16-post-compiler-go-reference.{json,md}`;
- `2026-07-16-post-compiler-compiled-selection.{json,md}`;
- `2026-07-16-post-compiler-compiled-selection-variance.{json,md}`;
- `2026-07-16-post-compiler-stdlib-source-state.json`.

## Current target state

| Application | Able | Go | Ratio | Status |
| --- | ---: | ---: | ---: | --- |
| JSON | 0.7840 s | 1.4080 s | 0.56x | meets |
| QuickSort | 2.0500 s | 3.0451 s | 0.67x | meets |
| Monte Carlo Pi | 0.2080 s | 0.2693 s | 0.77x | meets |
| PiDigits | 1.4100 s | 1.6853 s | 0.84x | meets |
| Base64 | 2.3600 s | 2.3670 s | 1.00x | meets |
| Matrix Multiply | 1.1380 s | 1.0848 s | 1.05x | meets |

The largest absolute excesses are:

| Application | Able | Go | Absolute excess | Ratio |
| --- | ---: | ---: | ---: | ---: |
| Binary Trees | 30.6420 s | 6.3222 s | 24.3198 s | 4.85x |
| Sudoku Masks | 9.3260 s | 0.5674 s | 8.7586 s | 16.44x |
| K-Nucleotide | 5.6440 s | 0.0667 s | 5.5773 s | 84.62x |
| TapeLang Alphabet | 5.2060 s | 1.9656 s | 3.2404 s | 2.65x |
| Regex Suffix Audit | 2.0920 s | 0.0414 s | 2.0506 s | 50.53x |

The retained wide-integer carrier is clearly product-visible: Fixed Width 128
is now 0.2100 seconds versus the earlier 9.2880-second scorecard row, and
Rational Series is 0.1600 seconds versus 2.6260 seconds. These cross-cohort
figures demonstrate the scale of the retained lowering change, but the broader
row deltas are not treated as causal because both the dirty stdlib state and
workstation conditions differ.

## Selection interpretation

The current top four long-running owners were already profiled after the
caller-owned result ABI changes and remain structurally distinct:

- Binary Trees is recursive identity-bearing nominal allocation and GC;
- Sudoku Masks is recursive candidate-Array construction and growth;
- K-Nucleotide is string conversion, primitive map hashing/equality, and
  integer boundaries;
- TapeLang is an allocation-light generated method/checked-mutation loop.

Regex Suffix has a separate regex thread/result construction path. Reprofiling
the same source revision would add sample depth but would not create a shared
compiler descendant. The common `mallocgc`, scan, and generated-function
parents have different Able owners, so they do not authorize a compiler rule.

The short rows do expose a different cross-family band: Array Slice Window,
Dependency Plan, Distance Field, Document Audit, Reverse Complement, and RMS
Norm all carry roughly 80-110 milliseconds of whole-process time despite unlike
main algorithms. Existing initialization attribution already assigns about
58-61 milliseconds, 38 MB, and 707k allocations to the linked interpreter
package in no-bootstrap compiled binaries. The rejected lazy fixed-integer-cache
trial proved that removing one initializer in place changes the GC heap goal and
regresses allocation-heavy Binary Trees, so that experiment must not be retried.

## Verification and cleanup

- Portable corpus: 34 applications, 35 canonical sources, one diagnostic
  source, and 78 bounded fixtures.
- Reviewed selection: 34 compiled and 28 bytecode rows.
- Go references: 170/170 verified; no timeout or failure.
- Compiled Able: 170/170 verified; no timeout or failure.
- Generated comparison workspaces were removed by the harness.
- No current scoreboard was promoted from this compiled-only refresh because
  its bytecode half was not measured against the same stdlib state.

## Next recommendation

Design and measure a generated-binary runtime-boundary split that removes the
concrete tree-walker/bytecode interpreter package from ordinary compiled-only
applications, beginning with a compile/link inventory and a minimal interface
contract rather than another lazy-global experiment.

Why: the current scorecard proves that direct generated kernels can match or
beat Go, while many unlike short applications share an approximately
80-110-millisecond process floor. Existing `inittrace` evidence assigns most of
that floor to interpreter package initialization that compiled-only programs do
not semantically need. This is broad product overhead, not an application,
container, or benchmark fast path, and it can improve many real short-running
binaries at once.

What it entails: classify the 41 bridge interpreter operations and 21-61
generated signatures from the existing dependency audit into compile-time-only,
compiled-runtime, and dynamic-fallback contracts; define the smallest runtime
interfaces/value services needed by generated binaries; keep the full
interpreter adapter only behind actual dynamic-evaluation boundaries; then
build three unlike no-bootstrap applications and one dynamic-fallback control.
Measure linked symbols, binary size, `inittrace`, RSS, allocations, and at least
five alternating complete-process runs, with Binary Trees as the GC-pacing
guard. Retain the split only if it preserves all fallback semantics, removes
the interpreter initializer rather than disguising it, improves short programs
broadly, and does not regress allocation-heavy or long-running applications.
Continue to defer WASM.
