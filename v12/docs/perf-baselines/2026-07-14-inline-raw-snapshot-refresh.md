# Post-repair bytecode refresh — 2026-07-14

## Decision

Keep no bytecode VM, compiler, generated-runtime, canonical-stdlib, or
benchmark-source performance change from this refresh. The preceding change
repairs raw integer argument ownership across nested inline calls; it is a
correctness repair, not a performance experiment. This document establishes a
fresh, verifier-backed normal-process baseline before any future candidate is
considered.

## Normal-process evidence

`dependency_plan`, `word_frequency`, and `document_audit` each ran three times
in bytecode mode on CPU 12. The immediate three-sample host preflight passed
(2.94%, 0.00%, and 0.00% busy), every run completed within the 45-second
limit, and every output passed its canonical verifier.

| Application | Runs | Verification | Mean real time |
| --- | ---: | --- | ---: |
| Dependency Plan | 3/3 | verified | 0.4767 s |
| Word Frequency | 3/3 | verified | 1.4467 s |
| Document Audit | 3/3 | verified | 0.3467 s |

The raw verifier hashes and per-run summaries are retained temporarily in
`v12/tmp/perf/20260714_inline_raw_snapshot_refresh/`. These are complete
process measurements, so they include loading and startup. They must not be
read as a performance delta from earlier measurements on different cores; they
only confirm the repaired VM executes three unlike real applications correctly
under the normal benchmark harness.

## Reproducible warmed-profile invocation

`bench_perf` now accepts `--bytecode-runtime-calls N`, and
`bench_compare_external` forwards it to the same warmed Go benchmark. This
keeps a bounded CPU capture in the external catalog's canonical invocation:
the target, input argument, run directory, source-root policy, and verifier
association cannot drift into a hand-written profiling command. It is generic
benchmark infrastructure and changes no VM, compiler, generated runtime,
stdlib, fixture, or application behavior.

The integration guard ran Document Audit through `bench_compare_external` with
one and then two warmed `main()` calls. Both completed successfully; the second
run reported the configured two calls and used the catalog's required
`word-frequency/corpus.md` argument. These short integration checks are not
performance evidence. The existing normal-process measurements remain the
output-verifying semantic guard for the application.

## Bounded-profile status

The required warmed bytecode-runtime profile gate is complete. Each capture
used `GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, the canonical external
stdlib, and the catalog's canonical source/input setup. CPU 1, CPU 2, and CPU
14 respectively passed their immediate three-sample preflights before the
Document Audit, Word Frequency, and Dependency Plan runs. The profile starts
only after loading and one warm `main()` call; it therefore measures repeated
whole-program execution rather than parser, loader, or bootstrap work.

| Application | Core | Calls | ns/op | B/op | allocs/op | CPU samples |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Document Audit | 1 | 900 | 13,194,300 | 379,870 | 564 | 11.79 s |
| Word Frequency | 2 | 7 | 1,080,599,552 | 48,394,771 | 631,191 | 7.54 s |
| Dependency Plan | 14 | 28 | 246,154,577 | 2,194,236 | 33,242 | 6.88 s |

The normal-process runs above remain the verifier-backed semantic checks. The
warmed Go benchmark deliberately does not capture application stdout, but it
uses those same catalog targets, input paths, run directories, and source-root
rules.

## Cross-application attribution

The three applications do not repeat a new material concrete descendant.
Document Audit is lazy-iterator/member-cache dominated:
`lookupCachedMemberMethodEntry` is 28.58% cumulative and
`finishInlineReturn` is 7.89%. Word Frequency is named-call/return and
string-key map work: `finishInlineReturn` is 14.19% cumulative and
`hashMapFindEntryWithHash` is 4.77%. Dependency Plan is Array/member/index and
integer-map work: member lookup is 13.52% cumulative,
`runtime.mapaccess2_fast64` is 7.99%, and `finishInlineReturn` is 5.96%.

`finishInlineReturn` is the only concrete VM leaf represented materially in
all three profiles. It is not eligible for another experiment: the prior
generic guard-order candidate was neutral-to-mixed across unlike workloads and
was reverted. Member-cache work appears only in Document Audit and Dependency
Plan; the string-map and Array/index leaves are each application-shaped. This
does not authorize a return, map, Array, Queue, iterator, text, graph, or
benchmark exception.

The retained profiles are
`v12/interpreters/go/.profiles/20260714_inline_raw_snapshot_{document_audit,word_frequency,dependency_plan}.cpu.pprof`.
The matching machine-readable reports are the top-level
`{document_audit,word_frequency,dependency_plan}_profile.{json,md}` files
under `v12/tmp/perf/20260714_inline_raw_snapshot_refresh/`. They, the raw
profiles, and their temporary benchmark binaries are cleanup-eligible when
active performance investigation pauses.

## Next selection boundary

Do not retry return, member-cache, map, Array/index, raw-integer, or
type-match micro-variants from this evidence. The follow-up semantic/coverage
audit reconfirmed that there is currently no missing spec-defined behavior that
can honestly become a portable cross-language timing application: the 32
portable applications are supplemented by 77 intentionally local semantic
fixtures, while dynamic-module, user-extern, and host boundaries have no fair
foreign-runtime counterpart. Do not invent a synthetic timing loop or repeat
profiles of unchanged code. Resume implementation selection only after a real
cross-cutting semantic/compiler change, or a newly needed portable application,
exposes a previously untried concrete leaf across at least three unlike
applications; then guard that candidate on the stable generality suite. The
full rationale is in
`2026-07-14-feature-benchmark-roadmap-reconciliation.md`.
