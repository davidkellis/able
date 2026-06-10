# External Application Scoreboard

- Source measurements through: `2026-07-17T19:54:54.103001Z`
- Scope: verifier-backed Able application launches only; tree-walker is intentionally excluded from this performance target.
- Guard: each source scorecard records its process count, CPU-affinity when used, runtime settings, and per-process timeout.
- Compiled: `0/3` selected rankable rows meet the 95%-of-Go target.
- Bytecode: `0/3` selected rankable rows meet both 95%-of-Python and 95%-of-Ruby targets.
- Canonical Able source fingerprints: `6` row fingerprints in JSON; `6` came from the measured source report and the remainder are current-source legacy fingerprints.
- Verifier/declared-input contracts: `6` row fingerprints in JSON; `6` were captured before the timed launch and the remainder are current-contract legacy reconstructions.
- Canonical stdlib runtime sources: `70` `.able` files, tree SHA-256 `f7a470aae4fba342e5bbc3fce53ee26fa6f96df71dde18e057e044520624dafc`; Git `219eff222c28406487231713753641bc49ee5b9a` (dirty).
- Matched reference source fingerprints: `9` comparison fingerprints in JSON; `9` came from measured reference reports and the remainder are current-source legacy fingerprints.
- `unranked` means a partial, timed-out, failed, or unavailable matched run/reference; it is never counted as a pass or fail.
- `Unranked reason` identifies whether the Able launch or its required reference prevents ranking; reference-unavailable does not infer why that source has no valid ratio.

| Benchmark | Mode | Able status | Able (s) | Go / ratio | Python / ratio | Ruby / ratio | Target | Unranked reason |
| --- | --- | --- | ---: | --- | --- | --- | --- | --- |
| `regex_set_audit` | `compiled` | `verified` | 0.1260 | 0.0054 / 23.33x | n/a | n/a | `miss` | — |
| `regex_stream_audit` | `compiled` | `verified` | 0.1380 | 0.0047 / 29.36x | n/a | n/a | `miss` | — |
| `array_slice_window` | `compiled` | `verified` | 0.0900 | 0.0053 / 16.98x | n/a | n/a | `miss` | — |
| `regex_set_audit` | `bytecode` | `verified` | 4.9600 | n/a | 0.0236 / 210.17x | 0.0536 / 92.54x | `miss` | — |
| `regex_stream_audit` | `bytecode` | `verified` | 4.1780 | n/a | 0.0198 / 211.01x | 0.0496 / 84.23x | `miss` | — |
| `array_slice_window` | `bytecode` | `verified` | 0.7240 | n/a | 0.0298 / 24.30x | 0.0606 / 11.95x | `miss` | — |

## Source scorecards

- `v12/docs/perf-baselines/2026-07-17-regex-carrier-scorecard-compiled.json` — `custom` (`2026-07-17T19:53:50.959584Z`)
- `v12/docs/perf-baselines/2026-07-17-regex-carrier-scorecard-bytecode.json` — `custom` (`2026-07-17T19:54:54.103001Z`)

Regenerate after a new verifier-backed source scorecard with:

```sh
just bench-scoreboard
```

To replace the selected sources, pass each new scorecard explicitly, for example
`just bench-scoreboard --input path/to/compiled.json --input path/to/bytecode.json`.

Validate the checked-in report without running performance workloads with:

```sh
just bench-scoreboard-check
```
