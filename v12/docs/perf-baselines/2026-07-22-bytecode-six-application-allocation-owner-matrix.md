# Bytecode six-application allocation-owner matrix

Date: 2026-07-22

## Decision

Retain no VM, compiler, tree-walker, canonical-stdlib, benchmark, fixture,
language, or WASM change from this tranche. Five exact measured-main allocation
counters per application and six higher-resolution allocation profiles expose
no new removable semantic allocation that is material in at least three
unlike families.

Three apparent intersections survive the mechanical symbol gate. Integer
boxing and primitive Array lease/view ownership are already-closed generic
families. Generic positional struct construction is large in Concurrent Event
Routing, Fixed Width 128, and Word Frequency, but the common Go allocator
serves different nominal definitions with incompatible lifetimes: async
event/route records, checked `UInt128` results, and returned UTF-8 decode
results. There is no shared pooling, frame-ownership, or scalar-transport rule
to implement safely.

## Protocol

The workload set matches the preceding CPU-owner refresh. One fixed current
test binary, SHA-256
`7d6040f01b5396a91fce31125a284f4892a226bf5081e8ff9025ba62c2909ab6`
and Go build ID `3fc961f5f7986c29154bc4349c8bef659d60b78d`, ran every process.
Every run used normal typechecking, canonical external `able-stdlib`,
source-root-only loading, one warmup, one measured `main()` call,
`GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and a 55-second cap.

Allocation attribution used separate `alloc_objects` and `alloc_space` views
of 64 KiB-rate profiles. The runtime harness suspends memory sampling while it
loads, lowers, typechecks, and warms the program, then resumes it for measured
`main`. Package initialization occurs before that suspension, so named init
owners such as `initBytecodeSmallIntBoxCache`, `init.func1`, and testing alarm
setup were explicitly excluded.

Whole-profile measured-main roots were retained instead of focusing only on
`BenchmarkBytecodeProgramRuntime`. That distinction is necessary for Event
Routing: allocations performed by executor goroutines are not descendants of
the benchmark goroutine, even though they are required measured-main work.

All six profile processes and all eighteen new unprofiled counter processes
passed. Each mean below also retains the immediately preceding clean CPU
counter, giving five complete samples per application. Timing variation did
not select the result; the authoritative allocation means retain every sample.

## Exact counter stability

| Application | Samples | Mean B/op | Byte span | Mean objects/op | Object span |
| --- | ---: | ---: | ---: | ---: | ---: |
| Array Slice Window | 5 | 14,170,147.2 | 37,256 | 422,262.40 | 62 |
| Concurrent Event Routing | 5 | 284,531,776.0 | 850,040 | 2,808,239.80 | 153 |
| Distance Field | 5 | 368,061,289.6 | 31,760 | 26,000,136.80 | 59 |
| Fixed Width 128 | 5 | 1,242,248,091.2 | 37,288 | 30,858,357.40 | 62 |
| Reverse Complement | 5 | 213,587,099.2 | 38,864 | 3,542,935.00 | 96 |
| Word Frequency | 5 | 48,398,153.6 | 239,248 | 637,276.40 | 89 |

The wider Event Routing and Word byte spans are still 0.30% and 0.49% of
their means. Object counts—the allocation-owner admission metric—vary by at
most 153 objects out of 2.8 million and 89 out of 637 thousand respectively.

## Cross-application allocation intersection

Percentages below use conservative whole-profile denominators in Array, Event
Routing, Distance, Fixed Width, Reverse Complement, and Word order. An exact
site must clear 1% flat allocation objects in at least three unlike
applications before semantic reconciliation.

| Exact allocation site | Array | Event | Distance | Fixed | Reverse | Word | Reconciliation |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| `bytecodeBoxedIntegerValue` | 1.92% | 1.65% | 0% | 6.89% | 42.07% | 5.58% | Broad symbol, but two different source lines; both boxing mechanisms are closed |
| `NewStructInstancePositionalSized` | 0% | 6.58% | 0% | 8.80% | 0% | 21.62% | Same allocator, different definitions and observable lifetimes |
| `arrayStoreRegisterLeaseCleanupLocked` | 4.61% | 1.80% | 0% | 0% | 0% | 6.19% | Exact primitive-Array ownership allocation; already closed across nine applications |
| `ArrayStoreValueViewFromHandle` | 2.11% | 1.05% | 0% | 0% | 0% | 4.75% | Same closed Array shell/view/lease family |

Allocation space agrees. Positional structs account for 7.63%, 22.95%, and
27.73% of Event, Fixed, and Word space. Integer boxing accounts for 6.74%,
31.87%, and 2.68% of Fixed, Reverse, and Word space, plus 1.08% in Array.
Array tracking/shell ownership reaches 2.91%, 1.45%, and 4.37% in Array,
Event, and Word.

## Owner reconciliation

### Integer boxing

Exact line attribution divides the five-program symbol:

- Array Slice Window, Event Routing, and Word Frequency allocate at
  `bytecode_vm_small_int_boxing.go:249`, the direct large-`i64` result path.
- Fixed Width and Reverse Complement allocate at line 295, after the bounded
  dynamic cache has no reusable entry or capacity.

The three-application line-249 boundary was already material in the preceding
five-family allocation matrix. Those values cross observable stack,
environment, return, type, and field boundaries. Raw carriers, pointer cells,
dynamic map reuse, cache locking, cache limits, and producer-side alternatives
have all completed repeated broad gates. The current profiles expose no new
ownership proof or redundant allocation within that line.

### Positional nominal structs

All three profiles allocate the `StructInstanceValue` at
`runtime/struct_instance.go:6`, but their consumers prove that this is a shared
constructor rather than a shared removable allocation:

- Event Routing samples 179,680 constructor objects; 175,068 (97.4%) execute
  under spawned serial-executor tasks. The program constructs `EventRecord`,
  `EventTask`, `AcceptedRoute`, `RejectedRoute`, and `RoutedEvent` values that
  cross Result, channel/task, future, union, call, and collection boundaries.
- Fixed Width samples 2,256,024 constructor objects. Its caller tree directly
  reaches `execUInt128AddMemberFast`, `execUInt128SubMemberFast`,
  `bytecodeUInt128InstanceFromValue`, and ordinary named struct literals. These
  are checked `UInt128` values returned from member calls and carried through
  loop state.
- Word Frequency samples 142,987 constructor objects. The prior identity and
  lifetime census established `Utf8DecodeResult` as its dominant definition;
  those values cross an inline return, typed pattern, field reads, and multiple
  frame slots before discard.

No nominal definition repeats across these three applications, and their
lifetimes are async/escaping, loop-carried checked results, and
return-to-deconstruction respectively. Generic pooling would violate mutable
identity and alias behavior. Frame ownership is invalid for the return,
channel, task, union, and loop boundaries. The earlier returned-nominal gate
also failed its required three-program/two-definition common-lifetime test.

### Array ownership

The lease cleanup and handle-view sites are genuinely the same semantic
family in Array, Event, and Word. They are not open, however. The coverage-wide
allocation census already found the lease allocation in nine applications and
closed it after the primitive-Array shell/lease work. Cleanup registration is
what safely releases identity-bearing backing state after aliases die; removing
it leaks tracked state, while sharing it conflates independent owners. Existing
owner-local tracking avoids duplicate registration for the same lease.

## Per-application owners

The dominant profiles confirm the split below the broad Go GC ancestry:

| Application | Dominant measured allocation objects |
| --- | --- |
| Array Slice Window | Raw integer results, tracked i32 cache growth, string building, and Array shells |
| Concurrent Event Routing | Runtime environments and mutexes, async event/route structs, member caches, and strings |
| Distance Field | Raw float normalization/materialization plus unary and Ratio-native result values |
| Fixed Width 128 | `math/big` limbs, integer arithmetic/clones, `UInt128` results, and integer boxes |
| Reverse Complement | Host `u8` value conversion and integer boxing |
| Word Frequency | UTF-8 result structs, String host builtins, Array shells, and integer boxes |

There is no common open semantic allocation beneath these six owner sets. No
candidate was built, so no wall-time A/B claim is inferred from allocation
profiles.

## Verification

No temporary runtime instrumentation or candidate code was added. The
unchanged bytecode suite passes:

```text
GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1 \
  go test ./pkg/interpreter -run TestBytecode -count=1 -timeout 60s
