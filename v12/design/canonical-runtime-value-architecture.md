# Canonical Runtime-Value Architecture Decision

Date: 2026-07-12

## Decision

Do not introduce a universal tagged `Cell` as the replacement for
`runtime.Value` in the current Go runtime. This completes the architecture
design gate opened by the performance ledger and rejects a prototype before it
can turn into a partial VM, compiler, container, or benchmark optimization.

The stable dynamic ABI remains `runtime.Value`. `runtime.RawValue` remains a
bytecode-private transport for its existing integer and float lanes. Compiled
statically resolved code continues to use host-native values, as required by
the staged AOT policy in the v12 specification; it must not gain a new dynamic
carrier merely to share an interpreter representation.

This is not a claim that the current `runtime.Value` use is optimal. It means
that a new universal carrier is not presently a sound *general* optimization
in safe Go. The next credible broad investigation is to locate and remove
unnecessary existing dynamic-carrier crossings from otherwise-static compiled
code, not to replace the dynamic ABI everywhere.

## Scope and semantic invariants

This decision covers the interpreter/tree-walker, bytecode VM, shared Go
runtime, generated AOT binaries, native functions, and extern bridges. The
following language observations are non-negotiable:

- Every ordinary dynamic value has an observable `Kind`, concrete type and
  type-pattern behavior. This includes primitive width/suffix information,
  `nil`, `void`, bool, char, float, and identity-bearing nominal values.
- Arrays, maps, struct fields, closures, interface/union views, iterators,
  errors, packages, and futures may retain values after a VM frame returns.
  Their values therefore cannot borrow a stack cell or a transient conversion
  buffer.
- Hash/equality, casts, interface adaptation, diagnostics, native calls, Go
  externs, and `spawn`/Future result delivery must observe the same values in
  the tree walker, bytecode VM, and compiled executable.
- The AOT policy permits `runtime.Value` only for explicit dynamic crossings,
  irreducible runtime-polymorphic adaptation, host/ABI glue, and runtime
  service payloads. Static locals, direct calls, typed Array hot paths, static
  control flow, and statically representable nominal values must remain
  host-native.

The design changes none of those rules and adds no source-level type, syntax,
or standard-library API. In particular, it does not authorize a `HashMap`,
`UInt128`, `Array u8`, generator, Future, or application-name branch.

## Current boundary model

The current Go runtime has a deliberately small public dynamic contract:
`runtime.Value` is an interface with `Kind() Kind`. It carries concrete scalar
objects and reference-bearing values through environments, `[]Value`
collections, maps, struct fields, function/native calls, interface dispatch,
errors, packages, iterators, and futures. `RawValue` intentionally has a
different `Kind()` result and can only represent materialized values plus raw
integer or float payloads. It therefore cannot implement `runtime.Value`.

The valid representation split remains:

| Layer | Carrier | Lifetime and permitted use |
| --- | --- | --- |
| Bytecode operand/slot fast lanes | `RawValue` and VM-private typed cells | transient; box before an ordinary dynamic boundary |
| Tree walker and dynamic runtime | `runtime.Value` | stable dynamic ABI and storage contract |
| Static compiled code | native Go scalars, slices, structs, interfaces, and generated carriers | default for statically resolved Able semantics |
| Dynamic/host/service boundary from compiled code | `runtime.Value` | only the four staged-AOT categories in the specification |

This is an architectural boundary, not merely an implementation convention.
Moving a raw primitive into ordinary storage without a complete language value
would break closure escape, type matching, native/extern calls, or a future
that outlives its creating bytecode frame. Moving all static AOT values into a
dynamic cell would violate the v12 AOT policy and make direct compiled code
less Go-like.

## Universal-cell feasibility

The architecture considered the only potentially broad alternative: a single
internal `Cell` capable of holding every scalar bit pattern or a reference to
every nominal value. It must carry the full 64-bit integer payload *and* an
integer suffix, distinguish f32/f64, bool, char, nil and void, and preserve a
GC-visible reference when it carries an object. It must also support high-bit
`u64`/`usize`, i128/u128 BigInt promotion, NaN behavior, and all existing
dynamic `Kind`/type-match behavior.

