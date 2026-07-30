# Default parser fixture-corpus lane retained

Date: 2026-07-30

## Decision

Retain an explicit full-mode parser package lane in the default v12 handoff
runner.

The default runner previously included `pkg/parser` in its broad
`go test -short` package invocation. All 16 parser fixture-family tests call
one deliberate short-mode guard, so the package returned success after
skipping the entire source-to-AST fixture corpus. The corpus had been restored
to ordinary non-short Go testing on 2026-07-14, but the routine handoff did not
exercise it.

The runner now executes `./pkg/parser` once without `-short` and with a
one-minute package bound, then excludes that package from the remaining
non-compiler short-mode set. This is a package-level scheduling rule, not a
fixture-, feature-, benchmark-, nominal-type-, or container-specific rule.

## Conditional-skip inventory

The audit covered 617 active non-WASM v12 Go test files and found 140 explicit
`Skip`, `Skipf`, or `SkipNow` calls. Every call has an intentional category:

| Category | Calls | Disposition |
| --- | ---: | --- |
| short mode | 88 | 87 compiler guards use the explicit non-short `--compiler` release lane; the one parser guard is now covered by the default full-mode parser lane |
| Go toolchain unavailable | 23 | legitimate toolchain-availability guard |
| canonical stdlib unavailable | 3 | legitimate external-source availability guard |
| benchmark or diagnostic probe selection | 4 | explicit opt-in measurement/debug control |
| compiled CLI integration | 1 | explicit non-short `--compiled-cli` release lane |
| temporary-filesystem layout | 1 | platform/setup validity guard |
| fixture/audit subset selection | 20 | explicit empty, diagnostic-only, target, or batched-subset control |

There are zero unclassified explicit skip calls. The short-mode topology is
exactly 87 compiler guards, one parser guard, and zero elsewhere.

Compiler short-mode semantics remain unchanged: the default handoff uses
bounded short shards, while `run_all_tests.sh --compiler` supplies the
separate non-short release matrix. Generated-Go CLI integration remains under
`--compiled-cli`, which sets `ABLE_RUN_COMPILED_CLI_INTEGRATION=1`.
Benchmark/profiling probes and unavailable-toolchain/stdlib controls remain
opt-in or conditional by design.

## Reproduction

Before the runner correction, an exact JSON-event run of the parser fixture
families under `-short` reported:

- 16 top-level fixture families skipped;
- zero fixture families passed; and
- a green package result in 0.003 seconds.

The same exact family selection without `-short` reported:

- all 16 fixture families passed;
- zero fixture families skipped;
- all 176 source fixture subtests passed; and
- a green package result in 0.125 seconds.

The focused final runner-lane invocation passed all 16 families and all 176
subtests with zero skips in 0.142 seconds. The integrated default handoff lane
passed the complete parser package in 0.149 seconds.

## Implementation and verification

Only `v12/run_all_tests.sh` changed. Its retained SHA-256 is
`d81d9cbc49fe661bf50c5b7662cb76609923a2853a1edda09378382c11036e6f`.
The file has 416 lines.

Focused checks pass:

- shell syntax and whitespace validation;
- exact short-mode skip reproduction;
- exact non-short family and 176-subtest execution;
- the live 140-call classification with zero unknowns;
- exact short-mode topology; and
- the focused full parser package under the one-minute bound.

The default `./run_all_tests.sh` suite passes:

- every evidence, selection, coverage, threshold, cleanup, and kernel
  preflight;
- standalone parser Go binding;
- the new full parser package lane;
- all remaining non-compiler packages;
- all 34 compiler short-mode batches; and
- the complete bytecode fixture pass in 186.280 seconds.

The full parser package is no longer duplicated in the remaining short-mode
package set. The longest non-compiler aggregate was the interpreter at
200.105 seconds, and the longest compiler batch was 223.708 seconds. These are
cumulative package/batch durations, not individual-test durations.

The independent evidence gates remain green with 130 fully sampled rows, zero
actionable frontier groups, 23 current closures, zero invalidations, and an
empty selector. Test-runner scheduling is outside every checked production,
semantic-source, canonical-stdlib, specification, and benchmark scope, so no
performance refresh is required.

## Scope and cleanup

No parser implementation, AST, compiler, generated runtime, runtime package,
interpreter, bytecode VM, canonical stdlib, benchmark, language, dependency,
fixture, or WASM source changed.

Local `HEAD` and `origin/master` remain
`9c32f2777536da2c948327720acc75187973a6d9` with zero divergence. The index is
empty. Nothing was staged, committed, or pushed, and the unchanged 34-path
deferred WASM boundary was not touched.

All Go and Node state used a disk-backed `/var/tmp` workspace. Its exact size
was 2,440,980 KiB. The separate 20,483-byte raw skip inventory was also
removed. No task-owned cache remains.

The machine-readable companion is
`2026-07-30-default-parser-fixture-corpus-lane-retained.json`.

## Next

Audit silent non-WASM Go test suppression that does not emit an explicit skip,
especially environment/configuration early returns and build-constrained test
files.

Why: this tranche closed and classified explicit skip events, but a test can
still return successfully before assertions or be excluded from the compiled
test binary by a build tag.

What it entails: inventory conditional early returns, active test build
constraints, runner tags, and zero-test package results; distinguish child
process helpers and intentionally disabled experiments from normal semantic
coverage; reproduce one real blind spot before retaining any change. Keep
JS/WASM-only constraints outside the active workstream.

Why it matters: the handoff must prove that semantic tests executed, not only
that package processes exited successfully. That evidence protects native
lowering and interpreter correctness from silent coverage loss.
