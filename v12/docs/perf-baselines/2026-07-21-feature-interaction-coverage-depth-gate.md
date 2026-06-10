# Feature-interaction coverage-depth gate

Date: 2026-07-21

## Decision

Keep the portable Policy Record Dispatch application, its Go/Python/Ruby
implementations, checked-in input, independent verifier, catalog contract, and
feature-coverage annotations. Promote it in compiled and bytecode modes.

Keep no compiler, generated-runtime, bytecode-VM, canonical-stdlib, or language
performance change. Both Able profiles reproduce exact owners already closed
by unlike applications. A policy-, Regex-, Result-, pattern-, map-, or nominal-
specific fast path would be benchmark specialization rather than a generally
applicable optimization.

## Coverage-depth selection

The pre-tranche matrix had no empty pair but nine depth-one pairs. Eight of
those depended only on Concurrent Event Routing and joined lexical bindings or
patterns to text/files, closures, program entry, inherent methods, interface
dispatch, Option/Result, stdlib/Regex, or nominal/union types.

Policy Record Dispatch supplies a different serial workload. It reads 32
records from a real entry argument, validates complete records with Regex
captures, returns `Result PolicyRecord`, applies a capturing scoring callback,
destructures nominal records, matches accepted/rejected union states in
inherent methods, and aggregates through the public `Map` interface. Default
execution repeats the bounded corpus 64 times and emits:

```text
2048:1536:512:384,384,384,384:769419123:611618
```

The Able, Go, Python, and Ruby programs use the same records, validation,
scoring, aggregation, and output contract. The canonical and sibling Able
sources are byte-identical with SHA-256
`cd6089f205ff542c9c9c07c5526d47b9fc4287f9fd9c7081947d363610ec5263`.

Recomputing all 55 interactions across the 11 discriminating portable/mixed
families gives:

| Measure | Before | Current |
| --- | ---: | ---: |
| zero-depth pairs | 0 | 0 |
| depth-one pairs | 9 | 1 |
| pairs strengthened by the application | — | 45 |

All eight targeted depth-one pairs now have depth two. The remaining
depth-one pair is `concurrency × lexical_blocks_bindings_patterns`; it still
depends only on Concurrent Event Routing. Evidence:
`2026-07-21-feature-interaction-coverage-depth-matrix.{json,md}`.

## Repeated admission evidence

Every number is an arithmetic mean of five fresh verifier-backed processes.
The separated promotion reports name only the applicable comparison family,
so the strict evidence checker can prove five complete reference samples for
each mode.

| Mode/reference | Able mean | Reference mean | Ratio |
| --- | ---: | ---: | ---: |
| compiled / Go | 0.2260 s | 0.0056 s | 40.36x |
| bytecode / Python | 8.7820 s | 0.0245 s | 358.45x |
| bytecode / Ruby | 8.7820 s | 0.0519 s | 169.21x |

The initial independent cohort measured 0.2340 seconds compiled and 7.7500
seconds bytecode. The refreshed compiled promotion cohort measured 0.2260
seconds, producing a 0.2300-second mean across ten Able launches. Bytecode was
more volatile: the promotion cohort measured 8.7820 seconds and a third
five-process cohort measured 7.8780 seconds. The required pooled 15-run Able
mean is 8.1367 seconds, or 332.11x Python and 156.78x Ruby against the fresh
native-substring references. Every process retained the same verified output
hash; no single volatile cohort selected an implementation decision.

The mode-aware selection now contains 85 rows: 46 compiled and 39 bytecode.
The aggregate contains 92 full-status rows and exactly five successful Able
and applicable reference samples for every selected row.

## Exact attribution

One preserved compiled binary with SHA-256
`fb29cbeb6163156e5df48c255642952660ed82b878de34d64fed4a273eab3929`
produced a main-phase CPU profile with 190 ms of samples. The exact application
path is canonical Regex work: `Regex.match` is 36.84% cumulative,
`find_nfa_span` and `find_token_match` are 15.79% cumulative, and concrete NFA
move/accept leaves appear alongside Go allocation/GC. This is the established
`compiled-regex` owner; its closure, capture, carrier, index, arena, and call
alternatives already completed broad gates.

One loaded, lowered, warmed bytecode artifact measured one full default
`main()` at 6,639,195,209 ns/op, 139,533,024 B/op, and 1,453,755 allocs/op.
Its 6.86 seconds of CPU samples report `execCallMemberArraySlot` at 21.72%
cumulative, cached validated Array-slot lookup at 7.14%, `finishInlineReturn`
at 8.02%, raw integer slot storage at 3.06%, and generic call dispatch below
the VM loop. Those exact member/cache, return/frame, and raw-integer families
are current and closed. The profile exposes no new non-nominal leaf in three
unlike applications, so no candidate passes admission.

## Scope and verification

- No benchmark-specific opcode, lowering, nominal/container rule, stdlib
  shortcut, runtime patch, or WASM work was introduced.
- Five Go, five Python, and five Ruby reference processes verified.
- Ten compiled and fifteen bytecode Able comparison processes verified across
  the independent, promotion, and volatility cohorts.
- The feature-interaction generator contract, JSON/catalog syntax, selection
  evidence, current scoreboard replay, frontier tests, and the focused Able
  execution checks pass.

## Next recommendation

Audit and close the sole remaining depth-one interaction,
`concurrency × lexical_blocks_bindings_patterns`, before selecting another
implementation experiment.

Why: every current performance group is exact and closed, while that one pair
still relies on a single event-routing program. It is the last place where one
application shape can masquerade as broad feature-interaction evidence.

What it entails: first inspect the existing concurrency applications for a
substantial, already-executed lexical/destructuring-pattern path and annotate
one only if the coverage is real and verifier-backed. If none qualifies, add
one bounded non-Regex concurrent application with source-equivalent Able, Go,
Python, and Ruby implementations plus an independent verifier. Profile both
Able modes, but admit performance code only if a new concrete generic leaf
repeats in at least three unlike applications and passes the established
compiled/bytecode guards. Continue to defer WASM.
