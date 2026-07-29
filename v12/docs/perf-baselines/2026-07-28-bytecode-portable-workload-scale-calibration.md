# Bytecode portable workload-scale calibration

Date: 2026-07-28

## Decision

The seven-row portable scale calibration succeeds and admits a benchmark
contract/tooling promotion. It does not admit a compiler, runtime, or bytecode
VM optimization.

Retain the calibration evidence, but leave the production catalog, benchmark
sources, external suite, current scorecard, and 63-compiled/56-bytecode
selection unchanged in this tranche. Permanent promotion requires a general
mode-aware argument contract so compiled rows retain their canonical large
workloads while bytecode, Python, and Ruby use the calibrated workloads.

## Scope

The calibrated applications are the seven compiled-selected rows that lack a
selected bytecode row:

- Binary Trees;
- N-Body;
- Quicksort;
- Sudoku Masks;
- TapeLang Alphabet;
- Fib;
- Matrix Multiply.

All calibration occurred in disk-backed temporary copies under `/var/tmp`.
No canonical Able, Python, Ruby, Go, verifier, input, compiler, interpreter,
runtime, stdlib, language, dependency, selection, scorecard, or WASM source
was changed.

The measured Able source state was
`5bbe9e4594bfb2b7e7f8c84b9fbed9314c0ee8db`. The bytecode CLI SHA-256 was
`9d3559c1295b110ed95834f35cf08a040decbf4c2ea212370e47433a3becde4f`.
The canonical stdlib source state was
`219eff222c28406487231713753641bc49ee5b9a`; its existing dirty worktree was
read but not changed.

## Calibrated contracts

| Application | Canonical contract | Calibrated bytecode/reference contract | Preserved work |
| --- | --- | --- | --- |
| Binary Trees | depth 21 | depth 15 | perfect-tree allocation, stretch/long-lived trees, recursive checks, per-depth batches |
| N-Body | 500,000 steps | 50,000 steps | identical five-body state, `dt`, float updates, and energy calculation |
| Quicksort | 10,000,000 integers | first 500,000 integers | identical comma-delimited input shape, parsing, Quicksort, and first/last-ten output |
| Sudoku Masks | ten puzzles × ten passes | ten puzzles × one pass | identical corpus prefix, bit masks, constrained-cell search, recursion, and solution validation |
| TapeLang Alphabet | seven-level decimal delay | five-level decimal delay | same parser/interpreter, reverse-alphabet program, loop behavior, and byte-exact output |
| Fib | `fib(45)` | `fib(40)` | identical naive recursive definition and 32-bit result |
| Matrix Multiply | 1000×1000 | 400×400 | identical matrix construction, transpose, cubic multiplication, and center result |

The calibrated Quicksort input is the first 500,000 values of the canonical
10,000,000-value file. It is 4,945,018 bytes and has SHA-256
`104c96eba59e80a137f9d39a8a67d7c91a0f7573a8eb9ca24bf9f95b7c140e21`.
The calibrated TapeLang input has SHA-256
`381f864256b3c636e4db0278ec11dda44e83f3f20a42317bfc370bb2083458b6`.
It is retained as
`2026-07-28-bytecode-portable-scale-tapelang-input.tape`.
The exact temporary source and output identities are in
`2026-07-28-bytecode-portable-scale-source-manifest.tsv`.

## Measurement protocol

- One logical CPU, pinned to CPU 0.
- `GOMAXPROCS=1`, `GOMEMLIMIT=1GiB`, and `GOGC=50`.
- Current source-root-only Able package resolution and a persistent external-Go
  cache.
- Five independent processes per application and language.
- Rotating launch order:
  Able/Python/Ruby, Python/Ruby/Able, Ruby/Able/Python.
- A 60-second per-process ceiling.
- Exact output comparison for six applications.
- Two-line numeric N-Body comparison with `1e-9` tolerance.

All 105 timed samples completed and verified. No individual process exceeded
24.5 seconds. The raw retained samples are in
`2026-07-28-bytecode-portable-scale-balanced-runs.tsv`.

## Results

| Application | Able (s) | Python (s) | Ruby (s) | Able/Python | Able/Ruby |
| --- | ---: | ---: | ---: | ---: | ---: |
| Binary Trees | 11.625683 | 0.584139 | 0.581757 | 19.902247× | 19.983757× |
| N-Body | 8.892018 | 0.209770 | 0.349334 | 42.389369× | 25.454187× |
| Quicksort | 12.433842 | 0.685404 | 0.690791 | 18.140901× | 17.999437× |
| Sudoku Masks | 24.084107 | 1.739264 | 2.137725 | 13.847295× | 11.266231× |
| TapeLang Alphabet | 20.495694 | 0.633506 | 0.750641 | 32.352822× | 27.304263× |
| Fib | 0.114012 | 5.287008 | 4.291891 | 0.021565× | 0.026565× |
| Matrix Multiply | 0.735755 | 3.214674 | 3.143816 | 0.228874× | 0.234032× |

