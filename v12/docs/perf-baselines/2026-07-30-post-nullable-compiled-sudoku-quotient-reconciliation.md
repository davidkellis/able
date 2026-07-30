# Post-nullable compiled Sudoku quotient reconciliation

## Decision

Reconcile `compiled-sudoku-quotient` as causally current and retain no
production change.

The current strict Sudoku Masks application keeps its search state, masks,
indices, quotient, and remainder in native Go `int32` carriers. The primitive
nullable change reaches only `main`'s absent return after all 100 requested
solves finish. That return is an allocation-free `value + valid` carrier and
does not enter `find_best_empty`, `solve`, `solve_with_masks`, or
`square_index`.

Signed Euclidean division remains material only in Sudoku. Its two generated
call sites both use the general signed helper with a positive constant divisor
of three; the helper's positive path performs native Go `/` and `%`. Rational
Series uses the distinct two-word `runtime.Int128.DivMod` path, and the Regex
and K-Nucleotide controls do not make the exact signed-`i32` helper material.
The required three-unlike-application breadth therefore remains one.

## Strict boundary and execution gate

Sudoku Masks was rebuilt from the retained compiler with `--no-fallbacks`.
The generated dependency graph has 96 packages and omits
`pkg/interpreter`. The exact binary passed the public verifier from the
benchmark-suite working directory with arguments `10 10`.

| Gate | Result |
| --- | --- |
| Strict build | passed |
| Dependency packages | 96 |
| Interpreter dependency | absent |
| Public verifier | passed |
| Validation smoke | 1.50 s, 9,896 KiB peak RSS |

The smoke is an execution check, not timing evidence. Its output SHA-256 is
`35a81e448daf9986f2a9b7c3a873dc6216bd55c969efec50c9b1de6d866659ec`.
The authoritative scorecard retains five successful Able and Go processes.

## Current-source carrier reach

The benchmark source has changed since the retained July 21 quotient record,
principally to accept positive command-line counts. The current source and
generated application body were therefore audited rather than assuming the
old reach:

| Artifact | SHA-256 |
| --- | --- |
| Able source | `222b321f579d7b2a84f4bc0fd379064a7ebe554bd83169782b28d04eaaab90e0` |
| Benchmark contract | `23fb9a195d292c88efb87180cb2153f5ad893a960aed03cec070a37f2906ada6` |
| Public verifier | `2898dd09ba0cb05362589c83f95c2d8ae979b09f791cd9f2e63c762bfa221ce7` |
| Generated module | `aed73c15ab545c9a3f56e7c6d7b9d3aad9c3bfe98c2e2a15d376a79057e9f0b2` |
| Generated application bodies | `af94b597e2146cdcd5ea62100def9c16cc2328ebe2f78b828f703f8699c524aa` |
| Strict binary | `7014eaeb9a4b0cf4ae7db905df49e16134c5568be03cefc9acd71e7e7ac94e87` |

The generated application section has 15 textual nullable-`i32` references.
Every one belongs to `main`'s signature, error exits, or final absent return.
It has zero nullable helper calls. The material search ranges contain:

- zero nullable-`i32` references;
- zero nullable helper calls;
- zero `runtime.Value` references; and
- zero bridge conversions.

The registered entry wrapper converts `main`'s absent carrier to the required
runtime nil value once, after the complete solve loop. This is a real static
to-host boundary, but it is neither hot nor an allocation owner. No
compiler/interpreter boundary occurs in the search.

## Quotient and search audit

The current generated application has exactly two
`__able_divmod_signed[int32]` calls, both in `square_index`. They compute the
row quotient and remainder for a positive index and positive constant divisor
three. `square_index` is reached from the material mask initialization and
search paths.

The relevant helper first rejects zero, then uses this native fast path when
the dividend is non-negative and divisor positive:

```go
q := a / b
r := a % b
```

