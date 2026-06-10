# Bytecode corpus-wide exact-owner census closure

Date: 2026-07-26

## Decision

Retain no VM, runtime, compiler, tree-walker, canonical-stdlib, benchmark,
fixture, language, dependency, or WASM change from this tranche.

All 54 rankable bytecode rows now have compatible current-artifact CPU,
sampled-allocation, and independent allocation-counter evidence. After
reconciling 407 material exact CPU symbols and 133 material exact allocation
lines against the completed/rejected-route ledger, only one general,
non-nominal removable leaf passed the breadth gate:
`bindMemberMethodTemplate`.

The candidate removed real allocations in four unlike semantic families, but
five-pair public comparisons slowed Binary Event Log by 2.64% and Rational
Series by 3.52%. It therefore failed the broad performance gate and was fully
reverted. Rebuilt interpreter and CLI artifacts exactly match the frozen
baseline hashes.

This result closes local profile-driven bytecode optimization for the current
corpus until new benchmark evidence, a semantic change, or an architectural
invalidation exposes a different concrete owner. The machine-readable
companion is
`2026-07-26-bytecode-corpus-owner-census-closure.json`.

## Corpus and artifact contract

The census uses the exact 54-row rankable selection from
`external-scoreboard-current.json`: 51 target misses, 22 semantic families,
and 147.327 seconds of total governing-target excess.

Both frozen artifacts were built once from repository commit
`237406eccdfb025a519d898daedadee1c8d13a7b` with Go 1.26.4:

| Artifact | SHA-256 |
| --- | --- |
| ordinary `cmd/able` CLI | `5f1108bc9596e74dd37e29fdb863bf8fa517e91935fd7db83ceecc940b896666` |
| `pkg/interpreter` benchmark binary | `5069b6dff944d7e68aeb38fb9b85dab990b4d29c842a6cfed04fe66897cb01ab` |

The interpreter binary is byte-for-byte identical to the July 24 and earlier
July 26 current-artifact profile records. The canonical external stdlib
remained the dirty 70-file source tree at commit
`219eff222c28406487231713753641bc49ee5b9a`; no stdlib file was changed.

Every process used CPU 6, one logical CPU, `GOMAXPROCS=1`,
`GOMEMLIMIT=1GiB`, `GOGC=50`, source-root-only loading, normal typechecking,
the canonical external stdlib, one warm main call, and the executor required
by the row. CPU profiles targeted at least 1.5 seconds of sampled measured-main
work, repeating short main calls up to ten times. Allocation profiles used a
64-KiB sampling rate. The normal cap was 59 seconds; K-Nucleotide alone used
180 seconds because one warmed and one measured call take about 84 seconds.

The workspace was disk-backed under `/var/tmp`, and output handling was
byte-safe for Mandelbrot.

## Refresh completeness

All phases completed successfully:

| Evidence | Passed | Aggregate wall |
| --- | ---: | ---: |
| CPU profile processes | 54/54 | 333.275s |
| sampled-allocation profile processes | 54/54 | 295.956s |
| independent allocation-counter processes | 54/54 | 284.425s |

K-Nucleotide was the longest row at 84.867s CPU, 85.946s allocation, and
83.272s counter wall time. Those are separate process/corpus measurements,
not individual Go tests.

Each profile was decoded with the exact frozen binary. CPU ranking used exact
flat symbols; allocation ranking used exact source lines with inlining
preserved. An owner was material in a row at 1% flat or more. Owners were
ranked by unlike-family reach, target-miss reach, total target excess, and
flat-weighted target excess.

## Reconciliation of the mechanical ranking

The widest CPU owners are not admissible leaves:

| Owner | Family reach | Miss reach | Disposition |
| --- | ---: | ---: | --- |
| `runResumable` | 20 | 47 | aggregate opcode dispatcher |
| Go collector scan | 19 | 41 | symptom over different semantic allocations |
| slot-to-stack append | 16 | 27 | completed carrier route |
| raw-integer inspection | 13 | 34 | completed typed integer lane |
| call-frame pop/push and inline return | 8-13 | 11-25 | closed broad frame/ABI route |
| Go map/hash leaves | 6-15 | 11-32 | differing interpreter maps and cache consumers |
| member-cache validation | 7 | 16 | retained dependency-safe cache; cheaper validation already rejected |

The only remaining CPU pair at the exact three-family breadth threshold was
the static-member cache lookup and receiver identity path. It is not repeated
name resolution: it constructs the polymorphic cache key and checks method,
scope, member-name, receiver-type/alias, owner, import, and implementation
versions. Distance Field, RMS Norm, Wide Integer Records, and Concurrent
Policy Callbacks use different concrete receiver kinds. Removing the shared
parent checks would weaken required cache invalidation, while specializing a
receiver kind would violate the generality rule. The previously retained
package identity already removes the avoidable string construction beneath
this path.

The widest allocation lines likewise reconcile to:

- test-process/package initialization, which is not measured-main work;
- observable positional nominal values with different definitions and
  lifetimes;
