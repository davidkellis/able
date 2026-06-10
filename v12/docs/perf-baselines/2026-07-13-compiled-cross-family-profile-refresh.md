# Compiled cross-family profile refresh — 2026-07-13

## Decision

Keep no compiler, bridge, runtime, bytecode VM, or canonical-stdlib change.
Fresh compiled generated-main profiles of four high-miss applications do not
identify a concrete language-level operation that is material across unrelated
program shapes. The visible walls are numeric math/package switching, text-map
boxing and conversion, recursive search allocation, and named `Tape` methods.

This explicitly rejects a K-Nucleotide map shortcut, an N-body math/package
shortcut, a Sudoku search specialization, and a `Tape` nominal-type lowering.
None would improve Able programs generally.

## Method

All processes used the canonical external `able-stdlib`, CPU 15,
`GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, a 45-second cap, and each
benchmark's canonical Ruby verifier. The selected applications deliberately
cover unrelated high-miss families:

| Application | Family | Clean CPU run | Verification |
| --- | --- | ---: | --- |
| N-body | primitive numeric kernel | 0.4800s | verified |
| K-Nucleotide | text/counting and map use | 4.2500s | verified |
| Sudoku Masks | recursive search | 10.0400s | verified |
| TapeLang Alphabet | user nominal methods | 3.9600s | verified |

`ABLE_GO_PHASE_CPU_PROFILE_DIR` captured bootstrap and generated-main phases
without changing allocation sampling. Exact phase allocation snapshots were
also attempted with `ABLE_GO_PHASE_PROFILE_DIR`. N-body and TapeLang completed;
the allocation-heavy K-Nucleotide and Sudoku runs timed out under the normal
45-second application guard because exact sampling and boundary profile writes
are diagnostic instrumentation. They are recorded as diagnostic caps, not
performance results. Lightweight process allocation profiles completed for
those two applications and are used only for allocation ownership.

## Main-phase CPU attribution

| Application | Main samples | Material attribution | Interpretation |
| --- | ---: | --- | --- |
| N-body | 380ms | generated `sqrt` 220ms cumulative, `abs` 80ms flat, bridge environment swaps 60ms cumulative | numeric/package path only |
| K-Nucleotide | 4.03s | `mallocgc` 1.60s cumulative, `HashMap.raw_set_spec` 1.16s, primitive key equality/hash, `AsUint` | text/counting map and carrier conversion path |
| Sudoku Masks | 9.80s | `find_best_empty` 9.63s cumulative, `mallocgc` 6.06s, `growslice` 4.27s | recursive search allocation path |
| TapeLang Alphabet | 3.84s | `Tape.inc` 1.04s flat, `Tape.get` 350ms flat, `Tape.move` 100ms | named user nominal type path |

`runtime.tryDeferToSpanScan` appears in K-Nucleotide and Sudoku, but their
parents differ: map/string boxing and conversion in the former, recursive
search allocation in the latter. It is an allocator scanning frame, not a
shared Able semantic operation.

## Allocation ownership

The exact completed phase counters are:

| Application | Main allocated bytes | Main allocations | Main GCs |
| --- | ---: | ---: | ---: |
| N-body | 2,455,952 | 333 | 0 |
| TapeLang Alphabet | 2,840,040 | 5,098 | 0 |

The lightweight process allocation profiles make the two capped diagnostics
unambiguous:

- K-Nucleotide allocated 1,555.55MB / 30.68M objects. `bridge.ToInt` owns
  35.8% of bytes and 39.7% of objects, `bridge.ToUint` 23.3% / 25.8%, and
  generated String conversion 23.4% / 19.3%. Their callers are K-Nucleotide's
  `HashMap.raw_get_spec`/`raw_set_spec`, `String.starts_with`, and byte
  conversion. This is not shared with the other three shapes and cannot
  authorize a map-specific or carrier-only shortcut.
- Sudoku Masks allocated 2,916.26MB / 157.11M objects; generated
  `find_best_empty` owns 98.6% of bytes and 99.6% of objects. The cost is
  entirely its recursive search strategy, not a general runtime boundary.

Retained artifacts live in `v12/interpreters/go/.profiles/` under:

- `20260713_compiled_cross_family_cpu_*`
- `20260713_compiled_cross_family_alloc_*`
- `20260713_compiled_cross_family_process_alloc_*`

## Why no candidate is justified

The only apparently reusable primitive bridge work, `ToInt`/`ToUint`, is
material in K-Nucleotide's text/map flow alone in this set; N-body is dominated
by math and package switching, Sudoku by recursion, and TapeLang by its named
methods. The common GC scanning parent is reached from separate allocations.
A generalized rewrite would therefore optimize one benchmark family while
risking unrelated applications—the exact failure mode this suite is meant to
prevent.

No canonical `able-stdlib` update is needed.

## Next

Do not iterate on these four individual hotspots. Refresh the next compiler
selection only when an external application lane shows the same concrete
primitive or dynamic-boundary descendant in at least two unrelated programs;
then use the complete external generality suite as the acceptance gate. In
parallel, retain the bytecode-versus-Python/Ruby scorecard as its independent
selection lane rather than projecting compiled-runtime evidence onto the VM.
