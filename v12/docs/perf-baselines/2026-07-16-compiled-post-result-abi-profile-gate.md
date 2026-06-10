# Compiled post-result-ABI profile gate

Date: 2026-07-16

## Decision

Keep no compiler, runtime, stdlib, or benchmark optimization from this gate.
Fresh current measurements still show large compiled-product gaps, but Binary
Trees, Sudoku Masks, K-Nucleotide, and TapeLang Alphabet do not share one
concrete material compiler-owned descendant across three unlike programs.

The gate therefore rejects a speculative common allocation, boxing, map, or
method shortcut. In particular, it adds no `Node`, Sudoku, HashMap, TapeLang,
nominal-type, or application-specific lowering rule.

## Measurement setup

All four Able sources were rebuilt with `--no-fallbacks` against the canonical
external stdlib. Each normal timing row is the arithmetic mean of five
independent processes with:

- `GOMAXPROCS=1`
- `GOGC=50`
- `GOMEMLIMIT=1GiB`
- a 55-second process timeout
- the catalog working directory, arguments, and public Ruby verifier

Every one of the 20 timed outputs passed its verifier. Each application
produced one stable SHA-256 across all five runs.

| Application | Mean | Median | Range | CV | Previous scorecard | Change |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Binary Trees | 28.324 s | 27.411 s | 27.225-31.922 s | 6.37% | 29.106 s | -2.69% |
| Sudoku Masks | 9.145 s | 8.978 s | 8.879-9.514 s | 2.94% | 8.276 s | +10.50% |
| K-Nucleotide | 3.450 s | 3.456 s | 3.412-3.477 s | 0.71% | 3.418 s | +0.94% |
| TapeLang Alphabet | 3.710 s | 3.540 s | 3.431-4.390 s | 9.55% | 3.420 s | +8.47% |

The previous scorecard is not a paired A/B control, so these changes describe
the refreshed workstation samples rather than attributing them to a particular
compiler change. Against the scorecard's existing Go reference means, the
current Able/Go ratios are 5.271x, 16.932x, 62.615x, and 2.015x respectively;
all remain outside the 95%-of-Go goal.

## Current CPU owners

Separate CPU-only profile processes used the same runtime settings and passed
the same verifiers. Their generated-main profiles divide immediately below
generic Go runtime parents:

| Application | Current material owner | Profile evidence |
| --- | --- | --- |
| Binary Trees | recursive identity-bearing `Node` construction | `make_tree` is 80.19% cumulative; `mallocgc` is 73.71% cumulative and GC scanning dominates flat time |
| Sudoku Masks | recursive best-cell candidate arrays and growth | `find_best_empty` is 88.63% cumulative; `growslice` is 38.09% and `mallocgc` 52.36% cumulative |
| K-Nucleotide | string conversion, primitive map hashing/equality, and integer boundaries | `ToInt`/`ToUint` plus generated string/map helpers dominate; no tree, Sudoku-buffer, or tape descendant repeats |
| TapeLang Alphabet | allocation-free generated loop/method work | `execute` owns 100% cumulative; `Tape.inc`, `Tape.get`, and `Tape.move` account for the remaining flat samples, with no allocation/GC sample |

Runtime allocation and GC are common parents in only the first three, and
their compiler/source descendants are different semantic operations. TapeLang
is an important negative control: an allocation shortcut cannot address its
current wall.

## Bounded allocation-profile result

Exact phase allocation mode sets `runtime.MemProfileRate=1`. Under the project
limit, Binary Trees, Sudoku, and K-Nucleotide did not finish within 55-60
seconds, even though their ordinary and CPU-profile processes did. They wrote
valid bootstrap and main-start snapshots but no main-end snapshot, so those
partial files were not used to rank candidates.

TapeLang completed and reported 282,624 main bytes / 4,278 allocations. Of
those, 270 KiB / 4,225 allocations are the once-only lazy common-i32 box table
on its first runtime boundary. Moving that table between phases would not
remove complete-process work, does not appear in TapeLang's material CPU
samples, and does not explain the other three applications' distinct hot
loops. The start/end profile subtraction is dominated by the profiler's own
snapshot serialization, so the phase counters, not that subtraction, are the
authoritative TapeLang totals.

The inability to obtain full exact profiles for the allocation-heavy programs
inside one minute is recorded as a profiling bound, not worked around with a
longer test or an unverified partial result.

## Generality gate

No exact application-loop descendant is both material and shared by at least
three unlike programs:

- recursive nominal allocation is specific to the Binary Trees lifetime;
- Sudoku's growing candidate Array is a different source/data-structure wall;
- K-Nucleotide alone crosses the primitive HashMap/string/boxed-integer lane;
- TapeLang is dominated by generated primitive array/method execution.

Generic Go runtime symbols such as `mallocgc` are consequences, not enough
evidence to select one legal Able lowering change. The previously closed raw
integer, bridge conversion, result-slot reuse, source-container, and lazy-cache
families are not reopened from these parent samples.

## Verification and cleanup

- 20/20 normal timing processes passed their public verifiers.
- Four CPU-profile processes passed their public verifiers.
- TapeLang's exact allocation process passed its public verifier.
- Focused caller-owned nominal-result tests and compiler builds pass after the
  gate with no production candidate present.
- Generated trees, binaries, outputs, and profiles are temporary and removed
  after this record.

## Next recommendation

Run a compiled allocation-light loop/lowering gate across TapeLang Alphabet,
Matrix Multiply, and NBody, with one structurally different primitive-array or
numeric application as a guard.

Why: this gate rules out one shared allocation solution for the large misses,
while TapeLang demonstrates that generated primitive loop and direct-method
code can still be about 2x Go without allocating. Matrix Multiply and NBody
exercise different numeric and array shapes and already have matched Go
references. Requiring a repeated generated operation across them is a better
test of a broadly useful lowering improvement than optimizing one tree, map,
or tape program.

What it entails: rebuild strict current binaries, collect five verifier-backed
processes per application, capture bounded CPU line profiles and Go inlining/
escape diagnostics, and reconcile exact generated operations such as checked
primitive arithmetic, static method calls, and primitive array get/store.
Trial a compiler change only if the same concrete language-level lowering is
material in at least three unlike programs; guard code size, startup, checked
semantics, and the allocation-heavy Binary Trees workload. Keep bytecode work
queued and continue to defer WASM.
