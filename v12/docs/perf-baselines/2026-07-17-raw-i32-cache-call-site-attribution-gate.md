# Raw-i32 cache call-site attribution gate

Date: 2026-07-17

## Decision

Complete the call-site attribution gate and retain no runtime candidate. A
shared site does exist: `stackRawI32Value(...)` owns every cache request above
65,535 in Rational Series, Array Slice Window, and the Array Map fixture.
However, that site is not an avoidable boxing boundary. For values inside the
cache it returns an existing stable `runtime.Value` and allocates nothing; the
boxed operand stack requires that interface value.

Changing the site alone would either replace a cheap array lookup with a
mutable pointer carrier or require another whole operand-stack representation.
The former changes dispatch and snapshot ownership and repeats the raw-cell
shape already rejected by broad guards. The latter repeats the class of
per-operation tag/reconciliation costs measured by the rejected f64 operand
lane. Neither is a bounded candidate justified by these counters.

All temporary counters, output plumbing, runner code, binaries, and raw results
were removed. The cache, operand stack, compiler, stdlib, fixtures, and
language remain unchanged.

## Protocol

Temporary site IDs wrapped all 19 production calls to
`bytecodeRawI32SlotCachedValue(...)`. Atomic counters recorded total requests,
requests above 65,535, 131,071, 196,607, and the current 262,143 cache maximum,
plus each site's maximum value. Collection reset immediately before `main`,
so loading and initialization did not contaminate the application phase. The
wrapper then called the unchanged cache function and returned its exact value.

The gate ran Rational Series, Array Slice Window, and Array Map twice. Both
counter JSON and stdout were byte-identical between processes for each
workload. Both runs of each external program passed its Ruby verifier; both
Array Map processes exited successfully. One K-Nucleotide process completed
under the 55-second cap and passed its Ruby verifier. Its earlier value census
was deterministic and its 25.2 million calls made a second diagnostic process
unnecessary.

## Attribution result

The high-value counts at the shared `stack_value` site were:

| Workload | All site requests | `stack_value` requests | >65,535 | >131,071 | >196,607 | Site maximum |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Rational Series | 855,324 | 455,322 | 115,884 | 76,447 | 37,010 | 262,143 |
| Array Slice Window | 967,121 | 330,599 | 13,560 | 9,040 | 4,520 | 262,138 |
| Array Map fixture | 348,026 | 120,000 | 41,885 | 4,536 | 0 | 140,073 |

No other site in those three programs requested a value above 65,535. Thus
`stack_value` owns 100% of every high-value band in three structurally unlike
programs and passes the stated cross-program attribution threshold.

K-Nucleotide demonstrates why this is not a universal cache shortcut:

| K-Nucleotide site | Requests | >65,535 | Share of >65,535 | >196,607 | Share of >196,607 | >262,143 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| `owned_slot_store` | 14,880,248 | 6,094,848 | 54.72% | 2,031,616 | 40.82% | 0 |
| `stack_value` | 8,439,640 | 3,145,728 | 28.24% | 1,048,576 | 21.07% | 0 |
| `call_frame_copy` | 1,896,987 | 1,896,987 | 17.03% | 1,896,987 | 38.11% | 1,896,987 |

`call_frame_copy` owns all K-Nucleotide misses above the current cache, while
owned typed-slot storage leads its in-cache high-value traffic. A cache-wide or
single-site representation change would therefore move costs differently
across programs rather than remove one common wall.

## Restoration and verification

After diagnostic removal:

- no site-ID, counter, environment-variable, JSON-output, or runner reference
  remains in production source;
- focused raw integer, slot, frame, Array index, and comparison tests pass;
- focused CLI execution/stat-output tests pass; and
- the benchmark-selection contract check passes.

The diagnostic counts are selection evidence only. Atomic instrumentation was
never used as timing evidence.

## Next recommendation

Attribute `stackRawIntegerValue(...)` traffic one level deeper to the generic
producer families that feed it: binary replacement, cast, match, iterator,
index, slot load, and return.

Why: the current gate finds the transport boundary but also proves that the
transport operation itself is already allocation-free. A useful optimization
must avoid a raw-integer-to-boxed-stack-to-raw-integer round trip across a
multi-operation typed region, fused producer/consumer pair, or direct typed
store. Producer attribution can distinguish that opportunity from mandatory
dynamic, aggregate, call, and return escapes.

What it entails: add temporary producer IDs at the small centralized set of
`appendRawIntegerStack(...)` and `replaceTop2RawIntegerUnchecked(...)` callers;
run the same three primary workloads plus K-Nucleotide and unrelated integer
recursion/text controls; and require one semantic producer family to dominate
at least three unlike programs. Build only a generic typed-region or
producer/consumer fusion if that rule passes. Keep the boxed stack and stable
escape semantics, reject split traffic, and remove all diagnostics afterward.
