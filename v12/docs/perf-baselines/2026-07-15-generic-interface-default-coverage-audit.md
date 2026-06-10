# Generic interface-default coverage audit — 2026-07-15

## Question

The prior external guard found only Document Audit and Lexical Rollup on the
public lazy `Iterator<T>` default-method route. This audit determines whether
that is a missing portable language-coverage family that needs a third
application, or merely insufficient evidence to select a performance change.

## Inventory

`bench_external_suite_csv coverage` lists 33 portable applications, and
`just bench-catalog-check` confirms all 33 have canonical Able sources and
their cross-language/verifier lanes. The current feature reconciliation
continues to cover every v12 family that has a fair portable Go/Python/Ruby
contract; dynamic modules and user-authored externs intentionally remain
fixture-only.

A source inventory of the relevant public generic operations finds:

| Application | Operation | Classification |
| --- | --- | --- |
| Document Audit | `lazy().filter(...).map(...).collect(...)` | `Iterator<T>` default-method pipeline |
| Lexical Rollup | `lazy().filter(...).map(...).collect(...)` | `Iterator<T>` default-method pipeline |
| Option/Result Configuration | `Result<T>.map(...)` | generic named-union method, not an interface default |
| Channel Rollup | manual `Array` and `Channel` loops | no generic default-method pipeline |

The repaired standard-library operation is the `Iterator<T>.filter` default
method in `../able-stdlib/src/core/iteration.able`; it returns `Iterator<T>`.
The local iterator regression remains the third semantic control for its
generic return metadata, but it is deliberately not presented as an external
application.

## Decision

No portable application is missing. The three-unlike-applications rule is an
admission bar for a performance implementation candidate, not a requirement
to manufacture three application programs for every internal implementation
route. The existing two applications already provide genuine cross-language
coverage of the public iterator/default-method feature, while the full catalog
and fixtures cover all other current v12 feature families.

Do not add an application whose sole purpose is to create a third
`Iterator.filter` profile. That would be a synthetic timing loop, would not
make the existing two pipeline applications more unlike, and would invite an
Iterator-, Array-, or benchmark-shaped optimization. Keep no compiler, VM,
canonical-stdlib, external-benchmark, or fixture behavior change from this
audit.

## Verification

```sh
just bench-catalog-check
cd v12/interpreters/go
env GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1 \
  go test ./pkg/interpreter \
  -run 'TestIteratorDefault(MethodRetainsInterfaceGenericReturn|InterfaceMethodValueCache)$|TestBytecodeLinkedListIterator(Collect|FilterMap)BenchWarmup$' \
  -count=1 -timeout 55s
```

Both pass. The preceding five-run external correctness/performance guard is
recorded in `2026-07-15-interface-default-external-guard-decision.md`.

## Next eligible work

Wait for a material language-wide semantic/compiler/portability change that
is naturally used by at least three unlike verifier-backed applications. Then
measure repeated workstation averages and profile the shared concrete leaf.
Until such a change exists, do not reopen closed call/frame, raw-value,
map/Array, iterator, nominal-container, scheduler-identity, or WASM work.
