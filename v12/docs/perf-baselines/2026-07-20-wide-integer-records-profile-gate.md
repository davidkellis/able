# Wide Integer Records profile gate (2026-07-20)

## Decision

Keep the new Wide Integer Records application, its Go/Python/Ruby references,
verifier, benchmark registration, and the general standalone primitive
stringification correction. Keep no compiler, generated-runtime, bytecode VM,
or canonical-stdlib performance candidate from this tranche.

The third unlike wide-numeric application establishes the required operation
breadth. Fresh profiles nevertheless repeat two already-closed implementation
families rather than supplying invalidation evidence for them:

- compiled programs publish package environments through `SwapEnv` and
  `sync/atomic.StorePointer`; the general execution-context alternatives have
  already caused material N-Body and K-Nucleotide wall-time regressions; and
- bytecode programs repeatedly call `bytecodeRawIntegerValueInfo`; carrier,
  split-extractor, producer-fusion, and store-side variants have already failed
  unlike-program wall-time guards.

Recurrence confirms that these walls matter. It does not overturn their causal
wall-time failures, so no source A/B candidate was admitted.

## Workload and parity

Wide Integer Records parses 12,000 deterministic signed and unsigned decimal
records through the public `UInt128` and `Int128` APIs, validates range
extremes, and computes remainder checksums. It is unlike the checked arithmetic
sequence in Fixed Width 128 and the GCD/division series in Rational Series.

The canonical and sibling Able sources are byte-identical. Tree-walker,
bytecode, compiled Able, Go 1.26, Python 3.14, and Ruby 4.0 all pass the same
verifier and produce:

```text
12000:319817629:953087384
0:340282366920938463463374607431768211455
-170141183460469231731687303715884105727:170141183460469231731687303715884105727
```

## Standalone compiler correction

The first generated application printed its first line and then failed while
stringifying an `Int128`/`UInt128` result. Generated nominal methods had reduced
the wrapper to a primitive `runtime.IntegerValue`, but `bridge.Stringify`
unconditionally required an interpreter. Standalone generated binaries
deliberately have no interpreter.

`bridge.Stringify` now handles only primitive runtime values directly when no
interpreter exists: String, Bool, Char, Integer, Float, and Nil. With an
interpreter it retains the ordinary language-level stringification route, and
non-primitive values without an interpreter still return an error. This is a
general standalone boundary correction, not an `Int128`, `UInt128`, container,
or benchmark-specific lowering rule.

Focused bridge tests cover every primitive fallback and the non-primitive
error. The generated Wide Integer Records binary also passes the public
end-to-end verifier. An earlier broad compiler test regex matched multiple
expensive integration tests and reached its 59-second command ceiling; the
focused package test and verifier-backed generated program are the bounded
correctness evidence.

## Repeated timing evidence

Five-process reference means were 0.0238 seconds Go, 0.0636 seconds Python,
and 0.1579 seconds Ruby. Two independent verifier-backed five-process Able
cohorts measured:

| Cohort | Compiled | Bytecode |
| --- | ---: | ---: |
| A | 0.154 s | 5.110 s |
| B | 0.164 s | 5.092 s |
| Pooled ten-run mean | 0.159 s | 5.101 s |

All 20 Able executions verified. The subsequent complete promoted scorecard
is an independent third five-process cohort: compiled Able 0.206 seconds
against Go 0.0245 seconds (8.41x), and bytecode Able 5.116 seconds against
Python 0.0632 seconds (80.95x) and Ruby 0.1384 seconds (36.97x). Across all
three Able cohorts, the 15-run means are 0.1747 seconds compiled and 5.106
seconds bytecode. The row is a large product miss, not a favorable synthetic
case.

## Three-application profile attribution

Generated programs were built once and profiled through 30 separate verified
main-phase launches per application. The merged samples were 4.63 seconds for
Fixed Width 128, 1.25 seconds for Rational Series, and 2.35 seconds for Wide
Integer Records. `sync/atomic.StorePointer` beneath package-environment
`SwapEnv` was the exact leaf in all three, with 16.85%, 10.40%, and 30.64% flat
CPU respectively.

That exact generic family is already rejected by broad causal evidence. A
fixed execution-context design regressed N-Body 54.7%; a later
package-linkage variant regressed K-Nucleotide 16.6%. The third occurrence
confirms ownership but supplies no reason those unrelated regressions would
disappear. A named wide-type compiler shortcut would also violate the shared
nominal-lowering rule.

One warmed bytecode call per application measured:

| Application | ns/op | B/op | allocs/op | CPU samples |
| --- | ---: | ---: | ---: | ---: |
| Fixed Width 128 | 6,991,443,158 | 1,242,273,864 | 30,858,407 | 6.97 s |
| Rational Series | 3,456,643,619 | 129,989,232 | 1,405,731 | 3.44 s |
| Wide Integer Records | 4,752,583,244 | 652,306,864 | 19,360,660 | 4.73 s |

`bytecodeRawIntegerValueInfo` is the exact VM leaf in all three at 2.87%,
2.91%, and 2.54% flat CPU. Its callers differ: checked UInt arithmetic in
Fixed Width, division/casts/returns in Rational, and parsing, comparisons,
bitwise work, and member dispatch in Records. Previous general carrier,
split-extractor, producer-fusion, and store variants improved counters or one
consumer but failed broad wall-time guards. No new profile fact invalidates
those results.

## Scorecard, breadth, and audit

The promoted complete scorecard contains 71 selected rows and 78 full-status
rows, with five successful Able/reference samples for every selected row. The
selection contains 39 compiled and 32 bytecode rows. The full bytecode lowering
audit covers 118 programs, 449 functions, and 21,285 instructions; Wide
Integer Records contributes five functions and 305 instructions.

Operation depth now has 17 sufficient portable operations, one insufficient
portable operation, and three local-only operations. Wide nominal arithmetic
is sufficient across three unlike applications. No WASM work was performed.

## Next recommendation

Add one unlike portable regex-consuming application, such as deterministic log
routing and redaction, before reopening regex internals. Regex NFA matching is
the only remaining insufficient portable operation depth: Suffix, Set, and
Stream exercise three public APIs but still share one related NFA-audit
workload family.

The next tranche should build a bounded ordinary application with Able,
Go/Python/Ruby references and one verifier, register compiled and bytecode
rows, run repeated five-process cohorts, and then profile it alongside the
existing regex applications. Advance a candidate only if the same exact
generic regex/runtime leaf is material in the unlike application and survives
the complete non-regex guard set. This is the narrowest way to distinguish a
real-program regex wall from an optimization that only accelerates the current
NFA benchmark family.
