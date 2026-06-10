# Compiled external threshold-control gate — 2026-07-13

## Decision

Keep the external benchmark system report-only. There are two unlike clear
inside controls—QuickSort and JSON—and a third, recursive structural control
that is correctly `boundary`. The repository has a non-timing evidence-integrity
check for these retained classifications. It verifies report provenance only;
it never runs a benchmark and does not make a commit fail for being slower.

The follow-on JSON protocol observed a 20.04% half spread, wider than the
initial control set. The provisional classification guard is therefore widened
from 12% to **21%**. This still classifies the clearly separated controls, but
it reinforces why a raw target or a median-only result must not become a
compiler threshold.

No compiler, VM, bridge, runtime, canonical-stdlib, or benchmark-source change
is selected by this measurement work.

## Protocol

Every control retains five independent one-process Go/Able samples. The
initial four-control aggregate matches freshly rebuilt Go and Able samples by
repeat index. The JSON and BinaryTrees follow-ons improve that protocol by
interleaving every pair: a freshly built and verified Go reference immediately
precedes a freshly built and verified Able application. Timed processes are
pinned to CPU 15, use `GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1`, have a
45-second cap, and check Able stdout with the benchmark's canonical Ruby
verifier. Generated Go build time is outside the timed process.

The controls intentionally span distinct application families:

| Control | Role | Family |
| --- | --- | --- |
| QuickSort | clear inside control | Array/index/algorithmic static code |
| JSON | clear inside control | text, file, and JSON codec work |
| BinaryTrees | structural boundary control | recursive nominal allocation and traversal |
| Base64 | near-boundary control | byte/string and host codec work |
| N-body | clear outside control | primitive numeric/package work |
| K-Nucleotide | clear outside control | text/map boxing and conversion |

The retained aggregates have individual samples, medians, range, and sample
CV: `2026-07-13-compiled-threshold-protocol.{json,md}`,
`2026-07-13-compiled-json-threshold-protocol.{json,md}`, and
`2026-07-13-compiled-binarytrees-threshold-protocol.{json,md}`. The temporary
per-pair source reports remain cleanup-eligible under their matching
`v12/tmp/perf/2026-07-13-compiled*-threshold-protocol` directories.

## Guard-band classification

The target ratio is `1 / 0.95 = 1.0526x`. The largest observed half spread is
20.04% (JSON), so this report uses a provisional **21%** guard band solely for
classification:

- inside only when every paired ratio is at or below `0.8316x`;
- outside only when every paired ratio is at or above `1.2737x`;
- otherwise boundary.

| Application | Go-ratio samples | Median | CV | Classification |
| --- | --- | ---: | ---: | --- |
| QuickSort | 0.64x, 0.69x, 0.79x, 0.74x, 0.74x | 0.74x | 7.90% | inside |
| JSON | 0.44x, 0.63x, 0.48x, 0.67x, 0.67x | 0.63x | 19.07% | inside |
| BinaryTrees | 0.81x, 0.93x, 0.92x, 0.92x, 0.99x | 0.92x | 7.09% | boundary |
| Base64 | 1.09x, 1.01x, 1.03x, 0.99x, 0.99x | 1.01x | 4.28% | boundary |
| N-body | 15.49x, 12.43x, 13.95x, 13.69x, 15.11x | 13.95x | 8.61% | outside |
| K-Nucleotide | 51.29x, 59.49x, 64.59x, 64.54x, 59.32x | 59.49x | 9.08% | outside |

QuickSort and JSON are comfortably inside even under the widened guard, while
both miss controls are comfortably outside. Base64 remains the important
boundary result: its median is below the raw target but one pair is 1.09x, so a
median-only or single-run threshold would misclassify it. JSON likewise shows
why a guard must follow the widest observed spread: its ratio range is wide,
but every sample remains safely inside. BinaryTrees supplies the corresponding
structural warning: its raw median meets the 95%-of-Go target, but four of five
samples lie above the guard's inside cutoff and it must remain boundary.

## Consequence and next gate

The `bench_external_threshold_controls --check` preflight now makes the
classification reproducible alongside the existing scoreboard check. It is not
a timing test and must remain non-enforcing. Do not turn either inside control
into a standalone compiler requirement, and do not select a runtime
optimization from it.

Do not add more compiled controls merely to obtain a favorable result. The
compiled set now spans Array/index, text/codec, and recursive nominal
allocation, and it identifies no shared removable compiler cost. The bytecode
JSON/PIDigits protocol is recorded separately in
`2026-07-13-bytecode-threshold-protocol-gate.md`; it supplies two clear
multi-reference bytecode controls but likewise identifies no VM candidate.
The next work is profile-led comparison across unlike bytecode target misses,
not another hunt for a favorable scorecard row.
