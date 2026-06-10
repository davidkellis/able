# Runtime, tree-walker, and bytecode release-boundary review

Date: 2026-07-27

## Decision

The current runtime carrier, tree-walker, bytecode VM, runtime-support,
execution-fixture, and wrapper delta is completely classified and passes its
focused and repository-wide release gates. Treat the 658 reviewed paths as an
audited final-state boundary for the next dependency-ordered review.

The review found one real scheduler correctness failure, so this boundary was
not purely non-mutating. Retain the general `SerialExecutor.Flush`
dequeue-to-active handoff correction described in
`2026-07-27-serial-executor-flush-handoff-retained.md`. No performance
optimization was advanced: the correction changes scheduler observation, not
a hot carrier, materialization owner, or compiled/interpreted boundary.

Do not stage or commit this boundary yet:

- eight changed `*_wasm.go` shims remain `hold-deferred-wasm`;
- shared runtime definitions have compiler/AOT consumers that must be checked
  in the next dependency boundary; and
- no maintainer authorization to stage, commit, or rewrite history was given.

No standard-library source, language contract, dependency, benchmark,
reference implementation, fixture expectation, compiler lowering, or WASM
path changed. Nothing was staged, committed, reset, reverted, or rewritten.

## Deterministic inventory

The reviewed boundary is the union of current tracked modifications and
visible untracked files under:

- `v12/interpreters/go/pkg/interpreter/**`;
- `v12/interpreters/go/pkg/runtime/**`;
- `v12/interpreters/go/pkg/driver/**`;
- `v12/interpreters/go/pkg/profilehook/**`;
- `v12/interpreters/go/pkg/stdlibpath/**`;
- `v12/fixtures/exec/**`; and
- `v12/abletw` and `v12/ablebc`.

The deterministic manifest contains one header plus 666 sorted path rows:

| State | Paths |
| --- | ---: |
| Tracked modified | 277 |
| Untracked | 389 |
| **Total** | **666** |

| Family | Paths |
| --- | ---: |
| Bytecode VM | 181 |
| Deferred WASM shim | 8 |
| Engine guard or fixture | 208 |
| Execution fixture | 78 |
| Execution wrapper | 2 |
| Runtime carrier | 57 |
| Runtime support | 13 |
| Shared engine | 65 |
| Tree-walker | 54 |
| **Total** | **666** |

Six hundred fifty-eight paths are reviewed source, tests, fixtures, support,
or wrappers. Eight are held deferred WASM shims. There are zero unclassified
paths.

The manifest is
`2026-07-27-runtime-treewalker-bytecode-release-boundary-manifest.tsv`, has
667 lines and 116,528 bytes, and has SHA-256
`2b8b9a54db9487f54ca7d58a78a4c7ca4c36b2f338478a6df98fe4eb090404e0`.
Each row records state, family, disposition, line count, byte count, file
SHA-256, and path.

The earlier release-consolidation inventory projected 664 paths. The final
boundary has two additional untracked files created by the retained Go
source-hygiene split: the interpreter Array member extraction and the
slot-constant overflow test extraction. The current 666-path manifest is
authoritative.

The 277 tracked paths contain 34,315 added and 7,297 removed lines relative to
Git. That is the accumulated v12 final state, not a claim that this review
authored the historical delta.

## Runtime and engine ownership

### Carrier contract

The reviewed final state agrees with
`v12/design/canonical-runtime-value-architecture.md`:

- `runtime.Value` is the stable dynamic ABI used by the interpreters, host
  boundaries, services, and irreducibly dynamic values;
- `RawValue` is bytecode-private and keeps primitive register, slot, stack,
  call, and return values unboxed until a semantic escape requires a
  `runtime.Value`;
- compiled static primitives and static Arrays are expected to use native Go
  carriers; and
- non-primitive nominal values use the shared semantic encoding rather than
  structure-specific compiler lowering.

The bytecode paths materialize at environment, generic, dynamic, host,
nominal, public-result, error, and suspension boundaries. Searches found no
production branch keyed to an external benchmark application. Existing
canonical stdlib numeric and Array/String VM handling is the documented
language/kernel boundary and is not precedent for named-container compiler
rules.

### Arrays and ownership

`pkg/runtime` owns the canonical Array store and primitive-capable handles.
The tree-walker and bytecode VM share the live lease/last-owner lifetime
contract documented in `v12/design/array-handle-lifetime.md`.

The retained bytecode frame observer remains diagnostic only; no production
frame-local release policy was added. The closed Array, primitive-element,
raw-slot, and materialization censuses remain authoritative.

### Tree-walker, bytecode, and shared engine

The tree-walker remains the direct reference evaluator. The active bytecode VM
uses the in-place VM contract in `v12/design/bytecode-vm-v2.md`, including
typed primitive lanes, sound materialization, inline program frames, shared
nominal/interface resolution, and explicit fallback behavior.

