# Post-nullable compiled regex reconciliation

## Decision

Reconcile `compiled-regex` as causally current and retain no production
change.

The primitive nullable value carrier materially reaches all six rows through
the canonical NFA acceptance paths. The generated acceptance functions
already use direct Go `value + valid` carriers, native `int32` state arrays,
and direct primitive returns. They perform no normal-path conversion to
`runtime.Value` and allocate no result objects.

Fresh merged profiles across three unlike regex applications attribute the
remaining acceptance cost to the native state scan and `accepts_state` call,
not to nullable match boilerplate or a compiled/interpreted boundary. The
larger repeated owners remain NFA closure, move, and thread upsert work whose
general alternatives have already failed broad gates.

## Strict boundary and execution gate

Every application was rebuilt from the retained compiler with
`--no-fallbacks`. Each exact binary passed its public Ruby verifier.

| Application | Packages | Interpreter dependency | Verified smoke |
| --- | ---: | --- | --- |
| Config Validation Extraction | 96 | absent | 1/1 |
| Log Routing Redaction | 96 | absent | 1/1 |
| Policy Record Dispatch | 96 | absent | 1/1 |
| Regex Set Audit | 96 | absent | 1/1 |
| Regex Stream Audit | 96 | absent | 1/1 |
| Regex Suffix Audit | 96 | absent | 1/1 |

Smoke durations are execution checks, not timing evidence. The authoritative
scorecard retains five verifier-backed Able and Go processes per row.

## Causal carrier reach

Generated support definitions were excluded from the reach decision. The
actual application-to-NFA paths are:

| Application | Material primitive-carrier path |
| --- | --- |
| Config Validation Extraction | `Regex.match/is_match -> find_token_match -> find_nfa_span -> regex_nfa_accepting_thread -> __able_nullable[int32]` |
| Log Routing Redaction | `RegexSet.matches -> regex_set_accepting_start -> __able_nullable[int32]`; replacement also reaches `regex_nfa_accepting_thread` |
| Policy Record Dispatch | schema `Regex.match -> find_nfa_span -> regex_nfa_accepting_thread -> __able_nullable[int32]` |
| Regex Set Audit | `RegexSet.matches -> regex_set_accepting_start -> __able_nullable[int32]` |
| Regex Stream Audit | scanner acceptance reaches `regex_nfa_accepting_thread`; scanner history also uses a direct `__able_nullable[char]` |
| Regex Suffix Audit | `Regex.match -> find_nfa_span -> regex_nfa_accepting_thread -> __able_nullable[int32]` |

Both shared acceptance functions are identical in shape across the generated
modules:

- `regex_nfa_accepting_thread` returns `__able_nullable[int32]`, scans
  `Array<i32>` through its native Go carrier, and constructs a present result
  with `__able_some`;
- `regex_set_accepting_start` also returns `__able_nullable[int32]` directly;
- normal execution contains zero `bridge.To*`, `bridge.As*`,
  `runtime.Value`, or nullable-to-runtime conversion calls;
- the only bridge call in either generated match expansion is the unreachable
  defensive non-exhaustive-match error path.

This establishes that the compiler/interpreter boundary targeted by the
nullable tranche is absent from the material regex path.

## Fresh three-family profile gate

Config Validation Extraction, Log Routing Redaction, and Regex Stream Audit
were profiled because they exercise batch validation, combined set/replace,
and streaming scanner APIs respectively.

Each application used the exact strict binary and public verifier. CPU
evidence merged 40 separately verified processes per application. Allocation
evidence used one additional exact verified process. Every process used CPU
12, `GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`, and a 90-second timeout.
All 120 CPU-profile processes and all three allocation processes verified.

| Application | CPU samples | `accepting_thread` flat | cumulative | Leading shared NFA owners |
| --- | ---: | ---: | ---: | --- |
| Config Validation Extraction | 370 ms | 20 ms (5.41%) | 30 ms (8.11%) | `add_closure` 10.81%, `upsert_thread` 8.11%, `move` 2.70% flat |
| Log Routing Redaction | 840 ms | 10 ms (1.19%) | 10 ms (1.19%) | `add_closure` 13.10%, `move` 13.10%, `upsert_thread` 11.90% flat |
| Regex Stream Audit | 870 ms | 20 ms (2.30%) | 50 ms (5.75%) | `add_closure` 16.09%, `upsert_thread` 12.64%, `move` 9.20% flat |

Line attribution is consistent:

- Config assigns 10 ms to the bounds test, 10 ms to the direct native Array
  read, and 10 ms cumulative to `regex_nfa_accepts_state`.
