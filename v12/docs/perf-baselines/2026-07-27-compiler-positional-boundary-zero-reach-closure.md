# Compiler positional-boundary zero-reach closure

## Decision

**rebase-compiler-scope-after-complete-zero-reach-proof**.

The positional-nominal correctness repair changed the compiler-production
scope and conservatively invalidated 12 compiled performance closures. A
strict census of all 63 portable applications proves that no current
application executes either corrected branch:

- 41 binaries retain the Ratio helper, but all 41 verifier-backed sentinel
  runs avoid the corrected Ratio field read;
- the remaining 22 binaries do not retain that helper; and
- zero applications emit or link map-literal spread code.

The sentinel positive control, the exact formerly failing Ratio fixture,
reaches the probe and panics with the expected marker. The zero application
hits are therefore evidence of no changed-path reach rather than a disabled
probe.

No fresh timing cohort is required because there is no reached application to
measure. The compiler scope is rebased without changing a scorecard number,
closure definition, closure disposition, evidence input, or selection
identity. This is an evidence closure, not a performance improvement.

Machine-readable summary:
`2026-07-27-compiler-positional-boundary-zero-reach-closure.json`.
Deterministic 63-row census:
`2026-07-27-compiler-positional-boundary-reach-census.tsv`.

## Method

One retained `ablec` binary strictly emitted every application selected by the
portable coverage catalog. Each emission used:

- the catalog-selected target and source-root policy;
- the canonical external stdlib;
- `--no-fallbacks`;
- `GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`;
- an individual 60-second compiler bound; and
- disk-backed `/var/tmp`.

Every generated module was independently checked with `go list -deps` and then
linked under an individual 60-second Go-build bound. All 63 graphs omit
`able/interpreter-go/pkg/interpreter`.

The deterministic census records source, generated-source, and binary hashes
plus four reach layers:

1. generated Ratio-helper occurrences;
2. final linked Ratio symbol and diagnostic presence;
3. a runtime sentinel placed exactly before the corrected `num` accessor; and
4. map-spread source and linked-diagnostic presence.

The runtime probe used Go's overlay facility. Original generated files were
not modified. Each linked application ran once with its catalog working
directory, program arguments, setup, executor, and public Ruby verifier.

The positive control used
`06_12_03_stdlib_numeric_ratio_divmod`, the fixture that originally exposed
the positional-storage defect. Its instrumented binary exited nonzero with:

```text
panic: __ABLE_RATIO_POSITIONAL_BOUNDARY_REACHED__
```

The census was generated twice and the outputs were byte-identical. Its
SHA-256 is
`72ee4aa385e6d07d5087bd57f4783ebbd5a2b77231317f7523caa8b2341d66c7`.

## Results

| Property | Result |
| --- | ---: |
| Portable applications | 63 |
| Strict emissions and linked binaries | 63 |
| Interpreter-free final graphs | 63 |
| Generated Ratio-helper references | 189 |
| Binaries retaining Ratio helper and diagnostic | 41 |
| Binaries without retained Ratio helper | 22 |
| Public-verifier sentinel runs | 41/41 pass |
| Corrected Ratio field-read hits | 0 |
| Ratio positive-control hits | 1 |
| Applications with generated map-spread sites | 0 |
| Applications with linked map-spread diagnostic | 0 |

The 41 linked helpers are retained through shared numeric-runtime paths. Their
presence does not mean the program passes a runtime `Ratio` value into the
helper. The exact sentinel distinguishes that registration/linkage reach from
the corrected semantic branch.

`rational_series` does use statically compiled `Ratio` carriers internally,
but its operations remain native `*Ratio` calls and do not retain the runtime
struct-conversion helper. It therefore does not exercise the corrected
runtime-value boundary.

## Measurement decision

The tranche required five-or-more verifier-backed Able/Go measurements for
every application that executes a corrected branch. The reached set is empty,
so the required timing-row count is zero.

Running Rational Series merely because of its name would not measure this
change. Running the 41 linker-retained applications would likewise time
unchanged paths. Either choice would manufacture an A/B claim without changed
execution.

The existing scorecard remains synchronized and complete:

- 119 selected rows and 126 full-status rows;
- five successful Able/reference samples for every selected row;
- 20 retained source reports and 33 retained reference reports; and
- unchanged selection identity
  `e6a6ccacc9620f9e1b89e2510cab52a85114ddbf2f41e33abae7c6d8a70241f8`.

## Evidence rebase

A trial closure-ledger bootstrap changed only the compiler-production tree
hash:

```text
cc6df4aa7f65a2de5c3053be1cc216bb48d66d8e92b476d9aa338af6380f35a5
  ->
30e71f61e48405c87191288c7683da108e5de641d314ac1570f778ac4770b13d
```

All 21 closure records, frontier inputs, and selection identity were
byte-equivalent. The exact bootstrap was then applied. The checked selector
now reports 21 current closures and zero invalidations.

## Retained scope

Retained:

- the general compiler correctness fixes and regression guards from the prior
  release tranche;
- bounded compiler release sharding;
- the deterministic reach-census tool;
- the 63-row census;
- this decision pair; and
- the compiler-scope evidence rebase.

No additional compiler, generated-runtime, runtime, interpreter, bytecode VM,
language, canonical-stdlib, benchmark workload, dependency, or WASM production
change advanced.

Cleanup removed the exact 5.6 GiB generated/sentinel workspace and 4.9 GiB
reproducible Go cache. No RAM-backed `/tmp/able-*` path remains; the small
disk-backed reusable extern cache was preserved.

## Next recommendation

Audit remaining generated direct reads from
`StructInstanceValue.Fields` and classify them before changing code.

Why: Ratio and map spread were two instances of the same representation
mistake. Other direct reads may be intentional writes, legacy map-backed
services, or latent semantic reads that will fail when given positional
nominals.

What it entails: enumerate every generated `Fields[...]` read, distinguish
writes and proven map-only values from semantic named-field reads, map each
candidate to unlike fixtures/applications, and migrate only representation-
agnostic semantic reads to `__able_struct_named_field_value`. Require focused
positive controls plus the full compiler release matrices; if no remaining
unsafe read exists, retain no production code.

Why it is important: this prevents another correctness failure at the
compiled/runtime boundary while preserving native positional carriers and
avoiding broad boxing or named-container lowering rules. Do not begin WASM
work.
