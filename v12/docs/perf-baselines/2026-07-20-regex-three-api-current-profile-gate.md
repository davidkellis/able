# Regex three-API current profile gate (2026-07-20)

## Decision

Keep the benchmark-scale and scorecard-maintenance changes from this tranche,
but keep no compiler, generated-runtime, bytecode VM, or canonical-stdlib
performance change.

Regex Suffix Audit already supplied the third unlike public API shape requested
by the preceding numeric/wide gate. It was not missing; its 16,384-word default
made the bytecode row time out and therefore hid it from strict selection. The
four implementations now use the same 512-word default as Regex Set and Regex
Stream, the verifier expects the corresponding deterministic output, and the
bytecode row is selected. This is a workload rebalance across Able, Go, Python,
and Ruby, not an Able-only shortcut.

Fresh current-source profiles reproduce the same canonical NFA algorithm in
all three applications. Every concrete material operation is either an
already-retained general improvement or an already-rejected design with broad
wall-time evidence. No new exact child clears the project's admission bar, so
no runtime candidate was built.

## Benchmark and source reconciliation

The canonical and external Able copies now both use `fs.read_lines_prefix`,
and the external Set and Stream copies were synchronized to the same existing
canonical reader while this family was checked. Suffix changed uniformly:

- Able default word limit: 16,384 to 512;
- Go, Python, and Ruby default word limit: 16,384 to 512;
- verifier output: `65536:5356:925804` to `2048:228:36640`; and
- bytecode selection: previously excluded timeout, now a five-run verified
  selected row.

A separate 128-word profiling argument remains available for bounded diagnostic
runs. It does not replace the 512-word scorecard application.

Bounded semantic checks produced identical output in Able bytecode, Go,
Python, and Ruby:

| Application | 128-word output |
| --- | --- |
| Regex Suffix | `512:56:8952` |
| Regex Set | `512:28:5660` |
| Regex Stream | `512:56:12024` |

The new default Suffix output also matched in all four languages.

## Promoted application evidence

The targeted Suffix refresh used five independent verifier-backed processes
for Able and every required reference. Historical non-Suffix rows were kept
unchanged through filtered source artifacts; the obsolete Suffix timeout and
old-scale reference rows were not reused. The promoted scorecard now contains
65 selected rows (36 compiled and 29 bytecode) and 72 full-status rows. Every
selected row has exactly five successful Able/reference samples.

Current promoted regex rows are:

| Application | Compiled Able / Go | Bytecode Able / Python | Bytecode Able / Ruby |
| --- | ---: | ---: | ---: |
| Regex Suffix | 25.93x | 201.68x | 88.55x |
| Regex Set | 21.89x | 215.76x | 94.72x |
| Regex Stream | 22.69x | 178.90x | 69.07x |

Suffix's refreshed process means are 0.140 s compiled and 3.852 s bytecode,
against 0.0054 s Go, 0.0191 s Python, and 0.0435 s Ruby. A separate same-source
five-process gate measured 0.150 s compiled and 3.380 s bytecode. The spread is
consistent with the workstation policy: use repeated means for status and use
profiles for attribution, rather than selecting a candidate from a single
timing.

The refreshed current scorecard has four established compiled target meets out
of 36 and two bytecode target meets out of 29. Suffix is a large miss in both
modes and therefore a useful discriminator, not a synthetic pass.

## Profile contract

One preserved `interpreter.test` binary with SHA-256
`c99b399a1883d0ffa8b351a8762aae377b621fd82bea6cb09a8c1f5ed1f94af8`
served every bytecode profile. The current generated Suffix binary had SHA-256
`1f92fba0e5ac7db691688fb832fba7ef5919e5b58a68b1849a985ecce023e9bd`;
the current Set and Stream binaries had SHA-256
`a1239684639d61857418f6924b6daa360c1f0b6d7c6cc2a180ff63cc2e181a50`
and `b6dff6d83876a9c1f284f57e56bd24d7b39f53f21570724d4b894cbbc4bcba7b`.

Runs were serial with `GOMAXPROCS=1`, `GOGC=50`, a 1 GiB Go memory limit,
and a 55-second per-process cap. Bytecode CPU runs loaded and typechecked once,
warmed `main`, and measured three calls. Compiled CPU profiles merged 40
separate verified launches per application so short mains accumulated useful
samples without executing multiple mains in one process.

## Compiled attribution

All three generated profiles are led by the same exact canonical functions:

| Application | Samples | `add_closure` cumulative | `move` cumulative | `upsert_thread` cumulative |
| --- | ---: | ---: | ---: | ---: |
| Suffix | 1.42 s | 28.87% | 30.28% | 14.08% |
| Set | 1.11 s | 39.64% | 31.53% | 17.12% |
| Stream | 1.34 s | 29.85% | 29.85% | 9.70% |

