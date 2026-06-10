# v12 feature-to-benchmark coverage audit (2026-07-12)

## Outcome

### 2026-07-13 addendum

The original audit later identified one distinct cross-language concurrency
gap: no application selected among public `Channel.await_receive` arms and an
`Await.default` arm. `await-channel-mux` fills that boundary with matching
Able, Go, Python, and Ruby programs. It also found that the existing public
`regex_suffix_audit` and `regex_set_audit` applications were present in the
dedicated `regex-text` suite but missing from the full application inventory.
The corpus therefore reached 27 external applications. The scanner surface
was still fixture-only, so `regex_stream_audit` now adds the 28th application:
it exercises public chunked `RegexScanner.feed` / `next` / `flush` semantics
with ordinary buffered-record counterparts in Go, Python, and Ruby. The added
rows have pinned reference screens; the material product misses repeat only
known generic parents or already-kept canonical-regex work, not a new
performance candidate. The detailed records are
`2026-07-13-await-channel-mux-{coverage,reference-profile-gate}.md` and
`2026-07-13-regex-application-coverage-refresh.md`, plus
`2026-07-13-regex-stream-application-coverage.md`.

### 2026-07-14 addendum

`mutex_ledger` is the 29th portable application. It covers public
`Mutex` and `able.concurrency.with_lock` through four contended workers that
update one shared nominal ledger. Its Go, Python, and Ruby counterparts share
an exact-output verifier, and the app is present in the dedicated
`mutex-ledger`, `concurrency`, and broad `coverage` suites. The coverage work
also repaired generic bytecode aggregate-value materialization and the
multi-waiter mutex handoff in both bytecode and generated compiled runtime.
`mutex_await_journal` is now the 30th: four workers use public
`Mutex.await_lock` and `await` with an `ensure` unlock callback. It is also
verifier-backed in every language and belongs to the dedicated
`mutex-await-journal`, `concurrency`, and `coverage` suites. Together they
repaired generic await registration/re-arm and async-context propagation, and
they repeated ordinary lambda-bytecode lowering strongly enough to retain a
bounded environment-shape-aware cache. See
`2026-07-14-mutex-ledger-application-coverage.md` and
`2026-07-14-mutex-await-journal-application-coverage.md`.

At the time of the original audit, the active benchmark corpus had no
unrepresented, application-shaped v12 feature family that should be filled by
inventing a new benchmark. It had 22 cross-language Able applications and 77
bounded local performance fixtures.
The audit did find a harness coverage defect: `bench_bytecode_audit
--suite corpus-full` included only the 16-program `generality` scorecard plus
fixtures, silently omitting six existing application programs. The catalog now
provides a 22-program `coverage` suite, and `corpus-full` uses it before adding
the 77 fixtures.

`generality` remains the stable 16-program performance baseline. `coverage`
is an explicit broader application inventory, not a replacement ratio series.

## Original audit corpus

| Corpus | Count | Role |
| --- | ---: | --- |
| External-style applications | 22 | ordinary programs with Able, Go, Python, Ruby, and semantic-verifier lanes in the sibling benchmark repository |
| `generality` application baseline | 16 | stable cross-runtime ratio/status scorecard |
| Additional application coverage | 6 | Fixed Width 128, Rational Series, Word Frequency, Document Audit, Lexical Rollup, and Channel Rollup |
| Local `fixtures/bench` programs | 77 | bounded parity and regression guards, including features without fair foreign-runtime analogues |

programs, 377 lowered functions, and 18,081 bytecode instructions. This is a
lowering-coverage check, not a timing result.
The original `bench_bytecode_audit --suite corpus-full` succeeded with all 99
programs, 377 lowered functions, and 18,081 bytecode instructions. This is a
lowering-coverage check, not a timing result. The addendum above records the
three later concurrency applications separately.
programs, 377 lowered functions, and 18,081 bytecode instructions. This is a
lowering-coverage check, not a timing result.

## Feature matrix

