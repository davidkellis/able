# Truthiness/cast control-closure refresh

Date: 2026-07-21

## Decision

The `compiled-current-control` and `bytecode-iterator-control` closures are
current against the post-truthiness/cast shared interpreter semantics. Keep no
compiler, VM, runtime, canonical-stdlib, application, or reference change.

The repeated timing cohorts reproduce real target gaps, but the semantic paths
that triggered invalidation are not reached. Current compiled main profiles
remain entirely in direct generated application bodies. Two exact bytecode
censuses per application record no non-primitive/Error truthiness fallback and
no explicit runtime cast. Invalidation therefore does not expose a candidate.

## Frozen contract

- The v12 spec SHA-256 remains
  `4f0405b86c122993723e8617abd6f825d9a8ff858d4c72acaf4e33469452f080`.
- The canonical external stdlib source-tree SHA-256 remains
  `43ff2e68e59c8be7fb1024c86a1f61a0eea84596279b4f0e146511d66c5308d8`.
- Executables were built before timing. Every timed process passed its public
  verifier. Each process used its catalog run directory, arguments, CPU
  budget, and a 55-second cap.
- Ordinary rows retain five independent processes and arithmetic means without
  outlier removal. Matrix Multiply and TapeLang retain order-balanced second
  cohorts because Matrix crossed its prior target classification and TapeLang's
  first Able cohort had 15.25% coefficient of variation.
- Able executions use the runner's guarded `GOMEMLIMIT=1GiB`, `GOGC=50`, and
  catalog-resolved CPU policy.

## Repeated timing

| Application | Mode | Able samples | Able mean | Limiting reference | Reference samples | Reference mean | Ratio |
| --- | --- | ---: | ---: | --- | ---: | ---: | ---: |
| Matrix Multiply | compiled | 10 | 1.350 s | Go | 10 | 1.206 s | 1.119x |
| TapeLang Alphabet | compiled | 10 | 5.013 s | Go | 10 | 2.523 s | 1.987x |
| Array Slice Window | bytecode | 5 | 0.788 s | Python | 5 | 0.0331 s | 23.807x |
| Dependency Plan | bytecode | 5 | 0.734 s | Python | 5 | 0.0204 s | 35.980x |
| Document Audit | bytecode | 5 | 0.322 s | Python | 5 | 0.0183 s | 17.596x |
| Lexical Rollup | bytecode | 5 | 0.498 s | Python | 5 | 0.0194 s | 25.670x |
| Option/Result Configuration | bytecode | 5 | 0.968 s | Python | 5 | 0.0199 s | 48.643x |

All 115 timing processes represented by these decisions verified with zero
failures and zero timeouts. After pooling, Matrix's Able coefficient of
variation is 5.50% and TapeLang's is 13.31%. The five bytecode Able coefficients
range from 4.70% to 12.33%.

Matrix no longer meets the 95% target in this current cohort. That is a guard
classification observation, not evidence that the truthiness/cast correction
caused the gap: the changed paths have zero reach, and the current profile is
99.31% direct generated `matmul`. TapeLang's larger miss likewise remains in
its direct application algorithm and methods.

## Reach and profile gate

Fresh generated application files for Matrix and TapeLang contain zero calls
to `__able_truthy`, `__able_cast`, `__able_try_cast`, `bridge.IsTruthy`, or
`bridge.Cast`. One verified main-only profile per current binary confirms the
static result:

| Application | Main CPU samples | Exact owners | Changed helper samples |
| --- | ---: | --- | ---: |
| Matrix Multiply | 1.44 s | `matmul` 99.31% flat; `build_matrix` 0.69% | 0 |
| TapeLang Alphabet | 5.08 s | `execute` 64.57%; `Tape.inc` 26.97%; `Tape.get` 6.30%; `Tape.move` 0.98% flat | 0 |

For bytecode, temporary opt-in counters were placed at the shared semantic
boundary, used only in untimed main-only census processes, then fully removed.
The second process reproduced every first-process count exactly:

