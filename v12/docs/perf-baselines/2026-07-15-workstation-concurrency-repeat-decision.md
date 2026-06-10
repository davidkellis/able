# Workstation concurrency repeat decision — 2026-07-15

Five verifier-backed workstation runs per Able mode, with fresh five-run
Go/Python/Ruby references, confirm substantial concurrent-application gaps:
Future Await Race compiled is 23.33x Go; await-channel-mux compiled is 82.63x
Go; mutex-await-journal compiled is 233.33x Go. The matching bytecode rows
remain 1.65–8.59x Python and 2.11–3.88x Ruby. These are averaged product
screens, not a scorecard promotion.

They do not admit a new optimization. The completed fresh CPU profiles split
Option/Result into generic-union/allocation work, while all three concurrent
applications share the already rejected `bridge.currentGID` -> `runtime.Stack`
identity route beneath method/await helpers. Do not retry its fixed-context ABI:
independent N-body and K-Nucleotide guards already rejected it.

Keep no compiler, VM, generated-runtime, canonical-stdlib, or workload change.
The machine-readable repeated screen is
`2026-07-15-workstation-concurrency-repeat.json`.
