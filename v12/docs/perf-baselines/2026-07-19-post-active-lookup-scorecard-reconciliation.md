# Post-active-lookup full scorecard reconciliation

Date: 2026-07-19

## Decision

Promote the complete post-active-lookup scorecard and retain no additional
compiler, VM, runtime, stdlib, workload, reference, verifier, language, or WASM
change from this measurement tranche.

The promoted cohort reports four of 36 compiled rows meeting the 95%-of-Go
target and two of 28 selected bytecode rows meeting both the 95%-of-Python and
95%-of-Ruby targets. Cross-cohort evidence keeps the bytecode result stable.
Compiled Fib and Base64 flip from the preceding full cohort, and older
independent controls also make Monte Carlo Pi volatile; the conservative
established compiled set is therefore Binary Trees, QuickSort, and JSON.

## Complete evidence

The reviewed manifest still contains 64 selected rows: all 36 compiled rows
and 28 bytecode rows. Eight additional bytecode rows retain fresh bounded
status. The refresh collected:

- 320 selected Able processes, five for every selected row;
- 180 fresh Go processes, five for every compiled row;
- 280 fresh Python/Ruby processes, five of each required language for every
  selected bytecode row; and
- one Able plus one Python and Ruby status probe for each of the eight excluded
  bytecode rows.

All 64 selected rows have exactly five successful Able and required-reference
samples. Every successful Able output passed its public verifier. Six excluded
Able bytecode probes timed out: Binary Trees, QuickSort, Sudoku Masks, NBody,
TapeLang Alphabet, and Regex Suffix. Fib and Matrix Multiply completed but
remain status-only. No timeout or partial row contributes to a target claim.

Each process used its catalog CPU budget from the allowed `0-15` pool,
`GOGC=50`, `GOMEMLIMIT=1GiB`, and a 55-second cap. The canonical stdlib state
contains 70 Able sources at tree SHA-256
`64b66a5b49cf3779912010d288ea0bcd0256c291dd58fe1bda705ee22dee6863`.
The selection manifest SHA-256 remains
`f46fccb1a156cb0574e4c0418683454feb8412b19089b778d7bc7820c57325b0`.

## Product result and variance

| Mode | Current cohort meets | Current applications |
| --- | ---: | --- |
| Compiled versus Go | 4 / 36 | Binary Trees, QuickSort, JSON, Monte Carlo Pi |
| Bytecode versus Python and Ruby | 2 / 28 | JSON, PiDigits |

The preceding full cohort reported six compiled meets. Fib changes from 0.90x
to 1.06x Go and Base64 from 1.02x to 1.10x. Both are now misses. Monte Carlo Pi
is 0.85x in the current cohort and 0.97x in the preceding cohort, but an older
independent five-run control measured 1.12x. Treat all three as threshold-
volatile. Binary Trees, QuickSort, and JSON are the established compiled meets.

Every selected bytecode classification is identical across the two complete
cohorts. JSON is currently 0.32x Python / 0.48x Ruby and PiDigits 0.67x / 0.27x.
Base64 is the nearest split result: it beats Python at 0.83x but misses Ruby at
1.27x, so it remains a product miss.

The strict two-cohort variance report retains ten successful Able samples and
ten samples for every required reference component. Large sustained bytecode
rows are stable enough to select profile families: K-Nucleotide has 3.16%
Able CV, Distance Field 3.14%, Regex Set 5.49%, Regex Stream 5.33%, Rational
Series 6.58%, and Reverse Complement 7.08%. Short concurrency and semantic
rows remain noisier, as expected on the workstation.

## Stable miss families

| Family | Current bytecode evidence | Existing exact attribution |
| --- | --- | --- |
| Text/map/graph | K-Nucleotide 32.48x/34.11x, Word Frequency 74.84x/22.17x, Dependency Plan 33.06x/11.59x | HashMap/String/hash/equality; HashMap find/String split; Queue/Deque/graph and checked integers |
| Regex NFA | Regex Set 215.76x/94.72x, Regex Stream 178.90x/69.07x | Array/member access and canonical NFA work; only two related applications and several retained/rejected NFA trials |
| Numeric/wide | Fixed Width 18.94x/11.93x, Rational 37.91x/29.55x, Distance 10.07x/18.03x, RMS 5.47x/8.93x, Mandelbrot 5.22x/3.14x | wide nominal, i128, float geometry, raw-float, and loop owners split; prior carrier/typed-block gates are closed |
| Bytes/output | Reverse Complement 179.38x/61.05x, FASTA 10.97x/7.30x, Base64 0.83x/1.27x | byte-array transport, bulk output, and codec work; `write_all`, conversion, and capacity gates already landed or closed |
| Concurrency | six selected rows miss by 2.39x-15.11x Python and 2.84x-11.32x Ruby | scheduler, channel, future, and mutex work forms a separate runtime cluster |

Ratio magnitude alone does not reopen rejected return/frame, raw carrier,
typed-block, Array-growth, String/byte, or regex representation candidates.

## Next recommendation

Profile the current text/map/graph cohort: K-Nucleotide, Word Frequency,
Dependency Plan, and Document Audit.

Why: the first three are unlike bioinformatics, text-frequency, and graph
applications with stable large misses and a plausible shared map/hash/equality
boundary. Document Audit is a text-heavy non-map discriminator. This is the
strongest remaining way to determine whether the aggregate Go-map and call
parents hide one generic Able operation or merely combine different cache,
environment, and container owners.

Use one ordinary verifier-backed process for K-Nucleotide and bounded repeated
main calls for the shorter programs. Attribute flat CPU and sampled allocation
cost to exact `HashMap.find`, hash/equality dispatch, String conversion/split,
member/name lookup, and Go map leaves by their Able owner. Admit a candidate
only if the same exact language/kernel or stdlib operation is material in at
least three programs. Any compiler work must use shared nominal lowering, not a
named-container fast path. Guard Document Audit, Array Slice Window, JSON, and
the current target meets with repeated alternating averages. Continue to defer
WASM.

## Artifacts and checks

- Promoted aggregate: `2026-07-19-post-active-lookup-scorecard-refresh.json`
  and `.md`.
- Strict two-cohort variance: `2026-07-19-post-active-lookup-scorecard-variance.json`
  and `.md`.
- The checked-in current scoreboard matches the dated aggregate byte-for-byte.
- The selection protocol tests, five-sample evidence check, and current
  scoreboard synchronization check pass.
