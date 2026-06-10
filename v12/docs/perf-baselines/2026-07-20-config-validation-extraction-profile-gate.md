# Configuration validation and extraction profile gate (2026-07-20)

## Decision

Keep the portable Configuration Validation and Extraction application, its
Go/Python/Ruby references, verifier, corpus, and benchmark-contract
registrations. Keep no compiler, bytecode-VM, generated-runtime, or canonical
stdlib performance candidate.

The application supplies the third unlike portable workload family required
by the regex operation-depth gate. Current profiles reproduce the same exact
canonical NFA work in compiled mode and the same generic Array-slot member
traffic in bytecode mode, but expose no new material child and invalidate none
of the broad candidate rejections already recorded by the Regex API and Log
Routing/Redaction gates. A named Regex lowering, configuration-shaped fast
path, or NFA-specific bytecode opcode remains prohibited.

## Portable application contract

The checked-in corpus contains 24 deployment records. Twelve are valid; the
other twelve are partitioned into three schema failures, two version
failures, two port failures, two replica failures, and three region failures.
One anchored schema expression captures service, semantic-version components,
host, port, replica count, and region. Separate public Regex values validate
the extracted fields. The default four-round deterministic result is:

`96:48:12,8,8,8,12:253568`

The one-round `--profile` scale emits
`24:12:3,2,2,2,3:63392`; it exists only for bounded diagnostics. Scorecard and
verifier runs use the default scale. Able, Go, Python, and Ruby share the same
input, validation rules, extraction order, error categories, checksum, and
output. The canonical and external Able sources are byte-identical.

## Repeated timing evidence

All targeted executions were verifier-backed. Three independent five-process
Able cohorts were retained because the bytecode cohort means moved enough to
justify an additional workstation sample:

| Cohort | Compiled mean | Bytecode mean |
| --- | ---: | ---: |
| A | 0.124 s | 1.630 s |
| B | 0.112 s | 1.380 s |
| C | 0.100 s | 1.296 s |
| Pooled 15 runs | 0.112 s | 1.435 s |

The targeted five-run references measured 0.0048 seconds for Go, 0.0198
seconds for Python, and 0.0410 seconds for Ruby. The pooled Able ratios are
23.33x Go, 72.49x Python, and 35.01x Ruby. These are material product misses,
not a synthetic target-meeting benchmark.

The independent promoted full scorecard retained five fresh verified samples
for every selected row. Its application means are 0.118 seconds compiled
versus 0.0043 seconds Go (27.44x), and 1.702 seconds bytecode versus 0.0196
seconds Python (86.84x) and 0.0445 seconds Ruby (38.25x). Across the targeted
and promoted Able cohorts, all 20 compiled and all 20 bytecode outputs
verified; their pooled Able process means are 0.1135 and 1.5020 seconds.

Base64 compiled landed only 0.4% beyond the target cutoff in the promoted
cohort, so it received the required volatility follow-up. A separate ten-run
cohort measured 2.709 seconds Able versus 2.891 seconds Go (0.94x), while the
promoted cohort measured 2.602 versus 2.4625 seconds (1.06x). Pooling all 15
samples per implementation gives 2.673 seconds Able versus 2.748 seconds Go
(0.97x). The promoted five-run ledger remains unchanged and reports Base64 as
a miss; the pooled evidence classifies it as a volatile boundary guard, not a
stable regression or a newly established win.

## Bounded compiled attribution

One current generated binary with SHA-256
`1ead0074a22d9ecf979f69ed7e06542abef69ba1b7fbac7b1d02f502a0ff6334`
served 40 separately verified one-process CPU launches. The merged 0.40
seconds of CPU samples report:

| Exact canonical owner | Flat | Cumulative |
| --- | ---: | ---: |
| `regex_nfa_add_closure` | 2.50% | 12.50% |
| `regex_nfa_move` | 0.00% | 15.00% |
| `regex_nfa_upsert_thread` | 2.50% | 2.50% |
| `regex_nfa_threads_new` | 2.50% | 5.00% |

