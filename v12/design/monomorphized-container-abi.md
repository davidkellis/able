# Monomorphized Container ABI (Array<T>, Go AOT)

Date: 2026-02-26  
Owners: Able compiler/runtime contributors

## Purpose

Define a staged ABI plan to lower statically-typed `Array<T>` hot paths into host-native Go representations while preserving v12 semantics and explicit dynamic-boundary behavior.

This design directly supports:

- `spec/full_spec_v12.md` AOT boundary contract (no interpreter fallback for static code),
- reduced `runtime.Value` usage in static compiled paths,
- host-native polymorphism-first lowering for compiled Go targets.

## Problem Statement

Current Go AOT output already removes major dynamic fallback paths, but `Array` data still flows through dynamic carrier storage (`[]runtime.Value` via `ArrayStore*`) even when element type is statically known. This causes:

- per-element boxing/unboxing on hot `push/get/set/index` paths,
- avoidable GC pressure in sieve-style workloads,
- extra conversion overhead that limits compiled-mode speedups.

## Goals

- Keep language-visible `Array` semantics identical (aliasing, mutation, bounds behavior, clone/reserve/len/capacity behavior).
- Preserve explicit dynamic-boundary conversion behavior (`dynimport`, `dyn.eval`, etc.).
- Lower statically-known container paths to host-native typed storage first.
- Gate rollout behind a compiler/runtime flag so behavior can be validated incrementally.

## Non-Goals

- No semantic changes to `Array` in the language spec.
- No requirement to monomorphize every type in stage 1.
- No interpreter behavior changes in this design unit.

## Semantic Constraints (Must Hold)

1. Assignment aliasing must remain handle-backed: `b = a` shares storage identity.
2. Mutations through one alias are visible through all aliases to the same handle.
3. Bounds and error behavior must match existing kernel/stdlib contracts.
4. Crossing compiled<->dynamic boundary must preserve existing conversion/error semantics.
5. Static code must not silently route into interpreter fallback.

## Proposed ABI Shape

### 1) Keep Able-level struct shape unchanged

`Array T` remains:

```able
struct Array T {
  length: i32,
  capacity: i32,
  storage_handle: i64
}
```

This preserves source-level behavior and existing generated code assumptions.

### 2) Introduce typed storage classes behind handles

Add monomorphized stores in Go runtime for staged element types:

- `i32`
- `bool`
- `u8`

Each typed store keeps concrete slices (`[]int32`, `[]bool`, `[]byte`) and capacity/length semantics equivalent to current `ArrayStore`.

### 3) Handle tagging (runtime-internal)

Use tagged `storage_handle` values to route operations:

- legacy dynamic store handle (existing path),
- monomorphized typed store handle (new path).

Tagging is runtime-internal and opaque to Able programs.

### 4) Runtime API additions (Go runtime package)

Add typed operations for staged types (example names):

- `MonoArrayNewI32`, `MonoArraySizeI32`, `MonoArrayCapacityI32`
- `MonoArrayReadI32`, `MonoArrayWriteI32`, `MonoArrayReserveI32`, `MonoArrayCloneI32`

Equivalent sets for `bool` and `u8`.

Also add boundary adapters:

- typed store -> `[]runtime.Value` (for explicit dynamic-boundary crossing),
- `[]runtime.Value` -> typed store (when re-entering typed compiled contexts, if safe).

## Compiler Lowering Rules

1. Detect statically monomorphic `Array<T>` where `T` is in staged set (`i32`, `bool`, `u8`).
2. Lower hot operations to typed runtime APIs:
   - literals,
   - index read/write,
   - `push/get/set/len` intrinsics,
   - tight loops where static type is preserved.
3. Keep existing dynamic carrier path for non-staged element types or unresolved polymorphic cases.
4. Preserve host-native polymorphism priority:
   - concrete/static direct call,
   - host-native polymorphism (Go interfaces/generic specialization),
   - dynamic carrier adaptation only for residual non-representable cases.

## Dynamic Boundary Rules

Typed stores are internal compiled representation. At explicit dynamic boundaries:

- convert to dynamic values using existing boundary contract,
- surface conversion mismatch as `Error` (no silent coercion),
- preserve mutation/identity semantics required by spec.

## Rollout Plan (Flag-Gated)

Feature flag (naming proposal):

- compiler option: `--experimental-mono-arrays`
- env mirror for tests: `ABLE_EXPERIMENTAL_MONO_ARRAYS=1`

Stages:

1. **Stage 0 (scaffold)**:
   - add runtime typed stores + tag routing under flag,
   - no default behavior change.
2. **Stage 1 (hot-path subset)**:
   - enable compiler lowering for `Array i32|bool|u8` on literals/index/push/get/set/len.
3. **Stage 2 (wider coverage)**:
   - extend to more primitives and typed stdlib collection internals.
4. **Stage 3 (default-on candidate)**:
   - promote to default after parity/perf gates remain green.

## Validation Gates

### Correctness

- Existing compiler strict static checks remain green.
- `./run_stdlib_tests.sh` remains green (treewalker + bytecode reference behavior unchanged).
- Compiler regression tests prove no reintroduction of dynamic helper paths in staged typed-array codegen.
- Dynamic-boundary fixtures still pass with explicit conversions.

### Performance

Use `v12/bench_perf` report-only script:

- `noop`, `sieve_count`, `sieve_full` compiled mode baselines,
- interpreter modes may timeout (expected today for sieve benchmarks).

Promotion criteria for stage 1:

- measurable compiled `sieve_count` improvement vs pre-flag baseline,
- measurable compiled GC count reduction on `sieve_count`,
- unchanged program output.

## Risks

- Handle-tag routing bugs could violate aliasing semantics.
- Boundary conversion bugs could leak typed/internal representations.
- Partial lowering may create mixed-path complexity; tests must pin exact behavior.

## Open Questions

1. Should boundary adapters eagerly normalize typed handles to dynamic handles, or keep dual representation alive across boundary-return paths?
2. What is the minimal staged type set after `i32|bool|u8` that yields clear gains (`i64`, `f64`, `String`)?
3. Do we need dedicated diagnostics when monomorphic lowering is skipped in an otherwise eligible context?

## Commit-Ready Checklist

- [ ] Add runtime typed-array store types + handle-tag router (flag-gated).
- [ ] Add typed runtime APIs for `i32|bool|u8` create/read/write/len/capacity/reserve/clone.
- [ ] Add boundary conversion helpers typed<->dynamic for explicit dynamic crossing.
- [ ] Add compiler flag plumbing (`--experimental-mono-arrays`, env mirror).
- [ ] Lower `Array i32|bool|u8` literals/index/push/get/set/len to typed APIs under flag.
- [ ] Add compiler regression tests for typed-path emission and helper non-regression.
- [ ] Add boundary regression fixtures for explicit conversion behavior.
- [ ] Record pre/post perf snapshots with `v12/bench_perf` (`noop`, `sieve_count`, `sieve_full`).
- [ ] Update `spec/TODO_v12.md` with stage status and any residual limits.
