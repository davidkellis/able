# Post-Slot-Not-Nil Bytecode Scorecard Reconciliation

Date: 2026-07-18

## Decision

Do not promote the candidate scorecard. Fully revert the generic
`LoadSlot -> JumpIfNotNil` quickening admitted by the preceding bounded gate,
and retain the existing promoted scorecard as the current product record.

The candidate was semantically general and improved its three census-admitted
applications in the original preserved-binary gate, but a full selected-suite
refresh exposed a repeatable Mandelbrot regression. An exact causal A/B using
the same preserved baseline and candidate binaries measured the candidate
9.13% slower. That exceeds the broad guard bar, so local wins cannot justify
retaining the opcode.

No compiler, runtime value ABI, language, benchmark, fixture, canonical
`able-stdlib`, or promoted scorecard change is retained. No WASM work was
performed.

## Full primary refresh

All 27 selected bytecode rows received five independent ordinary processes
under CPU pool 0-3, each row's recorded one-CPU/executor contract,
`GOMEMLIMIT=1GiB`, `GOGC=50`, the canonical external stdlib, and a 55-second
per-process cap. Existing five-run Python/Ruby references were reused only
through their matched source, input, verifier, and execution-contract
fingerprints.

The three non-overlapping primary reports contain 135/135 successful Able
launches. Every output passed its public verifier and every sample remains in
the JSON evidence:

- `2026-07-18-post-slot-not-nil-bytecode-generality.json`;
- `2026-07-18-post-slot-not-nil-bytecode-async.json`; and
- `2026-07-18-post-slot-not-nil-bytecode-coverage.json`.

Against the preceding primary cohort, 15 row means improve, one is unchanged,
and 11 regress. The median movement is -0.22%. The unweighted sum changes from
108.284 seconds to 107.578 seconds (-0.65%). These cross-cohort numbers include
workstation movement and are not causal evidence for individual rows.

The complete candidate scorecard still reports 7/35 compiled rows meeting the
95%-of-Go target and 3/27 bytecode rows meeting both interpreter targets.
Base64, JSON, and PiDigits remain the bytecode meets; no classification flips.
`2026-07-18-post-slot-not-nil-bytecode-scorecard-rejected.{json,md}` retains
this complete non-promoted replay, and the strict five-sample evidence checker
accepts all 62 selected and 70 full-status rows.

## Expanded cohorts

The six short concurrency rows and I-Before-E, Base64, JSON, Mandelbrot, and
Reverse Complement received a second independent five-run cohort before the
candidate decision. Those add 55/55 verified launches. Collection stopped
before the planned coverage expansions once the causal Mandelbrot gate made
the rejection decisive; no partial report is presented as complete evidence.

Selected ten-run pooled means versus the preceding primary cohort are:

| Application | Candidate pooled | Previous primary | Change |
| --- | ---: | ---: | ---: |
| Mandelbrot | 7.451 s | 6.624 s | +12.48% |
| Reverse Complement | 7.093 s | 7.070 s | +0.33% |
| Base64 | 3.134 s | 3.136 s | -0.06% |
| JSON | 0.875 s | 0.854 s | +2.46% |
| I-Before-E | 0.581 s | 0.584 s | -0.51% |
| Future Pipeline | 0.502 s | 0.524 s | -4.20% |
| Mutex/Await Journal | 0.240 s | 0.224 s | +7.14% |

The remaining expanded concurrency changes range from -9.13% to +3.61% and
remain ordinary short-row workstation variability. Mandelbrot is different:
both independent five-run means are stable and slow (7.486 and 7.416 seconds),
while the dynamic census recorded the candidate pair at only 0.05% of its
instruction triples. That combination required a same-session causal binary
gate rather than another cross-cohort average.

## Exact Mandelbrot causal gate

The candidate binary has SHA-256
`b70d98ada7d7f58903f9591a810d005807ebc712f105b13cfe9e955d131c2def`.
The baseline was rebuilt after removing only the candidate opcode,
quickener, and dispatch path and exactly reproduced SHA-256
`b406e483ee6f05bc7c9d3681ffe8da64de297c3b3a2bec7b63310d122cfe63e0`.
These are the same two fingerprints used by the original local gate.

Five order-balanced pairs ran on CPU 0 with the scorecard memory/GC contract.
All 10 outputs passed the public verifier with one stdout hash, and no sample
was discarded:

| Variant | Samples | Mean | User mean | Range |
| --- | ---: | ---: | ---: | ---: |
| Baseline | 5 | 6.508 s | 6.448 s | 6.18-6.76 s |
| Candidate | 5 | 7.102 s | 7.044 s | 7.00-7.15 s |

The candidate is 9.13% slower by wall time and 9.24% slower by user CPU. The
candidate's added opcode/switch/finalization code is the only binary source
difference, but this gate does not claim which compiler layout or dispatcher
effect causes the loss. The result is sufficient to reject the implementation.
All ten rows are retained in
`2026-07-18-mandelbrot-slot-not-nil-causal-ab.tsv`.

## Restoration and verification

- Production source contains no `JumpIfSlotNotNil`, quickener, helper, or
  candidate test.
- A clean restored build exactly matches the baseline SHA-256 above.
- `go test ./pkg/interpreter -run 'TestBytecode' -count=1 -timeout 60s`
  passes in 25.427 seconds on the restored tree.
- The rejected candidate scorecard and the unchanged promoted scorecard both
  pass strict five-sample evidence validation.
- `bench_external_scoreboard --check`, manifest validation, seven execution-
  contract tests, two preserved-report tests, seven selection tests, and two
  refresh-protocol tests pass.
- The canonical stdlib remains 70 source files at tree SHA-256
  `f7a470aae4fba342e5bbc3fce53ee26fa6f96df71dde18e057e044520624dafc`
  and Git `219eff222c28406487231713753641bc49ee5b9a`, dirty.
- Temporary binaries, runners, benchmark workspaces, and failed scratch paths
  are removed after retaining the compact reports and timing ledger.

## Next recommendation

Run a bounded bytecode dispatcher-layout reconciliation before adding another
opcode or fused instruction.

Why: the slot-not-nil operation was generic, semantically safe, and locally
profitable, yet its small addition made an unrelated numeric hot loop 9.13%
slower. Until that sensitivity is understood, another otherwise reasonable
quickening can silently exchange wins among applications based on Go switch or
machine-code layout rather than removed Able work. This is now a shared VM
engineering constraint, not Mandelbrot-specific tuning.

What it entails: rebuild the exact baseline/candidate pair, collect bounded
main-only CPU profiles and disassembly/layout summaries for Mandelbrot plus
Reverse Complement, Option/Result Config, Word Frequency, and two numeric
controls, and separate executed-op savings from dispatcher/code-placement
movement. Test at most one general layout-stabilizing design—such as keeping
rare quickened operations behind an existing stable secondary dispatch—only
if the same mechanism explains the cross-workload differences. Require
order-balanced verifier-backed wins or neutrality across all controls; revert
if it merely reorders which program regresses. Continue to defer WASM.

## Follow-up

This recommendation is complete. The exact layout effect was reproduced and
one existing-dispatch design was tested over 90 verified causal processes. It
removed the Mandelbrot regression but produced expanded 3%-4% regressions in
three other applications, so it was reverted. See
`2026-07-18-bytecode-dispatcher-layout-reconciliation.md`.
