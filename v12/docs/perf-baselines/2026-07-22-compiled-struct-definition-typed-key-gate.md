# Compiled struct-definition typed-key gate — 2026-07-22

## Decision

Retain the general bridge change. `bridge.Runtime` now keys its nominal struct
definition cache with the comparable pair `(environment pointer, name)` rather
than formatting `"<pointer>:<name>"` for every lookup. This preserves
environment scoping and qualified-name behavior while removing two allocations,
formatting, and string hashing from each warm lookup. The rule names no Able
nominal type, container, stdlib API, or benchmark.

The nominal-definition code was also moved from the 1,049-line `bridge.go` into
a 157-line focused module. `bridge.go` is now 902 lines. No canonical-stdlib,
bytecode VM, language, benchmark, or WASM change was needed.

## Admission and profile protocol

Fifty verifier-backed main-only CPU profiles were refreshed from the retained
lazy-environment ABI binaries: five Binary Event Log profiles and fifteen each
for Option/Result Config, Manifest Normalization, and Policy Record Dispatch.
They merged to 2.63, 1.51, 1.61, and 1.89 seconds of samples. All four put
allocation/GC at the common wall, but their generated Able bodies differed.

One-object main-phase allocation subtraction then identified the same concrete
leaf in all four owners. `Runtime.StructDefinition`'s formatted cache key owned
557,060 objects in Binary Event Log, 12,866 in Option/Result, 38,920 in
Manifest, and 28,680 in Policy. This met the three-unlike-application admission
rule before implementation.

All benchmark launches used one logical CPU, `GOMEMLIMIT=1GiB`, `GOGC=50`, and
a 55-second cap. Every result passed the public suite verifier. Direction-
reversed cohorts retained every workstation sample.

## Allocation gate

Five lightweight exact-counter runs per candidate application show that the
predicted objects disappear. Allocated bytes also fall in every owner.

| Application | Current objects | Candidate objects | Object delta | Current bytes | Candidate bytes | Byte delta |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Binary Event Log | 4,014,876.4 | 3,457,792.8 | -13.875% | 261,858,484.8 | 244,031,233.6 | -6.808% |
| Option/Result Config | 1,065,408.4 | 1,052,543.4 | -1.208% | 38,180,118.4 | 37,768,609.6 | -1.078% |
| Manifest Normalization | 956,391.4 | 917,488.4 | -4.068% | 44,630,545.6 | 43,246,787.2 | -3.100% |
| Policy Record Dispatch | 950,050.6 | 921,374.6 | -3.018% | 48,470,915.2 | 47,357,025.6 | -2.298% |

Post-change one-object profiles contain no `structCacheKey` allocation. The
remaining `fmt.Sprintf` samples in Manifest and Policy have different owners
and are not evidence that the removed key remains.

## Wall-clock gate

Each owner used twenty processes per variant, split evenly across forward and
reverse ordering.

| Application | Current mean | Candidate mean | Delta | Current median | Candidate median |
| --- | ---: | ---: | ---: | ---: | ---: |
| Binary Event Log | 0.6395 s | 0.5510 s | -13.839% | 0.610 s | 0.540 s |
| Option/Result Config | 0.1495 s | 0.1475 s | -1.338% | 0.150 s | 0.150 s |
| Manifest Normalization | 0.1785 s | 0.1660 s | -7.003% | 0.180 s | 0.160 s |
| Policy Record Dispatch | 0.2125 s | 0.1990 s | -6.353% | 0.190 s | 0.190 s |

Five unlike compiled controls are neutral-to-better. Binary Trees used six
processes per variant because each process is about thirty seconds; N-Body
used ten, and the other controls used eight.

| Control | Current mean | Candidate mean | Delta | Current median | Candidate median |
| --- | ---: | ---: | ---: | ---: | ---: |
| N-Body | 0.1650 s | 0.1480 s | -10.303% | 0.150 s | 0.145 s |
| Binary Trees | 30.6633 s | 30.7283 s | +0.212% | 30.665 s | 30.655 s |
| K-Nucleotide | 3.0363 s | 3.0063 s | -0.988% | 3.075 s | 3.025 s |
| Matrix Multiply | 1.1588 s | 1.1525 s | -0.539% | 1.160 s | 1.165 s |
| Mutex Ledger | 0.5575 s | 0.5588 s | +0.224% | 0.550 s | 0.555 s |

The two positive control estimates are below the timer's useful workstation
resolution and are bracketed by improvements in the other controls. In total,
318 benchmark/profile processes verified with no benchmark failure or timeout.

## Verification and harness notes

The full bridge package, focused struct/standalone/generated-union compiler
tests, and focused interpreter tests pass. The focused compiler group completed
in 55.95 seconds. A monolithic compiler-package attempt was deliberately not
given a larger timeout: it reached the required 60-second cap while actively
compiling the unrelated iterator-pipeline test. An initial fresh `ablec` build
also reached the 55-second process cap, so the runtime-only candidate binaries
were rebuilt from the retained identical generated sources with the changed
bridge module. Neither diagnostic attempt is benchmark evidence.

## Next direction

Audit generated static type-expression lifetime, beginning with the repeated
`ast.Ty("Error")` construction in native-union `try_from_value` helpers. The
post-change exact profiles attribute 106,496, 116,064, 4,099, and 2,051
`NewIdentifier` objects to this family in the four unlike owners, with an equal
number of `NewSimpleTypeExpression` objects. A generator-wide immutable static
type table could remove those objects from every matching union conversion,
but only if the audit proves no matcher mutates the AST nodes and package/name
resolution remains identical. Prototype one shared representation, test
qualified/generic/nullable/error matching, then use these owners and unlike
controls for the same repeated allocation and wall-clock gate.
