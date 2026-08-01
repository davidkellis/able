# Nominal ownership lowering retained by default

Date: 2026-07-31

## Decision

Retain the proven caller-owned nominal-result lowering path as the ordinary
compiled default. Retain `ablec --no-nominal-ownership` as an explicit
diagnostic opt-out, and keep `--experimental-nominal-ownership` as a
compatibility spelling for callers that still pass the former opt-in.

The execution rule itself is unchanged from the retained opt-in prototype. It
is still a general structural compiler rule driven by the fail-closed
interprocedural ownership proof. It contains no application, benchmark,
container, source-family, or non-primitive nominal name check.

No runtime, interpreter, bytecode VM, stdlib, language, dependency, benchmark,
frozen-workspace, or WASM source changed.

## Default and opt-out contract

Ordinary `compiler.Options{}` and ordinary `ablec` builds now prepare and
consume the nominal-ownership execution report. The execution path still
admits only:

- a proven locally fresh nominal source;
- an unconditional same-region replacement;
- a direct fresh successor or one-field embedded fresh successor;
- complete generated native-interface implementation sets; and
- call sites without capture, storage, retained/returned alias, dynamic
  dispatch, conditional replacement, or unresolved ownership.

`compiler.Options.DisableNominalOwnership` and
`ablec --no-nominal-ownership` leave the old fresh-allocation path available
for controlled A/B diagnosis. Proof-report collection remains independently
opt-in; enabling or disabling JSON diagnostics does not alter execution.

The old `ExperimentalNominalOwnership` Go option remains source-compatible,
but no longer gates the default execution rule. Passing both CLI spellings is
an error because they request contradictory states.

## High-resolution five-application gate

An independent 120-triplet cohort rotated all six
opt-out/default/equivalent-Go orders on CPUs 12-15 with `GOMAXPROCS=4`,
`GOGC=50`, and `GOMEMLIMIT=1GiB`. All 1,800 processes produced output accepted
by the applications' public Ruby verifiers.

| Application | Opt-out | Default | Go | Default change | Default / Go |
| --- | ---: | ---: | ---: | ---: | ---: |
| Concurrent Graph Visitors | 0.006098 s | 0.005884 s | 0.001703 s | -3.51% | 3.456x |
| Concurrent Packet Codecs | 0.007062 s | 0.005708 s | 0.001553 s | -19.18% | 3.674x |
| Concurrent Tree Folds | 0.005089 s | 0.004631 s | 0.001536 s | -8.98% | 3.015x |
| Concurrent Audio Voices | 0.011983 s | 0.010925 s | 0.001808 s | -8.83% | 6.043x |
| Concurrent Scene Tiles | 0.006060 s | 0.005427 s | 0.001388 s | -10.44% | 3.911x |

A separate 60-triplet cohort improved four rows and showed a 2.59% Graph
regression. These applications run for only a few milliseconds, so launch
noise is material. The larger independent cohort resolved Graph in the
expected direction; the exact allocation counters below establish the
mechanism independently of launch timing.

The full per-process cohort is
`2026-07-31-nominal-ownership-default-balanced-timings.json`.

## Exact allocation gate

Five main-phase `MemStats` runs per default and opt-out product all passed the
same output contract.

| Application | Opt-out bytes / objects | Default bytes / objects | Change |
| --- | ---: | ---: | ---: |
| Concurrent Graph Visitors | 1,201,898 / 46,906 | 803,874 / 38,698 | -33.12% / -17.50% |
| Concurrent Packet Codecs | 2,407,104 / 49,406 | 567,192 / 16,621 | -76.44% / -66.36% |
| Concurrent Tree Folds | 929,992 / 24,779 | 537,032 / 16,592 | -42.25% / -33.04% |
| Concurrent Audio Voices | 3,690,915 / 98,547 | 1,591,338 / 65,774 | -56.89% / -33.26% |
| Concurrent Scene Tiles | 1,197,952 / 37,089 | 406,434 / 24,784 | -66.07% / -33.18% |

The compact counter report is
`2026-07-31-nominal-ownership-default-allocation-summary.json`.

## Official scorecard and broad verification

The official process harness used five independent runs per application,
public output verification, CPUs 12-15, and the pinned
`go1.26.5` toolchain contract. Its 10 ms process timer is intentionally
coarser than the balanced cohort:

