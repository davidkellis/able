# Compiled compute/control cohort closure — 2026-07-25

## Decision

Retain no compiler, generated-runtime, runtime, interpreter, VM, canonical
stdlib, language, dependency, or WASM change.

Tapelang Alphabet, Matrix Multiply, and Fib do not share one material semantic
primitive-arithmetic, Array-access, or control-flow owner. Their strict
generated applications already use native Go carriers and omit
`pkg/interpreter`, but their remaining costs divide below that broad parent:

- Tapelang is flattened instruction dispatch, native Array reads/writes, and a
  checked `i32` tape-cell update.
- Matrix is a native `float64` multiply-add loop with explicit Able Array
  range guards.
- Fib is direct native `i32` recursion with value-plus-control returns and
  nil-control checks.

The required three-unlike-application admission gate therefore fails before
implementation. No A/B candidate or benchmark-specific rule is justified.

## Protocol

A current `cmd/ablec` binary was built once with SHA-256
`8a64cddbb3c20b341ea20205c75257b558ac05cbdfe4369c06157a00381cc30e`.
It built each application with `-build -no-fallbacks`. All build and profile
work lived under disk-backed `/var/tmp/able-compute-control-20260725`.

Every measured process used:

```text
taskset -c 5
GOMAXPROCS=1
GOMEMLIMIT=1GiB
GOGC=50
```

The applications used their public benchmark arguments and sibling Ruby
verifiers. Each row received three exact main-phase allocation runs and ten
main-only CPU-profile runs. All 39 profile outputs verified. The ten CPU
profiles per row were merged only after verification.

All three final dependency graphs contain 96 packages and omit
`able/interpreter-go/pkg/interpreter`.

## Governing scorecard rows

These are the five-run current scorecard means that admitted the profiling
work:

| Application | Able | Go | Able / Go | Absolute excess | Main GC |
| --- | ---: | ---: | ---: | ---: | ---: |
| Tapelang Alphabet | 3.8180 s | 2.0027 s | 1.9064x | 1.8153 s | 0 |
| Matrix Multiply | 1.1820 s | 0.9801 s | 1.2060x | 0.2019 s | 8 |
| Fib | 3.6340 s | 3.3510 s | 1.0845x | 0.2830 s | 0 |

## Exact allocation evidence

The three main-phase allocation runs were exact within each application:

| Application | Allocated bytes per run | Allocations per run | GC counts |
| --- | ---: | ---: | --- |
| Tapelang Alphabet | 282,984 | 4,277 | 0, 0, 1 |
| Matrix Multiply | 32,897,352 | 8,018 | 8, 8, 8 |
| Fib | 504 | 10 | 0, 0, 0 |

Fib and Tapelang confirm that their hot CPU costs are not allocation or GC
walls. Matrix's bytes are its expected native matrix backing storage; its
merged CPU profile remains 98.93% flat in the compute body rather than in GC
or conversion code.

## CPU and generated-assembly attribution

### Tapelang Alphabet

The ten profiles contain 37.06 seconds of samples:

| Generated function | Flat CPU | Flat share | Owner |
| --- | ---: | ---: | --- |
| `__able_compiled_fn_execute` | 23.47 s | 63.33% | dispatch and native Array reads |
| `__able_compiled_method_Tape_inc` | 10.29 s | 27.77% | Array read/write and checked `i32` add |
| `__able_compiled_method_Tape_get` | 2.61 s | 7.04% | inlined native Array read |
| `__able_compiled_method_Tape_move` | 0.52 s | 1.40% | checked position update and growth |

Generated disassembly confirms that `Tape.inc` reads an `int32` from the
native element slice, widens the addition, checks the signed range, and writes
the native result. The overflow branch is semantically required for arbitrary
programs and tape state. `Tape.get` is control-free and already inlined.

### Matrix Multiply

The ten profiles contain 11.23 seconds of samples.
`__able_compiled_fn_matmul` owns 11.11 seconds flat, or 98.93%. Its hot inner
loop operates on native `float64` element slices and uses direct counted-loop
increments. It retains two explicit Able range guards before its two Array
loads.

Source attribution concentrates 6.97 seconds on the two-byte `INCL`
implementing `k++`. Disassembly shows that this is not evidence of a distinct
increment-lowering owner: the multiply, add, Array loads, guards, and branch
form the complete tight loop around that sample, and no separate increment
helper or checked arithmetic remains. Treating the line sample alone as a new
loop-lowering owner would misdiagnose the generated machine code.

### Fib

