# Performance Competitiveness Vision

Date: 2026-05-27

## Purpose

This document is the handoff map for making Able fast enough to be credible in
the sibling `../benchmarks` suite.

The target is deliberately higher than "faster than the old interpreter":

- compiled Able should be competitive with equivalent hand-written Go when the
  program is statically representable;
- bytecode Able should be competitive with mainstream bytecode/interpreter
  runtimes on the same logic;
- both runtimes must implement the v12 spec, share the same AST contract, and
  preserve tree-walker parity for observable behavior.

This is not permission to add benchmark-specific shortcuts. The route to speed
is a better compiler, a better VM representation, and reusable stdlib/runtime
surfaces.

## Non-Negotiable Constraints

- `spec/full_spec_v12.md` is the semantic authority.
- The Go tree-walker remains the behavioral reference.
- The bytecode VM must stay in strict semantic parity with the tree-walker.
- The compiler must keep static paths native and explicit dynamic boundaries
  narrow.
- Only primitive Able types may receive primitive-specific compiler lowering.
- Non-primitive nominal types, including stdlib containers, must lower through
  shared nominal/carrier/dispatch machinery.
- Array and String may have native compiler/VM treatment because they are core
  language/kernel boundary types, not because a benchmark happens to use them.
- Every optimization needs a guardrail test and a benchmark/profiling reason.

## End State

### Compiled Able

Compiled Able should look like direct Go for static code:

- primitive values lower to Go scalars;
- arrays lower to compiler-owned native slice carriers;
- structs lower to native structs or pointers;
- unions/results/options/interfaces lower to generated native carriers;
- loops, branches, patterns, calls, and dispatch lower to direct Go control
  and direct calls when the target is statically resolved;
- `runtime.Value`, `any`, interpreter dispatch, runtime array stores, and
  dynamic call helpers appear only at explicit dynamic language or host ABI
  boundaries.

The compiler is "fast enough" only when source audits and external benchmarks
both agree. If generated Go for a hot static path still manipulates
`runtime.Value`, it is not done even if a small benchmark looks good.

### Bytecode Able

The bytecode interpreter should evolve into VM v2, not fork into a second
interpreter:

- keep the existing lowering pipeline, diagnostic AST nodes, call-frame stack,
  resume machinery, and fallback runtime helpers;
- add typed internal storage for primitives so hot local operations do not
  continually box into `runtime.Value`;
- add typed operand-stack/register lanes where boxed stack traffic dominates;
- quicken stable call/member/index sites after the first successful proof;
- add VM-native Array/String bytecodes for canonical kernel APIs;
- box back to the existing `runtime.Value` path at every dynamic/spec boundary
  that cannot prove the optimized representation is valid.

The bytecode VM is "fast enough" when its remaining overhead is mostly the
algorithm and unavoidable dynamic semantics, not repeated lookup, boxing,
allocation, and helper dispatch for statically obvious primitive/container
operations.

### Stdlib

The canonical external `../able-stdlib` must grow normal reusable APIs needed
to express the benchmark suite:

- encoding, digest, JSON, deterministic RNG, byte-buffer/string-builder,
  bigint, IO, and text-processing features;
- APIs should be useful library surfaces, not benchmark-only shims;
- host-backed implementation is acceptable where the language/runtime boundary
  makes that the right abstraction, but compiler lowering must not special-case
  a named stdlib container or algorithm.

## Current State

The current external scoreboard already shows that the compiled path can be
competitive when static lowering is clean:

- `fib`: compiled near Go; bytecode completes through a guarded recurrence
  kernel; tree-walker times out.
- `binarytrees`: compiled near or ahead of Go; bytecode and tree-walker time
  out.
- `matrixmultiply`: compiled near Go; bytecode is currently ahead of the Go
  reference after guarded mono-f64 row/storage kernels; tree-walker times out.
- `quicksort`: compiled now completes in Go range; bytecode still times out at
  full external scale, though the 1MB prefix and reduced in-tree hotloop have
  improved materially.
- `sudoku`: compiled ahead of Go; bytecode is slower than Go but no longer in
  the timeout class.
- `i_before_e`: compiled close to Go; bytecode remains substantially slower
  than Go but usable.
- `base64` and `json`: implemented through reusable stdlib APIs; JSON is
  currently ahead of the Go/Python/Ruby references in both compiled and
  bytecode modes.
- `monte_carlo_pi`: implemented through deterministic Park-Miller sampling;
  compiled is in the broad Go range and faster than Ruby/Python, while
  bytecode is still dominated by boxed numeric arithmetic and GC.

