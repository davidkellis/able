# Compiled Dynamic-Boundary Reachability Decision (2026-07-12)

## Method and result

`just bench-compiled-boundary-audit -- --suite coverage --timeout 10` compiled
every one of the 22 application benchmarks with the new debug-only
`-dynamic-boundary-telemetry` mode. Each run used the normal canonical stdlib,
source-root-only discovery where declared by the catalog, a 10-second process
cap, `GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`. The generated telemetry
binary is not a timing candidate: its counters use atomics, so its elapsed
time is deliberately not compared with the normal compiled baseline.

Eighteen completed runs were verifier-backed. BinaryTrees, Sudoku, Sudoku
Masks, and Fixed Width 128 reached the cap and retain status only. The durable
row data is in
`2026-07-12-compiled-dynamic-boundary-reachability.json`; its Markdown view is
the matching `.md` file.

| Counter | Completed-run total | Interpretation |
| --- | ---: | --- |
| explicit dynamic call | 2,700 | named dynamic-call bridge entries |
| residual polymorphic call | 2,713 | callable-value dispatch entries |
| host/ABI call | 2,690 | host output/extern entries |
| runtime service call | 2 | `spawn`/await service crossings |

These counters are event categories, not mutually exclusive costs. A dynamic
call that reaches a host function increments both applicable categories.

## Attribution

The largest counts identify a repeated, language-level host boundary rather
than a hidden generic collection or value carrier wall:

- I-Before-E records 1,629 explicit/residual events and 1,628 host events.
  Its source calls `print(word)` for each failing word.
- PiDigits records 1,001 explicit/residual events and 1,000 host events. Its
  source prints every completed ten-digit output line.
- The remaining completed applications are at 0–36 events each. K-Nucleotide
  is the only row with a modest residual-call excess (36 residual, 21 host),
  but it is a single map/callback/text program rather than a repeated boundary.
  Channel Rollup is the only run with runtime-service events (two), so it does
  not authorize a scheduler rule.
- Mandelbrot records zero events despite being a material compiled miss. Fib,
  MatrixMultiply, Word Frequency, Document Audit, Lexical Rollup, and Rational
  Series record only their final output or small host setup crossings. This
  confirms the previous static-worker audit: their numerical, collection, and
  iterator execution is not secretly routing through the dynamic carrier.

The repeated I-Before-E/PiDigits event is ordinary observable output. It is
not a benchmark artifact: both programs intentionally emit their result lines,
and any general Able program that prints statically known primitive or String
values takes the same generated host path today. It is also not permission to
special-case either application, text corpus, digit algorithm, or nominal
container.

## Decision

Keep the telemetry mode and corpus harness; no production runtime behavior is
changed. The sweep rejects a global `runtime.Value`, map, iterator, scheduler,
or collection optimization from these counts. The telemetry's normal-build
absence is enforced by `TestCompilerDynamicBoundaryTelemetryIsOptIn`; the full
corpus run removed its generated `.gotmp` workspace after writing the durable
reports.

The evidence authorized, and the following tranche completed, one bounded
profile gate: generic static `print` lowering for statically known,
single primitive/String arguments. Its direct host-output prototype preserved
output but failed the broad compiled guard: it improved I-Before-E and
PiDigits, while materially regressing MatrixMultiply, JSON, and Word
Frequency. The candidate was fully reverted. See
`2026-07-12-compiled-static-builtin-print-gate.md` for the verifier-backed A/B
data and decision.

## Next recommendation

Do not revisit direct `print` lowering without new evidence that explains the
guard regressions. Refresh bounded bytecode profiles for Reverse Complement,
K-Nucleotide, I-Before-E, and Mandelbrot with Base64 and PiDigits controls.
Why: the output boundary has exhausted its generic compiler gate, while those
interpreter applications remain material verified gaps. Accept a follow-on
only when an identical non-nominal VM leaf repeats across independent workers
and remains neutral on the controls.
