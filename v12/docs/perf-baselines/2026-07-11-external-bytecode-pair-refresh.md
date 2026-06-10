# Verified external bytecode pair refresh (2026-07-11)

## Selection

The current external scorecard has two independently verified bytecode rows
that miss both the Ruby and Python performance floors: I-Before-E and Monte
Carlo Pi. Base64 is the unrelated verified guard because it is now essentially
at Python parity. This refresh remeasures all three before choosing another VM
change; it does not infer a candidate from one large ratio.

## Verification and method

Normal bytecode CLI runs used `taskset -c 2`,
`GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1`, the canonical external stdlib, and a
45-second guard. All three passed their unchanged upstream Ruby verifier:

| Application | Bytecode time | Ruby time | Python time | Bytecode/Python |
| --- | ---: | ---: | ---: | ---: |
| I-Before-E | 0.580 s | 0.100 s | 0.130 s | 4.46x |
| Monte Carlo Pi | 2.940 s | 1.420 s | 1.680 s | 1.75x |
| Base64 guard | 3.270 s | 2.210 s | 3.310 s | 0.99x |

The direct bytecode runtime benchmark then loaded and warmed each application
before CPU profiling repeated `main()` calls. It used the same limits and the
same real external working directory/input. Retained profiles under
`v12/interpreters/go/.profiles/` are:

- `20260711_i_before_e_external_pair_30x.cpu.pprof` (7.90 seconds sampled)
- `20260711_monte_carlo_pi_external_pair_3x.cpu.pprof` (7.53 seconds sampled)
- `20260711_base64_external_pair_guard_3x.cpu.pprof` (8.13 seconds sampled)

| Application | Warmed result |
| --- | --- |
| I-Before-E | 269,262,136 ns/op; 9,061,881 B/op; 1,923 allocs/op |
| Monte Carlo Pi | 2,560,246,245 ns/op; 177,804,304 B/op; 22,222,114 allocs/op |

## Attribution

The profiles have no material shared leaf:

| Application | Material work |
| --- | --- |
| I-Before-E | `execCallMember` is 28.99% cumulative; `lookupCachedMemberMethodEntry` is 13.92%; inline returns, cache validation, and Array/string calls make up the rest of the VM-heavy path. |
| Monte Carlo Pi | Float-slot store/cast work dominates: `execStoreSlotCastSlotFloatConstDiv` is 36.65%, its discard fast path 33.07%, and the fused float compare is 16.47%; allocation/boxing is also material. |
| Base64 guard | Exact native calls reach codec encode/decode and MD5; those host kernels are 73.68% flat and the native-call chain is 82.90% cumulative. |

`runResumable` is common only as the bytecode dispatcher parent. The I-Before-E
member-cache work does not appear materially in either numeric/codec profile;
the Monte Carlo float lane does not appear in I-Before-E or Base64; and Base64
cannot justify a codec- or MD5-specific path. This reproduces the prior
cross-workload conclusion with current source and current verifier output.

## Decision

Keep no VM, compiler, tree-walker, or `able-stdlib` code. The profile pair
rejects a member-cache, raw float, Array, codec, or host-native candidate as a
general bytecode improvement. In particular, do not re-open the previously
rejected float store/cast variants merely because Monte Carlo remains a large
external miss.

## Next recommendation

Audit the benchmark catalog's feature coverage and verification status before
collecting another profile pair. The current verified bytecode misses are
already profiled and disjoint, so another profile-only loop would optimize
measurement noise. The audit should map v12 language/runtime features to
benchmark programs and cross-language verifier/reference coverage, then add or
complete only a missing application-shaped benchmark. A future performance
candidate must still repeat in two independently verified applications and an
unrelated guard.
