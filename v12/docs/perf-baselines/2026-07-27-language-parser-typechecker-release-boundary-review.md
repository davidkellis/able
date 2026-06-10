# Language, parser, and typechecker release-boundary review

Date: 2026-07-27

## Decision

The current language/AST/parser/typechecker worktree delta is completely
classified and passes its focused and repository-wide release gates. Treat
the language-owned paths as an audited final-state boundary for the next
dependency-ordered review.

Do not stage or commit this boundary yet:

- `spec/full_spec_v12.md` and `spec/TODO_v12.md` also govern runtime,
  compiler, standard-library, and deferred work, so they remain
  `hold-cross-boundary`;
- `parser.c`, `grammar.json`, and `node-types.json` are generated artifacts
  and must be reproduced from the grammar rather than reviewed as independent
  handwritten source; and
- no maintainer authorization to stage, commit, or rewrite history was given.

This tranche changed no language, AST, parser, typechecker, runtime,
interpreter, VM, compiler, standard-library, dependency, benchmark, reference,
or WASM behavior. It added only this review record, its JSON summary and TSV
manifest, and updated forward-only handoff documentation.

## Deterministic inventory

The reviewed boundary is the union of current tracked modifications and
visible untracked files under:

- `spec/full_spec_v12.md` and `spec/TODO_v12.md`;
- `v12/fixtures/ast/**`;
- `v12/parser/**`;
- `v12/interpreters/go/pkg/ast/**`;
- `v12/interpreters/go/pkg/parser/**`; and
- `v12/interpreters/go/pkg/typechecker/**`.

The deterministic manifest contains one header plus 82 sorted path rows:

| State | Paths |
| --- | ---: |
| Tracked modified | 54 |
| Untracked | 28 |
| **Total** | **82** |

| Family | Paths |
| --- | ---: |
| AST fixtures | 7 |
| Go AST | 7 |
| Go parser | 25 |
| Go typechecker | 26 |
| Grammar source | 7 |
| Grammar corpus | 3 |
| Parser documentation | 2 |
| Generated parser | 3 |
| Shared specification contract | 2 |
| **Total** | **82** |

Seventy-seven paths are reviewed source, tests, fixtures, or documentation;
three are regenerate-only parser assets; two are held shared contracts. There
are zero unclassified paths.

The manifest is
`2026-07-27-language-parser-typechecker-release-boundary-manifest.tsv`, has
83 lines and 13,008 bytes, and has SHA-256
`5200fc5d8e698b828213f20659209d71caa52e4f71da468636979f3337d3482a`.
Each row records state, family, disposition, line count, byte count, file
SHA-256, and path.

The earlier worktree inventory contained 76 paths in this boundary. The six
additional paths are the four grammar modules created by the retained grammar
split plus later focused guards already present in the final worktree. The
current 82-path manifest, not the earlier planning count, is authoritative for
this boundary.

## Ownership and evidence

### Specification contract

The two spec files contain language text and forward status spanning parser,
runtime, compiler, standard-library, and deferred areas. Relevant reviewed
contracts include contextual integer-literal adoption and widening, explicit
interface casts, source re-export identity, and named-implementation import
collisions. Their parser/typechecker portions agree with the implementation,
but the files cannot safely be assigned to a parser-only release unit.

### AST and source syntax

The AST delta adds explicit export statements and module exports, shared AST
walking, and origin propagation through that walker. The parser, AST fixtures,
tree-sitter grammar/corpus, and typechecker agree on source re-exports and
named implementations.

The governing evidence is:

- the 2026-07-14 source re-export and named-implementation entries in
  `LOG.md` and `v12/LOG.md`;
- `v12/design/reexport-named-implementation-import-audit.md`;
- `2026-07-20-origin-skip-known-gate.md`; and
- `2026-07-27-tree-sitter-grammar-modularization-retained.md`.

### Parser implementation

The final parser state includes general startup allocation reuse, cached
tree-sitter node kinds and field IDs, reduced redundant span work, phase
observation used by measurements, and the source-syntax support above. These
are general parser rules, not benchmark or named-container special cases.

The retained evidence is:

- `2026-07-20-cross-mode-startup-allocator-gate.md`;
- `2026-07-20-parser-node-kind-cache-gate.md`;
- `2026-07-20-parser-field-id-cache-gate.md`;
- `2026-07-20-parser-redundant-span-dispatch-gate.md`;
- `2026-07-20-parser-residual-span-caller-pair-gate.md`; and
- `2026-07-27-tree-sitter-grammar-modularization-retained.md`.

