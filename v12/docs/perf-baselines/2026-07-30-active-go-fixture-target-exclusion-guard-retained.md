# Active Go fixture target-exclusion guard retained

Date: 2026-07-30

## Decision

Retain a general fixture-inventory guard that rejects unclassified exclusions
of an active execution target.

The previous correctness tranche found a stale `skipTargets: ["go"]` entry
that silently prevented a canonical v12 diagnostic fixture from reaching the
only active Go fixture harness. The correction restored that fixture, but it
did not prevent the same metadata error from recurring elsewhere.

The retained guard makes active-target exclusions fail closed. It is based on
target lifecycle and explicit policy, not fixture names, language features,
benchmark applications, nominal types, or container identities.

## Policy

`v12/fixture-target-exclusion-policy.json` currently classifies:

- `go` as an active target;
- `ts` as a retired target; and
- four permitted active-exclusion classifications:
  `intentional-contract-exclusion`, `known-correctness-gap`,
  `platform-limitation`, and `toolchain-limitation`.

An active-target exclusion must have one exact policy entry with its fixture
path, target, permitted classification, and non-empty reason. Retired target
metadata remains readable historical context and needs no active allowlist
entry.

The validator also rejects:

- malformed or duplicate manifest targets;
- targets absent from both lifecycle sets;
- duplicate policy entries;
- active/retired lifecycle overlap;
- invalid or incomplete classifications; and
- stale allowlist entries whose corresponding manifest exclusion no longer
  exists.

The current repository inventory contains 176 fixture manifests, zero active
target exclusions, one retired `ts` exclusion, and zero allowlist entries.

## Implementation

- `v12/scripts/check-fixture-target-exclusions.mjs` performs the recursive,
  deterministic inventory and exposes the validator for tests.
- `v12/scripts/check-fixture-target-exclusions.test.mjs` contains seven
  positive and negative policy tests.
- `v12/run_all_tests.sh` runs both the repository check and the policy tests
  in its preflight section, before expensive Go package execution.

The retained SHA-256 identities are:

- policy:
  `64a4f14752249161a9e03233614dcd14775a660fa1f10bcef69c58af1c075d95`;
- validator:
  `95c2d2b93dedbee3d08c12fae6c6c9aff2dbeb2f8fc65f39b31c39c153a773f1`;
- tests:
  `82f5a55c82b8a9edac2d7e2a891bfcdc21b5fbfd62f148a85f0338260060dcbd`;
  and
- integrated runner:
  `61d4e549f40ba0a3a462849c2dea5e05492b732a531da2890b62d45f5f3011d9`.

All changed source files remain below 1,000 lines.

## Verification

Focused validation passes:

- Node syntax checks for the validator and its tests;
- JSON parsing for the policy;
- the live 176-manifest inventory;
- all seven policy tests in 96.863 milliseconds;
- shell syntax validation for the integrated runner; and
- whitespace validation for the runner change.

The default `./run_all_tests.sh` v12 suite also passes:

- execution-coverage and the new target-exclusion preflights;
- the 130-row scoreboard with five Able and reference samples per row;
- feature/application, selection, threshold, cleanup, and kernel gates;
- standalone parser binding;
- all non-compiler Go packages;
- all 34 bounded compiler batches; and
- the complete bytecode fixture pass in 184.367 seconds.

The longest non-compiler package aggregation was the interpreter at 201.876
seconds. The longest compiler batch was 215.248 seconds. These are cumulative
package/batch durations rather than individual test durations.

The independent evidence gates pass with 130 rows, zero actionable frontier
groups, 23 current closures, zero invalidations, and an empty selector. This
test-inventory policy is outside compiler, runtime, shared-interpreter,
canonical-stdlib, specification, and benchmark performance scopes, so no
measurement refresh is required.

## Scope and cleanup

No compiler, generated runtime, runtime package, interpreter, bytecode VM,
parser, canonical stdlib, benchmark, language, dependency, or WASM source
changed. No performance implementation was admitted.

Local `HEAD`, `origin/master`, and remote `master` remain
`9c32f2777536da2c948327720acc75187973a6d9` with zero divergence. The index
is empty. Nothing was staged, committed, or pushed, and the unchanged
34-path deferred WASM boundary was not touched.

All test state used a disk-backed `/var/tmp` workspace. Its exact size was
2,440,504 KiB; it was removed after verification, and no task-owned cache
remains.

The machine-readable companion is
`2026-07-30-active-go-fixture-target-exclusion-guard-retained.json`.

## Next

Audit active non-WASM conditional Go test skips and harness opt-outs, then
retain a further guard only if a normal semantic test can be silently disabled
without an explicit classification.

Why: manifest exclusions are now fail closed, but `t.Skip`, environment
switches, short-mode branches, and toolchain availability checks are separate
coverage mechanisms.

What it entails: inventory those mechanisms, distinguish infrastructure-only
controls from semantic coverage opt-outs, confirm the normal v12 runner's
behavior, and add policy only for a reproducible unclassified semantic blind
spot. Do not turn legitimate bounded-run or unavailable-toolchain controls
into failures.

Why it matters: native-lowering correctness depends on broad semantic guards
remaining active. Closing the remaining skip mechanisms prevents performance
work from appearing green because coverage disappeared rather than because
the generated program is correct.
