# Regex NFA active-state index gate

Date: 2026-07-17

## Decision

Keep no compiler, bytecode VM, runtime, benchmark, or canonical-stdlib code
change from this tranche. Fresh profiles reproduced one concrete shared source
wall in Regex Suffix Audit, Regex Set Audit, and Regex Stream Audit: every
successful closure insertion enters `regex_nfa_upsert_thread`, which linearly
scans the active `Array RegexNFAThread` for an existing state.

A temporary canonical `able-stdlib` candidate added a state-to-position index
to each active thread set. Its final form reused two thread/index buffers so it
did not rebuild the index on every character. The candidate was general NFA
machinery: it did not inspect a pattern, input, benchmark, public regex API, or
nominal type in the compiler. It nevertheless failed the cross-consumer bar
because normal Regex Stream execution regressed materially. The candidate was
fully reverted.

## Fresh profile evidence

Each bytecode profile used the applications' existing 128-word `--profile`
mode, five warmed `main()` calls, `GOMEMLIMIT=1GiB`, `GOGC=50`,
`GOMAXPROCS=1`, and one logical CPU. Call tracing was enabled on both sides of
the causal comparison, so its overhead is matched rather than mistaken for a
normal-process result.

Before the candidate, the two hottest source operations in all three traces
were the length/read pair in the linear upsert scan:

| Application | scan `len` calls | scan `read_slot` calls |
| --- | ---: | ---: |
| Regex Suffix Audit | 1,177,680 | 970,920 |
| Regex Set Audit | 2,730,820 | 2,403,860 |
| Regex Stream Audit | 1,181,940 | 972,180 |

The retained compiled Suffix profile independently attributed 47.7% cumulative
CPU and 41.6% of sampled allocation objects to
`regex_nfa_add_closure`/`regex_nfa_upsert_thread`. This is a real shared NFA
wall, not a dispatcher-parent inference.

The final reusable-buffer candidate changed the matched traced measurements as
follows:

| Application | Baseline ns/op | Candidate ns/op | Change | Allocation change |
| --- | ---: | ---: | ---: | ---: |
| Regex Suffix Audit | 983,635,285 | 972,142,787 | -1.2% | -4.3% |
| Regex Set Audit | 1,375,325,081 | 1,241,634,981 | -9.7% | -3.7% |
| Regex Stream Audit | 1,100,553,073 | 1,096,674,303 | -0.4% | -5.1% |

The profile result was promising but insufficient for retention. Tracing
charges per call and therefore magnifies the benefit of deleting traced scan
operations.

## Normal-process and compiled gates

Five verifier-backed default bytecode processes were measured with the
candidate and again immediately after its revert:

| Application | Candidate mean | Reverted mean | Candidate change |
| --- | ---: | ---: | ---: |
| Regex Set Audit | 5.130 s | 5.164 s | -0.7% |
| Regex Stream Audit | 5.814 s | 4.792 s | +21.3% |

The Stream candidate batch contained one 7.77-second outlier. Removing it does
not change the decision: the remaining candidate mean is 5.325 seconds and its
median is 5.42 seconds, versus a 4.62-second reverted median. All ten outputs
passed their canonical verifiers.

The compiled side did benefit. Seven alternating full Suffix launches against
preserved binaries changed from 1.254 seconds to 1.201 seconds (-4.2%) with an
identical output hash. A compiled win in one consumer cannot override a large
bytecode regression in another consumer.

The non-regex Word Frequency guard completed five clean warmed processes after
the revert at 1,202,077,471 ns/op, 48,863,597 B/op, and 631,299 allocs/op. The
candidate had no compiler or VM component and `able.text.regex` is outside
Word Frequency's import closure, so no non-regex runtime path was eligible to
change.

## Correctness and cleanup

- All regex exec fixtures (`14_02` through `14_25`) pass in bytecode mode after
  the revert.
- The same fixture slice passes tree-walker/bytecode parity.
- Every measured normal Regex Set and Stream process passed its Ruby verifier.
- `regex_nfa.able` remains below the 1,000-line repository limit.
- The state-position helpers, scanner fields, signatures, and buffer swaps are
  absent from canonical `able-stdlib`; no production candidate remains.

## Next recommendation

Evaluate immutable initial capture-vector reuse across ordinary matching,
Regex Set, and Regex Scanner. `regex_nfa_new_captures` currently constructs a
fresh initialized capture Array at every candidate start boundary, even though
capture transitions clone that Array before writing and zero-capture programs
cannot mutate it at all. The compiled Suffix allocation profile attributes
15.2% of objects to this helper.

The experiment should hoist one initialized capture template per match/set
execution and one per scanner, prove that every capture write still targets a
clone, then gate the same three regex applications in compiled and bytecode
modes plus the full regex fixture slice. This is narrower than replacing the
whole thread-set representation, removes allocation rather than trading a
linear scan for index maintenance, and remains independent of pattern shape,
benchmark input, and public regex API.
