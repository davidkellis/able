# Bytecode Target-Miss Text Pair Profiles

This is a profile-selection tranche, not a runtime or stdlib change. It turns
the refreshed scorecard's text target misses into a direct comparison between
two independent Able programs, with an unrelated map/return guard.

## Method

The direct bytecode runtime benchmark loaded and warmed each program before
sampling CPU. It used the canonical stdlib pinned at
`/home/david/sync/projects/able-stdlib/src`,
`GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1`, and CPU affinity `0`. Program output
and benchmark success were the correctness checks.

| Workload | Input and role | Iterations | Result | CPU samples |
| --- | --- | ---: | ---: | ---: |
| I-Before-E | whole checked-in word list; direct text loop | 10 | 241,090,476 ns/op; 9,067,952 B/op; 1,928 allocs/op | 2.40 s |
| Lexical-Rollup | first 16,384 words; public iterator/filter/map/collect pipeline | 20 | 115,410,498 ns/op; 10,238,102 B/op; 259 allocs/op | 2.29 s |
| Word-Frequency | separate corpus; map/return guard | 5 | 1,164,617,162 ns/op; 50,034,160 B/op; 508,684 allocs/op | 5.80 s |

Retained profiles are
`20260710_i_before_e_bytecode_text_pair_10x.cpu.pprof`,
`20260710_lexical_rollup_bytecode_text_pair_20x.cpu.pprof`, and
`20260710_word_frequency_bytecode_text_pair_guard_5x.cpu.pprof` under
`v12/interpreters/go/.profiles/`.

## Evidence

The two target-miss text applications do not expose the same concrete text
wall:

| Workload | Material concrete work |
| --- | --- |
| I-Before-E | direct named/member calls, member-cache validation (11.3% cumulative), inline return (12.5%), and `String.contains` (3.8%) |
| Lexical-Rollup | generator execution (60.3%), typed patterns/type matching (12.7%), iterator-next/member/native dispatch, inline return (12.2%), and member-cache validation (7.4%) |
| Word-Frequency guard | HashMap lookup (5.7%), raw integer extraction (3.6%), cached named calls, map/type matching, and inline return (14.7%) |

`execStringContainsMemberFast(...)` has 90 ms/3.8% in I-Before-E and no sample
in either Lexical-Rollup or Word-Frequency. It is therefore a direct-loop text
leaf, not a shared string primitive candidate. Conversely,
`lookupCachedMemberMethodEntry(...)` and `finishInlineReturn(...)` occur in all
three profiles, but they are the existing generic member-cache and inline-return
families. The prior broad A/B work rejected both; their distinct callers here
do not create a new semantic subpath or justify a retry.

## Decision

Keep no runtime, compiler, or canonical-stdlib code. A shared VM dispatcher or
a repeated previously rejected parent is not sufficient evidence for a generic
optimization. In particular, do not add a word-list/string-shaped fast path,
do not reopen member-cache validation or return-guard rewrites, and do not
change `String.contains` based on a one-program sample.

## Next recommendation

Perform a feature-to-benchmark coverage audit across the v12 specification and
the full Able/Go/Python/Ruby suite, then add only missing cross-language
programs that combine underrepresented language features with ordinary
application work. Recent independent profiles repeatedly diverge below the VM
dispatcher, so more micro-optimization attempts would only revisit rejected
parents. The audit should map each language feature to existing benchmark
coverage and reference availability, identify missing real-program
combinations, and make any new workload verifiable in every supported mode.
That broader evidence is needed to find a generic wall that benefits real
programs rather than one benchmark family.
