# Preserved compiled cohort comparison

Two-phase protocol: build all binaries first, then time 2 cohorts with 5 runs each. Cohort spread limit: 15.0%.

| Benchmark | Able mean | Cohort spread | Go mean | Able / Go | Eligible |
|---|---:|---:|---:|---:|:---:|
| word_frequency | 0.2000 s | 4.0816% | 0.0063 s | 31.7460x | yes |
| document_audit | 0.0810 s | 18.9189% | 0.0051 s | 15.8824x | no |
| dependency_plan | 0.0720 s | 5.7143% | 0.0046 s | 15.6522x | yes |

Overall promotion eligible: **no**
