# Generic Extern-Host Cache Prewarm

## Decision

Retain the generic Go extern-host cache prewarm. It removes the repeated
uncached `go build -buildmode=plugin` cold-start wall identified in the paired
profiles without changing the interpreter's execution semantics, any Able
benchmark source, or a named stdlib module.

## Design

- `ABLE_EXTERN_CACHE_DIR` selects a persistent cache root; the existing
  temporary-directory cache remains the default when it is unset.
- A host-module key now includes the complete rendered extern state, cache
  format version, Go runtime version, OS/architecture, `GOEXPERIMENT`, and
  `GOFLAGS`. It no longer includes an interpreter-instance session number, so
  an identical state is shareable by independent interpreter instances.
- If `plugin.Open` rejects a stale or corrupt artifact, the runtime builds a
  uniquely named replacement and, when the cache remains writable, atomically
  records it as the selected artifact. The successful replacement remains
  usable even if that optional cross-process marker write fails.
- `able cache prewarm` discovers every visible canonical Able package through
  the normal loader naming rules, registers Go extern declarations only, and
  compiles their host modules without evaluating Able declarations or calling
  host functions. `able setup --prewarm-extern` exposes the same explicit
  installation path.
- The benchmark base image sets
  `ABLE_EXTERN_CACHE_DIR=/able/v12/cache/extern` and runs that generic command
  while building. Its current canonical stdlib/kernel source discovers 59
  packages and prewarms 10 Go host modules.

No `able-stdlib` source change was needed: prewarming is driven by all
discoverable declarations, not an allow-list such as `able.fs`.

## Fresh-container guard

The profiled Able bytecode images were rebuilt from the prewarmed base and run
as new containers three times. Each retained its established output.

| Workload | Prior uncached cold process | Prewarmed fresh container | Effect |
| --- | ---: | ---: | --- |
| Channel-Rollup | 4.462s | 0.408s average | 10.92x faster process wall |
| Word-Frequency | 5.38s | 1.527s average | 3.52x faster process wall |
| Lexical-Rollup guard | 4.67s | 0.428s average | 10.90x faster process wall |

The Channel-Rollup and Lexical-Rollup runs printed `16384:4828:502100`; the
Word-Frequency run printed `1937:11878177`.

`fib` is the no-extern control. With the persistent cache it printed
`1134903170` in 113.722ms; with `ABLE_EXTERN_CACHE_DIR` redirected to a new
empty directory it printed the same value in 98.706ms and created zero cache
artifacts. The cache is therefore absent from programs that do not use Go
extern hosting.

## Verification

- Focused driver/interpreter/CLI tests pass, including canonical package
  discovery, prewarming without evaluation, cache-root selection, state-hash
  invalidation, and a subprocess guard that corrupts a plugin then verifies
  stale-artifact rebuild and marker selection.
- `able cache prewarm` succeeds inside the rebuilt `able-v12-base` image;
  a subsequent container finds all 10 `.so` artifacts and repeats the command
  successfully.
- The base and four bytecode benchmark images used for the guard were rebuilt
  from the current source.

## Next recommendation

Refresh the full external scorecard with both fresh-container and warmed
process classifications, then profile only the residual bytecode execution
wall. Why: this change removes a roughly four-second packaging cost and makes
the remaining interpreter performance signal visible; mixing it with prior
uncached Docker rows would misstate progress toward the Python/Ruby target.
The work entails rebuilding the representative Able images, collecting
multi-run bytecode/compiled rows with cache state recorded, preserving a
no-extern control, and selecting a runtime candidate only when one concrete
leaf repeats across independent programs.
