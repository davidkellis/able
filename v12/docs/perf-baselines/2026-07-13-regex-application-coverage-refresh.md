# Public regex application coverage refresh — 2026-07-13

## Decision

Promote the existing verifier-backed Regex Suffix Audit and RegexSet Audit
applications from the dedicated `regex-text` suite into the full `coverage`
inventory, then add Regex Stream Audit as the public chunked-scanner row.
Keep no VM, compiler, bridge, runtime, canonical-stdlib, or benchmark-source
performance change in this tranche.

This is a coverage repair, not permission for a regex-only optimization. The
applications exercise public `RegexBuilder`, `Regex`, `RegexSet`, and
`RegexScanner` APIs through ordinary canonical-stdlib code, and the sibling
benchmark sources match the v12 example entries. The existing regex gates have
already identified and kept general tagged-NFA traversal and thread-reuse
improvements; they also rule out an API-, pattern-, or corpus-specific path.

## Fresh verifier-backed comparison

All reference and Able launches used CPU 15, three processes per row, a
45-second process cap, `GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`.
Python was 3.14.5 and Ruby was 4.0.5. Each successful Able process was checked
by the matching sibling Ruby verifier.

| Application | Go (s) | Python (s) | Ruby (s) | Compiled Able (s) | Able/Go | Bytecode Able (s) | Able/Python | Able/Ruby |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Regex Suffix Audit | 0.0339 | 0.0396 | 0.0768 | 3.5667 | 105.21x | cap-bound | n/a | n/a |
| RegexSet Audit | 0.0045 | 0.0207 | 0.0459 | 0.2033 | 45.18x | 8.5133 | 411.27x | 185.47x |
| Regex Stream Audit | 0.0066 | 0.0202 | 0.0499 | 0.2133 | 32.32x | 6.4167 | 317.66x | 128.59x |

Compiled Regex Suffix Audit verified 3/3 with stdout hash
`48835ea1a1741c659d1b6b215a56e6611e525366596e08e9a10ec985106f598a`.
All three bytecode suffix processes correctly reached the cap before producing
a verifiable result. Compiled and bytecode RegexSet Audit both verified 3/3
with stdout hash
`3d8f861a312416f95b95d59be62614f0ffc7918e86fb984fe6035ca7b7b28da2`.
All three Regex Stream Audit modes verified 3/3 with stdout hash
`dd7801f0b104d8bd47aaef64d685bc65a06263851f12c8df8cf75260d09e717b`.
Its detailed coverage record is
`2026-07-13-regex-stream-application-coverage.md`.

The initial two-application direct commands were:

```text
./v12/bench_refresh_interpreter_refs \
  --benchmarks regex_suffix_audit,regex_set_audit --runs 3 --timeout 45 \
  --cpu-affinity 15 --python-bin python3.14 --ruby-bin ruby

./v12/bench_refresh_go_refs \
  --benchmarks regex_suffix_audit,regex_set_audit --runs 3 --timeout 45 \
  --cpu-affinity 15 --gomaxprocs 1

GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1 ./v12/bench_compare_external \
  --benchmarks regex_suffix_audit,regex_set_audit --modes compiled,bytecode \
  --languages go,python,ruby --runs 3 --timeout 45 --cpu-affinity 15
```

## Why no new performance candidate follows

The fresh runtime rows establish two material product misses, but do not alter
the ownership evidence. The earlier suffix, ordinary `Regex.is_match`, and
RegexSet profiles already showed general tagged-NFA move/closure and
thread-allocation work across three public API shapes. That evidence led to
the kept immutable outgoing-transition index, matcher-local closure stack, and
upserted-thread reuse in canonical `able-stdlib`; all are API-independent.

The bytecode profiles from the same cross-application gate instead diverged
from a non-regex text control at concrete leaves. Reprofiling unchanged sources
solely because the rows now belong to `coverage` would collect another ratio,
not test a new generic hypothesis. No `able-stdlib` edit is therefore needed.

## Catalog and ledger effect

`coverage` now has 28 application-shaped programs. `generality` remains the
stable 16-program timing baseline. The bytecode ledger records Regex Suffix
Audit as its eighth cap-bound row, RegexSet Audit as a verified target miss,
and Regex Stream Audit as the sixteenth verified target miss (18 completed
Python/Ruby comparisons in total).

`v12/bench_bytecode_audit --suite coverage` passes with all 28 applications,
124 lowered functions, and 6,872 bytecode instructions. This is a lowering
coverage check rather than a runtime result; it proves the new catalog members
remain part of the ordinary bytecode guard.

## Cross-API profile follow-up

The bounded Scanner/Suffix/RegexSet/I-Before-E profile gate is complete. It
kept no source change: the regex-only compiled NFA family is already covered
by prior generic stdlib work, and bytecode repeats only already-rejected VM
parents across the unlike control. See
`2026-07-13-regex-stream-cross-api-profile-gate.md`.

## Next recommendation

Select a new spec-defined semantic boundary with shared fixtures and a
portable verifier-backed application before more performance work. The full
public regex family is now both application-covered and profile-gated, so
reprofiling it unchanged would not improve generality evidence.
