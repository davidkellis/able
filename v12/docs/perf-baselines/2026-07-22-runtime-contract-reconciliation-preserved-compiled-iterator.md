# Preserved compiled cohort comparison

Two-phase protocol: build all binaries first, then time 2 cohorts with 5 runs each. Cohort spread limit: 15.0%.

| Benchmark | Able mean | Cohort spread | Go mean | Able / Go | Eligible |
|---|---:|---:|---:|---:|:---:|
| array_slice_window | 0.1010 s | 2.0000% | 0.0063 s | 16.0317x | yes |
| binary_event_log | 0.5970 s | 1.6892% | 0.0108 s | 55.2778x | yes |
| dependency_plan | 0.1280 s | 56.0000% | 0.0048 s | 26.6667x | no |
| document_audit | 0.1240 s | 53.0612% | 0.0054 s | 22.9630x | no |
| lexical_rollup | 0.1230 s | 5.0000% | 0.0052 s | 23.6538x | yes |
| option_result_config | 0.2100 s | 18.7500% | 0.0049 s | 42.8571x | no |

Overall promotion eligible: **no**
