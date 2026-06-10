# Static compiled-program metadata-retention audit — 2026-07-14

## Decision

Accepted and implemented: omit only package-interface `DefaultImpl` block
payloads from a generated executable when the compiler has selected its
no-bootstrap static launcher. Keep every other definition and signature field,
and retain the current payloads unchanged in any target that requires
interpreter bootstrap.

This is not an AST re-encoding attempt and not a `HashMap`, iterator, text, or
canonical-stdlib special case. It is a target-wide retention rule following the
language's static-versus-dynamic execution boundary.

## Language and launcher boundary

The v12 specification requires statically resolved compiled code to execute
without interpreter involvement; dynamic features are the explicit exception.
The generated launcher already makes that distinction once:

| Target condition | Launcher behavior | Metadata rule |
| --- | --- | --- |
| No dynamic features, fallbacks, or unseedable imports | `RegisterIn(nil, entryEnv)` then compiled `RunRegisteredMain` | Candidate may omit package default-body AST payloads. |
| Dynamic feature, fallback, or unseedable import | creates an interpreter and evaluates the program before `RegisterIn` | Preserve current metadata exactly. |

The metadata mode must be derived from this one launcher decision. A generated
runtime `if interp != nil` is insufficient: it still emits JSON literals and
the decoder into static source, and it risks two definitions of the boundary.

## Current payload inventory

The package-definition generator directly constructs struct, union, interface
signature, generic-constraint, where-clause, base-interface, and visibility
metadata. Those fields remain needed by static type matching, native interface
adapters, generic constraints, constructors, diagnostics, and import seeding.
They do not use JSON decoding.

`DecodeNodeJSON` is emitted by three helpers, with different retention rules:

| Emitter | Current use | Static no-bootstrap disposition |
| --- | --- | --- |
| `renderBlockExpressionExpr` in a package interface signature | Reconstructs `FunctionSignature.DefaultImpl` during `RegisterIn` | Candidate: omit only this block from the runtime interface node. |
| `renderBlockExpressionExpr` in a local interface definition | Builds a runtime-local interface value while user code executes | Retain; local-default static semantics need their own proof. |
| `renderMethodsDefinitionExpr` / `renderImplementationDefinitionExpr` | Local declarations call interpreter-backed bridge registration | Retain; the bridge explicitly requires an interpreter. |

Thus the historical phrase “definitions, methods, implementations, and
defaults” is too broad for the current package-start path: the current static
startup decoder work is package-interface default bodies. Local methods and
implementation definitions are not emitted by package registration.

## Concrete generated-binary evidence

A fresh, no-bootstrap Word Frequency build with canonical external
`able-stdlib` produced 1,497 generated Go files (201 MiB before immediate
cleanup). It emits 39 package-start `DecodeNodeJSON` calls, all in interface
default signatures:

| Generated package definition file | Calls |
| --- | ---: |
| `able.collections.enumerable` | 16 |
| `able.collections.map` | 1 |
| `able.core.iteration` | 6 |
| `able.kernel` | 16 |

The previous compiled coverage profiles independently showed this decoder
family under `RegisterIn` in short real applications. The generated source
inspection makes the object of a safe retention audit concrete without
attributing the whole startup cost to one benchmark.

## Why the candidate preserves static default dispatch

The compiler retains the original default-body AST while compiling. Before
rendering, `collectDefaultImplMethods` turns an omitted implementation method
into a compiled function, and native interface adapters specialize the same
default body for static calls. The runtime `FunctionSignature.DefaultImpl`
field is used by the interpreter to build an interpreted default method; the
static path has no interpreter and invokes the compiled helper/dictionary
entry instead.

The no-bootstrap execution fixture
`10_15_interface_default_generic_method` passed from the current tree. It
uses a generic interface default through an interface dictionary and therefore
proves the static dispatch route is active. Existing compiler source audits
also assert direct compiled default-method calls for generic and sibling-method
cases.

This proof is deliberately limited to package-level interface defaults. It
does not authorize eliding local interface default bodies, local methods, or
local implementation declarations.

## Rejected alternatives and constraints

The two previous whole-contract approaches remain rejected:

- Recursive Go constructors made normal generated application builds spend
  more than six CPU minutes parsing emitted source.
