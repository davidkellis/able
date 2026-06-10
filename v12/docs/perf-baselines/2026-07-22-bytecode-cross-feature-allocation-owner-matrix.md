# Bytecode cross-feature allocation-owner matrix

Date: 2026-07-22

## Decision

Retain no VM, runtime, compiler, canonical-stdlib, benchmark, fixture,
language, or WASM change from this tranche. A higher-resolution, measured-main
allocation matrix across five unlike feature families exposes no new removable
semantic allocation shared by at least three applications.

Several exact Go allocation sites appear broad until their ownership and
lifetime are reconciled. The strongest new-looking site,
`bytecodeStackSnapshotValue`, is not one mechanism: Binary Event Log and
iterator collect allocate an immutable raw-`i64` snapshot at line 122, while
Matrix Multiply allocates a copied `FloatValue` at line 139. Other broad sites
are already-closed families: large-`i64` boxing, immutable type-expression
construction, Error-interface name slices, and runtime environments with
different lexical/call/concurrency lifetimes.

## Protocol and sampling correction

The workload set is unchanged from the preceding CPU matrix: Binary Event Log,
Matrix Multiply, linked-list iterator collect, Option/Result Config, and
Validated Job Pipeline. The fixed test binary SHA-256 was
`883a6ff6eaa828076a32dd5738f8844e2bec1803778bf2be86fb7e19ec650a00`.
Every process used normal typechecking, canonical external `able-stdlib`, one
warmup, source-root-only loading, `GOMEMLIMIT=1GiB`, `GOGC=50`,
`GOMAXPROCS=1`, and a 59-second cap.

An initial ordinary-rate profile pass was valid, but Go memory sampling starts
before the benchmark can suspend loader/warmup sampling. Consequently,
package initialization—especially `initBytecodeSmallIntBoxCache`—dominated the
short applications. All attribution below therefore focuses stacks through
`BenchmarkBytecodeProgramRuntime`, excluding process initialization, and uses
a second 64 KiB sampling pass to improve resolution. All five higher-resolution
processes passed. Their profile SHA-256 values are:

| Application | Allocation profile SHA-256 |
| --- | --- |
| Binary Event Log | `44050e1c1fedfd7e6b64998c77f6d484cdbc331e2a6f279ebf4363ac50c0b767` |
| Matrix Multiply | `a56f0b4acdf9bc24919ed4b577799e7d668a4ee659b74832aa864d2038d22d3c` |
| Iterator collect | `93a0bfa93bcce593941276359ef3be44c89b77867f24664be226448c78347c7f` |
| Option/Result Config | `e294760b9a55cc26de8d42c64d6ff60dddeebc898f29a1c1561fef16b04dfc05` |
| Validated Job Pipeline | `760eb71da36efc21c586ae23135192f54cad8929fcd1585befe597aade25912f` |

One first high-resolution launch accidentally omitted the explicit
`ABLE_BENCH_SKIP_TYPECHECK=0` override. The benchmark's historical empty-value
default skips typechecking, and Binary failed during warmup with an unresolved
`is_ok` method. It never entered measured `main`, contributes no profile or
counter to this gate, and was replaced by the complete five-process pass above.
This is recorded as a harness configuration error, not a language failure.

The percentages below are normalized against each focused benchmark subtree,
not the whole-process denominator printed by `pprof`.

## Exact counter stability

Four independent one-call processes are retained per application: the prior
CPU calibration, the ordinary-rate allocation profile, an unprofiled exact
run, and the 64 KiB allocation profile. Allocation counts are deterministic to
within 54-62 objects and 31-37 KiB across these runs. Arithmetic means retain
every workstation sample; profile wall times are intentionally not promoted as
performance comparisons.

| Application | Mean B/op | Byte span | Mean objects/op | Object span |
| --- | ---: | ---: | ---: | ---: |
| Binary Event Log | 465,574,830 | 37,368 | 8,361,450.50 | 62 |
| Matrix Multiply | 308,613,900 | 31,696 | 14,032,580.75 | 59 |
| Iterator collect | 8,709,922 | 32,264 | 193,021.25 | 61 |
| Option/Result Config | 67,922,822 | 31,320 | 1,167,830.50 | 54 |
| Validated Job Pipeline | 13,793,654 | 37,304 | 201,420.25 | 61 |

## Focused allocation-object intersection

Shares are sampled flat allocation objects divided by the sampled focused
benchmark subtree. A row is shown when it clears 1% in at least three unlike
applications or is required to explain a misleading broad parent.

| Exact allocation site | Binary | Matrix | Iterator | Option | Validated | Reconciliation |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| `bytecodeBoxedIntegerValue` | 10.35% | 0% | 8.06% | 1.77% | 31.28% | Same line, but required large-`i64` materialization; cache/raw mechanisms already rejected |
| `NewSimpleTypeExpression` | 9.15% | 0% | 0% | 10.24% | 9.39% | Immutable type-expression family already failed its unlike iterator guard |
| `NewIdentifier` | 8.80% | 0% | 0% | 9.58% | 18.77% | Same type-expression family |
| `errorInterfaceNames` | 4.39% | 0% | 0% | 13.08% | 18.77% | Exact prior candidate removed objects but regressed Binary's expanded mean |
| `newEnvironmentBase` | 1.60% | 0% | 0% | 4.65% | 1.56% | Scope, slotless-call, and binding-set lifetimes differ |
| `bytecodeStackSnapshotValue` | 2.85% | 25.45% | 38.00% | 0.38% | Exact symbol, but two raw-`i64` owners and one float-copy owner |

