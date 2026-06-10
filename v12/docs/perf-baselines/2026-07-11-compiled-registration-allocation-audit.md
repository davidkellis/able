# Compiled registration allocation audit (2026-07-11)

## Purpose

After the rejected lazy small-integer cache change, this tranche re-measured
the generated launcher registration boundary for the current I-Before-E, JSON,
and ReverseComplement applications. It used the real suite directories,
verifier-checked output, CPU `2`, `GOMEMLIMIT=1GiB`, `GOGC=50`, and
`GOMAXPROCS=1`. Phase allocation mode is attribution only; it forces exact
allocation sampling and is not used as a timing result.

## Current shared boundary

Three verified phase-allocation launches per application put bootstrap in a
narrow common band:

| Application | Bootstrap bytes/launch | Allocations/launch |
| --- | ---: | ---: |
| I-Before-E | 4,214,139 | 12,621 |
| JSON | 4,229,512 | 12,867 |
| ReverseComplement | 4,232,179 | 12,691 |

Merged exact allocation-profile diffs contain profiler/serialization overhead,
which is excluded from source selection. The repeated production descendants,
per launch, are `newInterpreter` (about 584 KiB cumulative),
`DecodeNodeJSON` (about 229 KiB), diagnostic origin registration (about
157 KiB), and compiled-package registration (about 528 KiB cumulative).

The first two are not new candidates: the prior phase work already rejected
lazy constructor-cache sizing and both alternative metadata codecs on broad
application guards. The current generated inventories make the remaining
diagnostic allocation concrete:

| Application | Packages | Diagnostic nodes | Compiled calls | Compiled methods | Definition decode calls |
| --- | ---: | ---: | ---: | ---: | ---: |
| I-Before-E | 12 | 963 | 131 | 134 | 28 |
| JSON | 13 | 989 | 136 | 144 | 28 |
| ReverseComplement | 12 | 1,021 | 135 | 134 | 28 |

Every launcher knew the complete diagnostic-node count before it inserted the
first origin into the interpreter map. The old generated loop let that internal
map grow repeatedly.

## Retained generic change

Generated `__able_register_diag_nodes()` now tells the bridge the exact count;
the interpreter pre-sizes its diagnostic-origin map once, then keeps the same
span and origin registration for every node. The reservation is inert when a
static no-bootstrap launcher has no interpreter. It is an implementation-wide
diagnostic metadata improvement, not a language-container or benchmark rule.

On I-Before-E, a verified candidate phase allocation snapshot reduced bootstrap
from the 4.21 MB / 12.62k-allocation baseline mean to 4,132,976 bytes and
12,562 allocations. `AddNodeOrigin` drops out of the candidate's exact
allocation diff; its map storage is allocated once at the known capacity.

## Process guard

Every measured process passed its public verifier. Baseline and candidate were
interleaved at the same CPU/memory/GC settings.

| Application | Baseline | Candidate | Runs per side |
| --- | ---: | ---: | ---: |
| I-Before-E | 0.0960 s | 0.0940 s | 10 |
| ReverseComplement | 0.0900 s | 0.0900 s | 10 |
| JSON | 0.7120 s | 0.7140 s | 5 |

The short-application timing movement is deliberately treated as
neutral-to-small because the host timer resolves to 10 ms. The retain decision
rests on the exact shared allocation reduction plus no verifier or process
regression across the independent controls, not on the 2.1% I-Before-E point
estimate.

## Verification

- Generated diagnostic registration source test.
- Interpreter origin-reservation preservation test.
- Compiler execution harness.
- Compiler diagnostic parity for division-by-zero and unhandled raise fixtures.
- Interpreter runtime-diagnostic tests.

No `able-stdlib` change was needed.

## Next recommendation

Refresh the fair pinned compiled scorecard and application profiles after this
retained registration change, then select a new candidate only from a material
post-bootstrap leaf shared by two target misses. The remaining measured
bootstrap descendants are either already rejected (metadata decoding and
constructor cache timing) or too small to justify another map/capacity tweak.
The refresh should include the verified I-Before-E, ReverseComplement, JSON,
and Mandelbrot lanes and preserve bytecode as an independent non-regression
control; do not pursue more registration micro-optimizations without a larger
repeated source cost.
