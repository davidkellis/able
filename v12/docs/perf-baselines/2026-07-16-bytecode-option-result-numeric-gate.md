# Bytecode Option/Result and nominal-numeric gate

Date: 2026-07-16

## Decision

Keep no bytecode VM, compiler, runtime, canonical-stdlib, application,
verifier, reference, benchmark-source, or spec change. Fresh warmed CPU,
allocation, opcode, and call-trace evidence for Option/Result Configuration,
Rational Series, and Fixed Width 128 does not repeat one new concrete VM
descendant materially across all three.

The selection hypothesis was that canonical `Result` use might expose a
shared generic-union, method-call, type-match, or nominal-value transport wall.
It does not. Option/Result Configuration repeatedly lowers closure bodies and
generic-union method metadata; Rational Series spends on ordinary call
environments, Rational construction, and receiver type metadata; Fixed Width
128 spends on UInt128 calls, `math/big` arithmetic/cloning, and UInt128 value
construction. Shared call/return/raw-integer frames are broad parents or
previously rejected leaves.

## Method

The existing runtime harness loaded and typechecked each source program once,
warmed `main`, then measured one complete `main()` call. Every process used
canonical `../able-stdlib`, `ABLE_SOURCE_ROOT_ONLY=1`,
`ABLE_BENCH_SKIP_TYPECHECK=0`, `GOMEMLIMIT=1GiB`, `GOGC=50`, and
`GOMAXPROCS=1`.

CPU profiles ran with bytecode stats and trace disabled. Allocation profiles
ran separately, and stats/trace ran in a third process. Thus trace maps,
counters, exact allocation sampling, and CPU profiling do not become each
other's candidates.

The benchmark counter gives exact allocated bytes and allocation counts for
the measured call. Option/Result and Rational allocation attribution used
`MemProfileRate=1`. That made those diagnostic processes substantially slower,
so Fixed Width used a 64 KiB sample rate to keep the individual test below one
minute; its exact totals still come from the benchmark counter. Allocation
profile percentages are attribution evidence, not timing evidence.

| Application | Clean CPU-profile runtime | Exact B/op | Exact allocs/op | CPU samples |
| --- | ---: | ---: | ---: | ---: |
| Option/Result Configuration | 3,991,534,554 ns | 974,015,112 | 7,737,870 | 3.98 s |
| Rational Series | 3,474,320,087 ns | 129,986,080 | 1,405,719 | 3.47 s |
| Fixed Width 128 | 6,701,311,634 ns | 1,242,271,592 | 30,858,404 | 6.69 s |

These are warmed attribution rows, not replacements for the promoted external
scorecard.

## Option/Result Configuration

The hot path recreates and lowers callback lambdas inside repeated generic
union method calls. `lowerFunctionDefinitionBytecodeWithMethodSetEnv` is
39.70% cumulative CPU and 61.02% cumulative sampled allocated bytes.
`seedBytecodeLoweringStructDefs` is 13.57% cumulative CPU and its inner
allocator alone is 16.68% of sampled allocation objects. Generic-union
matching/expansion is visible but smaller: `staticGenericUnionMethodMatches`
is 6.28% cumulative CPU.

The exact allocation profile attributes 43.20% of sampled objects to function
lowering. Large flat children include struct-definition seeding, union-type
AST construction, identifiers/simple-type expressions, frame analysis, and
generic-union substitution. `runtime.mallocgc` is 48.24% cumulative CPU and
GC span scanning is 45.48%, reflecting the nearly 1 GB allocation rate.

The trace reports 24,576 inline calls at each of six hot sites: `or_else`,
`map`, `ok_or_else`, `and_then`, `unwrap_or_else`, and `is_ok`. The existing
bounded lambda-program cache does not reuse these programs across the newly
created lexical binding-shape states.

## Rational Series

Rational does not repeat the Option lowering wall. Its allocation profile is
instead almost perfectly partitioned among ordinary mechanisms:

- `runtime.newEnvironmentBase`: 300,003 objects (22.94%) and 40.15% of bytes;
- `runtime.NewStructInstancePositionalSized`: 300,001 objects (22.94%) and
  26.77% of bytes;
