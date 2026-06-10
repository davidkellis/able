# Compiled backing-slice growth gate

Date: 2026-07-16

## Decision

Keep no compiler, generated-runtime, bytecode VM, canonical-stdlib,
application, verifier, reference, or benchmark-source change. Fresh
post-bootstrap profiles for Reverse Complement, Lexical Rollup, Array Slice
Window, and Base64 do not repeat one backing-growth operation. The earlier
Reverse Complement and Lexical Rollup `runtime.growslice` resemblance splits
into different generated callers on the current source, while Array Slice
Window and Base64 use exact result allocations rather than growth.

No capacity, append, conversion, or Array-lowering candidate was built. A
candidate based on this evidence would optimize one source shape or named use
rather than a compiler mechanism shared by unlike applications.

## Method

The current compiler built each application once with canonical
`../able-stdlib`, `ABLE_SOURCE_ROOT_ONLY=1`, and the default monomorphic Array
lowering. The resulting binaries ran from their catalog working directories
with `GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`.

CPU profiles used the generated launcher's main-phase hook so package loading
and registration were excluded. Reverse Complement, Lexical Rollup, and Array
Slice Window each contributed 30 independent process profiles. Base64
contributed three because its main phase is long enough to provide 6.36 CPU
seconds. Exact allocation statistics came from a separate one-process phase
snapshot, without CPU sampling. These are attribution runs rather than a new
promoted scorecard.

All four binaries completed normally. Their stdout SHA-256 values were:

| Application | SHA-256 |
| --- | --- |
| Reverse Complement | `db06a593d68950aa91293fb57ffed87ad57040c4ac7ab81ae9dc43d50b400bb7` |
| Lexical Rollup | `a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604` |
| Array Slice Window | `155f89122475c7b282637dbf2ecba6d19771d396e801b581cb1d1b0cef64103e` |
| Base64 | `5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316` |

## Main-phase attribution

| Application | CPU samples | Main allocated bytes / objects | Slice allocation route |
| --- | ---: | ---: | --- |
| Reverse Complement | 690 ms | 22,357,944 / 368 | `runtime.growslice` is 390 ms cumulative. Generated `slice_bytes_from_offset` contributes 310 ms while copying the unwritten suffix for `write_all`; `reverse_complement_fasta` contributes 60 ms while growing user output/sequence buffers. |
| Lexical Rollup | 240 ms | 4,870,896 / 31,323 | `runtime.growslice` is only 10 ms (4.17%), below generic `Iterator.collect` through the ordinary `Extend<Array<i64>>` append. The retained bounded-line API removed the old eager full-file growth wall. |
| Array Slice Window | 60 ms | 3,908,824 / 24,334 | No `growslice` sample. `Array.slice` creates an independent result with exact capacity; `runtime.makeslice` has one 10 ms sample. The allocation profile attributes 24,002 objects and about 1.37 MiB to these required result copies. |
| Base64 | 6.36 s | 2,204,171,864 / 1,361 | No `growslice` sample. Host base64 encode/decode and MD5 work dominate CPU. Encode/decode allocate exact large result slices through `runtime.makeslice`, accounting for almost all bytes but very few objects. |

Generated-source inspection agrees with the profiles:

- Reverse Complement's offset helper appends one byte at a time into a fresh
  `Array<u8>`, while its transformation preallocates the main output and
  sequence buffers according to application values.
- Lexical Rollup reaches the shared Array `Extend` implementation from
  `Iterator.collect`; it is no longer a material growth wall.
- Array Slice Window preallocates `end - start` capacity and appends exactly
  that many copied `i32` values. The fresh backing is required because Able
  `Array.slice` returns an independently mutable Array.
- Base64's exact host codec outputs are required results, not geometric
  backing growth.

The only common names below these programs are Go allocation and GC parents.
They do not encode a shared compiler decision. Preallocating Reverse
Complement's offset copy, changing `Array.slice` into a view, specializing
`Iterator.collect`, or altering host codec allocation would each address a
different semantic operation. None passes the cross-application admission
gate.

## Bootstrap observation

Although bootstrap was deliberately excluded from candidate selection, the
exact phase snapshots recorded a repeated cold-start allocation band:

| Application | Bootstrap allocated bytes | Bootstrap allocations |
| --- | ---: | ---: |
| Reverse Complement | 3,074,656 | 7,570 |
| Lexical Rollup | 3,089,208 | 8,023 |
| Array Slice Window | 2,754,944 | 3,942 |
| Base64 | 2,688,872 | 3,098 |

This table does not authorize a startup change. It establishes a current,
cross-family selection lead that can be attributed separately from the
already-rejected main-phase growth hypothesis.

### Follow-up correction

The subsequent bootstrap tranche added an allocation-only phase mode and
proved that roughly 2.4 MiB of every value above came from starting the CPU
profiler inside the combined diagnostic mode. Corrected bootstrap allocations
are 627,064 bytes for Document Audit, 396,392 for Dependency Plan, 283,880 for
Array Slice Window, and 225,008 for Base64. The corrected attribution and
rejected shared-type metadata trial are recorded in
`2026-07-16-compiled-bootstrap-allocation-gate.md`; do not use the combined
values above as production startup costs.

## Verification and cleanup

The four current binaries built successfully and completed their canonical
program invocations under the memory/GC guard. Generated trees, binaries,
per-process profiles, and allocation snapshots were temporary evidence and
are removed after this record. No source test suite was required because no
runtime or compiler code changed.

## Next recommendation

Reconcile exact compiled bootstrap allocation stacks across three short,
unlike applications—Document Audit, Dependency Plan, and Array Slice
Window—with Base64 as a long-running guard. Require the same concrete
registration, environment, metadata, or runtime-initialization descendant in
all three before designing a candidate.

Why: this tranche rules out backing growth as a shared generated-main wall,
while its independent phase counters expose roughly 2.7–3.1 MB of cold-start
allocation in every compiled binary. Startup is a broad product cost and
dominates many real applications whose Go equivalents finish in a few
milliseconds. Removing a proven shared initialization allocation would help
ordinary binaries rather than one benchmark kernel.

What it entails: build each binary once, capture allocation snapshots before
and after registration without CPU-profiler startup, diff and classify the
largest concrete allocation stacks, and reconcile them against the existing
static-metadata-retention and generic-interface-registration records. Do not
retry decoded default-AST removal, alternate metadata codecs, or literal
interface-name matching. Advance only a newly repeated descendant, then add
static/dynamic/bootstrap semantic controls and use alternating repeated
preserved-binary measurements on the workstation.
