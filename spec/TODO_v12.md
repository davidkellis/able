# Able v12 Spec TODOs

This list tracks the remaining v12 items after audit; completed work should be removed.

## Parser gaps

- Newline + space-delimited type application can still consume next-line identifiers. Semicolons are used in cast fixtures as a temporary workaround; needs a real parser fix (scanner/grammar refactor).
