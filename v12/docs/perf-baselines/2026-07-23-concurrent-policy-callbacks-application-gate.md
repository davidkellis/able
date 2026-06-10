# Concurrent Policy Callbacks application gate — 2026-07-23

## Decision

Retain the portable `concurrent_policy_callbacks` application, its signed
policy input, source-equivalent Go/Python/Ruby implementations, exact verifier,
catalog and coverage memberships, two complete measurement cohorts, and
bounded profiles. Retain no compiler, generated-runtime, bytecode VM,
tree-walker, canonical-stdlib, language, dependency, or WASM change.

The workload raises the sole minimum-depth concurrency × functions/closures ×
interface-dispatch interaction from three to four applications. Compiled
execution reproduces the closed goroutine-identity owner. Bytecode proves that
both the interface-selected methods and their callback arguments reach inline
frames, then reproduces closed integer, arithmetic, member/cache, frame, and
general call families. Neither mode exposes a new exact generic leaf that
passes the broad admission rule.

## Application contract

The program reads 32 signed policy records from `policies.txt`, creates 2,048
tasks over 64 rounds, and sends them through four long-lived workers. Each
task dynamically selects one of two user-defined `DecisionPolicy`
implementations. The selected interface method invokes one of three captured
callback strategies six times while processing a nominal record. The collector
checks worker and Future totals and computes schedule-independent policy,
callback, bucket, score, checksum, and adjustment totals:

```text
2048:2048:2048:1024:1024:683,682,683:515,497,550,486:985793755:988602070:516821
```

Tree-walker, bytecode, compiled Able, Go 1.26, Python 3.14, and Ruby 4.0 emit
that exact output. Its SHA-256 is
`7f1eeebf4548e851a416d06dddf41316bb9ebd4baa5f0e9e6b5265405883f210`.
The catalog passes the real input path, uses the goroutine executor, assigns
four logical CPUs to compiled/Go and one to interpreter lanes, and isolates
the explicit source root. Canonical and external Able sources are identical.
The existing public stdlib Array, Channel, Future, file, argument, and text
APIs were sufficient.

## Coverage result

The application genuinely covers lexical bindings and patterns; nominal
types and generic arrays; expressions, arrays, text, and files; captured
closures stored as callable data; control flow; inherent methods; user-defined
interfaces and implementations; nullable Channel receive handling;
concurrency; packages/imports; stdlib protocols; and real program entry.

The promoted catalog contains 54 portable applications and 108 status rows.
Both modes are selected, so the strict frontier contains 54 compiled and 47
bytecode rows. The priority concurrency × functions/closures × interface
dispatch triple increases from three to four independent applications.

## Repeated measurements

Every lane received two independent five-process cohorts, and every sample was
retained. All 50 timed processes passed the exact verifier with zero failures
and zero timeouts.

| Lane | Processes | Pooled mean | Cohort A | Cohort B | Limiting ratio |
| --- | ---: | ---: | ---: | ---: | ---: |
| Able compiled | 10 | 0.518000 s | 0.5420 s | 0.4940 s | 106.448× Go |
| Go 1.26 | 10 | 0.004866 s | 0.0049 s | 0.0048 s | — |
| Able bytecode | 10 | 0.400000 s | 0.3820 s | 0.4180 s | 5.926× Python / 7.356× Ruby |
| Python 3.14 | 10 | 0.067501 s | 0.0768 s | 0.0582 s | — |
| Ruby 4.0 | 10 | 0.054378 s | 0.0574 s | 0.0514 s | — |

Cohort means moved by 9.27% for compiled, 2.06% for Go, 9.00% for bytecode,
27.56% for Python, and 11.03% for Ruby. The Python reference lane was
volatile, so the pooled means preserve all ten workstation samples rather than
selecting a favorable cohort. Every selected and pooled ratio remains an
unambiguous target miss.

## Ownership and admission

Three compiled main-only profiles merge to 1.07 seconds of CPU samples.
`bridge.currentGID` owns 94.39% cumulatively. Both generated interface
implementation bodies, their adapters, all three captured callback bodies,
and the worker task sit beneath that same wall. This independently reproduces
the exact generic owner already seen across unlike concurrency applications.
Its fixed-context replacement failed broad concurrent and serial guards, so it
is not retried.

Three clean warmed bytecode-main profiles use ten application calls apiece,
average 137,946,116 ns/op, 7,218,931 B/op, and 185,445 allocs/op, and merge to
4.13 seconds of CPU samples. `execCallOpcode` is 43.34% cumulative,
`execCallMember` 32.69%, `execBinary` 23.73%, and
`bytecodeRawIntegerValueInfo` 5.08%. The separately instrumented full
application records 28,939 inline-call hits and zero misses, plus 11,305
resolved member inline hits. Interface coercion is only 2.42% cumulative and
does not form an exact three-unlike-application owner. Residual concrete
leaves repeat the established integer, arithmetic, environment/cache-lock,
frame/return, and member-dispatch families.

## Evidence

- two Go, Python/Ruby, and Able cohorts:
  `2026-07-23-concurrent-policy-callbacks-{go-reference,interpreter-reference,comparison}-{a,b}.{json,md}`;
- clean compiled merged profile:
  `.profiles/20260723_concurrent_policy_callbacks_compiled_merged.cpu.pprof`;
- clean warmed bytecode merged profile:
  `.profiles/20260723_concurrent_policy_callbacks_bytecode_runtime_merged.cpu.pprof`;
- readable profile tables and separate inline counters:
  `2026-07-23-concurrent-policy-callbacks-{compiled,bytecode}-profile-top.txt`
  and
  `2026-07-23-concurrent-policy-callbacks-bytecode-runtime-stats.json`.

## Verification

- exact output parity in tree-walker, bytecode, compiled Able, Go, Python, and
  Ruby;
- ten verifier-backed timed processes per compiled, bytecode, and reference
  lane;
- three clean compiled and three clean warmed bytecode profiles;
- focused catalog, selection, coverage, operation-depth, matrix, triple, and
  scorecard checks;
- every added source file remains below 1,000 lines;
- JSON, source-identity, whitespace, and diff checks.

## Next recommendation

Recompute and promote the governed interaction frontier, then continue with
its highest-ranked remaining minimum-depth portable interaction.

Why: this application raises the requested callable/interface interaction to
depth four without exposing an open implementation leaf. The next application
must use a materially different semantic shape if that interaction remains
the sole shallow frontier; repeating this worker-and-numeric-policy shape
would add timing volume without improving evidence breadth.

What it entails: use the promoted 54-application scorecard to select one
deterministic application shape, build source-equivalent Able/Go/Python/Ruby
lanes and an exact verifier, take two five-process cohorts, and admit a
production change only for an exact generic owner repeated across at least
three unlike programs. Update canonical `able-stdlib` only for a reusable API
or correctness defect, and do not begin WASM work.
