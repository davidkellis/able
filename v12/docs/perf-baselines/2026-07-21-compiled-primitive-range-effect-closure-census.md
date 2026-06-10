# Compiled primitive range/effect closure census

Date: 2026-07-21

## Decision

Retain the conservative range/effect extension to the report-only
`compiled-control-census` analyzer. Keep no compiler, generated-runtime,
bridge, stdlib, application, reference, language, or WASM performance change.

Primitive range closure adds three universally control-free direct bodies
across the seven-application cohort. Two are cold, main-unreachable numeric
sign helpers emitted with K-Nucleotide's stdlib closure. The third is
Mandelbrot's `pixel_byte`, whose checked shift is universally safe because its
local bit loop proves a shift in `[0,7]`. The prior main profile samples no
material control check around that function, and Mandelbrot's 120-ms profile
is selection-only.

No application contains a main-reachable direct function that is fallible for
the full parameter domain but control-free for every observed internal direct
call. The call-site-specialization count is zero in all seven rows. Fib remains
the sole application with a material eligible control-ABI cost, so the
three-unlike-family admission gate fails and no value-only ABI prototype is
justified.

## Retained analyzer contract

The existing generated-Go control fixed point now reports two additional
conservative closures:

- **universal range closure** starts every primitive parameter at its complete
  representable domain and discharges a primitive control dependency only when
  all corresponding operations are proven safe;
- **closed direct-call closure** starts at `__able_compiled_fn_main`, excludes
  entry wrappers and host/dynamic callers, propagates integer argument
  intervals through main-reachable direct calls, and computes the same effect
  closure over the aggregate interval seen at every callee.

The analyzer understands generated integer literals and casts, bounded
addition/subtraction/multiplication, signed and unsigned checked helpers,
division/modulo nonzero and signed-minimum rules, shift count and result bounds,
inline widened overflow guards, terminating condition refinements, direct
recursion, and common generated counted-loop guards.

Loop conditions may bound an induction variable. The initial lower bound is
retained only when the loop has one structurally direct positive update;
otherwise the lower side widens. Other loop-carried integer values are
invalidated before analyzing the body. This prevents a first-iteration zero
from incorrectly proving a growing mask or accumulator safe. Unknown
expressions, host calls, container elements, object fields, unsupported
control-producing shapes, and nonrepresentable unsigned ranges remain
unproven. A 4,096-update cap widens recursive observations to the full declared
type rather than trusting nonconvergence.

The extension changes JSON evidence only. It does not rewrite generated Go or
alter compiler decisions. Classifications are based on syntax, types, ranges,
and call reach—not package, function, container, stdlib, or benchmark names.

## Protocol

The same frozen compiler used by the preceding exact-leaf and control-effect
tranches was rebuilt at SHA-256
`87535396e8c8a471f3dd84d44ce7678554426f1ac84d5e730d729349f83fa913`.
It emitted fresh normal Go modules against canonical `../able-stdlib` with
`ABLE_SOURCE_ROOT_ONLY=1`, `GOMAXPROCS=1`, `GOGC=50`, a 1-GiB memory limit, and
a 59-second cap. Generation and analysis were serialized.

No compiler or runtime code changed, so the immediately preceding two
verified main-only CPU processes per application remain the source-exact
materiality evidence. Repeating identical workstation timing would add noise,
not a new executable comparison. The new analyzer was run over all seven fresh
modules after its final loop-carried-state guard.

## Fixed-point census

| Application | Base control-free / direct | Universal range-free | Main-reachable direct | Closed range-free | Call-site-specializable |
| --- | ---: | ---: | ---: | ---: | ---: |
| `fib` | 10 / 53 | 10 | 2 | 1 | 0 |
| `sudoku_masks` | 58 / 287 | 58 | 9 | 1 | 0 |
| `base64` | 11 / 63 | 11 | 3 | 0 | 0 |
| `pidigits` | 75 / 232 | 75 | 7 | 0 | 0 |
| `k_nucleotide` | 90 / 416 | 92 | 22 | 5 | 0 |
| `tapelang_alphabet` | 61 / 291 | 61 | 13 | 4 | 0 |
| `mandelbrot` | 36 / 175 | 37 | 4 | 1 | 0 |

`Closed range-free` includes functions already control-free. The final column
counts only functions newly safe under all main-reachable direct-call ranges;
it is the specialization admission column and is zero throughout.

