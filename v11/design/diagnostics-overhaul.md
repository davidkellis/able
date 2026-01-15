# Diagnostics Overhaul (v11)

## Goals
- Provide a shared diagnostics shape for parser, typechecker, and runtime errors.
- Emit consistent human-readable messages across `ablets` and `ablego`.
- Preserve warning/error severity while keeping diagnostics non-fatal in warn mode.

## Diagnostic Shape
- `severity`: `error` or `warning`.
- `message`: human-readable summary, no trailing punctuation.
- `code`: optional stable identifier for tooling.
- `span`: primary source span with `path`, `start(line,column)`, `end(line,column)`.
- `notes`: optional list of `{ message, span? }` for secondary context.

## Formatting (CLI + Fixtures)
- Prefix with `typechecker:` for type diagnostics, `parser:` for parser diagnostics, `runtime:` for runtime errors.
- Warnings prepend `warning:` to preserve compatibility with existing error-only formatting.
- Location prefix uses the primary span start: `path:line:column`.
- If a package label is available, prefix the message with `<package>:`.

Example:
- `warning: typechecker: src/main.able:10:5 mypkg: redundant union member i32`
- `typechecker: src/main.able:10:5 mypkg: undefined identifier 'x'`

## Warning Policy
- Warnings should be emitted for redundant or ambiguous declarations that do not alter runtime behavior.
- Warnings do not block evaluation when `ABLE_TYPECHECK_FIXTURES=warn`.
- Warnings must be surfaced in fixture diagnostics so parity can compare TS/Go outputs.

## Primary + Secondary Spans
- Primary span points at the most relevant token or construct.
- Notes may point at related declarations or conflicting symbols.
- If a span is unavailable, diagnostics should still render with message-only output.

## Union Normalization Rules
- Expand type aliases before comparing union members.
- Expand nullable (`?T`) into `nil | T` for normalization.
- Flatten nested unions (union members of union members).
- Dedupe equivalent members using alias-expanded equivalence.
- Emit a warning for each redundant member detected.
- Collapse single-member unions to that member.
- If the normalized union is exactly `nil | T`, prefer `?T` internally.

## Alias Equivalence
- Aliases are treated as equal to their fully-expanded target for union dedupe and warnings.
- Diagnostics should mention both names when helpful (e.g., `UserID` alias of `u64`).
