# Feature/benchmark roadmap reconciliation — 2026-07-14

## Decision

Do not add a new cross-language performance application in this tranche. The
previous feature-led recommendation was too broad: the active v12 roadmap has
no unfinished parser, AOT, stdlib, or regex language boundary that can
honestly supply a new portable timing row. Inventing a busy loop for dynamic
imports, user-authored externs, or host callbacks would violate the benchmark
generality policy.

No VM, compiler, runtime, canonical-stdlib, fixture, or benchmark source
changed.

## Evidence

`spec/TODO_v12.md` records no outstanding parser, compiler-AOT, stdlib
externalization, or regex API gap. The remaining active runtime backlog is
WASM host callbacks plus source/module loading and scheduler/filesystem ABI
work; those are browser-host portability work, not a fair Go/Python/Ruby
runtime comparison.

The current coverage matrix has 32 portable external applications and 77
bounded local fixtures. The local-only dynamic-package and user-extern rows
are intentional: their observable module/host semantics do not have an honest
like-for-like foreign-runtime counterpart. The static catalog check confirms
each portable row has its Able target, suite/run directory, verifier, and
default Go/Python/Ruby source lane. The current full lowering audit passed:

| Check | Result |
| --- | --- |
| `just bench-catalog-check` | 32 portable applications, 77 local fixtures, 109 total programs |
| `v12/bench_bytecode_audit --suite corpus-full` | 109 programs, 410 lowered functions, 20,205 instructions |
| `just wasm-smoke` | tree-walker and bytecode both evaluated the AST input to `3` under Node |

The WASM result validates only the existing pre-parsed-AST execution boundary.
It deliberately does not claim source parsing, module loading, timers,
filesystem access, or browser extern callbacks.

## Consequence for performance selection

The current portable suite remains the authoritative performance screen. New
source changes still require a concrete leaf repeated across unlike existing
applications and verification against the broad corpus. A feature that cannot
form a fair cross-language application remains a semantic/portability fixture
guard only; it must not become a synthetic timing target.

## Next recommendation

The output and static source/module slices are complete. The Go/JS adapter
forwards host output, the portable JavaScript resolver receives source bytes
through an injected provider, and a real browser smoke parses the approved
Able source closure before running both interpreters. No filesystem/module-root
ABI was added.

The follow-up three-application compiled profile gate is also complete:
Reverse Complement, Monte Carlo Pi, and PiDigits expose allocation/slice,
checked-integer, and `math/big` leaves respectively. None repeats across the
unlike applications, so it selects no source change. Keep the WASM scope
bounded and do not manufacture another unchanged-scorecard profile pass. The
verifier-backed applications remain the only timing authority; reopen an
optimization only when a broadly applicable change yields a repeated concrete
leaf. See
`2026-07-14-compiled-byte-numeric-bigint-profile-gate.md`.

## Follow-on: Array slice coverage — 2026-07-14

The completed `Array.slice(start, end) -> !Array T` contract is the narrow
exception that can be represented honestly in every reference language. The
new `array-slice-window` application takes overlapping, independently copied
numeric batches and produces a verifier-backed checksum. Go uses `make` plus
`copy`, rather than its aliasing subslice expression, so all implementations
measure the same shallow-copy container semantics. Python and Ruby use their
ordinary copying range operations.

It is deliberately catalogued only in the broader `coverage` suite, which now
has 32 portable applications, not the
stable `generality` screen. One new API carrier may establish behavioral and
lowering coverage but cannot select an Array-specific VM/compiler change. Any
future optimization must still expose the same concrete leaf in at least three
unlike verifier-backed applications.

## Follow-on: Deployment plan coverage — 2026-07-14

`dependency-plan` is the 32nd portable `coverage` application. It turns the
existing local topological-sort fixture shape into a normal application model:
given a deterministic 1,024-service deployment graph, it derives indegrees,
maintains a FIFO of ready services, and produces a checksum across twelve plan
resolutions. Able tree-walker, bytecode, and compiled modes, plus Go, Python,
and Ruby references, all verify `1024:12595200`.

The app uses only general `Array` and `Queue` behavior. It is not added to the
stable `generality` timing screen and does not authorize graph-, queue-, or
deployment-specific lowering. The fresh no-timing corpus audit is 109 programs,
410 functions, and 20,205 instructions. Any later candidate must expose a
concrete leaf shared by this and at least two unlike verified applications,
after the quiet-host preflight passes.
