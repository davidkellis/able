# Cross-application bytecode profile gate

Date: 2026-07-17

## Decision

Complete the Reverse Complement, Rational Series, Word Frequency, and Array
Slice Window profile gate and retain no VM, runtime, compiler, stdlib,
benchmark, fixture, or language change. The refreshed profiles reproduce a
real shared primitive-metadata lookup wall, but that exact representation has
already failed two unlike-application gates. The other repeated-looking
families either have disjoint semantic owners or are previously rejected
raw-integer and return/frame candidates.

No candidate was eligible, so the production tree was not changed and no
benchmark-specific fast path was built. Canonical `../able-stdlib` also needed
no change. WASM remains deferred.

## Method

- Used canonical external `able-stdlib` with `GOMEMLIMIT=1GiB`, `GOGC=50`,
  `GOMAXPROCS=1`, and CPU 0 affinity.
- Loaded and typechecked each program once per process, warmed `main()` once,
  forced GC, and measured one subsequent call. Five independent processes per
  application provide workstation averages instead of single-run claims.
- Used the catalog's canonical input contracts: Reverse Complement read
  `reverse-complement-input.fasta`, Word Frequency read `corpus.md`, and the
  other two programs took no arguments.
- Captured one separate bounded CPU profile per application. Reverse
  Complement and Rational Series used one measured call, Word Frequency used
  three, and Array Slice Window used six. Every process stayed below the
  55-second limit.
- Reconciled fully qualified concrete symbols below `runResumable` and call
  dispatch. A candidate required the same material operation in at least
  three unlike applications and could not retry a previously rejected
  representation.

## Five-process warmed measurements

All 20 processes completed without a timeout or failure.

| Application | Mean ns/op | Observed ns/op span | B/op | allocs/op |
| --- | ---: | ---: | ---: | ---: |
| Reverse Complement | 6,412,132,109 | 6,308,645,861-6,488,583,609 | 705,411,446 | 10,894,478 |
| Rational Series | 3,683,855,732 | 3,615,265,804-3,806,922,074 | 129,951,880 | 1,405,666 |
| Word Frequency | 1,182,351,545 | 1,160,917,233-1,198,727,032 | 47,740,130 | 631,449 |
| Array Slice Window | 500,800,221 | 488,406,338-523,958,066 | 14,158,600 | 422,250 |

The allocation means closely reproduce the 2026-07-16 single-process owner
census: Reverse Complement differs by less than 0.01%, Rational Series and
Array Slice Window are exact at the displayed precision, and Word Frequency
is 0.39% higher. The prior census was therefore representative, while the new
rows provide the requested independent-process averaging.

## Concrete CPU reconciliation

| Application | CPU samples | Primitive metadata | Raw integer extraction | Return/frame | Application-specific work |
| --- | ---: | ---: | ---: | ---: | --- |
| Reverse Complement | 6.36 s | `lookupIntegerInfo` 1.73% cumulative | 3.93% flat | `finishInlineReturn` 1.89% cumulative | Array slot member 33.49% cumulative; mono-u8 reads, pushes, boxing, and handle maps |
| Rational Series | 4.10 s | `lookupIntegerInfo` 4.15% cumulative | 7.80% flat | return 7.56%; frame pop 2.68% | casts 11.22%, division/modulo 5.61%, nominal Rational calls |
| Word Frequency | 3.62 s | `lookupIntegerInfo` 3.04% cumulative | 1.93% flat | return 14.64%; frame pop 1.10% | `hashMapFindEntryWithHash` 7.18% flat and nominal/type-match work |
| Array Slice Window | 3.30 s | `lookupIntegerInfo` 7.27% cumulative | 6.36% flat | return 3.33%; frame pop 1.21% | casts 22.73%, Array slot members 16.06%, arithmetic |

`lookupIntegerInfo(...)` is the one exact material leaf shared by all four
applications. It reads the closed twelve-entry primitive integer metadata
table. This is generic primitive language machinery, but it is not a new
candidate: the 2026-07-11 map-attribution tranche replaced it with a closed
metadata switch and found no stable application win. The stricter 2026-07-16
gate repeated that full switch and a membership-only classifier. Despite much
faster isolated lookup benchmarks, iterator collect regressed 7.16% and Array
map regressed 3.68% with the full switch; the narrower classifier recovered
iterator by 2.28% but regressed Array map 6.21%. Both were removed. This gate
does not overrule those averaged unlike-program failures with profile
percentages.

The Go string/Swiss-map leaves also recur, but their Able owners diverge:
primitive metadata and environment lookup dominate Rational and Array Slice;
known-type and type-match caches contribute in Word Frequency; and Reverse
Complement additionally spends 8.81% in integer-keyed boxed-value and Array
handle maps. Replacing one nominal map or one container's storage would not be
a shared operation.

`bytecodeRawIntegerValueInfo(...)` recurs in every program, but the generic
carrier and direct i32/i64 store variants have already failed broad guards.
`finishInlineReturn(...)` and `popCallFrameFields(...)` recur in three or four,
but the slotless-return guard reorder and caller-owned frame-result carrier
were neutral or regressive. No new subpath appears here that would justify
reopening those candidates.

## Verification and cleanup

The focused integer/numeric/cast/type-match tests and the runtime package pass
against the unchanged production tree. Temporary benchmark binaries, JSON
summaries, and raw profiles were removed after their aggregate evidence was
recorded here.

## Next recommendation

Profile the compiled `main` phases for the same four unlike applications and
compare them with their current Go references.

Why: this bytecode gate has exhausted its only exact shared leaf without a new
eligible representation, while the compiled rows remain 9.04x-29.85x slower
than Go and the project goal applies equally to generated binaries. Reusing
the same byte/array, numeric/nominal, map/text, and slicing cohort lets the next
gate ask whether compiler-generated code or a shared runtime bridge repeats
across unlike programs rather than changing the workload to fit a hypothesis.

What it entails: collect five verifier-backed compiled process samples plus
phase-separated main CPU and allocation profiles under the same one-CPU
limits; normalize generated-code, primitive operator, allocation, and runtime
bridge descendants; and admit a candidate only when the same generic lowering
or semantic operation is material in at least three applications. Do not add
per-container nominal lowering, retry the rejected unused-operator dependency
cut, or optimize bootstrap work that is absent from the measured `main`
phase.
