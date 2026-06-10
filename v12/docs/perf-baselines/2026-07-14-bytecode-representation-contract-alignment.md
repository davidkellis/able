# Bytecode representation-contract alignment — 2026-07-14

## Decision

Keep the shared representation corrections and test-contract updates below.
They make tree-walker and bytecode results agree at public value boundaries;
they are not a performance candidate and were not benchmarked as one.

## Changes

- The tree-walker's simple named-literal path now converts a canonical
  `Array { length, capacity, storage_handle }` literal into `ArrayValue`, just
  as the regular tree path and bytecode literal path already do. It no longer
  leaks a generic `StructInstanceValue` merely because all fields were named.
- A no-coercion return materializes ordinary raw integer carriers before the
  public value boundary. VM-owned integer return scratch remains raw only while
  it is immediately copied into the caller's internal stack lane, preserving
  the existing allocation-free nested-return path.
- Tests now describe the current generic lowering contract: an unresolved
  top-level identifier receiver is a static-member *candidate*. Runtime
  resolution still handles ordinary methods and callable fields, while safe
  navigation remains a normal member call. The tests no longer require an
  Array- or iterator-specific opcode for an unknown receiver.
- The current canonical `Int128` / `UInt128` constructors are source methods,
  so the trace regression asserts only the arithmetic and checked-conversion
  operations that are actual VM fast paths.
- An Array is an aggregate escape boundary. Tests now require a raw VM scalar
  to materialize when appended to tracked Array storage, and require metadata,
  length, and capacity operations to preserve that materialized state.
- The pending-await cancellation test uses a 250 ms timer window rather than a
  15 ms scheduling race. It still waits past the timer deadline and verifies
  that cancellation prevents the callback.

No file in the external `able-stdlib` checkout changed.

## Focused verification

All commands used the project-local Go cache and completed successfully:

```text
go test ./pkg/interpreter -run '^TestArrayWithCapacityFromSource' -count=1 -timeout 60s
go test ./pkg/interpreter -run '^(TestBytecodeVM_ArrayStructLiteralNamedFastReturnsArrayValue|TestToArrayValueReadsPositionalArrayStructMetadata)$' -count=1 -timeout 60s
go test ./pkg/interpreter -run '^(TestBytecodeVMProgramReturnNoCoercionFastPath|TestBytecodeVM_.*Return)' -count=1 -timeout 60s
go test ./pkg/interpreter -run '^(TestBytecodeVM_LoweringUsesStaticCandidateForUnknownArrayGetReceiver|TestBytecodeVM_LoweringUsesStaticCandidatesForUnknownNewReceivers|TestBytecodeVM_CallMemberOpcodeExecutesMethodCall|TestBytecodeVM_CallMemberFallsBackToCallableField|TestBytecodeVM_LoweringUsesStaticCandidateForUnknownIteratorNext)$' -count=1 -timeout 60s
go test ./pkg/interpreter -run '^TestBytecodeVM_CanonicalStdlibNumericStructMethodsUseFastPathsFromSource$' -count=1 -timeout 60s
go test ./pkg/interpreter -run '^(TestInterpreterEnsureArrayStateMaterializesTrackedRawValues|TestInterpreterEnsureArrayStateForMetadataPreservesMaterializedValues|TestBytecodeVMArrayLengthAssignmentPreservesMaterializedValues|TestInterpreterMemberAssignArrayCapacityPreservesMaterializedValues)$' -count=1 -timeout 60s
go test ./pkg/interpreter -run '^TestAwaitCancellationStopsPendingArms$' -count=20 -timeout 60s
go test ./pkg/interpreter -run '^TestExternStructArrayFieldCoercesIntoHostMap$' -count=100 -timeout 60s
```

The last two stress checks passed in 5.677 s and 0.155 s respectively.

## Go extern-host loader follow-up

The apparent package-level memory failure was reproducible as a Go-plugin
loader failure instead. A fresh process passed each extern integration test in
isolation, but loading the independent prewarm, corrupt-artifact, and struct
coercion plugins in one process ended in `SIGBUS` in `dlopen` at the third
distinct plugin. It persisted with a fresh cache, `GOGC=off`, a large Go memory
limit, and both the Go build directory and extern cache moved from `/tmp` to
disk-backed project storage. The artifacts totalled only about 18 MiB, so this
is not an Able allocation or cache-size regression.

The generated module is now built as its module directory (`go build ... .`),
rather than as a standalone source file. That gives it the unique module
identity declared in its content-addressed `go.mod`; the cache version is now
`v4` so incompatible anonymous artifacts cannot be reused. This is a general
host-module correctness repair, although it alone cannot overcome the Go
runtime loader failure on this host.

Go plugins are process-global and cannot be unloaded. The integration tests
therefore run each plugin-loading case in a fresh test process, while direct
fast-invoker tests exercise the shared adapter with typed Go functions instead
of compiling redundant plugins. Real generated-symbol, cache recovery, host
coercion, and result-conversion integration coverage remains present. The
reduced extern suite passes in 2.204 s under a fresh disk-backed cache.

This test isolation is deliberately not an application-runtime workaround.
The Go `plugin` package itself documents dynamic plugins as a fragile boundary
and recommends a statically built executable or IPC when independently built
modules must grow at runtime. Program evaluation and `able cache prewarm` now
register the complete dependency-ordered Go extern set first and build one
content-addressed host image: each Able package compiles as a separate Go
subpackage, while the image exports package-qualified wrappers. That preserves
same-named extern functions across packages without opening one plugin per
package. The two-package regression verifies both values execute through the
same plugin. Definitions introduced dynamically after program preparation keep
the old per-package fallback; moving that unrestricted dynamic boundary behind
typed IPC remains future work. Do not add package- or benchmark-specific
exemptions.

The serial executor also starts its worker lazily on first queued task. New
interpreter instances that never run a future no longer leave idle goroutines
behind; executor behavior is covered by a first-task lifecycle test. This was
safe resource hygiene but did not resolve the plugin loader fault.

## Full-suite status

A bounded complete package verification run with
`GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1` passed the old 184 s and 192 s failure
points and completed. A normal `go test ./pkg/interpreter -timeout 300s`
rerun also completed, followed by the cached verification:

```text
ok   able/interpreter-go/pkg/interpreter (cached)
```

The isolated semantic checks listed above remain green.

## Next recommendation

Refresh the verifier-backed application scorecard with a prewarmed host image.
Why: normal loaded applications no longer need one dynamic plugin per extern
package, so the previous scorecard cold-start boundary is obsolete. The work
entails output-checked runs for unlike programs plus an explicit guard for the
dynamic-definition fallback. Select a VM/compiler candidate only if a concrete
leaf repeats across those refreshed workloads; this host correction provides no
license for a micro-optimization.
