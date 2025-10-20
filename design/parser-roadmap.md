# Able v10 Parser Roadmap

Date: 2025‑10‑19  
Owner: Able Agents

## Preconditions

- AST contract documented in `design/ast-contract.md`.
- Typechecker plan established (`design/typechecker-plan.md`); any structural
  changes must land there before parser implementation begins.
- Concurrency semantics anchored on the Go interpreter’s goroutine model.
- Safe navigation (`?.`) is deferred to the v11 language cycle; the v10 parser
  must only recognise the existing `.` member access operator.

## Objectives

1. Parse Able v10 source into the existing AST without introducing new node
   shapes.
2. Provide meaningful syntax diagnostics (line/column) while still allowing the
   interpreter to execute well-formed inputs.
3. Share fixtures with the Go interpreter by emitting JSON identical to the
   current `fixtures/ast` layout.

## High-level phases

1. **Grammar definition**
   - Derive a formal grammar from `spec/full_spec_v10.md` and validate it
     against existing fixtures.
   - Encode precedence/associativity rules for expressions, including async
     constructs and cooperative helpers.
2. **Tree-sitter grammar**
   - Implement the Able grammar as a Tree-sitter module (C ABI) so the parser
     can be embedded in editors and tooling.
   - Reuse Tree-sitter’s lexer interface to handle UTF-8 and incremental parsing
     (no custom lexer needed).
3. **Parser integration**
   - Wrap the generated Tree-sitter parser in Go (and other runtimes) to produce
     AST nodes via the helpers in `pkg/ast/dsl.go`.
   - Implement error recovery by inspecting Tree-sitter error nodes and mapping
     them to diagnostics.
4. **Integration harness**
   - Build an `ablec` CLI prototype that reads source, runs the parser, and
     serialises the AST to JSON for fixture parity.
   - Add end-to-end tests that take sample programs through `parser -> AST ->
     interpreter` to ensure behaviour matches the current hand-crafted ASTs.

## Testing strategy

- Reuse existing fixtures by re-parsing their Able source (when available) and
  comparing serialised AST output.
- Introduce parser-specific regression suites covering tricky grammar cases
  (nested patterns, generics, async constructs).
- When the typechecker becomes available, integrate it into the pipeline to
  catch semantic issues early.

## Open questions / future work

- **Source spans on AST** – Tree-sitter supplies byte/row/column information; we
  should consider extending `nodeImpl` with span data once the binding layer is
  in place. Coordinate with the typechecker plan if we go this route.
- **Error recovery heuristics** – Determine whether we favour LL-style
  predictive parsing with synchronisation sets or a GLR-based approach for
  ambiguous constructs.
- **Tooling integration** – Plan how the parser will be invoked by the CLI and
  future language server efforts once typechecking is in place.
