# Regex immutable initial capture-template reuse

Date: 2026-07-17

## Decision

Keep the canonical `able-stdlib` change. Ordinary matching, Regex Set
execution, and Regex Scanner now construct one initialized capture vector per
execution or scanner and reuse it for each candidate start. Capture vectors
are copy-on-write inside the NFA: the only two write sites,
`regex_nfa_capture_start` and `regex_nfa_capture_end`, clone their input before
calling `write_slot`. The shared template therefore remains immutable.

No compiler, bytecode VM, runtime, benchmark, reference implementation,
verifier, pattern, or public regex API changed. Zero-capture and tagged-capture
programs use the same ownership rule.

## Ownership proof

The canonical regex source contains exactly two capture-array writes. Both
have this sequence:

1. clone the incoming Array with `regex_nfa_clone_captures`;
2. write only to the clone; and
3. return the clone for the new NFA thread.

Active threads may share the initialized template until a group transition is
taken, but no thread can mutate it. Accepted captures are also cloned before
they replace `best_captures`. Scanner reset now points `best_captures` back to
the immutable template rather than allocating another unused initialized
vector.

## Warmed bytecode gate

Each side used five independent benchmark processes, three timed `main()`
calls per process, the existing 128-word `--profile` input, one logical CPU,
`GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`.

| Application | Baseline ns/op | Candidate ns/op | Time | Bytes/op | Allocs/op |
| --- | ---: | ---: | ---: | ---: | ---: |
| Regex Suffix Audit | 847,212,361 | 788,791,110 | -6.9% | -16.3% | -12.6% |
| Regex Set Audit | 1,415,347,459 | 1,053,554,913 | -25.6% | -9.2% | -8.5% |
| Regex Stream Audit | 1,178,308,039 | 882,390,517 | -25.1% | -15.2% | -11.0% |

The allocation reductions are direct measurements from the warmed Go
benchmark, not CPU-profile attribution. They cover both capture-bearing
Suffix/Stream programs and the zero-capture Regex Set program.

## Normal-process bytecode gate

Whole-process Set and Stream results were volatile in separated batches, so a
temporary baseline stdlib snapshot and the retained candidate were run with
the same bytecode binary in paired order, then in reversed order. Five pairs
were collected in each direction.

| Application | Baseline mean | Candidate mean | Change |
| --- | ---: | ---: | ---: |
| Regex Set Audit | 4.993 s | 4.983 s | -0.2% |
| Regex Stream Audit | 4.451 s | 4.472 s | +0.5% |

These complete-process results are neutral within workstation variance. Every
output hash matched the canonical verifier hash. They rule out the large
Stream regression that rejected the preceding active-state-index experiment.

## Compiled gate

Baseline and candidate Suffix binaries were preserved and alternated for ten
full verifier-sized launches on one logical CPU:

| Side | Mean | Output hash |
| --- | ---: | --- |
| Baseline | 1.067 s | `48835ea1...f598a` |
| Candidate | 1.004 s | `48835ea1...f598a` |

The candidate improves compiled Suffix by 5.9%. Sampled compiled allocation
profiles also show `regex_nfa_new_captures` falling from about 1.96 million
objects to about 0.19 million objects. Total sampled objects are noisier
because Go's allocation profile is sampled; the exact bytecode B/op and
allocs/op measurements above are the allocation selection evidence.

## Correctness

- The complete regex exec-fixture slice (`14_02` through `14_25`) passes in
  bytecode mode.
- The complete slice passes tree-walker/bytecode parity.
- The compiled RegexSet fixture passes in 54.742 seconds.
- A compiled incremental-scanner fixture reached the 55-second repository
  ceiling while its child executable was still running. It produced no
  semantic failure and was not rerun beyond the one-minute policy.
- Compiled Suffix and normal Set/Stream application hashes match their
  canonical verifiers.
- All changed Able source files remain below 1,000 lines.

## Next recommendation

Refresh post-change profiles for all three regex applications and evaluate the
remaining `RegexNFAThread` construction/upsert wall as an ownership problem,
not another sidecar index. The preceding index experiment proved that deleting
the linear scan alone can be outweighed by per-character index initialization
and scanner maintenance. The remaining compiled Suffix profile attributes
roughly half of sampled objects to `regex_nfa_upsert_thread`.

The next tranche should first design and prove a bounded, matcher-owned thread
arena or primitive state/start carrier whose records cannot escape a match or
scanner. It must preserve earliest-start replacement, tagged-capture identity,
closure work-stack ordering, and scanner retention across `feed` calls. Only
then should one representation candidate be measured across Suffix, Set, and
Stream in both compiled and bytecode modes. Do not reintroduce the rejected
per-character state-position arrays or add an API-, pattern-, or benchmark-
specific path.
