# Scorecard Verifier and Declared-Input Fingerprints

## Decision

Each verifier-backed external benchmark row now carries a compact contract for
the correctness verifier and its explicitly declared program arguments. The
contract records:

- the verifier script path and SHA-256 when a verifier exists;
- every declared program argument, including literal values such as sizes;
- the path and SHA-256 of each declared argument that resolves to a regular
  file from the benchmark's run directory; and
- a SHA-256 over that normalized manifest.

`bench_compare_external` captures the contract before it launches a timed
Able process and labels it `measured`. It intentionally does not attempt to
discover arbitrary files opened implicitly by a benchmark; the catalog is the
declared input contract.

## Legacy cohort and promotion

The July scorecard reports predate the capture field. The report-only
scoreboard reconstructs their contracts from `bench_external_catalog.sh`,
labels them `current`, and validates every referenced verifier/input file on
each regeneration. A changed verifier, declared input, or literal argument
makes `just bench-scoreboard-check` stale. A bare `just bench-scoreboard`
also refuses to overwrite the current report after a contract digest changes;
a refreshed explicit scorecard is required to promote the new measurement.

The generated current report contains 64 row contracts, with 64 verifier
fingerprints and 22 declared input-file entries. All are `current` legacy
reconstructions; no timing workload was run for this provenance update.

## Verification

- `bash -n v12/bench_external_catalog.sh v12/bench_compare_external`
- `./v12/bench_external_catalog.sh contracts ../benchmarks fib i_before_e document_audit`
- `python3 -m py_compile v12/bench_external_scoreboard`
- `./v12/bench_external_scoreboard --write-current`
- `./v12/bench_external_scoreboard --check`
- In-memory fresh-contract normalization and bare-promotion-drift guard check.

No VM, compiler, runtime, stdlib, benchmark program, reference timing, or
performance behavior changed.
