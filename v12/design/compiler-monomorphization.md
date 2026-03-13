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
| `Array i32` | native Go `[]int32` or compiler-owned wrapper around `[]int32` |
| `Array String` | native Go `[]string` or compiler-owned wrapper around `[]string` |
| `Array Point` | native Go `[]*Point` or compiler-owned wrapper around `[]*Point` |
| `Array (A | B)` | compiler-owned wrapper over native union carrier elements |
| `Point` | `*Point` |
| `A | B` | generated Go interface plus native variants |
| `?Point` | nil-capable `*Point` |

`any` is acceptable only as a temporary implementation escape hatch or at an
explicit dynamic boundary. It is not the target representation for compiled
union values or generic containers.

## Arrays

### Required direction

- Static compiled arrays must use compiler-native Go storage.
- Array literals should become native slice literals or wrapper construction.
- `len`, `push`, `pop`, `get`, `set`, indexing, iteration, and cloning should
  operate directly on that native storage.
- The compiler must not treat kernel `Array { length, capacity, storage_handle }`
  as the internal compiled representation.

### Boundary rule

If a compiled array crosses into explicit dynamic behavior, the compiler may
generate adapters:

- native compiled array -> runtime/kernel boundary value
- boundary value -> native compiled array

That conversion is boundary logic only. It must not leak back into the default
static lowering path.

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
- compiler-only rewrites of the kernel `Array` struct shape that hide
  `storage_handle` behind `Elements []runtime.Value`;
- union lowering that stops at `any` instead of generating native carrier
  interfaces;
- struct-local boxing into `runtime.Value` to paper over missing static ABI
  work.

## Execution Order

1. Define the compiler-native array ABI for static code.
2. Specialize array operations over native storage.
3. Remove struct boxing that was added to preserve dispatch identity.
4. Introduce generated Go interfaces for union carriers.
5. Narrow `any` usage until it exists only at explicit dynamic boundaries or
   clearly logged residual cases.
