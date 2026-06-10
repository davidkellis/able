# Filesystem, I/O, and text roadmap reconciliation — 2026-07-14

## Decision

Retire the old broad `able.fs` / `able.io` / `able.text.string` roadmap
placeholder. It does not name an unimplemented v12 behavior or a selected
cross-application performance boundary.

No compiler, bytecode VM, runtime, canonical-stdlib, fixture, or benchmark
source changed in this tranche.

## API and coverage inventory

The canonical `able-stdlib` already provides the relevant reusable surface:

| Area | Existing public surface | Semantic evidence |
| --- | --- | --- |
| Filesystem | open/stat/directory, byte/text/line read and write, copy, rename, remove, metadata, `Path` helpers | `06_12_21_stdlib_fs_path`, `06_12_22_stdlib_io_temp` |
| I/O | byte/string conversion, read/write/flush/close, `read_all`, line helpers, `BufferedReader`, `BufferedWriter` | `06_12_22_stdlib_io_temp`, blocking-I/O fixture |
| Text | UTF-8 length, substring, split, replace, prefix/suffix, byte/char/grapheme iteration, `StringBuilder` | `06_12_01_stdlib_string_helpers`, string and regex fixtures, bounded builder workloads |

Portable applications exercise the timing-relevant file paths without adding a
synthetic workload: `read_lines` appears in I-Before-E, K-Nucleotide, Sudoku,
Word Frequency, Document Audit, Lexical Rollup, Channel Rollup, and all three
regex audits; `read_bytes` appears in QuickSort, Reverse Complement, and
TapeLang; JSON uses `read_text`; and K-Nucleotide, Reverse Complement,
Mandelbrot, TapeLang, and PiDigits use the ordinary I/O conversion or output
paths.

The existing performance record already covers the only candidate classes that
this old heading could imply. General primitive `String.bytes`,
`String.from_bytes_unchecked`, `String.chars`, and `utf8_decode` lowerings
were retained only after unlike-workload gates. The general StringBuilder
conversion candidate was reverted after regressing two of three controls.
File ingress fast paths have likewise already been profiled and must not be
reopened without new shared evidence.

## Verification

The focused fixtures passed with the canonical sibling stdlib in all Go
execution modes:

```text
go test ./pkg/interpreter -run 'TestExecFixtures/(06_12_01_stdlib_string_helpers|06_12_21_stdlib_fs_path|06_12_22_stdlib_io_temp|06_12_23_stdlib_os|06_12_24_stdlib_process|06_12_25_stdlib_term)$' -exec-mode=treewalker
go test ./pkg/interpreter -run 'TestExecFixtures/(06_12_01_stdlib_string_helpers|06_12_21_stdlib_fs_path|06_12_22_stdlib_io_temp|06_12_23_stdlib_os|06_12_24_stdlib_process|06_12_25_stdlib_term)$' -exec-mode=bytecode
ABLE_COMPILER_EXEC_FIXTURES='06_12_01_stdlib_string_helpers,06_12_21_stdlib_fs_path,06_12_22_stdlib_io_temp,06_12_23_stdlib_os,06_12_24_stdlib_process,06_12_25_stdlib_term' go test ./pkg/compiler -run '^TestCompilerExecFixtures$'
```

## Consequence

Do not infer a file, line-reader, text, UTF-8, StringBuilder, or benchmark
special case from the current performance gaps. Reopen this area only when a
new specified API behavior is required or profiles identify the same concrete
language-level leaf in at least two unlike applications.
