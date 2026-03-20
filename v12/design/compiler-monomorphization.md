# Compiler Monomorphization Design

This note refines [compiler-native-lowering.md](./compiler-native-lowering.md)
for generic/container-heavy code generation.

## Goal

Use monomorphization and static specialization to keep compiled code on native
Go carriers instead of dynamic/interpreter carriers.

Monomorphization is a means to the native-lowering end state. It is not a
license to keep `Array<T>` on top of `runtime.ArrayValue`, `ArrayStore*`,
`runtime.Value`, or `any` on otherwise static paths.

## Representation Targets

| Able type | Target compiled representation |
| --- | --- |
| `Array i32` | compiler-owned wrapper around `[]int32` |
| `Array String` | compiler-owned wrapper around `[]string` |
| `Array Point` | compiler-owned wrapper around `[]*Point` |
| `Array (A | B)` | compiler-owned wrapper over native union-carrier elements |
| `Point` | `*Point` |
| `A | B` | generated Go interface plus native variants |
| `?Point` | nil-capable `*Point` |

`any` is acceptable only as a temporary implementation escape hatch or at an
explicit dynamic boundary. It is not the target representation for compiled
union values or generic containers.

## Arrays

### Required direction

- Static compiled arrays must use compiler-owned native Go storage.
- Array literals should become native wrapper construction over typed Go slices.
- `len`, `push`, `pop`, `get`, `set`, indexing, iteration, and cloning should
  operate directly on that native storage.
- The compiler must not treat kernel `Array { length, capacity, storage_handle }`
  plus runtime handles as the authority for static array storage.

### Boundary rule

If a compiled array crosses into explicit dynamic behavior, the compiler may
generate adapters:

- native specialized array wrapper -> runtime/kernel boundary value
- boundary value -> native specialized array wrapper

That conversion is boundary logic only. It must not leak back into the default
static lowering path.

### Revised mono-array design reference

The detailed staged plan now lives in
[monomorphized-container-abi.md](./monomorphized-container-abi.md).

Important update from the 2026-03-19 audit:

- the older typed-runtime-store / handle-tag rollout plan is superseded as the
  final architecture;
- future mono-array work must target compiler-owned specialized wrappers over
  native Go slices;
- existing runtime-store experiments are historical scaffolding only.

## Unions And Generics

For generic code that cannot be fully specialized yet:

- prefer generated Go interfaces and native wrapper types;
- use `any` only as a temporary residual fallback;
- document every residual `any` use as a staged limit, not as the target ABI.

Pattern matching should eventually compile to native Go type checks over those
generated interfaces/variants rather than runtime-value inspection.

## Current Work To Avoid

The following are not acceptable as final monomorphization outcomes:

- `Array<T>` lowering that still depends on `runtime.ArrayValue` /
  `ArrayStore*` on static paths;
- typed runtime-store / handle-tag routing as the primary compiled
  representation of mono arrays;
- compiler-only rewrites of the kernel `Array` struct shape that hide
  `storage_handle` behind `Elements []runtime.Value` and stop there;
- union lowering that stops at `any` instead of generating native carrier
  interfaces;
- struct-local boxing into `runtime.Value` to paper over missing static ABI
  work.

## Execution Order

1. [x] Freeze the compiler-native mono-array ABI in docs and guardrail tests.
2. [x] Emit specialized array wrapper types for staged element kinds.
3. [x] Specialize the first direct typed hot-path slice over those wrappers:
   explicit typed literals, `push/get/set/read_slot/write_slot`, direct index
   read/write, and wrapper/lambda/runtime boundary conversion.
4. [x] Widen specialization coverage to constructors / stdlib factories,
   unannotated local inference, clone / iteration / pattern paths, and other
   residual generic `*Array` surfaces.
5. [x] Re-measure compiled benchmark deltas once the widened slice lands.
   - Recorded in
     `v12/docs/perf-baselines/2026-03-19-mono-array-widened-compiled.md`.
   - Result: `bench/noop` and `bench/sieve_count` stayed flat on wall-clock,
     while `bench/sieve_full` stayed flat on wall-clock but dropped timed GC
     from `3.00` to `1.00`.
   - Interpretation: the widened slice is reducing allocation pressure on the
     heaviest staged array benchmark, but the next material runtime win now
     depends on shrinking residual generic `*Array` / `runtime.Value` paths.
6. [x] Narrow the first residual generic/`runtime.Value` array slice until the
   remaining static mismatches are either closed or explicitly logged.
   - Inferred local `Array T` bindings now keep enough recoverable metadata for
     generic static helper results (`get`, `pop`, `first`, `last`,
     `read_slot`) to stay on native nullable carriers where representable.
   - Static and residual runtime-backed array `set` / index-assignment success
     is back in parity at `nil`, restoring the compiled `nil | IndexError`
     contract for the array helper fixtures.
   - Runtime-backed iterator interface boundaries now accept raw
     `*runtime.IteratorValue` carriers directly, and generator stop now stays
     iterator completion instead of becoming a generic runtime error.
   - Result: the previously-blocked `06_08_array_ops_mutability`,
     `06_12_02_stdlib_array_helpers`, and
     `06_12_18_stdlib_collections_array_range` compiler fixtures are green
     again on the current native-lowering arc.
