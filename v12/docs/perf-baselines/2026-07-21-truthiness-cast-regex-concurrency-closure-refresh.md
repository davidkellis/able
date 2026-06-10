# Truthiness/cast regex-concurrency closure refresh

Date: 2026-07-21

## Decision

The `compiled-regex` and `bytecode-concurrency` closures are current against
the post-truthiness/cast shared interpreter semantics. Keep no compiler, VM,
runtime, canonical-stdlib, benchmark, verifier, reference, language, or WASM
change from this tranche.

Fresh repeated timing confirms large product gaps but does not admit a generic
candidate. Only compiled Policy Record Dispatch enters the generated
truthiness bridge, 2,048 times; the other five compiled programs have zero
truthiness/cast reach and all six have zero explicit casts. The eleven
bytecode concurrency programs make 0-541,058 primitive truthiness checks per
main but never reach the changed Error fallback or explicit-cast boundary.

## Frozen contract

- The v12 spec SHA-256 remains
  `4f0405b86c122993723e8617abd6f825d9a8ff858d4c72acaf4e33469452f080`.
- The 70-file canonical external Able stdlib source-tree SHA-256 remains
  `43ff2e68e59c8be7fb1024c86a1f61a0eea84596279b4f0e146511d66c5308d8`.
- Executables were built before timing. Every process passed its public
  verifier and used its catalog directory, arguments, executor policy, CPU
  budget, and a 55-second application-process cap.
- Arithmetic means retain every successful sample. Because short first lanes
  were volatile, every row received a matched second cohort; no application or
  sample was selectively removed.
- Able executions use the catalog's bounded memory policy. Compiled rows are
  serial on CPU 0; concurrency rows use the goroutine executor on CPU 0.

## Repeated timing

| Application | Mode | Able samples | Able mean | Limiting reference | Reference samples | Reference mean | Ratio |
| --- | --- | ---: | ---: | --- | ---: | ---: | ---: |
| Config Validation Extraction | compiled | 10 | 0.120 s | Go | 10 | 0.004363 s | 27.502x |
| Log Routing Redaction | compiled | 10 | 0.122 s | Go | 10 | 0.005458 s | 22.353x |
| Policy Record Dispatch | compiled | 10 | 0.198 s | Go | 10 | 0.005337 s | 37.099x |
| Regex Set Audit | compiled | 10 | 0.116 s | Go | 10 | 0.006088 s | 19.053x |
| Regex Stream Audit | compiled | 10 | 0.119 s | Go | 10 | 0.005109 s | 23.294x |
| Regex Suffix Audit | compiled | 10 | 0.124 s | Go | 10 | 0.005514 s | 22.487x |
| Await Channel Mux | bytecode | 10 | 0.240 s | Ruby | 10 | 0.102693 s | 2.337x |
| Channel Rollup | bytecode | 10 | 0.430 s | Python | 10 | 0.054978 s | 7.821x |
| Concurrent Event Routing | bytecode | 10 | 2.835 s | Python | 10 | 0.032368 s | 87.586x |
| Concurrent Text Index | bytecode | 10 | 0.594 s | Python | 10 | 0.067398 s | 8.813x |
| Dependency Wave Validation | bytecode | 10 | 0.501 s | Python | 10 | 0.034640 s | 14.463x |
| Future Await Race | bytecode | 10 | 0.140 s | Python | 10 | 0.032590 s | 4.296x |
| Future Pipeline | bytecode | 10 | 0.409 s | Python | 10 | 0.064056 s | 6.385x |
| Mutex Await Journal | bytecode | 10 | 0.189 s | Python | 10 | 0.025762 s | 7.336x |
| Mutex Ledger | bytecode | 10 | 0.325 s | Python | 10 | 0.039178 s | 8.295x |
| Mutex Work Queue | bytecode | 10 | 0.323 s | Python | 10 | 0.031110 s | 10.383x |
| Validated Job Pipeline | bytecode | 10 | 0.770 s | Ruby | 10 | 0.058546 s | 13.152x |

All 450 timing processes represented by these decisions verified with zero
failures and zero timeouts: 120 compiled/Go processes and 330
bytecode/Python/Ruby processes. Pooled compiled Able CVs range from 11.12% to
37.06%; pooled bytecode Able CVs range from 3.37% to 25.23%. Short references
remain naturally noisy, with maxima of 30.52% Python and 23.78% Ruby. Every
sample remains in its arithmetic mean, and every classification is far from
the 1.053x ceiling corresponding to the 95%-throughput target.

