# File/byte fresh scorecard (2026-07-12)

## Method

Reverse Complement and K-Nucleotide refreshed Go 1.26.4, Ruby 4.0.5, and
Python 3.14.5 references with three CPU-2-pinned verifier-backed runs each.
Compiled Able used three runs; Reverse Complement bytecode used three runs,
while K-Nucleotide bytecode used one bounded run because each process consumes
most of the 45-second guard. All runs used `GOMEMLIMIT=1GiB`, `GOGC=50`, and
`GOMAXPROCS=1`.

| Benchmark | Mode | Able (s) | Go ratio | Ruby ratio | Python ratio | Status |
| --- | --- | ---: | ---: | ---: | ---: | --- |
| Reverse Complement | compiled | 0.1400 | 9.09x | 1.92x | 5.53x | verified 3/3 |
| Reverse Complement | bytecode | 6.8533 | 445.02x | 93.88x | 270.88x | verified 3/3 |
| K-Nucleotide | compiled | 4.0933 | 71.19x | 3.12x | 3.06x | verified 3/3 |
| K-Nucleotide | bytecode | 42.5200 | 739.48x | 32.41x | 31.81x | verified 1/1 |

Fresh reference means are Reverse Complement Go/Ruby/Python
`0.0154s`/`0.0730s`/`0.0253s` and K-Nucleotide
`0.0575s`/`1.3119s`/`1.3368s`. The existing three-family Base64 control
remains healthy for bytecode (0.94x Ruby, 0.75x Python), so it is not evidence
for a generic file/byte slowdown.

## Profile gate

No profile was retaken. The VM source has not changed since the current-source
paired file/byte audit in
`2026-07-11-bytecode-file-byte-reference-gate.md`. Reverse Complement's
material work is direct primitive-byte `Array` reads/pushes, stack
materialization, and boxed integers. K-Nucleotide instead is call-name/cache,
inline-return, rolling raw-u64 binary, and map work. Their common dispatcher
parents are not a concrete shared helper; previous array-slot, raw-carrier,
stack-append, and map variants did not clear broad guards.

## Decision

Keep no VM, compiler, tree-walker, or `able-stdlib` change. Do not add a
FASTA, DNA, `Array u8`, HashMap, file, or benchmark-specific lowering from
these ratios. K-Nucleotide's one-run bytecode row is a valid bounded status
measurement, not a basis for variance claims or a standalone optimization.

## Next recommendation

Re-rank the remaining compiler target with a current, source-aligned
cross-feature scorecard before selecting another implementation candidate.
Why: the numerical and file/byte families have each reconfirmed exhausted or
disjoint paths, while current large compiler misses now span different source
shapes. The work entails fresh Go references and verifier-backed compiled Able
rows for a bounded set of remaining misses, with profiles only after two rows
share a concrete lowering/runtime helper and retain neutral controls.
