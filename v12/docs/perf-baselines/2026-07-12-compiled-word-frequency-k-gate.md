# Compiled Word Frequency / K-Nucleotide conversion gate (2026-07-12)

## Method

Fresh Go 1.26.4 references and compiled Able each ran three independent,
CPU-2-pinned, verifier-backed processes with `GOMEMLIMIT=1GiB`, `GOGC=50`,
`GOMAXPROCS=1`, and a 45-second limit. JSON was the non-counting control.

| Benchmark | Compiled Able (s) | Fresh Go (s) | Able/Go | Status |
| --- | ---: | ---: | ---: | --- |
| Word Frequency | 0.2633 | 0.0051 | 51.63x | verified 3/3 |
| K-Nucleotide | 5.8033 | 0.0586 | 99.03x | verified 3/3 |
| JSON control | 0.7600 | 1.4582 | 0.52x | verified 3/3 |

Both target rows are material misses while the independently shaped JSON
control remains healthy. The canonical external verifiers accepted every Able
output.

Current compiled binaries built against the sibling canonical stdlib then ran
collector-free CPU and cumulative-allocation captures: eight verified Word
Frequency launches produced 1.23 seconds of CPU samples; two verified
K-Nucleotide launches produced 7.79 seconds. The profile artifacts and
generated binaries were temporary and are not retained.

## Attribution gate

`bridge.ToUint` repeats, but not as an eligible local lowering candidate.

| Workload | Material CPU descendants | `bridge.ToUint` CPU | `bridge.ToUint` allocation |
| --- | --- | ---: | ---: |
| Word Frequency | `String_split` 38.2% cumulative; `__able_hash_map_find_entry` 38.2% flat; HashMap set/get | 8.1% cumulative | 20.5 MB across eight captures |
| K-Nucleotide | count windows; raw HashMap set/get; string validation/conversion; allocation/GC | 16.8% cumulative | 1.075 GB across two captures |

K-Nucleotide also has 30.4% of allocation space in `bridge.ToInt`; Word
Frequency's profile is instead primarily a string split and map-entry-search
shape. Both measurements contain `runtime.convT`, but that is the unavoidable
interface materialization associated with the existing `runtime.Value`
boundary, not evidence for a safe named-collection shortcut.

The result confirms the constraints in
`design/raw-value-boundary-audit.md`: `RawValue` is bytecode-local and cannot
replace `runtime.Value` across compiled functions, nominal values, interfaces,
collections, callbacks, and dynamic boundaries. A HashMap, string-key,
FASTA, small-count cache, or raw-map lowering would make these two benchmarks
faster without proving a language-wide value representation.

## Decision

Keep no compiler, VM, runtime, benchmark, or `able-stdlib` change. This
closes the compiled primitive-boxing hypothesis as a local optimization. The
repeated helper is real telemetry, but its only demonstrated material contexts
are two text/counting HashMap applications and the required general-value
boundary has no semantics-preserving local replacement.

## Next recommendation

Refresh a source-aligned compiled Able-versus-Go status ledger for the
remaining feature families in the `generality` suite, reporting verified rows,
timeouts, and missing references separately. Why: the text/counting
primitive-boundary lane is now both repeated and correctly rejected, while the
current compiler target still needs broad coverage to find a pair that shares a
concrete helper outside a named container or source shape. The work entails
bounded fresh Go and compiled Able processes under the existing guardrails,
then profiles only for two material misses with the same non-nominal helper.
