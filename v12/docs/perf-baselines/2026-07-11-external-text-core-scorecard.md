# External Text and Byte-Core Scorecard

This is a bounded scorecard tranche, not a runtime, compiler, or standard
library change. It refreshes three independently shaped text/byte programs
before selecting another performance candidate.

## Method

`v12/bench_compare_external` ran `i_before_e`, `base64`, and `json` in
compiled and bytecode modes, three times each, against the checked-in
`../benchmarks/results.json` Go, Ruby, and Python references. The Able runs
used the canonical stdlib at `/home/david/sync/projects/able-stdlib/src`, CPU
affinity `2`, `GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and a 45-second
per-run cap. Every requested mode completed all three runs without a timeout
or launch failure.

The retained machine-readable and Markdown artifacts are
`v12/tmp/scorecard-text-core/text-core.{json,md}`. The wrapper labels the
report `core` because explicit benchmark selection leaves the preset label
unchanged; the artifact's benchmark list is the authoritative scope.

## Results

| Benchmark | Mode | Able (s) | Relevant reference (s) | Able/reference |
| --- | --- | ---: | ---: | ---: |
| I-Before-E | compiled | 0.1500 | Go 0.0500 | 3.00x |
| I-Before-E | bytecode | 0.5667 | Ruby 0.1000; Python 0.1300 | 5.67x; 4.36x |
| Base64 | compiled | 2.7167 | Go 2.2000 | 1.23x |
| Base64 | bytecode | 3.3200 | Ruby 2.2100; Python 3.3100 | 1.50x; 1.00x |
| JSON | compiled | 0.7267 | Go 1.3600 | 0.53x |
| JSON | bytecode | 0.8467 | Ruby 1.5600; Python 2.8700 | 0.54x; 0.30x |

A ratio above approximately `1.053x` misses the stated 95%-of-reference-speed
floor. I-Before-E misses its compiler and bytecode targets decisively. Base64
misses the compiled-Go and bytecode-Ruby comparisons but reaches the Python
floor. JSON clears every relevant comparison.

## Decision

Keep no code. The results do not identify a shared concrete boundary: the
I-Before-E miss is a direct file/text scan, Base64 is a reusable codec/byte
transfer path, and JSON is already ahead through a different host boundary.
Neither a text-shaped lowering rule nor a codec-specific VM shortcut would be
supported by this scorecard. No `able-stdlib` source changed.

## Next recommendation

Run the next bounded external shard for `matrixmultiply`, `monte_carlo_pi`,
and `mandelbrot` with the same compiled/bytecode, three-run, pinned settings.
Those kernels exercise arrays, scalar numeric loops, and float-heavy control
flow independently. Their results can distinguish a cross-kernel compiler or
VM wall from a text/byte boundary; only then profile a repeated miss alongside
an unrelated guard before considering a generic code change.
