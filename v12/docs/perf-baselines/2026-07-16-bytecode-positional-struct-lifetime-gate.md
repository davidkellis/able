# Bytecode Positional-Struct Lifetime Gate — 2026-07-16

## Decision

Keep no VM, runtime, compiler, canonical-stdlib, application, or benchmark
source change. The repeated positional named-struct allocator is real, but its
hot values are not frame-local. The dominant Word Frequency and K-Nucleotide
values cross an inline return boundary before pattern deconstruction; the
dominant Regex Set and Stream values additionally cross calls and collection
storage. No hot definition or safely non-escaping lifetime shape repeats in
three unlike programs, so no scalar-replacement or pooling candidate is
admitted.

Temporary diagnostics were completely removed after collection. Struct fields
remain mutable and struct functional update remains fresh-allocation behavior,
as required by the v12 specification.

## Method

An opt-in, off-timing diagnostic sampled one in every 256 constructions at
each `StructLiteralNamedFast` source site. It keyed samples by runtime struct
identity and recorded:

- definition, source origin, location, and immediate next opcode;
- frame-slot storage and traversal through more than one slot index;
- field read, pattern, member access, mutation, and discard;
- call, collection, aggregate, environment, spawn/yield, and return
  boundaries.

The ordinary runtime harness loaded and typechecked each source once, warmed
`main`, reset the diagnostic, and then ran one measured `main` call. Runs used
canonical external `able-stdlib`, `ABLE_SOURCE_ROOT_ONLY=1`,
`ABLE_BENCH_SKIP_TYPECHECK=0`, `GOMEMLIMIT=1GiB`, `GOGC=50`, and
`GOMAXPROCS=1`. These deliberately instrumented times are not performance
measurements and do not replace the promoted external scorecard.

## Construction census

| Program | Total constructions | Sampled | Dominant definitions |
| --- | ---: | ---: | --- |
| Word Frequency | 130,880 | 513 | `Utf8DecodeResult` 130,879 |
| Regex Set | 69,725 | 326 | `RegexNFAThread` 65,392; `RegexCodePoint` 3,810 |
| Regex Stream | 59,925 | 307 | `RegexNFAThread` 41,952; `RegexCaptureTag` 8,752; `RegexCodePoint` 8,585 |
| K-Nucleotide | 4,233,470 | 16,540 | `Utf8DecodeResult` 4,233,440 |

The allocator's apparent four-program commonality therefore splits into two
different semantic families. `Utf8DecodeResult` repeats in two unrelated text
consumers. `RegexNFAThread` repeats in the two related NFA applications. No
hot definition repeats across three unlike programs.

## Lifetime census

Every one of the 17,049 sampled `Utf8DecodeResult` values in Word Frequency
and K-Nucleotide was returned from `utf8_decode`, pattern-tested in its caller,
read by field, stored through caller/callee frame slots, and then discarded.
None of those samples entered a collection, aggregate, environment, mutation,
or unrelated call. This is not frame-local allocation, although it may be a
future candidate for a general cross-inline-return scalar transport if a third
unlike application and static proof repeat the same shape.

Every sampled hot `RegexNFAThread` in Set (256/256) and Stream (164/164)
crossed call and collection boundaries in addition to field reads, patterns,
returns, and frame slots. The source confirms the semantic need: the generic
NFA upsert routine writes the fresh thread into an Array, returns it, or both.
The thread contains an Array capture field and is mutable, so pooling or
frame-owning it would be unsafe without a materially different ownership
proof.

Across all samples, only one low-frequency parser-expression site in each
regex program had no observed return/call/aggregate/mutation/collection
boundary. Each site executed 26 or fewer times. They are neither material nor
shared, and cannot justify a runtime path.

## Candidate gate

No implementation candidate was built:

- generic struct pooling violates identity and mutable alias semantics;
- frame-owned values cannot represent the observed return and collection
  boundaries;
- specializing `Utf8DecodeResult`, `RegexNFAThread`, or a named container
  would violate the broad-optimization and nominal-lowering rules;
- a general cross-frame scalar return needs a third unlike hot consumer plus
  a static proof that no identity, mutation, aggregate, collection, dynamic
  call, closure, or concurrency boundary observes the materialized value.

## Next recommendation

Run a returned-nominal selection gate across three unlike existing
applications outside the regex family—JSON, Document Audit, and Reverse
Complement—alongside the retained Word Frequency/K-Nucleotide evidence.
Attribute hot struct results that follow the exact return → typed-pattern or
field-read → discard shape, and require at least three programs plus at least
two different nominal definitions before designing anything. If that gate
passes, the implementation work would be a generic lowering/dataflow proof and
cross-inline-call scalar transport that materializes at every observable
escape; otherwise close returned-nominal transport and return to compiler
residual selection. This is next because the current diagnostic found a large
allocation shape that is potentially removable across a call boundary, but
also proved that ordinary frame ownership and pooling are unsafe.
