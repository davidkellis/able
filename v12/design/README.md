# v12 Design Notes

v12 is Go-first (tree-walker + bytecode). The TypeScript interpreter was removed
from the active toolchain. Some design notes still reference TypeScript/Bun
workflows for historical context; treat those references as archived unless a
future non-Go runtime is revived.

Authority map:
- `compiler-go-lowering-spec.md` supplies active Able -> Go guardrails and
  target shapes.
- `compiler-go-lowering-plan.md` is the historical compiler-completion record
  with active high-level guardrails.
- `compiler-native-lowering-guardrails.md` is the short-form compiler contract.
- `compiler-native-lowering.md` and `compiler-aot.md` are historical records.
- `spec/full_spec_v12.md` is the language authority when a design note and the
  implementation diverge.

Interpretation rules for the rest of this directory:
- documents that describe bytecode, parser, runtime, stdlib, or concurrency
  mechanics remain useful design notes, but they do not override the compiler
  lowering spec/plan;
- documents that still discuss TypeScript/Bun are historical or future-runtime
  notes unless they have been explicitly refreshed for the current Go-first
  toolchain;
- named stdlib/container examples in compiler docs are proof cases for shared
  lowering machinery, not licenses to add nominal-type-specific compiler rules.

Compiler architecture references:
- `compiler-go-lowering-spec.md`: active Able -> Go guardrails and target shapes
- `compiler-native-lowering-guardrails.md`: concise native-lowering contract
- `compiler-go-lowering-plan.md`: historical completion path and proof record
- `compiled-split-receiver-method-abi.md`: active dual-entry compiled instance-
  method ABI, compatibility boundary, and measured allocation rationale
- `compiler-noncapture-effect-audit.md`: closed proof requirements for any
  future loop-carried nominal storage reuse

Bytecode/runtime architecture references:
- `truthiness-cast-runtime-alignment.md`: current cross-mode Error truthiness,
  explicit-cast failure, and performance-evidence dependency record
- `bytecode-vm-v2.md`: concise active bytecode VM contract, boundaries, and
  performance-selection gate; the adjacent `*-historical-*.md` records retain
  the superseded architecture and experiment chronology
- `portable-vm-backend-abi.md`: checked decision closing C-ABI, portable-JIT,
  and direct-codegen backends against the current Go object graph; it defines
  the backend-neutral semantic-ABI prerequisite for reconsideration
- `shared-runtime-semantic-abi-feasibility.md`: conditional design for one
  pointer-free program/value/effect ABI and shared semantic heap; it admits
  only a standalone codec/layout spike, not a backend or runtime migration
- `semantic-abi-codec-layout-spike.md`: retained stage-zero pointer-free cell,
  deterministic image codec/validator, generated manifests, and three-function
  no-AST-fallback shadow evidence; execution-complete lowering remains next
- `semantic-abi-shadow-image-lowering.md`: retained version-2 typed register,
  call-target, CFG, match/raise, and host-resume images for three unlike whole
  functions
- `shared-value-heap-conformance-contract.md`: generated layouts for every
  shared/host kind plus the retained deterministic identity, mutation, root,
  tracing, cycle, and host-lifetime model
- `shared-value-heap-go-binding-conformance.md`: test-only identity-preserving
  current-Go graph adapter and historical record of the three gaps it found
- `shared-value-heap-contract-reconciliation.md`: owned wide scalars,
  inspectable Hasher state, and an explicit Iterator host-driver/root boundary;
  its checked 31-kind matrix admits one bounded live integration pilot
- `../docs/perf-baselines/2026-07-22-runtime-contract-performance-evidence-reconciliation.md`:
  checked 21-closure scope review and 120-process direct-consumer guard that
  permits the bounded live integration pilot without claiming a timing win
- `../docs/perf-baselines/2026-07-22-shared-value-heap-production-pilot.md`:
  checked rejection of a semantically correct call/return cell veneer after
  120 verified selection runs; future shared-runtime work must first identify
  a closed native-ownership region that removes rather than adds conversion
- `shared-runtime-closed-region-cutover.md`: checked closure of production
  shared-runtime migration; four of five material hot functions require
  interpreter-instance ownership, while the sole bounded primitive cut covers
  only one family against the required three
- `bytecode-frame-array-ownership.md`: release-disabled Array-provenance
  diagnostic boundary; its historical companion retains the sidecar proposal
  and rejected explicit-release evidence
- `array-handle-lifetime.md`: active generic ArrayStore last-owner contract;
  its historical companion retains the completed rollout and retention evidence
- `testing-cli-protocol.md`: active `able test` command and external-stdlib
  framework contract; its historical companion retains deferred protocol ideas
- `testing-plan.md`: active split between Go implementation verification and
  external-stdlib user tests; its historical companion retains the completed
  consolidation chronology
- `typechecker-plan.md`: active Go checker ownership, integration, and
  evidence gate; its historical companion retains the completed bootstrap and
  retired TypeScript/Bun roadmap
- `parser-ast-coverage.md`: active parser feature/fixture/test matrix;
  `parser-roadmap.md` and `parser-node-kind-inventory.md` retain the completed
  grammar bring-up sequence and must not be treated as live syntax backlogs
- `reexport-named-implementation-import-audit.md`: active package-import and
  source-re-export identity/privacy matrix
- `performance-competitiveness-vision.md`: active 95%-target scorecard and
  cross-application performance-candidate gate; its historical companion
  retains dated benchmark and experiment lessons
- `concurrent-document-pipeline-performance-gate.md`: portable feature-depth
  expansion and the independent repeated-evidence/profile admission decision
- `manifest-normalization-performance-gate.md`: serial files/callback/error
  interaction coverage and a three-application candidate rejected by guards
