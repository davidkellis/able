# Bytecode register-admission closure census

Date: 2026-07-21

## Decision

The current minimum semantic closure is one opcode family. Proceed next with
a cold-fallback operand-register `MemberAccess` vertical slice, but do not
retain any census or executable prototype from this tranche and do not make a
performance claim yet.

`MemberAccess` is the only sole missing family in materially reached functions
across six unlike applications and 253,448 dynamic function entries. Every one
of those functions also passed the existing operand-transfer and CFG-merge
contracts when `MemberAccess` was temporarily admitted. This clears the
predeclared three-application selection bar before implementation.

No compiler, VM, runtime, stdlib, workload, fixture, reference implementation,
scorecard, language, or WASM performance change is retained.

## Method

The recovered register translator was extended only for diagnosis:

- a descriptor inventory classified all 143 current bytecode opcodes;
- each rejected function reported its complete sorted set of unsupported
  semantic families instead of only its first blocker;
- transfer underflow, invalid/stale slot references, and incompatible CFG
  merge identities/depths/slot versions received distinct rejection reasons;
- admission was weighted at every real dynamic function entry while the
  ordinary bytecode VM remained the execution authority;
- `MemberAccess`, `StructLiteralNamedFast`, and `CallMember` were then admitted
  one family at a time to expose any transfer/merge wall behind the semantic
  gate.

Materiality was fixed before selection at the greater of 1,000 dynamic entries
or 1% of an application's attempts. This excludes startup-only functions while
allowing shorter applications to contribute real repeated work. Counts are
deterministic events, so repeated timing and arithmetic means do not apply to
this diagnostic tranche. Every second-stage run reproduced its prior
verifier-backed output byte-for-byte under a 55-second, one-process,
`GOMAXPROCS=1`, 1-GiB guard.

## Baseline census

The 14 applications produced 9,312,126 dynamic admission attempts. The
bounded prototype already admitted 2,024,588 entries, almost entirely the
previously established Distance Field and Option/Result call-compatible
functions. Of the remaining entries, 7,268,082 stopped at an unsupported
semantic closure. Two already-supported shapes exposed precise later walls:
16,384 `merge_slot_versions` failures in Validated Job Pipeline and 3,072
`merge_value_identity` failures in Concurrent Event Routing.

| Application | Attempts | Already admitted | Later state/CFG failure |
| --- | ---: | ---: | --- |
| Future Pipeline | 40,993 | 0 | none behind semantic gate |
| Fixed Width 128 | 3,000,016 | 1 | none behind semantic gate |
| Distance Field | 4,000,009 | 2,000,000 | none behind semantic gate |
| Mandelbrot | 80,018 | 4 | none behind semantic gate |
| Array Slice Window | 36,014 | 0 | none behind semantic gate |
| Option/Result Config | 293,789 | 24,576 | none behind semantic gate |
| Word Frequency | 673,776 | 1 | none behind semantic gate |
| Reverse Complement | 62 | 3 | none behind semantic gate |
| Concurrent Text Index | 117,493 | 1 | none behind semantic gate |
| Validated Job Pipeline | 182,030 | 0 | 16,384 merge-slot-version entries |
| Dependency Wave Validation | 120,703 | 0 | none behind semantic gate |
| Concurrent Event Routing | 763,196 | 1 | 3,072 merge-value-identity entries |
| Matrix Multiply | 4,014 | 0 | none behind semantic gate |
| Monte Carlo Pi | 13 | 1 | none behind semantic gate |

## Minimum shared closures

Three sole-family closures clear the material three-application bar. Each
second-stage admitted delta exactly equaled the original sole-family count;
none turned into a transfer or merge rejection.

| Sole added family | Material applications | Dynamic entries unlocked | State/CFG result |
| --- | ---: | ---: | --- |
| `MemberAccess` | 6 | 253,448 | all admitted |
| `StructLiteralNamedFast` | 5 | 1,068,347 | all admitted |
| `CallMember` | 3 | 66,325 | all admitted |

`MemberAccess` breadth consists of Concurrent Text Index (66,925), Concurrent
Event Routing (22,545), Dependency Wave Validation (29,688), Validated Job
Pipeline (65,549), Word Frequency (35,958), and Future Pipeline (32,783).

`StructLiteralNamedFast` has the largest raw total, but 1,000,004 of its
1,068,347 entries come from Fixed Width 128. The rest span Matrix Multiply,
Concurrent Event Routing, Word Frequency, and Array Slice Window. This remains
a generic struct-literal opcode rather than a named-type special case, but its
selection evidence is more concentrated.

`CallMember` spans Concurrent Text Index, Dependency Wave Validation, and Word
Frequency. It meets the minimum breadth exactly, but it adds member lookup,
arbitrary dispatch, inline-frame/suspension, and return reconciliation to the
already specified call continuation ABI.

## Interpretation

The previous first-blocker map was stale, but the larger architecture remains
viable. Complete closure accounting changes the next step from another guessed
opcode to a one-family slice with measured breadth. The second-stage runs also
show that semantic admission is the actual wall for the selected functions;
their operand identity and CFG state already fit the current translator.

`MemberAccess` is preferred over the higher-volume struct-literal candidate
because it reaches the most unlike applications and its one-input/one-output
semantics make it the smallest reliable executor extension. It is preferred
over `CallMember` because ordinary member value/field resolution does not
require adding another call/suspension boundary in the same tranche.

## Next recommendation

Implement the register-native `MemberAccess` vertical slice and rerun dynamic
reach before timing.

Why: it is the widest minimum closure, it clears the selection bar in six
unlike applications, and the exact candidate functions have already passed
transfer and CFG validation. This is broad language behavior, not a benchmark,
stdlib-container, or nominal-type special case.

What it entails: recover the documented operand-register engine and existing
call/continuation ABI; add `MemberAccess` transfer, operand emission, execution,
error/diagnostic parity, and cold fallback before effects; add focused safe and
ordinary field/method-value guards; then rerun admission and actual semantic-IR
reach across the six applications plus unlike controls. Only if at least three
applications execute material register work should the tranche run repeated,
order-balanced verifier-backed A/B processes and compare arithmetic means.
Retain the slice only for a broad win without guarded regressions. Continue to
defer WASM.
