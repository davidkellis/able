# Post-`write_all` full scorecard reconciliation (2026-07-18)

## Decision

Promote the fresh 36-application scorecard and retain no additional compiler,
VM, runtime, stdlib, application, reference, verifier, or language change from
this measurement tranche.

The promoted snapshot reports six of 36 compiled rows meeting the 95%-of-Go
target and two of 28 selected bytecode rows meeting both the 95%-of-Python and
95%-of-Ruby targets. An independent five-run cohort for the four compiled rows
near the threshold shows that Fib and Monte Carlo Pi still flip with workstation
load. A conservative cross-cohort interpretation therefore treats four
compiled rows and two bytecode rows as established passes.

## Complete evidence

The expanded catalog contains 36 portable applications and 72 full-status
application/mode rows. The reviewed manifest selects all 36 compiled rows and
28 bytecode rows; the other eight bytecode rows retain one bounded status probe.

The refresh collected five independent, verifier-backed Able processes for
every selected row, five fresh Go processes for every compiled row, and five
fresh Python and Ruby processes for every selected bytecode row. That is 320
selected Able executions and 460 selected reference executions. The aggregate
evidence checker accepted all 64 selected rows with five successful samples per
required lane. There were no output-verifier or correctness failures.

Every application process had a 55-second cap. The excluded bytecode status
rows are Fib, Binary Trees, Matrix Multiply, QuickSort, Sudoku Masks, NBody,
TapeLang Alphabet, and Regex Suffix. Fib and Matrix Multiply completed their
Able probes; the other six timed out. These remain visible as status, never as
partial performance evidence.

The canonical stdlib fingerprint is 70 Able sources at tree SHA-256
`64b66a5b49cf3779912010d288ea0bcd0256c291dd58fe1bda705ee22dee6863`.
The selected-row manifest SHA-256 is
`f46fccb1a156cb0574e4c0418683454feb8412b19089b778d7bc7820c57325b0`.

## Product result

The promoted single-cohort result is:

| Mode | Promoted target result | Meets |
| --- | ---: | --- |
| Compiled versus Go | 6 / 36 | Fib, Binary Trees, QuickSort, Base64, JSON, Monte Carlo Pi |
| Bytecode versus Python and Ruby | 2 / 28 | JSON, PiDigits |

The independent near-threshold cohort used fresh Go and Able processes:

| Compiled application | Cohort A Able/Go | Cohort B Able/Go | Classification |
| --- | ---: | ---: | --- |
| Fib | 0.90x | 1.06x | volatile |
| Binary Trees | 0.94x | 0.72x | established meet |
| Base64 | 1.02x | 0.97x | established meet |
| Monte Carlo Pi | 0.97x | 1.12x | volatile |

QuickSort and JSON remain comfortably inside the compiled target and were
stable in the previous independent cohorts. The conservative current conclusion
is therefore four established compiled passes (Binary Trees, QuickSort, Base64,
and JSON), two volatile compiled passes, and two established bytecode passes.

The nearest selected misses are compiled Matrix Multiply at 1.22x Go and
PiDigits at 1.34x Go, plus bytecode Base64 at 0.75x Python but 1.18x Ruby.
Ratios for very short programs expose a broad launch/runtime floor: many
compiled semantic and concurrency applications take 0.1-0.9 seconds while
their Go references take 0.004-0.015 seconds. That is real product cost, but
the prior interpreter-package boundary experiment is closed: removing the
unused initializer improved startup yet regressed allocation-heavy Binary
Trees through changed GC pacing. Do not retry that representation or compensate
with heap ballast or GC-policy changes.

Large sustained misses remain structurally diverse. K-Nucleotide, Fixed Width,
numeric loops, regex NFA programs, concurrency programs, and short nominal/text
programs do not share one already-admitted inner helper. A ratio ranking alone
therefore does not authorize a candidate.

## Next recommendation

Run a cross-mode primitive-Array capacity/backing-growth profile gate across
Base64, Reverse Complement, FASTA Generation, Lexical Rollup, and Array Slice
Window.

This is the strongest general next direction because it spans unlike encoding,
bioinformatics, text aggregation, and slicing applications; the existing
compiled hypothesis already repeats generated backing-slice growth in Reverse
Complement, Lexical Rollup, and Array Slice Window; FASTA adds an independent
large-output consumer; and Base64 is both a compiled guard and the closest
selected bytecode miss, only 18% behind Ruby while already beating Python.

The tranche should collect preserved generated-main CPU/allocation profiles and
warmed bytecode CPU/allocation profiles under the same source fingerprints.
Normalize evidence to exact generic Array capacity/backing-growth descendants,
not broad `growslice`, GC, VM-dispatch, or application parents. Build a candidate
only if one language/kernel Array rule is material in at least three unlike
applications and preserves nominal/identity semantics. Guard JSON, Monte Carlo
Pi, Word Frequency, and allocation-heavy Binary Trees. Do not add Base64,
bioinformatics, stdlib-container, or benchmark-specific lowering, and continue
to defer WASM.

## Artifacts and checks

- Promoted aggregate:
  `2026-07-18-post-write-all-selection-refresh.json` and `.md`.
- Independent threshold evidence:
  `2026-07-18-post-write-all-near-threshold-go-b.json` and
  `2026-07-18-post-write-all-near-threshold-compiled-b.json`.
- The checked-in current scoreboard is regenerated from the same explicit
  source reports and passes the five-sample evidence check.
