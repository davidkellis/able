# Primitive Conversion Boundary Audit

## Decision

Keep the opt-in fixed-pointer execution-context ABI unchanged and make no
primitive conversion change. The repeated `__able_int64_from_value` samples in
the prior file-backed profile set are not a general numeric or collection
conversion cost.

## Static call-path audit

Generated source has two call sites outside the helper definition in the
string renderer and twelve in the concurrency renderer:

- `String_to_builtin`: validates and converts each runtime byte-array element
  while forming a Go string.
- `char_from_codepoint`: validates the numeric code point.
- Channel and mutex helpers: decode opaque runtime handle values.

No Array, HashMap, numeric, nominal-coercion, or ordinary compiled arithmetic
renderer calls this helper. It is therefore not the raw numeric extraction
path for compiled primitive arrays or arithmetic.

## Non-file control profiles

To break the shared file-ingestion confounder, current
`-experimental-execution-context` binaries were built and output-checked under
`GOMAXPROCS=1`, `GOMEMLIMIT=1GiB`, and `GOGC=50`:

| Workload | Program shape | Output | CPU-profile launches | Result |
| --- | --- | --- | ---: | --- |
| `array_map_i32_small` | in-memory primitive Array construction, map, and sum | `1097192358` | 100 | no `__able_int64_from_value` samples |
| `linked_list_iterator_collect_i64_small` | in-memory nominal collection plus lazy map/filter/collect/reduce | `382455000` | 50 | no `__able_int64_from_value` samples |

Merged normal profiles are retained as:

- `.profiles/20260710_fixed_context_array_map_i32_main_collector_free.cpu.pprof`
- `.profiles/20260710_fixed_context_linked_list_iterator_collect_i64_main_collector_free.cpu.pprof`

The Array-map profile is dominated by its arithmetic/map operation and has a
small sample set because each process is very short. The LinkedList profile is
dominated by iterator dispatch, channels, and allocation. Neither contains the
audited helper, which is the relevant negative result.

Together with the file-backed profiles, this rules out a generic
`__able_int64_from_value` optimization: its apparent recurrence came from
shared text input and concurrency handle decoding, not every program's
primitive numeric work. No compiler, runtime, benchmark, or `able-stdlib`
source changed.

## Next recommendation

Return to the retained fixed-context candidate's broader application scorecard
and compare its ordinary process wall and generated-main profiles against the
default compiler on Word-Frequency, Document-Audit, Lexical-Rollup,
Channel-Rollup, and non-file controls. Why: the ABI candidate has passed its
initial three controls and now has no eligible residual micro-optimization;
the next decision is whether its benefit is broad enough to justify moving
toward default enablement. The work entails matched fresh binaries, output
checks, repeated wall runs, and default-versus-opt-in profiles. Do not enable
the option by default or change dynamic compatibility boundaries without that
broad comparison.
