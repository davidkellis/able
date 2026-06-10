# Compiled four-application main-profile gate

Date: 2026-07-17

## Decision

Complete the compiled Reverse Complement, Rational Series, Word Frequency,
and Array Slice Window gate and retain no compiler, generated-runtime,
bytecode VM, canonical-stdlib, benchmark, fixture, or language change. Fresh
generated-main CPU and exact allocation profiles divide below generic runtime
parents into four different semantic owners. No exact compiler-owned operation
is material in three unlike applications.

The gate therefore adds no byte-application, Rational, HashMap, String, Array,
or named-container lowering rule. It also does not retry closed backing-slice,
checked-arithmetic, bridge-conversion, package-environment, or nominal-result
candidates. WASM remains deferred.

## Measurement contracts

The normal comparison lane used canonical external `able-stdlib`, CPU 0,
`GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, a 55-second process limit, and
the catalog working directory and arguments. All 20 Able processes and all 20
fresh Go-reference processes passed their public verifiers with one stable
stdout hash per application.

| Application | Able five-run mean | Fresh Go mean | Able / Go | Able observed range |
| --- | ---: | ---: | ---: | ---: |
| Reverse Complement | 0.1600 s | 0.0240 s | 6.67x | 0.12-0.25 s |
| Rational Series | 0.1820 s | 0.0156 s | 11.67x | 0.13-0.25 s |
| Word Frequency | 0.2800 s | 0.0076 s | 36.84x | 0.25-0.39 s |
| Array Slice Window | 0.1800 s | 0.0061 s | 29.51x | 0.17-0.19 s |

Because the short Able rows were volatile, the current compiler also built
strict no-fallback product binaries once and each received two additional
ten-process verifier-backed timing cohorts. The combined 20 clean launches
were:

| Application | Preserved-binary mean | Fresh Go mean | Able / Go |
| --- | ---: | ---: | ---: |
| Reverse Complement | 0.0895 s | 0.0240 s | 3.73x |
| Rational Series | 0.1030 s | 0.0156 s | 6.60x |
| Word Frequency | 0.1995 s | 0.0076 s | 26.25x |
| Array Slice Window | 0.0680 s | 0.0061 s | 11.15x |

Ten alternating same-binary pairs tested whether `GODEBUG=gctrace=1`, which
the comparison harness uses to report GC counts, explained the difference.
Traced versus clean means changed Reverse Complement -1.1%, Rational 0.0%,
Word Frequency -2.1%, and Array Slice +1.4%. The instrumentation is not the
cause. A separately rebuilt low-level `ablec` Array binary had identical
generated Go source to the product binary outside unused copied stdlib/kernel
directories and averaged 0.064 s over ten runs. The first comparison cohort
therefore reflects workstation/build-adjacent variance, not a different
generated application implementation. No scorecard was promoted from this
targeted diagnostic.

## Phase profiling

Every strict binary passed with no fallback. CPU-only phase profiles excluded
bootstrap and exact allocation sampling. Short mains used repeated independent
verifier-backed launches, merging only `main.cpu.pprof`: 20 Reverse
Complement, 20 Rational Series, 15 Word Frequency, and 330 Array Slice Window
profiles. The high Array count was required because its generated main is
only a few milliseconds.

Separate allocation-only phase processes used exact sampling without the CPU
profiler. All four completed and verified inside 55 seconds. Phase counters,
not profiler snapshot serialization stacks, are the allocation authority.

| Application | Merged CPU samples | Main bytes / allocations | Material concrete owners |
| --- | ---: | ---: | --- |
| Reverse Complement | 570 ms | 19,905,928 / 101 | `runtime.growslice` 50.88% cumulative; generated offset-byte copy 35.09%; reverse-complement transform 47.37%; large input/output slices |
| Rational Series | 900 ms | 256 / 15 | `Int128.DivMod` 44.44% cumulative, `Uint128.DivMod` 23.33% flat, Rational build 60.00%, environment swap/restore 12.22% |
| Word Frequency | 2.05 s | 31,184,888 / 720,431 | generated HashMap find 44.88% cumulative/41.95% flat; `String.split` 33.66%; allocation 23.41%; string/byte conversions |
| Array Slice Window | 1.36 s | 1,441,352 / 24,012 | checked multiply 22.06%, checked add 11.03%, `Array.slice` 33.82% cumulative, allocation 25.00% |

The allocation attribution agrees with the counters:

- Reverse Complement allocates a few large input, output, and offset-copy
  slices. This is the already-reconciled backing-growth family.
- Rational Series' caller-owned `Rational` results make the loop essentially
  allocation-free. Its remaining wall is primitive i128 division and package
  call state, not allocation.
- Word Frequency creates hundreds of thousands of string/byte conversion and
  formatting values below `String.split` and HashMap operations.
- Array Slice Window creates exactly 24,002 independent `Array.slice` results,
  as required by mutable-copy semantics.

## Generality reconciliation

`runtime.tryDeferToSpanScan` appears in every CPU profile, but its callers are
not one Able operation. Reverse Complement scans large byte-slice growth,
Word Frequency scans string/map conversion allocations, Array Slice scans
independent result copies, and allocation-free Rational Series samples
background/previous heap scanning around primitive division. A Go collector
leaf is a consequence, not a legal compiler candidate.

The actual language-level descendants do not repeat across three programs:

- backing growth/copy is material only in Reverse Complement;
- i128 division and package environment restoration are material only in
  Rational Series;
- HashMap probing and String conversion are material only in Word Frequency;
- checked small-integer arithmetic reaches Rational and Array, but is material
  only in Array and has already failed broader lowering gates;
- independent Array slice allocation is material only in Array Slice Window.

Changing any one would optimize one application family. No candidate advanced
to an A/B gate, and canonical `../able-stdlib` needed no change.

## Verification and cleanup

- 20/20 normal Able comparison processes verified.
- 20/20 fresh Go reference processes verified.
- 40 additional clean preserved-binary processes and 80 alternating
  clean/gctrace processes verified.
- 385 CPU-profile processes and four exact allocation-only processes verified.
- Focused profile-hook and compiler CLI/build tests pass against the unchanged
  production implementation.
- An initially broad CLI test regex also selected
  `TestBuildOutputOutsideModuleRoot`; its child Go build reached the 60-second
  repository ceiling under workstation load. The timeout was not extended.
  Narrow argument/no-fallback tests and the independently completed `ablec`
  package passed, so the timeout is recorded as test-duration noise rather
  than a product result.
- Generated source trees, binaries, outputs, and raw profiles were removed
  after this aggregate evidence was recorded.

## Next recommendation

Add a preserved-binary compiled timing lane to the external comparison tools
and use it to reconcile short compiled rows before selecting another compiler
optimization.

Why: the first build-adjacent five-run means were 45%-165% slower than later
preserved-binary cohorts even though the generated implementation and
gctrace setting were not responsible. That variance is larger than most
candidate effects and can misclassify both wins and regressions on a
workstation. The project already requires repeated averaging; the harness
should make the build-once/run-many contract explicit and machine-readable.

What it entails: teach the compiled comparison lane to build each selected
binary once, record its source/generated/binary provenance, finish all builds
before timing, and then run at least two order-balanced verifier-backed timing
cohorts. Keep the existing full build-and-run integration lane, but use the
preserved-binary lane and fresh matched Go references for optimization
selection. Add JSON sample/provenance coverage and a variance check that
rejects short rows when independent cohorts disagree materially. After that
reconciliation, select the next compiler or stdlib operation only from a
concrete descendant repeated across at least three unlike applications.
