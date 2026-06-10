# Compiled Fib / Monte Carlo Cross-Pair Attribution

## Decision

Keep no compiler, runtime, canonical-stdlib, or benchmark-source change. The
two verified 1.17x Able/Go rows in the refreshed external scorecard do not
share a material generated helper or lowering boundary. A change for either
would optimize a particular program shape rather than a broadly applicable
Able program.

## Evidence and source applicability

This tranche cross-attributes the bounded phase profiles captured earlier on
2026-07-11 instead of needlessly rerunning the same processes. The profiles
were collected from the exact external benchmark directories, pinned to CPU 2,
with `GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and a 45-second guard.
Fib and QuickSort use the collector-free CPU-only phase hook; Monte Carlo Pi
and Mandelbrot use the equivalent numeric-kernel capture. The scorecard then
independently re-established the current-source 1.17x Fib and Monte Carlo Pi
ratios.

The source files that produce generated callable bodies and shared generated
runtime helpers are byte-identical to the retained profile builds:

| File | SHA-256 |
| --- | --- |
| `pkg/compiler/generator_binary.go` | `30a3ba8e89da9b427dea85bf9ca5dc90ad13151b06383b802d07943c58ff4cad` |
| `pkg/compiler/generator_render_runtime.go` | `7eb4dfb344cc442fc51d3dc433c8c3458452d50f208d5054f6911c3c7355113b` |

`generator_render_main.go` has changed only with the later profile-launcher
work, so these captures are used for callable-body attribution, not as a
replacement for the current end-to-end scorecard. Artifacts remain in
`v12/tmp/compiled-structural-profiles/` and
`v12/tmp/compiled-numeric-profiles/`.

## Main-phase CPU results

| Workload | Samples | Material leaves | Consequence |
| --- | ---: | --- | --- |
| Fib | 9.56 s | generated `fib`: 9.51 s / 99.48% flat | Direct recursive body; no sampled helper below it. |
| Monte Carlo Pi | 1.53 s | generated `approx_pi`: 1.02 s / 66.67%; checked signed multiply: 0.26 s / 16.99%; signed div/mod: 0.24 s / 15.69% | Scalar Park--Miller update and point-in-circle test. |
| QuickSort control | 5.22 s | generated `quicksort`: 3.11 s / 59.58%; `swap`: 0.70 s / 13.41%; checked multiply: 0.62 s / 11.88%; parsing: 0.60 s / 11.49% | It overlaps Monte Carlo only at checked multiply, but is already 0.98x Able/Go. |
| Mandelbrot control | 0.59 s | generated `pixel_byte`: 0.52 s / 88.14%; checked shift: 0.03 s / 5.08% | Independent floating-point escape kernel. |

The exact phase-allocation mode establishes only a coarse baseline: Fib main
allocated 2,676,344 bytes in 950 allocations; Monte Carlo Pi 2,451,072 bytes
in 288; QuickSort 299,313,128 bytes in 1,250; and Mandelbrot 2,908,840 bytes
in 285. Its allocation profiling materially perturbs short programs, so it is
not used to select a leaf.

## Why no candidate is authorized

Fib does not call either of Monte Carlo Pi's material arithmetic helpers.
Although checked signed multiplication repeats in QuickSort, that control is
already within the Go-performance floor and its remaining work is partitioning,
swapping, and input parsing. Mandelbrot instead confirms that another numeric
miss has its own generated-body wall. Thus no leaf recurs materially in the
two selected misses plus an independently shaped control.

The checked arithmetic helpers have also already received their generic
non-negative fast paths; inventing a Fib-recursion shortcut, a random-number
rule, or a benchmark/named-type special case would violate the generality
requirement. No `able-stdlib` change is warranted.

## Next recommendation

Resolve the full external Sudoku status gap before selecting another
optimization: perform a bounded, verifier-backed input-size ladder for both
compiled and bytecode modes to distinguish an input/contract issue from a
general backtracking, string, or control-flow performance wall. The current
full rows time out in both modes, so they cannot supply a performance ratio or
justify a source-specific fix. Profile only if the ladder exposes a concrete
boundary that also appears in an independently shaped workload; retain the
unchanged full verifier as the acceptance gate.
