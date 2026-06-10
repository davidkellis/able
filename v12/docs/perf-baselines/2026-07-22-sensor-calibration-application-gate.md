# Sensor calibration application gate — 2026-07-22

## Decision

Retain the portable, non-concurrent `sensor_calibration` application, its
deterministic text fixture, source-equivalent Go/Python/Ruby implementations,
catalog and coverage memberships, repeated comparison cohorts, and bounded CPU
profile evidence. Retain no compiler, generated-runtime, bytecode VM, canonical-stdlib,
language, or WASM change.

The application adds a validation-and-transformation workload over signed
numeric text rather than another scheduler or string-normalization workload.
Its profiles repeat already-closed generic string, array, type-match, stack,
and raw-integer costs; they do not expose a new exact shared leaf that meets the
three-unlike-application admission rule.

## Application contract

The fixture contains 32 pipe-delimited sensor records. The workload validates
sensor identifiers, kinds, signed raw values, offsets, scales, and quality
markers; returns nominal `CalibrationError` values through the `Error`
interface; constructs valid nominal `SensorReading` values; then applies
inherent validation, acceptance, bucketing, calibration, and checksum methods.
Accepted, dropped, and invalid paths all materially change the final result.

The normal 256-round result is:

```text
8192:6144:1024:1024:2048,2048,2048:3054096921:398181
```

The bounded one-round profile result is:

```text
32:24:4:4:8,8,8:10180651:272972
```

Able tree-walker, Able bytecode, Able compiled, Go 1.26, Python 3.14, and Ruby
4.0 produce the same verified normal result. The catalog passes `readings.txt`
as a real program argument. The benchmark contract SHA-256 is
`0dc288da992f3174fb68c03ffd3e0837fa168186947317a3fcd5d04aa64ad301`;
the input SHA-256 is
`4308cf3733e1ec250744813bc2bfa9686b867ed1ea87b944cbd4a45b6a95651a`.

The first 1,024-round calibration was reduced uniformly because the tree-walker
exceeded the project’s one-minute test cap. The retained 256-round workload
completes in about 19.4 seconds under the tree-walker and about 2.6 seconds in a
direct bytecode smoke run, while remaining long enough to measure the compiled
body. No discarded calibration result is included in a retained mean.

## Coverage result

The application is portable and belongs to ten feature families: lexical
bindings and patterns; nominal types, generics, and unions; expressions,
arrays, text, and files; control flow; inherent methods; interfaces and
dispatch; `Option`/`Result`/exceptions; packages and imports; stdlib protocols;
and real program entry. It does not claim closure/callable or concurrency
coverage.

The checked catalog now contains 50 portable applications and 93 selected
performance rows: 50 compiled and 43 bytecode. The generated 165-triple
interaction frontier retains zero depth-zero or depth-one triples, raises
minimum depth from the reconstructed baseline’s one to three, and improves 160
triples. The generated records are
`2026-07-22-sensor-calibration-interaction-triple-frontier.{json,md}`.

## Repeated measurements

Every timed lane received five independent verifier-backed workstation
processes. All 25 processes passed, all samples were retained, and all outputs
have SHA-256
`e96cf1e366228f34478289660b4478b345bc069ac6e6633900d9805f0340edbb`.

| Lane | Successful processes | Mean | Ratio |
| --- | ---: | ---: | ---: |
| Able compiled | 5/5 | 0.3180 s | 67.660× Go |
| Go 1.26 | 5/5 | 0.0047 s | — |
| Able bytecode | 5/5 | 2.8480 s | 97.534× Python / 35.869× Ruby |
| Python 3.14 | 5/5 | 0.0292 s | — |
| Ruby 4.0 | 5/5 | 0.0794 s | — |

The compiled samples are `0.41`, `0.28`, `0.31`, `0.29`, and `0.30` seconds.
The bytecode samples are `3.03`, `2.98`, `2.63`, `2.77`, and `2.83` seconds.
Go samples are `0.004883290`, `0.004483136`, `0.004518740`, `0.004651709`,
and `0.004943736` seconds. Python samples are `0.030836833`, `0.028708358`,
`0.028470497`, `0.030738091`, and `0.027339617` seconds. Ruby samples are
`0.084917134`, `0.074263309`, `0.084780188`, `0.072339541`, and
`0.080630882` seconds. None were rejected as workstation noise.

## Ownership and admission

Three verifier-backed compiled main-only profiles merge to 570 ms of CPU
samples. Parsing owns 52.63% cumulatively, `String.split` owns 36.84%, builtin
string materialization owns 26.32%, validated byte access owns 21.05%, array
construction/storage owns 17.54%, signed integer parsing owns 12.28%, and
string-to-byte conversion owns 10.53%. These are cumulative compositions of
the already-investigated generic string/array identity machinery, GC, and the
application’s parser. No new exact compiler leaf repeats broadly enough to
admit a candidate.

Three bytecode-runtime profiles average 2.402 seconds per operation and merge
to 7.77 seconds of CPU samples. `runResumable` owns 90.22% cumulatively; exact
flat VM leaves include checked stack append at 2.19%, raw integer extraction at
2.06%, call-frame pop and stack snapshot at 1.67% each. Typed-pattern matching
is 7.34% cumulative. The stack, raw-integer, snapshot, and type-match families
are already closed by wider unlike-application gates, and the new profile does
not invalidate those results. A workload-specific sensor or record fast path
would violate the project’s generality rule.

## Source fingerprints

- canonical/external Able source:
  `28a29926e6511185eda8621d69f09109a3f16bcff69de6f41e0d4cdc4c0d480d`
- Go source:
  `89c3de53add27fe7b865a16fe379bb264c4681ae4d2e5d8802c5b83489f5e0d6`
- Python source:
  `03ddbcc562997e1901ff546c9a8e5212b59de2b51a8bd69d7295ffbc618d29bf`
- Ruby source:
  `18fc9a0df5f593ed69981b767c2d9061df49bb803aa61fc39c21a0007ecc6db8`
- verifier:
  `b1f77a04897780b5599795b4f1e827fb9ab433d2df3aed662b913d0bf98fc0bd`

## Verification

- exact output parity across both interpreters, compiled Able, and all three
  reference languages;
- five verifier-backed processes per timed lane, averaged without discarding
  volatile samples;
- explicit source, verifier, argument, and input-file fingerprints;
- three compiled and three bytecode main-only CPU profiles;
- catalog, selection, coverage, operation-depth, pair-, and triple-interaction
  checks;
- all new source files remain below 1,000 lines;
- `git diff --check`.

## Completed follow-up

The 50-row compiled and 43-row bytecode selection has been promoted and the
current frontier and architecture budget have been regenerated. See
`2026-07-22-sensor-calibration-scorecard-reconciliation.md` for the repeated
promotion cohort, current totals, and next structural-strategy gate.
