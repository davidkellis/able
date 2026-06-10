# Benchmark host preflight — 2026-07-14

## Decision

Do not record a fresh scorecard or CPU profile on the current host. A three
sample readiness check for the pinned CPU found material unrelated work and
I/O wait, so a timing refresh would not distinguish Able execution from host
contention.

This tranche adds a generic preflight rather than changing the VM, compiler,
runtime, canonical stdlib, application sources, or benchmark workload.

## Evidence

`v12/bench_host_cpu_check --cpu 15 --samples 3 --interval 1` uses the deltas
in Linux `/proc/stat` for that one logical CPU. Its default requires every
one-second sample to remain at or below 5% ordinary busy time and 5% I/O-wait
time. The current host failed all three samples:

| Sample | Busy | I/O wait |
| --- | ---: | ---: |
| 1 | 25.51% | 0.00% |
| 2 | 21.42% | 19.38% |
| 3 | 16.83% | 19.80% |

These are host-readiness observations, not benchmark measurements and must not
be compared with an Able, Go, Python, or Ruby result.

## Tooling contract

Run the check immediately before a pinned timing command:

```sh
just bench-host-check --cpu 15
```

The external scorecard refresh has an opt-in guard that performs the same
check before it creates report or temporary-workspace paths:

```sh
just bench-scorecard-refresh --cpu-affinity 15 --require-quiet-cpu
```

The direct timing entry points accept the same flag: `bench_perf`,
`bench_compare_external`, `bench_refresh_go_refs`, and
`bench_refresh_interpreter_refs`. The grouped scorecard checks only at its
outer boundary, so it does not ask every child process to repeat the same
three-second sample. Standalone manual measurements should pass the flag at
their own entry point.

`--require-quiet-cpu` intentionally accepts only one CPU number, because a
multi-CPU affinity set cannot give a meaningful one-core contention reading.
The check does not reserve the CPU; another process can still start after it
passes, so use it directly before the benchmark command. The ordinary
scorecard options remain unchanged for existing automation, while new manual
performance evidence should use the guard.

## Next measurement gate

A passing preflight does **not** by itself reopen benchmark selection or
authorize an unchanged scorecard rerun. Use it only after a material
cross-cutting semantic/compiler change, or a newly needed spec-defined
portable application, has supplied a reason to measure again. Then run the
bounded verifier-backed scorecard under the normal one-core, 1 GiB,
45-second safeguards. Retain a candidate only if unlike applications expose a
new repeated concrete leaf; a dispatcher-parent split never authorizes a VM,
compiler, bridge, scheduler, or stdlib specialization.
