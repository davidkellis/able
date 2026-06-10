# Grouped external scorecard refresh tooling

`just bench-scorecard-refresh` now performs the complete current application
refresh as bounded independent groups instead of relying on one monolithic
benchmark invocation. It refreshes fresh Go/Python/Ruby references, measures
the 16 generality and six standalone async applications in compiled and
bytecode modes, preserves every source scorecard, and aggregates them into one
dated report.

Run it from the repository root:

```sh
just bench-scorecard-refresh --cpu-affinity 15
```

The default is three timed processes per row, a 45-second cap, and the
standard one-process guards (`GOMEMLIMIT=1GiB`, `GOGC=50`,
`GOMAXPROCS=1`). Use `--tag` to name a retained refresh, `--output-dir` to
place its reports elsewhere, `--keep` to retain temporary captures, and
`--dry-run` to inspect all commands without creating files.

The command refuses to overwrite existing dated evidence and never updates
`external-scoreboard-current.*`; that remains an explicit `just
bench-scoreboard` decision. This is measurement plumbing only: it adds no VM,
compiler, bridge, runtime, stdlib, or benchmark behavior.
