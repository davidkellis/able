# Bytecode typed loop-block gate (2026-07-18)

## Decision

Keep no compiler, VM, runtime, stdlib, benchmark, or fixture change. A broad
class of loop-contained typed scalar basic blocks passed the predeclared
coverage gate in three unlike applications, so it earned a prototype. The
prototype then made all three admitted workloads neutral or slower. The
second, out-of-line opcode dispatcher cost more than the four-instruction
batching boundary saved, so it was removed.

The original bytecode instructions remained present throughout the trial, and
the candidate was disabled whenever bytecode statistics were enabled. After
rejection, all block annotations, planning, execution code, and tests were
removed.

## Admission rule

The temporary observer found maximal straight-line regions inside loops
identified by backward jumps and basic-block leaders. It admitted regions of
at least four instructions containing only already-lowered primitive scalar
constants, arithmetic, casts, local-slot stores, stack discard, and an
unconditional jump. Calls, allocation, member or operator dispatch, dynamic
lookup, environment mutation, raises, yields, identity-bearing operations, and
unproven constants were excluded.

Before measurement, a class was declared material only when it:

- executed at least 10,000 times;
- covered at least 5% of main's bytecode instructions; and
- repeated in three unlike applications spanning multiple workload families.

Existing production recurrence and float-loop kernels remained enabled. The
observer counted only instructions reached on the production execution path.

## Coverage census

The same bounded protocol used for the primitive-integer census covered the
exact 35-application external selection with `GOMAXPROCS=1`, a 55-second
process limit, canonical external stdlib sources, main-only counters, and each
application's public verifier. Twenty-nine full applications completed and
verified.

The dynamically material rows were:

| Application | Main instructions | Block executions | Covered instructions | Share | Class |
| --- | ---: | ---: | ---: | ---: | --- |
| Monte Carlo Pi | 106,341,149 | 11,111,000 | 44,444,000 | 41.7938% | mixed integer/float, length 4 |
| RMS Norm | 68,000,226 | 2,000,000 | 8,000,000 | 11.7647% | mixed integer/float, length 4 |
| Fixed Width 128 | 60,446,436 | 1,000,000 | 4,000,000 | 6.6174% | integer, length 4 |
| Future Await Race | 229,438 | 18,432 | 110,592 | 48.2013% | integer, length 6 |

PiDigits and Word Frequency reached admitted code but stayed below the
materiality threshold. The remaining completed full applications had no
material admitted block.

Binary Trees, QuickSort, Sudoku Masks, N-Body, TapeLang Alphabet, and Regex
Suffix exceeded the full-size bound. Existing smaller sources closed five of
those exclusions: Binary Trees, N-Body, TapeLang, and the exact reduced Sudoku
source had zero qualifying blocks, while a profiled 128-word Regex Suffix run
also had zero and independently verified as `512:56:8952`.

QuickSort used the unchanged external source with the first 4,000 canonical
input numbers. Its output was independently checked by sorting the same input
with Ruby. It executed 1,243,321 main instructions; 35,580 mixed length-four
blocks covered 142,320 instructions, or 11.4468%. This made mixed typed scalar
blocks material in three unlike families: deterministic random/numeric work,
floating-point vector work, and file parsing/sorting.

## Shape attribution

Exact shape attribution confirmed that the breadth did not come from one
benchmark-shaped sequence:

- Monte Carlo Pi alternated a checked integer multiply/modulo store with a
  slot-to-float cast/divide store.
- RMS Norm used discard, float binary store, integer slot/constant store, and
  jump.
- QuickSort used constant, integer subtraction, integer multiply/add store,
  and constant.

The shapes had no hash collisions. Because the exact sequences differed, the
prototype was one shared small executor over the bounded safe opcode set, not
a sequence-specific superinstruction.

## Prototype and causal gate

The program metadata pass annotated only the first instruction of each safe
block. The normal VM dispatcher called a small out-of-line executor, which
used the existing opcode helpers and left every original instruction intact as
a cold fallback. Statistics mode continued through the ordinary dispatcher so
its instruction accounting remained exact.

Both baseline and candidate CLIs were preserved before the decision. Every
timed output passed the applicable Ruby verifier; the 50,000-number QuickSort
timing input was independently sorted and checked. Runs alternated baseline
and candidate on logical CPU 0 with `GOMAXPROCS=1`. Monte Carlo Pi and
QuickSort were extended to ten pairs because their first five pairs were
volatile.

| Workload | Pairs | Baseline mean | Candidate mean | Mean delta | Median paired delta |
| --- | ---: | ---: | ---: | ---: | ---: |
| Monte Carlo Pi | 10 | 2.749152 s | 2.869319 s | +4.37% | +4.65% |
| RMS Norm | 5 | 4.909369 s | 4.934215 s | +0.51% | +2.17% |
| QuickSort, 50,000 canonical numbers | 10 | 1.164481 s | 1.213043 s | +4.17% | +1.76% |

All deltas are regressions; positive means slower. The candidate therefore
failed before guard expansion was necessary. Its binary was only 12,560 bytes
larger, so size was not the limiting factor. The failure is execution cost:
entering another function, switching over the same opcodes again, and checking
the block boundary did not amortize over four instructions.

## Restoration and verification

The restored binary is 45,813,432 bytes, exactly the baseline size. Its two Go
content IDs match the preserved baseline; only link action IDs differ because
the binaries were written under different output names.

Final verification after removing all temporary code:

```text
go test ./pkg/interpreter -run '^TestBytecode' -count=1 -timeout 55s
ok   able/interpreter-go/pkg/interpreter  25.135s
```

## Next recommendation

Profile the three admitted families inside their existing typed opcode
handlers, with no secondary dispatcher. Attribute raw slot extraction,
sidecar synchronization, boxed stack conversion, and helper-call cost in Monte
Carlo Pi, RMS Norm, and QuickSort, and advance only when the same primitive
wall repeats materially in all three profiles.

The census proves that typed scalar work is broad and hot; the rejected
prototype proves that batching four instructions through a second dispatch
layer is the wrong mechanism. The next tranche should take one bounded CPU
profile per admitted application, compare exact handler-level costs, and build
at most one in-dispatch candidate that removes a shared representation or
helper boundary. It must then use alternating preserved binaries and the text,
wide-integer, concurrency, float-control, and allocation-heavy guards. This
keeps the search focused on a demonstrated cross-family wall without turning
the three observed opcode sequences into special cases. Continue to defer
WASM.
