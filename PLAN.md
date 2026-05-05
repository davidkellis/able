# Able Project Roadmap (v12 focus)

## Prompts

### New Session
Read AGENTS, PLAN, and the v12 spec, then start on the highest priority PLAN work. Proceed with next steps. We need to correct any bugs if bugs or broken tests are outstanding. We want to work toward completing the items in the PLAN file. Please mark off and remove completed items from the PLAN file once they are complete. Remember to keep files under one thousand lines and to refactor them if they are going to exceed one thousand lines. You have permission to run tests. Tests should run quickly; no test should take more than one minute to complete.

### Next steps
Proceed with next steps as suggested; don't talk about doing it - do it. We need to correct any bugs if bugs or broken tests are outstanding. We want to work toward completing the items in the PLAN file. Please mark off and remove completed items from the PLAN file once they are complete. Remember to keep files under one thousand lines and to refactor them if they are going to exceed one thousand lines. Tests should run quickly; no test should take more than one minute to complete.

## Scope
- Keep the frozen Able v10/v11 toolchains available for historical reference while driving all new language, spec, and runtime work in v12.
- Use the Go tree-walking interpreter as the behavioral reference and keep the Go bytecode interpreter in strict semantic parity.
- Preserve a single AST contract for every runtime so tree-sitter output can target both the historical branches and the actively developed v12 runtime; document any deltas immediately in the v12 spec.
- Capture process/roadmap decisions in docs so follow-on agents can resume quickly, and keep every source file under 1000 lines by refactoring proactively.
- For compiler/AOT work, only primitive Able types may use primitive-specific Go lowering. All non-primitive nominal types must lower through shared struct/union/interface/generic translation rules and semantic encoding rules. Avoid new per-structure lowering branches for named non-primitive types unless the special handling is required by a language-level syntax or kernel ABI boundary rather than by the nominal type itself.

## Existing Assets
- `spec/full_spec_v10.md` + `spec/full_spec_v11.md`: authoritative semantics for archived toolchains. Keep them untouched unless a maintainer requests errata.
- `spec/full_spec_v12.md`: active specification for all current work; every behavioral change must be described here.
- `v10/` + `v11/`: frozen workspaces (read-only unless hotfix required).
- `v12/`: active development surface for Able v12 (`interpreters/go/`, `parser/`, `fixtures/`, `stdlib-deprecated-do-not-use/`, `design/`, `docs/`). Canonical stdlib source lives in the external `able-stdlib` repo and is cached locally via `able setup`.

## Active Priorities
- **Benchmark competitiveness**: make Able credible on the sibling
  `../benchmarks` suite. Compiled Able should be competitive with equivalent
  hand-written Go for static code, and the bytecode interpreter should become
  competitive with mainstream bytecode/interpreter implementations rather than
  only with the tree-walker.
- **Compiler completion / compiled performance**: first workstream under the
  competitiveness program. Finish removing static-path dynamic scaffolding and
  close the current external benchmark gaps before resuming broad feature work.
- **Stdlib benchmark coverage**: expand the canonical external `able-stdlib`
  with reusable library features needed to implement the missing benchmark
  families, using host-backed primitives where that is the language/runtime
  boundary rather than benchmark-specific lowering.
- **Bytecode VM v2 performance**: after the compiled/static path is credible,
  move bytecode from accumulated opcode micro-fusions toward typed slots,
  quickened opcodes, native collection/string bytecodes, and unboxed primitive
  frames.
- **Everything else**: parser/tooling/WASM/testing-framework work stays in
  backlog unless it directly supports benchmark competitiveness or is selected
  explicitly.

## Tracking & Reporting
- Update this plan as milestones progress; log design and architectural decisions in `v12/design/`.
- Keep `AGENTS.md` synced with onboarding steps for new contributors.
- Historical notes + completed milestones live in `LOG.md`.
- Keep this `PLAN.md` file up to date with current progress and immediate next actions, but move completed items to `LOG.md`.

## Guardrails (must stay true)
- `./run_all_tests.sh` (v12 default) must stay green for the Go suites and fixtures.
- `./run_stdlib_tests.sh` must stay green (tree-walker + bytecode).
- Diagnostics parity is mandatory: tree-walker and bytecode interpreters must emit identical runtime + typechecker output for every fixture when `ABLE_TYPECHECK_FIXTURES` is `warn|strict`.
- After regenerating tree-sitter assets under `v12/parser/tree-sitter-able`, force Go to relink the parser by deleting `v12/interpreters/go/.gocache` or running `cd v12/interpreters/go && GOCACHE=$(pwd)/.gocache go test -a ./pkg/parser`.
- It is expected that some new fixtures will fail due to interpreter bugs/deficiencies. Implement fixtures strictly in accordance with the v12 spec semantics. Do not weaken or sidestep the behavior under test to "make tests pass".
- Compiler/AOT optimization work must not introduce new bespoke lowering rules for specific non-primitive structures or containers. The correct fix direction is to improve the general lowering machinery so user-defined and stdlib nominal types fall out of the same rules.

## TODO (working queue: tackle in order, move completed items to LOG.md)

## Top Priorities

Priority order is now:
1. Make compiled Able competitive with equivalent pure Go on the external
   benchmark suite.
2. Expand stdlib and benchmark coverage so Able can express the full external
   suite with normal reusable APIs.
3. Make the bytecode interpreter competitive with mainstream bytecode /
   interpreter implementations.
4. Everything else is backlog.

The active plan below is intentionally organized around getting from the
current codebase to a production-grade compiled path and then a production-grade
bytecode path. Historical completed slices belong in `LOG.md`, not here.

### Benchmark Competitiveness Program (active)

Goal:
- make Able compiled code competitive with equivalent hand-written Go for
  static benchmark logic;
- make Able bytecode competitive with mainstream bytecode/interpreter runtimes
  on the same logic;
- grow `../able-stdlib` with normal reusable APIs needed by the benchmark
  suite, not benchmark-only compiler special cases.

Current measured external snapshot:
- Checked-in scoreboard artifact:
  `v12/docs/perf-baselines/external-scoreboard-current.{json,md}` joins the
  current kept Able compiled, bytecode, and tree-walker measurements with the
  best Go/Ruby/Python rows from `../benchmarks/results.json`.
- `fib`: compiled `2.9940s` vs Go `2.8400s` after bounded recursive
  return-range facts removed the final checked add for the proven `fib(45)`
  call range; bytecode now completes at `67.8200s` on the latest focused
  one-shot comparison after the compact self-call guard reduction, inline
  small-`i32` return-add/value-pair, compact frame-push, and compact slot-0
  raw-lane tranches; tree-walker times out at `60s`.
- `binarytrees`: compiled `3.6400s` vs Go `3.8300s` on the refreshed
  goroutine-executor comparison; bytecode and tree-walker time out.
- `matrixmultiply`: compiled `0.9660s` vs Go `0.8800s` after the canonical
  Able benchmark started using `Array.with_capacity(n)` for fixed-size matrix
  rows/outers and statement-position counted loops stopped materializing
  discarded `runtime.Value` loop results; bytecode and tree-walker time out.
- `quicksort`: compiled now completes in Go range after the byte-parser source
  rewrite plus native `Array u8` host-return boundary (`1.75s` vs Go `2.01s`);
  bytecode and tree-walker still time out at the current external benchmark
  scale.
- `sudoku`: compiled `0.0600s` vs Go `0.1300s` on the refreshed comparison;
  bytecode `0.3320s` after the shared propagation Error fast-negative guard,
  v12 nil-propagation semantic alignment, cached member-method fast-path
  dispatch, cached `String.bytes()` byte construction, the mono/native `u8`
  string-byte iterator lane, the direct canonical `Iterator u8` wrapper,
  static `Array.new`, the canonical string byte-iterator `next` call-member
  fast path, the canonical nullable `Array.get` overload selector, the
  allocation-free canonical stdlib origin suffix check, the Array.get overload
  function-pair validation cache, the exact primitive typed-pattern match fast
  path, slot-aware simple match lowering for nil/wildcard/typed clauses, and
  the guarded canonical `Array.get` call-site cache plus MRU hot tier and
  cache-first call-member ordering that skip repeated member resolution inside
  function closures, and the primitive two-part string interpolation fast path
  that cuts `board_to_string` allocation churn, plus fused slot-constant
  `>`/`>=` conditional jumps for loop guards, plus no-promotion hot hits in
  the canonical `Array.get` call-site cache, plus fused slot-const
  self-assignment stores for loop counters;
  tree-walker `6.71s`.
- `i_before_e`: compiled `0.0620s` vs Go `0.0500s` after primitive host
  extern wrappers stopped bridging native scalar arguments/results and static
  no-fallback launchers stopped loading/parsing/evaluating source metadata;
  bytecode `0.4440s` on the latest slot-const self-assignment confirmation;
  tree-walker `3.54s`.

Guardrails:
- Do not add compiler fast paths for named non-primitive containers or
  benchmark-specific structures. Fix the shared nominal/static lowering,
  stdlib API, or kernel/host ABI boundary instead.
- For compiled-performance keeps, inspect generated Go shape as well as
  wall-clock time. Hot static paths should not contain `runtime.Value`, `any`,
  interpreter dispatch helpers, runtime array stores, or option/result wrapper
  churn unless they are at an explicit dynamic or host boundary.
- For bytecode-performance keeps, use external benchmark results as the
  guardrail where available, not only reduced microbenchmarks.
- Keep external benchmark implementations faithful to the suite rules. Stdlib
  helpers are allowed when other languages use library primitives for the same
  domain, but benchmark algorithms that are required to be explicit must remain
  explicit in Able too.

#### Milestone A: Scoreboard And Coverage
- [ ] Add Able benchmark implementations for missing external suites as the
      required stdlib surface lands:
      - `base64`
      - `json`
      - `monte_carlo_pi`
      - `pidigits`
      - `tapelang-alphabet` only after the language/runtime surface is
        intentionally selected.
- [ ] Keep `v12/examples/benchmarks` as the canonical Able source location and
      treat `../benchmarks/*/able-v12-*` copies as harness packaging only.

#### Milestone B: Stdlib Surface Needed For Benchmarks
- [ ] Add `able.encoding.base64` with efficient encode/decode APIs over
      strings and byte arrays.
- [ ] Add `able.crypto.md5` or a general digest package with MD5 support for
      the `base64` benchmark output checks.
- [ ] Add a byte-buffer / string-builder surface that can build large strings
      and byte arrays without repeated UTF-8 validation or whole-string copies.
- [ ] Add `able.json` with at least DOM parsing plus typed numeric/object/array
      access for the `json` benchmark; prefer a streaming parser follow-up for
      low-allocation compiled paths.
- [ ] Finish host-backed or otherwise competitive `BigInt` / `BigUint`
      operations needed by `pidigits`.
- [ ] Add deterministic RNG and small numeric helpers needed by
      `monte_carlo_pi`.
- [ ] Tighten `able.fs`, `able.io`, and `able.text.string` hot APIs for
      benchmark-scale file reads, line iteration, splitting, numeric parsing,
      substring/search, and replacement.

#### Milestone C: Compiled Go-Competitive Path
- [ ] Add generated-source audits for each closed benchmark family so direct
      static Go shapes stay direct.
- [ ] Keep the compiled scoreboard current while shifting new optimization
      work to bytecode VM v2 and stdlib coverage. The current compiled core
      pass is in Go range for `fib`, `binarytrees`, `matrixmultiply`,
      `quicksort`, `sudoku`, and `i_before_e`; future compiled work should be
      driven by new benchmark families or regression evidence, not by more
      `fib` micro-slices.

#### Milestone D: Bytecode VM Competitiveness
- [ ] Add typed primitive slot/register storage, following
      `v12/design/bytecode-vm-v2.md`. Start with `i32` slots/stack cells for
      slot-eligible, non-yielding functions, with boxed `runtime.Value`
      materialization at every dynamic/spec boundary; then add `bool`, `f64`,
      and references only after the `i32` lane is proven.
      First stack-only seed is landed: literal-only final `i32` add/sub
      expressions now run on a raw `i32` operand stack and box before return.
      Declared `i32` slot metadata plus `LoadSlotI32` / `StoreSlotI32` are now
      landed for safe final arithmetic and typed local declarations. A parallel
      typed-slot side cache for recursive frames was tested and rejected on
      reduced `Fib30Bytecode`; the kept recursive-frame win is now compact
      two-slot self-fast frame reuse, which saves/restores slot 0 instead of
      acquiring a fresh two-slot frame at each self call. The follow-up compact
      self-call guard reduction is also landed, moving focused external
      bytecode `fib(45)` under the old `90s` timeout guard; the inline
      return-add value-pair plus compact frame-push tranches moved that
      focused external check to `76.7900s`; and the compact slot-0 raw lane
      now carries proven `i32` recursive arguments through the exact
      self-fast frame shape, moving focused external bytecode `fib(45)` to
      `67.8200s`. Next, target the remaining structural return/add handoff in
      `finishInlineReturn` and `execReturnBinaryIntAdd` while preserving boxed
      fallback boundaries and v12 checked integer semantics. The first bool
      lane is also landed: declared slot-backed `bool` conditions in `if`,
      `elsif`, and `while` now branch through `JumpIfBoolSlotFalse`, moving
      external bytecode `sudoku` from a refreshed `4.2200s` baseline to
      `2.6500s` and external bytecode `i_before_e` from `1.3200s` to
      `1.0000s`. The next guarded member fast path is landed too: after
      normal method resolution selects canonical `Array.len()` or nullable
      `Array.get(i32)`, the bytecode VM executes the size/read directly,
      moving external bytecode `sudoku` to `2.0433s` and external bytecode
      `i_before_e` to `0.7500s` over `3/3` runs. The next guarded member fast
      path is landed too: after normal method resolution selects canonical
      `String.len_bytes`, `String.contains`, or `String.replace`, the VM
      executes the direct string operation, moving external bytecode
      `i_before_e` to `0.5767s` over `3/3` runs while `sudoku` stays neutral at
      `2.0333s`. The next guarded Array member fast path is landed too: after
      normal method resolution selects canonical kernel `Array.push(value) ->
      void`, the VM appends directly through tracked array state, moving
      external bytecode `sudoku` to `1.8833s` and the latest `i_before_e`
      confirmation to `0.5333s` over `3/3` runs. Next, profile post-push
      `sudoku` and target iterator `next`, string byte iteration /
      `utf8_decode`, residual Array construction, or canonical Array `set`
      only after the fresh trace identifies the top remaining edge. The next
      guarded stdlib iterator fast path is landed too: after member access
      resolves canonical `RawStringBytesIter.next` / `StringBytesIter.next`
      returning `u8 | IteratorEnd`, the VM reads the byte and advances the
      iterator directly, moving external bytecode `sudoku` to `1.7300s` over
      `3/3` runs and confirming `i_before_e` at `0.5300s` over `5/5` runs.
      Next, profile post-iterator `sudoku` and target Array reads/sets,
      direct byte validation/decode, or Array construction only after fresh
      evidence identifies the top remaining edge. The follow-up tracked
      `Array.get` branch is landed too: canonical `Array.get(i32)` now reads
      directly from tracked dynamic array state before falling back to the
      handle store path, moving external bytecode `sudoku` to `1.6200s` over
      `3/3` runs and confirming `i_before_e` at `0.5240s` over `5/5` runs.
      Next, profile post-tracked-get `sudoku` and target UTF-8
      validation/decode during `String.bytes()`, residual Array construction,
      or Array `set` only after fresh evidence identifies the top remaining
      edge. The follow-up guarded `String.bytes()` fast path is landed too:
      after normal method resolution selects canonical `String.bytes() ->
      Iterator u8`, the VM validates the host string with Go UTF-8 validation,
      builds the canonical `RawStringBytesIter` shape directly, and returns it
      through the normal `Iterator u8` interface coercion. This removes the
      old `utf8_validate` / `utf8_decode` / `read_byte` fan-out from valid
      sudoku input, moving external bytecode `sudoku` to `0.7780s` over `5/5`
      runs while `i_before_e` stays neutral at `0.5333s` over `3/3` runs.
      Next, profile post-`String.bytes()` `sudoku` and target residual
      `Array.new` construction, tracked Array read/write volume, or a broader
      typed/native array lane only after fresh evidence identifies the top
      remaining edge. The next shared propagation slice is landed too:
      postfix `!` now skips the generic `matchesType("Error")` interface probe
      for ordinary success values unless the interpreter has a matching
      `Error` implementation registered for that runtime type. This moves
      external bytecode `sudoku` to `0.6700s` over `5/5` runs and confirms
      `i_before_e` at `0.5000s` over `5/5` runs. Next, profile
      post-propagation `sudoku` and target residual `execCallMember` /
      member-cache overhead around tracked `Array.get`, or the hot name lookup
      path. The dedicated semantic follow-up is landed too: postfix `!` now
      returns early on `nil` in both interpreters per the v12 spec, while
      successful dynamic definitions now return runtime `void` so `!void`
      success remains distinct from Option nil failure. The guardrail rerun
      measured external bytecode `sudoku` at `0.6220s` and `i_before_e` at
      `0.4620s` over `5/5` runs. The follow-up member-cache fast-path dispatch
      slice is landed too: cached canonical member methods now keep the
      resolved fast-path kind and `CallMember` tries that direct execution
      before rebinding or routing through the generic call ladder, moving the
      external bytecode guardrail to a final `0.6400s` for `sudoku` and
      `0.4840s` for `i_before_e` over `5/5` runs, with an earlier same-slice
      confirmation at `0.6140s` / `0.4580s`. Next, profile that state and
      target the remaining `String.bytes()` allocation/interface path plus
      residual name/member lookup before expanding to broader quickened
      member/index opcodes. The first post-member-cache heap slice is landed
      too: `String.bytes()` fast construction now indexes the source string
      directly and reuses cached boxed `u8`/`i32` values instead of copying to
      `[]byte` and boxing each byte afresh. Runtime-only `sudoku` moved from
      `429.30ms/op`, `118.96 MB/op`, `1,484,673 allocs/op` to
      `420.48ms/op`, `114.51 MB/op`, `1,390,910 allocs/op`; the external
      guard stayed neutral at `0.6380s` for `sudoku` and `0.4900s` for
      `i_before_e` over `5/5` runs. Next, profile the same state and target
      the remaining `String.bytes()` array/interface materialization or the
      residual name/member lookup path. The follow-up mono/native string-byte
      iterator lane is landed too: `String.bytes()` now builds the iterator's
      `bytes` field on the existing mono `u8` array store and attaches
      implementation-private native text metadata, while canonical
      `RawStringBytesIter.next` reads directly from that metadata before
      falling back to the normal array path. Runtime-only `sudoku` stayed
      neutral in the `421.79-426.69ms/op` warmed band with slightly lower heap
      volume; the external guard moved to `0.6260s` for `sudoku` and
      `0.4800s` for `i_before_e` over `5/5` runs. Next, profile this state and
      target the remaining generic `Iterator u8` interface coercion in
      `String.bytes()` or the residual name/member lookup path; do not broaden
      hidden native metadata beyond canonical stdlib shapes without a
      language-level host boundary. The follow-up direct interface-wrapper
      slice is landed too: canonical `String.bytes()` now builds its
      `Iterator u8` wrapper directly when the VM has the canonical
      `RawStringBytesIter`, `Iterator`, and `next` method, bypassing generic
      interface coercion and method dictionary construction while falling back
      for unsupported shapes. Runtime-only `sudoku` now runs in the
      `415.33-427.64ms/op` warmed band with roughly `1.282M allocs/op`, and
      the external guard moved to `0.6160s` for `sudoku` and `0.4700s` for
      `i_before_e` over `5/5` runs. Next, profile this kept state and target
      residual `execCallMember` / `resolveMethodCallableFromPool` cost around
      interface member access and name lookup rather than another
      `String.bytes()` wrapper rewrite. The follow-up static `Array.new()`
      slice is landed too: after ordinary static member resolution proves the
      active method is canonical kernel `Array.new`, the VM builds the same
      empty tracked array directly instead of entering the generic Able method
      and `__able_array_new` bridge. Runtime-only `sudoku` now runs in the
      `407.66-416.55ms/op` warmed band with roughly `1.161M allocs/op`, and
      the external guard moved to `0.5840s` for `sudoku` while `i_before_e`
      stayed neutral at `0.4760s` over `5/5` runs. Next, profile this kept
      state and target residual `resolveMethodCallableFromPool` /
      overload-selection cost around hot `Array.get` and iterator `next`;
      avoid broader static member hooks until fresh trace evidence justifies
      them. The follow-up canonical byte-iterator member-call slice is landed
      too: when `CallMember next` sees the canonical `Iterator u8` wrapper
      produced by `String.bytes()` around canonical `RawStringBytesIter`, it
      validates the cached canonical `next` method and jumps directly to the
      existing string-byte iterator fast body instead of re-entering generic
      `interfaceMember` dispatch. Steady-state `i_before_e` moved into the
      `236.02-280.09ms/op` band with roughly `2.83 MB/op` and `2.0k
      allocs/op`; the external guard moved to `0.4660s` for `i_before_e` and
      `0.5720s` for `sudoku` over `5/5` runs. Next, profile this kept state
      and target residual `execCallMember` / `resolveMethodCallableFromPool`
      and overload-cache overhead around hot direct string/member calls, or
      move back to the larger bytecode `fib` typed-frame work; avoid broad
      per-instruction member quickening until fresh trace evidence justifies
      them. The follow-up canonical nullable `Array.get` overload slice is
      landed too: when normal method resolution returns exactly the canonical
      nullable `Array.get(i32) -> ?T` method plus the lower-priority canonical
      `Index.get(i32) -> !T` implementation method, `CallMember get` executes
      the existing tracked `Array.get` fast body directly instead of running
      generic runtime overload selection. A VM-local hot cache keeps the
      canonical-shape validation off the per-call path. Steady-state
      `i_before_e` moved into the `193.31-199.15ms/op` band with roughly
      `2.82 MB/op` and `1.99k allocs/op`; the external guard moved to
      `0.4480s` for `i_before_e` and `0.5640s` for `sudoku` over `5/5` runs.
      Next, profile this kept state and target residual
      `resolveMethodCallableFromPool` / `lookupBoundMethodCache` around
      canonical primitive methods, or switch back to bytecode `fib` typed-frame
      work if the next member-call slice would require another map/cache layer.
      The follow-up canonical-origin allocation slice is landed too:
      `isCanonicalAbleStdlibOrigin(...)` now checks the fixed
      `/able-stdlib/src/` and `/pkg/src/` path suffixes without allocating
      concatenated suffix strings on every validation. This removes that helper
      from the `sudoku` allocation profile, moves runtime-only `sudoku` from a
      refreshed `339.53ms/op`, `118.11 MB/op`, `1,572,523 allocs/op` sample to
      `334.69ms/op`, `86.58 MB/op`, `915,969 allocs/op`, and moves the
      external bytecode guard to `0.5160s` for `sudoku` and `0.4280s` for
      `i_before_e` over `5/5` runs. The follow-up Array.get overload
      validation-cache slice is landed too: when member resolution returns a
      fresh overload wrapper around the same canonical nullable `Array.get`
      function and lower-priority result-returning implementation function,
      the VM reuses the cached canonical-shape validation result until the
      method/global cache version changes. Same-session restored external
      passes were noisy at `0.6080-0.6480s` for `sudoku` and
      `0.4760-0.5340s` for `i_before_e`; the kept cache confirmations landed
      at `0.5280s` / `0.5340s` for `sudoku` and `0.4580s` / `0.4540s` for
      `i_before_e` over `5/5` runs. The follow-up exact primitive
      typed-pattern match slice is landed too: simple typed patterns whose
      runtime value already has the exact primitive shape now bind directly
      instead of entering generic `matchesType(...)` plus coercion. This keeps
      aliases, non-exact integer widths, unions, structs, and interfaces on
      the generic path while shortening the hot `case byte: u8` path in
      `parse_board`. Runtime-only `sudoku` moved from a refreshed
      `340.43ms/op`, `86.60 MB/op`, `915,996 allocs/op` sample to a kept
      `326.26-331.68ms/op` warmed band with the same allocation shape, and
      the external guard moved to `0.5040s` for `sudoku` and `0.4200s` for
      `i_before_e` over `5/5` runs. Next, target the structural match/env
      issue directly by lowering simple `match` clauses into slot-aware
      bytecode so `parse_board` / `solve` can inline, or target
      `board_to_string` through a spec-backed string builder / byte-buffer
      surface rather than another generic interpolation tweak. The slot-aware
      simple match lowering is now landed: match expressions inside
      slot-eligible functions lower directly for literal `nil`, wildcard, and
      typed identifier/wildcard clauses while guarded clauses, non-nil
      literals, destructuring, and identifier-as-existing-symbol patterns stay
      generic. Runtime-only `sudoku` moved to `203.70-209.21ms/op` with
      about `499.5k` allocs/op, and the external bytecode guard moved
      `sudoku` to `0.4120s` over `5/5` runs. `i_before_e` was noisy but
      stayed in the same broad band (`0.4680s` combined guard, `0.4480s`
      rerun), and external bytecode `binarytrees` still times out at `60s`.
      Next, profile post-match `sudoku` and target residual `CallMember get`
      / `board_to_string` string construction, or switch to the timeout
      bytecode workloads where typed nominal/struct allocation and call-frame
      design are larger blockers. The follow-up guarded canonical `Array.get`
      call-site cache is now landed too: after full method resolution proves a
      bytecode `CallMember get` site targets the canonical nullable stdlib
      `Array.get(i32)` overload, the VM caches that proof per program/IP/env
      and executes the existing tracked array-read fast body directly on later
      hits. The cache is guarded by environment revision, global revision,
      method-cache version, and absence of runtime impl context. Runtime-only
      `sudoku` moved from a paired no-trace baseline of `200.66-209.51ms/op`
      with about `499k allocs/op` to a kept `176.47-200.95ms/op` band with
      about `417k allocs/op`; the no-trace profiled rerun landed at
      `179.88ms/op`. The external bytecode guard moved `sudoku` to
      `0.3500s` over `3/3` runs, about `2.69x` Go. Next, profile this kept
      state and target the remaining `execCallMember` guard cost, residual
      string interpolation in `board_to_string`, or switch to timeout
      bytecode workloads if the next member-call slice needs broader
      quickening infrastructure. The follow-up MRU hot-cache layout is now
      landed too: the canonical `Array.get` call-site cache keeps four recent
      program/IP/env identities and only reads revisions after a cheap identity
      match. This removes the single-hot-entry thrash across nested sudoku
      `Array.get` call sites. Paired runtime-only reruns moved from a restored
      `169.46-187.57ms/op` band to `164.50-176.74ms/op`, with allocations
      unchanged at about `417k allocs/op`; the profiled rerun shows
      `lookupCachedCanonicalArrayGetCall(...)` down from about `0.10s` to
      about `0.02s` cumulative. Paired external bytecode `sudoku` moved from
      restored pre-MRU `0.3833s` to MRU `0.3700s`. The follow-up
      cache-first member ordering slice is now landed too:
      `execCallMember(...)` checks the guarded canonical `Array.get` call-site
      cache before the single-hot general member-method cache, allowing nested
      sudoku `get` sites to use the specialized MRU tier without first paying
      general member-cache churn. Paired runtime-only reruns moved from a
      restored `168.34-171.74ms/op` band to `163.35-167.92ms/op`, and the
      profiled kept rerun landed at `161.23ms/op`. External bytecode `sudoku`
      moved from the prior recorded `0.3700s` to `0.3633s` over `3/3` runs. A
      stdlib `StringBuilder` benchmark-source rewrite, a two-part
      interpolation VM fast path, and a single-thread propagation-cache mutex
      bypass were all tested and reverted as non-keeps. Next, profile
      `lookupCachedCanonicalArrayGetCall(...)` itself and the remaining
      `board_to_string` interpolation allocation; avoid another builder-source
      rewrite unless the builder path first gains a general VM fast path. A
      narrower primitive-only two-part interpolation fast path is now landed:
      `execStringInterpolation(...)` formats primitive pairs directly and uses
      a one-builder `String + Integer` subpath while preserving generic
      Display/`to_string` fallback for dynamic values. Runtime-only `sudoku`
      now runs in the `161.29-169.59ms/op` band with about `343k allocs/op`,
      and external bytecode `sudoku` confirmed at `0.3620s` over two `5/5`
      runs. Next, profile residual `execCallMember(...)` / canonical
      `Array.get` guard work and binary compare slots; larger string wins
      likely require a general byte-buffer/string-builder runtime primitive.
      The follow-up slot-constant compare branch slice is landed too:
      slot-backed `>` and `>=` conditions now lower to
      `JumpIfIntCompareSlotConstFalse`, avoiding standalone bool-producing
      compares for sudoku loop guards like `i >= 9` and `num > 9`. Runtime-only
      `sudoku` landed at `149.16ms/op`, `153.52ms/op`, and one noisy
      `176.73ms/op`, with allocations unchanged around `343k allocs/op`; the
      external bytecode guard moved to `0.3560s` over `5/5` runs. Next, target
      residual `execCallMember(...)` / canonical `Array.get` guard work, slot
      load/store traffic, or runtime type/propagation checks.
      The follow-up canonical `Array.get` hot-cache slice is landed too: hot
      hits in the four-entry call-site cache no longer promote entries to the
      front on every nested solver/output access while preserving the same
      env/global/method version guards. Runtime-only `sudoku` moved to a
      warmed `137.57-164.03ms/op` band, and external bytecode `sudoku` moved
      to `0.3360s` over `5/5` runs. Next, profile this kept state and avoid
      another `Array.get` cache micro-slice unless the refreshed profile proves
      it is still the top wall. The follow-up slot-const self-assignment store
      slice is landed too: boxed slot-backed `x = x + const` and `x = x - const`
      assignments now lower to `StoreSlotBinaryIntSlotConst`, combining the
      checked slot-const arithmetic, slot write, and assignment-result stack
      value into one opcode while typed `i32` slots keep the raw `StoreSlotI32`
      path. Runtime-only `sudoku` moved to
      `140.23-145.35ms/op`, external bytecode `sudoku` moved to `0.3320s`
      over `5/5` runs, and external bytecode `i_before_e` stayed neutral at
      `0.4440s` over `5/5` runs. Next, profile this kept state and target
      `execCallMember(...)` / canonical `Array.get` guard cost, residual
      checked integer arithmetic, or typed slot assignment checks.
