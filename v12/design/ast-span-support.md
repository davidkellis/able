# Able v11 AST Span Support

## Motivation

Program-wide diagnostics now include per-package summaries and source hints, but
those hints currently surface only best-effort file paths. To unlock precise
line/column reporting across the CLI, fixtures, and downstream tooling, the
canonical AST needs to carry span metadata for every node. This document
captures the plan for introducing span support without breaking existing
consumers.

## Goals

- Annotate every AST node constructed by the parser with a start/end span
  recorded in `nodeImpl`.
- Thread the span data through the driver loader into the typechecker so
  diagnostics can report exact locations.
- Preserve backwards compatibility for embedders by providing helper functions
  (e.g. `SetSpan`) that allow future tools or hand-built ASTs to opt-in to span
  metadata gradually.
- Cover the change with parser/unit/CLI tests so regressions are caught quickly.

## Approach

1. **AST foundation**
   - Extend `nodeImpl` with a `Span` struct (`start`, `end`, each containing
     one-based `line`/`column` fields).
   - Expose helper setters/getters (`ast.SetSpan`, `Node.Span()`) so code
     constructing nodes manually can annotate them.

2. **Parser instrumentation**
   - Add parser-level helpers (`annotateSpan`, `spanFromNode`) that map
     tree-sitter `Point`s to the AST `Span`.
   - Update every parser routine that constructs AST nodes to call
     `annotateSpan` immediately after creation.
     * This includes module-level constructs (imports, package statements),
       declarations (structs, unions, interfaces, methods, impls), expressions,
       patterns, and type expressions.
     * For composite nodes that reuse constituent nodes (e.g. tuple struct
       fields), make sure to annotate both the composite and the nested
       definitions.

3. **Driver + typechecker integration**
   - Replace the existing reflection-based origin map with span-aware
     origin data derived from the parser (while keeping a fallback for older
     modules).
   - Adapt the typechecker diagnostic formatting to include `path:line:column`
     when span information is available, retaining the previous behaviour for
     nodes that still lack spans.

4. **Testing**
   - Parser unit tests: extend golden tests to assert span fields for a
     representative sample of nodes (identifiers, struct fields, expressions,
     etc.).
   - Typechecker tests: verify diagnostics now include line/column offsets.
   - CLI tests: ensure `able run` outputs the enhanced diagnostic format and
     remains stable for success cases.
   - Fixture harness: confirm spans are attached when fixtures are parsed so
     warn-mode checking reports precise locations.

5. **Documentation & communication**
   - Update README/design docs to mention span availability.
   - Highlight that custom AST builders should call `ast.SetSpan` when they
     want span-aware diagnostics.

## Incremental delivery

Given the breadth of parser changes, the work will land in incremental commits:

1. Land AST + helper scaffolding (already prototyped).
2. Instrument declarations (structs, functions, methods, impls) with spans +
   tests.
3. Instrument expressions/patterns/type expressions + tests.
4. Switch loader/typechecker/CLI to consume spans, update diagnostics/tests.

Each step will include targeted tests so we never regress existing behaviour.
