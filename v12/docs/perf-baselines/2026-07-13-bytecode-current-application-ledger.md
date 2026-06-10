# Current bytecode application ledger — 2026-07-13

## Purpose

This is the provenance-preserving 28-application bytecode status ledger. It
joins the fresh 16-program `generality` screen with twelve fresh extension
applications, including Future Pipeline, Future Await Race, Await Channel
Mux, Regex Suffix Audit, and RegexSet Audit. It is a provenance ledger, not a
synthetic aggregate: each row keeps the source report's process count, timeout
status,
verifier result, and CPU-15/1-GiB guard. No ratio is inferred from a timeout
or unavailable foreign reference.

The bytecode target is Able/reference no greater than `1.0526x`, meaning Able
executes at least 95% as fast as the corresponding Python or Ruby program.
The stable scorecard used one verifier-backed Able process per row as a bounded
status screen; the extension scorecard used three. Both use current canonical
stdlib, Python 3.14.5, Ruby 4.0.5, `GOMEMLIMIT=1GiB`, `GOGC=50`,
`GOMAXPROCS=1`, CPU 15, and a 45-second per-process guard.

| Application | Bytecode (s) | Able/Python | Able/Ruby | Status and provenance |
| --- | ---: | ---: | ---: | --- |
| Fib | 0.1600 | n/a | n/a | verified 1/1; foreign references cap-bound |
| BinaryTrees | n/a | n/a | n/a | bytecode cap-bound |
| MatrixMultiply | 4.4800 | n/a | n/a | verified 1/1; foreign references cap-bound |
| QuickSort | n/a | n/a | n/a | bytecode cap-bound |
| Sudoku | n/a | n/a | n/a | bytecode cap-bound |
| Sudoku Masks | n/a | n/a | n/a | bytecode cap-bound |
| I-Before-E | 0.5500 | 5.40x | 4.20x | verified 1/1, miss |
| Base64 | 4.3500 | 1.10x | 1.74x | verified 1/1, miss |
| JSON | 1.2000 | 0.46x | 0.69x | verified 1/1, meets both |
| Monte Carlo Pi | 2.6800 | 1.74x | 1.61x | verified 1/1, miss |
| PiDigits | 2.7100 | 0.67x | 0.26x | verified 1/1, meets both |
| Mandelbrot | 6.7900 | 5.22x | 3.48x | verified 1/1, miss |
| Reverse Complement | 6.7500 | 160.71x | 80.26x | verified 1/1, miss |
| K-Nucleotide | n/a | n/a | n/a | bytecode cap-bound |
| N-body | n/a | n/a | n/a | bytecode cap-bound |
| TapeLang Alphabet | n/a | n/a | n/a | bytecode cap-bound |
| Fixed Width 128 | 10.1600 | 27.50x | 14.17x | verified 3/3, miss |
| Rational Series | 4.9567 | 46.37x | 35.08x | verified 3/3, miss |
| Word Frequency | 1.7500 | 89.74x | 29.51x | verified 3/3, miss |
| Document Audit | 0.3667 | 25.12x | 8.75x | verified 3/3, miss |
| Lexical Rollup | 0.5000 | 26.88x | 9.54x | verified 3/3, miss |
| Channel Rollup | 0.5633 | 13.64x | 9.93x | verified 3/3, miss |
| Future Pipeline | 0.5267 | 8.51x | 7.39x | verified 3/3, miss |
| Future Await Race | 0.1700 | 5.50x | 3.36x | verified 3/3, miss |
| Await Channel Mux | 0.2700 | 2.26x | 2.74x | verified 3/3, miss |
| Regex Suffix Audit | n/a | n/a | n/a | bytecode cap-bound; 3/3 at 45 s |
| RegexSet Audit | 8.5133 | 411.27x | 185.47x | verified 3/3, miss |
| Regex Stream Audit | 6.4167 | 317.66x | 128.59x | verified 3/3, miss |

