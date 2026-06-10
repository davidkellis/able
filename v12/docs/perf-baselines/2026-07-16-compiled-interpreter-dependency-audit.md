# Compiled Interpreter Dependency Audit (2026-07-16)

## Decision

This tranche keeps no runtime or compiler code change. The generated compiled
runtime does pay a large, unused interpreter-package cost in no-bootstrap
binaries, but that cost is not reachable through one safely movable helper.
It is retained by two architectural roots:

1. generated registration and launcher signatures use the concrete
   `*interpreter.Interpreter` type; and
2. `compiler/bridge.Runtime` stores the same concrete type and exposes the
   dynamic boundary through 41 interpreter methods.

Moving one primitive operation, exit helper, thunk type, or cache initializer
would leave the interpreter package reachable. Removing only the eager bytecode
cache was already rejected because it changed normal Go GC pacing and slowed the
allocation-heavy Binary Trees control by 7.74%. No candidate met the broad
benchmark bar, so none was advanced.

## Scope and method

The audit used three unlike verifier-backed compiled applications built from the
current tree and canonical `../able-stdlib` source:

- Array Slice Window: short, array/slice and iterator work;
- TapeLang Alphabet: allocation-light sustained scalar/control-flow work; and
- Binary Trees: allocation-heavy nominal construction and concurrency.

Each application was built once with retained generated source. The audit then
used:

- generated-source import and symbol counts;
- `go list -deps -json` reverse-import inspection;
- `go tool nm -size` linker reachability;
- `GODEBUG=inittrace=1`; and
- one cumulative allocation profile per application under `GOGC=50`,
  `GOMAXPROCS=1`, the goroutine executor, and a 60-second process timeout.

The profiling launches were diagnostic one-process runs, not selection timing
samples. Each output SHA-256 matched the preceding verifier-backed build run.
Performance selections continue to require repeated alternating processes and
combined means on this workstation.

## Exact package cost

All three binaries have the same two reverse-import edges:

```text
able/compiled -> able/interpreter-go/pkg/interpreter
able/interpreter-go/pkg/compiler/bridge -> able/interpreter-go/pkg/interpreter
```

The interpreter package accounts for essentially identical linked code/data in
every application:

| Application | Binary bytes | Linked interpreter symbols | Interpreter symbol bytes |
| --- | ---: | ---: | ---: |
| Array Slice Window | 25,435,696 | 3,480 | 6,790,729 |
| Binary Trees | 24,309,312 | 3,480 | 6,790,729 |
| TapeLang Alphabet | 29,830,320 | 3,479 | 6,790,713 |

The bridge itself is small (65-91 KiB of linked symbols) and has no measurable
initializer allocation. The interpreter initializer is the common fixed wall:

| Application | Interpreter init clock | Exact bytes | Exact allocations |
| --- | ---: | ---: | ---: |
| Array Slice Window | 59 ms | 38,002,696 | 707,336 |
| Binary Trees | 58 ms | 38,002,504 | 707,336 |
| TapeLang Alphabet | 61 ms | 38,002,680 | 707,336 |

Parser, driver, and typechecker initialization together remain below 0.1 ms and
allocate only about 24 KiB. `compiler/bridge` initialization takes 0.004-0.005
ms and allocates zero bytes.

## Why there is no bounded extraction

Top-level generated source uses only a few interpreter names beyond the
concrete registration type:

- `CompiledThunk`;
- `ExecutorKindFromEnvironment`;
- `ExitCodeFromError`;
- `ApplyBinaryOperatorFast`; and
- `ApplyUnaryOperatorFast`.

That small visible list is misleading. Depending on the application, 13-53
generated files import the interpreter package, chiefly because 21-61 generated
functions accept `*interpreter.Interpreter`. The bridge separately imports the
package and its concrete runtime field calls 41 operations spanning evaluation,
calls, type matching/coercion, iterator/range/concurrency behavior, member/index
access, hashing/equality, diagnostics, and error construction.

The linker result confirms the collateral: all three applications retain the
same 3,479-3,480 interpreter symbols even though their compiled work is unlike.
The exported linked roots are only `ApplyBinaryOperatorFast`,
`ExecutorKindFromEnvironment`, `MaterializeRuntimeValues`, and
`TypecheckProgram`, but their implementation graph and the concrete bridge type
retain the monolithic package. Extracting any one root cannot remove its
initializer.

## Allocation-profile reconciliation

The cumulative allocation profiles distinguish the fixed package problem from
application work:

- Array Slice Window: 100% of the 24.35 MiB sampled allocation profile is
  interpreter initialization.
- TapeLang Alphabet: 100% of the 23.99 MiB sampled allocation profile is
  interpreter initialization.
- Binary Trees: `__able_compiled_fn_make_tree` owns 9,228.64 MiB, or 99.59% of
  the 9,266.77 MiB sampled total; interpreter initialization is immaterial to
  total allocation volume but materially changes its initial GC heap goal.

The profile totals are sampled allocation-space values; the `inittrace` rows
above are the exact initializer allocation counts. Together they explain both
sides of the rejected lazy-cache result: interpreter baggage dominates short
applications, while its live heap unintentionally reduces collection frequency
in a program allocating more than 9 GiB.

## Verification and cleanup

- All three retained build runs passed their external verifiers.
- All three allocation-profile launches reproduced the verified output hashes.
- No source was changed for an experimental candidate.
- The temporary generated trees, binaries, linker listings, and profiles were
  removed after the audit.

## Next recommendation

Reconcile generic compiled nominal-construction allocation across Binary Trees,
Sudoku Masks, and at least one unlike nominal/union-heavy application before
attempting interpreter-package isolation.

Why: the dependency audit found a real fixed 59 ms / 38 MB startup wall, but the
previous guard proved that removing it safely depends on lowering the compiled
runtime's allocation pressure rather than manipulating GC policy. Binary Trees
currently allocates 9.23 GiB almost entirely in one generated nominal
constructor and is 5.4x the Go reference, making general nominal representation
and construction the larger product gap.

What it entails: collect fresh bounded CPU/allocation profiles and generated Go
escape-analysis evidence for three unlike allocation-heavy programs; reconcile
the concrete repeated descendant in the shared nominal translation and semantic
encoding pipeline; and advance a candidate only if the same constructor,
boxing, field-carrier, or escape pattern repeats across all guards. Any retained
change must use the general nominal lowering rules, must not name a benchmark or
stdlib container, and must pass repeated alternating compiled application
averages plus compiler correctness and no-bootstrap/no-fallback tests. Once that
allocation wall is reduced, return to an interpreter-neutral dynamic-boundary
interface so no-bootstrap binaries can omit the interpreter package without GC
ballast or policy overrides.
