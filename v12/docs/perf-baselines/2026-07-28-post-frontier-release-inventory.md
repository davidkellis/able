# Post-frontier release inventory

Date: 2026-07-28

## Decision

The inventory of work accumulated after the preceding consolidation commit is
complete. No path is unmatched and no deferred WASM path entered the retained
boundary.

The maintainer explicitly authorized exact staging and one local
consolidation commit for the 181-path retained boundary. This authorization
does not include pushing, resetting, deleting, rewriting, or modifying the
34-path deferred WASM boundary.

## Snapshot identity

The inventory was captured before this record, its JSON companion, or its TSV
manifest existed:

- `HEAD`: `5bbe9e4594bfb2b7e7f8c84b9fbed9314c0ee8db`;
- `origin/master`: `5bbe9e4594bfb2b7e7f8c84b9fbed9314c0ee8db`;
- branch relation: zero commits ahead and zero behind;
- staged paths: zero;
- index SHA-256:
  `2d58f619060aeb960c342c2aaef79a0181439b121d5c78f443f83a3bdd73353e`;
- visible dirty paths: 212;
- tracked modifications: 74;
- untracked files: 138;
- deleted or renamed paths: zero; and
- tracked diff: 6,478 additions and 4,492 deletions.

The source NUL-delimited porcelain snapshot has SHA-256
`f355f1c5c876beab17060b6c5a96df81b1e01f35521724e4c2efd236a9c5ba07`.

The exact 212-row manifest is
`2026-07-28-post-frontier-release-inventory.tsv`. It records state, review
order, boundary, disposition, governing dated record, line count, byte count,
SHA-256, and path. It contains 51,897 bytes and has SHA-256
`13154fd551ada27f96b02538c0b321343672d13283bd1f585a63e9bda9f7ee66`.

The three inventory metadata files are intentionally not self-referential
manifest rows. The manifest preserves the immutable pre-record state; current
content identities must be refreshed before any later exact-index operation.

## Dependency-ordered review boundaries

| Order | Boundary | Tracked | Untracked | Paths | Disposition |
| ---: | --- | ---: | ---: | ---: | --- |
| 1 | benchmark contract implementation | 13 | 0 | 13 | retain |
| 2 | benchmark contract evidence | 4 | 104 | 108 | retain |
| 3 | bytecode profile coverage | 0 | 6 | 6 | retain |
| 4 | bytecode semantic-boundary closure | 4 | 2 | 6 | retain |
| 5 | derived architecture reconciliation | 37 | 0 | 37 | retain |
| 6 | compiler release-readiness | 5 | 0 | 5 | retain |
| 7 | handoff and release tooling | 3 | 0 | 3 | retain |
| hold | deferred WASM | 8 | 26 | 34 | hold outside release boundary |

The pre-record retained boundary is exactly 178 paths: 66 tracked
modifications and 112 untracked files. It contains 133,064 lines and
7,633,794 bytes. Its sorted newline-delimited path-list SHA-256 is
`f85b12851614a5f716463feab8d719ee53889beda76f099e27e0c670630b5b44`.

The deferred boundary is exactly the same 34-path set as the preceding
post-consolidation inventory. Its current path-list SHA-256 is
`7f29de27f03da83138df19696be182a5139aa6536ba0bdb363a1a23ffafd1f4e`.
There are no generated-local or unmatched manifest paths.

## Dated-record coverage

The seven new chronological entries in both `LOG.md` and `v12/LOG.md` contain
the same seven unique dated record references, and every record exists.
Per-path governing-record coverage is:

| Governing record | Paths |
| --- | ---: |
| bytecode unranked-coverage admission search | 4 |
| bytecode portable workload-scale calibration | 20 |
| mode-aware benchmark contract closure | 97 |
| bytecode current-source profile coverage closure | 6 |
| bytecode semantic-boundary reach closure | 6 |
| v12 correctness and release-readiness closure | 45 |
| deferred WASM hold | 34 |

Review order reflects dependencies:

1. The benchmark applications and general mode-aware runner contract establish
   which workloads and inputs are executable.
2. The 126-row scorecard, selected cohorts, calibration, and stability
   evidence depend on that contract.
3. Current-source CPU/allocation coverage depends on the selected bytecode
   rows.
4. The semantic-boundary census and closure ledger depend on that coverage.
5. Cross-engine budgets and feasibility records derive from the reconciled
   frontier and closure ledger.
6. Compiler timing repairs and the 640-partition release matrix validate the
   resulting release surface.
7. The plan and logs describe the verified state and next authority boundary.

This sequencing prevents a reviewer from evaluating generated evidence before
the sources, selection contract, and causal closures that define it.

## Validation

- All 212 manifest line, byte, and SHA-256 identities reproduce exactly.
- All 75 dirty JSON files parse.
- All 14 dirty Go files pass `gofmt -l`.
- `git diff --check` passes.
- None of 51 dirty maintained Go, Python, shell, JavaScript, or Able source
  paths reaches 1,000 lines. The largest is
  `v12/bench_external_catalog.sh` at 997 lines.
- The 132 dirty performance-evidence paths total 3,168,352 bytes.
- No v10, v11, deprecated in-tree stdlib, or external `able-stdlib` path is
  present.
- No secret-like filename or common private-key/service-token signature is
  present.
- The deferred WASM path set exactly matches the preceding inventory.
- The index remains empty with the same SHA-256 captured above.

The project cleanup dry run reports 4.01 GiB across the ignored
`v12/tmp`, `v12/interpreters/go/.gocache`, and
`v12/interpreters/go/tmp` generated paths. They were identified but not
deleted because this tranche is explicitly non-mutating. No manifest path was
created, modified, or removed by the cleanup check.

No compiler production path, runtime, interpreter, VM, canonical stdlib,
language, dependency, benchmark behavior, fixture, reference implementation,
or WASM behavior changed during this inventory.

## Exact post-record candidate

If a maintainer later authorizes an exact local consolidation, the candidate
is the sorted union of:

- all 178 pre-record paths marked `retain`; and
- this Markdown record, its JSON companion, and the 212-row TSV manifest.

That candidate contains exactly 181 paths. Its sorted newline-delimited
path-list SHA-256 is
`21a3aa4c94d62cfeaa3da7b87a7fe354b253800c299233cb9b8ca0abc2c32510`.
The complement must remain exactly the 34 deferred WASM paths.

The exact NUL-delimited pathspec has SHA-256
`fd41801a0261058210ab24e5c31660ea1ea59ed990b1208a8edde1056a8596f8`.
No broad `git add` pathspec is used.

After the final handoff edits, the refreshed 178 non-self identity rows contain
133,167 lines and 7,639,403 bytes. Their 27,370-byte
line/byte/content/path manifest has SHA-256
`0a9635e0b0e99f740b9cb803167dd0cd5627af14bf2e8c0b4a90b833eb1f0a2a`.

The exact candidate and deferred complement were revalidated immediately
before staging. Cached identity, JSON, Go formatting, whitespace, source-size,
scope, common-secret, evidence, and path-complement checks passed before the
one authorized local commit. No push occurred.

## Recommendation

Next obtain explicit maintainer authorization before publishing the local
branch.

Why: the exact retained boundary is consolidated locally, but pushing changes
external repository state and was not included in the local-commit
authorization.

What it entails: inspect the final commit and branch divergence, confirm the
remote destination and exact commit range, and push only after separate
explicit authorization.

Why it is important: this prevents publishing inactive WASM or an unintended
commit range while preserving the verified native-carrier and interpreter-free
state. Production performance mutation remains paused until a checked evidence
invalidation or a new exact open owner reaches three unlike application
families.
