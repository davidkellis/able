# Compiler + Interpreter Vision (Historical Direction and Current VM Record)

This is an architecture-history document, not a language or lowering
specification. It preserves earlier typed-core-IR and shared-runtime ideas for
context while recording the execution contracts that actually exist in the Go
reference implementation.

## Current authority and implementation boundary

- [`spec/full_spec_v12.md`](../../spec/full_spec_v12.md) defines Able
  semantics. The AST and the Go tree-walker and bytecode VM must remain in
  parity with it.
- The active bytecode implementation is the Go source under
  `interpreter-go/pkg/interpreter/bytecode_*.go`; its focused tests are the
  executable record for instruction, frame, cache, and return behavior. This
  document deliberately does not define a serialized bytecode ABI.
- The active compiler is the direct AST-to-Go generator in
  `interpreter-go/pkg/compiler`. `compiler-go-lowering-spec.md`,
  `compiler-go-lowering-plan.md`,
  `compiler-native-lowering-guardrails.md`, strict
  compiler tests, and `docs/compiler-full-matrix.md` describe its current
  lowering and verification boundary. `compiler-aot.md` is a historical
  proposal, not additional authority.
- `PLAN.md` owns performance selection. A runtime or compiler optimization
  needs a concrete, non-nominal shared leaf that recurs across unlike verified
  applications; a faster isolated benchmark is not enough.

### Current Go execution record

The Go bytecode VM uses an in-memory instruction stream and the same runtime
value model as the tree-walker. It retains an operand stack, but it is not a
stack-only VM: eligible functions use analysed slot-indexed local frames, and
proved eligible `i32` functions may use a raw register frame. Inline calls use
VM call frames that preserve caller program, instruction position, stack
boundary, local-frame state, lookup state, coercion metadata, and raw-frame
state. Calls outside that inline path retain the ordinary semantic dispatch
bridge. The exact eligibility, invalidation, and fallback rules are source and
test contracts, not promises made by this design record.

The current compiler typechecks, discovers the static module graph and dynamic
boundaries, then emits Go directly from the AST. It has no standalone typed
core IR and no general shared VM/compiler runtime ABI. Strict no-fallback
policies protect statically representable paths; dynamic features cross the
existing explicit boundary machinery.

No architectural migration, new backend, serialized bytecode format, or
performance exception is authorized by the historical sections below.

## Historical goals
- Keep full Able expressiveness (dynamic interface values, metaprogramming, concurrency).
- Compiled output should be close to host performance (Go first, then other targets).
- Interpreted execution should move toward the performance of modern interpreted languages.
- A single semantic core must drive both interpreter and compiler backends.

## Historical recommended execution model
- Build a typed core IR from the AST.
- Two backends:
  1) Bytecode VM interpreter (fast, portable).
  2) Host-language codegen (Go first), emitting calls into a shared runtime library.
- Compiled artifacts bundle the runtime and, where needed, explicit dynamic
  boundary machinery.
- Dynamic features route to runtime or VM entry points only when the source
  program crosses an explicit dynamic boundary.

## Historical interface-dispatch direction
- Dynamic/runtime interface values may carry dictionary-style metadata (see
  `v12/design/interface-dispatch-dictionaries.md`).
- Static compiled code should use direct calls or generated native interface
  carriers/helpers, not interpreter-style dictionary dispatch, when the target
  is statically representable.
- Dictionaries can remain part of the interpreter/dynamic runtime model and may
  still be useful at explicit dynamic boundaries.

## Historical interpreter-modernization direction
- Move from tree-walking to bytecode or SSA-based VM:
  - Lower AST -> typed IR -> bytecode.
  - Use inline caches for member/method lookups and dictionary dispatch.
  - Keep value representations shared with the runtime to reduce conversions.
- Start with a stack-based bytecode to minimize compiler complexity, then consider register-based once stabilized.
- Maintain determinism for concurrency semantics (`spawn` / `Future` handles) while improving throughput.

## Historical stack-oriented bytecode snapshot

This was an early stack-oriented description of the VM. It predates slot
layouts, inline VM call frames, raw scalar carriers, and the register-frame
work recorded above. It remains useful as design history only; do not use its
instruction list or its call-frame statements as a current opcode contract.

### Program model
- **Program**: linear instruction stream. Each program runs with a single operand stack.
- **Environment**: lexical environments are explicit; `enter_scope` pushes a new env and
  `exit_scope` pops.
- **Values**: VM uses the same `runtime.Value` types as the tree-walker.
- **Control flow**: absolute jump targets by instruction index.

