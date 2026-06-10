# Tree-sitter grammar modularization retained

Date: 2026-07-27

## Decision

Retain the behavior-preserving tree-sitter grammar modularization and the
corpus-expectation correction found while verifying the current parser.

The 1,650-line handwritten `grammar.js` was the final known active handwritten
source file above the project's 1,000-line limit. Its rules now live in
cohesive CommonJS modules, assembled in their original order by an 86-line
entry point:

| File | Lines | Responsibility |
| --- | ---: | --- |
| `grammar.js` | 86 | Grammar metadata, conflicts, and ordered rule assembly |
| `grammar/helpers.js` | 130 | Precedence, separators, line operators, and literal patterns |
| `grammar/declarations.js` | 420 | Module, declaration, statement, and control-flow rules |
| `grammar/expressions.js` | 650 | Expression precedence, calls, lambdas, and iterators |
| `grammar/literals-patterns-types.js` | 418 | Literals, patterns, types, and lexical rules |

`package.json` now publishes `grammar/**` with `grammar.js`; an
`npm pack --dry-run --ignore-scripts` inventory confirmed that all four
modules are present. No dependency changed.

## Generated-language identity

Before extraction, `tree-sitter generate` reproduced the existing generated
assets exactly. After extraction, regeneration produced the same hashes:

| Artifact | SHA-256 before and after |
| --- | --- |
| `src/parser.c` | `054369173a78160f15ebef87bced7c4717a4aaf85cfb04ded76eb6ab1ed4c6e3` |
| `src/grammar.json` | `c9ed2e44a74e11f3a333ad6a9df51ce24696bd84f4314ce60c8069e048088b08` |
| `src/node-types.json` | `07e23952fdcb7e586c116b28bc50b8927865f7354825e4d69779b95d677d20b1` |

The split therefore changes neither the exported grammar nor the generated
concrete-syntax/AST contract. Generated `parser.c` remains the intentional
generated-source exception to the line limit.

## Corpus correction

The first complete corpus run exposed one stale expectation in
`expressions_partials.txt`: the checked-in tree omitted the
`expression_list`/`expression_statement` wrappers for a single-expression
inline lambda, while the already-current generated parser emitted the
declared CST shape.

Because the generated parser was byte-identical before and after the source
split, this mismatch predated the modularization. Only that expected tree was
mechanically refreshed. No source input, grammar rule, generated artifact, Go
AST mapping, or language behavior changed. A forced native parser rebuild then
passed all 28 corpus cases.

## Verification

- `node --check` passed for the entry point and all four modules.
- `tree-sitter generate` retained all three generated hashes exactly.
- The forced native tree-sitter corpus passed 28/28.
- The package dry-run included every required grammar module.
- The forced Go parser relink passed `go test -a ./pkg/parser -count=1
  -timeout 55s`; the reported package test time was 0.150 seconds.
- `git diff --check` passed.
- The active handwritten Go/JavaScript scan found no file at or above 1,000
  lines after excluding dependencies, generated build workspaces, and
  generated `parser.c`.

The complete `./run_all_tests.sh` handoff passed in 683.69 seconds at
4,609,656 KB peak RSS:

- all coverage, scoreboard, selection, and threshold contracts passed;
- all non-compiler packages passed;
- all 33 compiler batches passed;
- the known specialization-heavy batch 19 completed in 127.664 seconds;
- batch 29 completed in 66.728 seconds;
- the complete bytecode fixture pass completed in 85.523 seconds.

No test body changed, so the prior individual-test timing audit remains
authoritative for aggregate batches 19 and 29.

Canonical stdlib verification passed in both modes in 53.04 seconds at
803,072 KB peak RSS:

- tree-walker: 17 seconds;
- bytecode: 15 seconds.

## Cleanup

The reviewed default cleanup removed 214 MiB of reproducible local Go cache,
an empty local temp tree, and 140 KiB of Python bytecode. The three empty,
disk-backed `/var/tmp` work directories created for this tranche were also
removed. A post-cleanup dry run reports no default project-local candidate.

The reusable `/var/tmp/able-go-cache` and bounded
`/var/tmp/able-compiled-test-cache` remain. No profile evidence was removed.

## Scope

No compiler lowering, generated execution code, runtime, interpreter, VM,
canonical stdlib, language, dependency, benchmark, reference application, or
WASM behavior changed.

## Recommendation

Keep production performance mutation paused because the authoritative profile
and generated-artifact checks still identify no non-closed owner shared by
three unlike applications.

Next perform the first non-mutating release-consolidation review boundary for
language, parser, and typechecker paths.

Why: all known handwritten size violations are now closed, but the long-lived
dirty worktree still needs subsystem-level review before any maintainer can
safely authorize commits.

What it entails: map the final language/parser/typechecker paths to their
retained records, verify generated/source boundaries and spec alignment,
exclude local/generated and deferred-WASM paths, rerun the focused gates, and
produce a deterministic review manifest. Do not stage, commit, reset, or
rewrite history without explicit authorization.

Why it is important: this makes the verified native-lowering state auditable
and releasable without inventing per-tranche provenance or disturbing the
closed compiler/interpreter boundary evidence.
