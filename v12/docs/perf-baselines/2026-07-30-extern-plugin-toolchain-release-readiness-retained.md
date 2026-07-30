# Extern plugin toolchain release-readiness closure

Date: 2026-07-30

## Decision

Retain one general interpreter correctness fix: generated Go extern plugins
are now built with the exact Go toolchain that built the running Able host.

The canonical stdlib lane exposed the defect as:

`plugin was built with a different version of package runtime`

The v12 Go module selects Go 1.26.5, while the temporary generated extern
module declares only `go 1.22` and therefore used the ambient Go 1.26.4
toolchain. Go plugins require exact package/toolchain identity. The extern
builder now removes inherited `GOTOOLCHAIN` entries and supplies
`runtime.Version()` when it is a valid released Go version. Development
toolchain strings preserve the caller environment.

This is a shared language/runtime build rule. It contains no benchmark,
application, container, nominal type, or primitive-specific branch. No
compiler lowering, generated runtime, bytecode VM, canonical stdlib,
language, dependency, benchmark, reference application, or WASM source
changed.

Also retain one evidence-contract repair. The sustained workload-depth test
now derives its broad Go-duration ceiling from the current report and proves
that duration gating uses the Go reference by mutating a temporary generic
row. It no longer pins pre-captured-callable scorecard values.

## Canonical extern verification

- Focused environment-selection unit tests pass for released and development
  Go versions.
- Focused generated-extern integration tests pass.
- The complete canonical stdlib suite passes: tree-walker in 21 seconds and
  bytecode in 16 seconds, including filesystem, temporary-file, I/O, OS, and
  Go extern coverage.
- `../able-stdlib` did not change.

## Affected performance refresh

Changing shared interpreter source correctly invalidated the checked
performance closures. Exact reach review found that the selected Base64 and
JSON bytecode applications exercise `extern go`; strict compiled applications
remain fallback-free and interpreter-free.

Five successful, verifier-backed runs were retained for every Able and
reference row:

| Application | Current bytecode | Previous bytecode | Change | Python | Ruby |
| --- | ---: | ---: | ---: | ---: | ---: |
| Base64 | 2.9760 s | 3.9740 s | -25.11% | 4.0818 s | 2.7211 s |
| JSON | 0.8800 s | 1.2080 s | -27.15% | 2.6764 s | 1.7105 s |

Base64 beats Python and is 1.09x Ruby. JSON beats both references. The fresh
bytecode measurements used an explicitly cold isolated extern cache, so the
result does not depend on a stale plugin.

The refreshed current scorecard contains 132 selected rows, 66 compiled and
66 bytecode, with five successful Able/reference samples per row and 34
retained measurement sources. The combined frontier has ten guards, 122
misses, zero actionable groups, and 277.200421 seconds of aggregate target
excess, down 1.168842 seconds from 278.369263. All 23 checked performance
closures are current with zero invalidations, and the downstream architecture
evidence chain is green.

## Release verification

- `./run_all_tests.sh --compiler --compiled-cli` passes:
  - compiler bridge, all core batches, and all explicit parity outliers;
  - 128/128 fallback partitions;
  - 128/128 compiled-execution partitions;
  - 128/128 strict-dispatch partitions;
  - 128/128 interface-lookup partitions;
  - 128/128 boundary-marker partitions; and
  - generated-Go CLI integration in 1303.019 seconds.
- The final ordinary `./run_all_tests.sh` passes every preflight and
  noncompiler package, all 34 bounded compiler batches, and the bytecode
  fixture corpus in 89.459 seconds.
- `go vet ./...`, `go build ./...`, and runner shell-syntax checks pass.
- The complete architecture-budget check passes, including the 23-closure
  ledger and all downstream feasibility/ADR contracts.
- Maintained source files changed in this tranche remain below 1,000 lines.

Five release core batches exceeded one minute only as aggregates. An exact
125-test JSON replay found a 24.32-second maximum:
`TestCompilerExperimentalExecutionContextFixtureParity`. The two long default
compiler aggregates retain their exact audits at 26.050 and 11.440 seconds
maximum per named test. The 640 generated-program audit partitions remained
below 34 seconds each. No individual-test timing repair is required.

An initial diagnostic default invocation incorrectly applied a one-minute
timeout to an aggregate package. The named fixture implicated at interruption
passed independently in 0.830 seconds in the tree-walker and 0.865 seconds in
bytecode. The intended bounded runner subsequently passed in full.

## Evidence identities

- extern builder:
  `8f388d679a079093a6e48ec6f0d1938ba8a12d475ed9565f6832b952b2e3b276`
- current scorecard:
  `3652695dc7b1576ed4729ef30a7688b171114cda9b4ce269132fd868b37849f3`
- combined frontier:
  `d819bb32c402d690b9e89926b5bb8a0fddd6f18d9b966f8c85b3d9f73baa0a76`
- closure ledger:
  `e91bc0ea504ec7e79615b01265cba0359da1378790f596bde33af9345ba35c62`

## Temporary-file cleanup

Before measurement, five inactive Able/Go caches were checked for running
processes and open handles, then removed:

| Path | Removed |
| --- | ---: |
| `/var/tmp/able-v12-extern-go` | 85,776 KiB |
| `/var/tmp/able-parser-gocache` | 72,912 KiB |
| `/var/tmp/able-go-build-cache` | 8 KiB |
| `/var/tmp/able-compiled-test-cache` | 1,739,488 KiB |
| `/var/tmp/able-go-cache` | 28,550,168 KiB |
| **Total** | **30,448,352 KiB** |

All large work for this tranche used the exact disk-backed workspace
`/var/tmp/able-v12-release-readiness-20260730.5Ij6Mv`. Its final 28,929,240
KiB of isolated Go cache, generated applications, timing JSON, and temporary
build state was removed after this record was written. No broad `/tmp`,
repository, caller-owned, or unrelated cache cleanup was performed.

## Next

Keep production performance mutation paused and refresh the non-mutating v12
release-consolidation inventory.

Why: the current 132-row execution surface, 640-partition compiler matrix,
canonical stdlib, and 23 checked closures are green, while the frontier admits
no general owner shared by three unlike applications.

What it entails: capture a deterministic read-only inventory of changed and
untracked v12 paths; map them to dated records and dependency-ordered review
boundaries; identify unmatched or generated-local paths; and update the
consolidation record without staging, committing, deleting, resetting,
rewriting history, or touching deferred WASM work.

Why it matters: this turns the verified native-carrier and interpreter-free
compiled state into an auditable release candidate while protecting the
extremely dirty worktree. Production performance work should resume only when
a new application, correctness repair, source change, or observer result
invalidates a checked closure and exposes one exact material owner across
three unlike programs.