### Instruction encoding (in-memory)
The historical VM used an in-memory struct representation rather than a
serialized bytecode format. Each instruction carried inline operands:
- `op`: opcode enum.
- `name`: identifier payload for `load_name`, `assign_name`, `call_name`, `break_label`.
- `operator`: operator token for `binary`, `unary`, compound assignments, member/index sets.
- `value`: literal `runtime.Value` for `const`.
- `target`: absolute instruction index for jumps.
- `argCount`: arity for `call`, `call_name`, `array_literal`, `string_interp`, `exit_scope`.
- `loopBreak`/`loopContinue`: jump targets recorded by `loop_enter`.
- `node`: AST node used for diagnostics and fallback evaluation (match/rescue/ensure/etc.).
- `program`: nested bytecode program for placeholder lambdas and iterator literals.
- `safe`/`preferMethods`: member-access flags (safe navigation + method preference).

### Execution + call frames
- The VM itself is not re-entrant; nested execution (e.g., placeholder lambdas) spins up
  a new VM instance.
- This historical text's claim that calls have no explicit VM frame has been
  superseded by the current inline-call frame implementation.
- `call` expects the callee below its arguments on the stack; arguments are pushed
  left-to-right, then popped in reverse to preserve order before invoking the callee.
- `call_name` pops only arguments and resolves the callee by name (including `Type.method`).
- Runtime diagnostics attach to AST nodes captured on each instruction where needed.

### Concurrency + async
- `spawn` lowers to a VM opcode that schedules an async task via the interpreter executor.
- Async bytecode tasks preserve VM state across `future_yield`; the VM resumes at the
  instruction after the yield call.
- Await/select execution cooperates with the serial executor by returning an internal
  yield error (`errSerialYield`) until resumed.

### Instruction set (stack effects)
Notation: `S` is the operand stack. Effects are shown as `... -> ...`.

#### Constants and locals
- `const <Value>`: `S -> S, value`
- `load_name <id>`: `S -> S, env[id]`
- `declare_name <id>`: `S, value -> S, value` (defines in current scope)
- `assign_name <id>`: `S, value -> S, value` (updates existing binding)
- `assign_name_compound <id> <op>`: `S, value -> S, value` (reads env[id], applies operator)
- `assign_pattern <op>`: `S, value -> S, value` (binds destructuring/typed patterns)

#### Stack helpers
- `dup`: `S, x -> S, x, x`
- `pop`: `S, x -> S`

#### Expressions + operators
- `binary <op>`: `S, left, right -> S, result`
  - `+` handles string concatenation; all other ops delegate to runtime helpers.
- `unary <op>`: `S, value -> S, result`
- `range`: `S, start, end -> S, range`
- `cast`: `S, value -> S, coerced`
- `string_interp <n>`: `S, parts... -> S, string`
- `propagate`: `S, value -> S, value` (raises if value is `Error`/`!T`)
- `or_else`: opcode that evaluates the main expression via fallback, binds errors in a
  new scope, and runs the handler inline.

#### Calls + member access
- `call <n>`: `S, callee, args... -> S, result`
- `call_name <id> <n>`: `S, args... -> S, result` (resolves by name, supports `Type.method`)
- `member_access`: `S, receiver -> S, value` (supports safe access + method preference)
- `implicit_member`: `S -> S, value` (resolves `#member` inside current function)
- `member_set <op>`: `S, value, receiver -> S, result`
- `implicit_member_set <op>`: `S, value -> S, result`

#### Functions + lambdas
- `make_function <lambda>`: `S -> S, function`
- `define_function <fn>`: `S -> S, nil` (defines in env)
- Placeholder lambdas lower to a closure that evaluates the captured expression with
  placeholder bindings; they can execute bytecode when available.

#### Definitions
- `define_struct <struct>`: `S -> S, nil`
- `define_union <union>`: `S -> S, nil`
- `define_type_alias <alias>`: `S -> S, nil`
- `define_methods <methods>`: `S -> S, nil`
- `define_interface <iface>`: `S -> S, nil`
- `define_implementation <impl>`: `S -> S, nil`
- `define_extern <extern>`: `S -> S, nil`

#### Imports
- `import <stmt>`: `S -> S, nil`
- `dynimport <stmt>`: `S -> S, nil`

#### Data literals
- `struct_literal`: `S -> S, struct`
- `map_literal`: `S -> S, map` (handles spread via runtime helpers)
- `array_literal <n>`: `S, elements... -> S, array`
- `iterator_literal`: `S -> S, iterator` (generator body runs in bytecode when supported; falls back to tree-walker)

#### Indexing
- `index_get`: `S, object, index -> S, value`
- `index_set <op>`: `S, value, object, index -> S, result`