Regeneration reproduced the generated parser assets at these identities:

| Asset | SHA-256 |
| --- | --- |
| `src/parser.c` | `054369173a78160f15ebef87bced7c4717a4aaf85cfb04ded76eb6ab1ed4c6e3` |
| `src/grammar.json` | `c9ed2e44a74e11f3a333ad6a9df51ce24696bd84f4314ce60c8069e048088b08` |
| `src/node-types.json` | `07e23952fdcb7e586c116b28bc50b8927865f7354825e4d69779b95d677d20b1` |

### Typechecker

The reviewed typechecker delta covers source re-export surfaces and collision
diagnostics, explicit interface-cast typing, contextual integer-literal
adoption and fixed-width bounds, method-selection provenance, forward
declaration hydration, and the narrow
`PrepareProgramForEvaluation` path used by execution.

The source re-export behavior is governed by the July 14 log entries and
design audit. Generic-union method provenance is covered by
`2026-07-22-bytecode-typechecked-generic-union-call-gate.md`. Program
preparation, integer behavior, casts, and forward hydration are governed by
the specification, chronological logs, focused tests, and
`v12/design/typechecker-plan.md`; they are log-governed final-state work, not
unowned production edits.

Two unchanged base files, `self_type_patterns.go` and `decls_types.go`, are not
`gofmt`-canonical. They do not appear in the 82-path changed-state manifest and
were not modified by this non-mutating review. This inherited formatting debt
is not assigned to the current boundary and should be addressed only if those
files enter an authorized source-hygiene change.

## Verification

Focused verification passed:

- `gofmt -l` reported no changed or untracked Go path in the boundary;
- every reviewed handwritten code file remains below 1,000 lines;
- `go vet ./pkg/ast ./pkg/parser ./pkg/typechecker`;
- `go test ./pkg/ast ./pkg/parser ./pkg/typechecker -count=1 -timeout 55s`;
- focused cast, integer, method-selection, re-export,
  named-implementation-collision, and forward-struct typechecker guards;
- focused `cmd/able` and interpreter integration guards;
- tree-sitter regeneration with byte-identical generated hashes;
- all 28 forced native tree-sitter corpus cases;
- `./v12/export_fixtures.sh --check imports/source_reexport_syntax`;
- a forced Go parser relink with a fresh disk-backed Go cache; and
- `git diff --check`.

The required complete `./run_all_tests.sh` handoff passed in 11:34.57 at
4,633,264 KB peak RSS. All non-compiler packages, all 33 bounded compiler
batches, and the final 85.244-second bytecode fixture pass were green. The two
known long aggregate compiler batches completed in 134.184 and 69.186 seconds;
the prior per-test timing record remains authoritative that no individual test
exceeds one minute.

Canonical standard-library modes were not rerun because this tranche made no
production or contract change. The immediately preceding grammar tranche
already passed both modes in 53.04 seconds, reporting 17 seconds for
tree-walker and 15 seconds for bytecode.

Cleanup removed 140 KiB of reproducible project-local Python cache and an
empty generated Go temporary directory. It also removed the tranche-specific
72 MiB forced-relink cache, empty full-suite temporary directory, temporary
manifest source, and tree-sitter generation log under `/var/tmp`. The guarded
project cleanup now reports no generated candidates. Reusable disk-backed Go
and compiled-test caches remain.

## Recommendation

Next perform the second non-mutating release-consolidation review boundary:
runtime carriers, the tree-walking interpreter, the bytecode VM, execution
fixtures, runtime support, and their CLI wrappers.

Why: these components are the first consumers of the now-audited language,
AST, parser, and typechecker state. Reviewing them next preserves dependency
order and can expose any boxing, native-carrier, or compiled/interpreted
boundary inconsistency before the compiler/AOT boundary is considered.

What it entails: deterministically re-inventory the current runtime/VM path
set; map every production path and guard to retained records; separate shared
runtime contracts, generated/local state, and deferred WASM; verify both
tree-walker and bytecode semantics with focused gates; and emit a manifest and
decision record without staging, committing, resetting, or rewriting history.

Why it is important: it proves that the reference execution engines implement
the reviewed contract and gives the later compiler/AOT review a stable,
auditable semantic boundary. Keep performance mutation paused unless that
review discovers a correctness failure or a new exact owner meeting the
three-unlike-application evidence gate. Do not begin WASM work.
