# Compiled Static-Builtin `print` Gate (2026-07-12)

## Decision

Rejected and reverted. The candidate removed the named dynamic-call and
native-adapter path for an unshadowed, one-argument `print` with a statically
native String, boolean, integer, float, or character argument. It emitted
direct Go stdout output while leaving every runtime, nullable, nominal,
interface, callable, and unknown value on the existing adapter.

The candidate preserved output semantics, but it is not a broad performance
win. It helped both output-heavy applications and one concurrency guard while
materially regressing three unrelated compiled applications. No compiler,
runtime, benchmark, fixture, or `able-stdlib` source remains changed by this
experiment.

## Method

Each baseline/candidate pair used the same source, canonical external stdlib,
benchmark input, verifier, `GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, CPU
2 affinity, 45-second process bound, and three independent compiled runs. The
only switch was `ablec -experimental-static-builtin-print`. Every row completed
and was verified three of three times. The raw per-variant summaries are in
the sibling `2026-07-12-compiled-static-print-gate/` directory.

| Application | Baseline | Candidate | Change |
| --- | ---: | ---: | ---: |
| I-Before-E | 0.1133s | 0.1100s | -2.9% |
| PiDigits | 1.5167s | 1.5000s | -1.1% |
| Fib | 3.4933s | 3.4800s | -0.4% |
| MatrixMultiply | 1.2000s | 1.2633s | +5.3% |
| JSON | 0.7867s | 0.8633s | +9.7% |
| Word Frequency | 0.2700s | 0.3500s | +29.6% |
| Channel Rollup | 2.0067s | 1.8967s | -5.5% |

The I-Before-E and PiDigits results validate the reachability finding: their
repeated ordinary output crosses the dynamic host adapter. They do not justify
making that crossing a compiler fast path. The JSON, MatrixMultiply, and Word
Frequency regressions exceed normal measurement noise and violate the broad
guard policy, so the prototype and its CLI option were removed rather than
left dormant.

## Next recommendation

Do not reopen Array/extern release or integer-box cache policy work. The
completed tagged reuse probe finds K-Nucleotide at 95.77% dynamic-cache hits,
Reverse Complement with 4.59M hits even after saturation, and zero dynamic
tier use in both controls. It rejects an adaptive/scoped policy premise; see
`2026-07-12-bytecode-dynamic-box-reuse.md`. The canonical runtime-value
architecture design was already completed and rejected a prototype; the later
normal-build opcode census also finds no remaining shared VM leaf. See
`2026-07-12-bytecode-opcode-census.md`. Resume feature completion rather than
another unchanged-source compiler or VM micro-tranche.
