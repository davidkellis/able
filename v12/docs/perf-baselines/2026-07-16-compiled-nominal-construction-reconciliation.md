# Compiled Nominal-Construction Reconciliation (2026-07-16)

## Decision

This tranche keeps no compiler, runtime, benchmark, or canonical-stdlib code
change. Fresh verifier-backed allocation profiles, CPU profiles, generated Go,
and Go escape analysis reject a shared nominal-construction candidate across
Binary Trees, Sudoku Masks, and Fixed Width 128.

The three large allocation totals have different semantic owners:

- Binary Trees performs one required heap allocation for each returned `Node`.
- Sudoku Masks allocates the source-requested transient `Array i32` values and
  their backing slices in `find_best_empty`.
- Fixed Width 128 sends the primitive Able `i128`/`u128` family through
  `runtime.Value`, `math/big`, compiler bridge casts, AST type construction, and
  interpreter primitive operations.

There is no constructor, boxing, field-carrier, or escape descendant repeated
across all three. A `Node`, Sudoku position-array, or `UInt128` nominal rule
would violate the generality requirement, so no candidate was built.

## Method

All three applications were compiled from the current tree with canonical
`../able-stdlib`, retained generated source, a 60-second per-process timeout,
and their external Ruby verifiers. Diagnostic allocation and CPU launches used
`GOMAXPROCS=1`, `GOGC=50`, and the goroutine executor. CPU and allocation data
came from separate processes, and every profile launch reproduced the verified
output SHA-256.

The generated modules were then rebuilt with:

```text
go build -gcflags='able/compiled=-m=2'
```

This provided source-line escape decisions without modifying generated or
production code. The one-run build measurements below are smoke/verification
rows, not selection timing claims:

| Application | Verified real | User | GC cycles | Output SHA-256 |
| --- | ---: | ---: | ---: | --- |
| Fixed Width 128 | 5.27 s | 7.93 s | 128 | `eceabf5869b1abca8d6dd228b64a09f89e4e98ba8cabc4833ffee1218dafa56a` |
| Sudoku Masks | 5.79 s | 7.74 s | 95 | `35a81e448daf9986f2a9b7c3a873dc6216bd55c969efec50c9b1de6d866659ec` |
| Binary Trees | 4.08 s | 40.21 s | 69 | `341de11a51feab3d8122b4b5d6a68b038a2d14434aa9bc2372f39300bf5f48e1` |

Binary Trees' smoke row used the benchmark's normal parallel goroutine
execution, explaining real time below aggregate CPU time. The normalized
single-P profiles below are used for attribution.

## Allocation reconciliation

| Application | Sampled allocation space | Sampled objects | Dominant exact owner |
| --- | ---: | ---: | --- |
| Binary Trees | 9,357.99 MiB | 611,706,081 | `make_tree`: 99.68% bytes, 99.93% objects |
| Sudoku Masks | 2.87 GiB | 158,625,943 | `find_best_empty`: 98.72% bytes, 99.69% objects |
| Fixed Width 128 | 3.90 GiB | 94,826,320 | mixed `math/big`, bridge, AST, and interpreter primitive work |

### Binary Trees

Line attribution places all material flat allocation on the two successful
returns:

- leaf `&Node{Left: nil, Right: nil}`;
- branch `&Node{Left: left, Right: right}`.

Go escape analysis confirms both values escape through the function result.
The control/error returns are cold and have no sampled flat allocation. The
generated representation is already a native `*Node` with two `*Node` fields,
and the profile is approximately 16 bytes per returned node. Replacing this
with a benchmark-shaped arena, flattening rule, or source-specific tree
representation would change the general nominal model rather than remove
compiler scaffolding.

### Sudoku Masks

Line attribution places all material allocation on the source's per-empty-cell
`Array.new()` and three pushes:

- 63,833,501 sampled objects at the `&__able_array_i32{}` wrapper;
- 32,309,731 at the first append growth; and
- 61,997,988 at the later append growth.

Escape analysis confirms the wrapper escapes because it may become `best`, and
all three appends escape through the wrapper's stored slice. CPU attribution is
also source-local: `find_best_empty` owns 88.97% cumulative CPU,
`runtime.growslice` 38.62%, and `runtime.mallocgc` 50.70%. Prior corrected
Sudoku work already closed a source-shape, nested-array, bit-mask, or default
capacity optimization; this profile independently reaches the same decision.

### Fixed Width 128

Fixed Width does not expose a nominal-constructor wall comparable to Binary
Trees. Its largest allocation owners are:

- `math/big.nat.make`: 34.58% of bytes / 37.24% of objects;
- `bridge.ToUint`: 17.22% / 17.36%;
- `ast.NewSimpleTypeExpression`: 11.87% / 8.19%;
- `ast.NewIdentifier`: 10.51% / 7.25%; and
- interpreter `evaluateBitwise`: 8.84% flat bytes, 37.08% cumulative.

The generated `uint128_from_u128` function accepts `runtime.Value`. Each
primitive shift/mask uses `__able_binary_op`, each cast emits `ast.Ty("u64")`,
and the mask literal constructs a new `big.Int`. Escape analysis confirms the
`runtime.Value` parameter leaks to the runtime operator, the type AST and
identifier nodes escape through `__able_cast`, and the mask BigInt escapes
through the binary operation. CPU attribution agrees: allocation consumes
52.88% cumulatively, interpreter `evaluateBitwise` 36.59%, and bridge
`AsUint` 9.40%.

This is a primitive-carrier limitation beneath ordinary stdlib nominal code,
not evidence for a `UInt128` special case.

## Verification and cleanup

- Three compiled build launches passed their external verifiers.
- Six independent profiling launches reproduced the verified output hashes.
- All CPU and allocation processes completed under 60 seconds.
- All three generated modules passed Go escape-analysis builds.
- No production, benchmark, or `able-stdlib` source was changed.
- The temporary 996 MiB generated/profile workspace was removed.

## Next recommendation

Audit a native compiled carrier for the primitive Able `i128` and `u128` types
across Fixed Width 128 plus the independent `int128_accumulate_small` and
`uint128_accumulate_small` benchmark fixtures.

Why: compiler carrier selection currently maps both primitive types directly
to `runtime.Value`, while all <=64-bit integer primitives use native Go scalar
carriers. This forces otherwise-static 128-bit source through dynamic boxing,
AST type construction, bridge conversion, `math/big`, and interpreter
operations. The project rules explicitly permit primitive-specific Go lowering
but prohibit naming the stdlib `Int128`/`UInt128` structs, so the primitive
boundary is the only legitimate shared candidate exposed by this tranche.

What it entails: first census every generated `i128`/`u128` operation and
conversion in the three benchmark programs and the numeric correctness corpus.
If both signed and unsigned programs repeat the same dynamic-carrier wall,
prototype compiler-owned two-word primitive carriers with direct checked
arithmetic, comparison, bitwise, shift, and cast semantics. Boxing to
`runtime.IntegerValue` must remain only at explicit dynamic/host boundaries.
The candidate must preserve overflow, signed ordering, cast, literal, extern,
interface, and interpreter parity semantics, and must pass repeated alternating
compiled averages for both fixture families and Fixed Width plus unrelated
small-integer, Rational, Binary Trees, and startup guards. Do not add a nominal
type, method name, benchmark, or stdlib-package branch.
