# Compiler struct-field reach balanced measurement

`21` rotating cohorts produced `21` independent processes per lane.
Every successful stdout was checked by the application's public verifier.

| Application | Baseline mean | Candidate mean | Candidate vs baseline | Go mean | Candidate/Go time | Candidate as Go performance |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| `await_channel_mux` | 0.254228s | 0.250492s | -1.47% | 0.004894s | 51.18x | 1.95% |
| `future_await_race` | 0.034133s | 0.035581s | +4.24% | 0.004366s | 8.15x | 12.27% |
| `mutex_await_journal` | 0.391587s | 0.394607s | +0.77% | 0.004099s | 96.28x | 1.04% |
| `mutex_work_queue` | 0.945521s | 0.987128s | +4.40% | 0.004447s | 221.96x | 0.45% |

Raw samples: `2026-07-27-compiler-struct-field-reach-balanced-21-samples.tsv` (`283260be52e982e10c4f943a6cdd1407feca11b7de5bfc8c6bd715114545361f`).