- completed integer/float boxing and typed-slot lanes;
- retained Array storage, leases, views, and capacity semantics;
- String host result construction and text builders already closed by the
  text gates;
- environments with differing synchronous and concurrent lifetimes;
- required explicit interface wrappers;
- AST/type-expression and error-name candidates already rejected by repeated
  broad gates.

None of those categories authorizes a new general leaf.

## Admitted experiment

`bindMemberMethodTemplate` accounted for at least 1% of sampled allocation
objects in 12 target misses spanning:

- Binary Event Log (`binary-nominal`);
- Rational Series (`wide-numeric`);
- Manifest Normalization (`generic-union`);
- nine concurrency applications.

The cache-miss path built a temporary bound callable only for
`storeCachedMemberMethod` to immediately extract and retain its unbound
template. The candidate replaced that temporary wrapper with a general
template-bindability predicate covering native functions, Able functions, and
overload sets. It did not name a benchmark, stdlib container, nominal type,
method, source shape, or primitive operation.

Focused member/static-cache, shadowing, and invalidation tests passed.

## Repeated measured-main A/B

Five independent processes per side used the fixed baseline and candidate
benchmark binaries. Positive time means slower:

| Application | Time | Bytes | Allocations |
| --- | ---: | ---: | ---: |
| Binary Event Log | -0.24% | -1.26% | -1.99% |
| Rational Series | +1.39% | -4.92% | -14.23% |
| Manifest Normalization | +2.02% | -0.48% | -1.83% |
| Concurrent Signal Dispatch | -0.61% | -0.06% | -0.05% |
| RMS Norm zero-reach guard | +0.83% | neutral | neutral |
| Word Frequency zero-reach guard | +1.32% | +0.16% | neutral |

The allocation removal is deterministic and especially large in Rational
Series, but measured-main time is mixed. Allocation reduction alone is not
enough to retain a VM change.

## Public verifier-backed A/B

Five interleaved launches per baseline/candidate/Python/Ruby lane all passed
their public output verifiers:

| Application | Baseline | Candidate | Change | Candidate/Python | Candidate/Ruby |
| --- | ---: | ---: | ---: | ---: | ---: |
| Binary Event Log | 5.428s | 5.571s | +2.64% | 33.58x | 23.18x |
| Rational Series | 3.818s | 3.952s | +3.52% | 40.15x | 29.28x |
| Manifest Normalization | 1.441s | 1.447s | +0.39% | 87.62x | 29.24x |
| Concurrent Signal Dispatch | 1.457s | 1.396s | -4.15% | 22.77x | 17.38x |

The candidate loses on two materially affected unlike applications, is
essentially neutral on a third, and wins only the concurrency row. It remains
far outside the required 1.052632x external target. This fails the broad
retention bar even though it reduces allocation counts.

## Revert and verification

The predicate experiment was removed. A fresh restored test binary and CLI
exactly reproduced the baseline SHA-256 values above, proving no candidate
runtime delta remains.

The restored focused family passed:

```text
GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1 \
  go test ./pkg/interpreter \
  -run 'TestBytecodeVM_(StaticMember|MemberMethod)|TestBytecode.*Member|Test.*Member.*Cache' \
  -count=1 -timeout 60s
ok able/interpreter-go/pkg/interpreter 0.445s
```

The complete `./run_all_tests.sh` handoff passed: coverage and scorecard
contracts, every non-compiler package, all 32 bounded compiler batches, and
the final bytecode fixture corpus. The aggregate non-compiler interpreter
package completed in 93.323s and the final fixture pass in 86.235s. Existing
compiler batches 5, 19, 28, and 29 took 69.611s, 183.124s, 102.960s, and
110.409s respectively. Those are pre-existing multi-test batch aggregates;
this tranche added no test and no production code. Their size is test-sharding
debt to address separately from this performance decision.

## Next recommendation

Refresh the current strict compiled scorecard and final dependency graphs,
then rank the largest shared generated/native owner or explicit
`runtime.Value` boundary.

Why: the corpus-wide bytecode census is the strongest remaining local-VM
selection test, and it now terminates without an admissible retained leaf.
Compiled native lowering remains the primary path to Go parity, and retained
compiler-bridge/runtime corrections landed after the last full compiled
scorecard.

What it entails: rebuild and verifier-check every strict `--no-fallbacks`
compiled row; prove final graphs still omit `pkg/interpreter`; repeat
equivalent Go measurements; then profile the widest target misses and admit
only a shared general generated-code/native-runtime owner. Any
`runtime.Value` work must correspond to an explicit dynamic, irreducibly
polymorphic, host/ABI, or runtime-service boundary.

Why it is important: this directly tests the immediate goal—Able primitives,
static Arrays, and general nominal encodings staying on equivalent Go
carriers without crossing into the interpreter. It moves effort back to the
execution mode where near-Go performance should be achievable without
weakening language semantics or inventing benchmark-specific lowering.
