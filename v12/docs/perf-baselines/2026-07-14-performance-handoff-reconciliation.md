# Performance handoff reconciliation

The performance vision dated May 2026 still directed follow-on work toward a
typed `i32` quicksort frame/register slice and companion Monte Carlo/PIDigits
micro-optimization paths. That is obsolete. Subsequent bounded gates rejected
the whole-function primitive tier, repeatedly separated the apparent numeric,
text, map/counting, and async parents into different material children, and
rejected the only broad compiled async ABI candidate as a default.

The vision now records the actual selection rule: do not restart old
quicksort/frame, raw-float, scheduler, Array, Mutex, Future, map, or named
container experiments. Use the grouped scorecard refresh after a material
cross-cutting semantic/compiler change, or add a genuinely needed portable
application with shared fixture and foreign-reference coverage. Only profile a
concrete descendant that repeats in at least three unlike verified misses.

Before a future measurement, run `just bench-catalog-check` to confirm the
current portable/local inputs and pass the single-CPU quiet-host guard
(`just bench-host-check --cpu CPU`, then `--require-quiet-cpu` on the timed
entry point). A quiet host alone does not reopen selection.

This is a roadmap correction. It changes no runtime, compiler, stdlib,
benchmark, or language behavior.
