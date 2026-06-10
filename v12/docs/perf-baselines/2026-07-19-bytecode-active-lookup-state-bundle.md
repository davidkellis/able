# Bytecode active-lookup state bundle

Date: 2026-07-19

Decision: retain the generic VM state-layout change.

## Selection

The preceding operand-register IR census identified ordinary `CallName`
execution as the only shared blocker that cleared the cross-program admission
bar, but it did not identify a safe operation to optimize. Fresh bounded CPU
and sampled-allocation profiles were therefore collected for Fixed Width 128,
Distance Field, Mandelbrot, Option/Result Config, Word Frequency, and Reverse
Complement.

Each application ran in its own process with its catalog input and working
directory, the canonical external stdlib, `GOMAXPROCS=1`, `GOGC=50`,
`GOMEMLIMIT=1GiB`, and a 55-second cap. Every run passed its public verifier.
The pre-change binary was preserved as
`/tmp/able-call-boundary-profile-20260719` with SHA-256
`b406e483ee6f05bc7c9d3681ffe8da64de297c3b3a2bec7b63310d122cfe63e0`.

`switchRunProgramWithActiveLookupState` was the exact shared call-boundary
operation in four materially exposed programs:

| Application | Pre-change cumulative CPU |
| --- | ---: |
| Fixed Width 128 | 3.53% |
| Distance Field | 4.19% |
| Option/Result Config | 2.44% |
| Word Frequency | 4.00% |
| Mandelbrot | 0.16% |
| Reverse Complement | no sample |

The allocation profiles did not identify a shared call-frame allocation:
frames were already pooled. Their dominant owners instead diverged into wide
integers, Float/Ratio work, float boxing, interface/type work, parser/array
work, and array/integer work. The selected CPU operation remained generic and
cross-program; allocation-specific candidates did not.

## Change

The VM previously kept the active program and its six lookup-cache pointers or
slices as seven independent fields. Inline calls captured, installed, and
restored them field by field. The retained change stores those fields in the
existing `bytecodeActiveLookupProgramState` value and copies that value as the
unit of state at the program-switch boundary.

This changes representation only. It adds no opcode, language rule, nominal
type special case, benchmark recognition, or alternate scheduler/error path.
Global, lexical, call-name, and index lookup behavior is unchanged. No stdlib,
compiler, tree-walker, or WASM change was needed.

The candidate binary was preserved during measurement as
`/tmp/able-call-boundary-candidate-20260719` with SHA-256
`57fe72757410edc3bc8974e0a7d5ee0be12a1bca0764f27183f887fdd8990eb1`.

## Repeated verifier-backed A/B

Measurements used one allowed workstation CPU, `GOMAXPROCS=1`, `GOGC=50`,
and `GOMEMLIMIT=1GiB`. Baseline-first and candidate-first order alternated by
pair. All samples, including slow samples, are included. Every one of the 80
executions passed its public verifier. Fixed Width received five additional
pairs after its first set was mixed; Array Slice Window was added as a
low-duration independent guard.

| Application | Pairs | Baseline mean | Candidate mean | Change | Candidate wins |
| --- | ---: | ---: | ---: | ---: | ---: |
| Distance Field | 5 | 5.942 s | 5.722 s | -3.70% | 5/5 |
| Option/Result Config | 5 | 0.884 s | 0.846 s | -4.30% | 5/5 |
| Word Frequency | 5 | 1.590 s | 1.534 s | -3.52% | 4/5 |
| Fixed Width 128 | 10 | 8.134 s | 8.136 s | +0.02% | 6/10 |
| Array Slice Window | 5 | 0.730 s | 0.732 s | +0.27% | 3/5 |
| Reverse Complement | 5 | 4.490 s | 4.442 s | -1.07% | 2/5 |
| Mandelbrot | 5 | 6.700 s | 6.400 s | -4.48% | 3/5 |

The three unrelated workloads with material pre-change exposure improved by
3.52%-4.30% and the candidate won 14 of their 15 pairs. Fixed Width was neutral
after ten pairs and Array Slice was neutral. Mandelbrot's apparent improvement
is not treated as causal evidence because its pre-change exposure was only
0.16% and one baseline sample was slow. Reverse Complement is likewise a
low-exposure control.

Raw timing samples are preserved in
`2026-07-19-bytecode-active-lookup-state-ab-samples.tsv`.

## Causal profile check

Fresh post-change CPU profiles used the same bounded one-process protocol.
The selected operation fell in three of the four materially sampled programs:

| Application | Before | After |
| --- | ---: | ---: |
| Distance Field | 4.19% | 1.91% |
| Word Frequency | 4.00% | 2.72% |
| Fixed Width 128 | 3.53% | 2.39% |
| Option/Result Config | 2.44% | 2.53% |

Option/Result's post profile contained only 790 ms of CPU samples, so its
0.09-point movement is below useful sampling resolution. The timing result is
supporting evidence there, while the other three profiles establish the
intended mechanism.

## Correctness

The focused call-name, lookup/index-cache, program-switch, inline-return, and
pool suite passed in 0.587 seconds. `go test ./pkg/interpreter -run
'TestBytecode' -count=1 -timeout 60s` passed in 26.942 seconds. Split bytecode
parity groups 02-04, 05-08, 09-11, 12-13, and 14 all passed in 11.977, 28.493,
7.843, 12.607, and 20.942 seconds respectively. Every command remained below
the one-minute test ceiling.

## Next bounded target

Do not retry the previously rejected return-guard reorder or empty-frame
special cases. After this change, `finishInlineReturn` remains broad and its
concrete `popCallFrameFields` child accounts for approximately 4.17% Distance
Field, 2.72% Word Frequency, 2.53% Option/Result, and 1.79% Fixed Width CPU.

The next tranche should instrument which full-frame restore components are
actually live at this boundary across those four applications: integer
registers, implicit slots, value-slot integer/float sidecars, environment and
scope state, control bases, ownership, and active lookup state. A candidate is
admissible only if the same removable copy or generically packable restore unit
repeats in at least three unlike programs. This narrows the remaining shared
return wall without selecting a benchmark, nominal type, or speculative branch
order.
