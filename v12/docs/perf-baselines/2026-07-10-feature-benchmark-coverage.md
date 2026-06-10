# v12 Feature-to-Benchmark Coverage Audit

This audit maps the active v12 feature surface to the benchmark corpus before
selecting more runtime work. It distinguishes broad cross-language application
coverage from local fixtures: a feature is not performance-comparable merely
because it has an interpreter regression test.

## Inventory

The audit covers 76 local benchmark fixtures, 22 canonical external-style Able
programs, and the sibling cross-language benchmark suites. The full external
catalog remains 15 established comparison programs; independently maintained
Word-Frequency, Document-Audit, and Lexical-Rollup applications supplement it.
New Channel-Rollup is registered as a special suite until an isolated Docker
publication adds reference result rows without overwriting the dirty shared
`../benchmarks/results.json`.

## Feature matrix

| v12 area | Current benchmark coverage | Cross-language status |
| --- | --- | --- |
| Lexing, blocks, bindings, loops, and pattern control flow | all external programs; Sudoku, K-Nucleotide, Tapelang, graph and iterator fixtures | broad |
| Primitive, big, fixed-width, ratio, struct, union, and generic nominal values | MatrixMultiply, Pidigits, Fixed-Width-128, Rational-Series, Array/collection fixtures | broad for numeric/structural values; type-alias/HKT-heavy cases remain local |
| Arrays, strings, bytes, regex, JSON, files, and host-backed stdlib calls | Base64, JSON, I-Before-E, Reverse Complement, K-Nucleotide, Word-Frequency, Document-Audit, Lexical-Rollup | broad |
| Functions, lambdas, callbacks, methods, interfaces, and iterator protocols | Lexical-Rollup, Document-Audit, collection and linked-list iterator fixtures | broad for normal callback/interface traffic |
| Result/error matching and raising | Base64, JSON, regex and String-builder fixtures | success paths are comparable; rescue/ensure/rethrow paths remain local |
| `spawn`, `Future`, channels, and scheduler flushing | BinaryTrees plus channel/future/mutex fixtures | Channel-Rollup now adds an ordinary cross-language channel application; await/cancellation/mutex remain local-only |
| Packages/modules and normal imports | every external Able program imports canonical stdlib modules | broad for static imports; dynamic imports remain local-only |
| Host interop | fs, JSON, Base64, codecs, and OS argument paths | broad at the stdlib boundary; user-authored target-specific extern bodies remain local-only |
| Tooling/test framework | fixture and Docker verification harnesses | verified as infrastructure, not a runtime timing family |

The remaining gaps are deliberately not filled with synthetic loops. In
particular, cancellation/await, mutex contention, dynamic import, and
rescue/ensure need application-shaped workloads with a fair Go/Python/Ruby
counterpart before they can inform a generic optimization.

## Added Channel-Rollup application

Channel-Rollup reads the first 16,384 words of the checked-in ENABLE list,
sends them from a spawned producer through a buffered `Channel String`, filters
and scores them in a spawned worker, flushes the task scheduler, and reduces
the resulting `Channel i64`. It combines file input, Strings, typed channels,
`spawn`, `Future` flushing, pattern matching, and numeric reduction while
remaining independent of the existing direct iterator, map, and tree workloads.

Canonical Able source is
`v12/examples/benchmarks/channel_rollup/channel_rollup.able`; the sibling
`../benchmarks/channel-rollup` package supplies equivalent Able bytecode,
tree-walker, compiled, Go 1.26, Python 3.14, and Ruby 4.0 implementations.
All six Docker images and all five locally invoked runtimes produced the shared
verified output:

```
16384:4828:502100
```

`channel-rollup` and the two-program `concurrency` suite are available through
both `bench_bytecode_audit` and `bench_compare_external`. The lowering audit
finds three functions and 220 instructions with no speculative float/array
opcode shape. A one-run bytecode catalog smoke completes successfully at
0.5200 seconds. The isolated six-image timing artifact is
`2026-07-10-channel-rollup-docker-publication.{md,json}`; it is report-only and
does not modify the shared result source.

All three Able Docker rows use the goroutine executor. The tree-walker now
completes this ordinary channel program under that executor after the generic
array-store registry and interpreter alias-tracking synchronization repair;
the timing publication is therefore a like-for-like concurrent comparison.

## Decision

Keep the new benchmark and its harness integration; keep no VM, compiler, or
stdlib optimization. Channel-Rollup fills the material cross-language
concurrency gap without changing language semantics, adding a named-container
rule, or encoding a benchmark-shaped opcode.

## Next recommendation

Measure the generic synchronization boundary across independent array-heavy
applications before choosing another performance candidate. The concurrent
tree-walker correctness gate is now green, while the bytecode profile pair
still does not identify a shared optimization leaf. The work entails bounded
array, iterator, and channel suite rows in bytecode and compiled modes; retain
the repair only when it is neutral across that breadth.
