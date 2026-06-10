# Regex NFA matcher-owned thread-arena gate

Date: 2026-07-17

## Decision

Keep no production change. A bounded double-buffered arena correctly reused
`RegexNFAThread` records across NFA input boundaries, reduced allocation
volume in all three representative regex applications, but regressed warmed
bytecode time in all three. The candidate and its temporary arena source file
were fully reverted.

No compiler, bytecode VM, runtime, benchmark, pattern, verifier, or public API
changed. The retained immutable initial capture-template optimization remains
in canonical `able-stdlib`.

## Ownership and lifetime result

The experiment used two matcher-owned arenas rather than one:

1. the current arena owned every record reachable from the active thread set;
2. the spare arena was reset before a move or scanner reclosure;
3. the new active set was constructed only in the spare arena; and
4. the arenas swapped only after the previous active set and closure scratch
   references were dead.

Within a generation, replacement allocated another arena slot instead of
mutating an existing active record, because the closure stack can still hold
the replaced record. Scanner arenas lived on `RegexScanner` and therefore
survived across `feed` calls. Thread capture arrays could safely cross arena
generations because the only capture write sites clone before mutation.

The complete regex fixture slice passed both bytecode and tree-walker modes
with the candidate, so the lifetime model preserved observed semantics. This
closes the ownership question even though the representation failed the
performance gate.

## Warmed bytecode gate

Each candidate row is the mean of five independent processes with three timed
`main()` calls per process, the existing 128-word `--profile` input, one
logical CPU, `GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`. The baseline is
the immediately preceding retained capture-template gate under the identical
contract.

| Application | Baseline ns/op | Arena ns/op | Time | Bytes/op | Allocs/op |
| --- | ---: | ---: | ---: | ---: | ---: |
| Regex Suffix Audit | 788,791,110 | 861,633,974 | +9.23% | -11.70% | -6.21% |
| Regex Set Audit | 1,053,554,913 | 1,121,963,631 | +6.49% | -12.72% | -10.05% |
| Regex Stream Audit | 882,390,517 | 904,928,786 | +2.55% | -21.78% | -10.86% |

The direction repeats across all three workloads: the arena saves heap bytes
and objects, but the extra arena dispatch plus repeated nominal-field writes
cost more than fresh record construction in the bytecode VM. This is not a
single volatile timing or an allocation-attribution inference.

Compiled builds were not run after this decisive gate. A candidate that
regresses all three bytecode consumers cannot be retained as the shared stdlib
representation, so compiled results could not change the selection decision.
Avoiding three multi-minute builds also kept the tranche within the project's
bounded-test process.

## Correctness and cleanup

- With the candidate, the complete regex exec-fixture slice (`14_02` through
  `14_25`) passed in bytecode mode in 22.141 seconds and tree-walker mode in
  12.201 seconds.
- After the revert, the complete bytecode slice passed in 12.221 seconds.
- SHA-256 checks confirmed `regex_nfa.able`, `regex_set.able`, and
  `regex_scanner.able` exactly matched their pre-experiment sources.
- The temporary `regex_nfa_thread_arena.able` file was removed.
- All candidate source files remained below 1,000 lines.

## Next recommendation

Evaluate a primitive active-thread carrier, but only as one shared NFA
representation used by ordinary matching, Regex Set, and Regex Scanner. The
arena proved that object reuse is the wrong mechanism, while the allocation
reductions confirm that nominal thread construction is still material. A
primitive carrier can remove per-thread nominal allocation and nominal-field
mutation together rather than trading one for the other.

This entails designing parallel state/start/capture storage with one active
length and two move buffers, preserving earliest-start replacement, closure
order, tagged-capture identity, filtering, and scanner retention across
`feed`. It must not recreate the rejected per-character state-position index,
and it must pass repeated Suffix, Set, and Stream bytecode gates before any
compiled builds. If the three-way bytecode gate is neutral or negative, stop
regex representation work and return to cross-application VM profiles.
