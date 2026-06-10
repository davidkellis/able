# Concurrent Signal Dispatch application gate — 2026-07-23

## Decision

Retain the portable `concurrent_signal_dispatch` application, signed numeric
input, source-equivalent Go/Python/Ruby implementations, exact verifier,
catalog and coverage memberships, two complete measurement cohorts, and
bounded profiles. Retain no compiler, generated-runtime, bytecode VM,
tree-walker, canonical-stdlib, language, dependency, or WASM change.

The workload fills the concurrency × arrays/files × user-defined interface
dispatch frontier with application-shaped work. Compiled execution reproduces
the closed goroutine-identity owner. Bytecode execution reproduces closed
Array, raw-integer, call, and type/member families. Neither profile exposes a
new exact generic leaf that passes the three-unlike-application rule.

## Application contract

The program reads 32 signed samples from `signals.txt`, creates 2,048 tasks,
and sends them through four long-lived workers. Each task selects either a
`WeightedKernel` or `DeltaKernel` through the ordinary user-defined
`SignalKernel` interface, processes shared read-only `Array i64` data, and
emits one nominal result. The collector checks worker/Future totals and
computes schedule-independent kernel counts, worker buckets, values, energy,
and checksum:

```text
2048:2048:2048:1024:1024:499,506,529,514:1009575853:1049137146:557952
```

Tree-walker, bytecode, compiled Able, Go 1.26, Python 3.14, and Ruby 4.0 emit
that exact output. Its SHA-256 is
`cb24d4b4ebe05455c69d9232b5bee08e27e1f864776c80996a0253ece9d69d94`.
The catalog passes the real input path, uses the goroutine executor, assigns
four logical CPUs to compiled/Go and one to interpreter lanes, and isolates
the explicit source root. Canonical and external Able sources are identical.
The existing public stdlib Array, Channel, Future, file, and argument APIs were
sufficient.

## Coverage result

The application genuinely covers lexical bindings and patterns; nominal
types, generics, and unions; arrays, text, and files; control flow; inherent
methods; user-defined interface dispatch; Option handling; concurrency;
packages/imports; stdlib protocols; and real program entry. It intentionally
does not claim closure/callable coverage.

The catalog and promoted scorecard contain 52 portable applications, 104
status rows, and 97 selected rows. The 165-triple interaction frontier remains
at minimum depth three with no depth-zero or depth-one gaps and 163
improvements over the reconstructed baseline. The priority concurrency ×
arrays/files × interface-dispatch triple increases from three to four
independent applications. The promoted selected frontier contains 52 compiled
and 45 bytecode rows, eight snapshot meets, 89 misses, five established
guards, zero actionable local groups, and `164.126842` seconds of aggregate
target excess.

## Repeated measurements

Every lane received two independent five-process cohorts, and every sample was
retained. All 50 timed processes passed the exact verifier with zero failures
and zero timeouts.

| Lane | Processes | Pooled mean | Cohort A | Cohort B | Limiting ratio |
| --- | ---: | ---: | ---: | ---: | ---: |
| Able compiled | 10 | 0.283000 s | 0.2700 s | 0.2960 s | 53.905× Go |
| Go 1.26 | 10 | 0.005250 s | 0.0052 s | 0.0053 s | — |
| Able bytecode | 10 | 1.670000 s | 1.6140 s | 1.7260 s | 26.403× Python / 19.385× Ruby |
| Python 3.14 | 10 | 0.063250 s | 0.0612 s | 0.0653 s | — |
| Ruby 4.0 | 10 | 0.086150 s | 0.0922 s | 0.0801 s | — |

Cohort means moved by 9.19% for compiled, 1.90% for Go, 6.71% for
bytecode, 6.48% for Python, and 14.05% for Ruby. All lanes stay below the
15% volatility guard, and the pooled arithmetic means preserve normal
workstation noise instead of selecting a favorable cohort.

## Ownership and admission

Three clean compiled main-only profiles merge to 1.12 seconds of CPU samples.
`bridge.currentGID` and its `runtime.Stack` descendant own 93.75%
cumulatively. Channel receive reaches 28.57%, channel send 19.64%, and the
selected `SignalKernel.evaluate` entry 16.07%, all beneath the current-context
wall. This independently reproduces the exact generic owner already seen
across unlike concurrency applications. Its fixed-context replacement failed
broad concurrent and serial guards, so it is not retried.

Three warmed bytecode-main profiles average 1,313,981,626 ns/op,
108,910,669 B/op, and 2,456,332 allocs/op, merging to 3.92 seconds of CPU
samples. `execArrayReadSlot` is 39.80% cumulative but only 0.51% flat;
`bytecodeRawIntegerValueInfo` is 3.57% flat; `invokeFunction` is 54.34%
cumulative; binary execution is 17.86%; and member/type-generic work is
scattered. These are existing semantic families whose concrete descendants
either split by workload or already failed broad guards. An interface-name,
signal-kernel, task, or named-container special case would violate the
generality rules.

## Evidence

- two Go, Python/Ruby, and Able cohorts:
  `2026-07-23-concurrent-signal-dispatch-{go-reference,interpreter-reference,comparison}-{a,b}.{json,md}`;
- clean compiled merged profile:
  `.profiles/20260723_concurrent_signal_dispatch_compiled_clean_merged.cpu.pprof`;
- warmed bytecode merged profile:
  `.profiles/20260723_concurrent_signal_dispatch_bytecode_runtime_merged.cpu.pprof`;
- readable profile tables:
  `2026-07-23-concurrent-signal-dispatch-{compiled-clean,bytecode}-profile-top.txt`.

## Verification

- exact output parity in tree-walker, bytecode, compiled, Go, Python, and Ruby;
- ten verifier-backed timed processes per compiled, bytecode, and reference
  lane;
- three clean compiled and three warmed bytecode profiles;
- focused catalog, selection, coverage, operation-depth, matrix, triple, and
  scorecard checks;
- every added source file remains below 1,000 lines;
- JSON, whitespace, and diff checks.

## Next recommendation

Complete `portable-concurrent-callable-data-application-frontier`.

Why: this application raises concurrency × arrays/files × interface dispatch
from three to four programs without exposing a new open local mechanism. The
remaining highest-ranked minimum-depth interaction is concurrency ×
arrays/files × functions/closures, still represented by only three portable
applications and adjacent to `34.721474` seconds of current target excess.
Another interface or VM micro-change would repeat closed evidence; an unlike
callable-driven application can expose a genuinely shared semantic owner or
strengthen that guard.

What it entails: add one deterministic file-driven application whose workers
transform numeric or structured Array data through first-class functions or
closures, with source-equivalent Able/Go/Python/Ruby lanes and an exact
schedule-independent verifier. Take two five-process cohorts per lane, profile
only an exact owner already present in two unlike applications, and admit a
change only after broad application guards. Update canonical `able-stdlib`
only for a reusable API or correctness defect, and do not begin WASM work.
