# External Threshold Controls

This is a reproducible **report-only** classification of retained five-pair application evidence. It never runs a benchmark and is not a commit performance threshold.

- Raw target: Able/reference ratio at or below `1.0526x` (95% of the applicable Go, Python, or Ruby reference speed).
- Guard band: `21%`; `inside` requires every pair against every required reference at or below `0.8316x`. `outside` requires every pair against at least one required reference at or above `1.2737x`.
- Controls: `4` inside, `2` boundary, `2` outside.
- A control is evidence for broad policy only; it never authorizes a benchmark-specific runtime or compiler optimization.

| Control | Mode | References | Family | Verified pairs | Ratio ranges | Medians | CVs | Classification |
| --- | --- | --- | --- | ---: | --- | --- | --- | --- |
| `quicksort` | `compiled` | go | array/index algorithmic static code | 5 | go: 0.64x–0.79x | go: 0.74x | go: 7.90% | `inside` |
| `json` | `compiled` | go | text, file, and JSON codec work | 5 | go: 0.44x–0.67x | go: 0.63x | go: 19.07% | `inside` |
| `binarytrees` | `compiled` | go | recursive nominal allocation and traversal | 5 | go: 0.81x–0.99x | go: 0.92x | go: 7.09% | `boundary` |
| `base64` | `compiled` | go | byte/string codec boundary | 5 | go: 0.99x–1.09x | go: 1.01x | go: 4.28% | `boundary` |
| `nbody` | `compiled` | go | primitive numeric/package work | 5 | go: 12.43x–15.49x | go: 13.95x | go: 8.61% | `outside` |
| `k_nucleotide` | `compiled` | go | text/map boxing and conversion | 5 | go: 51.29x–64.59x | go: 59.49x | go: 9.08% | `outside` |
| `json` | `bytecode` | python, ruby | text, file, and JSON codec work | 5 | python: 0.28x–0.36x; ruby: 0.37x–0.53x | python: 0.33x; ruby: 0.49x | python: 10.09%; ruby: 15.21% | `inside` |
| `pidigits` | `bytecode` | python, ruby | native BigInt arithmetic and formatted output | 5 | python: 0.58x–0.71x; ruby: 0.21x–0.27x | python: 0.59x; ruby: 0.23x | python: 8.86%; ruby: 9.20% | `inside` |

## Retained aggregate reports

- `v12/docs/perf-baselines/2026-07-13-compiled-threshold-protocol.json`
- `v12/docs/perf-baselines/2026-07-13-compiled-json-threshold-protocol.json`
- `v12/docs/perf-baselines/2026-07-13-compiled-binarytrees-threshold-protocol.json`
- `v12/docs/perf-baselines/2026-07-13-bytecode-json-pidigits-threshold-protocol.json`

Regenerate this report after replacing retained evidence:

```sh
just bench-threshold-controls
```

Validate report provenance without running performance workloads:

```sh
just bench-threshold-controls-check
```
