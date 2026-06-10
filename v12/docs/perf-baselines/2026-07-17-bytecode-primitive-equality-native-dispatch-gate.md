# Bytecode primitive equality native-dispatch gate — 2026-07-17

## Decision

Keep primitive-first equality resolution for `==` and `!=`. The operator path
now uses the interpreter's existing primitive interface matrix and native
`Eq`/`PartialEq` callables before consulting Able implementation bodies.
Non-primitive values retain the existing interface resolver and custom nominal
implementations continue to execute as Able functions.

This reconciles operator dispatch with explicit interface calls and kernel
hash-map equality, which already prefer the same primitive-native resolver. It
is a primitive language/kernel boundary, not a benchmark, stdlib container, or
application-name special case. No compiler, stdlib, language, fixture, or
retained benchmark-source change was needed.

## Semantic basis

Section 14.1.5 of the v12 specification makes primitive implementations
intrinsic: boolean, integer, character, and string `Eq` implementations are
always in scope and cannot be redefined; floats implement `PartialEq` but not
`Eq`. The interpreter already encodes that matrix in
`primitiveImplementsInterfaceMethod(...)` and implements the corresponding
native semantics in `primitiveEqNativeMethod(...)`.

Before this change, `applyEqualityInterface(...)` bypassed that boundary and
called `findMethodCached(...)` directly. As a result, canonical primitive
comparisons selected Able `FunctionValue` bodies and paid general invocation,
frame, return, and coercion costs. Explicit interface lookup and hash-map
lookup did not have this inconsistency.

The retained resolver:

- asks the primitive matrix whether the current `Eq` or `PartialEq` candidate
  is implemented;
- skips nominal lookup for a recognized primitive when the current interface
  is not implemented, allowing floats to proceed from `Eq` to `PartialEq`;
- turns the existing receiver-relative native template into the unbound
  two-argument callable expected by operator dispatch;
- falls back unchanged to the ordinary interface resolver for every
  non-primitive value.

The arity adjustment is shared with `resolveInterfaceMethod(...)`; it does not
introduce a second native-call convention.

## Repeated performance gate

Every timing is an independent process with one warmup and one measured call,
the canonical external `able-stdlib`, `GOMAXPROCS=1`, `GOGC=50`,
`GOMEMLIMIT=1GiB`, CPU 0, and skipped benchmark typechecking. Samples alternate
order. All workstation outliers remain in the arithmetic means; volatile short
controls use ten samples per side.

| Workload | Samples/side | Baseline mean | Candidate mean | Result |
| --- | ---: | ---: | ---: | ---: |
| Boolean Reconciliation | 5 | 708.567 ms | 514.284 ms | 27.42% faster |
| Temporary custom nominal `Eq` | 5 | 215.232 ms | 213.084 ms | 1.00% faster |
| Run-length encode | 5 | 2.3956 s | 1.3512 s | 43.60% faster |
| Unicode Scalar Pipeline | 3 | 6.4315 s | 4.8371 s | 24.79% faster |
| Iterator Collect guard | 10 | 435.933 ms | 434.245 ms | 0.39% faster |
| Numeric Array Map guard | 10 | 75.217 ms | 74.226 ms | 1.32% faster |
| String Split/Join guard | 10 | 1.0077 s | 1.0233 s | 1.55% slower |
| Bounded Reverse Complement guard | 10 | 1.061 ms | 0.835 ms | 21.32% faster |

The Unicode candidate mean retains a 5.573-second candidate process between
two 4.45-second samples. Iterator retains a 516-millisecond candidate process.
Array Map retains a 121-millisecond baseline process and a 91-millisecond
candidate process. Reverse Complement retains three 1.46-1.59 millisecond
baseline processes. No sample was discarded.

Split/Join initially appeared 6.9% slower over three baseline-first pairs, but
CPU profiles showed no material equality path and one profiled pair favored the
candidate. Expanding to ten samples with candidate-first pairs reduced the full
mean difference to 1.55%, within the established 5% broad guard.

The full Base64 application was not rerun because it allocated about 2.2 GB per
measured call and exhausted swap in the preceding tranche. A fixed small FASTA
input provided the bounded byte/array/index/file-I/O guard instead.

The nominal control keeps exactly 52 allocations and 40,176 bytes per measured
call on both sides. Iterator and Array Map allocation counts are also identical.
The primitive-heavy rows remove a small native-dispatch setup tail in addition
to their CPU gains.

## Post-change profile

A five-call Boolean CPU profile puts the complete equality interface path at
12.13% cumulative, down from the preceding clean baseline's 36.59%.
`applyCachedEqualityDispatch(...)` is 8.20% cumulative and the generic
two-argument callable shell is 7.87%; the native equality body itself is 3.93%.
The detached Able equality method VM/frame subtree is gone.

Within the remaining native path, `primitiveCanonicalValue(...)` is 2.95%
cumulative, primitive argument coercion is 1.64%, primitive value comparison is
1.31%, and equality-cache lookup is 1.97%. Those are follow-up evidence, not
separate changes admitted by this tranche.

## Correctness and cleanup

- New tests prove bool, char, String, integer, and float operators cache native
  callables; float dispatch selects `PartialEq`.
- A custom nominal `Eq` test proves its cached callable remains an Able
  `FunctionValue`.
- Mixed integer/float equality, IEEE NaN and signed-zero behavior, custom
  `Hash`/`Eq`, operator interfaces, cached equality coercion, primitive explicit
  interface lookup, and bytecode/tree-walker parity tests pass.
- The complete `pkg/interpreter` and `pkg/runtime` suites pass within their
  55-second bounds.
- Changed Go files remain below 1,000 lines and `git diff --check` passes.
- The custom nominal program, FASTA input, profiles, and test binaries are
  removed after recording the result.

## Next recommendation

Refresh bounded CPU profiles for Boolean, Run-length, Unicode, custom nominal
`Eq`, Iterator Collect, and Numeric Array Map after native primitive dispatch.
Specifically reconcile the remaining generic native-call shell,
`primitiveCanonicalValue`, raw integer extraction, cached call-name work,
array reads, and casts. Admit a direct cached primitive-equality plan only if
the callable shell and duplicate canonicalization remain material in at least
three unlike primitive consumers.

Why: this tranche removed the detached Able frame wall and changed the profile
substantially; immediately optimizing the old call tree would risk chasing a
stale owner. The next work entails fresh CPU-only profiles, per-primitive call
shape counts, mixed numeric/float/nominal safety tests for any selected plan,
and the same repeated text/iterator/byte/numeric performance gate. WASM remains
deferred.
