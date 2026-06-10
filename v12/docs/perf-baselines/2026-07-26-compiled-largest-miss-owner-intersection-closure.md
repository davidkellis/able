# Compiled largest-miss owner intersection closure — 2026-07-26

## Decision

Retain no compiler, generated-runtime, runtime, interpreter, bytecode VM,
canonical-stdlib, application, language, dependency, or WASM change.

The four largest absolute strict compiled misses do not share an exact
main-phase CPU or allocation owner. Their dominant work separates into
primitive dispatch and checked Array mutation, runtime-backed generic map
storage, search and checked arithmetic, and the already-rejected goroutine
identity boundary. Neither parity control shares those target descendants.

The required three-unlike-application admission gate therefore fails before
implementation. No prototype or A/B timing cohort is justified.

## Protocol

The current corrected compiler had SHA-256
`6ea26d6ee1e9b6b5447371bb01adc8052738ec08b13177db28b972b8c30c7bcc`.
It rebuilt all six applications with `--no-fallbacks`. Every final graph
contains 96 packages and omits `pkg/interpreter`.

All build and profile work used the disk-backed workspace
`/var/tmp/able-main-owner-intersection-20260726`. Serial applications and the
serial control used CPU 5, `GOMAXPROCS=1`, and the serial executor. Mutex Work
Queue and Binary Trees used CPUs 5,10,15,11, `GOMAXPROCS=4`, and the goroutine
executor. Every run used `GOMEMLIMIT=1GiB` and `GOGC=50`.

Each application received ten independent main-only CPU profiles. Five
applications received three exact `MemProfileRate=1` start/end allocation
profiles and exact pre-observer main MemStats deltas. Binary Trees' first exact
end-snapshot allocation run exceeded the one-minute process guard while
serializing the profile, so that incomplete profile was excluded. Three
lighter exact main MemStats runs replaced it; they preserve total allocation
evidence without treating the profiling observer as application work.

The retained evidence comprises six smoke runs, 60 CPU-profile runs, 15 exact
allocation-profile runs, and three Binary Trees MemStats runs. All 84
processes passed their public verifiers.

## Governing scorecard rows

| Application | Role | Able | Go | Excess |
| --- | --- | ---: | ---: | ---: |
| Tapelang Alphabet | target | 3.7940 s | 2.0119 s | +1.7821 s |
| K-Nucleotide | target | 1.7260 s | 0.0556 s | +1.6704 s |
| Sudoku Masks | target | 1.8260 s | 0.6382 s | +1.1878 s |
| Mutex Work Queue | target | 0.9440 s | 0.0044 s | +0.9396 s |
| Pidigits | serial control | 1.1740 s | 1.1715 s | +0.0025 s |
| Binary Trees | concurrent control | 10.4000 s | 10.8498 s | -0.4498 s |

## CPU owner attribution

The ten verified profiles for each application were merged only within that
application.

| Application | Merged CPU | Dominant exact descendants |
| --- | ---: | --- |
| Tapelang Alphabet | 37.16 s | `execute` 64.16% flat; `Tape.inc` 27.13%; `Tape.get` 6.73%; `Tape.move` 1.29% |
| K-Nucleotide | 16.72 s | `primitiveHashMapKeyEqual` 15.85% flat; `primitiveHashMapHash` 7.60%; `mallocgc` 30.98% cumulative |
| Sudoku Masks | 17.65 s | `find_best_empty` 30.71% flat; checked multiply 13.26%; signed divmod 11.78%; `bit_count` 11.33% |
| Mutex Work Queue | 19.08 s | `currentGID` 92.35% cumulative through `runtime.Stack`; scheduler lock/spin work below it |
| Pidigits | 11.67 s | native `math/big`: `mulAddVWW` 43.10%, `lshVU` 12.85%, `subVV` 8.31%, `addVV` 7.37% |
| Binary Trees | 393.20 s | native allocation/GC: `tryDeferToSpanScan` 25.49% flat, `mallocgc` 71.37% cumulative, `make_tree` 77.74% cumulative |

Generated-caller inspection confirms the distinctions:

- Tapelang is already a native primitive/Array loop with required checked
  mutation and Array semantics. This reproduces its earlier compute closure.
- K-Nucleotide converts statically specialized keys and values into the
  runtime-backed generic map service. Its largest allocation leaves are
  `bridge.ToUint` and `bridge.ToInt`. A narrow repair would name HashMap; the
  general typed-storage architecture remains unproven across unlike nominal
  families.
- Sudoku's largest allocation site is the generated `find_best_empty` body,
  which allocates a native three-element position Array when it finds a new
  best cell. That is application search work, not a compiler rule.
