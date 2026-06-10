# Bytecode Array Retention Revalidation (2026-07-12)

## Decision

Keep the existing generic last-owner ArrayStore reclamation and make no new
runtime, compiler, benchmark, fixture, or `able-stdlib` change. The fresh
post-scope probe confirms that Base64, I-Before-E, and Reverse Complement all
reach zero ArrayStore handles, revisions, states, and direct backing bytes
after their interpreter scope ends and three forced Go collections run.

This corrects the interpretation of the preceding miss-profile heap summaries.
Those profiles capture in-use heap after a timed call while the benchmark's
interpreter, loader, and program remain reachable; they are useful for
in-scope allocation attribution, not for proving final ArrayStore retention.
Do not reopen the rejected frame-local explicit-release experiment or add a
codec, text, or file-specific cleanup rule.

## Method

`TestBytecodeProgramRuntimeRetention` loaded exactly one application in a
fresh process, invoked `main`, recorded ArrayStore state while the interpreter
still owned the program, returned from that scope, forced three GCs, and wrote
the final state. All runs used the canonical external stdlib,
`GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, CPU 2 affinity, and a 60-second
test bound. I-Before-E used its normal `wordlist.txt` argument.

| Application | Live ArrayStore state before scope release | Direct backing before scope release | After three GCs | Final live heap |
| --- | ---: | ---: | --- | ---: |
| Base64 | 19 `u8` states | 530,133,901 B | 0 handles, revisions, states, and bytes | 33,317,416 B |
| I-Before-E | 2 dynamic states | 2,765,184 B | 0 handles, revisions, states, and bytes | 34,192,920 B |
| Reverse Complement | 3 dynamic + 3 `u8` states | 66,245,599 B | 0 handles, revisions, states, and bytes | 127,142,808 B |

The retained reports are in
`2026-07-12-bytecode-array-retention-revalidation/` beside this record.
`BackingBytes` is direct ArrayStore-owned storage; it intentionally excludes
objects transitively retained by dynamic values or unrelated Go/plugin state.

Focused semantic verification also passed:

- ArrayStore owned-`u8`, alias, final-release, stale-handle, and atomic-move
  tests.
- Bytecode ownership transfer/error barriers, nested Array returns, and
  shared tree-walker/bytecode alias parity.
- Compiled Array carrier alias, receiver/interface writeback, and clone/move
  tests.

## Interpretation

The external result paths are no longer a generic ArrayStore lifetime
candidate. Base64's large in-scope `u8` graph and I-Before-E's dynamic
String-array graph are released by the same language-level owner protocol,
and Reverse Complement confirms that ordinary file/byte Array state follows
the same rule.

The follow-up post-scope heap profiles are complete. Reverse Complement and
K-Nucleotide share a bounded process-global dynamic-`i32` box-cache root;
they do not share an ArrayStore, loader, or host-retention root. A generic
cache-cap reduction failed the broad verified gate and was reverted. Full
attribution and the A/B result are in
`2026-07-12-bytecode-postscope-heap.md`.

## Next recommendation

The dynamic-cache reuse follow-up is now complete and rejects another cache
policy: K-Nucleotide is 95.77% hits, Reverse Complement still has 4.59M hits
despite saturating, and both controls bypass the tier. See
`2026-07-12-bytecode-dynamic-box-reuse.md`. The canonical-carrier design was
already completed and rejected a prototype; the later opcode census also
found no remaining shared VM leaf. Resume feature completion rather than
another unchanged-source micro-tranche, and profile any new shared boundary
through the full cross-language guard matrix.
