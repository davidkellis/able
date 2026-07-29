# Compiled frontier admission closure

Date: 2026-07-29

## Decision

Retain no production change.

The published Go 1.26.5 scorecard and its checked performance frontier admit
no general compiled optimization candidate. All 63 compiled rows have five
successful Able processes and five successful equivalent-Go processes. Seven
meet the 95%-of-Go target, 56 miss it, the geometric-mean Able/Go ratio is
4.637116x, and positive target excess totals 5.353263 seconds.

All 63 strict generated dependency graphs remain interpreter-free. The
remaining gaps therefore do not come from accidentally entering the
tree-walking interpreter. The checked 23-closure evidence ledger has zero
invalidations, so the Go 1.26.5 timing refresh did not reopen an earlier
compiler, runtime, semantic-boundary, or native-carrier owner.

No compiler, generated-runtime, runtime, interpreter, bytecode VM, canonical
stdlib, benchmark, language, dependency, nominal-lowering, or WASM production
source changed. No prototype or A/B cohort was started because the
three-unlike-program admission gate failed first.

The machine-readable companion is
`2026-07-29-compiled-frontier-admission-closure.json`.

## Evidence contract

The reviewed repository commit is
`b28a896ab45a592ed3155908f07ae3324923fd25`, published on
`origin/master`. The governing inputs are:

| Evidence | SHA-256 |
| --- | --- |
| Go 1.26.5 compiled scorecard | `57936ab2bdc3eb205b96b88933d735c5051dca979a05f553bce6bf17bb5ff4fa` |
| Go 1.26.5 cross-mode frontier | `29105aa91d3f3db2c78aac3e27d960763892f1d61d891278db3fb681f6b4d3f7` |
| checked performance-closure ledger | `3589828d4cfb03ed7b536f1be6f22dfac34b1e0181c45839f444281ab5b14873` |

The dated frontier reproduces byte-for-byte from the selected scorecard,
selection manifest, evidence manifest, and stability manifest. The scorecard
evidence check confirms all 126 selected cross-mode rows have five successful
Able/reference samples. The ledger independently hashes relevant production,
stdlib, spec, benchmark-source, and evidence scopes.

## Compiled residual ranking

The three largest compiled groups contain 75.05% of the positive target
excess, but each fails a distinct admission requirement:

| Group | Misses | Excess | Share | Disposition |
| --- | ---: | ---: | ---: | --- |
| `compiled-text-map` | 7 | 1.655053s | 30.92% | closed: no shared leaf |
| `compiled-concurrency` | 23 | 1.493684s | 27.90% | closed: rejected candidate |
| `compiled-sudoku-quotient` | 1 | 0.868842s | 16.23% | closed: insufficient breadth |

### Text and generic-map work

The exact descendants split among checked arithmetic, UTF-8 and String work,
boxing, allocation, and required generic nominal storage semantics.
K-Nucleotide supplies most of this group's excess, but application arithmetic
and the runtime-backed generic Map boundary remain distinct owners.

A Map-specific rule would violate the prohibition on named-container and
non-primitive nominal lowering. A general typed generic-nominal storage design
would need to preserve identity, aliasing, equality, hashing, and dynamic
fallback semantics; the current profiles do not identify one local concrete
operation shared by three unlike programs.

### Concurrency environment recovery

`bridge.currentGID` and `runtime.Stack` are a real owner across 23 concurrency
rows. Prior profiles place that path at 85%-93% cumulative CPU in three unlike
applications. Removing it safely requires explicit execution context through
runtime-callable and captured-lambda entries, package-environment proof, and
dynamic compatibility entries.

That is the broad execution-context ABI route explicitly excluded by the
current handoff. Its earlier default-context experiment regressed unrelated
N-Body wall time by 54.7%. Channel-, Mutex-, Future-, Awaitable-, or
application-specific shortcuts would violate the generality rule.

### Sudoku quotient work

Sudoku's signed Euclidean division and generated search body remain material,
but the owner occurs in one application. Exact control and relational-proof
censuses found no material counterpart in two other unlike programs. It
therefore cannot justify a compiler rule.

## Remaining groups

The other compiled residuals are already classified:

- current-control applications have separate generated bodies;
- float and wide numeric candidates previously reversed sign or failed broad
  gates;
- byte-output rows expose no shared generated/runtime descendant;
- regex rows share one canonical NFA family rather than unlike semantics;
- iterator/control rows split after the retained native bound-method and
  generic-storage improvements; and
- target-meeting rows remain protected guards.

Grouping these owners under labels such as boxing, checks, allocation, or
native code would erase the concrete mechanism and semantic-reason
requirements.

## Verification

- `bench_scorecard_evidence_check.py` confirms 126 selected rows with five
  successful Able/reference samples each.
- The dated frontier check reproduces 126 rows and zero actionable groups.
- All seven focused performance-ledger tests pass, with one intentional skip.
- The checked performance-evidence ledger reports 23 closures and zero
  invalidations.
- The remaining worktree boundary is still the pre-existing deferred WASM
  set; this tranche did not inspect or modify those files.

## Next recommendation

Pause production performance mutation until a concrete admission invalidation
exists.

Why: every current residual owner is required, already rejected, below the
breadth gate, or lacks one exact shared concrete leaf. Repeating a closed
experiment would target one application, one nominal family, or a known broad
regression.

What it entails: run the checked evidence ledger after a retained compiler,
runtime, language, canonical-stdlib, benchmark-source, or broad-application
change. Refresh diagnostics-off CPU and allocation evidence only for an
invalidated closure. Advance production code only if one material mechanism
then repeats in at least three unlike verifier-backed applications and passes
five-or-more balanced A/B/reference pairs.

Why it is important: this keeps primitives and static Arrays in native Go
carriers and strict applications outside the interpreter without accumulating
benchmark-specific, named-container, non-primitive nominal, or unsafe ABI
rules. The 95%-of-Go goal remains active; this closure defines the evidence
required to resume implementation. Do not begin WASM work.