- bound member templates: 200,004 objects (15.30%);
- receiver-type identifiers and simple-type ASTs: 400,006 objects (30.59%).

CPU is ordinary VM work: `execCallMember` is 16.43% cumulative,
`finishInlineReturn` 10.66%, and `bytecodeRawIntegerValueInfo` 4.90% flat.
The trace is led by 1.2 million inline `rational_abs_i128`/
`rational_gcd_i128` calls plus 300,001 `rational_build` calls. No lambda
lowering or generic-union descendant is material.

## Fixed Width 128

Fixed Width also does not repeat the Option wall. The existing UInt128 member
fast path is 33.33% cumulative CPU; its add child is 27.95%. `math/big.nat.make`
is 13.45% cumulative CPU and accounts for 57.39% of sampled allocation
objects/50.82% of sampled bytes. Big-integer cloning and arithmetic account
for most surrounding work.

The other large allocation families are UInt128 nominal construction
(`NewStructInstancePositionalSized`, 8.35% of sampled objects and 22.13% of
bytes), boxed/raw integer results, and big-integer copies. The trace contains
one million `uint128_add_fast` calls, 220,963 `uint128_sub_fast` calls, and two
million inline comparisons. No generic-union or repeated lambda-lowering
descendant is material.

## Cross-program admission result

`execCallOpcode`, `execCallMember`, and `runResumable` occur across the cohort,
but they are dispatcher parents whose concrete children differ. The only
smaller overlapping leaves are `finishInlineReturn` and raw-integer extraction.
Both have already failed broad generic candidates, and neither is the dominant
allocation source in Option/Result.

Positional nominal construction repeats only in Rational and Fixed Width, not
Option/Result. Repeated lambda/generic-union lowering is material only in
Option/Result. Big-integer work is material only in Fixed Width. The required
three-unlike-program gate therefore fails, so no candidate was built. This
avoids an Option/Result shortcut, a Rational/UInt128 named-nominal rule, or a
retry of known mixed VM micro-optimizations.

## Verification and cleanup

Direct current bytecode processes passed each external Ruby verifier. Their
stdout SHA-256 values are:

| Application | SHA-256 |
| --- | --- |
| Option/Result Configuration | `28e46b27a6dceeaa15968e9db7a6728f4a2b35f87a89ff7bf561db18cad53112` |
| Rational Series | `127f0f44ee4870b57a188a7948f80a0a5d14584a326c345a48d4285594069f0c` |
| Fixed Width 128 | `eceabf5869b1abca8d6dd228b64a09f89e4e98ba8cabc4833ffee1218dafa56a` |

Focused generic-union, lambda cache, UInt128/raw-integer, and stdlib Result
tests pass on the unchanged production tree. Diff hygiene passes. The runtime
test binary, CPU/allocation profiles, stats, traces, and stdout captures are
temporary and removed after this record.

## Next recommendation

Run a temporary coverage-wide lambda-cache miss census before considering the
large Option/Result lowering wall. Count cache hits/misses by lambda AST,
binding-shape state/revision, and repeated same-AST/different-state misses
across every portable bytecode application; remove the counters afterward.

Why: Option/Result spends 39.70% cumulative CPU and 61.02% cumulative sampled
bytes in generic lambda/function lowering even though the VM already has a
bounded immutable-program cache. That is potentially a broad closure problem,
but both this gate and the previous Option/Dependency/Document gate find it
material in only one application. A corpus census can identify whether other
real programs miss for the same reason without prematurely weakening the
cache's lexical-shape correctness key.

What it entails: distinguish misses caused by new binding-shape state IDs from
actual shape revisions, nominal/type-definition dependencies, cache eviction,
and genuinely new lambda ASTs. Profile only applications with repeated
same-AST misses, and trial a dependency-derived lexical-shape fingerprint only
if at least three unlike verifier-backed programs share the mechanism. Any
candidate must preserve shadowing, imports, nominal-definition identity,
generic type bindings, mutation visibility, and bounded retention. Do not key
only on lambda source identity or add an Option/Result/member-name fast path.
