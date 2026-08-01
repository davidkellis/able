# Composite-interface contract performance-evidence reconciliation

Date: 2026-07-30

## Decision

**rebaseline-reviewed-v12-spec-scope**.

All 66 shared benchmark sources remain valid under the composite-interface Self-pattern rule, and all 66 strict generated modules plus their compact native-boundary counts exactly match the prior census. The ten compiled and nine bytecode set-cover cohorts provide 470 verified Able/Go/Python/Ruby processes across all 23 closures. No production execution path changed, so the data admit a reviewed v12-spec scope rebase and no optimization candidate.

This is an evidence rebase, not a performance win. No compiler, runtime, interpreter, VM, stdlib, dependency, benchmark, or WASM production change is retained.

## Scope and native-boundary proof

- All 23 closures were invalidated only by `scope-content-drift:v12-spec`.
- Strict generation: 66/66 selected sources.
- Generated modules matching the prior census: 66/66.
- Compact native-boundary counts matching the prior census: 66/66.
- Set-cover dependency guard: 10 verified strict applications, zero interpreter dependencies, zero runtime-service crossings.

## Repeated-process evidence

All 470 Able/reference processes verified; zero failed or timed out. Each benchmark/mode and each reference has two opposite-order five-process cohorts. `volatile` means at least one Able/reference cohort spread exceeds 15%; volatile rows remain descriptive and do not admit a candidate.

| Benchmark | Mode | Able mean | Cohort spread | Reference ratio | Volatile |
| --- | --- | ---: | ---: | --- | :---: |
| reverse_complement | compiled | 0.0440s | Able 20.00% / ref 10.95% | 2.70x Go | yes |
| concurrent_event_routing | compiled | 0.0380s | Able 23.53% / ref 0.55% | 7.61x Go | yes |
| distance_field | compiled | 0.0300s | Able 0.00% / ref 88.67% | 1.74x Go | yes |
| array_slice_window | compiled | 0.0300s | Able 0.00% / ref 30.82% | 6.34x Go | yes |
| quicksort | compiled | 1.8450s | Able 4.09% / ref 6.15% | 0.67x Go | no |
| policy_record_dispatch | compiled | 0.0750s | Able 14.29% / ref 15.37% | 14.56x Go | yes |
| fib | compiled | 3.7550s | Able 0.59% / ref 8.82% | 1.13x Go | no |
| k_nucleotide | compiled | 1.6330s | Able 4.64% / ref 3.55% | 27.81x Go | no |
| fixed_width_128 | compiled | 0.1030s | Able 10.20% / ref 5.09% | 18.66x Go | no |
| sudoku_masks | compiled | 1.5960s | Able 4.35% / ref 9.34% | 2.39x Go | no |
| reverse_complement | bytecode | 4.4150s | Able 52.14% / ref 8.41% | 168.02x Python / 57.30x Ruby | yes |
| future_await_race | bytecode | 0.1540s | Able 2.63% / ref 20.31% | 4.80x Python / 3.14x Ruby | yes |
| monte_carlo_pi | bytecode | 2.7220s | Able 0.59% / ref 5.31% | 1.73x Python / 1.60x Ruby | no |
| document_audit | bytecode | 0.2930s | Able 0.68% / ref 18.87% | 18.44x Python / 6.75x Ruby | yes |
| nbody | bytecode | 9.5950s | Able 0.44% / ref 2.30% | 42.58x Python / 26.60x Ruby | no |
| config_validation_extraction | bytecode | 1.3580s | Able 6.38% / ref 31.67% | 69.54x Python / 33.59x Ruby | yes |
| fib | bytecode | 0.2160s | Able 5.71% / ref 0.95% | 0.04x Python / 0.05x Ruby | no |
| i_before_e | bytecode | 0.5410s | Able 7.28% / ref 16.43% | 5.94x Python / 4.34x Ruby | yes |
| rational_series | bytecode | 4.2290s | Able 0.14% / ref 9.71% | 37.62x Python / 28.46x Ruby | no |

## Interpretation

The new contract changes only front-end acceptance of invalid composite interface declarations. Complete strict generation proves that no benchmark or canonical-stdlib relationship is newly invalid. Exact generated-module and static native-boundary equality proves there is no changed compiled execution shape to profile as a candidate. The repeated cohorts confirm current execution without converting workstation variance into a causal claim.

All 23 closure families are intersected by the mode-specific set-cover cohorts. The reviewed shared spec scope can therefore be rebased while every earlier performance disposition remains unchanged.

## Cleanup

The exact 1,515,636 KiB disk-backed task workspace and 48 KiB task-created project Python cache were removed after retaining the checked evidence. No task artifact used RAM-backed `/tmp`, and no task-owned workspace remains.

## Next

Specify the interface-dictionary capture contract at static upcasts before attempting further native interface lowering.

Why: generic methods, default bodies, `Self` returns, and ambiguity at upcast remain unresolved in the v12 spec. Those semantics determine when a compiled interface value can keep a native Go dictionary instead of falling back to a runtime representation.

What it entails: define dictionary capture and ambiguity rules in the spec, add shared checker/interpreter fixtures, implement the same rule in tree-walker, bytecode, and strict compiled lowering only where required, then rerun the evidence selector and repeated set-cover guards. Do not begin WASM.

Why it matters: settling the semantic boundary first makes later reductions in boxing or compiled/interpreted crossings sound, general, and applicable to unlike programs.
