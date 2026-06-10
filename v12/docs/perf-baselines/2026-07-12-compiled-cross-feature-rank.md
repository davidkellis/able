# Compiled cross-feature rank (2026-07-12)

## Method

This source-aligned ranking refreshed three Go 1.26.4 reference runs and
three verifier-backed compiled-Able runs for five materially different
workloads. Every process was pinned to CPU 2 and used `GOMEMLIMIT=1GiB`,
`GOGC=50`, and `GOMAXPROCS=1`; the per-run timeout was 45 seconds. The
comparison consumed the fresh Go JSON through
`bench_compare_external --go-reference-json`, rather than using stored Go
ledger rows.

| Benchmark | Primary shape | Compiled Able (s) | Fresh Go (s) | Able/Go | Status |
| --- | --- | ---: | ---: | ---: | --- |
| I-Before-E | file, line search, string output | 0.1267 | 0.0664 | 1.91x | verified 3/3 |
| Reverse Complement | file, primitive bytes, output | 0.1367 | 0.0164 | 8.34x | verified 3/3 |
| K-Nucleotide | file, strings, generic counting | 4.4867 | 0.0589 | 76.17x | verified 3/3 |
| Mandelbrot | numeric floating-point kernel | 0.1900 | 0.0494 | 3.85x | verified 3/3 |
| Sudoku Masks | recursive arrays and search | 10.1333 | 0.5781 | 17.53x | verified 3/3 |

The measurement work directory was deliberately ephemeral. The harness report
recorded a valid verifier result and stdout hash for every compiled run.

## Attribution gate

The rank does not identify a repeated compiler helper. The large rows divide
across short file/byte startup and output (Reverse Complement), a generic
string/counting boundary (K-Nucleotide), a direct float kernel (Mandelbrot),
and recursive array/search work (Sudoku Masks). I-Before-E is a separate
line-search/string-output shape.

The retained K-Nucleotide compiled profile is useful only as a hypothesis for
the next comparison: it has material generic value-conversion/allocation
(`bridge.ToUint` and `runtime.convT`) beneath string conversion and nominal
map calls. It does not show that those costs are material in any other ranked
row. The existing completed audits also already reject a named `HashMap`,
FASTA, raw-map, or float special case. The static-external launcher is retained
from its prior broad A/B evidence; it does not establish a main-body leaf for
this scorecard.

## Decision

Keep no compiler, VM, benchmark, or `able-stdlib` source change. In
particular, do not specialize lowering for K-Nucleotide, `HashMap`, string
keys, Sudoku masks, or Mandelbrot. A rank orders work; it is not sufficient
evidence for a reusable optimization.

## Next recommendation

Refresh source-aligned compiled Able-versus-Go rows for Word Frequency beside
K-Nucleotide, retaining JSON or Base64 as an allocation/host-work control.
Why: Word Frequency is an independent ordinary text/counting program that can
test whether K's generic primitive-to-runtime conversion and allocation wall
recurs without making a named-container rule. The work entails fresh pinned,
verifier-backed Go and compiled rows first. Take bounded CPU and allocation
profiles only if both are material misses and the same conversion or runtime
helper is visible in both; any candidate must improve that primitive boundary
generically and remain neutral on the controls.