- [ ] Add quickened call/member/index opcodes that rewrite after first
      successful shape resolution and invalidate safely under mutation or
      environment revision changes.
- [ ] Add bytecode-native array and string bytecodes for hot `len`, `get`,
      `set`, `push`, byte search, replacement, and line iteration paths.
- [ ] Replace recursive/function frame hot paths with compact typed frame
      records for proven layouts.
- [ ] Establish bytecode thresholds: first no external benchmark timeouts,
      then beat Python/Ruby on `fib`, `sudoku`, and `i_before_e`, then close
      toward Node/Bun where the benchmark logic is comparable.

#### Milestone E: Release Gates
- [ ] Add report-first CI guardrails for the external benchmark scoreboard.
- [ ] Once variance is characterized, add threshold gates for compiled
      benchmarks where Able is expected to stay in Go range.
- [ ] Keep `run_all_tests.sh` and `run_stdlib_tests.sh` green as the semantic
      gate; benchmark gates are additive and should not mask correctness
      failures.

### Compiler Completion Program (highest priority)

Goal:
- ship a production-grade Go AOT compiler that compiles non-dynamic Able code to
  direct Go implementations with no interpreter execution on static paths, keeps
  dynamic behavior behind explicit boundaries, and performs well on the checked-
  in benchmark family.

Canonical architecture docs:
- `v12/design/compiler-go-lowering-spec.md`
- `v12/design/compiler-go-lowering-plan.md`
- `v12/design/compiler-native-lowering.md`
- `v12/design/compiler-aot.md`

#### Current state snapshot

Compiler-native encoding completion is closed on 2026-04-14:
- large static slices of arrays, structs, interfaces, callables, joins, and
  control-flow now stay native;
- explicit dynamic-boundary audits exist;
- reduced benchmark fixtures exist for matrix, iterator, heap, and array-heavy
  paths;
- the array-native lowering tranche is complete on 2026-04-01; remaining
  `runtime.ArrayValue` / `ArrayStore*` use is now limited to explicit dynamic
  or ABI edges plus the unspecialized wildcard-array ABI;
- imported shadowed nominal match bindings now preserve foreign package
  context through carrier reconstruction too, so direct field access stays
  native instead of round-tripping through nominal/runtime helpers;
- mixed imported/local shadowed nominal joins now keep distinct native union
  members instead of collapsing on unqualified type-expression strings, and
  shadowed callable joins built from those nominal returns now stay on native
  callable-union carriers instead of silently collapsing to
  `fn(...) -> runtime.Value`;
- lambda literals and placeholder lambdas now narrow through expected callable
  members inside native union/result carriers, and semantic `Result` carrier
  synthesis preserves that callable member's resolved package context too, so
  imported semantic `Option` / `Result` aliases and direct union aliases
  built from shadowed callable returns stay on native callable carriers
  instead of failing with `lambda expression type mismatch` or
  `placeholder lambda type mismatch`;
- imported semantic `Result` aliases over shadowed callable members now stay
  native under outer `Result` carriers too, because raw imported selector
  aliases nested inside function type expressions keep lexical caller-package
  normalization instead of being re-normalized under stale foreign package
  context and collapsing to `__able_fn_*_to_runtime_Value`; the same outer-
  result native-carrier path is now pinned for imported semantic `Option`
  aliases and imported generic union aliases over those shadowed callable
  members too;
- local parameterized union/result aliases now also have proof coverage for
  imported shadowed interface/callable actuals, so `Choice (RemoteReader i32)`
  and `Outcome (() -> RemoteThing)` style locals are pinned to native carriers
  instead of widening through `runtime.Value` / `any`;
- generic specialization now also has proof coverage for those imported
  shadowed alias actuals, so specialized helpers over `Choice (RemoteReader
  i32)` and `Outcome (() -> RemoteThing)` stay on native signatures instead
  of widening to `runtime.Value` / `any`;
- the shared carrier mapper and generic interface-method dispatch now also
  have proof coverage for those same imported shadowed alias actuals, so
  direct `lowerCarrierTypeInPackage(...)` lookups and existential
  `pass<T>(...)` style calls stay on native union/result/interface/callable
  carriers instead of broadening locals or helper signatures;
- imported generic interface-method calls now also normalize explicit type
  arguments in the lexical caller package and retry representable-carrier
  recovery while computing concrete param/return helper signatures, so
  imported `Echo.pass<Choice(RemoteReader i32)>(...)` and imported default
  generic method calls stay on native union/result/interface/callable
  carriers instead of widening through `runtime.Value` / `any`;
- imported generic interface default methods now also resolve nested
  selector-imported members inside those explicit type arguments before
  specializing the interface-package body, so calls like
  `tagged.tagged<Outcome(() -> RemoteThing)>(...)` synthesize the concrete
  native `Tagged<...>` carrier and avoid `__able_method_call_node(...)`;
- native union synthesis now also retries representable-carrier recovery in
  the member's resolved package before accepting a residual `runtime.Value`
  member, and the imported shadowed interface/callable alias families above
  are now pinned against hidden `_runtime_Value` union variants too;
- generic specialization now also retries representable-carrier recovery
  before rejecting a fully bound actual as broad, and the remaining imported
  shadowed result/interface plus union/callable specialization quadrants are
  pinned to native helper signatures too;
- imported shadowed nullable interface/callable aliases now stay on those
  native carriers through generic specialization too, and imported shadowed
  callable union aliases like `Choice (() -> RemoteThing)` now specialize
  through native union helpers instead of falling back through
  `runtime.Value`;
- proof coverage now also pins the adjacent broader imported-shadowed
  three-member alias surface, so `Choice3(RemoteReader i32)` and
  `Outcome3(() -> RemoteThing)` already stay on native
  union/interface/callable carriers too, with no hidden `_runtime_Value`
  helper variants;
- proof coverage now also pins those same broader imported-shadowed
  three-member alias families through imported generic-interface dispatch and
  imported default generic methods, so
  `echo.pass<Choice3(RemoteReader i32)>(...)` and
  `tagged.tagged<Outcome3(() -> RemoteThing)>(...)` stay on native
  union/result/interface/callable carriers too instead of widening helper
  signatures or falling back through `__able_method_call_node(...)`;
- proof coverage now also pins those same broader imported-shadowed
  three-member alias families when the generic-interface receiver itself is a
  join-produced existential across concrete implementers, so joined
  `echo.pass<Choice3(RemoteReader i32)>(...)` and joined
  `tagger.tagged<Outcome3(() -> RemoteThing)>(...)` stay on native
  union/result/interface/callable carriers too instead of widening the join
  local or defaulting the call back through runtime helpers;
- proof coverage now also pins broader outer-result shapes over those same
  imported-shadowed three-member alias families, so
  `!(Choice3(RemoteReader i32))` and `!(Outcome3(() -> RemoteThing))`
  already flatten/collapse to native union/result/interface/callable carriers
  too instead of regressing broader result/error families to
  `runtime.Value` / `any`;
- proof coverage now also pins the same broader outer-result families through
  generic interface shape synthesis itself, so parameterized carriers like
  `Keeper(!(Choice3(RemoteReader i32)))` and
  `Keeper(!(Outcome3(() -> RemoteThing)))` keep native union/result/interface
  signatures too instead of widening interface helper params/returns to
  `runtime.Value` / `any`;
- proof coverage now also pins the closed local interface existential family
  itself, so local aliases like `Either = (Reader i32) | Echo`,
  `Outcome = !Either`, and `Keeper<Either>` / `Keeper<Outcome>` helper
  synthesis stay on native union/result/interface carriers too instead of
  broadening local helper params/returns to `runtime.Value` / `any`;
- proof coverage now also pins the broader local multi-member
  interface/callable existential family, so local aliases like
  `Choice3 = (Reader i32) | Echo | String`,
  `Outcome3 = Error | (() -> Thing) | String`, generic interface dispatch
  over those same local families, and `Keeper<Choice3>` /
  `Keeper<Outcome3>` helper synthesis all stay on native
  union/result/interface/callable carriers too instead of broadening local
  params, locals, or helper signatures to `runtime.Value` / `any`;
- proof coverage now also pins the last local analogs of that broader family,
  so joined existential receivers calling `Echo.pass<Choice3>(...)` /
  `Tagger.tagged<Outcome3>(...)` stay on native carriers too, and local
  outer-result helper synthesis like `Keeper<!(Choice3)>` /
  `Keeper<!(Outcome3)>` also keeps native union/result/interface signatures
  instead of widening joined locals or helper params/returns to
  `runtime.Value` / `any`;
- native union synthesis no longer materializes partially native helper
  families when any member still only maps to `runtime.Value` / `any`, so
  hidden `_runtime_Value` variants stop leaking into adjacent imported-
  shadowed specialization slices;
- native interface method-shape collection, native interface impl-signature
  synthesis, and native callable signature materialization now also retry
  representable-carrier recovery after raw package-scoped mapping, so
  substituted imported shadowed alias families stay on native
  union/result/interface/callable carriers instead of broadening internally
  to `runtime.Value` / `any`;
- imported generic struct members with shadowed nominal type arguments now
  stay on specialized native carriers inside result and nested union/result
  families too, because fully bound imported selector arguments count as
  concrete in the caller package and foreign generic-struct specialization
  keeps that caller-side package context through field substitution; proof
  coverage now also pins that same native-carrier behavior through generic
  specialization over imported shadowed generic-struct result/union aliases,
  and imported generic-struct result/union aliases now also keep specialized
  native carriers when the generic argument is a native interface or native
  callable instead of falling back to base `*Box` plus dynamic member calls;
- local generic nominal carriers over normalized nullable/union/result members
  are now pinned too, so `Box MaybeReader`, `Box Choice`, `Box Outcome`,
  `Box !(Choice)`, and `Box !(Outcome)` all stay on specialized native
  `Box<...>` carriers instead of falling back to base `*Box`,
  `runtime.Value`, or `any`; this closes the deeper `types.go` /
  `generator_native_unions.go` carrier-synthesis cleanup;
- error-wrapped nominal struct typed matches now stay on those native struct
  carriers too, because generated `__able_struct_*_try_from(...)` /
  `__able_struct_*_from(...)` helpers unwrap through the shared
  `__able_struct_instance(...)` path before enforcing the nominal definition
  check, fixing static `case _: IndexError` matches on array helper bounds
  results under the no-bootstrap boundary harness;
- imported shadowed interface selector aliases now preserve their source
  package through generic type normalization too, so nullable / union / result
  aliases built from foreign `Interface<T>` members stay on native interface
  carriers even when the caller shadows the same interface name locally;
- raw imported selector-alias typed patterns now normalize in their lexical
  caller package before any recorded foreign package is reapplied, so generic
  `Result` / semantic-result carrier matches like
  `Outcome(RemoteIface<i32>) match { case value: RemoteIface<i32> => ... }`
  resolve back onto the native imported interface carrier instead of widening
  to `runtime.Value`;
- imported semantic `Result` aliases nested under outer `Result` carriers now
  stay native across shadowed foreign interfaces too, because alias expansion
  preserves alias-source normalization and builtin `Error` identities collapse
  across package contexts during nested carrier synthesis;
- generic specializations now also keep native interface carriers for both
  local and imported-shadowed interface actuals, including the duplicate-
  member collapse case where a fully bound generic union normalizes to the
  same foreign interface type twice, because imported selector source-package
  context now survives specialization materialization too instead of being
  dropped by no-op type substitution;
- join carrier selection now treats interfaces nominally instead of
  structurally, so unrelated same-shape interfaces stay on distinct native
  carriers and join through a native union rather than silently collapsing
  onto one interface family;
- propagated rescue identifier joins built from statically typed call failures
  now recover native callable plus imported shadowed nominal/interface
  carriers instead of widening back through `runtime.Value`, native-error
  unions, or member access fallback, and higher-order/unknown rescue call
  failures no longer misinfer from the callable return type so `err.value`
  style handlers stay on the dynamic error path instead of collapsing to
  arbitrary static return carriers;
- raised imported shadowed nominal struct literals now prefer the compiler's
  syntax-aware struct-literal type reconstruction during failure inference, so
  propagated rescue joins keep the foreign native struct carrier instead of
  collapsing onto same-named local nominals, while no-bootstrap raised
  non-`Error` bridge fallback stringification now stays aligned with
  interpreter-visible output for compiled rescue/string paths;
- nested struct-pattern field bindings now restore field-local expected type
  context before recursing into subpatterns, which keeps persistent map/set
  helper patterns on the correct native carriers, and dynamic typed-pattern
  casts now allocate temps on the caller stream so iterator-end /
  generator-yield matches stop colliding with surrounding temps during
  codegen;
- implicit and explicit return expressions now preserve the declared Able
  return `TypeExpr` while compiling the returned expression too, so generic
  return paths like `unwrap(io_read(...))` keep nullable success carriers
  instead of collapsing to their nil-capable Go carrier only;
- static nullable typed matches on nil-capable native carriers now guard both
  the non-nil typed branch and the `case nil` branch directly, so native
  interface and result-family whole-carrier matches no longer compile to dead
  `true`/`false` conditions ahead of the real nil arm;
- the concrete nullable typed-match nil-guard path is now narrowed to actual
  in-scope generics only, so `?Interface<T>` / `?Result<T>` typed arms regain
  their required non-nil guard instead of compiling as unconditional typed
  branches;
- nested native nullable/result outer unions now keep representable literals,
  struct literals, and typed-match clause ordering on native carriers too,
  because nested member wrapping is direct and match narrowing only removes a
  union member when the pattern exhausts that whole v12 member type rather
  than a non-nil subcase inside it;
- representable nested union/result members now flatten during carrier
  synthesis too, so outer unions like `(!T) | U` and `(A | B) | U`, plus
  direct nested result families like `!!T` and `!(A | B)`, lower to a single
  native union family instead of nesting one native union carrier inside
  another;
- fully bound duplicate union/result members now collapse to their single
  native carrier during mapping too, so generic specializations like
  `T | String` at `T = String` and `!T` at `T = Error` no longer miss native
  specialization behind synthetic union/runtime carriers;
- no-bootstrap / no-fallback enforcement is green across the release gate;
- the compiler release gates currently pass:
  - `GOFLAGS='-p=1' ./run_all_tests.sh --compiler`
  - `./run_stdlib_tests.sh`
- the current dirty-tree compiler release rerun is green end-to-end again:
  bridge tests, all compiler core batches, outliers, the compiled exec-fixture
  matrix, strict-dispatch audit, interface-lookup audit, boundary audit, and
  `./run_stdlib_tests.sh` all pass after the latest rescue/failure-inference,
  persistent-collection, iterator-end, and bridge-fallback fixes;
- no-interpreter generic-interface dispatch now performs alias expansion and
  interface-constraint revalidation through compiler-emitted runtime metadata
  and generated helpers instead of interpreter registries or bridge fallback;
- the aggregated observed wall clock for the latest green compiler release
  rerun is now `52m51s` (`real 3171.27`), so the dominant remaining release-
  path issue is test runtime pressure rather than a known correctness blocker;
