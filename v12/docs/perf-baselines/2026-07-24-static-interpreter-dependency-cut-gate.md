# Static interpreter-dependency cut gate

Date: 2026-07-24

## Decision

Retain an enforceable `ablec --no-fallbacks` option and make the compiled
boundary-audit runner use it for every application. The audit now also records
whether the final Go dependency graph links
`able/interpreter-go/pkg/interpreter`.

Reject and revert the unconditional static interpreter-package cut. It removes
a real fixed startup and binary-size cost and improves six unlike guards, but
it regresses allocation-heavy Binary Trees by 10.60% in an exact-source
comparison. That violates the broad-applicability rule.

No Able syntax, v12 semantics, bytecode VM, tree-walker, canonical
`able-stdlib`, dependency, or WASM change was needed.

## What “no fallback” now means

`able build --no-fallbacks` already enforced the compiler policy, but `ablec`,
which is the compiler entry used by `bench_perf` and the boundary audit, did
not accept that option. Consequently, application audits could not prove that
their generated binaries were built under the strict lowering policy.

`ablec -no-fallbacks` now sets both `RequireNoFallbacks` and
`RequireStaticNoFallbacks`. A focused CLI regression proves that a residual
lowering is rejected with `fallback not allowed`.

`bench_compiled_boundary_audit` now:

- passes `--no-fallbacks` to every compiled build;
- retains verifier status and runtime boundary telemetry; and
- runs `go list -deps` on the final generated module and records
  `interpreter_dependency`.

This closes the tooling loophole between fixture-level strict compilation and
whole-application audits.

## Current lowering versus linking

The strict core audit compiled and verified Fib, Binary Trees,
MatrixMultiply, Quicksort, Sudoku Masks, and I-Before-E. Every row had:

- `strict_no_fallbacks: true`;
- successful verified execution; and
- no interpreter dependency under the experimental package-cut candidate.

The remaining dynamic-boundary telemetry events in those binaries were not
compiled-to-interpreted transitions: the interpreter package was absent.
They were generated runtime, host-output, callable-adapter, or scheduler
events that share the boundary telemetry vocabulary. This distinction is
important: static AST lowering can be complete for a program while generated
code still uses boxed runtime services or host adapters.

The durable core audit is
`2026-07-24-static-no-interpreter-core-reach.json`.

## Package-cut candidate

Static fallback-free generated code retained the interpreter package only for:

- `interpreter.ApplyBinaryOperatorFast`; and
- `interpreter.ApplyUnaryOperatorFast`.

The candidate emitted those calls only for interpreter-bootstrap programs.
Static programs retained their generated Ratio operator and bridge error paths
but did not import the interpreter package. Dynamic/metaprogramming programs
retained the original interpreter behavior.

Six strict static application builds contained no interpreter dependency.
Their binary sizes fell by 32.47% to 37.20%:

| Application | Linked binary | Interpreter-free | Change |
| --- | ---: | ---: | ---: |
| Wide Integer Records | 20,137,528 B | 13,598,208 B | -32.47% |
| Fixed Width 128 | 16,384,632 B | 10,534,472 B | -35.71% |
| Rational Series | 16,463,752 B | 10,603,216 B | -35.60% |
| Unicode Scalar Pipeline | 17,891,336 B | 11,578,824 B | -35.28% |
| Concurrent Packet Codecs | 19,163,848 B | 12,796,240 B | -33.23% |
| Mutex Work Queue | 15,537,528 B | 9,757,552 B | -37.20% |

## Repeated performance gate

Each of the six initial guards used two opposite-order ten-process cohorts per
side, public verifiers, CPU 0, `GOMEMLIMIT=1GiB`, `GOGC=50`, and a 60-second
runtime limit. All 240 processes passed verification.

| Application | Linked mean | Interpreter-free mean | Change | Linked GC | Interpreter-free GC |
| --- | ---: | ---: | ---: | ---: | ---: |
| Wide Integer Records | 0.1765 s | 0.0975 s | -44.76% | 3.75 | 8.45 |
| Fixed Width 128 | 0.2025 s | 0.1155 s | -42.96% | 5.00 | 26.00 |
| Rational Series | 0.1285 s | 0.0620 s | -51.75% | 3.20 | 0.00 |
| Unicode Scalar Pipeline | 0.2680 s | 0.1785 s | -33.40% | 5.00 | 21.35 |
| Concurrent Packet Codecs | 0.4495 s | 0.3825 s | -14.91% | 4.55 | 6.00 |
| Mutex Work Queue | 0.7345 s | 0.6995 s | -4.77% | 4.05 | 11.45 |

The six-row geometric mean improves 34.08%. The higher GC counts show why a
large-allocation guard was still required: removing approximately 38 MB of
interpreter initialization also removes the enlarged initial Go heap goal.

## Allocation-heavy rejection guard

Binary Trees received an exact-source comparison. Both binaries were generated
from the same source tree. The linked variant differed only by restoring the
interpreter import and the two boxed fast-operator calls.

Three reverse-order verified processes per side produced:

| Variant | Mean | Mean GC |
| --- | ---: | ---: |
| Interpreter linked | 29.4900 s | 144.00 |
| Interpreter-free | 32.6167 s | 205.33 |

The interpreter-free binary is 10.60% slower and performs 42.6% more
collections. An independent comparison against the preserved July 22 linked
binary reproduced a 10.77% loss. The package-cut candidate was therefore
reverted completely.

The result does not justify retaining unused interpreter initialization as an
architecture. It identifies the prerequisite: generated nominal construction
currently allocates enough memory that the interpreter's unrelated live heap
acts as accidental GC ballast. GC policy manipulation or replacement ballast
would hide the lowering defect rather than fix it.

## Verification

- Full `go test ./cmd/ablec`: pass in 10.331 seconds.
- Focused static launcher/no-bootstrap compiler gate: pass in 8.947 seconds.
- Strict core application audit: 6/6 compiled, ran, and verified.
- Initial broad guard: 240/240 processes verified.
- Exact Binary Trees rejection guard: 6/6 processes verified.
- Files remain below 1,000 lines.

## Next

Profile and reduce general generated nominal-construction allocation across
Binary Trees, Sudoku Masks, and at least one unlike nominal/union-heavy
application. Then repeat the interpreter-package cut.

This is first because interpreter execution is already absent from the strict
static core programs; the remaining package dependency is linker/startup
baggage, and its safe removal is blocked by a real native-code allocation
problem. Completing nominal lowering means generated structs must have Go-like
allocation and escape behavior rather than allocating through semantic
encoding or pointer-heavy construction on every value.

The next tranche should collect bounded allocation profiles and Go escape
analysis, identify a repeated constructor/field/return escape in the shared
nominal translation pipeline, implement only a general nominal rule, and gate
it across the three unlike allocation shapes plus the six package-cut guards.
After that guard is neutral or faster, remove the two interpreter operator
roots again and expand the strict dependency audit from the core suite to all
portable applications.
