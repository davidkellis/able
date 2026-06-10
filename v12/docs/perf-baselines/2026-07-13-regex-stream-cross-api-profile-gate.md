# Regex streaming cross-API profile gate — 2026-07-13

## Decision

Keep no VM, compiler, bridge, runtime, canonical-stdlib, or benchmark source
change. The new public `RegexScanner` application completes the cross-API
evidence, but its current costs either belong to the already-optimized generic
tagged-NFA family or do not repeat outside regex.

This is a negative selection result, not evidence that a scanner-, pattern-,
corpus-, or chunk-shape path is valid. It prevents the fresh streaming row from
turning one product miss into a narrow optimization.

## Method

All collection used current canonical `able-stdlib`, CPU 15, and the bounded
interpreter guard `GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1`.

The warmed bytecode profiles ran the normal program-runtime benchmark with no
parser/bootstrap sample in the target phase:

| Program | Profile scale | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: | ---: |
| Regex Stream Audit | 5x, 128 words | 1,403,538,090 | 60,945,201 | 557,305 |
| Regex Suffix Audit | 5x, 128 words | 1,271,968,030 | 56,647,166 | 496,307 |
| RegexSet Audit | 5x, 128 words | 2,011,775,715 | 72,038,672 | 594,379 |
| I-Before-E control | 20x | 249,555,743 | 9,063,477 | 1,926 |

The four retained bytecode CPU profiles are
`v12/interpreters/go/.profiles/20260713_{regex_stream_audit,regex_suffix_audit,regex_set_audit}_bytecode_cross_api_5x.cpu.pprof`
and
`v12/interpreters/go/.profiles/20260713_i_before_e_bytecode_cross_api_20x.cpu.pprof`.

The compiled half used only generated-program `main` CPU phase profiles: 20
verified Stream runs, 5 verified Suffix runs, 20 verified RegexSet runs, and
20 verified I-Before-E runs. Each process checked its ordinary output before
its sample was accepted. It deliberately did not collect allocation profiles,
so bootstrap/parser work and allocation instrumentation cannot be mistaken for
an application execution candidate.

## Attribution

The bytecode profiles share `execCallOpcode`, but that is the dispatcher parent
already rejected by broad call-name, raw-cell, frame, and return experiments.
Their concrete descendants differ:

| Program | Relevant concrete descendants |
| --- | --- |
| Stream | cached identifier lookup 10.89%; array-slot member call 14.33%; inline return 7.02% |
| Suffix | cached identifier lookup 7.58%; array-slot member call 15.96%; inline return 7.27% |
| RegexSet | cached identifier lookup 16.40%; array-slot member call 16.50%; inline return 4.30% |
| I-Before-E | member-method lookup 15.49%; member call 30.99%; array-slot member call 7.65%; inline return 12.88% |

Thus cached identifier and array-slot work is a regex-only VM shape here,
while the ordinary text control is member-cache/call dominated. The common
dispatcher cannot justify another generic VM change without a repeated concrete
child across unlike programs.

Compiled generated-main samples do repeat a canonical regex family:

| Program | Public entry | `regex_nfa_move` | `regex_char_codepoint` | `runtime.mallocgc` |
| --- | ---: | ---: | ---: | ---: |
| Stream | `RegexScanner.feed` 77.21% | 31.16% | 41.40% | 57.67% |
| Suffix | `Regex.match` / `find_nfa_span` 84.71% / 78.24% | 37.65% | 44.49% | 49.71% |
| RegexSet | `RegexSet.matches` 79.37% | 49.78% | 56.95% | 58.30% |
| I-Before-E | line reading / validation | not present | not present | 22.39% |

The repeated regex descendants are not new candidates. The canonical stdlib
already has the generic outgoing-transition index, reusable closure scratch,
and upserted-thread reuse selected by the preceding NFA profile gates. The
character-to-codepoint route was separately rejected by the compiled character
generality controls because it did not repeat across non-regex character
programs. I-Before-E confirms that the residual family remains regex-specific,
not a general text/runtime allocation wall.

## Result and cleanup

No source experiment followed. The evidence does not permit a `RegexScanner`,
`RegexSet`, pattern, chunking, or corpus specialization, and it does not
reopen the previously rejected generic VM or character-conversion directions.
`able-stdlib` needs no change.

The generated compiled-profile workspace was a temporary, agent-created
`/tmp/able-regex-cross-api.*` directory (3.0 GiB, including generated source
trees). It was removed after the profile summaries above were recorded; only
the small bytecode evidence profiles remain in the project profile directory.

## Next recommendation

Return to feature-led selection from the active v12 roadmap rather than
reprofiling unchanged regex paths. The work should first add a missing
spec-defined behavior to the shared fixtures and a portable verifier-backed
application, then use a bounded profile gate only if that application exposes
a material miss. This is the right next step because the public regex API
family is now covered and its remaining concrete costs are either already
addressed or isolated from unlike programs; a new semantic boundary is the
credible source of a broadly useful, previously unmeasured optimization.