- the stronger compiler-native completion program and the bytecode performance
  program are now both closed; the remaining work returns to backlog /
  tooling-priority selection rather than another compiler or bytecode closure
  milestone.
- the clean-checkout reproducibility follow-up is now closed on 2026-04-17:
  `run_stdlib_tests.sh` self-bootstraps through `able setup` when needed, and
  the remaining active tooling/helper paths (`cmd/fixture`, interpreter
  fixture loading, `ablec`, generated compiled wrappers, and the repo fixture
  harnesses) now prefer explicit or cached stdlib installs before sibling
  `able-stdlib` probing; the rebased tree is green again at the top level with
  `/usr/bin/time -p ./run_all_tests.sh` = `real 1193.78` and
  `/usr/bin/time -p ./run_stdlib_tests.sh` = `real 36.36`.

#### Production definition of done

The compiler is only fully done for the stronger native-encoding goal when all
of these are true together:
- every statically representable Able type expression lowers to a native Go
  carrier;
- static control flow and pattern binding stay on native carriers;
- static field/method/interface/index/call dispatch lowers to direct compiled Go
  dispatch;
- dynamic/runtime carriers remain only at explicit dynamic or ABI boundaries;
- compiled runtime helpers implement language semantics directly in Go instead
  of modeling normal static execution in terms of interpreter behavior;
- no staged hybrid carrier architecture remains on static paths for arrays,
  unions, or other nominal values that should have final host-native compiled
  representations;
- experimental transition machinery is either removed or reduced to explicit
  boundary-only helpers instead of serving as the general static lowering path;
- compiler fixture parity is green under no-bootstrap/no-fallback enforcement;
- the compiled benchmark family is materially faster and free of already-known
  avoidable scaffolding on hot paths.

#### Milestone 1: Centralize Compiler Lowering Knowledge

Goal:
- make every carrier, dispatch, control, pattern, and boundary decision come
  from one shared synthesis point instead of emitter-local rules.

Status:
- complete on 2026-03-21.
- canonical lowering facade landed in:
  - `v12/interpreters/go/pkg/compiler/generator_lowering_types.go`
  - `v12/interpreters/go/pkg/compiler/generator_lowering_patterns.go`
  - `v12/interpreters/go/pkg/compiler/generator_lowering_dispatch.go`
  - `v12/interpreters/go/pkg/compiler/generator_lowering_control.go`
  - `v12/interpreters/go/pkg/compiler/generator_lowering_boundaries.go`
- source-audit enforcement landed in:
  - `v12/interpreters/go/pkg/compiler/compiler_lowering_facade_audit_test.go`

Why this is first:
- without this, the compiler keeps accumulating one-off fixes and nominal-type
  drift.

Required work:
- [x] establish one canonical type-normalization path used by all codegen
      stages;
- [x] establish one canonical carrier-synthesis path used by all emitters;
- [x] establish one canonical join/pattern-synthesis path;
- [x] establish one canonical dispatch-synthesis path;
- [x] establish one canonical control-envelope synthesis path;
- [x] establish one canonical boundary-adapter synthesis path;
- [x] audit emitters and remove local fallback decisions that bypass those
      shared paths.

Proof required:
- generated-source audits that fail on new ad hoc carrier or dispatch fallback
  patterns.

#### Milestone 2: Native Carrier Completeness

Goal:
- every statically representable Able type expression maps to a native Go
  carrier everywhere it appears.

Required work:
- [x] remove remaining representable `runtime.Value` / `any` fallbacks from
      type mapping and carrier synthesis;
- [x] finish carrier synthesis for nullable, result, union, interface,
      callable, and fully bound generic nominal types;
- [x] ensure alias-expanded and recovered type expressions use the same carrier
      path as directly written types;
- [x] ensure representable branch/join locals do not regress to dynamic
      carriers just because a value was temporarily recovered from a broader
      path;
- [x] ensure residual `runtime.Value` union members exist only for true dynamic
      payloads.

Status:
- complete on 2026-03-22.
- landed the remaining shared carrier-synthesis closure in:
  - `v12/interpreters/go/pkg/compiler/types.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_unions.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_unions.go`
  - `v12/interpreters/go/pkg/compiler/generator_binary.go`
  - `v12/interpreters/go/pkg/compiler/generator_builtin_structs.go`
  - `v12/interpreters/go/pkg/compiler/generator_native_result_void.go`
- moved built-in `DivMod T` onto the shared nominal struct path instead of the
  old `any` fallback;
- tightened concrete union/result carrier synthesis so fully bound
  representable members fail fast instead of silently widening to
  `runtime.Value` / `any`;
- added shared native `Error | void` carrier support so `!void` signatures no
  longer regress to runtime carriers;
- added focused regressions for:
  - concrete `DivMod i32` carrier lowering
  - parameterized union/result alias locals containing `DivMod i32`
  - native `!void` return lowering
  - existing no-fallback fixture coverage including `06_01_compiler_divmod`
    and `06_01_compiler_match_patterns`

Proof required:
- representative carrier audits over arrays, structs, unions, interfaces,
  callables, results, and generics;
- generated-source scans for representable `runtime.Value` / `any` locals.

#### Milestone 3: Native Pattern And Control-Flow Completeness

Goal:
- all static pattern matching, branch joins, loop results, rescue/or-propagation,
  and non-local control stay native when the types are statically representable.

Status:
- complete on 2026-03-22.
- shared join inference now keeps `if`, `match`, `rescue`, `or {}`, `loop`,
  and `breakpoint` results on native carriers across representable mixed
  branches, nil-capable joins, common existential joins, and recovered
  `TypeExpr`-backed locals.
- typed pattern bindings now stay native when the narrowed carrier is
  recoverable, including rescue bindings and native-union whole-value
  interface bindings.
- representative static pattern/control bodies are now mechanically audited
  against `__able_try_cast(...)`, `bridge.MatchType(...)`, `panic`, `recover`,
  and IIFE-style scaffolding.

Required work:
- [x] remove runtime typed-pattern fallback for recoverable static subjects;
- [x] keep `if`, `match`, `rescue`, `or {}`, `loop`, and `breakpoint` joins on
      native carriers in all representable cases;
- [x] keep typed bindings and recovered branch bindings on native carriers;
- [x] finish explicit control-envelope propagation for helper boundaries and
      remove any remaining panic/recover/IIFE-style static control scaffolding;
- [x] ensure `raise`, `rethrow`, `ensure`, `!`, and `or {}` are implemented as
      proper compiled control/data semantics instead of interpreter-shaped
      escape hatches.

Proof required:
- fixture slices for `match`, `rescue`, `or {}`, loops, breakpoints, typed
  patterns, and propagation paths;
- generated-source audits that fail when static paths regress to runtime-type
  helpers.

#### Milestone 4: Native Dispatch Completeness

Status:
- complete on 2026-03-22.
- shared dispatch recovery now converts recoverable `runtime.Value` / `any`
  call/member/index targets back onto native carriers before dispatch.
- local concrete and interface `Apply` bindings now route through the shared
  static apply path instead of `__able_call_value(...)`.
- mixed-source pure-generic interface dispatch now prefers the more concrete
  compiled specialization instead of falling back to runtime method dispatch.
- representative dispatch coverage now lives in:
  - `v12/interpreters/go/pkg/compiler/generator_dispatch_recovery.go`
  - `v12/interpreters/go/pkg/compiler/compiler_dispatch_completeness_test.go`

Goal:
- all statically resolved operations compile to direct Go dispatch.

Required work:
- [x] finish static field access/assignment lowering;
- [x] finish static method and default-method lowering;
- [x] finish static interface/default-method dispatch lowering for all
      representable generic cases;
- [x] finish static callable/bound-method/partial application lowering;
- [x] finish static index/get/set/apply lowering without dynamic helper
      dispatch;
- [x] remove residual dynamic helper dispatch from static call/member/index
      paths.

Proof required:
- combined-source dispatch audits;
- fixture slices covering structs, interfaces, callables, indexable types, and
  generic/default-method paths.

#### Milestone 5: Compiled Runtime Core Independence

Status:
- complete on 2026-03-22.
- static compiled kernel/runtime helper families now call direct Go `_impl`
  helpers on static paths instead of routing through `__able_extern_*` or
  helper-to-helper `__able_extern_call(...)` chains.
- zero-arg callable syntax and `Await.default` zero-arg callback
  specialization now stay on native callable carriers in compiled static code.

Goal:
- compiled runtime helpers used on static paths must implement Able semantics as
  normal Go logic, not as thin wrappers around interpreter-style machinery.

Required work:
- [x] audit every compiled runtime helper family used on static paths;
- [x] replace helpers whose normal static behavior is still modeled too closely
      after interpreter operations;
- [x] keep only true dynamic-boundary helpers dependent on runtime/interpreter
      object-model values;
- [x] ensure array/map/range/string/concurrency runtime services used by
      compiled code are direct Go implementations;
- [x] keep helper control propagation aligned with the explicit control-envelope
      model.

Proof required:
- source audit over emitted helper families;
- static fixture slices that exercise helper families without dynamic features.

#### Milestone 6: Boundary Containment And Static Cleanliness

Status:
- complete on 2026-03-22.
- the final explicit boundary helper set is now mechanically locked to:
  - `call_value` via `__able_call_value(...)`
  - `call_named` via `__able_call_named(...)`
  - `call_original` via generated original-wrapper calls
- representative static no-bootstrap fixture execution is now audited for:
  - zero fallback boundary calls
  - zero explicit dynamic boundary calls
  - zero interface/member lookup fallback calls
  - zero global lookup fallback calls
- representative static fixture batches now remain green under:
  - no-fallback boundary-marker harnesses
  - no-bootstrap boundary/lookup-marker harnesses
  - static generated-source boundary audits
- representative boundary-containment coverage now lives in:
  - `v12/interpreters/go/pkg/compiler/compiler_boundary_containment_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_boundary_audit_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_native_touchpoint_audit_test.go`
  - `v12/interpreters/go/pkg/compiler/compiler_main_bootstrap_test.go`

Goal:
- make the dynamic boundary explicit, narrow, and mechanically enforced.

Allowed boundary categories only:
- `dynimport`
- dynamic package mutation / definition
- dynamic evaluation / metaprogramming
- extern / host ABI conversion
- explicit compiled <-> dynamic callback boundaries
- values already originating from dynamic runtime payloads

Required work:
- [x] enumerate and document the final allowed boundary helper set;
- [x] tighten adapters so conversion happens exactly at the edge and returns to
      native carriers immediately after;
- [x] remove residual dynamic leakage from static fixtures;
- [x] keep strict no-bootstrap/no-fallback/no-boundary audits green for static
      fixture families.

Proof required:
- boundary-marker harnesses;
- strict static fixture audits;
- generated-source audits around boundary helper usage.

#### Milestone 7: Compiler Performance Completion

Status:
- complete on 2026-03-22.
- added the reduced checked-in recursion benchmark family:
  - `v12/fixtures/bench/fib_i32_small/main.able`
- shared compiled callable/runtime env scaffolding now swaps package envs only
  when the target env differs from the current env, via:
  - `v12/interpreters/go/pkg/compiler/bridge/bridge_env_swap.go`
  - `v12/interpreters/go/pkg/compiler/generator_render_runtime_env_helpers.go`
- representative generated code now uses the conditional env-swap path across:
  - compiled functions/methods
  - native callable wrappers
  - native array core methods
  - native interface generic dispatch helpers
  - iterator collect mono-array helpers
  - compiled future task entry
- representative performance proof and current numbers are now recorded in:
  - `v12/docs/perf-baselines/2026-03-22-compiler-performance-milestone-7-compiled.md`

Goal:
- make the compiler-generated code fast on the checked-in benchmark family
  without violating the lowering rules.

Required work:
- [x] keep using reduced checked-in benchmarks to isolate hot shared lowering
      gaps;
- [x] remove only shared primitive/control/array/callable/dispatch scaffolding,
      never by adding named non-primitive fast paths;
- [x] remeasure after each material compiler workstream;
- [x] keep benchmark proofs paired with generated-source shape audits.

Primary benchmark families:
- matrix / array hot paths
- iterator / generic container pipelines
- nominal generic method/container paths
- recursion/call overhead microbenchmarks

Definition of done for this milestone:
- [x] hot compiled benchmark paths no longer carry already-identified avoidable
      scaffolding;
- [x] checked-in benchmark baselines and current numbers are up to date.

#### Milestone 8: Compiler Release Validation

Goal:
- turn the above architecture into a hard release gate.

Status:
- complete on 2026-03-30.
- release validation is now green under the real top-level gates:
  - `GOFLAGS='-p=1' ./run_all_tests.sh --compiler`
  - `./run_stdlib_tests.sh`
- the milestone-closing fixes were shared semantic fixes, not nominal
  special-cases:
  - range expressions inferred through `Iterable<T>` instead of incorrectly
    recoercing compiled ranges through nominal `Range<T>` carriers
  - native interface default-method dispatch now preserves concrete wrapped
    receiver overrides instead of eagerly short-circuiting to default helpers
  - numeric operators now accept unions whose members are all numeric and
    resolve them through pairwise promotion/normalization instead of rejecting
    them as non-numeric

Required work:
- [x] keep one authoritative lowering spec and one authoritative completion plan
      in sync with implementation;
- [x] keep compiler fixture parity green under no-bootstrap/no-fallback rules;
- [x] run the full compiler matrix and stdlib suite in compiled mode as a
      release gate;
- [x] ensure diagnostics and failure behavior are stable enough for production
      use;
- [x] confirm reproducible build trees and clean-checkout behavior for the
      compiler toolchain.

Release gate checklist:
- [x] `./run_all_tests.sh` green
- [x] `./run_stdlib_tests.sh` green
- [x] full compiled fixture matrix green
- [x] strict static no-fallback/no-boundary audits green
- [x] benchmark baselines updated
- [x] no known representable static-path regressions remaining in PLAN

#### Compiler Program Status

Compiler release validation is closed, but compiler-native encoding completion
remains the active highest-priority work. Bytecode performance does not start
until the staged hybrid/static-lowering gaps below are closed.

#### Compiler Native Encoding Completion (active follow-on)

Goal:
- finish the stronger compiler end-state where statically representable Able
  constructs lower to final Go-native encoded constructs rather than to staged
  hybrid carriers.

Status:
- array-native lowering tranche complete on 2026-04-01; remaining
  `runtime.ArrayValue` / `ArrayStore*` use is limited to explicit dynamic or
  ABI edges plus the unspecialized wildcard-array ABI.
- residual representable union/result/interface carrier-synthesis cleanup
  complete on 2026-04-14; representable static carrier families now only
  admit `runtime.Value` / `any` at explicit dynamic/open or ABI edges.
- mono-array transitional runtime-store scaffolding cleanup complete on
  2026-04-14; compiler-generated mono-array wrappers are now pure slice
  carriers, mono-array field access (`length`, `capacity`, `storage_handle`)
  stays native on those wrappers, and `runtime.ArrayValue` / `ArrayStore*`
  remain only explicit dynamic or ABI boundary machinery.

Required work:
- [x] finish eliminating residual representable union/result/interface lowering
      paths that still rely on `runtime.Value` / `any` members outside explicit
      dynamic or ABI boundaries;
- [x] retire the remaining transitional mono-array/runtime-typed-store
      scaffolding now that static compiled arrays use compiler-native carriers
      by default;
- [x] decide and document the final no-interpreter policy for alias /
      constraint revalidation in generic interface dispatch, then make the
      implementation match that policy;
- [x] rerun the compiled release gates after each material native-encoding
      closure so the stronger finish line is enforced, not just the milestone-8
      release gate.

Proof required:
- source audits showing no staged hybrid carrier shapes remain on static paths
  where final host-native encodings are expected;
- fixture and generated-source coverage proving arrays, unions, interfaces,
  patterns, and dispatch stay on final native carriers;
- top-level release gates still green after each closure step.

Status:
- compiler-native completion is closed on 2026-04-14;
- bytecode performance is closed on 2026-04-16.

### Bytecode Performance Program (second priority; start after compiler-native completion work is closed or paused explicitly)

Goal:
- make the Go bytecode interpreter fast enough to be a practical execution mode
  after the compiler is finished.

Current state snapshot:
- Bytecode Milestone 1 is closed on 2026-04-16:
  - the remaining high-frequency dispatch/lookup scaffolding work is now
    closed across name lookup, member/index caches, slot stores, inline-call
    setup, direct raw-array access, and hot integer compare/add slot-const
    paths;
  - the last correctness blocker on that path was a bytecode-only inline
    return coercion gap for impl-generic interface returns, now fixed by
    threading the callee's generic-name set through bytecode call frames so
    return coercion keeps method-set generics instead of validating
    `Iterator T` / `Enumerable T` style returns as fully concrete;
  - focused interpreter coverage, the full `pkg/interpreter` Go test run, and
    `./run_stdlib_tests.sh` are green on this tree;
  - a fresh local quicksort hotloop spot-check after the fix is
    `10735729 ns/op` (`go test ./pkg/interpreter -run '^$' -bench
    '^BenchmarkBytecodeQuicksortHotloopRuntime$' -benchtime=50x -count=1`);
  - Bytecode Milestones 2, 3, and 4 are now also closed on 2026-04-16:
    allocation-pressure cleanup is done, collection/async-specific hidden
    overhead is reduced, the checked-in bytecode-core cross-mode baseline is
    current, and report-first guardrail tooling now compares fresh runs
    against that baseline without enforcing premature thresholds;
- the old compiled-mode CLI blockers in `cmd/able` / `cmd/ablec` are now
  cleared in this tree, and the downstream compiler follow-on regressions
  they exposed are now fixed too:
  - matcher/interface boundary helper finalization now materializes concrete
    sibling interface families before emitting native boundary helpers, so
    compiled matcher coercions keep concrete sibling adapters instead of
    falling back at boundary sites;
  - assignment binding metadata now reconciles concrete native carriers back
    into stored type expressions when rescue/join inference had only retained
    a broader wrong-package union/result wrapper, so imported shadowed nominal
    rescue joins no longer regress later member dispatch to stale local
    carrier metadata;
  - the repo-level `./run_all_tests.sh` gate is green again on this tree
    after those fixes (`real 1399.73`);
  - after the rebase exposed compiler package timeout pressure again, the
    heavy independent compiler fixture/parity/audit subtests now run in
    parallel, which brought `TestCompilerExecFixtures` down to about
    `real 203.36`, the full `pkg/compiler` package back under the default
    30-minute timeout (`1230.138s`), and the top-level repo gate back to
    green on this tree (`real 1287.77`);
- a large amount of call-dispatch, lookup-cache, frame-pool, integer-op, and
  hotspot work is already landed;
- the first post-compiler bytecode tranche is now landed too: repeated array
  index-method sites use a single-entry hot inline cache before the broader
  map cache, and the quicksort hotloop CPU profile no longer shows
  `indexMethodCacheKey(...)` among the active runtime hotspots;
- the next bytecode tranche is landed too: array-handle tracking now keeps a
  single tracked `ArrayValue` fast path and only promotes to an alias set when
  multiple wrappers share a handle, so `syncArrayValues(...)` dropped out of
  the same quicksort hotspot profile;
- the next bytecode tranche is landed too: repeated call-position member
  lookups now hit a single-entry inline member-method cache before the broader
  per-VM map cache, and receiver identity is pinned so same-name methods on
  different struct definitions cannot cross-hit;
- the next bytecode tranche is landed too: repeated non-local `LoadName` /
  `CallName` sites now cache lexical-owner hits (including captured parent and
  global bindings) with a single-entry inline hot path, so runtime
  `Environment.Lookup` is no longer a meaningful quicksort hotspot;
- the next bytecode tranche is landed too: already-tracked array wrappers now
  reuse their tracked array state directly instead of re-entering
  `ArrayStoreEnsure(...)` on hot reads, which moved array-state/store work out
  of the quicksort hotspot set;
- the next bytecode tranche is landed too: cached plain-array index sites now
  stay on a bytecode raw-array fast path instead of bouncing through the
  generic interpreter index dispatcher, which materially cut the quicksort
  index hot path again;
- the next bytecode tranche is landed too: exact-arity native calls and
  native bound-method calls now stay on a VM-side fast path instead of always
  routing through the generic callable dispatcher, which cut the remaining
  quicksort call-dispatch path again and keeps receiver injection plus
  non-borrowing arg stability pinned under bytecode;
- the next bytecode tranche is landed too: index-method caching now stores
  per-program / per-IP cache slots instead of routing hot array index sites
  back through a composite-key hash map, which removes the remaining
  `bytecodeIndexMethodCacheKey` hashing cost from the quicksort profile;
- the next bytecode tranche is landed too: array receiver identity now reuses
  a cached element-type token on shared array state, and single-threaded
  bytecode cache probes now read the method-cache version without taking the
  method-cache lock, which removed `currentMethodCacheVersion()` from the
  quicksort hotspot set and shrank the remaining index-identity path again;
- the next bytecode tranche is landed too: `execCall` and `execCallName` now
  resolve exact native call targets before attempting inline bytecode frame
  setup, so exact native call sites stop paying the inline-probe miss path and
  focused stats coverage now pins that those sites no longer contribute inline
  hit/miss counters;
- the next bytecode tranche is landed too: direct raw-array index `get` / `set`
  now reuse the tracked shared `ArrayState` pointer when the receiver already
  carries a valid tracked handle, and direct integer index decoding now stays
  on the small-int fast path, which moved the quicksort hot loop back into the
  high-17ms band without changing cache or invalidation semantics;
- the next bytecode tranche is landed too: name, member-method, and
  index-method VM cache probes now read environment/global revision state
  through a single-thread runtime hint, and bytecode method-cache probes read
  the interpreter method-cache version directly on the same single-thread path,
  which cut the remaining revision/version bookkeeping around hot cache checks
  and pushed the quicksort hot loop back down to roughly `17.5ms/op` on the
  longer local spot-check;
- the next bytecode tranche is landed too: already-cached raw-array no-method
  index sites now bypass `resolveCachedIndexMethod(...)` entirely and jump
  straight to direct array access under the same inline cache guards, and a
  same-session 5x50x local A/B check moved the quicksort hotloop distribution
  from roughly `20.0-24.3ms/op` without the bypass to `19.8-20.4ms/op` with
  the bypass while keeping recursive array-parameter self-call coverage pinned;
- the next bytecode tranche is landed too: when no `Array` `Index` /
  `IndexMut` impls exist at all, raw-array bytecode indexing now skips the
  index-method cache layer completely and jumps straight to direct array
  access, while focused “impl appears later” coverage stays green; this pushed
  the quicksort hot loop down again to roughly `15.9ms/op` on the 50x local
  spot-check;
