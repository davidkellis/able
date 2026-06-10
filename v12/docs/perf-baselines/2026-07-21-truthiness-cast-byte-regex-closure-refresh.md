# Truthiness/cast byte-regex closure refresh

Date: 2026-07-21

## Decision

The `compiled-byte-output` and `bytecode-regex` closures are current against
the post-truthiness/cast shared interpreter semantics. Keep no compiler, VM,
runtime, canonical-stdlib, benchmark, verifier, reference, language, or WASM
change from this tranche.

Fresh repeated timing confirms important product gaps, but exact reach rejects
the corrected semantic boundary as their cause. None of the four compiled
programs enters generated truthiness or explicit-cast helpers. The six
bytecode regex programs make 208,060-1,565,212 ordinary truthiness checks per
main, but none reaches the changed non-primitive Error fallback or explicit
cast boundary. No concrete leaf passes the profile admission rule.

## Frozen contract

- The v12 spec SHA-256 remains
  `4f0405b86c122993723e8617abd6f825d9a8ff858d4c72acaf4e33469452f080`.
- The canonical external stdlib source-tree SHA-256 remains
  `43ff2e68e59c8be7fb1024c86a1f61a0eea84596279b4f0e146511d66c5308d8`.
- Executables were built before timing. Every timing process passed its public
  verifier and used its catalog directory, arguments, serial CPU budget, and
  a 55-second cap.
- Arithmetic means retain every successful sample. Base64, FASTA Generation,
  and Pi Digits received matched second compiled/Go cohorts because the first
  cohort was volatile or close to the 95%-of-Go target. No sample was removed.
- Able executions use the catalog's bounded memory and serial CPU policy.

## Repeated timing

| Application | Mode | Able samples | Able mean | Limiting reference | Reference samples | Reference mean | Ratio |
| --- | --- | ---: | ---: | --- | ---: | ---: | ---: |
| Base64 | compiled | 10 | 2.521 s | Go | 10 | 2.577516 s | 0.978x |
| FASTA Generation | compiled | 10 | 0.104 s | Go | 10 | 0.012775 s | 8.141x |
| Pi Digits | compiled | 10 | 1.264 s | Go | 10 | 1.194157 s | 1.058x |
| Reverse Complement | compiled | 5 | 0.176 s | Go | 5 | 0.018503 s | 9.512x |
| Config Validation Extraction | bytecode | 5 | 1.282 s | Python | 5 | 0.023363 s | 54.873x |
| Log Routing Redaction | bytecode | 5 | 2.892 s | Python | 5 | 0.017588 s | 164.433x |
| Policy Record Dispatch | bytecode | 5 | 7.280 s | Python | 5 | 0.024965 s | 291.614x |
| Regex Set Audit | bytecode | 5 | 4.164 s | Python | 5 | 0.019942 s | 208.804x |
| Regex Stream Audit | bytecode | 5 | 3.750 s | Python | 5 | 0.018835 s | 199.093x |
| Regex Suffix Audit | bytecode | 5 | 3.256 s | Python | 5 | 0.018069 s | 180.197x |

All 160 timing processes represented by these decisions verified with zero
failures and zero timeouts: 70 compiled/Go processes and 90
bytecode/Python/Ruby processes. The pooled compiled CVs are 11.95% Able and
6.46% Go for Base64, 26.51% Able and 4.59% Go for FASTA, and 6.16% Able and
6.46% Go for Pi Digits. Reverse Complement's five-run Able CV is 17.79%; the
largest bytecode CV is Regex Stream Audit at 13.88%. Every sample remains in
its arithmetic mean.

Base64 now clears the compiled target. Pi Digits is close but remains just
outside it: 1.058x Go means 94.48% of Go throughput, versus the 95% goal.
FASTA and Reverse Complement remain structural gaps. Every bytecode regex row
is far outside the 1.053x ceiling corresponding to 95% of the faster Python or
Ruby reference. This focused refresh does not rewrite the full scoreboard.

## Exact reach

Temporary debug counters were placed immediately at the changed semantic
boundaries, used only in untimed processes, and then removed.

