# ArrayStore Reclamation Validation

## Decision

Keep generic last-owner ArrayStore reclamation. The checked source programs
reach zero live ArrayStore handles, revisions, states, and direct backing bytes
after their interpreter scope becomes unreachable and three final GCs run. The
externally supplied Base64 verifier still passes through the normal bytecode
CLI. No benchmark-shaped VM path, compiler lowering rule, or stdlib change was
introduced.

## Method

`TestBytecodeProgramRuntimeRetention` is an opt-in test-harness probe. It
loads and runs exactly one target in one `go test` process, records an
ArrayStore snapshot while the interpreter owns the program, returns from that
scope, runs three Go collections, then writes a JSON report. The normal VM and
CLI do no additional diagnostic work unless
`ABLE_BENCH_RUNTIME_RETENTION_OUT` is set.

Every run used canonical `able-stdlib` source with:

```text
GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1
ABLE_STDLIB_ROOT=/home/david/sync/projects/able-stdlib/src
```

The raw reports are in `.profiles/20260711-array-store-retention/`.
`BackingBytes` is deliberately direct registry-owned slice backing only; it
does not count transitive objects referenced by dynamic values.

| Program | Live state before final GC | Direct backing before final GC | State and backing after final GC |
| --- | ---: | ---: | --- |
| External Base64 | 18 u8 states | 397,100,413 B | 0 states; 0 B |
| Base64 small control | 17 u8 states | 12,784,425 B | 0 states; 0 B |
| String split/join | 2,040 dynamic + 2,001 u8 states | 3,754,606 B | 0 states; 0 B |
| Iterator collect | 4 dynamic states | 589,824 B | 0 states; 0 B |
| Numeric Array map | 6 dynamic + 1 i32 state | 780,000 B | 0 states; 0 B |

The externally supplied Base64 application was also run through the ordinary
bytecode CLI from `../benchmarks/base64`; its unchanged `verify.rb` passed in
3.33 seconds after the command cache was warm. This is the same application
that previously retained 2.13 GiB after forced final GC. Its new post-GC
ArrayStore snapshot is empty and its live heap after final GC is 33.31 MB.

## Current bounded bytecode readings

The profile refresh deliberately uses the same three independent source
families used for ordinary VM guards. These are fresh current-state readings,
not an A/B claim against a pristine pre-reclamation tree.

| Program | Fixed run count | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: | ---: |
| Base64 small control | 1 | 94,388,791 | 12,948,080 | 370 |
| String split/join | 5 | 1,080,425,218 | 49,155,822 | 549,677 |
| Iterator collect | 5 | 250,251,024 | 3,254,545 | 29,090 |
| Numeric Array map | 20 | 70,068,128 | 852,588 | 333 |

The resulting CPU profiles are retained beside the snapshots. They do not show
a new large common caller:

- Split/join is still mainly inline call/return and type-match work
  (`execCallOpcode` 28.6% cumulative; `finishInlineReturn` 22.8%).
- Iterator collect remains member-dispatch/cache validation work
  (`execCallMember` 40.0% cumulative; cached member lookup 7.2%).
- Numeric Array map is raw primitive and call/frame work; its largest shared
  helper is `bytecodeRawIntegerValueInfo` (7.1% flat), while the same helper is
  smaller but present in the split/join and iterator controls.

The reclaim path is not a hot profile node, and the retained backing falls to
zero in every probe. The profile numbers therefore authorize keeping the
lifetime correction, but do not authorize a new return, member-cache, or
Array-specific CPU optimization.

## Follow-up

The raw-integer caller/carrier diagnostic follow-up is complete and rejected
its direct-store candidate. See
`v12/docs/perf-baselines/2026-07-11-raw-integer-extraction-diagnostics.md`.
The next investigation is the common `mapaccess2_faststr` samples, beginning
with map-owner and hit/miss classification rather than a cache or lookup change.