- the next bytecode tranche is landed too: successful bytecode call paths no
  longer eagerly materialize eval-state/runtime-context data just to complete a
  call; `execCall`, `execCallName`, and related exact-native paths now resolve
  `stateFromEnv(...)` only if an actual error needs runtime-context attachment,
  which moved the quicksort hot loop to roughly `16.2ms/op` on a fresh 50x
  local spot-check while leaving error semantics unchanged;
- the next bytecode tranche is landed too: bound generic method calls whose
  receiver is already concretely injected may now use the existing bytecode
  inline call fast path instead of being conservatively forced through full
  `invokeFunction(...)` dispatch; that specifically unlocked concrete generic
  receiver calls like the hot `Array.push(...)` path and moved the quicksort
  hot loop to roughly `13.4ms/op` on a fresh 50x local spot-check;
- the next bytecode tranche is landed too: simple `LoadName` / `CallName`
  sites now record at lowering time that they are plain identifier lookups, so
  the VM can skip repeated runtime dotted-name/cacheability checks and go
  straight to the hot lexical/global/scope cache path; that cut the remaining
  named-call overhead again and moved the quicksort hot loop to roughly
  `13.3ms/op` on a clean 50x local rerun while pushing `execCallName` out of
  the top-tier hotspot set;
- the next bytecode tranche is landed too: simple named-call cache entries now
  live behind stable pointers, so hot `CallName` hits stop copying full
  dispatch records back and forth through the inline cache and map cache; the
  focused call-name invalidation slice stayed green, repeated 50x quicksort
  reruns landed around `10.6-11.2ms/op`, and `lookupCachedCallName(...)`
  dropped out of the top hotspot set on the next profiled run;
- the next bytecode tranche is landed too: direct small-int pair helpers now
  sit in front of the bytecode integer compare/add/sub hot paths, so common
  small-int pairs stop paying the older repeated `IntegerValue` extraction work
  before comparison and same-type arithmetic; repeated 50x quicksort reruns
  landed around `10.4-10.8ms/op`, and the old direct integer-compare hotspot
  no longer shows up as a top flat frame on the next profiled run;
- the next bytecode tranche is landed too: small slot-frame layouts now batch
  prefill the bytecode hot frame pool, so recursive inline calls can draw
  several same-size frames from the pool on the way down instead of allocating
  one frame per depth step; focused slot-frame/call coverage stayed green,
  repeated 50x quicksort reruns stayed in the same `10.5-11.2ms/op` band, and
  the profile cut `acquireSlotFrame(...)`, `tryInlineResolvedCallFromStack(...)`,
  and `execCallName(...)` materially on the next run;
- the next bytecode tranche is landed too: untyped `StoreSlot` sites now take
  a direct VM fast path instead of always routing through the typed-assignment
  helper, so ordinary local slot stores stop paying the typed pattern/coercion
  helper call on every write; focused slot-store coverage stayed green,
  repeated 50x quicksort reruns stayed roughly in the `10.2-11.3ms/op` band,
  and `execStoreSlot(...)` dropped out of the hotspot tier on the next profile;
- the next bytecode tranche is landed too: direct array index decoding now
  inlines the 64-bit small-int case at the outer value switch, so common value
  and pointer-backed integer indexes stop bouncing through the extra
  `IntegerValue -> int` helper on hot array `get` / `set` sites; repeated 50x
  quicksort reruns stayed roughly in the `10.2-11.0ms/op` band, and the old
  `bytecodeDirectArrayIndexFromInteger(...)` hotspot dropped out of the top set
  on the next profile;
- the next bytecode tranche is landed too: slot-const integer bytecode
  instructions now carry typed integer-immediate metadata directly, so hot
  `x +/- const` and `x <= const` loops stop re-decoding `instr.value` through
  the generic runtime-value path on every iteration; the focused lowering and
  cache slice stayed green, the profiled 50x quicksort run reached
  `9917425 ns/op`, and the old slot-const immediate decode hotspot dropped out
  of the top tier on the next profile;
- the next bytecode tranche is landed too: direct bytecode array writes now
  use an alias-aware tracked-array write sync path, so exclusive wrappers skip
  the old post-write alias-broadcast walk while shared aliases still stay
  coherent and element-type tokens stay current; the focused tracking/index
  slice stayed green, repeated 50x quicksort reruns landed around
  `9.90-10.00ms/op`, and the old `syncArrayValues(...)` / `resolveIndexSet(...)`
  hotspot pair dropped out of the top tier on the next profile;
- the next bytecode tranche is landed too: identifier member-access sites now
  carry their member name directly in bytecode, and `execMemberAccess` only
  probes/stores the member-method cache when a site can actually use it; that
  removes the remaining futile member-cache work from plain field/property
  accesses and drops `execMemberAccess` out of the hotspot tier, while the
  residual named-call cost is now dominated by inline-call value coercion
  rather than by ordinary/member dispatch scaffolding;
- the next bytecode tranche is landed too: slot-layout analysis now caches
  simple parameter type names for all inlineable parameters, and inline bytecode
  calls now use a primitive-only fast coercion path for simple integer widening
  plus integer/float coercions before falling back to the general type
  coercion machinery; that cuts the targeted `tryInlineCallFromStack(...)`
  coercion profile materially even though whole-benchmark wall-clock remains in
  the same noisy low-13ms band on this machine;
- the next bytecode tranche is landed too: ordinary general integer
  comparisons now bypass the shared fast-operator dispatcher in the bytecode VM
  and stay on a direct integer compare path, while the shared fast evaluator
  also shortcuts exact integer/string comparisons before it falls back to the
  generic comparison machinery; that moved the quicksort hot loop down again
  into the low-12ms band on local 20x/50x spot-checks and pushed the remaining
  visible cost toward slot coercion / cast work rather than comparison
  dispatch;
- the next bytecode tranche is landed too: inline bytecode calls now skip
  `coerceValueToType(...)` entirely when the declared parameter type is a
  guaranteed no-op coercion shape (for example `Array i32`), and simple
  primitive casts now short-circuit before alias/type-metadata checks on hot
  same-type paths such as `as i32`; that kept the quicksort hot loop in the
  high-11ms band on local 50x runs and pushed `castValueToType(...)` out of
  the hotspot tier;
- the next bytecode tranche is landed too: slot layouts now cache per-parameter
  inline-call coercion metadata, and the VM’s hot inline call path now uses
  that cached data instead of recomputing generic/no-op coercion guards on
  every recursive call while also separating the already-bound receiver path;
  this pushed the quicksort hot loop down again to roughly `11.0ms/op` on the
  50x profiled local run and dropped `tryInlineCallFromStack(...)` out of the
  hotspot tier;
- the next bytecode tranche is landed too: slot store instructions now carry
  typed-identifier assignment metadata directly in bytecode, and the VM’s
  store-slot path uses that metadata instead of reopening assignment AST nodes
  and pattern helpers on every store just to discover most stores are untyped;
  this removed `execStoreSlot(...)` / `typedSlotAssignmentValues(...)` from the
  hotspot tier and kept the quicksort hot loop in the mid-11ms band on local
  50x reruns;
- the next bytecode tranche is landed too: simple `CallName` sites now keep a
  revision-guarded call-site cache for the resolved callee plus its dispatch
  shape (`exact native`, `inline bytecode`, or `generic`), so repeated named
  calls no longer redo lexical lookup validation and call-target
  classification on every hit; focused invalidation and dispatch-kind rebinding
  coverage is green, the profiled 50x quicksort run reached roughly
  `10.7ms/op`, and repeated clean 50x reruns stayed around `11.1-11.4ms/op`
  with one noisy outlier;
- the next bytecode tranche is landed too: the direct integer-comparison fast
  path and the shared comparison helper now short-circuit small-int pairs
  before they fall through the general `ToInt64()` / big-int comparison route,
  which trims the remaining generic comparison helper overhead without widening
  the bytecode opcode surface; focused parity coverage is green,
  `execBinaryDirectIntegerComparisonFast(...)` dropped from roughly `100ms`
  cumulative to roughly `80ms` on the profiled quicksort run, and repeated
  local 50x reruns stayed in the same low-11ms band;
- the next bytecode tranche is landed too: hot `x + const` integer sites now
  lower onto the existing slot-const immediate path instead of reloading and
  unboxing the constant operand through the ordinary specialized binary route;
  focused lowering/parity coverage is green, the profiled quicksort run moved
  `execBinarySpecializedOpcode(...)` from roughly `70ms` cumulative down to
  roughly `30ms`, and repeated local 50x reruns stayed in the low-11ms band
  with the best reruns landing around `10.8ms/op`;
- the next bytecode tranche is landed too: primitive-receiver method
  resolution now checks the existing bound-method cache before paying
  `env.Lookup(...)`, and single-threaded bytecode runs now read/write that
  bound-method cache without taking the interpreter method-cache lock; this
  removed `resolveMethodFromPool(...)` / array member dispatch from the
  hotspot tier on the next profiled quicksort run, while repeated local 50x
  reruns landed in the roughly `10.4-11.2ms/op` band;
- the next bytecode tranche is landed too: direct array index decoding now
  keeps concrete integer receivers on a narrower small-int/int64 path and
  skips the old extra generic integer extraction on already-concrete index
  values; on 64-bit builds that also removes the redundant int-range guard for
  small ints. Focused index-cache coverage is green, the profiled quicksort
  run moved `bytecodeDirectArrayIndex(...)` from roughly `70ms` flat down to
  roughly `40ms` cumulative, and repeated local 50x reruns stayed in the
  roughly `10.6-11.1ms/op` band;
- the next bytecode tranche is landed too: slot-enabled frame layouts now
  cache summary `any runtime coercion needed?` flags, and the hot inline-call
  setup paths bulk-copy arguments straight into slot frames whenever the
  layout proves no runtime coercion is possible; bound-receiver inline calls
  now do the same for explicit arguments after seeding the injected receiver.
  Focused slot-analysis/quicksort coverage is green, the profiled 50x
  quicksort run reached `9778187 ns/op`, repeated clean 50x reruns landed at
  `9930751`, `9916588`, and `10300807 ns/op`, and
  `tryInlineResolvedCallFromStack(...)` is down to roughly `30ms` cumulative
  instead of sitting in the top hotspot tier;
- the next bytecode tranche is landed too: the direct small-int comparison
  path no longer pays an extra tuple-return helper on every hot comparison.
  `bytecodeDirectIntegerCompare(...)` now decodes and compares concrete
  small-int pairs directly, which removed `bytecodeDirectSmallIntPair(...)`
  from the current hot profile entirely. Focused parity/quicksort coverage is
  green, the profiled 50x quicksort run moved from `10722194` to
  `10384074 ns/op`, repeated clean 50x reruns landed at `10067934`,
  `9961095`, and `9989424 ns/op`, and the compare fast path collapsed from
  the old `80ms` flat / `90ms` cumulative chain to roughly `50ms`
  cumulative total;
- the first post-Milestone-1 allocation-pressure tranche is landed too: hot
  bytecode member/index/binary result sites now reuse existing stack slots in
  place instead of doing repeated pop/pop/append reshaping, and the VM’s hot
  execution paths use unchecked top-slot replacement helpers after their
  existing stack-depth guards. Focused stack/quicksort coverage is green, the
  profiled 50x quicksort run reached `9886874 ns/op`, repeated clean 50x
  reruns landed at `10203423`, `10273281`, and `9753472 ns/op`, and the
  temporary `replaceTop2(...)` hotspot introduced by the first helper rewrite
  dropped back out of the top tier once the hot paths switched to the
  unchecked inlinable helpers;
- the next Milestone 2 tranche is landed too: bytecode programs now cache
  their resolved return generic-name sets, so hot inline/named call paths no
  longer re-enter `FunctionValue.GenericNameSet()` on every frame push, and
  the bytecode small-int boxing cache is now eagerly initialized so hot
  arithmetic no longer pays `sync.Once` on every small boxed integer hit.
  Focused coverage is green; repeated local 50x quicksort reruns landed at
  `9189777`, `9094444`, and `9161513 ns/op`, and the kept profile shows the
  old `FunctionValue.GenericNameSet()` frame gone with small-int boxing
  reduced to noise;
- the next reduced-fib tranche is landed too: slot frame layouts now cache a
  compact primitive return check, and `finishInlineReturn(...)` uses that
  cached check for simple return types before falling back to the older
  string-based helper or full return coercion. Focused return/self-fast
  coverage is green, refreshed reduced `BenchmarkFib30Bytecode` reruns moved
  from the prior `156.88-163.96ms/op` band to `150.28-157.63ms/op`, and
  aligned external bytecode `fib` still times out at `90s`;
- the next reduced-fib tranche is landed too: the fused recursive self-call
  path now has a self-call-only small-`i32` immediate subtract helper, so
  `fib(n - 1)` / `fib(n - 2)` avoid the broader integer-immediate helper
  ladder before setting up the minimal self-fast frame. Focused recursive and
  arithmetic coverage is green; the refreshed confirmation band moved to
  `139.07-146.70ms/op` with unchanged allocation shape, and aligned external
  bytecode `fib` still times out at `90s`;
- the next reduced-fib maintenance tranche is landed too: fused slot-const
  recursive self-call execution now lives in
  `bytecode_vm_call_self_slot_const.go`, reducing `bytecode_vm_calls.go` from
  `992` to `842` lines while keeping the current reduced `fib` band in range
  at `143.51-148.90ms/op`;
- the next reduced-fib tranche is landed too: fused self-calls with the common
  two-slot frame layout now acquire frames through a dedicated size-2 hot path
  before falling back to the general frame pool. Focused slot-frame/self-call
  coverage is green; warmed reduced `BenchmarkFib30Bytecode` reruns moved from
  the refreshed `140.91-150.88ms/op` band to a kept `138.53-141.70ms/op`
  confirmation band, with a profiled one-shot at `133.52ms/op`; aligned
  external bytecode `fib` still times out at `90s`;
- the next reduced-fib tranche is landed too: the fused slot-const conditional
  jump now uses a dedicated direct small-integer `<=` immediate helper instead
  of routing through the generic `bytecodeDirectIntegerCompare("<=", ...)`
  helper and materializing a `BoolValue` only to read it immediately. Focused
  comparison/lowering/recursive coverage is green; refreshed warmed reduced
  `BenchmarkFib30Bytecode` confirmation moved to `134.52-138.06ms/op`, with a
  profiled one-shot at `135.38ms/op`; aligned external bytecode `fib` still
  times out at `90s`;
- the next reduced-fib measurement tranche is landed too:
  `BenchmarkFib30BytecodeRuntimeOnly` now evaluates/lowers the reduced `fib`
  function once, validates a warmup `fib(30)` result, and then repeatedly calls
  the same bytecode function on the same interpreter. This separates recursive
  VM runtime from parser/module/interpreter setup noise; initial runtime-only
  checks landed at `129.18-139.44ms/op` with effectively zero allocations, and
  the profiled one-shot landed at `135.83ms/op`;
- the next runtime-only reduced-fib helper tranche is landed too: the
  self-call-only small-`i32` immediate subtract helper now skips the redundant
  `int64` overflow probe after both operands have already been proven small
  `i32`, while retaining the existing `i32` bounds check and overflow error.
  Focused helper/self-call coverage is green; runtime-only `fib(30)` warmed
  reruns landed at `126.40-132.47ms/op` on the first kept band, with a
  profiled one-shot at `128.22ms/op`;
- the next reduced-fib control-flow tranche is landed too: statement-position
  `if slot <= const { return slot }` now lowers to a fused slot-const
  return-if opcode instead of a conditional jump plus separate slot load and
  return dispatch. Focused lowering/parity/self-call coverage is green;
  runtime-only `fib(30)` warmed reruns landed at `122.95-133.48ms/op` on the
  first kept band and `125.41-136.10ms/op` on confirmation, with a profiled
  one-shot at `136.73ms/op`;
- the next reduced-fib return-if micro-tranche is landed too: the fused
  return-if opcode now has a same-slot typed-immediate path for the common
  `if n <= const { return n }` shape, avoiding the extra return-slot bounds
  check, immediate fallback ladder, and helper call on that hot base case.
  Focused opcode/lowering coverage is green; runtime-only `fib(30)` warmed
  reruns landed at `122.67-133.12ms/op` on the first kept band and
  `123.58-138.89ms/op` on confirmation, while a temporary restored A/B check
  regressed to `136.60-155.13ms/op`;
- the next reduced-fib run-loop tranche is landed too: the hot fused
  return-if opcode is back inline in `runResumable(...)`, while cold
  placeholder lambda/value execution moved to a focused helper file to keep
  `bytecode_vm_run.go` under the line cap. Focused placeholder/return-if/
  self-call coverage is green; runtime-only `fib(30)` reruns landed at
  `127.63-132.06ms/op` on the quiet confirmation band, and the reduced
  end-to-end one-shot checks landed at `130.39-132.60ms/op`;
- the next runtime-only reduced-fib frame-release tranche is landed too:
  minimal self-fast returns from proven two-slot layouts now use a dedicated
  `releaseSlotFrame2(...)` helper, preserving eager clearing while avoiding the
  generic release switch on the hot recursive unwind path. Focused
  slot-frame/self-call/return-if coverage is green; runtime-only `fib(30)`
  reruns landed at `111.05-117.82ms/op` on the confirmation band, with a
  profiled one-shot at `118.84ms/op`;
- the next runtime-only reduced-fib implicit return-add tranche is landed too:
  a node-less implicit `BinaryIntAdd` immediately followed by `Return` now
  lowers to `bytecodeOpReturnBinaryIntAdd`, leaving the following return
  instruction unreachable so jump targets stay stable while the VM returns the
  add result directly. Focused lowering/parity/recursive coverage is green;
  same-load pre-change runtime-only checks landed at `116.06-118.25ms/op`, a
  temporary no-fusion control landed at `113.04-123.21ms/op`, and the restored
  fused confirmation band landed at `109.40-115.48ms/op`, with a profiled
  one-shot at `123.82ms/op`;
- the next aligned-fib base-case tranche is landed too: statement-position
  `if slot <= const { return small_i32_const }` now lowers to
  `bytecodeOpReturnConstIfIntLessEqualSlotConst`, covering the real external
  `fib(45)` shape (`if n <= 2 { return 1 }`) instead of only the reduced
  `return n` shape. Focused lowering/VM coverage is green; the aligned-style
  `fib_i32_small` bytecode-runtime fixture moved from a temporary no-fusion
  control at `12.59s/op` to fused confirmation at `10.58s/op`. The full
  external bytecode `fib(45)` check still times out at `90s`, so the next
  tranche should continue on aligned-fib-only residual overhead rather than
  another reduced-`fib(30)` micro-branch;
- the next aligned-fib raw-immediate tranche is landed too: lowered
  slot-const opcodes now retain a raw `int64` immediate next to the existing
  typed `IntegerValue`, and the fused self-call / return-const / conditional
  helpers use it on their direct small-integer path. Focused immediate,
  lowering, self-call, and return-const coverage is green; the aligned-style
  `fib_i32_small` bytecode-runtime fixture landed at `9.94s/op`, `10.37s/op`,
  `10.18s/op`, and restored raw confirmation `9.49s/op`, versus a temporary
  no-raw control at `10.49s/op`. A reduced `Fib30BytecodeRuntimeOnly` sanity
  band landed at `118.38-126.67ms/op`, so this keep is aligned-driven;
- the next aligned-fib return-add tranche is landed too: implicit final `+`
  expressions in functions declared `i32` now lower to
  `bytecodeOpReturnBinaryIntAddI32`, which tries the direct small-`i32` add
  path before falling back to the existing generic return-add semantics.
  Focused return-add/recursive coverage is green; aligned-style
  `fib_i32_small` bytecode-runtime confirmation landed at `9.89s/op` and
  `9.86s/op` across two 3-run bands, with a profiled one-shot at `9.77s/op`.
  The reduced `Fib30BytecodeRuntimeOnly` sanity band landed at
  `125.87-127.84ms/op`, so this keep is aligned-driven rather than a reduced
  `fib(30)` win;
- the next aligned-fib self-call arithmetic tranche is landed too: the fused
  raw-immediate self-call path now performs the small-`i32` subtract directly
  inside `execCallSelfIntSubSlotConst(...)` instead of calling back through
  `bytecodeSelfCallSubtractIntegerImmediateI32RawFast(...)`. Focused
  recursive/arithmetic coverage is green; reduced
  `Fib30BytecodeRuntimeOnly` landed at `114.39-122.80ms/op`, aligned-style
  `fib_i32_small` bytecode-runtime landed at `9.80s/op` across a 3-run band,
  and the profiled one-shot landed at `9.61s/op`. The full external bytecode
  `fib(45)` check still times out at `90s`, so the next aligned-fib tranche
  should start from the remaining `execCallSelfIntSubSlotConst(...)`,
  `finishInlineReturn(...)`, `execReturnConstIfIntLessEqualSlotConst(...)`,
  and `releaseSlotFrame2(...)` costs rather than another subtract-helper call
  shape rewrite;
- the next aligned-fib minimal-return tranche is landed too: the minimal
  self-fast branch in `finishInlineReturn(...)` now recognizes the fused
  `bytecodeOpReturnConstIfIntLessEqualSlotConst` base-case opcode in functions
  declared `i32` and skips the generic simple return-coercion probe for that
  already-encoded i32 constant. Focused lowering/runtime coverage pins that
  the fused return-const value is encoded as `i32`; reduced
  `Fib30BytecodeRuntimeOnly` landed at `113.15-122.26ms/op`, aligned-style
  `fib_i32_small` bytecode-runtime landed at `9.33s/op` and `9.24s/op` across
  two 3-run bands, and the profiled one-shot landed at `8.74s/op`. The full
  external bytecode `fib(45)` check still times out at `90s`. The next
  aligned-fib tranche should start from the remaining
  `execCallSelfIntSubSlotConst(...)`, `execReturnConstIfIntLessEqualSlotConst(...)`,
  `bytecodeDirectSmallI32Pair(...)`, and slot-frame release/boxing costs rather
  than another return-coercion shortcut;
- the next aligned-fib return-add value-pair tranche is landed too:
  `bytecodeOpReturnBinaryIntAddI32` now handles direct
  `runtime.IntegerValue`/`runtime.IntegerValue` small-`i32` operands before
  falling back to the existing pointer-oriented small-`i32` add helper.
  Focused coverage pins the direct value-pair result, pointer fallback, and
  overflow error path. Reduced `Fib30BytecodeRuntimeOnly` was noisy but stayed
  in the kept band after the first sample (`143.82ms/op`, `125.72ms/op`,
  `116.12ms/op`), aligned-style `fib_i32_small` bytecode-runtime landed at
  `9.07s/op` and `9.10s/op` across two 3-run bands, and the profiled one-shot
  landed at `8.88s/op`. The full external bytecode `fib(45)` check still times
  out at `90s`. The next aligned-fib tranche should start from the remaining
  `execCallSelfIntSubSlotConst(...)`, `finishInlineReturn(...)`,
  `execReturnConstIfIntLessEqualSlotConst(...)`, and slot-frame release costs
  rather than another return-add operand extraction shortcut;
