# Compiled generated static-closure reach census — 2026-07-22

## Decision

Close generated static-closure pruning as a current application-runtime
candidate. Retain no compiler, generated-runtime, bridge, canonical-stdlib,
application, benchmark, language, bytecode, or WASM execution change.

A fresh six-application normal-source/link census found no new linked closure.
It found two already-distinct classes:

1. 193-600 KiB of declaration-only generated Go per application is already
   absent from the linked executable. Omitting it may improve compiler
   throughput, but cannot improve application runtime, startup, linked size,
   or instruction-cache behavior by the admission rule used here.
2. Every executable retains the same interpreter-package initializer through
   the unconditional generic binary/unary operator helper definitions. A
   general reach-based candidate already removed precisely this root, cut
   binary size 35.8%-39.1%, improved short applications, and then regressed
   Binary Trees 3.6% across 15 paired verified runs. The later independent
   primitive-runtime boundary reproduced the failure at +5.07% mean and
   +7.36% median with 44% more collections. That candidate remains rejected;
   this tranche does not retry it.

All remaining linked generated callables, methods, interface entries, exports,
and package initializers are rooted by the language's registration/dynamic
semantic machinery. A benchmark input not exercising a registration does not
make the registered function unreachable for all valid inputs.

## Cohort and method

The analyzable cohort deliberately spans three serial and three concurrent
applications:

| Application | Contract shape | Scale role |
| --- | --- | --- |
| Fib | serial recursion | minimal/fast-build control |
| Option/Result Config | serial nominal/union/callable | boundary-heavy control |
| Word Frequency | serial file/text/map | one-CPU build-timeout control |
| Future Pipeline | goroutine/Future/channel | small concurrent control |
| Concurrent Document Pipeline | goroutine/file/nominal/callable | one-CPU build-timeout control |
| Mutex Ledger | goroutine/mutex/nominal | third independent concurrent control |

Normal generated output used the current compiler, canonical
`../able-stdlib`, source-root-only catalog policy where applicable,
`GOMEMLIMIT=1GiB`, `GOGC=50`, and a 55-second per-phase/process cap. The first
four applications except Word Frequency built and verified directly under the
initial bounded process. Word Frequency and Concurrent Document Pipeline
completed source generation but their one-CPU Go builds reached the cap; their
trees were built separately with four Go workers inside 55 seconds solely to
obtain linker evidence, then both binaries passed their catalog verifiers.
Those diagnostic build/run times are not scorecard or candidate timings.

Concurrent Event Routing was the intended large concurrent timeout control,
but it reached the 55-second source-generation bound before writing a complete
Go tree. It remains an explicit exclusion; Mutex Ledger supplies the required
third complete concurrent program. No fact is inferred from Event Routing's
partial attempt.

For each complete tree, the census inspected only generated root-package Go
files, parsed top-level function declarations, counted exact identifier reach,
classified declaration-only bodies, counted registration/init references, and
joined declarations to `go tool nm -size`. `GODEBUG=inittrace=1` launches used
the catalog CPU/executor/input contract and passed the public verifier.

## Generated and linked closure

| Application | Go files | Go bytes | Functions | Declaration-only functions / bytes | Linked declared functions | Binary bytes |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Fib | 15 | 1,106,734 | 1,526 | 422 / 235,878 | 731 | 14,068,816 |
| Option/Result Config | 23 | 1,428,238 | 1,848 | 428 / 210,532 | 953 | 14,964,568 |
| Word Frequency | 67 | 5,065,981 | 6,111 | 1,134 / 599,751 | 2,473 | 21,370,896 |
| Future Pipeline | 15 | 1,176,614 | 1,549 | 385 / 193,371 | 796 | 14,248,152 |
| Concurrent Document Pipeline | 55 | 4,663,210 | 5,580 | 1,066 / 567,493 | 2,284 | 20,172,248 |
| Mutex Ledger | 27 | 1,643,312 | 2,114 | 466 / 238,168 | 1,009 | 15,459,352 |

There are 259 exact declaration-only names common to all six applications.
They comprise 76 compiled-entry adapters, 42 extern wrappers, 41 function
thunks, 18 struct conversions, 18 nullable conversions, 12 callable
conversions, 11 Array conversions, five interface conversions, five union
conversions, four diagnostic helpers, five arithmetic helpers, and 22 other
runtime/helper functions. Go already omits these bodies from the executable.
They are a compile-throughput/source-size opportunity, not a binary-runtime
opportunity.

