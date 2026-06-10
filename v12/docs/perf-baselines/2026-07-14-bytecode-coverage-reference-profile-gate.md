# Bytecode coverage-reference and profile gate — 2026-07-14

## Decision

Keep no bytecode VM, compiler, canonical-stdlib, or benchmark-source
optimization from this tranche. Fresh Python/Ruby references make three
feature-rich application misses measurable, but the host-image CPU profiles
still divide their material work among map lookup/type matching and iterator
member dispatch. The only recurring iterator cache and inline-return families
have already failed broad guard workloads, so they are not eligible for a
retry.

## Reference-aware scorecard

The existing `bench_refresh_interpreter_refs` and
`bench_compare_external` lanes ran the installed Python 3.14 and Ruby 4.0
implementations, then normal bytecode processes, three times apiece on CPU 15.
Every process passed its public Ruby verifier and produced the same stdout
digest across languages.

| Application | Able bytecode | Python | Able/Python | Ruby | Able/Ruby |
| --- | ---: | ---: | ---: | ---: | ---: |
| Word Frequency | 1.4933 s | 0.0302 s | 49.45x | 0.0636 s | 23.48x |
| Document Audit | 0.3100 s | 0.0264 s | 11.74x | 0.0543 s | 5.71x |
| Lexical Rollup | 0.4033 s | 0.0296 s | 13.62x | 0.0621 s | 6.49x |

The retained rows are [fresh interpreter references](2026-07-14-bytecode-coverage-interpreter-refs.md)
and the [comparison scorecard](2026-07-14-bytecode-coverage-scorecard.md).
`results.json` did not contain these newer benchmark families, so the fresh,
verified local rows are the appropriate reference source rather than a stale or
missing stored result.

An explicit `--benchmarks` selection now reports its suite as `custom` instead
of incorrectly inheriting the default `generality` or `core` label. This is
reporting metadata only; it changes neither workload selection nor timing.

## Bounded warmed CPU evidence

The normal CLI measurements include startup. To examine VM work, each program
was loaded through the complete-program host-image path, warmed once, then
profiled only while repeatedly calling `main` on CPU 15.

| Application | Calls | Warmed ns/op | B/op | allocs/op | CPU samples |
| --- | ---: | ---: | ---: | ---: | ---: |
| Word Frequency | 7 | 1,221,084,533 | 49,778,874 | 631,330 | 8.52 s |
| Document Audit | 900 | 7,546,444 | 364,388 | 211 | 6.76 s |
| Lexical Rollup | 64 | 120,706,885 | 10,229,614 | 252 | 7.70 s |

| Application | Material descendants | Result |
| --- | --- | --- |
| Word Frequency | `hashMapFindEntryWithHash` 6.22% cumulative; `runtime.mapaccess2_faststr` 8.33%; named-call and type-match work | Generic string-key map and typed-result workload. |
| Document Audit | `lookupCachedMemberMethodEntry` 15.53%; cached `next()` 8.28%; `finishInlineReturn` 12.57% | Lazy iterator member/cache work plus text predicates. |
| Lexical Rollup | `lookupCachedMemberMethodEntry` 10.91%; cached `next()` 5.97%; `finishInlineReturn` 15.71%; type matching 11.43% | Lazy iterator work combined with typed-pattern and text work. |

The two iterator applications independently exercise lazy `filter → map →
collect → reduce`, so their recurring `Iterator.next()` call is real protocol
work. It remains mediated by the generic member cache in order to preserve
callable-field shadowing, dynamic definition, interface, and lexical-version
semantics. Its concrete validation/return descendants are the same categories
already measured and rejected in prior broad tests; Word Frequency does not
share that leaf. This does not authorize a named iterator, map, text, or
benchmark fast path.

Ignored raw profiles retained while development continues:

- `v12/interpreters/go/.profiles/20260714_external_word_frequency_host_image.cpu.pprof`
- `v12/interpreters/go/.profiles/20260714_external_document_audit_host_image.cpu.pprof`
- `v12/interpreters/go/.profiles/20260714_external_lexical_rollup_host_image.cpu.pprof`

They are cleanup-eligible when this investigation pauses.

## Follow-up

The semantics-first audit is complete in
`v12/design/iterator-stable-call-plan-audit.md`. It finds that a new stable
`next` call plan would either violate callable-field/interface-overlay/dynamic
invalidation semantics or repeat the existing member-cache checks. No prototype
is authorized. The next measurement priority is a fresh compiled-vs-Go
coverage scorecard for the same verified applications, where the project still
needs direct evidence toward its 95% AOT goal.