- Mutex Work Queue repeatedly obtains goroutine identity with
  `currentGID`/`runtime.Stack`. Multiple prior cohorts already established
  this cost and rejected the only known general fixed-context ABI because it
  regressed independent applications.
- Pidigits and Binary Trees spend their time in the expected native Go
  arithmetic and allocation machinery. They do not hide a target-only Able
  boundary.

## Exact allocation evidence

The means below come from three exact main-phase MemStats deltas, captured
before end-snapshot serialization.

| Application | Allocated bytes | Allocations | GC |
| --- | ---: | ---: | ---: |
| Tapelang Alphabet | 282,984 | 4,277 | 0.00 |
| K-Nucleotide | 614,264,368 | 16,232,599 | 337.00 |
| Sudoku Masks | 156,389,341 | 7,802,767 | 125.67 |
| Mutex Work Queue | 17,443,352 | 314,845 | 13.67 |
| Pidigits | 298,929,032 | 25,218 | 272.67 |
| Binary Trees | 9,820,588,397 | 613,771,228 | 205.00 |

Exact start/end allocation attribution was stable across the five completed
profile sets:

- Tapelang: 4,225 objects in lazy common-`i32` box initialization at its
  output boundary; not CPU-hot or shared materially.
- K-Nucleotide: 7,999,998 objects in `ToUint`, 3,961,373 in `ToInt`,
  approximately 1,000,129 in nullable-`i32` decoding, and 233,356 in native
  String conversion.
- Sudoku: exactly 7,787,560 objects in `find_best_empty` per run.
- Mutex Work Queue: run one contains 60,046 `currentGID` objects, 53,737
  `call_value_fast` objects, and 40,960 mutex-awaitable struct encodes.
- Pidigits: its largest non-observer leaf is 9,081 `math/big.nat.make`
  allocations.

`runtime/pprof` snapshot serialization appears in the end-minus-start
profiles. It is an observer and is excluded from application attribution.

## Candidate reconciliation

| Candidate | Result | Reason |
| --- | --- | --- |
| Primitive/Array compute lowering | closed | material in Tapelang only; already native and previously profiled |
| Runtime-backed generic map storage | not admitted | material in K-Nucleotide only in this unlike cohort; named-container treatment is forbidden |
| Sudoku search/position allocation | rejected | application-specific semantic work |
| Checked arithmetic | insufficient breadth | material in Sudoku; the prior broad arithmetic gates were neutral or mixed |
| Goroutine identity lookup | closed | material in Mutex only here; the fixed-context ABI repeatedly failed broad controls |
| Shared allocation leaf | none | exact target allocation owners do not intersect |

Combining these different descendants under a broad label such as “generated
runtime” would discard the semantic and control distinctions the admission
gate exists to protect. The correct result is no code.

## Artifact identity

| Application | Binary SHA-256 |
| --- | --- |
| Tapelang Alphabet | `762ac2241bb76072ffb9198e36a3256235a3c6ab08d11c0b903cdc0504c78be6` |
| K-Nucleotide | `431269dc2eb51dda5b9bbe60bdccb4c1f9320c686de3aac5d2a771c70f3da539` |
| Sudoku Masks | `6104c744d5e03ca35245f05343df312db452274b458e8a99528fbe6ebef11637` |
| Mutex Work Queue | `5c6bae479606c6c959cb22498c0ae4dfb98069fa0ad62a34f7aae7d82e922e86` |
| Pidigits | `b3b95a76e14b853312f083acab32dc4b1cfcf731767774a6eaef9ac37aa112e0` |
| Binary Trees | `a92e956f57ebb4080f9a488d59f3f152f9dcf8e265f1c6c931f71dda14c7a388` |

The machine-readable companion is
`2026-07-26-compiled-largest-miss-owner-intersection-closure.json`. Raw
binaries, generated modules, profiles, and outputs are disposable after this
record.

## Next recommendation

Build a corpus-wide exact-owner frequency census over the current 61-row
strict compiled scorecard.

Why: the four largest absolute misses are heterogeneous. A general compiled
win may instead be a smaller exact leaf repeated across many applications,
which a top-four-only intersection cannot select.

What it entails: normalize current compatible main-only CPU/allocation
profiles by exact symbol and generated caller; refresh only stale or missing
rows; rank non-closed owners by material reach and total scorecard excess; and
admit a prototype only when one exact compiler/generated-runtime owner is
material in at least three unlike applications and non-material in controls.
If none exists, document that profile-driven local compiled optimization is
exhausted for the current corpus and advance to typed bytecode slot work.

Why it is important: this is the broadest remaining evidence test for a
general native-lowering improvement. It can find a widespread lowering tax
without inventing benchmark or named-container rules, and it gives a concrete
stopping condition rather than repeatedly profiling unrelated local maxima.
Do not begin WASM work.
