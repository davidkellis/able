# Bytecode unranked-coverage admission search

Date: 2026-07-28

## Decision

Retain no production change.

The reviewed selection contains 63 compiled rows and 56 bytecode rows. The
seven compiled-selected applications without a selected bytecode row divide
into:

- five current 90-second Able bytecode timeouts: Binary Trees, N-Body,
  Quicksort, Sudoku Masks, and TapeLang Alphabet;
- two verifier-passing but unranked rows: Fib and Matrix Multiply, both with
  `python_reference_unavailable` in the current scorecard.

Bounded profiles across the five unlike timeout applications do not identify
an open compiler, runtime, or VM owner that repeats broadly enough to admit an
implementation experiment. The common exact VM leaves belong to already
closed stack-carrier, inline-return, frame, raw-integer, or Array routes.
Main-only instruction censuses show the relevant current inline and Array
fast paths hitting with negligible or zero fallback. Reopening those routes
would not satisfy the current three-unlike-application admission gate.

## Inputs and method

The source and verifier identities, current scorecard status, and reference
timings are preserved in
`2026-07-28-bytecode-unranked-coverage-status.tsv`.

All Able runs used the current `5bbe9e4594bfb2b7e7f8c84b9fbed9314c0ee8db`
source state, the current canonical stdlib source, source-root-only package
resolution, and a freshly built bytecode CLI. Its SHA-256 is
`9d3559c1295b110ed95834f35cf08a040decbf4c2ea212370e47433a3becde4f`.
The disposable build and profile workspace was disk-backed under `/var/tmp`.

For each timeout application:

1. run the current standard workload under the catalog executor;
2. collect a graceful 15-second Go CPU profile with `GOMAXPROCS=1`,
   `GOMEMLIMIT=1GiB`, and `GOGC=50`;
3. separately collect a five-second, main-only bytecode instruction census;
4. preserve the raw artifact identities and compact profile tops, then delete
   the raw disposable workspace after verification.

The 15-second exits are deliberate cutoffs, not new full-run measurements.
The current scorecard's 90-second runs remain the authority for timeout
classification.

Fib and Matrix Multiply were each executed five times against their public
Ruby-suite verifier. These are fresh Able measurements only. Existing Ruby
scorecard timings are useful descriptive context, but they were not refreshed
as paired A/B references in this tranche.

## Fresh verifier-backed runnable rows

| Application | Able bytecode runs (s) | Mean (s) | Verification | Current ranking blocker |
| --- | --- | ---: | --- | --- |
| Fib | 0.15, 0.16, 0.19, 0.16, 0.12 | 0.1560 | 5/5 pass, stable output | Python reference unavailable |
| Matrix Multiply | 4.75, 4.68, 4.48, 4.56, 4.48 | 4.5900 | 5/5 pass, stable output | Python reference unavailable |

For context only, dividing these fresh Able means by the existing scorecard
Ruby timings gives 0.00318 for Fib and 0.09444 for Matrix Multiply. These are
not paired current-reference ratios and do not change selection.

## Bounded timeout profiles

| Application | Samples (s) | Distinct current owner shape |
| --- | ---: | --- |
| Binary Trees | 14.90 | allocation/GC, named-struct construction, typed patterns, recursive returns |
| N-Body | 14.90 | float simulation calls, members, stack values, and returns |
| Quicksort | 14.48 | input parsing at cutoff; raw integers, casts, and Array reads |
| Sudoku Masks | 14.92 | calls, raw bitwise operations, frames, and Array reads |
| TapeLang Alphabet | 14.92 | member dispatch, named-struct plans, and Array slots |

The retained top-40 reports are under
`2026-07-28-bytecode-timeout-admission-profiles/`. Raw CPU and stats artifact
hashes are in
`2026-07-28-bytecode-timeout-admission-profile-manifest.tsv`.

No exact open leaf repeats across all five applications. The recurring exact
leaves reconcile as follows:

- `appendSlotStackValueChecked`: visible in all five profiles; the checked
  stack-carrier route is already closed;
- `finishInlineReturn`: visible in four retained tops; the inline-return
  family is already closed;
