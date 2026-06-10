# Host-image bytecode miss profiles — 2026-07-14

## Decision

Keep no new bytecode VM, compiler, canonical-stdlib, or benchmark-source
optimization from this tranche. Two unlike, output-verified programs reach the
normal call/return dispatcher, but their costly descendants are different.
I-Before-E is cached text-member dispatch; Word Frequency is generic string-key
map lookup and typed-pattern work. Reopening a call-name, raw-cell, or
inline-return micro-optimization would repeat variants that already failed the
broad benchmark bar.

## Method

Each application first ran through the normal bytecode CLI with its public Ruby
verifier, using CPU 15 and the persistent complete-program extern-image cache.
The steady-state harness then loaded and prepared the whole program once,
warmed `main`, and profiled only repeated `main` calls. This excludes loader and
one-time image-build work while exercising the same normal image path.

| Application | Normal bytecode elapsed | Verification | Warmed calls | Warmed ns/op | B/op | allocs/op | CPU samples |
| --- | ---: | --- | ---: | ---: | ---: | ---: | ---: |
| I-Before-E (`wordlist.txt`) | 0.5100 s | passed | 30 | 263,979,579 | 9,062,031 | 1,922 | 7.89 s |
| Word Frequency (`corpus.md`) | 1.4800 s | passed | 7 | 1,221,084,533 | 49,778,874 | 631,330 | 8.52 s |

The profile repeat counts were selected from a one-call warmed measurement to
produce roughly eight seconds of CPU samples while each test process remains
well below the one-minute limit. The ignored raw profiles are retained locally
for this active investigation and are cleanup-eligible when work pauses:

- `v12/interpreters/go/.profiles/20260714_external_i_before_e_host_image.cpu.pprof`
- `v12/interpreters/go/.profiles/20260714_external_word_frequency_host_image.cpu.pprof`

## CPU evidence

| Program | Material profile evidence | Interpretation |
| --- | --- | --- |
| I-Before-E | `execCallOpcode` is 4.19 s cumulative (53.11%); `execCallMember` 2.26 s (28.64%); `lookupCachedMemberMethodEntry` 1.13 s (14.32%); `finishInlineReturn` 0.91 s (11.53%) | Text/file member-call and cache validation dominate. |
| Word Frequency | `execCallOpcode` is 3.26 s cumulative (38.26%); `execCallName` 1.91 s (22.42%); `finishInlineReturn` 1.28 s (15.02%); generic `hashMapFindEntryWithHash` 0.53 s (6.22%); `runtime.mapaccess2_faststr` 0.71 s (8.33%) | String-key map lookup, type matches, and named-call work dominate. |

`runResumable`, `execCallOpcode`, and `finishInlineReturn` recur only as broad
VM parents. The concrete call cache shapes differ (member versus name), and
the map leaf is present only in the map-shaped program. The relevant generic
return and raw-cell/call-name variants have already been measured against
unlike workloads and either made no material difference or regressed them.
This profile pair therefore supplies no new reusable leaf or safe source-level
candidate. It also does not justify a `HashMap`, iterator, text, or benchmark
exception.

## Result

The complete-program host image remains correct under real programs, and both
profiles are clean steady-state evidence rather than plugin-startup noise. No
runtime change is warranted. JSON remains the existing independent guard
workload: it already runs faster than both reference interpreters in the
host-image scorecard, so any later VM candidate must retain that result as well
as improving more than one residual miss. The external `able-stdlib` checkout
did not need a change.

## Follow-up

The first coverage-reference expansion is complete for Word Frequency,
Document Audit, and Lexical Rollup. Its fresh ratios and two iterator profiles
are recorded in `2026-07-14-bytecode-coverage-reference-profile-gate.md`.
They confirm large misses but no new reusable VM leaf, so a generic iterator
protocol design audit—not another member-cache micro-variant—is the only
appropriate follow-up.
