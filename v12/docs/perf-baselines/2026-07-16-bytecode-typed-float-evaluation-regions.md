# Bytecode Typed-Float Evaluation Regions

Date: 2026-07-16

## Outcome

Retain the first generic typed-float evaluation region. Statically proven
multi-operation `f32`/`f64` expressions over local value slots and float
literals now execute from one compact postfix plan with a fixed stack-local
scratch area. Intermediate primitive results remain raw; the VM materializes
one ordinary slot value at the region exit.

The candidate clears the broad gate. Repeated fresh-process means improve
Distance Field by 5.51% and RMS Norm by 21.63%, while reducing their allocation
counts by 15.79% and 50.00%. Reduced NBody is neutral at -1.14%, and a volatile
ten-process matrix cohort is neutral at +1.94% with identical allocations.
Non-float controls are mixed inside workstation noise and have unchanged
allocation shapes.

No eligibility depends on a benchmark, source name, function name, stdlib
type, or non-primitive nominal type. This is a primitive `f32`/`f64` bytecode
lowering rule. No `able-stdlib` change is needed.

## Design

The lowering pass accepts exact-float expression trees containing local value
slots, float literals, and `+`, `-`, `*`, or `/`. It requires at least two
arithmetic operations, allowing the existing single-operation fused opcodes to
retain priority. The plan is postfix, has a verified result depth of one, and
is capped at a scratch depth of 16.

The VM evaluates a valid plan with fixed `[16]float64` and float-kind arrays.
`f32` results normalize after every operation. If a runtime slot violates its
static proof, the same plan takes a cold generic-value fallback instead of
assuming an invalid representation. The region exit stores an immutable raw
float carrier through the ordinary slot boundary; this is important because
reusing a mutable pointer there made later stack snapshots observe subsequent
slot updates and regressed allocation behavior.

Exact unary-negated float expressions now preserve their simple `f32`/`f64`
fact. This is a general type-fact correction: it lets slots initialized with a
negative float participate in later exact-float analysis, but the unary
expression itself is not a special region opcode.

The experiment began from an exact pre-candidate test binary:

```text
/tmp/able-bytecode-float-region-baseline.test
sha256 66ca53548084162386fb6dbc50d89b5348992db864ddaad4b36d6ad519bfc212
size   37,350,552 bytes
```

## Broad benchmark gate

Times are arithmetic means of separate workstation processes. Three samples
were used except for matrix, whose short and volatile process time received
ten samples. Positive wall change is slower.

| workload | baseline mean | candidate mean | wall change | allocation change |
| --- | ---: | ---: | ---: | ---: |
| Distance Field | 6.121 s | 5.784 s | -5.51% | about 512,052,149 B / 38,000,142 to 464,051,500 B / 32,000,136; -9.38% bytes / -15.79% allocs |
| RMS Norm | 6.041 s | 4.734 s | -21.63% | about 592,051,565 B / 52,000,162 to 384,049,557 B / 26,000,140; -35.13% bytes / -50.00% allocs |
| reduced NBody | 1.667 s | 1.648 s | -1.14% | effectively identical; 2 fewer allocs and about 50 fewer bytes |
| `matrixmultiply_f64_small` | 0.3265 s | 0.3328 s | +1.94% | identical: 47,636,112 B / 1,187,689 allocs |
| string split/join control | 0.9846 s | 0.9444 s | -4.08% | effectively identical |
| iterator collect control | 0.4067 s | 0.3873 s | -4.76% | identical apart from one fresh-process allocation |
| numeric array/map control | 0.0633 s | 0.0648 s | +2.39% | identical: 869,448 B / 239 allocs |

Distance baseline samples were 6.493, 5.948, and 5.922 seconds; candidate
samples were 5.881, 5.637, and 5.834 seconds. RMS baseline samples were 6.614,
5.754, and 5.755 seconds; candidate samples were 5.036, 4.651, and 4.515
seconds. Those two primary wins are larger than their observed candidate
variation and are corroborated by stable allocation-count reductions.

The matrix means have 6.13% baseline and 8.33% candidate coefficients of
variation, while the 1.94% mean movement has no allocation movement. It is
therefore a neutral guard rather than evidence of a regression. The short
array/map control has the same interpretation: its 2.39% movement is below
both cohorts' approximately 7.4% variation, with exactly unchanged
allocations.

## Verification

Focused lowering, parity, mixed-kind, `f32` normalization, cold fallback,
loop-carried target, and single-operation-priority guards pass. The complete
bytecode VM family also passes under the one-minute limit.

```text
go test ./pkg/interpreter -run 'TestBytecodeVM_(LoweringEmitsTypedFloatRegion|TypedFloatRegion|LoweringEmitsFloat|Float.*Parity)|TestBytecodeFloatSimpleTypeCheck' -count=1 -timeout 60s
ok  able/interpreter-go/pkg/interpreter  0.067s

go test ./pkg/interpreter -run '^TestBytecodeVM' -count=1 -timeout 60s
ok  able/interpreter-go/pkg/interpreter  18.185s

```

The unfiltered package command was attempted, but
`TestExecFixtureParity/14_19_regex_scanner_boundaries` reached the 60-second
timeout while parsing its module. A second package run with
`TestExecFixtureParity` excluded also exhausted the package deadline under
workstation load. Neither run reported a typed-region failure before the hard
timeout. These package-wide aggregates are currently unsuitable as the
project's under-one-minute gate; all benchmark programs used for the gate
completed successfully through the candidate VM.

The line-cap refactor moved the existing call-name slot-argument helper into
`bytecode_lowering_call_name.go`; `bytecode_lowering.go` is now 968 lines.

## Next recommendation

Refresh post-region CPU and allocation profiles, then extend the region model
to generic statically typed external leaves only if the same boundary cost
repeats across numeric workloads.

Why: Distance and RMS benefit because their arithmetic interiors are composed
mostly of proven local float slots. Reduced NBody and matrix do not reduce
allocations because important values arrive through calls or aggregate reads,
which deliberately terminate this first region. That boundary is now the
largest plausible shared opportunity, but it must be confirmed after removing
the local-expression cost rather than inferred from old profiles.

What it entails: capture bounded one-process CPU and allocation profiles for
Distance, RMS, reduced NBody, and matrix using the retained binary; quantify
region coverage and the remaining cost of calls, typed array reads, boxing,
and materialization; then, only if multiple profiles agree, add a generic
region input form that evaluates a statically exact `f32`/`f64` subexpression
once in language order and feeds its raw result into the plan. Guard error and
side-effect ordering, mixed kinds, aliases, joins, recursion, and dynamic
fallback. Re-run the same four numeric workloads and three non-float controls
with repeated process means. Continue to defer WASM.
