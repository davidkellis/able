# Bytecode negative-field plan admission — 2026-07-17

## Decision

Reject lowering-time negative-field metadata before implementation. Exact
runtime exposure reaches only 53 calls across two fixture programs and two
nominal definitions. The two application controls have zero eligible calls,
so the proposed plan does not meet the required three-program admission bar.

No runtime, lowering, compiler, stdlib, fixture, language, or tree-walker
change is retained. Temporary metadata, counters, binaries, and reports are
removed after this record is written.

## Method

The temporary lowering pass reused the existing positive named-struct member
plan safety model. At a member-call instruction it recorded an absent-field
marker only when lowering already knew the receiver's exact nominal struct
definition and that immutable definition did not declare the member name.

Runtime diagnostics then counted a site only when the actual receiver's
definition matched the lowering-time definition. No field probe was skipped,
so these runs measured candidate exposure without changing dispatch behavior.
Counters reset after benchmark warmup and covered one measured `main()` call
per fresh process under CPU 0, `GOMAXPROCS=1`, `GOGC=50`, and
`GOMEMLIMIT=1GiB`.

The external applications used their canonical benchmark contracts:

- Document Audit ran from the external benchmark root with
  `word-frequency/corpus.md`.
- Lexical Rollup ran from the same root with
  `i-before-e/wordlist.txt`.

## Exact exposure

| Workload | Matched executions | Definition/member | Result |
| --- | ---: | --- | --- |
| Run-length encode | 49 | `StringBuilder.finish` | eligible but immaterial relative to the workload's hundreds of thousands of other member calls |
| Iterator Collect | 4 | `LinkedList.lazy` | eligible but immaterial |
| Document Audit | 0 | none | no application exposure |
| Lexical Rollup | 0 | none | no application exposure |

The plan does not cover Run-length's hot `StringBuilder.push_char` and
`push_string` calls because those receiver slots originate from constructor
call results rather than an exact nominal slot proof. It does not cover the
application pipelines because their receivers flow through chained call
results, generic containers, and interfaces. Those are precisely the ordinary
real-program shapes that the application controls were intended to test.

Extending lowering's exact nominal dataflow merely to make this candidate
appear would be a substantially larger semantic analysis change, not evidence
that negative-field metadata itself is material. It would also still leave the
interface-heavy pipelines uncovered. The candidate therefore stops at the
admission gate; wall-clock A/B testing would measure effectively zero exposure
in both application controls.

## Closure

This closes the current field/member micro-optimization sequence. Do not retry
the same-dispatch boolean shortcut, small-definition maps, inline field-name
storage, fixed string-iterator indices, or exact-slot-only negative-field
metadata without a new application-level profile showing material exposure.

## Next recommendation

Pivot the next tranche to compiled application performance and refresh
main-only CPU/allocation profiles for Word Frequency, Document Audit, Lexical
Rollup, and Dependency Plan against their verified Go programs.

Why: only six of 34 compiled applications currently meet the 95%-of-Go target,
while the last two bytecode field/member candidates failed first at the broad
wall gate and then at real-application exposure. The compiler gap is much
larger, and these four applications cover string maps, filesystem/text
processing, lazy iterator pipelines, arrays, and nominal data without relying
on one benchmark family.

What it entails: preserve the current verified input/output contracts; collect
fresh main-only CPU and allocation profiles under the one-process memory
guardrails; reconcile leaves across all four; and admit code only for a shared
runtime or general nominal-lowering cost material in at least three unlike
programs. Do not add named-container lowering rules, benchmark-specific
helpers, or WASM work.
