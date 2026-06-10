# Compiled pinned Go-reference refresh (2026-07-11)

## Purpose

Replace the stored Go figures with current, like-for-like reference processes
before using the 95%-of-Go goal to select compiler work. The stored rows mixed
parallel and very short historical Docker measurements with the current
single-core Able lane.

## Reusable reference runner

`v12/bench_refresh_go_refs` now builds each sibling Go 1.26 reference binary
outside the measurement, runs it from the same suite directory/input, pins its
process with `taskset`, and invokes the suite `verify.rb` after every run. It
emits JSON and Markdown reports and is available as:

```text
just bench-go-reference --runs 3 --timeout 45 --cpu-affinity 2
```

The runner uses `GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`, matching the
compiled Able refresh. It supports valid randomized benchmark output: the
Monte Carlo Go implementation produces distinct but verifier-accepted pi
approximations, so its row is `verified-nondeterministic` rather than an
invalid measurement.

## Fair current scorecard

All rows are three normal processes with externally verified output. Full
machine-readable data is in
[`2026-07-11-compiled-pinned-go-refresh.json`](2026-07-11-compiled-pinned-go-refresh.json).

| Application | Able compiled | Pinned Go | Able/Go | Result |
| --- | ---: | ---: | ---: | --- |
| Fib | 3.3567 s | 3.1752 s | 1.06x | Just outside the 95%-speed floor. |
| BinaryTrees | 31.2600 s | 34.3861 s | 0.91x | Able is faster in the fair one-core goroutine lane. |
| QuickSort | 1.9700 s | 2.7058 s | 0.73x | Go-competitive control. |
| I-Before-E | 0.1367 s | 0.0687 s | 1.99x | Short text-process direction signal. |
| Base64 | 2.5000 s | 2.6438 s | 0.95x | Meets the floor. |
| JSON | 0.7767 s | 1.5587 s | 0.50x | Go-competitive control. |
| Monte Carlo Pi | 0.2133 s | 0.2178 s | 0.98x | Meets the floor; Go output is valid but randomized. |
| PiDigits | 1.3733 s | 1.2814 s | 1.07x | Just outside the floor. |
| Mandelbrot | 0.1567 s | 0.0534 s | 2.93x | Short float/control direction signal. |
| ReverseComplement | 0.1367 s | 0.0167 s | 8.19x | Short byte/text direction signal. |

The previous apparent BinaryTrees compiler gap was a comparison artifact: the
one-core, verified Go reference is slower than the equally pinned Able binary.
The fair comparison also moves Base64 and Monte Carlo Pi into the target band.

## Selection decision

Keep no compiler, runtime, or `able-stdlib` performance change. Fib and
PiDigits are the only nontrivial fair rows just outside the floor, but their
existing current-source profiles are recursion versus BigInt arithmetic, with
no shared compiler helper. QuickSort and JSON remain independent
Go-competitive controls. The large I-Before-E, Mandelbrot, and
ReverseComplement ratios are based on 17--69 ms Go processes; they need more
normal-process samples before they can justify a lowering investigation.

## Next recommendation

Run higher-repeat fair measurements for I-Before-E, Mandelbrot, and
ReverseComplement in both compiled Able and Go modes, retaining their existing
verifiers and inputs. Why: these are the only large fair ratios, but their Go
reference processes are too short for a three-run average to distinguish a
generic text/byte or float/control wall from scheduling noise. The work entails
30 or more separately launched pinned processes per side, followed by a
ReverseComplement phase profile only if the text rows remain materially slow;
compare it to the existing I-Before-E/Base64/JSON profiles before considering a
generic compiler or stdlib change.
