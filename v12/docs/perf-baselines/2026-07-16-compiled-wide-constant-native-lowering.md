# Compiled Wide-Integer Constant Native Lowering (2026-07-16)

## Decision

Retain native static lowering for evaluated module constants whose primitive
types are Able `i128` or `u128`.

The preceding two-word carrier tranche made wide literals and operations
native, but the compiler's evaluated-module-constant table still recognized
native integer carriers only through 64 bits. Every static reference to an
`i128/u128` module constant therefore rebuilt a `runtime.IntegerValue`, parsed
it through decimal `big.Int`, and immediately converted it back to the native
wide carrier.

The retained change extends the existing primitive constant-carrier mapping to
`runtime.Int128` and `runtime.Uint128`, then uses the common wide literal
encoder. It does not recognize an application, package, stdlib module,
`Int128`, `UInt128`, `Rational`, or any other nominal type. Dynamic/module ABI
registration remains boxed as before; only statically proven constant reads
stay native.

## Profile selection

Strict no-fallback binaries for `int128_accumulate_small`,
`uint128_accumulate_small`, Fixed Width 128, and Rational Series each received
one separate bounded main-phase CPU process and one exact allocation-only
process. Runs used `GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`, and a
55-second process guard. Every output reproduced its established hash and the
two external applications passed their Ruby verifiers.

Before the candidate, exact main-phase totals were:

| Workload | Allocated bytes | Allocations | Repeated exact owner |
| --- | ---: | ---: | --- |
| signed accumulation | 328,002,136 | 8,000,045 | boxed `i128` module constants through `Int128FromValue` |
| unsigned accumulation | 86,400,600 | 3,400,015 | boxed `u128` mask through `Uint128FromValue` |
| Fixed Width 128 | 142,981,024 | 5,883,875 | boxed `u128` mask through `Uint128FromValue` |
| Rational Series | 244,801,848 | 6,000,043 | boxed `i128` range constants through `Int128FromValue` |

`math/big.nat.make` owned 48-84% of allocation bytes across the four profiles,
and `strings.NewReader` plus the generated constant-unboxing helper repeated
in every program. Generated source showed the exact missing boundary: native
wide operands combined with `runtime.NewBigIntValue(...)` constants followed
by `runtime.Int128FromValue(...)` or `runtime.Uint128FromValue(...)`.

## Repeated A/B gate

Preserved pre-change and candidate binaries ran in alternating order with
order reversal. Each side received seven independent processes per workload.

| Workload | Baseline mean | Candidate mean | Change | Speedup |
| --- | ---: | ---: | ---: | ---: |
| signed accumulation | 0.437023 s | 0.081060 s | -81.45% | 5.39x |
| unsigned accumulation | 0.234353 s | 0.069394 s | -70.39% | 3.38x |
| Fixed Width 128 | 0.372524 s | 0.121756 s | -67.32% | 3.06x |
| Rational Series | 0.393979 s | 0.075441 s | -80.85% | 5.22x |

Every baseline/candidate output had one matching SHA-256. Fixed Width and
Rational also passed their external verifiers.

## Post-candidate allocation state

Fresh exact candidate profiles confirm removal rather than displacement:

| Workload | Candidate bytes | Byte change | Candidate allocations | Allocation change |
| --- | ---: | ---: | ---: | ---: |
| signed accumulation | 16,000,888 | -95.12% | 1,000,022 | -87.50% |
| unsigned accumulation | 16,000,360 | -81.48% | 1,000,013 | -70.59% |
| Fixed Width 128 | 35,536,224 | -75.15% | 2,220,986 | -62.25% |
| Rational Series | 4,800,272 | -98.04% | 300,016 | -95.00% |

No boxed-wide-constant conversion remains in any generated program. Residual
allocation now belongs to returned nominal instances: generated `Int128`,
`UInt128`, and `Rational` construction. These are three different nominal
definitions and cannot be removed with a named-type compiler rule.

## Scorecard refresh

The standard external harness produced a dedicated five-process, verifier-
backed comparison rather than partially relabeling untouched rows in the
32-application cohort:

| Application | Compiled Able | Go | Able/Go | Previous selected ratio |
| --- | ---: | ---: | ---: | ---: |
| Fixed Width 128 | 0.1380 s | 0.0049 s | 28.16x | 1895.51x |
| Rational Series | 0.0940 s | 0.0118 s | 7.97x | 222.54x |

Machine-readable and rendered artifacts:

- `2026-07-16-compiled-wide-constant-comparison.json`
- `2026-07-16-compiled-wide-constant-comparison.md`

## Guards and verification

- The direct wide constant source audit requires zero fallbacks and rejects
  generated `NewBigIntValue`, `Int128FromValue`, `Uint128FromValue`, or
  `new(big.Int)` in static constant consumers.
- Existing <=64-bit module constant lowering remains native.
- Runtime wide arithmetic, direct primitive operations, cast/boundary checks,
  and strict `Int128`/`UInt128` no-bootstrap fixtures pass.
- Repository-wide Go compilation passes.
- A same-session 60-pair `sum_u32_small` guard is neutral: candidate mean is
  +1.75%, trimmed mean +1.16%, and median -3.39%, with identical hashes.
- Five alternating full Binary Trees pairs favor the candidate: 13.981 s
  versus 15.036 s (-7.02%), with identical verified output.
- No canonical stdlib, benchmark, reference implementation, spec, VM, or WASM
  source changed.

## Next recommendation

Run a returned-small-nominal selection gate across signed accumulation,
unsigned accumulation, Fixed Width 128, and Rational Series before attempting
another compiled optimization.

Why: after constant boxing is removed, exact allocation profiles repeat
returned nominal construction in all four workloads, but the owners are three
different nominal definitions. The only legitimate shared opportunity would
be a general compiler rule proving that a small nominal return can be kept as
value/scalar state across compiled calls without changing identity, mutation,
aliasing, dynamic-interface, or error-control semantics. A rule specialized to
`Int128`, `UInt128`, or `Rational` is forbidden.

What it entails: use generated Go escape analysis plus a temporary compile-time
consumer census to classify constructor returns by immediate field
deconstruction, method chaining, mutation, identity observation, collection or
environment storage, dynamic boxing, and escape. Require the safe shape in at
least three unlike programs and at least two nominal definitions before
building a candidate. If that gate fails, close this branch and select the
largest remaining compiled scorecard miss outside wide numerics. Preserve
small-integer, Binary Trees, startup, interface-boundary, and error-control
guards. Do not begin WASM work.
