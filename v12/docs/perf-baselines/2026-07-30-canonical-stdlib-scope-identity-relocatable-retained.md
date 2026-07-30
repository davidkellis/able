# Canonical stdlib scope identity relocatable

Date: 2026-07-30

## Decision

Retain a general performance-evidence correction: a scope override changes
where the ledger reads files, but it does not change the configured semantic
identity of that scope.

The canonical stdlib can now live at any adjacent checkout path without
invalidating evidence when its selected relative paths and contents are
identical. Real stdlib content drift still invalidates all closures that
depend on it.

No compiler, runtime, interpreter, VM, stdlib, benchmark, language, fixture,
dependency, or WASM behavior changed.

## Correction

`hash_scope` previously wrote the resolved filesystem root into each checked
scope record. Most roots are inside Able and became repository-relative, but
`../able-stdlib/src` resolves outside the Able checkout and therefore became
an absolute workstation path.

The ledger now records the configured scope root:

`../able-stdlib/src`

The resolved default or `--scope-override` path is used only to enumerate and
read files. The existing content contract is unchanged:

- match the configured include and exclude patterns;
- retain each matched path relative to the scope root;
- hash every matched file;
- hash the ordered relative-path/content records; and
- invalidate on definition, path-set, or content drift.

An override is therefore a relocation mechanism, not a semantic scope
redefinition.

## Regression coverage

Two new tests use separate temporary canonical-stdlib directories:

1. identical relative paths and contents at different roots retain an empty
   selector and the semantic root `../able-stdlib/src`; and
2. changing the relocated file content selects all 23 dependent closures with
   exactly `scope-content-drift:canonical-stdlib`.

The existing test-only compiler exclusion and native bytecode `*_wasm.go`
exclusion guards also pass.

The ledger suite now runs ten tests: nine pass and the conditional
invalidated-Markdown test skips because the current selector is empty.

## Checked baseline

The rebuilt closure baseline changed only its canonical-stdlib root identity:

| Property | Before | After |
| --- | --- | --- |
| root | `/home/david/sync/projects/able-stdlib/src` | `../able-stdlib/src` |
| files | 70 | 70 |
| tree SHA-256 | `382d256e…` | `382d256e…` |
| baseline SHA-256 | `15fcba70…` | `aaf867b3…` |

The evaluated JSON remains byte-identical at
`2d0f331971c938a7c086d8be95aa255bc9c34dfb6d7943686ce7d7debdc423b2`.
The evaluated Markdown remains byte-identical at
`dcf133cbdd3fe987b9c541deca3eb3a6a84375c418145c9ba9021542f332a8ea`.
All 23 closures are current, invalidations are zero, and the selector is
empty.

Retained implementation identities:

- ledger tool:
  `d6e64acc4217b445fc43d49aa8a4c5e572fb97fe7e0046758f79bc79109eedc7`;
- ledger tests:
  `3adec4b6b378178dfc40b4dac1a4614023c487d805ce5428d2c420c9173d1ad0`;
- checked baseline:
  `aaf867b3b78add52fbf5b36efae45a4c097b9ae9e142f7bf826c754c942d3171`.

## Validation

Both the developer worktree and a clean source-only export pass:

- external scoreboard: 130 rows;
- repeated evidence: 130 full-status rows, five successful Able/reference
  samples per row, and 31 retained source/reference reports;
- relocation tests: four pass;
- performance frontier: 130 rows and zero actionable groups;
- frontier tests: five pass;
- evidence ledger: 23 current closures and zero invalidations; and
- ledger tests: nine pass and one conditional skip.

The clean proof started from published commit
`9a4eac717e1a46e10195e17c562b389fa452dcfd`, overlaid only the ledger tool,
test, and rebuilt baseline, and contained 10,966 regular Able files. It
contained none of the three deferred bytecode WASM proof files.

The adjacent canonical stdlib reproduced 70 runtime sources and scorecard
tree SHA-256
`6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`.
No scope override or developer-worktree path was used.

## Cleanup

All clean-export and test state lived under disk-backed `/var/tmp`. The exact
140,948 KiB workspace contained no Python, Go-build, or project cache and was
removed after preserving this record. No task state used RAM-backed `/tmp`.

## Next

Refresh the exact retained/deferred release inventory without staging,
committing, or pushing.

Why: the worktree now combines the failed reconciliation record, the retained
three-file ledger correction, this record, and handoff updates with the
unchanged deferred WASM boundary.

What it entails: expand every untracked file, record state, line, byte, and
SHA-256 identities, reproduce the exact 34-path deferred complement, validate
format/scope/evidence/cleanup gates, and define the retained candidate.

Why it matters: the portability correction can later enter history without
absorbing deferred WASM work or any unrelated dirty path.
