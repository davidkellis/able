# Bytecode i32 Proof Profile Selection Gate

**Date:** 2026-07-13
**Decision:** no new primitive-frame candidate

## Method

MatrixMultiply and Mandelbrot were selected because each includes
concrete-`i32` function parameters admitted by the retained proof metadata,
while their primary work differs: matrix/Array construction and multiplication
versus per-pixel float iteration and byte output.

Normal bytecode CLI controls used CPU 15, `GOMEMLIMIT=1GiB`, `GOGC=50`,
`GOMAXPROCS=1`, and a 45-second cap. Both passed their canonical verifiers:
MatrixMultiply in 4.48 seconds and Mandelbrot in 7.00 seconds.

The warmed runtime benchmark then loaded and warmed each program once before
profiling one `main()` call under the same one-core memory/GC limits:

| Application | Runtime result | CPU samples |
| --- | ---: | ---: |
| MatrixMultiply | 4.596s/op; 308,626,200 B/op; 14,032,456 allocs/op | 4.56s |
| Mandelbrot | 7.221s/op; 618,923,664 B/op; 76,303,152 allocs/op | 7.18s |

Profiles are retained at:

- `v12/interpreters/go/.profiles/20260713_i32_frame_gate_matrixmultiply.cpu.pprof`
- `v12/interpreters/go/.profiles/20260713_i32_frame_gate_mandelbrot.cpu.pprof`

## Attribution

MatrixMultiply is led by `tryExecF64DotLoop` (27.19% cumulative), canonical
Array slot/member work (28.07%), and Array pushes (25.22%). Mandelbrot is led
by `execJumpIfFloatMulAddMulCompareConstFalse` (24.51%) and its normalized
float-slot update chain (about 22%). These are distinct descendants; the
Mandelbrot float lane is also an already rejected performance direction.

`appendSlotStackValueChecked` appears in both profiles (7.02% MatrixMultiply;
6.41% Mandelbrot), but line attribution places its samples on the ordinary
non-i32 `runtime.Value` snapshot branch. The i32-specific store helper is only
1.97% and 2.65% cumulative respectively. Thus neither shared helper supports
the typechecker-proven i32 frame representation, and the snapshot observation
does not reopen previously rejected raw-cell variants.

## Decision

Keep no VM/compiler/runtime/stdlib change. The first typed-frame experiment
was already reverted for a broad regression, and this paired profile audit
finds no second, concrete i32 descendant to justify another representation
design or a typed call/return ABI. Resume feature-driven benchmark coverage
instead.
