# Able Testing Matchers (Draft)

Able's stdlib ships a growing set of expectation helpers layered on top of `able.testing.expect`. This note captures the current matcher catalog and shows how to compose it in test files.

## Usage Basics

```able
import able.testing.rspec{describe, expect, eq, be_truthy}

describe("Math") { suite =>
  suite.it("adds numbers") { _ctx =>
    expect(2 + 2).to(eq(4))
    expect(1 < 2).to(be_truthy())
  }
}
```

- Use `suite.it_only` to focus a single example (tagged with `focus`; others are filtered out automatically).
- Use `suite.it_skip` to mark pending examples (emits `case_skipped`).

- `expect(value).to(matcher)` raises `AssertionError` when the matcher reports failure.
- `expect(value).not_to(matcher)` negates the matcher with matcher-specific failure messages.
- `suite.it_only` tags an example with `focus`, limiting execution to focused tests.
- `suite.it_skip` tags an example with `skip`, producing a `case_skipped` event.

Custom matchers can be built at the call site:

```able
import able.testing.rspec{expect, matcher}

let even = matcher("expected value to be even", "expected value to be odd", fn(value: i64) -> bool {
  value % 2 == 0
})

expect(4).to(even)

let even_with_details = matcher_with_details(
  "expected value to be even",
  "expected value to be odd",
  fn(value: i64) -> bool { value % 2 == 0 },
  fn(value: i64) -> ?String { `actual ${value}` }
)

expect(6).to(even_with_details)
```

## Example Options

Use `example_options()` to configure tags, focus, or skip behaviour programmatically:

```able
import able.testing.rspec{describe, example_options}

describe("options") { suite =>
  let opts = example_options().tag("fast").focus()
  suite.it_opts("focused", opts, { _ctx => {} })
}
```

- `.tag("fast")` appends a tag.
- `.tags([...])` merges multiple tags.
- `.focus()` marks the example as focused (`focus` tag).
- `.skip()` marks the example as skipped (`skip` tag), producing a `case_skipped` event.

## Equality & Strings

| Matcher | Description |
|---------|-------------|
| `eq(expected)` | Structural equality using `==`. |
| `eq_string(expected)` | String comparison with diff-friendly details. |
| `start_with(prefix)` / `end_with(suffix)` | String prefix/suffix checks. |
| `include_substring(substring)` | Ensures substring appears somewhere in the string. |

## Truthiness & Nil

| Matcher | Description |
|---------|-------------|
| `be_truthy()` / `be_false()` | Boolean truthiness checks. |
| `be_nil()` | Matches `nil` for option-like results. |

## Collections & Arrays

| Matcher | Description |
|---------|-------------|
| `be_empty_array()` | Requires length zero. |
| `contain(value)` | Verifies any array element equals `value`. |
| `contain_all(values)` | Expects every element in `values` to appear in the subject array. |

```able
expect(numbers).to(contain(42))
expect(numbers).to(contain_all([1, 2, 3]))
```

## Numeric Ranges

| Matcher | Description |
|---------|-------------|
| `be_within(delta, target)` | Accepts `target - delta <= actual <= target + delta`. |
| `be_greater_than(threshold)` / `be_less_than(threshold)` | i64 ordering helpers. |
| `be_between(lower, upper)` | Inclusive bounds check for i64 values. |

## Regex

`match_regex(pattern)` delegates to `able.text.regex.regex_is_match`. Until the stdlib regex engine lands, the helper returns a `RegexError` and the matcher falls back to string equality for compatibility.

## Errors

| Matcher | Description |
|---------|-------------|
| `raise_error()` | Expects any `Error` to be thrown. |
| `raise_error_with_message(message)` | Requires an error whose message equals `message`. |

Use inside function expectations: `expect(fn() { dangerous() }).to(raise_error())`.

## Snapshots

Snapshots capture text output and compare it against stored expectations.

```able
import able.testing.snapshots{snapshot_clear, snapshot_set_update_mode}

snapshot_clear()
snapshot_set_update_mode(true)
expect(render_view()).to(match_snapshot("dashboard"))

snapshot_set_update_mode(false)
expect(render_view()).to(match_snapshot("dashboard"))
```

- Enable update mode (CLI flag `--update-snapshots`) to record/refresh snapshots. Otherwise tests fail when snapshots are missing or diverge.
- Custom storage implementations can be injected via `match_snapshot_with_store(name, SnapshotStore { ... })`.
- Current default store is in-memory; CLI will eventually provide a file-backed store.

## Snapshot Store Helpers

`able.testing.snapshots` exposes:

| Helper | Purpose |
|--------|---------|
| `snapshot_set_update_mode(bool)` | Toggles update behaviour (record/overwrite). |
| `snapshot_clear()` | Resets the in-memory store (useful in tests). |
| `SnapshotStore(read, write, update)` | Constructs custom stores (e.g., file-backed). |

## Reporters Recap

- `DocReporter` emits readable lines: `Suite example … ok`.
- `ProgressReporter` shows dot/progress output; call `finish()` after run.
- CLI front-ends should pick the appropriate reporter based on `--format`.

## TODO / Future Work

- Implement file-backed snapshot store and diff formatting.
- Extend matcher set with structural diffs, numeric closeness for integers/floats of other widths, and custom matcher authoring docs.