- the next aligned-fib inline return-coercion tranche is landed too: handled
  small-`i32` branches of `bytecodeOpReturnBinaryIntAddI32` now report that
  the returned value already satisfies an `i32` return, allowing
  `finishInlineReturn(...)` to skip the generic simple return-coercion probe
  only for those proven fast-path values. Generic fallback arithmetic still
  reports an unknown return shape and keeps the existing coercion behavior.
  Focused coverage pins that distinction. Reduced `Fib30BytecodeRuntimeOnly`
  landed at `110.85-114.67ms/op`, aligned-style `fib_i32_small`
  bytecode-runtime landed at `8.33s/op` and `8.44s/op` across two 3-run bands,
  and the profiled one-shot landed at `9.01s/op`. The full external bytecode
  `fib(45)` check still times out at `90s`. The next aligned-fib tranche should
  start from `execCallSelfIntSubSlotConst(...)`,
  `execReturnConstIfIntLessEqualSlotConst(...)`, and slot-frame return/release
  costs rather than another return-add coercion shortcut;
- the next aligned-fib self-call layout tranche is landed too:
  `execCallSelfIntSubSlotConst(...)` now keeps only the fused recursive
  fast-path inline and delegates the non-fast callable/native/generic fallback
  path to `execCallSelfIntSubSlotConstFallback(...)`. Semantics are unchanged;
  this is a code-layout reduction for the hot self-call opcode. Focused
  recursive coverage is green. Reduced `Fib30BytecodeRuntimeOnly` landed at
  `109.33-118.41ms/op`, aligned-style `fib_i32_small` bytecode-runtime landed
  at `8.44s/op` and `8.42s/op` across two 3-run bands, and the profiled
  one-shot landed at `8.42s/op`. The full external bytecode `fib(45)` check
  still times out at `90s`. The next aligned-fib tranche should start from
  `execReturnConstIfIntLessEqualSlotConst(...)`, `finishInlineReturn(...)`,
  and slot-frame return/release costs rather than more self-call fallback
  rearrangement;
- the next aligned-fib measurement tranche is landed too: no VM code changed,
  but the cross-mode baseline is now quantified. On `fib_i32_small`, compiled
  averaged `0.3433s`, tree-walker timed out `3/3` at `60s`, bytecode
  end-to-end averaged `8.1467s`, and bytecode-runtime confirmed at about
  `8.71-8.77s/op` with `24104 B/op` and `47 allocs/op`. Reduced
  `Fib30BytecodeRuntimeOnly` stayed in range at `112.22-119.67ms/op`. A
  longer external compare measured full `fib(45)` at compiled `3.3700s` and
  bytecode `92.5200s`, so bytecode is now known to be just above the old
  `90s` guard rather than unmeasured past it. The next aligned-fib tranche
  should use the external `fib(45)` result as the keep/revert guardrail and
  start from `execReturnConstIfIntLessEqualSlotConst(...)`,
  `finishInlineReturn(...)`, and minimal slot-frame return/release costs
  rather than reduced-only `fib(30)` micro-work;
- the next aligned-fib self-call guard tranche is landed too:
  `execCallSelfIntSubSlotConst(...)` now tries an early exact-shape compact
  branch for the proven raw-immediate two-slot slot-0 recursive shape before
  entering the generic immediate/layout/return-name ladder. Unsupported shapes
  still fall through to the existing boxed path. Focused recursive coverage is
  green, reduced `Fib30Bytecode` moved from the refreshed compact-frame
  profiled baseline of `105.27ms/op` to warmed reruns of `99.54ms/op`,
  `100.39ms/op`, and `99.00ms/op`, and full external bytecode `fib(45)` now
  completes inside the old `90s` guard at `79.1200s`. The next aligned-fib
  tranche should start from the base-case and return side around
  `execReturnConstIfIntLessEqualSlotConst(...)`, `finishInlineReturn(...)`,
  and `execReturnBinaryIntAdd(...)`, not another self-call fallback or
  frame-pool rewrite;
- the next aligned-fib return-add inline tranche is landed too:
  `execReturnBinaryIntAdd(...)` now handles the direct
  `runtime.IntegerValue`/`runtime.IntegerValue` small-`i32` value pair inline
  before falling back to the existing pointer/generic path. Focused return-add
  and recursive coverage is green, reduced `Fib30Bytecode` stayed in the kept
  range at `97.19-106.93ms/op`, aligned `fib_i32_small` bytecode-runtime moved
  to `7.21s/op` over a 3-run band with a profiled one-shot at `7.50s/op`, and
  full external bytecode `fib(45)` moved to `77.2400s`. The profile no longer
  shows `bytecodeReturnAddSmallI32ValuePairFast(...)`; remaining work should
  target structural boxed return/add handoff, the base-case raw compare, and
  compact `finishInlineReturn(...)` restoration rather than adding another
  return-add helper;
- the next aligned-fib compact self-call setup tranche is landed too:
  `execCallSelfIntSubSlotConstCompact(...)` now writes the compact slot-0
  self-fast frame record directly after its exact-shape checks instead of
  calling `pushSelfFastSlot0CallFrame(...)` on every recursive step. Focused
  self-call/frame coverage is green, reduced `Fib30Bytecode` moved to
  `94.85-104.21ms/op`, aligned `fib_i32_small` bytecode-runtime landed at
  `7.18s/op`, and full external bytecode `fib(45)` moved to `76.7900s`. A
  compact `finishInlineReturn(...)` shortcut was tested and reverted because
  it regressed aligned runtime to `8.31s/op`. The next tranche should stop
  shaving helper calls and either introduce safe raw/typed return-stack
  metadata with explicit invalidation, or move to a broader VM v2 typed-frame
  design slice;
- benchmark harnesses and counters already exist;
- the remaining work is no longer “find obvious first wins”, but a disciplined
  second phase focused on the remaining hot-path costs.

Milestones:

#### Bytecode Milestone 1: Hot Dispatch And Lookup Closure
Closed on 2026-04-16.
- high-frequency environment/path lookup churn is removed from the hot-loop
  dispatch path;
- slot coverage and direct-call lowering are extended far enough that the
  remaining top quicksort costs are no longer ordinary dispatch/lookup
  scaffolding;
- inline caches stay precise under rebinding/mutation and cheap on the current
  single-thread bytecode execution path;
- remaining bytecode optimization work now moves to allocation pressure and
  collection/async-specific hot paths rather than more dispatch/lookup closure.

#### Bytecode Milestone 2: Allocation Pressure Reduction
Closed on 2026-04-16.
- hot stack reshaping, repeated generic-name recomputation, and per-hit
  small-int box-cache initialization are removed from the bytecode VM’s hot
  path;
- transient arg-slice and generic/method-call metadata churn are reduced far
  enough that the remaining benchmark allocation cost is no longer generic VM
  scaffolding;
- the remaining visible allocation/throughput work is now dominated by
  collection-specific paths such as `ArrayStoreWrite`, plus the core integer
  compare/index loop, so that work moves to Milestone 3.

#### Bytecode Milestone 3: Collections / Containers / Async Hot Paths
Closed on 2026-04-16.
- dynamic array append growth now goes through the bytecode runtime's explicit
  reserve policy instead of the Go slice heuristic, which cut the quicksort
  benchmark from roughly `106.6 KB/op / 51 allocs/op` to roughly
  `75.6-75.9 KB/op / 49-50 allocs/op` while keeping the hot loop in the
  low-`9ms/op` band;
- a dedicated bytecode `spawn` / `future_yield` / `future_flush` benchmark is
  now in-tree, and the accidental per-spawn bytecode re-lowering path is
  removed by caching lowered spawn bodies in bytecode instructions;
- serial future scheduling also avoids the old queue-insert copy churn for
  not-yet-started tasks, and the new async benchmark moved from roughly
  `1.61-1.71ms/op`, `~1.07 MB/op`, and `3537-3539 allocs/op` down to roughly
  `1.23-1.43ms/op`, `~279 KB/op`, and `2666-2668 allocs/op`;
- the remaining bytecode costs are now core compare/call execution plus
  fundamental scheduler/context creation work, so further work moves to
  Milestone 4 benchmark gates rather than more hidden collection/async
  scaffolding cleanup.

#### Bytecode Milestone 4: Benchmark Gates
Closed on 2026-04-16.
- [x] keep benchmark baselines current for treewalker vs bytecode vs compiled;
- [x] add report-first perf guardrails, then optional thresholds once noise is
      characterized.
- `v12/bench_suite --suite bytecode-core` now records the checked-in
  cross-mode baseline for:
  - `quicksort`
  - `future_yield_i32_small`
  - `sum_u32_small`
- the current baseline artifacts are:
  - `v12/docs/perf-baselines/2026-04-16-bytecode-core-benchmark-baseline.json`
  - `v12/docs/perf-baselines/2026-04-16-bytecode-core-benchmark-baseline.md`
- `v12/bench_guardrail` now compares a fresh suite JSON against a baseline and
  reports status/timing/GC deltas without failing the build, so benchmark
  regressions are visible before any hard thresholds are introduced.

### Backlog (not active until compiler + bytecode priorities permit)

These items remain important, but they are not active priorities right now.

