# Able v12 Typechecker (Go)

Date: 2025-10-21  
Maintainers: Able Agents

The Go typechecker is now feature-complete for the v12 surface. This document
explains the architecture, the checking pipeline, and—most importantly—how the
constraint solver enforces interface/impl obligations. The goal is that any new
contributor can read this file, understand how the checker works end-to-end, and
confidently extend it.

---

## High-level Goals

1. **Spec fidelity** – evaluate Able v12 modules exactly as described in
   `spec/full_spec_v12.md`, including generics, where clauses, interfaces, and
   async helpers.
2. **Inference-first** – minimise explicit annotations. The checker infers types
   for expressions, patterns, and statements, recording results in side tables
   instead of mutating the AST.
3. **Reusable metadata** – the checker exposes an `InferenceMap` and the set of
   resolved obligations so the interpreter, future compiler, and tooling can
   reuse the same information.
4. **Actionable diagnostics** – report rich, contextual errors (e.g. “via method
   ‘format’” or “via method set”) while continuing checking when possible.

---

## Package Layout

`pkg/typechecker` is split into focused files that mirror the interpreter:

| File | Responsibility |
| ---- | -------------- |
| `checker.go` | Public entry point (`CheckModule`), orchestration, diagnostic aggregation |
| `decls.go` | Module-level declaration collection (structs, unions, interfaces, functions, impls, method sets) |
| `env.go` | Lexical symbol tables used during checking |
| `types.go` | Type definitions (`PrimitiveType`, `StructType`, `InterfaceType`, `FunctionType`, etc.) plus helper structs (`GenericParamSpec`, `WhereConstraintSpec`, `ConstraintObligation`) |
| `inference.go` | `InferenceMap` (AST node → resolved type) |
| `literals.go`, `member_access.go`, `patterns.go`, `range_expression.go`, `array_literal.go`, etc. | Expression/statement/pattern checkers grouped by feature |
| `concurrency.go` | Async-specific helpers (`spawn`, `future_*` builtins) |
| `constraint_solver.go` | Obligation resolution (see “Constraint Solving” below) |
| `type_utils.go` | Substitution helpers used across checking and solving |

Tests live alongside the code (`checker_*_test.go`) and mirror the feature
breakdown.

---

## Checking Pipeline

`CheckModule(*ast.Module)` runs in four phases:

1. **Initialisation**
   - Reset the `InferenceMap`, diagnostics slice, and internal state.
   - Seed the global environment with builtins (e.g. `print`, `future_*`).

2. **Declaration Collection (`decls.go`)**
   - Walk top-level statements and register structs, unions, interfaces,
     function signatures, impl blocks, and method sets in the global
     environment.
   - For each declaration we capture:
     - Generic parameters (`[]GenericParamSpec`)
     - Where clauses (`[]WhereConstraintSpec`)
     - Function signatures (`FunctionType` values with param/return types)
     - Impl/method-set metadata (`ImplementationSpec`, `MethodSetSpec`)
   - Duplicate declarations produce diagnostics but do not stop the pass.
   - While collecting we also build an initial set of `ConstraintObligation`
     records representing generic/where requirements tied to declarations.

3. **Body Checking**
   - Iterate module body statements again and evaluate them with a fresh lexical
     environment.
   - Statements delegate into specialised helpers (e.g. `checkBlock`,
     `checkWhile`, `checkFunctionDefinition`), each of which:
       - Resolves types recursively.
       - Updates the `InferenceMap` for every visited node.
       - Threads diagnostics upward when type requirements are violated.
   - Expression helpers reside in feature-specific files; they combine direct
     inference with data pulled from `InferenceMap` and the global env.
   - Pattern checking populates bindings inside scope environments so subsequent
     expressions observe inferred types.

4. **Constraint Solving**
   - After bodies are checked, `resolveObligations()` verifies that all gathered
     `ConstraintObligation` records are satisfied (details in the next section).
   - Diagnostics emitted during solving are appended to the main slice and
     include contextual hints (e.g. `via method 'format'`).

If any unexpected errors occur (e.g. nil AST), the checker returns an error in
addition to diagnostics. Otherwise callers receive the diagnostics slice and the
checker’s internal state (environments, inference map) remains available for
inspection.

---

## Key Data Structures

