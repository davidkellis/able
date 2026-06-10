# Regex suffix-audit baseline — 2026-07-12

`regex-suffix-audit` is a verifier-backed application benchmark in the sibling
`../benchmarks` corpus. It reads the first 16,384 lines of the checked-in
ENABLE word list, builds one anchored public `RegexBuilder` expression for
lowercase words ending in `ing`, `tion`, or `ment`, reads the named suffix
capture, and repeats the audit four times. Every implementation must print:

```
65536:5356:925804
```

The workload exercises normal file input, text values, anchors, character
classes, alternation, repetition, captures, and aggregation. It is not a
request for a word-list, suffix, regex, or container-specific fast path.

## Fresh references

Three verifier-checked process runs under the one-core lane produced:

| Implementation | Average real time |
| --- | ---: |
| Go 1.26.4 | 0.0418s |
| Python 3.14.5 | 0.0409s |
| Ruby 4.0.5 | 0.0879s |

Commands:

```text
v12/bench_refresh_go_refs --benchmarks regex_suffix_audit --runs 3 --timeout 45
v12/bench_refresh_interpreter_refs --suite regex-text --runs 3 --timeout 45
```

## Able baseline

With `GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, canonical external stdlib,
the benchmark-root working directory, and source-root-only discovery:

| Mode | Result |
| --- | --- |
| compiled | verifier passed, 4.8800s real, 4.7500s user, 0.0500s sys, 75 GCs |
| tree-walker | exceeded the 90s application guard |
| bytecode | exceeded the 90s application guard |

The compiled result is about 0.9% of the Go-reference speed in this initial
single-run lane. It is a correctness-validated baseline, not a stable
microbenchmark claim and not an optimization target by itself.

## Correctness repairs exposed by the workload

The workload required no regex-specific runtime or compiler path. It did
expose and close three general defects:

- embedded kernel caches now refresh when their contents change without a
  version change;
- type declaration exports hydrate forward struct references before another
  package accesses their fields;
- standalone compiled binaries register package-qualified struct definitions
  with the compiler bridge, preserving nominal identity when packages reuse a
  short struct name such as `Span`.

## Next evidence gate

Capture bounded profiles for this workload, `regex_is_match_small`, and a
non-regex text workload. Only pursue a candidate if the same material
stdlib/VM/compiler leaf repeats; otherwise record the difference and avoid
benchmark-shaped optimization.
