# Interface-dictionary capture performance-evidence reconciliation

Date: 2026-07-30

## Decision

**Rebaseline the reviewed interface-dictionary production scopes and admit no
performance candidate.**

All 23 closures were selected because the tranche changed the compiler,
bytecode VM, shared interpreter semantics, and v12 spec. The complete
mode-specific set cover now supplies current evidence for those scopes.

## Structural proof

- Strict static generation: 66/66 applications succeeded with
  `--no-fallbacks`.
- Dynamic-boundary audit: 10/10 compiled applications verified.
- Final interpreter dependencies: 0/10.
- Runtime-service calls: 0.
- Expected telemetry totals: 156 explicit dynamic calls, 169 residual
  polymorphic calls, and 148 host-ABI calls.

The generated applications therefore remain fallback-free and
interpreter-free. The new captured dictionaries do not create a hidden
compiled/interpreted crossing.

## Repeated evidence

Ten compiled and nine bytecode applications cover all 23 closures. Every Able
application and every Go/Python/Ruby reference has two opposite-order
five-process cohorts on CPU pool `12,13,14,15`.

- Verified timed processes: 470/470.
- Failures: 0.
- Timeouts: 0.
- Stable benchmark/mode rows: 14.
- Volatile rows: 5.

| Application | Mode | Able mean | Able cohort spread | Reference ratio | Volatile |
| --- | --- | ---: | ---: | --- | :---: |
| reverse_complement | compiled | 0.0440s | 20.00% | 1.36x Go | yes |
| concurrent_event_routing | compiled | 0.0360s | 25.00% | 2.98x Go | yes |
| distance_field | compiled | 0.0350s | 33.33% | 1.13x Go | yes |
| array_slice_window | compiled | 0.0300s | 0.00% | 2.90x Go | no |
| quicksort | compiled | 1.8030s | 3.96% | 0.39x Go | no |
| policy_record_dispatch | compiled | 0.0750s | 14.29% | 6.25x Go | no |
| fib | compiled | 3.5900s | 1.69% | 0.68x Go | no |
| k_nucleotide | compiled | 1.5040s | 4.63% | 12.38x Go | no |
| fixed_width_128 | compiled | 0.0940s | 13.64% | 7.89x Go | no |
| sudoku_masks | compiled | 1.5000s | 7.18% | 1.21x Go | no |
| reverse_complement | bytecode | 3.9430s | 51.95% | 130.78x Python / 50.09x Ruby | yes |
| future_await_race | bytecode | 0.1350s | 4.55% | 4.45x Python / 2.59x Ruby | yes |
| monte_carlo_pi | bytecode | 2.4200s | 2.85% | 1.61x Python / 1.52x Ruby | no |
| document_audit | bytecode | 0.2810s | 3.62% | 20.03x Python / 6.35x Ruby | no |
| nbody | bytecode | 8.8780s | 0.41% | 40.45x Python / 25.98x Ruby | no |
| config_validation_extraction | bytecode | 1.2180s | 3.00% | 70.59x Python / 31.06x Ruby | no |
| fib | bytecode | 0.1840s | 4.44% | 0.04x Python / 0.04x Ruby | no |
| i_before_e | bytecode | 0.4680s | 0.86% | 5.37x Python / 3.81x Ruby | no |
| rational_series | bytecode | 3.9400s | 0.20% | 37.78x Python / 28.88x Ruby | no |

`future_await_race` is volatile because a reference cohort spread exceeds 15%
even though Able's spread is 4.55%. The other volatile rows are identified by
their Able spread. All volatile ratios remain descriptive.

An unrelated Marketlab pipeline imposed sustained host load during this
reconciliation. The opposite-order application and reference cohorts mostly
reproduced closely; the five rows above preserve the exceptions. These
measurements establish current verified execution and closure coverage, not a
causal speed improvement.

## Interpretation

The correctness change is safe to retain:

- no strict application regained an interpreter dependency;
- no runtime-service crossing appeared;
- the complete compiled and bytecode set cover executes correctly;
- the isolated three-application A/B remains effectively neutral.

The evidence does not expose one new material owner shared across three unlike
applications. Existing broad misses retain their prior dispositions. In
particular, k-Nucleotide's primitive HashMap bridge cost remains a one-family
owner and cannot justify a named-container rule.

## Evidence

- `2026-07-30-interface-dictionary-capture-pre-reconciliation-invalidation.json`
- `2026-07-30-interface-dictionary-capture-strict-static-census.json`
- `2026-07-30-interface-dictionary-capture-boundary-audit.json`
- `2026-07-30-interface-dictionary-capture-compiled-cohorts.json`
- `2026-07-30-interface-dictionary-capture-go-references-forward.json`
- `2026-07-30-interface-dictionary-capture-go-references-reverse.json`
- `2026-07-30-interface-dictionary-capture-bytecode-forward.json`
- `2026-07-30-interface-dictionary-capture-bytecode-reverse.json`
- `2026-07-30-interface-dictionary-capture-interpreter-references-forward.json`
- `2026-07-30-interface-dictionary-capture-interpreter-references-reverse.json`

The checked compact summary is
`2026-07-30-interface-dictionary-capture-performance-evidence-reconciliation.json`.

## Next

Profile native dictionary runtime reach in
`concurrent_graph_visitors`, `concurrent_event_routing`, and
`validated_job_pipeline`.

Why: these interface-heavy applications have three of the largest
generated-source contractions in the 66-application census, while the current
algorithmic A/B trio did not execute an interface adapter as a hot owner.

What it entails: capture current CPU, `alloc_space`, and typed-boundary
profiles; prove whether exact-`Self` native adapters execute materially; select
one exact repeated boundary owner; then use a pre-change compiler and
equivalent Go applications for balanced five-or-more-pair A/B. Retain no code
if the owner is absent, one-family, or only launch-dominated.

Why it matters: this is the shortest general path from the newly sound
dictionary semantics to fewer `runtime.Value` conversions and fewer
compiled/interpreted crossings in real interface-heavy programs.
