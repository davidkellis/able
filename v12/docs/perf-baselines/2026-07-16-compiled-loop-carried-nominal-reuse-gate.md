# Compiled loop-carried nominal reuse gate

Date: 2026-07-16

## Decision

Keep one independent benchmark fixture and one semantic regression test, but
make no compiler result-storage change. Loop-carried nominal reuse has a
material ceiling, yet the same statically safe ownership shape does not repeat
across the required three unlike programs and two nominal definitions.

The direct signed and unsigned accumulation fixtures are one related workload
family. A new user-defined `RecurrenceState` benchmark independently repeats
their direct replacement shape, but Fixed Width does not: one worker replaces
its carried value through a conditional fresh-result branch, while the other
conditionally adopts a candidate that may remain `best` for arbitrarily many
iterations. A single reusable candidate slot would mutate the still-live best
object before the next comparison.

No `Int128`, `UInt128`, Fixed Width, benchmark, package, or nominal-specific
rule was added.

## Independent recurrence fixture

`v12/fixtures/bench/nominal_recurrence_small` defines a plain user nominal with
two `i64` fields and advances it 500,000 times through a standalone function.
It has no dependency on the wide-numeric stdlib definitions.

The retained caller-owned result ABI produces:

| Main bytes | Main allocations | Exact repeated owner |
| ---: | ---: | --- |
| 8,000,400 | 500,013 | one loop-local `RecurrenceState` result slot promoted per iteration |

Generated Go escape analysis reports that the `_into` result pointer leaks to
its returned result and the loop-local slot moves to the heap because the
carried `state` pointer survives the iteration. The fixture prints
`200349387` and `830330818`; compiled, bytecode, and tree-walker execution all
agree.

## Temporary ceiling experiment

A generated-source-only experiment passed the current `state` pointer back as
the `_into` result slot. This is not retained compiler code. For this audited
function, every old field is read into scalar temporaries before the final
two-field write, and no alias of `state` exists.

The experiment preserves SHA-256
`5b505e4cbfdaaa9999417849dd7724f9882609cadf0d37f7ab5f605db897bc7e`
and changes the exact main phase from 8,000,400 bytes / 500,013 allocations to
264 bytes / 11 allocations. Twenty-five alternating processes improve from
82.067 ms to 73.421 ms (-10.54%). This confirms materiality without claiming a
general semantic proof.

## Cross-program lifetime reconciliation

| Workload | Carried nominal | Update shape | Reuse requirement |
| --- | --- | --- | --- |
| signed accumulation | `Int128` | unconditional fresh call chain replaces `acc` | prove receiver and every intermediate cannot be retained |
| unsigned accumulation | `UInt128` | same related accumulation shape | same proof; not an independent application family |
| user recurrence | `RecurrenceState` | unconditional fresh function result replaces `state` | prove the input parameter and result have no aliases or captures |
| Fixed Width modular add | `UInt128` | `sum` is conditionally adopted or replaced by `sum.sub(...)` | track both branch results and prove the old value dies on every branch |
| Fixed Width ordered select | `UInt128` | a fresh candidate is adopted only when greater | track which candidate slot owns `best`; it may remain live indefinitely |

Only the direct recurrence shape repeats in the related signed/unsigned pair
and the independent user fixture. Fixed Width supplies a second nominal
definition across the cohort but not the same lifetime rule. The required
three-unlike-program proof therefore fails.

## Why a local syntactic rule is unsafe

Able structs have reference semantics. The source spelling
`state = advance(state)` does not prove that `state` is unaliased:

- an earlier local may still reference it;
- the callee may store the parameter in an Array, map, module binding, closure,
  or dynamic interface before returning a fresh value; and
- a conditional candidate may become the carried value while its allocation
  site is reused on later iterations.

The retained regression test explicitly keeps an old `State`, advances the
current state twice, and requires the old fields to remain unchanged. A future
reuse implementation must also carry an interprocedural non-capture/effect
fact through every statically called function and method. The compiler does
not currently have such an ownership fact, and Go escape analysis occurs too
late to authorize changing Able object identity.

Two slots do not solve the general case: an alias may remain live for more
than two iterations, exactly as Fixed Width's `best` can. Pooling or rotating
identity-bearing structs is therefore rejected.

## Verification

- The new fixture builds with `--no-fallbacks` and its exact allocation run
  completes under the 55-second guard.
- Bytecode and tree-walker executions reproduce both expected output lines.
- Focused compiler caller-owned-result and retained-old-result tests pass.
- The generated-only ceiling binary reproduces the same output hash.
- No compiler, runtime, stdlib, application, reference, verifier, bytecode VM,
  spec, or WASM behavior changed.
- Temporary generated binaries, profiles, and escape logs are cleanup-eligible
  and removed after this record.

## Next recommendation

Run a fresh post-result-ABI compiled selection/profile gate across Binary
Trees, Sudoku Masks, K-Nucleotide, and TapeLang Alphabet.

Why: loop reuse is closed until a broader ownership/effect system exists. The
last full scorecard predates the retained wide carriers, native constants, and
caller-owned result ABI. Binary Trees remains the largest known non-wide
absolute miss and is already a verified no-regression guard, but its nominal
allocation wall historically differs from Sudoku's search buffers,
K-Nucleotide's map/string work, and TapeLang's allocation-light dispatch. A
fresh comparison is necessary before choosing another shared compiler wall.

What it entails: rebuild strict current binaries, run at least five independent
verifier-backed processes per application, and collect separate bounded CPU
and exact main-allocation profiles. Reconcile concrete descendants across the
four programs and advance only a language-level operation repeated in at least
three unlike applications. Do not add a `Node`, Sudoku, HashMap, TapeLang, or
other named nominal/application shortcut. Keep bytecode performance queued and
continue to defer WASM.
