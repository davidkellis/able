# Static interpreter-package cut retained

Date: 2026-07-24

## Decision

Retain the static interpreter-package cut. A fallback-free application that
does not require interpreter bootstrap no longer imports or links
`able/interpreter-go/pkg/interpreter`. Dynamic/metaprogramming programs retain
the interpreter import and the original binary/unary operator semantics.

The widened strict audit found two general lowering gaps and both were fixed:

- a primitive unary operand inherited an outer `Result` union expectation,
  forcing native `i64` through `runtime.Value` before negation; and
- native `Result` equality fell back to boxed comparison because
  `runtime.ErrorValue` contains a Go map.

Neither fix names a benchmark, container, Option, Result, or stdlib type.
Unary operators now preserve their natural primitive carrier before the shared
outer union/interface coercion. Native union equality now implements the
reference interpreter rule that runtime error values never compare equal,
while comparable members use their native Go comparison.

No Able syntax, language semantics, bytecode VM, tree-walker, canonical
`able-stdlib`, dependency, or WASM change was needed.

## Strict semantic and dependency audit

All 29 applications compiled with `ablec --no-fallbacks`, passed their public
verifier, and reported no final interpreter dependency:

- generic-union/callable: Option/Result Config, Validated Job Pipeline, and
  Dependency Wave Validation;
- core: Fib, Binary Trees, Matrix Multiply, Quicksort, Sudoku Masks, and
  I-Before-E;
- data/numeric/text: Array Slice Window, Lexical Rollup, Word Frequency,
  Binary Event Log, Fixed Width 128, Rational Series, Wide Integer Records,
  Unicode Scalar Pipeline, Base64, JSON, K-Nucleotide, and N-Body; and
- nominal/concurrency: Manifest Normalization, Policy Record Dispatch,
  Sensor Calibration, Concurrent Text Index, Concurrent Document Pipeline,
  Concurrent Event Routing, Concurrent Packet Codecs, and Mutex Work Queue.

Five rows initially exhausted the audit runner's 60-second preparation budget.
They were rebuilt sequentially without a preparation timeout to avoid memory
pressure; each then verified within the 60-second execution limit and had no
interpreter dependency. Sensor Calibration was the only real execution
failure. Its generated `parse_i64` originally contained one boxed unary call;
after the general unary correction it contains one native checked signed
subtraction, zero boxed unary calls, and verified in three final processes.

A wider compiler test exposed the separate generic `Result<u32>` equality
case. The specialized matcher now contains no `__able_binary_op`; both success
and error-member execution guards pass.

## Binary size

The current strict binaries remain 32.49%-37.23% smaller than the preserved
interpreter-linked binaries:

| Application | Linked bytes | Final bytes | Change |
| --- | ---: | ---: | ---: |
| Wide Integer Records | 20,137,528 | 13,593,952 | -32.49% |
| Fixed Width 128 | 16,384,632 | 10,530,184 | -35.73% |
| Rational Series | 16,463,752 | 10,598,920 | -35.62% |
| Unicode Scalar Pipeline | 17,891,336 | 11,578,656 | -35.28% |
| Concurrent Packet Codecs | 19,163,848 | 12,792,008 | -33.25% |
| Mutex Work Queue | 15,537,528 | 9,753,272 | -37.23% |

## Repeated performance guard

The six established guards used five verified processes each on CPU 0 with
`GOMEMLIMIT=1GiB`, `GOGC=50`, and a 60-second execution limit. Wide Integer
Records received a second five-process cohort because its first result was
volatile. All 35 processes verified.

| Application | Preserved linked mean | Final mean | Change |
| --- | ---: | ---: | ---: |
| Wide Integer Records | 0.1765 s | 0.1020 s (10 runs) | -42.21% |
| Fixed Width 128 | 0.2025 s | 0.1140 s | -43.70% |
| Rational Series | 0.1285 s | 0.0640 s | -50.19% |
| Unicode Scalar Pipeline | 0.2680 s | 0.1820 s | -32.09% |
| Concurrent Packet Codecs | 0.4495 s | 0.3860 s | -14.13% |
| Mutex Work Queue | 0.7345 s | 0.6880 s | -6.33% |

