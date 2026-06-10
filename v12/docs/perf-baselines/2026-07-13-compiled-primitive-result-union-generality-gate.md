# Compiled primitive-result union generality gate — 2026-07-13

## Decision

Keep no compiler, runtime, or canonical-stdlib change. The residual
pointer-backed `Utf8DecodeResult` union allocation is a material String
decoder cost, but it does not repeat across the independent numeric and map
controls. A general union-representation rewrite would therefore optimize the
current String traversal shape without evidence that it improves application
programs broadly.

The language specification permits implementations to choose a union's
internal representation, but representation freedom does not justify an
unproven global change. There is no candidate to retain in this tranche.

## Bounded profiles

Fresh compiled generated-main profiles used the external canonical
`able-stdlib`, `GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`. They use one
control from each required shape:

| Workload | Shape | Main bytes / allocations / GCs |
| --- | --- | ---: |
| `run_length_encode_small` | String character/result-union traversal | 103,799,128 / 3,213,294 / 6 |
| `array_map_i32_small` | primitive numeric Array work | 2,576,104 / 112 / 0 |
| `hashmap_i32_small` | generic map lookup returning `?i32` | 2,956,808 / 12,351 / 0 |

The String profile still attributes about 722,243 allocation objects (31.7%
of its allocation profile) to the canonical `utf8_decode` function, which
returns the pointer-backed `Utf8DecodeResult` union payload. In contrast,
numeric array-map has no material application allocation attribution after
the profile-writer frames are excluded. The map profile attributes only about
751 objects to `__able_nullable_i32_from_value`; its larger visible work is
the separate primitive scalar `bridge.ToInt` conversion through the generic
`Map.get`/`HashMap.raw_get` path.

The compact artifacts are retained under
`v12/interpreters/go/.profiles/20260713_primitive_result_union_`. The phase
stats are based on `runtime.MemStats` snapshots immediately around the
registered main phase, so they are the allocation decision metric; allocation
profile writer frames are not treated as application work.

## Why no candidate is justified

The union representation appears in one of three independent workload
families, while the numeric control is essentially allocation-free and the
map option carrier is a small part of its profile. Replacing all generated
unions with a new value/pointer scheme would risk broad semantic and escape
behavior for no demonstrated cross-program return. It would violate the
project's performance rule against encoding a narrow benchmark wall as a
global optimization.

## Verification

- All three compiled profile runs completed successfully under the normal
  CPU, timeout, and memory guardrails.
- No source behavior changed; no compiler, interpreter, or stdlib test needs
  a new semantic expectation.
- `git diff --check` passes.

The repository-wide `./run_all_tests.sh` remains blocked before Go tests by
the existing untracked `exec/12_09_nested_spawn_native_context` fixture missing
from the already-modified exec coverage index. This tranche leaves that
fixture and index untouched.

## Next recommendation

Profile the shared primitive `bridge.ToInt` conversion through the generic
`Map` interface across `HashMap`, `TreeMap`, and `PersistentMap`, with a
numeric Array control. It is the non-union map cost visible here, and any
follow-up must improve one primitive scalar conversion in the common nominal
translation boundary—not a named-map implementation. Only retain a candidate
if the same conversion is material and timing-positive in multiple map
implementations without harming the numeric control.
