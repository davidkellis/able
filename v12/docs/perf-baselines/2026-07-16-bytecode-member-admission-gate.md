# Bytecode member-cache admission gate

Date: 2026-07-16

## Decision

Keep no new runtime code. The proposed hotness admission rule does not match
the measured behavior, and a generic same-parent sibling-scope shortcut was
neutral to slower across the broad gate. All census hooks and the shortcut
were removed; the previously retained dependency-validated member-method cache
is unchanged.

## Path census

A temporary stats-only census split member-cache hits into the exact-
environment stable path and the dependency-complete path. Each application ran
one warmed measured call in a separate process with canonical
`../able-stdlib`, source-root-only resolution, `GOMEMLIMIT=1GiB`, `GOGC=50`,
and `GOMAXPROCS=1`.

| Application | Hits | Misses | Exact-env hits | Dependency checks | Inline dependency hits |
| --- | ---: | ---: | ---: | ---: | ---: |
| Option/Result Configuration | 129,164 | 5,236 | 0 | 134,400 | 129,164 |
| Await Channel Mux | 2,554 | 6 | 0 | 2,560 | 2,554 |
| Mutex Await Journal | 4,086 | 10 | 0 | 4,096 | 4,086 |
| Mutex Ledger | 16,376 | 8 | 0 | 16,384 | 16,376 |
| Document Audit | 11,934 | 28 | 0 | 11,962 | 7,901 |
| Lexical Rollup | 91,402 | 27 | 0 | 91,429 | 91,402 |
| Numeric Array Map | 5 | 10 | 5 | 10 | 0 |
| Linked-list Iterator Collect | 100,003 | 25 | 7,999 | 92,029 | 91,998 |
| String Split/Join | 28,031 | 31 | 14,012 | 14,050 | 0 |

The remaining hits in Document Audit and String Split/Join came through the
map-backed cache after dependency validation. The central result is that the
guard cost is not dominated by a few cold member sites. Document Audit,
Lexical Rollup, iterator collect, and String Split/Join all enter dependency
validation thousands to tens of thousands of times. Delaying cross-environment
owner capture until a second observation would remove at most initial-site
work and would leave the repeated path intact. No threshold candidate was
built because the prerequisite diagnosis was false.

## Rejected same-parent shortcut

A second generic trial tested whether repeated sibling call scopes could avoid
full ancestor-chain owner discovery. It admitted a cache entry only when:

- the old and current transient environments had the exact same parent;
- neither transient scope owned or currently bound the member name, canonical
  receiver type, or receiver-type alias;
- program/IP, receiver identity, method version, environment-family state,
  impl context, and all relevant per-name revisions still matched; and
- the captured mutable member owner retained its revision.

This was a lexical-shape mechanism with no benchmark, method, receiver,
container, channel, mutex, or stdlib special case. Focused cache and shadowing
tests passed. Its shallow current-scope probes, parent checks, and name-revision
locks nevertheless cost about as much as the owner walk they replaced.

## Clean repeated A/B

The retained-cache baseline and same-parent candidate were separate test
binaries. Processes were alternated, and arithmetic means are reported for
this shared workstation. Option/Result used three runs per side; the other
material applications used five.

| Application | Runs/side | Baseline ns/op | Candidate ns/op | Time | Bytes | Allocs |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Option/Result Configuration | 3 | 722,575,709 | 729,818,379 | +1.00% | -0.00% | 0.00% |
| Await Channel Mux | 5 | 41,711,272 | 41,996,168 | +0.68% | -0.00% | -0.00% |
| Mutex Await Journal | 5 | 51,546,463 | 52,633,787 | +2.11% | +0.29% | +0.02% |
| Mutex Ledger | 5 | 188,082,789 | 186,456,123 | -0.86% | +0.03% | 0.00% |

The guard gate used ten runs per side for Document Audit and Lexical Rollup
and five for the fixture guards:

| Guard | Runs/side | Time | Bytes | Allocs |
| --- | ---: | ---: | ---: | ---: |
| Document Audit | 10 | +7.03% | -0.00% | -0.03% |
| Lexical Rollup | 10 | +1.23% | -0.00% | -0.00% |
| Numeric Array Map | 5 | -7.89% | 0.00% | 0.00% |
| Linked-list Iterator Collect | 5 | +4.42% | 0.00% | 0.00% |
| String Split/Join | 5 | +2.19% | -0.04% | -0.00% |

The material gate is neutral-to-mixed, while three substantial guards regress.
Allocation shape is unchanged, confirming that this was only a more expensive
validation ordering. The shortcut was fully reverted.

## Verification and cleanup

After removing the census and candidate:

- focused member-method, call-member, static-member, shadowing, generic-member,
  and lambda tests pass;
- the member-cache source returned to 928 lines and its direct-cache tests to
  183 lines;
- no compiler, canonical stdlib, benchmark source, verifier, reference, or
  language-spec change remains.

The temporary census binaries, A/B binaries, reports, and benchmark outputs
are cleanup-only and are removed after final verification.

## Next recommendation

Refresh the verifier-backed external bytecode scorecard across the portable
coverage suite, then cluster the largest remaining Python/Ruby performance
misses by shared VM descendant before selecting another candidate.

Why: the two retained cache tranches materially changed Option/Result and the
three concurrency applications, making the old comparative scoreboard stale.
This gate also closes the current member-cache validation policy: both
thresholding and a cheaper sibling-parent probe lack a broad win. A fresh
coverage ranking will move work to the largest current product gaps instead of
continuing to optimize a now-small and noisy local wall.

What it entails: run each portable bytecode application repeatedly with its
external verifier and the same workstation averaging policy; compare current
means with the available Python and Ruby references; rank misses by ratio and
absolute Able time; then take bounded CPU/allocation profiles for three unlike
applications from the largest repeated cluster. Advance only an exact generic
descendant shared by all three. Do not begin WASM work or add benchmark,
stdlib-type, nominal-container, or source-shape special cases.
