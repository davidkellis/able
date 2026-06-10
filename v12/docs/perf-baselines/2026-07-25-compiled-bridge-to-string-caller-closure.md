# Compiled bridge.ToString caller closure

Date: 2026-07-25

## Decision

Retain no production code.

The exact `bridge.ToString` allocation leaf repeats in Concurrent Event
Routing, Word Frequency, and Sensor Calibration, but its semantic parents do
not. Event and Sensor primarily materialize successful native pattern-
assignment results into `runtime.Value`; Word boxes String keys for the
runtime-backed generic Map store; Event also converts a channel payload at an
explicit runtime-service boundary.

No single general nominal/interface carrier rule reaches all three
applications. A pattern-result liveness correction would miss Word entirely,
a runtime Map key correction would miss Event and Sensor and may not be
implemented as a named HashMap/String special case, and the channel conversion
is required by the current scheduler payload ABI.

No compiler, generated runtime, canonical stdlib, interpreter, tree-walker,
bytecode VM, language, dependency, or WASM change was made.

Machine-readable evidence is in
`2026-07-25-compiled-bridge-to-string-caller-closure.json`.

## Exact caller attribution

The immediately preceding retained binaries and their three-sample exact
allocation profiles were used unchanged. `pprof` caller attribution for one
representative exact sample is:

| Application | `bridge.ToString` objects | Exact direct callers |
| --- | ---: | --- |
| Concurrent Event Routing | 38,274 | `__able_array_String_to`: 18,816 (49.16%); `__able_struct_EventRecord_to_seen`: 15,360 (40.13%); `__able_struct_EventTask_to_seen`: 4,096 (10.70%); two diagnostic/setup objects |
| Word Frequency | 35,959 | `HashMap.raw_get`: 17,979 (50.00%); `HashMap.raw_set`: 17,979 (50.00%); one diagnostic/setup object |
| Sensor Calibration | 48,898 | `__able_array_String_to`: 48,896 (effectively 100%); two diagnostic/setup objects |

### Event and Sensor: successful pattern-result materialization

The Array and `EventRecord` parents occur after native destructuring has
already checked shape and assigned concrete Go fields. `compilePatternAssignment`
then converts the entire successful right-hand value to `runtime.Value` so the
assignment expression can return its specified success value. In these two
applications the parent immediately checks for a mismatch error and discards
the successful value.

Avoiding that work requires a general liveness-aware pattern-expression result
rule. Such a rule is independently promising, but it has no occurrence under
Word Frequency's exact `bridge.ToString` leaf and therefore fails this
tranche's three-unlike-program admission bar.

### Word: runtime-backed generic Map keys

Word Frequency's calls are evenly split between Map lookup and insertion.
The generated specialization has a native Go `string` key, but the shared Map
store preserves keys as `runtime.Value` so it can apply Able Hash/Eq semantics
and store arbitrary nominal keys. Inserted keys escape into the store.

Removing this conversion requires a general typed generic-Map storage design.
A `HashMap String` or other named-container compiler branch is prohibited, and
no such storage redesign has a shared owner in Event or Sensor.

### Event: channel payload conversion

`EventTask` conversion occurs when sending a nominal value through the shared
runtime scheduler channel. The payload must remain available after the send
returns, so its String field cannot borrow ephemeral conversion storage.
Changing this route would require a broad typed scheduler/channel payload ABI,
which is outside this tranche and is not shared by Word or Sensor.

### Why `bridge.ToString` itself was not changed

`runtime.Value` is an interface and `runtime.StringValue` contains a Go string
header. Values stored in runtime Arrays, structs, Maps, and channels must own a
stable interface payload. Replacing the conversion locally with another
pointer wrapper does not remove the required escaping object. Global
interning would add retention and synchronization semantics, while changing
`runtime.Value` to a tagged carrier would be a broad runtime ABI redesign.

The shared leaf name therefore does not identify one shared removable
boundary.

## Existing repeated measurements

No candidate was admitted, so manufacturing a baseline/candidate comparison
would be misleading. The frozen binaries are exactly those measured in the
immediately preceding twenty-cohort record:

| Application | Current Able mean | Go mean | Able / Go |
| --- | ---: | ---: | ---: |
| Concurrent Event Routing | 104.598 ms | 3.077 ms | 33.990x |
| Word Frequency | 18.661 ms | 3.765 ms | 4.957x |
| Sensor Calibration | 25.491 ms | 3.259 ms | 7.821x |

That record also contains ten CPU profiles and three exact allocation profiles
per application. This closure adds exact caller localization; it does not
reinterpret or replace those repeated measurements.

## Verification

- All three unchanged public benchmark verifiers pass.
- All three unchanged strict dependency graphs omit `pkg/interpreter`.
- Candidate binary SHA-256 values match the preceding retained record.
- `go test ./cmd/ablec -count=1 -timeout 60s` passes in 5.978 seconds.
- The worktree's existing production changes and untracked files were
  preserved.

## Next

Add a guarded canonical primitive `String.split` native lowering candidate
that operates on proven Go String carriers and returns a fresh native
`Array String`.

This is next because the same semantic pipeline now repeats in all three
applications: `String.split` owns substantial cumulative CPU and
`slice_bytes` allocates 33,101 Event, 33,684 Word, and 63,991 Sensor objects.
The work entails matching the canonical method/signature only, using native Go
splitting for valid String/delimiter values, preserving empty-delimiter
codepoint behavior and invalid UTF-8 errors, returning a fresh Array carrier,
and retaining the Able implementation for exceptional/dynamic cases. It must
pass generated-source, Unicode/error, ownership, strict-dependency, exact-
allocation, and repeated baseline/candidate/Go gates.

This is important because it can remove the shared intermediate mutable
`Array u8` slicing and decode loop while keeping primitive Strings and static
Arrays on their native Go carriers. It is a primitive String rule, not a
named-container or non-primitive nominal special case.

Do not begin WASM work.