The July 26 owner closures found no exact non-closed production owner shared
by three unlike applications. This boundary found no invalidation of those
profiles, so no new CPU/allocation experiment or performance code was
retained.

`compiled_thunk.go` and related registration surfaces are reviewed as runtime
definitions. Their generated compiler consumers are intentionally verified
again in the next compiler/AOT boundary.

### Scheduler correction

Race-enabled focused verification exposed a window in which the serial worker
had removed a runnable task from the queue but had not published it as active.
`Flush()` could observe an empty queue and return before the resumed task ran.

The retained correction accounts only for that handoff under the executor
lock. It preserves the existing rule that Flush may return when all remaining
work is blocked. A deterministic regression guard, repeated race checks, the
blocked mutex fixture, the complete interpreter package, both execution
modes, and both canonical stdlib modes pass.

### Runtime support and wrappers

The driver, profile hook, and stdlib-path packages provide package discovery,
source origins/re-exports, opt-in phase observation, and canonical stdlib-root
resolution. Diagnostic environment controls remain opt-in and inert during
ordinary execution.

`v12/abletw` and `v12/ablebc` build mode-specific cached binaries and execute
them from the caller's working directory. From a clean disk-backed directory,
both wrappers produced the same five-line result for
`02_lexical_comments_identifiers`.

### Deferred WASM

These eight current changed shims remain held:

- `bytecode_i32_frame_proof_wasm.go`;
- `bytecode_scalar_proof_wasm.go`;
- `bytecode_type_proofs_wasm.go`;
- `extern_host_wasm.go`;
- `interpreter_typecheck_wasm.go`;
- `interpreter_typechecker_state_wasm.go`;
- `runtime_diagnostics_wasm.go`; and
- `static_receiver_type_hints_wasm.go`.

No WASM implementation, design, verification, or performance work was
performed.

## Verification

Focused verification passed:

- `gofmt -l` reported no changed or untracked Go path in the boundary;
- every reviewed code file remains below 1,000 lines;
- `go vet ./pkg/runtime ./pkg/driver ./pkg/profilehook ./pkg/stdlibpath
  ./pkg/interpreter`;
- focused runtime/support package tests;
- the complete `TestBytecode` group;
- focused shared-engine and tree-walker groups;
- the opt-in `able_bytecode_box_reuse` build-tag lane;
- runtime Array race checks;
- the scheduler/Future/await/channel/async-bytecode race lane;
- the deterministic Flush handoff guard 50 times;
- the formerly stalled mutex-contention serial fixture 10 times;
- the complete short interpreter package in 75.624 seconds;
- both execution wrappers with identical output;
- wrapper Bash syntax checks; and
- `git diff --check`.

The required complete `./run_all_tests.sh` handoff passed in 10:20.17 at
4,742,496 KB peak RSS. All non-compiler packages, all 33 bounded compiler
batches, and the final 83.750-second bytecode fixture pass were green. The two
long aggregate compiler batches completed in 111.448 and 64.351 seconds; the
existing per-test timing record remains authoritative that no individual test
exceeds one minute.

The final canonical `./run_stdlib_tests.sh` handoff passed in 37.13 seconds at
870,304 KB peak RSS. Tree-walker reported 16 seconds and bytecode reported 15
seconds. No canonical stdlib change was required.

All substantial builds used disk-backed `/var/tmp` state. Guarded cleanup
removed 332 MiB of project-local wrapper cache, 140 KiB of Python cache, two
empty generated directories, 288 KiB of tranche inventory scratch files, and
the empty tranche working directories. The cleanup dry run now reports no
generated project candidates. Reusable disk-backed Go caches were retained.

## Recommendation

Next perform the third dependency-ordered release-consolidation review:
compiler/AOT lowering, the semantic ABI, compiler-facing CLI paths, and their
generated-code guards.

Why: the language contract and both reference execution engines are now
audited. The compiler consumes those boundaries, and it is the component that
must prove the user's immediate goal: ordinary static Able values stay in
equivalent native Go carriers and fallback-free applications never cross into
the interpreter.

What it entails: deterministically inventory current compiler, semantic-ABI,
and compiler CLI paths; map primitive and static-Array carriers, shared
nominal encoding, dynamic escape points, caches, generated runtime helpers,
and strict package cuts to retained evidence; reproduce generated snippets;
verify representative `--no-fallbacks` dependency graphs omit
`pkg/interpreter`; run focused compiler/semantic-ABI/CLI guards; and emit a
manifest and decision record without staging, committing, resetting, or
rewriting history.

Why it is important: this establishes whether the final generated Go really
preserves the representation strategy needed to approach native Go
performance. It also isolates any true boxing or compiler/interpreter-boundary
regression before new profiling is considered. Keep production performance
mutation paused unless the audit discovers a correctness invalidation or one
new exact owner material in three unlike applications. Do not begin WASM work.
