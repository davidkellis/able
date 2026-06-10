# Bytecode Cold-Process Profile Pair

This tranche separates cold-process startup from warmed VM execution before
selecting another performance change. It keeps no runtime, compiler, stdlib,
or benchmark-source code change.

## Method

- Used the rebuilt `able-v12-base` image, mounted current v12 benchmark
  sources and sibling benchmark inputs, and ran one normal bytecode process
  per workload. This includes the same fresh-container boundary represented by
  the Channel-Rollup Docker publication.
- Each capture used `ABLE_GO_CPU_PROFILE`, a fresh `TMPDIR`/`GOCACHE`, and
  the canonical stdlib already in the base image. Channel-Rollup used
  `ABLE_EXECUTOR=goroutine`; Word-Frequency and Lexical-Rollup were serial.
- Captures are retained under `v12/interpreters/go/.profiles/` as
  `20260710_{channel_rollup,word_frequency,lexical_rollup}_cold_docker_bytecode.cpu.pprof`.
- Verified output: Channel-Rollup and Lexical-Rollup each printed
  `16384:4828:502100`; Word-Frequency printed `1937:11878177`.

## Profile evidence

| Workload | Process duration | CPU samples | Material attributed work |
| --- | ---: | ---: | --- |
| Channel-Rollup | 4.64s | 690ms | loader 21.74% cumulative; parser 17.39%; async task path 21.74%; no material shared channel leaf |
| Word-Frequency | 5.53s | 1.43s | VM execution 76.22%; loader 11.89%; parser 9.09%; HashMap find 4.90% flat, unique to this map workload |
| Lexical-Rollup guard | 4.67s | 460ms | loader 28.26%; parser 26.09%; tree-sitter parse 15.22%; generator/GC work distinct from the target misses |

Loader and tree-sitter parsing recur, but their samples are much smaller than
the observed process wall. The parent CPU profiles do not include the child
`go build -buildmode=plugin` work used by the extern-host boundary.

## Cold extern-host cache experiment

Before the retained prewarm work, `buildExternModule(...)` wrote and compiled
a host plugin beneath `os.TempDir()`. The old interpreter-instance salt was
also stable for the first ordinary CLI process, so a cache could be reused by
a later normal process on the same filesystem; a Docker container started
with an empty filesystem cache. Two runs in one fresh container gave:

| Workload | First process | Cached second process | Difference |
| --- | ---: | ---: | ---: |
| Channel-Rollup | 4.54s | 0.39s | 4.15s |
| Word-Frequency | 5.38s | 1.33s | 4.05s |

Both workloads import the ordinary `able.fs` extern module, but the observed
cost is not `fs`-specific: it is the generic externally hosted Go-plugin build
performed for every uncached extern state. The same cache behavior explains
why host process rows and fresh Docker rows cannot be compared as VM-only
measurements.

## Decision

Do not optimize `runResumable`, channel execution, HashMap lookup, parser
parents, or any benchmark program from these profiles. The profile pair instead
identifies one eligible broad design direction: persistent, ABI-compatible
extern-host plugin caching and generic image/setup prewarming. It must be
designed for all canonical extern states, with safe invalidation when the Able
binary or extern state changes; do not prewarm only `fs` or add a
Channel-Rollup-specific image rule.

## Next recommendation

Completed by the retained generic cache/prewarm implementation documented in
`2026-07-10-extern-host-cache-prewarm.md`. Its next measurement must classify
fresh-container and warmed-process rows separately before selecting another
runtime optimization.
