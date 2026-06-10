# Primitive kernel-helper native lowering

Date: 2026-07-16

## Decision

Retain a compiler-only primitive lowering for the six scalar kernel helpers:

- `__able_char_from_codepoint(i32) -> char`
- `__able_char_to_codepoint(char) -> i32`
- `__able_char_simple_fold_next(char) -> char`
- `__able_f32_bits(f32) -> u32`
- `__able_f64_bits(f64) -> u64`
- `__able_u64_mul(u64, u64) -> u64`

Statically typed calls now pass native Go scalars to typed generated helpers
and receive native scalar results. The old `[]runtime.Value` helper remains
the semantic fallback for dynamic calls. Local bindings that shadow a kernel
helper name continue to use ordinary callable resolution. Invalid codepoints
retain source-aware error control, and numeric bit/multiply semantics are
unchanged.

This is a primitive language/kernel boundary, not a Regex, HashMap, benchmark,
or named nominal rule. No bytecode VM, stdlib, application, verifier, reference,
or benchmark source changed.

## Audit

The audit separated two superficially similar boundaries:

1. Source-level Go externs already lower primitive arguments and results
   directly. Generated filesystem code, for example, returns primitive host
   results without `HostValueToRuntime`; no change was warranted there.
2. Toolchain kernel helpers still boxed every scalar argument, built a
   `[]runtime.Value`, called an `_impl` helper, and unboxed the scalar result.
   The same generated form existed for character conversion, float bitcasts,
   and wrapping `u64` multiplication.

Generated post-change code proves that canonical Regex calls
`__able_char_to_codepoint_native(value)` and the canonical `Hasher.write_u8`
calls `__able_u64_mul_native(mixed, FNV_PRIME)`. No static call site uses the
boxed `_impl([]runtime.Value{...})` form for these six helpers.

## Exact allocation gate

Main-phase allocation snapshots used preserved binaries built immediately
before and after the compiler change.

| Application | Baseline bytes | Candidate bytes | Change | Baseline allocations | Candidate allocations | Change |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Regex Suffix (128 words) | 7,176,840 | 5,716,928 | -20.3% | 134,648 | 104,882 | -22.1% |
| Word Frequency | 33,967,208 | 33,968,944 | +0.01% | 722,227 | 722,248 | +0.00% |
| K-Nucleotide | 1,605,612,944 | 1,605,543,672 | -0.00% | 35,748,392 | 35,748,005 | -0.00% |
| Document Audit | 2,777,752 | 2,777,784 | +0.00% | 2,058 | 2,058 | 0.00% |

The Regex profile loses all 29,628 allocations formerly attributed to
`__able_char_to_codepoint_impl`. The other applications prove that merely
emitted helpers do not change allocation behavior when the helpers are not hot.

## Repeated timing gate

Alternating preserved baseline/candidate binaries avoids compiler-version and
source drift. High-resolution process timing alternated which side ran first.

| Application | Pairs | Baseline mean | Candidate mean | Change |
| --- | ---: | ---: | ---: | ---: |
| Regex Suffix Audit | 10 | 0.949809 s | 0.702732 s | -26.0% |
| Word Frequency | 20 | 0.134259 s | 0.135690 s | +1.1% |
| Document Audit | 20 | 0.032236 s | 0.031749 s | -1.5% |
| K-Nucleotide | 5 | 1.982 s | 1.990 s | +0.4% |

The unrelated rows are neutral and their exact allocation counts are also
neutral.

Verifier-backed five-process runs completed 5/5 for every requested row. Regex
Suffix is stable at 0.750 seconds (2.67% CV). The short Set and Stream rows were
volatile, so each received a second five-process batch:

| Application | Candidate samples | Candidate mean | Prior immediate mean | Change |
| --- | ---: | ---: | ---: | ---: |
| Regex Suffix Audit | 5 | 0.750 s | 1.604 s | -53.2% |
| Regex Set Audit | 10 | 0.089 s | 0.146 s | -39.0% |
| Regex Stream Audit | 10 | 0.095 s | 0.146 s | -34.9% |

The prior means come from the immediately preceding redundant-cast tranche,
with the same canonical stdlib source. The preserved-binary A/B is the causal
gate; the verifier-backed reports confirm application output and broader API
coverage.

Evidence artifacts:

- `2026-07-16-primitive-kernel-helper-candidate-regex.json`
- `2026-07-16-primitive-kernel-helper-candidate-regex-variance.json`
- `2026-07-16-primitive-kernel-helper-controls.json`
- `2026-07-16-primitive-kernel-helper-controls-variance.json`
- `2026-07-16-primitive-kernel-helper-volatile-repeat.json`

## Correctness and containment

- Primitive-helper lowering, shadowing, fallible error-control, ordinary Go
  extern, and direct runtime-core tests pass.
- The compiled primitive-hash and regex-core fixtures pass.
- Tree-walker and bytecode primitive-hash/regex-core fixtures pass; no bytecode
  implementation changed.
- `ablec` builds.
- All touched Go source files remain below 1,000 lines; moving the existing
  helper registry reduced `generator_exprs_calls_lambda.go` to 983 lines.
- Diff hygiene and promoted-scoreboard replay pass.

## Next recommendation

Refresh bounded bytecode CPU/allocation/trace profiles for K-Nucleotide, Word
Frequency, and Regex Set/Stream, then advance only a VM or kernel boundary that
repeats across at least three of those unlike programs.

Why: this tranche removes the current compiled primitive-helper wall, while the
bytecode scorecard remains much farther from the Python/Ruby target. These
applications combine `u64`-key maps, String-key maps, NFA arrays, integer
traffic, and unlike algorithms, making them a strong screen for a genuinely
shared raw-integer extraction, lookup, boxing, return, or type-match cost.

What it entails: collect fresh one-process bounded profiles under the 1 GiB
guardrails, retain per-call traces and exact allocation counts, reconcile them
against already-closed String/NFA candidates, implement a candidate only if
the same leaf repeats across three workloads, and gate it with semantic tests,
two five-execution batches for volatile rows, and unrelated numeric/iterator
controls.