7. [x] Extend the staged specialized array set to `f64` and re-measure a
   nested `Array (Array f64)` benchmark target.
   - `Array f64` now lowers to `*__able_array_f64` on the staged specialized
     path.
   - Native nullable propagation now handles pointer-backed scalar carriers
     such as `*float64` when a concrete `float64` is required, which keeps
     nested expressions like `rows.get(j)!.get(i)!` on the static path instead
     of forcing the surrounding `push(...)` back through
     `__able_method_call_node(...)`.
   - The full compiled `v12/examples/benchmarks/matrixmultiply.able` path no
     longer aborts early under mono arrays; at `60s` it now matches mono-off
     by timing out rather than failing.
   - New reduced benchmark fixture:
     `v12/fixtures/bench/matrixmultiply_f64_small/main.able`
   - Snapshot:
     `v12/docs/perf-baselines/2026-03-19-mono-array-f64-matrixmultiply-small-compiled.md`
   - Result: mono on `5.4833s` / `280.00` GC vs mono off `45.3133s` /
     `3568.67` GC over 3 compiled runs.
8. [x] Remove the remaining generic outer row shell for nested staged
   mono-array carriers.
   - Nested typed arrays such as `Array (Array f64)` now lower to dedicated
     outer wrappers like `*__able_array_array_f64` over `[]*__able_array_f64`
     instead of generic `*Array` with `[]runtime.Value` row storage.
   - Rendered mono-array converters and native `Array` core helpers now
     handle pointer-backed specialized elements directly, including nilable
     `read_slot` / `pop` results and explicit nested runtime boundary
     conversion.
   - Snapshot:
     `v12/docs/perf-baselines/2026-03-19-mono-array-nested-wrapper-compiled.md`
   - Result: mono on `5.7233s` / `252.00` GC vs mono off `44.5167s` /
     `3550.67` GC over 3 compiled runs.
   - Interpretation: this closes the last generic outer-row carrier on the
     representative nested static path, but does not add a second visible
     wall-clock step beyond the earlier `f64` slice by itself.
9. [x] Widen compiler-owned array wrapper synthesis to broader native carrier
   element families.
   - Arrays of generic inner arrays, native interface carriers, native
     callable carriers, and other representable pointer-backed carrier types
     now synthesize dedicated outer wrappers instead of falling back to
     generic `*Array` / `runtime.Value` storage.
   - Dynamic-boundary callback coverage now explicitly includes
     interface-carrier arrays and callable-carrier arrays.
   - This tranche is architectural rather than benchmark-driven; no new
     shared benchmark snapshot was added because the current benchmark set does
     not materially exercise these carrier families.
10. [x] Extend the staged specialized scalar set to the text family and
    re-measure it on a char-heavy compiled benchmark.
   - `Array char` now lowers to `*__able_array_char` over `[]rune`.
   - `Array String` now lowers to `*__able_array_String` over `[]string`.
   - Nested `Array (Array char)` now lowers to `*__able_array_array_char`.
   - Native result propagation for `!Array char` now re-wraps specialized
     success branches through static coercion instead of routing them back
     through `_from_value(__able_runtime, ...)`.
   - New reduced benchmark fixture:
     `v12/fixtures/bench/zigzag_char_small/main.able`
   - Snapshot:
     `v12/docs/perf-baselines/2026-03-19-mono-array-zigzag-char-small-compiled.md`
   - The initial mono-off comparison from this slice was invalid because the
     mono-off compiled path for nested text rows was producing the wrong
     result.
11. [x] Preserve compiler-owned carrier-array identity even with staged scalar
    mono arrays disabled, and correct the text benchmark baseline.
   - Carrier-array wrappers for already-native compiler carriers now synthesize
     regardless of `ExperimentalMonoArrays`, while staged scalar
     specializations remain flag-gated.
   - This fixes the mono-off `Array (Array char)` bug where the outer generic
     `*Array` boxed rows through `runtime.Value` and then lost row mutation
     identity on `rows[idx]!.push(...)`.
   - Corrected benchmark result: mono on `0.9567s` / `88.00` GC vs mono off
     `1.0500s` / `384.00` GC over 3 compiled runs.
   - Interpretation: with a valid mono-off baseline, the text-scalar slice is
     directionally correct and does not require a text-path rollback.
12. [x] Extend the staged specialized scalar set to the remaining primitive
    numeric family and benchmark a representative unsigned path.
   - Staged wrappers now also cover `Array i8`, `Array i16`, `Array u16`,
     `Array u32`, `Array u64`, `Array isize`, `Array usize`, and `Array f32`.
   - Focused regressions pin the generated wrappers for those types and cover
     dynamic-boundary callback conversion for `Array u32` / `Array f32`.
   - New reduced benchmark fixture:
     `v12/fixtures/bench/sum_u32_small/main.able`
   - Snapshot:
     `v12/docs/perf-baselines/2026-03-19-mono-array-u32-sum-small-compiled.md`
   - Result: mono on `1.0933s` / `185.33` GC vs mono off `1.6800s` /
     `21.33` GC over 3 compiled runs.
   - Interpretation: the remaining primitive scalar family now shows a real
     compiled wall-clock win on a typed unsigned workload, even though GC
     count is not the improving metric on that benchmark.

## Next Category

The next monomorphization category is now beyond primitive scalars:

1. Widen native storage coverage to broader non-scalar/container carrier
   families that still fall back to generic `*Array` / `runtime.Value`.
2. Add focused reduced benchmarks for each new family when the existing set
   does not already exercise it directly.
3. Re-measure after each wider carrier reduction slice lands.
