# Compiled primitive nullable value carrier

## Decision

Compiled primitive nullables use a generated value-plus-valid carrier:

```go
type __able_nullable[T any] struct {
	value T
	valid bool
}
```

This is a compiler representation rule for primitive Able types, not a new
language value or a nominal-type exception. It replaces the prior `*T`
carrier, whose non-nil cases required an escaping Go allocation.

`?Error` remains `*runtime.ErrorValue`. Nullable structs, unions, interfaces,
and other non-primitive nominal values continue through the shared nominal
translation and semantic encoding rules. Static Arrays remain native Go
Array carriers. Dynamic and host boundaries still convert through the
existing runtime representation.

## Semantics

- `nil` is the zero carrier: `valid == false`.
- A present value has `valid == true`, including a present primitive zero.
- Matching, `or`, safe navigation, propagation, equality, and nil comparison
  test `valid` and read `value` only on the present branch.
- Static calls, returns, locals, interface adapters, union paths, and Array
  methods preserve the carrier without pointer construction.
- Runtime/dynamic encoding creates the existing runtime nullable value only at
  an explicit boundary. Decoding reconstructs the typed carrier.
- `?Error` retains its pointer carrier because the error runtime object is
  already the semantic payload and is not a primitive scalar box.

The compiler emits shared generic construction and conversion helpers rather
than a separate algorithm for each nominal type. Primitive helper stems exist
only to connect the existing width-specific runtime bridge functions.

## Soundness matrix

| Boundary | Retained behavior |
| --- | --- |
| Static parameter and return | `__able_nullable[T]` stays typed |
| Local, assignment, and join | zero carrier represents absent |
| Pattern match and binding | `valid` guards `value` |
| `or` and propagation | present payload remains unboxed |
| Nil and nullable equality | validity participates in equality |
| Safe navigation | result construction uses value carrier |
| Static Array read/mutation | nullable element/result remains typed |
| Interface and generic adapter | generated typed signatures preserve carrier |
| Union path | shared native-union lowering preserves carrier |
| Dynamic/runtime boundary | explicit encode/decode helper |
| Host boundary | existing runtime contract remains authoritative |
| `?Error` | existing pointer carrier retained |

The executable regression for present `0_i64` distinguishes it from absent
`nil`. Focused tests also cover nullable integer, float, character, string,
Array, match, union, interface, Result, safe-navigation, and propagation
paths.

## Performance admission

Three unlike strict applications were measured from a frozen pre-change
compiler and the candidate compiler with five verifier-backed processes each.
Three independent exact main-phase allocation measurements were also taken
per side.

| Application | Wall baseline | Wall candidate | Allocation objects baseline | Allocation objects candidate |
| --- | ---: | ---: | ---: | ---: |
| Generic Slot Buffer | 0.0560 s | 0.0340 s | 264,215 | 1,046 |
| Inventory Reconciliation | 0.1220 s | 0.1100 s | about 553,060 | 282,723 |
| Transaction Ledger Audit | 0.0480 s | 0.0400 s | 115,269 | 105,029 |

The intended nullable allocation owner disappears in all three profiles.
Remaining dynamic-map conversions are explicit runtime boundaries rather than
evidence that static scalar nullables should be pointer-backed.

All strict builds used `--no-fallbacks`, passed their public verifiers, and
omit `pkg/interpreter` from the generated dependency graph. Equivalent Go
controls averaged 0.0046, 0.0079, and 0.0060 seconds respectively. The change
therefore removes a general lowering tax but does not by itself close the
remaining application-level Go gap.

## Non-goals

- No named-container or non-primitive nominal special case.
- No broad compiled/interpreter ABI change.
- No change to bytecode, tree-walker, language semantics, canonical stdlib,
  dependencies, or WASM.
- No attempt to erase required runtime values at explicit dynamic or host
  boundaries.

The complete retained evidence and verification record is
`../docs/perf-baselines/2026-07-30-compiled-primitive-nullable-value-carrier-retained.md`.
