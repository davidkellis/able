# Release-consolidation index review

Date: 2026-07-27

Decision: commit the exact reviewed v12 release-consolidation boundary after
explicit maintainer authorization, while retaining the deferred WASM boundary
outside history.

## Exact staging boundary

The four dependency-ordered release manifests produce an original stageable
union of 5,054 paths:

- 5,046 source, test, documentation, and evidence paths whose recorded line,
  byte, and SHA-256 identities were revalidated before staging; and
- eight non-self-referential final-boundary metadata files: its review pair,
  manifest index, and five manifest parts.

This review and its JSON companion add two release metadata paths, producing a
final staged index of 5,056 paths.

The complement is exactly 34 deferred WASM paths: eight tracked modifications
and 26 untracked paths. No broad `git add` pathspec was used. The exact
manifest-derived path list was supplied to Git, and the resulting cached path
list was compared byte-for-byte with that list.

## Manifest identities

All 5,046 manifest-recorded paths match their final line count, byte count,
and SHA-256 identity. The final documentation/evidence manifest has conceptual
combined SHA-256
`8bc063622617d82a04e644a62d21b4ec56a4b24faed163a3d0a7319670c1c0cc`.
Its part identities are:

- part 01:
  `45990748cbcdc6331074f079b3cd5609cdfb99d6deadd331a4bd3ca0b15f960c`;
- part 02:
  `d28ba8a74d1394ca13b9a4461d2f19ca9abf528f721a0ca770379db916157f29`;
- part 03:
  `46d96a8c1516d44bb6c60cf5f30bea28045f90e12d6f62b68068e24c784055ed`;
- part 04:
  `f5ae39b7af3c7006df050222a8c3624adb85ae6a11c94f84f5a3c0698a1aa9d8`;
- part 05:
  `db6ff3e16a6942c33b30742cb0ec8ec6545880ea62e379803e22f41fd664e71c`.

The preceding language, runtime, and compiler manifest SHA-256 identities are,
respectively:

- `5200fc5d8e698b828213f20659209d71caa52e4f71da468636979f3337d3482a`;
- `2b8b9a54db9487f54ca7d58a78a4c7ca4c36b2f338478a6df98fe4eb090404e0`;
- `9072a4a4762cf9736979f69612e168df63bd2b064c469b49751e0d5579cd84ce`.

## Cached-index findings

The first `git diff --cached --check` exposed 19 pre-existing whitespace
defects hidden in previously untracked files: three trailing-space cases and
16 extra blank lines at end of file. They were corrected mechanically. The
two affected executable contract programs pass:

- `python3 v12/bench_refresh_external_scorecard_test.py`: two tests;
- `python3 v12/bench_shared_value_heap_production_pilot_test.py`: two tests.

Only the affected final-manifest rows and their derived identities changed.
The final cached whitespace check passes.

The staged boundary contains no v10/v11 path, deprecated in-tree stdlib path,
deferred WASM path, generated-local cache, or secret-like filename. A cached
content scan found no common private-key or service-token signature. The only
two binary artifacts are the reviewed gzip profiles:

- `2026-07-21-validated-job-file-entry-bytecode-profile.pb.gz`;
- `2026-07-21-validated-job-file-entry-compiled-main-profile.pb.gz`.

The one atomic commit containing this review is the authorized history
operation. No path outside the exact index entered it. No reset, revert,
history rewrite, dependency change, stdlib implementation change,
benchmark-specific optimization, or WASM work occurred.

## Next recommendation

Keep production performance mutation paused until an authoritative
invalidation identifies a non-closed owner material in at least three unlike
applications.

Why: the committed release state already reconciles the broad coverage
catalog, current scoreboard, frozen profiles, and closure ledger, with no
admitted general production candidate.

What it entails: refresh only affected evidence when a broad application,
retained semantic/source change, correctness failure, or new shared owner
invalidates the current state; then apply the existing verifier-backed,
repeated three-unlike-application admission protocol.

Why it is important: this prevents repeated closed experiments and preserves
the native-carrier, general nominal-lowering, and minimal
compiled/interpreted-boundary rules while waiting for evidence that can
support another general change.
