# Primitive-value representation feasibility (2026-07-12)

## Outcome

Do not prototype a new primitive carrier in the current runtime. The proposed
change is not a VM-local optimization: a carrier that crosses all ordinary
Able boundaries must replace the shared `runtime.Value` storage and dispatch
contract. The current `runtime.RawValue` cannot do that safely, and a partial
conversion would reintroduce the same materialization boundary under a new
name.

This is a design rejection, not a claim that primitive representation work can
never help. A future proposal is eligible only as a language-wide runtime
architecture project with the proof matrix below; it is not eligible as a
benchmark, named-container, compiler-lowering, or bytecode opcode fast path.

## Established contracts

Able values have concrete runtime types, and primitive width, assignment,
patterns, interface dispatch, collections, host conversion, and asynchronous
results are observable language behaviour. The active specification requires
concrete types in value positions and runtime typed-pattern checks; it also
defines mutable array/struct/map values, dynamic interface values, unions, and
host handles. An implementation may choose an internal representation, but it
must preserve those behaviours in the tree walker, bytecode VM, and generated
binary.

The current Go contract is deliberately simple:

| Boundary | Current stable carrier | Why it cannot admit VM-private raw cells |
| --- | --- | --- |
| Environments and closures | `runtime.Value` | names may escape a bytecode frame and be read by either interpreter or a callback |
| Arrays, maps, struct fields, errors, packages | `runtime.Value` | storage, mutation, hashing, reflection, and nominal payloads need stable language values |
| Interfaces, unions, patterns, dynamic calls | `runtime.Value` | type matching and method lookup require a public `runtime.Kind` plus concrete value semantics |
| Native and extern calls | `[]runtime.Value` / `runtime.Value` | host conversion and generated bridge helpers use concrete integer, float, bool, char, array, and nil representations |
| Futures and iterators | `runtime.Value`; optional `RawValue` iterator/native lane | futures can outlive a VM frame; the existing raw lane is explicitly transient and materializes for ordinary callers |
| Compiler generated code | `runtime.Value` plus primitive Go locals | emitted bridges, error controls, closures, and dynamic fallbacks share the runtime contract |

The bytecode VM has a separate transport representation: raw integer/float
cells on its stack and slots, `NativeFunctionValue.RawImpl`, and
`IteratorValue.NextRaw`. Its public exits deliberately use
`bytecodeMaterializeRawValue` / `RawValue.Materialize` before a value enters
the stable contract.

## Candidate analysis

### 1. Make `runtime.RawValue` implement `runtime.Value`

Rejected. Go cannot give `RawValue` both its current `Kind() RawValueKind` and
the `runtime.Value` method `Kind() runtime.Kind`; changing that method would
break the raw transport API. More importantly, ordinary runtime and compiler
code use concrete type switches and methods such as `IntegerValue.ToInt64`,
integer suffix checks, `FloatValue.TypeSuffix`, `BoolValue`, `CharValue`, and
`NilValue`. Passing a raw integer as a `runtime.Value` would make those checks
fail unless every consumer gained an unwrapping branch. That moves a generic
test onto every hot and cold value operation while still omitting bool, char,
nil, void, references, and nominal values.

### 2. Extend `RawValue` with all primitive types but materialize for ordinary
storage and calls

Rejected. This is the existing VM-local design with more tags. It may avoid a
specific producer allocation, but it cannot remove the call, closure,
environment, collection, interface, extern, or future boundary. It therefore
cannot be demonstrated as a broadly applicable improvement, and previous
raw-cell/call-return experiments already regressed their broad guards.

### 3. Introduce a universal tagged runtime cell

Not rejected in principle, but not eligible for a performance prototype. This
would replace `runtime.Value` at every storage and dispatch boundary with a
new canonical representation that can hold every primitive bit pattern and a
reference to every nominal value. It requires all three runtimes, generated
compiler code, native APIs, host reflection, equality/hash semantics,
concurrency hand-off, and error paths to use the same abstraction. It also
needs a specified representation for `u64` high-bit values, `u128`/`i128`
BigInt promotion, NaN/float suffixes, nil/void, and identity-bearing values.

