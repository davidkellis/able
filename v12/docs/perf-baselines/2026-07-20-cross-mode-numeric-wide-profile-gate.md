# Cross-mode numeric/wide profile gate (2026-07-20)

## Decision

Keep no compiler, generated-runtime, bytecode VM, canonical-stdlib,
application, verifier, reference, or language change from this tranche.
Preserved generated-main profiles and warmed bytecode profiles for Fixed Width
128, Rational Series, Distance Field, RMS Norm, and a Mandelbrot float-loop
discriminator do not identify one new exact operation that is both material in
three unlike applications and open to a generic implementation.

The only exact three-program VM allocation owner is normalized raw-float
transport in Distance Field, RMS Norm, and Mandelbrot. That is not a new
candidate: the exact carrier substitution, owned-cell/slot-sidecar variants,
typed float operand lane, native scalar return, and producer-fusion designs
have already received multi-program correctness and wall-time gates. They
reduced allocations or moved them to another boundary while slowing guarded
applications. Repeating that design would optimize a profile counter against
the existing wall-time evidence.

## Admission rule

Before collection, a candidate had to satisfy all of these conditions:

- the same concrete primitive, checked/wide arithmetic, conversion,
  nominal-result, coercion, or float operation had to be material in at least
  three unlike programs;
- a compiler candidate had to be a primitive lowering rule or part of shared
  nominal lowering, never a rule for UInt128, Rational, or another named
  non-primitive type;
- a VM candidate had to be a general runtime rule rather than a benchmark or
  opcode-sequence exception; and
- a previously rejected carrier, typed-block, call/return, member-cache, or
  Array-growth design needed new invalidation evidence before it could reopen.

No operation cleared that bar, so no source A/B candidate was built.

## Reproducibility contract

One `ablec` and one `interpreter.test` binary were built before collection and
preserved for the entire gate:

| Artifact | SHA-256 |
| --- | --- |
| `ablec` | `a3c29290a8412612ac1126498294d6e2799f0ef91f56550534fbdd38bdadbde3` |
| `interpreter.test` | `53e137130679ee6feab651a14a482a81c958c18aa4cb413e4043faedb4b0d517` |

The external canonical stdlib contained 70 Able sources at aggregate SHA-256
`4b61c593a4837f812baf9f9abe2a4d65c0a6259c0e0a36c2322d35b134f65562`.
Every process used one logical CPU, `GOMAXPROCS=1`, `GOGC=50`, a 1 GiB Go
memory limit, and a 55-second process cap.

Generated programs were built once. Main-phase CPU profiles merged 30
separate launches per application so short compiled programs accumulated
useful samples without profiling multiple `main` calls in one process. Every
one of the 150 launches passed its public Ruby verifier. Allocation snapshots
used one separately verified generated-main launch and the phase profiler's
exact main-only counters.

The bytecode runtime harness loaded and lowered each application once, warmed
`main`, then profiled measured `main` calls. CPU runs used 2 calls for Fixed
Width, 6 for Rational, 5 for Distance, 6 for RMS, and 5 for Mandelbrot.
Allocation ownership used a separate one-call process with 64 KiB sampling;
the benchmark counters remain exact while symbol percentages are attribution
evidence. All ordinary bytecode CLI outputs also passed their public
verifiers. Mandelbrot's CPU process completed in 54.72 seconds; every other
profile process had more headroom.

## Target context

The current promoted scorecard makes these stable product misses rather than
microbenchmark-only targets:

| Application | Compiled Able / Go | Bytecode Able / Python | Bytecode Able / Ruby |
| --- | ---: | ---: | ---: |
| Fixed Width 128 | 42.14x | 18.94x | 11.93x |
| Rational Series | 12.66x | 37.91x | 29.55x |
| Distance Field | 11.45x | 10.07x | 18.03x |
| RMS Norm | 12.83x | 5.47x | 8.93x |
| Mandelbrot | 3.20x | 5.22x | 3.14x |

These ratios are context from the repeated external scorecard. Profile times
below are attribution measurements and do not replace its workstation means.

## Generated compiler attribution

The exact generated owners separate by program:

| Application | Merged main CPU samples | Dominant exact work | Exact main allocation |
| --- | ---: | --- | ---: |
| Fixed Width 128 | 4.29 s | atomic environment publication, bridge swapping, generated modular-add/ordered-select checks, checked multiply | 35,536,192 B / 2,220,984 allocations |
| Rational Series | 1.70 s | `Uint128.DivMod` 27.65% flat; `Int128.DivMod` 49.41% cumulative; rational GCD/build | 256 B / 15 allocations |
| Distance Field | 250 ms | `sqrt` 36% flat and `hypot` 60% cumulative | 144 B / 7 allocations |
| RMS Norm | 360 ms | generated `sqrt` 50% cumulative and the numeric loop | 248 B / 13 allocations |
| Mandelbrot | 1.53 s | generated `pixel_byte` 95.42% cumulative | 86,320 B / 66 allocations |

