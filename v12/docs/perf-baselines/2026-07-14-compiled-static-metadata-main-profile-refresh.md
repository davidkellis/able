# Compiled static-metadata main-profile refresh — 2026-07-14

## Decision

Keep no compiler, generated-runtime, bytecode VM, canonical-stdlib, or
benchmark-source change. The retained static metadata rule removed its
reachable startup decoder work, but the refreshed user-main profiles do not
identify a repeated concrete leaf across the three unlike applications.

## Method

Each current generated binary was built from the same benchmark run directory
used by `bench_compare_external`, with the canonical external `able-stdlib`.
Every profile launch used CPU 15, `GOMEMLIMIT=1GiB`, `GOGC=50`, and
`GOMAXPROCS=1`. The CPU-only generated phase hook
`ABLE_GO_PHASE_CPU_PROFILE_DIR` wrote separate phases; this refresh merged only
`main`, without allocation snapshots or compiler/bootstrap samples. Profiles were merged with
`go tool pprof -proto`; every output was accepted by the benchmark's Ruby
verifier.

| Application | Verified launches | Main samples | Output SHA-256 |
| --- | ---: | ---: | --- |
| Word Frequency | 20 | 2.92 s | `7dc1dae393e2c070eb0b9c9e611e154b2e6cce1b4a4268aa1bc73f8ff0e2fd07` |
| Document Audit | 50 | 0 s | `0dad030a80c8a883cbb56fbcfebfd530d521075e15d5d91ba538bc93e66c0aab` |
| Lexical Rollup | 30 | 0.93 s | `a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604` |

Document Audit completes its user-main work below the CPU sampler's startup
resolution: all 50 phase files were written, but their merged `main` profile
has no samples. Its short normal process remains valid and verified; this is a
measurement limitation, not a timeout or semantic failure. The refresh does
not inflate its corpus or alter the benchmark source merely to create samples.

## Attribution

Word Frequency is map/text shaped. `__able_hash_map_find_entry` is 47.95% flat
and 49.66% cumulative. `String_split` is 34.25% cumulative, while allocation
and garbage-collection descendants account for the rest of the material work.

Lexical Rollup is file/iterator/allocation shaped. The compiled `read_lines`
path is 27.96% cumulative; `runtime.mallocgc` is 24.73%; `IteratorValue.Next`
is 16.13%; and `strings.genSplit` is 11.83%. Its material flat samples are GC
scan/write-barrier and slice-allocation work, not a map probe.

The tiny Document Audit bootstrap profile has no remaining
`DecodeNodeJSON` sample. Its few samples are registration/allocation noise and
cannot rank another static-startup candidate. The static metadata rule already
removed the only previously proven unreachable default-body payload; local
metadata and dynamic/bootstrapped semantics remain intentionally unchanged.

## Why no candidate is eligible

`__able_hash_map_find_entry`, `read_lines`, iterator traversal, and GC growth
are not the same leaf. Selecting one from one application would either create
a workload-shaped optimization or add a prohibited compiler lowering rule for
a nominal container. The source-visible `HashMap` probe may be investigated
only as a canonical-stdlib algorithm shared by independently authored map
programs; this refresh does not establish that evidence.

The temporary generated binaries and raw phase profiles were removed after
extracting these summaries. No `able-stdlib` change was required.

## K-Nucleotide follow-up

The recommended map follow-up completed with three CPU-15, output-verified
K-Nucleotide launches and 10.46 seconds of generated-main samples. Its output
SHA-256 was `d628623daa677c00673df8d4961d14eb271b4112cdcb65b8233f07b69d7b49b8`.

K-Nucleotide does call the exact generated `__able_hash_map_find_entry` helper
seen in Word Frequency, but it is only 10.04% cumulative (1.05% flat), rather
than Word Frequency's 49.66% cumulative. K-Nucleotide is instead dominated by
the surrounding primitive-key/value representation work: `runtime.mallocgc`
40.44% cumulative, `runtime.convT` 38.15%, raw HashMap get/set 28.78%/28.30%,
`primitiveHashMapKeyEqual` 7.74%, `primitiveHashMapHash` 4.97%, and
`bridge.ToInt` 16.54%. Lexical Rollup remains the non-map control and has no
material map-probe path.

This repeats the existing rejected representation trial rather than creating a
new candidate. The shared collision-safe entry index made a scaled large-map
control much faster but regressed real K-Nucleotide by 6%; its lazy variant
also regressed the small, high-frequency K-Nucleotide maps. Do not retry that
index or retune its threshold. The fresh profile confirms why: reducing only
the scan leaves the material primitive conversion/allocation path untouched.

## Next recommendation

Do not run another unchanged compiler map/profile tranche. The current portable
suite and active v12 roadmap expose no new cross-language performance boundary,
and the existing compiler, iterator, map, value-representation, and numeric
directions have all been selected or rejected with broad evidence. Keep the
verified scorecards as regression guards. The next eligible engineering work is
an unfinished semantic/portability boundary from the active roadmap, with
fixture parity first; performance selection resumes only when that work yields
a concrete leaf repeated across unlike real applications. This avoids turning a
known `HashMap`, corpus, key-type, iterator, or file shape into a synthetic
optimization target.
