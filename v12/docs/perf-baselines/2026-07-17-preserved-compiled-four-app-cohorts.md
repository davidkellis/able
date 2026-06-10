# Preserved compiled cohort comparison

Two-phase protocol: build all binaries first, then time 2 cohorts with 5 runs each. Cohort spread limit: 15.0%.

| Benchmark | Able mean | Cohort spread | Go mean | Able / Go | Eligible |
|---|---:|---:|---:|---:|:---:|
| reverse_complement | 0.1060 s | 12.0000% | 0.0160 s | 6.6250x | yes |
| rational_series | 0.1250 s | 19.2982% | 0.0130 s | 9.6154x | no |
| word_frequency | 0.2150 s | 6.7308% | 0.0053 s | 40.5660x | yes |
| array_slice_window | 0.0840 s | 4.8780% | 0.0043 s | 19.5349x | yes |

Overall promotion eligible: **no**