#### Control flow + loops
- `jump <target>`: `S -> S`
- `jump_if_false <target>`: `S, cond -> S` (jumps on falsy)
- `jump_if_nil <target>`: `S, value -> S` (jumps on nil)
- `break_label <label>`: `S, value -> S` (raises break signal for breakpoint labels)
- `enter_scope`: pushes environment
- `exit_scope [n]`: pops `n` environments (defaults to 1)
- `loop_enter <break, continue>`: pushes loop frame for delegated break/continue handling
- `loop_exit`: pops loop frame

#### Iteration
- `iter_init`: `S, iterable -> S` (push iterator handle)
- `iter_next`: `S -> S, value, done`
- `iter_close`: `S -> S` (closes top iterator)
- `bind_pattern`: `S, value -> S` (binds loop pattern in current env)
- `yield`: `S, value -> S, nil` (emit from generator)

#### Errors + rescue
- `raise`: delegates to tree-walker; returns `raiseSignal`
- `rethrow`: delegates to tree-walker; returns `raiseSignal`
- `match`: `S, subject -> S, result` (clause bodies and guards use fallback eval)
- `ensure`: opcode that evaluates the try expression via fallback, runs the ensure
  block inline, then rethrows any captured error or yields the try result.
- `rescue`: opcode that evaluates the monitored expression via fallback, matches
  clauses inline, and returns the handled value or rethrows.

#### Spawn / await
- `spawn`: schedules async task; returns `Future`
- `await`: `S, iterable -> S, result` (collects await arms; cooperates with the serial executor)

#### Eval fallbacks
- `eval_expression`: evaluate expression in tree-walker (used for unsupported nodes)
- `eval_statement`: evaluate statement in tree-walker (definitions/imports, etc.)

#### Return
- `return`: `S, value -> (exit)` (returns from program)

## Historical compiler direction (Go first)
- Emit Go code from typed IR.
- Generate explicit runtime calls for:
  - interface coercion + dictionary dispatch
  - dynamic metaprogramming (expr-eval)
  - concurrency scheduling (`spawn`)
- Keep the runtime ABI stable so the interpreter and compiler stay in lockstep.

## Historical typed-core-IR proposal
The core IR is a typed, SSA-like, control-flow graph used by both the VM and
codegen. It is intentionally small and explicit about runtime calls so we can
preserve v12 semantics while keeping backends simple.

### Concrete IR node set (draft)
This is a **working** node list for the first IR cut. It is intentionally
conservative and mirrors current runtime semantics (tree-walker + bytecode).

#### Units
- **Program**: packages + entrypoints.
- **Package**: symbol table, type table, constants, functions.
- **Function**: generic params + typed params + return type + CFG.
- **Block**: ordered instruction list + single terminator. Blocks can take
  **parameters** (block args) to model SSA phi merges.

#### Values
- **Temps**: immutable SSA values produced by instructions.
- **Locals**: mutable slots for `:=` / `=` semantics; explicit load/store.
- **Type refs**: runtime type tokens (`T_type`) for diagnostics + reflection.

#### Core instructions (draft)
**Binding + locals**
- `local_alloc name, type -> local`
- `local_set local, value`
- `local_get local -> value`
- `define_global name, value`

**Constants**
- `const <T> value`

**Aggregates**
- `struct_new Type { field: value, ... } -> value`
- `struct_get value.field -> value`
- `struct_set value.field = new_value -> value`
- `array_new [values...] -> value`
- `array_get value[index] -> value`
- `array_set value[index] = new_value -> value`
- `map_new [(key, value)...] -> value`
- `map_get value[key] -> value`
- `map_set value[key] = new_value -> value`
- `map_delete value[key] -> value`

**Union / option / result**
- `union_new Variant(value) -> value`
- `union_tag value -> tag`
- `union_payload value -> payload`
- `option_none -> value`
- `result_ok value -> value`
- `result_err value -> value`

**Operators + casts**
- `binary op, left, right -> value`
- `unary op, value -> value`
- `cast value, target_type -> value`
- `range start, end, inclusive -> value`
- `string_interp [parts...] -> value`
- `propagate value -> value` (raises on `Error` or `!T`)

**Calls + dispatch**
- `call fn, args... -> value`
- `call_method receiver, method_name, args... -> value`
- `call_static Type, method_name, args... -> value`
- `call_interface iface_value, method_name, args... -> value`
- `call_apply value, args... -> value` (for `Apply` interface)
- `coerce_interface value, iface_type -> iface_value` (build dictionary)

**Control flow**
- `jump block, args...`
- `branch cond, then_block, else_block`
- `match value -> [case -> block], default -> block`
- `return value`
- `raise value`

**Loops**
- Structured lowering into blocks with back-edges.
- `break`/`continue` lower into `jump` with value-carried block args.

**Concurrency**
- `spawn fn, args... -> Future<T>`
- `await value -> value`
- `future_status handle -> value`
- `future_value handle -> value`
- `future_cancel handle -> value`
- `future_yield -> value`
- `future_flush -> value`

