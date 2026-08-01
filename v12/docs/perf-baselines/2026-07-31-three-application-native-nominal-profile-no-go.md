# Three-application native-nominal profile closure

Date: 2026-07-31

## Decision

Retain no production code.

Fresh interpreter-free CPU and exact-allocation profiles for Binary Event Log,
Concurrent Document Pipeline, and Versioned Telemetry Pipeline expose no exact
general compiler or generated-runtime owner that is material and removable in
all three.

Heap-backed nominal records are the only broad parent. Their exact descendants
are different:

- Binary Event Log constructs `EventRecord` pointers returned through a native
  static union;
- Concurrent Document Pipeline converts `DocumentTask` and `DocumentScore`
  pointers through the explicit runtime-backed Channel payload ABI; and
- Versioned Telemetry Pipeline stores `Sample` pointers in a native specialized
  `Array Sample`.

Replacing all three with copied Go structs would violate v12's mutable
reference semantics. The compiler has no whole-program read-only/identity
proof that could make those copies unobservable. Channel payload conversion is
material only in the concurrent row, while Binary's HashMap bridge and
Telemetry's native interface adapters are likewise row-specific.

Checked integer arithmetic samples in all three, but it is sparse in
Concurrent Document Pipeline and is an explicitly closed route. No candidate
therefore entered an A/B implementation cohort.

No compiler, generated runtime, runtime package, interpreter, bytecode VM,
canonical stdlib, language, dependency, benchmark, fixture, frozen workspace,
or WASM source changed.

## Current products and controls

All applications were regenerated with the current compiler and
`--no-fallbacks`. Each strict binary passed its public verifier, contains 96
dependencies, and omits the exact package
`able/interpreter-go/pkg/interpreter`.

| Application | Current Able / Go | Ratio | Binary bytes | Binary SHA-256 |
| --- | ---: | ---: | ---: | --- |
| Binary Event Log | 0.0900 / 0.0097 s | 9.278x | 13,924,232 | `5466e6ff...5ca21` |
| Concurrent Document Pipeline | 0.0440 / 0.0043 s | 10.233x | 13,623,248 | `4aaaf7e3...ad2f8` |
| Versioned Telemetry Pipeline | 2.0680 / 0.2078 s | 9.952x | 12,949,752 | `4ee47f92...7b642` |

The current `ablec` SHA-256 is
`8b93aa13336d973854f3573c2b7e6fcb751847844c1444f80551dc63481a1d30`.
The locally active toolchain is Go 1.26.5.

Serial applications used CPU 12 and `GOMAXPROCS=1`. Concurrent Document
Pipeline retained its catalog contract: CPUs 12-15 and `GOMAXPROCS=4`. Every
run used `GOGC=50`, `GOMEMLIMIT=1GiB`, and a 58-second process bound.

## Repeated CPU ownership

The two short applications require many independent launches because Go CPU
profiles sample at 100 Hz. Only each generated launcher's registered main
phase was retained and merged:

| Application | Able profiles | Merged samples | Largest current owners |
| --- | ---: | ---: | --- |
| Binary Event Log | 40 | 1.76 s | checked multiply 21.02%; primitive HashMap equality 8.52%; checked add 8.52%; divmod 6.82%; `parse_event` 47.16% cumulative; allocation 13.64% cumulative |
| Concurrent Document Pipeline | 300 | 0.54 s | Channel state 12.96% cumulative; Channel send 18.52%; `score_line` 12.96%; native-to-runtime record conversion and recovery; allocation descendants remain shallow |
| Versioned Telemetry Pipeline | 10 | 19.40 s | checked multiply 25.15%; divmod 17.99% cumulative; checked add 10.88%; allocation 20.93% cumulative; policy adapters 12.16% and 6.24% cumulative |

All 350 Able CPU processes passed their public verifier and reproduced one
stable output hash per application.

Five Go reference profiles per application also passed the same verifiers.
The profiling wrapper repeated only the unchanged reference main while
discarding intermediate output: 100 calls for Binary, 200 for Concurrent
Document, and 10 for Telemetry. Their merged samples were 2.29, 0.72, and
10.97 seconds.

The generated/reference operation comparison is:

| Application | Generated Able | Equivalent Go |
| --- | --- | --- |
| Binary Event Log | checked arithmetic, pointer record in native union, runtime-backed HashMap | machine arithmetic, by-value `(eventRecord, bool)`, native `map[int64]int64` |
| Concurrent Document Pipeline | pointer records encoded into and decoded from runtime Channel storage | `chan task` and `chan score` carry Go values |
| Versioned Telemetry Pipeline | native `[]*Sample`, checked arithmetic, native interface adapters | `[]sample`, machine arithmetic, Go interface calls |

There is no interpreter execution in any row. A `runtime.Value` payload
transition is material for Concurrent Document Pipeline only; it is an
explicit scheduler-service ABI rather than a fallback into the interpreter.

## Exact main allocations

Three lightweight main-phase `MemStats` deltas per Able and Go product avoid
the one-object profiler's serialization cost:

