# Channel-Rollup Isolated Docker Publication

This report publishes a refreshed Docker result per implementation for the new
Channel-Rollup application. It is isolated and report-only: the dirty
shared `../benchmarks/results.json` was not read, modified, or replaced.

## Provenance

- Input: first 16,384 records of `i-before-e/wordlist.txt` (1,743,363 bytes;
  SHA-256 `3f16130220645692ed49c7134e24a18504c2ca55b3c012f7290e3e77c63b1a89`).
- Expected and verified output for every image: `16384:4828:502100`.
- Images: Able v12 compiled, bytecode, tree-walker; Go 1.26; Python 3.14; and
  Ruby 4.0.
- Method: one direct `docker run` per locally built image; each image invokes
  the shared `time.sh` high-resolution process-wall timer.
- Executor policy: every Able image sets `ABLE_EXECUTOR=goroutine`, matching
  the concurrent workload.
- Refresh: `able-v12-base` and all six images were rebuilt after the completed
  generic dynamic and primitive ArrayStore synchronization repair. Every
  rebuilt image again verified the expected output.

The JSON companion records the exact values. These one-run rows are valid only
within Channel-Rollup. They establish a target-miss selection signal, not a
replacement for a multi-host or shared-results publication.

| Mode | Able (s) | Go (s) | Able/Go | Ruby (s) | Able/Ruby | Python (s) | Able/Python |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Compiled | 1.655721179 | 0.006852194 | 241.63x | 0.073451352 | 22.54x | 0.105122927 | 15.75x |
| Bytecode | 4.462437608 | 0.006852194 | 651.24x | 0.073451352 | 60.75x | 0.105122927 | 42.45x |
| Tree-walker | 27.374979696 | 0.006852194 | 3995.07x | 0.073451352 | 372.70x | 0.105122927 | 260.41x |

These are refreshed one-run rows. The prior pre-completion row set is
superseded for current-runtime provenance, not used as a before/after claim.
The tree-walker row remains semantic verification rather than a performance
selection signal; its one-run process timing is especially variable.

## Selection

Channel-Rollup materially misses both the compiled Go target and the bytecode
Ruby/Python target. Its bounded bytecode profile pair does not identify a
material shared scheduler/channel descendant with BinaryTrees: only generic
async parents repeat, while concrete descendants diverge and the full
BinaryTrees target times out. Do not optimize the new benchmark source,
channel capacity, or one runtime implementation in isolation.

## Next recommendation

The follow-up multi-run target-miss ledger is
`2026-07-10-target-miss-ledger-refresh.{md,json}`. It confirms current Able
process completion across Word-Frequency, Document-Audit, Lexical-Rollup, and
Channel-Rollup but deliberately has no reference ratios: the dirty shared
external JSON contains no matching reference rows. The next work should profile
the cold-process bytecode boundary in two independent target-miss applications
with a sequential guard. Why: the Docker process rows and host local process
rows differ materially, so a loader/stdlib/runtime initialization wall must be
separated from warm steady-state VM work before selecting an optimization. Do
not tune a channel, executor, container, or benchmark source from these rows.
