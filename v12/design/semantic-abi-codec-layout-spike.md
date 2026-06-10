# Semantic ABI codec/layout spike

Date: 2026-07-22

Status: retained; advance to non-executing shadow image lowering

## Decision

**`retain-codec-layout-spike-advance-shadow-image-lowering`**.

The stage-zero immutable data plane satisfies its retention gates. Keep the
isolated pure-Go `internal/semanticabi` package and advance only to an
execution-complete, still non-executing shadow lowering. This tranche changes no
tree-walker, bytecode VM, compiler, runtime value, stdlib, benchmark source,
production dispatch, or application output.

The retained spike does not prove that the proposed ABI can execute Able. Its
current shadow records demonstrate exhaustive syntax coverage, stable IDs,
serialization, validation, and source identity. They intentionally do not yet
encode real registers, value flow, basic-block edges, cleanup regions, or call
resolution.

## Retained implementation

`v12/interpreters/go/internal/semanticabi` contains:

- a 16-byte `Cell { tag:u32, aux:u32, payload:u64 }` with no pointer-bearing
  field;
- generation-checked 64-bit object identities with zero index/generation
  reserved;
- a generated exhaustive manifest for all 31 current `runtime.Kind` values;
- a generated dense 23-operation structural semantic manifest, separate from
  the ordinary VM's private fused opcodes;
- a manifest SHA-256 identity embedded in every image;
- an eight-section canonical little-endian image codec;
- deterministic bounds, type, opcode, operand, source, callsite, function,
  block, capability, and ABI-identity validation;
- a shadow lowering package that rejects unsupported AST nodes instead of
  retaining an AST pointer or evaluation fallback.

The serialized sections are:

1. header and semantic-manifest identity;
2. strings and symbols;
3. types and nominal metadata;
4. source spans and callsite identities;
5. constants and immutable data;
6. packages, functions, and signatures;
7. semantic records and control-flow indices;
8. host effects and capabilities.

The codec rejects reordered, overlapping, truncated, version-mismatched, or
manifest-mismatched sections before decoding. The post-decode validator rejects
unknown tags/opcodes, wrong operand counts/classes, bad table indices, reversed
source spans, invalid function ranges, and out-of-range block targets. Repeating
the same malformed decode produces the same prefixed error.

## Generated manifests

`manifestgen` parses the canonical `pkg/runtime/values.go` Kind declaration and
combines it with the deliberately small semantic-operation specification. It
generates stable numeric tags/opcodes and the manifest identity. The checked
generator fails when runtime kinds, classifications, operation shapes, or the
checked-in generated file diverge.

The kind classification remains exactly:

- 7 immediate or overflow-backed primitive kinds;
- 20 shared-heap or immutable-metadata kinds;
- 4 host-registry kinds.

No current bytecode opcode number is exposed by this ABI.

## Representative shadow images

The spike parses the real benchmark sources and shadow-encodes the whole target
function. Every visited canonical AST node must produce one structural semantic
record. An unrecognized node aborts lowering and prevents the function from
receiving the `shadow-eligible` flag.

| Application | Function | Nodes / records | Image bytes | Host effects | AST fallbacks |
| --- | --- | ---: | ---: | --- | ---: |
| Fixed Width 128 | `ordered_select_checksum` | 70 | 3,596 | none | 0 |
| Distance Field | `main` | 89 | 4,641 | `able.math.hypot`, `able.host.print` | 0 |
| Array Slice Window | `rolling_checksum` | 147 | 6,905 | none | 0 |

For all three images:

- repeated encodes are byte-identical;
- decode/re-encode is byte-identical;
- every source path/span and function-call callsite round-trips;
- every instruction operand validates against its declared table class;
- the encoded bytes contain no AST or Go object reference;
- ordinary interpreters and the compiler do not import the spike.

Checked evidence:
`v12/docs/perf-baselines/2026-07-22-semantic-abi-codec-layout-spike.json`.

## What this does not establish

The shadow images are not executable IR. Their single block and preorder
records prove only that the selected syntax can be represented without carrying
AST nodes. They do not yet prove evaluation order, register allocation,
short-circuiting, branches, loop backedges, exact call targets, match decision
trees, raise/unwind behavior, return coercion, or continuation boundaries.

The structural operation manifest may change during the next stage. Only the
value-cell layout, versioned section discipline, stable-ID rule, deterministic
codec, and validation approach are retained by this decision.

## Next recommendation

Complete **`semantic-abi-shadow-image-lowering`**.

Replace the three functions' preorder structural records with execution-
complete but non-executing semantic images: explicit registers/value flow,
real basic blocks and branch targets, resolved call/member identities, typed
constants and operations, match/raise regions, and complete host-effect
continuation metadata. Compare the shadow lowering against the ordinary
interpreter/compiler semantic analysis, but do not dispatch or execute it.

Why: the codec/layout risks are now retired. The next unresolved risk is whether
real Able semantics can be expressed without AST fallback or a benchmark-
specific operation. Proving that before live value migration keeps the work
reversible and tests the shared architecture against wide integers, floating-
point native effects, and generic Array/match/error behavior.

Retain the next stage only if all three functions have real control-flow and
data-flow validation, every call/effect is explicit, no AST fallback remains,
the operation set is generic across the three families, checked images remain
deterministic, and production execution remains unchanged.

Exclusions remain: no foreign heap, cgo runtime, JIT/backend dependency,
executable memory, production dispatch, `runtime.Value` migration, benchmark
branch, named-container/non-primitive-nominal special case, or WASM.

## Reproduction

```sh
just bench-semantic-abi-codec-check
just bench-architecture-budget-check
just bench-evidence-ledger --check
```
