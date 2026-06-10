# Compiled coverage scorecard and profile gate — 2026-07-14

## Decision

Keep no compiler, generated-runtime, bytecode VM, canonical-stdlib, or
benchmark-source change. The three new verifier-backed compiled rows are all
material misses relative to their equivalent Go programs, but their generated
`main` costs diverge. Their shared cold-start metadata decoding is already a
known compiler boundary with two rejected general replacements. It does not
authorize retrying either representation experiment or adding a map, iterator,
file, text, or benchmark fast path.

## Scorecard

Fresh Go 1.26.4 reference binaries ran on CPU 15 with `GOMAXPROCS=1`; current
Able compiled binaries used the same CPU lane. Each reference and Able row
used three processes and the external suite's Ruby verifier. Able builds used
the canonical external `able-stdlib`; no stdlib source changed.

| Application | Go real (s) | Able real (s) | Able / Go | Verification |
| --- | ---: | ---: | ---: | --- |
| Word Frequency | 0.0054 | 0.2100 | 38.89x | 3/3 verified |
| Document Audit | 0.0040 | 0.0800 | 20.00x | 3/3 verified |
| Lexical Rollup | 0.0044 | 0.1200 | 27.27x | 3/3 verified |

The machine-readable records are
`2026-07-14-compiled-coverage-go-refs.json` and
`2026-07-14-compiled-coverage-scorecard.json`. They are the timing authority;
the profiles below are attribution only.

## Bounded generated-main profiles

Each application was compiled once, then run from its exact external-suite
working directory with the CPU-only phase hook
`ABLE_GO_PHASE_CPU_PROFILE_DIR`, CPU 15, `GOMEMLIMIT=1GiB`, `GOGC=50`, and
`GOMAXPROCS=1`. The 24 Word Frequency, 60 Document Audit, and 48 Lexical
Rollup launches all passed their Ruby verifier and each application produced
one stdout hash. Generated source trees and binaries were removed immediately
after profile merging.

| Application | Merged main samples | Material work | Selection result |
| --- | ---: | --- | --- |
| Word Frequency | 3.52 s | `__able_hash_map_find_entry` 46.59% flat; `String_split` 36.36% cumulative | Text-key map lookup and split path only. |
| Document Audit | 60 ms | iterator generator/filter and `String.contains`; the short sample is sparse | No reliable shared descendant. |
| Lexical Rollup | 1.44 s | `fs_read_lines` 34.72% cumulative; allocation/GC and `growslice` 28.47%/21.53% cumulative | File-line materialization and allocation path only. |

The retained `main` and bootstrap artifacts are in
`v12/interpreters/go/.profiles/` under the
`20260714_compiled_coverage_` prefix. They are small diagnostic files and are
eligible for the project cleanup recipe.

## Shared bootstrap boundary is not a candidate

Document Audit and Lexical Rollup again reach
`interpreter.DecodeNodeJSON` under generated `RegisterIn` while reconstructing
definition/default metadata. The current bootstrap profile attributes 30 ms of
150 ms to that family for Document Audit and 70 ms of 120 ms for Lexical
Rollup. Word Frequency's bootstrap capture is too short to sample the decoder
reliably, but the earlier five-workload rebaseline recorded the same boundary
in every workload.

This does not reopen metadata decoding:

- A complete generated Go-constructor replacement made ordinary application
  builds spend more than six CPU minutes parsing the generated source.
- A compact tagged codec was neutral for Document Audit and Lexical Rollup and
  regressed the independent I-Before-E process guard by 2.49%.

Those are broad compiler-throughput and verifier-backed process-wall failures,
not benchmark-specific decisions. The repeated leaf is real but currently has
no eligible general implementation.

## Why no optimization landed

Word Frequency cannot justify a `HashMap` or string-key special case; the
compiler must continue using the shared nominal lowering path. Lexical
Rollup's read-lines allocation path does not recur in the map workload, and
Document Audit is too short to establish a stable common iterator leaf. The
small environment-switch sample in Lexical Rollup is also a previously
rejected fixed-context ABI family, not a new target.

## Next recommendation

Perform a semantics-first design audit of static compiled-program metadata
retention before writing another prototype. The scorecard shows that short
real applications are strongly affected by common bootstrap work, while the
two previously tried complete metadata encodings are disqualified. The audit
should inventory which emitted definition/default AST payloads are actually
needed by fully static binaries, which are required only by dynamic or
interpreter-fallback semantics, and what a compact shared representation would
need to preserve. It should add no code until it identifies a third general
design that retains diagnostics, defaults, dynamic behavior, and build
throughput. This is the only broad route exposed by these rows; tuning any one
application would improve a benchmark while risking ordinary programs.
