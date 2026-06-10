# Multi-Source External Performance Ledger

This report extends the refreshed generality catalog with independently
Docker-verified Word-Frequency, Document-Audit, and Lexical-Rollup rows. It is report-only:
it does not alter `../benchmarks/results.json` or any published result source.

Ratios are meaningful only within a row. The inputs, runners, and publication
environments differ across sources, so this ledger does not rank workloads
against each other.

## Source and Input Ledger

| Workload | Input | Able/reference source | Measurement status |
| --- | --- | --- | --- |
| Generality catalog | Per-suite catalog inputs | `2026-07-10-external-scorecard-recheck.json`; `../benchmarks/results.json` references | One pinned-host Able run per mode; 25-second cap |
| Word-Frequency | `corpus.md`, 131,072 bytes | `../benchmarks-word-frequency-publish-20260710/results.json` | Six Docker-verified rows |
| Document-Audit | Same 131,072-byte corpus | `../benchmarks-document-audit-publish-20260710/results.json` | Six Docker-verified rows |
| Lexical-Rollup | First 16,384 records of `wordlist.txt`, 1,743,363-byte checked-in source | `../benchmarks-lexical-rollup-publish-20260710/results.json` | Six Docker-verified rows |
| I-Before-E control | `wordlist.txt`, 1,743,363 bytes | Recheck report above and `../benchmarks/results.json` | One pinned-host Able run per mode; 25-second cap |
| Channel-Rollup | First 16,384 words of the same word list | `2026-07-10-channel-rollup-docker-publication.json` | Refreshed six direct Docker image runs after rebuilding the Able base; every Able mode uses the goroutine executor; report-only |

## Comparable Application Rows

| Workload | Mode | Able (s) | Go (s) | Able/Go | Ruby (s) | Able/Ruby | Python (s) | Able/Python |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Word-Frequency | compiled | 0.150955343 | 0.011430693 | 13.21x | 0.083066069 | 1.82x | 0.044039261 | 3.43x |
| Word-Frequency | bytecode | 6.623927764 | 0.011430693 | 579.49x | 0.083066069 | 79.74x | 0.044039261 | 150.41x |
| Document-Audit | compiled | 0.036843904 | 0.003030783 | 12.16x | 0.043371155 | 0.85x | 0.019010632 | 1.94x |
| Document-Audit | bytecode | 4.290248590 | 0.003030783 | 1415.56x | 0.043371155 | 98.92x | 0.019010632 | 225.68x |
| I-Before-E | compiled | 0.180000000 | 0.050000000 | 3.60x | 0.100000000 | 1.80x | 0.130000000 | 1.38x |
| I-Before-E | bytecode | 0.580000000 | 0.050000000 | 11.60x | 0.100000000 | 5.80x | 0.130000000 | 4.46x |
| Lexical-Rollup | compiled | 0.059276659 | 0.003364905 | 17.62x | 0.055081010 | 1.08x | 0.022383732 | 2.65x |
| Lexical-Rollup | bytecode | 4.519517367 | 0.003364905 | 1343.13x | 0.055081010 | 82.05x | 0.022383732 | 201.91x |
| Channel-Rollup | compiled | 1.655721179 | 0.006852194 | 241.63x | 0.073451352 | 22.54x | 0.105122927 | 15.75x |
| Channel-Rollup | bytecode | 4.462437608 | 0.006852194 | 651.24x | 0.073451352 | 60.75x | 0.105122927 | 42.45x |

## Selection Result

The ledger does not authorize an optimization candidate. Word-Frequency,
Document-Audit, and I-Before-E all exercise realistic text/file programs, but
their concrete profiles differ: map lookup and return coercion in
Word-Frequency, member-cache/iterator work in Document-Audit, and string
containment/text processing in I-Before-E. Their input sizes also differ.

Document-Audit and Lexical-Rollup are independent public
Iterable/filter/map/collect controls, while I-Before-E is the independent
text-file control. The paired bytecode profiles for the two pipeline programs
repeat only the previously rejected member-cache and inline-return candidates;
their remaining descendants differ. Together with Word-Frequency, the ledger
now has the requested coverage and still does not authorize an iterator/text
or named-container optimization.

Channel-Rollup adds the first ordinary cross-language channel/Future
application and materially misses both dynamic-language targets. Its warmed
profile pair with reduced BinaryTrees and a sequential Lexical-Rollup guard
does not repeat a material concrete scheduler, channel, or call/value leaf;
the shared `runTask` / `runResumable` parents are not optimization targets.
It therefore authorizes neither a channel-, scheduler-, nor
benchmark-specific implementation change. The refreshed Docker row runs the
retained primitive-handle repair, while the accompanying multi-run Able-only
ledger has no clean external reference rows and so cannot create a ratio-based
candidate. The tree-walker goroutine-executor repair remains verified in the
same concurrent Docker workload.
