# Bytecode scalar-character box-reuse gate — 2026-07-17

## Decision

Keep the generic primitive-boundary change. `__able_char_to_codepoint` now
returns the existing immutable boxed `i32` value for a codepoint when one is
available through the VM's shared integer cache, and retains the general
bounded-cache/fallback behavior for the rest of Unicode. The change does not
inspect a benchmark, regex, string, automaton, or nominal-container identity.

Also keep the new verifier-backed `unicode_scalar_pipeline` application in the
portable coverage catalog. Able, Go, Python, and Ruby construct the same mixed
ASCII/Latin/Greek/CJK/emoji text, traverse scalar values, round-trip every
codepoint, and emit the same weighted checksum and UTF-8 byte length.

## Profile admission

Fresh one-process profiles used the current bytecode test binary,
`GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`, CPU 0, canonical external
`able-stdlib`, a warmed `main`, and one bounded measured call.

| Workload | Baseline | Exact `initStringHostBuiltins.func8` allocation |
| --- | ---: | ---: |
| Run-length encode | 2.709 s; 4,150,667 objects | 2,053,555 objects (44.51%); 94 MiB (47.85%) |
| Automata DFA | 3.576 s; 1,763,565 objects | 983,084 objects (41.47%); 45 MiB (33.07%) |
| Regex Stream | prior bounded profile | 797,389 objects (25.04%); 36.50 MiB (19.20%) |
| Levenshtein | 18.32 ms; 16,166 objects | no material sample |

The same leaf is therefore material in three unlike algorithms: ordinary
run-length text processing, general automata/DFA construction and matching,
and streaming captured regex. Levenshtein is a useful negative control.

## Candidate

Both value and pointer `char` arguments previously constructed a fresh
`runtime.IntegerValue{i32}` and converted it to the `runtime.Value` interface
on every call. They now call the already-general
`boxedOrSmallIntegerValue(IntegerI32, codepoint)` primitive helper. Existing
fixed and extended `i32` caches cover the benchmark corpus, including its emoji;
the existing bounded dynamic cache handles larger valid scalars. Integer values
are immutable at this boundary, so reuse does not change Able identity or
mutation semantics.

## Repeated runtime gate

Each row uses independent processes. Initial baseline-first triples were
extended with candidate-first interleaved samples wherever load order appeared
material. The table reports all samples, while the text below records the
order-balanced subsets used to resolve volatility.

| Workload | Samples/side | Baseline mean | Candidate mean | Timing | Allocation result |
| --- | ---: | ---: | ---: | ---: | ---: |
| Run-length encode | 3 | 2.8361 s | 2.5789 s | 9.07% faster | 4,150,597 -> 2,230,689 objects (-46.3%) |
| Automata DFA | 8 | 3.5238 s | 3.5005 s | 0.66% faster | 1,763,508 -> 687,687 objects (-61.0%) |
| Levenshtein control | 3 | 19.186 ms | 18.548 ms | 3.33% faster | about 16,116 -> 9,394 objects (-41.7%) |
| Unicode Scalar Pipeline | 3 | 6.8347 s | 6.3666 s | 6.85% faster | 14,039,344 -> 8,730,883 objects (-37.8%) |
| Regex Stream | 3 | 678.253 ms | 651.984 ms | 3.87% faster | about 192,172 -> 135,412 objects (-29.5%) |
| Reverse Complement guard | 3 | 1.111 ms | 1.081 ms | 2.68% faster | effectively identical |
| Base64 guard | 8 | 70.727 ms | 71.056 ms | 0.47% slower; neutral | effectively identical |
| Iterator Collect guard | 8 | 427.293 ms | 434.281 ms | 1.64% slower overall | identical |
| Numeric Array Map guard | 3 | 74.619 ms | 73.789 ms | 1.11% faster | identical |
| String Split/Join guard | 3 | 1.0926 s | 1.0165 s | 6.96% faster | effectively identical |

The first Automata triple looked 2.3% slower, but five candidate-first pairs
were 2.35% faster and made the eight-sample mean positive. Iterator Collect's
first baseline-first triple looked 4.6% slower; its five candidate-first pairs
were 436.858 ms versus 436.954 ms, a 0.02% improvement. Base64's corresponding
five-pair subset was 0.28% slower. These order-balanced subsets classify both
unrelated guards as neutral rather than stable regressions.

A post-change Run-length profile no longer attributes any material allocation
to the char-to-codepoint closure. Its measured call used 66.37 MB and 2,230,762
objects; the next dominant shared allocation is now equality call-argument
ownership (`ensureMutableCallArgs`, 42.22% flat objects) beneath cached equality
dispatch.

## New external application status

Five verifier-backed source runs average 0.0100 s for Go, 0.2649 s for Python,
and 0.3726 s for Ruby. Five current Able process runs average 0.2340 s compiled
and 6.9060 s bytecode. Thus this new row is 23.40x Go in compiled mode, and
26.07x Python / 18.53x Ruby in bytecode mode. The retained optimization is a
real improvement, but this application remains far from both 95% product goals.
Its fresh five-run rows satisfy the bounded protocol, so it is explicitly added
to both modes of the reviewed strict selection manifest. The next complete
scorecard refresh must incorporate it before replacing the promoted scoreboard.

## Verification

- Focused char/String-host tests pass, including ASCII and non-ASCII
  char-to-codepoint round trips.
- Able compiled and bytecode smoke runs match the verifier.
- Go, Python, and Ruby sources all match the same verifier.
- `just bench-catalog-check` reports 35 portable applications, 36 canonical
  sources, one diagnostic source, 78 local fixtures, and 113 combined programs;
  all five feature-coverage protocol tests pass.
- The reviewed selection contains 63 rows: all 35 compiled applications and
  28 verified bytecode applications, including Unicode Scalar Pipeline.

## Next recommendation

Profile the post-change equality argument-ownership wall across Run-length,
Automata DFA, Unicode Scalar Pipeline, and one non-text equality-heavy program.
Admit a candidate only if `ensureMutableCallArgs` beneath cached equality
dispatch is the same material leaf in at least three unlike applications.

Why: the retained change removed the admitted scalar box wall, and Run-length's
new profile now assigns 42.22% of allocated objects to generic equality
call-argument ownership. The work entails fresh bounded post-change profiles,
an audit of whether primitive immutable equality arguments can safely remain
borrowed, a focused semantic test for mutable/user-defined equality, and
repeated scalar plus iterator/byte/numeric guards. This targets a larger shared
VM cost without introducing char-, regex-, or benchmark-specific lowering.
WASM remains deferred.
