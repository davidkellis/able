# Bytecode transient-scope member-method cache

Date: 2026-07-16

## Decision

Keep dependency-validated reuse of hot/direct member-method cache entries across
transient sibling call scopes. The previous cache required the exact
`*runtime.Environment` pointer even when two short-lived call scopes belonged
to the same environment family and exposed the same method and receiver-type
definitions. Hot member sites therefore resolved and stored the same method on
nearly every call.

The retained path still requires the same program, instruction, member name,
receiver identity, method-cache version, environment-family state, impl
context, and per-name shape revisions. It additionally records and compares
the exact environment owners of the visible member binding, canonical receiver
type, and receiver-type alias. A mutable member owner must also retain its
captured revision. The old exact-environment stable-shape path remains the
cheapest first probe. Only its dependency-complete fallback may cross an
environment pointer, and a successful hit refreshes the transient environment
and aggregate shape revision in the inline entry.

This is a general lexical dependency rule. It contains no benchmark, method,
receiver, nominal type, concurrency primitive, or stdlib-container special
case.

## Post-lambda-cache profile refresh

Each application ran in its own clean process after load, typecheck, and one
warm call, with canonical `../able-stdlib`, source-root-only resolution,
`GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`. CPU and allocation profiles
were separate processes under the 55-second guardrail. Exact benchmark
allocation counters, rather than process-start samples present in allocation
profiles, are authoritative for bytes and object counts.

| Application | Profile calls | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: | ---: |
| Option/Result Configuration | 5 | 2,495,668,671 | 301,738,388 | 4,118,969 |
| Await Channel Mux | 50 | 77,783,863 | 12,542,108 | 192,031 |
| Mutex Await Journal | 50 | 85,713,302 | 10,262,676 | 198,860 |
| Mutex Ledger | 20 | 248,951,321 | 29,691,622 | 534,494 |

The refreshed profiles no longer share the old lambda-lowering wall. They do
share repeated member-method resolution and cache storage:

- Option/Result spent 52.37% cumulative CPU below `execCallMember`; its
  allocation profile attributed 135,442 sampled objects and 90.59 MB to
  `storeCachedMemberMethod`. Cache stats reported 134,400 misses and no hits.
- Await Channel Mux attributed 8.53% cumulative CPU to member lookup and 2,598
  sampled objects/1.54 MB to member-cache storage. Stats reported 2,560 misses
  and no hits.
- Mutex Await Journal attributed 2,117 sampled objects/1.48 MB to cache
  storage. It reported 2,041 hits and 2,055 misses, so the remaining repeated
  misses were still material even though some sites already reused entries.
- Mutex Ledger attributed 16,564 sampled objects/11.55 MB to cache storage,
  25.6% of sampled allocation space. It reported 16,384 misses and no hits.

Other leaves diverged. Generic-union expansion and type matching were material
only in Option/Result. Raw integer-result construction repeated in the three
concurrency applications, but that generic family was already tested and
rejected by the cross-family raw-integer gate. Scheduler, channel, and mutex
descendants were not shared with Option/Result.

## Rejected environment allocation trial

A generic construction trial coallocated an environment with its mutex. It
did not reduce allocation counts and crossed unfavorable Go size classes. Five
alternating runs per side regressed Await Channel Mux by 10.32%, Mutex Await
Journal by 4.36%, and Mutex Ledger by 5.37%; allocated bytes rose 8.79%,
15.62%, and 30.56% respectively. The runtime and test changes were completely
removed before the retained member-cache candidate was measured.

## Candidate refinement

The first safe cross-scope cache design retained several owner/version pairs
in every inline entry. It produced the expected material-workload wins but
made the VM/cache structures large enough to regress short guards. A second
design moved the state to a heap object, which added an owner-state allocation
per store and still failed the guard bar. Both were removed.

The retained representation needs only two additional owner pointers for the
receiver type and its alias. The existing member owner and revision already
cover the mutable method-name dependency. Tests prove reuse across sibling
call scopes and rejection when a sibling shadows either the member name or the
receiver nominal type.

## Clean repeated A/B

Baseline and candidate test binaries were built from the same worktree before
and after the candidate, then run as alternating clean processes. The table
reports arithmetic means, as required for this shared workstation. Material
applications used five runs per side except the longer Option/Result workload,
which used three.

