# Bytecode Register-Block Census and Slot-Not-Nil Gate

Date: 2026-07-18

## Decision (superseded by the full-suite gate)

This bounded census initially admitted one general bytecode quickening:
adjacent `LoadSlot` and `JumpIfNotNil` instructions were finalized as a direct
`JumpIfSlotNotNil` instruction at the original load position. The original
branch instruction remained in place so jumps that entered it directly
preserved the ordinary stack-based behavior.

The subsequent selected-bytecode refresh found a reproducible Mandelbrot
regression that this seven-workload A/B set did not cover. A new exact
preserved-binary five-pair gate measured the candidate 9.13% slower, so the
quickening was fully reverted and this report's keep decision is superseded by
`2026-07-18-post-slot-not-nil-bytecode-scorecard-reconciliation.md`. The
census and original A/B evidence remain useful records of why the candidate
was tried and why a smaller guard set was insufficient.

This is a register-form transport optimization, not a benchmark or nominal-
type rule. It reads the existing slot and raw-i32 sidecars without creating a
temporary stack value, taking a stack snapshot, or dispatching the second
instruction. Nil, out-of-range slots, non-nil values, and raw-i32 slots retain
the behavior of the original instruction pair.

No compiler, runtime value ABI, language, benchmark, fixture, or canonical
`able-stdlib` source changed.

## Bounded dynamic census

A temporary opt-in observer counted same-program, adjacent three-instruction
windows. Each complete process ran on CPU 0 with `GOMAXPROCS=1`, the canonical
external stdlib, the catalog working directory and arguments, a public output
verifier, and the existing 55-second profile cap. The observer was deliberately
serial and was fully removed before either A/B binary was built.

| Application | Verified wall | Counted triples |
| --- | ---: | ---: |
| Fixed Width 128 | 8.82 s | 33,162,969 |
| Reverse Complement | 7.66 s | 58,834,619 |
| Distance Field | 5.99 s | 42,000,160 |
| Mandelbrot | 8.25 s | 177,782,940 |
| Array Slice Window | 0.84 s | 5,785,640 |
| Word Frequency | 1.62 s | 10,574,650 |
| Option/Result Config | 0.92 s | 1,805,966 |
| Lexical Rollup | 0.93 s | 776,736 |

K-Nucleotide exceeded the cap with the observer enabled. Its partial record
was removed and was not used as negative or positive evidence. A first partial
Reverse Complement record was likewise removed after a complete rerun. The
eight complete raw JSON snapshots are retained in
`2026-07-18-bytecode-block-census/`.

The exact material windows repeated across at least three applications as
follows:

| Exact window | Material occurrences | Decision |
| --- | --- | --- |
| `Pop -> StoreSlotBinaryIntSlotConst -> Jump` | Array 5.19%, Distance 4.76%, Mandelbrot 8.83%, Reverse 3.46% | Already-specialized loop store and jump; no remaining stack transport to remove |
| `StoreSlotNew -> Pop -> LoadSlot` | Fixed 3.02%, Option 13.23%, Reverse 10.31%, Word 10.41% | Allocation/discard crosses identity-bearing slot semantics; unsafe to scalar-replace from adjacency |
| `LoadSlot -> LoadSlot -> LoadSlot` | Array 5.19%, Reverse 3.40%, Word 2.63% | Removes dispatch only unless a later consumer supplies type and lifetime proof; not an independently useful block |
| `Pop -> LoadSlot -> JumpIfNotNil` | Option 4.08%, Reverse 3.46%, Word 4.68% | Admitted: direct slot branch removes one load snapshot, stack round trip, and dispatch with an unchanged fallback |

This evidence rejects a general basic-block interpreter for now. The profitable
unit is the smaller semantic pair, and all exits remain in the existing
bytecode program.

## Preserved-binary A/B gate

The baseline binary was built after removing the census observer and before
the candidate (`b406e483ee6f05bc7c9d3681ffe8da64de297c3b3a2bec7b63310d122cfe63e0`).
The candidate binary contains only the slot-not-nil quickening relative to that
baseline (`b70d98ada7d7f58903f9591a810d005807ebc712f105b13cfe9e955d131c2def`).

Runs alternated baseline/candidate order by pair, used CPU 0,
`GOMAXPROCS=1`, canonical catalog inputs and working directories, and the
external stdlib. Every stdout capture passed its public Ruby verifier. Per the
workstation protocol, every sample and outlier is included in the arithmetic
mean; volatile or nearly neutral rows received additional pairs.

| Application | Samples per binary | Baseline mean | Candidate mean | Change |
| --- | ---: | ---: | ---: | ---: |
| Array Slice Window | 10 | 0.6310 s | 0.6160 s | -2.38% |
| Distance Field | 10 | 5.3830 s | 5.4730 s | +1.67% |
| Fixed Width 128 | 5 | 7.2260 s | 6.9940 s | -3.21% |
| Lexical Rollup | 20 | 0.4130 s | 0.4310 s | +4.36% |
| Option/Result Config | 10 | 0.7550 s | 0.7320 s | -3.05% |
| Reverse Complement | 5 | 6.2880 s | 6.0740 s | -3.40% |
| Word Frequency | 20 | 1.3920 s | 1.3850 s | -0.50% |

The three applications that admitted the exact window improve by 0.50%-3.40%.
Five of seven total guards improve. The two control regressions remain below
5% after expanded sampling, and their user-time changes agree with wall time
(+1.67% and +4.01%). The candidate therefore provisionally cleared this
bounded gate without discarding noisy samples, but did not clear the later
full selected-suite gate. All 160 individual verified rows, stdout hashes, and
timings are retained in
`2026-07-18-slot-not-nil-register-block-ab-samples.tsv`.

## Correctness and cleanup

- Focused quickening, direct/fallback, raw-i32, and typed-match guards pass.
- `go test ./pkg/interpreter -run 'TestBytecode' -count=1 -timeout 60s`
  passes in 24.293 seconds.
- The complete exec-fixture parity matrix passes when split into four bounded
  prefix partitions: 28.843s, 10.432s, 14.518s, and 21.954s.
- The monolithic package command reached its required 60-second cap while the
  fixture matrix was still progressing; the active fixture passed alone in
  0.976s. The timeout was not extended.
- The temporary census observer, incomplete census records, raw process
  profiles, A/B binaries, stdout/stderr captures, and runner are removed after
  this report. Only the eight complete compact census JSON files and the
  160-row timing ledger remain.
- No WASM work was performed.

## Follow-up

The recommended selected-bytecode refresh was completed. It rejected this
candidate, restored the exact pre-candidate binary, and left the promoted
scorecard unchanged. See
`2026-07-18-post-slot-not-nil-bytecode-scorecard-reconciliation.md` for the
full evidence and next recommendation.
