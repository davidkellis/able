# Bytecode lambda dependency cache

Date: 2026-07-16

## Decision

Keep a generic dependency-derived cache key for immutable lambda bytecode
programs. The old key used the shared environment-family binding-shape
revision. Temporary call scopes advance that global revision even when they do
not change anything visible from a lambda's definition environment, so hot
lambdas were lowered again on every evaluation.

The retained key consists of the lambda AST identity, the environment-family
state ID, and the exact visible shape of every identifier referenced anywhere
in that lambda. The dependency shape records whether each name resolves and
the identity of a visible struct definition, if any. Identifier names are
collected once per lambda, sorted, and retained under the same lock and bounded
reset policy as the bytecode cache.

This is not a source-only key. Different environment families remain isolated;
referenced binding additions and referenced nominal-definition replacements
invalidate the entry. Unrelated scope topology and ordinary value updates do
not invalidate immutable bytecode. Closure values and runtime generic bindings
remain attached to each `FunctionValue`, rather than to the shared program.

## Coverage-wide miss census

A temporary counter classified every lambda-program lookup as a new AST miss,
environment-state miss, same-state binding-shape revision miss, or eviction
revisit. It ran one warmup plus one measured call in a fresh process for each
portable bytecode application selected by the external benchmark catalog. The
K-Nucleotide process completed its warm call but its second call exceeded the
55-second per-process limit; its warm-call census is included. The other 25
applications completed both calls.

Across 26 applications the original cache made 271,374 lookups and had zero
hits. There were 19 new-AST misses, 271,355 same-state revision misses, no
state-ID misses, no eviction revisits, and 2,116 bounded cache resets.

| Application | Lookups | New AST | Revision misses | Resets |
| --- | ---: | ---: | ---: | ---: |
| Option/Result Configuration | 245,760 | 5 | 245,755 | 1,919 |
| Mutex Ledger | 16,384 | 1 | 16,383 | 127 |
| Await Channel Mux | 5,120 | 5 | 5,115 | 39 |
| Mutex Await Journal | 4,096 | 1 | 4,095 | 31 |
| Document Audit | 6 | 3 | 3 | 0 |
| Lexical Rollup | 6 | 3 | 3 | 0 |
| K-Nucleotide (warm call only) | 2 | 1 | 1 | 0 |

The other 19 selected applications made no lambda-program cache lookup. The
same revision-churn mechanism is therefore material in four unlike programs:
generic Option/Result callbacks, channel selection, mutex journaling, and a
mutex-protected ledger.

## Rejected cause hypothesis

A first generic trial seeded the temporary parameter-analysis environment in a
single construction step instead of calling `DefineWithoutMerge` per
parameter. Census counts and benchmark behavior were unchanged. Temporary
subphase counters then showed no binding-shape delta inside lambda lowering;
the shared revision had already changed while constructing ordinary call
scopes. The trial was removed before the retained candidate was built.

## Clean repeated A/B

Production test binaries contained neither census locks nor diagnostic output.
Baseline and candidate processes were alternated. Each process loaded and
typechecked once, warmed `main`, then timed one complete call with canonical
`../able-stdlib`, source-root-only resolution, `GOMEMLIMIT=1GiB`, `GOGC=50`,
and `GOMAXPROCS=1`. The four material applications used three independent
processes per side; the two low-frequency lambda guards used five per side.
The table reports arithmetic means, as required for this shared workstation.