Distance and RMS share square-root geometry, but only two programs do.
Fixed Width's bridge/GC and generated nominal-result allocations do not occur
in the allocation-light Rational, Distance, RMS, or Mandelbrot mains.
Rational is a multiword division/GCD workload, while Mandelbrot spends almost
all main CPU in its primitive pixel loop. There is no exact compiler owner in
three programs and therefore no admissible primitive or shared-nominal
lowering candidate.

## Bytecode attribution

The repeated warmed CPU processes reported:

| Application | ns/op | B/op | allocs/op | Profile samples |
| --- | ---: | ---: | ---: | ---: |
| Fixed Width 128 | 8,044,531,266 | 1,242,225,660 | 30,858,325 | 16.01 s |
| Rational Series | 4,166,797,487 | 129,885,965 | 1,405,553 | 24.89 s |
| Distance Field | 6,116,276,339 | 368,019,907 | 26,000,091 | 30.45 s |
| RMS Norm | 4,751,955,373 | 288,016,553 | 20,000,094 | 28.40 s |
| Mandelbrot | 6,768,394,946 | 615,102,036 | 76,302,969 | 33.71 s |

The concrete descendants again split:

- Fixed Width is led by Go map/GC work around wide nominal calls, with raw
  integer extraction only 2.06% flat. Its allocation is chiefly `math/big`,
  UInt128 positional results, boxed integers, and BigInt cloning.
- Rational is led by raw integer extraction, binary execution, environment
  lookup, calls/returns, and Rational construction. Its allocation owners are
  call environments, positional Rational values, bound methods, and raw
  integer results.
- Distance is led by static/native calls and float geometry: call execution is
  47.62% cumulative, cached static-member calls 8.24%, and its typed float
  region 3.91%.
- RMS is led by its typed float region (10.77% cumulative), static/native calls,
  raw float reads, snapshots, and return work.
- Mandelbrot is led by the fused float condition (21.21% cumulative), raw-float
  normalization/boxing (21.95% cumulative), and allocation/GC work.

The allocation profile makes the apparent shared float wall explicit:

| Exact flat allocation owner | Distance | RMS | Mandelbrot |
| --- | ---: | ---: | ---: |
| normalized raw-float slot value | 124.89 MB / 45.20% objects | 91.07 MB / 42.54% objects | 580.70 MB / 99.63% objects |
| materialize raw float | 135.21 MB / 32.62% objects | 45.95 MB / 14.31% objects | below the displayed tier |
| stable stack snapshot | below the displayed tier | 46.07 MB / 14.35% objects | below the displayed tier |

This is the same wall recorded by the coverage-wide allocation-owner census
and the typed-handler, f64 operand-lane, float-sidecar, native-scalar-result,
and typed-float-region gates. The true operand lane cut allocation by
5.3%-30.3% across four unlike numeric programs but slowed all four by
12.7%-23.6%; pointer/sidecar alternatives moved boxing to call and snapshot
boundaries. The current profile supplies recurrence, not invalidation.

Raw integer extraction repeats materially only in Fixed Width and Rational.
Square-root/native float geometry repeats only in Distance and RMS. Generic
call/return, dispatcher, allocator, and map parents are broad aggregation
points whose concrete children differ, and their relevant local designs are
already closed. Consequently no candidate advanced to repeated A/B timing.

## Verification and cleanup

- 150/150 compiled profile launches passed the public verifiers.
- 5/5 ordinary bytecode CLI executions passed the public verifiers.
- Every admitted profile completed under the one-minute rule.
- The production compiler, runtime, VM, benchmark sources, references,
  verifiers, spec, and external canonical stdlib are unchanged.
- No WASM work was performed.
- Preserved binaries and raw profiles were removed after this record was
  written.

## Next recommendation

Completion correction: Regex Suffix Audit already was the third portable
discriminator, with Able, Go, Python, Ruby, and verifier lanes. Its 16,384-word
default caused the bytecode timeout that made it look absent from strict
selection. The follow-up tranche uniformly bounded it to 512 words, selected
the verified bytecode row, reconciled the scorecard, and profiled Suffix, Set,
and Stream. The same canonical NFA operations repeat in all three, but every
material concrete option is already retained or rejected by broad wall-time
evidence, so no runtime candidate was kept. Record:
`v12/docs/perf-baselines/2026-07-20-regex-three-api-current-profile-gate.md`.
