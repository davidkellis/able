# Verified bytecode file/byte-reference gate (2026-07-11)

## Coverage and method

Reverse Complement was the remaining external file/byte-transform benchmark
without current Python/Ruby comparisons. The sibling suite now contains Python
3.14 and Ruby 4.0 implementations that read the shared FASTA file, preserve
headers, apply the complete ASCII complement mapping, reverse each sequence,
and wrap output at 60 bytes. Both outputs were checked byte-for-byte with the
checked-in 2 MB solution, in addition to the existing suite verifier.

K-Nucleotide is the independently shaped file/byte control: it has current
references and processes the same broad kind of FASTA input, but its rolling
frequency-map and call structure differ from Reverse Complement's direct
byte-array transform. All measurements use CPU `2`, a 45-second process cap,
the sibling working directory, and the canonical external stdlib.

| Application | Python 3.14 | Ruby 4.0 | Validation |
| --- | ---: | ---: | --- |
| Reverse Complement | 0.0276 s | 0.0825 s | exact solution SHA-256 and verifier, 3/3 |
| K-Nucleotide control | 1.4002 s | 1.3860 s | verifier, 3/3 |

The current three-process Able bytecode row for Reverse Complement verifies
the same solution hash but remains well short of the interpreter floor:

| Application | Bytecode | Bytecode/Python | Bytecode/Ruby | Validation |
| --- | ---: | ---: | --- |
| Reverse Complement | 7.3633 s | 266.79x | 89.25x | exact verifier, 3/3 |
| K-Nucleotide control | timeout at 45 s | n/a | n/a | not run after cap |

The K-Nucleotide timeout is status evidence, not an averaged zero or a failed
verification. A prior current-code 40.54-second CPU capture remains the
bounded attribution control; the runtime sources did not change between it
and this refreshed reference coverage.

## Warm-profile attribution

The Reverse Complement runtime benchmark loaded and warmed the external
program before one profiled `main()` invocation. It measured
`7,329,818,535 ns/op`, `705,448,536 B/op`, and `10,894,384 allocs/op`.
Its 7.29 CPU seconds identify a direct primitive-array path:

| Reverse Complement VM path | Cumulative CPU |
| --- | ---: |
| `execCallMemberArraySlot(...)` | 30.73% |
| `appendSlotStackValueChecked(...)` | 22.50% |
| `bytecodeBoxedIntegerValue(...)` | 21.54% |
| `execCallOpcode(...)` | 39.51% |
| `execArrayPushMemberFast(...)` | 14.68% |
| `ArrayStoreMonoPrimitiveReadInfoIntoFresh(...)` | 5.08% |

`execCallOpcode(...)` and `runResumable(...)` are dispatcher parents, not
candidates. K-Nucleotide's current capture instead has call-name/cache,
inline-return, rolling raw-integer binary work, and no material direct
array-read/push or stack-materialization child. The earlier independent
ByteHistogram comparison likewise did not repeat Reverse Complement's direct
array lane. Existing array-slot cache/proof, raw-carrier, and stack-append
variants have already failed broad array, string, and list controls.

## Decision

Keep no bytecode VM, compiler, tree-walker, or `able-stdlib` change. The new
references make the largest file/byte bytecode gap comparable to both target
interpreters, but the independent FASTA control rejects an `Array u8`, DNA,
file, stack, or array-member specialization. No benchmark-specific source
fusion is authorized.

## Next recommendation

Complete the remaining external interpreter-reference coverage for N-body
(Python and Ruby) and PiDigits (Python), then publish a fresh pinned status
scorecard before taking another performance candidate. This closes the last
cross-language gaps in the completed external suite and is more useful than
re-profiling the now-disjoint byte family. It entails portable source/container
recipes, verifier-backed multi-process rows, and no runtime change unless a
new benchmark pair exposes the same concrete leaf.
