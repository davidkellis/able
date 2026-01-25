# Able Test CLI & Framework Protocol (Draft)

## Goals

- Keep tests co-located with production sources so they share package scope and can exercise private APIs when needed.
- Allow builders to omit test code by default while keeping discovery trivial when `able test` runs.
- Support multiple user-space testing frameworks (xUnit, property-based, etc.) behind a consistent CLI contract.
- Make the CLI responsible for locating test modules, applying user-supplied filters, orchestrating execution, and reporting results.
- Keep the protocol purely in Able code where possible so alternative runtimes can adopt it without bespoke host support.

## Non-Goals (Separation of Concerns)

- This protocol is **not** for interpreter fixtures, parity harnesses, or spec
  conformance testing. Those are language implementation tests with separate
  tooling and conventions.
- The user-facing test framework does **not** replace or inform AST/exec fixture
  suites; the two systems are intentionally unrelated.

## File Naming & Build Profile

- Any source ending in `.test.able` or `.spec.able` is considered a **test module**.
- Test modules live in the same package directories as production code; they compile into the same package namespace.
- Standard builds (`able build`, `able run`, `able check`, etc.) ignore test modules unless the caller explicitly enables the `--with-tests` profile flag.
- `able test` enables the `--with-tests` profile automatically and compiles both production and test modules in a single type-checking session so privacy rules remain intact.
- When the compiler emits artifacts, it writes production outputs under the normal target directories and routes test artifacts into a sibling cache namespace (e.g., `target/test/…`) to avoid polluting release products.

## `able test` Command Surface

`able test [OPTIONS] [TARGETS…]`

- **Discovery scope**
  - Defaults to the root package; optional positional arguments restrict the search to specific manifest targets or directories.
  - The CLI recursively searches for `.test.able` / `.spec.able` within the resolved scope, honouring manifest `tests: false` opt-outs (future work).
- **Filtering**
  - `--path PATTERN` / `--exclude-path PATTERN` narrow modules by relative file path (glob syntax).
  - `--name REGEX` / `--exclude-name REGEX` apply to logical test identifiers reported by frameworks.
  - `--tag TAG` / `--exclude-tag TAG` filter by arbitrary labels emitted by frameworks during discovery.
  - Filters are applied by the CLI after framework discovery but before execution so frameworks can expose richer metadata without re-implementing filter logic.
- **Execution controls**
  - `--shuffle [SEED]`, `--fail-fast`, `--parallel N`, `--repeat COUNT` are passed verbatim to frameworks via the run context; frameworks choose how to honour unsupported features.
  - `--list` triggers discovery without execution, printing the resolved plan in a human-readable (or `--format json`) format.
- **Outputs**
  - Human-readable progress (default), TAP-style text (`--format tap`), or JSON events (`--format json`).
  - Exit codes: `0` success, `1` failures, `2` discovery/runtime errors, other codes reserved for CLI/tooling faults.

## Framework Integration Protocol

All test frameworks implement the `able.test.Framework` interface provided by the standard library. Frameworks are ordinary Able modules packaged alongside user code or shipped via the stdlib. Each test file explicitly imports the framework it intends to use (for example, `import able.spec{describe, expect}`); simply importing the module registers the framework instance with the CLI.

```able
package protocol

struct DiscoveryRequest {
  include_paths: Array String,
  exclude_paths: Array String,
  include_names: Array String,
  exclude_names: Array String,
  include_tags: Array String,
  exclude_tags: Array String,
  list_only: bool
}

struct RunOptions {
  shuffle_seed: ?i64,
  fail_fast: bool,
  parallelism: i32,
  repeat: i32
}

struct MetadataEntry {
  key: String,
  value: String
}

struct TestDescriptor {
  framework_id: String,
  module_path: String,
  test_id: String,
  display_name: String,
  location: ?SourceLocation,
  tags: Array String,
  metadata: Array MetadataEntry
}

struct TestPlan {
  descriptors: Array TestDescriptor
}

enum TestEvent =
  | case_started { descriptor: TestDescriptor }
  | case_passed { descriptor: TestDescriptor, duration_ms: i64 }
  | case_failed { descriptor: TestDescriptor, duration_ms: i64, failure: Failure }
  | case_skipped { descriptor: TestDescriptor, reason: ?string }
  | framework_error { message: string }

struct Failure {
  message: string,
  details: ?string,
  location: ?SourceLocation
}

interface Reporter {
  fn emit(self: Self, event: TestEvent) -> void
}

interface Framework {
  fn id(self: Self) -> string
  fn discover(self: Self, request: DiscoveryRequest, register: TestDescriptor -> void) -> void | Failure
  fn run(self: Self, plan: TestPlan, options: RunOptions, reporter: Reporter) -> void | Failure
}
```

