# Post-nullable compiled byte-output reconciliation

## Decision

Reconcile `compiled-byte-output` as causally current and retain no production
change.

FASTA Generation and Reverse Complement already keep their material byte work
in native generated `[]uint8` storage. Their generated application bodies
contain zero nullable-`u8` carriers, zero nullable-byte helper calls, and zero
runtime-value byte conversions. Reads are direct bounds checks plus
`Elements[index]`; writes are direct assignments or Go `append`.

Fresh Mandelbrot evidence supplies the required unlike third application for
the shared bulk-output path. It has the same native Array and borrowed
`[]uint8` output representation, but its material owner remains the direct
numeric pixel loop. The only shared runtime conversion is the explicit
one-call host result boundary after `io_write`, not a per-byte or nullable
carrier cost.

## Strict boundary and execution gate

All three applications were rebuilt from the retained compiler with
`--no-fallbacks`. Each exact binary passed its public Ruby verifier.
Mandelbrot is a breadth control rather than a row in the two-row
`compiled-byte-output` closure.

| Application | Role | Packages | Interpreter dependency | Verified smoke |
| --- | --- | ---: | --- | --- |
| FASTA Generation | closure row | 96 | absent | 1/1 |
| Reverse Complement | closure row | 96 | absent | 1/1 |
| Mandelbrot | unlike output control | 96 | absent | 1/1 |

Smoke durations were 0.02, 0.01, and 0.05 seconds respectively. They are
execution checks, not timing evidence. The authoritative scorecard retains
five verifier-backed Able and Go processes per row.

## Generated native-carrier audit

The complete generated modules each contain 11 textual
`__able_nullable[uint8]` references and six nullable-byte helper-name
references. Those references belong to shared support:

- generic runtime-value conversion helpers;
- text-decoding `read_byte` and `slice_bytes` functions; and
- `slice_bytes_from_offset`, which is reached by `write_all` only after a
  successful positive partial write.

The normal bulk-output path begins with offset zero and passes the original
Array to `io_write`. The generated host call borrows `bytes.Elements` directly
as `[]uint8`. Go's file write reports an error for a short write, so the
nullable-byte suffix loop is not part of a successful ordinary process.

The exact application-body census is:

| Application | Nullable `u8` sites | Runtime byte conversions | Direct `.Elements` references | Application-body SHA-256 |
| --- | ---: | ---: | ---: | --- |
| FASTA Generation | 0 | 0 | 19 | `24bada0eca0ba90528d60b839e5f2a15206180c4c81c2e5f7587e81199f50e9e` |
| Reverse Complement | 0 | 0 | 35 | `790e1de641f6dbc753ffce8ad9b8b92c0231485d9d3172647ec848f5903e94d0` |
| Mandelbrot | 0 | 0 | 7 | `5a6f8736c3173627567194afb21cc6c83431f77e93d2558ae0616240be423f7a` |

FASTA pre-sizes its output and appends native bytes. Its material
`append_random` loop calls direct generated integer and byte functions, then
appends the returned `uint8`.

Reverse Complement receives `os.ReadFile`'s `[]uint8` result through the
generated native fast path without per-byte conversion. Its transform reads
the data, sequence, and complement table directly and appends native bytes to
pre-sized output. The sequence buffer's geometric growth is real but remains
a Reverse-specific descendant rather than a three-application candidate.

Mandelbrot's `render` builds the native byte Array directly from `pixel_byte`
results. Its application body contains no nullable byte or runtime-value
conversion.

## Output-boundary audit

All three generated modules have the same general output shape:

1. `write_all` accepts `*__able_array_u8`.
2. The first write receives the caller's original Array.
3. `io_write` passes `able_borrowed_bytes(bytes.Elements)` to the synchronous
   host function.
4. The host result crosses the required `IOError | i32` boundary once.
5. A copied suffix is created only after a positive partial write.

This preserves the retained general `write_all` optimization. It also rejects
the earlier byte-array deoptimization hypothesis for compiled code: there is
no dynamic Array and no per-byte runtime representation at output.

The one-call handle/result bridge repeats in three programs, but it is not a
material shared leaf. Retained profiles place FASTA in direct random-sequence
arithmetic, Reverse Complement in transform/copy/backing/GC work, and
Mandelbrot at 95.42% cumulative CPU in its generated `pixel_byte` loop.
Removing a once-per-process boundary cannot explain the current 2.06x-3.75x
application gaps.

## Selective profile and candidate gate

Fresh profiling and A/B implementation were not admitted:

- the changed nullable-byte carrier has zero application-body reach in all
  three applications;
- material Arrays already use native `[]uint8` backing;
- input and output byte slices use generated native host fast paths;
- the conditional partial-write suffix is not ordinary successful-process
  work;
- the only shared required host conversion executes once; and
- retained material descendants split among FASTA arithmetic, Reverse
  transform/backing work, and Mandelbrot numeric rendering.

No exact changed residual is material in three unlike applications. Repeating
profiles would not identify an implementable general compiler or runtime
rule.

## Current row state

The current five-process scorecard means are:

| Application | Able compiled | Go | Able / Go |
| --- | ---: | ---: | ---: |
| FASTA Generation | 0.0480 s | 0.0163 s | 2.9448x |
| Reverse Complement | 0.0600 s | 0.0160 s | 3.7500x |
| Mandelbrot control | 0.1100 s | 0.0533 s | 2.0638x |

These remain product misses. The direct generated carriers show that
nullable-byte boxing, dynamic byte Arrays, and interpreter fallback are not
their residual cause.

## Scope and cleanup

No compiler, generated runtime, runtime package, interpreter, bytecode VM,
canonical stdlib, benchmark, language, dependency, or WASM source changed.
No Array, I/O, bioinformatics, image, named-container, non-primitive nominal,
or benchmark-specific rule was introduced.

`go test ./cmd/ablec` passed in 5.888 seconds. The machine-readable record is
`2026-07-30-post-nullable-compiled-byte-output-reconciliation.json`.

After retaining this evidence, the exact 418 MiB disk-backed compiler,
generated-module, binary, smoke-output, audit, and Go-cache workspace was
removed. No matching tranche artifact remains in `/var/tmp` or `/tmp`.

## Next

Reconcile `compiled-sudoku-quotient` against the primitive nullable carrier.

Why: it is the smallest remaining invalidated compiler closure. Sudoku's
signed Euclidean division was previously material in one application, while
the newly current wide-numeric evidence distinguishes Rational's different
two-word `DivMod` representation and operation mix.

What it entails: strictly rebuild Sudoku Masks, confirm its dependency graph
remains interpreter-free, and trace nullable-`i32` plus quotient/remainder
paths through `square_index` and the search loop. Reuse the retained two-profile
merge and unlike controls. Admit no candidate unless the exact same primitive
quotient leaf is material in three unlike current applications.

Why it matters: Euclidean division is a legitimate primitive operation and
can be expensive, but a one-algorithm shortcut would not be a general
lowering improvement. This review can close the compiler drift without
turning Sudoku's structure into a compiler rule.