The large trees are not arbitrary copies of one cold helper. Their loaded
packages create 10-138 compiled-call registrations, 58-166 compiled-method
registrations, and 125-280 interface-dispatch registrations after excluding
the shared helper definition occurrence. Package initialization is separately
ordered and rooted. Removing those bodies based on the observed benchmark
input would break dynamic calls, method/interface selection, function values,
exports, or initialization for other valid inputs.

## Exact linked fixed cost

| Application | Interpreter symbols / bytes | Init clock | Init bytes | Init allocations |
| --- | ---: | ---: | ---: | ---: |
| Fib | 328 / 4,487,390 | 91 ms | 37,984,072 | 707,115 |
| Option/Result Config | 328 / 4,487,390 | 72 ms | 37,984,344 | 707,115 |
| Word Frequency | 423 / 4,587,486 | 74 ms | 37,984,184 | 707,115 |
| Future Pipeline | 328 / 4,487,390 | 51 ms | 37,984,712 | 707,118 |
| Concurrent Document Pipeline | 423 / 4,587,486 | 47 ms | 37,984,968 | 707,119 |
| Mutex Ledger | 328 / 4,487,390 | 72 ms | 37,984,616 | 707,118 |

All six generated root packages import the interpreter only for
`ApplyBinaryOperatorFast` and `ApplyUnaryOperatorFast` inside
`__able_binary_op` and `__able_unary_op`. Fib, Option/Result, Future Pipeline,
and Mutex Ledger contain no caller beyond those definitions and retain the
328-symbol initializer closure. Word Frequency and Concurrent Document
Pipeline each contain one actual binary-helper caller and retain 423 symbols.

This is exactly the root classified by the July 17 unused-helper and primitive
runtime-boundary gates. It is broad, linked, and removable, but it is not an
open candidate: removing its 38-MB initial live heap changes Go's GC goal and
repeatedly slows the stable allocation-heavy Binary Trees guard. Retaining
unused heap ballast, changing GC policy by workload, or selecting the import
by benchmark would violate the performance and generality rules.

## Why no candidate was admitted

- Declaration-only source is already linker-dead, so it cannot close a
  compiler-versus-Go application-runtime gap.
- Registration-rooted source is semantically reachable even when one measured
  input does not execute it. Current telemetry is reach evidence for that
  input, not a whole-language dead-code proof.
- The only new-looking linked-size/startup opportunity is identical to a
  general candidate already rejected twice by repeated application guards.
- Source-generation/build timeouts are real tooling evidence, but do not
  authorize relabeling compiler-throughput work as application-speed work.
- No named nominal, container, package, stdlib type, application, or benchmark
  rule was considered or added.

## Verification

- Six complete normal binaries pass their public catalog verifiers.
- Six `inittrace` launches pass the same verifiers.
- Every generated-tree and executable fingerprint is retained in the companion
  JSON.
- The Event Routing exclusion stopped at the 55-second source-generation
  bound; no test or process was extended past one minute.
- The current static-root, no-bootstrap, dynamic-main, and boundary-clean
  compiler tests pass.
- Compiler bridge tests pass.
- No compiler, runtime, stdlib, application, benchmark, or scorecard source
  changed.

## Next recommendation

Run a compile-throughput-only generated-helper emission gate, explicitly
separate from the application-runtime target.

Why: this census found 259 exact declaration-only helpers across all six
unlike programs and 193-600 KiB of already-unlinked Go per application, while
large normal trees exceed the one-minute single-worker build guardrail. That
is broad, current, generic evidence for compiler/source throughput. It is not
evidence for faster application execution, so the candidate must preserve the
linked closure and runtime heap shape exactly rather than reopening the
rejected interpreter-root removal.

What it entails: add generator-owned reach bookkeeping before rendering and
omit only helpers with no compiled body, registrar, export, dynamic-fallback,
interface, package-init, extern, or launcher reference. Keep the binary/unary
operator helper/import closure unchanged in this tranche. Prove generated
source coverage with dynamic-call, export/re-export, function-value,
interface-default, extern, package-init, and concurrent-entry fixtures. Across
the same six applications, compare repeated complete generation/build times,
source bytes, peak disk, and peak RSS; require identical application output,
interpreter init bytes/allocations, and non-build-ID linked symbol sets. Run
normal application guards only to demonstrate runtime neutrality, not to claim
a speedup. If source reduction is not broad or build throughput is neutral,
revert it and return directly to scorecard-driven runtime work. Continue to
exclude named nominal/container/application rules and WASM.