### Registration

- Each framework module calls `able.test.registry.register_framework(framework: Framework)` during evaluation (typically directly in the module body that users import).
- Importing the module executes that registration; test files do not need additional wiring beyond the `import`.
- Frameworks may register multiple logical suites by returning unique `framework_id` values or by publishing multiple `Framework` instances.

### Discovery Flow

1. CLI loads every discovered test module (triggering their registration calls).
2. The CLI instantiates a `DiscoveryRequest` populated with path filters; name/tag filters remain empty during this phase to let frameworks emit full metadata.
3. For each registered framework, the CLI invokes `discover`. The framework enumerates its test cases and invokes the supplied `register` callback for each `TestDescriptor`.
4. The CLI aggregates descriptors across frameworks, applies name/tag filters, and constructs the final `TestPlan`.

### Execution Flow

1. CLI groups descriptors by `framework_id` to preserve framework-specific metadata and ordering guarantees.
2. For each group, it calls `run` with the subset of descriptors and the caller-supplied `RunOptions`.
3. Frameworks are responsible for honouring `repeat`, `shuffle_seed`, `parallelism`, and `fail_fast` semantics. If unsupported, they may fall back to defaults but must report the limitation via a `framework_error` event.
4. During execution, frameworks stream structured events to the provided `Reporter`. The CLI multiplexes these events into the chosen output format.
5. Frameworks return `Failure` when execution cannot proceed (e.g., missing fixtures). The CLI treats that as an errored run and stops subsequent frameworks unless `--keep-going` is set (future extension).

### Minimal Framework Example

```able
package mypkg

import able.test.protocol.{Framework, DiscoveryRequest, TestDescriptor, TestPlan, Reporter, RunOptions, TestEvent, Failure}
import able.test.registry{register_framework}

struct MiniFramework;

impl Framework for MiniFramework {
  fn id(self: Self) -> String { "mypkg.mini" }

  fn discover(self: Self, request: DiscoveryRequest, register: TestDescriptor -> void) -> void | Failure {
    register(TestDescriptor {
      framework_id: self.id(),
      module_path: "mypkg/foo.test",
      test_id: "additions_pass",
      display_name: "adds numbers",
      location: nil,
      tags: [],
      metadata: Array.new()
    })
  }

  fn run(self: Self, plan: TestPlan, options: RunOptions, reporter: Reporter) -> void | Failure {
    for descriptor in plan.descriptors {
      reporter.emit(TestEvent.case_started { descriptor })
      ## Run the test body…
      reporter.emit(TestEvent.case_passed { descriptor, duration_ms: 3 })
    }
  }
}

## Module entrypoint (executed when users import this module)
register_framework(MiniFramework);
```

## CLI Responsibilities

- **Module lifecycle**: load/execute test modules exactly once per invocation; ensure registration happens before discovery.
- **Filter enforcement**: apply name/tag/path filters centrally so frameworks can remain simple.
- **Concurrency**: manage cross-framework parallelism; frameworks control internal parallel scheduling but must respect the `parallelism` hint when possible.
- **Reporting**: normalize events from all frameworks into human/TAP/JSON formats. Preserve ordering guarantees (e.g., respecting per-framework serial mode when requested).
- **Exit codes**: synthesize overall status based on aggregated events and detected failures/errors.

