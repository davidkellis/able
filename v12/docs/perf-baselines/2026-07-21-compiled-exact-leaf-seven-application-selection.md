# Compiled exact-leaf seven-application selection

Date: 2026-07-21

## Decision

Keep no compiler, generated-runtime, bridge, stdlib, application, reference,
language, or WASM performance change. Seven unlike current generated programs
do not share one exact compiler-controlled CPU leaf that is material in at
least three applications.

The repeated Go GC leaves in four profiles remain a parent over incompatible
allocation mechanisms. The only newly isolated compiler-owned opportunity is
the nil `*__ableControl` result/check around direct recursive Fib calls, and it
is CPU-material in only one application. It therefore does not authorize an
infallible-function ABI prototype yet.

## Protocol

One current compiler CLI was frozen at SHA-256
`87535396e8c8a471f3dd84d44ce7678554426f1ac84d5e730d729349f83fa913`.
It emitted normal generated Go against canonical `../able-stdlib` with
`ABLE_SOURCE_ROOT_ONLY=1`. Each generated module supplied:

- two independent main-only CPU-profile processes;
- two independent lightweight exact main-allocation processes;
- the public scorecard arguments, working directory, input, and Ruby verifier;
  and
- `GOMAXPROCS=1`, `GOGC=50`, a 1-GiB Go memory limit, and a 59-second cap.

Every one of the 28 measured outputs passed its verifier. Arithmetic means are
reported because the host is a working workstation. CPU profiles were merged
only within one application; observer processes were not timing evidence.

The larger generated modules did not produce a binary during the first
bounded wrapper invocation, but left complete generated modules. A bounded
`go build` over those exact preserved modules completed the frozen binaries;
no source, build flag, or semantic option changed.

## Current evidence

| Application | Family | Profile-process mean | Merged CPU | Main bytes / objects | Exact leading owner |
| --- | --- | ---: | ---: | ---: | --- |
| `fib` | primitive recursion | 3.715 s | 7.28 s | 144 / 6 | generated `fib` 99.86% flat |
| `sudoku_masks` | bitmask search | 2.065 s | 3.98 s | 156,370,400 / 7,802,591 | `find_best_empty`; signed div/mul; bit count |
| `base64` | byte transformation | 2.830 s | 5.38 s | 2,201,553,544 / 129 | Go Base64 encode/decode and MD5 |
| `pidigits` | arbitrary-width integer | 1.370 s | 2.61 s | 298,859,228 / 24,231 | `math/big` multiply/shift/add/subtract |
| `k_nucleotide` | text/map | 2.940 s | 5.74 s | 1,224,369,032 / 27,734,575 | primitive HashMap equality/hash and allocation |
| `tapelang_alphabet` | state-machine dispatch | 3.740 s | 7.33 s | 282,552 / 4,274 | generated `execute`, `Tape.inc`, and `Tape.get` |
| `mandelbrot` | float numeric control | 0.130 s | 0.12 s | 86,328 / 66 | generated `pixel_byte`; sample is selection-only |

Mandelbrot's 120 ms profile is too small for a percentage-level admission
claim. It is retained as a verified absence/control row, not used to promote a
leaf.

## Exact-owner reconciliation

### Generated control results

Fib's direct recursive body owns 99.86% flat CPU. Its two post-call
`control != nil` checks account for 1.32 seconds, or 18.13% of the merged
profile. The function allocates essentially nothing, so this is a real direct
ABI/code-shape observation.

The same syntax is not material elsewhere:

- TapeLang samples 10 ms (0.14%) on the control check after its hot
  `Tape.inc` call;
- Sudoku samples 10 ms (0.25%) on the corresponding check after `bit_count`;
- Mandelbrot samples zero CPU on the check after `pixel_byte`; and
- Base64, Pidigits, and K-Nucleotide end in distinct host, wide-integer, and
  bridge/map operations rather than a shared direct-control check.

Removing or changing the generated control ABI from this evidence would be a
Fib optimization. Several other bodies also return a nil control on the
verified path, but path outcome is not a proof that the function is
semantically infallible for every input.

