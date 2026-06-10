# Bytecode text/map/graph profile gate

Date: 2026-07-19

## Decision

Complete the current K-Nucleotide, Word Frequency, Dependency Plan, and
Document Audit bytecode gate and retain no VM, runtime, compiler, canonical-
stdlib, benchmark, fixture, language, or WASM change.

The profiles do not support the proposed shared HashMap hypothesis. Exact Able
HashMap entry search is material only in Word Frequency. K-Nucleotide reaches
the same helper, but at only 1.38% cumulative CPU, while Dependency Plan and
Document Audit do not sample it materially. The one exact CPU parent that is
material in all four programs is cached member-method lookup. That family is
not a new candidate: its dependency validation, direct-cache layout, hotness
admission, same-parent scope shortcut, and map-entry rewrite policy have
already received focused semantic tests and repeated broad A/B gates. The
remaining children do not identify a new removable operation shared by three
unlike programs.

## Bounded profile contract

The profile binaries were built once from the current tree. Their SHA-256
fingerprints were:

- ordinary `able` CLI:
  `52e707cae58aa7b74937907ca43cc9e7df04874f76a2779e58c2355742f445a2`;
- interpreter test binary:
  `53e137130679ee6feab651a14a482a81c958c18aa4cb413e4043faedb4b0d517`.

Every process used canonical `../able-stdlib`, one workstation CPU,
`GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`, and a 55-second cap.
K-Nucleotide used one ordinary process because a warm plus measured call
cannot fit below the project limit. It completed in 45.84 seconds, passed its
public Ruby verifier, and produced stdout SHA-256
`d628623daa677c00673df8d4961d14eb271b4112cdcb65b8233f07b69d7b49b8`.

The shorter programs used the existing runtime benchmark: load and typecheck
once, warm `main` once, then execute a fixed number of measured calls in the
same process. Separate CPU and allocation runs avoided combining profiler
overheads.

| Application | CPU calls | CPU runtime | Allocation runtime | Bytes/op | Allocs/op |
| --- | ---: | ---: | ---: | ---: | ---: |
| Word Frequency | 6 | 1,268,871,380 ns/op | 1,337,823,894 ns/op | 47,265,345 | 631,324 |
| Dependency Plan | 20 | 350,015,053 ns/op | 335,492,029 ns/op | 1,921,084 | 26,740 |
| Document Audit | 1,000 | 12,055,119 ns/op | 14,279,901 ns/op | 367,586 | 487 |

These warmed calls intentionally discard program output. Correctness remains
anchored by the current five-process verifier-backed scorecard on identical
application/input fingerprints; the ordinary K-Nucleotide diagnostic was
verified directly. The profile gate changes no program or measurement
classification.

## Exact CPU ownership

| Exact operation | Word Frequency | Dependency Plan | Document Audit | K-Nucleotide | Interpretation |
| --- | ---: | ---: | ---: | ---: | --- |
| `hashMapFindEntryWithHash` cumulative | 7.79% | no material sample | no material sample | 1.38% | one material application, not a shared map wall |
| `fastPrimitiveHashMapKeyEqual` cumulative | no material sample | no material sample | no material sample | 0.99% | K-Nucleotide-only key shape |
| `lookupCachedMemberMethodEntry` cumulative | 3.04% | 14.24% | 34.00% | 2.55% | broad, but the completed member-validation family |
| `memberMethodLexicalStateHeader` cumulative | below 1% | 1.29% | 3.42% | 0.44% | material in only two rows and already A/B-rejected |
| `bytecodeRawIntegerValueInfo` cumulative | 1.19% | 1.73% | 0.33% | 5.86% | caller mix and prior broad raw-carrier gates remain closed |

The broad `runtime.mapaccess2_faststr`, AES string hashing, and Go map-control
leaves do recur, but their Able owners differ. Word Frequency uses language
HashMap lookup. Dependency Plan and Document Audit primarily use VM lexical,
member, environment, and instruction caches. K-Nucleotide combines those
caches with a smaller language HashMap component. Optimizing the Go runtime
leaf as if it represented one Able operation would conflate unrelated maps.

The member-method result is real but not novel. The retained direct and
dependency-validated caches already removed broader resolution work. A later
path census found repeated dependency validation rather than a few cold sites;
a generic sibling-scope shortcut then changed Document Audit by +7.03%,
Lexical Rollup by +1.23%, iterator collect by +4.42%, and split/join by +2.19%
in repeated baseline/candidate means. The shortcut was removed. This new
profile supplies no different exact child or invalidation proof that would
justify retrying that policy.

## Allocation ownership

The sampled allocation profiles split even more clearly:

| Application | Leading exact owners |
| --- | --- |
| Word Frequency | positional UTF-8 result structs: 32.75% of bytes and 21.28% of objects; String host conversion: 20.68% cumulative bytes |
| K-Nucleotide | positional UTF-8 result structs: 41.00% of bytes and 29.45% of objects; String host conversion: 34.52% cumulative bytes |
| Dependency Plan | Array member materialization: 25.88% of bytes and 48.12% of objects; primitive Array storage/promotion/growth |
| Document Audit | file read and line materialization: 79.43% cumulative bytes; iterator member values: 43.51% of objects |

Word Frequency and K-Nucleotide repeat `Utf8DecodeResult`, but the completed
positional-struct identity/lifetime census found this definition in only those
two unlike applications, short of the three-program and two-definition gate.
Dependency Plan is Array/graph dominated, while Document Audit is file-line
and iterator dominated. Process-global integer-cache initialization and broad
Array-store/GC helpers are excluded as candidates by the existing allocation-
owner census and representation gates.

No exact allocation owner is both new and material in three unlike programs.
Consequently no candidate, timing A/B, or downstream guard cohort is warranted.

## Verification and cleanup

- All four CPU profiles contain usable samples: 45.55 seconds for
  K-Nucleotide, 7.57 seconds for Word Frequency, 6.95 seconds for Dependency
  Plan, and 12.00 seconds for Document Audit.
- The three repeated-main CPU and allocation benchmark processes passed.
- The ordinary K-Nucleotide process completed below one minute and verified.
- Focused bytecode tests pass on the unchanged production tree.
- Raw profiles, binaries, stdout/stderr captures, and temporary top reports
  are cleanup-only and are removed after this record is written.
- No WASM work was performed.

## Next recommendation

Run a current cross-mode numeric/wide profile gate over Fixed Width 128,
Rational Series, Distance Field, and RMS Norm, with Mandelbrot as a float-loop
discriminator.

Why: text/map/graph now joins return/frame, raw-carrier, member-validation,
Array-growth, and regex-local designs as a closed local search. Numeric/wide
is the largest remaining stable family with several unlike applications and
potentially shared language-level checked arithmetic, conversion, and nominal
result boundaries in both generated code and bytecode. Looking across both
product modes also gives compiler work a fair turn instead of continuing to
subdivide exhausted bytecode parents.

What it entails: collect preserved generated-main and bounded warmed-bytecode
CPU/allocation profiles under the current source fingerprints; separate
checked primitive arithmetic, wide nominal dispatch, raw float/integer
transport, result construction, type/coercion, and runtime-call ownership by
exact descendant. Advance at most one operation that is material in three
unlike programs and is expressible through primitive rules or shared nominal
lowering. Any candidate then needs focused semantics plus alternating,
verifier-backed workstation averages for all admitted programs and unrelated
text, Array, concurrency, and current-target guards. Do not reopen rejected
typed-block/carrier designs, add named-type or benchmark special cases, or
begin WASM work.