## Future Enhancements

- Manifest-driven opt-outs (`package.yml` → `tests: false` or custom globs).
- Tag conventions (`slow`, `integration`, `focus`) and CLI shortcuts.
- Artifact attachments (logs) via extended event metadata.
- Ability for frameworks to expose configuration schemas so `able test` can surface `--framework-opt key=value` flags automatically.

## Open Questions

- Do we want a lightweight mechanism for frameworks to register hierarchical suites (suite → case)? Current descriptors assume flat IDs with optional metadata; we may extend `TestDescriptor` with a tree path.
- Should `SourceLocation` become part of the compiler-emitted metadata for lambdas to avoid manual strings?
- How does this protocol adapt to remote execution (e.g., distributed runners)? Might require serializing plans/events over stdin/stdout or sockets.

## Stdlib Spec-Style Framework (Draft)

### Import & Registration

- Shipped as `package able.spec`.
- Test files opt in explicitly: `import able.spec{describe, expect}` (or use `_` to bring all helpers into scope).
- Module initialization registers the framework with `able.test.registry.register_framework`, so no extra boilerplate is required.

### DSL Surface

- Top-level entry points: `describe(name: string, opts?: SuiteOptions, body: Suite -> void) -> void`.
  - Aliases: `context`, `feature`.
- Within the suite builder:
  - `suite.it(name: string, opts?: ExampleOptions, body: ExampleCtx -> void | !void)`.
  - `suite.it_skip(...)`, `suite.it_only(...)` (future opt-in focus helpers).
  - Hooks: `suite.before_each(body: ExampleCtx -> void | !void)`, `suite.after_each(...)`, `suite.before_all(...)`, `suite.after_all(...)`.
  - Suite-level configuration: `suite.configure(fn(opts: SuiteConfig) -> void)`, `suite.tag(tags: Array string)`.
- `ExampleCtx` carries per-example helpers and exposes `skip(reason?: string)`, `tag(...)`, and `metadata`.
- Examples and hooks register eagerly when the module loads, ensuring discovery is deterministic.

### Expectations & Matchers

- `expect(value)` returns `Expectation T` with:
  - `fn to(self: Self, matcher: Matcher T) -> void`
  - `fn not_to(self: Self, matcher: Matcher T) -> void`
- `Matcher` is an interface with `fn matches(self: Self, actual: T) -> MatcherResult`.
- `MatcherResult` includes `passed: bool`, `message: string`, and optional structured `details`.
- `Expectation.to` raises `AssertionError { message, details }` when `passed` is false. `AssertionError` conforms to the shared `Error` interface.
- Built-in matchers: `eq`, `be_nil`, `be_true`, `raise_error`, etc. Negation uses matcher-provided failure messages so diffs remain meaningful.
- Unexpected exceptions (anything other than `AssertionError`) surface as framework errors with stack traces, while still tagging the example as failed.

### Metadata, Tags, and Filtering

- Each example inherits a breadcrumb array describing its nested suite path. The framework emits:
  - `display_name`: human-readable string (breadcrumbs joined by space).
  - `full_id`: canonical ID using `::` separators for CLI filters (e.g., `Array::when empty::pushes`).
- Tags can be applied at suite or example level; children inherit suite tags by default. `suite.it_tags` allows per-example overrides, `suite.it_only` tags examples with `focus`, and `suite.it_skip` tags with `skip` (emitting `case_skipped`).
- Descriptors carry the full tag set so `--tag`/`--exclude-tag` filters work consistently. Path filters compare against the suite’s declared `module_path` (defaulting to the joined suite key), which test files can override via `suite.module_path("spec/foo")`.
  - Additional metadata keys recorded today: `suite_path`, `module_path`, and a comma-separated `tags` string. CLI consumers can inspect `MetadataEntry` pairs for extensible attributes.

### Execution Policy

