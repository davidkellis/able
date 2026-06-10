# Strict Two-Cohort External Scorecard

## Decision

The first complete mode-aware scorecard refresh is valid and promoted. Two
independent cohorts used the same reviewed 58-row selection manifest, the same
64-row timeout-preserving full coverage, and the same canonical stdlib source
state. Every selected Able row and every applicable fresh Go/Python/Ruby
reference row retained exactly five successful samples. Cohort B is now the
current scorecard; the separate variance report preserves both cohorts.

No compiler, VM, stdlib, application, verifier, foreign reference source, or
selection-manifest source changed during collection.

## Protocol result

- Full status per cohort: 64 application/mode rows.
- Strict selection per cohort: 32 compiled plus 26 bytecode rows.
- Selected result: 58/58 verified in cohort A and 58/58 verified in cohort B.
- Excluded full-status result: bytecode Binary Trees, QuickSort, Sudoku Masks,
  N-Body, TapeLang Alphabet, and Regex Suffix Audit each timed out 5/5 in both
  cohorts.
- Manifest semantic SHA-256:
  `11caaee63c66fa2e235249640fe0ce44833dc6ed9946d7d0a1e840997345c132`.
- Canonical stdlib: 69 sources, tree SHA-256
  `785a6fd058c179379b1a153529fb340151a11b96d9014394cc40dbd87e1882ab`,
  Git head `219eff222c28406487231713753641bc49ee5b9a`, with the same dirty-state
  flag in both cohorts.
- Strict validation also proved disjoint source reports, identical coverage,
  identical manifest identity, and five verifier-backed samples for every
  selected Able/reference row.

The refresh ran sequentially with `GOMEMLIMIT=1GiB`, `GOGC=50`,
`GOMAXPROCS=1`, a 90-second per-process cap, and no quiet-CPU requirement.
This follows the workstation policy: independent repeated processes are
averaged, and variability remains visible rather than requiring an idle host.

## Distance from the targets

Using the two-cohort mean ratios and the 95%-of-reference threshold
(`Able/reference <= 1.0526`):

- compiled: 4 of 32 selected applications are at target—Base64, JSON, Monte
  Carlo Pi, and QuickSort;
- bytecode versus Python: 5 of 26 are at target—Base64, Fib, JSON, Matrix
  Multiply, and Pi Digits;
- bytecode versus Ruby: 4 of 26 are at target—Fib, JSON, Matrix Multiply, and
  Pi Digits.

The largest repeatable compiled misses include Fixed Width 128 (`1686.74x`
Go, Able CV `2.26%`), Channel Rollup (`274.89x`, `2.80%`), Rational Series
(`199.81x`, `3.74%`), Future Pipeline (`159.76x`, `3.40%`), Mutex Ledger
(`142.09x`, `4.35%`), and Mutex Await Journal (`133.72x`, `4.65%`).

The largest repeatable bytecode/Python misses include Regex Set (`327.36x`,
Able CV `7.03%`), Regex Stream (`276.25x`, `4.80%`), Reverse Complement
(`252.71x`, `4.22%`), Option/Result Config (`233.81x`, `4.46%`), Word
Frequency (`83.76x`, `3.36%`), and Rational Series (`42.93x`, `2.11%`).

## Variability review

Eight selected rows have Able CV above 15%: compiled Array Slice Window,
Dependency Plan, Document Audit, K-Nucleotide, Lexical Rollup, Regex Stream,
and Reverse Complement, plus bytecode JSON. Most are very short processes and
therefore sensitive to timer/startup noise. Compiled K-Nucleotide also moved
from 3.322 seconds in A to 4.518 seconds in B and must not drive a candidate
from its aggregate ratio alone.

The high-priority families above are not inferred from a single noisy ratio:
fixed-width/rational work, six distinct concurrency applications, and multiple
regex/text/nominal controls repeat large gaps with low Able CV. Existing
profiles still forbid reopening their broad parent costs or previously
rejected raw-value, float, map, return, startup, and global fixed-context
variants without new concrete evidence.

## Promotion and artifacts

Cohort B was promoted only after strict validation and review. The current
scoreboard replay check passes, and its selected status is 58 verified rows.

- `2026-07-15-strict-cohort-a-refresh.{json,md}`
- `2026-07-15-strict-cohort-b-refresh.{json,md}`
- `2026-07-15-strict-cohort-variance.{json,md}`
- `external-scoreboard-current.{json,md}`

All dated group reports and independent reference artifacts remain beside the
aggregates. Temporary cohort workspaces were removed automatically; only the
small durable reports remain.

## Follow-up resolution and next recommendation

The proposed selective-context direction was reconciled against retained
history before a new performance gate. Its first program-level formulation was
the already-rejected 2026-07-14 spawn-selected fixed-context candidate, which
regressed Mutex Ledger by 10.0%; the broader unchanged fixed-context descendants
are also closed by the retained ABI gates. No duplicate compiler code remains.

The next tranche should be feature-led: identify missing non-WASM spec-defined
coverage, establish cross-interpreter/compiler parity and a portable
verifier-backed application, then profile only if a concrete descendant repeats
across at least three unlike applications. The full rationale and the new
manifest-driven five-selected/one-status collection policy are recorded in
`2026-07-16-selected-samples-status-probes.md`.
