# Compiled post-static-extern scorecard (2026-07-12)

## Method

This refresh evaluates the retained static-extern launcher change against
freshly rebuilt Go 1.26 references, rather than the stored external corpus.
Each Go reference and compiled Able row used ten verifier-backed launches on
CPU 2, with `GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and a 45-second cap.

| Benchmark | Compiled Able median (s) | Fresh Go median (s) | Able / Go | Valid runs |
| --- | ---: | ---: | ---: | --- |
| I-Before-E | 0.1450 | 0.0637 | 2.28x | 10/10 each |
| Reverse Complement | 0.1340 | 0.0167 | 8.02x | 10/10 each |
| JSON | 0.8010 | 1.6016 | 0.50x | 10/10 each |
| Mandelbrot | 0.1500 | 0.0526 | 2.85x | 10/10 each |

The preceding same-session A/B and exact bootstrap-allocation measurements
remain the retain basis for the launcher change. This broader scorecard is not
used to overstate a sub-10ms launch effect; it establishes the remaining
application-level gaps against fair, current Go references.

## Decision

Keep the static-extern launcher change and make no further code change in this
tranche. JSON remains twice as fast as its fresh Go reference. The material
misses have disjoint remaining costs:

- I-Before-E: text value/search and allocation work.
- Reverse Complement: a very short file/byte process whose generated `main`
  remains too short for useful CPU sampling; the retained startup reduction is
  the supported shared improvement.
- Mandelbrot: the direct floating-point escape kernel.

No concrete leaf repeats across the misses, so this is not evidence for a
text, FASTA/byte, floating-point, named-container, or benchmark-specific
lowering. `able-stdlib` needs no change.

## Next recommendation

Refresh fresh-Go compiled status for an independently shaped, main-dominated
pair such as QuickSort and Sudoku, retaining JSON and this scorecard as
controls. The post-launch misses above diverge, so repeating their profiles
would mostly produce noise. First establish pinned Go/Able rows; generate
main profiles only if both new rows are material misses and the same concrete
helper repeats.
