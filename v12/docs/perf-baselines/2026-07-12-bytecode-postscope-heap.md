# Bytecode Post-Scope Heap Attribution (2026-07-12)

## Decision

Keep the new test-only post-scope heap-profile diagnostic, but keep no runtime
cache-policy change. The profile distinguishes ArrayStore reclamation from
ordinary Go heap retention after a program scope ends. It found a real shared,
bounded global dynamic-`i32` box cache in Reverse Complement and K-Nucleotide,
then rejected the only generic cap-reduction candidate against verifier-backed
application guards.

No normal VM, CLI, compiler, benchmark, fixture, or `able-stdlib` path gained
a profile branch or allocation. `ABLE_BENCH_RUNTIME_RETENTION_HEAP_PROFILE`
is read only by the test harness and writes a heap profile after the existing
scope teardown and three forced GCs. The report's test-only
`dynamic_integer_box_cache_size` snapshot likewise does not instrument the VM
hot path.

## Method

Each application ran in its own CPU-2-pinned process with canonical external
stdlib, `GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`. The retention test
runs `main`, records ArrayStore state, releases the complete interpreter
scope, forces three GCs, then writes JSON and the opt-in heap profile.
Reverse Complement and K-Nucleotide are independent FASTA/file-byte workers;
Base64 and I-Before-E are controls.

| Application | ArrayStore states before teardown | States after teardown | Dynamic `i32` cache entries | Post-scope heap profile |
| --- | ---: | ---: | ---: | ---: |
| Reverse Complement | 6 | 0 | 262,144 | 53.42 MB |
| K-Nucleotide | 24 | 0 | 240,549 | 108.25 MB |
| Base64 control | 19 | 0 | 0 | 31.66 MB |
| I-Before-E control | 2 | 0 | 0 | 33.11 MB |

All reports and readable `pprof` summaries are retained in the sibling
`2026-07-12-bytecode-postscope-heap/` directory. The normal application
timings below ran through `bench_perf`, pinned to the same CPU, with the
upstream Ruby verifier after every successful process.

## Attribution

The exact repeated post-scope root is the dynamic `i32` cache maintained by
`bytecodeBoxedIntegerValue`:

- Reverse Complement retains 23.33 MB (43.67% of the profile) from that
  helper after filling the 262,144-entry cap.
- K-Nucleotide retains 22.89 MB (21.15%) from the same helper with 240,549
  entries. Its separate 41.05 MB `ArrayEnsureCapacity` root does not repeat in
  Reverse Complement.
- Base64 and I-Before-E retain zero dynamic integer-cache entries. Their
  profiles are instead dominated by the fixed small-int cache and, for
  I-Before-E, ordinary external String-array bridge allocations.

This is a generic cache-policy signal, not an ArrayStore release defect: all
four ArrayStore snapshots are empty after teardown, and the cache is an
intentional process-global bounded optimization. The static small-int and
extended-`i32` caches remain separate from the dynamic tier.

## Generic cap experiment

The sole candidate changed the dynamic cache cap for every supported integer
kind from 262,144 to 16,384 entries. It retained the fixed small-int and
extended-`i32` caches, changed no language semantics, and produced verifier
identical stdout. It was nevertheless reverted because the independent
K-Nucleotide control regressed across three runs and I-Before-E also regressed.

| Application | Baseline | 16,384-entry candidate | Change | Verification |
| --- | ---: | ---: | ---: | --- |
| Reverse Complement | 7.33 s | 6.90 s | -5.9% | 1/1 each |
| K-Nucleotide | 44.02 s mean (44.66, 43.02, 44.39) | 45.22 s mean (46.27, 44.85, 44.55) | +2.7% | 3/3 each |
| Base64 control | 3.41 s | 3.33 s | -2.3% | 1/1 each |
| I-Before-E control | 0.55 s | 0.61 s | +10.9% | 1/1 each |

The candidate raised K-Nucleotide's GC count from 37--38 to 43 in every
candidate run. Its Reverse Complement gain therefore cannot justify a global
memory-for-speed policy that slows a different independent file/byte worker
and an unrelated text control. The original 262,144-entry policy is restored.

## Dynamic-cache reuse follow-up

The requested build-tagged reuse measurement is complete. It records only
dynamic-tier lookup/hit/insert/full-cache-miss events after program setup; the
normal build selects a false compile-time constant and has no diagnostic work.
The data rejects another cache-policy candidate: K-Nucleotide is 95.77% hits
over 5,690,960 dynamic `i32` lookups with no saturation, while Reverse
Complement has 4,591,000 useful hits (48.94%) as well as 4,527,195 requests
after the cap. Base64 and I-Before-E do not use the dynamic tier. There is no
repeated low-reuse class on which to base an adaptive, scoped, eviction, or
clear-on-teardown policy.

The full method, raw reports, and exact counts are in
`2026-07-12-bytecode-dynamic-box-reuse.md` and its sibling artifact directory.
The existing 262,144-entry cap remains unchanged.

## Verification

Focused tests pass for retention configuration/profile output, dynamic and
fixed integer boxing, `i64` dynamic-cache bypass, stack snapshot boxing, and
the allocation-free unary small-int path. The four post-scope probe processes
completed with zero ArrayStore state; all eight baseline/candidate application
rows passed their available Ruby verifier.

## Next recommendation

The canonical runtime-value architecture design was already completed and
rejected a universal carrier prototype. The later normal-build opcode census
also finds no remaining shared eligible VM leaf; see
`2026-07-12-bytecode-opcode-census.md`. Do not take another unchanged-source
micro-tranche. Resume language/runtime feature completion and profile a new
shared execution boundary only after it is represented in the cross-language
application and fixture matrix.