## Hot blocker reconciliation

### Fib

The recursive parameter closure is `[1,45]`, matching the guarded recurrence
and its literal root call. The compiler already used this fact to remove
recursive arithmetic overflow checks, so `fib` was base control-free before
this tranche. Its nil-control propagation remains the one material ABI owner
at 18.13% of merged CPU.

### Sudoku Masks

The direct divisions by literal three in `square_index` are universally safe.
The subsequent multiply/add chain is not universally safe for arbitrary i32
rows and columns. Main-reachable search paths also recover row, column, mask,
and choice values through Arrays. A primitive interval census cannot preserve
the element provenance needed to prove those values are Sudoku indices and
9-bit masks.

Consequently `bit_count` correctly retains the `i32` minimum-minus-one
overflow case, and `square_index` remains fallible. Proving the benchmark's
container invariants would require a separate relational Array-element proof,
not a primitive call-range extension. The prior matching control checks are
only 0.25% of CPU, so that larger proof is not admitted from this row.

### Base64 and Pidigits

Range closure adds no direct body. Base64 remains in Go Base64/MD5 kernels.
Pidigits proves several individual generated arithmetic helpers safe, but the
enclosing functions still cross BigInt entry and host boundaries. Those
different dependencies prevent a value-only effect closure.

### K-Nucleotide

`sign` and `sign_i64` become universally range-free because their only checked
operation is literal `0 - 1`. Both are emitted stdlib helpers and are not
reachable from the compiled main. Hot encode/window/count paths retain
loop-carried u64 shifts, map operations, text conversion, and count overflow;
the analyzer deliberately does not freeze their first-iteration zero values.

### TapeLang Alphabet

The already control-free `Tape.get` remains eligible and inlined. `Tape.inc`
and `Tape.move` combine bounded deltas with mutable cell/position state that can
accumulate across arbitrary programs, so their additions remain capable of
overflow. Parse-time literal subtractions are safe but do not remove the
functions' runtime/error dependencies.

### Mandelbrot

`pixel_byte` is newly universally range-free. Its local loop proves
`bit ∈ [0,7]`, hence `7 - bit ∈ [0,7]`, and `1 << (7 - bit)` fits i32. This is
a genuine general compiler proof, independent of image dimensions or public
inputs.

It does not pass the performance gate: the preceding verified main profile is
only 120 ms, samples zero CPU in the post-call control check, and attributes
the visible body to direct float work. Building an ABI candidate around this
row would optimize an unmeasured cost.

## Verification

Focused analyzer tests cover:

- safe recursive control SCCs and fallible propagation;
- handled controls and constructed-control reassignment;
- universal versus internal call-site arithmetic safety;
- a universally bounded local-loop shift;
- full loop-call argument coverage; and
- rejection of a growing loop-carried shift value.

Repeated focused analyzer runs pass in under 0.01 seconds. All seven source
modules parse and produce deterministic JSON classifications. No performance
candidate or canonical stdlib source changed.

## Reproduction

```text
cd v12/interpreters/go
go test ./cmd/compiled-control-census -count=1
go run ./cmd/compiled-control-census -output /tmp/control-range-census.json \
  /path/to/generated/module
```

## Next recommendation

Close the compiled control-ABI branch for now and run the bytecode typed-register
architecture feasibility and target-budget audit already identified by the
cross-mode residual model.

Why: two increasingly broad compiler censuses find only one material eligible
application, while the bytecode VM repeatedly exposes stack/slot transport as
a three-family parent whose partial fast paths have reached diminishing
returns. Before writing another local VM candidate, the project needs to know
whether a complete typed register IR can remove enough dispatch, boxing, slot,
and allocation work to approach Python/Ruby performance without narrowing Able
semantics.

What it entails: measure semantic operations versus executed VM instructions,
dispatch/allocation budget, and target speedup across Fixed Width 128, Distance
Field, Concurrent Event Routing, and unlike guards; enumerate the full
language/stdlib closure required for typed registers; identify fallback and
deoptimization boundaries; estimate implementation stages and correctness
fixtures; and admit implementation only if the modeled end-to-end savings meet
the target across unlike families. Keep compiler work on independently
measured exact owners, retain the stack VM until parity is proven, and do not
begin WASM work.