### Types (`types.go`)

- **PrimitiveType / IntegerType / FloatType** – concrete scalar types.
- **StructType / StructInstanceType** – named structs (optionally generic) and
  instances with field/type data.
- **InterfaceType** – interface metadata: name, generic parameters, where
  clauses, and a map of method signatures.
- **FunctionType** – parameter types, return type, type parameters, where
  clauses, and any constraints collected during checking.
- **AppliedType** – generic application (e.g. `Array<T>`); used for both type
  expressions and constraint arguments.
- **FutureType** – async handle representation.
- **UnknownType** – placeholder when inference cannot resolve a type.

### Environments (`env.go`)

`Environment` provides a lexical scope with optional parent reference. Entries
are `Type`, `FunctionInfo`, or other checker-specific metadata. The checker
pushes/pops scopes around blocks, functions, lambdas, and pattern destructuring.

### Inference Map (`inference.go`)

`InferenceMap` is a `map[ast.Node]Type`. After checking, it contains resolved
types for every expression, statement result, and pattern. The interpreter (and
future tooling) can reuse this map for execution or IDE feedback.

### Obligations (`types.go`)

```go
type ConstraintObligation struct {
    Owner      string      // e.g. "fn useFormatter"
    TypeParam  string      // type parameter name
    Constraint Type        // required interface/expression
    Subject    Type        // concrete type under test
    Context    string      // e.g. "via method 'format'"
    Node       ast.Node    // origin for diagnostics
}
```

Obligations are produced during declaration collection (generic params, where
clauses) and during expression checking (method sets with their own constraints).
The solver consumes this slice to ensure every requirement is met.

---

## Constraint Solving

The solver lives in `constraint_solver.go` and runs after expression checking so
that all obligations have concrete subjects. It performs the following steps:

1. **Collect Resolution Target**
   - Each obligation’s `Constraint` is normalised into an
     `interfaceResolution` value (`resolveConstraintInterfaceType`).
   - This handles three cases:
     - Direct `InterfaceType` references.
     - `AppliedType` nodes (e.g. `Formatter<Wrapper>`).
     - Struct names that refer to interfaces (for legacy shorthand).
   - If the constraint cannot be resolved (unknown interface, wrong arity) a
     diagnostic is emitted immediately.

2. **Interface Arity Validation**
   - The solver checks that the number of provided type arguments matches the
     interface’s type parameters. It distinguishes between “no arguments
     provided” vs. “wrong count” to keep errors actionable.

3. **Satisfaction Check (`obligationSatisfied`)**
   - For each obligation we test:
     1. `subjectMatchesInterface`: Does the subject already *is-a* the interface
        (e.g. it’s typed as `Formatter<T>`)? If so, we’re done.
     2. `implementationProvidesInterface`: Does an explicit impl (`impl Formatter for Wrapper`) satisfy the requirement? This uses `ImplementationSpec` data collected earlier and performs type argument substitution.
     3. `methodSetProvidesInterface`: Does a method set (`methods Wrapper { ... }`) provide the required methods with compatible signatures? If yes, we also gather any secondary obligations the method set introduces (e.g. its own where clauses).
   - If none of the above succeed, the solver returns `false` along with a
     human-readable `detail` string that includes method names, conflicting
     signatures, etc.

4. **Method Set Evaluation**
   - `methodSetProvidesInterface` iterates all method sets registered for the
     subject type. For each set:
       - `matchMethodTarget` checks whether the method set’s target matches the
         concrete subject and returns a substitution map plus a “score” that
         indicates specificity.
       - Interface methods are substituted with `ifaceSubst`, which binds
         interface type parameters to the target (`Self`) and constraint
         arguments.
       - Actual method signatures are substituted with `combined` (method set
         generics + interface arguments), then compared using
         `functionSignaturesCompatible`.
       - If a method set satisfies the interface but introduces additional
         obligations, those obligations are recursively resolved via
         `obligationSetSatisfied`.
       - When failures occur, `annotateMethodSetFailure` decorates the detail
         string with the candidate label (“methods for Wrapper”) and context
         (“via method ‘format’” or “via method set”).

