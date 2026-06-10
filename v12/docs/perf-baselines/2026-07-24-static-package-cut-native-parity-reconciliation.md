# Static package-cut and native-parity reconciliation

Date: 2026-07-24

## Decision

Keep no new compiler, runtime, VM, canonical-stdlib, benchmark, or language
change from this tranche. The experimental static interpreter-package cut was
reapplied, widened beyond the original six guards, found to be semantically
required by a strict no-fallback application, and reverted completely.

The tranche also corrects the premise that Binary Trees exposes a deficient
nominal representation. Its generated `Node` and constructor have the same Go
shape as the reference program, and the interpreter-free Able binary matches
the native Go reference under the normal Go GC policy. The earlier loss comes
from removing accidental heap ballast, not from an extra Able node carrier.

The newly exposed lowering gap is instead a statically known primitive value
crossing a generic callable ABI as `runtime.Value`. In Validated Job Pipeline,
the `Result i64.map<i64>` callback receives `value runtime.Value`; all callback
arithmetic consequently calls `__able_binary_op`. That operation still uses
shared primitive semantics rooted in the interpreter package even though no
AST fallback is reachable.

## Nominal and Array reconciliation

Current generated source plus the earlier exact allocation/escape profiles
agree on three unlike shapes:

- Binary Trees emits `type Node struct { Left, Right *Node }` and returns
  `&Node{...}` exactly like the Go reference. Its approximately 16-byte
  allocation per returned source node is required by Able reference identity
  and by the source's live tree.
- Sudoku Masks emits `&__able_array_i32{Elements: make([]int32, 0, 3)}` and
  direct Go `append` operations for the source-requested three-element
  position Array. The pointer header preserves Array identity and shared
  length mutation; the backing storage is a native Go slice.
- Policy Record Dispatch emits ordinary native Go structs for
  `PolicyRecord`, `PolicyDecision`, and related records, plus generated Go
  interfaces/wrappers for language unions. Its dominant allocation remains
  regex/string work, not the Binary Trees constructor or Sudoku position
  Array.

No constructor, field carrier, return escape, or semantic-encoding allocation
repeats across those three. A tree arena, fixed Sudoku tuple, or named record
rule would be application/type-specific and was not attempted.

## Binary Trees normal-GC parity

The exact-source interpreter-linked and interpreter-free Able binaries from
the prior package-cut gate were rerun in three direction-reversed pairs with
`GOMAXPROCS=4`, `GOMEMLIMIT=1GiB`, the goroutine executor, CPUs 0-3, and Go's
normal `GOGC` setting. The equivalent Go 1.26 source received three processes
under the same CPU and memory settings. Every output passed the public Ruby
verifier.

| Variant | Samples | Wall samples (s) | Mean |
| --- | ---: | --- | ---: |
| interpreter-linked Able | 3 | 7.21, 7.56, 7.68 | 7.4833 s |
| interpreter-free Able | 3 | 7.55, 7.75, 8.15 | 7.8167 s |
| native Go reference | 3 | 8.06, 7.61, 7.85 | 7.8400 s |

The package cut is 4.45% slower than the linked binary but 0.30% faster than
the Go-reference mean, which is ordinary workstation noise. This proves the
generated tree allocation is already Go-equivalent. The linked binary's
advantage comes from its unrelated approximately 38 MB initialization raising
the initial heap target. Under the diagnostic `GOGC=50` protocol that effect
is amplified into the previously recorded 10.60% loss.

This does not justify ballast or a GC-policy override. It closes
nominal-construction as the explanation for the package-cut guard.

## Strict dependency audit

The package cut again passed the strict core audit:

- Fib;
- Binary Trees;
- Matrix Multiply;
- Quicksort;
- Sudoku Masks; and
- I-Before-E.

All six compiled with `ablec --no-fallbacks`, ran, passed their public
verifiers, and reported
`interpreter_dependency: false`. Durable rows:
`2026-07-24-static-package-cut-core-final.json`.

An exploratory coverage audit then verified additional unrelated static
applications including Base64, JSON, PiDigits, K-Nucleotide, N-Body,
Fixed Width 128, Rational Series, Wide Integer Records, Word Frequency,
future/channel/mutex applications, Unicode Scalar Pipeline, Array Slice
Window, Dependency Plan, Inventory Reconciliation, Option/Result Config, and
Concurrent Text Index. Regex applications with known residual lowering were
correctly rejected by the strict build gate.

The coverage sweep was stopped when two eligible concurrency applications
timed out; it is diagnostic evidence only and was not retained as a completed
aggregate report.

## Semantic rejection guard

Validated Job Pipeline provided the decisive failure:

- interpreter-free exact binary: no output after 10 seconds, 0.05 seconds user
  CPU;
- the same generated source with only the interpreter import and the two
  `ApplyBinaryOperatorFast` / `ApplyUnaryOperatorFast` calls restored:
  correct verified output;
- three independent restored-binary runs: 3/3 verified, 1.1367-second wall
  mean, identical output hashes.

The generated `transform` function converts the statically known
`Result i64` to a runtime union value and passes its `map<i64>` closure through
`__able_fn_runtime_Value_to_runtime_Value`. The closure parameter, its
loop-carried `state`, and the results of `*`, `+`, and `%` therefore remain
`runtime.Value` and call `__able_binary_op`.

Without `interpreter.ApplyBinaryOperatorFast`, the no-bootstrap bridge has no
concrete interpreter to perform those boxed primitive operations. Control
exits the worker before it sends its completion sentinel, while `main` waits
for all sentinels. The package cut therefore changes observable behavior; it
cannot be retained merely because AST fallback is absent.

Production generator source and tests were restored before verification.
Focused compiler tests pass after the revert. No canonical stdlib change was
appropriate.

## Next recommendation

Complete native generic-callable lowering, starting with primitive
specializations of the shared `T -> U` callable ABI used by `Result.map`,
`Option.map`, `Array.map`, and iterator mapping.

Why this is next: this tranche found a concrete compiled-to-boxed transition
that remains inside a strict no-fallback binary. The source and instantiated
generic types prove that Validated Job Pipeline's callback is `i64 -> i64`,
yet generated code erases it to `runtime.Value -> runtime.Value`. That is both
a correctness dependency on shared interpreter semantics and a hot
performance boundary. Fixing it advances the user's requirement that primitive
Able values stay native Go primitives; merely moving the existing boxed
operator helper to another package would remove a linker edge without
removing the runtime boundary.

What it entails: census strict generated applications for
`__able_fn_runtime_Value_to_runtime_Value` callbacks whose instantiated
parameter/result types are primitive; require the same carrier loss in at
least three unlike applications and across more than one generic API. Add
typed generated callable forms such as `func(int64) (int64, *control)` through
the existing general generic specialization machinery, keep runtime adapters
only at actual dynamic/interface boundaries, and lower callback arithmetic
directly to Go operators. Preserve union/error propagation, closure capture,
arity, identity, overflow, and dynamic-call semantics. Gate with Validated Job
Pipeline, Dependency Wave Validation, an Array/iterator map application,
Option/Result Config, the strict core suite, and the existing six package-cut
performance guards. Only after those boxed operator sites disappear should
the interpreter-package cut be retried.