The geometric mean across the fourteen Able/reference ratios is 4.232748×.
Fib and Matrix Multiply already exceed both reference targets. The five
remaining misses are large but have different application owners. Able
process coefficient of variation is 0.771%–6.712%, with the largest
percentage belonging to the 0.114-second Fib launch-floor row.

The compact machine-readable summary is
`2026-07-28-bytecode-portable-scale-summary.tsv`.

## Full-run profile reconciliation

One complete, verifier-passing Able CPU profile was collected for each of the
five misses at the calibrated scale. The retained top-40 reports are under
`2026-07-28-bytecode-portable-scale-profiles/`.

Only two exact functions appear in all five retained tops:

- `(*bytecodeVM).runResumable`, the aggregate dispatch loop;
- `bytecodeRawIntegerValueInfo`, an already-closed raw-integer carrier route.

Other functions appearing in at least four profiles are:

- `execCallOpcode`;
- `execStoreSlot`;
- `finishInlineReturn`;
- `bytecodeStackSnapshotValue`;
- `popCallFrameFields`.

These reconcile to already-closed call, store, return, stack, and frame
families. Application-specific owners remain:

- Binary Trees: allocation/GC, named structs, typed patterns, recursion;
- N-Body: float operations, calls, members, and Arrays;
- Quicksort: input parsing, integer decoding, casts, and Array reads;
- Sudoku Masks: bit operations, recursive calls, frames, and Array reads;
- TapeLang: member dispatch, named-struct plans, and Array slots.

No open exact compiler/runtime/VM owner repeats in three unlike applications.
Therefore this tranche enters no production A/B gate.

## Evidence

- Balanced runs:
  `2026-07-28-bytecode-portable-scale-balanced-runs.tsv`
  (`5ab58ba38a7318c6f8bd7a0d84b84ae711b2b81182b271513fb32c5662e83912`)
- Summary:
  `2026-07-28-bytecode-portable-scale-summary.tsv`
  (`5ac14291a4e1dd8a4c7a35456bea1a99cab011d8fd9311b824e6fefd615303aa`)
- Source manifest:
  `2026-07-28-bytecode-portable-scale-source-manifest.tsv`
  (`33f3576b9ab462baf69e309433add29d1b97496ceedb6f58a60ef698089c8bb8`)
- Calibrated TapeLang input:
  `2026-07-28-bytecode-portable-scale-tapelang-input.tape`
  (`381f864256b3c636e4db0278ec11dda44e83f3f20a42317bfc370bb2083458b6`)
- Profile run manifest:
  `2026-07-28-bytecode-portable-scale-profile-runs.tsv`
  (`b846c339877c700764c21df23c040ceb036222bb67ededf50f591e14037cc124`)
- Raw-profile identity manifest:
  `2026-07-28-bytecode-portable-scale-profile-manifest.tsv`
  (`9853577b5e8beae4966902c185fc39a6349625a14f496466f05679d8df1c19e3`)
- Current selection manifest:
  `v12/bench-selection-manifest.json`
  (`dac2450e10f73655271c2e03b236e3d2c0b4dfe83e8bdfcfda9bee4efdba9d23`)
- Current scorecard:
  `external-scoreboard-current.json`
  (`9df24ffed73a2ad39060eb2229f7588d617426fa456c3b8df3ff045d5392b53c`)

## Next recommendation

Implement and verify the general mode-aware benchmark workload contract, then
promote these seven rows only through a complete scorecard refresh.

Why: the calibration proves all seven portable scales, but the current catalog
has one program-argument contract per application. Replacing the canonical
workload would weaken compiled coverage and invalidate existing Go
comparisons.

What it entails:

1. make benchmark arguments mode-aware while preserving the current compiled
   defaults;
2. capture verifier/input contracts per benchmark and mode;
3. pass the bytecode contract to the Python/Ruby reference runner;
4. expose general workload arguments in the seven Able and reference programs,
   add the two calibrated input assets, and extend public verifiers to accept
   the declared scale;
5. add focused catalog/contract regression tests;
6. run a complete five-process compiled and bytecode scorecard refresh;
7. select all seven bytecode rows only if all 126 rows satisfy the evidence
   contract.

Why it is important: this can move bytecode coverage from 56/63 to 63/63
without shrinking compiled workloads or mixing incomparable inputs. It also
turns five timeouts into stable, full-run evidence while preserving the
barrier against benchmark-specific compiler and VM rules.
