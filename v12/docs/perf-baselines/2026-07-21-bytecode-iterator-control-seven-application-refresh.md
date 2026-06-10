# Bytecode iterator/control seven-application refresh

## Outcome

This refresh retains no bytecode VM, runtime, compiler, canonical stdlib,
benchmark, fixture, language, or WASM change. Array Slice Window, Dependency
Plan, Document Audit, Lexical Rollup, and Option/Result Configuration were
profiled with Mandelbrot and Word Frequency as unlike numeric and text/map
controls.

Two independent high-sample CPU processes and two independent exact first-
`main()` allocation processes completed for every row from one frozen current
artifact. Array Slice Window adds required independent Array copying, cast,
integer, and Array-ownership work, but it does not add a new concrete operation
to the cached-member/return frame shared by the other iterator/control rows.
The only exact Able leaves material in at least three selected rows are
already-completed member-cache, active-lookup, call/return-frame, and raw-
integer families.

## Reproducibility contract

- Go: `go1.26.4 linux/amd64`
- repository HEAD: `237406eccdfb025a519d898daedadee1c8d13a7b`
- frozen interpreter test binary SHA-256:
  `6758a13355a1adeebe0984098679c3ad344b0f8cf1a8642694e873e3dd12d53e`
- canonical stdlib tree SHA-256:
  `7c2910abf320846b7aa8f95fe336be830bb7d69b9b36a10172cbcaa134dc05aa`
- CPU 0, `GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`, serial executor,
  and a 59-second ceiling for every process
- catalog run directory and arguments, including source-root-only isolation
  where declared
- program load, typecheck, lowering, and one warm `main()` before each CPU
  measured region
- fixed CPU call counts chosen to produce roughly five seconds or more of
  samples per process: Array Slice 10, Dependency 20, Document 500, Lexical
  40, Option/Result 7, Mandelbrot 1, and Word Frequency 4
- exact allocation processes used the existing one-main retention probe after
  load/typecheck; sampled heap profiles were ownership supplements, not exact
  allocation totals

The first one-call CPU attempt was discarded before selection because Document
Audit produced only 20 ms of merged samples. Replacing it with fixed repeated-
main counts is the established bounded method and prevents 10 ms profiler
quantization from selecting a candidate.

## CPU and allocation measurements

CPU timings include profiling overhead and are diagnostic arithmetic means of
the two processes. Each row's profiles nevertheless contain 13.55-17.53
seconds of merged samples. Exact allocation values are arithmetic means of two
separate processes; their durations are shown only to document workstation
spread.

| Application | CPU 1 ns/op | CPU 2 ns/op | CPU mean ns/op | Exact duration mean ns | Bytes mean | Allocs mean |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Array Slice Window | 759,373,349 | 1,010,354,555 | 884,863,952 | 543,871,830 | 17,000,376 | 415,169 |
| Dependency Plan | 338,002,445 | 506,457,475 | 422,229,960 | 307,216,734 | 22,482,304 | 340,042 |
| Document Audit | 13,323,580 | 13,919,040 | 13,621,310 | 22,772,251 | 880,160 | 3,238 |
| Lexical Rollup | 166,216,079 | 250,735,011 | 208,475,545 | 137,405,667 | 3,397,520 | 32,251 |
| Option/Result Configuration | 1,076,016,763 | 1,323,251,700 | 1,199,634,232 | 736,438,630 | 76,888,384 | 1,305,328 |
| Mandelbrot control | 6,656,503,621 | 8,966,230,430 | 7,811,367,026 | 6,462,406,827 | 615,170,440 | 76,303,277 |
| Word Frequency control | 1,975,311,916 | 1,649,373,421 | 1,812,342,669 | 1,191,422,084 | 54,255,848 | 625,862.5 |

The workstation caused material CPU-time movement in several profiled rows,
so both processes remain in every mean. Allocation identities stayed exact or
near-exact: Dependency, Lexical, and Option/Result matched exactly; Array Slice
differed by 32 bytes; Word Frequency differed by seven allocations.

