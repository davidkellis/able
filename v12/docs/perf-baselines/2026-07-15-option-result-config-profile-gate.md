# Option/Result Configuration Three-Application Profile Gate — 2026-07-15

## Decision

Keep no VM, compiler, canonical-stdlib, or benchmark-workload performance
change. The new Option/Result Configuration application exposes a large
generic-lambda/lowering cost, but that concrete descendant does not recur in
the two unlike control applications. The only concrete three-way VM leaf is
the already-rejected `finishInlineReturn(...)` family, so it is not a new
candidate.

## Correctness and measurement contract

The normal compiled and bytecode processes for Option/Result Configuration,
Dependency Plan, and Document Audit were already verified three times each by
their external Ruby verifiers on CPU 11:

- `2026-07-15-option-result-config-three-application-comparison.json`
- `2026-07-15-option-result-config-three-application-comparison.md`

The warmed runtime harness intentionally suppresses program output, so its
in-process rows are not verifier rows. The permanent
`TestBytecodeOptionResultConfigGenericUnionMethodsRemainResolvedWhenWarm`
instead loads the canonical external stdlib and calls the application `main`
twice, asserting `1024:18221610432` both times.

The warm harness now sets `ABLE_BENCH_SKIP_TYPECHECK=0` during its one-time
setup. Generic named unions have structural runtime values, so ordinary
overloaded method dispatch needs the checked receiver fact. Typechecking,
loading, lowering, and the explicit warmup stay outside the measured loop;
this does not change the separate trusted CLI `--skip-typecheck` lane.

## Bounded warmed profiles

Each capture used one process, `GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`,
a 45-second cap, and the immediate three-sample `--require-quiet-cpu` gate.
The host cores differ, so these numbers are profiling windows and allocation
descriptions, not cross-core speed deltas.

| Application | Quiet CPU | Timed calls | ns/op | B/op | allocs/op | CPU samples |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Option/Result Configuration | 0 | 2 | 3,863,064,628 | 946,641,888 | 7,343,709 | 7.69 s |
| Dependency Plan | 6 | 28 | 245,856,276 | 2,189,663 | 33,246 | 6.87 s |
| Document Audit | 12 | 900 | 10,217,399 | 382,511 | 611 | 9.16 s |

The machine-readable results are:

- `2026-07-15-option-result-config-bytecode-runtime-cpu-profile.json`
- `2026-07-15-dependency-plan-bytecode-runtime-cpu-profile.json`
- `2026-07-15-document-audit-bytecode-runtime-cpu-profile.json`

The associated local CPU profiles are retained under
`v12/interpreters/go/.profiles/` while active development continues.

## Leaf comparison

Option/Result Configuration is dominated by repeated runtime lowering of its
ordinary lambda/function-definition expressions:
`lowerLambdaExpressionBytecodeWithEnv(...)` is 39.01% cumulative and its
`lowerFunctionDefinitionBytecodeWithEnv(...)` child is 37.71%. Its allocation
and GC samples are correspondingly large. This is a legitimate generic VM
cost, but neither Dependency Plan nor Document Audit shows it as a material
leaf.

Dependency Plan instead concentrates in Array/member/index work:
`execCallMember(...)` is 23.14% cumulative,
`lookupCachedMemberMethodEntry(...)` 10.48%,
`finishInlineReturn(...)` 9.75%, and direct Array index paths make up the next
distinct group.

Document Audit is generator and member-cache work:
`execCallMember(...)` is 41.70% cumulative,
`lookupCachedMemberMethodEntry(...)` 28.49%, and
`finishInlineReturn(...)` 9.83%. Its iterator generator route accounts for
the surrounding call-opcode parent.

`execCallOpcode(...)` and `execCallMember(...)` recur in all three profiles,
but they are broad dispatcher parents. The only common concrete child is
`finishInlineReturn(...)`: 3.90% for Option/Result Configuration, 9.75% for
Dependency Plan, and 9.83% for Document Audit. Its generic guard-order
experiment was already tested and removed because it was neutral-to-mixed on
the broad guard set. Do not retry it without materially new evidence.

The cached-member path is strong in Dependency Plan and Document Audit but is
only 1.43% in Option/Result Configuration; the lambda-lowering path is strong
only in Option/Result Configuration. Neither clears the three-unlike-program
admission rule.

## Follow-up

The bytecode selection gate is closed with no candidate. The next performance
tranche should profile generated compiled binaries across the same three
applications, using a generic repeated-process profile/merge mechanism if
needed to obtain enough samples. The current verified full-process comparison
shows that compiled performance remains an important gap, while this bytecode
gate rules out spending that effort on an Option/Result-specific VM shortcut.
