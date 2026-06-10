# Qualified impl diagnostic reconciliation and cross-mode profile refresh

## Decision

Update one stale fixture expectation and retain no compiler, VM, runtime, or
standard-library performance change.

The `alias_reexport_impl_ambiguity` runtime diagnostic now names the canonical
interface identity, `alias_reexport_impl_ambiguity_base.Taggable`. The runtime
was already correct: v12 re-exports and aliases preserve the original nominal
identity, and package qualification distinguishes unrelated interfaces that
share a short name. The fixture still expected the older short label.

The typechecker diagnostic remains source-facing and continues to say
`Taggable`; only the expected runtime error was reconciled. No language rule,
AST, parser, compiler, interpreter, or canonical `able-stdlib` source changed.

## Correctness controls

The exact failing fixture reproduced before the manifest correction and passed
after it. The full Go fixture-parity traversal then passed, including the
qualified-interface identity and other ambiguity fixtures.

- exact alias/re-export ambiguity fixture;
- alias/re-export plus generic/concrete ambiguity controls;
- focused typechecker implementation, alias, and re-export tests; and
- complete `TestFixtureParityStringLiteral` traversal.

## Warmed bytecode refresh

One clean interpreter test binary ran four unrelated loader-driven programs.
Each process loaded and warmed once, then measured repeated `main()` calls for
at least five seconds while collecting CPU samples. Profile overhead was kept
outside any baseline/candidate comparison because no candidate was built.

| Application | Iterations | ns/op | B/op | allocs/op | CPU samples |
| --- | ---: | ---: | ---: | ---: | ---: |
| Document Audit | 512 | 11,569,853 | 367,233 | 489 | 6.45 s |
| Dependency Plan | 21 | 298,498,779 | 1,957,110 | 27,825 | 6.25 s |
| Future Pipeline | 42 | 131,695,012 | 14,145,700 | 655,490 | 27.02 s across goroutines |
| Configuration Validation/Extraction | 7 | 780,523,482 | 15,664,217 | 156,124 | 5.55 s |

The concrete owners diverge:

- Document Audit is cached member lookup and iterator/member-call work;
- Dependency Plan is member/index calls, Go map access, and Array storage;
- Future Pipeline is integer arithmetic plus goroutine scheduling/locking;
- Configuration Validation/Extraction is canonical Array-slot calls,
  named-field access, and Regex/NFA application work.

`appendSlotStackValueChecked` is the only exact leaf visible in all four at
0.93%, 2.08%, 2.15%, and 3.42% flat. This is not new evidence: the coverage
census already found it in thirteen applications, and two general
ordering/carrier candidates failed broad wall-time guards. The current shares
are smaller than several of those rejection cohorts. `runResumable`, call
dispatch, return handling, Go map internals, and atomic counters are parents or
split into different semantic consumers.

## Compiled generated-main refresh

Three current binaries were built once and profiled through the CPU-only
generated phase hook. They were selected because each has a multi-second user
main, avoiding startup-dominated short-process profiles. Every profiled output
passed its public Ruby verifier.

| Application | Main CPU samples | Material owner |
| --- | ---: | --- |
| K-Nucleotide | 3.70 s | allocation/GC, primitive HashMap equality/hash, integer and String conversion |
| Base64 | 2.57 s | Go `encoding/base64` and `crypto/md5` kernels (94.94% flat) |
| TapeLang Alphabet | 3.68 s | direct generated `execute`, `Tape.inc`, `Tape.get`, and `Tape.move` bodies |

There is no compiler-controlled concrete leaf shared across the three. K-
Nucleotide's generic-runtime boundary, Base64's host kernels, and TapeLang's
direct generated control loop separate immediately below the registered-main
wrapper. Runtime GC scanning occurs in K-Nucleotide and Base64 but not
TapeLang, and its allocation callers differ.

Five additional independent launches per preserved binary all verified:

| Application | Mean | Min-max | Verification |
| --- | ---: | ---: | ---: |
| K-Nucleotide | 2.332 s | 2.22-2.54 s | 5/5 |
| Base64 | 2.364 s | 2.26-2.46 s | 5/5 |
| TapeLang Alphabet | 3.840 s | 3.67-4.01 s | 5/5 |

## Normal bytecode controls

The external-comparison harness ran five independent candidate-free bytecode
processes for each profiled source family. All twenty outputs verified.

| Application | Mean | Verification |
| --- | ---: | ---: |
| Document Audit | 0.268 s | 5/5 |
| Dependency Plan | 0.440 s | 5/5 |
| Future Pipeline | 0.402 s | 5/5 |
| Configuration Validation/Extraction | 1.266 s | 5/5 |

## Verification and cleanup

- `go test ./pkg/interpreter -run '^TestFixtureParityStringLiteral$' -count=1 -timeout 55s`
- focused interpreter ambiguity fixtures;
- focused typechecker implementation/alias/re-export tests;
- four warmed bytecode benchmark/profile processes;
- three verifier-backed compiled profile processes;
- fifteen additional verifier-backed compiled launches;
- twenty normal verifier-backed bytecode launches; and
- `git diff --check`.

All performance artifacts were generated under `/tmp` and are cleanup-only.
No performance candidate was admitted, so no A/B timing claim is made.

## Next selection

Follow-up completed by
`2026-07-20-current-iterator-control-profile-refresh.md`. Current generated-
main profiles still split across the four applications, while warmed bytecode
repeats only the already-completed dependency-validated member-cache family.
No candidate advanced.

Refresh the one material no-shared-leaf frontier whose compiled evidence is
still explicitly pre-current: the iterator/control cohort spanning Document
Audit, Dependency Plan, Lexical Rollup, and Option/Result Configuration. Pair
those repeated generated-main profiles with current warmed bytecode profiles
for the same applications.

Why: this tranche refreshed long-running compiled owners and a broad VM
screen, but the frontier ledger still marks the short compiled iterator/control
binaries as source-compatible rather than current-profiled. They cover common
real-program behavior and are the best remaining place where current code
could invalidate an old no-shared-leaf decision.

This entails building each binary once, collecting enough independent CPU-only
main-phase launches to overcome sampler resolution, verifying every output,
and intersecting exact leaves with the bytecode cohort. Advance a candidate
only when the same generic, non-nominal child is material in at least three
unlike applications and is not an already-rejected member-cache, stack,
return/frame, nullable/union, Array-growth, or environment-swap design. Do not
begin WASM work.
