# Versioned Telemetry Pipeline application gate

## Decision

Retain Versioned Telemetry Pipeline as the 66th source-equivalent portable
application, together with its catalog, feature, operation-depth, selection,
scorecard, frontier, and profile evidence. Retain no compiler, generated
runtime, interpreter, VM, stdlib, language, dependency, or WASM code change in
this tranche.

The application fills the corpus gap identified by the sustained-work audit:
it is both broad and sustained. It covers ten of the 11 discriminating feature
families and its five-process Go mean is 0.2201 seconds, above the 0.200-second
floor.

The current strict compiled profile also opens one general next candidate.
A statically monomorphic captured callable is retained on erased
`runtime.Value` carriers, so native `*Sample` arguments cross the semantic
encoding boundary and its primitive `i64` body falls through dynamic
`__able_binary_op` and `math/big` work. The same erased captured-callable
boundary is already material in Manifest Normalization and Binary Event Log,
which supplies the required three unlike applications.

## Application contract

The default workload performs 16.5 million deterministic telemetry updates.
The `--short` contract performs 160,000 updates for bytecode, Python, and Ruby.
Both keep the following work in the sustained loop:

- application-owned generic nominal storage backed by static `Array`;
- primitive Array reads and writes;
- nullable reads and periodic `Result` matching;
- dynamically selected interface implementations;
- a capturing scorer closure;
- nominal construction and field access;
- control flow and arithmetic.

A generic readable interface and an iterator consume the final window. Able,
Go, Python, and Ruby implement the same contract. The canonical and external
Able sources are byte-identical at
`8701c4502963c99452439b882584aef3c8ea6b64bea024ae29447f75f9e6de20`.

Default output:

```text
16500000:13208878:3291122:13208878:3302220,3302214,3302218,3302226:868785125:113857858:917094197
```

Short output:

```text
160000:128104:31896:128104:32024,32028,32025,32027:939045114:549240532:733988122
```

Every language passed the public verifier. The strict build used
`--no-fallbacks`; its final dependency list omits
`able/interpreter-go/pkg/interpreter`.

## Repeated comparison

All measurements used CPU 12 and five independent verifier-backed Able and
reference processes. Compiled Able and Go used Go 1.26.5. No sample was
discarded.

| Mode | Able mean | Reference mean | Able/reference |
| --- | ---: | ---: | ---: |
| compiled | 51.4180 s | Go 0.2201 s | 233.6120× |
| bytecode | 3.3040 s | Python 0.2023 s | 16.3322× |
| bytecode | 3.3040 s | Ruby 0.1258 s | 26.2639× |

All five compiled samples finished between 50.94 and 51.79 seconds. All five
bytecode samples finished between 3.11 and 3.55 seconds. Both modes had zero
timeouts and failures.

## Compiled owner

One main-only strict CPU profile contributed 51.57 seconds of samples:

| Owner | Cumulative CPU | Share |
| --- | ---: | ---: |
| erased local `scorer` body | 34.78 s | 67.44% |
| `__able_binary_op` | 35.46 s | 68.76% |
| `__able_struct_Sample_to_seen` | 8.64 s | 16.75% |

Generated Go declares the closure as
`__able_fn_runtime.Value_runtime.Value_to_runtime.Value`. Each hot invocation
converts two already-native `*Sample` values, performs field access through the
runtime structure representation, and routes seven primitive operations
through the dynamic binary helper. The surrounding loop otherwise retains
native `int64`, `int32`, static Array, nominal, and interface carriers.

The ordinary sampled allocation profile attributes 608,030,785 cumulative
objects, or 73.26% of the captured objects, beneath the erased scorer.
`__able_binary_op` owns 633,056,182 cumulative objects (76.28%), and Sample
conversion owns 130,359,460 (15.71%). Most descendants are `math/big`,
`bridge.ToInt`, and runtime struct construction caused by the erased carrier.

This is the exact missing third-family evidence for the retained July 25
captured-callable closure:

- Manifest Normalization converts native `ManifestRecord` arguments at its
  erased captured `normalizer`;
- Binary Event Log converts native `EventRecord` arguments at its erased
  captured `scorer`;
- Versioned Telemetry converts native `Sample` arguments and additionally
  loses primitive lowering inside its erased captured `scorer`.

The next candidate may infer and preserve one concrete generated callable
carrier only when all statically known uses agree. It must retain erased
carriers at dynamic, conflicting, escaping, host, and runtime-service
boundaries. It may not name any application, record, `Result`, container, or
non-primitive nominal type.

## Bytecode closure

Three current cold-process profiles contributed 9.63 seconds of CPU samples.
They repeat the already closed bytecode owners: `runResumable`, binary
dispatch and raw-integer inspection, cached identifier/member lookup,
call/member dispatch, frame/return work, and casts. No new exact VM leaf is
specific to this application or shared by a newly eligible three-program
cohort, so no bytecode candidate is opened.

Readable profile summaries are under
`2026-07-30-versioned-telemetry-pipeline-owner-profiles/`. Repeated comparison
sources are
`2026-07-30-versioned-telemetry-pipeline-{compiled,bytecode}.json`.

## Next recommendation

Prototype the general statically monomorphic captured-callable carrier rule.

Why: Versioned Telemetry supplies the missing third unlike application for a
boundary already proven material in Manifest Normalization and Binary Event
Log. Its current scorer alone owns 67.44% of compiled CPU and 73.26% of
sampled allocations.

What it entails: extend local callable inference through nested and deferred
uses only when every statically known call agrees on one concrete parameter
and result signature. Preserve native generated carriers through the callable
body and invocation, while retaining erased carriers for dynamic, conflicting,
escaping, host, and runtime-service uses. Add focused positive and negative
compiler guards, then run order-balanced five-or-more-process baseline,
candidate, and Go cohorts across all three applications.

Why it matters: this directly advances the project goal that compiled Able
should behave like generated native Go. It removes a proven compiled/runtime
boundary and restores primitive arithmetic lowering without introducing a
benchmark, container, or nominal-type special case.

Do not begin WASM work.
