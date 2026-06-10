# Clean four-workload bytecode profile reconciliation

Date: 2026-07-17

## Decision

Complete the Rational Series, Array Slice Window, numeric Array Map, and
K-Nucleotide CPU/allocation reconciliation and retain no bytecode VM, runtime,
compiler, canonical-stdlib, benchmark, fixture, or language change. Clean
profiles reproduce several exact shared VM costs, but every material shared
leaf belongs to a generic representation or policy already rejected by broad
application gates. The remaining return, allocation, and semantic descendants
split by workload.

No candidate was eligible. In particular, this tranche does not retry raw
integer carrier changes, integer-metadata switches, frame-result ABI changes,
Array growth-factor changes, or the rejected slot/constant fusion. WASM
remains deferred.

## Protocol

- Used canonical external `../able-stdlib/src`, CPU 0, `GOMAXPROCS=1`,
  `GOGC=50`, `GOMEMLIMIT=1GiB`, and a 55-second process cap.
- Rational Series, Array Slice Window, and Array Map used the retained runtime
  benchmark boundary: load and typecheck once, warm `main()` once, force GC,
  then profile only repeated measured calls. Counts were 4, 20, and 100.
- K-Nucleotide cannot fit both a warm and measured call below one minute. It
  therefore used one ordinary verifier-backed bytecode process. Its 39.85 CPU
  seconds below `runResumable` make load/typecheck contamination immaterial.
- CPU and sampled allocation profiles were captured in each bounded process.
  The benchmark `B/op` and `allocs/op` counters are authoritative for the three
  measured-main runs; package initialization necessarily remains visible in
  Go allocation profiles and was excluded from owner reconciliation.
- A candidate required the same concrete material descendant in at least
  three unlike workloads and could not retry a representation already rejected
  by repeated broad guards.

K-Nucleotide's stdout passed its public Ruby verifier. The other three used
the already output-guarded benchmark harness and completed successfully.

## Measured regions

| Workload | Measured calls | CPU samples | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: | ---: | ---: |
| Rational Series | 4 | 14.76 s | 3,700,039,792 | 129,897,098 | 1,405,571 |
| Array Slice Window | 20 | 10.95 s | 549,586,358 | 15,238,539 | 426,502 |
| Array Map | 100 | 7.43 s | 74,515,476 | 805,442 | 119 |
| K-Nucleotide | 1 ordinary process | 39.85 s | n/a | sampled whole process | sampled whole process |

The absolute Map time includes simultaneous CPU and allocation sampling and is
not promoted as a new wall-time baseline. These runs were attribution inputs,
not an A/B claim. No candidate existed, so the workstation averaging rule did
not require candidate/control timing cohorts.

## Exact CPU intersection

Percentages are cumulative unless the helper is a leaf.

| Concrete helper | Rational | Slice | Map | K-Nucleotide | Reconciliation |
| --- | ---: | ---: | ---: | ---: | --- |
| `execBinary` | 9.28% | 15.07% | 9.83% | 15.28% | parent with different arithmetic/bitwise children; the only shared slot/constant fusion was just rejected |
| `runtime.mapaccess2_faststr` | 6.98% | 12.24% | 6.06% | 7.13% | integer metadata, environments, type caches, and member caches have different owners |
| `bytecodeRawIntegerValueInfo` | 5.89% flat | 6.58% flat | 6.33% flat | 6.55% flat | exact shared leaf, but unified extraction is faster than the rejected split and raw-carrier/store variants failed broad guards |
| `lookupIntegerInfo` | 3.12% | 7.31% | 3.23% | 3.54% | exact shared metadata read; both full-switch and membership-switch replacements regressed unlike applications |
| `execStoreSlotOpcode` | 2.98% | 4.75% | 3.10% | 3.89% | parent already reduced through generic raw-slot pooling/store cleanup; descendants are carrier-specific |
| `execLoadSlotOpcode` | 2.03% | 3.01% | 3.36% | 3.81% | stack/raw snapshot family whose ordering and carrier trials failed broad controls |
| `finishInlineReturn` | 9.82% | 1.10% | 13.59% | 16.14% | material in three, but its descendants split below the parent |

The return subtree is not a hidden common leaf. Rational spends 2.37% in
`popCallFrameFields`; Map spends 4.31% flat boxing a returned raw i32 and 3.77%
in frame pop; K-Nucleotide spends 4.24% in return coercion/type matching and
3.99% in frame pop. Slice spends only 0.18% in frame pop. The shared frame
helper is the exact ABI whose caller-owned result-structure candidate
regressed Split/Join, iterator collect, and Distance Field. Reopening it from
these profiles would optimize a closed shape.

## Allocation reconciliation

Allocation owners also divide:

- Rational Series is dominated by nominal Rational construction, per-call
  environments, and runtime static-receiver type expressions.
- Array Slice Window allocates required independent slice results, raw integer
  arithmetic results, Array value/lease shells, and tracked primitive-cache
  growth.
- Array Map has only 119 measured allocations per call. Its sampled bytes are
  mostly result backing growth and tracked raw-cache backing growth rather than
  object churn.
- K-Nucleotide's ordinary-process profile is dominated by String/byte host
  conversion (380.10 MiB sampled), positional result structures (496.56 MiB),
  Array backing growth (120.85 MiB), and native-boundary i32 materialization
  (62 MiB).

`runtime.ArrayEnsureCapacity` is the one repeated allocation-space leaf:
27.17% of Slice, 42.05% of Map, and 9.48% of K-Nucleotide sampled bytes. It is
not a new candidate. The generic 1.75x post-4096 growth experiment increased
iterator allocation, was neutral on text, and moved numeric wall time only at
noise level. In these profiles the helper consumes at most 0.46% CPU, and the
three callers request different required results. A different growth factor
would trade memory among applications rather than remove shared semantic work.

## Verification and cleanup

- All four bounded profile processes completed; K-Nucleotide verified.
- No source candidate was built, so no candidate-specific A/B or source test
  was represented as performance evidence.
- Focused bytecode VM and benchmark-selection tests pass against the unchanged
  production implementation.
- Temporary binaries, stdout, and raw profiles were removed after this
  aggregate record was written.

## Next recommendation

Profile a fresh bytecode text/byte conversion cohort: K-Nucleotide, Reverse
Complement, Base64, and Word Frequency, using a non-text numeric/nominal guard.

Why: the current numeric/Array cohort has exhausted every exact shared leaf,
while K-Nucleotide exposes 380 MiB of sampled allocation in a canonical
String/byte host boundary. Text ingestion, byte conversion, and output are
ordinary application operations also exercised by the other three programs;
they offer a plausible shared semantic boundary without inventing a named
container or benchmark-specific rule.

What it entails: collect bounded main-only CPU and allocation profiles where a
warm call fits, use one ordinary bounded process for K-Nucleotide, and attribute
the String host-builtins to exact operations and callers. Admit a candidate
only if the same conversion/copy/materialization step is material in at least
three text applications and is not one of the previously rejected broad
String/byte representations. Any candidate must use repeated order-balanced
A/B means, preserve UTF-8 and mutable Array semantics, pass the numeric/nominal
guard, and leave canonical `../able-stdlib` changes generic to the String/byte
API rather than to an application.
