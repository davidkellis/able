# Compiler: No Panic-Based Flow Control

## Principle

The compiled Go output must not use `panic`/`recover` for normal control flow (break, continue, return, loop exit). Go panics are reserved exclusively for exceptions — Able's error propagation via `bridge.Raise`, `rescue`, and `or_else`. Using panics for structural flow control is slow, obscures intent, and requires IIFE wrappers with defer/recover stacks.

## Goals

1. **Eliminate all IIFEs** — Replace `func() T { stmts; return expr }()` with inline setup lines hoisted to statement context. Where a true function boundary is needed, use a named helper function instead.
2. **Eliminate panic-based flow control** for break, continue, and return. Use Go's native control flow (labeled break/continue, regular return, labeled switch for breakpoints).
3. **Keep panic for exceptions only** — `bridge.Raise`/`rescue`/`or_else` model truly exceptional error propagation. Go's `panic`/`recover` is appropriate here. Converting exceptions to return-value signaling would require signal checks after every function call — significant complexity for something that rarely fires.

## Architecture: Lines Hoisting

The core technique is `compileExprLines`: every expression compiler returns `([]string, string, string, bool)` — setup lines separated from the final expression value. Lines are hoisted to the nearest statement context. This eliminates the need for IIFEs because statements execute in the function body (where Go `return` works normally) rather than inside an anonymous closure.

Example — an Able block expression used as a value:

```able
x := {
    if something { return 42 }
    other_value
}
```

Before (IIFE + panic):
```go
x := func() runtime.Value {
    if something { panic(__able_return{value: wrap(42)}) }
    return other_value
}()
// enclosing function has defer/recover to catch __able_return
```

After (lines hoisted):
```go
if something { return wrap(42) }  // regular Go return
x := other_value
```

The setup lines (including the `if/return`) hoist to the enclosing function body. The expression (`other_value`) is used directly. No IIFE, no panic.

For deeply nested expressions like `f(g(complex_expr))`, each nesting level hoists its sub-expression's lines:
```go
// lines from complex_expr
__tmp := complex_expr_value
f(g(__tmp))
```

## Implementation Plan

### Phase 1: Complete `compileExprLines` Conversion — DONE

All expression compilers that can produce setup lines now return `([]string, string, string, bool)`. `wrapLinesAsExpression` is down to 2 call sites (canonical `compileExpr` fallback + `compileFunctionCall` array intrinsics caller).

Dead duplicate cases in `compileExprExpected` were removed — lines-based types are now exclusively routed through `compileExprLines`.

### Phase 2: Remove `iifeDepth` and `__able_return` Panic — DONE

- `compileBlockExpression` no longer wraps in IIFEs — returns lines + tail expression directly
- `compileReturnStatement` always emits regular Go `return` (no conditional panic)
- Removed `iifeDepth` field from `compileContext`
- Removed `needsReturnRecover` mechanism
- Removed `__able_return` panic type and its defer/recover handler from function rendering
- Removed `iifeDepth++` from match clause compilation

### Phase 3: Convert Breakpoint Expressions to Labeled Switch — DONE

Breakpoint expressions (`#label { ... break #label value ... }`) now use Go's labeled `switch` for the labeled break:

```go
var __result runtime.Value = runtime.NilValue{}
__label: switch { default:
    // body...
    if something {
        __result = value
        break __label
    }
    __result = other_value
}
```

`break #label value` compiles to `__result = value; break __label`. The Go label and result temp are propagated through `breakpointGoLabels` and `breakpointResultTemps` maps on `compileContext`.

**Dead code removed:** All panic-based flow control signal types and functions:
- `__able_break`, `__able_break_value` — loop break (replaced by labeled `break`)
- `__able_continue_signal`, `__able_continue` — loop continue (replaced by labeled `continue`)
- `__able_break_label_signal`, `__able_break_label` — breakpoint break (replaced by labeled switch `break`)
- `__able_continue_label_signal`, `__able_continue_label` — never used
- Lambda recover handler simplified — no longer re-panics flow control signals

## Completed Work

**Loop break/continue (done):** Loops use Go's native labeled break/continue:

```go
__able_tmp_1: for {
    // body
    if condition { break __able_tmp_1 }
    if other { continue __able_tmp_1 }
}
```

Break-with-value uses a temp variable:

```go
var __able_tmp_2 runtime.Value = runtime.NilValue{}
__able_tmp_1: for {
    // body
    __able_tmp_2 = someValue
    break __able_tmp_1
}
```

**Dead code removed:** `blockHasBreakContinueRescue` and related analysis functions.

## Exceptions: Permanent Panic

`bridge.Raise`, `rescue`, and `or_else` model Able's error propagation. These will continue to use Go's `panic`/`recover`:

- `bridge.Raise(err)` panics with the error value
- `rescue` expressions use `defer/recover` to catch raised errors
- `or_else` expressions use `defer/recover` to provide fallback values

This is appropriate because: (a) exceptions are truly exceptional — they fire rarely, (b) Go's panic is zero-cost on the happy path, (c) converting to return-value signaling would require `if sig != nil` checks after every function call, adding overhead to every call site for a condition that rarely triggers.

## Lambda Break/Continue

The compiler does not support `break`/`continue` inside lambdas — `loopDepth` resets to 0 when compiling a lambda body, so break/continue statements are rejected. If this is ever needed, Go's labeled break/continue cannot cross function boundaries, so a panic-based signal would be required as a narrow exception.

## Guidelines for New Code

- **Never** introduce new panic-based flow control for break, continue, return, or loop exit.
- **Always** return `([]string, string, string, bool)` from new expression compilers. Do not create new `wrapLinesAsExpression` call sites.
- **Prefer** Go's native control flow: labeled `break`/`continue` for loops, regular `return`, labeled `switch` for non-loop breaks.
- **Use** temp variables for break-with-value and expression results.
- **Watch for `:=` shadowing** when restructuring `return` to `if/else`: `result, err := ...` inside an `if` block creates a new scoped variable shadowing the outer one.