| v12 family | Application-shaped coverage | Local guard coverage | Performance-use status |
| --- | --- | --- | --- |
| Lexical forms, bindings, blocks, operators, conditionals, loops, recursion, and patterns | Fib, QuickSort, Sudoku/Sudoku Masks, K-Nucleotide, TapeLang | automata, graph, knapsack, sieve, iterator-match | broad application timing |
| Primitive widths, f32/f64, BigInt, fixed-width 128, and rational arithmetic | MatrixMultiply, Monte Carlo Pi, Mandelbrot, N-body, PiDigits, Fixed Width 128, Rational Series | bigint/biguint/int128/uint128 and numeric fixtures | broad numeric controls; do not infer one shared leaf from ratios alone |
| Arrays, strings, bytes, files, codecs, regex, JSON, and stdlib/host-backed I/O | Base64, JSON, I-Before-E, Reverse Complement, K-Nucleotide, Word Frequency, Document Audit, Regex Suffix Audit, RegexSet Audit, Regex Stream Audit | byte histogram, regex, string, MD5, run-length fixtures | broad application timing |
| Structs, named fields, methods, nullable values, unions, interfaces, generic collections, and dynamic method traffic | BinaryTrees, K-Nucleotide, TapeLang, lexical/document rollups | linked-list iterator, persistent/tree collection, interface and union fixtures | ordinary nominal/iterator traffic has independent guards |
| Functions, closures, callbacks, partial/member calls, and iterators | Lexical Rollup and Document Audit | array/linked-list/lazy iterator fixture families | broad controls plus local call-shape guards |
| Result/error matching and `raise` | Base64 and JSON exercise result/error branches on ordinary input | rescue/ensure/rethrow, result propagation, and error-path fixtures | only success-path application comparisons are fair; exception timing stays local |
| `spawn`, channels, Future handles, scheduler flush | BinaryTrees, Channel Rollup, Future Pipeline, Future Await Race, Await Channel Mux, Mutex Ledger, Mutex Await Journal | await-batch, future fanout/yield, channel pipeline, concurrent queue | broad concurrent application timing; retain scheduler policy controls locally |
| Mutex and cancellation | Mutex Ledger covers public Mutex/`with_lock`; Mutex Await Journal covers public `await_lock`/Awaitable; cancellation remains local | mutex counter, await-mutex, and cancellation/fairness fixtures | compare both Mutex acquisition styles; keep cancellation and scheduler-policy timing local |
| Packages, static imports, CLI entry/arguments, and canonical stdlib loading | every external application | package/config/import fixtures | broad for static packages; timing includes real launch paths |
| Dynamic packages/imports and interpreted metaprogramming | none by design | dynimport/dynamic-package fixtures | local semantic guard only; no synthetic foreign-runtime comparison |
| User-authored extern code and host conversion | applications exercise canonical stdlib host boundaries | inline-extern and host result/argument fixtures | local semantic/performance guard; target-language bodies are not portable benchmarks |

The local-only rows are not missing coverage. Their observable scheduling,
exception, dynamic-module, or target-language semantics do not have an honest
like-for-like Go/Python/Ruby program. A busy loop labeled with such a feature
would overfit the benchmark and weaken the performance policy.

## Catalog repair

`bench_external_catalog.sh` originally exposed `coverage` with every 22
application-shaped entry. It now contains all 30: the original 22, Future
Pipeline, Future Await Race, Await Channel Mux, Mutex Ledger, Mutex Await
Journal, Regex Suffix Audit, RegexSet Audit, and Regex Stream Audit.
`bench_bytecode_audit --suite corpus-full` and its
`generality-full` alias use this wider catalog, so a future opcode/lowering
candidate is checked against all existing application sources before it can be
kept. The stable `core`, `full`, and `generality` suite memberships are
unchanged.

The six previously omitted applications already had equivalent sibling sources
and verifiers. Regex Stream Audit is the one new program; it has the same
portable incremental-record behavior in every language. No compiler, VM,
runtime, `able-stdlib`, or benchmark algorithm changed.

## Cross-API regex follow-up

The Scanner/Suffix/RegexSet/I-Before-E profile gate is complete and selected
no source candidate. The three public APIs repeat existing generic tagged-NFA
work only within regex; the unrelated text control does not. The bytecode
commonality is a dispatcher parent, not a new concrete child. See
`2026-07-13-regex-stream-cross-api-profile-gate.md`.

## Next recommendation

The three-application Awaitable gate, fresh CPU-pinned scorecard, and compiled
generated-main follow-up are complete. The new compiled profiles again repeat
`bridge.currentGID` / `runtime.Stack` in Channel Rollup, Future Pipeline, and
Future Await Race, but the only known generic remedy remains disqualified by
independent N-body and K-Nucleotide guards. See
`2026-07-14-awaitable-cross-app-profile-gate.md`,
`2026-07-14-external-scorecard-async-refresh.md`, and
`2026-07-14-compiled-async-main-phase-refresh.md`.

The subsequent roadmap reconciliation found that there is no remaining v12
semantic boundary that can honestly add a portable timing row: parser, AOT,
stdlib, and regex gaps are complete, while dynamic modules and user externs
remain intentionally local-only. Do not invent a synthetic benchmark. The
next active product work is the WASM `able_host` stdout/stderr forwarding
slice, verified through Node and kept outside performance selection. See
`2026-07-14-feature-benchmark-roadmap-reconciliation.md`.
