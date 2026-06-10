# Fresh coverage-application lanes (2026-07-12)

## Method

This refresh measures the six applications newly included in the broad
`coverage` catalog—Fixed Width 128, Rational Series, Word Frequency, Document
Audit, Lexical Rollup, and Channel Rollup—with JSON as an unrelated control.
Fresh Go 1.26.4, Python 3.14.5, and Ruby 4.0.5 references used three
CPU-2-pinned, verifier-backed processes each. Compiled and bytecode Able used
three processes under the same 45-second guard, canonical stdlib, and output
verifier. The Channel Rollup lane uses its declared goroutine executor.

The first Fixed Width/Rational pass correctly recorded launcher errors rather
than timings: the sibling input directories contain same-named reference
packages. `bench_perf --source-root-only` now sets `ABLE_SOURCE_ROOT_ONLY=1`
for explicit-entry builds/runs, so the entry directory remains the source root
while the process CWD remains available for input files. The normal CLI and
compiler defaults are unchanged. The catalog marks only those two data-layout
applications for this generic launcher mode.

## Fresh verified results

| Benchmark | Mode | Able (s) | Go ratio | Python ratio | Ruby ratio | Status |
| --- | --- | ---: | ---: | ---: | ---: | --- |
| Fixed Width 128 | compiled | 7.3567 | 1268.40x | 20.91x | 11.35x | verified 3/3 |
| Fixed Width 128 | bytecode | 7.8000 | 1344.83x | 22.17x | 12.03x | verified 3/3 |
| Rational Series | compiled | 2.0833 | 156.64x | 20.38x | 15.70x | verified 3/3 |
| Rational Series | bytecode | 3.9300 | 295.49x | 38.45x | 29.62x | verified 3/3 |
| Word Frequency | compiled | 0.2433 | 46.79x | 12.67x | 4.78x | verified 3/3 |
| Word Frequency | bytecode | 1.4233 | 273.71x | 74.13x | 27.96x | verified 3/3 |
| Document Audit | compiled | 0.1033 | 25.20x | 7.65x | 2.47x | verified 3/3 |
| Document Audit | bytecode | 0.3067 | 74.80x | 22.72x | 7.32x | verified 3/3 |
| Lexical Rollup | compiled | 0.1267 | 30.17x | 7.45x | 2.55x | verified 3/3 |
| Lexical Rollup | bytecode | 0.3933 | 93.64x | 23.14x | 7.91x | verified 3/3 |
| Channel Rollup | compiled | 1.5733 | 302.56x | 39.43x | 29.80x | verified 3/3 |
| Channel Rollup | bytecode | 0.4500 | 86.54x | 11.28x | 8.52x | verified 3/3 |
| JSON control | compiled | 0.7267 | 0.49x | 0.28x | 0.43x | verified 3/3 |
| JSON control | bytecode | 0.8200 | 0.56x | 0.32x | 0.48x | verified 3/3 |

The fresh reference means are included in the generated reports. In particular,
the control remains faster than all three references, so these process gaps do
not authorize a generic startup or external-reference adjustment.

## Profile gate

No new runtime or compiler profile is authorized. Current-source evidence
already separates every apparent pair:

- Fixed Width 128 is the public `UInt128` checked-member/two-word path;
  Rational Series is ratio/nominal arithmetic. Their paired profiles diverge,
  and the raw-value audit rejects a `UInt128`, BigInt, or local boxing rule as
  a general value representation change.
- Word Frequency is string-key map/name-call work. Document Audit is
  iterator/member-cache/public-return work; Lexical Rollup is filesystem,
  generator, typed-pattern, and iterator work. Their shared dispatcher names
  are parents, not common leaves; previous member-cache, call-name, raw-cell,
  and inline-return variants failed broad guards.
- Channel Rollup's goroutine scheduler/task work is not material in the serial
  iterator/text controls or BinaryTrees in the same way. It does not authorize
  a channel, Future, scheduler, or coroutine-shape shortcut.

The fresh ratios prove these applications belong in broad guards. They do not
override prior source-aligned attribution or permit a named nominal type,
container, corpus, fixed-width, ratio, or concurrency special case.

## Decision and next recommendation

Keep the generic source-root launcher repair and no performance code. The
comparison path now faithfully measures all six application lanes, while their
results confirm no repeated eligible helper across the coverage extension.

Next, publish one consolidated current 22-application cross-runtime status
ledger by combining the versioned 16-program `generality` baseline with these
six fresh lanes. Why: the project now has current verified evidence for every
application-shaped benchmark, but it is split between reports; one ledger will
make the 95%-of-Go and near-Python/Ruby gaps auditable before any architecture
or implementation proposal. The work entails provenance-preserving joins only,
explicit timeout/status handling, and a source/profile audit only if the
consolidated rank reveals a previously unseen repeated non-nominal helper.
