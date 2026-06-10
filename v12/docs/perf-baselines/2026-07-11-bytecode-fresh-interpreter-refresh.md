# Bytecode fresh Python/Ruby reference refresh

## Decision

Keep no bytecode VM, compiler, canonical-stdlib, or benchmark-source
performance change. Fresh references expose a real text/iterator rollup
cluster, but steady-state profiles split it into distinct concrete execution
paths. `execCallOpcode` is a broad VM parent, not sufficient authorization for
another raw-value, call-name, return, or named-container specialization.

## Fair reference lane

`v12/bench_refresh_interpreter_refs` now runs the sibling source references
under the same CPU `2` / 45-second process guard used by Able, with every
successful stdout checked by the suite verifier. This refresh used Python
3.14.5 (`python-3.14` source) and Ruby 4.0.5 (`ruby-4.0` source), three
independent processes per completed row.

The runner records unavailable source/verifier rows and stops after the first
timeout while retaining the requested and attempted-run counts. That prevents
repeating a known 45-second timeout merely to manufacture an average. The
current Fib lane is therefore bounded status evidence rather than a comparison
ratio.

The shared external catalog now supplies a run directory. Document Audit,
Lexical Rollup, and Channel Rollup intentionally pass paths relative to the
benchmark repository root; all Go, Able, Python, and Ruby runners now use that
declaration. The correction made their previously failing reference and Able
lanes verifier-clean without changing any runtime behavior.

## Verified three-process scorecard

All Able rows below are independent normal bytecode processes and all outputs
verified. Ratios use only the freshly measured reference rows, never the
stored corpus figures.

| Application | Able bytecode | Python 3.14 | Able/Python | Ruby 4.0 | Able/Ruby |
| --- | ---: | ---: | ---: | ---: | ---: |
| I-Before-E | 0.5900 s | 0.0829 s | 7.12x | 0.1237 s | 4.77x |
| Base64 | 3.0667 s | 3.8880 s | 0.79x | 2.4628 s | 1.25x |
| JSON | 0.8433 s | 2.5497 s | 0.33x | 1.6868 s | 0.50x |
| PiDigits | 2.3533 s | unavailable | — | 12.4325 s | 0.19x |
| Word Frequency | 1.5233 s | 0.0245 s | 62.18x | 0.0639 s | 23.84x |
| Document Audit | 0.3300 s | 0.0147 s | 22.45x | 0.0433 s | 7.62x |
| Lexical Rollup | 0.4467 s | 0.0173 s | 25.82x | 0.0496 s | 9.01x |
| Channel Rollup | 0.5100 s | 0.0427 s | 11.94x | 0.0658 s | 7.75x |

Sudoku remains a one-process bytecode timeout. Fixed Width 128 and Rational
Series reference rows verify (Python/Ruby 0.3996/0.7503 s and 0.1228/0.1706 s)
but their current Able rows fail before timing, so they are retained as status
rows rather than used to select VM work.

## Warm-main separation and CPU evidence

The bytecode-runtime harness loads, lowers, and warms once, then times only
repeated `main()` calls. Its normal-process outputs are already verified above;
the warm measurements classify where their wall time lives.

| Application | Cold process | Warm main | Warm profile result |
| --- | ---: | ---: | --- |
| Word Frequency | 1.5233 s | 1.4960 s/op | Real VM wall: 47.9 MB and 603k allocations/op. `execCallOpcode` is 36.5% cumulative, `execCallName` 22.8%, typed-pattern matching 7.8%, and generic hash-map lookup 4.9%. |
| Document Audit | 0.3300 s | 9.30 ms/op | Startup-dominated. Its short warm profile only samples scattered VM dispatch/caching work and cannot justify a runtime candidate. |
| Lexical Rollup | 0.4467 s | 133.62 ms/op | Real iterator/generator VM work: generator execution 54.5%, `execCallOpcode` 51.5%, member-method lookup 12.5%, and type matching 13.0% cumulative. |
| Channel Rollup | 0.5100 s | 171.26 ms/op | Real scheduler/call work: asynchronous tasks 71.8%, `execCallOpcode` 49.4%, member calls 26.3%, lookup 16.9%, plus atomic/lock costs. |

Word Frequency's material lower leaf is generic map lookup/name-call work;
Lexical Rollup's is generator/member/type dispatch; Channel Rollup adds
spawned-task synchronization. Although all run through `runResumable` and
`execCallOpcode`, no concrete descendant repeats materially across the real
misses. The prior raw-cell, return-guard, and small primitive micro-variants
also failed their broad guards, so retrying them based on this parent would
optimize noise.

No `able-stdlib` source changed. The fresh reports, captured output, and CPU
profiles were temporary analysis artifacts and were removed after this record
was written.

## Next recommendation

Add fresh verifier-backed Python and Ruby references for the existing
K-Nucleotide application, then compare and profile it with Word Frequency and
I-Before-E under the same warm bytecode lane. K-Nucleotide is an independent,
real text-and-frequency application and can determine whether Word Frequency's
generic map/name-call cost recurs outside one program. The work entails source
references plus verifier wiring in the sibling benchmark suite, a pinned
three-process scorecard, and merged warm profiles with JSON as a control.
Only a concrete repeated map or call-dispatch leaf—not the `HashMap` nominal
name or a benchmark corpus—would authorize a VM change.
