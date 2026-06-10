# Typed generic-nominal storage reach census

Date: 2026-07-27

## Decision

Retain the deterministic census tool and evidence only. Retain no compiler,
generated runtime, runtime, interpreter, VM, stdlib, language, dependency, or
WASM production change.

Runtime-backed typed storage is broad: 25 of the 62 strict applications execute
at least one specialized HashMap or Channel operation. It is not one shared
nominal-storage owner, however. Fourteen applications reach HashMap storage,
14 reach Channel payload storage, and only three overlap. Direct generic
nominal fields are already native, and the application corpus contains no
user-defined generic storage declaration.

The candidate therefore fails the language-general representation gate. A
HashMap-only change would be a forbidden named-container rule. A Channel
payload change would reopen the already-separated scheduler ABI. Combining
their unrelated semantic contracts under one implementation project would be
a broad runtime-service redesign rather than one evidenced lowering rule.

Machine-readable results are in
`2026-07-27-typed-generic-nominal-storage-reach-census.json`; the 62-row source
and generated-code manifest is
`2026-07-27-typed-generic-nominal-storage-census.tsv`.

## Method

The current portable coverage catalog was emitted with one retained compiler,
the catalog-selected source root, `ABLE_SOURCE_ROOT_ONLY=1`, and
`--no-fallbacks`. The workspace lived under disk-backed `/var/tmp`.

All 62 applications emitted successfully. Independent `go list -deps` checks
found 62 valid final graphs and zero
`able/interpreter-go/pkg/interpreter` dependencies. The census records exact
source and generated `compiled.go` hashes, carrier declarations, specialized
method reach, application structs, and generated Map interfaces. A second run
was byte-identical.

The report classifies a storage site by:

- nominal origin: direct generic nominal, HashMap, or Channel;
- carrier: native primitive/Array/application nominal or `runtime.Value`;
- operation: direct field access, map operation, or channel send/receive;
- identity and aliasing obligation;
- generated interface exposure; and
- whether materialization comes from fallback or an explicit runtime service.

## Corpus result

| Property | Result |
| --- | ---: |
| Strict generated applications | 62 |
| Interpreter-free final graphs | 62 |
| Native Array variants, per-application sum | 1,164 |
| Distinct native Array variants | 55 |
| Direct generic variants, per-application sum | 310 |
| Distinct direct generic variants | 36 |
| `runtime.Value` fields in those direct variants | 0 |
| Application struct declarations | 95 |
| Application generic struct declarations | 0 |
| Applications with executable HashMap operations | 14 |
| Applications with executable Channel operations | 14 |
| Applications with either boundary | 25 |
| Applications with both boundaries | 3 |

The 36 distinct direct specializations span `ArrayIterator`,
`ChannelIterator`, `Deque`, `DivMod`, `DivModResult`, `Indexed`, `Queue`, and
`Span`. Their fields use concrete primitives, native Array pointers, or native
nominal pointers. This is the expected general nominal lowering behavior.

### HashMap origin

The 14 applications are Backup Dedup, Binary Event Log, Configuration
Validation Extraction, Concurrent Event Routing, Concurrent Text Index,
Dependency Wave Validation, Inventory Reconciliation, K-Nucleotide, Log
Routing Redaction, Policy Record Dispatch, the three regex audits, and Word
Frequency. They emit 15 concrete carrier variants and 125 specialized methods.
Eleven expose a generated Map interface; Backup, Binary Event Log, and
K-Nucleotide use concrete methods.

The generated signatures accept native keys and values, then encode them for
the `runtime.Value` HashMap helpers. That store owns Able hashing and equality,
mutation, aliasing, cloning, iteration, insertion identity, and nullable
recovery. Dependency Wave supplies the strongest application-nominal payload
control, `HashMap i64 WaveResult`, but the storage nominal is still the stdlib
HashMap.

### Channel origin

Fourteen applications execute specialized Channel operations. Sixteen contain
a Channel carrier, but Mutex Ledger and Mutex Work Queue emit no corresponding
specialized methods and are not counted as executable reach. Ten of the 14
carry application-defined nominal payloads.