- Log assigns its only 10 ms to the native loop header.
- Stream assigns 10 ms to the loop header, 10 ms to the direct Array read,
  and 30 ms cumulative to `regex_nfa_accepts_state`.
- None of the three profiles samples carrier construction, `.valid`, `.value`,
  nullable match branches, defensive match failure, or a bridge conversion.

The exact allocation profiles contain 72,066, 100,810, and 15,292 sampled
objects respectively. Both acceptance functions have zero flat and zero
cumulative allocation objects. Allocation remains in distinct regex
operations such as thread-array creation, capture cloning, Unicode decoding,
and scanner feed work.

The merged CPU profile SHA-256 fingerprints are:

- Config Validation Extraction:
  `b3c3d02ed74a5c8956ce6199d68caa94e94b27cdcbc02239c6580ffe4706216f`
- Log Routing Redaction:
  `75856bf29e8e309bf86f1eec38f17ea58020fdceb95f16e3d05353c108ed74b2`
- Regex Stream Audit:
  `44a86ab8741b50945668b4c5352a15e27c7303bc1146f8cb9eaaf416f50ef773`

Their exact allocation profile fingerprints are recorded in the
machine-readable companion file.

## Residual-owner disposition

The material nullable path is already lowered as intended. Removing the
remaining acceptance scan would require maintaining a state-to-position or
accepting-state index. That exact general stdlib algorithm was previously
measured and rejected: normal Regex Stream bytecode became 21.3% slower.
Matcher thread arenas also regressed Suffix, Set, and Stream by 9.2%, 6.5%,
and 2.6%, and character specialization failed the generality gate.

The larger fresh owners—`add_closure`, `move`, and `upsert_thread`—are the
same canonical NFA owners covered by the retained closure-scratch, capture
template, and primitive-thread-carrier work. The nullable representation does
not invalidate the prior arena, index, character, raw-integer, Array-member,
or call/carrier A/B decisions.

There is therefore no new exact compiler, generated-runtime, or canonical
stdlib successor to test. Implementing another acceptance index would repeat
a closed regex-specific algorithm; adding a compiler rule for Regex or its
nominal structures would violate the nominal-lowering guardrail.

## Current row state

The current five-process scorecard means remain:

| Application | Able compiled | Go | Able / Go |
| --- | ---: | ---: | ---: |
| Config Validation Extraction | 0.0480 s | 0.0038 s | 12.6316x |
| Log Routing Redaction | 0.0420 s | 0.0041 s | 10.2439x |
| Policy Record Dispatch | 0.0960 s | 0.0053 s | 18.1132x |
| Regex Set Audit | 0.0680 s | 0.0054 s | 12.5926x |
| Regex Stream Audit | 0.0700 s | 0.0050 s | 14.0000x |
| Regex Suffix Audit | 0.0620 s | 0.0054 s | 11.4815x |

These misses remain important product gaps, but the evidence does not support
attributing them to primitive nullable boxing or interpreter fallback.

## Scope and cleanup

No compiler, generated runtime, runtime package, interpreter, bytecode VM,
canonical stdlib, benchmark, language, dependency, or WASM source changed.
No named-container, non-primitive nominal, regex-nominal, or
benchmark-specific rule was introduced.

`go test ./cmd/ablec` passed in 5.944 seconds. The machine-readable record is
`2026-07-30-post-nullable-compiled-regex-reconciliation.json`.

After retaining this evidence, the exact 1,092 MiB disk-backed generated
module, binary, profile, and Go-cache workspace was removed. No matching
tranche artifact remains in `/var/tmp` or `/tmp`.

## Next

Reconcile `compiled-float-numeric` against the retained primitive nullable
carrier.

Why: its four rows—Distance Field, Mandelbrot, N-Body, and RMS Norm—are the
smallest remaining invalidated family whose material work uses a primitive
carrier class directly covered by the compiler change. The carrier supports
nullable `f64`, and prior float-owner records include raw-float lowering
questions that now need causal review.

What it entails: strictly rebuild all four applications, verify their graphs
remain interpreter-free, and trace whether any material float result,
accumulator, or control path actually uses the new nullable carrier. Reuse the
current Distance Field architecture profile and create fresh profiles only if
a changed path reaches one exact residual that can plausibly repeat across at
least three unlike rows.

Why it matters: native float arithmetic should be one of the clearest places
for compiled Able to approach Go. Proving direct float carriers—or locating
one remaining shared boxing boundary—keeps the work aligned with native Go
lowering instead of spending another tranche on already-closed regex
algorithms.
