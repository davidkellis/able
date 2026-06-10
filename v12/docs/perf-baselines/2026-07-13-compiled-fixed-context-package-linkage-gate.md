# Fixed-context package-linkage gate — 2026-07-13

## Scope

This gate tested a generic refinement of the opt-in fixed-pointer
execution-context ABI. The candidate removed the per-package-entry
`__able_context_with_environment` clone identified by the preceding N-body
profile. An execution context carried task payload only; compiled package
entries retained the dynamic bridge environment swap, and `spawn` derived its
child environment from the active bridge environment.

The change applied to every generated package entry and every spawned task. It
did not inspect a benchmark, function name, package name, math operation, task
count, or nominal container.

The candidate was reverted. It removed the known allocation wall and repaired
N-body, but it created a repeated material K-Nucleotide regression. A default
generated-call ABI must be broadly neutral or better, not an exchange between
application families.

## Semantic gate

Before timing, the candidate passed generated-source coverage, the race-built
nested-spawn execution test, fixture parity, and dynamic named/value boundary
parity. The relevant focused command was:

```sh
go test ./pkg/compiler \
  -run 'TestCompilerExperimentalExecutionContext(UsesFixedPointerForStaticCalls|ThreadsStaticSpawnKernelCalls|NestedSpawnExecutes|FixtureParity|DynamicBoundaryParity|DynamicNamedAndValueBoundaries)$' \
  -count=1 -timeout 90s
```

The same coverage passed after the revert. All timed executions below passed
their canonical Ruby verifier with deterministic output hashes.

## N-body allocation check

Paired three-run measurements used CPU 15, `GOMEMLIMIT=1GiB`, `GOGC=50`,
`GOMAXPROCS=1`, and a 45-second cap:

| ABI | Mean | GC mean | Result |
| --- | ---: | ---: | --- |
| Default | 0.4600s | 3.00 | verified |
| Payload-only package linkage | 0.4467s | 3.00 | verified, 2.9% lower |

The paired generated-main profiles confirm the intended local effect:

- Default: 350 ms samples over 345.25 ms wall.
- Candidate: 360 ms samples over 356.20 ms wall.
- Neither candidate profile contains `__able_context_with_environment` or a
  `runtime.mallocgc` allocation descendant. The remaining material work is
  generated `sqrt`/`abs` and existing bridge environment swaps.

Retained profiles:

- `v12/interpreters/go/.profiles/20260713_fixed_context_package_linkage_nbody_default/main.cpu.pprof`
- `v12/interpreters/go/.profiles/20260713_fixed_context_package_linkage_nbody_candidate/main.cpu.pprof`

## Serial generality gate

The same guarded three-run screen repeated the four non-N-body serial controls
that selected as suspicious in the preceding ABI scorecard:

| Application | Default | Candidate | Change | Decision |
| --- | ---: | ---: | ---: | --- |
| Sudoku Masks | 10.0933s | 9.8733s | -2.2% | neutral/better |
| K-Nucleotide | 4.1267s | 5.1767s | +25.4% | repeat required |
| Mandelbrot | 0.1400s | 0.1633s | +16.6% | too short for authority |
| Reverse Complement | 0.1500s | 0.1233s | -17.8% | better |

The K-Nucleotide selection loss was immediately repeated under the same
guards:

| K-Nucleotide pass | Default | Candidate | GC default/candidate |
| --- | ---: | ---: | ---: |
| First | 4.1267s | 5.1767s | 57.33 / 61.67 |
| Repeat | 4.1600s | 4.4867s | 57.67 / 61.33 |
| Six-run mean | 4.1434s | 4.8317s | 57.50 / 61.50 |

The six-run candidate result is **16.6% slower** and has more garbage
collections. It is a stable, verifier-backed loss in a separate text/map-heavy
application, so no Channel Rollup/Future Pipeline confirmation or broad
scorecard is warranted for this version.

## Decision

- Revert the payload-only package-linkage refinement; retain no compiler,
  bridge, runtime, bytecode VM, or canonical-stdlib source change from it.
- Keep the previously existing fixed-context ABI opt-in and leave the compiler
  default unchanged.
- Do not infer a K-Nucleotide-, map-, math-, package-, or benchmark-specific
  remedy from this gate. The refinement itself was generic; its failed
  generality gate means it is not eligible to keep.

## Next

Return to profile-led selection among the largest current compiled-versus-Go
misses. Refresh bounded generated-main CPU and allocation evidence across
unrelated numeric, text, recursive, and nominal controls, and proceed only if
one language-level primitive operation or dynamic-boundary cost repeats in
multiple applications. A future execution-context change must be screened
against N-body, K-Nucleotide, and the independent concurrency pair before any
default decision.