- Default order: definition order. The framework reshuffles when CLI passes `RunOptions.shuffle_seed`.
- Parallel runs: the CLI supplies `RunOptions.parallelism`; suites execute serially unless marked `suite.configure(fn(opts) { opts.allow_parallel = true })`. Documentation will emphasize state isolation when enabling parallelism.
- Hooks obey standard spec-style semantics; `before_each`/`after_each` wrap each example, `before_all`/`after_all` run once per suite per worker.

### Reporting Defaults

- Default reporter: documentation-style lines (`Array pushes onto empty array … ok`). Colourised output when terminal capabilities allow.
- Alternate reporters (progress dots, TAP, JSON) consume the same event stream; CLI chooses via `--format`.
- Failures display matcher messages and optional diff snippets derived from `MatcherResult.details`.
- Stdlib provides `able.test.reporters.DocReporter` and `able.test.reporters.ProgressReporter` helpers that accept a simple `(string -> void)` sink; CLI or other tooling can wrap these for console output. Additional helpers like `eq_string` and `be_empty_array` surface richer failure details out of the box.
  - Matchers now cover equality (`eq`, `eq_string`), truthiness (`be_truthy`, `be_false`, `be_nil`), collections (`be_empty_array`, `contain`, `contain_all`), numeric tolerances/ordering (`be_within`, `be_greater_than`, `be_less_than`), regex checks (`match_regex`, currently a placeholder until runtime regex lands), and error expectations (`raise_error`, `raise_error_with_message`). Each returns `MatcherResult.details` with message-ready diagnostics for reporters.

### Harness Helpers

- `able.test.harness.discover_all` aggregates descriptors across all registered frameworks while applying the requested filters (path, name, tag).
- `able.test.harness.run_plan` groups descriptors per framework and reuses the shared protocol to execute them, returning the first framework failure if one occurs.
- `able.test.harness.run_all` combines discovery and execution, making it easy for thin CLI front-ends to wire the stdlib facilities together with any reporter implementation.

### CLI Wiring Checklist (Future Work)

When the `able test` command is implemented, it should:

1. **Load test modules** matching the `.test.able` / `.spec.able` suffix for the selected packages (respecting CLI path filters). Module evaluation triggers framework registration automatically via imports.
2. **Build a `DiscoveryRequest`** from the CLI flags (`--path`, `--name`, `--tag`, `--exclude-*`, `--list`, etc.) and pass it to `able.test.harness.discover_all`. Surface any returned `Failure` as a discovery error and exit with code `2`.
3. **Render discovery-only results** (`--list`) by iterating the descriptors and printing `framework_id`, `display_name`, `module_path`, and tags/metadata; skip execution.
4. **Select a reporter** based on `--format` (doc/ progress/ TAP/ JSON). The doc/progress reporters in `able.test.reporters` can be wrapped to write to stdout/stderr; JSON/TAP reporters can consume the same `TestEvent` stream.
5. **Construct a `RunOptions`** record from CLI flags (`--shuffle`, `--fail-fast`, `--parallel`, `--repeat`). The CLI is responsible for providing a deterministic `shuffle_seed` when shuffling.
6. **Invoke `able.test.harness.run_plan` or `run_all`** with the descriptors and reporter. Stream events to the reporter immediately; failures returned from the harness should map to exit status `1` (test failures) or `2` (framework/runtime error) depending on context.
7. **Handle exit codes** consistently: `0` for success, `1` when any example fails, `2` for discovery/runtime errors, reserving higher codes for CLI faults.
8. **Propagate metadata** (suite path, tags, module path) into any structured output (JSON, TAP diagnostics) so downstream tooling can filter without re-running discovery.

### Configuration Surface

- Suite-level configuration methods set defaults (e.g., fail-fast within the suite, default tags). Helper modules can encapsulate shared policies.
- CLI flags (`--fail-fast`, `--tag`, `--parallel`, `--shuffle`) override suite defaults for the current run, passed through `RunOptions`.
- No external config files initially; all behaviour lives in code, keeping the experience low ceremony and explicit.

Feedback on the interface and event vocabulary is welcome before we wire up the runtime support.
