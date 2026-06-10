# Bytecode map-lookup attribution — 2026-07-11

## Scope

This diagnostic tranche investigated repeated `runtime.mapaccess2_faststr`
samples after the raw-integer carrier candidate was rejected. The question was
whether one interpreter-owned map and one invalidation rule recurred across the
text, iterator, and numeric controls strongly enough to justify a generic VM
lookup optimization.

All profile and benchmark commands used the existing one-process limits:
`GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`.

## Attribution

Fresh bounded CPU profiles show `runtime.mapaccess2_faststr` in each workload,
but not behind a single cache:

| Control | Direct map-access samples | Share | Main interpreter callers |
| --- | ---: | ---: | --- |
| `string_split_join_small` | 0.50 s / 5.39 s | 9.28% | `matchesTypeWithoutRuntimeValue`, `matchesType`, `tryFastSimpleTypeCoercion`, `lookupKnownTypeNameCache`, function invocation |
| `linked_list_iterator_collect_i64_small` | 0.06 s / 1.25 s | 4.80% | raw-integer matching/coercion, `lookupKnownTypeNameCache`, invocation, type matching |
| `array_map_i32_small` | 0.10 s / 1.40 s | 7.14% | `matchesTypeWithoutRuntimeValue`, `lookupKnownTypeNameCache`, `bytecodeSimpleIntegerTargetKind` |

The common fixed map is `integerInfos`, keyed by the 12 language-defined
integer suffixes. The other contributors have distinct semantics: the known
type-name cache is interpreter and package-registry state, while the interface
and union maps participate in nominal type matching. The existing bytecode
identifier, call-name, and member-method paths already use site-local direct
or hot caches and were not the repeated profile owner here.

## Candidate and decision

I temporarily replaced `lookupIntegerInfo`'s string-keyed map lookup with a
closed `switch` returning precomputed integer descriptors. It was a generic
primitive-language change, not a source-program or container special case.

The candidate produced the same allocation shapes, but did not establish a
broad wall-time win:

| Control | Candidate samples | Allocation shape |
| --- | --- | --- |
| split/join (5x) | 1.17–1.56 s/op | 48.7–49.8 MB/op; 549.8k–550.0k allocs/op |
| iterator collect (5x) | 305–314 ms/op | 3.445 MB/op; 29,346–29,348 allocs/op |
| numeric Array map (20x) | 67.6–90.0 ms/op | about 0.910 MB/op; 443–457 allocs/op |

The immediate restored run had substantial wall-clock variation in the long
text control (1.19–2.27 s/op), while collect was comparable or faster after
restore (246–336 ms/op). With no allocation reduction and no stable advantage
across the controls, the switch was removed. It must not be reintroduced as a
microbenchmark-led replacement.

## Outcome

No runtime or compiler code remains from this tranche. The results narrow the
next investigation to the shared *type-match* execution paths rather than a
general map implementation change. Any future candidate must first establish
which type-expression/value categories repeat across the suite and preserve
the nominal type rules for interfaces, unions, and user-defined structures.
