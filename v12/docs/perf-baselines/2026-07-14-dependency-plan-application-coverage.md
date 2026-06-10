# Dependency-plan application coverage — 2026-07-14

## Scope

`dependency-plan` adds a real portable application where the corpus previously
had only the bounded `toposort_small` fixture. It models release planning for a
deterministic 1,024-service graph. A service unlocks its downstream services
when all predecessors are complete; Kahn's algorithm derives a FIFO deployment
order and contributes each service/order pair to a checksum. Twelve planning
passes produce the output `1024:12595200`.

The portable row has matching Able source under
`v12/examples/benchmarks/dependency_plan/`, an external `run.able`, Go 1.26,
Python 3.14, Ruby 4.0, Docker build lanes, and a single output verifier under
`../benchmarks/dependency-plan/`. It is deliberately in `coverage`, not the
stable `generality` timing suite.

## Verification

All of the following produced or accepted `1024:12595200`:

- Go, Python, and Ruby reference implementations.
- Able tree-walker and bytecode execution.
- An Able-compiled binary built directly into `/tmp`.
- `just bench-catalog-check` and `just bench-scoreboard-check`.
- `v12/bench_bytecode_audit --suite corpus-full`: 109 programs, 410 lowered
  functions, and 20,205 instructions.

No timing or profile was run: the last CPU readiness check found the pinned
host core busy. Static lowering confirms that Dependency Plan resembles the
existing local graph/BFS/topological fixtures rather than revealing a new
shared performance leaf.

## Selection boundary

This application permits future breadth checks for ordinary dependency-plan
code. It does not permit Array-, Queue-, graph-, or deployment-specific
optimization. A source change remains eligible only after a quiet-host
verifier-backed comparison and profiles show the same concrete descendant in
Dependency Plan and at least two unlike applications.
