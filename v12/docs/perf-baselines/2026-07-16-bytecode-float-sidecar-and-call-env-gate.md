# Bytecode Float Sidecar and Call-Environment Gate

Date: 2026-07-16

## Outcome

Complete this tranche with no retained VM, compiler, stdlib, fixture, or
benchmark production change.

The existing raw-float slot sidecar cannot be enabled independently of operand
and call-frame ownership. Sidecar-aware central reads repair ordinary
arithmetic and return correctness, but moving raw floats through reusable
pointer stack cells adds work at the stable inline-call boundary and regresses
the profiled scalar workload. A separate conservative call-environment shortcut
improves Distance Field and RMS Norm, but regresses reduced NBody in an adjacent
five-process control and therefore fails the broad benchmark rule.

## Float-sidecar reconciliation

The first experiment made discarded fused float stores authoritative in the
existing frame-owned sidecar. Adding sidecar fallback reads to
`slotMaterializedValue`, `slotStackValueChecked`,
`appendSlotStackValueChecked`, and `slotStoredValue` cleared the prior
arithmetic and return failures. The remaining focused failures asserted the old
internal representation directly (`vm.slots[n]` must contain a raw carrier),
rather than observable Able behavior.

That repair did not remove allocation, however. A one-process Distance Field
probe remained at 512,052,160 B/op and 38,000,142 allocs/op and measured
6,375,106,645 ns/op. Moving the box from a discarded slot store to the next
operand-stack load only shifts the allocation.

The next experiment added reusable, stack-index-owned raw-float pointer cells,
analogous to existing integer stack cells. It reconciled direct extraction,
materialization, snapshots, casts, type information, array materialization,
and generic slot stores. Distance Field then measured:

| State | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: |
| bounded retained baseline | 6,134,317,804 | 512,083,872 | 38,000,201 |
| float sidecar + stack cells | 6,776,238,184 | 752,054,616 | 40,000,175 |

The candidate regressed time 10.46%, bytes 46.86%, and allocations 5.26%.
Inline Able calls require stable argument values: a reusable pointer carrier
cannot be copied into callee slots without aliasing caller stack storage, while
snapshotting it moves the box to the call boundary. Distance's `hypot` wrapper,
RMS's math wrapper, and NBody's ordinary functions all exercise this boundary.
The sidecar and pointer-cell changes were reverted.

## Call-environment fallback candidate

The refreshed profiles showed one smaller concrete call-setup descendant in
all three programs: `inlineResolvedCallEnvForBindings` consumed 0.98% of
Distance, 1.22% of RMS, and 3.55% of reduced NBody. A conservative frame-layout
proof marked only ordinary non-generic functions with no method set, impl
context, or explicit runtime type-binding use. Proven calls returned the
captured closure without consulting the runtime generic-binding plan or
receiver-shape logic. Generic calls, methods, constrained functions, impl
contexts, and slotless functions retained the existing path.

Focused generic, method, float, and return tests passed. One-process smoke
measurements improved Distance to 5,967,987,547 ns/op and RMS to
5,684,365,159 ns/op, without changing allocation counts. Reduced NBody did not
agree. An adjacent five-process comparison using the same source and build
conditions measured:

| reduced NBody | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: |
| shortcut disabled | 1,712,850,111 | 97,613,565 | 6,161,775 |
| shortcut enabled | 1,775,362,485 | 97,613,574 | 6,161,775 |

The shortcut regressed reduced NBody 3.65%. The proof is semantically sound,
but the code-shape/cache trade is not a broad performance win. It and its tests
were reverted rather than retaining a scalar-benchmark preference.

## Verification after reverts

```text
go test ./pkg/interpreter -run 'TestBytecodeVM_.*Float|TestBytecodeVM_.*Return|TestBytecodeVM_.*Slot.*Restore|TestBytecodeVM_.*CallFrame|TestBytecodeVM_InlineGeneric|TestBytecodeVM_.*Array' -count=1 -timeout 60s
ok  able/interpreter-go/pkg/interpreter
```

## Next recommendation

Design a typed f32/f64 register-frame proof parallel to the retained i32
register frame, rather than making a generic raw pointer carrier flow through
the operand stack.

Why: the current 6.16M-52.0M allocation wall remains real, but this tranche
demonstrates that local carrier substitutions merely move boxing to stable
call arguments. A typed register frame can keep proven primitive-float params
and locals unboxed across ordinary Able inline calls, fused arithmetic, and
returns, while materializing only at dynamic language boundaries. This is a
primitive-wide VM rule and is applicable beyond these benchmarks.

What it entails: extend frame analysis with typechecker-backed float slot and
parameter eligibility; allocate/pool float register frames beside i32 frames;
seed raw float call arguments into callee registers; teach fused float loads,
stores, arithmetic, and simple returns to consume them; preserve boxed values
for closures, dynamic calls, aliases, patterns, containers, interfaces, and
escaping snapshots; add mixed f32/f64, aliasing, recursion, error, generic,
method, and call-frame restoration tests; then repeat Distance, RMS, reduced
NBody, and unlike float guards. Continue to defer WASM.
