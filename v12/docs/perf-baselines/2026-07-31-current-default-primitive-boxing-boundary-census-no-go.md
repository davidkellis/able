# Current-default primitive-boxing boundary census: no-go

Date: 2026-07-31

## Decision

Retain no compiler or runtime change. Do not launch a runtime profile or A/B
prototype from this tranche.

The ordinary current compiler generated all 66 strict applications and linked
the interpreter in none of them. Across the direct-main-reachable compiled
bodies, only five exact primitive-to-`runtime.Value` identities span three
unlike workload groups. Two are one-shot output conversions in `main`; the
other three are the canonical runtime-backed `HashMap` service boundary. No
non-named semantic category clears the admission gate.

This is the intended negative result: compiled application work continues to
use native Go primitive carriers, and the remaining broad encodes occur only
where values enter an explicit host or runtime-service ABI.

## Protocol and coverage

The disk-backed census used the current ordinary compiler with
`--no-fallbacks`, two generation workers, a 55-second per-process timeout, and
the existing schema-4 exact semantic-parent analyzer. Nominal proof-report
collection was disabled; ordinary default ownership execution remained
enabled.

- 66 applications selected and 66 generated successfully;
- 66 final dependency graphs resolved;
- zero failed or dependency-failed rows;
- zero interpreter-linked products;
- 15 boundary categories and 1,304 exact category/callee/parent identities;
- generation ranged from 391 ms to 13.164 seconds, averaging 4.252 seconds;
- generated modules ranged from 1,124,525 to 9,145,785 bytes; and
- generated modules ranged from 30,172 to 213,704 Go lines.

The primitive-encode slice contains 40 exact identities, 225
identity/application reaches, and 307 static sites. Nineteen identities reach
at least three applications, but only five span three workload groups. The
widest identity reaches 29 applications and six groups. No direct-reachable
`ToBool`, `ToRune`, or `ToFloat32` identity appears.

The compact JSON companion records the exact compiler, analyzer, frontier, and
raw census SHA-256 identities.

## Exact cross-group review

| Encode and semantic parent | Apps | Groups | Sites | Classification |
| --- | ---: | ---: | ---: | --- |
| `ToString` at `main` | 29 | 6 | 38 | one-shot print/host ABI |
| `ToDynamicI64` at `HashMap.raw_set` | 11 | 3 | 19 | named `HashMap` runtime-service ABI |
| `ToInt` at `HashMap.with_capacity` | 5 | 3 | 5 | named `HashMap` runtime-service ABI |
| `ToFloat64` at `main` | 5 | 3 | 9 | one-shot print/host ABI |
| `ToDynamicI64` at `HashMap.raw_get` | 4 | 3 | 6 | named `HashMap` runtime-service ABI |

Generated-source inspection confirms the classification:

- Distance Field computes with native `float64` and boxes only its final value
  for `print`.
- Fib computes with native `i32` and boxes only its final value for `print`.
- Versioned Telemetry Pipeline boxes its final String for `print`; its fresh
  profiled `Sample` pointer work remains inside the native specialized Array.
- Concurrent Event Routing's `main_ctx` i64 encodes are child-spawn return
  values entering Future scheduler payloads, while its final String encode is
  output.

The wider `main_ctx`, Channel, and Mutex identities do not broaden the result.
They repeat only in one or two groups and are explicit concurrency
scheduler/runtime-service ABI. Other `HashMap` variants reach at most two
groups.

The exact primitive-encode breadth is unchanged from the 2026-07-30
semantic-parent census: 40 identities, 19 reaching three applications, five
reaching three groups, with maxima of 29 applications and six groups. The
newer current-default products therefore do not invalidate that closure.

## Why no profile or prototype followed

The plan required static breadth before expensive profiling:

1. a category must occur in at least three unlike current applications;
2. its grouping may not depend on a container, nominal type, application, or
   source-family name; and
3. it must plausibly be material application work rather than cold host ABI or
   an already-closed runtime boundary.

The output identities fail the third condition. The `HashMap` identities fail
the second, and fresh K-Nucleotide/Inventory profiles already show that their
materiality belongs specifically to that runtime-backed map boundary.
Profiling these identities again could not authorize a general compiler rule.

No compiler, generated runtime, interpreter, bytecode VM, parser, language,
canonical stdlib, dependency, benchmark, fixture, frozen workspace, or WASM
source changed.

## Verification and cleanup

The complete census stayed within the one-minute per-process rule. Focused
analyzer tests and the census wrapper syntax check pass. The exact 364,736 KiB
task workspace and its pointer were removed from `/var/tmp` after compact
evidence publication; no task artifact was placed in RAM-backed `/tmp`.

## Next recommendation

Run the mode-aware evidence selector and ordinary release checks once against
this closed current state. If they remain empty and green, pause speculative
performance mutation and refresh the v12 correctness/stdlib completeness
selection audit before choosing another bounded tranche.

Why: this whole-corpus census exhausts the current primitive-boxing route
without finding a legal, general owner.

What it entails: verify all performance-closure identities and the
zero-actionable frontier, run focused compiler/census guards, then inventory
the current spec TODOs, fixtures, and canonical stdlib tests for one
reproducible non-WASM gap.

Why it matters: it prevents repeated work on cold host ABI or prohibited
named-container rules while ensuring a future native-lowering opportunity is
opened by real semantic, compiler, stdlib, or benchmark invalidation.
