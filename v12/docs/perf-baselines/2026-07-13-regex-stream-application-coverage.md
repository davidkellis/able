# Streaming regex application coverage — 2026-07-13

## Decision

Add `regex_stream_audit` as the 28th verifier-backed application in the
external `coverage` inventory and as the dedicated `regex-stream` suite. Keep
the stable 16-program `generality` scorecard unchanged. This tranche changes
no VM, compiler, runtime, or canonical-stdlib performance code.

The app closes the remaining public regex surface gap. It reads 512 ordinary
words from ENABLE, feeds each word and its trailing newline as separate chunks
to `RegexScanner`, drains currently decidable matches, and finalizes each of
four streams with `flush()`. It uses anchored multiline matching and a named
capture through the ordinary public `Regex.scan` API. Go, Python, and Ruby use
the same incremental line-buffer rule: a record remains pending until its
newline arrives. The shared verifier output is `2048:228:49240`.

This is not a scanner-specific performance target. In particular, it does not
authorize a path for this pattern, corpus, newline chunking, or named capture.
The app exists so future generic NFA work has a public streaming application
alongside ordinary matching and `RegexSet` classification.

## Fresh verifier-backed comparison

All reference and Able processes used CPU 15, three runs, a 45-second
per-process cap, `GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`. Python was
3.14.5 and Ruby was 4.0.5. Every successful Able run passed the sibling Ruby
verifier and produced stdout SHA-256
`dd7801f0b104d8bd47aaef64d685bc65a06263851f12c8df8cf75260d09e717b`.

| Application | Go (s) | Python (s) | Ruby (s) | Compiled Able (s) | Able/Go | Bytecode Able (s) | Able/Python | Able/Ruby | Tree-walker (s) |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Regex Stream Audit | 0.0066 | 0.0202 | 0.0499 | 0.2133 | 32.32x | 6.4167 | 317.66x | 128.59x | 41.8133 |

The tree-walker row is retained as a semantic comparison, not a performance
goal. All three modes completed and verified 3/3 inside the guard.

The exact commands were:

```text
./v12/bench_refresh_go_refs \
  --benchmarks regex_stream_audit --runs 3 --timeout 45 --cpu-affinity 15 \
  --gomaxprocs 1 --output-json /tmp/regex-stream-go-refs.json

./v12/bench_refresh_interpreter_refs \
  --benchmarks regex_stream_audit --runs 3 --timeout 45 --cpu-affinity 15 \
  --python-bin python3.14 --ruby-bin ruby \
  --output-json /tmp/regex-stream-interpreter-refs.json

GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1 ./v12/bench_compare_external \
  --benchmarks regex_stream_audit --modes compiled,bytecode,treewalker \
  --languages go,python,ruby --runs 3 --timeout 45 --cpu-affinity 15 \
  --reference-json /tmp/regex-stream-interpreter-refs.json \
  --go-reference-json /tmp/regex-stream-go-refs.json
```

## Lowering guard and conclusion

`v12/bench_bytecode_audit --suite coverage` passes with 28 applications, 124
lowered functions, and 6,872 bytecode instructions. The stream application
itself lowers to three functions and 257 instructions; it is therefore part
of the normal all-application opcode/lowering guard rather than a side probe.

The new row is a substantial current product miss, but one application cannot
identify a broadly applicable leaf. No profile or source experiment follows
from this coverage tranche. The existing public-regex work already kept only
API-independent NFA traversal and thread-reuse improvements in `able-stdlib`.

## Cross-API profile follow-up

The required Scanner/Suffix/RegexSet/I-Before-E gate is complete and selected
no source candidate. Its compiled profiles repeat the already-addressed
generic tagged-NFA family only within the three regex APIs, while the bytecode
profiles share only VM parents rather than a new concrete descendant. The
record is `2026-07-13-regex-stream-cross-api-profile-gate.md`.

## Next recommendation

Return to feature-led coverage from the active v12 roadmap. A new candidate
must begin with a missing shared semantic boundary and portable
verifier-backed application, because this now-complete regex family cannot
justify scanner-, pattern-, corpus-, or generic-call-path tuning.
