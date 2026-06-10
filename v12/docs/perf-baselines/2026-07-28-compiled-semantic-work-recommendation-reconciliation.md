# Compiled semantic-work recommendation reconciliation

Date: 2026-07-28

## Decision

Retain no production change.

The recommendation to audit TapeLang Alphabet, Sudoku Masks, and N-Body was
stale. The first two applications had already received a complete
semantic-work normalization and generated-source/assembly audit on
2026-07-26. The subsequent corpus-wide relational safety observer had already
included N-Body and established that its material fixed-length Array proof did
not repeat in three unlike applications.

Fresh current strict artifacts confirm that later compiler work did not
invalidate those decisions:

- normalized Able and Go sources are byte-identical to the audited sources;
- all five refreshed graphs remain interpreter-free;
- the current relational observer reproduces the decisive TapeLang, Sudoku,
  and N-Body proof counts;
- current hot assembly preserves the earlier calls and branches; and
- Matrix Multiply and Pidigits remain current repeated target controls.

No compiler, generated-runtime, runtime, interpreter, bytecode VM,
canonical-stdlib, benchmark, reference, language, dependency,
non-primitive-nominal, or WASM production source changed. No A/B candidate was
admitted because no new general operation is material in three unlike
applications.

## Prior closure

The relevant completed chain is:

1. `2026-07-26-compiled-semantic-work-equivalence-closure.md` identified
   unequal TapeLang and Sudoku source work.
2. `2026-07-26-benchmark-semantic-work-equivalence-normalized.md` made both
   language pairs structurally comparable.
3. `2026-07-26-normalized-compiled-generated-owner-closure.md` found no
   removable generated owner shared by normalized TapeLang, normalized
   Sudoku, and Fib.
4. `2026-07-26-compiled-relational-safety-proof-census-closure.md` proved
   current bounds conditions across the strict corpus and found dynamic
   materiality only in N-Body.
5. `2026-07-26-performance-frontier-reconciliation.md` explicitly removed
   semantic-work re-auditing from the frontier unless a concrete invalidation
   appeared.

The 2026-07-28 broad owner refresh accidentally recommended the completed
semantic-work route again. This record corrects that handoff.

## Current artifact contract

Five applications were freshly emitted with:

```text
--no-fallbacks --experimental-execution-context
```

The compiler SHA-256 was
`55c7e9cb6f911406510b67a9f09ccece55cc4cb9a111cf8ceebb234adfc13871`.
One smoke process per application was used only to verify fresh artifacts and
public outputs. It was not used as performance-selection evidence. Current
performance remains the five-run 2026-07-28 broad scorecard:

| Application | Able mean | Go mean | Ratio | Role |
| --- | ---: | ---: | ---: | --- |
| TapeLang Alphabet | 3.9660s | 3.0651s | 1.2939x | normalized miss |
| Sudoku Masks | 1.5620s | 0.6740s | 2.3175x | normalized miss |
| N-Body | 0.0820s | 0.0350s | 2.3429x | unlike miss |
| Matrix Multiply | 1.0340s | 0.9870s | 1.0476x | target control |
| Pidigits | 1.2200s | 1.1864s | 1.0283x | target control |

All five smoke outputs match those scorecard output hashes. All generated
dependency graphs omit `pkg/interpreter`.

The complete source, reference, output, generated-source, binary, and compiler
hash summary is
`2026-07-28-compiled-semantic-work-recommendation-reconciliation-source-artifact-summary.tsv`
with SHA-256
`8273ef8d24bb8b04f07038278b53621deb8a2aff1cf1fac8c76a5925d876368f`.

TapeLang and Sudoku retain the exact normalized source identities:

| Source | Current and normalized SHA-256 |
| --- | --- |
| Able TapeLang | `426a40e33840f3a0e9e62d5f9b9519a6840edd2733b8031df6a280e4b782fdb8` |
| Go TapeLang | `e53f36775b30177debe262c6693c40ff0c3db9a552fcf2dd8cb6c5a37174f775` |
| Able Sudoku | `fd918bf940bf22973fb205b1f1fa1be9de07a8774ad79138c9bc8a95acb43d6b` |
| Go Sudoku | `0731d83377fde41add93bd6f503a7b42b9f612b8c1a07a2d3ec440e301b7ee99` |

Their exact dynamic operation counts therefore remain authoritative:
TapeLang performs 1,213,347,694 equivalent flat dispatches; Sudoku performs
1,918,450 solve/find calls, 155,394,450 cell scans, 64,090,010 empty-cell
evaluations, and 166,923,250 bit-count iterations.

## Current relational proof result

The retained report-only `compiled-control-census` observer was rerun over
each exact generated module:

| Application | Reachable Array checks | Proven safe | Call-site specializable functions |
| --- | ---: | ---: | ---: |
| TapeLang Alphabet | 13 | 0 | 0 |
| Sudoku Masks | 32 | 4 | 1 |
| N-Body | 52 | 32 | 0 |
| Matrix Multiply | 8 | 0 | 0 |
| Pidigits | 2 | 0 | 0 |

The summary is preserved in
`2026-07-28-compiled-semantic-work-recommendation-reconciliation-control-summary.tsv`
with SHA-256
`cc55812f7aec1604fed51146ba2233e81cd8a5adbf542f1d64fdc7f2739fd452`.
Selected hot-function facts are in the companion hot-control TSV with
SHA-256
`4470c8d3bf2c77ee6bbd790f4e1916285c0c02d520cf34dcf1dcada863aa1a57`.