| Candidate | Result | Reason |
| --- | --- | --- |
| `RawValue` expanded with more primitive tags | rejected | It still materializes at ordinary storage, call, nominal, host, and Future boundaries, and its incompatible `Kind()` contract makes it non-ABI-compatible. |
| Safe `Cell` holding a `runtime.Value` reference plus raw payload/metadata | rejected | A full payload, suffix/tag metadata, and an interface reference require at least multiple words beyond the two-word dynamic-interface slot on the target Go ABI. `[]Cell` storage would grow and still preserve the interface indirection for references. |
| Pointer-plus-payload `Cell` | rejected | Retaining only an untyped pointer loses the concrete Go dynamic type needed by current generic operations. Recovering it requires a universal heap header/type switch migration and adds a new allocation/indirection boundary. |
| Pointer tagging or NaN-boxing | rejected | It cannot encode all 64-bit integer payloads plus suffixes safely, and pointer manipulation would make GC visibility, Go unsafe rules, host interop, and portability part of the language runtime contract. |
| Universal heap-boxed object model | rejected | It adds allocation and indirection to scalars and static paths, contradicting both the performance target and the AOT native-lowering rule. |

The second row is deliberately conservative: an implementation can optimize a
particular layout, but it cannot erase the metadata and reference requirements
above. A layout that only works for `i32`, floats, or a selected collection is
the existing VM-local typed-lane strategy, not a canonical language carrier.

## Why a full migration is not justified

The 22-application ledger shows material misses across independent application
families, but the attributed leaves remain distinct: checked multiword and
ratio arithmetic, numeric kernels, map/name conversion, iterator/generator
work, byte-array work, recursive search, and scheduling. That makes a
language-wide value change tempting, but it does not prove that an additional
per-value tag, allocation, or type dispatch will improve any of them.

The compiler tree currently contains many textual `runtime.Value` references;
that is an inventory signal, not proof that each reference violates the AOT
policy. Some are required host, dynamic, error, scheduler, or interface
adaptation boundaries. Replacing every one with a cell would obscure the
distinction the specification makes between static native lowering and
intentional dynamic behavior.

No benchmark is evidence for relaxing this decision. A successful fixed-width,
string, map, numeric, or VM microbenchmark would be rejected unless it first
proves a complete representation and all dynamic-boundary semantics.

## Required proof if this decision is reopened

A future carrier project must first supply all of the following before code or
benchmarks decide its fate:

1. A safe-Go, target-supported layout whose measured dynamic slot and
   allocation costs are no worse than the current carrier for primitive and
   reference storage. It must state how Go's garbage collector sees every
   retained reference.
2. A complete encoding and conversion specification for every primitive width,
   high-bit unsigned values, BigInt promotion, f32/f64, NaN/infinity, bool,
   char, nil, void, and every reference/identity-bearing value.
3. Exact rules for value escape, mutation, aliasing, hash/equality, `Kind`,
   typed patterns/casts, union/interface adaptation, diagnostic formatting,
   host reflection, native/extern calls, iterator yield, and Future hand-off.
4. Cross-runtime source fixtures for every row in
   `primitive-value-representation-feasibility.md`: primitives, calls and
   closures, environment escape, nominal values, collections, host/native,
   concurrency, and observable parity. The fixtures must pass in the tree
   walker, bytecode VM, and compiled executable.
5. A migration plan that leaves statically compiled locals and direct calls
   native. Any use in generated code must be assigned to a permitted AOT
   dynamic-boundary category.
6. A three-run, verifier-backed 22-application guard showing improvement in
   more than one independent family without material regressions. A win with
   an uncovered semantic row or only one application is a rejection.

Until all six conditions are written and demonstrated, reopening the carrier
would be speculative architecture work rather than a performance candidate.

## Completed follow-up: existing dynamic-carrier boundaries

The required source-and-generated-code audit of the compiler's existing
`runtime.Value` crossings is complete in
`v12/docs/perf-baselines/2026-07-12-compiled-dynamic-carrier-boundary-audit.md`.
It classifies lowering by semantic boundary, never by an application or nominal
type name, and finds no shared material dynamic carrier in the sampled static
workers. Its one static-loop residual is cold control bookkeeping, not a
performance candidate.

The audit applied these gates, which remain required before any future lowering
change:

1. Enumerate every generated `runtime.Value` local, parameter, result, slice,
   and helper in representative statically compiled applications.
2. Assign each occurrence to one of the four permitted dynamic categories, or
   identify it as a candidate static-lowering escape with its source AST and
   lowering phase.
3. Prove that a candidate is statically resolved and can preserve the same
   type, overflow, error/control, interface, and lifetime behavior using a
   native Go carrier.
4. Add a generator/source audit test and a cross-runtime semantic fixture
   before changing lowering. Use the full 22-application ledger as the broad
   performance guard only after at least two unrelated applications exercise
   the same semantic boundary.

The opt-in reachability telemetry sweep is complete across the 22-application
corpus. Its result is recorded in
`v12/docs/perf-baselines/2026-07-12-compiled-dynamic-boundary-reachability-decision.md`:
it rejects a global dynamic-carrier replacement and isolates repeated generic
static print output as the only material shared boundary eligible for a
separate formatting-parity and broad-guard profile gate.
