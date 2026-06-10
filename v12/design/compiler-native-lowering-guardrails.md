# Compiler Native Lowering Guardrails

## Status and authority

This is the concise active contract for native Go lowering. It constrains the
direct AST-to-Go compiler; it is not a separate IR, runtime ABI, completion
plan, or source of benchmark-specific optimization work.

- `spec/full_spec_v12.md` defines Able semantics.
- `interpreter-go/pkg/compiler` typechecks the shared AST, collects the static
  program graph, classifies dynamic features, enforces fallback policy, and
  renders Go directly.
- `compiler-go-lowering-spec.md` supplies the detailed language-independent
  rules and target shapes; `compiler-go-lowering-plan.md` and
  `compiler-native-lowering.md` are historical completion records.
- `PLAN.md` selects performance candidates only after a concrete non-nominal
  leaf recurs across unlike verified applications.

## Active guardrails

1. Only primitive Able types may have primitive-specific lowering rules.
2. All non-primitive nominal types—including stdlib and user-defined
   containers—use shared nominal lowering. A named type is never a reason for
   a compiler fast path.
3. Dedicated lowering is permitted only for language syntax, kernel/host ABI
   boundaries, and required runtime services; it attaches to that boundary,
   never to a named nominal type.
4. Static compiled execution must use direct Go carriers, control, and
   dispatch where statically representable. Interpreter evaluation and runtime
   dispatch are reserved for explicit dynamic boundaries.
5. Lowering knowledge belongs in shared synthesis helpers, not emitter-local
   branches. A structural refactor needs a semantic or broadly measured reason.

## Static representation and control

| Able family | Static compiled representation |
| --- | --- |
| primitives and `String` | native Go scalars |
| structs | native Go structs or pointers |
| unions/interfaces | generated native carriers and adapters |
| nullable/result values | native nil-capable or result/union carriers |
| arrays | native slices or compiler-owned slice wrappers |
| callables | Go functions or generated callable carriers |
| dynamic values | boundary adapters only |

Static control uses ordinary Go branches, loops, and returns. Generated helper
boundaries may use explicit return-based control signals; normal Able control
must not use IIFEs or `panic`/`recover`.

Native primitive and Array carriers are mandatory, not experimental modes.
Legacy `ExperimentalMonoArrays` options and CLI/environment spellings may be
accepted for compatibility, but no setting may select runtime-store carriers
for a statically representable Array. Likewise, `_ = expression` must evaluate
and discard a static native value without first boxing it into `runtime.Value`.

## Boundary discipline

Dynamic carriers such as `runtime.Value`, `runtime.ArrayValue`, and bridge
dispatch helpers are allowed only at explicit dynamic language features,
extern/host ABI edges, or values already originating from dynamic payloads.
A valid crossing converts native carrier to boundary form at the edge, performs
dynamic work, then converts back immediately when the result is statically
representable.

The generated dynamic entry surface is limited to `__able_call_value(...)`,
`__able_call_named(...)`, and generated `call_original` wrappers, plus the
necessary extern/host adapters. Carrier-conversion helpers are adjacent to
those edges, never a general static execution substrate.

## Proof and change selection

Every compiler change needs focused behavioral tests and generated-source or
boundary proof appropriate to the path it changes. Static paths must remain
no-fallback under the strict compiler policies; dynamic tests must demonstrate
only documented crossings. The full matrix is a long-running confidence
workflow, not a default sub-minute test.

Allocation breadth alone is not candidate breadth. The 2026-07-21
[generated-allocation shape reconciliation](../docs/perf-baselines/2026-07-21-compiled-generated-allocation-shape.md)
found exact String-conversion and unsigned-boxing helpers allocating in three
unlike families, but each was CPU-material in only two. Selection therefore
requires the same exact lowering/runtime shape to be material in CPU or other
modeled end-to-end cost in at least three unlike applications; a shared Go GC,
allocator, conversion, or generated-wrapper parent is insufficient.

For prior array, union, interface, control, dispatch, boundary, benchmark, and
release milestones, see [the historical native-lowering record](./compiler-native-lowering.md).
