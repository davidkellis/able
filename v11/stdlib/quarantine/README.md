Quarantined copy of the pre-existing v11 stdlib.

Use this as the source of truth while rebuilding the stdlib incrementally from an empty baseline. When restoring a module, copy it back into `v11/stdlib/src/`, fix parser/typechecker/test issues, and keep the tree green before proceeding to the next module. Recommended order is captured in PLAN.md and `v11/stdlib/src/README.md`.
