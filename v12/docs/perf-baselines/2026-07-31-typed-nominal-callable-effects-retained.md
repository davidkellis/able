# Typed nominal callable effects retained

Date: 2026-07-31

## Decision

**Retain the opt-in typed effect analysis and no production lowering change.**

The compiler can now compute a conservative fixed point over top-level
functions, methods, local functions, and statically resolved lambdas. For each
nominal parameter it records:

- reachable field or index mutation;
- capture or storage;
- return aliasing; and
- an unknown or dynamic call.

Direct and monomorphic callable edges propagate these facts transitively.
Concrete nominal parameter types also propagate through those edges, which
lets unannotated nested lambdas participate without introducing a global or
name-based resolver. Ambiguous rebinding, interface dispatch, and unresolved
calls remain fail-closed. Whole-reference aliases through typed expressions
and storage inside struct, Array, or Map literals are also tracked; ordinary
field-value reads do not falsely escape the containing record.

Array element storage is classified as capture at the language-kernel
`__able_array_write` boundary. This is a semantic kernel rule, not an Array or
other named-container lowering fast path.

The report is diagnostic only. Ordinary compilation does not run it, and the
result does not select carriers or alter generated Go.

## Full strict census

All 66 selected compiled applications generated with `--no-fallbacks`.

The opt-in census reported:

| Measure | Total |
| --- | ---: |
| Callables | 34,885 |
| Parameters | 55,814 |
| Nominal parameters | 23,281 |
| Read-only/non-escaping nominal parameters | 11,292 |
| Mutation parameters | 3,586 |
| Capture parameters | 460 |
| Return-alias parameters | 1,337 |
| Unknown-call parameters | 8,303 |

These totals include imported canonical-stdlib callables once per application;
they are coverage totals, not unique project-wide declarations.

The prior generated feasibility screen had 48 constructed nominal rows in 23
applications blocked by an unknown mutation-capable call. Joining those rows
to the typed report produced:

| Typed disposition | Rows | Applications |
| --- | ---: | ---: |
| Every instance read-only/non-escaping | 20 | 14 |
| Precisely resolved as capture/storage | 2 | 2 |
| Typed unknown remains | 21 | 14 |
| Generated helper name not directly matched | 5 | 5 |

Thus 22 of the 48 formerly opaque rows now have a precise safe-or-capture
answer.

Five rows are newly clear of every generated feasibility blocker:

- Binary Event Log `EventRecord`;
- Concurrent Event Routing `EventRecord`;
- Concurrent Graph Visitors `VisitState`;
- Concurrent Packet Codecs `CursorState`; and
- Concurrent Tree Folds `FoldState`.

The result is broad enough to justify the next admission audit, but it does not
by itself prove that a Go value carrier preserves Able identity, lifetime, and
storage semantics.

For the two profiled records:

- Binary Event Log `EventRecord` has four nominal parameter instances, all
  four read-only/non-escaping, with no mutation, capture, return alias, or
  unknown call.
- Versioned Telemetry `Sample` has six instances: two read-only scorer
  parameters and four precise Array-storage captures, with no mutation,
  return alias, or unknown call.

The Telemetry result is especially useful: the old unknown-call blocker is
gone, while the real storage lifetime remains visible and fail-closed.

## Generated-output control

The full census was repeated with the same current compiler and harness but
with effect collection disabled. Both runs were 66 successful / 0 failed.

Across enabled and disabled runs:

- all 66 generated module hashes match exactly;
- all 66 generated boundary analyses match exactly;
- all aggregate native-boundary categories match exactly; and
- all semantic-parent boundary maps match exactly.

The raw enabled and disabled census SHA-256 values are respectively
`58857f3652df39eda3750de286501c50e7447bf9374b16b4c368fdcfc8ba502d`
and
`a4bf9d1c524fb7a8e6fc580930aa405159864457ff0c5efa6b5772e1987749b3`.
The large raw reports were disposable and are summarized here rather than
retained in the repository.

Binary Event Log and Versioned Telemetry strict binaries also passed their
public Ruby verifiers, and both dependency graphs omit
`able/interpreter-go/pkg/interpreter`. Binary's top-level generated Go SHA-256
was identical with collection enabled and disabled:
`4198279afdaa61da80ef56a2180e128b99fdb89cbf8497f9c6b2a43338e426da`.

No repeated runtime A/B cohort was run because the opt-in diagnostic is
executed after rendering and the complete generated-output control proves
that executable code is unchanged. Timing such identical binaries would only
measure workstation noise.

## Verification

Passed:

- direct and transitive read-only effects;
- direct and transitive mutation;
- unannotated monomorphic nested-lambda type propagation;
- closure capture and return aliasing;
- conditional whole-reference aliases and aggregate-literal storage;
- unknown function-typed parameter calls;
- Array kernel storage capture;
- ambiguous local callable and interface-dispatch negatives;
- opt-in/report isolation from generated Go;
- existing structural alias, executable alias/escape, retained-old-result,
  callee-capture, and conditional-candidate guards;
- `go test ./cmd/able-generated-boundary-census ./cmd/ablec`; and
- strict executable verification for Binary Event Log and Versioned
  Telemetry.

The focused effect suite completed in about 1.1 seconds. The combined
caller-owned/effect guard set completed in 2.889 seconds. The command packages
completed in 0.002 and 5.032 seconds.

`go test ./pkg/compiler -short -count=1 -timeout=60s` reached the aggregate
package timeout while an existing audit row was starting. The named row,
`TestCompilerPersistentSortedQueueMethodsStayNative`, passed independently in
2.347 seconds. No individual observed test exceeded one minute.

After retaining the compact evidence, the two exact disk-backed task
workspaces totaling 1,761,224 KiB were removed. No `/tmp/able-*` or other
`/var/tmp/able-*` workspace remains.

## Scope and exclusions

No generated runtime, runtime package, interpreter, bytecode VM, parser, AST,
language spec, dependency, canonical stdlib, benchmark, frozen workspace, or
WASM change was required. No carrier rule, named-container rule,
non-primitive nominal special case, execution-context ABI change, or
benchmark-specific branch was introduced.

The compiler-source addition invalidated the 11 compiled performance closures
and the cross-family architecture closure through
`scope-content-drift:compiler-production`. The exact 66-module equality above
is the reconciliation authority: the reviewed closures can be atomically
rebaselined without changing any timing disposition or admitting a production
candidate.

## Next

Join the typed callable facts to exact generated boundary and storage sites
for the five newly clear rows, then admit a prototype only if one
representation/lifetime mechanism repeats materially in at least three unlike
applications.

Why: callable non-mutation is now proven, but a native Go value carrier must
also preserve observable identity, return ownership, union/interface
encoding, and storage lifetime.

What it entails: expose a stable diagnostic link from generated call sites to
their callable summary; inventory conversion, writeback, allocation, and
storage descendants for the five rows; select one shared owner; then, only if
the three-application gate passes, implement a general rule with semantic
guards and five-or-more balanced public-verifier baseline/candidate/Go pairs.
If no shared owner passes, retain no code and document the no-go.

Why it matters: this is the shortest sound route from the new proof to less
boxing and allocation while keeping compiled Able entirely on native Go
carriers and preserving Able reference semantics.
