# Compiled Generated-Go Diagnostics Census

Date: 2026-07-18

## Decision

Keep no compiler, generated runtime, bytecode VM, canonical-stdlib,
application, fixture, or language change from this tranche.

Normal generated Go for 32 of the 35 selected compiled applications completed
both Go escape analysis and SSA bounds-check diagnostics. The three canonical
regex applications exceeded source generation's 55-second bound and an
independent 59-second retry, so they remain explicit exclusions rather than
partial evidence. The complete status and count matrix is retained in
`2026-07-18-compiled-generated-go-diagnostics-census.tsv`.

The static census does not expose a candidate that is both executed and
material in three unlike applications. Almost all broad diagnostics belong to
unconditionally emitted runtime, registration, error, conversion, or stdlib
code. Current CPU and allocation profiles divide below those broad parents.
The few repeated application-body bounds checks are String byte reading,
filesystem path cleanup, and three related iterator consumers; none meets the
cross-family runtime-materiality rule.

No WASM work was performed.

## Method

The compiler CLI was built once. Each application was then emitted normally,
without telemetry or a compiler flag that changes generated semantics. The
generated `able/compiled` module was built with:

```text
-gcflags='able/compiled=-m=2 -d=ssa/check_bce/debug=1'
```

Only diagnostics owned by generated application-module `.go` files were
retained. The normalizer kept:

- final `escapes to heap` decisions, excluding the preceding explanatory flow
  trace;
- `moved to heap` decisions;
- `leaking param` decisions, recorded separately because parameter leakage is
  not itself proof of a new allocation;
- `Found IsInBounds` and `Found IsSliceInBounds` sites.

Duplicate location/message reports were removed. Each remaining location was
mapped to its enclosing generated function and source line. Numbered generated
temporaries and numbered package-output filenames were normalized so the same
generator shape could be compared across applications. Source SHA-256,
generated file/byte counts, raw line counts, selected location counts, and
normalized shape counts remain in the TSV.

Each completed Go diagnostic build took 10-21 seconds. Generated trees,
binaries, raw logs, and the dedicated Go cache were deleted application by
application or at audit completion.

## Coverage

The 32 complete applications generated 1,272 Go files and 94,484,743 source
bytes. Go emitted 5,712,620 raw diagnostic lines. Normalization retained
514,814 unique final locations across 373,804 per-application shapes:

| Diagnostic | Locations | Notes |
| --- | ---: | --- |
| escape | 354,601 | Placement decisions, including inlined caller copies. |
| moved | 1,262 | Named locals moved to the heap. |
| parameter leak | 155,697 | Flow facts, not direct allocation counts. |
| index bounds | 2,594 | Remaining Go index checks. |
| slice bounds | 660 | Remaining Go slice checks. |

Within generated compiled Able bodies, as distinct from registration and
general runtime support, there were 34,676 escape sites, 446 moved-local sites,
71 index checks, and 23 slice checks. Static counts describe emitted code, not
execution frequency.

The three excluded regex rows are still represented in the matrix. Both
attempts stopped source generation before a complete Go tree existed, so no
escape or bounds count was inferred from partial output. Their current runtime
profiles from the preceding helper census remain valid for profile
intersection, but they are not included in static absence claims here.

## Bounds-check reconciliation

The broadest checks are cold generated-runtime shapes present in all 32
complete trees: type-expression comparison, interface dispatch selection,
runtime Array conversion, generic runtime HashMap operations, task-queue
insertion, launcher formatting, and launcher argument slicing. The preceding
execution census records zero selected hot-path calls through generic
interface dispatch and checked generic Array indexing, while current profiles
do not identify these runtime checks as shared material owners.

Only three compiled-Able body families repeat:

| Shape | Emitted applications | Interpretation |
| --- | ---: | --- |
| canonical `read_byte(bytes, idx) -> ?u8` | 14 | One check in the body and one through its entry wrapper; the same scalar success value also moves to the heap through `__able_ptr`. |
| `Path.parent` / path normalization pop | 11 | Imported filesystem setup code; absent from current main-profile owners. |
| generic `Enumerable` temporary indexing | 5 | Word Frequency, Document Audit, Lexical Rollup, Channel Rollup, and Dependency Plan; material iterator work repeats only in the related Document/Lexical pair. |