| Application | Able mean bytes / objects / GC | Go mean bytes / objects / GC |
| --- | ---: | ---: |
| Binary Event Log | 9,321,784 / 171,069.67 / 8 | 3,416 / 26 / 0 |
| Concurrent Document Pipeline | 647,554.67 / 10,471 / 0 | 63,957.33 / 62.33 / 0 |
| Versioned Telemetry Pipeline | 430,788,213.33 / 13,325,303 / 351 | 17,096 / 21 / 0 |

One separate `runtime.MemProfileRate=1` shape process per Able application
passed its verifier. Start/end subtraction includes the known profile-writer
allocations, so exact counters above determine totals and generated line
attribution determines shape:

- Binary: 53,248 fresh `EventRecord` pointers are 31.13% of the exact
  allocation count. The remaining material sites are Result/error union
  wrapping, error struct conversion, HashMap storage, and a dynamic integer
  boundary.
- Concurrent Document: 1,024 fresh `DocumentTask` pointers and approximately
  1,028 `DocumentScore` pointers cross Channel send/receive. Runtime struct
  creation, native reconstruction, integer conversion, and String conversion
  together account for most of the 10,471 allocations.
- Telemetry: 13,208,878 fresh `Sample` pointers are 99.13% of the exact
  allocation count. The native specialized carrier is
  `type __able_array_Sample struct { Elements []*Sample }`; no hot Sample
  conversion to `runtime.Value` occurs.

The equivalent Go escape diagnostics and exact counters confirm that none of
the three references allocates one record object per logical item/update.

## Owner matrix and admission

| Candidate parent | Binary | Concurrent Document | Telemetry | Admission |
| --- | --- | --- | --- | --- |
| Checked integer arithmetic | material | sparse | material | rejected; breadth below three and route already closed |
| Nominal pointer allocation | material | material | dominant | broad parent only; distinct union, Channel, and generic-storage lifetimes |
| Runtime nominal conversion | error/HashMap paths | material Channel payload | absent from hot Sample path | breadth one for the shared payload shape |
| HashMap hashing/equality | material | absent | absent | row-specific runtime service |
| Channel scheduler/storage | absent | material | absent | row-specific runtime service |
| Native interface adapters | absent | absent | material | row-specific |
| Allocation/GC | material | exact counts material, CPU shallow | material | aggregate consequence, not one legal lowering |

The apparent Go/Able representation difference is real, but not currently one
sound optimization. Section 5.3.1 of the v12 specification requires all Able
values, including structs, to retain reference semantics: mutation through one
alias must be visible through every alias. Whole-program Go escape analysis
runs too late to authorize merging or copying Able identities.

The previously retained caller-owned nominal result ABI handles proven
nonescaping fresh small results. These current records escape through a union,
concurrent runtime storage, or generic storage. The prior non-capture/effect
gate also records that the compiler has no parameter-retention plus
caller-alias/liveness proof. Reusing or value-copying these objects without
that proof would repeat a closed unsafe route.

Accordingly:

- no generated-source ceiling was mistaken for a legal compiler candidate;
- no named record, Channel, HashMap, or application rule was added;
- no broad execution-context or checked-arithmetic experiment was reopened;
- no baseline/candidate timing cohort was manufactured; and
- the current compiled performance gap remains open.

The machine-readable companion is
`2026-07-31-three-application-native-nominal-profile-no-go.json`.

## Verification and cleanup

- 3/3 strict build smokes passed public verifiers.
- 350/350 Able main-only CPU processes passed.
- 9/9 Able lightweight allocation processes passed.
- 3/3 Able one-object allocation-shape processes passed.
- 15/15 repeated Go CPU processes passed.
- 9/9 Go lightweight allocation processes passed.
- 3/3 Go reference smokes passed.
- Every individual process stayed below one minute.
- Large modules, binaries, caches, and raw profiles lived under disk-backed
  `/var/tmp`, never RAM-backed `/tmp`.
- The exact 494,304 KiB task workspace was removed after the compact evidence
  was recorded.

## Next

Run a bounded read-only nominal-carrier feasibility gate before attempting
another compiled optimization.

Why: this tranche supplies three unlike applications where immutable-by-use
records account for substantial allocation, but the language requires
reference semantics and the current compiler cannot prove that value copying
is unobservable. Another profile sweep would rediscover the same parent.

What it entails: define a conservative whole-program eligibility proof for a
nominal instantiation: no reachable field mutation, no dynamic/host identity
exposure, no unknown mutation-capable call, and safe behavior through native
unions, specialized generic storage, and runtime-service adapters. Census all
66 strict applications and canonical stdlib, reuse the existing alias/mutation
negative guards, and build generated-only ceilings for these three rows. Only
then prototype one shared carrier rule, and only if the same exact proof
admits all three without a type name, container name, or application rule.

Why it matters: a sound read-only proof could let the compiler represent
unobservably immutable Able records with Go values while retaining pointers
for every mutable or uncertain case. That is the credible route from the
measured allocation wall toward Go-native performance without changing Able
semantics.
