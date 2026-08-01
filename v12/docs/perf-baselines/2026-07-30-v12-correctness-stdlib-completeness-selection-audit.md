# v12 correctness and stdlib completeness selection audit

Date: 2026-07-30

## Decision

Select no implementation change.

The current v12 specification, `spec/TODO_v12.md`, active non-WASM design
authorities, parser/AST coverage, executable fixtures, canonical stdlib tests,
and broad benchmark catalog expose no already-specified, reproducible
correctness or canonical-stdlib completeness defect.

Retain this documentation-only audit. Do not change the compiler, runtime,
tree-walker, bytecode VM, canonical stdlib, parser, AST, fixtures, benchmark
measurements, dependencies, frozen workspaces, or deferred WASM work.

## Authority reconciliation

The audit used these exact authority identities:

- `spec/full_spec_v12.md`:
  `4f0405b86c122993723e8617abd6f825d9a8ff858d4c72acaf4e33469452f080`;
- `spec/TODO_v12.md`:
  `f2ff4fcfdf2136aa0debd19f4fa7abc0c4c26af65e56976782355db6e727e4b6`;
- exec coverage index:
  `9d6140e35e5a2434a7b9b56632832dc5d4f6683109491b953abf336c0b15c388`;
- active parser coverage:
  `f96ea97ccf9ca4c54bdf64613921e0fdf7827a6d1fa8aa5d51c4ca9d8f628c9f`;
- active typechecker contract:
  `1dc0d0f58da11eb644777189f00fea935c43eb270b100846734af5b8bb28ad67`;
- active testing contract:
  `330ca8129e443ddeef94ad4d9007166c6e57572f8e9cfe4d17768df433a07ee4`;
- active compiler lowering specification:
  `bc06053f6c0efe628461a0ee05d4166b3c4890fc9d006e6b08a3e4ab8e7cefbd`;
- native-lowering guardrails:
  `6a36aeb7e3da455c6c28874ddf064ead26f03671fa9ec8a6cca97b3dda5f4f1e`;
  and
- interface dictionary execution record:
  `8c173d1d0cc838ac2afee321e7c265d3349237856e5743772efa2f587abe2a6b`.

The top-level v12 TODO reports no current compiler-AOT gap, no stdlib
externalization gap, and no regex gap. Its native-container section is
historical and explicitly selects no new array/container workstream. The
active parser, typechecker, testing, compiler, bytecode, stdlib, and
performance documents likewise contain no implementation queue that overrides
`PLAN.md`.

## Candidate classification

The unresolved or stale markers in the full specification divide into three
classes:

### Language decisions required

- block-comment syntax;
- variance syntax and the wider inference/coercion/HKT rules;
- public names and contracts for wrapping, saturating, and checked arithmetic
  helpers;
- interaction between explicit and implicit `for`/`Self` patterns in
  composite interfaces;
- static exhaustiveness/refutability rules for open interface sets such as
  `Error`;
- the concrete versus existential representation of range-expression
  results;
- recursive `Self` constraints; and
- shared-data race/ownership guidance for concurrency.

These are not implementation bugs because the specification does not yet
choose an observable result. No parser, checker, interpreter, compiler, or
stdlib change is authorized until the relevant language rule is decided.

### Minimum contract already complete

- `FutureError` has the required `details`, `message()`, and `cause()` surface
  in the canonical stdlib and all Go execution paths. The specification
  explicitly permits platform enrichment.
- Non-exhaustive runtime matching already raises the specified error, while
  the unresolved item concerns future static rules for open sets and guarded
  patterns.

### Stale broad or completed prose

- Named-implementation invocation is already specified in §10.3.3 and covered
  by explicit function-position dispatch fixtures. The §13.5 “TBD”
  cross-reference is stale prose, not missing behavior.
- Dictionary capture at interface upcast is specified in §10.3.4 and covered
  for defaults, generic methods, `Self` returns, private factory impls,
  inherited interfaces, and per-value/cache invalidation behavior.
