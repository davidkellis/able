# Recursive-control selection check (2026-07-12)

## Purpose

Sudoku Masks is a material compiled-Able miss against Go, but one application
cannot justify a compiler or runtime change. This check screened the proposed
independent recursive controls before collecting profiles.

## Result

Fresh Go 1.26.4 QuickSort completed ten CPU-2-pinned, verifier-backed runs at
`2.9685s` mean. The immediately preceding unchanged compiled-Able ten-run row
is `2.0420s`, or `0.69x` fresh Go. QuickSort is therefore a healthy recursive
control, not a second compiler miss.

BinaryTrees is not a suitable substitute in this lane. Its external reference
is explicitly parallel; under the fair one-core (`GOMAXPROCS=1`, CPU 2) policy,
individual Go processes ran for tens of seconds. The ten-run refresh was
stopped rather than spend several minutes collecting a parallel allocation
workload that does not test the same recursive mutable-array shape as Sudoku
Masks. No partial captures are treated as a scorecard result.

## Decision

Do not profile Sudoku Masks alone and do not add a recursive, tree, array, or
solver-specific lowering. There is no repeated material miss: QuickSort is
faster than Go, and BinaryTrees is outside the selected execution model. No
compiler, VM, or `able-stdlib` source changed.

## Next recommendation

Return to a broad fresh-Go compiler scorecard and select the next pair from
two existing, single-threaded material misses rather than forcing a recursive
pair. Why: a shared optimization must follow evidence across ordinary programs,
not a desired category. The work entails pinned verifier-backed Go/Able rows
for a bounded candidate slice, then profiles only where two misses expose the
same concrete generated-code or runtime helper.
