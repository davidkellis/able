# Compiled static match-type expression gate — 2026-07-22

## Decision

Reject and revert both candidate scopes. No compiler, runtime, stdlib,
bytecode, language, benchmark, or WASM code change is retained.

The broad candidate emitted one package-level immutable AST value for every
statically rendered `bridge.MatchType` expression. It was semantically valid
and removed the profiled allocations, but Binary Trees regressed in both
ordering directions: 30.6550 seconds current versus 31.7717 seconds candidate
(+3.642%). That is too large for this project's unlike-control bar.

The narrowed candidate cached only native-union `try_from_value` expressions
and direct `Error` parameter expressions. This removed every Binary Trees
substitution, but repeated owner/control means remained mixed. It improved
Binary Event Log and Option/Result Config while slowing Manifest
Normalization, Policy Record Dispatch, and N-Body. The exact allocation win is
real, but it does not justify retaining a mixed wall-clock result.

## Admission and safety audit

Post-typed-key allocation profiles found the same generated AST construction
family in four unlike applications:

| Application | `NewIdentifier` objects | `NewSimpleTypeExpression` objects |
| --- | ---: | ---: |
| Binary Event Log | 106,496 | 106,496 |
| Option/Result Config | 116,064 | 116,064 |
| Manifest Normalization | 4,099 | 4,099 |
| Policy Record Dispatch | 2,051 | 2,051 |

The matcher audit found no mutation of these type-expression nodes.
`bridge.MatchType`, the bridge's interpreter-free matcher, and the
interpreter type matcher read the expression graph. The prototype therefore
shared only compiler-rendered immutable expressions; dynamic type inference
continued to construct dynamic nodes.

The implementation used the complete rendered expression as its identity and
emitted a deterministic package-level variable. It named no benchmark,
container, stdlib API, or user nominal type.

## Narrow allocation gate

Five direction-alternated exact main-phase counter runs per variant confirm
that the intended objects disappear:

| Application | Current objects | Candidate objects | Object delta | Current bytes | Candidate bytes | Byte delta |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Binary Event Log | 3,457,793.0 | 3,244,800.6 | -6.160% | 244,031,259.2 | 230,399,764.8 | -5.586% |
| Option/Result Config | 1,052,543.2 | 869,183.6 | -17.421% | 37,768,596.8 | 26,033,601.6 | -31.071% |
| Manifest Normalization | 917,505.8 | 909,415.6 | -0.882% | 43,397,881.6 | 43,889,289.6 | +1.132% |
| Policy Record Dispatch | 921,369.6 | 917,280.4 | -0.444% | 47,323,836.8 | 47,142,512.0 | -0.383% |

Manifest's allocated-byte counter remained volatile despite its stable object
reduction. This is another reason not to infer a broad speedup from the object
count alone.

## Narrow wall-clock gate

Every process used one logical CPU, `GOMEMLIMIT=1GiB`, `GOGC=50`, a
55-second per-process cap, and the public verifier. All samples are retained in
the companion JSON.

| Application | Runs/variant | Current mean | Candidate mean | Delta | Current median | Candidate median |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Binary Event Log | 20 | 0.5585 s | 0.5440 s | -2.596% | 0.555 s | 0.530 s |
| Option/Result Config | 20 | 0.1965 s | 0.1635 s | -16.794% | 0.180 s | 0.150 s |
| Manifest Normalization | 40 | 0.19075 s | 0.19350 s | +1.442% | 0.185 s | 0.190 s |
| Policy Record Dispatch | 40 | 0.2140 s | 0.2175 s | +1.636% | 0.210 s | 0.220 s |

Unlike controls were neutral-to-mixed:

| Control | Runs/variant | Current mean | Candidate mean | Delta |
| --- | ---: | ---: | ---: | ---: |
| N-Body | 30 | 0.16133 s | 0.16467 s | +2.066% |
| K-Nucleotide | 8 | 3.0150 s | 2.93375 s | -2.695% |
| Matrix Multiply | 16 | 1.21125 s | 1.21563 s | +0.361% |
| Mutex Ledger | 8 | 0.57125 s | 0.57375 s | +0.438% |
| Binary Trees | source identity | unchanged | unchanged | no eligible sites |

The positive estimates have noise intervals that cross zero, but Manifest,
Policy, and N-Body remained positive after expanded direction-reversed
cohorts. The established conservative bar rejects this pattern instead of
assuming the allocation reduction will eventually become a wall-clock win.

## Build and verification notes

Fresh generation completed within the cap for Option/Result and Manifest.
Binary Event Log and Policy generation each hit the 55-second cap. Their
candidate sources were therefore produced by mechanically applying the exact
tested generator transformation to retained generated sources. On
Option/Result, that transformation produced a byte-identical `compiled.go`
to the fresh generator output; the only omitted declarations were unused
validation-discovery globals outside the main phase.

The broad and narrow investigations recorded 788 verifier-backed benchmark
processes, with zero benchmark failures or timeouts. Focused native-result,
native-union, and interface-boundary compiler tests pass after the production
revert. The report JSON retains every decision sample and exact allocation
counter.

## Next direction

Refresh bounded bytecode profiles for split/join, iterator collect, and numeric
array/map applications, then admit only a larger VM wall that repeats in at
least three unlike applications. This is preferable to another type-expression
representation tweak: both global scopes have now demonstrated that removing a
small allocation leaf does not reliably improve broad wall time. The next
tranche should use one-process profiles under the existing memory/time caps,
compare raw integer extraction, map lookup, and return/type-match costs, and
prototype only the repeated generic owner with the strongest cross-program
weight.
