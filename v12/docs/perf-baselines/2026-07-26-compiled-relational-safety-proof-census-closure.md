# Compiled relational safety-proof census closure

Date: 2026-07-26

Decision: retain the report-only observer improvements and retain no
production compiler, generated-runtime, runtime, interpreter, VM, stdlib,
benchmark, language, dependency, or WASM change.

## Question

The normalized generated-owner closure found no removable CPU or allocation
owner shared by Tapelang Alphabet, Sudoku Masks, and Fib. This tranche asked
whether a conservative relational proof could nevertheless establish that
material generated safety checks were unreachable in at least three unlike
strict applications.

The admission bar remained:

1. one structural, syntax-independent proof rule;
2. material dynamic work in at least three unlike applications; and
3. only then a production prototype and five-or-more balanced
   verifier-backed A/B/Go comparisons.

Merely finding a statically redundant check did not satisfy the materiality
requirement.

## Report-only observer

`cmd/compiled-control-census` report schema 2 now observes:

- exact and interval-valued static-Array lengths and element domains;
- integer loop/index intervals, including exact identifier loop bounds;
- struct-field intervals;
- non-negative bitwise `AND`, `OR`, and `XOR` domains;
- successful value-return facts separately from error/control returns;
- aggregate facts passed into and returned from direct compiled calls;
- direct-call parameter mutation summaries;
- aliases created by generated temporaries;
- exact counted-loop growth for generated Array appends; and
- explicit strict and safe native-Array bounds conditions.

Unknown calls invalidate aggregate facts. Multiple direct call sites must all
provide a compatible fact. Branches merge conservatively, and only changing
facts widen during recursive fixed points. The observer never changes
generated code or runtime semantics.

New regression tests cover:

- a constructed Array length carried through a direct call and counted loop;
- invalidation at an unknown aliasing call;
- rejection when one of multiple direct call sites has unknown shape;
- struct-field and bitmask intervals crossing a direct call; and
- a successful Array return carried into the next direct call while an error
  return remains excluded.

The focused package completes in approximately 0.004-0.005 seconds.

## Strict corpus protocol

Fresh current artifacts were built for all 61 coverage applications with
`-no-fallbacks`. One smoke process per application was used only for artifact
and public-output verification; it was not used as performance-selection
evidence. All 61 processes passed their public verifiers with no timeout or
failure.

All work lived under disk-backed `/var/tmp` with the reusable disk-backed Go
cache. The 61 compiler binaries were byte-identical:

```text
28697a5adf4f73918f3d83fbcddc211407dc7e539240f64e4127a4e3dd4ddcab
```

Every final dependency graph omits
`able/interpreter-go/pkg/interpreter`. Sixty graphs contain 96 packages;
Base64 contains 119 because of its additional Go packages. The dependency
audit SHA-256 is:

```text
b78d576407e283c8413e310c2c066ea86a15b8c8fe1abaa4bad0a988139f880d
```

The final observer binary SHA-256 is:

```text
3c3fa13563f0ce3529751a4cabfab10d40a2925966ac90b15d9b47c5a3faa545
```

## Census result

The closed-main census found 1,006 reachable generated Array conditions. It
proved 89 conditions in 12 applications:

| Application | Reachable checks | Proven safe | Universal/local | Closed-call only |
| --- | ---: | ---: | ---: | ---: |
| NBody | 52 | 32 | 0 | 32 |
| Config Validation Extraction | 129 | 15 | 15 | 0 |
| Concurrent Policy Callbacks | 17 | 7 | 7 | 0 |
| Log Routing Redaction | 146 | 5 | 5 | 0 |
| Binary Event Log | 26 | 4 | 4 | 0 |
| Concurrent Signal Dispatch | 11 | 4 | 4 | 0 |
| Concurrent Stateful Pipeline | 7 | 4 | 4 | 0 |
| Concurrent Stencil Reduction | 11 | 4 | 4 | 0 |
| Concurrent Transform Chain | 11 | 4 | 4 | 0 |
| Sudoku Masks | 32 | 4 | 0 | 4 |
| Manifest Normalization | 8 | 3 | 3 | 0 |
| Sensor Calibration | 10 | 3 | 3 | 0 |

