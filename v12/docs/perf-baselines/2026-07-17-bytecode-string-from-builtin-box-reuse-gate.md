# Bytecode String-to-byte box-reuse gate — 2026-07-17

## Decision

Keep no interpreter, runtime, kernel, compiler, or canonical-stdlib change.

Fresh profiles established the exact `__able_String_from_builtin` byte-boxing
leaf in three independent applications. Reusing the VM's immutable cached
`u8` values is semantically general and reduces allocation in Word Frequency
and Regex Stream, but the final direct-lookup candidate regressed the
conversion-free Reverse Complement guard repeatedly. The broad-benefit bar
therefore fails, and all candidate code and tests were reverted.

## Profile admission

All runtime-harness profiles used the current bytecode test binary,
`GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`, canonical external
`able-stdlib`, a warmed `main`, and a bounded call count. K-Nucleotide used one
ordinary verifier-backed CLI process because a warmed harness call does not fit
the 60-second test limit.

| Workload | Runtime sample | Exact `initStringHostBuiltins.func5` allocation |
| --- | ---: | ---: |
| K-Nucleotide | 39.87 s CPU | 401.11 MiB flat / 473.57 MiB cumulative; 6,712,038 objects flat / 7,701,335 cumulative |
| Word Frequency | 10.54 s CPU | 72.05 MiB flat / 87.30 MiB cumulative; 1,213,780 objects flat / 1,457,473 cumulative |
| Regex Stream Audit | 6.17 s CPU | 1.50 MiB flat / 6.04 MiB cumulative; 32,769 objects flat / 113,654 cumulative |
| Reverse Complement | 19.89 s CPU | no material String host conversion |
| Base64 | 7.46 s CPU | no material String host conversion |
| Lexical Rollup | 2.69 s CPU | no String host conversion; its line reader remains in the extern fast path |

The repeated leaf is the conversion loop that stores a freshly boxed
`runtime.IntegerValue{u8}` interface for every byte. This satisfied the
three-application admission rule without relying on a benchmark name or a
nominal-container special case.

## Candidate sequence

Three generic implementations were tested.

1. Returning the existing monomorphic `u8` array removed the conversion boxes
   but changed downstream representation costs. Word rose from 631,154 to
   994,483 allocations per call and became about 2% slower; its mono-array
   reads alone materialized 4.15 million boxed values. Regex also slowed. This
   representation candidate was reverted.
2. Keeping the dynamic array while calling the general cached-integer
   dispatcher reduced Word allocations, but its per-byte kind/range dispatch
   made the three-run Regex mean 4.04% slower. This candidate was reverted.
3. Keeping the dynamic array and directly indexing the existing immutable
   `u8` cache removed that dispatch. This is the final candidate measured
   below; it was also reverted after the guard gate.

## Repeated final-candidate gate

The runtime lanes used independent, order-balanced processes. Each reported
Word sample averages eight warmed calls; each Regex sample averages two. K
uses verified ordinary processes. The workstation was active, so the decision
uses repeated means and adjacent-pair direction rather than a single sample.

| Workload | Baseline | Candidate | Allocation result | Timing result |
| --- | ---: | ---: | ---: | ---: |
| Word Frequency | 1.2100 s | 1.1986 s | 631,131 -> 500,058 objects/call (-20.77%) | 0.94% faster |
| Regex Stream Audit | 3.2789 s | 3.1844 s | 858,243 -> 839,141 objects/call (-2.23%) | 2.88% faster |
| K-Nucleotide | adjacent pairs | adjacent pairs | peak RSS consistently 5–7% lower | pair ratios +1.06%, -4.61%, +0.51%; neutral/volatile |
| Base64 guard | 2.4918 s | 2.5093 s | effectively identical | 0.70% slower; noise-level |
| Reverse Complement guard | 6.4859 s | 6.7436 s | effectively identical | 3.97% slower |

A fresh adjacent Reverse baseline/candidate pair was 6.4652 s versus 6.8118 s,
again 5.36% slower. Reverse's clean profile contains no material
`__able_String_from_builtin` call, so this is not saved work being traded for
other work. Whether the effect is binary layout or host noise, it repeats too
strongly to accept a broadly enabled runtime change. The unrelated
numeric/nominal guard was not run after this decisive rejection.

K-Nucleotide output was verified after every ordinary process. Focused String
host tests pass, and the final reverted-source VM/String selection suite passes
in 17.626 s. A whole-package `go test ./pkg/interpreter -timeout=60s` reached
the existing fixture-parity workload `14_15_regex_function_replace` at the
suite timeout; it reported no candidate-related failure before timing out.

## Next recommendation

Build a bytecode scalar-character generality gate around the remaining large
String host leaf, `__able_char_to_codepoint`, using Regex Stream plus two
ordinary non-regex character-processing programs. First inventory whether the
current external suite materially exercises `char` traversal, Unicode scalar
conversion, and case folding; add one cross-language benchmark if that feature
coverage is missing. Then collect bounded CPU/allocation profiles and admit a
primitive-boundary candidate only if the same exact leaf is material in three
independent algorithms.

Why: Regex Stream attributes 19.20% of sampled allocation bytes and 25.04% of
objects to the relevant String-host closure, making it the next concrete text
wall, but regex variants alone are not independent evidence. The work entails
coverage inventory, at most one benchmark-suite expansion, fresh bounded
profiles, and repeated Word/Regex/K plus Reverse/Base64/numeric guards for any
generic candidate. WASM remains deferred.
