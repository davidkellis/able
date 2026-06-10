# Verified bytecode numeric-reference gate (2026-07-11)

## Selection and method

This tranche tests whether the two independently written floating-point
applications, Mandelbrot and Monte Carlo Pi, identify a new general bytecode
VM improvement. K-Nucleotide is the unrelated text/frequency-map control: a
numeric change must not be inferred from the `runResumable` dispatcher parent
or from one program's source shape.

The sibling `mandelbrot` suite now includes Python 3.14 and Ruby 4.0 buffered
PBM renderers. The existing verifier accepts both outputs byte-for-byte against
the fixture. Monte Carlo Pi already had matching Python 3.14 and Ruby 4.0
references; its verifier accepts its intentionally nondeterministic sampled
output. Three fresh pinned reference processes used `taskset -c 2`, a
45-second cap, and the normal suite working directory:

| Application | Python 3.14 | Ruby 4.0 | Reference validation |
| --- | ---: | ---: | --- |
| Mandelbrot | 1.2036 s | 1.9124 s | exact verifier |
| Monte Carlo Pi | 1.4946 s | 1.6299 s | verifier-accepted nondeterministic output |
| K-Nucleotide control | 1.3185 s | 1.3588 s | exact verifier |

Current Able bytecode used the same pinning, canonical external stdlib,
verifier, and a 45-second cap:

| Application | Bytecode | Bytecode/Python | Bytecode/Ruby | Validation |
| --- | ---: | ---: | ---: | --- |
| Mandelbrot | 6.9667 s | 5.79x | 3.64x | exact verifier, 3/3 |
| Monte Carlo Pi | 2.7200 s | 1.82x | 1.67x | verifier, 3/3 |

## Warm-profile attribution

One-process runtime benchmarks load and warm the program before measuring
`main()`. The CPU samples cover one warmed invocation, with `GOMEMLIMIT=1GiB`,
`GOGC=50`, and `GOMAXPROCS=1`:

| Application | Warmed result | Material VM work |
| --- | --- | --- |
| Mandelbrot | 7,322,514,512 ns/op; 618,923,712 B/op; 76,303,149 allocs/op | fused float compare jump 22.25% cumulative; normalized-float discard store 14.97%; generic binary work 23.63% cumulative; allocation/boxing is material |
| Monte Carlo Pi | 2,902,933,551 ns/op; 177,862,128 B/op; 22,222,195 allocs/op | cast-slot float divide store 39.79% cumulative; fused float compare jump 13.84%; raw-float store normalization 20.07% cumulative; integer PRNG recurrence is also material |
| K-Nucleotide control | prior current-code 40.54-second CPU capture | call opcode 29.0% cumulative, inline return 18.1%, binary work 14.7%; primitive-map equality/hash below 1% flat |

The only material numeric leaf shared by Mandelbrot and Monte Carlo Pi is
`execJumpIfFloatMulAddMulCompareConstFalse(...)`. It does not appear in the
K-Nucleotide control. More importantly, it is the exact float compare/store
lane already subjected to broad A/B work: the instruction-indexed quickened
plan-table variant regressed Mandelbrot and MatrixMultiply, while other raw
float representation/store variants were neutral or regressed their diverse
guards. The surrounding costs also differ: Mandelbrot's general binary and
per-pixel normalized-store path does not repeat Monte Carlo's cast/divide and
integer recurrence path.

## Decision

Keep no VM, compiler, tree-walker, or stdlib code. The fresh references make
the current numeric gaps comparable and verifier-backed, but they reconfirm a
previously exhausted float lane rather than reveal a new concrete operation.
Do not reopen float compare/store, raw-float carrier, or quickened-plan-table
micro-variants from this result. No `able-stdlib` change is required.

## Next recommendation

Refresh the feature-coverage and cross-runtime status for an application family
that still lacks current Python/Ruby references—starting with the
file/byte-transform family—and only then form a pair with a distinct concrete
runtime boundary. The numeric family has now repeated its only plausible leaf
and rejected it with an unrelated map/call control; another numeric profile
would optimize already-measured noise. The next tranche should add verifier
backed references where needed, run the same pinned scorecard lane, and profile
only if two independently shaped applications expose a *new* shared leaf.
