# Compiled `StringBuilder` generality gate — 2026-07-13

## Decision

Keep no compiler, runtime, kernel, or canonical-stdlib performance change.

The three applications do repeat `StringBuilder.push_string` and
`StringBuilder.finish`, but the first general kernel-boundary candidate did
not meet the timing bar. It lowered statically typed
`__able_String_from_builtin` calls directly to a monomorphized `Array u8`
instead of allocating boxed runtime values one byte at a time. That rule is a
valid language-kernel boundary rather than a `StringBuilder` special case, but
it regressed two of the three independent applications. The candidate was
reverted.

## Method

Each generated binary used canonical external `able-stdlib`, `GOMEMLIMIT=1GiB`,
`GOGC=50`, and `GOMAXPROCS=1`. `ABLE_GO_PHASE_PROFILE_DIR` retained only
start/end allocation snapshots and phase stats around user `main`; generated
source/build trees were removed. CPU captures were retained, but their samples
are dominated by Go's allocation-profile snapshot bookkeeping, so allocation
stacks—not those CPU samples—determine this gate.

| Workload | Main phase delta | Shared builder attribution | Other material path |
| --- | ---: | --- | --- |
| `string_builder_small` | 93,786,352 B / 1,941,096 allocs | `push_string`: 774,668 objects (46.60% cumulative); `finish`: 694,259 (41.76%) | String-to-byte conversion and UTF-8 validation. |
| `run_length_encode_small` | 330,158,808 B / 7,870,476 allocs | `push_string`: 1,316,986 (20.39%); `finish`: 824,485 (12.77%) | `String.chars`: 2,400,451 (37.17%) and iterator-next conversion. |
| `byte_histogram_small` | 14,396,528 B / 239,304 allocs | `push_string`: 105,688 (43.94%); `finish`: 84,918 (35.30%) | The same String-to-byte conversion family. |

The retained artifacts in `v12/interpreters/go/.profiles/` are prefixed with:

- `20260713_string_builder_string_builder_small_compiled_`
- `20260713_string_builder_run_length_encode_small_compiled_`
- `20260713_string_builder_byte_histogram_small_compiled_`

## Candidate timing gate

The candidate applied only when the input was statically primitive `String`
and the expected result was statically `Array u8`; all dynamic kernel calls
kept the existing helper. Three-run compiled timings and stdout hashes were:

| Workload | Baseline | Candidate | Result |
| --- | ---: | ---: | --- |
| `string_builder_small` | 0.2033 s | 0.2067 s | 1.7% slower |
| `run_length_encode_small` | 0.4933 s | 0.4867 s | 1.3% faster |
| `byte_histogram_small` | 0.0967 s | 0.1067 s | 10.3% slower |

Every baseline/candidate stdout hash matched, and each lane retained the same
average GC count. The mixed result is insufficient for a shared compiler
optimization, so the temporary lowering and its tests were removed.

## Next recommendation

Profile and inspect the primitive `String.from_bytes` boundary—especially its
`Array u8` to canonical `String` result path—in these same three applications.
`StringBuilder.finish` is material in every profile, while a compiler rule for
the named `StringBuilder` type would violate the nominal-lowering policy. A
candidate is justified only if the same primitive conversion leaf recurs and
can preserve `Result String` invalid-UTF-8 behaviour for every caller.
