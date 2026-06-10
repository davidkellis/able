# Compiled byte, numeric, and BigInt profile gate — 2026-07-14

## Decision

Keep no compiler, generated-runtime, bytecode VM, canonical-stdlib, or
benchmark-source change. The three current compiled misses exercise unlike
program shapes and do not contain a common concrete leaf. Optimizing any one
of their dominant paths now would be a benchmark-specific bet rather than a
generally applicable Able improvement.

## Method

The current compiler built fresh binaries for Reverse Complement, Monte Carlo
Pi, and PiDigits. Each binary ran from its catalog working directory with the
canonical external stdlib and its normal catalog arguments. Captures were
pinned to CPU 15 with a 45-second cap, `GOMEMLIMIT=1GiB`, `GOGC=50`, and
`GOMAXPROCS=1`.

The generated binary wrote its post-bootstrap `main` CPU profile through
`ABLE_GO_PHASE_CPU_PROFILE_DIR`. Twenty Reverse Complement, ten Monte Carlo
Pi, and eight PiDigits launches were merged independently with `go tool pprof
-proto`. Every run passed the benchmark's canonical Ruby verifier, and the
output hash was stable within each application. These are diagnostic profiles,
not replacement timing measurements; the CPU-pinned external scorecard remains
the performance authority.

## Result

| Application | Current compiled scorecard result | Launches | Merged samples | Dominant concrete work |
| --- | --- | ---: | ---: | --- |
| Reverse Complement | 0.1400 s; 8.75x Go | 20 | 510 ms | `runtime.growslice` 250 ms (49.02%), allocation/GC-assist work |
| Monte Carlo Pi | 0.3800 s; 1.73x Go | 10 | 1.22 s | generated `approx_pi` body 1.21 s (99.18%); signed division 210 ms and checked multiplication 190 ms |
| PiDigits | 1.3600 s; 1.10x Go | 8 | 9.93 s | `math/big` multiplication and division: `nat.mul` 4.37 s (44.01%) and BigInt division 4.04 s (40.68%) |

Reverse Complement is dominated by growing and collecting byte slices. Monte
Carlo Pi is a raw numeric recurrence whose visible generated helpers are
checked signed multiplication and division. PiDigits crosses the existing
generic BigInt host boundary and spends its time in Go's `math/big` algorithms.
Neither `growslice`, the signed integer helpers, nor `math/big` work repeats
across the other two applications.

## Why no candidate is eligible

The only superficially compiler-owned leaf, the checked signed arithmetic in
Monte Carlo Pi, is absent from the byte-transform and BigInt applications.
Changing it based on this one numeric row would violate the three-unlike-
application selection rule. Conversely, adding a special reverse-complement
buffer rule or a named BigInt/PiDigits lowering rule would violate the nominal
lowering and benchmark-generality policy.

The temporary generated binaries and profile artifacts were removed after
inspection. No `able-stdlib` change was needed.

## Next recommendation

Do not manufacture another profiling tranche from the same unchanged scorecard
just to find a code change. Keep the current verified corpus and scorecard
check as the regression gate, and reopen performance selection only after a
new broadly applicable implementation or semantic change produces a repeated
leaf across unlike applications. That is the shortest path toward real-program
speed rather than fast isolated benchmarks.
