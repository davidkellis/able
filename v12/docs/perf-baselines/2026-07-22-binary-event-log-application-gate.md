# Binary event log application gate — 2026-07-22

## Decision

Retain the non-concurrent `binary_event_log` application, its deterministic
binary fixture and source-equivalent Go/Python/Ruby implementations, catalog
memberships, repeated comparison cohorts, and bounded CPU profiles. Retain no
compiler, generated-runtime, bytecode VM, canonical-stdlib, language, or WASM
change.

The application adds real binary file input, nominal record construction,
`Result` success/error handling, interface dispatch, captured scoring logic,
arrays, and hash-map aggregation. Its current profiles repeat important generic
families, but the compiler leaf has breadth two rather than the required three,
and the bytecode leaf belongs to an already rejected broad candidate family.

## Application contract

The fixture begins with an eight-byte magic value and a fixed header, followed
by 64 variable-length records. Each record contains a timestamp, source,
severity, kind, payload length, payload bytes, and checksum. Deliberately bad
checksums, invalid severities, and empty payloads exercise error paths. The
workload rereads and processes the fixture for 1,024 normal rounds; the profile
variant uses 4,096 rounds to amortize parser and startup cost.

The normal result is:

```text
65536:53248:12288:22:13312,14336,12288,13312:648192:26622961321:945064
```

The profile result is:

```text
262144:212992:49152:22:53248,57344,49152,53248:2592768:106482112317:476207
```

Able tree-walker, Able bytecode, Able compiled, Go 1.26, Python 3.14, and Ruby
4.0 produce the corresponding exact output through the same verifier. The
catalog passes `events.bin` as a real program argument and retained comparison
rows fingerprint the fixture, verifier, and language sources. The fixture is
1,547 bytes with SHA-256
`b1a8e4b24ab6f0ae169112d22fc9fe44ae5c604bfda22168084d34ab861bc9`.

Exploratory 16- and 128-round scales were discarded before retained reporting
because compiled measurements were dominated by process startup. They are not
included in any mean. The 1,024-round contract keeps tree-walker correctness
well below the one-minute project cap while giving the compiled body enough
work to measure.

## Coverage result

The application is portable and belongs to ten feature families: lexical
bindings and patterns; nominal types, generics, and unions; expressions,
arrays, text, and files; closures and callables; control flow; interfaces and
dispatch; `Option`/`Result`/exceptions; packages and imports; stdlib protocols;
and program entry. It does not claim inherent-method or concurrency coverage.

The checked catalog now contains 49 portable applications and 91 selected
performance rows: 49 compiled and 42 bytecode. The three-family interaction
frontier retains minimum depth three; the high-density expressions/files ×
closures × `Result` interaction rises from five to six current applications.

## Repeated measurements

Two independent five-process cohorts were retained for every timed lane. The
pooled means below include every successful workstation sample; no outlier was
removed.

| Lane | Processes | Pooled mean | Limiting ratio |
| --- | ---: | ---: | ---: |
| Able compiled | 10 | 0.718000 s | 57.750× Go |
| Go | 10 | 0.0124329915 s | — |
| Able bytecode | 10 | 7.881000 s | 34.122× Python / 25.019× Ruby |
| Python | 10 | 0.2309635655 s | — |
| Ruby | 10 | 0.3149952737 s | — |

The promoted second cohort also retains all five samples: compiled 0.688
seconds versus Go 0.0096 seconds (71.667×), and bytecode 6.896 seconds versus
Python 0.2385 seconds (28.914×) and Ruby 0.3194 seconds (21.590×). Both pooled
and promoted results are clear target misses.

## Ownership and admission

Three verified compiled profile processes merge to 6.98 seconds of CPU samples.
`runtime.mallocgc` is 51.15% cumulative, and the generated general nominal
conversion for `EventRecord` is also 51.15% cumulative. Static generic-union
method dispatch is 29.80% cumulative, with the generic fast callable path at
24.93%. The union-dispatch mechanism is also material in Option/Result Config,
so its exact breadth is two. That is not enough to admit a candidate, and a
special case for `EventRecord` would violate the general nominal-lowering rule.

Three verified bytecode profile processes merge to 134.45 seconds of CPU
samples. `runResumable` is 82.62% cumulative; named descendants include call
opcode dispatch at 31.06%, member calls at 20.77%, binary operations at 9.95%,
typed-pattern matching at 6.39%, and raw integer extraction at 2.20% flat. The
raw-integer leaf now repeats at material weight across at least six unlike
applications, but prior generic extractor/carrier candidates failed broad
guards. The remaining type-match and Go-map items are cumulative parents whose
concrete consumers diverge. This evidence does not invalidate those closures.

## Verification

- exact output parity across both interpreters, compiled Able, and all three
  reference languages;
- two verifier-backed five-process cohorts per timed lane, averaged without
  deleting volatile samples;
- explicit source, verifier, argument, and input-file fingerprints;
- three verified compiled and three verified bytecode profile processes;
- catalog, selection, coverage, pair-, and triple-interaction checks;
- scoreboard, performance-frontier, and evidence-ledger checks;
- all new source files remain below 1,000 lines;
- `git diff --check`.

## Next recommendation

Run a report-only compiled allocation-shape census across Binary Event Log,
Option/Result Config, and at least one unlike nominal/`Result` application before
building another candidate.

Why: this tranche found a large general nominal-conversion/allocation wall and
raised static generic-union dispatch to exact breadth two. A third unlike
application would make the mechanism eligible for a general compiler/runtime
design; failure to repeat would close it without risking a nominal-type-specific
optimization.

What it entails: preserve current binaries, collect bounded allocation and CPU
evidence for three or more verifier-backed applications, separate general
nominal conversion from `Result` method dispatch and GC ancestry, and identify
the exact generated rule shared across them. Only then prototype a shared
nominal translation or semantic-encoding improvement, with unrelated compiled
target guards and bytecode workloads protecting against regressions. Do not
begin WASM work.
