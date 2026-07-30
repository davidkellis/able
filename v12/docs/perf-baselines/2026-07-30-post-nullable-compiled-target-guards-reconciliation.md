# Post-nullable compiled target-guard reconciliation

## Decision

All six compiled target guards remain protected. Retain no production change.

Fresh post-carrier strict builds and generated-source reach analysis find no
primitive nullable carrier in any material protected owner. QuickSort has an
outer absent nullable result, but its recursive sorting and Array-mutation
body remains carrier-free. No new profile or timing cohort was warranted.

## Strict boundary and execution gate

Every application was rebuilt with `--no-fallbacks` and passed one
contract-accurate public-verifier smoke process:

| Application | Packages | Interpreter dependency | Verified smoke |
| --- | ---: | --- | --- |
| Base64 | 119 | absent | 1/1 |
| Binary Trees | 96 | absent | 1/1 |
| JSON | 96 | absent | 1/1 |
| Monte Carlo Pi | 96 | absent | 1/1 |
| PiDigits | 96 | absent | 1/1 |
| QuickSort | 96 | absent | 1/1 |

Base64's additional packages come from its native codec integration. The
strict graph still contains no interpreter dependency.

Smoke durations are not promoted as timing evidence. The target
classification remains based on the two independent five-process stability
cohorts and the current five-process scorecard.

## Primitive-carrier reach

| Application | Generated functions/entries | Application carrier references | Material-owner references |
| --- | ---: | ---: | ---: |
| Base64 | 6 | 0 | 0 |
| Binary Trees | 20 | 0 | 0 |
| JSON | 6 | 0 | 0 |
| Monte Carlo Pi | 6 | 0 | 0 |
| PiDigits | 12 | 0 | 0 |
| QuickSort | 8 | 16 absent-result references | 0 |

The Binary Trees count includes the ordinary wrappers and the spawn-gated
execution-context bodies. Its recursive `make_tree`/`check_tree` owners have
no carrier reach.

Base64's application body calls native `encode_bytes`, `decode_bytes`, and
MD5 `hex`. Those imported material bodies also contain no primitive nullable
carrier.

JSON's material `f64_field_means` host decoder and application body contain no
primitive carrier. A one-time file-read error path contains
`__able_nullable_error_to_value`; `?Error` was explicitly excluded from the
retained representation change and is therefore not changed-mechanism reach.

Monte Carlo and PiDigits remain direct scalar and `math/big` kernels with no
carrier reach. PiDigits also has a direct five-process nullable baseline and
candidate cohort: 1.258 and 1.276 seconds respectively. The 1.43% candidate
increase did not overturn the guard: both cohorts verified, and the
authoritative current five-process row is 1.234 seconds versus Go at 1.2218.

QuickSort's generated `main` returns the absent result of its final output
loop. The 16 references are the value-carrier return type, absent literals,
and error returns; there is no `__able_some` construction. Its recursive
`quicksort` owner contains zero carrier references, and its current ratio has
a large target margin.

## Current guard state

| Application | Able compiled | Go | Able / Go |
| --- | ---: | ---: | ---: |
| Base64 | 2.4480 s | 2.6198 s | 0.9344x |
| Binary Trees | 10.7140 s | 10.6498 s | 1.0060x |
| JSON | 1.2160 s | 1.4767 s | 0.8235x |
| Monte Carlo Pi | 0.2040 s | 0.2024 s | 1.0079x |
| PiDigits | 1.2340 s | 1.2218 s | 1.0100x |
| QuickSort | 1.9140 s | 2.6816 s | 0.7138x |

All ratios remain below the 1.052632x threshold corresponding to at least 95%
of Go throughput.

## Scope

No compiler, generated runtime, runtime package, interpreter, bytecode VM,
canonical stdlib, benchmark, language, dependency, or WASM source changed.
No application, named-container, non-primitive nominal, or benchmark-specific
rule was introduced.

The machine-readable record is
`2026-07-30-post-nullable-compiled-target-guards-reconciliation.json`.
After verification, the exact 728 MiB generated-module/binary workspace,
125 MiB focused-test cache, and generated Python cache were removed. No
matching `/var/tmp` artifact remains.

## Next

Reconcile `compiled-iterator-control` against the retained carrier.

Why: Generic Slot Buffer belongs to this closure and is the application where
the primitive nullable representation was materially dominant. Its prior
pointer-allocation owner was intentionally removed, so this is the nearest
remaining closure with proven causal reach.

What it entails: register the retained Generic Slot Buffer A/B and exact
allocation evidence, then scan the other iterator/control rows for primitive
carrier reach. Reprofile only a materially reached row whose residual owner
could now repeat across three unlike applications.

Why it matters: the retained optimization must update the closure it actually
changed and determine whether removing the nullable allocation reveals a new
general successor, without reopening closed call, frame, Array, or nominal
routes.
