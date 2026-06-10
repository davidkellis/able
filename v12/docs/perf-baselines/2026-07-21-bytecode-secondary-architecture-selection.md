# Bytecode secondary architecture selection

Date: 2026-07-21

## Decision

Keep no VM, compiler, runtime, stdlib, benchmark, or language performance
change. A fresh six-application selection cohort found one new Able-owned
exact leaf in three unlike programs, but the only eligible general candidate
regressed all five measured applications and was fully removed.

The ordinary small named-struct rule remains unchanged: definitions with at
most four fields use direct comparisons, while larger definitions carry an
immutable name-to-index map.

## Reconciliation and cohort

The semantically separate operand-register engine is not an open architecture
candidate. Today's completed register gates admitted and executed broad
`MemberAccess` work, removed frame allocation, and hoisted fruitless
continuation probes, but the final engine remained 5.75% slower in the pooled
Word Frequency guard. The ordinary VM remains authoritative.

This tranche therefore selected high-excess rows not used together in the
earlier six-application exact-leaf sweep:

- K-Nucleotide: text, UTF-8, integer-key HashMap, and call/return work;
- Mandelbrot: float arithmetic and byte output;
- Wide Integer Records: parsed records and checked wide nominal arithmetic;
- Unicode Scalar Pipeline: char iteration, ratios, and StringBuilder;
- Log Routing Redaction: regex, text iteration, Results, and NFA structures;
- Inventory Reconciliation: numeric-key map aggregation.

Concurrent Event Routing was initially selected, but the one-main retention
probe directly invokes `main` and cannot flush goroutine-executor work while
that call is blocked at a channel receive. Its timeout is a diagnostic-harness
limitation, not a product result; the normal verifier-backed scorecard row
passes. Log Routing Redaction replaced it. No concurrency timing or profile is
used in this decision.

## Profile protocol

An ordinary restored interpreter test binary was frozen before collection.
Each application received two fresh measured-main CPU processes with no event
statistics. Processes used canonical `able-stdlib`, the catalog run directory
and arguments, `GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`, and a 60-second
cap. The retained one-main harness excluded loading/lowering and final GC.

| Application | Merged profile duration | CPU samples | Mean main time |
| --- | ---: | ---: | ---: |
| K-Nucleotide | 88.28 s | 87.92 s | 44.150 s |
| Mandelbrot | 13.31 s | 13.29 s | 6.671 s |
| Wide Integer Records | 10.06 s | 10.05 s | 5.044 s |
| Unicode Scalar Pipeline | 6.52 s | 6.51 s | 3.267 s |
| Log Routing Redaction | 5.20 s | 5.17 s | 2.603 s |
| Inventory Reconciliation | 4.77 s | 4.76 s | 2.402 s |

At the 1% flat threshold, the broad intersections were the already-closed
dispatcher, call, raw-integer, stack transport, Go map/hash, allocation, and
GC families. One new exact Able leaf cleared the three-program rule:
`structDefinitionNamedFieldIndex(...)` was 1.19% flat in Wide Integer
Records, 1.38% in Unicode Scalar Pipeline, and 2.71% in Log Routing Redaction.

## Identity and field census

A temporary `able_named_field_stats` build counted definition/field identities
after load and before `main`. The instrumentation was used only for deterministic
counts, never timings.

- Wide Integer Records performs millions of `Int128.low/high`,
  `UInt128.low/high`, and method-name misses, plus byte-iterator fields.
- Unicode performs 3,538,976 `RawStringCharsIter.offset` lookups, 1,769,504
  each for `bytes` and `len_bytes`, and independent StringBuilder lookups.
- Log performs millions of lookups across `RegexNFAThreads`, `RegexNFAIndex`,
  transitions, code points, ranges, NFA state, char iterators, and
  StringBuilder.

Caller trees put all three under
`structNamedFieldValue → structDefinitionNamedFieldIndex`. The definitions,
field names, successful reads, and negative method-name probes are diverse,
so this is a genuine general nominal operation rather than one shared stdlib
definition.

## Candidate and repeated gate

The sole candidate built `NamedFieldIndices` for every non-empty named struct,
not only definitions with more than four fields. This reused the existing
immutable general table and preserved dynamic definitions, aliases, positional
storage, and all field semantics. It did not name an application, stdlib type,
container, or benchmark.

Focused struct/member tests passed. Frozen baseline and candidate binaries then
received five fresh one-main pairs per application with alternating launch
order. Arithmetic means are required on this workstation. Positive changes
are regressions.

| Application | Baseline mean | Candidate mean | Change |
| --- | ---: | ---: | ---: |
| Wide Integer Records | 5.0848 s | 5.3485 s | +5.19% |
| Unicode Scalar Pipeline | 3.3986 s | 3.4476 s | +1.44% |
| Log Routing Redaction | 2.6262 s | 2.7602 s | +5.10% |
| Inventory Reconciliation | 2.4088 s | 2.4287 s | +0.83% |
| Mandelbrot | 6.3345 s | 6.4989 s | +2.60% |

The map regresses both primary and unrelated rows. Hashing a small immutable
field set costs more than one-to-four direct comparisons even when the direct
helper is profile-visible. The result independently confirms the older
single-fixture decision and supplies the broader evidence previously missing.
No long K-Nucleotide candidate cohort was warranted after every completed row
regressed.

The field-index candidate, its changed test expectation, the complete
build-tagged census, and report plumbing were removed.

## Restored verification

The restored full bytecode family passes:

```text
go test ./pkg/interpreter -run 'TestBytecode' -count=1 -timeout 60s
ok   able/interpreter-go/pkg/interpreter  25.422s
```

One fresh ordinary bytecode process for each selected application passed its
catalog Ruby verifier. Times were 43.71 s K-Nucleotide, 5.86 s Mandelbrot,
4.74 s Wide Integer Records, 3.70 s Unicode, 3.16 s Log Routing, and 2.58 s
Inventory. These one-process values are correctness smokes, not performance
selection evidence.

## Next recommendation

Shift the next tranche to the compiled runtime boundary and re-evaluate
interpreter-package isolation against the current Binary Trees guard.

Why: the ordinary VM, separate register engine, dynamic integer cache, and now
small nominal-field map all lack a broadly profitable next change. The
compiler's retained interface boundary still leaves only binary/unary operator
roots pulling the whole interpreter package into static binaries. The prior
interpreter-independent primitive package removed about 55 ms, 38 MB of init
allocation, and 5.6 MB of binary size from every static application, but was
rejected when the then-current Binary Trees took about 31.5 seconds and became
5.07% slower through changed GC pacing. Current Binary Trees is about 10.2
seconds, meets the Go target, and has substantially changed allocation/GC
behavior, so that old guard result is now stale enough to justify one bounded
revalidation.

What it entails: recover the already-proven generic primitive operator boundary
without changing Able semantics; freeze source-matched baseline/candidate
binaries; prove static dependency removal and dynamic fallback parity; then run
alternating five-process means for short startup-heavy programs, allocation-
light TapeLang, current Binary Trees, and unrelated compiled/bytecode guards.
Retain it only if Binary Trees keeps its target and no broad guard regresses.
Do not add heap ballast, alter GC policy, specialize a nominal type, or do WASM.
