# Compiled numeric release-readiness closure — 2026-07-26

## Decision

Retain four general compiler corrections and the release-test cleanup described
below. The corrections restore canonical compiled numeric behavior without
changing Able semantics, the canonical stdlib, runtime arithmetic, either
reference interpreter, or the static nominal-lowering rules.

Do not infer a broad performance win from this correctness tranche. The only
execution-path reduction is at a statically recoverable generic interface
boundary. Census that shape before considering any further production
lowering.

## Failing release gates

The ordinary release audit first found repository hygiene failures:

- `pkg/compiler/debug_gen_test.go` permanently wrote generated Go beneath
  `v12/interpreters/go/tmp/able-debug`, so ignored stale files entered
  `go vet ./...` and `go build ./...`;
- two exhaustive source switches retained unreachable trailing returns.

After those were corrected, `./run_all_tests.sh --compiled-cli` exposed two
generated-Go build failures:

1. the canonical UInt128 suite intentionally evaluates
   `340282366920938463463374607431768211456_u128` inside `raise_error()`, but
   generated Go attempted an invalid constant conversion to `runtime.Uint128`;
2. the canonical math suite evaluates `1e308 * 1e308`, but generated Go kept
   the expression constant and the Go front end rejected its overflow instead
   of producing IEEE positive infinity at runtime.

Once those build blockers were removed, the compiled suites exposed two
boundary defects:

- `runtime.Int128` and `runtime.Uint128` had no native-interface
  Go-to-`runtime.Value` encoding;
- recovery of `Matcher<Result<f64>>` probed the sibling `Matcher<f64>`
  interface before the known concrete `BeWithinMatcher`. In the compiled
  bootstrap environment that semantic probe returned an error before the
  concrete native adapter could be recovered.

The existing interpreter-free Int128 execution fixture passed throughout,
which separated native arithmetic from the generic matcher boundary.

## Retained general rules

- Explicit or contextually typed fixed-width integer literals are checked
  against their primitive range before emitting a Go cast. An out-of-range
  literal is represented as the shared runtime integer value so the normal
  typed boundary can produce recoverable Able control rather than invalid Go.
- Native float `+`, `-`, and `*` expressions evaluate operand temporaries
  before the operation. This preserves runtime IEEE behavior and prevents the
  Go compiler from folding an otherwise invalid overflowing constant.
- Native `i128` and `u128` interface adapters use the carriers'
  `IntegerValue()` encoding only when a generic runtime boundary actually
  requires `runtime.Value`.
- Native-interface recovery probes concrete adapters before sibling interface
  adapters. This avoids an unnecessary semantic interface probe when the
  concrete Go carrier is already known.

The release test that previously combined Int128, UInt128, and Rational was
split at canonical source-file boundaries. This preserves coverage and
improves failure localization. Deliberately restrictive GC samples put Int128
and Rational just over one minute, but ordinary isolated reruns completed in
48.98 and 50.01 seconds; the isolated UInt128 sample completed in 53.76
seconds. All three source-file gates therefore remain below the repository's
one-minute test limit under the normal test environment.

The binary-lowering source was also split without behavior change:
`generator_binary.go` is 931 lines and the extracted
`generator_binary_operations.go` is 123 lines.

## Verification

All work used disk-backed `TMPDIR` and build roots under
`/var/tmp/able-release-readiness-20260726`.

- focused literal, IEEE overflow, wide matcher encoding, sibling-interface
  ordering, and related interface guards: pass;
- interpreter-free canonical Int128 execution fixture: pass;
- direct compiled canonical Int128 and UInt128 suites: all 11 cases pass;
- direct compiled math/core numeric suites: all 9 cases pass;
- `go vet ./...`: pass;
- `go build ./...`: pass;
- complete `./run_all_tests.sh --compiled-cli`: pass,
  `able/interpreter-go/cmd/able` in 1,647.592 seconds;
- canonical stdlib tree-walker suite: pass in 20 seconds;
- canonical stdlib bytecode suite: pass in 17 seconds.

No canonical-stdlib, runtime, interpreter, VM, language, dependency,
benchmark, reference, or WASM change was required.

## Next

Run a report-only census of statically known concrete-to-lifted-interface
coercions across the 61 strict applications and representative compiled
stdlib packages.

Why: this correctness failure exposed one remaining box/recover path for a
fully known concrete value:

`BeWithinMatcher -> runtime.Value -> Matcher<Result<f64>>`.

What it entails: classify exact concrete type, source/target interface shape,
generated helper, and callsite; verify final dependency graphs remain
interpreter-free; and advance a direct native coercion only if one general
shape is material in at least three unlike applications. Otherwise retain no
additional production code and close the observation.

Why it is important: it directly tests the project goal of lowering known Able
values to native Go carriers and crossing the compiled/interpreted boundary as
little as semantics allow, without authorizing a matcher-, Result-, stdlib-,
container-, or benchmark-specific fast path.
