# Dependency-plan cross-application profile gate — 2026-07-14

## Decision

Keep no bytecode VM, compiler, generated runtime, canonical-stdlib, or
benchmark-source optimization from this tranche. The new Dependency Plan
application supplies an unlike Array/Queue/topological-workload control, but
its material bytecode descendants do not form a new shared leaf with Word
Frequency and Document Audit. The only three-way recurring VM leaf is the
previously rejected `finishInlineReturn` family.

## Verifier-backed normal-process comparison

Each fresh Go, Python, and Ruby reference row ran three times and passed its
public Ruby verifier. The retained Able comparison also ran each mode three
times and verified every captured Able output. The references ran on CPU 6;
the retained Able rows ran on CPU 1. Both selected cores passed their own
immediate three-sample quiet-host preflight. These short complete-process
measurements include loading/startup, so they establish real application gaps
but do not identify an inner VM or generated-code optimization by themselves.

| Application | Mode | Able | Go | Able/Go | Ruby | Able/Ruby | Python | Able/Python |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Dependency Plan | compiled | 0.1000 s | 0.0035 s | 28.57x | 0.0427 s | 2.34x | 0.0189 s | 5.29x |
| Dependency Plan | bytecode | 0.4600 s | 0.0035 s | 131.43x | 0.0427 s | 10.77x | 0.0189 s | 24.34x |
| Word Frequency | compiled | 0.2033 s | 0.0049 s | 41.49x | 0.0473 s | 4.30x | 0.0170 s | 11.96x |
| Word Frequency | bytecode | 1.3933 s | 0.0049 s | 284.35x | 0.0473 s | 29.46x | 0.0170 s | 81.96x |
| Document Audit | compiled | 0.0967 s | 0.0038 s | 25.45x | 0.0393 s | 2.46x | 0.0132 s | 7.33x |
| Document Audit | bytecode | 0.2967 s | 0.0038 s | 78.08x | 0.0393 s | 7.55x | 0.0132 s | 22.48x |

The raw reports are in
`v12/tmp/perf/20260714_dependency_plan_profile_gate/` as `comparison.*`,
`go-refs.*`, and `interpreter-refs.*`.

## Bounded warmed bytecode evidence

The complete-program benchmark loads the real target, warms `main` once, and
then profiles only repeated calls. Every profile was preceded by a passing
three-sample host check, used `GOMEMLIMIT=1GiB`, `GOGC=50`, and
`GOMAXPROCS=1`, and was bounded by a 45-second process guard.

| Application | Core | Calls | ns/op | B/op | allocs/op | CPU samples |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Document Audit | 12 | 900 | 10,109,989 | 379,218 | 482 | 9.07 s |
| Word Frequency | 12 | 7 | 1,035,678,718 | 48,395,809 | 631,191 | 7.23 s |
| Dependency Plan | 4 | 28 | 246,373,815 | 2,190,229 | 33,241 | 6.89 s |

Document Audit is materially lazy-iterator/member-cache work:
`lookupCachedMemberMethodEntry` accounts for 28.22% cumulative, while
`finishInlineReturn` is 7.39%. Word Frequency is generic string-key HashMap,
raw-integer, and typed-result work: `hashMapFindEntryWithHash` is 7.33% flat
and `finishInlineReturn` is 15.77% cumulative. Dependency Plan is generic
Array/Queue/member and integer-index work: member lookup is 13.35% cumulative,
`runtime.mapaccess2_fast64` is 8.85%, and `finishInlineReturn` is 9.14%.

`finishInlineReturn` is therefore the only concrete three-way VM descendant.
It cannot be selected: the prior generic guard-order candidate was
neutral-to-mixed and was reverted across unlike workloads. The member-cache
leaf occurs only in the two non-map callers; Word Frequency does not repeat
it. The remaining map, Array/Queue, raw-integer, and type-match descendants
are application-shape-specific. This evidence does not permit a HashMap,
Array, Queue, iterator, text, graph, or benchmark exception.

## Harness parity repair

The first warmed Dependency Plan attempt exposed a harness-only mismatch:
`bench_perf --source-root-only` set the documented environment flag, but the
complete-program bytecode benchmark still added the input-data CWD as a user
source root. A sibling `run.able` with the same package name then created a
duplicate-package load error. `buildExecSearchPaths` now honors the existing
flag in this test-only benchmark path, and a regression test proves that an
explicit entry can share a package name with a data-directory source without
making that directory a source root. Normal fixture and CLI search behavior is
unchanged.

## Verification

- `TestBuildExecSearchPathsSourceRootOnlyExcludesWorkingDirectoryPackage`
  passes in 0.059 seconds.
- `just bench-catalog-check` confirms the complete 32-application / 77-fixture
  corpus (109 programs), and `git diff --check` passes.
- The aggregate `go test ./pkg/interpreter` parity corpus did not complete
  inside either the 60- or 90-second project guard on this host; it timed out
  in the existing exhaustive `TestExecFixtureParity` corpus. This result is
  not claimed as a full-package pass and is separate from the focused
  source-root-only regression above.

## Follow-up

Retain the scorecard and profiles as regression evidence, but do not start a
new return/cache/map/collection experiment. Reopen implementation selection
only when a newly needed, verifier-backed unlike application repeats a
concrete descendant that has not already failed broad controls. The temporary
reports, profile binaries, host-image cache, and `.pprof` files are
cleanup-eligible when active performance investigation pauses.
