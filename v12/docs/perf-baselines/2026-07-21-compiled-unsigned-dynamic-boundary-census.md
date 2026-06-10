# Compiled Unsigned Dynamic-Boundary Census

Date: 2026-07-21

## Decision

Keep no compiler, bridge, runtime, stdlib, or benchmark change. The unsigned
dynamic-value boundary is a large concrete allocation owner in K-Nucleotide,
but it is not material in any of seven other unlike unsigned-heavy compiled
applications. It therefore fails the required three-application breadth gate.

No unsigned cache or named HashMap rule was prototyped. Building a candidate
after finding only one material application would measure a K-Nucleotide
optimization rather than a generally demonstrated language implementation
improvement.

## Census scope

Eight current applications were generated and built with one compiler before
profiling:

- K-Nucleotide: `u64` rolling keys crossing specialized map operations;
- Reverse Complement: native `u8` input and transformation arrays;
- Base64: native `u8` encoding, decoding, and digest data;
- Fasta Generation: native `u8` generators and output;
- Mandelbrot: native `u8` pixel output;
- Fixed Width 128: native `u64` limbs and wide arithmetic;
- Wide Integer Records: native `u64` fields, parsing, and wide arithmetic; and
- Sudoku Masks: native `u8` masks in the search loop.

This covers map keys, text/byte transforms, encoders, generators, numeric
output, wide integers, parsing, and bit-mask search. Searching the complete
portable application source set found K-Nucleotide as the only benchmark that
uses an unsigned primitive as a `HashMap` key; adding another benchmark solely
to manufacture the same profile shape would not be independent breadth
evidence.

All generated trees contain general unsigned conversion wrappers because they
embed the shared runtime. Generated presence alone was not counted as
materiality. Each application had to execute `bridge.ToUint` materially in its
verified measured main phase.

## Escape classification

Current Go escape and inlining output reports `bridge.ToUint` at inline cost
109, above the inline budget of 80. Its small-integer result escapes when
returned as `runtime.Value`; the `big.Int` fallback also escapes. A call that
actually reaches this bridge therefore allocates, but native unsigned code
that never crosses the dynamic boundary does not pay that cost.

This distinction explains the census: most applications retain their hot
unsigned values in native variables and specialized arrays. Their generated
fallback and conversion wrappers are cold. K-Nucleotide alone sends a hot
native `u64` through the dynamic boundary on every specialized map get/set.

## Exact main allocation census

Seven applications completed exact start/end main-phase allocation snapshots
under one logical CPU, `GOGC=50`, a 55-second cap, and their public Ruby
verifier. `ToUint` and `runtime.NewSmallInt` were absent even with pprof node
and edge thresholds set to zero.

| Application | Main bytes | Main allocations | Main GCs | Concrete owner |
| --- | ---: | ---: | ---: | --- |
| Base64 | 2,201,553,528 | 128 | 21 | large native byte backing/output |
| Fasta Generation | 1,058,528 | 445 | 0 | strings and output backing |
| Fixed Width 128 | 35,536,192 | 2,220,984 | 3 | loop-carried wide nominal results |
| Mandelbrot | 86,320 | 66 | 0 | pre-sized pixel/output storage |
| Reverse Complement | 9,314,944 | 64 | 1 | input and transformed byte backing |
| Sudoku Masks | 156,370,688 | 7,802,594 | 11 | `find_best_empty` result flow |
| Wide Integer Records | 10,737,920 | 640,052 | 1 | signed/unsigned parsing and byte conversion |

The allocation-heavy counterexamples are especially important. Fixed Width's
2.22 million objects are split between `modular_add_checksum` and
`ordered_select_checksum`; Sudoku's 7.79 million dominant objects belong to
`find_best_empty`; Wide Records is led by `parse_unsigned`, `parse_signed`, and
`String.bytes`. None descends from the unsigned bridge.

Base64's 2.20 GiB with only 128 allocations is native buffer growth rather
than scalar boxing. Reverse Complement and Mandelbrot similarly prove that
hot `u8` work need not cross the dynamic value boundary.

## K-Nucleotide outlier

Exact one-object sampling makes K-Nucleotide exceed the bounded profiling
contract, so it used a normal cumulative allocation profile plus a separate
main-only CPU profile. Both runs completed in under three seconds and passed
the public verifier.

| K-Nucleotide allocation metric | Count/share |
| --- | ---: |
| Total sampled objects | 23,037,298 |
| `bridge.ToUint` objects | 7,493,290 / 32.53% |
| `bridge.ToUint` sampled bytes | 343.02 MiB / 28.64% |
| `HashMap.raw_get` callers | 3,550,028 / 47.38% of `ToUint` |
| `HashMap.raw_set` callers | 3,943,262 / 52.62% of `ToUint` |

Both generated call sites box `u64` keys. The application invokes the map only
for frequency windows of length one and two, so every hot key is in `0..15`
and is reused millions of times. Longer target searches remain native and do
not use the map. This is an excellent single-application cache shape, but no
second material consumer exists in the current suite, much less the required
three unlike applications.

The main CPU profile is dominated by primitive map equality/hash work and Go
allocation/GC descendants. That is consistent with the allocation evidence,
but it does not turn the same map operation into a cross-application owner.

## Gate result and verification

- 8/8 generated application binaries built successfully.
- 8/8 allocation-profile outputs passed their external verifier.
- K-Nucleotide's separate main-only CPU output also verified.
- `go test ./pkg/profilehook ./pkg/compiler/bridge -count=1 -timeout 60s`
  passed.
- The focused generated boundary-routing compiler test passed.
- No candidate was admitted, so candidate A/B timings and unrelated runtime
  guards were intentionally not run.
- No canonical stdlib, compiler, bridge, runtime, bytecode VM, benchmark,
  fixture, reference, scorecard, spec, or WASM source changed.

## Next recommendation

Reconcile the promoted compiled performance frontier after the retained
dynamic-`i64` change before selecting another implementation candidate.

Why: the retained boundary cache materially changed K-Nucleotide, Inventory,
and Validated Job, while this census closes the immediately adjacent unsigned
idea. The checked-in aggregate still describes the pre-change compiled
ranking, so choosing another hotspot from it risks optimizing a wall that has
already moved.

What it entails: run five independent verifier-backed compiled processes for
the selected portable application suite under each catalog execution contract,
reuse or refresh source-matched Go references, rebuild the target-excess and
feature-interaction frontier, and profile only exact owners repeated in at
least three unlike remaining misses. Advance a candidate only after that
breadth gate, then repeat the established startup, allocation-heavy,
allocation-light, dynamic-fallback, and bytecode semantic guards. Do not add
named-container rules or resume WASM work.