One allocation-only phase launch recorded 4,971,520 main-phase allocated
bytes and 107,406 allocations. The exact profile again attributes material
application allocations to NFA thread arrays, capture clones, codepoint
arrays, builtin/nominal String conversion, and primitive bridge conversion.
This is the same ownership family already covered by retained closure scratch,
immutable initial captures, and primitive active-thread carriers.

## Bounded bytecode attribution

One preserved bytecode test binary with SHA-256
`368ecc79119e4bfe44640a166fd38f6731a4151c03674359f5572d21fdf076e8`
loaded and typechecked once, warmed once, and measured one default application
call. It reported 769,129,140 ns/op, 14,445,280 B/op, and 156,773 allocs/op.
Its 0.76 seconds of CPU samples report:

| Exact owner | Flat | Cumulative |
| --- | ---: | ---: |
| `execCallMemberArraySlot` | 2.63% | 23.68% |
| `tryStoreRawIntegerSlotValue` | 3.95% | 3.95% |
| `finishArrayReadSlotMemberFast` | 1.32% | 3.95% |
| `finishInlineReturn` | 1.32% | 5.26% |

The generic Array-slot member parent matches the 21.02%-26.14% cumulative
band in the Regex API and Log Routing/Redaction profiles. The lower children
remain split across canonical slot lookup, reads/writes, integer transport,
calls, returns, and runtime Array storage; no missing cache or distinct
invalidation case appears.

## Three-family reconciliation

Operation depth now records portable/sufficient breadth:

- Regex Suffix, Set, and Stream are three public API shapes over one audit
  family;
- Log Routing and Redaction is a classification/privacy application; and
- Configuration Validation and Extraction is a schema/capture application.

The compiled and bytecode frontier groups are therefore closed as rejected
candidate families rather than closed for insufficient breadth. Every
material concrete option is already resolved:

- outgoing-transition indexing, closure scratch, immutable initial captures,
  and primitive active-thread carriers are retained;
- matcher-owned thread arenas and reusable state-position indexes regressed
  the broad regex gate;
- character specialization failed non-regex generality; and
- generic Array-slot/cache, raw-integer, call/return, and carrier alternatives
  failed unlike-program wall-time gates or are already retained.

The new recurrence supplies the required breadth but no causal reason to
reopen any rejection. No performance candidate was built.

## Verification

- Go, Python, Ruby, and default Able bytecode outputs match the external
  verifier; the bounded Able tree-walker emits the expected one-round result;
- all 30 targeted Able timing samples and all 15 targeted reference samples
  verified;
- the promoted scorecard contains 75 selected rows and 82 full-status rows,
  with exactly five successful Able/reference samples for every selected row;
- the static bytecode audit covers 120 programs, 460 functions, and 21,955
  instructions;
- the catalog records 41 portable applications, 42 canonical sources, one
  diagnostic source, and 79 bounded local fixtures; and
- no WASM work was performed.

## Next recommendation

Run a cross-mode cold-start versus steady-state decomposition over at least
three unrelated short compiled applications and three unrelated short
bytecode applications.

Why: the new full scorecard shows many short ordinary programs clustered near
0.10-0.32 seconds compiled and 0.16-0.91 seconds bytecode while their native
or scripting references finish in a few milliseconds. Current main-only
profiles explain hot algorithm walls but deliberately exclude process setup,
generated registration, source loading, typechecking, and lowering. That
fixed cost is broadly paid by real CLI programs and may be the widest
remaining un-attributed wall now that every operation-depth family is
sufficient and every current frontier group is closed.

What it entails: select unrelated short programs from text/iterator,
option/result, and numeric or concurrency families; preserve binaries; record
separate OS/Go initialization, generated bootstrap/registration, stdlib
loading, typecheck/lowering, and measured-main phases; and compare those phase
means with normal five-process wall time. Exclude the already-rejected lazy
bytecode integer-box cache and compiled reachability/cache designs. Advance at
most one exact generic startup owner present in at least three unlike programs
per affected mode, then gate it against long-running target guards and
repeated alternating cohorts. This is runtime/toolchain work only, not WASM.
