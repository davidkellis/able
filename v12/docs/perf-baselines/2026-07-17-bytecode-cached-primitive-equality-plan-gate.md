# Bytecode cached primitive equality plan gate — 2026-07-17

## Decision

Keep a direct primitive plan in cached equality-dispatch entries. Once `==` or
`!=` has resolved a primitive receiver through the intrinsic `Eq`/`PartialEq`
matrix, subsequent comparisons now execute the same canonical coercion and
comparison helpers directly instead of rebuilding a native callable invocation.

This is a primitive language/runtime boundary, not a benchmark, application, or
stdlib-container special case. Custom nominal implementations continue through
their cached Able callable. No compiler, stdlib, language, fixture, or retained
benchmark-source change was required.

## Admission evidence

Fresh profiles were collected from the clean primitive-native baseline before
implementing the candidate. The remaining generic callable shell was material
in three unlike primitive consumers:

| Workload | Total CPU | Complete equality path | Cached/callable shell |
| --- | ---: | ---: | ---: |
| Boolean Reconciliation | 2.74 s | 13.87% | 10.58% |
| Run-length encode | 4.37 s | 14.65% | 10.07% / 9.84% |
| Unicode Scalar Pipeline | 4.17 s | 6.24% | 3.36% |

The temporary custom nominal `Eq` profile was deliberately excluded from the
primitive plan: its general callable shell was material, but it represents
user-defined semantics. Iterator Collect and Numeric Array Map did not have a
material equality shell and served as unrelated controls.

## Implementation

The equality cache now records whether a resolved method belongs to a primitive
receiver. On a primitive cache hit, `applyCachedPrimitiveEquality(...)`:

- unwraps and canonicalizes each operand once;
- uses the existing receiver-relative primitive coercion rules;
- uses the existing primitive equality implementation;
- negates the result for `!=`.

The native method implementation was factored into canonical-input helpers so
the ordinary explicit-interface path retains identical semantics without
duplicating logic. The plan contains no branches for individual benchmark or
nominal type names.

## Repeated workstation gate

Every timing is an independent process with one warmup and one measured call,
the canonical external `able-stdlib`, `GOMAXPROCS=1`, `GOGC=50`,
`GOMEMLIMIT=1GiB`, CPU 0, and skipped benchmark typechecking. All valid samples,
including workstation outliers, remain in the arithmetic means. Volatile rows
were expanded rather than trimmed.

| Workload | Samples/side | Baseline mean | Candidate mean | Result |
| --- | ---: | ---: | ---: | ---: |
| Boolean Reconciliation | 5 | 552.271 ms | 513.880 ms | 6.95% faster |
| Temporary custom nominal `Eq` | 10 | 242.588 ms | 247.297 ms | 1.94% slower |
| Run-length encode | 5 | 1.6177 s | 1.3575 s | 16.08% faster |
| Unicode Scalar Pipeline | 20 | 5.0138 s | 5.0240 s | 0.20% slower; neutral |
| Iterator Collect guard | 10 | 492.032 ms | 464.163 ms | 5.66% faster |
| Numeric Array Map guard | 10 | 82.692 ms | 79.313 ms | 4.09% faster |
| String Split/Join guard | 20 | 1.2366 s | 1.2788 s | 3.42% slower |
| Bounded Reverse Complement guard | 10 | 1.194 ms | 1.248 ms | 4.45% slower |

Unicode retains one 9.63-second baseline process and two 8.77-9.68-second
candidate processes. Its ordinary candidate cluster is faster, but the required
full mean is correctly reported as neutral. The first ten Split/Join samples
showed a 9.08% candidate regression after an unmatched workstation transition:
candidate sample 8 took 2.55 seconds and the subsequently recovered baseline
sample took 1.00 second. No sample was discarded; expansion to twenty samples
put the full difference inside the 5% guard.

Boolean removes three measured allocations per call. Custom nominal, Iterator,
and Numeric Array Map allocation counts are identical between sides. The small
Reverse Complement setup tail varies by a handful of allocations on both sides.

## Post-change profile

A five-call Boolean profile reduces the complete equality path from 13.87% to
4.72% cumulative. `applyCachedEqualityDispatch(...)` falls from 10.58% to
1.18%, and `callCallableValue2Mutable(...)` disappears from the equality tree.
Primitive canonicalization falls from 4.01% to 0.79%. The direct operation is
small enough to inline into its caller.

The new dominant shared candidates are outside equality. The refreshed
pre-candidate profiles put `bytecodeRawIntegerValueInfo(...)` at 3.65% Boolean,
3.36% Unicode, 10.34% Numeric Array Map, 1.82% Iterator, 1.41% custom nominal,
and 1.37% Run-length. The post-change Boolean profile still puts it at 2.36%
flat while array reads and casts dominate the cumulative tree.

## Correctness and worktree reconciliation

- Focused primitive `==`/`!=`, mixed numeric, float/NaN, custom nominal,
  operator-interface, hash/equality, native interface-resolution, and
  bytecode/tree-walker parity tests pass.
- Tests prove bool, char, String, integer, and float entries retain a direct
  primitive plan; float still selects `PartialEq`; custom nominal `Eq` retains
  an Able `FunctionValue` and no primitive plan.
- `go test ./pkg/runtime -count=1 -timeout 55s` passes.
- The all-in-one interpreter suite is not currently green in the shared dirty
  worktree: four existing truthiness/cast fixture failures and one error-message
  fixture mismatch occur identically in the clean pre-candidate and candidate
  test binaries, after which the package reaches its 55-second cap. They are not
  caused by this tranche and are not hidden as a passing result.
- Changed Go files remain below 1,000 lines and `git diff --check` passes.

## Next recommendation

Audit raw integer extraction across Boolean, Unicode, Numeric Array Map,
Iterator Collect, Run-length, and the custom nominal control. First census the
call sites and carrier shapes reaching `bytecodeRawIntegerValueInfo(...)`, then
profile descendants under casts, array reads/indexing, comparisons, and length
operations. Implement a candidate only if one redundant generic extraction
step repeats in at least three unlike programs.

Why: the equality callable wall is now closed, while raw integer extraction is
the clearest repeated leaf across numeric, text, array, iterator, and nominal
workloads. The work entails temporary shape counters, a general carrier-level
plan rather than a benchmark or container-name branch, focused mixed-width and
signed/unsigned safety tests, and the same repeated text/iterator/byte/numeric
performance gate. WASM remains deferred.