| Application | Mode | Census processes | Truthy checks/process | Changed Error fallback | Explicit casts/process | Cast failures |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| Base64 | compiled | 1 | 0 | 0 | 0 | 0 |
| FASTA Generation | compiled | 1 | 0 | 0 | 0 | 0 |
| Pi Digits | compiled | 1 | 0 | 0 | 0 | 0 |
| Reverse Complement | compiled | 1 | 0 | 0 | 0 | 0 |
| Config Validation Extraction | bytecode | 2 | 208,060 | 0 | 0 | 0 |
| Log Routing Redaction | bytecode | 2 | 733,267 | 0 | 0 | 0 |
| Policy Record Dispatch | bytecode | 2 | 1,565,212 | 0 | 0 | 0 |
| Regex Set Audit | bytecode | 2 | 939,980 | 0 | 0 | 0 |
| Regex Stream Audit | bytecode | 2 | 560,897 | 0 | 0 | 0 |
| Regex Suffix Audit | bytecode | 2 | 634,199 | 0 | 0 | 0 |

Every retained census process completed under its guard and passed the public
verifier; both bytecode counts reproduce exactly. One preliminary full-opcode
statistics attempt was killed before producing output, so the exact counters
were rerun without unrelated full statistics, reducing memory pressure while
preserving the same semantic census. It is not counted above.

All bytecode checks return through primitive truthiness cases before the
corrected Error matcher. Because no changed path is reached in either closure,
the profile admission gate fails and no CPU profile or candidate is warranted.
The concise census is retained in
`2026-07-21-truthiness-cast-byte-regex-closure-reach.json`; raw compiled
telemetry is retained in
`2026-07-21-truthiness-cast-byte-regex-closure-compiled-reach.json`. No
diagnostic counter, binary, raw profile, or generated package remains in
production code.

## Exact timing artifacts

- Initial compiled Able and Go:
  `2026-07-21-truthiness-cast-byte-regex-closures-compiled.json`
  (`d14ab71e469853f622d9b16b257f051a70f590966b9a2518350d5d86dee94ed7`),
  `2026-07-21-truthiness-cast-byte-regex-closures-go-reference.json`
  (`4b74e23054a37f4c9ba49e374d65f1d4fb9c291f99c88090430eae4954f7480e`).
- Matched second compiled and Go cohorts:
  `2026-07-21-truthiness-cast-byte-regex-closures-compiled-c2.json`
  (`0c1ecbb8347055ce7145c67d148635d7dff684175af141640d1f78f2bc7ce274`),
  `2026-07-21-truthiness-cast-byte-regex-closures-compiled-c2-go-reference.json`
  (`e5702504aad727503c9749a144b1f985995c6e4f786d1f2373a40ecf3df491a3`).
- Bytecode Able and interpreter references:
  `2026-07-21-truthiness-cast-byte-regex-closures-bytecode.json`
  (`6f6151f965bda9d4bcb80403de47d97c81c4819efd85c70d08b3e83ed1578f28`),
  `2026-07-21-truthiness-cast-byte-regex-closures-interpreter-reference.json`
  (`0abbe8c4956da17b46cc453d6015457646d60a1bdb20bd90f1dbc4b6492310ea`).
- Compiled reach telemetry:
  `2026-07-21-truthiness-cast-byte-regex-closure-compiled-reach.json`
  (`6fc71c68cf71a55f7df98c76c4186c905aecc25394b451ec9659d6b976215eac`).

## Next recommendation

Refresh `compiled-regex` and `bytecode-concurrency` next.

Why: compiled regex is the compiler counterpart to the VM regex family just
measured, so it can distinguish generated-runtime/stdlib regex costs from VM
dispatch costs. Bytecode concurrency is the nearest remaining invalidated VM
closure and exercises Future, cancellation, result, and Error-bearing control
paths unlike the byte/output and regex rows already closed. Together they add
new feature diversity without mixing mode-specific ownership.

What it entails: reuse the frozen sources and current toolchains; collect five
verified processes per ordinary lane and matched additional cohorts for
volatile workstation rows; run exact changed-path reach before profiling; and
profile only a materially reached concrete leaf. Advance only those two
closures. Build a candidate only if one generic mechanism is material in at
least three unlike applications and preserves current target guards. This is
the right next step because it completes the closest compiler counterpart and
tests the corrected semantics in a genuinely different VM subsystem. No WASM
work is involved.