Generated send and receive signatures accept or return native primitives and
nominal pointers, then encode the handle and payload for the runtime scheduler.
The scheduler owns ordering, blocking, wakeup, cancellation, closure,
concurrent payload lifetime, and task bookkeeping. These obligations are
different from map hashing and identity. The retained nonblocking Channel fast
path already localizes scheduler-payload recovery; removing stored payload
conversion would require the broader typed scheduler/channel ABI that prior
evidence deliberately excluded.

The three overlapping applications are Concurrent Event Routing, Concurrent
Text Index, and Dependency Wave Validation. Sharing a `runtime.Value`
conversion leaf does not make the two runtime services one semantic owner.

## Application-nominal allocation control

Dependency Wave Validation exercises both service families with native
application nominals:

```text
WaveTask   -> Channel WaveTask
WaveResult -> Channel WaveResult
WaveResult -> HashMap i64 WaveResult
```

The generated `WaveTask` fields are three `int64` values. `WaveResult` retains
an `int64` identifier and its concrete union carrier. The transition to
`runtime.Value` occurs only inside the generated specialized Channel and
HashMap methods.

Three verifier-backed exact main-phase allocation profiles used
`ABLE_EXECUTOR=goroutine` and `GOMAXPROCS=4`:

| Run | Main bytes | Main allocations | GC | HashMap operation routes | Channel operation routes |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 6,515,176 | 122,125 | 3 | 36,610 | 46,864 |
| 2 | 6,515,320 | 122,125 | 3 | 36,073 | 47,399 |
| 3 | 6,515,320 | 122,117 | 3 | 36,229 | 47,236 |
| Mean | 6,515,272 | 122,122.33 | 3 | 36,304 | 47,166.33 |

The top-level operation routes are disjoint. Relative to exact phase
allocation counts, HashMap accounts for 29.73%, Channel for 38.62%, and the
combined routes for 68.35%. This proves material cost in the user-nominal
control, but it also proves that the cost has two semantic parents.

## Admission analysis

| Required evidence | Result |
| --- | --- |
| Material in at least three unlike applications | Yes, separately for HashMap and Channel |
| Application-defined nominal payload | Yes, Dependency Wave for both; ten Channel apps overall |
| Multiple storage nominal origins | No; all runtime-backed origins are HashMap or Channel |
| Application-defined generic storage nominal | No; the corpus has zero declarations |
| One shared identity/aliasing contract | No |
| One general compiler/runtime correction | No |

Fallback materialization is not the cause: every surveyed program is strict
and interpreter-free. Ordinary generic nominal storage is not the cause
either: the direct generic carrier census and compiler guards show it already
stays native. The remaining transitions are explicit runtime-service ABIs.

No production candidate advances, and manufacturing an A/B comparison would
not satisfy the admission policy.

## Verification

- The census tool passes `bash -n` and produces byte-identical reruns.
- All 62 emissions succeeded; all 62 final graphs omit the interpreter.
- The Dependency Wave verifier passed for every allocation run with one stable
  output hash.
- Native generic method, bound generic field, generic interface touchpoint,
  caller-owned nominal alias/escape, generic Map signature, and retained
  Channel localization guards pass in 3.113 seconds.
- The worktree's existing changes and 34 deferred WASM paths were preserved.
- The 4.1 GiB disposable census workspace was removed after evidence
  validation.

## Next

Expand the portable suite only with a genuine unlike application/domain whose
natural design uses application-defined generic storage, then measure it before
reopening representation work.

This is next because the corpus now proves that runtime-service boxing is
expensive, but it has no application-defined generic storage nominal. That
missing control is the decisive distinction between a general nominal lowering
defect and two service-specific HashMap/Channel ABIs.

The work entails selecting a real feature-covering workload without
benchmark-only APIs or workload shaping, then implementing source-equivalent
Able/Go/Python/Ruby lanes whose hot state naturally lives in an
application-defined generic nominal backed by static Arrays. Verify strict
interpreter-free generation, native fields, identity and aliases, then collect
repeated public-verifier timings and exact allocation profiles. Do not encode
a named container or benchmark in compiler logic.

This is important because a native result closes the general nominal route and
keeps future work focused on explicit kernel services. A material
`runtime.Value` transition would instead provide the first admissible evidence
for a shared representation correction. Do not begin WASM work.
