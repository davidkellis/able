# Able v11 Tutorial Programs

Each file in this folder demonstrates one major Able language concept with small, runnable samples. Use the TypeScript interpreter helper from the repo root:

```
./v11/ablets examples/tutorial/01_basics_and_bindings.able
./v11/ablets check examples/tutorial/01_basics_and_bindings.able
```

Examples (run in order):
- `01_basics_and_bindings.able` — declarations vs assignment, type annotations, block expressions, interpolation, safe navigation, and pipe-forward basics.
- `02a_operators_and_builtin_types.able` — scalar literals, arithmetic (`/ // % /%`), comparisons, bitwise/shift ops, and truthiness for logical operators.
- `02_patterns_and_destructuring.able` — struct/array patterns, named and positional structs, nested destructuring, and wildcards.
- `03_structs_unions_and_match.able` — struct construction, unions, nullable shorthand, and exhaustive `match`.
- `04_functions_and_lambdas.able` — named functions, anonymous functions, lambdas, trailing lambdas, partial application placeholders, and pipe-driven callables.
- `05_methods_and_implicit_self.able` — `methods` blocks, implicit `self` (`fn #method`), `#field` shorthand, and UFCS-style calls.
- `06_interfaces_and_impls.able` — defining interfaces, implementing them for concrete types, static vs instance methods, and interface-typed values.
- `07_arrays_ranges_and_iteration.able` — array literals, mutation, for/while loops, ranges (`..` / `...`), and generator literals (`Iterator { ... }`).
- `08_control_flow_and_breakpoint.able` — `if/or` as expressions, `loop` results, `continue`/`break`, and labeled `breakpoint` early-exit.
- `09_options_results_and_else.able` — option/result shorthands (`?T` / `!T`), propagation `!`, and `else {}` handling with and without error binding.
- `10_exceptions_and_rescue.able` — `raise`, `rescue`, `ensure`, and `rethrow` with a custom `Error` implementation.
- `11_concurrency_proc_and_spawn.able` — `proc` vs `spawn`, cooperative helpers (`proc_yield`, `proc_flush`, `proc_pending_tasks`), and memoization of futures.
- `12_channels_mutex_and_await.able` — channel creation, buffered/unbuffered sends, `try_*` helpers, `await` with defaults, and a simple mutex guard.
- `13_packages_and_imports_main.able` — package naming, visibility with `private`, re-exports, and static imports across sibling modules (entrypoint uses the helpers).
- `14_host_interop.able` — `prelude <target>` setup, `extern` host bodies for Go/TypeScript, and a pure Able fallback.
