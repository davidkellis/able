# Truthiness/cast architecture closure refresh

Date: 2026-07-21

## Decision

The `compiled-concurrency` and `bytecode-register-architecture` closures are
current against the corrected truthiness/cast semantics. Keep no compiler, VM,
runtime, canonical-stdlib, benchmark, verifier, reference, language, or WASM
change from this tranche.

Fresh repeated timing confirms large product gaps. Compiled concurrency does
reach the generated truthiness and cast bridges broadly enough to admit
profiling, but six normal-build profiles sample none of those bridges. They
instead reproduce the already rejected `bridge.currentGID`/`runtime.Stack`
owner. The six-family bytecode architecture selection reaches no corrected
Error-truthiness fallback; five rows have no cast reach and Fixed Width 128 has
no cast reach either. No new concrete shared leaf exists.

## Frozen contract

- The v12 spec SHA-256 is
  `4f0405b86c122993723e8617abd6f825d9a8ff858d4c72acaf4e33469452f080`.
- The canonical 70-file stdlib source state is
  `6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`
  under the checked scorecard source-state encoding. The closure ledger's
  independently encoded canonical-stdlib scope also reports no content drift.
- Every executable was built before its timed processes. Each process used the
  catalog directory, arguments, verifier, executor, logical CPU budget, and a
  55-second process cap.
- Arithmetic means retain every successful sample. Every row received a
  matched second cohort so that ten Able and ten limiting-reference samples
  contribute to each reported ratio.
- Compiled concurrency uses the source-equivalent four-CPU goroutine contract.
  Bytecode architecture rows use one CPU and their catalog executor policy.

## Repeated timing

| Application | Mode | Able samples | Able mean | Limiting reference | Reference samples | Reference mean | Ratio |
| --- | --- | ---: | ---: | --- | ---: | ---: | ---: |
| Await Channel Mux | compiled | 10 | 0.395 s | Go | 10 | 0.005899 s | 66.965x |
| Channel Rollup | compiled | 10 | 0.573 s | Go | 10 | 0.006638 s | 86.316x |
| Concurrent Event Routing | compiled | 10 | 3.039 s | Go | 10 | 0.005796 s | 524.288x |
| Concurrent Text Index | compiled | 10 | 1.041 s | Go | 10 | 0.006847 s | 152.032x |
| Dependency Wave Validation | compiled | 10 | 1.384 s | Go | 10 | 0.005880 s | 235.358x |
| Future Await Race | compiled | 10 | 0.093 s | Go | 10 | 0.004459 s | 20.855x |
| Future Pipeline | compiled | 10 | 0.386 s | Go | 10 | 0.006224 s | 62.023x |
| Mutex Await Journal | compiled | 10 | 0.748 s | Go | 10 | 0.004529 s | 165.155x |
| Mutex Ledger | compiled | 10 | 0.804 s | Go | 10 | 0.004587 s | 175.261x |
| Mutex Work Queue | compiled | 10 | 1.946 s | Go | 10 | 0.004169 s | 466.759x |
| Validated Job Pipeline | compiled | 10 | 3.077 s | Go | 10 | 0.005951 s | 517.017x |
| Array Slice Window | bytecode | 10 | 0.637 s | Python | 10 | 0.041699 s | 15.276x |
| Concurrent Event Routing | bytecode | 10 | 2.834 s | Python | 10 | 0.036600 s | 77.431x |
| Distance Field | bytecode | 10 | 5.552 s | Ruby | 10 | 0.354541 s | 15.660x |
| Fixed Width 128 | bytecode | 10 | 7.286 s | Python | 10 | 0.359779 s | 20.251x |
| Reverse Complement | bytecode | 10 | 3.260 s | Python | 10 | 0.028810 s | 113.154x |
| Word Frequency | bytecode | 10 | 1.386 s | Python | 10 | 0.021841 s | 63.459x |

All 400 timing processes represented by these decisions verified with zero
failures and zero timeouts: 220 compiled/Go processes and 180
bytecode/Python/Ruby processes. Pooled compiled Able CVs are 6.05%-19.37%; the
very short Go lanes are naturally noisier, up to 23.69%. Pooled bytecode Able
CVs are 2.34%-9.68%; their short limiting references reach 20.71%. Every
sample remains in the arithmetic mean, and every ratio is far from the 1.053x
ceiling corresponding to the 95%-throughput target.

## Exact reach and profile gate

Temporary opt-in generated counters were used only for one untimed, verified
compiled process per application and then removed. Seven of eleven compiled
applications enter generated semantic truthiness, for 29,728 calls in total.
Six enter the generated explicit-cast bridge, for 627 calls in total. The
largest row has 9,862 truthiness calls; Await Channel Mux has the largest cast
count at 512. The raw census is retained in
`2026-07-21-truthiness-cast-architecture-closure-compiled-reach.json`.

