# Bytecode main-only stats inventory — 2026-07-15

## Scope

This is an opt-in instruction/counter inventory, not a timing measurement.
Each normal bytecode CLI process enabled `ABLE_BYTECODE_STATS=1`, wrote one
snapshot through `ABLE_BYTECODE_STATS_OUT`, and used
`ABLE_BYTECODE_STATS_MAIN_ONLY=1` to exclude load/typecheck/bootstrap work
before invoking `main()`. The three applications completed with their existing
Ruby verifiers. Counters do not establish CPU cost and are not scorecard input.

## Results

| Application | Main-only shape | Guarded fast-path evidence | Why it is not a shared candidate |
| --- | --- | --- | --- |
| Word Frequency | `LoadSlot` 4,406,730; `Pop` 2,465,647; `StoreSlotNew` 1,859,411; `CallName` 796,567; typed-pattern jump 778,084 | 424,785 direct-slot named calls; 117,772 member-cache hits / 30 misses; 140,752 Array-member fast hits / 9 fallbacks | Map/text/type-pattern and named-call work. |
| Future Pipeline | `LoadSlot` 2,203,692; `BinaryIntAdd` 1,589,255; `Binary` 540,681; slot integer-store 540,672 | 8,192 direct-slot named calls; 32,766 member-cache hits / 17 misses | Integer loop and inline/member-dispatch work. |
| Base64 | `LoadSlot` 2,000,087; `Jump` 1,000,063; slot integer-store 1,000,040; Array member call 1,000,000 | 999,999 Array-member fast hits / 1 fallback; only 25 named lookups | Canonical Array/kernel work. |

All three runs use the established guarded paths; no run recorded a generic
named-call fallback. Their only common high-count instructions are the VM
envelope (`LoadSlot`, `Jump`, and, where source results need discarding,
`Pop`). Those counters have no shared concrete descendant and cannot justify
a dispatcher, stack, call-frame, Array, text, iterator, or benchmark-specific
optimization.

## Reproduction

Run an ordinary bytecode CLI process with:

```sh
ABLE_BYTECODE_STATS=1 \
ABLE_BYTECODE_STATS_MAIN_ONLY=1 \
ABLE_BYTECODE_STATS_OUT=/tmp/able-bytecode-stats.json \
able --exec-mode=bytecode run path/to/main.able
```

The JSON mirrors `interpreter.BytecodeStatsSnapshot`; its current field names
are Go-exported names such as `TopOps`, `CallNameLookups`, and
`ArrayMemberSlotFastHits`. Use it for count attribution alongside CPU/allocation
profiles, never as a proxy for elapsed time.

## Decision

Keep no VM, compiler, runtime, canonical-stdlib, fixture, or benchmark change.
This inventory reinforces the existing candidate-admission rule: a future
change requires a concrete non-nominal material leaf shared by three unlike
verifier-backed applications and then a bounded broad performance gate.
