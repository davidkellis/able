# Compiled closed-benchmark source audit — 2026-07-15

## Scope

The seven established compiled benchmark families now have one shared
generated-source audit:

- Fib
- Binary Trees
- Matrix Multiply
- Quicksort
- Sudoku
- I Before E
- Pidigits

The audit compiles each canonical source under `v12/examples/benchmarks` with
the ordinary loader and canonical stdlib search paths.  Compilation requires
zero fallbacks.

It then inspects the source-owned compute kernels rather than the program
entry points.  `main` may legitimately cross file, argument, output, or
scheduler boundaries; those are not evidence that a numerical, recursive, or
collection compute kernel needs dynamic lowering.

## Guard

Three independently bounded tests (`...Numeric...`, `...Structural...`, and
`...Search...`) require each selected kernel to avoid the dynamic object
carrier and generic dynamic dispatch helpers:

- `runtime.Value` and `[]runtime.Value`
- callable, named-call, member-get/set, and method-call dynamic helpers
- broad `any` conversion
- ordinary call-frame push/pop scaffolding

The audited kernels cover recursive tree construction/checking, matrix
construction/multiplication, parsing and quicksort recursion, Sudoku board
search, text checking, and Pidigits arithmetic updates.  The test is a shared
static-lowering regression guard: it contains no benchmark-specific compiler
branch and does not assert a particular timing result.

The guard initially exposed two common lowering gaps rather than a benchmark
exception:

- an unannotated function whose body is inferred as `nil` used a
  `runtime.Value` return carrier, including bare returns and bare loop breaks;
- a `loop` used only as a statement still allocated a discarded result carrier.

The compiler now uses `runtime.NilValue` only for the first case—an inferred,
unannotated function return—so declared/generic `nil` types keep their shared
nominal/runtime boundary behavior. Discarded loops compile directly; a
`break value` still evaluates and discards that value so effects and errors are
preserved. These are language-level control-flow rules, not a container or
benchmark specialization.

## Verification

Under `GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`:

```text
go test ./pkg/compiler -run '^TestCompilerClosedBenchmarkNumericComputeKernelsStayStatic$' -count=1 -timeout 55s
go test ./pkg/compiler -run '^TestCompilerClosedBenchmarkStructuralComputeKernelsStayStatic$' -count=1 -timeout 55s
go test ./pkg/compiler -run '^TestCompilerClosedBenchmarkSearchComputeKernelsStayStatic$' -count=1 -timeout 55s
go test ./pkg/compiler -run '^(TestCompilerInferredNilLoopReturnUsesNativeCarrier|TestCompilerDiscardedLoopBreakValueStillEvaluates|TestCompilerCountedLoopStatementAvoidsDiscardedRuntimeValueProbe|TestCompilerLoopExpressionBreakValuesInferNativeUnion)$' -count=1 -timeout 55s
```

passed. The focused execution tests cover native inferred-`nil` returns,
bare-loop exits, and evaluation of a discarded `break value`. Each source-audit
group is independently bounded so a cold compiler cache cannot turn the
combined audit into one overlong test.

## Post-change checkpoint

A fresh verifier-backed, CPU-14-pinned three-run compiled checkpoint completed
for the two directly affected external applications:

| Benchmark | Able compiled | Go reference | Able/Go | Validation |
| --- | ---: | ---: | ---: | --- |
| Quicksort | 1.6600 s | 2.0100 s | 0.83x | 3/3 verified |
| Pidigits | 1.0767 s | 0.7400 s | 1.46x | 3/3 verified |

The evidence is retained in
`v12/docs/perf-baselines/2026-07-15-compiled-loop-carrier-benchmark.{json,md}`.
It is a valid post-change checkpoint, not a before/after delta: no same-host
pre-change sample was retained, so it does not by itself quantify the gain.

## Decision

The closed compiled families now have direct generated-source coverage in
addition to their external scorecard rows.  Future compiler changes that
reintroduce an object carrier or dynamic dispatch into one of these static
kernels fail locally before a timing regression has to be diagnosed.  This
does not reopen performance selection: future performance claims still need a
same-lane baseline plus the external scorecard’s material cross-cutting change
or three-unlike-program shared-hotspot evidence.
