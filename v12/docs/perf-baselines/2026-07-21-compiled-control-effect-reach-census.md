# Compiled control-effect reach census

Date: 2026-07-21

## Decision

Retain the report-only `compiled-control-census` analyzer and keep no compiler,
generated-runtime, bridge, stdlib, application, reference, language, or WASM
performance change.

Control-free direct functions are statically broad, but a value-only internal
ABI is CPU-material in only one of seven unlike current applications. Fib's
recursive body is genuinely control-free and its non-inlined nil-control
propagation is material. Other eligible functions are cold, already inlined,
or wrappers around a different dominant Go/bridge kernel. The three-family
reach gate therefore fails before an ABI prototype.

## Retained analyzer contract

`cmd/compiled-control-census` parses the root generated-Go package with the Go
parser and computes a conservative greatest fixed point over every local
function returning `*__ableControl`.

A function is control-free only when every externally returned control is:

- literal `nil`; or
- produced by another function in the same control-free fixed point.

This handles safe recursive strongly connected components such as Fib.
Unknown, dynamic, host, or directly constructed control sources remain
fallible. A function may call and consume a fallible operation internally and
still be externally control-free when every outward return is nil. The result
describes only the explicit `__ableControl` ABI; it does not claim that an
arbitrary Go host call cannot panic.

The JSON report includes each function's dependencies, conservative hazards,
direct call sites, and aggregate direct-compiled/control-free counts. The tool
does not rewrite or execute generated code and is inert outside explicit audit
runs.

## Protocol

The immediately preceding seven-application exact-leaf tranche supplied two
verified main-only CPU processes per application. No compiler or runtime code
changed afterward, so those source-exact merged profiles remain the reach
evidence rather than repeating identical workstation timing.

The same frozen compiler was rebuilt at SHA-256
`87535396e8c8a471f3dd84d44ce7678554426f1ac84d5e730d729349f83fa913`.
It emitted fresh normal Go modules against canonical `../able-stdlib` with
`ABLE_SOURCE_ROOT_ONLY=1`, `GOMAXPROCS=1`, `GOGC=50`, a 1-GiB limit, and a
59-second cap. Each root generated package was analyzed, then deleted before
the next application to bound disk use.

The compiler hash was checked against the preceding tranche's frozen artifact.

## Static fixed-point census

| Application | Direct functions | Control-free | Direct call sites | Calls to control-free direct functions |
| --- | ---: | ---: | ---: | ---: |
| `fib` | 53 | 10 | 59 | 16 |
| `sudoku_masks` | 287 | 58 | 630 | 127 |
| `base64` | 63 | 11 | 68 | 14 |
| `pidigits` | 232 | 75 | 459 | 145 |
| `k_nucleotide` | 416 | 90 | 959 | 241 |
| `tapelang_alphabet` | 291 | 61 | 642 | 140 |
| `mandelbrot` | 175 | 36 | 366 | 83 |

Static breadth is not dynamic materiality. Generated modules include imported
stdlib bodies, so broadly emitted helpers such as `read_byte`, conversion
utilities, path helpers, and String-builder operations can be control-free but
cold in the measured main.

## Hot reach reconciliation

| Application | Relevant classification | Main-profile result | Admission result |
| --- | --- | --- | --- |
| `fib` | recursive `fib` is control-free | 99.86% flat; nil-control checks are 18.13% | material |
| `sudoku_masks` | `bit_count`, `square_index`, search, and solve are fallible | checked arithmetic/search own the profile | not eligible |
| `base64` | small string/hex wrappers are free; encode/decode byte paths are fallible | Go Base64/MD5 kernels dominate | ABI checks not material |
| `pidigits` | several native BigInt helper bodies are free | `math/big` kernels dominate; no material generated check | ABI checks not material |
| `k_nucleotide` | `nucleotide_code` is free; target/count/map paths are fallible | HashMap equality/hash/boxing and allocation dominate | ABI checks not material |
| `tapelang_alphabet` | `Tape.get` is free; `Tape.inc`, move, and execute are fallible | `Tape.get` is inlined; mutation/dispatch dominate | value-only ABI already optimized away locally |
| `mandelbrot` | `pixel_byte` is fallible through checked shift | direct float body dominates the small profile | not eligible |

The classifications preserve Able semantics rather than observed-input
outcomes:

- Sudoku `bit_count` subtracts from an `i32` and retains overflow control even
  though current call sites pass small masks.
- `square_index` retains signed division/multiplication control.
- Mandelbrot's pixel helper retains checked signed-shift control.
- Tape mutation retains possible arithmetic overflow.

Removing those controls merely because public benchmark inputs do not trigger
them would be an unsound benchmark specialization. Proving bounded arguments
at every direct call site is a separate interprocedural range question.

Fib is the sole material eligible row. Tape's eligible getter is already
inlined, allowing Go to eliminate its nil result/check without a new ABI.
Pidigits and Base64 similarly reach host kernels below small generated
wrappers. Static control-free call-site counts therefore cannot be used as a
speedup proxy.

## Verification

- analyzer tests prove a safe recursive SCC remains control-free;
- a direct constructed control and its propagating caller remain fallible;
- a constructed-control reassignment after a safe result remains fallible;
- a function that consumes a fallible result but returns only nil control is
  classified externally control-free;
- all seven fresh generated modules parsed and produced deterministic census
  reports; and
- focused analyzer tests pass in 0.002 seconds.

No generated source, performance candidate, benchmark, canonical stdlib, or
reference changed. Temporary generated modules and audit binaries are removed
after recording the results.

## Reproduction

```text
cd v12/interpreters/go
go test ./cmd/compiled-control-census -count=1
go run ./cmd/compiled-control-census -output /tmp/control-census.json \
  /path/to/generated/module
```

## Next recommendation

Run an interprocedural primitive range/effect closure census before abandoning
the control-free ABI direction.

Why: Fib proves the ABI cost can be material, while Sudoku and Mandelbrot are
currently excluded by checked primitive operations whose arguments may be
provably bounded at all direct call sites. A general range proof could create
three-family eligibility without weakening error semantics; observed benchmark
inputs alone cannot.

What it entails: report the first fallibility blockers for hot direct
functions; propagate existing integer facts and call-site argument ranges
through the direct call graph; distinguish universally safe bodies from
call-site-specializable bodies; and join the proven closures with the existing
main-profile samples. Require material proven control-free calls in three
unlike families before emitting a value-only internal specialization. Preserve
the ordinary fallible wrapper for dynamic, interface, callback, exported, and
unproven calls, and guard overflow, division-by-zero, shifts, recursion,
raise/rescue/or-else, nullable/error, and host boundaries. Do not add function-
name, benchmark, nominal/container, stdlib, GC, or WASM special cases.
