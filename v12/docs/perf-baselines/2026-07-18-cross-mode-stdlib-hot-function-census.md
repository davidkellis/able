# Cross-mode canonical-stdlib hot-function census (2026-07-18)

## Decision

Keep no compiler, VM, runtime, canonical-stdlib, benchmark, fixture, or
language change. Twenty-five selected bytecode applications supplied complete
bounded first-call traces. No new exact canonical-stdlib function, and no
coherent source-file algorithm family, is material in three unlike selected
applications.

Reconciliation with retained generated-Go CPU profiles finds one genuine
three-application cross-mode owner: `able.math.sqrt` across Distance Field,
RMS Norm, and NBody. That is closed evidence rather than a new candidate. The
generic primitive bridge and compiler imported-raw-call proof are already
retained; the subsequent bytecode scalar-result and typed-native-region
representations improved the three math programs but regressed unrelated
controls and were removed. Repeating either would optimize a known local win
at the cost of other real programs.

The strongest uncovered lead is `able.io.slice_bytes_from_offset`, which
copies the entire payload before the first `write_all` attempt. It is material
in Mandelbrot and Reverse Complement, but not a third unlike application.
Changing it now would weaken the predeclared breadth gate after seeing the
result, so it remains a coverage-gated next direction.

## Admission rule

A source candidate required one fully qualified canonical-stdlib function or
coherent algorithm family to satisfy all of the following:

- at least 100 traced calls and at least 1% of the application's traced call
  sites;
- material execution in at least three unlike applications; and
- corroborating materiality in both bytecode and generated compiled-Go
  evidence.

Application names, broad dispatcher/allocation parents, unrelated functions
from the same package, and three variants of one algorithm did not qualify.
Any candidate also had to preserve shared nominal lowering and could not add a
named-container compiler or VM special case.

## Trace protocol

The selected bytecode manifest contains 27 applications. One interpreter test
binary was built and reused. Each process used the catalog target, working
directory, arguments, executor policy, canonical external stdlib, one logical
CPU, normal typechecking, and a 55-second outer cap.

The ordinary runtime benchmark clears its trace after a warm `main()` call.
That undercounts first-call-only work, so a temporary diagnostic binary moved
the trace reset immediately before warmup and retained warm plus measured call
sites. The source edit was restored immediately after collection. Counts are
therefore attribution shares, not timing samples; stateless programs usually
contribute two equivalent calls, while stateful first-call work remains
visible.

Twenty-five complete traces contain 116,647,436 call-site hits, of which
56,542,508 originate inside canonical stdlib source. K-Nucleotide could not
fit two traced calls below 55 seconds. Option/Result Config's repeated loaded
`main()` reached an ambiguous `or_else` overload in this steady-state harness.
Neither application is broken: fresh normal bytecode processes passed their
public verifiers in 40.88 and 0.90 seconds respectively.

Canonical stdlib identity during collection was 70 Able sources at tree
SHA-256 `f7a470aae4fba342e5bbc3fce53ee26fa6f96df71dde18e057e044520624dafc`
and Git `219eff222c28406487231713753641bc49ee5b9a` (dirty). The selection
manifest's file SHA-256 was
`19c0b7c5c9a41226cfff851b99ffeca46317ff7f8ab608378deb2c66153c06fe`;
its validated semantic SHA-256 was
`9976ccea0e85b2acf92b019727e81b0ce88a347828b3ffb675d869cae81eca7c`.

## Fresh bytecode owner census

Trace entries identify source origin and call-site line rather than a direct
callee object. Each canonical source site was normalized to its enclosing
function definition; source path plus definition line disambiguates overloads
and same-named methods. Primitive/kernel-native calls remained distinct from
Able function owners.

No owner appears in three applications even when the material-share threshold
is relaxed from 1% to 0.01%. At the declared 1% gate, the only repeated owners
are:

