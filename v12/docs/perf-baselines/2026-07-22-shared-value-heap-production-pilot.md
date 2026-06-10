# Shared value/heap production pilot

## Decision

**`reject-live-cell-call-veneer-require-native-shared-ownership`**.

The bounded opt-in pilot was implemented at the one generic Able-function
call/return boundary shared by the tree-walker and bytecode interpreter. Small
Bool, Char, Nil, Void, Integer, Float, and IteratorEnd values crossed as the
existing 16-byte pointer-free semantic cell. Identity-bearing values, wide
integers, and host values retained the exact original Go `runtime.Value`.
Reference and indirect cells were rejected rather than decoded through a
second heap or graph conversion.

The semantic vectors passed, but the performance gate did not. Across two
independent five-process cohorts in each configuration, the candidate
regressed four of six application/mode rows, was effectively neutral in one,
and improved only the volatile Dependency Plan bytecode row by 2.65%. The
complete live-path candidate was therefore reverted. No interpreter, runtime,
compiler, stdlib, language, dependency, or WASM production change remains.

## What was tested

The pilot wrapped `invokeFunction`, the shared semantic entry used for ordinary
Able function calls in both interpreters. It did not claim to cover bytecode
inline call frames, which bypass that boundary by design.

The adapter used shared cells only for direct immediate payloads. Every
identity-bearing or not-yet-shared value used exact pass-through fallback.
There was no graph snapshot, shared-heap allocation, foreign heap, named
container branch, or non-primitive nominal special case.

Focused conformance covered:

- exact small signed Integer and Float suffix/bit round trips;
- exact Array pointer identity through an argument and return;
- exact wide-Integer backing identity through fallback;
- unchanged raised-error propagation with no manufactured result;
- deterministic rejection of reference and indirect cells without shared
  ownership;
- the same vectors under tree-walker and bytecode execution.

A broader opt-in call/return, iterator, partial-call, rescue, and raise guard
set also passed under the one-minute test cap.

## Repeated-process evidence

Array Slice Window supplies the array/iterator family, Dependency Plan supplies
graph/control behavior, and Option/Result Config supplies nominal/error
behavior. Each row has two baseline and two candidate cohorts of five
independent verifier-backed processes. The second pair reversed both
application and mode order. All 120 selection runs verified; none failed or
timed out. Arithmetic means pool the ten samples in each configuration.

| Application | Mode | Baseline cohorts (s) | Candidate cohorts (s) | Baseline pooled (s) | Candidate pooled (s) | Delta | Result |
| --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| Array Slice Window | bytecode | 0.566 / 0.640 | 0.588 / 0.656 | 0.603 | 0.622 | +3.15% | regress |
| Array Slice Window | tree-walker | 6.074 / 7.628 | 6.832 / 6.930 | 6.851 | 6.881 | +0.44% | neutral/volatile |
| Dependency Plan | bytecode | 0.542 / 0.440 | 0.440 / 0.516 | 0.491 | 0.478 | -2.65% | isolated/volatile |
| Dependency Plan | tree-walker | 5.472 / 5.764 | 6.262 / 6.778 | 5.618 | 6.520 | +16.06% | regress |
| Option/Result Config | bytecode | 0.668 / 0.662 | 0.656 / 0.972 | 0.665 | 0.814 | +22.41% | regress/volatile |
| Option/Result Config | tree-walker | 1.956 / 1.676 | 1.784 / 2.092 | 1.816 | 1.938 | +6.72% | regress/volatile |

The tree-walker candidate also raised pooled GC counts in all three
applications: 14.1 to 18.4, 33.9 to 37.0, and 7.2 to 8.0. This is consistent
with temporary boundary slices and reconstructed immediate interface values
escaping into the current Go-owned execution path.

The volatile rows remain visible and are not used as precise point estimates.
They do not rescue the candidate: the stable Dependency Plan tree-walker row
regresses 16.06%, the stable Option/Result bytecode baseline becomes volatile
and 22.41% slower, and no broad improvement repeats across both interpreters.

## Interpretation

A cell veneer over Go-owned values adds packing, unpacking, interface
reconstruction, and temporary storage while removing none of the current
environment, collection, frame, or semantic-operation ownership work. Exact
fallback makes the adapter correct, but correctness alone cannot turn a
representation conversion at a narrow boundary into a speedup.

This rejection does not invalidate the standalone semantic ABI or its
conformance model. It closes the assumption that they can be introduced one
call boundary at a time for performance. A profitable shared representation
must be native on both sides of a semantically closed execution region and
must own its values for the lifetime of that region.

## Next recommendation

Complete a **`shared-runtime-closed-region-cutover-decision`** before another
live implementation candidate.

Map the minimum semantically closed production cut that can keep cells,
identity-bearing objects, environments, calls, returns, errors, and host roots
under one shared owner for an entire function-entry region. Measure its static
reach and model its implementation and target-excess budget against at least
three unlike whole applications. Advance only if the region avoids all
pack/unpack veneers and graph conversion, preserves exact pre-entry fallback,
and can remove at least 25% of target excess in every governing row.

Why: this pilot demonstrates that a narrow adapter pays migration cost without
removing the Go-owned work. The next useful question is therefore whether a
closed ownership cut exists at all—not how to make this rejected veneer
slightly cheaper. If no such cut can be staged without a wholesale duplicate
runtime, close shared-runtime production migration and return to the portable
application frontier. No WASM work is included.