Only negative operands enter the additional Euclidean-sign correction.
Consequently the hot Sudoku case already lowers to Go integer carriers and Go
division instructions behind a shared semantic helper. The remaining call
overhead is general primitive semantic machinery, not boxing, interpreter
fallback, dynamic dispatch, or a named-container rule.

The retained two-profile merge remains the best exact ownership evidence
because the material search body and quotient shape are unchanged:

| Owner | Retained share |
| --- | ---: |
| `__able_divmod_signed` flat | 11.51% |
| `__able_divmod_signed` cumulative | 12.53% |
| `square_index` cumulative | 35.04% |
| `find_best_empty` cumulative | 89.00% |

The source change does not alter those functions, and the freshly generated
`square_index` body SHA-256 is
`db7972f4be24b3101fd6a7cc64586ef5e2fa8805ec4cc743f3aa0b9ba635bee7`.
The relevant signed-helper body SHA-256 is
`14a0303919009ab72062777a6599845de03dda31350a5f745a72cdfe325af7d1`.

## Breadth and impact gate

The unlike controls remain disqualifying:

- Rational Series performs wide signed arithmetic through
  `runtime.Int128.DivMod`, not the generated signed-`i32` helper.
- Current Regex material paths are NFA closure, move, thread upsert, and
  acceptance scans; quotient work is confined to a cold DFA descendant.
- K-Nucleotide's material owner is text/map work; quotient formatting is
  negligible.

The exact signed-`i32` helper is therefore material in one unlike
application, not three.

Sudoku's current five-process mean is 1.9660 seconds versus 0.7184 seconds for
Go, a 2.736637x ratio. Even assigning the retained 12.53% cumulative share
zero cost gives only a 1.143249x idealized speedup and leaves a 2.393736x
ratio. It would still require another 2.274049x improvement to reach the
1.052632x target threshold. The quotient leaf cannot close the application
gap by itself.

Fresh CPU/allocation profiles and an A/B implementation were not admitted.
The changed carrier is outside the hot path, the exact quotient leaf still
lacks breadth, and perfect removal is insufficient. Repeating profiles cannot
turn this one-algorithm owner into a general candidate.

## Scope, verification, and cleanup

No compiler, generated runtime, runtime package, interpreter, bytecode VM,
canonical stdlib, benchmark, language, dependency, nominal special case, or
WASM source changed.

The focused executable division fixture and concrete-carrier guard passed:

```text
ABLE_COMPILER_EXEC_FIXTURES=06_01_compiler_division_ops \
  go test ./pkg/compiler \
  -run '^TestCompilerExecFixtures$|^TestCompilerDivModConcreteCarrierStaysNative$' \
  -count=1 -timeout 50s
```

It completed in 3.165 seconds. `go test ./cmd/ablec` also passed in 6.742
seconds. The machine-readable record is
`2026-07-30-post-nullable-compiled-sudoku-quotient-reconciliation.json`.

After retaining this evidence, the exact 310 MiB disk-backed compiler,
generated-module, binary, audit, smoke-output, and Go-cache workspace was
removed. No matching tranche artifact remains in `/var/tmp` or `/tmp`.

## Next

Reconcile `compiled-concurrency` against the primitive nullable carrier.

Why: it is the broadest remaining invalidated compiled frontier and the last
unreviewed compiled family before the cross-family architecture ledger can be
reconciled. Future, channel, and mutex applications are where primitive
results are most likely to cross generated concurrency and host boundaries.

What it entails: strictly rebuild and census all 23 concurrency rows for
interpreter-free graphs and causal primitive-nullable reach. Reuse the
retained post-spawn profiles for Await Channel Mux, Validated Job Pipeline,
and Concurrent Stateful Pipeline; refresh profiles only if a changed shared
path is materially reached. Preserve the rejected global integer cache,
`currentGID`, broad execution-context ABI, nominal-struct, and
application-specific routes.

Why it matters: compiled concurrency still misses Go by 5.33x-19.65x. This
review tests the central lowering claim at goroutine/channel boundaries and
can identify a genuinely general boxing reduction—or close the final family
without weakening Able's semantics.
