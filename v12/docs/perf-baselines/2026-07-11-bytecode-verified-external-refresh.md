# Verified external bytecode refresh (2026-07-11)

## Purpose

Re-rank the bytecode interpreter using only completed, verifier-backed external
applications before selecting another VM change. The JSON correction showed
that an unchecked one-run row is not a safe performance target.

## Method

- Ran three fresh normal bytecode processes per completed application under
  CPU `2`, `GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and the existing
  45-second per-process guard.
- The external wrapper ran each suite's setup hook and verifier after the
  timed process. Every row below has three matching verified outputs.
- Existing timeout rows (BinaryTrees, QuickSort, Sudoku, NBody, and Tapelang)
  and MatrixMultiply's no-verifier row remain status evidence, not timing
  candidates. Mandelbrot, ReverseComplement, and K-Nucleotide have no stored
  Ruby/Python comparison row, so their verified results likewise cannot select
  an interpreter-to-Ruby/Python optimization.

## Current completed rows

The short rows were sampled twice while isolating the harness's command-cell
limit; ranges below make that small process-level variation explicit rather
than presenting a false precise average.

| Application | Verified bytecode real time | Ruby ratio | Python ratio | Selection status |
| --- | ---: | ---: | ---: | --- |
| Fib | 0.1967 s | 0.00x | 0.00x | Faster than both references. |
| I-Before-E | 0.5967–0.6233 s | 5.97–6.23x | 4.59–4.79x | Current miss. |
| Base64 | 3.5800–4.1233 s | 1.62–1.87x | 1.08–1.25x | Current miss. |
| JSON | 0.8533–0.8767 s | 0.55–0.56x | 0.30–0.31x | Verified neutral control. |
| Monte Carlo Pi | 5.9233 s | 4.17x | 3.53x | Current miss. |
| PiDigits | 3.5567 s | 0.39x | n/a | Faster than the available reference. |
| Mandelbrot | 6.8900 s | n/a | n/a | Verified, but no relevant reference row. |
| ReverseComplement | 6.8767 s | n/a | n/a | Verified, but no relevant reference row. |
| K-Nucleotide | 43.3833 s | n/a | n/a | Verified, but no relevant reference row. |

K-Nucleotide retained its normal full input and completed all three processes
within the 45-second cap; it was monitored separately because each process is
longer than the interactive command cell. It is not reduced or treated as a
microbenchmark.

## Attribution

Fresh normal-process CPU profiles preserved the three current misses' distinct
walls:

- I-Before-E: call/member dispatch and stack-slot work (`execCallOpcode`
  21.43% cumulative and `appendSlotStackValueChecked` 7.14% flat).
- Base64: codec/MD5 host work beneath `execAndFinishExactNativeCall` (80.85%
  cumulative); MD5 block processing alone is 10.37% flat.
- Monte Carlo Pi: float store/cast/normalization and branch work
  (`execStoreSlotCastSlotFloatConstDiv` 36.47% cumulative and its discard fast
  path 31.77%).

JSON is a useful verified neutral control, but its 69.44% exact-native caller
leads to the independent `ableJsonF64FieldMeansFast` scanner and `strconv`
number parsing. Exact-native call dispatch is only a broad parent shared with
Base64, not a shared host-operation leaf. `runResumable` is similarly only a
common interpreter parent. No concrete low-level cost is material in two
misses and neutral on JSON.

## Decision

Keep no VM, compiler, or `able-stdlib` change. Do not tune a JSON scanner,
codec/MD5 host call, Float slot path, string/member call shape, or named
stdlib type from a single workload. The refreshed scorecard strengthens the
cross-program rule: even the remaining current misses are heterogeneous.

## Next recommendation

Apply the same three-run, verifier-backed refresh to completed compiled
applications with stored Go references. The compiler target is independent of
bytecode Ruby/Python parity, and its current scorecard may contain the same
one-run/unchecked selection error that JSON exposed. Re-rank only verified
compiled rows, then profile two current Go misses plus a Go-competitive control
only if their generated code exposes the same concrete lowering/helper cost.
