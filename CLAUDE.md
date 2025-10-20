# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

**IMPORTANT: Before starting any work, consult AGENTS.md for detailed contributor guidance, collaboration workflows, and project-specific conventions.**

## Project Overview

Able is an experimental programming language with a v10 specification and multiple reference interpreters. This workspace contains:

- **Authoritative specification**: `spec/full_spec_v10.md` - the source of truth for all language behavior
- **TypeScript interpreter**: `interpreter10/` - mature Bun-based implementation with comprehensive tests
- **Go reference interpreter**: `interpreter10-go/` - canonical runtime that must match the spec exactly
- **Shared AST fixtures**: `fixtures/ast/` - JSON test cases used by both interpreters for parity validation
- **Design docs**: `design/` - architectural notes and implementation plans

## Core Development Commands

### TypeScript Interpreter (`interpreter10/`)
```bash
cd interpreter10
bun install                    # Install dependencies
bun test                       # Run unit tests
bun run typecheck             # TypeScript typecheck only
bun run export:fixtures       # Export AST fixtures to JSON
bun run test:fixtures         # Run shared fixture tests
```

### Go Interpreter (`interpreter10-go/`)
```bash
cd interpreter10-go
go test ./...                  # Run all tests
GOCACHE=$(pwd)/.gocache go test ./...  # Run with local cache (used by CI)
```

### Repository-wide Testing
```bash
./run_all_tests.sh             # Run TS tests, fixtures, and Go tests together
```

## Architecture Principles

### AST Contract Alignment
- The AST structure is part of the language contract and must be identical across interpreters
- Go definitions in `interpreter10-go/pkg/ast` serve as canonical but must match TypeScript AST in `interpreter10/src/ast.ts`
- Any AST changes require coordinated updates across both interpreters and shared fixtures
- Future tree-sitter parser must emit compatible AST nodes for all runtimes

### Specification-First Development
- `spec/full_spec_v10.md` is authoritative - if code conflicts with spec, update code
- Record gaps in `spec/todo.md` and update PLAN files when work progresses
- Behavioral divergences between interpreters are treated as bugs to fix immediately

### Interoperability Requirements
- Shared fixtures under `fixtures/ast/` must pass in both interpreters
- When adding fixtures: run `bun run export:fixtures`, then `bun run test:fixtures`, then `go test ./pkg/interpreter`
- Use `scripts/export-fixtures.ts` to maintain canonical JSON representations
- Use `scripts/run-fixtures.ts` to validate behavior matches manifest expectations

## Code Organization

### TypeScript Interpreter Structure
- `src/ast.ts` - AST node definitions and DSL helpers
- `src/interpreter.ts` - single-pass tree-walk evaluator
- `test/*.test.ts` - Jest-compatible unit tests
- `index.ts` - public API exports

### Go Interpreter Structure
The Go interpreter is modularized into focused files:
- `pkg/interpreter/interpreter.go` - core interpreter and module evaluation
- `pkg/interpreter/eval_*.go` - evaluation subsystems (statements, expressions, definitions, imports, impl resolution)
- `pkg/interpreter/interpreter_*.go` - utilities (operations, signals, stringify, members, patterns, types)
- `pkg/interpreter/*_test.go` - test suites mirroring the modular structure
- `pkg/typechecker/` - Go-native typechecker (in development)

### Concurrency Implementation
- **TypeScript**: Uses cooperative scheduler with microtasks for deterministic test execution
- **Go**: Uses goroutines/channels with configurable executor (Serial for tests, Goroutine for production)
- Both expose the same helper functions: `proc_yield()`, `proc_cancelled()`, `proc_flush()`
- Cancellation semantics must be identical across runtimes

## Working with Fixtures

### Creating New Fixtures
1. Add fixture definition to `interpreter10/scripts/export-fixtures.ts`
2. Run `cd interpreter10 && bun run export:fixtures` to generate JSON
3. Add manifest expectations in the fixture definition
4. Run `cd interpreter10 && bun run test:fixtures` to verify TS behavior
5. Run `cd interpreter10-go && go test ./pkg/interpreter` to verify Go parity

### Fixture Manifest Format
```json
{
  "description": "Human-readable description",
  "entry": "module.json",           // Entry point (default: module.json)
  "setup": ["package.json"],        // Optional setup modules for multi-file scenarios
  "expect": {
    "result": { "kind": "i32", "value": 42 },
    "stdout": ["expected output"],
    "errors": ["expected error messages"]
  }
}
```

## Development Workflow

### Making Language Changes
1. Update spec first if changing semantics
2. Update both interpreters to maintain AST parity
3. Add/update shared fixtures to exercise new behavior
4. Run full test suite to ensure no regressions
5. Update relevant PLAN files and design notes

### Adding New Language Features
1. Document feature in spec with examples
2. Add AST nodes to both interpreters
3. Implement evaluation in TypeScript interpreter
4. Port implementation to Go interpreter
5. Create comprehensive fixtures covering edge cases
6. Update documentation and roadmap files

### Debugging Parity Issues
- Use shared fixtures to isolate divergent behavior
- Compare evaluation results between interpreters using same AST JSON
- Check `spec/todo.md` for known specification gaps
- Consult design notes in `design/` for implementation rationale

## Key Constraints

### AST Modifications
- Never modify AST structure without checking impact on both interpreters
- Field semantics, names, and relationships must remain identical
- Consider impact on future tree-sitter parser integration

### Test Coverage
- Every language feature must have corresponding fixtures
- Both interpreters must pass all shared tests
- TypeScript and Go should have equivalent test coverage

### Specification Compliance
- Implementation details may differ but observable behavior must be identical
- Error messages, evaluation order, and edge cases must match spec
- Record any spec ambiguities in `spec/todo.md` for resolution

## Current Priorities

From `PLAN.md`:
- Complete concurrency parity between interpreters (cancellation, yielding, memoization)
- Finish Go typechecker implementation
- Modularize TypeScript interpreter to match Go structure
- Strengthen interface and generics coverage
- Prepare for tree-sitter parser integration

**Refer to AGENTS.md for detailed contributor guidance, modularization targets, and collaboration workflows.**