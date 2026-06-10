# Source Equivalence and Bounded Line Reading

## Decision

Retain one broad canonical-stdlib repair: `able.fs.read_lines_prefix(path,
count)`. Five unrelated benchmark applications requested a bounded prefix but
used `fs.read_lines(path).take(count)`, which eagerly decoded and materialized
all 172,823 lines in the shared word list before discarding nearly all of them.
Their Go, Python, and Ruby counterparts stop reading at the requested prefix.

This is a real source/library-boundary mismatch, not a VM opcode or named
benchmark fast path. The new API is usable by any program that needs the first
N normalized text lines. The five applications now use it; no compiler or VM
special case was added.

## Equivalence audit

| Family | Algorithm and work | Classification |
| --- | --- | --- |
| Fixed Width 128 | All four languages perform the same 1,000,000 modular-add and ordered-select rounds. Able exercises the canonical `UInt128` nominal API, including primitive-u128 conversion and overflow semantics; the references use direct two-word/workload-specific operations. | No source rewrite. The remaining gap is canonical representation and semantic cost. |
| Rational Series | All four languages perform 50,000 equivalent rational updates. Able's canonical Rational uses i128 intermediates and checked division/overflow behavior; Go uses workload-specific unchecked int64 arithmetic, while Python/Ruby use their native arbitrary-precision integers. | No fair application-only rewrite. The remaining gap is stdlib representation and required semantics. |
| Concurrency | Channel, Future, cancellation, Await, and Mutex topologies, rounds, and payload work align with the references. | No algorithm mismatch except Channel Rollup's eager input boundary. |
| Regex/text | Outputs and requested matching work align. Go/Python/Ruby regex rows delegate to their mature native regex engines, while Able exercises its canonical Able NFA. Reverse Complement also uses native translate/reverse helpers in Python/Ruby versus explicit loops in Go/Able. | These ratios include real language-library implementation differences; do not disguise them with benchmark-only delegation. Word Frequency and Option/Result Config are structurally aligned. |

The bounded-input mismatch occurred in Channel Rollup, Lexical Rollup, Regex
Suffix Audit, Regex Set Audit, and Regex Stream Audit. The first two request
16,384 lines, and the regex applications request 512 lines (128 under their
profile settings).

## Retained API and implementation

`read_lines_prefix(path: Path | String, count: i32) -> Array String`:

- opens and validates the path even when `count <= 0`;
- returns at most `count` lines;
- recognizes LF, CRLF, bare CR, and a final unterminated line consistently
  with the existing line APIs; and
- closes the file as soon as the requested prefix or EOF is reached without
  reading the remainder.

The Go host boundary uses a bounded `bufio.Scanner` with the shared line
normalization contract. The TypeScript host boundary uses bounded chunked
reads and a streaming UTF-8 decoder. Both return a lossless normalized text
prefix through the existing error-union machinery, then the stdlib uses its
shared line splitter. Canonical stdlib tests cover CRLF, bare CR, LF, zero,
and negative counts. The v12 spec now states the public contract.

## Rejected formulations

Two broader-looking alternatives were tested and completely removed:

1. A generic `Iterator.collect_prefix<C>` passed tree-walker and bytecode
   stdlib tests, but compiled applications failed with a non-exhaustive match.
2. App-local `fs.lines` iteration encountered a typechecker limitation in the
   direct form. A manual `next()` form compiled, but Channel Rollup bytecode
   regressed to about 1.66 seconds from 0.634 seconds.

Those results are why the retained repair sits at the filesystem boundary,
where early termination can actually avoid the I/O, decode, and allocation
work in every execution mode.

## Focused performance gate

The baseline below is promoted 2026-07-15 cohort B. Candidate values are fresh
five-process verified means, except Channel Rollup bytecode, where two
five-process batches were combined because the first contained one workstation
outlier. Regex Suffix bytecode remains a bounded timeout.

| Application | Mode | Baseline (s) | Candidate (s) | Change |
| --- | --- | ---: | ---: | ---: |
| Channel Rollup | compiled | 1.232 | 1.186 | -3.7% |
| Lexical Rollup | compiled | 0.116 | 0.108 | -6.9% |
| Regex Suffix Audit | compiled | 2.744 | 2.624 | -4.4% |
| Regex Set Audit | compiled | 0.208 | 0.176 | -15.4% |
| Regex Stream Audit | compiled | 0.198 | 0.170 | -14.1% |
| Channel Rollup | bytecode | 0.634 | 0.625 | -1.4% |
| Lexical Rollup | bytecode | 0.510 | 0.434 | -14.9% |
| Regex Set Audit | bytecode | 5.380 | 5.092 | -5.4% |
| Regex Stream Audit | bytecode | 4.568 | 4.528 | -0.9% |

The subsequent full promoted cohort independently completed all 58 selected
rows at five verified runs and all six excluded rows at one fresh status probe.
Its affected means are 1.200/0.572 seconds for Channel Rollup,
0.102/0.444 for Lexical Rollup, 2.594/timeout for Regex Suffix,
0.190/4.972 for Regex Set, and 0.182/4.344 for Regex Stream
(compiled/bytecode). Because the source and canonical stdlib identity changed,
this cross-cohort comparison is confirmation rather than strict same-source
variance evidence.

The promoted scoreboard now contains 4/32 compiled rows meeting the Go target,
5/26 selected bytecode rows meeting Python, 4/26 meeting Ruby, and 4/26 meeting
both. Six excluded bytecode rows remain visible and unranked.

## Correctness and protocol verification

- Sequential canonical `able.fs` tree-walker and bytecode suites pass.
- All five changed applications pass focused compiled/bytecode output
  verification where the mode is selected.
- `just bench-scoreboard-check`, `just bench-catalog-check`,
  `just bench-selection-check`, and selected-status validation pass.
- The full refresh records 69 canonical stdlib sources with source-tree hash
  `44a1adeafa85b2aec82fa18b4adb1d2903f8103aa9c58953c4b89767f20c3052`.
- Diff hygiene passes, and `able-stdlib/src/fs.able` remains below the
  one-thousand-line limit at 957 lines.

The focused compiled fixture `06_12_28_stdlib_fs_lines` still fails during
generic method resolution:

`no matching impl method for able.kernel.Clone.clone target=Array<T>`

The same fixture was run with the entire new prefix API and application edits
removed and failed identically. Direct compiled benchmark applications using
`read_lines_prefix` pass. This therefore is a pre-existing compiler correctness
blocker in the current dirty source snapshot, not a regression introduced by
the retained API.

## Next recommendation

Fix the compiler's generic `Clone<Array<T>>` implementation-method resolution
failure before selecting another performance candidate.

Why: it is a reproducible, general compiler correctness defect at a shared
generic-interface boundary. It blocks the compiled resource-owning filesystem
fixture and could affect ordinary user-defined generic code; working around it
inside `fs` or a benchmark would conceal the real problem.

What it entails: minimize the fixture to the smallest Array-clone call that
loses its generic binding; trace nominal identity, interface arguments, and
method constraint substitution through compiler impl selection; repair that
shared resolution path; add a focused regression; then run the filesystem
fixture plus unrelated Array/Clone and generic-interface compiled controls.
Only if the correctness repair touches timed code should it advance to a
repeated cross-application performance gate.
