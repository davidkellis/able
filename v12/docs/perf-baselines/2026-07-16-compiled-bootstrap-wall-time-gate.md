# Compiled bootstrap wall-time gate

Date: 2026-07-16

## Decision

Keep no additional compiler, generated-runtime, bytecode VM,
canonical-stdlib, application, verifier, reference, benchmark-source, or spec
change. Repeated internal timestamps show that the complete generated
`RegisterIn` sequence costs only about 0.53-1.94 ms on average across Document
Audit, Dependency Plan, Array Slice Window, and Base64. No individual
registration phase is a material explanation for the roughly 60 ms
complete-process wall of the three short applications.

The temporary instrumentation was removed before this record. The retained
interface-dispatch capacity improvement from the preceding tranche remains;
this gate closes further generated-bootstrap micro-allocation work until new
evidence identifies a material wall.

## Method

A temporary compiler binary emitted timestamps around these generated
`RegisterIn` phases:

1. environment/runtime setup and entry struct seeding;
2. builtin compiled-call registration;
3. builtin compiled-method registration;
4. package method/implementation registration;
5. generated interface-dispatch registration;
6. package definition/callable registration and import seeding;
7. package initialization;
8. resolver installation.

Each phase stored its elapsed nanoseconds in a local scalar. One diagnostic
line was written only after every phase completed, so stderr I/O did not enter
an individual phase. The diagnostic compiler was then removed from the source
tree before application collection. Normal generated code has no timestamp,
environment check, import, or reporting overhead.

Each short application contributed three independent batches of 100 process
launches (300 total). Base64 contributed three batches of five launches (15
total) because its main workload takes more than two seconds. Launch order was
changed between batches. Every launch used `GOMEMLIMIT=1GiB`, `GOGC=50`, and
`GOMAXPROCS=1`; no quiet-CPU requirement was imposed. This follows the
workstation policy of averaging repeated processes and retaining variability.

All 915 launches completed. Every diagnostic line had the expected nine
fields, and sample outputs reproduced the canonical hashes:

| Application | SHA-256 |
| --- | --- |
| Document Audit | `0dad030a80c8a883cbb56fbcfebfd530d521075e15d5d91ba538bc93e66c0aab` |
| Dependency Plan | `96dc74508d9b7a476bafdef453b11e11f2f70279c58ccaa5dcb6d85c529c4a38` |
| Array Slice Window | `155f89122475c7b282637dbf2ecba6d19771d396e801b581cb1d1b0cef64103e` |
| Base64 | `5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316` |

## Combined attribution

Means include the workstation's scheduler/GC tails; medians show the ordinary
launch. Times below are milliseconds.

| Application | Setup mean / median | Method impl mean / median | Interface mean / median | Packages mean / median | Total mean / median | Total p95 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Document Audit | 0.204 / 0.166 | 0.205 / 0.151 | 0.314 / 0.271 | 1.140 / 0.880 | 1.940 / 1.588 | 4.302 |
| Dependency Plan | 0.099 / 0.073 | 0.198 / 0.113 | 0.331 / 0.276 | 0.487 / 0.444 | 1.179 / 0.987 | 3.407 |
| Array Slice Window | 0.066 / 0.062 | 0.095 / 0.093 | 0.225 / 0.190 | 0.363 / 0.324 | 0.793 / 0.734 | 1.254 |
| Base64 | 0.043 / 0.046 | 0.044 / 0.041 | 0.121 / 0.118 | 0.283 / 0.270 | 0.529 / 0.542 | 0.685 |

Builtin calls, builtin methods, package inits, and resolver installation make
up the small remainder. Package definitions/callables are the largest common
phase, but their mean is only 0.28-1.14 ms. Interface registration is
0.12-0.33 ms after the retained capacity improvement.

The immediately preceding clean preserved-binary gate measured complete
process means of 60.861 ms, 62.102 ms, 64.490 ms, and 2.23336 s respectively.
The entire measured `RegisterIn` mean is therefore only 3.19%, 1.90%, 1.23%,
and 0.024% of those walls. Even an impossible removal of all generated
registration would leave the performance targets essentially unchanged.

## Admission result

No production candidate was built. Pre-sizing another compiled-call or method
map could at best affect a fraction of a sub-millisecond phase, while adding
generated code and another benchmark gate. The evidence does not justify
optimizing that noise, and it does not authorize removing semantically
required definitions, callables, methods, interfaces, or package inits.

This result also distinguishes the retained interface-capacity change from a
reason to continue the branch: that candidate had a deterministic 7-10%
bootstrap-allocation reduction and no broad timing regression, but the new
wall-time budget proves that smaller remaining allocation descendants are not
currently useful product targets.

## Verification and cleanup

- The temporary generator markers and their `os`/`time` imports are absent
  from the production source.
- Focused compiler/interface and profiling-hook tests pass on the restored
  production tree.
- `go build ./cmd/ablec` passes.
- Diff hygiene passes for this record and the retained preceding tranche.
- The diagnostic compiler, four generated trees, binaries, and 915 raw timing
  records are temporary and removed after this record.

## Next recommendation

Return selection to bytecode and profile Option/Result Configuration together
with Rational Series and Fixed Width 128. Require the same concrete generic
union, method-call, type-match, or nominal-value VM descendant in all three
before admitting a candidate.

Why: Option/Result Configuration is a stable verifier-backed bytecode miss at
207.28x Python and 82.30x Ruby in the current scorecard, while Rational Series
and Fixed Width 128 are separate nominal/numeric applications that also use
canonical `Result` values and remain 22-42x behind Python. This trio can tell
us whether the gap is a general union/result transport cost or merely the
Option benchmark's method composition. It avoids returning to regex engine
work whose language/library algorithm mismatch is already documented.

What it entails: collect bounded warmed CPU and exact allocation profiles for
all three through the existing catalog, inventory union construction,
type-match, member/call, return, and nominal allocation descendants, and build
a candidate only if one exact VM mechanism is material across all three.
Preserve public `Option`/`Result` semantics and the shared nominal lowering
rules; do not add `Option`, `Result`, Rational, fixed-width, or benchmark-named
opcodes or fast paths. Gate any candidate with unrelated union/interface and
numeric controls plus repeated verifier-backed process means.