That is a language/runtime architecture project with a new performance model,
not a safe local change. Until its proof exists, replacing stable values with a
tagged cell would be an unbounded semantic migration whose likely overhead is
larger than the rejected local checks.

### 4. Add compiler-only raw lowering

Rejected. Generated programs intentionally exchange `runtime.Value` at
dynamic, native, nominal, and extern boundaries. A compiler-only carrier would
duplicate bytecode semantics, diverge from the tree walker, and either require
type-name exceptions or materialize at the same public boundaries. It cannot
be justified by `bridge.ToInt`/`bridge.ToUint` samples from a HashMap, string,
FASTA, or fixed-width-integer workload.

## Proof matrix required before a future design

The following matrix is the minimum gate for a new canonical carrier. Every
row must execute the same source fixture in the tree walker, bytecode VM, and
compiled binary; external/host rows also need the existing Go extern harness.
The listed current tests are boundary evidence only, not proof that a new
representation is valid.

| Area | Required fixture behaviour | Existing boundary evidence | Required new proof |
| --- | --- | --- | --- |
| Primitive widths | round-trip i8/i16/i32/i64/u8/u16/u32/u64/isize/usize plus high-bit u64/usize and i128/u128 promotion; preserve suffix, overflow, equality, and casts | `TestRawIntegerMaterializePreservesUnsignedHighBit` | one cross-runtime fixture including each width and error cases |
| Float and scalar identity | f32/f64 normalization, NaN/infinity policy, bool, char, nil, void, and type-pattern outcomes | bytecode raw-float/raw-integer tests | cross-runtime scalar and typed-pattern fixture |
| Calls and closures | direct calls, recursion, closures/captures, partial/bound methods, return/error/control transfer | raw exact-native and completed-run tests | source fixture crossing every call form with primitive values |
| Environment escape | declaration, assignment, shadowing, and a captured primitive after the creating frame exits | `TestBytecodeVM_EnvironmentNameWritesMaterializeRawI32Values` | shared closure/shadowing fixture |
| Nominal values | primitive struct fields, positional fields, nullable/union payloads, interfaces, method dispatch, and generic arguments | existing struct/interface tests | shared nominal/interface/union fixture |
| Collections | primitive Array read/write/promotion, HashMap key equality/hash, nested arrays/maps, iterator yield/collect | mono-array and raw iterator tests | shared mutable Array/HashMap/iterator fixture |
| Host/native | ordinary `Impl`, `RawImpl`, Go extern argument/result, host slices, and error paths | `TestBytecodeVM_StructCallableFieldRawImplReceivesRawI64Arg` | one extern plus native fixture proving no raw transport escapes host APIs |
| Concurrency | `spawn`, `Future`, cancellation, await, and iterator/future values outliving the originating VM frame | future/iterator suites | shared scheduling fixture with primitive result/error values |
| Observable parity | `Kind`, stringify, type-match, equality, hashing, diagnostics, and stdout must match across all three engines | materialization boundary tests | fixture-harness parity assertion and compiled executable test |

Any design must first state where materialization is impossible, where it is
permitted, ownership/lifetime rules for references, and how the tree walker and
compiler obtain exactly the same results. It must then pass this matrix before
benchmarking. A benchmark win with an uncovered row is a rejection, not a
candidate.

## Decision and historical next recommendation

Retain `runtime.RawValue` as bytecode-local transport and retain
`runtime.Value` as the stable language boundary. No runtime, compiler,
tree-walker, benchmark, or `able-stdlib` change follows from this audit.

The feature-surface audit and six additional verifier-backed application lanes
are complete. The resulting 22-application ledger did not reveal a missing
independent workload or a repeated local helper. The follow-on canonical
carrier architecture decision is recorded in
`v12/design/canonical-runtime-value-architecture.md`: it rejects a universal
`Cell` prototype in safe Go, retains `runtime.Value` as the dynamic ABI and
`RawValue` as VM-private transport, and directs the next investigation to
existing compiler dynamic-carrier boundaries under the staged AOT policy.