## Exact reach

Temporary debug counters were placed immediately at the changed semantic
boundaries, used only in untimed processes, and then removed.

| Application group | Mode | Census processes | Primitive/general truthy checks | Changed-path result |
| --- | --- | ---: | ---: | --- |
| Six regex applications | compiled | 6 | Policy: 2,048; other five: 0 | One application reaches generated truthiness; no casts |
| Eleven concurrency applications | bytecode | 22 | 0-541,058 per process | Zero Error fallbacks, casts, or cast failures |

Every census process completed under its guard and passed the public verifier;
both bytecode counts reproduce exactly. Policy's 2,048 bridge entries occur in
only one application, below the mandatory three-unlike-application breadth
rule. The bytecode paths all return through primitive truthiness cases before
the corrected Error matcher. Therefore neither closure passes profile
admission, no CPU profile or candidate was built, and prior exact ownership
evidence remains causal.

The concise census is retained in
`2026-07-21-truthiness-cast-regex-concurrency-closure-reach.json`; raw compiled
telemetry is retained in
`2026-07-21-truthiness-cast-regex-concurrency-closure-compiled-reach.json`.

## Exact timing artifacts

- Initial compiled Able/Go:
  `2026-07-21-truthiness-cast-regex-concurrency-closures-compiled.json`
  (`7736fadea4e918e4b04b4b509121abb3f7fae45cf5398aa9cb9f4c0ce209df3e`),
  `2026-07-21-truthiness-cast-regex-concurrency-closures-go-reference.json`
  (`e48dc746bbaf4dd3ac107d96a7bc6f94ac1c9fbdb9e99b2a264c09536853e33e`).
- Second compiled Able/Go:
  `2026-07-21-truthiness-cast-regex-concurrency-closures-c2-compiled.json`
  (`85f5162c60e1b48e74bb932008be977b76aaa7bdcec62d5530ab895dbf8c2c65`),
  `2026-07-21-truthiness-cast-regex-concurrency-closures-c2-go-reference.json`
  (`443fca0948c3918e42bafc410580dd7ad528547dcdf3fae6a92fd268014026d9`).
- Initial bytecode Able/interpreter references:
  `2026-07-21-truthiness-cast-regex-concurrency-closures-bytecode.json`
  (`562e70ed6b97e85b89413d2294e6f52a4bbe2cdd5c66aaabfc9249268cc9713e`),
  `2026-07-21-truthiness-cast-regex-concurrency-closures-interpreter-reference.json`
  (`9a053ff9a956aa40bdeef3230a897f6d6b025e32ba54595251a6dd2808f4b9f2`).
- Second bytecode Able/interpreter references:
  `2026-07-21-truthiness-cast-regex-concurrency-closures-c2-bytecode.json`
  (`5c8b5dc6eb9bd633762bc7fe8f6438c82f9572d7c0fff83abbccdd411ec01aec`),
  `2026-07-21-truthiness-cast-regex-concurrency-closures-c2-interpreter-reference.json`
  (`24932f01c53a1924bc7d65eb2e04a9389af70768f415e84918f4df034debea67`).
- Compiled reach telemetry:
  `2026-07-21-truthiness-cast-regex-concurrency-closure-compiled-reach.json`
  (`d2f9597ceeb7b540fa6f68522529b01c2f790a0b1d962224d7c76058e7b3c3df`).

## Next recommendation

Refresh `compiled-concurrency` and `bytecode-register-architecture` next.

Why: compiled concurrency is the compiler counterpart to the eleven VM
concurrency rows just measured, so it can separate generated runtime/Future
costs from VM dispatch and environment costs. Bytecode register architecture
is the remaining VM-wide invalidated closure and tests a cross-feature
transport wall rather than another semantic family. Together they continue
mode-specific reconciliation while increasing architectural breadth.

What it entails: reuse the frozen sources and current toolchains; collect five
verified processes per ordinary lane and matched additional cohorts for
volatile workstation rows; run exact changed-path reach before profiling; and
profile only a materially reached concrete leaf. Advance only those two
closures. Build a candidate only if one generic mechanism is material in at
least three unlike applications and preserves current target guards. This is
the right next step because it closes the direct compiler counterpart and the
remaining broad VM architecture question before revisiting narrow residual
closures. No WASM work is involved.
