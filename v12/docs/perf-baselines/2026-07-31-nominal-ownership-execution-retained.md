# Nominal ownership execution retained

Date: 2026-07-31

## Decision

Retain the opt-in generated execution path selected by
`ablec --experimental-nominal-ownership`.

The path consumes the retained interprocedural ownership proof and routes a
proven fresh nominal successor into caller-owned Go storage. Normal compilation
remains unchanged when the flag is absent.

This is a general structural compiler rule. It contains no application,
benchmark, container, source-family, or non-primitive nominal name check. No
runtime, interpreter, bytecode VM, stdlib, language, dependency, frozen
workspace, or WASM source changed.

## Execution boundary

An owned-result variant is emitted only for a callable with one proven fresh
successor summary. At an eligible replacement site:

- a direct successor is copied into the caller's existing nominal storage;
- a complete native-interface implementation set dispatches through one
  generated type switch whose concrete cases use the same owned-result ABI;
- a successor embedded at a one-field result path reconstructs the outer
  nominal value field-by-field around caller-owned successor storage; and
- nested successor calls use a local Go value slot until the enclosing result
  is complete, so later reads through the old source observe the old value.

The field-by-field outer reconstruction is required for Go escape analysis.
Copying an already-built outer value also copied its temporary pointer before
overwriting the field, so Go conservatively moved the temporary successor to
the heap. Constructing the returned outer value with the caller-owned pointer
present from the start removed that escape without weakening Able identity.

Parameter-origin, dynamic, capture/storage, retained-alias, returned-alias,
conditional/nonstraight, incomplete-interface, and deeper embedded-path cases
remain on fresh allocation.

## Five-application performance gate

Every baseline, candidate, and Go process produced the same application-specific
output hash previously accepted by its public Ruby verifier. Sixty rotating
baseline/candidate/Go triplets per application yielded 900 verified timing
processes on CPUs 12-15 with `GOMAXPROCS=4`, `GOGC=50`, and
`GOMEMLIMIT=1GiB`.

| Application | Baseline | Candidate | Go | Candidate change | Candidate / Go |
| --- | ---: | ---: | ---: | ---: | ---: |
| Concurrent Graph Visitors | 0.010099 s | 0.009300 s | 0.002967 s | -7.90% | 3.134x |
| Concurrent Packet Codecs | 0.011101 s | 0.008832 s | 0.002522 s | -20.44% | 3.502x |
| Concurrent Tree Folds | 0.007804 s | 0.006959 s | 0.002603 s | -10.83% | 2.674x |
| Concurrent Audio Voices | 0.014986 s | 0.013815 s | 0.002512 s | -7.81% | 5.499x |
| Concurrent Scene Tiles | 0.008077 s | 0.007218 s | 0.002117 s | -10.63% | 3.410x |

These applications are only 5-15 ms per launch, so a preliminary
30-triplet cohort was noisy and included one small regression. The final
60-triplet fully rotated cohort improved all five directions.

The result is material but does not close the project target: these candidates
remain 2.674x-5.499x slower than equivalent Go on this launch-dominated set.

## Exact allocation gate

Five lightweight main-phase `MemStats` runs per baseline and candidate all
passed their verifier-backed output contract.

| Application | Baseline bytes / objects | Candidate bytes / objects | Change |
| --- | ---: | ---: | ---: |
| Concurrent Graph Visitors | 1,202,038 / 46,907 | 803,896 / 38,698 | -33.12% / -17.50% |
| Concurrent Packet Codecs | 2,409,574 / 49,410 | 567,096 / 16,620 | -76.46% / -66.36% |
| Concurrent Tree Folds | 929,992 / 24,779 | 536,990 / 16,591 | -42.26% / -33.04% |
| Concurrent Audio Voices | 3,689,877 / 98,546 | 1,591,517 / 65,775 | -56.87% / -33.25% |
| Concurrent Scene Tiles | 1,197,830 / 37,089 | 406,584 / 24,784 | -66.06% / -33.18% |

The exact counters establish the intended heap removal independently of noisy
short-process timing.

## Artifact cost

The five candidates emitted 8-16 owned variants each. Generated source grew
0.465%-0.963%, and final binaries grew 0.197%-0.295%. Ordinary non-opt-in
generated products remain byte-for-byte free of the ownership helpers.

## Strict census

The strengthened census generated all 66 selected compiled applications with
`--no-fallbacks --experimental-nominal-ownership`.

- 66/66 generated successfully; zero failed.
- 66/66 final Go dependency graphs resolved with `go list -mod=mod -deps`.
- Zero dependency graphs contained `able/interpreter-go/pkg/interpreter`.
- Generation ranged from 356 ms to 11.892 seconds, with a 3.753-second mean.
- Disposable output totaled 288,134,658 bytes, 7,082,742 Go lines, and 3,316
  files before every row module was deleted.

The raw aggregate SHA-256 was
`4fd7450ef59d3b6809099e8dc1230ac9923f4975d60b44c6cc2153ca2af9b7c1`.
The compact machine-readable companion is
`2026-07-31-nominal-ownership-execution-retained.json`.

## Verification

- focused ownership execution, identity, caller-owned-result, and loop guards:
  5.055 seconds;
- interface lookup, native/generic interface, imported/shadowed alias,
  Result, and Option guards: 39.630 seconds;
- `go test ./cmd/ablec`: 5.283 seconds;
- `go test ./cmd/able-generated-boundary-census`: 0.002 seconds;
- five exact public verifier contracts across baseline, candidate, and Go;
- 900/900 balanced timing processes verified;
- 50/50 allocation-counter processes verified; and
- final 66/66 strict generation and dependency census.

No individual test or application generation exceeded one minute. The active
toolchain was Go 1.26.5 on linux/amd64.

## Cleanup

Three superseded prototype workspaces totaling about 2.0 GiB were removed
during development. After compact evidence publication, the final prototype,
three census workspaces, and stale extern-Go cache totaling 3,571,620 KiB were
also removed. No ownership task directory remains under `/var/tmp`, and no
Able task directory remains under `/tmp`.

## Next recommendation

Promote the proven ownership execution rule to ordinary compiled builds while
retaining an explicit diagnostic opt-out.

Why: the opt-in path improved all five unlike applications and passed all 66
strict generation and dependency rows, but ordinary users do not receive the
gains yet.

What it entails: invert the option plumbing, preserve a baseline opt-out,
rerun identity and interface guards, execute a broad public-verifier subset,
repeat the 66-row strict dependency census, and refresh the affected compiled
frontier rows.

Why it matters: this converts a validated 33%-76% allocation-byte reduction
and 8%-20% timing gain into normal compiled Able performance while preserving
a controlled comparison route and the same fail-closed semantic boundary.
