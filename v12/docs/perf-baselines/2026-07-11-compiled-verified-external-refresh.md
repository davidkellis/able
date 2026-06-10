# Verified external compiled refresh (2026-07-11)

## Purpose

Re-rank the compiler against stored Go results using three fresh, externally
verified application processes before selecting another AOT lowering change.
This is the compiled counterpart to the bytecode refresh: an unchecked or
single-run ratio is not sufficient evidence for a generic compiler path.

## Method

- Ran `bench_compare_external` in compiled mode for three normal processes per
  completed application, with CPU affinity `2`, `GOMEMLIMIT=1GiB`, `GOGC=50`,
  `GOMAXPROCS=1`, and a 45-second per-process guard.
- The wrapper built each current generated binary before timing, ran the
  external suite setup where required, and executed the suite verifier after
  every timed process. All rows below are 3/3 verified with stable stdout
  hashes.
- BinaryTrees ran in the harness's explicit goroutine executor. It is reported
  separately: its stored Go row was produced as a parallel workload, whereas
  this selection pass is pinned to one CPU. That 8.16x ratio is operational
  status, not a fair compiler-lowering target.

## Current scorecard

| Application | Verified compiled time | Stored Go time | Able/Go | Selection role |
| --- | ---: | ---: | ---: | --- |
| Fib | 3.3567 s | 2.8400 s | 1.18x | Moderate recursive/control miss. |
| QuickSort | 1.9700 s | 2.0100 s | 0.98x | Go-competitive control. |
| I-Before-E | 0.1367 s | 0.0500 s | 2.73x | Short-process text row; not a stable profile selector alone. |
| Base64 | 2.5000 s | 2.2000 s | 1.14x | Moderate codec/host miss. |
| JSON | 0.7767 s | 1.3600 s | 0.57x | Go-competitive host/parser control. |
| Monte Carlo Pi | 0.2133 s | 0.1800 s | 1.19x | Moderate scalar numeric miss. |
| PiDigits | 1.3733 s | 0.7400 s | 1.86x | BigInt-specific miss. |
| Mandelbrot | 0.1567 s | 0.0400 s | 3.92x | Short-process float/control row. |
| ReverseComplement | 0.1367 s | 0.0100 s | 13.67x | Short-process byte/text row. |
| BinaryTrees | 31.2600 s | 3.8300 s | 8.16x | Verified, but one-core versus stored parallel-reference status only. |

The short I-Before-E, Mandelbrot, and ReverseComplement rows are important
correctness and direction indicators, but their reference values are 10--50
ms. They cannot establish a precise relative compiler target without rerunning
the Go reference in the same pinned lane. MatrixMultiply still has no
verifier, Sudoku does not complete within the guard, and the rows without a
stored Go result remain out of this compiler selection pass.

## Attribution decision

The relevant generated-code sources are unchanged from the current
cross-workload phase profiles:

| File | SHA-256 |
| --- | --- |
| `pkg/compiler/generator_binary.go` | `30a3ba8e89da9b427dea85bf9ca5dc90ad13151b06383b802d07943c58ff4cad` |
| `pkg/compiler/generator_render_runtime.go` | `7eb4dfb344cc442fc51d3dc433c8c3458452d50f208d5054f6911c3c7355113b |

Those profiles already separate the qualified long/mid-duration workloads:

- Fib is almost entirely its direct generated recursive body.
- Monte Carlo Pi is generated scalar RNG/float control plus checked signed
  arithmetic; QuickSort shares only checked multiplication and is already
  Go-competitive.
- I-Before-E, Base64, and JSON split into string-search/environment work,
  codec/MD5 host work, and JSON numeric parsing respectively.
- Mandelbrot is a separate generated float pixel loop; PiDigits is BigInt
  arithmetic with no second independently shaped, verifier-backed Go miss.
- BinaryTrees' known full-workload cost is tree allocation/GC and executor
  behavior, not the removed generic `Flush` polling loop.

Thus no exact generated helper or lowering boundary is material in two current
qualified misses and neutral on QuickSort or JSON. Keep no compiler, runtime,
or `able-stdlib` performance change. In particular, do not specialize
recursion, a random-number expression, a codec, BigInt, a byte-complement
loop, or BinaryTrees allocation from one row.

## Next recommendation

Refresh the Go reference binaries themselves in the same pinned process lane
before another compiled optimization tranche. The BinaryTrees concurrency
mismatch and 10--50 ms reference rows show that stored results alone cannot
define a reliable 95%-of-Go threshold. The work entails running the sibling Go
implementations with the same input/setup, CPU affinity, process guard, and
three-run verifier contract, then publishing a current Able-versus-Go
scorecard. Only profile two current, fair, nontrivial misses plus a
Go-competitive control after that refresh.