| Application | Runs/side | Baseline ns/op | Candidate ns/op | Time | Baseline B/op | Candidate B/op | Bytes | Baseline allocs/op | Candidate allocs/op | Allocs |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Option/Result Configuration | 3 | 4,269,737,280 | 2,022,557,325 | -52.63% | 983,574,517 | 317,412,624 | -67.73% | 7,741,686 | 4,119,373 | -46.79% |
| Mutex Ledger | 3 | 460,253,707 | 224,297,252 | -51.27% | 237,784,939 | 32,171,669 | -86.47% | 814,078 | 534,830 | -34.30% |
| Await Channel Mux | 3 | 87,247,750 | 48,421,605 | -44.50% | 30,861,688 | 12,824,717 | -58.44% | 287,648 | 192,208 | -33.18% |
| Mutex Await Journal | 3 | 92,378,080 | 57,574,964 | -37.67% | 26,316,008 | 10,654,200 | -59.51% | 252,600 | 199,139 | -21.16% |
| Document Audit | 5 | 10,695,994 | 10,682,631 | -0.12% | 648,728 | 639,037 | -1.49% | 981 | 894 | -8.91% |
| Lexical Rollup | 5 | 122,873,702 | 118,722,190 | -3.38% | 2,467,328 | 2,457,621 | -0.39% | 14,990 | 14,902 | -0.59% |

The material wins repeat across four different language-feature families, and
the low-frequency guards are neutral to favorable. Candidate census screens
also changed the four material programs from zero hits to all but their first
per-AST lookups hitting: 245,755/245,760, 16,383/16,384, 5,115/5,120, and
4,095/4,096 respectively.

## Correctness and verification

Focused tests cover slot-lowered lambdas, generic lambdas and runtime type
bindings, reuse across unrelated binding additions, invalidation after a
referenced binding appears, reuse across an ordinary referenced-value update,
and invalidation after a referenced struct definition changes identity.

Verification completed:

- focused lambda and frame-layout tests: pass;
- `go test ./pkg/runtime -count=1 -timeout 60s`: pass;
- interpreter package excluding the two known fixture-harness blockers:
  pass in 30.275 seconds;
- fresh source-built bytecode CLI output passed the external Ruby verifier for
  Option/Result Configuration, Await Channel Mux, Mutex Await Journal, and
  Mutex Ledger.

The verified stdout SHA-256 values are:

| Application | SHA-256 |
| --- | --- |
| Option/Result Configuration | `28e46b27a6dceeaa15968e9db7a6728f4a2b35f87a89ff7bf561db18cad53112` |
| Await Channel Mux | `0a32fa2e9481ecb635562079d993288d911326d9f36aeea0d9d3058d1a494693` |
| Mutex Await Journal | `e6f87830a6a1bb1bf3889a6b77eae053bfd3630055a06dd8fd500201586bf61e` |
| Mutex Ledger | `58f2e2fda882a7c1e5032a2f34c4d9d05f40ced51199c0be0672ecc4d4cf14e4` |

The unfiltered interpreter package did not complete inside the mandatory
one-minute ceiling. `TestExecFixtureParity/14_22_regex_set_iter` was still in
the parser's pre-existing deeply nested expression recursion at timeout. With
that long harness skipped, the existing
`TestFixtureParityStringLiteral/errors/alias_reexport_impl_ambiguity` expected
an unqualified error fragment while the interpreter returned the now-qualified
interface name. Neither failure enters the changed lambda cache. Excluding
both fixture harnesses leaves the rest of the package green.

All temporary census/stage hooks, A/B binaries, benchmark outputs, and census
profiles are removed after this record. No compiler, canonical-stdlib,
application, benchmark-source, verifier, reference, or spec change was needed.

## Next recommendation

Refresh bounded clean CPU and allocation profiles for Option/Result
Configuration, Await Channel Mux, Mutex Await Journal, and Mutex Ledger after
the retained cache change, then admit a new candidate only if the same exact
residual descendant is material in at least three of them.

Why: repeated lambda lowering was the dominant shared wall and has now largely
disappeared. The old profiles no longer describe the programs' real hot paths,
and the remaining Option/Result, channel, and mutex work may diverge. Profiling
the post-change state prevents optimizing a wall that the cache already
removed or turning one workload's next bottleneck into a special case.

What it entails: one warmed CPU process and one separate allocation process per
application under the same 55-second/1-GiB guardrails, followed by exact-stack
reconciliation across generic-union/type-match, call-environment, member/map
lookup, return, scheduler, channel, and mutex descendants. Trial only one
generic mechanism that repeats across three unlike verifier-backed programs,
and use alternating repeated A/B means plus the two low-frequency lambda guards
before retaining it.
