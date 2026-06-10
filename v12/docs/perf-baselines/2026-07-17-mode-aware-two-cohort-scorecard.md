# Mode-aware two-cohort scorecard reconciliation — 2026-07-17

## Decision

The first complete mode-aware scorecard is valid. Keep cohort A as the current
scoreboard, retain cohort B as an independent control, and select no compiler,
bytecode VM, runtime, stdlib, application, reference, or verifier change from
this measurement tranche.

The reviewed selection contains all 34 compiled applications and 27 bytecode
applications. All 68 application/mode pairs remain visible as full-status
rows. The seven excluded bytecode rows retain one bounded status probe rather
than disappearing from the report.

## Evidence correction

The initial cohort-A aggregation exposed a promotion hole: a selected row could
reach the aggregate after a foreign interpreter stopped early, even though the
strict two-cohort reader would later reject its incomplete repeated evidence.
`bench_scorecard_evidence_check.py` now applies that exact five-run eligibility
rule to a single aggregate before promotion. The grouped refresher invokes it
unconditionally, and its dry-run test proves the gate is present.

Python Fib cannot supply five successful samples inside the repository's
sub-minute process rule. A 59-second repair attempt timed out on its first
sample. Fib therefore remains visible in full bytecode status but is no longer
part of the reviewed bytecode comparison set. This is an evidence-boundary
decision, not an Able performance pass or an application change. The resulting
selection has 61 rows and SHA-256
`d829d5ae1a06dd346e1a9b9a0e8f4d33405bc0bca74c630ac858cf3912b35bf5`.

Both final aggregates pass the new gate with 61 selected rows, 68 full-status
rows, and exactly five successful verifier-backed Able and reference samples
for every selected row. They use disjoint comparison reports and the same
canonical 69-file stdlib state, source hash
`f37de0ac91abf02ab7c2af47e66cc06c9a37b9e32d618f4b12aee6ff11587f1d`.

## Reconciled result

Strict variance over the two independent cohorts passes. Of the 61 selected
application/mode contracts:

- 51 miss the 95% target in both cohorts;
- six meet it in both cohorts; and
- four compiled rows move from miss to meet when their Go reference slows
  under workstation load.

The stable meets are compiled Binary Trees, QuickSort, and JSON, plus bytecode
Matrix Multiply, JSON, and PiDigits. Compiled Fib, I-Before-E, Matrix Multiply,
and Monte Carlo Pi are the four volatile classifications and must not be
treated as established wins.

Stable large misses span unlike applications. Examples using the less
favorable of the two cohort-level ratios include compiled Option/Result
Configuration at 54.74x Go, Dependency Plan at 24.00x, Document Audit at
21.25x, and Word Frequency at 29.85x. Bytecode Regex Set/Stream remain more
than 93x the faster interpreter reference, Reverse Complement more than 79x,
Rational Series more than 27x, and Word Frequency more than 26x. These are
product gaps, but a ratio alone does not admit an implementation candidate.

Cohort A remains current because cohort B was deliberately collected with
`--no-promote` as a variance control. The current replay, selection checks,
catalog checks, both single-cohort evidence checks, and strict variance pass.

## Next gate

Map a genuine package-boundary route for compiled binaries to stop importing
and initializing the full interpreter when their statically generated program
does not need it. Start with Document Audit, Dependency Plan, and Option/Result
Configuration, with Binary Trees and Base64 as allocation-heavy and longer-run
guards.

This is the best next direction because the two cohorts repeat large compiled
gaps across unlike short applications, while earlier bytecode cross-family
profiles have already rejected their only shared raw-carrier, return, and
stack children. Prior startup work also showed that generated registration is
too small to explain these process gaps, whereas the interpreter dependency
audit measured a shared roughly 58–61 ms and 38 MB package-initialization cost.

The tranche should be feasibility-first: inventory the exact generated and
bridge operations reachable in the three applications, define the smallest
semantic runtime interface/package cut that removes interpreter initialization,
and build a candidate only if the same cut applies to all three without a
nominal/application special case. It must not retry lazy fixed-integer-cache
initialization, dummy heap ballast, generated registration micro-tuning, or
per-container lowering. Any candidate must preserve dynamic fallback behavior
and pass repeated complete-process guards, especially Binary Trees where the
earlier lazy-cache experiment changed GC pacing.

## Artifacts

- promoted cohort A:
  `2026-07-17-mode-aware-full-scorecard-strict-refresh.json`
- independent cohort B:
  `2026-07-17-mode-aware-full-scorecard-cohort-b-refresh.json`
- strict variance:
  `2026-07-17-mode-aware-full-scorecard-variance.json`
- current scoreboard:
  `external-scoreboard-current.json`
