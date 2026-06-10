# Regex NFA primitive active-thread carrier

Date: 2026-07-17

## Decision

Keep the canonical `able-stdlib` change. Every regex NFA consumer now stores
active threads in one shared primitive parallel-array carrier instead of
allocating a `RegexNFAThread` nominal value for each successful upsert.
Ordinary matching, Regex Set, and Regex Scanner use the same representation
and algorithms.

No compiler, bytecode VM, runtime, benchmark, pattern, verifier, or public API
changed. The NFA still uses a linear state scan; this change does not
reintroduce the rejected per-character state-position index.

## Representation and invariants

`RegexNFAThreads` contains parallel state (`Array i32`), start (`Array i32`),
and capture-reference arrays plus one active length. Each matcher owns current,
spare, and closure-scratch carriers:

1. upsert linearly scans the active state prefix and preserves the existing
   earliest-start replacement rule;
2. a move resets only the spare active length, writes the next generation,
   and swaps current and spare after the previous generation is dead;
3. closure scratch stores primitive copies rather than active-set indices, so
   replacing an active state cannot retroactively alter an older LIFO entry;
4. Regex Scanner owns persistent current/spare carriers across `feed` calls
   and reclosure always writes into the spare before swapping; and
5. post-match filtering compacts the active prefix in place without changing
   relative order or capture identity.

Carrier backing arrays grow only to their observed high-water mark. Inactive
capture slots can retain bounded tag-array references until overwritten, but
the carrier cannot grow per input character and capture arrays contain tags,
not input buffers. The existing capture copy-on-write rule remains unchanged.

## Warmed bytecode prerequisite

Each side used five independent processes, three timed `main()` calls per
process, the existing 128-word `--profile` input, one logical CPU,
`GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`. Baselines are the retained
immutable-capture-template results under the same contract.

| Application | Baseline ns/op | Carrier ns/op | Time | Bytes/op | Allocs/op |
| --- | ---: | ---: | ---: | ---: | ---: |
| Regex Suffix Audit | 788,791,110 | 678,356,313 | -14.00% | -35.11% | -38.53% |
| Regex Set Audit | 1,053,554,913 | 880,771,274 | -16.40% | -42.90% | -53.68% |
| Regex Stream Audit | 882,390,517 | 725,493,855 | -17.78% | -48.28% | -46.42% |

Unlike the rejected arena, the primitive carrier removes nominal construction
and nominal-field mutation together. All three consumers improve in time,
bytes, and allocations.

## Whole-process bytecode gate

An exact temporary copy of the pre-candidate stdlib was alternated with the
canonical candidate using the same bytecode binary. Set and Stream used five
full-input pairs. Suffix's first five bounded-input pairs were volatile, so a
reversed-order five-pair cohort was added and the ten pairs were combined.

| Application | Baseline mean | Carrier mean | Change |
| --- | ---: | ---: | ---: |
| Regex Suffix Audit (`--profile`) | 1.320 s | 1.286 s | -2.58% |
| Regex Set Audit | 4.316 s | 4.342 s | +0.60% |
| Regex Stream Audit | 4.052 s | 3.836 s | -5.33% |

Set is neutral inside workstation/process variance; Suffix and Stream improve.
Every output hash matched within each application.

## Compiled gate

Baseline and carrier binaries were built from the exact temporary baseline
stdlib and canonical candidate, preserved, and alternated on one logical CPU.
Suffix used ten pairs; the shorter Set and Stream applications used twenty.

| Application | Baseline mean | Carrier mean | Change | Output hash |
| --- | ---: | ---: | ---: | --- |
| Regex Suffix Audit | 1.177 s | 0.945 s | -19.71% | `48835ea1...f598a` |
| Regex Set Audit | 0.0900 s | 0.0805 s | -10.56% | `3d8f861a...8da2` |
| Regex Stream Audit | 0.1020 s | 0.0810 s | -20.59% | `dd7801f0...717b` |

All compiled consumers improve and every alternating launch produced the
expected stable hash.

## Correctness

- The complete regex exec-fixture slice (`14_02` through `14_25`) passes in
  bytecode mode in 12.181 seconds.
- The complete slice passes in tree-walker mode in 12.284 seconds.
- Focused Set, incremental-scanner, boundary, and scanner-compaction fixtures
  pass before the full slices.
- All changed source files remain below 1,000 lines; `regex_nfa.able` is 968
  lines and the new carrier module is 39 lines.

## Next recommendation

Refresh the affected mode-aware external scorecard rows for Regex Set and
Regex Stream, with an unchanged bytecode control. Their previous rows were
more than 93x slower than the faster Python/Ruby reference, and the retained
change is large enough that the product-level ratios are now stale.

This entails five verifier-backed Able and reference samples under the current
scorecard contract for compiled and bytecode modes, strict evidence checking,
and comparison with the two-cohort baseline. The refresh determines how much
of the user-facing gap closed before selecting another implementation wall.
Afterward, return to bounded profiles across unlike remaining bytecode misses
and admit a VM/runtime candidate only if the same concrete descendant repeats
in at least three applications.
