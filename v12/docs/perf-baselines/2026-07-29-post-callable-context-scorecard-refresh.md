# Post-callable-context compiled scorecard refresh

Date: 2026-07-29

## Decision

Promote a source-aware refresh of the authoritative scorecard.

The retained await-gated callable execution-context rule changes generated Go
for exactly four of 63 strict compiled applications. This refresh therefore
uses new paired five-process Able/Go 1.26.5 evidence for the two async groups
containing those four rows and preserves the existing Go 1.26.5 evidence for
the other 56 compiled rows and all 63 unchanged bytecode rows.

All 63 compiled applications were nevertheless rebuilt and measured before
selection. That complete candidate cohort passed 315/315 Go executions and
315/315 Able executions with zero timeouts, failures, or verifier mismatches.
It was not promoted because unrelated host load inflated the byte-identical
59-row control surface: its compiled geometric ratio was 5.254291x and
positive target excess was 6.858632 seconds. Replacing stable evidence with
that known-noisy surface would misrepresent the production change.

## Promoted compiled result

| Measure | Previous | Post-activation |
| --- | ---: | ---: |
| Target passes | 7 / 63 | 7 / 63 |
| Geometric-mean Able/Go ratio | 4.637116x | 4.263718x |
| Positive target excess | 5.353263s | 4.750737s |

The geometric ratio improves 8.05%, and positive target excess falls 11.25%.
All seven established compiled guards retain their snapshot status.

The four source-changed rows moved as follows:

| Application | Previous Able | Current Able | Change | Current Go | Current Able/Go |
| --- | ---: | ---: | ---: | ---: | ---: |
| Future Await Race | 0.042s | 0.034s | -19.05% | 0.0049s | 6.94x |
| Await Channel Mux | 0.150s | 0.128s | -14.67% | 0.0055s | 23.27x |
| Mutex Await Journal | 0.198s | 0.034s | -82.83% | 0.0052s | 6.54x |
| Mutex Work Queue | 0.486s | 0.040s | -91.77% | 0.0055s | 7.27x |

The selected 126-row scorecard has SHA-256
`02a24f666353e8130e54f3c1034a4f9fd6c0c2040f36945c6743035035e1557e`.
Its evidence check confirms five successful Able and reference processes for
every selected row and exact Go 1.26.5 provenance for every compiled
comparison.

## Updated compiled frontier

The largest compiled groups after activation are:

| Group | Positive excess | Disposition |
| --- | ---: | --- |
| Text/map | 1.655053s | closed: no shared concrete leaf |
| Concurrency | 0.891158s | open: spawn-gated callable context |
| Sudoku quotient | 0.868842s | closed: one-application reach |
| Current control | 0.479474s | closed: separate generated bodies |

Concurrency excess falls from 1.493684 seconds to 0.891158 seconds. The four
await-bearing rows no longer share the old `currentGID` owner, but the other
19 concurrency applications remain byte-identical and retain the exact
`bridge.currentGID`/`runtime.Stack` boundary established by their current
profiles.

This invalidates the old conclusion that explicit callable context is merely
a rejected broad ABI. The ABI is now retained and causally improves every row
that selects it. The remaining candidate is activation breadth: programs
containing statically loaded `spawn` should be evaluated for the same ABI
without enabling it globally for serial programs.

## Verification and scope

- The selected scorecard contains 63 compiled and 63 bytecode rows.
- Every selected row retains five successful Able/reference processes.
- All 63 newly measured strict compiled applications passed their public
  verifier.
- No compiled target status changed.
- The full 63-application dependency census from the retention tranche remains
  authoritative: every graph is interpreter-free, 59 await-free generated
  sources are byte-identical, and only four await rows changed.
- The canonical stdlib source identity is unchanged.
- No runtime, interpreter, bytecode VM, stdlib, language, dependency,
  benchmark, fixture, nominal-special-case, or WASM change was made.

The workstation had no CPU below the strict 5% idle threshold because
unrelated Marketlab work was active. No unrelated process was paused or
modified. Pairing each small Go-reference group immediately with its Able
group reduced drift; the complete noisy control cohort was rejected before
promotion.

## Next

Design and evaluate a spawn-gated extension of scheduler-context activation.

Why: the retained await gate proves that explicit callable context removes the
dominant compiled/interpreted boundary safely, while 19 byte-identical
spawn-bearing concurrency applications still execute the old goroutine
identity recovery path. This is the largest now-open exact owner repeated
across unlike applications. Static source census also finds `spawn` in Binary
Trees, an established target guard, so the candidate has 20 newly reached
rows and 39 spawn-free zero-reach controls.

What it entails: make scheduler-context detection recognize statically loaded
`spawn` as well as `await`; keep await-free and spawn-free programs
byte-identical; preserve dynamic and host compatibility entries; add imported
spawn, captured callable, package-environment, cancellation, and nested-task
guards; then run a complete 63-application source/dependency census plus
five-or-more balanced A/B/Go measurements, allocation counters, and exact
`currentGID` counts for the 20 newly reached rows, with Binary Trees serving
as an explicit performance-regression guard.

Why it is important: this can extend native Go call behavior to the remaining
compiled concurrency families without boxing primitive carriers, entering the
interpreter, or reviving the rejected global execution-context overhead.
