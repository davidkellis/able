# Bytecode coverage-extension scorecard gate — 2026-07-13

## Scope

This gate brings the seven application-shaped benchmarks outside the stable
16-program `generality` scorecard into the independent bytecode-versus-Python/
Ruby selection lane:

- Fixed Width 128 and Rational Series: distinct fixed-width and rational
  numeric/nominal arithmetic;
- Word Frequency, Document Audit, and Lexical Rollup: map/text, public
  iterator, and filesystem/typed-pattern pipelines; and
- Channel Rollup and Future Pipeline: independent text-channel and numeric
  Future/goroutine applications.

Fresh Python 3.14.5 and Ruby 4.0.5 reference processes and bytecode Able
processes ran three times each, CPU-15 pinned, under the canonical stdlib,
`GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and a 45-second per-process
guard. The two goroutine applications retained their catalogued executor.
Every completed output passed its canonical Ruby verifier.

The generated evidence is retained alongside this decision:

- `2026-07-13-bytecode-coverage-extension-interpreter-refresh.{json,md}`;
- `2026-07-13-bytecode-coverage-extension-scorecard.{json,md}`.

## Fresh status

| Application | Able bytecode | Able/Python | Able/Ruby | Status |
| --- | ---: | ---: | ---: | --- |
| Fixed Width 128 | 10.1600 s | 27.50x | 14.17x | verified 3/3 |
| Rational Series | 4.9567 s | 46.37x | 35.08x | verified 3/3 |
| Word Frequency | 1.7500 s | 89.74x | 29.51x | verified 3/3 |
| Document Audit | 0.3667 s | 25.12x | 8.75x | verified 3/3 |
| Lexical Rollup | 0.5000 s | 26.88x | 9.54x | verified 3/3 |
| Channel Rollup | 0.5633 s | 13.64x | 9.93x | verified 3/3 |
| Future Pipeline | 0.5267 s | 8.51x | 7.39x | verified 3/3 |

These are target misses, not direct optimization instructions. A process ratio
also includes loading, lowering, canonical-stdlib initialization, and host
boundaries. It can select a profile family, but it cannot select a named
collection, a numeric width, an input corpus, an iterator expression, or a
task topology.

## Candidate selection

The result does not establish a new broadly applicable bytecode operation.
Current-source profile evidence already separates the apparent groups:

- Fixed Width 128 uses the public checked two-word `UInt128` path, while
  Rational Series uses ratio/nominal arithmetic. Their prior paired profiles
  diverge; a shared value-representation change would be a width- or
  ratio-specific shortcut.
- Word Frequency's material descendants are map/raw-integer work. Document
  Audit and Lexical Rollup instead spend their attributable work in generator,
  member, iterator, public-return, and typed-pattern paths. Their shared call
  and inline-return frames are existing generic parents/candidates that have
  already failed broad guards.
- Channel Rollup's material child is text/call work, while Future Pipeline's
  material child is numeric binary execution. Their shared executor and VM
  frames are parents only, as recorded by the normal-process pair in
  `2026-07-13-concurrency-profile-and-context-abi-gate.md`.

Therefore no newly profiled named type, scheduler structure, or one-program
child qualifies for a VM change. This decision retains no bytecode VM,
compiler, bridge, runtime, canonical-stdlib, or benchmark-source change.

## Next recommendation

Do not immediately reopen a raw numeric, member-cache, inline-return, map, or
scheduler prototype: each is either an individual workload path or a generic
candidate already rejected by broad guards. The next bytecode performance
tranche should begin only when two independent, verifier-backed applications
expose the same concrete non-nominal VM descendant that is not one of those
rejected families. It should then collect matched bounded profiles, implement
one semantic-level candidate, and accept it only through the full 23-program
`coverage` catalog plus local feature guards. This preserves the project goal:
faster ordinary Able programs, rather than a faster scorecard row.
