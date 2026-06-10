# Compiled generated-allocation shape reconciliation

Date: 2026-07-21

## Decision

Keep no compiler, generated-runtime, bridge, stdlib, application, or reference performance change. The same exact allocation shape is not CPU-material in three unlike families.

The retained diagnostic addition is `ABLE_GO_PHASE_STATS_DIR`, a lightweight main/bootstrap `runtime.MemStats` delta mode. It does not enable one-object pprof sampling and is inert when unset.

## Allocation amplification

| Application | Able bytes / objects | Go bytes / objects | Byte / object amplification | Logical normalization |
| --- | ---: | ---: | ---: | ---: |
| `k_nucleotide` | 1,224,424,432 / 27,734,581.5 | 4,959,208 / 85.0 | 246.90x / 326289.19x | 612.212 B, 13.867 objects per target-sequence bases |
| `fixed_width_128` | 35,535,960 / 2,220,982.0 | 400 / 7.0 | 88839.90x / 317283.14x | 17.768 B, 1.110 objects per checksum loop iterations |
| `distance_field` | 144 / 7.0 | 384 / 5.0 | 0.38x / 1.40x | 0.000 B, 0.000 objects per numeric loop iterations |
| `policy_record_dispatch` | 49,167,204 / 970,529.5 | 322,736 / 3,366.0 | 152.34x / 288.33x | 24007.424 B, 473.891 objects per records processed |
| `concurrent_event_routing` | 77,771,480 / 1,502,286.0 | 604,968 / 4,162.0 | 128.55x / 360.95x | 18987.178 B, 366.769 objects per routed tasks |

Distance Field is the counterexample: generated native float work allocates only seven objects. The other rows therefore do not identify a universal generated-Go allocation tax.

## Exact mechanism gate

| Mechanism | Applications | CPU shares | Allocation-object shares | Material CPU families | Disposition |
| --- | --- | --- | --- | ---: | --- |
| `builtin-string-conversion` | k_nucleotide, policy_record_dispatch, concurrent_event_routing | 22.03%, 15.38%, 0.64% | 25.98%, 7.80%, 11.33% | 2 | `insufficient-three-family-cpu-leverage` |
| `bridge-to-uint` | k_nucleotide, policy_record_dispatch, concurrent_event_routing | 14.75%, 3.85%, 0.00% | 35.86%, 11.37%, 9.37% | 2 | `causally-closed-and-insufficient-three-family-cpu-leverage` |
| `loop-carried-nominal-result` | fixed_width_128 | 89.29% | 98.63% | 1 | `single-family` |
| `regex-nfa-storage` | policy_record_dispatch | 46.15% | 31.00% | 1 | `single-family` |
| `goroutine-identity-discovery` | concurrent_event_routing | 91.26% | 11.08% | 1 | `single-family-and-causally-closed` |

`builtin-string-conversion` and `bridge-to-uint` allocate in three families, but Event Routing spends only 0.64% and less than the CPU sampling threshold in those paths. Its 91.26% `currentGID` path is a separate, previously rejected concurrency owner. Fixed Width is loop-carried nominal allocation, Policy is regex storage plus conversions, and K-Nucleotide is map boxing/text conversion.

The prior general `ToUint` branch-layout candidate improved Word Frequency but regressed K-Nucleotide, and the unsigned census found the hot reused-key shape only in K-Nucleotide. The new Policy/Event allocation recurrence does not invalidate that causal rejection because neither supplies comparable CPU leverage.

## Protocol and verification

Five generated trees were preserved before measurement. Every row received two independent main-only CPU processes and two independent lightweight allocation-counter processes under a 55-second cap, `GOMAXPROCS=1`, `GOGC=50`, and a 1-GiB limit. All outputs passed their public Ruby verifiers. Reference Go mains were wrapped in temporary copies with the same two-process MemStats boundary and also verified.

K-Nucleotide's exact one-object phase profile again hit the 55-second cap and is missing, not zero. Its ordinary cumulative sampled allocation profile completed and supplied caller attribution; the authoritative main counts come from the lightweight counter mode.

Focused profilehook tests and this report's contract tests pass. Temporary generated trees, binaries, profiles, escape logs, and copied Go references are removed after the checked record is generated.

## Next recommendation

Return to bytecode semantic cost, starting with a non-Array application family whose exact semantic helper is repeated across at least three unlike misses.

Why: this tranche closes the broad compiled allocation parent without admitting a safe local compiler candidate. The compiler's remaining owners are domain-specific and require larger representation work, while the bytecode target remains much farther from Python/Ruby and its architecture audit explicitly directs selection toward semantic operations rather than another transport tweak.

What it entails: refresh bounded CPU, allocation, and opt-in semantic counters for text/UTF-8 conversion across K-Nucleotide, Word Frequency, Policy Dispatch, and an unlike non-text control; reconcile one exact helper below call/slot parents; and prototype only if the same removable semantic conversion is material in three families with useful end-to-end leverage. Keep compiler guards, do not alter benchmark algorithms, do not add named-container rules, and do not begin WASM work.
