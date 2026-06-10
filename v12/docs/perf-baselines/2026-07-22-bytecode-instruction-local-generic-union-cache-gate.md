# Bytecode instruction-local generic-union cache gate

Date: 2026-07-22

## Decision

Keep no interpreter, bytecode VM, runtime, compiler, canonical-stdlib,
benchmark, fixture, language, or WASM change.

The requested stats-only census passed its admission rule. Binary Event Log,
Option/Result Config, and Manifest Normalization contain 151,416 successful
static generic-union method resolutions at 16 bytecode sites. Every one of the
151,400 repeat opportunities retained the same lexical family, member-name
revision, method owner/version, checked receiver type, and callable. Exact
environment-sensitive hits were zero because fresh call environments and
unrelated binding-shape changes occurred on every repetition.

A two-way instruction-local cache consequently removed repeat resolution and
produced large owner wins. It is nevertheless rejected. Every implementation
that discovered or consulted the cache from the ordinary call-member path
added work or changed the hot layout for unrelated union-shaped calls. In the
final order bracket, iterator collect averaged 439,526,755 ns/op across two
five-process candidate blocks versus 428,124,679 ns/op after full restoration
(+2.66%), with unchanged allocations. Split/join and numeric Array map were
neutral or faster, but one repeatable shared-workload regression fails the
broad benchmark bar.

## Census

| Workload | Successful calls | Sites | Stable name-guard repeats | Exact env hits | Binding-resolved calls |
| --- | ---: | ---: | ---: | ---: | ---: |
| Binary Event Log | 118,786 | 4 | 118,782 / 118,782 | 0 | 0 |
| Option/Result Config | 18,292 | 6 | 18,286 / 18,286 | 0 | 0 |
| Manifest Normalization | 14,338 | 6 | 14,332 / 14,332 | 0 | 0 |
| Policy Record Dispatch | 5,632 | 3 | 5,629 / 5,629 | 0 | 0 |
| String split/join | 0 | 0 | n/a | 0 | 0 |
| Iterator collect | 0 | 0 | n/a | 0 | 0 |
| Numeric Array map | 0 | 0 | n/a | 0 | 0 |

The stable guard consisted of bytecode program/instruction identity, lexical
binding-family identity, the member's name-specific binding revision, method
cache version, and the defining owner's value revision. This protects lexical
shadowing and method replacement without depending on the fresh call
environment or its unrelated binding-shape revision.

## Candidate evidence

The best bounded owner measurements show the value available if resolution can
be removed without taxing other member calls:

| Workload | Restored/control ns/op | Candidate ns/op | Time change | Bytes change | Objects change |
| --- | ---: | ---: | ---: | ---: | ---: |
| Binary Event Log | 7,552,184,210 | 4,874,167,893 | -35.46% | -51.13% | -39.81% |
| Option/Result Config | 1,024,979,221 | 746,937,802 | -27.13% | -29.13% | -25.33% |
| Manifest Normalization | 1,674,600,129 | 1,369,269,586 | -18.23% | -27.42% | -27.31% |
| Policy Record Dispatch | 7,252,571,409 | 6,654,542,688 | -8.25% | n/a | n/a |

Binary, Option, and Manifest controls use the adjacent restored measurements
from the preceding type-resolution gate; Policy uses the fresh three-process
compiled-out block and the mean of its two surrounding three-process candidate
blocks. All successful samples are retained.

The unrelated guard was decisive:

| Workload | Candidate ns/op | Restored/compiled-out ns/op | Change | Processes |
| --- | ---: | ---: | ---: | ---: |
| String split/join | 1,034,031,947 | 1,062,532,941 | -2.68% | 5 + 5 |
| Iterator collect | 439,526,755 | 428,124,679 | +2.66% | 10 + 5 |
| Numeric Array map | 77,013,176 | 77,460,721 | -0.58% | 5 + 5 |

Iterator allocations remained effectively identical at 192,559 objects per
operation. The regression therefore has no offsetting memory benefit and is
not explained by changed application work.

## Designs tested and removed

1. A one-way instruction table worked for Binary and Option but thrashed two
   Manifest programs whose hot instruction offsets collided.
2. A two-way table removed that collision and reduced each owner to one or two
   resolutions per site.
3. Checked-receiver eligibility, lazy atomic eligibility, a dedicated opcode,
   and a VM-local eligibility table all retained owner wins but added work to
   unrelated union-shaped member calls.
4. Consulting the cache only after the existing member-method cache missed
   reduced the semantic overlap. Iterator's ordinary member cache already hit
   100,003 times and missed only 21 in the diagnostic run, yet the final code
   still failed the wall-time bracket. It was also removed.

No nominal type, stdlib container, application, or benchmark name was used in
any candidate.

## Verification

- Twenty direct bytecode candidate processes (five each for Binary Event Log,
  Option/Result Config, Manifest Normalization, and Policy Record Dispatch)
  passed their public Ruby verifiers with zero failures or timeouts.
- Focused generic-union, member-cache, and truthiness fixture tests pass after
  full restoration:

```text
go test ./pkg/interpreter -run 'TestOptionResultConfigGenericUnionMethodsRemainResolvedWhenWarm|TestBytecodeVM.*Member.*Cache|TestExecFixtures/06_11_truthiness_boolean_context' -count=1 -timeout 60s
ok able/interpreter-go/pkg/interpreter 3.698s
```

Raw cleanup-eligible census and candidate artifacts remain under
`v12/.profiles/20260722_bytecode_generic_union_census/`,
`v12/.profiles/20260722_bytecode_generic_union_candidate/`, and
`v12/.profiles/20260722_bytecode_generic_union_final/`.

## Next recommendation

Audit whether the typechecker already retains exact per-call method-resolution
provenance for generic named-union member calls; add stats-only provenance if
it does not.

Why: runtime receiver shape is too broad—iterator collect executed 30,004
union-shaped member calls but zero successful static generic-union resolutions.
Only the semantic method selection can distinguish owner sites without adding
work to ordinary member-cache hits.

What it entails: trace checked `MemberAccessExpression`/`FunctionCall` facts
through bytecode lowering, count exactly proven generic-union method sites in
the four owners and unrelated guards, and prototype a dedicated lowering only
if the proof is immutable and general. Any later cache or direct-call opcode
must be emitted only for proven sites, preserve lexical/method invalidation,
name no nominal type, and leave the existing call-member hit instruction path
unchanged. Continue to defer WASM.