- `popCallFrameFields`: visible in three retained tops; frame restore is
  already closed;
- `bytecodeRawIntegerValueInfo`: visible in four retained tops; raw-integer
  carrier work is already closed;
- Array index/read work: material only in application subsets and already
  covered by the closed Array route.

## Fast-path census

| Application | Instructions | Primitive eligible | Inline hit/miss | Array fast/fallback | Generic call/member fallbacks |
| --- | ---: | ---: | ---: | ---: | ---: |
| Binary Trees | 23,821,542 | 20.00% | 4,787,650 / 0 | 0 / 0 | 0 / 0 |
| N-Body | 21,026,669 | 52.99% | 479,791 / 0 | 811,371 / 40 | 0 / 0 |
| Quicksort | 34,081,490 | 92.49% | 4 / 0 | 184,176 / 2 | 0 / 0 |
| Sudoku Masks | 28,400,081 | 67.60% | 572,879 / 0 | 119,648 / 6 | 0 / 0 |
| TapeLang Alphabet | 22,647,597 | 58.80% | 763,053 / 0 | 2,341,009 / 27 | 0 / 0 |

The detailed instruction mix is preserved in
`2026-07-28-bytecode-timeout-admission-stats-summary.tsv`. These censuses
confirm that the current inline-resolution and Array-specialized paths are
being selected. They do not expose a missed shared fallback that could admit
a general VM change.

## Scope and retained state

- No compiler, generated runtime, runtime, interpreter, bytecode VM,
  canonical stdlib, benchmark, language, dependency, nominal-special-case,
  or WASM production source changed.
- The 63 compiled-selected applications and 56 bytecode-selected applications
  remain unchanged.
- All 63 strict compiled graphs remain interpreter-free.
- Deferred dirty WASM paths were preserved without inspection or mutation.

## Evidence

- Coverage status:
  `2026-07-28-bytecode-unranked-coverage-status.tsv`
  (`9384caa267f51fd81e65e243ae13ecfa19efff3f8ef1337b5574a94ecb880a29`)
- Profile manifest:
  `2026-07-28-bytecode-timeout-admission-profile-manifest.tsv`
  (`458641124ed24131720dd95bdfe43727604d77054c5d44f0f6bb5029b70c007b`)
- Stats summary:
  `2026-07-28-bytecode-timeout-admission-stats-summary.tsv`
  (`f0b9a106644be4181a4524aedbbe08be087db35096cf6c30a5ac8dc2a6ca3adb`)
- Verifier runs:
  `2026-07-28-bytecode-unranked-verifier-runs.tsv`
  (`1f4ae1ddf7cd08926f0fc8a34fd9625e0e6d0e3b1115787044e735c2f9f12444`)
- Selection manifest:
  `v12/bench-selection-manifest.json`
  (`dac2450e10f73655271c2e03b236e3d2c0b4dfe83e8bdfcfda9bee4efdba9d23`)
- Current scorecard:
  `external-scoreboard-current.json`
  (`9df24ffed73a2ad39060eb2229f7588d617426fa456c3b8df3ff045d5392b53c`)
- Closure ledger:
  `performance-evidence-closure-ledger.json`
  (`f00e9fafd2ad222c018de0748093957687bfd555f3bceb711825ef23191c3ffc`)

## Next recommendation

Calibrate a portable, mode-aware workload-scale contract for the seven
unranked bytecode rows before changing the compiler or VM.

Why: standard workloads currently yield either 90-second Able timeouts or a
missing usable Python measurement, so they cannot produce repeated,
comparable Able/Python/Ruby ratios. The profiles do not justify another
production optimization experiment.

What it entails: use temporary benchmark copies to choose one smaller scale
per application that preserves the same algorithm, language features, input
shape, and public output verifier across Able, Python, and Ruby. Require every
language to complete comfortably under the timing ceiling, then collect five
or more balanced verifier-backed runs. Update the external benchmark suite
and selection only if a common portable contract is demonstrated.

Why it is important: this converts the catalog's largest bytecode evidence
gap into measurable cross-language data. Those full-run profiles and ratios
can either reveal a genuinely shared general owner or establish that the
remaining difference is workload-specific without weakening semantics.