- A complete tagged decoder was neutral for Document Audit and Lexical Rollup
  and regressed the independent I-Before-E process guard by 2.49%.

The candidate avoids both mechanisms: it emits less metadata only where the
static ABI cannot observe the field. It must not alter dynamic-program
metadata, diagnostic source mapping, signature types, constraints, base
interfaces, overload rules, or the shared nominal lowering pipeline.

## Implementation and validation

The implementation centralizes the exact bootstrap decision in
`requiresBootstrapExecution`. The launcher and package-definition renderer
both consume that decision. Library output, which has no generated launcher,
retains existing interface default metadata. The static renderer emits `nil`
for `FunctionSignature.DefaultImpl`; compiled default-method helpers still
come from the original compiler AST.

The following guards passed:

1. New source tests prove that a static launcher omits the package default AST
   while retaining the interface signature, and a fallback launcher retains
   the decoded AST payload.
2. The no-bootstrap generic dictionary-default fixture
   `10_15_interface_default_generic_method`, default-to-sibling compilation,
   generic constraints, and local interface/method/implementation metadata
   guards all pass.
3. A fresh canonical-stdlib Word Frequency static build uses
   `RegisterIn(nil, entryEnv)` and has zero `DecodeNodeJSON` calls in its
   generated `compiled*.go` and `main.go` files, down from the prior inventory
   of 39 package-default calls. A fresh build completed in 23.33 seconds;
   this is a completion guard, not a cross-machine build-time claim.
4. CPU-15, Ruby-verifier-backed coverage runs improved at three runs each:
   Word Frequency `0.2067s` (from `0.2100s`), Document Audit `0.0733s` (from
   `0.0800s`), and Lexical Rollup `0.1000s` (from `0.1200s`).
5. The one-run CPU-15 compiled generality suite verified every previously
   supported row. Its single pre-existing Sudoku timeout remains a timeout;
   no new failure or broad regression appeared.

The reproducible commands and scorecard outcomes are recorded in
`v12/docs/perf-baselines/2026-07-14-compiled-static-metadata-retention-gate.md`.

## Prototype admission criteria

The accepted implementation met all of the following evidence requirements:

1. Centralize the exact no-bootstrap decision before both launcher and package
   definition rendering; do not duplicate a looser dynamic-feature predicate.
2. In a static target, package default signatures render `nil` for
   `DefaultImpl` and static package definition files contain no
   `DecodeNodeJSON` calls. Dynamic/bootstrap targets retain the current payload
   byte-for-byte in behavior.
3. Add source and execution coverage for generic dictionary defaults,
   default-to-sibling calls, static constraints/base interfaces, and a dynamic
   program that still requires interpreter defaults.
4. Require compiler build-time checks plus output-verified, CPU-pinned
   compiled coverage and generality scorecards. Retain the change only when it
   has no material broad regression; do not accept a one-application startup
   win.

## Follow-up profile result and next recommendation

The bounded generated-main refresh is complete. Word Frequency repeats a
`__able_hash_map_find_entry` probe, Lexical Rollup is file/iterator/allocation
shaped, and Document Audit's user main is below phase-profiler sampling
resolution after the startup improvement. No leaf repeats across the three, so
no source experiment followed. Record:
`v12/docs/perf-baselines/2026-07-14-compiled-static-metadata-main-profile-refresh.md`.

That K-Nucleotide follow-up is complete and retains no source change. It calls
the same generated map-entry scan as Word Frequency, but only at 10.04%
cumulative; primitive conversion, allocation, and raw map get/set dominate
instead. The previous collision-safe index and lazy-index trials already
regressed real K-Nucleotide because its small maps make the representation
branch costlier than the scans. Do not retry either design or add a compiler
nominal-container branch. Record:
`v12/docs/perf-baselines/2026-07-14-compiled-static-metadata-main-profile-refresh.md`.

No further unchanged compiler map/profile pass is eligible. Keep the verified
scorecards as regression guards and return to an unfinished semantic or
portability boundary with fixture parity. Resume performance selection only
when it produces a concrete leaf that repeats across unlike real applications;
the current roadmap has no honest new cross-language timing row to manufacture.