## Exact flat-leaf intersection

At a one-percent flat CPU threshold, the complete Able-owned intersection in
three or more selected applications is:

| Exact symbol | Selected rows | Flat shares | Disposition |
| --- | ---: | --- | --- |
| `(*bytecodeVM).runResumable` | 5 | 2.22%-7.64% | dispatcher parent; exact-line gate closed |
| `(*bytecodeVM).switchRunProgramWithActiveLookupState` | 4 | 1.03%-1.55% | retained active-lookup family |
| `(*bytecodeVM).execCallMember` | 3 | 1.43%-2.21% | call/member parent |
| `(*bytecodeVM).lookupCachedMemberMethodEntry` | 3 | 2.80%-4.94% | completed dependency-validated member cache |
| `(*bytecodeVM).finishInlineReturn` | 3 | 1.33%-1.70% | completed/rejected return family |
| `(*bytecodeVM).popCallFrameFields` | 3 | 1.33%-1.88% | completed/rejected frame family |
| `bytecodeRawIntegerValueInfo` | 3 | 1.15%-1.77% | completed raw-integer family; also 1.46% in Word Frequency |

Array Slice Window's material exact children instead include primitive target
lookup/casting, raw integer transport, small-integer boxing, Array member
execution, and independent Array backing/ownership. Option/Result is dominated
by allocation, GC, union/type construction and matching. Dependency, Document,
and Lexical carry the cached member/return family, but the preceding current
gate already proved repeated dependency validation and rejected the safe same-
parent shortcut across broad guards. This refresh supplies no new invalidation
fact.

## Go map/hash and allocation reconciliation

`internal/runtime/maps.ctrlGroup.matchH2` and `aeshashbody` are flat in all five
selected programs. Caller traces show different semantic stores:

- Array Slice: primitive-type tables, alias/type caches, Array handle/lease
  tracking, and Array cleanup;
- Dependency: member caches, environment identifiers, mono-Array read/write
  state, and Array handles;
- Document and Lexical: member/name/environment caches plus text operations;
- Option/Result: union/type/interface caches and GC-heavy type construction.

Treating the Go leaf as one Able lookup would conflate different key shapes,
lifetimes, and invalidation contracts. The sampled allocation profiles likewise
split among required slice backing, graph/Queue arrays, text/iterator values,
lexical iterator work, and generic-union/type construction. Bootstrap cache
samples are excluded from measured-main candidate selection.

No new concrete CPU or allocation operation is material in three unlike
iterator/control families. Candidate admission therefore closes without an
A/B prototype.

## Correctness and public verification

- All seven normal CLI bytecode processes completed and passed their
  independent public-output verifiers.
- `go test ./pkg/interpreter -run TestBytecode -count=1 -timeout 60s` passed
  in 25.746 seconds.
- Frontier, evidence, JSON, scorecard, and diff checks are run after this
  record is incorporated.

## Next recommendation

Run a feature-interaction coverage-depth tranche and add the highest-value
portable application selected by it, rather than reopening another closed VM
helper.

Why: after this refresh every selected bytecode ownership group has current
exact evidence or a completed candidate gate, and the current frontier has no
actionable local group. The interaction matrix has no uncovered broad pair,
but many critical combinations—especially lexical bindings/patterns with
closures, methods, interfaces, Result/error handling, program entry, and
stdlib protocols—have depth one and are represented only by Concurrent Event
Routing. One application should not be the sole evidence for that much of the
language surface.

What it entails: recompute feature depth across all current applications,
select the weakest high-value interaction cluster, then build one bounded
source-equivalent Able/Go/Python/Ruby application with an independent verifier
and representative input. Profile both Able product modes and promote it only
after five verifier-backed processes per implementation. A runtime/compiler
candidate remains conditional on a new exact non-nominal leaf appearing in at
least three unlike applications. This expands real-program evidence without
adding benchmark-specific optimization or beginning WASM work.
