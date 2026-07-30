# Guarded generated-local cleanup retained

Date: 2026-07-30

## Decision

Retain the guarded cleanup of exactly four generated project-local paths and
the resulting exact release-candidate verification.

The cleanup removed 1,742,116 KiB of reproducible cache state. The seven
visible Python bytecode files disappeared with `v12/__pycache__`; no retained
candidate path or deferred WASM path changed. Nothing was staged, committed,
pushed, reset, reverted, or modified in production source.

## Pre-cleanup gate

The current cleanup preview matched the prior inventory exactly:

| Path | Reclaimed |
| --- | ---: |
| `v12/tmp` | 0 KiB |
| `v12/interpreters/go/.gocache` | 1,742,064 KiB |
| `v12/interpreters/go/tmp` | 0 KiB |
| `v12/__pycache__` | 52 KiB |
| **Total** | **1,742,116 KiB** |

The ordered 77-byte target list has SHA-256
`2a096f2774226f5e326502356962325a93d0240d6d03bf3468566c03aa891046`.
Each target resolved beneath the repository, contained zero tracked files,
and was selected by `scripts/cleanup.sh` without `--include-profiles`.

A read-only `/proc` scan checked every accessible process CWD, executable,
file descriptor, mapped file, and relevant `GOCACHE`, `GOTMPDIR`, `TMPDIR`,
and `PWD` environment value. It found zero owners across the four targets.
The preview was reproduced immediately before applying the cleanup.

The pre-cleanup worktree had 151 visible paths: the 110-path candidate, the
unchanged 34-path deferred WASM boundary, and seven generated-local files.
Its 9,970-byte NUL-delimited porcelain snapshot has SHA-256
`2678232e58c196a9ccae4b6845abd9926a034264de99ea73695b5f060f6c926b`.

## Cleanup result

The repository cleanup script removed only the four previewed paths. Its
181-byte output has SHA-256
`b21ca73de39fd96a957f7595373810fbfe94f226701c15a40d367b2a36520772`.
The immediate follow-up preview reports:

```text
cleanup: no generated project artifacts found
```

All audit state lived under disk-backed `/var/tmp`. The final 104 KiB task
workspace was removed after preserving this record.

The post-cleanup pre-record worktree has exactly 144 visible paths:

- 110 retained release-candidate paths;
- 34 deferred WASM paths; and
- zero generated-local paths.

Its 9,554-byte NUL-delimited porcelain snapshot has SHA-256
`be5a79ff4ed82494ab10b6c8e46a05508e7c9e5f6e72728de5f07efd529dcc1c`.
The index remains empty, and local `HEAD` and local `origin/master` remain
equal at `418886c70aee64b92b5bb3266ee5fe6453ac4320`.

## Exact candidate verification

The 110-path candidate reproduces the prior inventory's exact path identities:

- newline-delimited path list: 7,708 bytes, SHA-256
  `c1e7dc8f5ee052e52251ce62273f2770429d3c322acb3208750f1c306129dc21`;
- NUL-delimited pathspec: 7,708 bytes, SHA-256
  `08b25c1b1ca46a0721994ff256cd12be732155da186af2f60bffd236a9dde771`;
- content-identity rows: 110;
- total lines: 114,910;
- total bytes: 7,855,723;
- identity-manifest bytes: 17,372; and
- identity-manifest SHA-256:
  `dd5a3b64024b36490a72a5f772f454fec1831e7298d706dd38e153181da3bb06`.

All 34 deferred WASM paths reproduce the state, line, byte, and SHA-256
identities in the authoritative 148-row inventory. The deferred boundary was
not traversed by the cleanup script.

## Validation

- All 144 post-cleanup paths partition exactly into the candidate and deferred
  boundary.
- All 18 dirty Go files are formatted.
- All 55 dirty JSON files and nine dirty Python files parse.
- All four dirty shell programs and ten dirty JavaScript modules pass syntax
  validation.
- Whitespace, final-newline, source-size, and `git diff --check` gates pass.
- All 46 maintained dirty source files remain below 1,000 lines; the largest
  is `v12/bench_external_catalog.sh` at 883 lines.
- The extern builder, 132-row scoreboard, combined frontier, and 23-entry
  closure ledger reproduce their release-readiness identities.
- The guarded cleanup policy test passes, and its final repository preview is
  empty.
- The complete release suite remains retained for the exact unchanged
  production source. It was not rebuilt merely to verify deletion of ignored
  caches.

No compiler, generated runtime, runtime, interpreter, bytecode VM, parser
semantic, canonical stdlib, language, dependency, benchmark measurement,
fixture behavior, frozen workspace, or WASM behavior changed.

## Post-record candidate

Adding this record and its JSON companion produces an exact 112-path
candidate. Its sorted path identities are:

- newline-delimited path list: 7,868 bytes, SHA-256
  `12c33607dad6eb7033eef9e1fa56c6a825074efad46bee4307b1b232cd30991a`;
  and
- NUL-delimited pathspec: 7,868 bytes, SHA-256
  `6ab8db8d4ede1872794ff79deca560dbdf901d782b498acca5c6a11a71549790`.

Because PLAN and log edits follow the pre-record verification, the 110
non-self retained identities are refreshed separately:

- rows: 110;
- total lines: 115,003;
- total bytes: 7,860,995;
- manifest bytes: 17,372; and
- manifest SHA-256:
  `9da1bba6d8899130c7052bc14ad9c585a467ea28601bb808525e390e4b3fb33d`.

The final worktree contains exactly 146 visible paths: the 112-path candidate
and unchanged 34-path deferred WASM complement.

## Authorization

The maintainer subsequently authorized exact staging of the reviewed 112-path
candidate and creation of one local consolidation commit from
`418886c70aee64b92b5bb3266ee5fe6453ac4320`.

The authorization does not include a push, broad staging operation, reset,
revert, history rewrite, or modification of the deferred WASM boundary.

## Next

After the authorized local consolidation, verify the one-commit divergence
and obtain explicit maintainer authorization before publication.

Why: exact staging and one local commit are authorized, but remote mutation is
not.

What it entails: verify that the local branch is exactly one commit ahead,
that the commit has the reviewed parent and exact 112-path boundary, that the
index is empty, and that the 34 deferred WASM paths remain unchanged. Push
only that one commit to the intended branch if separately authorized.

Why it matters: independent post-commit verification and separate publication
authorization prevent deferred WASM, another commit, or an unintended ref
from reaching shared history.

Do not begin WASM work.
