# Compiled generality scorecard and CPU-miss gate — 2026-07-13

## Decision

Keep no compiler, runtime, or canonical-stdlib change. A fresh, pinned
compiled-versus-Go screen confirms that Able has substantial application gaps,
but CPU attribution separates those gaps into independent workload leaves. No
compiler/runtime operation is material in enough different shapes to justify a
general lowering change. In particular, this result does not authorize a
HashMap, `Tape`, Sudoku, FASTA, output, or application-specific fast path.

## Fresh scorecard

Go 1.26.4 references and compiled Able each used one CPU-15-pinned,
verifier-backed process with `GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and
a 120-second cap. One process is a selection screen, not a release timing
claim. The fresh Go rows and compiled comparison are retained in
`2026-07-13-compiled-generality-go-refresh.*` and
`2026-07-13-compiled-generality-scorecard.*`. Sudoku Masks was refreshed and
compared in the matching supplementary `compiled-sudoku-masks-*` artifacts
after its Go row was omitted from the first command list.

The compiler target is Able/Go at most `1.0526x` (at least 95% of Go speed).

| Benchmark | Compiled Able (s) | Fresh Go (s) | Able/Go | Result |
| --- | ---: | ---: | ---: | --- |
| Fib | 3.8100 | 3.2278 | 1.18x | target miss |
| BinaryTrees | 33.3800 | 34.0543 | 0.98x | meets target |
| MatrixMultiply | 1.1800 | 0.9830 | 1.20x | target miss |
| QuickSort | 1.8900 | 2.7048 | 0.70x | meets target |
| Sudoku | timeout | 0.1479 | n/a | cap-bound |
| Sudoku Masks | 9.1500 | 0.5746 | 15.92x | target miss |
| I-Before-E | 0.2700 | 0.0607 | 4.45x | target miss |
| Base64 | 2.6900 | 2.4982 | 1.08x | target miss |
| JSON | 0.8500 | 1.5526 | 0.55x | meets target |
| Monte Carlo Pi | 0.2300 | 0.2022 | 1.14x | target miss |
| PiDigits | 1.4200 | 1.2443 | 1.14x | target miss |
| Mandelbrot | 0.2000 | 0.0505 | 3.96x | target miss |
| Reverse Complement | 0.1200 | 0.0179 | 6.70x | target miss |
| K-Nucleotide | 3.7200 | 0.0710 | 52.39x | target miss |
| N-body | 0.4300 | 0.0400 | 10.75x | target miss |
| TapeLang Alphabet | 3.7300 | 1.9688 | 1.89x | target miss |

Fifteen rows are rankable; BinaryTrees, QuickSort, and JSON meet the target.
Every completed Able row was accepted by its canonical verifier. Sudoku remains
a timeout status rather than a manufactured ratio.

## CPU-only miss gate

Only the clear misses at or above `1.5x` were profiled. Each compiled profile
again used the same CPU/memory guards, a canonical verifier, and
`ABLE_GO_PHASE_CPU_PROFILE_DIR`, so allocation-snapshot work cannot select a
CPU target.

| Workload | CPU duration | Dominant attribution | Why it does not authorize a change |
| --- | ---: | --- | --- |
| I-Before-E | 15.26 ms | file-line helper, `String_len_bytes`, GC | too short for stable leaf attribution |
| Mandelbrot | 32.42 ms | generated `pixel_byte` | one numeric worker only, and too short for a broad claim |
| Reverse Complement | 7.90 ms | allocation/write barrier | too short for stable leaf attribution |
| K-Nucleotide | 3.51 s | generic map lookup/equality, primitive boxing, allocation/GC | the prior Word Frequency/K-Nucleotide gate already rejects a named-map or dynamic-carrier shortcut |
| N-body | 323.63 ms | primitive `sqrt`/`abs` and environment swaps | no matching material primitive-math leaf in the other controls |
| TapeLang Alphabet | 3.65 s | generated user `Tape` methods | a named nominal-type rule is prohibited and does not recur |
| Sudoku Masks | 8.89 s | recursive `find_best_empty` plus allocation/GC | recursive search/allocation, not a shared language boundary |

The long-enough profiles are disjoint. The common runtime allocation-scanning
frames arise from their separate callers and are not an Able semantic operation.
The short profiles are retained as status evidence only. Profile files use the
prefix `20260713_compiled_generality_miss_cpu_` in
`v12/interpreters/go/.profiles/`.

## Why no candidate is justified

K-Nucleotide's map/boxing evidence is real, but the existing independent Word
Frequency/K-Nucleotide gate already shows that its apparent commonality is a
text/counting `HashMap` shape, not a safe language-wide dynamic-value change.
N-body, TapeLang, Sudoku Masks, and the short text/numeric lanes do not repeat
that helper or each other. A candidate from any one would violate the broad
application rule and the staged-AOT prohibition on nominal-container special
cases.

No canonical `able-stdlib` change is needed.

## Verification

- All 16 fresh Go rows and all completed compiled rows were verifier-backed;
  Sudoku is explicitly recorded as a timeout.
- All seven CPU-only profile runs completed and their Able stdout was verified.
- No source behavior changed, so no new semantic test is required.
- `git diff --check` passes.

The repository-wide `./run_all_tests.sh` remains blocked before Go tests by
the existing untracked `exec/12_09_nested_spawn_native_context` fixture missing
from the already-modified exec coverage index. This tranche leaves that
fixture and index untouched.

## Next recommendation

Refresh the bytecode interpreter's same-suite comparison against fresh Python
and Ruby references, then profile only the repeated, target-miss application
families. The compiled scorecard is now a durable guard but offers no safe
generic candidate; the bytecode target is an independent project goal and may
expose a shared VM operation instead. Use pinned CPU/memory limits, canonical
verifiers, and CPU-only phase profiles, with the compiled scorecard retained as
a no-regression application guard.
