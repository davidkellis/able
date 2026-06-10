# Regex carrier targeted external scorecard refresh

Date: 2026-07-17

## Decision

Retain the primitive regex NFA carrier and keep the targeted scorecard as
fresh product evidence. Do not promote it to `external-scoreboard-current`:
the targeted report contains six application/mode rows, while full promotion
requires all 68 status rows and the exact five-run evidence for every one of
the 61 reviewed rows.

No compiler, bytecode VM, runtime, stdlib, benchmark, reference, verifier, or
application changed in this tranche. The measured stdlib is the retained
70-file canonical tree with SHA-256
`f7a470aae4fba342e5bbc3fce53ee26fa6f96df71dde18e057e044520624dafc`.

## Contract and evidence

The refresh reran the exact prior scorecard group:

- affected applications: Regex Set Audit and Regex Stream Audit;
- unchanged control: Array Slice Window, which does not import regex;
- modes: compiled and bytecode;
- five independent verifier-backed Able processes per row;
- five fresh Go processes for each compiled reference;
- five fresh Python and Ruby processes for each bytecode reference;
- catalog-resolved one-CPU execution, `GOMEMLIMIT=1GiB`, `GOGC=50`, and a
  55-second per-process cap.

All 15 Go, 30 Python/Ruby, 15 compiled Able, and 15 bytecode Able processes
completed and verified. Three-report variance checks for compiled and bytecode
accepted cohort A, cohort B, and the refresh with exactly five successful Able
and reference samples in every component.

## Refreshed scorecard

| Application | Mode | Able | Reference | Ratio | Target |
| --- | --- | ---: | ---: | ---: | --- |
| Regex Set Audit | compiled | 0.1260 s | Go 0.0054 s | 23.33x | miss |
| Regex Stream Audit | compiled | 0.1380 s | Go 0.0047 s | 29.36x | miss |
| Array Slice Window | compiled | 0.0900 s | Go 0.0053 s | 16.98x | miss |
| Regex Set Audit | bytecode | 4.9600 s | Python 0.0236 s / Ruby 0.0536 s | 210.17x / 92.54x | miss |
| Regex Stream Audit | bytecode | 4.1780 s | Python 0.0198 s / Ruby 0.0496 s | 211.01x / 84.23x | miss |
| Array Slice Window | bytecode | 0.7240 s | Python 0.0298 s / Ruby 0.0606 s | 24.30x / 11.95x | miss |

The bytecode application means improve relative to both previous cohorts:
Regex Set is 13.8% faster than cohort A and 4.4% faster than cohort B; Regex
Stream is 8.9% and 7.6% faster respectively. The unchanged bytecode control is
4.0% slower than cohort A and 2.8% slower than cohort B, consistent with
workstation/process variance rather than a broad unrelated speedup.

Fresh reference processes also vary, so product ratios are not expected to
move in exact proportion to Able time. The refreshed Regex Set ratios sit
between or below the prior cohorts; Regex Stream's Ruby ratio improves beyond
both prior cohorts while its Python ratio lies between them. The retained A/B
carrier gate remains the attribution evidence; this refresh measures current
user-facing distance to the external targets.

## Promotion boundary

The targeted aggregate renders and contains six verified, rankable rows. The
full selection/status check rejects it as a current-scoreboard replacement
because required rows such as `await_channel_mux`, Base64, and Binary Trees are
absent. `external-scoreboard-current` therefore remains unchanged and its
no-input synchronization check still passes. A future promotion requires a
complete grouped refresh, not row splicing or synthetic carry-forward.

## Artifacts

- targeted scoreboard:
  `2026-07-17-regex-carrier-targeted-scorecard.{json,md}`
- Able comparison reports:
  `2026-07-17-regex-carrier-scorecard-{compiled,bytecode}.{json,md}`
- fresh references:
  `2026-07-17-regex-carrier-scorecard-{go-reference,interpreter-reference}.{json,md}`
- three-cohort evidence:
  `2026-07-17-regex-carrier-scorecard-{compiled,bytecode}-variance.{json,md}`
- canonical source state:
  `2026-07-17-regex-carrier-scorecard-stdlib-source-state.json`

## Next recommendation

Return to cross-application bytecode profiles using Reverse Complement,
Rational Series, and Word Frequency, with Array Slice Window as a guard. These
are unlike stable misses covering byte/string mutation, numeric/nominal work,
and map/iterator processing; Array Slice supplies a fourth array-heavy path.
They therefore provide a better generality test than further regex-only work.

This entails fresh five-process warmed allocation measurements plus one
bounded CPU profile per application under the same one-CPU/OOM guardrails.
Normalize the concrete descendants below call dispatch and admit a candidate
only if the same VM/runtime operation is material in at least three unlike
applications. The next tranche should not retry the rejected raw-integer,
return/frame, per-container, or regex state-index candidates.