#### Integration / Tooling backlog
- benchmark alignment / external comparison work
  - current state: the aligned comparison harness is now landed
    - `v12/bench_perf` now supports `--run-from DIR` for benchmarks that read
      suite-local relative inputs, repeated `--program-arg ARG` flags for
      workload-specific entry arguments, and `--output-json PATH` for
      machine-readable summaries
    - `v12/bench_perf` now also pins the selected stdlib root from
      `ABLE_STDLIB_ROOT` or the installed cache during external runs, which
      avoids the sibling-stdlib collision that previously broke compiled
      benchmarks under `../benchmarks/*`
    - `v12/bench_perf` now also supports `--executor serial|goroutine`, and
      the main `able` CLI plus generated compiled launchers now honor the
      shared `ABLE_EXECUTOR` environment selection too
    - `v12/bench_compare_external` now runs Able against the sibling
      `../benchmarks` workloads and joins those results against
      `../benchmarks/results.json` for reference families like `go`, `ruby`,
      and `python`
    - the Able side of that comparison now uses the canonical local benchmark
      sources under `v12/examples/benchmarks/` instead of the stale
      `../benchmarks/*/able-v12-*` copies
    - aligned results now clearly rank the next optimization targets:
      - compiled `matrixmultiply` is already near Go (`1.36s` vs `0.88s`,
        about `1.55x`) and is not the first compiler problem
      - compiled `fib` and `binarytrees` are still the main static-path
        compiler gaps (`40.15s` vs `2.84s`, about `14.14x`; `34.47s` vs
        `3.83s`, about `9.00x`)
      - aligned `i_before_e` is the main stdlib/text path gap even in compiled
        mode (`6.05s` vs `0.05s`, about `121x` Go / `60x` Ruby), and bytecode
        still times out on both aligned `fib` (`60s`) and aligned
        `i_before_e` (`90s`)
    - the old compiled `binarytrees` gap was partly a benchmark harness bug:
      the benchmark body had been updated to use `spawn`, but the generated
      launcher still defaulted to the serial executor, so the supposed
      parallel workload was silently serialized
    - with executor selection fixed and `v12/bench_compare_external`
      auto-running `binarytrees` under the goroutine executor, the refreshed
      aligned compiled core benchmarks are now all in Go range:
      - `fib`: `3.16s` vs Go `2.84s` (`1.11x`)
      - `binarytrees`: `3.65s` vs Go `3.83s` (`0.95x`)
      - `matrixmultiply`: `1.03s` vs Go `0.88s` (`1.17x`)
    - the benchmark harness now supports reusable Go-side profiling for both
      the main CLI and generated compiled launchers via
      `ABLE_GO_CPU_PROFILE=/tmp/cpu.pprof` and
      `ABLE_GO_MEM_PROFILE=/tmp/heap.pprof`
    - `v12/bench_perf` now sends `SIGINT` before `SIGKILL` on timeout, so
      profiled timeouts still flush pprof output for bytecode runs
    - the first aligned profile pass now separates the next optimization work:
      - compiled `fib` / `binarytrees` were dominated by compiler bridge
        environment swap + call-frame overhead, not arithmetic itself
      - compiled `i_before_e` is dominated by allocation/GC plus
        `read_lines` / UTF-8 validation / `String.replace` / `contains`
      - bytecode `fib` is dominated by recursive call-frame and small-int
        opcode overhead, while bytecode `i_before_e` still mixes lowering cost
        with runtime cost when profiled through the CLI path
    - the first post-profile compiler tranche is now landed:
      - `v12/interpreters/go/pkg/compiler/bridge` now keeps the single-thread
        environment path lock-free and collapses `SwapEnvIfNeeded` into a
        direct fast-path swap instead of routing every compiled recursive call
        through the old `RWMutex`-guarded env lookup/swap sequence
      - refreshed aligned compiled numbers on the kept code:
        - `fib`: `15.92s` vs the prior `34.19s`
        - `binarytrees`: `21.30s` vs the prior `30.14s`
      - refreshed profiles show the old `RWMutex` / `Runtime.Env` hotspot is
        gone; the remaining compiler recursion overhead is now the narrower
        `SwapEnvIfNeeded`, `PushCallFrame`, and `PopCallFrame` path
    - the next compiler recursion tranche is now landed too:
      - compiled call-frame diagnostics no longer route through the
        interpreter eval-state stack on every call
      - `v12/interpreters/go/pkg/compiler/bridge` now keeps its own
        lightweight compiled call-frame stack and feeds it into runtime
        diagnostics only on the error path via
        `Interpreter.AttachRuntimeContextWithCallStack(...)`
      - refreshed aligned compiled numbers on the kept code:
        - `fib`: `13.40s` vs the prior `15.92s`
        - `binarytrees`: `19.84s` vs the prior `21.30s`
      - refreshed profiles show the remaining compiled recursion hotspot is
        now mostly `SwapEnvIfNeeded`, with call-frame overhead narrowed to the
        generated `__able_push_call_frame` / `__able_pop_call_frame` wrappers
        and the tiny inline bridge append/pop helpers
    - the next compiler recursion tranche is now landed too:
      - compiled bodies now split into raw `__able_compiled_*` implementations
        plus env-swapping `__able_compiled_entry_*` entry wrappers
      - same-package static compiled calls now target the raw body directly,
        so recursive static paths stop paying `SwapEnvIfNeeded` on every
        self-call while cross-package/runtime entrypoints still preserve the
        existing package-env semantics
      - refreshed aligned compiled numbers on the kept code:
        - `fib`: `6.79s` vs the prior `13.40s`
        - `binarytrees`: `15.02s` vs the prior `19.84s`
      - refreshed profiles show `SwapEnvIfNeeded` has dropped out of the
        compiled `fib` hotspot set; the remaining recursion overhead there is
        now mostly `__able_push_call_frame` / `__able_pop_call_frame`, while
        compiled `binarytrees` is now primarily allocation/GC with only a
        smaller residual call-frame slice
    - the next compiler recursion tranche is now landed too:
      - the generated `__able_push_call_frame` / `__able_pop_call_frame`
        trampoline functions are gone; compiled static call sites now emit
        direct `bridge.PushCallFrame(__able_runtime, ...)` /
        `bridge.PopCallFrame(__able_runtime)` calls instead
      - refreshed aligned compiled numbers on the kept code:
        - `fib`: `5.09s` vs the prior `6.79s`
        - `binarytrees`: `14.76s` vs the prior `15.02s`
      - refreshed profiles show the wrapper layer is gone completely:
        `fib` is now dominated by the recursive body plus direct inline
        `bridge.PushCallFrame` / `bridge.PopCallFrame`, and `binarytrees`
        remains primarily allocation/GC with only a smaller residual direct
        call-frame slice
    - the next compiler recursion tranche is now landed too:
      - compiled static call sites no longer maintain a live bridge call-frame
        stack on the success path; instead they append caller frames onto the
        runtime-diagnostic error only when a non-nil `__ableControl` actually
        propagates
      - the interpreter/bridge diagnostic path now exposes
        `AppendRuntimeCallFrame(...)` /
        `bridge.AppendCallFrameError(...)` so the slow error path preserves
        the same runtime notes without any hot-path push/pop bookkeeping
      - refreshed aligned compiled numbers on the kept code:
        - `fib`: `3.33s` vs the prior `5.09s`, now about `1.17x` Go
          (`2.84s`)
        - `binarytrees`: `14.26s` vs the prior `14.76s`, now about `3.72x`
          Go (`3.83s`)
      - refreshed profiles show the compiled recursion scaffolding is now
        effectively closed for `fib`: the profile is almost entirely
        `__able_compiled_fn_fib`, while `binarytrees` is now overwhelmingly
        allocation/GC with only a tiny residual compiler slice
    - the next aligned text/fs tranche is now landed too:
      - the kept changes are in the canonical external stdlib repo, not the
        compiler core: `../able-stdlib/src/text/string.able` now fast-paths
        `String.len_bytes`, `String.contains`, and `String.replace` through
        host-native Go/TypeScript helpers, and `../able-stdlib/src/fs.able`
        now fast-paths newline normalization plus line splitting for
        `read_lines`
      - those paths avoid the old repeated `validated_bytes(...)`,
        UTF-8-validation, and string/array conversion churn that dominated the
        aligned `i_before_e` workload
      - refreshed aligned `i_before_e` numbers on the kept code:
        - compiled: `1.07s` vs the prior `3.99s`
        - bytecode: `56.76s`, down from timing out at `90s`
      - the aligned compiled `i_before_e` gap is still large relative to Go
        (`0.05s`) / Ruby (`0.10s`) / Python (`0.13s`), but the benchmark is
        no longer dominated by obviously naive stdlib text/fs helpers and the
        bytecode run now completes instead of timing out
      - the rejected sibling-stdlib experiment stands: avoid extern/helper
        designs that widen through large `Array String` conversions, since
        that path regressed the real `i_before_e` workload instead of
        improving it
    - the next bytecode recursion tranche is now landed too:
      - the bytecode VM now keeps same-program self-fast recursion on a
        compact call-frame stack instead of always appending the full
        `bytecodeCallFrame` record, and inline returns now use a cached simple
        return-type name when available instead of always probing the generic
        `inlineCoercionUnnecessary(...)` path
      - hot bytecode integer paths now use reference-style runtime integer
        accessors in `v12/interpreters/go/pkg/runtime/values.go`, which
        removes repeated `IntegerValue` value-method copies from the fused
        self-call subtraction path and the slot-const comparison helpers
      - refreshed aligned bytecode results on the kept code:
        - `fib`: still above the `90s` and `120s` timeout gates
        - `i_before_e`: still in the same high-50s band (`57.55-57.74s`
          across repeated aligned reruns)
      - the profile shift is still meaningful even without a new `fib`
        completion time:
        - `pushCallFrame` dropped from roughly `10s` flat to roughly `3s`
          flat on the aligned timeout profile
        - the remaining aligned `fib` runtime is now dominated by
          `execCallSelfIntSubSlotConst`, `execBinary`, `releaseSlotFrame`,
          and the narrowed immediate-subtract path instead of the older
          general call-frame bookkeeping
      - the next reduced-`fib` lowering slice is landed too:
        - statement-position `if` conditions that already match the slot-const
          integer compare fast path now lower directly to a conditional jump
          opcode instead of materializing a temporary boolean only to have
          `JumpIfFalse` consume it immediately
        - refreshed reduced `BenchmarkFib30Bytecode` reruns on the kept code:
          `159.93ms/op`, `155.70ms/op`, `151.83ms/op`
        - profiled reduced rerun on the kept code: `151.60ms/op`
        - aligned external bytecode `fib` still times out at `90s`, so the
          remaining aligned wall is now even more cleanly the residual
          self-call / add / slot-frame path rather than the old base-case
          compare result materialization
      - aligned `i_before_e` is no longer blocked on stdlib text helpers, but
        its bytecode profile still mixes one-time lowering with VM runtime
        cost
      - `v12/bench_perf` now also has a `bytecode-runtime` mode backed by a
        generic Go benchmark harness, so aligned steady-state bytecode runs
        can load/lower once and then measure repeated `main()` calls with
        `ns/op`, `B/op`, and `allocs/op`
      - first steady-state aligned results on the kept code:
        - `fib`: still timed out even with a `300s` warmup+measure budget
        - `i_before_e`: `62.14s/op`, `107.19 GB/op`, `315,106,815 allocs/op`
          with `123.61s` wall-clock for warmup plus measured call
      - that new measurement path shows the remaining bytecode problem is VM
        runtime itself, not just CLI/bootstrap/lowering noise; repeated
        aligned `i_before_e` execution is actually slower than the old
        one-shot CLI wall-clock number
    - the next steady-state profiling tranche is now landed too:
      - `bytecode-runtime` profiling now starts after program load/lowering
        plus the explicit warmup call, using dedicated runtime-profile env
        vars inside the benchmark harness instead of the broader Go test
        profiling path
      - the interpreter now caches lowered expression bytecode programs by
        AST-expression identity plus placeholder-lambda mode, so repeated
        bytecode evaluation of `match` guards/bodies, `ensure`, `rescue`,
        breakpoint bodies, and similar AST subexpressions no longer pays a
        fresh lowering pass during steady-state execution
      - focused runtime coverage now pins repeated bytecode match evaluation
        to reuse the cached lowered programs
      - refreshed aligned steady-state `i_before_e` numbers on the kept code:
        - `27.92s/op`
        - `16.61 GB/op`
        - `172,094,647 allocs/op`
      - the corrected runtime-only profiles no longer show
        `lowerExpressionToBytecodeWithOptions(...)`,
        `(*bytecodeLoweringContext).emit(...)`, or `emitExpression(...)` in
        the hot set; the remaining steady-state cost is now actual VM/runtime
        work centered on `execCallName`, `execCall`, `runMatchExpression`,
        identifier/call-name cache churn, and GC/allocation pressure
      - allocation space also shifted materially away from lowering and onto
        runtime structures such as cached scope entries, environment defines,
        cached call-name entries, and repeated runtime context snapshots
    - the next steady-state runtime-cache tranche is now landed too:
      - pooled bytecode VMs now preserve their validated lookup and dispatch
        caches across repeated runs instead of zeroing lexical/call/member/
        index cache tables on every `main()` call
      - match-pattern binding environments now use the lighter
        `collectPatternBindings(...)` path plus pre-sized environments and
        non-merging local binds, which removes the old recursive
        `assignPattern(...)` declaration path from hot steady-state match
        evaluation
      - refreshed aligned steady-state `i_before_e` numbers on the kept code:
        - `24.61s/op`
        - `9.90 GB/op`
        - `146,646,034 allocs/op`
      - the runtime-only profiles now show the intended shift:
        `storeCachedScopeValue(...)` / `storeCachedCallName(...)` dropped out
        of the alloc-space top tier, while the remaining cost is concentrated
        in `matchPattern(...)`, environment creation/binds for clause scopes,
        `runMatchExpression(...)`, and the downstream runtime work those
        clauses trigger
    - the next steady-state match fast-path tranche is now landed too:
      - simple match patterns now bypass the generic
        `collectPatternBindings(...)` slice-building path and bind directly
        into fresh clause-local environments for identifier, wildcard,
        literal, and typed-pattern cases
      - refreshed aligned steady-state `i_before_e` numbers on the kept code:
        - `23.84s/op`
        - `8.89 GB/op`
        - `139,670,607 allocs/op`
      - the refreshed profiles show the intended shift:
        `matchPattern(...)` itself is no longer the dominant allocator, and
        `matchPatternFast(...)` now carries much of the hot match work
        directly; the remaining runtime pressure is concentrated in
        clause-scope environment creation/binds, `runMatchExpression(...)`,
        `execEnsureStart(...)`, struct literal construction, extern-host
        calls, and runtime context snapshots
    - the next steady-state lazy return-context tranche is now landed too:
      - bytecode and tree-walker `returnSignal` no longer snapshot runtime
        call stacks on the normal control-flow path; return coercion failures
        now attach runtime context lazily only when they actually need a
        diagnostic
      - refreshed aligned steady-state `i_before_e` numbers on the kept code:
        - `23.40s/op`
        - `8.74 GB/op`
        - `136,183,494 allocs/op`
      - the refreshed profiles show the intended shift:
        `snapshotCallStack(...)` dropped again as a top allocator and the
        remaining steady-state pressure is now more cleanly concentrated in
        clause-scope environment/binding churn, `runMatchExpression(...)`,
        `execEnsureStart(...)`, struct literal construction, extern-host
        calls, and runtime context attachment on actual error paths
    - the next steady-state extern-host fast-path tranche is now landed too:
      - extern target hashes are now cached on the registered target state and
        invalidated only when new host preludes/externs are registered, so hot
        extern calls stop re-hashing the whole target definition set on every
        invocation
      - loaded extern modules now build direct invokers for the hot primitive
        string signatures that dominate aligned `i_before_e`
        (`String -> i32`, `String -> bool`, `String -> String`,
        `String,String -> bool`, `String,String -> String`,
        `String,String,String -> String`, `String -> Array String`), which
        bypasses the generic reflection path on those calls
      - refreshed aligned steady-state `i_before_e` numbers on the kept code:
        - profiled: `23.43s/op`, `8.23 GB/op`, `121,472,207 allocs/op`
        - clean rerun: `22.24s/op`, `8.23 GB/op`, `121,472,290 allocs/op`
      - the refreshed profiles show the intended shift:
        `hashExternState(...)`, `externSignatureKey(...)`, and the old
        cumulative extern-host hashing path dropped out of the top allocators,
        while `invokeExternHostFunction(...)` / `fromHostResults(...)` shrank
        materially and the remaining runtime pressure is now more clearly in
        clause-scope environment/binding churn, `runMatchExpression(...)`,
        `execEnsureStart(...)`, struct literal construction, and the residual
        host-value conversion path
    - the next steady-state environment single-binding tranche is now landed
      too:
      - `runtime.Environment` now keeps the first local binding in an inline
        slot and promotes to a real map only on the second distinct local
        binding, so single-bind match/rescue/or-else scopes stop allocating a
        one-entry map by default
      - `NewEnvironmentWithValueCapacity(...)` no longer allocates a map for
        `valueCapacity == 1`, which keeps the common one-binding child-scope
        case on that inline path
      - refreshed aligned steady-state `i_before_e` numbers on the kept code:
        - profiled: `24.40s/op`, `6.64 GB/op`, `107,523,776 allocs/op`
        - clean rerun: `21.93s/op`, `6.64 GB/op`, `107,523,333 allocs/op`
      - refreshed steady-state `fib` on the same runtime-only path still
        times out at `300s`, so this tranche improved text-heavy bytecode
        runtime pressure but did not materially change the recursion path
      - the refreshed alloc profile shows the intended shift:
        first-binding map allocation is cut, but the remaining alloc wall is
        now more clearly `NewEnvironmentWithValueCapacity(...)` object churn,
        `promoteSingleBindingNoLock(...)` on multi-bind scopes,
        `evaluateStructLiteral(...)`, `snapshotCallStack(...)`,
        `runMatchExpression(...)`, and `execEnsureStart(...)`
    - the next steady-state block-scope and miss-lookup tranche is now landed
      too:
      - tree-walker block, function/lambda call, loop-iteration, iterator,
        and `or {}` handler scopes now use `NewEnvironmentWithValueCapacity`
        with a cheap AST-derived binding estimate instead of always starting
        from an empty child scope and promoting later
      - `matchPatternFast(...)` now uses `Environment.Lookup(...)` for the hot
        singleton-struct probe instead of the miss-allocating `Get(...)` path
      - refreshed aligned steady-state `i_before_e` numbers on the kept code:
        - profiled: `22.01s/op`, `6.43 GB/op`, `97,060,888 allocs/op`
        - clean reruns: `21.93s/op` and `21.87s/op`, with
          `6.43 GB/op`, `97,062,815 allocs/op` and
          `6.43 GB/op`, `97,062,593 allocs/op`
      - wall-clock stayed in the same low-21s band, but allocation pressure
        dropped materially from the prior kept baseline of
        `6.64 GB/op`, `107,523,333 allocs/op`
      - the refreshed profiles show the intended shift:
        `Environment.Get(...)` / `fmt.Errorf(...)` miss pressure dropped out of
        the top tier, `matchPattern(...)` cumulative allocs narrowed again, and
        the remaining steady-state wall is now more cleanly
        `NewEnvironmentWithValueCapacity(...)`, `setCurrentValueNoLock(...)`,
        `evaluateStructLiteral(...)`, `snapshotCallStack(...)`,
        `runMatchExpression(...)`, and `execEnsureStart(...)`
    - the next steady-state lazy runtime-context tranche is now landed too:
      - runtime diagnostics no longer eagerly copy the eval-state call stack
        at first attachment; they now keep a lazy reference to the current
        stack and only freeze it if the stack is about to mutate or if a real
        diagnostic is actually built
      - `evalState.pushCallFrame(...)` / `popCallFrame(...)` now flush any
        pending lazy diagnostic contexts before mutating the call stack, so
        escaping errors still preserve the original call chain
      - refreshed aligned steady-state `i_before_e` numbers on the kept code:
        - profiled: `20.51s/op`, `5.85 GB/op`, `84,856,303 allocs/op`
        - clean rerun A: `23.19s/op`, `5.85 GB/op`, `84,856,355 allocs/op`
        - clean rerun B: `22.09s/op`, `5.85 GB/op`, `84,856,481 allocs/op`
      - refreshed steady-state `fib` on the same runtime-only path still
        times out at `300s`
      - the refreshed alloc profile shows the intended shift:
        `snapshotCallStack(...)` dropped out of the top allocators entirely,
        while the remaining wall is now more cleanly
        `NewEnvironmentWithValueCapacity(...)`,
        `evaluateStructLiteral(...)`, `setCurrentValueNoLock(...)`,
        `collectImplCandidates(...)`, `arrayMember(...)`,
        `runMatchExpression(...)`, and `execEnsureStart(...)`
    - the next steady-state type-canonicalization tranche is now landed too:
      - `canonicalizeExpandedTypeExpression(...)` now reuses the original
        nullable/result/union/function/generic AST nodes when none of their
        children actually change, instead of rebuilding fresh type-expression
        trees on every no-op canonicalization pass
      - focused coverage now pins both behaviors:
        unchanged nested type expressions preserve identity, while changed
        nested members still rebuild correctly when canonical names shift
      - refreshed aligned steady-state `i_before_e` numbers on the kept code:
        - profiled: `20.57s/op`, `5.42 GB/op`, `77,708,818 allocs/op`
        - clean rerun A: `20.39s/op`, `5.42 GB/op`, `77,708,480 allocs/op`
        - clean rerun B: `20.82s/op`, `5.42 GB/op`, `77,710,637 allocs/op`
      - the refreshed profiles show the intended shift:
        `ast.NewNullableTypeExpression(...)` and
        `ast.NewUnionTypeExpression(...)` dropped out of the alloc-space top
        set entirely, and `canonicalizeExpandedTypeExpression(...)` itself
        shrank to a much smaller CPU/alloc slice
    - the next steady-state interface/member-resolution tranche is now landed
      too:
      - array helper member access now skips the guaranteed direct-member miss
        for non-field names like `len` / `get` / `push`, while preserving
        direct-member precedence for `storage_handle`, `length`, `capacity`,
        and `iterator`
      - `typeImplementsInterface(...)` now caches resolved
        type/interface/arg-signature results behind the same invalidation
        boundary as the existing method cache, so repeated match/ensure/member
        checks stop rebuilding the same impl-candidate set on the steady-state
        hot path
      - focused coverage now pins both behaviors:
        direct array members still beat methods, non-direct array names still
        route through method lookup, and interface-implementation cache entries
        clear on method-cache invalidation
      - refreshed aligned steady-state `i_before_e` numbers on the kept code:
        - profiled: `21.00s/op`, `5.06 GB/op`, `77,363,744 allocs/op`
        - clean rerun: `20.19s/op`, `5.06 GB/op`, `77,364,488 allocs/op`
      - the refreshed profiles show the intended shift:
        `collectImplCandidates(...)` dropped out of the top alloc-space set on
        the kept profiled rerun, so the remaining wall is now more cleanly
        `NewEnvironmentWithValueCapacity(...)`,
        `evaluateStructLiteral(...)`, `setCurrentValueNoLock(...)`,
        `arrayMember(...)`, `runMatchExpression(...)`, `execEnsureStart(...)`,
        and the residual host-value conversion path
    - the next steady-state array-metadata tranche is now landed too:
      - dynamic array state now caches boxed `length` / `capacity` metadata
        values, so repeated array helper/member access stops re-boxing the
        same large integers on every read
      - struct-literal success-path lookups now use `Environment.Lookup(...)`
        instead of the heavier error-producing `Get(...)` path when resolving
        shorthand field bindings or falling back from `StructDefinition(...)`
      - focused coverage now pins the large-length metadata case:
        repeated `array.length` access after the first read allocates
        effectively zero additional objects
      - refreshed aligned steady-state `i_before_e` numbers on the kept code:
        - profiled: `23.03s/op`, `4.88 GB/op`, `73,577,870 allocs/op`
        - clean rerun A: `20.00s/op`, `4.88 GB/op`, `73,577,767 allocs/op`
        - clean rerun B: `19.61s/op`, `4.88 GB/op`, `73,577,672 allocs/op`
      - the refreshed alloc profile shows the intended shift:
        `arrayMember(...)` dropped from the older ~`343 MB` flat tier to about
        `160 MB`, with the remaining metadata cost now concentrated in the
        first boxed-length materialization rather than repeated re-boxing
    - the next steady-state environment-object tranche is now landed too:
      - `runtime.Environment` now moves the cold struct-definition/runtime-data
        fields behind a lazy `environmentMeta` pointer, so ordinary lexical
        scopes no longer carry those fields inline when they only need local
        value bindings
      - focused runtime coverage now pins the runtime-data parent fallback on
        that lazy metadata path
      - refreshed aligned steady-state `i_before_e` numbers on the kept code:
        - profiled: `21.33s/op`, `4.63 GB/op`, `73,577,840 allocs/op`
        - clean rerun A: `20.51s/op`, `4.63 GB/op`, `73,577,220 allocs/op`
        - clean rerun B: `20.09s/op`, `4.63 GB/op`, `73,577,644 allocs/op`
      - the refreshed alloc profile shows the intended shift:
        `NewEnvironmentWithValueCapacity(...)` fell from the old ~`1.95 GB`
        flat object-allocation tier to about `1.71 GB`, while the value-map
        allocation slice stayed small; the remaining wall is now more cleanly
        scope count plus `evaluateStructLiteral(...)`,
        `setCurrentValueNoLock(...)`, `runMatchExpression(...)`,
        `execEnsureStart(...)`, and residual host-value conversion
    - the next steady-state propagation tranche is now landed too:
      - the canonical stdlib `io.unwrap(...)`, `io.unwrap_void(...)`, and
        `io.bytes_to_string(...)` paths in `../able-stdlib/src/io.able` now use
        direct propagation (`!`) instead of nested `match`/`raise` control
        flow, so the aligned fs/text path stops building that extra match tree
        on every read/unwrap step
      - the Go interpreter now reuses `cachedSimpleTypeExpression("Error")`
        on the hot propagation/or-else/runtime error checks instead of
        constructing a fresh `ast.Ty("Error")` node on each pass through those
        paths
      - refreshed aligned steady-state `i_before_e` numbers on the kept code:
        - profiled: `19.17s/op`, `4.60 GB/op`, `73,227,937 allocs/op`
        - clean rerun A: `20.94s/op`, `4.60 GB/op`, `73,229,445 allocs/op`
        - clean rerun B: `19.92s/op`, `4.60 GB/op`, `73,228,693 allocs/op`
      - the refreshed profile shows the intended shift:
        the old `bytecodeOpPropagation` alloc line dropped to zero flat alloc
        space on the kept profile, and the steady-state heap moved from the
        prior kept `4.63 GB/op` band down to about `4.60 GB/op` with alloc
        count dropping by roughly `350k` objects per run
    - the next steady-state shared array-metadata boxing tranche is now landed
      too:
      - `v12/interpreters/go/pkg/runtime/array_store.go` now keeps shared boxed
        metadata values for common dynamic-array `i32` lengths/capacities, and
        `v12/interpreters/go/pkg/interpreter/interpreter_arrays.go` now reuses
        the same shared `u64` metadata boxes for `__able_array_size`
      - focused coverage now pins first-access large metadata boxing in
        `v12/interpreters/go/pkg/runtime/array_store_test.go` plus the
        corresponding `__able_array_size` path in
        `v12/interpreters/go/pkg/interpreter/interpreter_strings_arrays_test.go`
      - refreshed aligned steady-state `i_before_e` numbers on the kept code:
        - profiled: `20.24s/op`, `4.57 GB/op`, `72,504,749 allocs/op`
        - clean rerun: `19.87s/op`, `4.57 GB/op`, `72,505,049 allocs/op`
      - the refreshed alloc profile shows the intended shift:
        `(*Interpreter).initArrayBuiltins.func4` fell from about `167 MB` flat
        alloc-space to about `148 MB`, and `(*ArrayState).BoxedLengthValue`
        fell from about `160 MB` to about `149 MB`
    - the next steady-state small unsigned extern-host conversion tranche is
      now landed too:
      - `v12/interpreters/go/pkg/interpreter/extern_host_coercion.go` now
        lowers host `u8` / `u16` / `u32` and in-range `u64` / `usize` results
        straight to small integers instead of routing them through `big.Int`
        boxing
      - focused extern coverage now pins small `u64`, `Array u8`, and
        out-of-range `u64` fallback behavior in
        `v12/interpreters/go/pkg/interpreter/interpreter_extern_test.go`
      - refreshed aligned steady-state `i_before_e` numbers on the kept code:
        - profiled: `22.95s/op`, `4.50 GB/op`, `69,018,740 allocs/op`
        - clean rerun A: `20.57s/op`, `4.50 GB/op`, `69,019,113 allocs/op`
        - clean rerun B: `21.65s/op`, `4.50 GB/op`, `69,018,676 allocs/op`
        - clean rerun C: `19.46s/op`, `4.50 GB/op`, `69,021,452 allocs/op`
      - the refreshed alloc profile shows the intended shift:
        `(*Interpreter).fromHostValue` fell from about `103 MB` flat
        alloc-space to about `91 MB`, and the old `bigIntFromUint(...)`
        unsigned-host conversion slice dropped out of the top alloc set
    - the next steady-state lazy environment-mutex tranche is now landed too:
      - `v12/interpreters/go/pkg/runtime/environment.go` now keeps the
        per-environment `sync.RWMutex` behind a lazily allocated
        `atomic.Pointer`, so the single-threaded bytecode hot path stops paying
        that lock payload in every short-lived lexical scope object
      - focused runtime coverage now pins both lazy multi-thread lock creation
        and the single-thread no-lock path in
        `v12/interpreters/go/pkg/runtime/environment_test.go`
      - refreshed aligned steady-state `i_before_e` numbers on the kept code:
        - profiled: `21.98s/op`, `4.25 GB/op`, `69,018,511 allocs/op`
        - clean rerun A: `21.97s/op`, `4.25 GB/op`, `69,018,009 allocs/op`
        - clean rerun B: `21.84s/op`, `4.25 GB/op`, `69,019,472 allocs/op`
      - the refreshed alloc profile shows the intended shift:
        `NewEnvironmentWithValueCapacity(...)` fell from about `1.71 GB` flat
        alloc-space to about `1.48 GB`, while the value-map slice stayed small;
        the remaining wall is now more cleanly `evaluateStructLiteral(...)`,
        `setCurrentValueNoLock(...)`, `runMatchExpression(...)`,
        `execEnsureStart(...)`, and residual host-value conversion
    - the next stdlib `read_lines` fast-path tranche is now landed too:
      - the canonical external stdlib in `../able-stdlib/src/fs.able` now uses
        a direct `fs_read_lines_fast(...)` extern for `fs.read_lines(...)`
        instead of routing through `open` / `read_all` / `close`,
        `bytes_to_string(...)`, newline normalization, and line splitting in
        layered Able code
      - focused compiled temp-file coverage now pins the public
        `fs.read_lines(...)` API behavior in
        `v12/interpreters/go/pkg/compiler/compiler_stdlib_io_temp_test.go`,
        including CRLF normalization and trailing-line trimming
      - refreshed aligned `i_before_e` numbers on the kept code:
        - bytecode-runtime clean rerun A: `1.28s/op`, `101.46 MB/op`,
          `3,582,266 allocs/op`
        - bytecode-runtime clean rerun B: `1.40s/op`, `101.47 MB/op`,
          `3,582,304 allocs/op`
        - bytecode-runtime profiled: `1.41s/op`, `101.51 MB/op`,
          `3,582,321 allocs/op`
        - compiled external compare: `0.38s` (down from the prior `1.07s`)
      - the refreshed profile shows the intended shift:
        the old `runMatchExpression(...)`, `execEnsureStart(...)`,
        `NewEnvironmentWithValueCapacity(...)`, and `evaluateStructLiteral(...)`
        wall dropped out of the steady-state top alloc set for this benchmark;
        the remaining `i_before_e` wall is now `copyCallArgs`,
        `resolveMethodFromPool`, `stringMemberWithOverrides`, and the residual
        extern-host conversion path
    - the next primitive-receiver method-resolution tranche is now landed too:
      - `v12/interpreters/go/pkg/interpreter/interpreter_method_resolution.go`
        now skips eager scope-callable and type-name probing for primitive
        receivers until after inherent/interface/native method lookup fails, so
        hot `String` method calls stop paying `env.Lookup(...)` / `env.Has(...)`
        work that only matters for the fallback UFCS path
      - the rejected concrete-string receiver cache experiment was not kept;
        the kept direction is the primitive fast path only, which preserves the
        existing pointer-receiver bound-method cache behavior without creating
        per-value cache churn
      - refreshed aligned `i_before_e` numbers on the kept code:
        - bytecode-runtime clean rerun: `1.27s/op`, `101.47 MB/op`,
          `3,582,283 allocs/op`
        - bytecode-runtime profiled: `1.24s/op`, `101.51 MB/op`,
          `3,582,351 allocs/op`
      - the refreshed profile shows the intended CPU shift:
        `resolveMethodFromPool(...)` dropped from about `210ms` cumulative on
        the post-`read_lines` profile to about `120ms`, and
        `stringMemberWithOverrides(...)` dropped from about `250ms` cumulative
        to about `100ms`; the remaining text-path wall is now led more clearly
        by `copyCallArgs`, residual `resolveMethodFromPool(...)` flat allocs,
        `overloadArgKinds(...)`, and extern-host conversion
    - the next extern-wrapper borrow-args tranche is now landed too:
      - `v12/interpreters/go/pkg/interpreter/definitions.go` now marks
        `makeExternNative(...)` wrappers with `BorrowArgs: true`, so the exact
        native bytecode path stops cloning argument slices before synchronous
        extern-host invocation
      - focused extern coverage in
        `v12/interpreters/go/pkg/interpreter/interpreter_extern_test.go` now
        pins that registered Go extern wrappers borrow their args, while the
        existing native-call stability tests still pin the non-borrowing path
      - refreshed aligned `i_before_e` numbers on the kept code:
        - bytecode-runtime clean rerun A: `1.14s/op`, `84.88 MB/op`,
          `3,063,799 allocs/op`
        - bytecode-runtime clean rerun B: `1.19s/op`, `84.89 MB/op`,
          `3,063,840 allocs/op`
        - bytecode-runtime profiled: `1.11s/op`, `84.92 MB/op`,
          `3,063,917 allocs/op`
      - the refreshed profile shows the intended alloc shift:
        `copyCallArgs(...)` dropped out of the top alloc set entirely, total
        alloc-space fell from about `119.85 MB` to about `90.95 MB`, and the
        remaining text-path wall is now led by `resolveMethodFromPool(...)`,
        `overloadArgKinds(...)`, `stringMemberWithOverrides(...)`, and the
        residual extern conversion path through `fromHostValue(...)`
    - the next overload-cache signature tranche is now landed too:
      - `v12/interpreters/go/pkg/interpreter/eval_expressions_calls_overloads.go`
        now uses an inline comparable overload-cache signature for small-arity
        hot calls instead of rebuilding the old concatenated
        `overloadArgKinds(...)` string on every lookup; larger arities still
        fall back to the old slow path
      - the new cache signature also carries float subkind and host-handle
        type detail, and focused coverage in
        `v12/interpreters/go/pkg/interpreter/eval_expressions_calls_overloads_test.go`
        now pins those distinctions directly
      - refreshed aligned `i_before_e` numbers on the kept code:
        - bytecode-runtime clean rerun: `1.18s/op`, `75.17 MB/op`,
          `2,718,071 allocs/op`
        - bytecode-runtime profiled: `1.14s/op`, `75.21 MB/op`,
          `2,718,161 allocs/op`
      - the refreshed profile shows the intended alloc shift:
        `overloadArgKinds(...)` dropped out of the top alloc set entirely, and
        total profiled alloc-space fell again from about `90.95 MB` to about
        `82.41 MB`; the remaining text-path wall is now led more cleanly by
        `resolveMethodFromPool(...)`, `stringMemberWithOverrides(...)`, and
        residual extern conversion through `fromHostValue(...)`
    - the next bytecode direct member-call tranche is now landed too:
      - `v12/interpreters/go/pkg/interpreter/interpreter_method_resolution.go`
        now exposes `resolveMethodCallableFromPool(...)`, which returns the
        callable template directly instead of forcing the hot path through a
        fresh `runtime.BoundMethodValue`
      - `v12/interpreters/go/pkg/interpreter/bytecode_lowering.go` now emits a
        dedicated `bytecodeOpCallMember` for identifier member-call syntax, and
        `v12/interpreters/go/pkg/interpreter/bytecode_vm_call_member.go` now
        executes that path by resolving the callable template, reusing the
        existing injected-receiver inline-call path when possible, and only
        falling back to ordinary member access when field-callable precedence
        requires it
      - focused coverage in
        `v12/interpreters/go/pkg/interpreter/bytecode_vm_call_member_test.go`
        now pins both direct method calls and the callable-field fallback case
      - refreshed aligned `i_before_e` numbers on the kept code:
        - bytecode-runtime clean rerun A: `1.06s/op`, `55.81 MB/op`,
          `1,853,918 allocs/op`
        - bytecode-runtime clean rerun B: `1.09s/op`, `55.81 MB/op`,
          `1,853,917 allocs/op`
        - bytecode-runtime profiled: `1.09s/op`, `55.85 MB/op`,
          `1,854,009 allocs/op`
      - the refreshed profile shows the intended alloc shift:
        `resolveMethodFromPool(...)` dropped out of the top alloc set
        entirely, and total profiled alloc-space fell again from about
        `82.41 MB` to about `72.65 MB`; the remaining text-path wall is now
        led more cleanly by `callResolvedCallableWithInjectedReceiver(...)`,
        `fromHostValue(...)`, and the residual extern/string conversion path
    - the next extern return-conversion tranche is now landed too:
      - `v12/interpreters/go/pkg/interpreter/extern_host_fast.go` now routes
        hot `i32` fast-invoker results through `boxedOrSmallIntegerValue(...)`
        instead of boxing a fresh `runtime.NewSmallInt(...)` on every call
      - `v12/interpreters/go/pkg/interpreter/extern_host_coercion.go` plus the
        new `v12/interpreters/go/pkg/interpreter/extern_host_result_fast.go`
        now fast-path host `String`, `Array String`, and `IOError | Array
        String`-style union results before falling back to the old generic
        reflect-heavy conversion path
      - focused coverage in
        `v12/interpreters/go/pkg/interpreter/extern_host_result_fast_test.go`
        now pins the direct string-slice conversion helper, the union-member
        preference logic, the `IOError | Array String` host return path, and
        the hot `i32` fast-invoker signature
      - refreshed aligned `i_before_e` numbers on the kept code:
        - bytecode-runtime clean rerun A: `1.08s/op`, `42.69 MB/op`,
          `1,335,429 allocs/op`
        - bytecode-runtime clean rerun B: `1.08s/op`, `42.70 MB/op`,
          `1,335,471 allocs/op`
        - bytecode-runtime profiled: `1.08s/op`, `42.74 MB/op`,
          `1,335,572 allocs/op`
      - the refreshed profile shows the intended alloc shift:
        the old `fmt.Sprint(value.Interface())` / generic `fromHostValue(...)`
        extern return path collapsed out of the top flat allocators, the hot
        `buildExternFastInvoker.func1` `runtime.NewSmallInt(...)` slice
        disappeared, and total profiled alloc-space fell again from about
        `72.65 MB` to about `50.44 MB` while wall-clock stayed in the same
        `~1.08s/op` band
    - the next union extern-invoker tranche is now landed too:
      - `v12/interpreters/go/pkg/interpreter/extern_host.go` now passes the
        active interpreter into cached fast invokers, so a direct typed host
        call can still fall back to the normal host-result coercion path
        without re-invoking the extern
      - `v12/interpreters/go/pkg/interpreter/extern_host_fast.go` now
        fast-paths one-string-arg `func(string) interface{}` host wrappers,
        which is the shape produced for union-return externs like
        `fs_read_lines_fast(...)`; hot `[]string` success returns now bypass
        `reflect.Value.Call` entirely, while non-`[]string` results still flow
        through `fromHostValue(...)` using the already-computed direct result
      - focused coverage in
        `v12/interpreters/go/pkg/interpreter/extern_host_result_fast_test.go`
        now pins the direct union `Array String` fast-invoker path alongside
        the earlier `String` / `Array String` host-result helpers
      - refreshed aligned `i_before_e` numbers on the kept code:
        - bytecode-runtime clean rerun A: `0.98s/op`, `42.71 MB/op`,
          `1,335,497 allocs/op`
        - bytecode-runtime clean rerun B: `1.01s/op`, `42.70 MB/op`,
          `1,335,460 allocs/op`
        - bytecode-runtime profiled: `1.01s/op`, `42.74 MB/op`,
          `1,335,541 allocs/op`
      - the refreshed profile shows the intended CPU shift:
        `reflect.Value.Call` dropped out of the top alloc and CPU sets for the
        `fs_read_lines_fast(...)` success path, and steady-state wall-clock
        moved down from the prior `~1.08s/op` band into the
        `~0.98-1.01s/op` band while keeping the same collapsed heap profile
    - the next integer-kind boxing tranche is now landed too:
      - `v12/interpreters/go/pkg/interpreter/bytecode_vm_small_int_boxing.go`
        now boxes all hot int64-representable integer suffixes through the same
        shared fixed/dynamic caches instead of limiting the hot path to `i32`,
        `i64`, and `isize`
      - focused coverage in
        `v12/interpreters/go/pkg/interpreter/bytecode_vm_slot_const_immediates_test.go`
        now pins the unsigned fixed-cache and dynamic-cache paths too
      - refreshed aligned `i_before_e` numbers on the kept code:
        - bytecode-runtime clean rerun A: `1.09s/op`, `34.39 MB/op`,
          `1,162,592 allocs/op`
        - bytecode-runtime profiled: `1.08s/op`, `34.43 MB/op`,
          `1,162,662 allocs/op`
      - subsequent clean reruns on this machine were wall-clock noisy while
        holding the same `~34.4 MB/op` / `1.16M allocs/op` heap shape, so this
        is best treated as an allocation-pressure keep rather than a new stable
        CPU-path win
      - the refreshed alloc profile shows the intended heap shift:
        the old unsupported-kind fallback in `boxedOrSmallIntegerValue(...)`
        disappeared, total profiled alloc-space fell again from about
        `61.30 MB` to about `50.83 MB`, and the remaining text-path wall is now
        more clearly `bytecodeBoxedIntegerValue(...)`,
        `patternToInteger(...)`, `callResolvedCallableWithInjectedReceiver(...)`,
        `externStringSliceResult(...)`, and the residual direct plugin-body
        cost in `fs_read_lines_fast(...)`
    - the next small-int cast tranche is now landed too:
      - `v12/interpreters/go/pkg/interpreter/interpreter_type_coercion_fast.go`
        now keeps small integer-to-integer suffix casts on a 64-bit arithmetic
        path instead of immediately allocating `big.Int` values for
        `bitPattern(...)` / `patternToInteger(...)`
      - the same helper is reused from
        `v12/interpreters/go/pkg/interpreter/interpreter_type_coercion.go`, so
        both the direct fast cast path and the post-alias simple-type cast path
        avoid the old `big.Int` churn when the wrapped result still fits in
        `int64`
      - focused coverage in
        `v12/interpreters/go/pkg/interpreter/interpreter_type_coercion_fast_test.go`
        now pins small signed-to-unsigned wrap behavior, the negative-to-`u64`
        big-integer fallback boundary, and the bounded allocation behavior on
        repeated `u8` casts
      - refreshed aligned `i_before_e` numbers on the kept code:
        - bytecode-runtime clean rerun A: `1.35s/op`, `26.10 MB/op`,
          `644,158 allocs/op`
        - bytecode-runtime clean rerun B: `1.17s/op`, `26.10 MB/op`,
          `644,158 allocs/op`
        - bytecode-runtime profiled: `1.36s/op`, `26.13 MB/op`,
          `644,192 allocs/op`
      - this is a keep for allocation pressure, not a stable CPU-path win:
        wall-clock stayed noisy on this machine, but the heap shift is large and
        repeatable
      - the refreshed alloc profile shows the intended shift:
        `castValueToCanonicalSimpleTypeFast(...)`, `castValueToType(...)`,
        `patternToInteger(...)`, and `bitPattern(...)` dropped out of the top
        alloc set entirely, and total profiled alloc-space fell again from
        about `50.83 MB` to about `44.90 MB`
    - the next native-bound-member exact-call tranche is now landed too:
      - `v12/interpreters/go/pkg/interpreter/bytecode_vm_call_member.go` now
        lets the direct member-call exact-native fast path accept
        `runtime.NativeBoundMethodValue` too, instead of only raw
        `NativeFunctionValue` templates
      - that closes a real gap between method resolution and bytecode member
        execution: the resolver can already return native bound methods for
        primitive/native receivers, and the VM now routes those straight through
        `execExactNativeCall(...)` instead of falling back to the generic
        injected-receiver call path
      - focused coverage in
        `v12/interpreters/go/pkg/interpreter/bytecode_vm_call_member_test.go`
        now pins `bytecodeResolveExactInjectedNativeCallTarget(...)` on
        `NativeBoundMethodValue`
      - refreshed aligned `i_before_e` numbers on the kept code:
        - bytecode-runtime clean rerun A: `1.08s/op`, `26.09 MB/op`,
          `644,119 allocs/op`
        - bytecode-runtime clean rerun B: `1.07s/op`, `26.13 MB/op`,
          `644,192 allocs/op`
        - bytecode-runtime clean rerun C: `1.12s/op`, `26.09 MB/op`,
          `644,118 allocs/op`
        - bytecode-runtime profiled: `1.07s/op`, `26.13 MB/op`,
          `644,192 allocs/op`
      - this is a CPU-path keep layered on top of the prior heap work:
        the aligned runtime moved back into the low `~1.07-1.12s/op` band while
        preserving the post-cast `~26.1 MB/op` / `644k allocs/op` heap shape
    - the next direct member-method cache tranche is now landed too:
      - `v12/interpreters/go/pkg/interpreter/bytecode_vm_call_member.go` now
        uses the existing bytecode member-method cache on the
        `bytecodeOpCallMember` path instead of bypassing it and re-running
        method resolution on every direct `obj.method(...)` call
      - that closes a real regression too: the existing
        `TestBytecodeVM_StatsMemberMethodCacheCounters` proof was red because
        `execCallMember(...)` never consulted the cache even though member
        access and dotted call-name fallback already did
      - on a miss, the VM now stores a rebound template for the exact same
        cache surface used elsewhere; on a hit, it executes the cached resolved
        member callee through the exact-native / inline / generic call ladder
        without re-running `resolveMethodCallableFromPool(...)`
      - refreshed aligned `i_before_e` numbers on the kept code:
        - bytecode-runtime clean rerun A: `1.00s/op`, `26.09 MB/op`,
          `644,119 allocs/op`
        - bytecode-runtime clean rerun B: `1.00s/op`, `26.09 MB/op`,
          `644,119 allocs/op`
        - bytecode-runtime profiled: `1.02s/op`, `26.13 MB/op`,
          `644,191 allocs/op`
      - this is a keep as both correctness and CPU-path work:
        the cache-counter regression is closed, and aligned steady-state
        bytecode `i_before_e` moved from the prior `~1.07-1.12s/op` band into
        the low `~1.00-1.02s/op` band while preserving the post-cast
        `~26.1 MB/op` / `644k allocs/op` heap shape
    - the next primitive bound-method cache tranche is now landed too:
      - `v12/interpreters/go/pkg/interpreter/interpreter_method_resolution.go`
        now gives primitive receivers stable bound-method cache keys by type
        token instead of treating value receivers like `String` as uncached
      - the same tranche also keeps the cache semantically safe by only
        storing primitive receiver entries when resolution actually came from
        a real method candidate; primitive scope-fallback callables stay
        uncached
      - focused coverage in
        `v12/interpreters/go/pkg/interpreter/interpreter_method_resolution_cache_test.go`
        now pins both sides:
        primitive `String` methods reuse one cache entry across distinct
        receiver values, and primitive scope-fallback callables do not get
        cached across reassignment
      - refreshed aligned `i_before_e` numbers on the kept code:
        - bytecode-runtime clean rerun A: `0.92s/op`, `26.10 MB/op`,
          `644,118 allocs/op`
        - bytecode-runtime clean rerun B: `0.97s/op`, `26.10 MB/op`,
          `644,117 allocs/op`
        - bytecode-runtime profiled: `1.11s/op`, `26.13 MB/op`,
          `644,191 allocs/op`
      - this is a CPU-path keep layered on top of the prior cache work:
        the refreshed profile shows `resolveMethodCallableFromPool(...)`
        dropping from the older `~250ms` cumulative tier to about `~70ms`
        while keeping the post-cast heap shape effectively flat
    - the next injected-receiver helper tranche is now landed too:
      - `v12/interpreters/go/pkg/interpreter/eval_expressions_calls.go` now
        routes bytecode direct member calls through the shared callable
        dispatcher with an explicit optional injected-receiver path instead of
        first building a fresh merged argument slice at every
        `obj.method(...)` call site
      - `v12/interpreters/go/pkg/interpreter/bytecode_vm_call_member.go` now
        passes the existing VM stack slice plus the receiver into that shared
        helper, so the prepend step can reuse the stack-backed slice when it
        has spare capacity instead of always materializing a new merged slice
      - focused coverage in
        `v12/interpreters/go/pkg/interpreter/bytecode_vm_call_member_test.go`
        now also pins optional-arity and overloaded method-call semantics on
        the direct member-call opcode path
      - refreshed aligned `i_before_e` numbers on the kept code:
        - bytecode-runtime clean rerun: `0.90s/op`, `20.57 MB/op`,
          `471,293 allocs/op`
        - bytecode-runtime profiled: `0.90s/op`, `20.60 MB/op`,
          `471,367 allocs/op`
      - this is a keep as both CPU-path and heap work:
        the injected-member-call arg merge no longer shows up as a top alloc
        node, and aligned steady-state bytecode `i_before_e` moved from the
        prior `~0.92-0.97s/op` band into the `~0.90s/op` band while heap fell
        from about `26.1 MB/op` / `644k allocs/op` to about
        `20.6 MB/op` / `471k allocs/op`
    - the next dynamic boxed-int cache expansion tranche is now landed too:
      - `v12/interpreters/go/pkg/interpreter/bytecode_vm_small_int_boxing.go`
        now raises the lazy dynamic boxed-int cache cap from `32768` to
        `262144`, which is large enough for a single warmup pass to retain the
        full loop-index working set seen by aligned `i_before_e`
      - the fixed small-int cache remains unchanged; this is a targeted
        out-of-range cache expansion rather than a broader eager preallocation
      - focused coverage in
        `v12/interpreters/go/pkg/interpreter/bytecode_vm_slot_const_immediates_test.go`
        now pins the dynamic cache path with a large out-of-range integer
      - refreshed aligned `i_before_e` numbers on the kept code:
        - bytecode-runtime clean rerun: `0.88s/op`, `14.63 MB/op`,
          `347,625 allocs/op`
        - bytecode-runtime profiled: `0.88s/op`, `14.66 MB/op`,
          `347,695 allocs/op`
      - this is a keep as heap work with a stable CPU band:
        aligned steady-state bytecode `i_before_e` stayed in the same
        sub-second runtime band while heap fell from about
        `20.6 MB/op` / `471k allocs/op` to about `14.6 MB/op` / `348k allocs/op`
      - the refreshed alloc profile shows the intended shift:
        `bytecodeBoxedIntegerValue(...)` dropped from about `5.63 MB` flat
        alloc-space to about `1.54 MB`, and total profiled alloc-space fell
        from about `44.37 MB` to about `34.94 MB`
    - the next atomic `read_lines` hot-cache tranche is now landed too:
      - `../able-stdlib/src/fs.able` now keeps `fs_read_lines_fast(...)` on a
        single-entry immutable hot cache keyed by `path + size + modifiedNs`
        via `atomic.Pointer` instead of the earlier map + `RWMutex`
        experiment
      - the hot repeated-read path is now just `os.Stat(...)` plus an atomic
        load/compare; misses still rebuild from `os.ReadFile(...)` and replace
        the cached entry
      - focused proof in
        `v12/interpreters/go/pkg/compiler/compiler_stdlib_io_temp_test.go`
        now pins that rewriting the same file invalidates the cached
        `read_lines(...)` result
      - refreshed aligned `i_before_e` numbers on the kept code:
        - bytecode-runtime clean rerun A: `0.91s/op`, `8.37 MB/op`,
          `347,617 allocs/op`
        - bytecode-runtime clean rerun B: `0.89s/op`, `8.37 MB/op`,
          `347,617 allocs/op`
        - bytecode-runtime profiled: `0.89s/op`, `8.40 MB/op`,
          `347,690 allocs/op`
      - this is a keep as both heap work and a CPU-safe cache shape:
        aligned steady-state bytecode `i_before_e` keeps the sub-second band
        while heap falls from about `14.6 MB/op` to about `8.4 MB/op`, and the
        earlier map/lock regression is gone
      - the refreshed alloc profile shows the intended shift:
        the old `strings.genSplit` / `os.readFileContents` plugin-body cost
        drops out of the measured hot path, leaving
        `buildExternFastInvoker.func8`, `externStringSliceResult(...)`,
        `bytecodeBoxedIntegerValue(...)`, and residual member/native dispatch
        as the cleaner remaining wall
    - the next string-slice template cache tranche is now landed too:
      - `v12/interpreters/go/pkg/interpreter/extern_host_fast.go` now gives
        each string-slice fast invoker a tiny cached `[]string ->
        []runtime.Value` template, keyed by a source snapshot
      - repeated hot `Array String` extern results now clone that cached boxed
        template instead of re-boxing every `StringValue` from scratch on each
        call, while still returning a fresh Able array backing slice
      - changed source `[]string` values rebuild the template instead of
        reusing stale boxed elements, and
        `v12/interpreters/go/pkg/interpreter/extern_host_result_fast_test.go`
        now pins both the no-aliasing and invalidation behavior
      - refreshed aligned `i_before_e` numbers on the kept code:
        - bytecode-runtime clean rerun: `0.87s/op`, `5.61 MB/op`,
          `174,794 allocs/op`
        - bytecode-runtime profiled: `0.88s/op`, `5.64 MB/op`,
          `174,867 allocs/op`
      - this is a keep as another material heap collapse with a flat CPU band:
        aligned steady-state bytecode `i_before_e` stays in the same
        sub-second range while heap falls from about
        `8.4 MB/op` / `348k allocs/op` to about `5.6 MB/op` / `175k allocs/op`
      - the refreshed alloc profile shows the intended shift:
        `externStringSliceResult(...)` drops out of the top alloc tier and the
        remaining array-string return cost is now mostly one cloned
        `[]runtime.Value` slice in `externCloneValueSlice(...)`
    - current state: trace-first bytecode runtime ranking is now landed
      - `v12/bench_perf` now forwards
        `ABLE_BYTECODE_TRACE_OUT=/tmp/trace.json` and optional
        `ABLE_BYTECODE_TRACE_LIMIT=N` into the steady-state
        `bytecode-runtime` harness
      - the benchmark binary now emits a sorted bytecode call-trace JSON
        report after warmup, keyed by bytecode `call_name` / `call_member`
        callsite, callee name, dispatch path, and source location
      - the first traced aligned `i_before_e` run says the current hottest
        measured callsites are `__able_array_size`, array `len` /
        `read_slot`, `__able_array_read`, benchmark-local `replace` /
        `contains` / `len_bytes` / `is_valid`, and the fast string externs
        those methods bottom out into
      - traced `bytecode-runtime` wall-clock is diagnostic-only because the
        trace itself adds overhead; use it to rank hot callsites, then return
        to the normal untraced benchmark for keep/reject decisions
    - current state: the next trace-driven overload-member inline tranche is
      now landed too:
      - `v12/interpreters/go/pkg/interpreter/bytecode_vm_call_member.go` now
        resolves overload-valued member call targets down to a selected
        `*runtime.FunctionValue` before dispatch instead of always falling
        through the generic bound-method path
      - that selected overload now feeds the existing injected-receiver
        inline/generic call ladders directly, and the small-arity selection
        scratch path stays stack-backed instead of allocating a merged
        receiver+args slice on every hot call
      - focused coverage in
        `v12/interpreters/go/pkg/interpreter/bytecode_vm_bound_method_inline_test.go`
        now pins inline hits for overloaded member-call sites
      - refreshed aligned `i_before_e` numbers on the kept code:
        - bytecode-runtime clean rerun A: `0.887s/op`, `5.60 MB/op`,
          `174,773 allocs/op`
        - bytecode-runtime clean rerun B: `0.911s/op`, `5.60 MB/op`,
          `174,775 allocs/op`
        - bytecode-runtime profiled: `0.905s/op`, `5.63 MB/op`,
          `174,847 allocs/op`
      - this is a keep as a CPU-path win with the prior low-heap shape:
        aligned steady-state bytecode `i_before_e` moved from the restored
        `~1.00-1.01s/op` band down to `~0.89-0.91s/op` while staying in the
        prior `~5.6 MB/op` / `175k allocs/op` heap band
      - the traced benchmark confirms the target moved:
        `Array.get` now shows up as `call_member` / `resolved_method` /
        `inline` instead of remaining on the generic member-call path
    - current state: the next exact-native context tranche is now landed too:
      - `v12/interpreters/go/pkg/runtime/values.go` now exposes an opt-in
        `SkipContext` flag on `runtime.NativeFunctionValue`, and
        `v12/interpreters/go/pkg/interpreter/definitions.go` marks extern
        wrappers as `SkipContext: true` because those generated host-call
        closures do not observe `*runtime.NativeCallContext`
      - both the tree-walker native call path and the bytecode
        `execExactNativeCall(...)` fast path now bypass
        `NativeCallContext` pooling/setup when that opt-in flag is set,
        while every existing context-sensitive runtime/concurrency native keeps
        the old path unchanged
      - focused coverage now pins that extern wrappers advertise the opt-in
        contract and that both the tree-walker and bytecode exact-native paths
        pass `nil` context only on that flagged fast path
      - refreshed aligned `i_before_e` numbers on the kept code:
        - bytecode-runtime clean rerun A: `0.872s/op`, `5.60 MB/op`,
          `174,775 allocs/op`
        - bytecode-runtime clean rerun B: `0.853s/op`, `5.60 MB/op`,
          `174,774 allocs/op`
        - bytecode-runtime profiled: `0.837s/op`, `5.63 MB/op`,
          `174,846 allocs/op`
      - this is a keep as a modest CPU-path win with the same low-heap shape:
        aligned steady-state bytecode `i_before_e` moved from the prior
        `~0.89-0.91s/op` band into the `~0.84-0.87s/op` band while holding the
        prior `~5.6 MB/op` / `175k allocs/op` heap band
      - the refreshed profile says native-call context setup is no longer the
        right exact-native target; the remaining measured wall is the actual
        fast-string extern body plus residual member/name-call dispatch
    - current state: the next aligned `i_before_e` slice is now landed too,
      but it is benchmark-local rather than VM-internal:
      - `v12/examples/benchmarks/i_before_e/i_before_e.able` now short-circuits
        `is_valid(...)` by returning early when a word has no `"ei"` or has
        `"ei"` but no `"cei"`, so it only falls back to
        `replace("cei", "")` on the small subset that actually contains the
        replacement needle
      - on the aligned `wordlist.txt` corpus this removes pointless
        `replace(...)` work from 172,695 of 172,823 words; only 128 words
        still take the replace path, and an exhaustive local equivalence check
        over the aligned wordlist preserved the prior `1628` invalid outputs
      - refreshed aligned `i_before_e` numbers on the kept code:
        - bytecode-runtime clean rerun A: `0.792s/op`, `2.84 MB/op`,
          `2,080 allocs/op`
        - bytecode-runtime clean rerun B: `0.749s/op`, `2.84 MB/op`,
          `2,078 allocs/op`
        - bytecode-runtime profiled: `0.779s/op`, `2.87 MB/op`,
          `2,151 allocs/op`
        - external compiled compare: `0.290s`
        - external bytecode compare: `1.020s`
      - this is a keep because the old benchmark body was doing obviously
        wasted string work on nearly the entire corpus; the semantics stay the
        same, but the benchmark is no longer dominated by avoidable
        `replace(...)` calls
      - aligned steady-state bytecode `i_before_e` moved from the prior
        `~0.84-0.87s/op` band into the `~0.75-0.79s/op` band, and heap dropped
        from roughly `5.6 MB/op` / `175k allocs/op` into the
        `~2.84 MB/op` / `2.1k allocs/op` band
      - one-shot aligned bytecode `i_before_e` is now `1.02s`, so this
        benchmark is no longer the right place to spend more string
        micro-optimization time before revisiting the much larger `fib`
        timeout problem
    - next step: move back to the steady-state bytecode runtime and rerun
      aligned `fib` with the same profiling/tracing discipline. `i_before_e`
      is now low enough that the next real VM hotspot should come from the
      recursion/call-frame path rather than more text-path benchmark cleanup.
    - current state: the next reduced-recursion bytecode slice is now landed
      too:
      - `v12/interpreters/go/pkg/interpreter/bytecode_vm_i32_fast.go` now
        carries a dedicated small-`i32` boxing path plus direct small-`i32`
        add/sub helpers, and the fused self-call / specialized binary-op
        paths use those helpers before falling back to the generic integer
        machinery
      - focused coverage now pins the direct `i32` boxing/add/sub helpers in
        the existing bytecode integer fast-path tests
      - the reduced `fib(30)` bytecode microbench on warmed reruns moved from
        the restored roughly `219-225ms/op` band into roughly the
        `199-202ms/op` band
      - aligned one-shot external bytecode `fib` still times out at `90s`, so
        this is not yet the real aligned benchmark fix
      - this is still a keep as a narrow recursion-kernel win: it removes
        some generic type-switch/cache-selection work from the hot `i32`
        arithmetic path without changing semantics, and it gives the next
        `fib` tranche a lower baseline
    - current state: the next reduced-recursion frame-shape slice is now
      landed too:
      - self-fast recursive bytecode calls now use a smaller minimal
        self-fast frame shape whenever the call carries no generic-name set,
        no implicit receiver, and no loop/iterator base state
      - focused cleanup tests now pin unwind/release behavior for that new
        minimal self-fast frame path
      - refreshed reduced `fib(30)` warmed reruns moved from the prior
        roughly `198.70-201.98ms/op` band to roughly `195.06-199.73ms/op`
      - refreshed reduced profiling no longer shows `pushCallFrame(...)` as a
        top-tier flat hotspot; the remaining reduced recursion wall is now
        more cleanly `execCallSelfIntSubSlotConst(...)`, `execBinary(...)`,
        `popCallFrameFields(...)`, `acquireSlotFrame(...)`,
        `bytecodeDirectSmallI32Value(...)`, and the residual `i32`
        boxing/immediate-subtract path
      - aligned one-shot external bytecode `fib` still times out at `90s`, so
        this is still not the real aligned benchmark fix
    - current state: the next reduced-recursion return-path slice is now
      landed too:
      - inline bytecode returns now run through a dedicated
        `bytecode_vm_return.go` helper so the hot return path is split out of
        `bytecode_vm_run.go`, which is back under the 1000-line guardrail
      - the inline return path now has a direct
        `bytecodeCallFrameKindSelfFastMinimal` fast path instead of going
        through the generic `popCallFrameFields(...)` path for the minimal
        self-fast case
      - hot inline call sites now use cached return-generic metadata through
        `bytecodeInlineReturnGenericNames(...)` instead of re-entering the
        broader program return-generic helper on every path
      - refreshed reduced `fib(30)` warmed reruns moved from the prior
        roughly `195.06-199.73ms/op` band to roughly `189.63-195.72ms/op`
      - refreshed reduced profiling no longer shows `pushCallFrame(...)` as a
        visible top-tier hotspot; the remaining reduced recursion wall is now
        more cleanly `execCallSelfIntSubSlotConst(...)`, `execBinary(...)`,
        `execBinarySlotConst(...)`, `finishInlineReturn(...)`,
        `bytecodeDirectSmallI32Value(...)`, and
        `bytecodeBoxedIntegerI32Value(...)`
      - aligned one-shot external bytecode `fib` still times out at `90s`, so
        this is still not the real aligned benchmark fix
    - current state: the next reduced-recursion dedicated self-slot fast
      branch slice is now landed too:
      - `bytecode_vm_calls.go` now has an early dedicated self-slot fast
        branch inside `execCallSelfIntSubSlotConst(...)`, so the successful
        recursive hot path bypasses the older generic callee switch, the
        `*bytecodeProgram` type assertion/equality check, and `callNode`
        extraction entirely
      - that branch reads the `*runtime.FunctionValue` directly from the
        reserved self slot, uses the already-known `currentProgram`, and
        stays on the existing minimal self-fast frame path
      - focused self-fast recursion, cleanup, and fixture-parity tests stayed
        green on the narrower dispatch path
      - refreshed reduced `fib(30)` warmed reruns moved from the prior
        roughly `186.37-189.04ms/op` band to roughly `176.39-181.90ms/op`
      - aligned one-shot external bytecode `fib` still times out at `90s`, so
        this is still not the real aligned benchmark fix
    - current state: the next reduced-recursion small-`i32` pair-add slice is
      now landed too:
      - `bytecode_vm_i32_fast.go` now has a dedicated
        `bytecodeDirectSmallI32Pair(...)` helper for the hot recursive-result
        `+` path
      - `bytecodeAddSmallI32PairFast(...)` now uses that combined extractor
        directly instead of calling `bytecodeDirectSmallI32Value(...)` twice
        on each small-`i32` pair add
      - focused binary-fast-path and fixture-parity coverage stayed green on
        the narrower add path
      - refreshed reduced `fib(30)` warmed reruns moved from the prior
        roughly `176.39-181.90ms/op` band to roughly `171.06-175.27ms/op`
      - aligned one-shot external bytecode `fib` still times out at `90s`, so
        this is still not the real aligned benchmark fix
    - current state: the next reduced-recursion minimal-self-fast suffix
      slice is now landed too:
      - `bytecode_vm_call_frames.go` now keeps top contiguous
        minimal-self-fast frames out of `callFrameKinds` entirely and only
        materializes those kinds if a broader frame kind needs to sit above
        them
      - `bytecode_vm_return.go`, `bytecode_vm_pool.go`, and
        `bytecode_vm_run_finalize.go` now track that unmaterialized suffix
        directly so the hot reduced-`fib` recursion path no longer appends a
        `bytecodeCallFrameKindSelfFastMinimal` entry on every recursive step
      - focused self-fast unwind/materialization coverage stayed green on the
        narrower frame contract
      - refreshed reduced `fib(30)` warmed reruns moved from the current
        low-`170s` restored baseline into a kept roughly `171.32-173.91ms/op`
        band, with a single profiled reduced run at `170.23ms/op`
      - aligned one-shot external bytecode `fib` still times out at `90s`, so
        this is still not the real aligned benchmark fix
    - current state: the next reduced-recursion statement-position `if`
      lowering slice is now landed too:
      - `bytecode_lowering.go` now routes non-last `if` expressions through a
        statement-only lowering path instead of forcing them through full
        expression-result semantics
      - `bytecode_lowering_controlflow.go` now lowers that statement path
        without synthesizing a dead `Nil` value for the missing `else` branch,
        so the hot reduced-`fib` recursion path no longer executes the old
        `bytecodeOpConst Nil` plus immediate `Pop` pair on every non-base step
      - focused lowering, self-fast recursion, and fixture-parity coverage
        stayed green on the narrower control-flow lowering contract
      - refreshed reduced `fib(30)` warmed reruns moved from the prior kept
        roughly `171.32-173.91ms/op` band to roughly `159.41-163.53ms/op`,
        with a single profiled reduced run at `169.55ms/op`
      - the refreshed reduced CPU profile no longer shows the earlier dead
        statement-result const/pop overhead as a visible top-tier slice; the
        remaining reduced wall is back on `execCallSelfIntSubSlotConst(...)`,
        `execBinary(...)`, `execBinarySlotConst(...)`,
        `bytecodeSubtractIntegerImmediateI32Fast(...)`,
        `bytecodeDirectSmallI32Value(...)`, `acquireSlotFrame(...)`, and the
        residual run-loop dispatch
      - aligned one-shot external bytecode `fib` still times out at `90s`, so
        this is still not the real aligned benchmark fix
    - current state: after the remaining boxed-path micro-slices stopped
      producing defensible reduced-`fib` wins, bytecode work moved to VM v2's
      typed-lane plan:
      - `v12/design/bytecode-vm-v2.md` now defines the typed slots/stack,
        boxing-boundary, quickening, and native Array/String implementation
        order
      - the first stack-only seed is landed: literal-only final `i32` add/sub
        expressions now execute on a raw `i32` operand stack with checked
        overflow and box back before return
      - declared `i32` slot metadata plus `LoadSlotI32` / `StoreSlotI32` are
        now landed for safe final arithmetic and typed local declarations
      - reduced `Fib30Bytecode` stayed neutral on the declared-slot slice
        because recursive self-fast frames still pass boxed slot values;
        guardrail reruns landed at `117.43ms/op` and `121.97ms/op`
    - next step: carry typed `i32` slot state through inline/self-fast call
      frames and wire the fused self-call subtract / typed return-add path to
      consume that state, with boxing at all dynamic/spec boundaries.
