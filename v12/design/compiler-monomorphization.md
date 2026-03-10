# Compiler Monomorphization Design

## Goal

Eliminate `runtime.Value` boxing from compiled code. Native Go types everywhere.
`runtime.Value` used ONLY at the compiled↔interpreted boundary.

## Type Mapping (TypeMapper)

### Current → Target

| Able Type | Current Go Type | Target Go Type |
|-----------|----------------|----------------|
| `i32` | `int32` | `int32` (unchanged) |
| `String` | `string` | `string` (unchanged) |
| `bool` | `bool` | `bool` (unchanged) |
| `Point` (struct) | `*Point` | `*Point` (unchanged) |
| `Array<i32>` | `*Array` (handle-based) | `[]int32` |
| `Array<String>` | `*Array` (handle-based) | `[]string` |
| `Array<Point>` | `*Array` (handle-based) | `[]*Point` |
| `Array<i32 \| String>` | `*Array` (handle-based) | `[]any` |
| `i32 \| String` | `runtime.Value` | `any` |
| `?i32` | `runtime.Value` | `any` (nil = absent) |
| `?Point` | `runtime.Value` | `*Point` (nil = absent) |
| `T -> U` (fn type) | `runtime.Value` | `func(T) U` or `any` |
| `T` (unbound generic) | `runtime.Value` | `any` |

### Key Principle: `any` replaces `runtime.Value`

Go's `any` (empty interface) provides the same polymorphism as `runtime.Value`
but accepts native Go types directly. A `[]any` can hold `int32`, `string`,
`*Point` values without boxing into `runtime.IntValue`, `runtime.StringValue`, etc.

Pattern matching on `any` uses native Go type switches:
```go
switch v := x.(type) {
case int32:  // Able i32
case string: // Able String
case *Point: // Able Point
}
```

## Array Representation

### Current: Handle-based indirection
```
Array struct { Length int32, Capacity int32, Storage_handle runtime.Value }
  → handle (int64) → global map → []runtime.Value or []int32 (mono)
```

### Target: Native Go slices
```
Array<i32>    → []int32
Array<String> → []string
Array<Point>  → []*Point
Array<T>      → []any   (when T is unbound or union)
```

Arrays are mutable and passed by reference. Go slices are reference types
(backed by an array pointer + length + capacity), but `append` may reallocate.
Two options:

**Option A: Bare slices.** `Array<i32>` = `[]int32`. Methods take `*[]int32`
receiver to support push/append. Caller must pass address.

**Option B: Slice wrapper struct.** `Array<i32>` = `*ArrayI32` where
`type ArrayI32 struct { Data []int32 }`. Pointer semantics natural.

Option A is simpler and more Go-idiomatic. The compiler already handles pointer
receivers for structs. Let's go with bare slices as the representation.

### Array Operations

| Operation | Current | Target |
|-----------|---------|--------|
| `[1,2,3]` | alloc handle + 3x map write | `[]int32{1, 2, 3}` |
| `arr.len()` | `ArrayStoreSize(handle)` | `int32(len(arr))` |
| `arr[i]` | `ArrayStoreRead(handle, i)` | `arr[i]` |
| `arr.push(v)` | `ArrayStoreWrite(handle, len, v)` + bookkeeping | `*arr = append(*arr, v)` |
| `arr.pop()` | `ArrayStoreRead` + `ArrayStoreSetLength` | `v := arr[len-1]; *arr = (*arr)[:len-1]` |
| `arr.get(i)` | bounds check + `ArrayStoreRead` | bounds check + `arr[i]` |
| `arr.set(i,v)` | bounds check + `ArrayStoreWrite` | bounds check + `arr[i] = v` |
| `arr.clone()` | `ArrayStoreClone(handle)` | `slices.Clone(arr)` (or `copy`) |
| `arr.filter(f)` | loop + dynamic dispatch | loop + native call |
| `arr.map(f)` | loop + dynamic dispatch | loop + native call |

## Union Types → `any`

Union types like `i32 | String` become `any` in Go. Values inside are native
Go types (`int32`, `string`), not runtime.Value wrappers.

Match patterns compile to type switches:
```go
// x: i32 | String
switch v := x.(type) {
case int32:
    // use v as int32
case string:
    // use v as string
}
```

Nullable `?T` also becomes `any` (or `*T` when T is a value type and nil can
be distinguished). For `?Point`, use `*Point` (nil = absent). For `?i32`,
must use `any` since `int32` has no nil.

## Boundary Conversion

At the compiled↔interpreted boundary, conversion functions translate between
native Go types and `runtime.Value`:

```go
func nativeToRuntimeValue(v any) runtime.Value {
    switch val := v.(type) {
    case int32:  return bridge.ToInt(int64(val), runtime.IntegerI32)
    case string: return bridge.ToString(val)
    case bool:   return bridge.ToBool(val)
    case *Point: return __able_struct_Point_to(rt, val)
    case []int32: return arrayI32ToRuntime(val)
    // etc.
    }
}

func runtimeValueToNative(v runtime.Value, targetType string) any {
    // inverse of above
}
```

## Implementation Order

1. **TypeMapper: union/nullable → `any`** — foundational type change
2. **TypeMapper: Array<T> → `[]ElemType`** — monomorphized arrays
3. **typeCategory: handle `any` and slice types** — downstream dispatch
4. **Runtime helpers: accept `any`** — truthy, try_cast, binary_op
5. **Array intrinsics: target native slices** — len, push, get, set
6. **Array literal compilation** — produce slice literals
7. **Match compilation: native type switches** — pattern matching on `any`
8. **Method compilation** — monomorphized array methods
9. **Boundary converters** — from/to at interop edges
10. **Struct field/param handling** — native types in structs and functions

Each step is testable independently. Steps 1-3 are the foundation.
Steps 4-7 make basic operations work. Steps 8-10 complete the picture.
