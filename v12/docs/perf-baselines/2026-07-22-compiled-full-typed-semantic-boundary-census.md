# Compiled full-suite typed semantic-boundary census — 2026-07-22

## Decision

Close the compiled typed/runtime semantic-boundary layer on the current
49-application corpus. Retain no compiler execution, generated-runtime,
bridge, canonical-stdlib, application, benchmark, language, bytecode, or WASM
performance change.

Retain two general benchmark-harness corrections: every requested build phase
now observes `bench_perf --timeout`, and `bench_compiled_boundary_audit`
deletes each generated application tree after recording its row unless
`--keep` is explicit. The first correction prevents an `ablec -build` process
from outliving the advertised bound; the second keeps a full audit's disk use
bounded by one application rather than 49.

The governing diagnostic attempted all 49 portable compiled applications.
Thirty-one completed and passed their public verifiers; eighteen hit the
55-second build bound, with zero execution or verification failures. The
successful rows found a real three-unlike-application intersection, but five
normal, telemetry-free CPU-profile processes per application show that its
three category names are cumulative parents over different existing costs,
not a shared CPU leaf. No execution candidate was admitted, so diagnostic or
profile wall times are not promoted as A/B performance results.

## Governing bounded census

The run used one main-only `typed-boundary` diagnostic process per scorecard
row, `GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, the catalog execution
contract, public verifiers, and a 55-second per-phase/process bound. The raw
JSON SHA-256 is
`5364859d3fbbb48a436020f311ba6b3f099f2911171a279021f417f6e88b478d`.
Telemetry counts reach; it does not time the boundary and is never used as
wall-time evidence.

Successful-row totals are:

| Category | Count |
| --- | ---: |
| Any → runtime | 337,358 |
| Runtime → integer | 100,302 |
| Runtime → struct / struct → runtime | 33,779 / 30,047 |
| Runtime → union / union → runtime | 68,091 / 118,554 |
| Runtime → interface / interface → runtime | 10,248 / 8,704 |
| Runtime → callable / callable → runtime | 0 / 136,704 |
| Error → control / control → error | 3,182,673 / 8,958 |

Twelve successful applications had no measured crossing at all: Fib, Matrix
Multiply, QuickSort, Sudoku Masks, Base64, Distance Field, RMS Norm, Fixed
Width 128, Rational Series, Wide Integer Records, Array Slice Window, and
Dependency Plan. The raw JSON is content-addressed above and its durable
summary is retained in the companion JSON. The largest new cross-family
intersection is:

| Application | Any → runtime | Struct → runtime | Callable → runtime | Error → control |
| --- | ---: | ---: | ---: | ---: |
| Option/Result Config | 312,864 | 6,432 | 122,880 | 374,688 |
| Dependency Wave Validation | 24,492 | 20,530 | 4,096 | 83,915 |
| Concurrent Document Pipeline | 0 | 3,076 | 1,024 | 12,318 |

The eighteen explicit timeouts are Binary Event Log, Reverse Complement,
K-Nucleotide, TapeLang Alphabet, Word Frequency, Document Audit, Lexical
Rollup, Channel Rollup, Regex Suffix Audit, Regex Set Audit, Regex Stream
Audit, Log Routing Redaction, Config Validation Extraction, Concurrent Text
Index, Validated Job Pipeline, Concurrent Event Routing, Manifest
Normalization, and Policy Record Dispatch. No absence claim is inferred from
those rows. Their existing current profile/telemetry evidence remains part of
the reconciliation: K-Nucleotide, Event Routing, and Policy were measured in
the preceding typed-boundary tranche, while the text/map, regex, concurrency,
and post-ABI descendants have current closed dispositions in the performance
frontier.

Two non-governing attempts are excluded. The first exposed the missing build
bound when Binary Event Log generation ran for more than 200 seconds. A later
whole-row timeout was also rejected because it could interrupt verification
after a valid execution. The retained implementation bounds each build and
execution phase independently, so verification is not silently truncated.

## Normal-build CPU materiality gate

The three intersecting applications were rebuilt normally: no telemetry
variables, atomics, branches, or environment checks were emitted. Each used
one preserved binary and five independent verifier-backed main-only CPU
profile launches. The five profiles per application were pooled only for
attribution.

| Application | Pooled samples | Shared-parent evidence | Exact descendant |
| --- | ---: | --- | --- |
| Option/Result Config | 0.55 s | struct conversion 5.45% cumulative; control helper 1.82% flat | ordinary map allocation plus generic-union type matching |
| Dependency Wave Validation | 5.60 s | struct conversions up to 15.54% cumulative; callable wrapper 22.32% cumulative; both zero flat | `Runtime.StructDefinition/currentEnv`, then `currentGID` / `runtime.Stack` at 94.11% cumulative |
| Concurrent Document Pipeline | 1.69 s | struct conversions up to 13.61% cumulative; callable wrapper 54.44% cumulative; both zero flat | `Runtime.StructDefinition/currentEnv`, then `currentGID` / `runtime.Stack` at 95.27% cumulative |

Merged profile SHA-256 values are:

- Option/Result Config:
  `35779aa4ff24ed633cc6cd19193a4376859cb0a55c5b4e0ba64a4f0f65307633`
- Dependency Wave Validation:
  `98770d048cf1f5ac5f4c339ed6fdd4125d0ebbac035f0c4fa356e7e60f235126`
- Concurrent Document Pipeline:
  `ce7070f09d3acdb78eb384b3171743fa2c564f7407275f89776e116650ca175e`

`control_from_error` samples only in the serial Option/Result profile. The
callable conversion functions are zero-flat wrappers. Struct conversion is
also zero-flat in all three; it descends into ordinary allocation/type
matching in the serial application but environment lookup in the concurrent
applications. That concurrent leaf is precisely the already-rejected
goroutine-identity architecture wall, not evidence for another struct or
callable conversion fast path.

## Why no candidate was built

- The same counter category does not imply the same cost. Here the exact
  children split before a generic change can remove work from all three.
- Optimizing the zero-flat wrappers would merely move the same allocation,
  generic-union matching, or execution-context work under a different parent.
- The serial descendants reproduce the retained/rejected post-ABI and static
  type-match gates. The concurrent descendants reproduce the separately
  rejected execution-context variants.
- Timed-out rows cannot support absence evidence, but their earlier current
  evidence does not add a new exact descendant. Re-running them with longer
  than one-minute tests would violate the project guardrail and would not
  change the three-app profile split already observed.
- A named `Result`, `HashMap`, channel, document, wave, regex, or benchmark
  lowering would violate the shared nominal rule and would not address the
  corpus-wide target gap.

## Verification

- Governing census: 31/49 verified, 18/49 bounded build timeouts, zero failed
  executions and zero failed verifications.
- Profile gate: 15/15 normal telemetry-free launches verified.
- `bash -n v12/bench_perf v12/bench_compiled_boundary_audit` passes.
- Both scripts' help paths pass, and both remain below 1,000 lines.
- A Binary Event Log smoke proved that generated-source compilation now exits
  with status 124 at the requested bound and leaves no child process.
- Audit artifacts were removed row by row; no retained application tree is
  required to reproduce the checked summary.

## Next recommendation

Measure generated static-closure reach and size before implementing another
compiled runtime fast path.

Why: the dynamic boundary layer now has no shared exact CPU leaf, while this
full pass newly demonstrated that diagnostic builds for 18 diverse programs
cannot finish inside the one-minute project guardrail. Earlier generated-Go
diagnostics also found broad runtime, registration, conversion, and stdlib
bodies emitted whether or not a workload executes them. A generic
reachability reduction could reduce compiler work and binary/code footprint
across real applications without naming a container or benchmark; it must not
be assumed to improve runtime until linked size, startup, and main profiles
show that effect.

What it entails: census emitted generated functions, bytes, registrations,
and linked symbols against conservative main/dynamic/export reach for at least
three serial and three concurrent applications, including both timeout and
fast-build controls. Preserve every function reachable through dynamic calls,
interfaces, exports, package initialization, externs, and concurrency entry
points. Advance only if one generic unreachable closure repeats broadly and
at least three unlike normal binaries show a shared linked/startup or
instruction-cache cost. Then prototype conservative whole-generated-package
pruning and compare five complete verifier-backed processes per side, plus
dynamic-fallback and package-initialization correctness controls. Otherwise
close this compiler-scale direction and return to scorecard-driven bytecode
architecture selection. Continue to exclude named nominal/container rules and
WASM.
