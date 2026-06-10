# Feature-to-Application Coverage Contract

## Decision

Do not add a new portable benchmark application or performance implementation
in this tranche. Reconciliation against the current v12 spec, TODO, fixtures,
portable catalog, and retained July 15 audit confirms that no specified
portable feature family is absent. Inventing a dynamic-import, host-extern, or
test-reporter timing loop would not create an equivalent Go/Python/Ruby
application and would weaken the benchmark policy.

The retained change turns that conclusion into a machine-checked contract so a
future spec or catalog change cannot silently invalidate it.

## Coverage result

`bench-feature-coverage.json` maps 15 feature families across all 16 normative
top-level spec sections (sections 2 through 17):

- all 32 portable applications appear in at least one feature family;
- all ordinary portable surfaces have cross-runtime application coverage;
- dynamic metaprogramming, user-authored host interop, and the testing
  framework are explicit `local_only` families with existing execution
  fixtures and rationale;
- Option/Result exceptions, concurrency cancellation/policy, and package
  dynamics are `mixed`: ordinary behavior has portable application coverage,
  while non-equivalent policy/host semantics stay in focused local fixtures;
  and
- the portable catalog itself remains the authority for comparable timing.

The checker derives the required section set from `spec/full_spec_v12.md`,
loads the current `coverage` catalog, and validates every declared fixture on
disk. It rejects unknown or missing spec sections, unmapped portable programs,
invalid classification, local-only claims that name a portable timing row,
duplicate entries, and missing fixture entry points.

## Verification

- `just bench-catalog-check`:
  - 32 portable applications;
  - 33 canonical sources including one diagnostic source;
  - 77 bounded local benchmark fixtures;
  - 109 programs in the combined lowering corpus;
  - 15 feature families covering 16 normative sections;
  - five feature-contract failure-mode tests pass.
- Focused tree-walker/bytecode fixture parity passes for dynamic package
  metaprogramming, Future cancellation/fairness, dynamic-import interface
  dispatch, user-authored inline externs, and test reporters (`1.296s`).
- `bench_bytecode_audit --suite corpus-full` completes under the normal
  `GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1` guard and includes all 109
  catalogued application/fixture programs.

No spec semantics, parser, compiler, VM, canonical stdlib, application,
foreign reference, verifier, scorecard, selection manifest, or benchmark
timing changed.

## Why no profile followed

This audit found no new application or changed semantic path. Profiling the
same unchanged corpus would reproduce the already-closed residual inventory,
not satisfy the rule requiring a new concrete non-nominal leaf across three
unlike applications. The correct result of the requested feature-led audit is
therefore a durable coverage contract and no fabricated optimization.

## Next recommendation

Audit source and algorithm equivalence for the largest stable target misses,
starting with Fixed Width 128, Rational Series, the concurrency family, and the
regex/text family. Compare Able, Go, Python, and Ruby sources at the level of
algorithm, input size, operation counts, data representation required by the
language contract, and library work delegated to the host. Record whether each
gap is runtime/compiler overhead, canonical-stdlib algorithm cost, or a source
shape mismatch.

Why: verifier equality proves output equivalence, but it does not prove that
the implementations perform equivalent work. The current profile inventory
has no reusable runtime leaf, while several ratios are orders of magnitude
apart. A source-equivalence audit can expose a broadly useful stdlib algorithm
repair or correct an unfair reference comparison without adding a named
compiler fast path. Only a mismatch or a shared semantic/library boundary
found in at least three unlike programs should advance to code and repeated
five-run A/B measurement.