Exact generated-main allocation counters were:

| Application | Bytes | Allocations |
| --- | ---: | ---: |
| Suffix | 8,834,008 | 201,510 |
| Set | 6,942,016 | 103,780 |
| Stream | 8,587,536 | 240,704 |

`regex_nfa_threads_new`, capture cloning, codepoint arrays, Array storage, and
GC account for the repeated allocation/CPU wall. This is the same ownership
family covered by the retained closure scratch, immutable initial capture
template, and primitive active-thread carrier changes, plus the rejected
thread-arena and state-index alternatives.

## Bytecode attribution

The current default warmed profiles reported:

| Application | ns/op | B/op | allocs/op | CPU samples |
| --- | ---: | ---: | ---: | ---: |
| Suffix | 2,917,263,593 | 45,277,613 | 543,510 | 8.71 s |
| Set | 3,462,616,282 | 40,647,296 | 338,494 | 10.35 s |
| Stream | 2,884,177,279 | 36,808,736 | 526,393 | 8.61 s |

`execCallMemberArraySlot` is 22.85%, 24.93%, and 21.02% cumulative,
respectively. Its exact traffic is the canonical parallel-array NFA carrier,
not three unrelated user Array algorithms. Current-source traces make the
shared sites explicit:

| `regex_nfa.able` site | Suffix hits | Set hits | Stream hits |
| --- | ---: | ---: | ---: |
| line 681 transition read | 238,276 | 380,268 | 244,820 |
| line 794 move transition read | 210,560 | 342,372 | 221,188 |
| line 631 carrier length | 187,240 | 283,000 | 187,640 |

The generic VM Array member and canonical-slot caches already serve these
sites. The traces do not expose a missing cache or a distinct invalidation
case.

## Candidate reconciliation

No candidate advanced because every repeated concrete option is closed by
current evidence:

- outgoing-transition indexing, reusable closure scratch, the immutable
  capture template, and primitive parallel active-thread carriers are retained;
- a matcher-owned thread arena reduced allocations but slowed Suffix, Set, and
  Stream by 9.2%, 6.5%, and 2.6%;
- a reusable state-to-position index removed the scan but made normal Stream
  bytecode 21.3% slower;
- character-to-codepoint specialization failed non-regex character generality;
  and
- generic raw-integer, Array member/cache, call/return, and carrier alternatives
  have already received unlike-program gates without a broad wall-time win.

The fresh profiles validate recurrence, but do not invalidate any rejection.
Adding a regex nominal compiler branch or a benchmark-shaped NFA opcode would
violate the shared nominal/runtime rules and is not authorized.

## Verification

- `just bench-catalog-check` passed: 36 portable applications, 79 bounded
  fixtures, and 115 combined programs.
- `just bench-selection-check` passed: 65 selected rows, manifest SHA-256
  `e7b35985b05134e1619be193cbe21ddce846cc2392efe78560e629de048d97dc`.
- the strict scorecard evidence check passed with five successful
  Able/reference samples for all 65 selected rows;
- `just bench-scoreboard-check` passed after targeted reconciliation;
- all `14_*regex*` exec fixtures passed in bytecode (12.048 s) and tree-walker
  (12.690 s) modes against the canonical external stdlib; and
- no WASM work was performed.

## Next recommendation

Run a compiled control-flow and call-boundary profile gate over Sudoku Masks,
TapeLang Alphabet, Fib, and Matrix Multiply.

Why: the current compiled product has only four established target meets out of
36. Regex, numeric/wide, text/map/graph, byte/output, concurrency, and several
representation/call-carrier ideas now have current closed evidence. Sudoku
Masks is the largest non-meeting compiled ratio in the remaining selection at
15.58x Go; TapeLang is 1.96x, Fib is 1.06x, and Matrix Multiply is 1.24x.
Together they exercise unlike bit-mask search, interpreter-style dispatch,
recursion, and primitive nested loops, which can reveal whether one generated
branch/call/environment operation is broadly material instead of optimizing
one algorithm.

What it entails: preserve one compiler and the four generated binaries; collect
main-only CPU and exact allocation profiles with verifier-backed repeated
process means; attribute primitive comparisons/bitwise operations, generated
function calls, recursion frames, loop control, environment publication, and
GC by exact descendant. Advance at most one primitive lowering or shared
generated-runtime operation present in at least three programs, then gate it
with alternating five-run averages and the current compiled target meets. Do
not add Sudoku-, TapeLang-, recursion-, or matrix-specific lowering, and do not
begin WASM.
