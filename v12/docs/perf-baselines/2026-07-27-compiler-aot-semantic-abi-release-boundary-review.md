# Compiler, AOT, and semantic-ABI release-boundary review

Date: 2026-07-27

## Decision

The current compiler lowering, generated runtime, compiler bridge, semantic
ABI, compiler-facing CLI, embedded-kernel, and analysis-tool delta is
completely classified and passes its focused and repository-wide release
gates. Treat all 340 paths as one reviewed final-state boundary.

The review found one real generated-runtime correctness failure. Retain the
general compiled serial-executor handoff correction described in
`2026-07-27-compiled-serial-executor-flush-handoff-retained.md`.

No performance candidate advanced. The correction aligns generated scheduler
semantics with the reference executor; it does not change a hot native
carrier, introduce boxing, add a compiler/interpreter crossing, or justify
reopening closed CPU/allocation profiles.

Do not stage or commit this boundary yet. The worktree remains a
multi-boundary long-running delta, shared spec/design/evidence paths still
need final consolidation review, deferred WASM remains held, and no
maintainer authorized a Git history operation.

## Deterministic inventory

The boundary is the union of current tracked modifications and visible
untracked files under:

- `v12/interpreters/go/pkg/compiler/**`;
- `v12/interpreters/go/internal/semanticabi/**`;
- `v12/interpreters/go/cmd/able/**`;
- `v12/interpreters/go/cmd/ablec/**`;
- `v12/interpreters/go/cmd/compiled-control-census/**`; and
- `v12/interpreters/go/go.mod`.

`cmd/ablewasm/main_js_wasm.go` is excluded and remains in the deferred WASM
boundary.

The manifest contains one header plus 340 sorted path rows:

| State | Paths |
| --- | ---: |
| Tracked modified | 174 |
| Untracked | 166 |
| **Total** | **340** |

| Family | Paths |
| --- | ---: |
| AOT CLI | 2 |
| AOT CLI guard | 1 |
| Compiled CLI | 17 |
| Compiled CLI guard | 10 |
| Compiler analysis tool | 10 |
| Compiler bridge | 22 |
| Compiler guard | 98 |
| Compiler lowering/runtime generation | 142 |
| Embedded kernel | 1 |
| Module metadata | 1 |
| Semantic ABI | 32 |
| Semantic-ABI tool | 4 |
| **Total** | **340** |

All 340 paths are reviewed and none is unclassified. The manifest is
`2026-07-27-compiler-aot-semantic-abi-release-boundary-manifest.tsv`, has 341
lines and 59,239 bytes, and has SHA-256
`9072a4a4762cf9736979f69612e168df63bd2b064c469b49751e0d5579cd84ce`.
Each row records state, family, disposition, line count, byte count, file
SHA-256, and path.

The release-consolidation inventory projected 335 paths. The final boundary
adds three untracked source-hygiene extraction files and the two files from
the retained compiled handoff correction: one regression guard and one
declaration extraction. The current 340-path manifest is authoritative.

The 174 tracked paths contain 6,242 added and 3,365 removed lines relative to
Git. This is the accumulated final state of the long-running v12 effort, not
a claim that this review authored the historical delta.

## Native lowering and boundary ownership

### Primitive and Array carriers

Static Able booleans, signed and unsigned integer widths, i128/u128, f32/f64,
characters, and strings retain native Go function signatures and carriers.
Static `Array T` values retain compiler-owned typed Go carriers through
construction, indexing, assignment, iteration, calls, returns, and supported
interface boundaries.

Verifier-backed source guards reject `runtime.Value`, generic bridge calls,
and interpreter fallbacks in representative static primitive and Array
closures. Direct canonical kernel helpers stay direct rather than chaining
through dynamic extern lookup.

### Shared nominal translation

Structs, unions, interfaces, generic nominals, stdlib containers, and
user-defined containers continue through shared nominal specialization and
semantic encoding. Searches found no production compiler branch keyed to an
external benchmark application or to TreeMap, TreeSet, LinkedList, Heap, or
another non-primitive named container.

Map-literal lowering to the canonical HashMap representation and generated
HashMap kernel helpers are language-syntax/kernel ABI boundaries, not
container-specific performance rules. Ordered-container tests demonstrate
the shared nominal pipeline rather than production named-container branches.

### Dynamic and interpreter boundaries

`runtime.Value` remains present where the language or host ABI is explicitly
dynamic: host services, reflection/metaprogramming, irreducibly polymorphic
values, bootstrap metadata, errors, and runtime services.

The `able` CLI and semantic-ABI binding report intentionally retain
interpreter imports because they support tree-walker/bytecode execution,
dynamic/bootstrap evaluation, or cross-engine reporting. Generated compiler
imports and bootstrap roots are conditional. Strict fallback-free
applications omit them.

### Strict package cut

Fib, Sudoku Masks, and Tapelang Alphabet compiled from current source with
`--no-fallbacks`. Every final `go list -deps` graph contains 96 packages and
zero `able/interpreter-go/pkg/interpreter` package. Top-level generated Go
sources contain zero interpreter package references.

Current primary generated-source hashes are:

| Application | `compiled.go` SHA-256 |
| --- | --- |
| Fib | `ff189e394c899d7ae776261f0006bfb2ca33f7043606d46b68dfcc2f233b3c61` |
| Sudoku Masks | `dedab220ab8fb130dbdc7ac726887e67ebb222bc0bf3d1918ead423dad930ebc` |
| Tapelang Alphabet | `957ab35d0d75df2faf37dee64bf81b2606eb15c5b49c7378157476e7931a4fd5` |

A second Fib emission reproduced all 14 top-level `compiled*.go` files
byte-for-byte. Three independent processes per application passed their
public Ruby verifiers:

| Application | Verified | Smoke wall mean |
| --- | ---: | ---: |
| Fib | 3/3 | 3.4200 s |
| Sudoku Masks | 3/3 | 1.5100 s |
| Tapelang Alphabet | 3/3 | 3.6033 s |

These are correctness smoke measurements, not balanced A/B/Go performance
evidence.

### Generated scheduler correction

The generated serial executor had the same dequeue-to-active Flush race
already corrected in the reference interpreter. It now counts only the
locked invisible handoff and clears that count as the task becomes active.
Blocked-work Flush semantics remain unchanged.

The emitted regression test passed under the race detector for 20
repetitions. The correction is a general compiler-owned runtime rule and is
not tied to a benchmark, container, or primitive representation.

### Semantic ABI, caches, and tools

The semantic ABI continues to model Go bindings, flow, shadowing, and heap
facts independently of application names. Compiler import/type normalization,
specialization, and analysis caches remain invalidated by binding growth and
normalization lifecycle rules. CLI compiled-test caching remains opt-in,
content-addressed, file-locked, inspectable, and explicitly bounded.

`go.mod` only makes the already-used `golang.org/x/sys v0.15.0` dependency
direct for the explicit OS file-lock implementation. `go mod verify` passes,
and the resolved module graph contains 64 modules.

## Verification

Focused verification passed:

- the generated serial-Flush source guard;
- the generated serial-Flush race test for 20 repetitions;
- generated goroutine-Flush notifier and blocked-task parity;
- static primitive, static Array, shared nominal, direct-kernel,
  interpreter-package-cut, dynamic-boundary, and unlike benchmark-source
  guards;
- semantic-ABI, bridge, `ablec`, compiled-control-census, and `able` package
  tests;
- `go vet` across compiler, bridge, semantic ABI, and compiler-facing CLIs;
- `go mod verify` and module-graph resolution;
- all changed or untracked boundary Go files are `gofmt` clean;
- the largest changed boundary Go file is 998 lines;
- `git diff --check`; and
- three current strict application builds, dependency graphs, reproducible
  generated source, and nine public-verifier processes.

The explicit disk-backed compiled-CLI lane passed all 42 cases:

- command: `./v12/run_all_tests.sh --compiled-cli`;
- wall time: 19:31.59;
- peak RSS: 3,399,964 KB; and
- current compiler identity was rebuilt rather than reusing the prior
  generation.

The required complete `./run_all_tests.sh` handoff passed:

- wall time: 15:14.81;
- peak RSS: 4,545,104 KB;
- all 33 compiler batches passed; and
- the final bytecode fixture corpus passed in 97.109 seconds.

The canonical `./run_stdlib_tests.sh` handoff passed:

- wall time: 50.25 seconds;
- peak RSS: 856,468 KB;
- tree-walker reported 20 seconds; and
- bytecode reported 19 seconds.

No canonical stdlib source change was required.

## Cache and cleanup discipline

The compiled lane intentionally invalidated its prior compiler-identity
generation. A reviewed dry run selected exactly 42 older valid entries.
Applying the 1536 MiB bound removed 42 entries and 1,580,708,024 bytes. The
retained current working set contains 42 valid entries and 1,580,796,720
bytes, with zero corrupt, staging, obsolete-schema, or unknown entries.

All substantial scratch and generated builds used disk-backed `/var/tmp`;
the RAM-backed `/tmp` was not used for this tranche's heavy artifacts. Final
guarded cleanup removed the exact 405 MiB disk-backed audit workspace and a
37 MiB stale RAM-backed Able extern cache. The current bounded compiled-test
cache and reusable disk-backed Go build cache remain.

## Scope

No canonical stdlib, external reference, benchmark, fixture expectation,
language contract, interpreter, bytecode VM, or WASM source changed. No Git
history operation was performed. The generated serial-executor correction is
the only production behavior authored during this review.

## Recommendation

Next perform the fourth dependency-ordered release-consolidation review for
canonical documentation, examples, benchmarks, performance evidence, and
shared held spec/design paths.

Why: the language contract implementation, reference engines, and
compiler/AOT/semantic-ABI boundaries are now independently audited. Remaining
visible changes are cross-boundary evidence and documentation that must be
classified before a maintainer can safely authorize staging.

What it entails: deterministically inventory remaining changed/untracked
documentation, examples, benchmark contracts, generated evidence, and shared
held paths; verify links and hashes; separate authoritative retained records
from disposable generated state; run the appropriate lightweight integrity
gates; and emit the final manifest and decision without touching deferred
WASM or Git history.

Why it is important: this completes the consolidation chain while preserving
the immediate compiled-performance architecture—native Go carriers for static
values and no interpreter package in strict applications. Performance
mutation remains paused unless a new broad application, correctness
invalidation, changed generated hot path, or report-only observation reveals
one exact general owner in at least three unlike programs.