The important conclusion is that the compiled direction is broadly correct.
The highest-risk remaining work is:

- missing benchmark families and stdlib surfaces;
- keeping compiled static paths native as new library features arrive;
- bytecode VM v2 representation work, especially for quicksort and
  binarytrees-style timeout families.

## Measurement Rules

Use external benchmarks as the final guardrail.

Recommended loop for a tranche:

1. Refresh the current external or reduced baseline.
2. Capture CPU and allocation profiles for the exact baseline.
3. Choose one bounded mechanism from the profile.
4. Add focused semantic coverage before benchmarking.
5. Run the focused tests.
6. Run a repeated benchmark band.
7. Keep only if wall time and allocation evidence are defensible.
8. If kept, update `PLAN.md`, `LOG.md`, and performance docs.
9. If rejected, revert the experiment and document the negative result.

Reduced and prefix benchmarks are useful when the full external benchmark still
times out, but they are not the finish line.

## Workstreams

### 1. Scoreboard And Coverage

Goal: know what remains, keep it measured, and avoid optimizing blind.

Required work:

- keep `v12/docs/perf-baselines/external-scoreboard-current.*` current;
- add Able implementations for missing benchmark families as stdlib support
  lands; after the pidigits coverage tranche, the remaining candidate is
  `tapelang-alphabet` if its language/runtime needs are intentionally selected;
- keep canonical sources in `v12/examples/benchmarks`;
- treat `../benchmarks/*/able-v12-*` as harness packaging only;
- add per-benchmark generated-source audits once a compiled family is closed.

Success condition:

- every external benchmark has an Able implementation or an explicit tracked
  language/stdlib blocker;
- compiled, bytecode, and tree-walker rows are refreshed often enough that the
  plan is driven by current evidence.

### 2. Stdlib Benchmark Surface

Goal: make Able capable of expressing the remaining benchmark suite with normal
library APIs.

Landed APIs:

- `able.encoding.base64` encode/decode over strings and byte arrays;
- `able.crypto.md5` MD5 hex helpers;
- `able.json` with small-DOM parsing, typed numeric/object/array access, and a
  fast numeric field projection helper for the external JSON benchmark;
- host-backed `able.fs.read_text` so benchmark-scale file reads do not convert
  a host byte array back through Able string validation one element at a time;
- `able.random.Random` with deterministic Park-Miller `next_i32`,
  `next_i64`, and `next_f64` helpers;
- `able.numbers.bigint_native.BigIntRef`, a reusable host-backed mutable BigInt
  reference API over Go `math/big`, which unlocks the pidigits benchmark.

Near-term APIs:

- byte-buffer / string-builder APIs for large output construction without
  repeated string copies or UTF-8 validation;
- tighter `able.fs`, `able.io`, and `able.text.string` hot APIs for file
  reads, line iteration, splitting, parsing, searching, and replacement.

Success condition:

- missing benchmark families can be implemented naturally in Able;
- compiled generated code for those APIs does not route hot static paths
  through interpreter containers or dynamic dispatch.

### 3. Compiler Go-Competitive Path

Goal: keep compiled Able in Go range as the benchmark suite broadens.

Current priority:

- maintain the existing compiler-native architecture rather than adding more
  benchmark-local lowering;
- add generated-source audits for every closed benchmark family;
- when a new stdlib API lands, prove its compiled hot path uses native carriers;
- fix shared carrier/dispatch/control lowering when new benchmarks expose
  dynamic scaffolding in static code.

Likely near-term compiler work:

- native byte-buffer/string-builder carriers and operations;
- native JSON DOM or streaming-token carriers where statically typed;
- source audits and cleanup for the native BigInt boundary now used by
  `pidigits`;
- source audits that reject `runtime.Value`, `runtime.ArrayValue`,
  `ArrayStore*`, `any`, interpreter dispatch, and panic/recover flow in closed
  static benchmark paths;
- no-fallback launch checks for new benchmark binaries.

Do not do:

- do not add a compiler branch for `HashMap`, `LinkedList`, `TreeMap`, `Heap`,
  `JsonObject`, or any other named non-primitive type just because a benchmark
  uses it;
- do not hide interpreter calls behind helper names on hot static paths;
- do not accept a benchmark win if the generated Go shape is moving away from
  native carriers.

Success condition:

- every implemented static benchmark compiles to straightforward Go-shaped
  code and stays within the agreed comparison band against hand-written Go.