- The full specification's broad standard-library, interface-dictionary,
  tooling, and package-manager TODO bullets do not identify one absent
  operation. Current canonical source and executable contracts supersede
  historical roadmap language.

None of these classes supplies an already-specified failing behavior suitable
for implementation.

## Executable evidence

The fixture-coverage checker reports 272 seeded fixtures, zero planned
fixtures, and 273 fixture directories.

Focused parser and typechecker controls pass for:

- generic and nested-generic composite interfaces;
- generic interface method substitution;
- implementation signature checking; and
- named-implementation direct, wildcard, renamed, collision, and privacy
  imports.

The following representative interface fixtures pass in tree-walker,
bytecode, and strict compiled execution:

- `10_01_interface_defaults_composites`;
- `10_04_interface_dispatch_defaults_generics`;
- `10_05_interface_named_impl_defaults`; and
- `13_08_static_interface_dictionary_capture`.

The canonical external stdlib suite passes in both tree-walker and bytecode
modes at sibling HEAD
`219eff222c28406487231713753641bc49ee5b9a`. It covers collections, persistent
collections, I/O, filesystem/path/temp, JSON, math/numerics, OS, random,
Rational/128-bit integers, regex replacement, String/StringBuilder, and the
user-test protocol. The sibling worktree was not modified.

The broad catalog remains complete:

- 66 portable applications;
- 67 canonical Able benchmark sources;
- one diagnostic benchmark source;
- 79 bounded local fixtures;
- 145 combined programs;
- 15 feature families and 16 normative sections; and
- zero insufficient portable operation-depth categories or actionable gaps.

The 132-row scorecard remains current with five Able/reference samples per
row. All 23 performance closures remain current, and the selector emits zero
bytes. Semantic selection therefore does not indirectly authorize a
performance tranche.

## Worktree and cleanup

The pre-audit fully expanded worktree contained exactly 42 paths: the prior
eight-path non-WASM documentation/test handoff and the unchanged 34-path
deferred WASM boundary. The index was empty, and local `HEAD` and
`origin/master` both remained
`c43c4e6825504cff6486ef6ce74aaed48aebc7e5`.

Verification created:

- a 266 MiB project-local Go cache;
- an empty project-local Go temporary directory;
- a 65 MiB `/tmp/able-v12-extern-go` cache; and
- a small disk-backed `/var/tmp` audit workspace.

The guarded project cleanup removed the first two and ended empty. The
task-owned extern cache had no open file owner and was removed exactly. The
audit workspace is removed after final verification.

This record and its JSON companion expand the non-WASM handoff to ten paths.
Its sorted newline-delimited list is 564 bytes with SHA-256
`f0c52b1e3ebda5d69bc8262f000cf29d00f7a42f0d55bd3cbc006d98286575cd`;
the equivalent NUL pathspec is 564 bytes with SHA-256
`4a4e29e99beb7cfea44678e20b99de94aeabc7b46c4810058b12ce96f94fc2a7`.
The final fully expanded worktree contains 44 paths: those ten paths plus the
byte-identical 34-path deferred WASM complement. Nothing is staged.

## Next

Run a spec-first composite-interface self-pattern contract tranche.

Why: the selection audit found no implementable correctness defect, but
§10.1.5 explicitly leaves the interaction between composite-interface
constituents' explicit and implicit `for` clauses unresolved. This is the
smallest bounded undecided language surface that directly affects native
interface carrier and adapter synthesis.

What it entails: inventory the currently accepted AST and checker forms;
define whether mixed explicit/implicit self patterns are rejected or how they
normalize; specify generic substitution and diagnostics; then add focused
parser, typechecker, tree-walker, bytecode, and strict compiled fixtures for
the chosen rule. Only after the rule is canonical should implementation
change.

Why it matters: a precise composite-interface contract lets both interpreters
and the compiler agree on one static interface shape, preventing unnecessary
dynamic/interface-boundary fallback while preserving shared nominal lowering.

Do not begin WASM work.
