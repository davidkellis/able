# Bytecode Returned-Nominal Selection Gate — 2026-07-16

## Decision

Close generic returned-nominal transport on the current application corpus and
keep no VM, runtime, compiler, canonical-stdlib, application, or benchmark
source change. JSON and Reverse Complement construct no generic positional
named structs during their measured main calls. Document Audit constructs one
`ArrayIterator`, whose identity crosses return, environment, call, field-read,
and mutation boundaries. None of the three supplies a material
return-to-deconstruction/discard result.

Combined with the retained Word Frequency and K-Nucleotide evidence, only two
unlike programs and one definition (`Utf8DecodeResult`) share the proposed
transport shape. The required three-program/two-definition gate fails, so no
cross-inline-call scalar carrier was designed or benchmarked. Temporary
identity instrumentation was completely removed.

## Method

The gate used two separate one-process passes after the runtime harness loaded,
typechecked, and warmed each source:

1. A trace-free allocation profile established whether generic positional
   struct construction was material.
2. An opt-in one-in-256 identity sampler recorded named-struct creation sites
   and subsequent slot, field, pattern, return, call, collection, aggregate,
   environment, mutation, yield/spawn, alias, and discard events.

Runs used canonical external `able-stdlib`, `ABLE_BENCH_SKIP_TYPECHECK=0`,
`GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`. Document Audit also used
`ABLE_SOURCE_ROOT_ONLY=1`. Reverse Complement output was discarded only at the
host shell after normal execution; the program still performed its full
output work. These diagnostic runs do not replace the promoted five-process
scorecard.

## Allocation gate

| Program | Runtime | Bytes/op | Allocs/op | Main allocation shape |
| --- | ---: | ---: | ---: | --- |
| JSON | 438,585,831 ns/op | 114,802,320 | 216 | input-file bytes; no sampled positional struct allocation |
| Document Audit | 10,250,524 ns/op | 648,728 | 981 | member-cache/environment work; no sampled positional struct allocation |
| Reverse Complement | 6,119,807,114 ns/op | 705,412,176 | 10,894,458 | mono-u8 materialization, integer boxing, and Array growth; no sampled positional struct allocation |

For all three allocation profiles, a focused
`NewStructInstancePositionalSized` query matched zero samples. The JSON byte
volume is one large input buffer rather than many result objects. Reverse
Complement is allocation-heavy, but in a different primitive Array lane.

## Identity census

| Program | Generic positional structs | Sampled definitions | Observed lifetime |
| --- | ---: | --- | --- |
| JSON | 0 | none | none |
| Document Audit | 1 | `ArrayIterator` | return, environment, call, field read, mutation, and discard |
| Reverse Complement | 0 | none | none |

The identity result agrees with the allocation profiles rather than merely
falling below profiler resolution. JSON delegates its parsing work through the
canonical JSON/host boundary without producing this VM struct-literal shape.
Reverse Complement uses primitive Arrays. Document Audit's single iterator is
long-lived protocol state, not a disposable scalar result.

The retained prior census remains:

- Word Frequency: 130,879 `Utf8DecodeResult` constructions;
- K-Nucleotide: 4,233,440 `Utf8DecodeResult` constructions;
- every sampled result crosses return, typed-pattern, field-read, and multiple
  frame-slot boundaries before discard.

That pair is material, but it is still one canonical String helper and one
nominal definition. Turning it into a VM representation rule would optimize a
specific stdlib result protocol without cross-language evidence.

## Candidate gate

No candidate was built because both admission dimensions fail:

- **Generality:** only two unrelated programs, not three, and only one nominal
  definition, not two.
- **Materiality:** the three new programs contribute zero, one, and zero
  relevant constructions respectively.

Generic struct pooling, frame ownership, or a `Utf8DecodeResult` exception
would also violate mutable alias semantics or the rule against nominal-type
special cases. The current runtime representation remains unchanged.

## Next recommendation

Return selection to compiled execution and test whether the generated
backing-slice growth leaf repeats across Reverse Complement, Lexical Rollup,
and Array Slice Window, with Base64 as an already-competitive byte-oriented
guard. First collect post-bootstrap main-phase CPU/allocation profiles and
attribute each `runtime.growslice` caller to generated capacity, append, or
conversion code. Advance only if the same generic capacity-propagation or
backing-growth descendant is material in all three unlike applications; do
not add an Array, benchmark, or named-container lowering rule. This is the
next lead because returned-nominal transport is now closed, while prior
compiled evidence already found concrete slice growth in Reverse Complement
and Lexical Rollup and has not reconciled it against the newer independent
Array Slice Window lane.
