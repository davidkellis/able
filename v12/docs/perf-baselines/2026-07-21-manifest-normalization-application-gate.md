# Manifest normalization application gate — 2026-07-21

## Decision

Retain the portable Manifest Normalization application, its source-equivalent
Go/Python/Ruby implementations, common input and verifier, catalog/coverage
memberships, selected compiled and bytecode rows, and current profile evidence.

Retain no compiler, generated-runtime, bytecode VM, canonical-stdlib, language,
or WASM performance change. Bytecode repeats closed owners. Compiled execution
admitted one generic String byte-conversion candidate, but its mixed-sign broad
wall-time gate required a complete revert.

## Application contract

The program reads 32 service manifests for 128 rounds. Each valid record passes
field validation, resolves an optional owner through `Option`, and is transformed
by a captured normalizer. Invalid records return an ordinary `Result` error.
The 24 valid and eight rejected records per round produce:

```text
4096:3072:1024:1024,1024,1024:1577539338:803938
```

Able, Go 1.26, Python 3.14, and Ruby 4.0 implement the same serial batch
transformation. The sibling Able source has only a distinct package name to
avoid loader collisions; its executable body is identical to the canonical
source. All runtime paths complete below the one-minute process cap.

The target interaction—expressions/files × closures/callables ×
Option/Result—now has three unlike applications: Concurrent Event Routing,
Manifest Normalization, and Policy Record Dispatch. Across all 165 weighted
triples, minimum depth remains two; six triples remain at depth two.

## Repeated measurements

All successful workstation samples are retained. Able has three independent
five-process cohorts per mode. Go, Python, and Ruby each have two five-process
cohorts.

| Lane | Processes | Pooled mean | CV | Limiting ratio |
| --- | ---: | ---: | ---: | ---: |
| Able compiled | 15 | 0.210000 s | 11.94% | 41.96x Go |
| Go | 10 | 0.00500455 s | 10.38% | — |
| Able bytecode | 15 | 1.753333 s | 5.59% | 71.22x Python / 32.15x Ruby |
| Python | 10 | 0.0246188 s | 23.85% | — |
| Ruby | 10 | 0.0545277 s | 11.16% | — |

All 60 normal timing processes verified with zero failures and timeouts. The
checked mode-specific promotion rows are compiled 0.216 seconds versus Go
0.0054 (40.00x), and bytecode 1.784 seconds versus Python 0.0227 (78.59x) and
Ruby 0.0536 (33.28x). The native startup spread cannot change either miss.

## Exact profiles

Three verified compiled main-phase profiles merge to 570 ms. The exact
`__able_string_to_builtin_impl` path is 15.79% cumulative, including per-byte
integer extraction and `fmt.Sprintf`. This makes String byte conversion
material in a third unlike application after K-Nucleotide and Policy Record
Dispatch.

One loaded/lowered/warmed bytecode artifact measured three `main()` calls at
1,339,783,125 ns/op, 86,717,941 B/op, and 1,238,790 allocs/op. Its 4.43 seconds
of CPU samples show the VM loop, Go map matching, call-frame push, raw integer
information, stack/slot storage, Array get, call dispatch, and inline return.
Every concrete VM child belongs to an already-completed broad family; no new
bytecode candidate was admitted.

## Generic candidate and rejection

The compiled candidate replaced eager `fmt.Sprintf("array element %d", idx)`
on every successful String byte with indexed integer extraction that formats
the same diagnostic only on failure. It was a primitive String/kernel boundary,
not a nominal or benchmark-specific fast path, and focused generated-source
tests preserved the error contract.

Every candidate and control process verified. The order-balanced owner gate
retained 35 additional processes:

| Application | Candidate mean | Current/control mean | Change |
| --- | ---: | ---: | ---: |
| Manifest Normalization | 0.212 s (5) | 0.210 s (15) | 1.0% slower |
| K-Nucleotide | 2.873 s (10) | 2.918 s (5) | 1.5% faster |
| Policy Record Dispatch | 0.210 s (10) | 0.204 s (5) | 2.9% slower |

Two of three material owners regressed and the sole improvement was small.
The candidate and its test were fully removed. Canonical `able-stdlib` was not
changed because no specified reusable API or behavior gap was found.

## Verification

- normal Able typecheck;
- tree-walker, bytecode, and compiled output parity;
- exact Go, Python, and Ruby parity through one verifier;
- 60 retained normal timing processes and 35 retained candidate/control
  processes;
- three verified compiled main profiles and one bounded warmed bytecode profile;
- catalog, coverage, selection, scorecard, triple, frontier, and evidence-ledger
  reconciliation;
- source files below 1,000 lines and `git diff --check`.

## Next recommendation

Audit the remaining depth-two concurrency × expressions/files × Option/Result
interaction, represented only by Concurrent Event Routing and Concurrent Text
Index, and prefer existing evidence before adding another application.

Why: it is now the highest-ranked shallow interaction, but compiled concurrency
already has a closed goroutine-identity wall. Inspecting existing exact paths
first avoids manufacturing a third benchmark that merely repeats that wall.

What it entails: compare the two applications' source operations and current
compiled/bytecode profiles; annotate another existing portable application only
if it genuinely exercises all three families. If breadth is still only two,
add a bounded non-routing file-validation workload with source-equivalent
Able/Go/Python/Ruby programs and one verifier. Measure repeated cohorts and
admit code only for a new concrete non-parent leaf in three unlike programs.
Continue to defer WASM.
