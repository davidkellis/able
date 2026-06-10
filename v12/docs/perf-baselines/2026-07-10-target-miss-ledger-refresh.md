# Current Target-Miss Ledger Refresh

This report is a small, report-only process-time refresh after rebuilding the
Channel-Rollup Docker images from the retained ArrayStore synchronization
repair. It does not read, modify, or replace the dirty
`../benchmarks/results.json` ledger.

## Method

- Canonical external stdlib; `GOMEMLIMIT=1GiB` and `GOGC=50`.
- Three normal processes per Able mode, including load/lower/build behavior as
  appropriate; no `GOMAXPROCS=1` cap because Channel-Rollup uses the goroutine
  executor.
- Sources are the checked-in v12 examples with the sibling benchmark inputs:
  Document-Audit, Lexical-Rollup, and Channel-Rollup match their sibling
  `run.able` byte-for-byte. Word-Frequency differs only in package name and
  formatting, with the same operations and input.
- The external JSON has no usable matching Go/Ruby/Python rows for these
  current process measurements, so the table intentionally records no ratios.
  Docker publication remains the cross-language source for Channel-Rollup.

## Results

| Workload | Mode | Able process average (s) | GC average | Status |
| --- | --- | ---: | ---: | --- |
| Word-Frequency | Compiled | 0.2300 | 8.00 | 3/3 passed |
| Word-Frequency | Bytecode | 1.4667 | 11.00 | 3/3 passed |
| Document-Audit | Compiled | 0.0967 | 6.00 | 3/3 passed |
| Document-Audit | Bytecode | 0.2700 | 8.67 | 3/3 passed |
| Lexical-Rollup | Compiled | 0.1033 | 6.67 | 3/3 passed |
| Lexical-Rollup | Bytecode | 0.3933 | 10.00 | 3/3 passed |
| Channel-Rollup | Compiled, goroutine | 2.1333 | 8.33 | 3/3 passed |
| Channel-Rollup | Bytecode, goroutine | 0.4567 | 9.33 | 3/3 passed |

## Decision

The ledger verifies current process completion across independent text/map,
public-iterator, and channel/Future applications. It does not select an
optimization: it has no clean reference ratios, and the prior profile pairs do
not identify a material concrete VM or compiler leaf shared across these
workloads. In particular, do not infer a win from differences between these
host-process rows and Docker process rows; the two environments have different
startup, cache, and runtime conditions.

## Next recommendation

Profile the cold-process bytecode boundary for two independent target-miss
applications, with a sequential guard, under Docker-like settings. Why: the
refreshed Docker and host-process rows differ enough that loader, stdlib, GC,
or initialization cost must be separated from warmed VM execution before any
candidate is credible. The work entails bounded one-process CPU profiles and
output checks for Channel-Rollup plus an independent text/map application,
with Lexical-Rollup as guard. Retain a candidate only if the same concrete
loader, stdlib, or runtime leaf recurs across both target-miss programs and
does not regress the guard; do not tune a benchmark source, channel capacity,
or scheduler parent.
