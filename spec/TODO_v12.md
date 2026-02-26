# Able v12 Spec TODOs

This list tracks the remaining v12 items after audit; completed work should be removed.

## Parser gaps
- None currently tracked (cast `as` line breaks fixed 2026-02-06).

## Compiler AOT gaps
- Evaluate whether no-interpreter static runtimes should add runtime-independent alias/constraint revalidation for generic interface dispatch, or keep this permanently compile-time-only and document it as final.