JSON has additional parser-specific checks and TapeLang has one parser pop.
Neither is a general lowering candidate.

Most importantly, current numerical and structural hot loops do not retain an
actionable Go check. Matrix Multiply and NBody have zero compiled-body bounds
sites. QuickSort's two index sites belong to imported `read_byte`, not its hot
recursive Array loop. This agrees with the earlier four-application loop gate:
Go already removes the secondary native check after the emitted Able guard at
profiled primitive Array accesses.

Changing `__able_slice_len` spelling does not address the residuals. The
helper is already inlineable, and the remaining checks are tied to nullable
result construction, independent slice relationships, or cold boundary code.

## Escape reconciliation

The broad static leaders are misleading without runtime attribution:

- fieldless `Less` / `Equal` / `Greater` result construction is emitted in all
  32 trees, but no current cross-application profile identifies it as a
  material allocation owner;
- empty/error return values for `Ratio`, `HashMap`, `Array`, `Channel`, and
  `Mutex` repeat because their compiled stdlib bodies are emitted, including
  cold failure exits;
- non-exhaustive-match error construction repeats in 29 applications but is a
  cold semantic error path in verified runs;
- iterator closure creation appears in 20 applications, while fresh profiles
  make generator allocation material in Document Audit and Lexical Rollup but
  not a third unlike application;
- required returned nominal allocations divide into Binary Trees nodes,
  Sudoku candidate arrays, independent Array copies, maps, and other owners
  already separated by allocation profiles.

The strongest new exact static shape is canonical `read_byte`. Its returned
nullable `u8` uses the shared pointer-backed scalar carrier; escape analysis
reports the helper local moved to the heap in all 14 trees that emit it. This
is a language-type representation, not a named-container rule, but emitted
breadth is not enough to change it. The July 13 primitive-result/union gate
found nullable allocation material in String traversal but immaterial beside
numeric Array and generic map controls. Current profiles likewise attribute
Word Frequency to HashMap find plus String split, Document/Lexical to different
iterator/text mixtures, Channel Rollup to scheduling, and Unicode work to its
text pipeline without isolating the nullable scalar allocation in three
unlike applications.

A global nullable carrier or `read_byte` special case therefore does not pass
admission. The check and allocation may share one cause, but a compiler ABI
prototype must wait for dynamic allocation attribution.

## Verification and restoration

- 32/35 selected generated modules completed `-m=2` and SSA bounds-check
  diagnostics.
- All 32 completed diagnostic builds stayed below 22 seconds.
- Regex Suffix, Set, and Stream each hit both the 55-second initial generation
  bound and the independent 59-second retry bound; no process was extended
  beyond one minute.
- Focused native-nullable, Array-helper, and Matrix scalar-loop compiler tests
  pass in 1.280 seconds.
- No production source, canonical stdlib, benchmark, verifier, reference, or
  runtime changed.
- Raw diagnostic logs, generated source trees, binaries, compiler probes, and
  audit caches are cleanup-only.

## Next recommendation

Run a dynamic main-phase allocation-attribution gate for shared nullable
scalar results before designing another compiled ABI candidate.

Why: this census found one exact generic source shape that combines a real
heap move and a residual bounds check across broad emitted code, but static
reachability cannot tell whether it is material. The previous global union
gate rejected a representation rewrite when only String traversal paid the
cost. A current cross-application allocation census can either establish new
broad evidence or close this path without speculative compiler surgery.

What it entails: collect exact main-phase allocation profiles for
I-Before-E, Word Frequency, Unicode Scalar Pipeline, and one non-text nullable
control such as Dependency Plan. Attribute allocation objects and bytes to
`__able_ptr` callers, especially canonical `read_byte`, and retain every
bounded verified process. Advance only if the same nullable-scalar result
shape is materially responsible in at least three unlike applications. If it
passes, prototype a generic internal value-plus-present ABI for statically
direct nullable primitive returns while preserving the existing pointer
carrier at dynamic/host boundaries; test nil propagation, early returns,
pattern matching, callback conversion, and retained-value semantics. Do not
add a String, `read_byte`, Array, map, regex, concurrency, or benchmark-specific
lowering. Continue to defer WASM.
