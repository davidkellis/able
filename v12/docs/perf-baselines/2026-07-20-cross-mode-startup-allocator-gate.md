# Cross-mode startup and Tree-sitter allocator gate

## Decision

Keep the generic Tree-sitter allocator restoration. The Go Tree-sitter
binding installs C-to-Go-to-C allocator callbacks even when its caller has not
provided a custom allocator. Able never replaces that allocator. Restoring
Tree-sitter's documented native defaults before the first Able parser is
created reduces the exact loader phase in three unrelated programs by
17.1%-18.6%, without changing parsing, AST construction, interpreter
semantics, generated application code, or canonical stdlib code.

This is not a benchmark or nominal-container fast path. Every source program
loaded by the CLI, bytecode interpreter, compiler, or test harness uses the
same parser boundary.

## Cohort and method

The cohort deliberately spans unrelated short applications:

- Document Audit: files, text, lazy iterators, and predicates.
- Dependency Plan: nominal collections, graph traversal, queues, and maps.
- Future Await Race: `spawn`, futures, await, and the goroutine executor.

Each normal comparison used ten independent processes per application and
mode under the catalog's source-root, argument, executor, CPU-budget, and Ruby
verifier contracts. The bytecode runtime harness also gained opt-in phase JSON
and exact loader CPU profile outputs. Those diagnostics run only during setup;
they do not execute in normal CLI runs or in the timed warmed `main()` loop.

Baseline phase means from ten fresh benchmark processes were:

| Application | Program load | Typecheck + module evaluation/lowering | Warm `main()` | Entry-to-ready |
| --- | ---: | ---: | ---: | ---: |
| Document Audit | 180.126 ms | 38.639 ms | 12.891 ms | 270.728 ms |
| Dependency Plan | 83.462 ms | 9.684 ms | 297.802 ms | 429.536 ms |
| Future Await Race | 48.102 ms | 4.702 ms | 9.825 ms | 91.137 ms |

Program loading was the only material setup phase repeated in all three
programs. Ten merged loader profiles per application attributed
`ModuleParser.ParseModule` 64.2%-69.4% cumulative and Tree-sitter's native
grammar parse alone 34.6%-38.8%. `runtime.cgocall` was 46.9%-58.7% flat, with
the binding's allocator callbacks visible underneath it. This exact owner did
not overlap the rejected lazy integer cache, generated reachability, bootstrap
cache, VM return, raw-integer, or named-container designs.

## Candidate result

`pkg/parser` now restores Tree-sitter's native allocator once, before creating
the first parser. Ten new phase processes per application produced:

| Application | Baseline load | Candidate load | Load change | Baseline ready | Candidate ready | Ready change |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Document Audit | 180.126 ms | 146.751 ms | -18.5% | 270.728 ms | 232.672 ms | -14.1% |
| Dependency Plan | 83.462 ms | 69.181 ms | -17.1% | 429.536 ms | 414.203 ms | -3.6% |
| Future Await Race | 48.102 ms | 39.135 ms | -18.6% | 91.137 ms | 81.759 ms | -10.3% |

The independent verifier-backed cold-process gate also improved all three
bytecode means:

| Application | Baseline bytecode | Candidate bytecode | Change | Verification |
| --- | ---: | ---: | ---: | ---: |
| Document Audit | 0.353 s | 0.295 s | -16.4% | 10/10 both cohorts |
| Dependency Plan | 0.506 s | 0.485 s | -4.2% | 10/10 both cohorts |
| Future Await Race | 0.144 s | 0.142 s | -1.4% | 10/10 both cohorts |

The warmed application-work guard remained healthy. Document Audit measured
11,958,123 -> 11,555,427 ns/op with 926 allocs/op; Dependency Plan measured
287,665,014 -> 271,941,668 ns/op with 11,970 allocs/op; Future Await Race
measured 9,673,026 -> 9,703,880 ns/op. These are workstation controls, not
claimed steady-state wins: the parser cannot execute inside the timed warmed
loop. Compiled launch rows all verified 10/10; their changes are likewise
control noise because generated binaries do not parse Able source at launch.

Post-candidate loader profiles no longer contain a recurring allocator
callback owner. The remaining shared wall is split between the real
Tree-sitter grammar parse, Go AST mapping through node-access C calls, and
origin annotation. That split does not yet justify another change.

Machine-readable and rendered normal-process records:

- `2026-07-20-cross-mode-startup-decomposition.json`
- `2026-07-20-cross-mode-startup-decomposition-scorecard.md`
- `2026-07-20-cross-mode-startup-allocator-candidate.json`
- `2026-07-20-cross-mode-startup-allocator-candidate-scorecard.md`

## Verification

- `go test ./pkg/parser ./pkg/driver -count=1 -timeout 60s`
- `go test ./pkg/interpreter -run 'TestLoadBytecodeProgramRuntimeBenchConfig|TestWriteBytecodeProgramRuntimeRetentionHeapProfile' -count=1 -timeout 60s`
- `go test ./cmd/able -run '^TestBuildEnvFalseAllowsFallbacks$' -count=1 -timeout 60s`
- `go test ./pkg/compiler -run '^TestCompilerIteratorFilterMapBenchmarkShapeExecutesFromNonMainSourcePackage$' -count=1 -timeout 60s`
- `go test ./pkg/interpreter -run '^TestExecFixtureParity/14_05_regex_nfa_wildcards_classes$' -count=1 -timeout 60s`

The broad `go test ./... -count=1 -timeout 60s` command exceeded the package
timeout in the large `cmd/able`, compiler, and interpreter aggregate suites.
All three tests named at the timeout passed individually in 9.756 s, 17.775 s,
and 0.741 s. Other reported packages, including parser and driver, passed.

## Next selection

Next split the post-candidate parser wall into native grammar parsing, node
access/Go AST mapping, and origin annotation across these three programs plus
one larger source-heavy control. Require one exact C accessor or AST mapping
operation to repeat materially before changing traversal. This is the next
useful step because native parsing and mapping are currently interleaved in
`runtime.cgocall`; optimizing their parent would risk a parser-specific guess,
whereas an exact repeated child could improve bytecode cold start and compiler
front-end time for every Able program. The work entails opt-in subphase
timings, repeated merged profiles, a general cursor/field-symbol/cached-node
candidate only if admitted by that evidence, and the same cold/warm verifier
gate. No WASM work is admitted.
