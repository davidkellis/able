# Regex redundant-cast profile gate

Date: 2026-07-16

## Scope and decision

Fresh bounded compiled and bytecode profiles were collected for Regex Suffix
Audit, Regex Set Audit, and Regex Stream Audit against the canonical sibling
`able-stdlib` checkout. The only retained change is a source-level cleanup in
`able.text.regex`:

```able
fn regex_char_codepoint(value: char) -> i32 {
  __able_char_to_codepoint(value)
}
```

The removed `as i32` was redundant because the kernel function is already
declared to return `i32`. No compiler, VM, workload, verifier, reference, or
benchmark-specific path changed.

This candidate is retained. It applies to every canonical NFA consumer and
removes work proportional to the number of characters examined, rather than
recognizing any benchmark, pattern, or input.

## Profile selection

The bounded profile workload reads 128 words via each application's existing
`--profile` mode. Each bytecode benchmark executes `main()` five times after
one load/lower/warmup, with `GOMEMLIMIT=1GiB`, `GOGC=50`, and
`GOMAXPROCS=1`. Compiled main-phase profiles use the same 128-word input.

The three CPU profiles repeated the already-retained NFA transition, closure,
and thread-upsert leaves. The new common child was in
`regex_char_codepoint`: compiled output constructed an `i32` type AST and
entered the generic cast bridge on every character even though the callee's
declared result type was already `i32`.

Pre-change bytecode profile means were:

| Application | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: |
| Regex Suffix Audit | 728,294,958 | 25,172,153 | 370,582 |
| Regex Set Audit | 982,896,874 | 28,119,792 | 396,687 |
| Regex Stream Audit | 844,554,444 | 28,653,452 | 425,611 |

The first post-change profiled batch was mixed: Suffix +3.1%, Set -0.9%, and
Stream -2.1%. Two clean five-execution batches were therefore run for each
side, as required for volatile workstation results:

| Application | Pre-change mean | Post-change mean | Change |
| --- | ---: | ---: | ---: |
| Regex Suffix Audit | 726,063,538 ns/op | 714,401,453 ns/op | -1.6% |
| Regex Set Audit | 989,004,086 ns/op | 973,933,263 ns/op | -1.5% |
| Regex Stream Audit | 863,697,527 ns/op | 846,364,497 ns/op | -2.0% |

Combining all three batches makes Suffix effectively neutral (-0.05%) while
Set and Stream remain better. Allocation counts are unchanged, which matches
the VM's already-cheap same-type cast behavior.

## Compiled causal evidence

The exact Suffix main-phase allocation snapshots changed as follows:

| Metric | Pre-change | Post-change | Change |
| --- | ---: | ---: | ---: |
| allocated bytes | 10,982,920 | 7,134,288 | -35.0% |
| allocations | 194,037 | 134,552 | -30.7% |

The pre-change allocation profile attributed 29,631 allocations each to
`ast.NewIdentifier` and `ast.NewSimpleTypeExpression` below
`regex_char_codepoint`. Both disappear after the cleanup: 59,262 type-AST
allocations are removed from this bounded Suffix execution.

Preserved binaries built immediately before and after the one-line change were
then timed with the same command, alternating where startup duration was
short. These direct measurements isolate the source change from compiler and
workstation drift:

| Application | Samples/side | Pre-change mean | Post-change mean | Change |
| --- | ---: | ---: | ---: | ---: |
| Regex Suffix Audit | 5 | 1.432 s | 0.908 s | -36.6% |
| Regex Set Audit | 10 | 0.088 s | 0.063 s | -28.4% |
| Regex Stream Audit | 10 | 0.086 s | 0.063 s | -26.7% |

The verifier-backed post-change harness also completed every requested run.
Its first five-process means were 1.604 seconds for Suffix, 0.144 seconds for
Set, and 0.146 seconds for Stream. A second independent Set/Stream batch was
0.148 and 0.146 seconds. The longer harness means include short-process
startup/scheduling noise; the preserved-binary A/B and allocation profiles
provide the candidate attribution.

## Generality and correctness gates

- Tree-walker and bytecode regex exec-fixture slices pass.
- Isolated compiled fixtures pass for core matching/streaming, RegexSet,
  incremental scanning, and builder/program behavior. Each completes in about
  50 seconds; an attempted eight-fixture parallel process hit the one-minute
  aggregate timeout while all eight compilers were still running, so it is not
  treated as a semantic failure.
- Document Audit, Lexical Rollup, and Reverse Complement each verify 5/5 in
  compiled mode with means of 0.062, 0.072, and 0.082 seconds. All are better
  than their preceding promoted control means, so no unrelated text/non-regex
  regression appears.
- The checked-in full scorecard is not relabeled from these focused reports.

Machine-readable and verifier-backed reports:

- `2026-07-16-regex-redundant-cast-candidate-compiled.json`
- `2026-07-16-regex-redundant-cast-candidate-compiled-repeat.json`
- `2026-07-16-regex-redundant-cast-controls-compiled.json`

## Next recommendation

Audit primitive-return extern lowering across unlike application families,
then profile a candidate only if the same runtime-value/type-AST round trip
repeats in at least three families. Start with regex character conversion, a
filesystem/text primitive-return helper, and a numeric/kernel primitive-return
helper.

Why: this tranche shows that a statically known primitive result can still pay
for runtime boxing, type-AST construction, and generic bridge work in compiled
code. Fixing that at the generic extern-result boundary could benefit all
programs that call primitive-returning kernel functions; merely deleting more
regex casts would be narrower and would risk optimizing library spelling
instead of the shared compiler wall.

What it entails: inventory generated Go and main-phase allocation profiles for
the three unlike families, identify the shared lowering decision, add compiler
tests for primitive types and dynamic/nullable fallbacks, implement a generic
primitive-only lowering when static proof is available, and gate it with five-
process A/B measurements plus unrelated compiled and bytecode controls.