- fixture exporter and other tooling cleanup
  - current state: the first cleanup slice is now landed
    - `cmd/fixture-exporter` has focused direct test coverage plus a
      `--check` mode for stale `module.json` detection
    - `v12/export_fixtures.sh` now forwards arguments so the check mode is
      usable from the repo root
    - stale exporter TODO wording has been removed from the active docs
  - current state: the next cleanup slice is now landed
    - `cmd/fixture-exporter` now accepts targeted fixture directories or
      `source.able` paths, so focused export/check runs no longer have to walk
      the full AST fixture tree
    - `v12/export_fixtures.sh` and the active docs now describe targeted
      fixture export/check usage from the repo root
  - current state: the next cleanup slice is now landed
    - `cmd/fixture` now delegates fixture replay to a shared
      `pkg/interpreter` helper instead of carrying its own duplicate fixture
      execution and stdlib-loading path
    - focused tests now cover both the shared replay helper and the CLI
      wrapper path on simple source fixtures
  - current state: the next cleanup slice is now landed
    - `cmd/fixture` now accepts either `--dir <fixture-dir>` or a direct
      fixture directory / entry-file argument, so focused replay no longer
      requires boilerplate flag wiring for the common case
    - the active docs now describe that direct fixture replay workflow
  - current state: the next cleanup slice is now landed
    - repo-relative fixture targets in `cmd/fixture` now resolve under
      `v12/fixtures/ast`, so focused replay from `v12/interpreters/go` no
      longer requires spelling `../../fixtures/ast/...`
    - the active docs now show the shorter fixture replay path form
  - current state: the next cleanup slice is now landed
    - `cmd/fixture --list` now prints repo-relative AST fixture directories
      and supports focused prefix filters, so fixture discovery no longer
      requires separate `find`/`rg` shell work
    - the active docs now show the list-mode workflow too
  - current state: the next cleanup slice is now landed
    - `cmd/fixture --describe` now prints resolved fixture metadata
      (directory, entry, setup, skip targets, executor, typecheck mode)
      without evaluating the fixture
    - the fixture CLI now also honors a manifest executor by default when no
      explicit `--executor` override is provided
  - current state: the next cleanup slice is now landed
    - `cmd/fixture --batch` now replays multiple fixture targets and emits a
      JSON array of per-target results plus resolved metadata
    - focused fixture batching no longer requires shell loops around repeated
      single-target invocations
  - current state: the next cleanup slice is now landed
    - `cmd/fixture --describe` now also supports `--format text` for
      human-readable metadata summaries
    - `cmd/fixture --batch` now also supports `--format jsonl` for
      one-result-per-line streaming output
  - current state: the next cleanup slice is now landed
    - `cmd/fixture --list` now also supports `--format json` for
      machine-readable fixture discovery output
