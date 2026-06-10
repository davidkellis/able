# Compiled positional named-field consumer reconciliation

Date: 2026-07-26

## Decision

Retain the general named-field consumer correction.

Generated native named structs already cross explicit runtime boundaries as
definition-ordered positional `runtime.StructInstanceValue` payloads. Generated
index helpers and the compiler host bridge now read either supported named
field representation through shared representation-aware accessors instead of
assuming the legacy `Fields` map.

This is the consumer-side completion of the retained general positional
encoding rule. It does not add a compiler carrier, lowering rule, or fast path
for `HashMap`, `Array`, `Box`, or any other non-primitive nominal type. The
existing Array/HashMap index branches remain language/kernel service
boundaries; only their handle retrieval now follows the shared nominal
encoding.

## Correctness failure and root cause

`TestCompilerHashMapNativeCarrierExecutes` constructs two native
`HashMap String i32` values, stores them in a static native Array, then indexes
the recovered values. The generated native-to-runtime converter correctly:

1. resolves the ordinary `HashMap` definition;
2. allocates a one-slot positional named-struct instance;
3. encodes the handle in slot zero; and
4. preserves the nominal definition and type arguments.

`__able_index` then incorrectly looked only in `inst.Fields["handle"]`.
Because the correct representation was positional, it reported
`hash map value missing handle`.

The same stale assumption existed for Array handles and in the generic
runtime-to-host struct bridge. The repository-wide gate exposed the latter as
`TestCompilerGoExternStructBoundaryExecutes: missing Box.value`. The exact
extern test failed five consecutive isolated baseline processes before the
bridge correction.

## Generated-module A/B

The exact failing and passing HashMap modules were preserved before cleanup.
Their generated trees differed only in four lines in `compiled.go`:

- `__able_index` Array `storage_handle`;
- `__able_index` HashMap `handle`;
- `__able_index_set` Array `storage_handle`; and
- `__able_index_set` HashMap `handle`.

Each direct `inst.Fields[...]` read became
`__able_struct_named_field_value(inst, fieldName)`. No generated definition,
carrier, call, registration, import, or application body changed.

## Retained implementation

- Generated index reads and writes use the existing common named-struct field
  accessor for runtime Array and HashMap handles.
- The compiler bridge has one package-local named-field accessor with the same
  map-or-definition-ordered-positional contract.
- Generic struct-to-host conversion uses that accessor.
- Compiler bridge String-byte and Array-handle consumers use the same accessor
  instead of independent map/positional conventions.
- Source guards prove arbitrary generated named structs use positional
  runtime storage and that both generated index helpers avoid direct handle-map
  reads.
- Direct bridge guards prove map-backed compatibility and positional named
  storage, including the complete `Box { value: i32 }` host conversion.

Map-backed values from compatible dynamic or host boundaries remain accepted.
Positional structs declared with `StructKindPositional` retain their distinct
ordered-tuple behavior.

## Strict application gate

Three unlike current applications were built with `--no-fallbacks` and run
three independent times. Every process passed its public verifier:

| Application | Verified | Mean wall time | Mean GC |
| --- | ---: | ---: | ---: |
| K-Nucleotide | 3/3 | 1.7367 s | 286.33 |
| Inventory Reconciliation | 3/3 | 0.1367 s | 13.00 |
| Word Frequency | 3/3 | 0.0633 s | 3.00 |

All three final dependency graphs omit
`able/interpreter-go/pkg/interpreter`. Inspection of each generated module
confirms the affected index read/write helpers use the shared accessor.
These runs are correctness/dependency guards, not a performance claim; this
tranche changes only previously failing or representation-dependent reads.

## Verification

Passing checks include:

- exact HashMap native-carrier execution and the new generated-source guard;
- exact Go-extern named-struct execution;
- the complete compiler Go-extern family;
- full `pkg/compiler/bridge`;
- generic Enumerable/Iterator/filter-map guards;
- generic interface-boundary and Result/Option specialization guards;
- broad imported and shadowed alias guards;
- HashMap `u64` keys, HashSet iteration, Array dynamic callbacks, HashMap
  dynamic callback success/failure, TreeMap, and strict Vector execution;
- `go test ./cmd/ablec -count=1`; and
- the standard `./run_all_tests.sh` handoff: coverage and selection contracts,
  every non-compiler package, all 32 bounded compiler batches, and the complete
  bytecode fixture pass.

The longest reported compiler figures are aggregate batch durations, not
individual test durations. The exact failing tests complete in under two
seconds each.

All touched files remain below 1,000 lines. `generator_render_runtime.go`
remains at 998 lines. `git diff --check` passes.

No canonical stdlib, external dependency, language, spec, tree-walker,
bytecode VM, runtime package, or WASM change was required.

## Cleanup

Generated A/B modules, strict application copies, binaries, Go caches, and
per-process output used the disk-backed
`/var/tmp/able-hashmap-handle-20260726` workspace. They are disposable after
this aggregate record is written; no tranche artifact belongs on RAM-backed
`/tmp`.

## Next recommendation

Refresh current bytecode CPU and exact-allocation profiles across at least
three unlike target-missing applications, then select the largest concrete
non-nominal owner that is material in all three.

Why: the 61-row compiled census has now retained its only newly admissible
shared owner and closed this correctness blocker. The active bytecode-v2
contract also confirms that pooled `i32` register lanes, value-slot `i32`
sidecars, and boundary materialization already exist, so the stale roadmap
instruction to begin typed `i32` slots would duplicate completed work.

What it entails: choose unlike verifier-backed applications from the current
bytecode scoreboard; rebuild one exact artifact per application; collect
repeated main-only CPU and allocation evidence under the existing resource
contract; and intersect exact semantic leaves rather than dispatcher, GC, or
stack-parent totals. Advance a prototype only if one general leaf clears the
three-unlike reach gate, preserves materialization at dynamic/host/error/resume
boundaries, and passes repeated A/B measurements against Python and Ruby plus
the broad bytecode guard. If no owner clears the gate, retain no code and
record the closure.

Why it is important: compiled programs retain interpreter-free native Go
lowering, while bytecode still meets its 95% target in only a small minority
of current rows. A fresh exact-owner intersection is the disciplined way to
find a VM change large and general enough to matter without repeating closed
typed-frame, Array, register-executor, named-container, or benchmark-specific
routes. Do not begin WASM work.
