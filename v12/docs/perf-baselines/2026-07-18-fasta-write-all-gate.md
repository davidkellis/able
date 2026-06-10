# FASTA / `write_all` breadth gate (2026-07-18)

## Decision

Retain one generic canonical-stdlib change: `able.io.write_all` now passes the
caller's byte array to the first host write and creates a suffix only after an
actual partial write. This removes an unconditional full-payload copy without
changing the partial-write, zero-write, error, empty-input, or buffered-writer
contracts.

The change cleared the predeclared breadth gate in three unlike, verifier-backed
bulk-output applications and improved the two applications with the largest
payloads in both generated-Go and bytecode modes. Two unrelated controls were
neutral-to-better after extra repetitions were used to average workstation
phase changes. No compiler, VM, runtime, nominal-lowering, or language special
case was added.

## Independent FASTA application

The new FASTA Generation application emits the conventional deterministic
three-record workload at `n = 100000`: repeated ALU data of length `2n`, IUB
random data of length `3n`, and Homo sapiens random data of length `5n`, with
the standard integer LCG and 60-byte line wrapping. Matched Able, Go, Python,
and Ruby sources produce exactly 1,016,745 bytes with SHA-256
`2907f3fb66fea247549c0f26b5b5d5cd1940a055574b72dad344283e1eb0fd10`.

Five fresh reference processes averaged 0.0258 seconds for Go, 0.2262 seconds
for Python, and 0.3106 seconds for Ruby. The pre-change Able means were 0.1260
seconds compiled and 3.2160 seconds bytecode in the initial application gate;
the immediately paired bulk-output baseline was 0.1280 and 3.4580 seconds.
All reference and Able outputs passed the byte-exact verifier.

FASTA is the 36th portable application and the 28th selected bytecode
application. It is selected in both product modes and is wired into the
`full`, `generality`, `coverage`, `text-bytes`, and new `bulk-output` suites.

## Attribution

The pre-change warmed FASTA bytecode trace contained 5,667,092 call-site hits.
The `able.io.slice_bytes_from_offset` loop contributed 3,050,237 hits, or
53.83%, by reading and pushing all 1,016,745 output bytes before the first host
write. This is the same owner previously measured at material shares in
Mandelbrot and Reverse Complement, satisfying the required three-unlike-
application rule.

After the change, the same helper contributes zero hits. The bounded warmed
runtime measurement moved as follows:

| Metric | Baseline | Candidate | Change |
| --- | ---: | ---: | ---: |
| Runtime | 3,782,318,327 ns/op | 1,922,112,866 ns/op | -49.18% |
| Allocated bytes | 158,832,048 B/op | 70,442,256 B/op | -55.65% |
| Allocations | 2,802,112 allocs/op | 1,817,091 allocs/op | -35.15% |

Trace instrumentation changes absolute runtime, so these figures establish
owner removal and allocation direction; the independent-process cohorts below
make the selection decision.

## Repeated process results

Every row used verifier-backed independent processes pinned to logical CPU 0.
The main bulk-output gate used five baseline and five candidate runs per
application/mode. Mandelbrot's bytecode samples were bimodal, so two additional
five-run pairs were collected and all 15 runs per side were averaged.

| Application | Mode | Baseline | Candidate | Change | Decision |
| --- | --- | ---: | ---: | ---: | --- |
| Reverse Complement | compiled | 0.1240 s | 0.1040 s | -16.13% | improve |
| Reverse Complement | bytecode | 6.5820 s | 4.2420 s | -35.55% | improve |
| FASTA Generation | compiled | 0.1280 s | 0.1120 s | -12.50% | improve |
| FASTA Generation | bytecode | 3.4580 s | 2.1660 s | -37.36% | improve |
| Mandelbrot | compiled, 10/side | 0.1470 s | 0.1380 s | -6.12% | improve |
| Mandelbrot | bytecode, 15/side | 6.4633 s | 6.5087 s | +0.70% | neutral |

The controls do not exercise this bulk-write copy. Their bytecode five-run
means were neutral-to-better: Distance Field moved from 5.740 to 5.716 seconds
(-0.42%), and Word Frequency moved from 1.558 to 1.478 seconds (-5.13%). Fast
compiled rows showed a host phase shift, so Distance Field was order-balanced
over 40 runs per side: 0.0955 seconds baseline versus 0.0935 seconds candidate
(-2.09%). Word Frequency's 20-run compiled means were 0.1965 versus 0.1920
seconds (-2.29%).

Raw samples and verification records are retained in the adjacent
`2026-07-18-*write-all*.json` and Markdown artifacts. No single process was
allowed to exceed 55 seconds.

## Correctness and scope

- The first attempt receives the original `Array u8`.
- If the host returns a positive partial count, later attempts receive newly
  allocated suffix arrays beginning at the exact accumulated offset.
- A zero or negative count still raises `BrokenPipe`; host errors still flow
  through `write`; empty input still performs no write.
- BufferedWriter continues to call the same `write_all` contract.
- The canonical `able.io` test suite passes in both Go interpreters, while all
  compiled and bytecode performance outputs above pass their public verifiers.
- FASTA also passes its public verifier in the tree-walker lane (17.96 seconds,
  below the per-process cap), completing three-runtime semantic coverage.
- No WASM work was performed.

## Next recommendation

Refresh the complete selected compiled and bytecode scorecards before choosing
another implementation candidate. The retained stdlib change materially moves
two existing rows, FASTA adds a new selected row in both modes, and the last
full product scorecards predate both facts. Five verifier-backed processes per
row will establish the current distance from the 95%-of-Go and 95%-of-
Python/Ruby goals and rank the remaining cross-application gaps.

After that refresh, profile the largest repeated compiled owner rather than an
individual slow application. The leading existing hypothesis is generic
generated backing-slice/capacity growth shared by Reverse Complement, Lexical
Rollup, and Array Slice Window, with Base64 as a competitive guard. Advance
only if the same general generated-code descendant is material in at least
three unlike applications; do not introduce an Array, stdlib-container, or
benchmark-specific lowering rule.
