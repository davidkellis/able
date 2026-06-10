# Log routing and redaction profile gate (2026-07-20)

## Decision

Keep the portable Log Routing and Redaction application, its Go/Python/Ruby
references and verifier, and the canonical `Regex.replace` result-boundary
correctness fix. Keep no compiler, generated-runtime, or bytecode-VM
performance candidate.

The application supplies one genuinely unlike use of the public regex engine
beyond the three related API audits. Exact NFA closure/move/thread work repeats
in compiled mode, and canonical Array-member traffic repeats in bytecode mode,
but the evidence still represents only two workload families. The operation
depth policy requires three unlike applications before a new candidate may
advance.

## Application contract

The checked-in 20-record corpus is processed four times. Five mutually
exclusive `RegexSet` patterns classify authentication errors, payment errors,
warnings, informational records, and debug records. Three public `Regex`
replacements redact email addresses, IPv4 addresses, and authentication
tokens. The deterministic result is:

`80:12,12,16,24,16:236836`

The `--profile` scale processes one round and emits
`20:3,3,4,6,4:59209`. It exists only to bound diagnostic profiles; scorecard
and verifier runs use the four-round default. Able, Go, Python, and Ruby
implement the same routing, replacement order, checksum, input, and output.

## Canonical stdlib correction

This application exposed a general public result-boundary bug in
`able.text.regex`. `Regex.replace` promised a builtin `String`, but the literal
and callback paths returned the nominal `able.text.string.String` builder
wrapper. Calling a builtin method such as `len_bytes()` on that result could
therefore collide with the wrapper's integer field and fail as a call on a
non-function.

Both replacement paths now use the existing `__able_String_to_builtin`
boundary on the completed builder buffer. The new canonical stdlib test proves
that literal and callback replacement results support builtin `String`
methods in both Go interpreters. This is a public type-contract fix, not a
regex benchmark fast path.

## Repeated timing evidence

All targeted Able executions and references were verifier-backed. Because the
compiled process is short and volatile on this workstation, three independent
five-process cohorts were retained instead of choosing the best batch:

| Cohort | Compiled mean | Bytecode mean |
| --- | ---: | ---: |
| A | 0.112 s | 2.818 s |
| B | 0.150 s | 3.016 s |
| C | 0.116 s | 2.948 s |
| Pooled 15 runs | 0.126 s | 2.927 s |

The targeted five-run references measured 0.0049 s for Go, 0.0187 s for
Python, and 0.0431 s for Ruby. Against those means, the pooled Able ratios are
25.71x Go in compiled mode, 156.54x Python in bytecode mode, and 67.92x Ruby
in bytecode mode. These are large product gaps and make the application useful
frontier evidence; they do not by themselves identify a general candidate.

The independent promoted full scorecard retained five fresh verified samples
for every selected row. Its application means are 0.100 s compiled versus
0.0043 s Go (23.26x), and 2.958 s bytecode versus 0.0181 s Python (163.43x)
and 0.0422 s Ruby (70.09x). The complete promoted ledger has 80 full-status
rows and 73 selected rows: all 40 compiled applications and 33 bounded
bytecode applications. Four selected compiled rows meet the Go target; two
selected bytecode rows meet both interpreter targets.

## Bounded profile attribution

One current generated binary with SHA-256
`ac5f1f78183871d7f364d0b31a8a7e278141dd3cc2aada5e2557a31ff43171a1`
served 40 verified one-process compiled CPU launches. Their merged 1.12 seconds
of samples report:

| Exact owner | Flat | Cumulative |
| --- | ---: | ---: |
| `regex_nfa_add_closure` | 13.39% | 18.75% |
| `regex_nfa_move` | 7.14% | 17.86% |
| `regex_nfa_upsert_thread` | 2.68% | 5.36% |

One exact main-phase allocation launch recorded 8,580,976 allocated bytes and
220,432 allocations. Allocation and GC remain material descendants of the same
canonical NFA implementation already profiled in the Suffix, Set, and Stream
audits.

One preserved bytecode test binary with SHA-256
`368ecc79119e4bfe44640a166fd38f6731a4151c03674359f5572d21fdf076e8`
loaded and typechecked once, warmed once, and measured one default application
call. It reported 2,423,206,361 ns/op, 20,422,216 B/op, 176,114 allocs/op, and
2.41 seconds of CPU samples. `execCallMemberArraySlot` was 26.14% cumulative;
canonical validated Array-slot lookup was 9.13%; raw-integer extraction was
2.90% flat. The existing three-API gate reports the same Array-member parent at
21.02%-24.93% cumulative.

## Breadth and candidate reconciliation

The profile recurrence is real, but the breadth is not yet sufficient:

- Regex Suffix, Set, and Stream deliberately audit three API shapes of one NFA
  algorithm and count as one workload family.
- Log Routing and Redaction is an ordinary classification/privacy application
  and adds a second family.
- The catalog therefore records `coverage_kind: mixed` and
  `breadth_status: insufficient` for `regex_nfa_matching`.

No candidate was built. Closure scratch, immutable initial captures,
primitive active-thread carriers, outgoing-transition indexing, and generated
primitive kernel calls are already retained. Thread arenas, state-position
indexes, character specialization, raw-integer carriers, generic Array-member
variants, and related call/return changes have already failed broad gates.
This new profile does not invalidate any of those decisions. A named Regex
lowering or NFA-shaped opcode remains prohibited.

## Verification

- canonical and sibling Able sources are byte-identical;
- Go, Python, Ruby, bytecode default, and bounded tree-walker outputs pass the
  external verifier;
- all 15 targeted compiled and all 15 targeted bytecode executions verified;
- the promoted scorecard evidence check passed with 73 selected rows, 80
  full-status rows, and exactly five successful Able/reference samples per
  selected row;
- the new canonical stdlib replacement test passes tree-walker and bytecode;
- the static bytecode audit now lowers 119 programs, 455 functions, and 21,608
  instructions;
- the rebuilt 73-row frontier reports six selected target meets, 67 misses,
  129.930 seconds above the aggregate target budget, and zero actionable
  groups; and
- no WASM work was performed.

## Next recommendation

Add one portable configuration-validation and extraction application using the
public regex APIs.

Why: regex/NFA matching is now the only operation-depth entry without three
unlike portable workload families. A deployment/configuration validator would
exercise field validation, version and endpoint recognition, and capture-based
extraction without repeating either an engine audit or log routing/redaction.
That would make a future profile decision honest: either the same exact NFA
owner recurs across three unlike applications, or regex performance work closes
without inventing a benchmark-shaped optimization.

What it entails: add one checked-in configuration corpus plus equivalent Able,
Go, Python, and Ruby implementations; verify identical extracted summaries and
error counts; register the application in coverage, selection, operation-depth,
scorecard, and static-audit contracts; collect repeated verifier-backed means;
then profile it under the same single-process memory and timeout guardrails.
Advance at most one generic compiler, VM, or canonical NFA operation only if it
is material in the audit family, log application, and configuration application
and passes unrelated target guards.
