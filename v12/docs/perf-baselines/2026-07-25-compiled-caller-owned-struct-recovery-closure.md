# Compiled caller-owned struct recovery closure

Date: 2026-07-25

## Decision

Retain no compiler or runtime code from the caller-owned generated-struct
recovery experiment.

The prototype added one general `from_into` recovery form, decoded into
caller-provided Go storage with failure-atomic commit, and connected recovered
function results to the existing caller-owned-result propagation. It named no
application type or container. The complete prototype was removed because it
improved only Concurrent Event Routing; Manifest Normalization and Binary
Event Log retained the same heap allocation counts.

Machine-readable evidence is in
`2026-07-25-compiled-caller-owned-struct-recovery-closure.json`.

## Caller attribution

The current strict binaries and their exact main-phase profiles reproduce the
candidate owner in all three applications:

- Concurrent Event Routing allocates 4,100 `EventTask` and 4,100
  `RoutedEvent` recovery objects. Their immediate callers are specialized
  compiled `Channel.receive` methods, whose results flow to compiled workers
  and the compiled collector loop.
- Manifest Normalization allocates 3,072 five-field `ManifestRecord` recovery
  objects in the erased normalizer closure immediately before the compiled
  `normalize` call.
- Binary Event Log allocates 53,248 five-field `EventRecord` recovery objects
  in the erased scorer closure immediately before the compiled `score_event`
  call.

The existing ten merged baseline CPU profiles per application support the same
caller shapes. The prototype was evaluated with three new exact allocation
runs per application.

## Prototype and semantic findings

The prototype:

1. decoded every ordinary generated nominal through a shared `from_into`
   helper;
2. converted all fields into temporary native storage;
3. committed the destination only after complete successful conversion;
4. kept allocating wrappers for dynamic and unproven lifetimes;
5. propagated caller storage through a recovered compiled return; and
6. preserved a nil recovered result rather than dereferencing caller storage
   when a channel closed.

Generated Event Routing then passed caller storage through both specialized
channel receive methods, and the selected 8,200 recovery objects disappeared.

That did not generalize to the other applications. Their erased closures
perform conservative runtime-origin writeback after the static call:

```text
runtime.Value -> native record -> compiled static call -> __able_struct_*_apply
```

Go escape analysis therefore moved the caller-owned record to the heap. Exact
profiles merely moved Manifest's 3,072 objects and Binary's 53,248 objects
from `__able_struct_*_from` to the generated caller's local declaration.
Manifest's compiled `normalize` body also closes over a field reached through
the record pointer, independently preventing a simple stack-lifetime proof.

The writeback is semantically necessary unless the compiler can prove the
callee neither mutates nor lets the nominal parameter escape. Suppressing it
without a conservative effect proof would break mutation, aliases, captured
references, nested/cyclic values, and dynamic-boundary observability.

## Exact allocation A/B

Three verified exact main-phase allocation runs used CPU 9,
`GOMAXPROCS=1`, `GOMEMLIMIT=1GiB`, and `GOGC=50`.

| Application | Bytes baseline -> prototype | Change | Allocations baseline -> prototype | Change |
| --- | ---: | ---: | ---: | ---: |
| Concurrent Event Routing | 4,484,386.67 -> 4,287,650.67 | -4.39% | 77,944.33 -> 69,744.67 | -10.52% |
| Manifest Normalization | 3,907,458.67 -> 3,907,485.33 | +0.001% | 84,933.67 -> 84,934.00 | +0.0004% |
| Binary Event Log | 62,725,448.00 -> 62,725,421.33 | -0.00004% | 1,078,006.00 -> 1,078,006.00 | 0.00% |

The three-unlike admission gate failed. Candidate CPU profiling and
twenty-cohort wall-time measurement were intentionally not advanced because
exact allocations already proved that two rows did not improve.

## Verification and cleanup

- All three prototype binaries built under `--no-fallbacks`, passed their
  public verifiers, and omitted `able/interpreter-go/pkg/interpreter` from
  their dependency graphs.
- All nine exact-allocation processes passed verification.
- After removing the prototype, focused caller-owned-result, nominal
  conversion, static-dispatch, and error-payload tests pass.
- `go test ./cmd/ablec -count=1 -timeout=60s` passes.
- Every prototype compiler/test file was byte-for-byte restored to the
  predecessor retained state, and the added recovery file was removed.
- No stdlib, runtime, interpreter, VM, language, dependency, or WASM change
  was retained.
- Raw artifacts:
  `/tmp/able-caller-owned-struct-recovery-20260725.fx9vQI`

## Next

Qualify read-only, non-escaping static nominal-parameter effects across at
least three unlike strict applications before revisiting writeback suppression
and caller-owned recovery.

This is next because the experiment proved that mandatory runtime writeback,
not the syntax of `__able_struct_*_from`, is what forces Manifest and Binary
records to escape. The work entails surveying for a third material
`from -> static call -> apply` application, building a conservative transitive
effect summary for mutation and escape through aliases, returns, closures,
calls, methods, imports, generics, interfaces, and dynamic dispatch, and only
then testing removal of redundant writeback.

This is important because a sound read-only/non-escaping proof could remove
the entire decode/use/re-encode boundary for unlike programs. Without that
proof, caller-owned recovery only relocates the same heap allocation and
cannot move compiled Able materially closer to native Go.

Do not begin WASM work.