5. **Diagnostics Emission**
   - If an obligation fails, the solver formats a diagnostic:
     ```
     typechecker: fn useFormatter constraint on T is not satisfied:
       Wrapper does not implement Formatter<Wrapper>: methods for Wrapper: method 'format' has incompatible signature (via method 'format')
     ```
   - Context strings originate from the obligation’s `Context` or from method
     set analysis (see `populateObligationSubjects` and
     `substituteObligations` in `type_utils.go`).
   - Diagnostics include the owning function/impl (`Owner`) and point back to
     the AST node that introduced the obligation (`Node`).

### Substitution Helpers (`type_utils.go`)

- `substituteType` recursively replaces type parameters with concrete types.
- `substituteFunctionType` applies substitutions to function signatures while
  pruning type parameters and where clauses that are fully instantiated.
- `substituteWhereSpecs` drops where clauses whose type parameters were replaced
  with concrete types—these would otherwise generate redundant obligations.
- `substituteObligations` rewrites nested obligations when method sets or impls
  introduce additional constraints.

These helpers ensure the solver works with concrete types and keeps diagnostics
readable.

---

## Expression & Pattern Checking Highlights

- **Member Access (`member_access.go`)**
  - Resolves struct fields, array indexes, interface methods, and methods
    provided via impls or method sets.
  - When accessing members on generic type parameters, obligations are recorded
    so the solver later verifies the required interface constraints.

- **Function Calls (`literals.go` + helpers)**
  - Handles generic inference, explicit type arguments, arity checks, and
    lambda inference.
  - Records inferred function types in the `InferenceMap` so downstream calls
    can inspect return/parameter types.

- **Patterns (`patterns.go`)**
  - Checks array/struct destructuring, typed patterns, and match expressions,
    inferring types for bound identifiers.
  - Guard expressions are evaluated in a cloned scope so bindings remain
    consistent after the guard.

- **Async (`concurrency.go`)**
  - `spawn` produces `FutureType` results.
  - Async helper calls (`future_cancelled`, `future_yield`, `future_flush`) enforce
    context rules via diagnostics and maintain the same semantics as the Go
    interpreter.

---

## Diagnostics

- Diagnostics are lightweight structs with the error message and a pointer to
  the originating AST node. Source span integration is postponed until the
  tree-sitter parser emits file/line information; the design and data structures
  already allow for spans once available.
- Messages follow a consistent prefix (`typechecker:`) so fixture manifests and
  tooling can parse them reliably.
- Constraint-related diagnostics leverage the context annotations described
  above to make ambiguous failures actionable.

---

## Usage Patterns

- **Interpreter Integration** – `Interpreter.EnableTypechecker` allows runtime
  callers to run the checker before executing a module. Fixture suites can
  toggle strictness via `ABLE_TYPECHECK_FIXTURES=warn|strict`.
- **Tooling** – external tools (e.g. future compiler, LSP) can re-run
  `CheckModule`, inspect `InferenceMap`, and leverage diagnostics without
  executing code.
- **Testing** – `./run_all_tests.sh --typecheck-fixtures=strict` runs TypeScript
  unit tests, exports fixtures, typechecks them in Go, and then executes the Go
  interpreter tests. The strict mode keeps the entire suite honest.

---

## Future Enhancements

1. **Source spans** – once the tree-sitter parser emits location info, extend
   diagnostics to carry spans and wire them into tooling.
2. **Incremental checking** – explore caching of declaration/type information to
   enable faster re-checks in long-running tooling.
3. **TypeScript parity** – port the Go checker approach into the Bun
   interpreter so fixtures can be enforced uniformly.
4. **Performance profiling** – profile substitution and method-set resolution
   for large modules; consider memoisation or specialised caches if needed.

---

## Cheatsheet

- **Entry point:** `typechecker.New().CheckModule(module)`
- **Obligations:** `ConstraintObligation` (owner, type param, constraint, subject, context, AST node)
- **Solver flow:** resolve interface → validate arity → direct match → impl → method set → diagnostics
- **Context strings:**
  - `"via method 'foo'"` – failure occurred inside a specific method obligation.
  - `"via method set"` – failure bubbled out of a method set where-clause.
- **Strict fixture mode:** `ABLE_TYPECHECK_FIXTURES=strict ./run_all_tests.sh`

With this mental model, you should be able to dig into any `pkg/typechecker`
file, understand how new features interact with the solver, and extend the
checker without surprises.
