# Able v12 Parser Workspace

This directory contains the active Tree-sitter grammar and Go source-to-AST
pipeline for Able v12. The Go parser maps source into the shared AST consumed
by both reference interpreters and the compiler; TypeScript is not part of the
active v12 toolchain.

The language authority is `spec/full_spec_v12.md`. The current feature,
fixture, and test matrix is `design/parser-ast-coverage.md`; the old parser
roadmap and node inventory are historical bring-up records, not syntax
backlogs.

## Current contract

- `tree-sitter-able/grammar.js` defines the active syntax and its generated
  `src/parser.c`, `src/grammar.json`, and `src/node-types.json` stay checked
  in with it.
- `interpreter-go/pkg/parser` maps the concrete tree into the canonical AST.
  It supports the current v12 declaration, expression, type, concurrency,
  error-handling, host-interop, import, and source-re-export surface.
- `spawn` accepts only function calls or blocks; the retired `proc` keyword is
  not grammar syntax. Safe navigation (`?.`) is also not active v12 syntax.
- The default Go parser package round-trips the shared source fixtures against
  their canonical `module.json` representations. `-short` deliberately omits
  that corpus.

## Verification

Run the Tree-sitter corpus after a grammar change:

```sh
cd v12/parser/tree-sitter-able
npm test
```

Run the complete Go parser package, including the default fixture corpus:

```sh
cd v12/interpreters/go
GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1 go test ./pkg/parser -count=1 -timeout 55s
```

For a source fixture change, verify that the generated canonical AST remains
synchronized:

```sh
./v12/export_fixtures.sh --check <fixture-path>
```

## Changing specified syntax

Only add syntax after it is defined in the v12 specification. Update the
grammar and regenerate its checked-in artifacts, add a corpus assertion and a
canonical AST fixture, then add focused Go mapper coverage. Because cgo does
not track an included generated `parser.c`, force the Go parser to relink after
regeneration:

```sh
cd v12/interpreters/go
GOCACHE=$(pwd)/.gocache go test -a ./pkg/parser
```

Finally run the appropriate tree-walker, bytecode, and strict compiler checks
for the new semantic surface. A benchmark requirement is never authority for a
new AST node or parser branch.