### Allocation and Go GC

`runtime.tryDeferToSpanScan` is sampled in Sudoku (3.27%), Base64 (5.58%),
Pidigits (7.66%), and K-Nucleotide (11.85%). This exact runtime leaf does not
identify one compiler transformation:

- Sudoku creates short candidate Arrays during the recursive search;
- Base64 allocates only 129 main objects but moves 2.20 GiB through large byte
  buffers;
- Pidigits allocates `math/big` values and limbs; and
- K-Nucleotide creates millions of map keys/values and conversion objects.

Fib has six main allocations and TapeLang has only 4,274. Their profiles have
no material GC leaf. The new rows therefore reinforce the existing generated-
allocation-shape decision: common GC machinery is an aggregate cost, not a
shared removable allocation shape.

### Primitive, Array, and host kernels

The remaining exact owners also split:

- TapeLang spends material CPU in explicit Array bounds/read/write work and a
  checked `i32` update;
- Sudoku spends material CPU in signed division, checked multiplication,
  bit counting, and its search body;
- Base64 is already in Go's `encoding/base64` and `crypto/md5` kernels;
- Pidigits is already in `math/big` kernels;
- K-Nucleotide is a HashMap equality/hash/boxing path; and
- Mandelbrot is direct generated float arithmetic.

No exact helper occurs as a material compiler-controlled leaf in three rows.
Combining these descendants under “primitive work,” “Array work,” or “host
work” would erase the mechanism distinction required by the project gate.

## Verification and cleanup

The focused recursion, control-to-error, raised-control, dynamic-callback, and
static-main compiler guards pass in 5.564 seconds. `git diff --check` passes.
No production or canonical-stdlib source changed. The frozen compiler,
generated modules, binaries, profiles, allocation reports, and captured
outputs are temporary and removed after this record is written.

## Artifact identity

| Application | Binary SHA-256 |
| --- | --- |
| `fib` | `142e604354ac5e6add24f55e30dcf516222db150c89fa44805747eb218382025` |
| `sudoku_masks` | `904ec3862d85631046fb5896bd192bf1c1866a2d60ada757dd0ab5f9bff76ba1` |
| `base64` | `60bace6a994e591c2c82b4e7b1d828cd46b9196860d39581cbe5303aed0be11e` |
| `pidigits` | `8b631face812f6d8fa785e301fd1fda08fc5747039cf0c57e5cba18a2eb7922f` |
| `k_nucleotide` | `ccc00582aaff63da3959afc82eb4830df586fa1076d24f0a6d5be03c88ee0dbc` |
| `tapelang_alphabet` | `47678f79115e88a0ffd835b745118e8c57c7d242d5f87bb12122daecceca59eb` |
| `mandelbrot` | `e227b944cf1feb825673985d6644d909dce543eac3fac31d645e5c18976bf7c4` |

## Next recommendation

Run a report-only compiled effect/control-ABI reach census before writing
another compiler candidate.

Why: the exact-leaf sweep found one genuinely compiler-owned cost—nil control
result propagation around pure Fib recursion—but only one material workload.
An effect fixed point can determine whether direct compiled functions that
cannot raise or return runtime errors are common and hot enough to justify a
general internal ABI, without weakening error semantics or selecting Fib by
name.

What it entails: classify direct compiled functions through a call-graph fixed
point as definitely control-free or potentially fallible; preserve the current
`*__ableControl` wrapper at dynamic, interface, callback, exported, and host
boundaries; combine the classification with main-only profile/call-site reach
across Fib, numeric, search, text, iterator, and concurrency applications; and
require material direct control-free calls in at least three unlike families.
Only then prototype an internal value-only signature and repeated source-
matched A/B cohorts. Include raise/rescue/or-else, overflow, division-by-zero,
dynamic-call, recursion, and nullable/error guards. Do not add function-name,
nominal-type, benchmark, stdlib-container, GC, or WASM special cases.
