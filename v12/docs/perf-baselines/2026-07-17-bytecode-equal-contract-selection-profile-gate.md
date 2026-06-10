# Bytecode equal-contract selection and profile gate

Date: 2026-07-17

## Decision

Keep no bytecode VM, compiler, runtime, canonical-stdlib, application,
reference, verifier, fixture, spec, or WASM change. A fresh equal-contract
screen covers all 28 applications in the reviewed bytecode selection. Only
JSON and PiDigits meet both interpreter targets among the 25 fully rankable
rows. Three unlike long misses were then profiled as ordinary one-process
bytecode executions; they share VM/call parents but no new concrete material
descendant.

The closest exact shared children are slot loading and checked operand-stack
append. They account for only 3-6% in the two numeric programs and 15-16% in
Reverse Complement, and their generic raw-carrier/ordering candidates have
already failed repeated broad guards. No candidate is admitted from a parent
frame or a previously rejected family.

## Measurement contract

The selected cohort comes directly from `bench-selection-manifest.json`.
Python 3.14, Ruby 4.0, and ordinary Able bytecode processes received CPU 0 and
a 55-second per-process cap. Able additionally used `GOMAXPROCS=1`,
`GOGC=50`, and `GOMEMLIMIT=1GiB`; concurrency applications used the catalog's
explicit goroutine executor. Each available row requested five independent
processes. Every successful stdout passed the application's public Ruby
verifier, and the reference harness stopped a row after its first timeout.

Reference implementations produced 265 verified executions, four timeouts,
and no failures across 269 attempts. Able produced 138 verified executions,
two K-Nucleotide timeouts, and no failures. Twenty-seven Able rows completed
5/5; K-Nucleotide completed 3/5 and remains unranked rather than being averaged
into a release claim.

The canonical stdlib source remained the external `../able-stdlib` checkout:
69 Able source files at Git head `219eff222c28406487231713753641bc49ee5b9a`,
with its pre-existing dirty worktree preserved.

## Fresh scorecard

Ratios are Able elapsed time divided by the matched reference mean; lower is
better. Meeting the target requires both ratios to be at most 1.0526x and all
three implementations to have complete repeated evidence.

| Application | Able mean | Able/Python | Able/Ruby | Status |
| --- | ---: | ---: | ---: | --- |
| Array Slice Window | 1.0000s | 22.62x | 13.59x | miss |
| Await Channel Mux | 0.3180s | 2.42x | 2.78x | miss |
| Base64 | 3.3420s | 0.69x | 1.10x | miss |
| Channel Rollup | 0.6080s | 10.86x | 9.57x | miss |
| Dependency Plan | 0.5640s | 30.49x | 10.52x | miss |
| Distance Field | 6.2360s | 9.43x | 16.08x | miss |
| Document Audit | 0.3900s | 23.35x | 7.66x | miss |
| Fib | 0.1940s | n/a | incomplete Ruby | unranked |
| Fixed Width 128 | 9.9520s | 17.40x | 11.39x | miss |
| Future Await Race | 0.3060s | 7.91x | 3.68x | miss |
| Future Pipeline | 0.6380s | 8.20x | 6.02x | miss |
| I-Before-E | 0.7120s | 5.62x | 4.49x | miss |
| JSON | 1.0760s | 0.35x | 0.60x | **meets** |
| K-Nucleotide | 49.4367s (3/5) | 34.76x | 36.48x | unranked mixed timeout |
| Lexical Rollup | 0.9680s | 43.02x | 18.44x | miss |
| Mandelbrot | 8.4380s | 6.71x | 3.97x | miss |
| Matrix Multiply | 5.8980s | n/a | incomplete Ruby | unranked |
| Monte Carlo Pi | 3.2980s | 1.79x | 1.75x | miss |
| Mutex Await Journal | 0.2700s | 10.67x | 5.16x | miss |
| Mutex Ledger | 0.4100s | 6.39x | 6.25x | miss |
| Option/Result Configuration | 1.3740s | 63.32x | 26.07x | miss |
| PiDigits | 2.8140s | 0.62x | 0.24x | **meets** |
| Rational Series | 4.5760s | 39.62x | 29.89x | miss |
| Regex Set Audit | 6.2180s | 286.54x | 127.94x | miss |
| Regex Stream Audit | 5.6560s | 261.85x | 107.53x | miss |
| Reverse Complement | 8.3780s | 243.55x | 101.55x | miss |
| RMS Norm | 4.8720s | 4.94x | 8.01x | miss |
| Word Frequency | 1.5780s | 64.41x | 24.05x | miss |

