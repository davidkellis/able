# K-Nucleotide fresh interpreter-reference gate

## Decision

Keep no bytecode VM, compiler, canonical-stdlib, or benchmark-source
performance change. K-Nucleotide is now a fair Python/Ruby comparison, but its
large bytecode gap does not share a concrete map or call/return leaf with Word
Frequency and I-Before-E. In particular, do not add a FASTA, k-mer, `HashMap`,
raw-u64, or return-shape exception.

## Reference coverage

The sibling benchmark suite now supplies Python 3.14 and Ruby 4.0 reference
sources (and container recipes) for K-Nucleotide. Both use the same rolling
two-bit frequency-map algorithm as the Go and Able programs, read the shared
generated FASTA input, and are checked by the existing tolerant frequency and
exact fragment-count verifier.

With CPU `2`, a 45-second guard, and three independent verifier-checked
processes, the fresh reference means were:

| Runtime | K-Nucleotide |
| --- | ---: |
| Python 3.14.5 | 1.5418 s |
| Ruby 4.0.5 | 1.4080 s |

The Python and Ruby stdout hashes match each other and the checked-in solution.
Fresh Word Frequency, I-Before-E, and JSON reference rows were collected in
the same lane as controls.

## Able status and attribution

One verifier-backed current normal bytecode process completed in 40.19 s:
26.1x the fresh Python reference and 28.5x Ruby. Repeating a 40-second
process three times would consume the full bounded benchmark budget without
changing that unambiguous selection signal, so one independent 40.84-second
verifier-checked CPU capture was used for attribution.

Its 40.54 seconds of CPU samples are overwhelmingly real VM execution, not
startup:

| K-Nucleotide VM path | Cumulative CPU |
| --- | ---: |
| `runResumable` | 96.8% |
| `execCallOpcode` | 29.0% |
| `execCallName` / cached call name | 18.3% / 10.9% |
| `finishInlineReturn` | 18.1% |
| `execBinary` | 14.7% |
| raw-integer metadata | 4.8% flat |
| generic primitive HashMap key equality/hash | below 1% flat each |

The high-level call parent overlaps the current Word Frequency profile, but
not its material child: Word Frequency has string-keyed map/name-call work
(`hashMapFindEntryWithHash` 4.9% flat, 47.9 MB and 603k allocations per warm
main), while K-Nucleotide is rolling raw-u64 bitwise, inline return, and frame
work. I-Before-E's independent profile is member/slot dispatch rather than
either map lane.

This also reproduces the earlier aligned K-Nucleotide/WordCount review: their
small direct cached call setup overlapped, but K-Nucleotide's return coercion
and WordCount's ordinary program return path did not. The raw-cell and return
guard variants around those boundaries already failed broad benchmark guards;
they are not reopened by a shared parent percentage.

No `able-stdlib` source changed. Temporary generated inputs, reports, and CPU
profiles were removed after this evidence was recorded.

## Next recommendation

Add fresh verifier-backed Python and Ruby references for both Mandelbrot and
Monte Carlo Pi, then compare their bytecode warm profiles with K-Nucleotide as
an unrelated map/call guard. They are independently shaped numeric applications
and can establish whether a concrete float store/cast/branch leaf repeats
before any VM work is attempted. The work entails source and container recipes
in the sibling benchmark suite, pinned three-process reference/bytecode rows,
and bounded warm profiles; retain a change only when the same primitive
operation and caller context recur in both numeric programs and remain neutral
on the text/map controls.