### 4. Bytecode VM v2

Goal: replace dynamic boxed traffic in hot bytecode paths with typed VM
representation and guarded quickening.

The next bytecode work should be architectural, not another series of tiny
helper shaves.

Primary VM-v2 pieces:

- typed layout metadata for slot-eligible bytecode programs;
- `i32` typed slots and operand/register cells;
- explicit materialization at dynamic/spec boundaries;
- typed inline call args and returns;
- bool branch lanes;
- `u8` and byte-array lanes for parsers and text scans;
- `f64` lanes for numeric loops where matrix work has already proven value;
- call/member/index quickening with version/shape guards;
- native Array/String bytecodes for canonical kernel APIs;
- resume/unwind coverage for typed frames before enabling typed lanes in
  yielding/async functions.

The first production target should be quicksort because full bytecode
quicksort still times out.

Quicksort-specific near-term path:

1. Start from a fresh 1MB prefix and full external profile.
2. Treat the kept tracked-array swap fast path as exhausted unless a fresh
   profile proves otherwise; the current reduced profile has moved on to
   compare, call setup, and slot arithmetic.
3. Design the first real raw `i32` typed frame/register slice, not another
   active-frame sidecar retrofit.
4. Carry `i`, `j`, parser `value`, and simple loop counters through raw typed
   slots without writing a non-pointer sentinel into `runtime.Value` on every
   update.
5. Box only at calls, returns, member/index dispatch, array writes, diagnostics,
   and other explicit dynamic/spec boundaries.
6. Keep v12 checked integer overflow and non-negative/range index behavior.
7. Re-run quicksort prefix bands and the full external timeout guard.

Parser byte-array lane:

- treat canonical `Array u8` and byte iteration as a VM-native boundary;
- avoid repeated `read_slot` method dispatch and byte boxing when shape guards
  prove canonical byte arrays;
- keep fallback for sparse/non-canonical arrays and all error cases.

Binarytrees near-term path:

- profile bytecode object construction, field access, recursion, and GC;
- add guarded direct struct/field bytecodes only through shared nominal layout
  metadata, not per-benchmark tree-node special cases;
- consider allocation pooling only if identity, mutation, and lifetime
  semantics are fully preserved;
- avoid a native binary-tree kernel.

Success condition:

- timeout bytecode families move into completed rows;
- bytecode hot profiles stop showing repeated boxing, generic lookup, and
  dynamic helper dispatch for statically obvious primitive/kernel operations.

## Rejected Paths Not To Repeat

These have already been tried or explicitly ruled out:

- whole-loop native quicksort scan/partition kernels;
- nominal `Array.swap` or benchmark-specific container special cases;
- swap call-site/body micro-quickening for quicksort;
- further tracked-array swap rewrites without fresh profile evidence;
- raw sentinel cache range tuning without a fresh profile reason;
- mutable pointer-shaped integer cells stored as `runtime.Value`;
- active-frame sidecar retrofits layered on the current dynamic slot model;
- untyped local propagation without v12 typechecker proof and a real typed
  representation;
- one-off read-slot proof-cache helper rewrites;
- disabled trace flag shaving;
- compiler fast paths for named non-primitive stdlib containers;
- accepting generated Go that merely wraps interpreter carriers more quickly.

## Recommended Assignment Order

If another model is taking over, assign work in this order:

1. Refresh the external scoreboard and confirm which families are missing,
   timed out, or regressed.
2. Return to bytecode quicksort with the real typed `i32` frame/register slice.
3. Use the Monte Carlo bytecode profile as the numeric-arithmetic companion
   case: optimize primitive numeric slots/arithmetic generally, not by adding
   a benchmark-specific RNG opcode.
4. Use pidigits as the host-extern/native-boundary companion case: optimize
   extern call overhead generally, not by adding pidigits-specific opcodes.
5. After quicksort completes externally, move to bytecode binarytrees.

## Definition Of Done

The performance program is complete when:

- all relevant `../benchmarks` families have Able implementations or explicit
  documented spec gaps;
- compiled Able is in the same practical performance class as Go on static
  benchmark logic;
- bytecode Able completes the benchmark suite and is competitive with
  mainstream bytecode/interpreter runtimes;
- generated-source audits prevent static compiled regressions back into
  interpreter carriers;
- VM guardrails prevent optimized bytecode paths from diverging from v12
  semantics;
- `PLAN.md`, `LOG.md`, and performance baselines explain the current state
  without requiring archaeology through old tranches.