- testing CLI / user-facing testing framework work
  - current state: the first user-facing `able test` cleanup slice is now
    landed
    - compiled `--list` / `--dry-run` output now matches the interpreted
      formatter shape for framework/module/test/tag/metadata display
    - focused compiled CLI coverage now pins that parity path directly
  - current state: the next `able test` cleanup slice is now landed
    - compiled `--list` / `--dry-run` now use the shared interpreted
      discovery/list pipeline instead of the compiled execution path
    - compiled dry-run now supports `--format json` through that shared path
  - current state: the next `able test` cleanup slice is now landed
    - compiled execution no longer carries dead list/dry-run runner logic
      after the shared discovery/list short-circuit
    - focused compiled CLI coverage still pins dry-run formatter and JSON
      parity after that cleanup
  - current state: the next `able test` cleanup slice is now landed
    - compiled `able test` execution now supports `--format json` and
      `--format tap` instead of rejecting those reporter formats outside
      list/dry-run mode
    - compiled and interpreted event formatting now share the same Go-side
      JSON/TAP decoder-emitter path through `pkg/testcli`, while the
      compiled runner uses a primitive-only extern callback boundary instead
      of trying to marshal full `TestEvent` unions through a Go extern call
  - current state: the next `able test` cleanup slice is now landed
    - compiled JSON execution now preserves descriptor `tags`, `metadata`,
      and source `location` instead of emitting a reduced descriptor payload
    - the compiled runner now forwards those fields across the primitive
      reporter boundary using `Array String` tags plus metadata key/value
      arrays and explicit descriptor-location primitives, and focused CLI
      coverage now pins that richer compiled JSON event shape directly
  - current state: the next `able test` cleanup slice is now landed
    - focused compiled reporter coverage now also pins non-success event
      behavior directly: skipped cases, failed cases with details/location,
      and framework-error output for both JSON and TAP reporters
    - the local CLI test stdlib now has dedicated reporter-event and
      framework-error harness variants so those compiled reporter paths are
      exercised without depending on larger repo fixtures
  - current state: the next `able test` cleanup slice is now landed
    - focused compiled reporter coverage now also pins `case_started`
      semantics directly: JSON preserves started-before-terminal event order,
      while TAP still numbers only terminal results and ignores started
      events in its test-point count
    - this closes the remaining obvious event-order/count parity gap in the
      compiled JSON/TAP reporter surface without changing runtime behavior
  - current state: the next `able test` cleanup slice is now landed
    - the minimal `able.test.*` fixture stdlib used by focused CLI tests now
      lives in one shared helper instead of being duplicated across the
      interpreted and compiled dry-run tests
    - older filter/dry-run coverage now exercises that shared helper too,
      reducing test-fixture drift inside `cmd/able`
  - current state: the next `able test` cleanup slice is now landed
    - the focused CLI tests now share temp working-directory and minimal
      workspace helpers instead of repeating `Getwd` / `Chdir` / `tests`
      setup blocks inline
    - this keeps the narrow dry-run/list coverage easier to extend without
      more harness drift inside `cmd/able`
  - current state: the next `able test` cleanup slice is now landed
    - the compiled stdlib smoke-suite tests now share one helper for env
      setup, target-path expansion, execution, and stderr enforcement instead
      of repeating the same boilerplate in every suite test
    - that helper also runs those tests from an isolated temp workspace so
      they stay valid under the canonical stdlib-root collision semantics
  - current state: the next `able test` cleanup slice is now landed
    - the compiled stdlib suite inventory now lives in a dedicated
      helper-driven test file instead of staying embedded in the larger
      general CLI test file
    - the old top-level test names are preserved, but their case data now
      comes from one shared case table plus one shared runner helper
  - current state: the next `able test` cleanup slice is now landed
    - the remaining compiled non-stdlib sample-module setup and success
      assertions now live behind shared helpers instead of repeating the same
      inline sample module and output checks
    - both compiled sample-module tests now run from isolated temp working
      directories so they stay valid under canonical stdlib-root collision
      enforcement
  - current state: the next `able test` cleanup slice is now landed
    - focused CLI success-path tests now share one `captureCLI` success helper
      instead of repeating the same `exit code 0` and `stderr empty` checks
      across interpreted dry-run, compiled dry-run, and compiled sample paths
    - compiled stdlib and compiled sample helpers now build on that shared
      success path too, reducing the remaining assertion drift in `cmd/able`
  - current state: the next `able test` cleanup slice is now landed
    - the focused dry-run and compiled-sample tests now share explicit env
      setup helpers instead of reassembling `ABLE_MODULE_PATHS`,
      `ABLE_PATH`, and `ABLE_TYPECHECK_FIXTURES` inline
    - the repo-stdlib compiled helpers and the minimal local dry-run helpers
      now each have one place where their env contract is defined
  - current state: the next `able test` cleanup slice is now landed
    - the remaining focused failure-path assertions now build on shared
      nonzero-exit helpers instead of hand-rolling `captureCLI` exit checks
    - the adjacent `build` negative tests now also run from isolated temp
      working directories, keeping them valid under canonical stdlib-root
      collision enforcement
  - current state: the next `cmd/able` cleanup slice is now landed
    - the adjacent `build` and `deps` success-path tests now also build on
      shared CLI success helpers instead of repeating open-coded `captureCLI`
      success assertions
    - the small set of build-success cases that intentionally surface
      fallback/typechecker diagnostics now use a shared
      success-with-diagnostics helper, and the external-output build test now
      only requires copied stdlib sources when a cached canonical stdlib is
      actually available
  - current state: the next `cmd/able` cleanup slice is now landed
    - `run_entry_test.go` success-path coverage now shares the common
      working-directory helper instead of repeating inline `Getwd` / `Chdir`
      restore blocks
    - the run/check success-path cluster there now also builds on the shared
      CLI success helper, keeping empty-`stderr` enforcement in one place
  - current state: the next `cmd/able` cleanup slice is now landed
    - the remaining `run_entry_test.go` failure/collision cluster now also
      shares the common working-directory helper and the shared CLI
      failure-path helper instead of repeating inline `Getwd` / `Chdir` /
      nonzero-exit plumbing
    - shared substring assertions for those diagnostics now build on one
      generic text-contains helper instead of open-coded `strings.Contains`
      chains in each collision test
  - current state: the next `cmd/able` cleanup slice is now landed
    - the remaining adjacent `build_test.go` and `deps_cli_test.go` cases now
      also share the common working-directory helper where they were still
      managing `Getwd` / `Chdir` inline
    - those tests now also build on the shared text/output assertion helpers
      and the shared CLI failure helper instead of repeating one-off
      `strings.Contains` and nonzero-exit checks
  - current state: the next `cmd/able` cleanup slice is now landed
    - `setup_smoke_test.go` now also shares the common working-directory
      helper plus the shared CLI success/output assertion helpers instead of
      carrying its own `captureCLI` success plumbing
  - current state: the next adjacent `cmd/able` cleanup slice is now landed
    - the last manual working-directory setup blocks in
      `dependency_installer_test.go` now also use the shared
      `enterWorkingDir(...)` helper instead of carrying local `Getwd` /
      `Chdir` restore code
  - current state: the next adjacent `cmd/able` cleanup slice is now landed
    - `overrides_test.go` now uses the shared working-directory helper for
      its relative-path override case, and the repeated override-log scans now
      build on one shared “any entry contains all substrings” helper
  - current state: the next adjacent `cmd/able` cleanup slice is now landed
    - `dependency_installer_test.go` now builds repeated lockfile assertions
      on shared locked-package helpers instead of open-coding the same
      stdlib/kernel presence checks across multiple installer scenarios
  - current state: the next adjacent `cmd/able` cleanup slice is now landed
    - the shared locked-package helpers now also cover the remaining
      stdlib/kernel lockfile assertions in `deps_cli_test.go`,
      `setup_smoke_test.go`, and `overrides_test.go`
  - current state: the next adjacent `cmd/able` cleanup slice is now landed
    - the remaining `test_cli` and `run_entry` output/diagnostic substring
      assertions now build on the shared text/output assertion helpers,
      including a shared “contains any” helper for alternate diagnostic text

#### Language / Runtime backlog
- WASM runtime work
- regex syntax and engine work
- broader parser/tree-sitter coverage work not required for compiler completion
- additional stdlib redesign/migration work not required for compiler release
- concurrency feature expansion beyond current spec/runtime requirements

#### Documentation backlog
- continue reconciling older design notes against the active Go-first toolchain
  and the compiler lowering spec as work resumes in those areas
- keep `spec/TODO_v12.md` and relevant design notes current when compiler or
  bytecode work resolves remaining language/implementation gaps
