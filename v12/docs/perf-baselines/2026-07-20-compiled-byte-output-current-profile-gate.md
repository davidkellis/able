# Compiled byte/output current-profile gate

Date: 2026-07-20

## Outcome

This gate refreshes current generated-main CPU and allocation ownership for
Base64, FASTA Generation, PiDigits, and Reverse Complement and retains no
compiler, generated-runtime, canonical-stdlib, benchmark, fixture, or language
change.

The four programs share an external label—each produces byte- or text-heavy
output—but not a material generated compiler/runtime descendant. Their exact
owners are respectively host codec/MD5 work, direct generated random-sequence
arithmetic, Go `math/big`, and application transform/copy/GC work. No candidate
meets the required three-unlike-application breadth.

## Reproducibility contract

- Go: `go1.26.4 linux/amd64`
- repository HEAD: `237406eccdfb025a519d898daedadee1c8d13a7b`
- compiler source tree SHA-256:
  `b485297e8dfec0e1697622042bf44d0caa11da8941e842d3170b68576891e86f`
- preserved `ablec` SHA-256:
  `114ead1081b975893ed867343dc259db73860998adba7e3b66aadebf9d1b85e2`
- canonical stdlib content-tree SHA-256:
  `5da10c658627de75d2e4dd6b2f9b5278ccdb77c0a6fd6b6d76b1c4f01737520d`
- CPU 0, `GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`, and a
  55-second per-process limit
- generated phase profiling excludes launcher/bootstrap CPU from the selected
  `main` profiles
- CPU and exact allocation profiles used separate processes.

Application identities were:

| Application | Able source SHA-256 | Preserved binary SHA-256 |
| --- | --- | --- |
| Base64 | `b4676ab1b4392ed4433d7a2ce57c7388907e4719494e6edce32728b071750108` | `49db02812d57cd2fd779d3019d9ad31c51a01b052227f4e708718457c766d05e` |
| FASTA Generation | `f8c67c9ab16e29d92904db2d58091f2512f83319a3bca5caf8ee90c37a2a96d7` | `b4ea612e229d8db10adc322b84ab94ac38d70c9108731d66d102dec364e763cb` |
| PiDigits | `236a74ef456b4a5ca3e33a743a0b3b8e8767db9cfa07dd619df870bff876d5cb` | `ff20a6f2c325a1b874847e095bb2a8afa77dcd07b043f3f1937bc41c30200958` |
| Reverse Complement | `4bd16d4b6c65362efc5b2515548f59686144553a91eadcc65a0cde6df537e5f9` | `293989341dbadd7c1f029635863ad1b346ba5c0573e81d8f1228410b8f17eeb0` |

## Current verifier-backed timing

Five fresh preserved-binary Able processes and five freshly built Go reference
processes completed per application. Every one of the 40 outputs passed the
public verifier.

The target permits Able / Go up to `1 / 0.95 = 1.0526x`.

| Application | Able mean | Go mean | Able / Go | Current cohort observation |
| --- | ---: | ---: | ---: | --- |
| Base64 | 2.3960 s | 2.5302 s | 0.9468x | meets |
| FASTA Generation | 0.0840 s | 0.0148 s | 5.6757x | misses |
| PiDigits | 1.1940 s | 1.2132 s | 0.9842x | meets |
| Reverse Complement | 0.0900 s | 0.0170 s | 5.2941x | misses |

Base64 and PiDigits are protected current-cohort target meets. One targeted
cohort does not replace the promoted two-cohort scorecard classification; it
does mean any later compiled candidate must guard these results.

## Generated-main CPU ownership

Base64 and PiDigits used two and three profile processes respectively. The
short FASTA and Reverse programs used 60 verified launches each so their merged
profiles contained 980 ms and 740 ms rather than relying on one sparse sample.

| Application | Main samples | Exact current owner |
| --- | ---: | --- |
| Base64 | 5.32 s | Go Base64 encode 44.55% cumulative, decode 26.88%, and MD5 block 13.35% |
| FASTA Generation | 0.98 s | `append_random` 91.84% cumulative; checked multiply 22.45%, signed `DivMod` 13.27%, and direct LCG/base selection |
| PiDigits | 3.80 s | `math/big` word multiplication, shifts, add, and subtract; `mulAddVWW` alone is 40.79% flat |
| Reverse Complement | 0.74 s | generated reverse/complement transform 79.73% cumulative, wrapping 22.97%, `memmove` 10.81%, and GC scanning |

The only repeated broad CPU envelope in three rows is Go GC scanning. Its
allocation roots are different: exact codec result slices in Base64,
`math/big` word storage in PiDigits, and generated transform/backing storage in
Reverse Complement. GC is therefore a downstream runtime parent, not an exact
compiler candidate shared by the programs. FASTA is allocation-light and does
not corroborate it.

## Exact main-phase allocation ownership

| Application | Allocated bytes | Allocations | GCs | Exact owner |
| --- | ---: | ---: | ---: | --- |
| Base64 | 2,201,553,528 | 128 | 20 | exact host encode/decode result slices |
| FASTA Generation | 1,058,576 | 446 | 0 | pre-sized generated output plus small String conversion objects |
| PiDigits | 298,858,792 | 24,224 | 17 | `math/big.nat.make` and digit printing |
| Reverse Complement | 9,314,944 | 64 | 1 | input buffer and generated transform/wrapped output backing |

This also closes the earlier capacity-growth hypothesis: FASTA remains
pre-sized, Base64's codec results are exact-sized, PiDigits is unrelated
BigInt storage, and Reverse's backing growth is not repeated in three unlike
applications. The retained `write_all` change has already removed the former
shared unconditional output copy.

## Reconciliation and decision

No generic candidate was admitted:

- host codec/MD5 optimization would affect Base64 only;
- checked primitive arithmetic is material in FASTA only;
- BigInt arithmetic/storage and repeated print conversion belong to PiDigits;
- transform/backing growth and wrapping belong to Reverse Complement;
- write syscalls are material only in FASTA and Reverse; and
- GC scanning reaches three programs, but no common compiler-controlled
  allocation descendant feeds it.

Adding a named Array, codec, BigInt, FASTA, or reverse-complement lowering would
violate the general nominal-lowering rule. Reopening rejected general capacity,
call-context, bridge, or output-copy candidates without a new exact child would
optimize a category label rather than shared generated work.

The group is current and closed as `closed-no-shared-leaf`. The bounded raw
artifacts remain under `/tmp/able-compiled-byte-output-20260720-a` for this
session. No WASM work was performed.

## Verification

- one preserved current compiler and four compiled application binaries built;
- 20/20 Able timing processes passed their public verifiers;
- 20/20 freshly built Go reference processes passed their public verifiers;
- 125/125 CPU-profile processes passed the same verifiers; and
- four exact allocation-profile processes completed and verified.

## Next recommendation

Refresh the complete selected compiled/bytecode scorecard in two independent
five-run cohorts, then regenerate the frontier from those current measurements.

Why: every currently selected ownership group is now either current or closed,
while this tranche moved Base64 and PiDigits from promoted misses to target
meets in one current cohort. A second full, independently averaged measurement
is needed to distinguish durable progress from workstation variance and to
rank the next real cross-application gap without reopening exhausted profile
families.

This entails preserved current compiler/interpreter identities, five
verifier-backed Able runs per selected mode and application in each cohort,
matching current Go/Python/Ruby references, strict variance checks, and a new
source/evidence-consistent frontier. If no new generic exact owner emerges,
the following tranche should extend benchmark coverage for underrepresented
language features rather than specialize a closed workload. Do not begin WASM.
