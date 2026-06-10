# Primitive-math sqrt coverage and kernel bridge

Date: 2026-07-16

## Decision

Retain two new portable benchmark applications and one generic primitive
kernel bridge, `__able_f64_sqrt(value: f64) -> f64`. Canonical
`../able-stdlib` continues to own the public `able.math.sqrt` domain contract;
after rejecting negative inputs and preserving its zero behavior, it delegates
the primitive operation to the bridge.

The compiler lowers statically resolved bridge calls directly to `math.Sqrt`.
The tree-walker and bytecode interpreter share the same native Go builtin and
return an Able `f64`. No NBody, Distance Field, RMS Norm, loop-shape, or
application-name rule was added. No non-primitive nominal lowering changed.
WASM implementation remains deferred.

## Coverage expansion

The preceding suite had only one full `sqrt` application, NBody. Two unlike
programs now close that evidence gap:

- `distance_field` advances a deterministic two-dimensional point recurrence
  and accumulates two million Euclidean distances through `math.hypot`;
- `rms_norm` advances four deterministic scalar channels, computes two
  million four-norms through `math.sqrt`, and reports both their sum and
  aggregate RMS.

They differ from NBody and from each other: geometry versus streaming
statistics, `hypot` versus direct `sqrt`, and one versus two output aggregates.
Neither uses NBody's arrays or solar-system update algorithm.

Each application has canonical Able source plus matched Go 1.26, Python 3.14,
and Ruby 4.0 sources, public tolerance-based verifier, README, and Docker lanes
in `../benchmarks`. The canonical and sibling Able sources are byte-identical.
Both are registered in:

- the `full`, `generality`, `coverage`, `numeric-structural`, and new
  `primitive-math` external suites;
- compiled and bytecode candidate-selection manifests; and
- feature coverage for expressions, control flow, and packages/imports.

The full static catalog now validates 34 portable applications, one diagnostic
application, and 78 bounded fixtures: 112 programs total.

## Current baseline

All processes used `GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`, a 55-second
guard, public verifiers, and arithmetic means over five independent launches.
The installed matched references were Go 1.26.4, Python 3.14.5, and Ruby 4.0.5.

| Application/runtime | Mean | Median | CV |
| --- | ---: | ---: | ---: |
| Distance compiled | 0.1222 s | 0.1207 s | 2.39% |
| Distance bytecode | 29.0347 s | 29.7369 s | 10.81% |
| Distance Go | 0.01076 s | 0.01065 s | 1.99% |
| Distance Python | 0.5098 s | 0.5088 s | 3.71% |
| Distance Ruby | 0.3121 s | 0.3100 s | 1.51% |
| RMS compiled | 0.1187 s | 0.1163 s | 3.34% |
| RMS bytecode | 21.3917 s | 21.2833 s | 3.38% |
| RMS Go | 0.00946 s | 0.00939 s | 2.44% |
| RMS Python | 0.8444 s | 0.8456 s | 5.89% |
| RMS Ruby | 0.6020 s | 0.5667 s | 13.46% |

All 50 baseline outputs passed their verifiers with one stable output per
application/runtime. Before the bridge, compiled Distance and RMS were 11.36x
and 12.55x their Go references. Bytecode was 56.96x Python / 93.03x Ruby for
Distance and 25.33x Python / 35.53x Ruby for RMS. NBody remained about 13.75x
Go in the preceding current gate, and full bytecode NBody timed out.

## Cross-program attribution

Ten independent compiled CPU profiles per new application reproduced NBody's
exact operation-level wall:

| Application | Pure-Able `sqrt`/`abs` | imported entry / environment work |
| --- | ---: | ---: |
| Distance Field | 45.61% cumulative | about 49% atomic/swap/restore cumulative |
| RMS Norm | 42.60% cumulative | about 54% atomic/swap/restore cumulative |
| NBody | 82.63% in the preceding gate | about 26% entry/atomic work |

Distance reaches `sqrt` through same-package `hypot`; RMS and NBody import
`sqrt` directly. The repeated primitive body is therefore not an artifact of
one source spelling.

Full bytecode profiles also show repeated `sqrt` call iterations layered over
ordinary VM float arithmetic, call-name/static-member dispatch, inline return,
slot snapshots, and allocation/GC. The VM does not contain a benchmark-shaped
sqrt opcode.

## Rejected pure-Able seed trial

The first candidate changed only Newton's initial guess for values above one
from `value` to `(value + 1) * 0.5`. Tree-walker and bytecode math tests passed,
and every application output passed its verifier.

Alternating compiled processes rejected it:

