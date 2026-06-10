# Concurrent Transform Chain application gate — 2026-07-23

## Decision

Retain the portable `concurrent_transform_chain` application, its signed
numeric input, source-equivalent Go/Python/Ruby implementations, exact
verifier, catalog and coverage memberships, two complete measurement cohorts,
and bounded profiles. Retain no compiler, generated-runtime, bytecode VM,
tree-walker, canonical-stdlib, language, dependency, or WASM change.

The workload fills the concurrency × arrays/files × functions/closures
frontier with application-shaped work. Compiled execution reproduces the
closed goroutine-identity owner. Bytecode proves that the dynamic callable
array already reaches the VM's inline-frame path, then reproduces closed Array,
integer, environment/cache, and dispatch families. Neither mode exposes a new
exact generic leaf that passes the broad admission rule.

## Application contract

The program reads 32 signed samples from `samples.txt`, creates 2,048 window
tasks, and sends them through four long-lived workers. Each worker constructs
an `Array` of three captured five-argument functions. Every task selects a
different starting position in that function array and applies all three
transforms repeatedly while scanning shared read-only numeric data. The
collector checks worker/Future totals and computes schedule-independent
buckets, values, energy, and checksum:

```text
2048:2048:2048:539,524,496,489:1008719842:1059875396:760310
```

Tree-walker, bytecode, compiled Able, Go 1.26, Python 3.14, and Ruby 4.0 emit
that exact output. Its SHA-256 is
`4695dbb758688b107a6c7917406311bad54f32f2f42310b2e6c2f58702468f57`.
The catalog passes the real input path, uses the goroutine executor, assigns
four logical CPUs to compiled/Go and one to interpreter lanes, and isolates
the explicit source root. Canonical and external Able sources are identical.
The existing public stdlib Array, Channel, Future, file, and argument APIs were
sufficient.

## Coverage result

The application genuinely covers lexical bindings and patterns; nominal
types and generic arrays; expressions, arrays, text, and files; captured
closures stored as callable data; control flow; inherent methods; nullable
Channel receive handling; concurrency; packages/imports; stdlib protocols;
and real program entry.

The promoted catalog contains 53 portable applications and 106 status rows.
Both modes are selected, so the strict frontier contains 53 compiled and 46
bytecode rows. The priority concurrency × arrays/files ×
functions/closures triple increases from three to four independent
applications.

## Repeated measurements

Every lane received two independent five-process cohorts, and every sample was
retained. All 50 timed processes passed the exact verifier with zero failures
and zero timeouts.

| Lane | Processes | Pooled mean | Cohort A | Cohort B | Limiting ratio |
| --- | ---: | ---: | ---: | ---: | ---: |
| Able compiled | 10 | 8.521000 s | 8.1760 s | 8.8660 s | 1,374.355× Go |
| Go 1.26 | 10 | 0.006200 s | 0.0065 s | 0.0059 s | — |
| Able bytecode | 10 | 2.844000 s | 2.8300 s | 2.8580 s | 19.533× Python / 15.533× Ruby |
| Python 3.14 | 10 | 0.145600 s | 0.1628 s | 0.1284 s | — |
| Ruby 4.0 | 10 | 0.183100 s | 0.2095 s | 0.1567 s | — |

Cohort means moved by 8.10% for compiled, 9.68% for Go, 0.98% for bytecode,
23.63% for Python, and 28.84% for Ruby. The reference lanes were volatile, so
the pooled means preserve all ten workstation samples rather than selecting a
favorable cohort.

## Ownership and admission

Three compiled main-only profiles merge to 55.24 seconds of CPU samples.
`bridge.currentGID` owns 98.21% cumulatively and `runtime.Stack` owns 98.01%;
the three generated captured-transform bodies each sit beneath that same wall
at 31–34% cumulative. This independently reproduces the exact generic owner
already seen across unlike concurrency applications. Its fixed-context
replacement failed broad concurrent and serial guards, so it is not retried.

Three warmed bytecode-main profiles average 2,395,497,747 ns/op,
120,647,216 B/op, and 3,363,444 allocs/op, merging to 7.14 seconds of CPU
samples. `execCallOpcode` is 25.21% cumulative, `execArrayReadSlot` 19.89%,
`execBinary` 16.81%, and `invokeFunction` 14.99%. The instrumented full
application records 442,368 dynamic `Call` opcodes, 454,754 inline-call hits,
zero inline-call misses, and 741,669 canonical Array-slot fast hits. Thus the
captured callable chain is already inlined; the cumulative invocation parent
also owns the worker task beneath it. Residual exact leaves are the established
Array, raw-integer, environment/cache-lock, frame/return, and arithmetic
families, with no new concrete child shared across three unlike applications.

## Evidence

- two Go, Python/Ruby, and Able cohorts:
  `2026-07-23-concurrent-transform-chain-{go-reference,interpreter-reference,comparison}-{a,b}.{json,md}`;
- clean compiled merged profile:
  `.profiles/20260723_concurrent_transform_chain_compiled_merged.cpu.pprof`;
- warmed bytecode merged profile:
  `.profiles/20260723_concurrent_transform_chain_bytecode_runtime_merged.cpu.pprof`;
- readable profile tables and inline counters:
  `2026-07-23-concurrent-transform-chain-{compiled,bytecode}-profile-top.txt`
  and `2026-07-23-concurrent-transform-chain-bytecode-runtime-stats.json`.

## Verification

- exact output parity in tree-walker, bytecode, compiled, Go, Python, and Ruby;
- ten verifier-backed timed processes per compiled, bytecode, and reference
  lane;
- three clean compiled and three warmed bytecode profiles;
- focused catalog, selection, coverage, operation-depth, matrix, triple, and
  scorecard checks;
- every added source file remains below 1,000 lines;
- JSON, source-identity, whitespace, and diff checks.

## Next recommendation

Complete `portable-concurrent-callable-interface-application-frontier`.

Why: after this promotion, concurrency × functions/closures × interface
dispatch is the only remaining minimum-depth interaction. It has three
portable applications, semantic weight nine, and `43.580000` seconds of
adjacent target excess. A workload that combines interface-selected behavior
with first-class callbacks inside workers can expose a shared call/dispatch
boundary that neither the interface-only nor callable-data applications could
isolate.

What it entails: build one deterministic, file-driven Able/Go/Python/Ruby
application whose workers process nominal records through user-defined
interface implementations and then invoke first-class strategy functions or
captured callbacks. Add an exact schedule-independent verifier, take two
five-process cohorts per lane, and profile only after parity is established.
Admit a production change only for an exact generic owner repeated across at
least three unlike programs. Update canonical `able-stdlib` only for a
reusable API or correctness defect, and do not begin WASM work.