| Application | Census processes | Primitive truthiness checks per process | Changed non-primitive/Error fallback | Explicit runtime casts | Cast failures |
| --- | ---: | ---: | ---: | ---: | ---: |
| Array Slice Window | 2 | 12,001 | 0 | 0 | 0 |
| Dependency Plan | 2 | 14,143 | 0 | 0 | 0 |
| Document Audit | 2 | 8,062 | 0 | 0 | 0 |
| Lexical Rollup | 2 | 83,196 | 0 | 0 | 0 |
| Option/Result Configuration | 2 | 24,576 | 0 | 0 | 0 |

The primitive truthiness switch was not changed by the correctness fix. The
new canonical/inherited Error lookup occurs only after that switch, and none
of these applications enters it. The new catchable cast-failure behavior also
cannot affect these rows because the shared explicit-cast boundary is never
called. Consequently the prior exact ownership profiles remain causally valid;
collecting new CPU/allocation profiles for zero-reach paths would not improve
candidate selection.

The concise machine-readable census is retained in
`2026-07-21-truthiness-cast-control-closure-reach.json`. No diagnostic branch,
counter, generated package, binary, or profile remains in production code.

## Exact timing artifacts

- Compiled Able:
  `2026-07-21-truthiness-cast-control-closures-compiled.json`
  (`46cc4a6a4bec0c1f4873de5bf029aacbb6da29417cc4f36b3613c62c7fb388a8`)
- Initial Go references:
  `2026-07-21-truthiness-cast-control-closures-go-reference.json`
  (`8e935bfad3376dcad132ceb763b802c737d621a11dfb122a66a337a72bc3078c`)
- Matrix second Able cohort:
  `2026-07-21-truthiness-cast-control-closures-matrix-c2-compiled.json`
  (`6fe41268fda34fee9816108328c1a79f88e95ee06921895aed6549243d0be570`)
- Matrix second Go cohort:
  `2026-07-21-truthiness-cast-control-closures-matrix-c2-go-reference.json`
  (`59086842c109cd22f56ed37e7e408a4d20cd1c4b682662b10aa499278389808b`)
- TapeLang second Able cohort:
  `2026-07-21-truthiness-cast-control-closures-tape-c2-compiled.json`
  (`3641358975f810a13284bf8350d5441d6dee33cc01d403da6de9d72940c68521`)
- TapeLang second Go cohort:
  `2026-07-21-truthiness-cast-control-closures-tape-c2-go-reference.json`
  (`6ddc4a5a2210120c6a19010c747b843147880b74234bd2003eb905e09de8a8ba`)
- Bytecode Able:
  `2026-07-21-truthiness-cast-control-closures-bytecode.json`
  (`f8f682956e46e64bfc86072c29dc3b70b1642a1047c2409c94a4fa30c0ced1f7`)
- Python/Ruby references:
  `2026-07-21-truthiness-cast-control-closures-interpreter-reference.json`
  (`3a685801d071f8c328884591713c78ef28c54b12c2db90e8382915b18d92bb78`)

## Next recommendation

Refresh `compiled-iterator-control` and `bytecode-float-numeric` next.

Why: the compiled iterator/control applications include Option/Result and
dynamic protocol boundaries where canonical Error truthiness could plausibly
reach compiled bridge helpers. The bytecode float/numeric family directly
exercises the specialized cast paths that were moved onto the catchable
explicit-cast boundary. These are the nearest remaining closures with a causal
route to the changed code, unlike the zero-reach closures closed here.

What it entails: reuse the current reference/source state; collect repeated
verified means; run exact main-only reach censuses first; and collect bounded
CPU/allocation profiles only for applications where the changed fallback or
cast wrapper is material. Advance only those two closures. A candidate remains
conditional on one concrete non-benchmark-specific leaf appearing in at least
three unlike applications and preserving all current target guards.
