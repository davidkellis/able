# Performance frontier reconciliation

Date: 2026-07-26

Decision: retain no compiler, generated-runtime, runtime, interpreter,
bytecode VM, canonical-stdlib, benchmark, reference, fixture, language,
dependency, or WASM change. Remove the stale completed semantic-work tranche
from `PLAN.md` and make new cross-program evidence a prerequisite for further
performance mutation.

## Why this reconciliation was required

The latest strict typed-boundary record recommended a semantic-work audit of
Tapelang Alphabet, Fib, and Sudoku Masks. Chronological logs and dated records
show that the proposed tranche and all of its immediate follow-ups were
already complete:

1. the original semantic-work audit identified inequivalent Tapelang and
   Sudoku source work;
2. the benchmark normalization made those programs structurally comparable;
3. generated source, assembly, CPU, allocation, and exact work counts found
   no removable owner shared with Fib;
4. the full strict relational proof census found material safe-check removal
   in only NBody;
5. typed bytecode `i32` storage was confirmed already present;
6. the 54-row primitive-materialization census closed broad raw-to-boxed
   boundaries after a mixed/regressive nominal-field experiment;
7. primitive Array-element materialization was not independently material in
   three unlike mechanisms;
8. a local static-return carrier experiment regressed time and allocation
   count and was reverted;
9. the corpus-wide bytecode exact-owner census closed local VM selection after
   the only admitted member-template experiment regressed two unlike rows;
10. the current strict compiled scorecard, AOT owner cohort, and full typed
    callsite census found no new shared native-carrier boundary.

Selecting semantic work or call/member/index quickening again would therefore
repeat completed work without a new artifact, workload, semantic change, or
owner.

## Current stopping conditions

### Compiled

- All 61 portable coverage applications build with `--no-fallbacks`, pass
  their public verifiers, and omit `pkg/interpreter` from final dependency
  graphs.
- Primitive and static-Array carriers remain native through strict generated
  call graphs. Residual conversions are startup-only, explicit host/dynamic
  boundaries, or required concurrency services.
- Normalized Tapelang, Sudoku, and Fib profiles have disjoint remaining
  owners. Tapelang and Sudoku retain reachable Able safety behavior; Fib's
  total-function control ABI is material in only one application.
- The relational observer proved 89 of 1,006 reachable Array conditions, but
  only NBody's proof family is dynamically material.
- The post-package-cut launch floor, generated registration, local
  `runtime.Value` boundaries, control/effect ABI, checked arithmetic, generic
  storage, and generated native-owner routes all have recorded closures.

### Bytecode

- Typed primitive registers, sidecars, raw slot/stack carriers, direct stores,
  and explicit materialization at dynamic boundaries are already retained.
- The current 54-row exact CPU/allocation census spans 22 semantic families
  and closes local profile-driven VM selection.
- Concrete call/member/index cache work, slot/stack transport, Array storage,
  static returns, frames, registers, boxing, materialization, GC symptoms, and
  launch/typecheck parents are completed or rejected routes.
- The remaining Python/Ruby gap is real, but no exact non-closed production
  owner is material in three unlike current applications.

## External evidence audit

The sibling benchmark repository contains 61 portable application
directories after excluding repository metadata, the shared `able-base`
image, and legacy diagnostic Sudoku. Normalizing hyphens and underscores, the
directory set exactly equals the 61-entry Able coverage catalog:

```text
catalog applications: 61
external applications: 61
external-only: 0
catalog-only: 0
normalized set SHA-256:
6b422bf432ebb26bfa23cd2d634fc89a2784caac7b77f2686823ec9edee96775
```

No benchmark source file was newer than the latest typed-boundary closure.
There is no unrepresented external application or source refresh that can
invalidate the current stopping conditions.

## Admission rule

Further performance production work requires at least one concrete
invalidation:

1. a broad external application is added outside the current catalog;
2. a retained compiler/runtime/VM/language/stdlib change invalidates frozen
   source or artifact identity;
3. a correctness failure demonstrates that a native carrier or semantic
   boundary is wrong; or
4. a report-only observer exposes one exact non-closed owner that is material
   in at least three unlike applications.

After an invalidation, refresh only the affected evidence first. A production
candidate must still be a general rule, preserve v12 semantics, pass focused
guards, and win five or more balanced public-verifier A/B/reference pairs.

## Verification

- The benchmark-directory/catalog set comparison is exact.
- `PLAN.md` no longer selects a completed semantic-work tranche or an already
  closed local VM category.
- No source, benchmark, stdlib, fixture, runtime, dependency, or WASM file was
  changed by this reconciliation.
- `git diff --check` passes.
- The complete `./run_all_tests.sh` handoff passes every contract,
  non-compiler package, all 32 compiler batches, and the 86.401-second
  bytecode fixture corpus. The known aggregate batches 19, 28, and 29 pass in
  184.422, 74.158, and 86.767 seconds; their prior repeated audit established
  that individual tests remain below one minute.
- `./run_stdlib_tests.sh` passes the complete canonical stdlib suite in both
  tree-walker and bytecode modes (20 and 15 seconds).
- The complete-suite log SHA-256 values before task-workspace cleanup were
  `ffb99a4e97f70ea734311c93575fbc28ec42bc2946dd7fa5542b757e26836569`
  and
  `32f3912a065594ae07ea9a1db1bad06c6189f89473c471a0246dc80d8db6c4d9`.

## Next recommendation

Keep performance mutation paused until one admission invalidation exists; use
the next session for correctness or release-readiness work selected from an
actual failing gate.

Why: the current corpus has exhausted both local compiled/native-boundary
selection and local VM exact-owner selection. More implementation without new
evidence would optimize an aggregate parent, one application, or required
Able semantics.

What it entails: run the ordinary correctness and release gates, repair any
real failure through the shared v12 semantics, and refresh only the affected
performance rows if the repair changes execution ownership. If the external
suite gains a new broad application, add it to coverage and profile it before
reopening a candidate.

Why it is important: this preserves the achieved interpreter-free compiled
architecture and native primitive carriers while preventing benchmark-specific
rules and noisy regressions. Do not begin WASM work.
