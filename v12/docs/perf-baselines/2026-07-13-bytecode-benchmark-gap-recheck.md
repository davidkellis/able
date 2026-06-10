# Bytecode Benchmark-Gap Recheck — 2026-07-13

## Decision

Keep no benchmark, VM, compiler, runtime, or canonical-stdlib change. The
proposed independent text/counting discriminator already exists: Word
Frequency is a separately authored, verifier-backed Python/Ruby/Able program
with current three-run coverage. Document Audit and Lexical Rollup provide two
additional independent text/pipeline applications.

Adding a near-duplicate text/counting benchmark would not make the selection
more general. The retained profiles already separate K-Nucleotide’s map/type
matching and raw-integer path from Word Frequency’s HashMap lookup/raw integer
extraction and the direct-text/iterator paths in I-Before-E and Lexical Rollup.
Their common call/return and member-cache parents have already failed broad
optimization guards.

## Evidence

The active feature audit records broad cross-language coverage for arrays,
strings, bytes, files, codecs, callbacks, methods, iterator protocols, static
imports, and host boundaries. The current bytecode extension scorecard also
has fresh three-run Python/Ruby references for Word Frequency, Document Audit,
and Lexical Rollup; all three Able outputs verify.

The established warmed text-pair profiles show distinct concrete leaves:

| Application | Material path |
| --- | --- |
| I-Before-E | direct member/named calls and String containment |
| Lexical Rollup | generator execution, typed patterns, iterator dispatch |
| Word Frequency | HashMap lookup, raw integer extraction, map/type matching |
| K-Nucleotide | call/return/type-match and raw-integer counting |

This is enough independent coverage to reject a new workload-shaped VM change.

## Next Recommendation

Add a cross-language, verifier-backed application for the actual remaining
coverage gap: the Future `Awaitable` protocol. Future Pipeline already covers
cancellation and cooperative yield, in addition to `spawn`, channels,
flushing, and ordinary futures. The new application must make repeated `await`
joins observable while retaining a deterministic cancellation probe, with
equivalent Able, Go, Python, and Ruby implementations, deterministic output,
and no privileged runtime path; it should expand product coverage before it is
ever used for performance selection.