The other 49 applications have no newly proven reachable bounds condition.
Notably, Tapelang Alphabet still has 13 reachable conditions and zero proven
safe conditions. Its mutable jump/program/Tape invariants remain outside this
conservative proof. Fib has no Array condition at all.

## Dynamic and profile reconciliation

The 89 static proofs divide into two materially different families.

### Fixed-length direct-call facts

NBody's seven Arrays are constructed with exactly five elements in `main`.
The observer carries those shapes into `advance`, `energy`, and
`offset_momentum`, then combines them with exact loop intervals. Its 32 proven
sites execute exactly 22,500,243 times:

- nine sites times five elements times 500,000 `advance` calls:
  22,500,000;
- 105 sites dynamically reached per `energy` call, called twice: 210; and
- 33 dynamic checks in the one `offset_momentum` call.

This is plausibly material and NBody's current five-run scorecard has at most
44.737 ms of positive target excess. The earlier repeated NBody CPU profile
also established material Array guard machinery.

Sudoku's four closed-call proofs cover only the row/column mask reads and
writes during initial clue loading. The public input's first ten puzzles have
226 clues and the benchmark repeats those puzzles ten times, so the four
sites execute exactly 9,040 times. The current repeated CPU profile is instead
led by `find_best_empty`, checked multiply, `bit_count`, and signed division;
the proved initialization sites cannot explain its 1.155-second target
excess. Recursive board, square-mask, and `Position.Choices` facts remain
conservatively unproved.

Therefore this exact interprocedural fixed-length proof is material in one
application, not three.

### Local literal-index facts

The other 53 proven sites are unrolled literal-index reads in application
`main` functions. Each executes once per process. They occur in ten unlike
applications, but removing three to fifteen one-shot comparisons cannot
explain 24.7-139.7 ms of current positive target excess. Those short rows are
already known to be dominated by the fixed launch/service floor.

This proof has broad syntactic reach and no material dynamic reach.

## Admission decision

No candidate passes:

| Proof | Unlike applications | Material applications | Decision |
| --- | ---: | ---: | --- |
| Closed direct-call fixed Array length plus loop interval | 2 | 1 (NBody) | reject: below three |
| Local fixed Array length plus literal index | 10 | 0 | reject: one-shot work |
| Aggregate “bounds proof” parent | 12 | 1 | reject: erases distinct facts and materiality |

Consequently no production prototype or repeated A/B/Go cohort was warranted.
No safety behavior was weakened, no named-container/non-primitive nominal
rule was added, and no interpreter boundary was introduced.

## Verification and cleanup

- `go test ./cmd/compiled-control-census -count=1 -timeout=60s` passes.
- 61/61 strict builds and 61/61 public verifier smoke processes pass.
- 61/61 final graphs omit `pkg/interpreter`.
- The complete `./run_all_tests.sh` handoff passes every contract,
  non-compiler package, all 32 compiler batches, and the 86.045-second final
  bytecode fixture corpus.
- Every observer source file remains below 1,000 lines. The repository-wide
  length checker still reports only pre-existing oversized source, test, and
  generated benchmark files outside this tranche.
- Generated modules, compiler copies, binaries, analyzer reports, and smoke
  outputs were removed after recording the evidence. The reusable
  disk-backed Go cache was retained; the task TMPDIR was emptied.

## Next recommendation

Begin typed bytecode `i32` slot storage as the next report/profile-driven
tranche.

Why: the compiled corpus has now exhausted local profile owners, boundary
owners, semantic-work differences, and conservative relational safety proofs
without finding a general production change material in three unlike
applications. The retained compiler already keeps strict programs
interpreter-free and uses native carriers. The bytecode goal remains far from
Python/Ruby, and the prior corpus census identifies primitive materialization
and slot transport as the remaining architecture-level direction.

What it entails: freeze three unlike bytecode applications with material
`i32` slot traffic and controls, add a typed internal `i32` slot carrier with
sound invalidation at polymorphic/dynamic boundaries, preserve all bytecode
semantics, then require repeated verifier-backed baseline/candidate/Python/
Ruby measurements. Retain it only if one general slot rule improves the broad
cohort.

Why it is important: this applies the same immediate architectural goal to
the VM—keep primitive values in their native carrier and cross boxed/dynamic
boundaries only when semantics require it—while avoiding more single-program
compiled exceptions. Do not begin WASM work.
