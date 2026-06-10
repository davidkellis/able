# Compiled generated-helper emission gate — 2026-07-22

## Decision

Reject the compile-throughput-only generated-helper omission candidate. Retain
no compiler, generated-runtime, bridge, canonical-stdlib, application,
benchmark, language, bytecode, or WASM code change.

The candidate was generic and semantically conservative. It parsed the whole
generated Go package, omitted only a unique private receiverless function with
zero identifier references in any generated file, and kept methods, exported
functions, `main`, `init`, Go-directed/bodyless declarations, the binary/unary
operator closure, and the four boundary-audit accessors used by post-compile
test harnesses. It deliberately retained mutually referenced dead groups.

That proof removed 9.36%-19.15% of generated Go and made the isolated Go
front-end compile 8.52%-17.13% faster in all six unlike applications. The
post-render reach pass itself was not free, however. Six alternating processes
on the volatile Word Frequency row put generation at 34.713 seconds baseline
versus 36.747 seconds candidate. Its 0.483-second isolated compile saving did
not repay the 2.033-second generation cost: the combined front-end path
regressed 3.84%. The broad gate therefore fails even though five other rows
improve.

## Candidate development

The first implementation parsed and reformatted every modified file. It passed
the synthetic reach tests but a broader semantic run correctly exposed four
private boundary-audit accessors referenced by generated harnesses that replace
`main.go` after `Compile` returns. Those accessors became explicit semantic
roots. Full-file reformatting also consumed enough time and memory to erase the
Go compile saving, so that implementation did not reach the final gate.

The final implementation retained the same Go-AST proof but spliced only the
parser-proven declaration and now-unused import byte ranges from source that
was already formatted. It did not run a second formatter. No application,
package, nominal type, container, stdlib API, or benchmark name participated
in either decision.

## Repeated measurements

Generation used immutable pre-change and candidate `ablec` binaries in
alternating order, `ABLE_SOURCE_ROOT_ONLY=1`, the canonical external stdlib,
`GOMEMLIMIT=1GiB`, `GOGC=50`, fresh output directories within the Go module,
and a 55-second process bound. Each application received three processes per
variant; Word Frequency received six because its workstation timings were
volatile.

The Go compile phase invoked `go tool compile` directly on only the generated
root package with the exact exported dependency map. This avoids both cached
root-package results and repeated dependency/link work. Each row received three
alternating processes per variant. Combined values below are the independently
measured generation and root-compile means; they are a front-end diagnostic,
not an application-runtime benchmark.

| Application | Generation baseline / candidate | Root compile baseline / candidate | Combined change |
| --- | ---: | ---: | ---: |
| Fib | 0.447 / 0.497 s | 1.257 / 1.137 s | -4.11% |
| Option/Result Config | 0.880 / 0.940 s | 1.810 / 1.500 s | -9.29% |
| Word Frequency | 34.713 / 36.747 s | 5.670 / 5.187 s | **+3.84%** |
| Future Pipeline | 0.623 / 0.690 s | 1.380 / 1.193 s | -5.99% |
| Concurrent Document Pipeline | 31.470 / 31.290 s | 5.057 / 4.460 s | -2.13% |
| Mutex Ledger | 1.290 / 1.297 s | 1.973 / 1.740 s | -6.95% |

One ordinary complete Go build per generated tree also moved in the expected
direction (candidate reductions of 0.29-1.03 seconds and 8%-12% peak RSS), but
those single builds are validation observations, not the repeated decision
metric. The expensive source-generation regression remains in the complete
compiler path and is not hidden by the later build saving.

## Source result

| Application | Baseline Go bytes | Candidate Go bytes | Change | Functions omitted |
| --- | ---: | ---: | ---: | ---: |
| Fib | 1,106,734 | 894,795 | -19.15% | 414 |
| Option/Result Config | 1,428,238 | 1,248,024 | -12.62% | 420 |
| Word Frequency | 5,065,981 | 4,591,600 | -9.36% | 1,127 |
| Future Pipeline | 1,176,614 | 1,011,189 | -14.06% | 377 |
| Concurrent Document Pipeline | 4,663,210 | 4,212,106 | -9.67% | 1,059 |
| Mutex Ledger | 1,643,312 | 1,438,506 | -12.46% | 458 |

## Semantic and runtime neutrality checks

- All 12 baseline/candidate binaries built and passed their public verifiers.
- After excluding compiler-numbered `main..stmp_N` constants, all six linked
  symbol-name/type/size sets are byte-for-byte identical. Raw symbol row counts
  are also identical within every pair.
- Interpreter initialization keeps the deliberately retained operator/import
  closure. Allocation counts match within every pair. Five pairs also match
  bytes exactly; Future Pipeline varied by 16 bytes with the same 707,318
  allocation count, which is treated as trace variance rather than evidence for
  a runtime claim.
- Dynamic/package-init/interface/source-reexport/extern/concurrent-focused
  compiler tests pass with the candidate. The restored ordinary compiler test
  selection and compiler bridge suite pass after removal.
- No application runtime speedup is claimed; the omitted declarations were
  already linker-dead.

The machine-readable summary is
`v12/docs/perf-baselines/2026-07-22-compiled-generated-helper-emission-gate.json`.

## Next recommendation

Implement one generator-native declaration reach ledger before trying another
emission candidate.

Why: this gate proves that omission itself is broadly useful to the Go
front-end, but reparsing already-rendered multi-megabyte output makes the heavy
Word Frequency compiler path slower. The next mechanism must eliminate that
new pass rather than tune its parser or special-case the heavy application.

What it entails: record generated declaration identities and identifier edges
at the existing emission sites, seed the same semantic roots, close reach once
after generator specialization stabilizes, and skip only unreferenced private
receiverless declarations while they are written. First check that bookkeeping
adds effectively zero time when omission is disabled; then repeat this exact
six-application generation/root-compile gate and the same verifier, symbol, and
initializer checks. Revert if Word Frequency or any other large unlike program
still regresses. Keep runtime scorecard work separate, because this direction
can improve compiler throughput but cannot improve already-linked application
execution.