Thus 2/25 rankable rows currently meet the 95%-of-Python and 95%-of-Ruby
objective. Base64 clears Python but misses Ruby, so it is correctly a product
miss rather than a partial pass. Fib and Matrix Multiply have incomplete
foreign-reference cohorts, while K-Nucleotide's Able cohort crossed the cap;
none is assigned a target result.

## Unlike normal-process profiles

The profile cohort was selected by absolute Able time and semantic diversity,
not by a hoped-for helper:

- Fixed Width 128 performs nominal two-word unsigned arithmetic;
- Distance Field performs primitive floating-point geometry; and
- Reverse Complement performs text parsing and mono-u8 Array transformation.

One additional ordinary process per application passed its verifier. These
were whole bytecode CLI profiles, not warmed repeated-main benchmark helpers.

| Application | Profile wall / CPU samples | Material concrete descendants |
| --- | --- | --- |
| Fixed Width 128 | 7.88s / 7.74s | UInt128 checked member path 28.94%; UInt128 add 23.90%; boxed integer conversion 12.14%; UInt128 extraction/materialization 10.47%/9.95%; allocation/GC |
| Distance Field | 5.99s / 5.87s | static-member calls 24.70%; float slot store 11.41%; binary operations 10.05%; inline return 8.35%; raw-float slot/value conversion |
| Reverse Complement | 7.52s / 7.39s | Array-slot member calls 29.50%; slot load 16.24%; checked stack append 15.43%; Array push 14.75%; boxed/snapshot integer work |

`runResumable`, the public call chain, and `execCallOpcode` are shared parent
frames. Intersecting every exact interpreter symbol at at least 2% cumulative
leaves only two children beyond those parents:

| Exact helper | Fixed Width | Distance Field | Reverse Complement |
| --- | ---: | ---: | ---: |
| `execLoadSlotOpcode` | 5.94% | 3.75% | 16.24% |
| `appendSlotStackValueChecked` | 4.65% | 3.07% | 15.43% |

Neither is material across all three. More importantly, the coverage-wide
symbol census already found stack append in 13 applications and tested two
generic ordering/carrier variants; both failed broad application guards. A
direct float slot-load shortcut was also neutral-to-regressive, while raw-cell
ownership variants moved boxing or violated stable argument snapshots. The
fresh result therefore confirms, rather than reopens, those decisions.

## Verification and cleanup

- 138 Able scorecard executions verified; two timed out; none failed.
- 265 Python/Ruby reference executions verified; four timed out; none failed.
- Three normal-process profile executions verified.
- No candidate-specific Go test was required because no source candidate was
  built.
- Generated CLIs, stdout captures, raw profiles, and temporary JSON workspaces
  were removed after this record and the PLAN handoff were written.

## Next recommendation

Make CPU/executor budgets first-class benchmark-contract metadata before the
next promoted combined scorecard.

Why: the preceding compiled reconciliation proved that a global one-P contract
misclassified Binary Trees by several-fold, while the current bytecode screen
correctly needs a single-CPU interpreter contract. One global affinity setting
cannot express both serial interpreter comparisons and intentionally parallel
compiled applications. If the suite is to drive optimization, its CPU budget
must be part of each portable application's audited contract rather than an
operator convention.

What it entails: add a small catalog field for the intended logical-CPU budget
and executor policy; make Able, Go, Python, and Ruby refresh paths derive one
shared `taskset` allocation from it; record the resolved contract in every JSON
row; reject aggregation when matched implementations used different budgets;
and add fast dry-run/schema tests for serial, parallel, concurrency, and
missing-metadata cases. Then regenerate the combined scorecard in bounded
five-process groups. This changes measurement infrastructure only; it must not
change benchmark algorithms, language/runtime behavior, stdlib code, or
reference implementations.