ok able/interpreter-go/pkg/interpreter 27.348s
```

Profiles, full pprof tables, counter samples, and the fixed test binary are
cleanup-eligible under `/tmp/able-six-owner-alloc-20260722`.

## Next recommendation

Move compiler selection up one architectural level: run a current
closed-world typed semantic-boundary coverage census across unlike high-excess
compiled applications.

Why: the bytecode CPU and allocation frontiers now both terminate in closed
mechanisms or semantically required objects. The current compiled scorecard
still has only five of 49 applications at the 95%-of-Go snapshot target, while
its last exact-owner refresh also found no three-family local helper. Continuing
to profile or reorder individual VM/compiler helpers would repeat exhausted
work. The remaining plausible general leverage is to keep more statically
proven execution inside native typed generated code and cross the dynamic
runtime boundary only where Able semantics require it.

What it entails: instrument the current compiler IR/generated code to count
dynamic transitions between typed generated values/calls and
`runtime.Value`, bridge, dynamic dispatch, nominal encoding, and effect/error
machinery. Reconcile those counts with target excess for at least Concurrent
Event Routing, Fixed Width 128, K-Nucleotide, Distance Field, and Policy Record
Dispatch. Admit a prototype only if one concrete shared primitive rule or
general nominal translation rule is material in at least three unlike
applications and its optimistic bound removes a meaningful portion of each
target gap. Then use fixed verifier-backed binaries and repeated alternating
process averages. Do not introduce a named nominal/container/stdlib/application
lowering, and continue to defer WASM.
