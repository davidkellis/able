# Mutex Await Journal application coverage and lambda-lowering gate — 2026-07-14

## Decision

Add `mutex_await_journal` as the 30th verifier-backed external application,
with its own `mutex-await-journal` suite and membership in `concurrency` and
the broad `coverage` inventory. It is deliberately unlike Mutex Ledger:
four workers repeatedly `await` the public `Mutex.await_lock` Awaitable and
release the acquired lock in the callback's `ensure` block. The work still
updates a shared nominal `Journal`, but it exercises readiness, registration,
wake, commit, callback, and re-arm behavior rather than `with_lock`'s direct
blocking path.

The active Able source is
`v12/examples/benchmarks/mutex_await_journal/mutex_await_journal.able`; the
portable sibling programs, Docker lanes, exact Ruby verifier, and README are
under `../benchmarks/mutex-await-journal`. All lanes print:

```text
4:512:2048:1056301196:108896:1056301196
```

No canonical `able-stdlib` change was needed. The existing public `Mutex`
API already specifies the callback ownership and `ensure`-based unlock
contract this application uses.

## General runtime repairs

The application exposed generic async defects, all repaired before timing:

1. A mutex Awaitable could observe the mutex locked, miss the unlock, and only
   then register its waker. Registration now checks and inserts while holding
   the mutex state lock; the generated compiled runtime mirrors the same
   recheck/cancel/wake behavior.
2. An await waker was treated as permanently one-shot. A wake means only that
   readiness must be checked again: another task can win the mutex before
   commit. Await state now clears its one-shot registrations after every wake
   and registers a reusable waker for the next wait cycle. Its blocked marker
   is atomic in both interpreters and the generated runtime.
3. Bytecode futures could jump from task-local execution into a reusable
   closure environment and lose the task payload, falling back to one shared
   VM call-frame stack. Direct bytecode calls now wrap the closure environment
   with the current payload when needed. Tree-walker helper calls do the same,
   so `await` remains valid inside an ordinary helper invoked by `spawn`.
4. Nested `ensure` inside a compiler-lowered lambda inherited a lambda-only
   `break` target and could emit invalid Go. `ensure` now establishes its own
   ordinary control boundary.

The focused public-Mutex await regression passes ten consecutive runs in both
tree-walker and bytecode modes; the compiled counterpart passes three
consecutive runs. A generated `-race` binary for the full four-worker
application also completes without a race report.

## Shared profile gate and kept optimization

Bounded bytecode-runtime profiles used
`GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1`, goroutine execution, and five steady
state `main()` calls. Both Mutex applications repeatedly lowered the same
runtime-created callback lambda. The baseline samples attributed about 29.7%
of Mutex Ledger CPU and 12.1% of Await Journal CPU to generic lambda lowering.

The kept response is a bounded (128-entry) cache of immutable lambda bytecode
programs, keyed by lambda AST identity plus lexical binding-shape state and
revision. A binding-shape change invalidates the key; ordinary local value
updates do not. Callable return metadata is populated before cache publication
and never rewritten when a later FunctionValue attaches the shared program.
This is ordinary closure lowering for every Able program, not a Mutex,
container, helper-name, or source-shape fast path.

| Bounded bytecode-runtime workload | Baseline | Cached | Change |
| --- | ---: | ---: | ---: |
| Mutex Ledger | 574,805,995 ns/op | 547,955,552 ns/op | 4.7% faster |
| Mutex Await Journal | 116,533,379 ns/op | 112,517,690 ns/op | 3.4% faster |
| String split/join guard | 1,043,194,732 ns/op | 994,067,976 ns/op | 4.7% faster |
| Linked-list iterator collect guard | 266,181,801 ns/op | 259,224,037 ns/op | 2.6% faster |
| Numeric Array map guard | 68,608,756 ns/op | 70,016,520 ns/op | 2.1% slower, within timing noise; allocations fell 329 to 95/op |

The candidate therefore clears the two-application gate and the established
unrelated guards. It stays in the bytecode interpreter; it does not change
compiler/AOT lowering rules or canonical stdlib source.

## Fresh verifier-backed comparison

Three fresh reference processes and three Able processes per mode all passed
the exact verifier. These unpinned measurements are product-status evidence,
not a replacement for the stable generality scorecard.

| Application / mode | Able real (s) | Go real (s) | Ruby real (s) | Python real (s) |
| --- | ---: | ---: | ---: | ---: |
| Mutex Ledger compiled | 0.7567 | 0.0042 | 0.0775 | 0.0394 |
| Mutex Ledger bytecode | 0.4067 | 0.0042 | 0.0775 | 0.0394 |
| Mutex Ledger tree-walker | 0.5967 | 0.0042 | 0.0775 | 0.0394 |
| Mutex Await Journal compiled | 0.7800 | 0.0038 | 0.0592 | 0.0287 |
| Mutex Await Journal bytecode | 0.1633 | 0.0038 | 0.0592 | 0.0287 |
| Mutex Await Journal tree-walker | 0.2200 | 0.0038 | 0.0592 | 0.0287 |

The applications remain material performance misses: bytecode is 5.69x
Python on Await Journal and 10.32x Python on Ledger; compiled is about 180x to
205x the matching Go lane. The coverage is retained precisely so later work
must improve ordinary programs rather than an isolated synthetic benchmark.

## Verification

- `TestPublicMutexAwaitLockSurvivesRepeatedGoroutineContentionInBothModes`
  passed ten times; its focused `-race` run passed.
- `TestCompilerPublicMutexAwaitLockRearmsAfterContention`,
  `TestCompilerEnsureInsideNativeLambdaBuildsAndRuns`, and
  `TestCompilerGoroutineMutexContentionCompletes` passed three times.
- `bench_bytecode_audit --suite mutex-await-journal` passed; the complete
  `coverage` audit passed with 30 applications, 130 functions, and 7,211
  instructions.
- `bench_compiled_boundary_audit --suite mutex-await-journal` completed with
  verified output and reported only its expected generic dynamic boundaries.

## Awaitable profile follow-up

The three-application Awaitable profile gate is complete and keeps no source
change. Mutex Await Journal and Await Channel Mux share only the
`completeAwait` dispatcher parent before diverging into mutex versus channel
commit implementations; Future Await Race has no material completion-path
sample. See `2026-07-14-awaitable-cross-app-profile-gate.md`. The next step is
a fresh verifier-backed compiled/bytecode scorecard, including the new async
coverage rows, before selecting another concrete cross-application leaf.