| Fully qualified owner | Application share | Application share | Breadth |
| --- | ---: | ---: | ---: |
| `able.io.slice_bytes_from_offset` | Mandelbrot 59.998% | Reverse Complement 33.516% | 2 |
| `able.math.sqrt` | RMS Norm 50.000% | Distance Field 33.333% | 2 |
| `able.core.iteration.filter` | Lexical Rollup 45.137% | Document Audit 30.944% | 2 |
| `able.text.regex_nfa.regex_nfa_upsert_thread` | Regex Set 35.363% | Regex Stream 29.447% | 2 |
| `able.text.regex_nfa.regex_nfa_add_closure` | Regex Set 19.201% | Regex Stream 16.527% | 2 |
| `able.text.regex_nfa.regex_nfa_move` | Regex Set 14.673% | Regex Stream 11.800% | 2 |

Grouping all functions by canonical source file does not create a hidden
three-way family. The repeated material files are `math.able`, `io.able`,
`text/regex_nfa.able`, `collections/array.able`, and
`core/iteration.able`; each reaches the 1% gate in exactly two applications.

## Compiled-profile intersection and closure

Fresh compiled profiling could not promote a function that already failed the
necessary three-selected-bytecode breadth condition, so the intersection used
fingerprint-compatible retained generated-main evidence rather than creating
hundreds of short noisy profiles.

| Owner/family | Cross-mode evidence | Decision |
| --- | --- | --- |
| `able.math.sqrt` | Ten compiled profiles per new math app plus retained NBody profiles established the exact owner in three unlike applications. Reduced-NBody bytecode evidence supplies the third interpreter lane. | Closed: the retained primitive bridge improved compiled Distance/RMS/NBody 19.8%-48.1% and bytecode Distance/RMS 72.1%-78.3%; retained imported raw calls then improved all three compiled programs about 39%-41%. Result-only and typed-native bytecode carriers were later rejected on broad controls. |
| regex NFA transition/closure | Compiled Suffix and bytecode Suffix/Set/Stream profiles establish a real shared regex algorithm. | Related regex consumers do not satisfy the unlike-application rule. The active-state index and thread arena were rejected; successful primitive thread/capture-template improvements are already retained. |
| `able.core.iteration.filter` | Retained compiled Document Audit samples attribute 60% cumulative to iterator generator/filter work; fresh bytecode shares are material in Document and Lexical. | Only two unlike applications, with no third compiled/bytecode owner. |
| `able.io.slice_bytes_from_offset` | Fresh bytecode traces expose full-payload copying in Mandelbrot and Reverse Complement; retained compiled Reverse attributes 310 ms to the same helper. | Strong general I/O lead, but only two unlike applications. No candidate this tranche. |

## Verification and cleanup

- All 25 complete trace processes finished below the one-minute limit.
- The two excluded applications passed fresh normal verifier-backed bytecode
  runs; they are exclusions from the diagnostic trace, not failed benchmarks.
- The focused trace unit tests pass in 0.156 seconds.
- Feature coverage remains 15 families, 16 normative sections, 35 portable
  applications, and three intentional local-only families. Selection and the
  complete 114-program catalog also validate.
- The temporary first-call trace source edit is fully restored.
- Raw trace JSON, binaries, and output logs are cleanup-only and are removed
  after extracting this compact machine record.
- No WASM work was performed.

## Next recommendation

Add a third independent, verifier-backed bulk-output application before
changing `able.io.write_all`; a conventional deterministic FASTA generator is
the best fit.

Why: the census's strongest new cross-mode source lead is not a vague runtime
parent. `write_all` currently copies the complete `Array u8` even at offset
zero, before knowing whether the host write is partial. That work is material
in two very different real applications and in compiled Reverse Complement,
but the project deliberately requires a third unlike consumer before changing
the shared algorithm. A real streaming-output application closes that
evidence gap and improves feature coverage without inventing a microbenchmark
to authorize the optimization.

What it entails: add matched Able, Go, Python, and Ruby FASTA-generation
programs with deterministic byte-for-byte or digest verification; register
the application in the external catalog, selection manifest, and feature
coverage index; collect repeated compiled/bytecode/reference cohorts and
bounded call/CPU attribution. Only if the third application reaches the same
`write_all` copy should canonical `../able-stdlib` try passing the original
payload on the first write and allocating a suffix only after an actual
partial write. Test zero writes, repeated partial writes, broken pipes,
buffered flushes, empty payloads, and ordinary small output, then use repeated
verifier-backed Mandelbrot, Reverse Complement, FASTA, and unrelated text/
numeric guards. Continue to defer WASM.