| Application | Baseline | Candidate | Change |
| --- | ---: | ---: | ---: |
| Distance Field, 20 pairs | 0.121921 s | 0.124231 s | +1.89% |
| RMS Norm, 20 pairs | 0.123124 s | 0.118250 s | -3.96% |
| NBody, 10 pairs | 0.368200 s | 0.358151 s | -2.73% |

Removing one iteration is not broad enough when the imported entry wrapper is
co-dominant, and the Distance regression fails the guard. The seed trial was
fully restored before the bridge was implemented.

## Retained primitive bridge

The retained implementation adds the primitive to the embedded kernel ABI,
typechecker builtins, interpreter native registration, compiler direct-helper
table, generated dynamic adapter, helper-result typing, and generated runtime
retention list. Static compiled `sqrt` contains a direct
`__able_f64_sqrt_native(value)` call, not `[]runtime.Value` materialization.

Alternating compiled results:

| Application | Baseline | Native bridge | Change |
| --- | ---: | ---: | ---: |
| Distance Field, 20 pairs | 0.128043 s | 0.102695 s | -19.80% |
| RMS Norm, 20 pairs | 0.125930 s | 0.105229 s | -16.44% |
| NBody, 10 pairs | 0.443778 s | 0.230215 s | -48.12% |

Five independent native-bridge bytecode launches per new application produced:

| Application | Baseline | Native bridge | Change |
| --- | ---: | ---: | ---: |
| Distance Field | 29.034744 s | 6.300897 s | -78.30% |
| RMS Norm | 21.391734 s | 5.962342 s | -72.13% |

Full bytecode NBody still cannot finish inside 55 seconds; the bridge removes
its sqrt iteration but not the separate array/VM wall. This is recorded as a
remaining timeout, not a passing performance row.

The new compiled ratios are 9.55x Go for Distance, 11.12x for RMS, and 7.52x
for NBody. New bytecode ratios are 12.36x Python / 20.19x Ruby for Distance
and 7.06x Python / 9.90x Ruby for RMS. The improvement is large and broad, but
the product targets remain unmet.

## Semantics and output

The public `sqrt` wrapper still raises `MathDomainError` before the primitive
for negative values and retains the explicit zero return. Expanded canonical
stdlib tests cover:

- zero and a sub-unit exact square;
- ordinary direct and method-call forms;
- a huge finite input (`1e300`);
- negative-domain failure;
- positive infinity; and
- NaN propagation.

Tree-walker and bytecode tests pass. The host operation is correctly rounded,
so RMS and NBody aggregates differ from the old tolerance-stopped Newton
result in their last few printed bits. All matched public verifiers accept both,
and the retained results agree with the operation used by the Go/Python/Ruby
references. Distance happens to retain identical stdout.

## Verification

- Primitive compiler helper generation and execution tests pass.
- Interpreter builtin numeric tests pass.
- Focused typechecker numeric tests pass.
- `go build ./cmd/able ./cmd/ablec` passes.
- Canonical stdlib math tests pass in tree-walker and bytecode modes.
- Both new sources build with `--no-fallbacks` and pass compiled/bytecode
  verifiers.
- Focused and corpus-full catalog checks pass.
- Every timing/profile process used for a retained claim passed its verifier.
- Temporary generated packages, binaries, profiles, and outputs are removed
  after this record.

## Next recommendation

Add an interprocedural package-environment-independence fact for compiled
static functions, then use it to route proven environment-independent imported
calls directly to their raw compiled bodies.

Why: after removing sqrt iteration, `SwapEnvIfNeeded`, restoration, and atomic
environment stores become the dominant exact compiled descendant in Distance,
RMS, and NBody: roughly 46–64% cumulative. Same-package calls already use raw
bodies. Extending that behavior only when a callee and its transitive static
dependencies read no package bindings, dynamic state, or environment-sensitive
metadata would benefit ordinary imported pure helpers without naming
`able.math` or any benchmark.

What it entails: compute a conservative fixed-point effect over compiled
functions and methods; mark any dynamic call, package/global lookup, closure,
runtime-generic environment access, or unknown extern as environment-dependent;
emit raw cross-package calls only for proven-independent targets; and retain
entry wrappers for runtime/dynamic callers. Test mutable package globals,
aliases, imports, generic specializations, closures, errors, concurrency, and
shadowing. Gate with Distance, RMS, and NBody plus I-Before-E or another
environment-dependent imported-call guard, repeated verifier-backed A/B runs,
and code-size/build-time checks. Keep bytecode float/call work queued and
continue to defer WASM.