The six-row geometric mean improves 33.24%. These current measurements also
match the earlier exact package-cut candidate cohorts closely, so the unary
and native-union corrections do not introduce a broad regression signal.

The allocation-heavy Binary Trees guard used three verified processes per
implementation under normal Go GC, `GOMAXPROCS=4`, `GOMEMLIMIT=1GiB`, and
CPUs 0-3:

| Implementation | Mean | Mean GC |
| --- | ---: | ---: |
| final interpreter-free Able | 7.4867 s | 86.0 |
| equivalent Go 1.26 source | 7.6500 s | 84.0 |

Able is 2.13% faster in this cohort. This confirms the earlier conclusion that
the `GOGC=50` rejection was an accidental heap-ballast effect rather than a
native nominal representation defect.

## Verification

Passing bounded checks:

```text
go test ./cmd/ablec -count=1 -timeout 60s

go test ./pkg/compiler \
  -run '^(TestCompiler(BroadNativeUnionExecutes|IfExpressionMixedBranchesInferNativeUnion|MatchExpressionMixedClausesInferNativeUnion|JoinFlattensExistingNativeUnionMembers|NestedResultUnionStringLiteralStaysNative|NestedResultUnionStructLiteralStaysNative|ResultReturnUsesNativeCarrier|ResultPropagationExecutes|NativeResultErrorEqualityMatchesReference|GenericExpectationResultCarrierExecutes|UnaryPrimitiveBranchPreservesCarrierBeforeResultWrap|StaticGeneratedCodeOmitsInterpreterRoots|BootstrapMainRetainsPackageInterfaceDefaultASTMetadata))$' \
  -count=1 -timeout 60s
```

The first command passed in 5.820 seconds and the second in 4.145 seconds.
Production and modified test files remain below 1,000 lines; unary lowering
was split into `generator_unary.go`.

`TestCompilerCanonicalStdlibExpectationResultArgumentStaysConcrete` separately
exceeds the required 60-second test budget while repeatedly refreshing native
interface adapters during join inference. It is a compile-time scalability
failure, not an application execution or package-cut semantic failure, and
must be corrected next.

## Evidence

- `2026-07-24-static-package-cut-final-wide-integer-a.json`
- `2026-07-24-static-package-cut-final-wide-integer-b.json`
- `2026-07-24-static-package-cut-final-fixed-width.json`
- `2026-07-24-static-package-cut-final-rational.json`
- `2026-07-24-static-package-cut-final-unicode.json`
- `2026-07-24-static-package-cut-final-packet.json`
- `2026-07-24-static-package-cut-final-mutex-work.json`
- `2026-07-24-static-package-cut-final-binarytrees-able.json`
- `2026-07-24-static-package-cut-final-binarytrees-go.json`
- `2026-07-24-static-package-cut-final-sensor.json`

## Next recommendation

Fix the canonical-stdlib native-interface specialization fixed point so every
individual compiler test and representative build completes within 60 seconds.
This is next because it is the only concrete broken gate remaining from this
tranche and because the same repeated adapter refresh inflates real application
build time.

The work should profile/count adapter refreshes for the canonical
`BigInt -> Result<i32> -> Expectation<Result<i32>>` path, cache completed
interface/actual/method shapes, and prevent `commonJoinInterfaceType` from
re-entering unchanged adapter discovery. It must preserve late specialization
discovery and generic constraint semantics, then pass the timed canonical test,
the generic Result/Matcher guards, and unlike Array/Iterator/interface builds.
After that, refresh compiled runtime profiles now that static binaries are
interpreter-free and select the largest repeated generated-runtime boundary
across at least three unlike applications.
