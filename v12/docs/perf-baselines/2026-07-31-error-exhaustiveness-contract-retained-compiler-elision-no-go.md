# Error exhaustiveness contract retained; compiler elision rejected

Date: 2026-07-31

## Decision

**Retain option A as the v12 language contract and retain conservative,
positive-only checker facts. Reject the compiler branch-elision candidate.**

The canonical rule is general to open interfaces:

- only unguarded clauses contribute static coverage;
- a binding and discard typed as the whole interface have identical coverage;
- enumerating visible concrete implementations never closes an open
  interface;
- an unproven ordinary match retains `Non-exhaustive match`; and
- an unmatched partial rescue re-propagates the original raised value.

The checker records only proven exhaustive facts in side tables keyed by the
canonical AST. Unknown types, guarded clauses, refutable Array patterns, and
unproven nested patterns produce no fact. No AST, parser, interpreter,
bytecode, runtime, stdlib, dependency, or WASM change was required.

## Cross-runtime contract

The new `08_01_error_exhaustiveness_open_set` fixture covers:

- `!i32` with success and whole-`Error` typed bindings;
- an open `Error` match with a concrete arm and whole-`Error` catch-all;
- a partial inner rescue whose unmatched error reaches an outer whole-`Error`
  handler; and
- a guarded binder whose unguarded typed `i32` arm closes the subject domain.

Tree-walker, bytecode, parity, and strict fallback-free compiled execution
produce the same output. Focused checker guards prove wildcard, binding,
nullable, `Result`, open-interface, rescue, exact-struct, guarded, concrete,
and Array cases.

## Compiler candidate and reach

The evaluated compiler candidate consumed only positive checker facts. It
omitted the generated non-exhaustive-match guard and unmatched-rescue
propagation branch when the fact was present. Missing facts retained the
existing dynamic branch.

The rule reached all three unlike applications:

| Application | Guard count | Generated Go bytes | Binary bytes |
| --- | ---: | ---: | ---: |
| Policy Record Dispatch | 350 → 206 (-41.14%) | 8,966,340 → 8,900,709 (-0.732%) | 22,125,200 → 22,065,120 (-0.272%) |
| Sensor Calibration | 216 → 165 (-23.61%) | 4,699,245 → 4,676,212 (-0.490%) | 13,867,216 → 13,849,392 (-0.129%) |
| Versioned Telemetry Pipeline | 171 → 155 (-9.36%) | 4,435,971 → 4,428,723 (-0.163%) | 12,959,112 → 12,952,320 (-0.052%) |

This proves broad generated-code reach, but code-size reach is not the runtime
retention bar.

## Repeated A/B

Each application and variant ran ten verifier-approved processes in two
balanced five-process cohorts: forward order ran baseline then candidate;
reverse order ran candidate then baseline. Runs used CPU 12,
`GOMAXPROCS=1`, a 58-second timeout, existing strict binaries built with
`--no-fallbacks`, and public output verification. All 60 Able processes passed
with no failures or timeouts.

| Application | Baseline wall | Candidate wall | Wall delta | Baseline user | Candidate user | User delta |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Policy Record Dispatch | 0.073 s | 0.067 s | -8.22% | 0.051 s | 0.052 s | +1.96% |
| Sensor Calibration | 0.030 s | 0.031 s | +3.33% | 0.016 s | 0.016 s | 0.00% |
| Versioned Telemetry Pipeline | 2.043 s | 2.076 s | +1.62% | 2.009 s | 2.041 s | +1.59% |

Policy’s wall improvement conflicts with its slightly worse user time and is
at launch-scale duration. Sensor is neutral to slightly worse. Telemetry is
the sustained row and is slightly worse, while its forward/reverse cohort
spread is larger than the candidate delta. The candidate therefore does not
show one broad, repeatable runtime improvement.

Fresh ten-process Go references passed public verification:

| Application | Go wall mean | Candidate/Go |
| --- | ---: | ---: |
| Policy Record Dispatch | 0.0051 s | 13.14× |
| Sensor Calibration | 0.0060 s | 5.17× |
| Versioned Telemetry Pipeline | 0.1929 s | 10.76× |

The candidate does not materially close the 95%-of-Go objective.

## Retention consequence

The generated-code optimization was removed in full. Compiled matches and
rescues continue to emit their dynamic fallback/re-propagation branches, so
compiler-production returns to its reviewed fingerprint. The checker facts
remain a reusable semantic result, but future lowering may consume them only
after a distinct general candidate clears the same cross-application bar.

This is deliberately stricter than retaining an optimization for small binary
reductions: the project requires verifier-backed repeated runtime evidence
across unlike programs.

## Evidence reconciliation

Before withdrawal, the selector reported all 23 closures invalidated by the
intentional v12-spec change and 12 compiled closures additionally invalidated
by compiler-production. After withdrawal, the compiler-production
invalidations disappeared; only the reviewed v12-spec semantic change
remained.

The 23 current dispositions remain applicable because:

- ordinary partial-match and partial-rescue runtime behavior is unchanged;
- tree-walker and bytecode execution code is unchanged;
- compiled generated behavior is unchanged after candidate withdrawal;
- benchmark sources and reference implementations are unchanged; and
- the candidate itself received fresh, balanced, verifier-backed rejection
  evidence.

Only the reviewed v12-spec scope is rebased. No closed performance result is
reclassified and no new performance owner is admitted.

## Verification

Focused checks passed for the complete typechecker package, the new fixture in
tree-walker and bytecode modes, tree-walker/bytecode parity, and strict
fallback-free compiled execution.

The complete `./run_all_tests.sh` gate passed in 585.91 seconds at
1,878,412 KiB peak RSS:

- 274 seeded exec fixtures, zero planned fixtures, and 275 fixture
  directories;
- 132 current scorecard rows and zero actionable frontier groups;
- all 844 compiler tests;
- canonical Result-normalization compiler outlier: 14.089 seconds;
- noisiest bounded compiler batch: 37.400 seconds;
- every tree-walker, bytecode, and parity shard; and
- 23 current closures, zero invalidations, and a current five-node
  architecture chain.

All task builds and measurements used disk-backed `/var/tmp`.

## Next

Use the mode-aware selector as the admission gate. It currently returns
nothing, so do not begin another performance mutation until a checked
production, semantic, stdlib, benchmark, or scorecard change invalidates a
closure.

Why: current compiled and bytecode owners already have reviewed dispositions,
and this candidate supplied a fresh no-go rather than a new shared hot owner.

What it entails: keep the evidence and release gates green during ordinary
correctness work. When a real invalidation appears, refresh only its named
closures and require a new three-unlike-application, verifier-backed A/B before
retaining production performance code.

Why it matters: native Go lowering remains the performance goal, but work must
follow measured shared runtime ownership rather than code-size reach or an
already rejected cold branch.
