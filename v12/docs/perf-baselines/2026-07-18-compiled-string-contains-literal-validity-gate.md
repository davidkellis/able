# Compiled literal `String.contains` validity gate

Date: 2026-07-18

## Decision

Retain a compiler-owned, primitive-String call-site lowering for canonical
`String.contains` calls whose receiver is a stable identifier and whose needle
is a String literal.

Able String literals are valid UTF-8 by construction. The lowered call checks
only the runtime receiver, calls Go's `strings.Contains` directly when that
receiver is valid, and enters the existing canonical Able method otherwise.
The fallback preserves `StringEncodingError` for strings created through
`String.from_bytes_unchecked`. Dynamic needles, computed receivers,
user-defined methods, and all non-String nominal types retain the normal
compiled-call path.

This is a primitive language-type rule, not an I-Before-E, iterator, regex,
filesystem, or named-container rule. The canonical external `able-stdlib`
required no change.

## Admission and provenance audit

Fresh generated-main profiles covered three existing applications with unlike
outer algorithms:

- I-Before-E directly filters file lines.
- Document Audit runs document predicates through a lazy iterator pipeline.
- Lexical Rollup maps and filters a bounded lexical pipeline.

All three naturally call canonical `String.contains`. Before the change,
I-Before-E's 50-profile merge contained 1.61 s of main samples:
`utf8.ValidString` was 120 ms / 7.45% flat, the compiled `String.contains`
entry was 290 ms / 18.01% cumulative, its method body was 120 ms / 7.45%
cumulative, and host substring search was 130 ms / 8.07%.

The broader validity fact proposed by the preceding tranche was rejected.
Unchecked String construction is public, and the canonical filesystem Go fast
paths currently construct Go strings without exposing a compiler-visible
validation proof. The hot receivers also cross parameters and lazy
iterator/container boundaries. Treating all String values or filesystem
results as valid would therefore be unsound.

The admitted fact is narrower but general: the literal needle is always valid.
Across one complete invocation of each application, the calls and bytes that
the old method revalidated were:

| Application | Calls | Receiver bytes | Literal-needle bytes | Receiver / needle |
| --- | ---: | ---: | ---: | ---: |
| I-Before-E | 174,707 | 1,588,093 | 351,170 | 4.52x |
| Document Audit | 7,765 | 527,703 | 48,196 | 10.95x |
| Lexical Rollup | 27,776 | 227,449 | 27,776 | 8.19x |

The lowering skips only the guaranteed-valid needle check and normal entry
wrapper. It deliberately retains receiver validation.

## Preserved-binary A/B

The baseline binaries were built and preserved before the compiler edit. The
candidate binaries were then built from the edited compiler. Every timed
process used CPU 0, `GOMAXPROCS=1`, `GOMEMLIMIT=1GiB`, `GOGC=50`, the catalog
working directory and arguments, and the public Ruby verifier. Forward and
reverse cohorts balanced ordering. All 300 processes verified, and every
sample—including the 0.20 s and 0.21 s baseline I-Before-E outliers—remains in
the arithmetic means and the companion JSON record.

| Application | Samples per binary | Baseline wall | Candidate wall | Change | Baseline user | Candidate user | Change |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| I-Before-E | 50 | 0.1268 s | 0.1218 s | -3.94% | 0.0886 s | 0.0838 s | -5.42% |
| Document Audit | 50 | 0.0920 s | 0.0908 s | -1.30% | 0.0600 s | 0.0604 s | +0.67% |
| Lexical Rollup | 50 | 0.1008 s | 0.1020 s | +1.19% | 0.0698 s | 0.0698 s | 0.00% |

Document Audit and Lexical Rollup remain below the 10 ms wall-clock timer's
useful resolution at the individual-process level; their opposite 1.2%-1.3%
wall movements and unchanged user CPU are treated as neutral, not wins or
regressions. I-Before-E was expanded from 10 to 50 samples because its cohort
means remained volatile. Its aggregate improvement is corroborated by user
CPU and the post-change profile rather than selected outlier removal.

The complete per-run record is
`2026-07-18-compiled-string-contains-literal-validity-gate.json`.

## Post-change attribution

Fifty fresh candidate phase-profile processes all verified. The merged main
profile fell from 1.61 s to 1.55 s of samples. `utf8.ValidString` fell from
120 ms to 90 ms, a 25% reduction, and the compiled `String.contains` entry
wrapper no longer appeared in the hot table. The remaining 90 ms is expected:
receiver validity is not proven and other canonical String operations still
validate. The profile therefore confirms that the wall result follows the
intended generic work removal.

## Correctness and controls

- The lowering unit tests cover canonical recognition, literal admission,
  dynamic-needle rejection, native search, and canonical fallback.
- An executable compiler regression proves that a valid receiver succeeds and
  an unchecked invalid UTF-8 receiver still raises an Able `Error` through the
  fallback.
- The existing checked `String.from_bytes` validation regression passed in
  20.624 s; the Unicode `String.chars` validation regression passed in
  30.857 s.
- Focused bytecode/tree-walker String conversion and canonical fast-path tests
  passed in 0.362 s. The compiler-only change does not alter either interpreter.
- A source/import-closure audit found no eligible literal `String.contains`
  call in Word Frequency, Reverse Complement, Regex Suffix/Set/Stream, or the
  canonical stdlib code they use; those unrelated controls cannot receive the
  new call-site form.
- `git diff --check` passed.

Two over-broad test invocations were discarded rather than reported as
successes. A four-test compiler group exceeded its 60-second package budget
while the fourth external compiled program was running. The standalone heavy
regex compiled fixture likewise exceeded 60 seconds waiting on its external
compiled program. Neither reached a failed semantic assertion; neither was
rerun with a prohibited longer timeout. Their relevant bounded individual and
static controls are listed above.

## Next recommendation

Run a post-change compiled selection refresh before starting another compiler
candidate.

Why: this retained rule improves a real shared primitive operation, but the
last promoted compiled scorecard predates it and only six of 34 applications
met the 95%-of-Go target. A current broad measurement is needed to determine
whether any product rows move materially and to select the next wall from
current evidence rather than continuing down the now-small String-validation
tail.

What it entails: rebuild the full selected compiled cohort once, run repeated
verifier-backed processes under the workstation averaging policy, refresh Go
references only where their source/toolchain fingerprints changed, and
reconcile the largest misses by concrete generated-main CPU/allocation owner.
Advance the next candidate only when the same compiler-owned descendant is
material in at least three unlike applications. Keep bytecode measurements in
the same status report, but do not begin WASM work.
