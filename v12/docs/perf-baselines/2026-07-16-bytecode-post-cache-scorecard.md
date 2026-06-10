# Bytecode post-cache scorecard and text-cluster gate

Date: 2026-07-16

## Decision

Keep no new runtime or stdlib code. The verifier-backed scorecard is refreshed
after the retained lambda and member-method caches, but the three highest-
impact text representatives do not share a new concrete VM descendant. Their
common low-level raw-integer leaf is the already-closed raw-integer family, and
their dominant CPU and allocation descendants otherwise diverge.

## Scorecard method

The reviewed bytecode selection manifest contains 26 portable applications.
Each ran in five independent complete processes with normal load and typecheck,
canonical `../able-stdlib`, source-root isolation where required,
`GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and a 55-second per-process cap.
No quiet-core requirement or CPU affinity was used; arithmetic means across
the independent processes are the workstation evidence.

All 130 launches completed. Every application had five successful runs, no
timeout or failure, five verifier passes, and one stable stdout SHA-256 across
its runs. Python/Ruby means below are the latest verified reference cohort in
the current scorecard; reference source hashes and benchmark input/verifier
contracts are unchanged. The Able rows are newly measured.

| Application | Current s | Prior s | Change | Able/Python | Able/Ruby |
| --- | ---: | ---: | ---: | ---: | ---: |
| Array Slice Window | 0.702 | 0.714 | -1.68% | 27.00x | 12.40x |
| Await Channel Mux | 0.208 | 0.252 | -17.46% | 1.76x | 2.24x |
| Base64 | 2.780 | 2.826 | -1.63% | 0.78x | 1.20x |
| Channel Rollup | 0.484 | 0.572 | -15.38% | 13.15x | 10.15x |
| Dependency Plan | 0.508 | 0.474 | +7.17% | 34.79x | 11.62x |
| Document Audit | 0.322 | 0.314 | +2.55% | 25.76x | 8.28x |
| Fib | 0.150 | 0.148 | +1.35% | 0.003x | 0.003x |
| Fixed Width 128 | 7.538 | 7.684 | -1.90% | 22.05x | 12.65x |
| Future Await Race | 0.164 | 0.158 | +3.80% | 5.58x | 3.29x |
| Future Pipeline | 0.460 | 0.496 | -7.26% | 7.76x | 6.68x |
| I-Before-E | 0.516 | 0.540 | -4.44% | 6.79x | 4.82x |
| JSON | 0.774 | 0.870 | -11.03% | 0.31x | 0.49x |
| K-Nucleotide | 38.320 | 40.602 | -5.62% | 31.54x | 32.09x |
| Lexical Rollup | 0.434 | 0.444 | -2.25% | 28.18x | 9.69x |
| Mandelbrot | 6.404 | 6.218 | +2.99% | 5.53x | 3.60x |
| Matrix Multiply | 4.450 | 4.646 | -4.22% | 0.09x | 0.10x |
| Monte Carlo Pi | 2.410 | 2.448 | -1.55% | 1.68x | 1.62x |
| Mutex Await Journal | 0.210 | 0.248 | -15.32% | 11.41x | 5.19x |
| Mutex Ledger | 0.364 | 0.684 | -46.78% | 12.21x | 7.60x |
| Option/Result Configuration | 0.772 | 3.358 | -77.01% | 47.65x | 18.92x |
| PiDigits | 2.204 | 2.246 | -1.87% | 0.58x | 0.23x |
| Rational Series | 3.984 | 4.014 | -0.75% | 41.37x | 31.72x |
| Regex Set Audit | 5.144 | 4.972 | +3.46% | 288.99x | 130.56x |
| Regex Stream Audit | 4.444 | 4.344 | +2.30% | 262.96x | 109.46x |
| Reverse Complement | 6.560 | 6.772 | -3.13% | 269.96x | 95.91x |
| Word Frequency | 1.418 | 1.446 | -1.94% | 81.97x | 29.98x |

The complete-process results confirm the retained caches are broad rather than
single-benchmark changes. Option/Result improved 77.01%, Mutex Ledger 46.78%,
Await Channel Mux 17.46%, Mutex Await Journal 15.32%, and Channel Rollup
15.38%. JSON also improved 11.03%. Small positive deltas remain within the
normal workstation/process band except Dependency Plan's +7.17%; that row is
not evidence for reverting changes whose allocation/runtime gates were already
performed with paired binaries, but it remains a future guard.

Only Base64, Fib, JSON, Matrix Multiply, and PiDigits currently meet or beat at
least one comparison interpreter; Base64 is still 1.20x Ruby and Monte Carlo
Pi remains 1.68x Python/1.62x Ruby. Most application families remain far from
the 95% interpreter target.

## Selection

Ranking by both absolute Able time and Python/Ruby ratio identifies a text
cluster:

- Regex Set Audit: 5.144 seconds, 288.99x Python, 130.56x Ruby;
- Reverse Complement: 6.560 seconds, 269.96x Python, 95.91x Ruby;
- Regex Stream Audit: 4.444 seconds, 262.96x Python, 109.46x Ruby;
- Word Frequency: 1.418 seconds, 81.97x Python, 29.98x Ruby.

Regex Set and Stream are related implementations, so they cannot alone satisfy
the unlike-program generality rule. The bounded profile gate selected Regex
Set, Reverse Complement, and Word Frequency: regex-set automata, bulk byte
transformation, and token/hash-map aggregation.

## Bounded profiles

Each representative was loaded and typechecked once, warmed once, then
profiled in a separate test process. CPU profiles used two measured calls for
Regex Set and Reverse Complement and five for Word Frequency, except Reverse
Complement was reduced to one call after two calls exceeded the 55-second
profile cap due profiling overhead. Exact allocation profiling used rate 1 for
Regex Set and Word Frequency. Reverse Complement rate 1 exceeded the cap, so
its bounded profile used rate 4096; its benchmark allocation counters remain
exact while profile proportions are sampled.

| Application | Profiled runtime | B/op | allocs/op | Dominant concrete CPU/allocation descendants |
| --- | ---: | ---: | ---: | --- |
| Regex Set | 4.153 s/call CPU | 120,674,148 | 1,733,997 | `execCallMemberArraySlot` 22.97% cumulative CPU; `arrayMember` 33.54% of sampled objects; match/thread structs 17.94%; string host results 17.12% |
| Reverse Complement | 6.011 s CPU | 705,449,656 | 10,894,564 | `execCallMemberArraySlot` 31.67% cumulative CPU; mono primitive Array shells 58.48% of sampled objects; integer boxing 41.45%; Array capacity growth 24.29% of sampled bytes |
| Word Frequency | 1.147 s/call CPU | 47,464,561 | 631,169 | generic calls 37.06% cumulative CPU; hash-map find 5.42%; string host results 22.77% of sampled objects; nominal result structs 22.15% |

The exact allocation count for Reverse Complement agrees with the prior
returned-nominal gate. Its large byte count is mono-u8 shell/boxing and backing
growth, not generic positional result structs.

## Reconciliation

The semantic label “text” does not survive concrete-stack reconciliation:

- Regex Set repeatedly converts regex/String/Array values and constructs
  automaton thread/match shells.
- Reverse Complement repeatedly transports primitive u8 Array values, boxes
  indices/bytes at stack boundaries, and grows output buffers.
- Word Frequency decodes Strings into nominal results and performs hash-map
  lookup/update calls.

`bytecodeRawIntegerValueInfo` appears in all three CPU profiles at 2.54%,
5.67%, and 2.97% flat. Its callers split among Array index/read, binary/
compare, coercion, and map/call work. This is the same raw-integer family that
the cross-family gate already tested with a general carrier interface and
rejected after Word Frequency regressed. Integer boxing is allocation-material
only in Reverse Complement; it is 2.76% of sampled Regex Set objects and 4.72%
of Word Frequency objects. Shared map frames likewise have unrelated owners:
ArrayStore handle maps in Regex/Reverse versus the language HashMap in Word.

No candidate meets the requirement of one concrete material descendant in all
three unlike programs. No benchmark, regex, corpus, String, HashMap, nominal
result, or primitive-Array special case was built.

## Verification and cleanup

The scorecard itself is the correctness gate: 130/130 processes passed their
external verifiers with stable per-application output hashes. No runtime,
compiler, canonical stdlib, benchmark, reference, verifier, or spec source was
changed. Temporary binaries, complete-process outputs, CPU/allocation profiles,
and comparison workspaces are removed after this record.

## Next recommendation

Build a coverage-wide, low-overhead CPU-symbol fingerprint census across the
same 26 selected bytecode applications, then cluster by exact concrete VM leaf
rather than by benchmark domain.

Why: ratio-based semantic clustering selected three “text” programs whose
actual costs diverged, while the scoreboard still has many large misses. The
project has also accumulated evidence closing raw integer, return, member-
validation, positional-result transport, scheduler, and several local Array
routes. A normalized symbol census can find a genuinely repeated remaining
leaf without manually forcing another application family or rerunning a closed
candidate.

What it entails: collect one bounded warmed CPU profile per selected
application in separate processes, normalize flat/cumulative time by measured
CPU samples, automatically group exact non-parent symbols appearing materially
in at least three unlike applications, and subtract the documented rejected
families. Only then collect allocation/trace evidence for the highest-impact
cluster and trial one generic candidate. Keep every process below 60 seconds,
use repeated verifier-backed A/B averages for any candidate, and do not begin
WASM work or add benchmark, stdlib-type, nominal-container, regex, or source-
shape special cases.
