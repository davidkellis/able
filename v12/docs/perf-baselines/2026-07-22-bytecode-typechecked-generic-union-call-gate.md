# Bytecode typechecked generic-union call gate

Date: 2026-07-22

## Decision

Retain the typechecker-proven generic-union member-call opcode and its
instruction-local cache.

The typechecker now preserves the declaration provenance of each successful
member-method selection. Bytecode lowering emits `CallGenericUnionMember` only
when that provenance says the selected declaration is a generic method set on
a named union. The rule uses language semantics; it names no union, container,
stdlib API, application, or benchmark.

This separation fixes the failure mode of the preceding rejected candidate.
The ordinary `CallMember` path is unchanged, and iterator collect executes
zero specialized instructions. Three unlike owners improve materially,
Binary Event Log is neutral, and all three unrelated controls are neutral or
faster. The candidate therefore passes the broad admission bar.

## Design

`CheckResult` now carries a per-package method-selection map alongside inferred
types. A selection records whether a method set or implementation supplied the
member, the selected declaration, its checked target, and whether it is a
generic named-union method set. Facts from packages with diagnostics are
discarded before lowering, matching the existing inference-fact policy.

The dedicated bytecode instruction:

- is emitted only for a typechecker-proven member call;
- preserves safe-member-call nil behavior and callable-field shadow fallback;
- resolves the generic-union overload set on the first execution;
- keeps a VM-local 16-set, two-way instruction cache; and
- invalidates on program/instruction/member mismatch, method-cache changes,
  lexical binding-shape changes, name-specific binding revisions,
  implementation-context changes, and defining-owner revision changes.

Overload selection for explicit arguments remains in the existing cached
member-call dispatcher. The ordinary runtime member lookup and ordinary
member-cache hit path receive no extra branch or table probe.

## Opcode census

Stats-enabled one-call measurements confirm the semantic selection boundary:

| Workload | Proven opcode executions | Ordinary `CallMember` | Cache hits | Cache misses |
| --- | ---: | ---: | ---: | ---: |
| Binary Event Log | 53,248 | 53,271 | not captured | not captured |
| Option/Result Config | 98,304 | 0 | 98,300 | 4 |
| Manifest Normalization | 13,312 | 21,504 | not captured | not captured |
| Iterator collect | 0 | 8,024 | 0 | 0 |

Binary and Manifest census binaries predated the final two counter fields, so
their opcode counts—not inferred counter values—are reported. Policy timing is
included below; its stats-enabled run exceeded the bounded instrumentation
window and was not used as a census claim.

## Repeated gate

Candidate and compiled-out control were fixed binaries built from identical
source except for the proof predicate. Every process performed one untimed
warmup and one measured `main()` call. Runs alternated order under
`GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, source-root-only loading, normal
typechecking, and a 59-second process cap.

Owner workloads used three independent processes per variant:

| Workload | Control ns/op | Candidate ns/op | Time | Bytes | Objects |
| --- | ---: | ---: | ---: | ---: | ---: |
| Binary Event Log | 6,969,838,432 | 6,955,794,582 | -0.20% | -0.0001% | -0.00004% |
| Option/Result Config | 778,828,979 | 697,281,655 | -10.47% | -11.07% | -10.51% |
| Manifest Normalization | 1,318,033,362 | 1,264,546,771 | -4.06% | -26.28% | -26.05% |
| Policy Record Dispatch | 6,949,612,819 | 6,676,023,023 | -3.94% | -2.27% | -3.49% |

The unrelated controls used five independent processes per variant because
they are shorter and workstation variance is proportionally larger:

| Workload | Control ns/op | Candidate ns/op | Time | Bytes | Objects |
| --- | ---: | ---: | ---: | ---: | ---: |
| String split/join | 1,043,558,417 | 1,013,183,389 | -2.91% | +0.06% | unchanged |
| Iterator collect | 412,418,334 | 410,401,172 | -0.49% | unchanged | unchanged |
| Numeric Array map | 68,840,005 | 68,074,958 | -1.11% | unchanged | unchanged |

The small split/join byte difference is 32.7 KiB amid workstation variation;
object count is exactly equal. Iterator and Array map have exactly equal byte
and object means. Unlike the rejected runtime-shape candidates, this change
does not select or tax iterator collect.

The candidate binary SHA-256 was
`cff167f6858e781d185a6a904e2c84e08d5d2d0314c04da9adce3e8477641230`;
the compiled-out control was
`6155fd427bc0797e15fd21cfa2fb2d98f00f77bad03276883b8b6233da4ecfc5`.
Every raw sample is retained in the companion JSON report.

## Verification

Focused semantic, invalidation, lowering-boundary, member-cache, and fixture
tests pass:

```text
go test ./pkg/typechecker ./pkg/interpreter -run 'TestGenericNamedUnionMemberAccessRecordsMethodSelection|TestBytecodeGenericUnionCallCacheInvalidatesMemberChanges|TestBytecodeLoweringUsesGenericUnionOpcodeOnlyForProvenCall|TestOptionResultConfigGenericUnionMethodsRemainResolvedWhenWarm|TestBytecodeVM.*Member.*Cache|TestExecFixtures/06_11_truthiness_boolean_context' -count=1 -timeout 60s
ok able/interpreter-go/pkg/typechecker
ok able/interpreter-go/pkg/interpreter
```

The full typechecker package passes. The interpreter package contains more
than one minute even in its broad `A-F` name range and correctly hit the
mandated 60-second aggregate timeout; each test named by the aggregate
timeouts passed alone:

```text
go test ./pkg/typechecker -count=1 -timeout 60s
ok able/interpreter-go/pkg/typechecker 0.026s

go test ./pkg/interpreter -run '^TestExecFixtures/06_12_01_stdlib_string_helpers$' -count=1 -timeout 60s
ok able/interpreter-go/pkg/interpreter 0.189s

go test ./pkg/interpreter -run '^TestExternModuleBuildsFastInvokerForHotStringSignatures$' -count=1 -timeout 60s
ok able/interpreter-go/pkg/interpreter 0.074s

go test ./pkg/interpreter -run '^TestBytecodeTracePrimitiveMemberCallsUseResolvedMethodPath$' -count=1 -timeout 60s
ok able/interpreter-go/pkg/interpreter 0.112s
```

No compiler, canonical-stdlib, benchmark, fixture, language, or substantive
WASM change was needed. WASM received only compatibility stubs for the new
host-side fact type.

## Next recommendation

Refresh bounded post-change CPU and allocation profiles for Option/Result
Config, Manifest Normalization, Policy Record Dispatch, and Binary Event Log,
then admit a candidate only if the same flat VM/runtime leaf recurs in at least
three unlike owners.

Why: this tranche removed the shared generic-union resolution wall from the
semantically proven sites, so the old profiles no longer describe the current
bottleneck. Binary's neutral result also shows that its remaining work is
elsewhere; broadening the proof to runtime union shape would recreate the
iterator regression.

What it entails: capture one clean bounded profile per owner with the same
one-process/1 GiB guardrails, reconcile exact flat descendants rather than
cumulative parents, and test the first genuinely shared leaf against
split/join, iterator collect, and numeric Array map with alternating repeated
processes. Likely areas are call-frame/return handling, raw scalar extraction,
or type matching, but no area should be selected before the refreshed profiles
agree. Continue to defer WASM.
