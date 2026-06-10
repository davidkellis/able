# External Scorecard Recheck: Selection

This tranche refreshes measurement and selection; it contains no interpreter,
compiler, or standard-library change. It is intentionally report-only so the
dirty shared external results source remains untouched.

## Fresh catalog measurement

The full 15-program `generality` catalog was rerun through
`v12/bench_compare_external` with `compiled,bytecode`, one run per mode, a
25-second per-mode cap, CPU affinity `0`, and
`GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1`. The canonical stdlib was pinned at
`/home/david/sync/projects/able-stdlib/src`. The harness report is
`2026-07-10-external-scorecard-recheck.{md,json}` beside this document.

The rerun retained a result row for every catalog mode: BinaryTrees timed out
in both modes; Quicksort, K-Nucleotide, and TapeLang timed out in bytecode;
and N-body failed in both modes. Those rows are status evidence, not ratios.

The comparable results reproduce the selection shape of the earlier refresh:

| Target family | Fresh representative misses | Why it is not a shared candidate |
| --- | --- | --- |
| Compiled vs Go | I-Before-E 3.60x, Mandelbrot 5.00x, Reverse Complement 19.00x | text/file work, float kernel, and byte processing are distinct paths |
| Bytecode vs Ruby/Python | I-Before-E 5.80x/4.46x, Base64 1.39x/0.93x, Monte Carlo 1.95x/1.65x | text scanning, host codec calls, and float-loop execution are different concrete families |
| Public applications | Word-Frequency 79.74x/150.41x, Document-Audit 98.92x/225.68x, Lexical-Rollup 82.05x/201.91x | independently verified inputs expose map/return, public iterable/member, and mixed text/iterator/native work respectively |

Ratios are Able/reference wall time; values above one miss the relevant target.
The public-application values are retained Docker-verified rows from the
same-day multi-source ledger. Their inputs and publication environments differ
from the catalog, so they are used to select profile families, never to rank
absolute workload cost against catalog rows.

## Selection

No implementation candidate is authorized. The catalog produces material
misses, but not the same concrete generic descendant across independently
featured programs. The existing independent Document-Audit/Lexical-Rollup
pair also repeats only `lookupCachedMemberMethodEntry(...)` and
`finishInlineReturn(...)`, both previously evaluated and rejected by broad VM
guards. Reopening either cache/return micro-tweak would be evidence-free.

No canonical `able-stdlib` edit is required: the result identifies no shared
stdlib API boundary whose improvement could be distinguished from a
workload-specific text, codec, iterator, map, or float path.

## Next recommendation

Profile the two independent target-miss text applications, I-Before-E and
Lexical-Rollup, with Word-Frequency as an unrelated map/return guard. Both
selected applications process the checked-in word-list input and miss the
bytecode Ruby/Python floor, but their sources differ substantially (direct
text loop versus public iterator/filter/map/collect pipeline). The work should
use warmed one-process bytecode CPU profiles under the same pinned stdlib,
OOM, and CPU guardrails. A runtime or stdlib candidate is permitted only if a
material `String`/typed-pattern/primitive-dispatch descendant—not merely the
VM dispatcher—repeats in both and survives the Word-Frequency guard. This
turns the scorecard’s broad target signal into a falsifiable generic hypothesis
without creating a benchmark-shaped fast path.
