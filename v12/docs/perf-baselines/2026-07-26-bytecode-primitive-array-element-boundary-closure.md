# Bytecode primitive Array element boundary closure

Date: 2026-07-26

## Decision

Retain no production change. The `CallMemberArraySlot` parent is hot and the
raw-`i32` materialization count is large, but boxing itself is not a material
CPU or allocation owner in three unlike applications. The three largest
regex-oriented rows also execute the same four canonical stdlib NFA sites, so
they do not establish independent Array storage shapes.

No candidate reached the implementation or A/B gate.

## Measurement contract

- CPU 6, `GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`.
- Disk-backed `/var/tmp` workspace and Go temporary directory.
- Canonical external stdlib with source-root-only module resolution.
- Catalog-selected serial executor and public verifier for every application.
- Three fresh one-main allocation-counter processes per application.
- One diagnostics-off main-only CPU profile per application.
- One measured-main sampled allocation profile per application.
- One main-only reason-aware materialization census per application.

The frozen CLI hash is
`f68f4e7314300999f257395da2a34d0c0d14e20557757f6535da89aa6466095c`;
the interpreter-test hash is
`aa525490df683d83e9b26f59723e2122139cbf209023829e89ed9475bbf7dad2`.
These match the restored artifacts from the preceding tranche.

Policy’s benchmark harness required typechecking during setup. Skipping it
leaves `map` overloads ambiguous, while the normal CLI and typechecked
one-main harness both execute correctly. CPU and allocation intervals still
cover only `main`.

## Storage-path reconciliation

The current generic primitive Array path is deliberate:

1. `writeArraySlotValueFastChecked` materializes a raw scalar before writing
   tracked state or the runtime Array store.
2. `appendTrackedArrayValueFast` materializes before appending tracked state.
3. Array metadata then extracts the same `i32` into
   `CachedI32Values`, preserving a native read/index sidecar.
4. Internal unmaterialized-state machinery remains available for specialized
   synchronization, but `ensureArrayState` and public/dynamic element access
   restore runtime values.

This matches the existing representation contract: Array storage is an
aggregate escape boundary. Changing it would reopen the previously corrected
identity/alias/public-value contract, so fresh evidence must show material
breadth rather than only frequent calls.

## Diagnostics-off profiles

| Application | Main mean | Bytes mean | Allocations mean | Array-slot CPU cumulative | Materialization CPU cumulative |
|---|---:|---:|---:|---:|---:|
| Policy Record Dispatch | 7.3607s | 141,728,285 | 1,391,441 | 19.89% | no samples |
| Regex Set Audit | 4.2416s | 45,764,632 | 327,971 | 27.34% | 2.03% |
| Log Routing Redaction | 2.9898s | 24,501,291 | 184,377 | 22.99% | 0.73% |
| Array Slice Window | 0.6019s | 17,512,859 | 415,598 | 20.97% | no samples |

The broad Array-slot parent includes dispatch, cache validation, Array
creation, backing growth, leases, value views, synchronization, and element
operations. It cannot be attributed to boxing as a unit.

Sampled allocation attribution makes that separation concrete:

| Application | Boxing/materialization bytes | Boxing/materialization objects |
|---|---:|---:|
| Policy Record Dispatch | 4.02% | 7.56% |
| Regex Set Audit | 1.23% | 2.21% |
| Log Routing Redaction | below sampled top set | below sampled top set |
| Array Slice Window | below sampled top set | below sampled top set |

Policy and Regex contain a measurable boxing allocation, but it does not repeat
materially in Log Routing or the independent Array/slice control. Therefore it
fails both the CPU and allocation breadth gates.

## Exact site census

| Application | `CallMemberArraySlot` raw-`i32` materializations | Sites |
|---|---:|---:|
| Policy Record Dispatch | 697,712 | 17 |
| Regex Set Audit | 566,764 | 17 |
| Log Routing Redaction | 346,302 | 18 |
| Array Slice Window | 288,259 | 2 |

For Policy, Regex Set, and Log Routing, almost all transitions come from
`regex_nfa_upsert_thread` at canonical stdlib lines 633, 637, 645, and 649:
two `write_slot` calls and two `push` calls for the thread-start Arrays. These
applications provide workload breadth but not independent mechanism breadth;
they reuse one implementation.

Array Slice contributes a distinct shape: 288,002 transitions are
`Array.slice` line 58 pushing copied `i32` elements. Despite the high count,
the materialization helper is absent from its CPU and sampled-allocation top
sets.

Frequency therefore overstates the optimization opportunity. The material
cost is a one-family regex allocation descendant, while the independent Array
shape is dominated by other Array work.

## Closure

The fresh census does not invalidate the earlier Array aggregate-boundary
decision or the rejected generalized raw-Array routes. A candidate that kept
raw values in generic Array storage would alter public/dynamic representation
and alias semantics to remove a cost that is not independently material in
three applications.

No VM, runtime, compiler, tree-walker, stdlib, benchmark, fixture, language,
dependency, or WASM file changed. Focused Array materialization, metadata,
alias, mutation, cached-`i32`, bounds, and runtime-store tests pass three
times.

The complete `./run_all_tests.sh` handoff passed every coverage, scoreboard,
threshold, non-compiler, and compiler-batch contract; the final bytecode
fixture corpus completed in 90.940 seconds.

The machine-readable companion is
`2026-07-26-bytecode-primitive-array-element-boundary-closure.json`.
Disposable raw profiles and diagnostic outputs were removed after checksums
and summarized evidence were recorded.

## Recommendation

Next reconcile candidate-static return materialization by immediate caller
consumer and ownership. The full census recorded 1,678,255 such transitions
across 39 applications; ordinary `Return` `i64` and `i32` shapes recur broadly,
while the 800,000 fused-binary count is essentially one-family work.

This entails classifying whether each returned primitive is immediately used
by a native opcode, copied into an existing typed slot/raw caller lane, or
actually escapes to a public/dynamic consumer. Refresh diagnostics-off CPU and
allocation profiles for at least three unlike material ordinary-return
applications. Prototype only a local handoff compatible with the existing
scratch/caller ownership model; do not reopen the broad frame or
execution-context ABI.

This is important because an internal static return is a direct native-carrier
boundary. Proving that a repeated subset can remain raw would advance the
no-boxing architecture without weakening public values, dynamic calls, error
control, interfaces, or unions. Do not begin WASM work.