**Dynamic + extern**
- `dyn_import package -> value`
- `dyn_eval expr_string -> value`
- `extern_call name, args... -> value`

#### IR invariants (draft)
- All blocks end with a single terminator (`return`, `raise`, `jump`, `branch`).
- All values are typed; locals have a declared type.
- `propagate` is explicit at each `!` site; error handling (`rescue`/`ensure`)
  is explicit in CFG.
- Method sugar, UFCS, and implicit `self` are lowered before IR emission.

### IR units
- **Program**: list of packages + entrypoints.
- **Package**: symbol table, type table, functions, constants.
- **Function**: generic params + typed params + return type + CFG.
- **Block**: ordered instruction list + single terminator.

### Type model
- **Primitive**: `bool`, `char`, `i32`, `i64`, `u64`, `f64`, `String`, `nil`.
- **Struct**: nominal struct with field list + field types.
- **Union**: tagged union (sum) with variant types.
- **Array / Map**: runtime-provided containers with element/key/value types.
- **Function**: `(T1, T2, ...) -> R`.
- **Interface**: interface name + type arguments (fully bound for existentials).
- **Future**: `Future T` handle type (per spec §12).
- **Result/Option**: `!T` and `?T` are unions in the IR.

### Values and variables
- **SSA temps**: immutable values produced by instructions.
- **Locals**: explicit mutable slots for `:=`/`=` binding semantics.
- **Type refs**: runtime type tokens when needed (e.g., `C.default()`).

### Core instruction set (representative)
- **Constants**: `const <T> value`.
- **Aggregate**:
  - `struct_new Type { field: value, ... }`
  - `struct_get value.field`
  - `struct_set value.field = new_value`
  - `array_new`, `array_get`, `array_set`, `array_push`
  - `map_new`, `map_get`, `map_set`, `map_delete`
- **Control flow**:
  - `branch cond -> block_true, block_false`
  - `jump block`
  - `match value -> [case/tag -> block]`
  - `return value`
  - `raise value`
- **Calls**:
  - `call fn(args...)`
  - `call_method receiver method(args...)`
  - `call_static Type.method(args...)`
  - `call_interface iface_value method(args...)` (dictionary dispatch)
- **Operators**:
  - arithmetic/compare/logical ops resolve to explicit calls or runtime helpers
  - `coerce_interface value -> interface` (build dictionary)
- **Concurrency**:
  - `spawn fn(args...) -> Future<T>`
  - `await value` (calls Awaitable protocol)
- **Dynamic**:
  - `dyn_eval expr_string` (or host-specific hook)
  - `dyn_import package`

### Notes
- Method sugar, UFCS, and implicit `self` are lowered before IR generation.
- Interface dictionary entries are bound during `coerce_interface` so runtime
  calls are direct and deterministic.
- All effectful operations (I/O, concurrency, metaprogramming) remain explicit
  runtime calls for parity across backends.

## Historical shared-runtime-ABI proposal
The runtime ABI is the shared contract implemented by the Go runtime and the VM
host layer. ABI calls are used by compiled code and the VM; the tree-walker can
reuse the same surface for consistency.

### Core services
- **Type metadata**: construct and compare runtime type tokens; used for
  generics (`T_type`) and diagnostics.
- **Errors**:
  - Construct `Error` values and standard error structs.
  - Propagate `!T` and `?T` via explicit helpers.
- **Strings**: create from bytes, slice, concat, interpolation helpers.
- **Collections**: allocate arrays/maps and perform primitive ops.

### Interface dispatch
- **Dictionary build**: `iface_build(value, iface_name, type_args) -> dict`
- **Dictionary lookup**: `iface_call(dict, method_name, args...)`
- **Default methods**: either stored in dict or resolved via interface metadata.

### Concurrency
- **Spawn**: `future_spawn(fn, args...) -> Future<T>`
- **Future handle**:
  - `future_status(handle) -> FutureStatus`
  - `future_value(handle) -> !T`
  - `future_cancel(handle) -> void`
  - `future_yield/flush/pending_tasks` helpers per spec
- **Await**: `await_poll(value)` drives `Awaitable` protocol in cooperative runtimes.

### Dynamic + externs
- **Dynimport**: load package/module objects dynamically.
- **Externs**: host function call boundary with value marshaling and error
  propagation.

### Diagnostics
- Shared formatting helpers for runtime/typechecker diagnostics so backends emit
  identical messages.

## Historical immediate next steps
- Define a typed core IR and document it.
- Prototype a minimal bytecode VM (values, calls, control flow).
- Add conformance tests that run on both tree-walker and VM backends.
- Sketch the Go codegen layer for a small subset of IR instructions.
