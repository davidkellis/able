# Bytecode three-shape post-compiler refresh

Date: 2026-07-22

## Decision

Keep no bytecode VM, compiler, runtime, canonical-stdlib, benchmark, fixture,
language, or WASM change. Fresh post-change profiles for text split/join,
iterator collection, and numeric Array mapping again share dispatcher parents
and several generic helper names, but not one new concrete semantic descendant.
Every material apparent intersection is either split among unlike callers or
belongs to a representation already rejected by repeated broad guards.

This refresh therefore closes the current handoff without manufacturing a
candidate. In particular, it does not retry raw-integer carrier changes,
integer-metadata switches, call-name/frame changes, return-guard reordering,
typed-pattern duplicate-precheck removal, or generic map replacement.

## Protocol

- Used canonical external `../able-stdlib/src`, `GOMAXPROCS=1`, `GOGC=50`,
  `GOMEMLIMIT=1GiB`, `--source-root-only`, and a 55-second cap per process.
- The retained benchmark harness loaded and lowered once, warmed `main()` once,
  forced GC, and measured only subsequent calls.
- Five independent processes per workload retain every workstation sample.
  Profiles were captured in separate clean processes and are attribution-only.
- Direct bytecode runs produced `191484`, `382455000`, and `1097192358`, the
  established outputs for split/join, iterator collect, and Array map.
- Cleanup-eligible raw records are under
  `v12/.profiles/20260722_three_shape_refresh/`; the checked-in JSON sibling
  contains the durable aggregate.

## Repeated warmed measurements

All 15 timing processes and all three profile processes completed without a
failure or timeout.

| Workload | Calls/process | Mean ns/op | Median ns/op | Span ns/op | CV | B/op | allocs/op |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| String split/join | 5 | 1,049,314,877 | 1,029,842,194 | 1,009,352,657-1,117,997,485 | 4.13% | 51,316,321 | 579,839 |
| Linked-list iterator collect | 20 | 430,248,930 | 430,671,761 | 414,590,604-448,746,065 | 3.05% | 8,363,839 | 192,559 |
| Numeric Array map | 75 | 77,668,204 | 76,242,584 | 72,652,783-84,721,969 | 6.72% | 805,804 | 119 |

The separate profiles used 8, 25, and 120 measured calls and collected 8.43,
11.04, and 9.34 seconds of CPU samples respectively.

## Concrete descendant reconciliation

| Helper/family | Split/join | Iterator collect | Array map | Reconciliation |
| --- | ---: | ---: | ---: | --- |
| `bytecodeRawIntegerValueInfo` flat | 1.19% | 4.98% | 7.17% | Text is typed-pattern/direct-value work, iterator is cast/immediate comparison, and numeric is cast/u8/integer transport. The generic carrier and raw-store representations already failed broad guards. |
| `runtime.mapaccess2_faststr` cumulative | 10.08% | 6.07% | 4.39% | Callers divide among primitive integer metadata, environment lookup, known-type caches, canonicalization, and nominal matching. The only fixed common table, integer metadata, already failed both full-switch and membership-switch gates. |
| `finishInlineReturn` cumulative | 17.20% | 6.61% | 9.64% | Text is dominated by return coercion/type matching; iterator mixes coercion, result append, and cleanup; numeric is materialization plus frame pop. Prior generic frame/result and guard-order candidates remain closed. |
| `execJumpIfNotTypedPattern` cumulative | 15.66% | 3.35% | 5.25% | Text enters full nominal/`Error` matching, iterator mixes generic/exact protocol paths, and numeric is mostly a no-runtime-value decision. The earlier category census admitted the shared `IteratorEnd` rule, which is already retained; the other categories did not repeat. |
| `matchesTypeWithoutRuntimeValue` under the typed-pattern jump | 2.25% | 2.26% | 3.64% | The same helper name does not imply the same type category. Removing only its duplicate precheck was previously neutral/regressive, and caching a type-only result would need invalidation for dynamic definitions while saving different semantic queries. |

Raw integer extraction is the largest exact flat leaf in two controls, but the
third control is both smaller and dominated by a different typed-pattern
caller. String-map access is broader only as a Go implementation primitive.
Return and typed-pattern execution remain aggregate VM families whose material
children diverge. None clears the three-unlike-program admission bar with a new
mechanism.

## Verification

Focused runtime-harness, typed-pattern, return, and raw-integer tests pass:

```text
go test ./pkg/interpreter -run 'TestLoadBytecodeProgramRuntimeBenchConfig|TestBytecodeMatch|TestBytecodeVM.*TypedPattern|TestBytecodeVM.*Return|TestBytecodeRawInteger' -count=1
ok able/interpreter-go/pkg/interpreter 0.701s
```

## Next recommendation

Profile the bytecode generic-union/`Result` cohort already established by
Binary Event Log, Option/Result Configuration, Manifest Normalization, and
Policy Record Dispatch, using these three shapes as unrelated controls.

Why: these are four unlike verifier-backed applications with a language-level
feature shared across them, while the current text/iterator/numeric screen has
again exhausted its local helper families. The recently retained compiler
direct-method rule does not improve bytecode execution, so the same cohort can
test whether bytecode has an independent, concrete union construction,
pattern, method-dispatch, or result-transport leaf.

What it entails: collect five independent verifier-backed warmed bytecode
means plus bounded CPU/allocation profiles; classify exact generic-union
construction, typed-pattern, resolved-call, and return descendants; and build
a candidate only if one generic semantic rule is material in at least three of
the four applications, spans at least two distinct union definitions, names no
nominal type, and remains neutral on split/join, iterator collect, numeric Array
map, and unrelated scorecard controls. Continue to defer WASM.
