# Compiled architecture and semantic-amplification target budget

## Decision

**no-go-current-compiled-architecture-mechanism**.

Keep no compiler candidate from this evidence. Select future compiler work only from a newly measured concrete mechanism that is material in three unlike applications; do not treat aggregate allocation, direct generated bodies, or host-library ancestry as one optimization.

The model is intentionally favorable: it makes each attributed exact owner completely free and adds no replacement, conversion, code-size, or secondary cost. CPU-profile cumulative shares are used as wall-time-removal ceilings, not as timing claims.

## Application target budgets

| Application | Family | Able / Go s | Required speedup | Largest exact owner | Perfect owner gain | Remaining target speedup |
| --- | --- | ---: | ---: | --- | ---: | ---: |
| `k_nucleotide` | text-map | 2.670 / 0.068445 (5 / 5 samples) | 37.06x | `builtin-string-conversion` (22.03%) | 1.28x | 28.89x |
| `fixed_width_128` | wide-numeric | 0.200 / 0.005724 (5 / 5 samples) | 33.19x | `loop-carried-nominal-result` (89.29%) | 9.34x | 3.55x |
| `distance_field` | float-numeric | 0.106 / 0.013843 (10 / 10 samples) | 7.27x | `none` (0.00%) | 1.00x | 7.27x |
| `policy_record_dispatch` | regex-dispatch | 0.198 / 0.005337 (10 / 10 samples) | 35.24x | `regex-nfa-storage` (46.15%) | 1.86x | 18.98x |
| `concurrent_event_routing` | concurrency | 3.039 / 0.005796 (10 / 10 samples) | 498.07x | `goroutine-identity-discovery` (91.26%) | 11.44x | 43.53x |

Every row remains outside the 95%-of-Go budget even after perfect removal of its largest attributed exact owner. The exact remaining multipliers are retained in the table rather than copied from an older timing cohort.

## Allocation amplification by logical work

| Application | Able bytes / objects per unit | Go bytes / objects per unit | Amplification |
| --- | ---: | ---: | ---: |
| `k_nucleotide` | 612.212 / 13.867291 | 2.480 / 0.000043 | 246.90x / 326289.19x |
| `fixed_width_128` | 17.768 / 1.110491 | 0.000 / 0.000003 | 88839.90x / 317283.14x |
| `distance_field` | 0.000 / 0.000003 | 0.000 / 0.000003 | 0.38x / 1.40x |
| `policy_record_dispatch` | 24007.424 / 473.891357 | 157.586 / 1.643555 | 152.34x / 288.33x |
| `concurrent_event_routing` | 18987.178 / 366.769043 | 147.697 / 1.016113 | 128.55x / 360.95x |

Distance Field is the decisive counterexample to a universal allocation representation tax: its generated main allocates less than the Go reference in bytes and only two more objects, yet remains well over its target budget. The four amplified rows have different exact lifetimes and owners.

## Shared exact-mechanism ceilings

| Mechanism | Material families | Perfect-removal bounds | Disposition |
| --- | ---: | --- | --- |
| `builtin-string-conversion` | 2 | k_nucleotide 1.28x -> 28.89x remaining, policy_record_dispatch 1.18x -> 29.82x remaining, concurrent_event_routing 1.01x -> 494.89x remaining | `insufficient-three-family-cpu-leverage` |
| `bridge-to-uint` | 2 | k_nucleotide 1.17x -> 31.59x remaining, policy_record_dispatch 1.04x -> 33.89x remaining, concurrent_event_routing 1.00x -> 498.07x remaining | `causally-closed-and-insufficient-three-family-cpu-leverage` |
| `loop-carried-nominal-result` | 1 | fixed_width_128 9.34x -> 3.55x remaining | `single-family` |
| `regex-nfa-storage` | 1 | policy_record_dispatch 1.86x -> 18.98x remaining | `single-family` |
| `goroutine-identity-discovery` | 1 | concurrent_event_routing 11.44x -> 43.53x remaining | `single-family-and-causally-closed` |

String conversion and unsigned bridge conversion reach three applications by allocation, but are CPU-material in only K-Nucleotide and Policy. None of the five exact mechanisms clears the three-unlike-family admission gate.

## Complete architectural partition

| Area | Kind | Material families | Disposition |
| --- | --- | ---: | --- |
| `primitive-check-control-abi` | concrete-compiler-boundary | 1 | `closed-insufficient-breadth` |
| `nominal-representation` | concrete-compiler-boundary | 1 | `single-family` |
| `bridge-conversion` | concrete-compiler-boundary | 2 | `closed-insufficient-breadth` |
| `environment-control-values` | concrete-runtime-boundary | 1 | `single-family-and-causally-closed` |
| `allocation-gc` | aggregate-parent | 4 | `observational-split-owners` |
| `direct-generated-application-bodies` | aggregate-parent | 5 | `not-one-mechanism` |
| `host-library-boundaries` | semantic-boundary | 2 | `different-required-kernels` |

The aggregate parents are deliberately not candidates. Allocation/GC has four material families but four distinct semantic owners. Direct generated bodies cover all five because they contain the applications themselves. Host kernels combine unrelated required contracts. Primitive control, nominal representation, bridge conversion, and concurrency state each remain below the three-family gate as concrete mechanisms.

## Evidence and measurement policy

The timing rows pool every successful process from the current post-fix family closure artifacts. Allocation rows are two independent exact measured-main processes per application. CPU ownership comes from bounded main-only profiles. This report does not rerun unchanged binaries or use observer processes as timing evidence.

The source contract checks SHA-256 identities for every current timing cohort plus the residual model, allocation report, architecture audit, exact-leaf audit, primitive range/effect closure, and concurrency audit. It also requires five unlike application families, reconciled application sets, positive logical units, Go references, and exact mechanism membership.

## Next recommendation

Refresh `compiled-sudoku-quotient`, the only remaining invalidated performance closure.

Why: this checked budget and the current cross-family census close both broad architecture questions without admitting a candidate. Sudoku quotient is now the sole stale closure, and resolving it will make the full 21-closure ledger current without using one narrow application to justify a compiler rule.

What it entails: reconcile current repeated compiled/Go Sudoku timing and exact quotient/corrected-path reach, add a second retained cohort only if the workstation result is volatile, and advance only that closure. Keep no quotient optimization unless the same concrete mechanism becomes material in at least three unlike applications. Do not begin WASM work.

## Reproduction

```sh
python3 v12/bench_compiled_architecture_budget_test.py
v12/bench_compiled_architecture_budget \
  --json-out v12/docs/perf-baselines/2026-07-21-compiled-architecture-target-budget.json \
  --markdown-out v12/docs/perf-baselines/2026-07-21-compiled-architecture-target-budget.md
```
