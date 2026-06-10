# Compiled generated-artifact drift closure

Date: 2026-07-27

## Decision

Retain no production compiler, generated-runtime, runtime, interpreter,
bytecode VM, canonical-stdlib, benchmark, language, dependency, or WASM
change.

The current compiler source identity differs from the compiler used by the
frozen normalized compiled profiles. A narrow report-only refresh therefore
rebuilt Tapelang Alphabet, Sudoku Masks, and Fib with the current compiler and
`--no-fallbacks`.

All eight recorded hot generated functions have exactly the same machine-code
size, instruction count, call count, conditional-branch count, and
unconditional-branch count as the frozen evidence. All nine current binary
executions passed their public verifiers, and all three final dependency
graphs remain interpreter-free.

The compiler-cache and compiled-test-cache tranches did not expose a new
generated execution owner. The three-unlike admission gate fails before CPU
or allocation profiling, a production prototype, or repeated A/B/Go timing is
justified.

The compact machine-readable companion is
`2026-07-27-compiled-generated-artifact-drift-closure.json`.

## Why this refresh was admissible

The authoritative frontier permits refreshing only evidence invalidated by a
retained implementation change. The frozen normalized generated-owner record
used compiler SHA-256:

```text
28697a5adf4f73918f3d83fbcddc211407dc7e539240f64e4127a4e3dd4ddcab
```

The current compiler built for this audit has SHA-256:

```text
2570b8372377e84406953583c6c440bc8a5c59392bc6ddbdc50d16cb9b3a6912
```

The intervening retained compiler changes were intended to change import/type
normalization and lookup cost rather than generated execution. The retained
compiled-test cache changes were also operational. This audit tested that
claim directly without reopening already-closed boundary, Array, control,
relational-proof, allocation, GC, or launch-floor experiments.

## Protocol

The selected applications are the unlike normalized cohort with exact frozen
machine-code evidence:

- Tapelang Alphabet: native flat primitive/Array dispatcher with reachable
  checked mutation;
- Sudoku Masks: native search with checked arithmetic and Array semantics; and
- Fib: native primitive recursion with a fallible-control result ABI.

Each application was emitted into a disk-backed `/var/tmp` workspace with:

```text
ablec -main -pkg main -o OUT --no-fallbacks ENTRY
go build -mod=mod -o BINARY .
```

The builds used Go 1.26.4 and the canonical external stdlib at commit
`219eff222c28406487231713753641bc49ee5b9a`.

Raw binary SHA-256 values differ from the frozen record because ordinary Go
builds retain workspace/build identity. Raw whole-binary identity was
therefore not used as execution-drift evidence. The comparison instead used
the exact named hot symbols and the same `go tool nm`/`go tool objdump`
machine-code metrics as the frozen record.

## Exact machine-code comparison

Every current metric equals the frozen normalized generated-owner metric:

| Function | Bytes | Instructions | Calls | Conditional jumps | Unconditional jumps |
| --- | ---: | ---: | ---: | ---: | ---: |
| Fib `fib` | 166 | 52 | 5 | 4 | 1 |
| Tapelang `execute` | 1,018 | 276 | 17 | 34 | 19 |
| Tapelang `Tape.inc` | 295 | 82 | 4 | 10 | 6 |
| Tapelang `Tape.move` | 248 | 66 | 4 | 5 | 3 |
| Sudoku `find_best_empty` | 784 | 198 | 10 | 13 | 5 |
| Sudoku `bit_count` | 98 | 33 | 2 | 3 | 2 |
| Sudoku `square_index` | 278 | 74 | 5 | 5 | 1 |
| Sudoku `solve_with_masks` | 1,477 | 348 | 24 | 23 | 6 |

The current primary generated source hashes are:

| Application | `compiled.go` SHA-256 |
| --- | --- |
| Tapelang Alphabet | `0a38432e2f5ab3f651cd72286f6d255ab52961dd004119f2c00e3570b62868ca` |
| Sudoku Masks | `2c85adecd8e26726cebda3f2a2d9e5b1bea51a5a9dd3582f438ab97809e2ed90` |
| Fib | `70b3e0eb1993b52e81dcb82d1192e353527e5008bdb6d4f26b02f8ca0c27ab27` |

These hashes freeze the current artifact for a future targeted drift check;
they are not compared with unavailable disposable full-source snapshots from
the prior tranche.

## Verifier and dependency result

Three independent current-binary processes per application passed the
existing public Ruby verifier:

| Application | Verified processes | Smoke wall mean |
| --- | ---: | ---: |
| Tapelang Alphabet | 3/3 | 3.8667 s |
| Sudoku Masks | 3/3 | 1.5433 s |
| Fib | 3/3 | 3.5733 s |

These means are smoke evidence only. They were not balanced with fresh Go
references and are not promoted as performance measurements.

Each final graph contains exactly 96 packages. All three `go list -deps`
results omit `able/interpreter-go/pkg/interpreter`, and the generated
`compiled*.go` sources contain no interpreter package reference.

## Admission decision

The current generated hot code is machine-identical in all three unlike
mechanisms. Consequently:

- there is no changed hot function to profile;
- there is no new exact CPU or allocation owner;
- there is no compiler/interpreter crossing to remove;
- there is no carrier drift to repair; and
- there is no general production candidate to measure.

Refreshing CPU/allocation profiles would repeat frozen evidence against
identical machine code. A production experiment would therefore violate the
new-evidence prerequisite. The correct result is no code.

## Verification and cleanup

- Three current strict builds passed with `--no-fallbacks`.
- Nine current binary processes passed public verification.
- Three 96-package dependency graphs are interpreter-free.
- Eight hot-symbol metric tuples exactly match their frozen values.
- No Able, Go reference, canonical stdlib, fixture, or production source
  changed.
- The v12 spec remained byte-for-byte unchanged.
- `./run_all_tests.sh` passed in 632.07 seconds at 4,657,832 KB peak
  RSS. Every non-compiler package, all 33 compiler batches, and the final
  86.660-second bytecode fixture corpus passed.

All generated modules, binaries, compiler copies, dependency lists, and smoke
reports are disposable after this record is complete. The reusable
disk-backed Go and compiled-test caches remain outside the task workspace.

## Next recommendation

Keep compiled production performance mutation paused. Use the next tranche
for correctness or release-readiness work selected by an actual failing gate,
or wait for one of the authoritative performance invalidations: a new broad
application, a generated-execution change, a carrier/boundary correctness
failure, or a new exact report-only owner material in three unlike
applications.

Why: the current compiler identity changed, but the hot generated machine
code did not. Profiling or editing identical execution would only repeat
closed work.

What it entails: run ordinary correctness/release gates, fix only a real
failure through shared v12 semantics, and refresh the affected compiled
profiles only if that repair changes generated execution. If a new broad
application arrives, add it to coverage and profile it before selecting a
candidate.

Why it is important: this preserves interpreter-free native carriers and the
goal of Go-equivalent execution while preventing benchmark-specific or
already-rejected compiler work. Do not begin WASM work.
