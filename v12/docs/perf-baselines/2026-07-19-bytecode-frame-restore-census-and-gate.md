# Bytecode full-frame restore census and gate

Date: 2026-07-19

Decision: retain no runtime change. The empty restore-state shortcut was
generic and correct, but it produced a material, consistent improvement in
only one of four unlike applications and therefore failed the broad benchmark
bar.

## Census

An opt-in temporary counter measured the live components restored by
`popCallFrameFields` in Fixed Width 128, Distance Field, Option/Result Config,
and Word Frequency. Each application ran once in its own verifier-backed
process using its catalog input and working directory, the canonical external
stdlib, `GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`, one allowed workstation
CPU, and a 55-second cap. `ABLE_BYTECODE_STATS_MAIN_ONLY=1` excluded program
loading and stdlib setup from the counts without enabling the broader bytecode
statistics paths.

All 7,967,551 observed pops were ordinary full frames. No observed frame was a
self-fast or minimal-self-fast pop.

| Application | Full pops | Caller and callee restore state empty | Caller state present | Callee state present |
| --- | ---: | ---: | ---: | ---: |
| Distance Field | 4,000,000 | 4,000,000 (100%) | 0 | 0 |
| Fixed Width 128 | 3,000,008 | 3,000,008 (100%) | 0 | 0 |
| Option/Result Config | 293,784 | 229,680 (78.2%) | 6,264 | 64,080 |
| Word Frequency | 673,759 | 635,864 (94.4%) | 37,895 | 0 |

The restore-state category covers caller/callee i32 register frames, implicit
slot state, i32 value-slot sidecars, and float value-slot sidecars. Across all
four applications, 7,865,552 pops (98.7%) had none of those components on
either side.

The nonempty cases were also narrow. Option/Result used caller i32 registers
6,240 times, callee i32 registers 57,648 times, caller implicit slots 24 times,
and callee implicit slots 6,432 times. Word Frequency used caller implicit
slots 37,895 times. No application used an i32 value-slot sidecar, float
sidecar, transient-scope release, iterator/loop-stack unwind, or non-nil array
ownership parent in an observed full pop.

This did not mean the frames themselves were semantically empty. Every pop
changed programs and restored an active lookup program. Slot frames changed on
all Fixed Width, Distance, and Word Frequency pops and on 269,208 Option/Result
pops. Active call-name lookup state occurred 3,000,003, 2,000,000, 593,871,
and 236,136 times respectively. Option/Result additionally restored active
scope lookup state 205,320 times, carried return generics 122,880 times, and
carried a return-coercion function 24,600 times; Word Frequency carried return
generics 63,843 times.

The temporary census code was removed after measurement. Its final diagnostic
binary SHA-256 was
`45b13cab4911c862826d39dfbc044ad05156a9ef8a734ef961587028da831c88`.

## Candidate

The admitted candidate added one generic check to the full-frame branch. When
both caller and callee lacked register, implicit-slot, and value-slot sidecar
state, it skipped the existing detach, field-clear, release, and restore
helpers for those components. Payload-bearing frames continued through the
unchanged restoration path.

The candidate did not change the ten-result `popCallFrameFields` ABI, repeat
the previously rejected caller-owned result struct, alter frame eligibility,
recognize an application or nominal type, or change calls, errors, scheduling,
lookups, coercion, ownership, or language semantics.

The preserved baseline binary SHA-256 was
`57fe72757410edc3bc8974e0a7d5ee0be12a1bca0764f27183f887fdd8990eb1`.
The candidate binary SHA-256 was
`cdef682bfdb876bcac61542d22506d32a490f456bee2c479c148cb33e3e74ba4`.

## Repeated verifier-backed A/B

Processes alternated baseline-first and candidate-first order by pair under
the same CPU, memory, GC, stdlib, input, working-directory, and timeout
contract as the census. Fixed Width completed five pairs. The other three
applications were extended from five to ten pairs because their initial means
were smaller than their workstation spread. All 70 executions passed their
public verifiers and every sample is included.

| Application | Pairs | Baseline mean | Candidate mean | Change | Candidate wins |
| --- | ---: | ---: | ---: | ---: | ---: |
| Fixed Width 128 | 5 | 8.224 s | 7.912 s | -3.79% | 5/5 |
| Distance Field | 10 | 5.808 s | 5.765 s | -0.74% | 5/10 |
| Option/Result Config | 10 | 0.867 s | 0.842 s | -2.88% | 6/10 |
| Word Frequency | 10 | 1.587 s | 1.572 s | -0.95% | 5/10 |

Option/Result's headline mean is driven by one 1.07-second baseline sample.
Removing that single slow sample yields a 0.844-second baseline mean versus
0.842 seconds candidate, which is neutral. Distance and Word Frequency are
also neutral by magnitude and pair wins. Fixed Width is a credible local win,
but one application is insufficient for retention under the cross-program
admission rule.

Raw samples are preserved in
`2026-07-19-bytecode-frame-restore-empty-state-ab-samples.tsv`.

## Revert and correctness

The shortcut and its predicate test were removed. The rebuilt post-revert
binary is byte-for-byte identical to the preserved baseline at SHA-256
`57fe72757410edc3bc8974e0a7d5ee0be12a1bca0764f27183f887fdd8990eb1`.
No production code, compiler, stdlib, tree-walker, language, fixture, or WASM
change remains from this tranche.

Verification after the revert:

- focused frame, inline-call, return, slot/value-frame, ownership, and program-
  switch tests passed in 0.635 seconds;
- `go test ./pkg/interpreter -run 'TestBytecode' -count=1 -timeout 60s`
  passed in 25.752 seconds;
- split tree-walker/bytecode parity groups 02-04, 05-08, 09-11, 12-13, and 14
  passed in 11.227, 28.074, 7.437, 12.483, and 21.196 seconds;
- every command remained below the project's one-minute test ceiling.

## Next recommendation

Close the empty restore-state idea and re-attribute the remaining
`finishInlineReturn` cost after the retained active-lookup-state improvement.
Collect fresh bounded CPU profiles for Fixed Width, Distance Field,
Option/Result, and Word Frequency, then separate return-type/coercion checks,
ownership completion, frame pop, program switching, control restoration, slot
release, and returned-value append/materialization by exact child and caller.

Why: the census proves that empty register/sidecar state is broad, but the A/B
gate proves skipping its helpers is not a broad wall. The previously retained
active-lookup bundle already changed the cost distribution, while older raw-
integer and return-guard rearrangements also failed. Fresh child attribution
is needed before another return-path edit. Admit a candidate only if the same
remaining operation is material in at least three unlike programs; otherwise
leave the return boundary and select a different shared VM leaf. Continue to
defer WASM.
