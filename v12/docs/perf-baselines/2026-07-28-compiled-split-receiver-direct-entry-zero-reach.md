# Compiled split-receiver direct-entry zero-reach closure

Date: 2026-07-28

## Goal

Determine whether the hot context-aware native-bound lookups preserve an
existing compiler-generated
`direct(runtime, environment, receiver, explicitArgs)` method entry. Advance
a compiler-owned split-receiver bound carrier only if that entry is present
across the four reached applications.

## Method

The retained callable-dispatch diagnostic now distinguishes successful
lookups whose receiver type and method resolve to a non-nil compiled direct
entry from lookups without one. Future Await Race, Await Channel Mux, Mutex
Await Journal, and Mutex Work Queue each ran in five fresh processes with
their catalog executor, logical CPU budget, canonical stdlib, strict
fallback-free compiler mode, and public verifier.

Generated source was also inspected across all emitted Go files, including
the split package registrars and method implementations. This avoids the
incorrect conclusion that direct entries do not exist merely because their
registrations are emitted outside `compiled.go`.

## Result

| Application | Context calls | Native-bound calls | Successful lookups | Direct entries |
| --- | ---: | ---: | ---: | ---: |
| Future Await Race | 912.6 | 912.6 | 912.6 | 0.0 |
| Await Channel Mux | 9,216.0 | 8,704.0 | 8,704.0 | 0.0 |
| Mutex Await Journal | 8,559.8 | 6,511.8 | 6,511.8 | 0.0 |
| Mutex Work Queue | 17,073.4 | 12,977.4 | 12,977.4 | 0.0 |

The compiler does generate and register general direct entries for ordinary
compiled methods. None is carried by this hot branch. The reached values are
instead:

- built-in Future methods, whose registry entries have no direct function;
- generated Awaitable methods stored as closure-backed
  `runtime.NativeBoundMethodValue` fields;
- generated AwaitRegistration and AwaitWaker closure-backed methods.

The Awaitable closures already own the actual service receiver. Their
`NativeBoundMethodValue.Receiver` exists to represent method binding, while
the native implementation either ignores the injected receiver or selects
the explicit argument from the end of the rebuilt slice.

All 20 executions passed their public verifier. All four dependency graphs
omit `pkg/interpreter`.

## Decision

Close the proposed compiled-entry carrier at zero reach. A carrier preserving
`entry.direct` would execute zero times in all four applications, so no
prototype, allocation experiment, or balanced Able/Go timing is admissible.
No compiler/runtime execution, interpreter, bytecode VM, canonical stdlib,
language, dependency, named-container/non-primitive nominal, or WASM change
was made.

The only retained code change is the reusable diagnostic counter extension.
The disk-backed measurement workspace and generated Python cache are removed
after the aggregate evidence is retained.

Machine-readable aggregate:

- `2026-07-28-compiled-split-receiver-direct-entry-zero-reach.json`

## Next

Prototype a diagnostic-only receiver-free representation for generated
closure-owned native kernel methods under the existing experimental callable
gate.

Why: these Future/Awaitable closure methods, not ordinary compiled method
entries, are the actual receiver-injection owner shared by all four programs.

What it entails: overlay the generated Awaitable, AwaitRegistration,
AwaitWaker, and applicable Future callable construction so the closure remains
a native callable without an unused bound receiver; guard direct, captured,
value/pointer, interface, nested, error, and dynamic compatibility behavior;
then run repeated allocation and balanced baseline/candidate/Go measurements
across the four reached applications plus unrelated serial controls.

Why it is important: it targets the measured common boundary directly and can
remove receiver boxing/slice construction at a language kernel boundary
without inventing a named nominal lowering rule or re-entering the
interpreter.
