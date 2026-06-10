# Bytecode register-native member-access gate

Date: 2026-07-21

## Decision

Reject and fully remove the register-native `MemberAccess` vertical slice.

The slice cleared correctness and dynamic-reach admission in all six selected
applications, but five order-balanced verifier-backed processes per side were
neutral or slower everywhere. Four applications regressed by 5.3%-15.0%, one
regressed 1.2%, and one was exactly neutral. The broad wall-time bar therefore
fails despite excellent semantic reach.

No compiler, VM, runtime, stdlib, workload, fixture, reference implementation,
scorecard, language, or WASM performance change is retained.

## Candidate

The recovered whole-function operand-register engine retained its previously
validated `CallName` continuation ABI, raw-i32 lane, slot operands, arithmetic,
stores, branches, loops, and returns. This tranche added one generic semantic
family:

- `MemberAccess` consumes one explicit register operand and produces one
  register result;
- safe `?.` returns `nil` directly when its register receiver is nil;
- ordinary access delegates to a value-returning form of the existing
  authoritative member resolver, preserving planned/direct struct fields,
  method preference, member-method caches, package/dynamic fallback,
  standard-error wrapping, and source diagnostics;
- unsupported whole functions still fall back before effects;
- the ordinary boxed operand stack remains empty throughout admitted
  member-access functions.

The candidate did not add a named struct, stdlib container, application, or
benchmark special case. The operand-register implementation remained opt-in
during measurement.

## Correctness and reach

Focused tests covered ordinary field access without operand-stack mutation,
safe nil short-circuiting, unsupported-function fallback, native and inline
`CallName`, inline suspend/resume, i32 register flow, loop CFG, and stale slot
references. Existing member/cache, call-name, inline-return, slotless-return,
minimal-return, and lookup-restoration guards passed.

The dynamic reach process enabled bytecode statistics only for census, used
the declared executor and canonical external stdlib, and ran with
`GOMAXPROCS=1`, `GOGC=50`, a 1-GiB memory limit, and a 55-second cap. All six
public verifiers passed.

| Application | Register entries | Semantic IR instructions | Suspend / resume |
| --- | ---: | ---: | ---: |
| Future Pipeline | 32,783 | 98,349 | 0 / 0 |
| Concurrent Text Index | 66,926 | 200,777 | 0 / 0 |
| Concurrent Event Routing | 22,546 | 67,637 | 0 / 0 |
| Dependency Wave Validation | 29,688 | 89,064 | 0 / 0 |
| Validated Job Pipeline | 65,549 | 196,647 | 0 / 0 |
| Word Frequency | 35,959 | 107,876 | 0 / 0 |

This clears the required three-unlike-application reach gate by a wide margin.
The reached functions are short: the instruction/entry ratios are almost
exactly three in every application.

## Preserved-binary A/B

The candidate binary was frozen after focused tests. The source was then fully
restored and used to build a separate baseline binary, so the baseline did not
contain either the register engine or the value-returning member-helper
refactor. Each application received five order-balanced baseline/candidate
pairs. The only candidate environment difference was enabling the register
engine; bytecode statistics were disabled. Every timed output passed its
public verifier.

Positive deltas are regressions.

| Application | Baseline mean | Candidate mean | Delta | Candidate wins | Baseline CV | Candidate CV |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Future Pipeline | 0.414 s | 0.476 s | +14.98% | 0/5 | 8.3% | 6.4% |
| Concurrent Text Index | 0.610 s | 0.658 s | +7.87% | 0/5 | 4.2% | 2.3% |
| Concurrent Event Routing | 3.114 s | 3.278 s | +5.27% | 1/5 | 8.8% | 7.5% |
| Dependency Wave Validation | 0.486 s | 0.492 s | +1.23% | 2/5 | 5.9% | 7.1% |
| Validated Job Pipeline | 0.874 s | 0.874 s | +0.00% | 3/5 | 8.3% | 3.9% |
| Word Frequency | 1.428 s | 1.524 s | +6.72% | 1/5 | 3.9% | 4.9% |

All CVs remain below the predeclared 15% extension threshold. The four clear
regressions also have only 0/5 or 1/5 candidate wins, so adding more workstation
samples would not turn this into a broad win.

## Interpretation

Semantic closure and wall-time value are different gates. `MemberAccess`
proved that a one-family extension can admit broadly useful functions, and the
state/CFG design is correct. However, each admitted invocation replaces only a
roughly three-instruction stack program. The current executor allocates and
zeros a register-cell slice sized to the bytecode program, enters a second
dispatch loop, and performs return reconciliation on every invocation. The
candidate's uniform neutral-to-negative result is consistent with that fixed
per-entry cost overwhelming the few transport dispatches removed. This is an
inference from the execution shape and A/B result; it still needs direct
profile/allocation attribution before an infrastructure candidate is built.

The result does not justify moving to `StructLiteralNamedFast` or `CallMember`
immediately. Both would use the same whole-function entry machinery;
`CallMember` also adds suspension complexity, while the struct-literal census
is concentrated in one application.

## Next recommendation

Run a bounded register-frame setup and allocation attribution gate across
Future Pipeline, Concurrent Text Index, and Word Frequency before attempting
another semantic family.

Why: these are unlike workloads with 32,783-66,926 reached register entries,
and all three regress despite successful translation. We need to distinguish
per-entry register-cell allocation/zeroing, executor entry/return dispatch,
and member-resolution cost rather than guessing which infrastructure change
would repay short functions.

What it entails: temporarily recover the same frozen candidate; collect clean
CPU and allocation profiles without bytecode event statistics; normalize
allocation counts and bytes by register entry; compare the candidate against
the restored baseline; and require the same exact setup owner to be material
in all three applications. Only then consider a general reusable/pool-backed
register-frame or caller-owned register-storage design, with cold fallback and
identity/continuation clearing guards. Re-run this six-application A/B before
adding another semantic family. Continue to defer WASM.
