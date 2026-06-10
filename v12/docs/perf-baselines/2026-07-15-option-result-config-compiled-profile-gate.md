# Option/Result Configuration compiled-profile gate — 2026-07-15

## Decision

Keep no compiler, generated-runtime, bytecode VM, canonical-stdlib, or
benchmark-source performance change. The three unlike verifier-backed
applications do not expose one new, material concrete compiled leaf. In
particular, do not turn the new generic-union support into an `Option`/`Result`
shortcut, retry the rejected execution-context ABI, or specialize Queue,
iterator, file, or Document Audit code.

## Method

The current generated binary for each application was built once with the
canonical external `able-stdlib`, then launched independently under the
generated launcher's CPU-only phase hook
`ABLE_GO_PHASE_CPU_PROFILE_DIR`. The hook writes distinct `bootstrap` and
registered-`main` CPU profiles without enabling exact allocation snapshots.
Per-application phase files were merged with `go tool pprof -proto`; profiles
from different binaries were never combined.

Every launch used `GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, a 45-second
cap, and `taskset` on one CPU. CPU 1, CPU 10, and CPU 11 respectively passed
the immediate three-sample quiet-host preflight just before the Option/Result,
Dependency Plan, and Document Audit launch sets. Every completed output passed
the public Ruby verifier, and each application's SHA-256 was stable.

| Application | Launches | Main CPU samples | Bootstrap CPU samples | Output SHA-256 |
| --- | ---: | ---: | ---: | --- |
| Option/Result Configuration | 50/50 | 6.78 s | 60 ms | `28e46b27a6dceeaa15968e9db7a6728f4a2b35f87a89ff7bf561db18cad53112` |
| Dependency Plan | 80/80 | 110 ms | 80 ms | `96dc74508d9b7a476bafdef453b11e11f2f70279c58ccaa5dcb6d85c529c4a38` |
| Document Audit | 100/100 | 240 ms | 220 ms | `0dad030a80c8a883cbb56fbcfebfd530d521075e15d5d91ba538bc93e66c0aab` |

Retained merged profiles:

- `v12/interpreters/go/.profiles/20260715_compiled_option_result_config_main_merged.cpu.pprof`
- `v12/interpreters/go/.profiles/20260715_compiled_option_result_config_bootstrap_merged.cpu.pprof`
- `v12/interpreters/go/.profiles/20260715_compiled_dependency_plan_main_merged.cpu.pprof`
- `v12/interpreters/go/.profiles/20260715_compiled_dependency_plan_bootstrap_merged.cpu.pprof`
- `v12/interpreters/go/.profiles/20260715_compiled_document_audit_main_merged.cpu.pprof`
- `v12/interpreters/go/.profiles/20260715_compiled_document_audit_bootstrap_merged.cpu.pprof`

The binaries, generated source, per-launch profiles, and captured output were
temporary measurement artifacts and were removed after the merged profiles
and this record were produced.

## Attribution

Option/Result Configuration has enough main samples to identify a real local
cost: `__able_static_generic_union_method_call` is 58.55% cumulative,
`__able_call_value_fast` is 45.28%, and `runtime.mallocgc` is 52.51%. The
static generic-union helper looks up an ordinary generated method and invokes
the native bound-method value. This is a generic language mechanism, but it
is material only in this application. The generated program body contains the
six `Option`/`Result` calls that exercise it; neither unlike control calls the
helper from its application body.

Dependency Plan's registered main is too short for a stable new optimization
selection. Its 110 ms sample set reaches Queue/Deque entry wrappers and
`bridge.SwapEnvIfNeeded` (45.45% cumulative). That bridge is the previously
tested fixed execution-context family: its broad default and package-linkage
experiments regressed independent N-body and K-Nucleotide guards. It is not
new evidence to retry the ABI, and Queue is a nominal stdlib type rather than
a language-level lowering boundary.

Document Audit remains below useful CPU-profile resolution even after 100
launches. `profilehook.StartMain` plus its GC descendants accounts for 62.5%
of the merged samples, while the small non-profiler remainder reaches the
expected `read_lines` and iterator generator/filter/next path. Its generated
application body directly composes file lines, `filter`, `map`, `collect`,
and `reduce`; it does not invoke the generic-union fast helper. The phase
hook's start cost is measurement overhead, not application work.

The bootstrap profiles are likewise dominated by CPU-profiler startup and GC.
Document Audit has a small package-registration/allocation residual, but it
does not recur as a material leaf in the other two applications. The already
completed static-metadata work removed the only demonstrated unreachable
startup decoder payload; this short-profile noise does not reopen it.

## Why no candidate is admitted

The one material source-level leaf is generic-union dispatch in Option/Result
Configuration. Dependency Plan instead reaches the rejected environment-swap
bridge through Queue/Deque wrappers, and Document Audit reaches file/iterator
work only at low sample resolution. Common runtime parents and profiler/GC
frames are not an implementation boundary. None is a new concrete descendant
that repeats across all three unlike verified programs.

This gate therefore makes no claim that the current compiled binaries meet the
Go target. The previous normal-process comparison still records the product
gap; the profile gate only says that this application trio does not identify a
safe, broadly applicable next compiler optimization.

## Verification

- All 230 profiled binary launches completed within the 45-second cap.
- All 230 outputs passed their canonical Ruby verifier and were stable per
  application.
- Each profile set was captured only after its own passing quiet-host
  preflight.
- `go tool pprof -proto` successfully merged the separate same-binary phase
  files listed above.

## Next recommendation

Do not take another unchanged compiled-binary profile or reopen generic-union
dispatch, Queue, iterator/file, bootstrap-decoder, or fixed-context work. The
next eligible performance tranche is a material cross-cutting semantic/compiler
change or a genuinely needed portable application that introduces a concrete
non-nominal operation in several unlike programs. Profile that operation across
at least three verifier-backed applications, then consider a compiler or
canonical-stdlib change only if its descendant—not merely a common runtime
parent—repeats and survives the full suite guard.
