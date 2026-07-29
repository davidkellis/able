# Shared-runtime closed-region cutover decision

Date: 2026-07-22

Status: production migration closed; no staged three-family cut exists

## Decision

**`close-shared-runtime-production-migration-no-three-unlike-closed-cut`**.

The performance opportunity is real, but the production ownership cut is not
staged broadly enough to pursue. All five hot-function models clear the 25%
target-excess bar. Only Distance Field can be enclosed below an entire
interpreter instance, and that is one float-numeric family rather than the
required three unlike families.

No runtime, interpreter, compiler, canonical-stdlib, dependency, benchmark,
language, or WASM implementation change is retained by this decision.

## Cut contract

- `single_semantic_owner`
- `exact_pre_entry_fallback`
- `no_live_frame_conversion`
- `no_identity_graph_conversion`
- `no_call_return_cell_veneer`
- `no_per_operation_go_transition`
- `no_duplicate_language_runtime`
- `bounded_below_interpreter_instance`

A cut must select ordinary-VM fallback before function entry, retain one
semantic owner throughout the region, and avoid live-frame conversion, object
graph conversion, call/return cell veneers, and per-operation Go transitions.
A whole interpreter-instance replacement is not a bounded migration stage.

## Static ownership reach

The current Go runtime stores semantic values across all 7
ownership domains below:

| Domain | Static source markers |
| --- | ---: |
| `value-representation-and-object-identity` | 3 |
| `collections-mutation-and-alias-storage` | 3 |
| `environments-closures-and-globals` | 3 |
| `calls-native-boundaries-and-host-roots` | 3 |
| `control-errors-and-evaluation-state` | 3 |
| `bytecode-slots-stacks-and-call-frames` | 3 |
| `definitions-types-and-dynamic-dispatch` | 3 |

These are not independent adapters. Arrays and maps contain `Value`; closures
capture environments; evaluation signals, bytecode slots, call frames, native
functions, definitions, and host roots all retain the same identity graph.

## Whole-function cut model

| Application | Hot function | Identity crosses entry/exit? | Minimum closed scope | Ownership domains | Target excess removed | Bounded? |
| --- | --- | --- | --- | ---: | ---: | --- |
| `fixed_width_128` | `ordered_select_checksum` | yes | `interpreter-instance-runtime` | 6 | 48.40% | no |
| `distance_field` | `main` | no | `function-plus-transitive-primitive-kernel` | 3 | 59.83% | yes |
| `word_frequency` | `split` | yes | `interpreter-instance-runtime` | 7 | 39.78% | no |
| `array_slice_window` | `rolling_checksum` | yes | `interpreter-instance-runtime` | 7 | 63.36% | no |
| `reverse_complement` | `reverse_complement_fasta` | yes | `interpreter-instance-runtime` | 7 | 53.14% | no |

Application-specific closure reasons:

- **`fixed_width_128`**: The hot function creates and returns a nominal UInt128 object into a legacy caller. Preserving its exact identity without a return veneer requires the caller, nominal constructors, dispatch metadata, environments, and frames to share the new owner.
- **`distance_field`**: The entry has no identity-bearing arguments or result. A bounded cut is possible only if hypot, sqrt, and the f64 sqrt kernel operation execute inside the semantic runtime; print remains one declared terminal host effect.
- **`word_frequency`**: Each hot split call receives legacy String identities and constructs an Array of Strings consumed by iterator and map logic. Avoiding per-call or graph adapters expands ownership through strings, arrays, iterators, maps, dispatch, environments, and frames.
- **`array_slice_window`**: The legacy caller owns the input Array and the function repeatedly slices, reads, type-matches, and may raise IndexError. A borrowed handle would cause per-operation exits; conversion would break alias and error identity.
- **`reverse_complement`**: The function receives a file-owned Array and returns a newly allocated Array to write_all while mutating several nested arrays. Exact identity and mutation require the I/O roots, containers, calls, environments, and frames to share one owner.

Distance Field is the one bounded exception because `main` has no
identity-bearing parameter or result. It would still require `hypot`, `sqrt`,
and the f64 sqrt kernel operation to execute inside the semantic runtime; the
final print can remain a declared host effect. Building that route would optimize
only one workload family, so it fails the broad-applicability gate.

The other four cuts reach legacy nominal/String/Array identities. Keeping those
objects Go-owned creates repeated semantic exits; converting them breaks alias
identity; migrating their transitive owners reaches environments, frames,
collections, dispatch, errors, and host roots. That is a wholesale runtime
cutover, not an incremental region.

## Budget conclusion

- Performance-material rows: 5/5.
- Bounded adapter-free rows: 1/5.
- Unlike bounded families: 1/3 required.
- Interpreter-instance cuts: 4/5.
- Production implementation admitted: no.

The standalone semantic ABI, codec, shadow images, and conformance models remain
valid research artifacts. They no longer constitute a production migration
roadmap under the current Go-owned runtime.

## Next recommendation

Complete **`portable-concurrent-numeric-application-frontier`**.

Add one deterministic numeric/data-parallel portable application covering concurrency, control flow, and arrays/files, with Able tree-walker, bytecode, compiled, Go, Python, and Ruby implementations and verifier-backed outputs.

Why: The current interaction frontier's highest-ranked minimum-depth triple is concurrency plus control flow plus arrays/text/files. Existing members are text-heavy; a numeric workload can distinguish scheduler cost from string, map, and dispatch costs.

The checked interaction frontier ranks
`concurrency + control_flow + expressions_arrays_text_files` first at depth
3, adjacent to
43.843579 seconds of current target excess.

Admission: Take five independent samples per timed lane, pool additional cohorts when volatile, profile both Able engines, and admit an optimization only when the same generic owner recurs across at least three unlike applications with a modeled material target reduction.

Exclusions: No WASM work, benchmark-name branches, named-container lowering rules, non-primitive nominal special cases, or retry of a closed shared-runtime/backend mechanism.

## Reproduction

```sh
python3 v12/bench_shared_runtime_closed_region_cutover_test.py
v12/bench_shared_runtime_closed_region_cutover --check
just bench-architecture-budget-check
```
