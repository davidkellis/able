# Compiled Public-Pipeline Pair: Profile Decision

This follow-up profiles the two independently published public-pipeline
applications, Document Audit and Lexical Rollup, against MatrixMultiply as an
unrelated numeric control. It is an attribution tranche, not a timing refresh.

## Method

Each canonical Able source was built once with the canonical stdlib pinned at
`/home/david/sync/projects/able-stdlib/src`. Generated binaries ran with
`GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and CPU affinity `0`.
`ABLE_GO_CPU_PROFILE` and `ABLE_GO_ALLOC_PROFILE` used the generic generated
binary hook. The two short application profiles were merged across independent
normal-process launches rather than changing their program semantics:

| Workload | CPU samples | Input |
| --- | ---: | --- |
| Document Audit | 60 launches, 330 ms sampled | 131,072-byte corpus |
| Lexical Rollup | 10 launches, 370 ms sampled | first 16,384 records of the checked-in 1,743,363-byte word list |
| MatrixMultiply | 1 launch, 1.02 s sampled | canonical 1000×1000 numeric kernel |

The retained merged CPU profiles are
`20260710_document_audit_compiled_pair.cpu.pprof`,
`20260710_lexical_rollup_compiled_pair.cpu.pprof`, and
`20260710_matrixmultiply_compiled_pair.cpu.pprof` under
`v12/interpreters/go/.profiles/`.

## Evidence

Document Audit is led by GC scan work (`runtime.tryDeferToSpanScan`, 27.3%),
generated-package registration/JSON node decoding, and String search. Lexical
Rollup is led by the same GC scan leaf (27.0%), but its callers are instead
`fs.read_lines`/`strings.Split`, the generated iterator filter, String
containment, and bridge environment swaps. The shared leaf is a Go runtime GC
implementation detail, not a shared Able primitive or semantic-boundary
operation.

The MatrixMultiply control has none of this shape: 98.0% of sampled CPU is its
already-direct `main.__able_compiled_fn_matmul` f64 kernel. It rules out
treating the pipeline process-startup/GC frame as a general compiler
optimization.

The original allocation profiles could not refine this decision. They were
cumulative from process start and dominated by pre-main interpreter
initialization (`initBytecodeSmallIntBoxCache` plus interpreter init). The
later exact phase-boundary allocation refresh is recorded in
`2026-07-10-compiled-phase-profiles.md`: it attributes bootstrap allocation to
generic interpreter construction and AST JSON decoding, while explicitly
removing the profile collector's own allocation paths.

## Decision

No compiler, VM, or stdlib change is authorized. Removing or reordering a
named collection path would be benchmark-specific; changing the Go GC/runtime
would not be an Able lowering optimization. The only common frame is too high
level and has different concrete callers. The existing direct MatrixMultiply
kernel remains an unrelated numeric guard.

The phase-profile follow-up is complete in
`2026-07-10-compiled-phase-profiles.md`. It separates bootstrap/package
registration from user `main` across both pipeline applications and i-before-e.
The repeated bootstrap descendant is generated AST JSON decoding, not the
previously ambiguous Go GC leaf. A full direct Go-constructor renderer was
then rejected before runtime A/B: its canonical-stdlib generated source spent
more than six CPU minutes in Go parsing without producing an output build. The
follow-up compact tagged codec also failed its broad process-wall guard:
Document Audit and Lexical Rollup were neutral while i-before-e regressed
2.49% across paired 50-launch runs. Both representations are reverted. The
next eligible evidence is phase-bounded allocation/registration attribution,
not another decoder representation or per-definition source expansion.

The allocation-counter follow-up is now complete in
`2026-07-10-compiled-phase-profiles.md`: bootstrap consistently allocates
about 4.1–4.3 MB and 12.1k–16.1k objects per launch, but counters cannot
attribute those bytes to one safe registry/map descendant. The next profiling
step is complete too: exact phase-differenced profiles identify eager generic
method-cache backing maps as the only new candidate. Its paired evaluation is
also complete and rejected: tiny/mixed process-start movement did not offset a
3.9% regression on the persistent member-call guard. The capacity layout is
restored; the next evidence should come from the broader bytecode VM workload
profiles rather than another startup-cache tweak.