The generated rows remain in
`2026-07-13-bytecode-generality-{interpreter-refresh,scorecard}.*` and
`2026-07-13-bytecode-coverage-extension-{interpreter-refresh,scorecard}.*`.
Future Await Race's independently generated pinned rows and profile decision
are recorded in `2026-07-13-future-await-race-reference-profile-gate.md`.
Await Channel Mux's equivalent record is
`2026-07-13-await-channel-mux-reference-profile-gate.md`. Regex's fresh
coverage/reference records are
`2026-07-13-regex-application-coverage-refresh.md` and
`2026-07-13-regex-stream-application-coverage.md`; the completed cross-API
selection record is `2026-07-13-regex-stream-cross-api-profile-gate.md`.

## What the current ledger establishes

Eighteen applications have completed bytecode/Python/Ruby comparisons. JSON
and PiDigits are the only two that meet both 95% targets. The other sixteen
are explicit verifier-backed misses. Eight applications remain bytecode
cap-bound, and Fib and MatrixMultiply have completed Able rows but no rankable
Python/Ruby reference under the shared guard.

The result makes the distance to the interpreter target visible across the
whole current application corpus, but it does not identify a common leaf. The
profile and generality gates already separate the misses into:

- raw-float/control-flow, host codec, BigInt, and byte/Array/boxing work;
- checked multiword and rational nominal arithmetic;
- map/text, generator/iterator/pattern, and filesystem/public-call work; and
- distinct text-channel and numeric-Future concurrency children.

Their common VM, executor, call, and allocation frames are parents. The
concrete generic float, member-cache, inline-return, raw-cell, map/boxing,
and scheduler directions have either been shown to be individual workload
paths or have already failed broad guard runs. A 28-row status ledger is not
permission to retest them.

## Decision and next recommendation

Keep no bytecode VM, compiler, bridge, runtime, canonical-stdlib, or benchmark
source change. The ledger closes the last split in the application-level
selection evidence: future VM work must start with a new semantic boundary or
a repeatable new concrete descendant, not another ratio or unchanged profile.

The recommended next tranche is language/runtime feature completion selected
from the active v12 roadmap, with its behavior added to the shared fixture and
application matrix before it receives performance attention. This is necessary
because all current application families now have coverage and the remaining
known VM hotspots are either disjoint or rejected; a newly implemented shared
semantic boundary is the only credible way to uncover a broadly useful,
previously unmeasured execution cost.

Future Await Race is the 24th row. It adds repeated Future-await joins to
the external `concurrency` and `coverage` catalogs while retaining a
cooperative cancellation probe. Its fresh pinned references and normal-
process profiles confirm a material miss but no new shared VM leaf; it does
not reopen rejected scheduler, raw-cell, frame, or fixed-context experiments.

Await Channel Mux is the 25th row. It covers public `Channel.await_receive`
arms plus `Await.default`, a different Awaitable boundary from Future Await
Race. Its three-run pinned screen is a material miss, but eight
normal-process captures add only loader/parser and generic VM/executor parents
to the bytecode evidence. Compiled capture independently repeats the
already-rejected `bridge.currentGID`/`runtime.Stack` wall. It therefore does
not reopen scheduler, raw-cell, frame, or fixed-context experiments. The
compiler/bridge correctness repair and profile decision are recorded in
`2026-07-13-await-channel-mux-{coverage,reference-profile-gate}.md`.

Regex Suffix Audit, RegexSet Audit, and Regex Stream Audit are rows 26 through
28. Together they cover public `RegexBuilder`, tagged-NFA matching, public
`RegexSet` classification, and chunked `RegexScanner` finalization through the
existing canonical stdlib. Their pinned screens add one bytecode cap and two
material verified misses. This does not reopen a VM candidate: the existing
suffix, ordinary-match, and RegexSet profile gates already identified and kept
general NFA traversal and thread-reuse improvements in the canonical stdlib,
while rejecting API- or corpus-specific paths. Scanner establishes product
coverage, not a scanner-specific fast path. The fresh reference/coverage
records are `2026-07-13-regex-application-coverage-refresh.md` and
`2026-07-13-regex-stream-application-coverage.md`. The completed cross-API
profile gate keeps no source change: only already-addressed NFA work repeats
within regex, while the unlike control shares no new concrete VM descendant.