These counts reproduce the July 26 stopping condition:

- N-Body's 32 proofs are the same fixed-five-element Array facts that execute
  22,500,243 times. They are dynamically material in N-Body only.
- Sudoku's four safe bounds checks remain clue-loading initialization sites,
  not its hot recursive search. The current observer additionally classifies
  `square_index_ctx` as call-site specializable for the closed benchmark
  graph, but that hot arithmetic proof remains Sudoku-only.
- TapeLang, Matrix, and Pidigits expose no new safe hot bounds family.

No one structural proof has material dynamic reach in three unlike
applications.

## Current assembly reconciliation

The experimental execution-context ABI changes whole generated artifacts, so
whole-file SHA identity alone cannot determine whether the earlier hot
analysis remains valid. Current exact-symbol disassembly supplies that check.

Compared with the normalized July 26 closure:

| Function | Current instructions | Change | Calls | Conditional / unconditional jumps |
| --- | ---: | ---: | ---: | ---: |
| TapeLang `execute_ctx` | 283 | +7 | 17 | 34 / 19 |
| TapeLang `Tape.inc_ctx` | 84 | +2 | 4 | 10 / 6 |
| TapeLang `Tape.move_ctx` | 69 | +3 | 4 | 5 / 3 |
| Sudoku `find_best_empty_ctx` | 205 | +7 | 10 | 13 / 5 |
| Sudoku `bit_count_ctx` | 34 | +1 | 2 | 3 / 2 |
| Sudoku `square_index_ctx` | 77 | +3 | 5 | 5 / 1 |
| Sudoku `solve_with_masks_ctx` | 348 | 0 | 24 | 23 / 6 |

Every listed call and jump count is identical to the prior normalized
assembly. The additional instructions are straight-line ABI/context
bookkeeping, not new semantic work or a repeated dynamic boundary.

Matrix's current `matmul_ctx` contains 399 instructions, one more than its
earlier 398-instruction form. Pidigits `next_term_ctx` contains 156, three
more than its earlier 153-instruction form. Both remain target controls, which
rules out treating these tiny static deltas as a general performance owner.

Current N-Body assembly remains a distinct bounds-heavy numeric family:
`advance_ctx`, `energy_ctx`, and `offset_momentum_ctx` contain 459, 273, and
180 instructions respectively. The current observer, rather than aggregate
function size, identifies its already-closed fixed-length proof opportunity.

All current symbol sizes, instruction counts, calls, jumps, and disassembly
hashes are preserved in
`2026-07-28-compiled-semantic-work-recommendation-reconciliation-assembly-summary.tsv`
with SHA-256
`f3f95243e0b00269770f71f765830f6c9f0d842a67485abbbf82d499adf0a5cc`.

## Candidate gate

| Candidate | Current material reach | Disposition |
| --- | ---: | --- |
| fixed-length Array bounds proof | N-Body only | closed below three unlike applications |
| closed-call `square_index` arithmetic proof | Sudoku only | one-family proof |
| checked tape mutations | TapeLang only | required for unrestricted inputs |
| total-function control ABI | no new three-way reach | prior Fib-only route; Fib now meets the noisy current target |
| execution-context straight-line delta | small static delta in several functions | no material repeated owner; Matrix/Pidigits controls also contain deltas |
| generic Array/result/check removal | different descendants | unsafe without the distinct proofs above |

No candidate entered implementation or the five-or-more-pair A/B/Go gate.
Grouping these distinct facts as “checks,” “native code,” or “function size”
would erase the proof and materiality requirements.

## Verification and cleanup

- Five fresh strict builds and smoke processes passed public verification.
- Five final generated dependency graphs omit `pkg/interpreter`.
- Source hashes match the normalized/equivalent contracts.
- The current control observer reproduces the prior proof-family stopping
  condition.
- `go test ./cmd/compiled-control-census -count=1 -timeout=60s`,
  `go test ./cmd/ablec -count=1 -timeout=60s`, selection-manifest checks, and
  the 21-closure evidence ledger pass.
- Compact evidence was retained before the exact disposable disk-backed
  469 MiB `/var/tmp/able-semantic-reconciliation-20260728-VFOSjB` workspace
  and its pointer were deleted.

## Next recommendation

Pause production performance mutation until a concrete admission invalidation
exists, and use the next tranche for current correctness and release-readiness
gates.

Why: the current broad refresh found no open shared owner, and this
reconciliation proves that its proposed semantic-work follow-up was already
completed and remains valid under the current compiler. Another local
performance experiment would target one application, an aggregate parent, or
required Able semantics.

What it entails: run the ordinary v12 correctness, stdlib, fixture,
dependency, and release checks; repair any real failure through shared
semantics; and refresh only performance rows whose execution ownership is
actually changed. Reopen performance production work only for a new broad
application, a retained compiler/runtime/language/stdlib change, a correctness
failure that invalidates a carrier, or a report-only observer finding one
exact open owner material in three unlike applications.

Why it is important: an evidence-triggered pause preserves the
interpreter-free compiled architecture and native primitive carriers while
preventing benchmark-specific changes and noisy regressions. The 95%-of-Go
goal remains active; the pause defines what new evidence is required to make
the next implementation scientifically actionable. Do not begin WASM work.
