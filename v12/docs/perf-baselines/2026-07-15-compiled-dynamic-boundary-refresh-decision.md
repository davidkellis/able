# Compiled dynamic-boundary refresh — 2026-07-15

## Decision

Keep no compiler, generated-runtime, VM, canonical-stdlib, or benchmark-source
change. The first full refresh after the portable catalog grew from 22 to 32
applications identifies repeated dynamic-boundary event categories, but no new
shared timed leaf eligible for an optimization.

## Method

`just bench-compiled-boundary-audit -- --suite coverage --timeout 45` generated
debug-only telemetry binaries for every catalogued portable application. The
normal sibling Ruby verifier checked each completed run. `GOMEMLIMIT=1GiB`,
`GOGC=50`, and `GOMAXPROCS=1` bounded each serial process. The telemetry build
uses atomic counter increments, so its elapsed times are not performance data.

The durable row report is
`2026-07-15-compiled-dynamic-boundary-reachability.{json,md}`. Thirty-one of
32 applications completed with verified output; Sudoku reached the existing
45-second process cap and is status-only.

| Counter | Verified-run total |
| --- | ---: |
| Explicit dynamic call | 6,830 |
| Residual polymorphic call | 5,911 |
| Host/ABI call | 2,813 |
| Runtime service call | 5,116 |

## Attribution

The two historic output-heavy rows remain exactly the already-rejected static
print boundary: I-Before-E records 1,629 explicit/residual and 1,628 host
events, while PiDigits records 1,001/1,000. The generic primitive/String print
candidate previously regressed unrelated compiled guards, so these counts do
not reopen it.

The new high counts are concurrency-shaped but are not one common lower cost:

| Application | Explicit | Residual | Service |
| --- | ---: | ---: | ---: |
| Future Pipeline | 1,155 | 1 | 7 |
| Future Await Race | 1,312 | 1 | 482 |
| Await Channel Mux | 1,538 | 1,025 | 2,560 |
| Mutex Await Journal | 1 | 2,049 | 2,052 |

The categories mark entries to `__able_call_named`, `__able_call_value`,
`__able_spawn`, and `__able_await`; they do not identify an inner CPU leaf or
prove equal work below those helpers. The existing three-application generated
main profiles already identify the only repeated inner async leaf as
`bridge.currentGID` / `runtime.Stack`, and the fixed execution-context ABI that
would remove it failed independent N-body and K-Nucleotide guards. The other
new rows diverge between named-call, callable-value, and scheduler-service
paths. A counter-only refresh cannot justify a scheduler, mutex, channel,
Future, closure, or application-specific rewrite.

## Harness maintenance

`bench_compiled_boundary_audit` now describes the catalog rather than a stale
hard-coded application count and accepts the conventional standalone `--`
separator forwarded by variadic `just` recipes. Normal builds remain telemetry
free.

## Next recommendation

Do not profile or modify this unchanged bridge family again. Reopen compiler or
VM performance selection only after a real cross-cutting source change exposes
a concrete descendant shared by at least three unlike verifier-backed
applications; use the full catalog as its regression gate.