| Application | Default Able | Go | Able / Go |
| --- | ---: | ---: | ---: |
| Concurrent Graph Visitors | 0.0300 s | 0.0040 s | 7.50x |
| Concurrent Packet Codecs | 0.0300 s | 0.0039 s | 7.69x |
| Concurrent Tree Folds | 0.0240 s | 0.0043 s | 5.58x |
| Concurrent Audio Voices | 0.0300 s | 0.0045 s | 6.67x |
| Concurrent Scene Tiles | 0.0220 s | 0.0036 s | 6.11x |

All 25 Able runs and 25 fresh Go-reference runs passed their verifiers. The
scorecard is
`2026-07-31-nominal-ownership-default-scorecard-five.{json,md}` and the fresh
reference report is
`2026-07-31-nominal-ownership-default-go-references.{json,md}`.

Default, opt-out, and equivalent-Go products for all five applications
produced their expected output hashes. Default-only Binary Event Log,
Concurrent Document Pipeline, and Versioned Telemetry Pipeline products also
passed their public contracts. All 13 inspected default/opt-out Able
dependency graphs omitted `pkg/interpreter`.

The aggregate external scoreboard still has 132 unique benchmark/mode rows.
Exactly the five measured compiled rows changed; the other 127 rows remain
from their existing verifier-backed source cohorts. The regenerated
performance frontier passes its exact check and has no actionable target
groups.

## Strict default census

The default-on census generated every selected application with
`--no-fallbacks`.

- 66/66 applications generated successfully; zero failed.
- 66/66 final Go dependency graphs resolved with
  `go list -mod=mod -deps`; zero failed.
- Zero dependency graphs contained `able/interpreter-go/pkg/interpreter`.
- Generation ranged from 378 ms to 14.983 seconds, with a 4.103-second mean.
- Disposable output totaled 288,134,658 bytes, 7,082,742 Go lines, and 3,316
  files before every row module was deleted.

The raw aggregate SHA-256 was
`b1c357bd772b8e58df51de50d6fed0bcd92d5892155a7452baa3a678d96aedb4`.
The census explicitly records
`nominal_ownership_execution_enabled: true` while the former experimental flag
is false.

## Verification

- focused ownership lifecycle, caller-result, and loop guards: 6.276 seconds;
- focused CLI default, opt-out, and mutual-exclusion guards: 5.165 seconds;
- broad interface, native/generic, imported/shadowed alias, Result, and Option
  guards: 37.942 seconds;
- `go test ./cmd/ablec`: 5.437 seconds;
- `go test ./cmd/able-generated-boundary-census`: 0.002 seconds;
- 18/18 broad default/opt-out/Go products passed public verification;
- 1,800/1,800 high-resolution timing processes verified;
- 50/50 exact allocation-counter processes verified;
- 25/25 official Able and 25/25 fresh Go-reference processes verified; and
- 66/66 strict generations and 66/66 dependency resolutions passed.

No individual test or generation exceeded one minute. All touched files remain
below 1,000 lines.

## Evidence-ledger consequence

Refreshing the five affected scoreboard rows does not authorize silently
advancing every historical compiled closure. The compiler-production scope
changed, so the checked ledger now selects 12 invalidations:

- 11 compiled performance closures; and
- the cross-family architecture-ownership closure.

The other 11 closures remain current. The exact selection is recorded in
`2026-07-31-nominal-ownership-default-ledger-invalidation.{json,md}`.
The checked ledger is intentionally not advanced by this tranche because the
other selected closures have not yet received fresh verifier-backed repeated
measurements.

## Cleanup

The exact 2,240,196 KiB disk-backed task workspace was measured before removal
and deleted after the compact evidence above was published. No task artifact
was written under RAM-backed `/tmp`, and no ownership-default task directory
remains under `/tmp` or `/var/tmp`. The removed raw build, census, and timing
artifacts are not recoverable; their compact evidence is retained in this
directory.

## Next recommendation

Refresh the 12 closures selected by the performance-evidence ledger, starting
with the 11 compiled closures and then reconciling the cross-family
architecture-ownership closure.

Why: making ownership lowering the default changed compiler-production
identity. The five measured applications prove the default rule itself, but
the ledger correctly refuses to treat older unrelated compiled dispositions
as current without fresh evidence.

What it entails: map each selected closure to its existing broad application
rows, run verifier-backed repeated default measurements against the pinned Go
references, confirm strict interpreter-free dependency graphs, update only
stale scorecard/evidence rows, and advance a closure only after its own scope
passes. Reconcile the cross-family closure last so shared evidence cannot mask
drift.

Why it matters: this restores a trustworthy compiler-wide performance baseline
after the default change. It tells us which costs remain real under native
nominal ownership and therefore where the next general lowering improvement
can be selected without relying on stale profiles.