That breadth admitted two normal-build CPU profiles each for Await Channel
Mux, Concurrent Event Routing, and Mutex Work Queue. Across 20.84 seconds of
CPU samples, `bridge.IsTruthy`, `bridge.Cast`, `CastValueToType`, and generated
`__able_truthy` receive zero flat and zero cumulative samples. In contrast,
`bridge.currentGID` is 74.07% and 96.55% cumulative in Await Channel Mux,
95.83% and 96.82% in Event Routing, and 94.81% and 93.04% in Mutex Work Queue.
This reproduces the existing general concurrency owner and supplies no new
truthiness/cast candidate.

The bytecode architecture members already have deterministic post-fix exact
censuses from their freshly advanced family closures:

| Application | Primitive truthiness/process | Error fallback | Explicit casts |
| --- | ---: | ---: | ---: |
| Array Slice Window | 12,001 | 0 | 0 |
| Concurrent Event Routing | 541,058 | 0 | 0 |
| Distance Field | 0 | 0 | 0 |
| Fixed Width 128 | 2,000,000 | 0 | 0 |
| Reverse Complement | 4 | 0 | 0 |
| Word Frequency | 415,293 | 0 | 0 |

Thus the bytecode closure never reaches the corrected Error fallback or a cast
leaf. Reintroducing the same counters would reproduce deterministic counts
without changing candidate admission. The existing checked architecture model
also remains decisive: making all six transport opcodes free leaves every row
7.69x-82.15x short of target, and its executable register candidates already
failed broad guards.

## Exact artifacts

- Initial and second compiled Able cohorts:
  `2026-07-21-truthiness-cast-architecture-closure-compiled.json`
  (`65abef4c0c9d99883568194e6a6d1ea6ab5ce6f78b46bd3cda9bee216153eeff`),
  `2026-07-21-truthiness-cast-architecture-closure-c2-compiled.json`
  (`082b5b67f84ce9e8ef9504850876011bf32905dc7982cf433574e0e1ed96f143`).
- Initial and second Go cohorts:
  `2026-07-21-truthiness-cast-architecture-closure-go-reference.json`
  (`46d2a4677ad7cf26310877fa8f665776161afa844eeaa53f98db26cd8de175d`),
  `2026-07-21-truthiness-cast-architecture-closure-c2-go-reference.json`
  (`ebad8cf10d8bf5b0695c15d1b453da2439344ac874cdf41a209ce71380562051`).
- Initial and second bytecode Able cohorts:
  `2026-07-21-truthiness-cast-architecture-closure-bytecode.json`
  (`025a46cb0dd4e553076ee94e00062459849ed50b3918ee5c2e7e76cc0b2f5fb7`),
  `2026-07-21-truthiness-cast-architecture-closure-c2-bytecode.json`
  (`45c58971d626a093f4f827c8fc5a381259248f9c5518c144e1b9ed78beeb37f0`).
- Initial and second interpreter-reference cohorts:
  `2026-07-21-truthiness-cast-architecture-closure-interpreter-reference.json`
  (`4da2bd4ca5eda0963782cfd0636ce8c6a0822632683fc767c498a1f098c2f2b2`),
  `2026-07-21-truthiness-cast-architecture-closure-c2-interpreter-reference.json`
  (`340b07fc8f279c9c355eca875f2a0cc5282c980d28f2b09c42362dcdfcd93e5b`).
- Compiled exact reach:
  `2026-07-21-truthiness-cast-architecture-closure-compiled-reach.json`
  (`4d11aff19acbb3aa21983ff97916169f568417586a820f2e586e9c4a32bd4326`).

## Next recommendation

Refresh `compiled-architecture-target-budget` and
`cross-family-architecture-ownership` next.

Why: these are the two remaining broad invalidated closures. The first asks
whether any compiler-wide mechanism can plausibly close the 95%-of-Go budget;
the second checks whether one owner now crosses compiler and VM families. The
fresh family closures already provide most post-fix reach inputs, so resolving
them before the narrow single-program Sudoku quotient closure preserves the
project's generality rule.

What it entails: recompute both checked ownership/budget models against the
current production scopes and newly refreshed family evidence; verify exact
source identities and logical-work normalization; add repeated timing only if
a model input no longer has a current matched cohort; and profile only a newly
material three-family concrete leaf. Advance only those two closures and build
a candidate only if it survives the established unlike-application guards.
No WASM work is involved.