| Application | Runs/side | Baseline ns/op | Candidate ns/op | Time | Baseline B/op | Candidate B/op | Bytes | Baseline allocs/op | Candidate allocs/op | Allocs |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Option/Result Configuration | 3 | 4,026,907,187 | 1,427,978,680 | -64.54% | 317,139,659 | 76,362,592 | -75.92% | 4,119,353 | 1,305,012 | -68.32% |
| Await Channel Mux | 5 | 73,360,596 | 52,595,775 | -28.31% | 12,824,741 | 9,940,120 | -22.49% | 192,208 | 166,639 | -13.30% |
| Mutex Await Journal | 5 | 75,842,648 | 69,585,688 | -8.25% | 10,687,728 | 8,871,029 | -17.00% | 199,177 | 190,978 | -4.12% |
| Mutex Ledger | 5 | 247,640,747 | 239,514,819 | -3.28% | 32,173,643 | 17,336,795 | -46.11% | 534,839 | 436,419 | -18.40% |

Diagnostic screens changed cache results to 129,164 hits/5,236 misses for
Option/Result, 2,554/6 for Await Channel Mux, 4,086/10 for Mutex Await Journal,
and 16,376/8 for Mutex Ledger. These screens had stats enabled and are not used
as timing evidence.

Five unlike low-frequency guards were also alternated. Their allocation-byte
means are neutral or slightly favorable except String Split/Join at +0.03%;
runtime means remain noisier on the shared workstation:

| Guard | Runs/side | Time | Bytes | Allocs |
| --- | ---: | ---: | ---: | ---: |
| Document Audit | 10 | +5.44% | -2.20% | +2.00% |
| Lexical Rollup | 10 | +2.51% | -0.58% | +0.11% |
| Numeric Array Map | 5 | +3.24% | -0.30% | +2.18% |
| Linked-list Iterator Collect | 5 | -2.17% | -0.23% | +0.00% |
| String Split/Join | 5 | +0.90% | +0.03% | +0.00% |

The short-guard timing costs are recorded rather than hidden. The candidate is
retained because it removes 4.12-68.32% of allocations and improves all four
independently selected material applications, while the guard allocation-byte
signal is essentially flat to favorable. The next tranche must specifically
address the extra validation tax at cold member sites rather than enlarging
the cross-scope entry again.

## Correctness and verification

Verification completed:

- focused member, shadowing, generic-member, alias/import impl, and lambda
  tests: pass;
- `go test ./pkg/runtime -count=1 -timeout 60s`: pass;
- interpreter package excluding the two known fixture-harness blockers: pass
  in 29.480 seconds;
- fresh source-built bytecode CLI output passed the external Ruby verifier for
  all four material applications plus Document Audit and Lexical Rollup.

The verified stdout SHA-256 values are:

| Application | SHA-256 |
| --- | --- |
| Option/Result Configuration | `28e46b27a6dceeaa15968e9db7a6728f4a2b35f87a89ff7bf561db18cad53112` |
| Await Channel Mux | `0a32fa2e9481ecb635562079d993288d911326d9f36aeea0d9d3058d1a494693` |
| Mutex Await Journal | `e6f87830a6a1bb1bf3889a6b77eae053bfd3630055a06dd8fd500201586bf61e` |
| Mutex Ledger | `58f2e2fda882a7c1e5032a2f34c4d9d05f40ced51199c0be0672ecc4d4cf14e4` |
| Document Audit | `0dad030a80c8a883cbb56fbcfebfd530d521075e15d5d91ba538bc93e66c0aab` |
| Lexical Rollup | `a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604` |

The known fixture-harness blockers are unchanged: the unfiltered parity
harness can exceed the mandatory one-minute ceiling in deeply nested regex-set
parser recursion, and the alias-reexport error fixture expects an unqualified
fragment while the interpreter emits the qualified interface name. Neither
path enters the changed cache.

No compiler, canonical stdlib, application, benchmark source, verifier,
reference implementation, or language-spec change was needed. All temporary
profile hooks and rejected candidates were removed.

## Next recommendation

Reduce the retained cache's cold-site validation tax without weakening its
lexical dependency checks, using a bounded hotness admission rule for the
cross-environment fallback.

Why: the large allocation and runtime wins prove that cross-transient-scope
reuse is a real shared mechanism, but Document Audit, Lexical Rollup, and the
small numeric Array guard expose a modest cost when a member site executes too
few times to amortize full owner discovery. The remaining risk is now the
policy for deciding when to pay for dependency-complete validation, not the
correctness or usefulness of the cache itself.

What it entails: add temporary per-site probe counters to classify cold versus
repeating member sites across the four material applications and the five
guards; trial one small bounded admission threshold that leaves exact-env hits
unchanged and begins cross-env owner capture only after repetition is observed;
then run alternating repeated A/B means. Retain it only if it preserves the
material allocation wins, removes the guard slowdown, and passes the sibling
shadowing tests. Do not special-case a member name, receiver type, benchmark,
stdlib container, channel, or mutex.