The corresponding allocation-space profiles agree on the first five broad
families. Boxed integers account for 7.55%, 6.92%, 1.37%, and 27.17% of focused
space in Binary, Iterator, Option, and Validated. Environment objects account
for 4.67%, 14.38%, and 5.44% in Binary, Option, and Validated. Identifier and
simple-type nodes each clear 8% in those same three applications. Stack
snapshot space is instead material only in Matrix and Iterator because
Binary's raw-`i64` carrier objects are small.

## Owner reconciliation

### Stack snapshots

Source-line attribution splits the apparent three-application exact symbol:

- Binary and Iterator allocate at `bytecode_vm_float_slot_update.go:122`,
  converting a mutable `bytecodeRawI64SlotCell` into an immutable
  `bytecodeRawI64ResultValue` for stack observation.
- Matrix allocates at line 139, copying a mutable `*runtime.FloatValue` into a
  stable value before pushing it on the operand stack.

The shared helper therefore contains two different alias-severing operations.
The raw-`i64` child has only two unlike owners; the float child has one in this
cohort. Reusing either mutable pointer would let a later slot update change an
already-pushed operand. The prior true operand-lane, pointer-cell, owned-cell,
and raw-carrier experiments either moved the box or regressed broad wall time.

### Large integer materialization

All four broad `bytecodeBoxedIntegerValue` samples allocate at line 249: the
intentional direct `i64` path. Direct parents still differ among environment
persistence, public return materialization, type coercion/matching, and struct
field filling. These are observable boundaries where VM-private mutable/raw
carriers cannot escape.

The alternative mechanisms are already closed. Dynamic map reuse sees
high-cardinality streams and its synchronization simplification regressed the
pooled Rational guard; extending raw cells across call/environment boundaries
also failed broad iterator/collection guards. The current matrix supplies no
new ownership proof that would make those carriers safely reusable.

### Type expressions and Error names

Binary, Option, and Validated rebuild identifier/simple-type nodes during
generic-union, Error, and static-receiver checks. Copy-on-write reuse of
unchanged immutable type-expression graphs previously reduced owner
allocations but regressed the unrelated iterator guard 2.48%. A compiler-wide
static-expression variant likewise produced mixed broad wall results.

`errorInterfaceNames` is even more exact: the same one- or two-name slice is
built in Binary, Option, and Validated. That exact immutable-array candidate
already reduced allocations in every owner, then failed because Binary Event
Log's complete eight-process arithmetic mean regressed 2.66%. No different
mechanism is exposed here.

### Environments and nominal values

The environment leaf does not share a lifetime:

- Binary is 99.7% non-transient `enterRuntimeScope` children;
- Option divides almost evenly between runtime scopes and slotless inline call
  environments;
- Validated's focused sample comes from `NewEnvironmentWithBindingSets`.

Pooling or eliding all three would conflate lexical scope, argument binding,
and concurrency-visible environment state. Existing scope analysis and
single-binding transient reuse already retain the safe generic cases.

Positional struct construction is material only in Binary and Iterator after
setup exclusion. Those instances are user-observable nominal values; the
prior lifetime census found return, pattern, call, mutation, and collection
escapes and closed generic pooling/scalar replacement on the current corpus.

## Verification

No temporary production instrumentation or candidate code was added. The
unchanged bytecode family passes:

```text
GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1 \
  go test ./pkg/interpreter -run TestBytecode -count=1 -timeout 60s
ok able/interpreter-go/pkg/interpreter 29.205s
```

Raw profiles and top reports are cleanup-eligible under
`/tmp/able-bytecode-cross-feature-alloc-20260722` and
`/tmp/able-bytecode-cross-feature-alloc64k-b-20260722`.

## Next recommendation

Return to the compiled frontier and refresh bounded main-only CPU plus exact
allocation profiles for Binary Event Log, Option/Result Config, Manifest
Normalization, and Policy Record Dispatch after the retained lazy native
environment/direct ABI.

Why: the bytecode CPU and allocation intersections now both terminate in
closed mechanisms, required semantic objects, or owners that split below a
shared helper. The compiled ABI removed the former bound-method, receiver-slice,
and native-context allocations, so its older profiles no longer identify the
next current compiler wall. The compiler also remains much farther from the
95%-of-Go product target on these applications.

What it entails: build fixed current generated binaries; retain repeated
verifier-backed main timings for stability; collect main-only CPU and exact
allocation profiles separately; and intersect concrete generated/runtime
leaves across at least three applications. Admit only a primitive rule or a
shared nominal translation/runtime mechanism, never a named container, union,
stdlib API, or benchmark branch. If the four union-heavy owners expose no new
leaf, broaden immediately to unlike non-union compiler misses instead of
retrying context, type-expression, or allocation-pool variants. Continue to
defer WASM.