The ten profiles contain 38.20 seconds of samples.
`__able_compiled_fn_fib` owns 38.06 seconds flat, or 99.63%. It takes and
returns native `int32`; retained range proof already emits its subtractions
and addition directly. The remaining generated difference from the reference
body is its `(int32, *__ableControl)` return plus nil-control checks after
recursive calls.

Fib has no Array access, allocation owner, interpreter crossing, or remaining
checked arithmetic owner.

## Control/effect admission

The existing conservative `compiled-control-census` analyzer was run over the
exact three generated modules:

| Application | Direct functions | Control-free | Main-reachable | Closed range-free | Call-site specializable |
| --- | ---: | ---: | ---: | ---: | ---: |
| Tapelang Alphabet | 291 | 62 | 13 | 4 | 0 |
| Matrix Multiply | 63 | 11 | 3 | 0 | 0 |
| Fib | 53 | 10 | 2 | 1 | 0 |

Fib's recursive body is control-free. Matrix's `matmul` remains fallible
through Array errors. Tapelang's `execute` and `Tape.inc` remain fallible
through Array/error propagation and overflow; only the already-inlined getter
is control-free. No row supplies a new call-site-specializable direct
function.

This independently reproduces the prior conclusion that a value-only control
ABI is material only in Fib. The current profiles provide no new
cross-application evidence to reopen that closed one-family route.

## Candidate reconciliation

| Candidate | Result | Reason |
| --- | --- | --- |
| Value-only control ABI | rejected | material only in Fib; the other hot bodies are semantically fallible |
| Native Array range-guard elimination | insufficient evidence and breadth | hot Array machinery and guards occur in Tapelang and Matrix, but Fib has none and Matrix guard-only materiality is not isolated |
| Checked `i32` arithmetic removal | insufficient breadth | material in Tapelang; already removed in Fib and Matrix hot loops |
| Counted-loop increment lowering | closed | direct increments already emit; Matrix's `k++` sample is aliasing |

Combining these different descendants under “native compute” would erase the
semantic distinction required by the project gate. The correct result is no
code.

## Verification and artifact identity

- 3 strict builds and 3 public-verifier smoke runs passed.
- 9 exact-allocation and 30 CPU-profile processes passed public verification.
- 3 final dependency graphs omit `pkg/interpreter`.
- 3 control-census reports parsed the exact generated modules.
- `go test ./cmd/compiled-control-census -count=1 -timeout=60s` passes in
  0.002 seconds.
- `go test ./cmd/ablec -count=1 -timeout=60s` passes in 5.835 seconds.

| Application | Binary SHA-256 | Generated `compiled.go` SHA-256 |
| --- | --- | --- |
| Tapelang Alphabet | `5baf0bc718836377c50092ae2f8fd814ceb9db3139d4f7b7f4b40504ec9d168d` | `a950094a7257d2c072f8e5becd45df88f57372f114855f05f7685d1dc95e0e91` |
| Matrix Multiply | `d0d5e7a74f2352f91998b8877f3b263f6b15b736a74f2724d0559924398a75e9` | `21ae097934336dbfcc4dccfcf00c90cc2d69bfc37017c9324bc190739af12d5f` |
| Fib | `0204de5ed0624b7398d172db52b8f149a9e73e4181a5717465aa08505eb1397a` | `518ce277cdf7207354d89b753b3f3632a699ec9b4d303ec64b7d6dd0801da458` |

The machine-readable companion is
`2026-07-25-compiled-compute-control-cohort-closure.json`. Raw generated
modules, binaries, profiles, and captured outputs are disposable after this
record.

## Next recommendation

Run a coverage-wide native Array range-proof admission census across current
strict compiled target misses.

Why: the fresh profiles establish hot native Array-access machinery and
explicit guards in two unlike applications, Tapelang and Matrix, but not in
Fib. They do not yet isolate guard-only materiality in Matrix. Implementation
still requires a third unlike, material application and the same provable
index/length relationship.

What it entails: inventory generated explicit Array guards across current
strict misses; intersect their exact sites with repeated CPU profiles; and
classify locally provable facts such as `0 <= i < array.len()` separately from
relational facts between different Arrays. Admit a prototype only when the
same syntax- and fact-based proof removes material guards in at least three
unlike applications. Preserve negative-index, out-of-range, mutation,
evaluation-order, nullable/error, alias, generic, interface, callback, and
dynamic-boundary behavior, then require twenty order-balanced
baseline/candidate/Go cohorts.

Why it is important: native Go Array carriers alone do not guarantee Go-like
machine code if redundant Able semantic guards remain after a general proof
can establish safety. This census can either identify a broad compiler rule
that closes that gap or reject the route without speculative or
benchmark-specific lowering. Do not begin WASM work.
