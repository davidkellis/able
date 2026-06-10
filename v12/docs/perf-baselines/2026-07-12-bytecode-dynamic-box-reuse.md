# Bytecode Dynamic Integer-Box Cache Reuse (2026-07-12)

## Decision

Keep the existing 262,144-entry bounded dynamic integer-box cache and make no
VM, compiler, benchmark, fixture, or `able-stdlib` change. The new data rules
out the only policy premise that could have justified another cache experiment:
there is no repeated low-reuse dynamic-cache class across independent
applications.

K-Nucleotide gets very high reuse before it reaches the cap. Reverse
Complement has a lower hit rate and saturates the cache, but it still receives
millions of hits. Base64 and I-Before-E do not enter this dynamic tier at all.
A cache policy that classifies any of those shapes would therefore be a
workload-specific rule, not a shared VM improvement.

## Method

`able_bytecode_box_reuse` is an opt-in Go build tag. Only a test binary built
with that tag increments the diagnostic counters; normal VM, CLI, and compiler
builds select a `false` compile-time constant, so the counter guards are
eliminated with no normal hot-path branch or allocation. The retention harness
resets the tagged counters after loader/program setup and immediately before
calling `main`, so each report covers application execution rather than
lowering or import setup.

Each source program ran in a fresh CPU-2-pinned process with the canonical
external stdlib, `GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and the existing
60-second test bound. `lookups` is every dynamic-tier request; a miss is
`lookups - hits`, split into cache insertion and a request after the cache has
reached its cap. The unchanged post-scope probe also continues to record final
ArrayStore and cache occupancy.

| Application | Dynamic i32 lookups | Hits | Hit rate | Inserts | At-cap misses | Final i32 entries |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Reverse Complement | 9,380,339 | 4,591,000 | 48.94% | 262,144 | 4,527,195 | 262,144 |
| K-Nucleotide | 5,690,960 | 5,450,412 | 95.77% | 240,548 | 0 | 240,549 |
| Base64 control | 0 | 0 | n/a | 0 | 0 | 0 |
| I-Before-E control | 0 | 0 | n/a | 0 | 0 | 0 |

K-Nucleotide also made eight large-`i64` requests, all through the preexisting
direct bypass. It did not use that kind's dynamic map. The one-entry difference
between K-Nucleotide's inserted and final `i32` values came from setup before
the counter reset and does not affect the main-call reuse result.

The raw one-process reports are retained in the sibling
`2026-07-12-bytecode-dynamic-box-reuse/` directory. All four probes completed
their existing test path; no timing claim is drawn from the tagged binary,
because its synchronization is deliberately diagnostic-only.

## Interpretation

The rejected 16,384-entry global-cap change slowed K-Nucleotide because that
worker is almost entirely dynamic-cache hits after its approximately 240k
unique values are inserted. Reverse Complement supplies the opposite pressure
(4.53M unique requests after saturation), but it also retains 4.59M hits. A
smaller global cap hurts the former; a high-cardinality eviction or scoped
policy would remove valuable reuse from the latter unless it recognized
benchmark-shaped behavior. The two controls exercise neither path, confirming
that their one-run cap timing variation was not cache-policy evidence.

The cache is a bounded, process-global trade-off that remains appropriate at
the current cap. Do not revisit cap, eviction, clearing, or per-program
policies without a new language-level lifetime or identity requirement.

## Next recommendation

The canonical runtime-value architecture design was already completed and
rejected a universal carrier prototype. The subsequent normal-build opcode
census likewise finds no new shared eligible VM leaf; see
`2026-07-12-bytecode-opcode-census.md`. Do not take another unchanged-source
micro-tranche. Resume language/runtime feature completion and use the complete
cross-language application/fixture matrix to profile a genuinely new shared
boundary before optimizing it. This is necessary to avoid repeatedly measuring
the same disjoint paths or introducing a benchmark-shaped rule.
