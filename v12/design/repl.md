# Able REPL Design (v11)

Status: Draft

## Goals
- Module-every-time evaluation (top-level expressions are allowed).
- Auto-print non-void results.
- Minimal commands: `:help`, `:quit`.
- Cross-platform, deterministic behavior.

## Dependencies
- Extern host execution is spec-correct (§16.1.2).
- Stdlib IO/term APIs (`able.io`, `able.term`).
- Dynamic evaluation API (`dyn.Package.eval` / `dyn.eval`).

## Evaluation Model
- Each user entry is evaluated as a module in a REPL session package.
- The REPL maintains a package name (e.g., `repl.session`) and evaluates every
  entry in that package to preserve bindings across lines.
- The last expression value is returned by `dyn.Package.eval` and printed if it
  is not `void`.

## Input Handling
- Use `able.term` for line editing when TTY is available.
- Fallback to `able.io.read_line` when `is_tty` is false.
- Continuation prompts are used for incomplete input (see parse errors).
- The REPL loop runs under the scheduler so blocking IO suspends only the
  current task, allowing other tasks to keep making progress.

## Parse Errors and Continuations
- `dyn.Package.eval` returns `ParseError { message, span, is_incomplete }`.
- If `is_incomplete` is true, the REPL accumulates more input.
- Otherwise, report the error and reset the input buffer.

## Output
- Use `Display.to_string` when available.
- Fallback formatting uses runtime formatting (existing `print` semantics).
- Do not print results of `void` expressions.

## Commands
- `:help` prints basic usage.
- `:quit` exits with code 0.

## API Surface
- `able.repl` module with `fn main() -> void` and `fn run() -> void`.
- `run()` reads input, evaluates entries, prints results, handles errors.

## CLI Integration
- Add `able repl` subcommand that loads `able.repl` and invokes `main()`.
- CLI config flags may be added later (history file, typecheck mode).
