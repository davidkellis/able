# Compiler generated struct-field read audit retained

## Decision

**retain-general-positional-aware-semantic-reads**.

The complete compiler source audit found 85 textual
`StructInstanceValue.Fields[...]`-style occurrences. Fourteen were generated
semantic reads that either assumed map-backed named storage or duplicated part
of the shared positional-aware lookup. Two additional functional-update paths
rejected positional-backed named structs before copying their fields.

All semantic named-field reads now use the existing shared accessor. The
change preserves map-first behavior for runtime-service overlays while also
resolving named fields from positional nominal storage. It adds no boxing,
interpreter fallback, container-specific storage rule, or non-primitive
nominal lowering branch.

This is a correctness and boundary-hardening result, not a performance claim.
Machine-readable summary:
`2026-07-27-compiler-generated-struct-field-read-audit-retained.json`.

## Retained corrections

The following generated or bridge paths now use the shared named-field
accessor:

- String `bytes` conversion;
- Array storage-handle conversion, mutation support, member reads, shape
  recognition, and formatting;
- generic runtime member reads and callable-field method lookup;
- FutureError `details`;
- synthetic Awaitable callable-field recognition; and
- direct-runtime and IR named-struct functional updates.

The functional-update correction is representation-agnostic. It preserves any
map-backed overlays, then copies every semantic definition field through the
shared accessor. A positional-backed named struct is therefore accepted
without first being converted to map-backed storage.

## Closure census

After the repair, 71 textual indexed-field occurrences remain:

| Category | Count | Disposition |
| --- | ---: | --- |
| writes | 53 | retained; construct or update explicit map-backed boundary/service values |
| compile-time AST/type metadata reads | 14 | out of runtime representation scope |
| shared accessor implementations | 2 | authoritative map-first/positional-aware boundary |
| same-map enumeration reads | 2 | safe; key is obtained by iterating that exact map |
| unclassified semantic reads | 0 | closed |

The two same-map reads are fallback string rendering and generated-main
formatting. They do not resolve a semantic field name independently of the map
being enumerated.

## Positive controls

New structural guards require the shared accessor and forbid the former
representation-specific reads in:

- String, Array, member-get, member-set, callable-member, FutureError, and
  generated-main helpers;
- direct runtime functional updates; and
- IR functional updates.

The existing bridge control constructs both map-backed and positional-backed
named instances and verifies that the shared accessor returns the same
semantic field from both representations. Existing synthetic Awaitable tests
also remain green after routing their map-backed callable fields through the
same accessor.

## Verification

Focused positional-field, Array-boundary, runtime-functional-update, IR, and
bridge guards passed. The final-code compiler release lane then passed:

| Matrix | Result |
| --- | --- |
| compiler bridge | pass |
| compiler core | 32/32 batches pass |
| compiler outliers | 3/3 pass |
| fallback audit | 24/24 shards pass |
| compiled execution | 24/24 shards pass |
| strict dispatch | 24/24 shards pass |
| interface-lookup bypass | 24/24 shards pass |
| boundary fallback markers | 24/24 shards pass |
| `go test ./cmd/ablec -count=1 -timeout=60s` | pass, 5.830s |
| `git diff --check` | pass |

Every release shard completed below one minute. The longer core figures are
multi-test batch aggregates, not individual tests. The largest touched
production file remains `generator_render_runtime_future.go` at 992 lines.

No runtime, interpreter, bytecode VM, language, canonical stdlib, dependency,
benchmark application, or WASM source changed in this tranche.

## Performance evidence state

The checked scorecard inputs remain complete, but the closure selector
conservatively invalidates 12 compiled-related closures solely because the
compiler-production scope changed. Nine bytecode-only closures remain current.
No closure was rebased and no timing claim was made.

The audit does not itself establish which current applications execute each
corrected helper branch. Measuring all compiled rows immediately would mix
changed and unchanged execution, while rebasing without reach data would make
stale evidence appear current.

## Cleanup

The complete release lane used a disk-backed `/var/tmp` workspace. Its exact
23 GiB directory was removed after verification. No `/tmp/able-*` directory
remains.

## Next recommendation

Run a strict generated-code and runtime reach census for every migrated
semantic-read family across all 63 portable applications.

Why: compiler-scope hashing correctly invalidates 12 closures, but helper
emission or linkage does not prove that a current application executes a
corrected branch. Some reached map-backed paths also add the shared accessor's
general representation checks and therefore require measured regression
evidence.

What it entails: strictly emit and link every application under disk-backed
`/var/tmp`; verify all final graphs remain interpreter-free; place separate
positive-controlled sentinels at String, Array, generic member, callable
member, FutureError, Awaitable, and functional-update branches; run each
reached binary with its catalog configuration and public verifier; and collect
at least five balanced baseline/candidate/equivalent-Go processes for every
reached application. Advance only the closure snapshots justified by those
results.

Why it is important: this keeps positional native carriers and the
compiler/interpreter boundary correct while ensuring the general accessor
rule does not silently move compiled applications away from the Go
performance target. Do not begin WASM work.
