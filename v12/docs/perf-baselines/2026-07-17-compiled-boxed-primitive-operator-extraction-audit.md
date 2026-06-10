# Compiled boxed primitive operator extraction audit (2026-07-17)

## Decision

Close the proposed boxed primitive operator package extraction without a
production candidate. The current dependency is real, but it does not define a
small shared semantic unit, and the only benefit in two of the three selected
applications is the same dead interpreter-link removal that already regressed
the stable Binary Trees control.

No compiler, bridge, interpreter, VM, runtime, stdlib, workload, verifier, or
scorecard source changed in this tranche.

## Current boundary

The compiler bridge is already concrete-interpreter-free. Its `Interpreter`
interface contains 60 operations, and static launchers can construct a bridge
runtime with no interpreter. In the audited generated root packages, the only
remaining concrete roots are:

- `interpreter.ApplyBinaryOperatorFast(...)`; and
- `interpreter.ApplyUnaryOperatorFast(...)`.

Those references live in unconditional generic helper definitions. Actual
generated call-site reachability is narrower:

| Application | Binary helper calls | Unary helper calls |
| --- | ---: | ---: |
| Document Audit | 1 | 0 |
| Dependency Plan | 0 | 0 |
| Option/Result Configuration | 0 | 0 |
| Binary Trees guard | 0 | 0 |
| Base64 guard | 0 | 0 |

Document Audit's call is the guarded nullable-`i32` comparison in canonical
`String.substring`. Dependency Plan and Option/Result do not exercise boxed
operator semantics at all; they benefit only if the unused interpreter link is
removed.

## Extraction inventory

`ApplyBinaryOperatorFast` and `ApplyUnaryOperatorFast` are only 451 source
lines, but their semantic closure is spread across the interpreter package:

- integer promotion/range/bit-pattern/float helpers: 315 lines;
- native integer overflow and Euclidean division helpers: 215 lines;
- comparison and string conversion helpers: 113 relevant lines;
- bitwise/shift/divmod implementations in the general arithmetic module;
- operator normalization and primitive classification in dispatch/control;
- ratio detection and interface unwrapping;
- bytecode raw integer/float extraction and raw integer result construction;
- bytecode small-integer box caches; and
- interpreter-private division/overflow/shift error values and their conversion
  to catchable Able standard errors.

The shared numeric/error names occur 137 times across 25 interpreter files.
Raw integer/float helpers occur across 51 interpreter files. Standard runtime
error wrapping occurs at 49 call sites in 20 files. This is not a bounded move
of two top-level functions.

## Rejected package shapes

### Move the existing fast path

Moving the implementation intact would expose bytecode raw carriers, box
caches, and interpreter standard-error machinery to generated applications.
It would invert the intended dependency boundary rather than create a small
primitive package.

### Copy a boxed-only engine

A standalone boxed copy would duplicate integer promotion, overflow,
Euclidean division, bitwise width, float/NaN, string, and unary semantics.
Parity tests can detect known differences, but they do not make two arithmetic
authorities safe to evolve. Error parity is especially structural: the current
interpreter recognizes a private standard-error type and turns it into a
catchable Able error. A copied `error` with the same message is not equivalent.

### Parameterize one engine over raw/boxed results

A callback or mode-parameterized engine would put a new branch/callback on the
VM's hot raw-integer path and require the standard-error contract to move into
another package. That is broader than the selected compiled boundary and would
need a new bytecode-wide performance gate. The present application census does
not justify that risk: only one selected compiled application executes the
boxed helper.

### Retain a small package only to preserve heap/link shape

For Dependency Plan, Option/Result, Binary Trees, and Base64, a clean primitive
package is reachable only from a dead generated helper. The Go linker can
discard the helper and operator code, reproducing the preceding dependency-cut
candidate. That candidate reduced binary size by 35.8%-39.1% but regressed
Binary Trees 3.6% over 15 paired runs. Its hot generated source was unchanged;
the dependency removal necessarily changed package state and binary layout,
but this gate did not attribute the precise mechanism. Adding package
initialization or retained data solely to alter that result would be heap
ballast, not a semantic operator layer, and is explicitly outside the project
bar.

## Verification

The restored operator and compiler boundary remain green:

```text
go test ./pkg/interpreter -run 'TestApplyBinaryOperatorFast|TestBytecodeVM_.*(Overflow|Division|Shift)|TestBytecodeVM_RawInteger' -count=1 -timeout 55s
go test ./pkg/compiler -run 'TestCompilerStaticGeneratedCodeRootsLimitedToOperators|TestCompilerMainSkipsProgramEvaluationWhenStaticAndFallbackFree|TestCompiler.*(DivisionByZero|ShiftOutOfRange|IntegerOverflow|UnaryOverflow)' -count=1 -timeout 55s
```

Both pass, in 0.066 seconds and 0.120 seconds respectively.

## Next direction

The compiled package-initialization branch is now exhausted under the broad
benchmark rule. Do not retry unused-helper omission, operator-package copying,
callback-parameterized VM arithmetic, lazy fixed-integer initialization, heap
ballast, or a nullable-comparison special case from this evidence.

Return to an implementation-level product gap rather than another generated
startup micro-boundary. The strongest explicit gap in the current evidence is
canonical regex: Regex Set and Regex Stream remain more than 100x behind the
faster interpreter reference, and the source-equivalence audit already showed
that foreign references use mature native engines while Able exercises its
portable NFA. The next tranche should refresh bounded profiles for Regex
Suffix, Set, and Stream plus a non-regex text guard, then improve only a shared
NFA transition/state-set/closure descendant in canonical `able-stdlib`. Do not
introduce benchmark-pattern recognition, host-regex delegation, or compiler
special cases for the Regex nominal types.
