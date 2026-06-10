# Benchmark Suite Report

- Generated: `2026-07-11T09:29:53Z`
- Suite: `fixture-generality`
- Git: `237406eccdfb025a519d898daedadee1c8d13a7b` (dirty)
- Machine: `Linux 7.1.3-200.fc44.x86_64 x86_64` on `davidlinux`
- CPU: `11th Gen Intel(R) Core(TM) i7-11800H @ 2.30GHz` (`16` logical cores)
- Go: `go version go1.26.4 linux/amd64`
- Runs per row: `2`
- Timeout: `90s`
- Build timeout: `240s`

- Experimental execution context: `false` (compiled mode only)

| Benchmark | Mode | OK/Runs | Timeouts | Errors | Real Avg (s) | GC Avg |
| --- | --- | --- | --- | --- | --- | --- |
| `array_map_i32_small` | `compiled` | `2/2` | `0` | `0` | `0.0400` | `6.00` |
| `channel_roundtrip_i32_small` | `compiled` | `2/2` | `0` | `0` | `3.4000` | `10.00` |
| `deque_i32_small` | `compiled` | `2/2` | `0` | `0` | `0.0400` | `6.00` |
| `dijkstra_heap_small` | `compiled` | `2/2` | `0` | `0` | `0.0400` | `6.00` |
| `future_fanout_i32_small` | `compiled` | `2/2` | `0` | `0` | `0.0500` | `6.00` |
| `heap_i32_small` | `compiled` | `2/2` | `0` | `0` | `0.0750` | `6.00` |
| `nbody_small` | `compiled` | `2/2` | `0` | `0` | `0.0400` | `6.00` |
| `persistent_set_i32_small` | `compiled` | `2/2` | `0` | `0` | `0.1150` | `11.00` |
| `persistent_sorted_set_i32_small` | `compiled` | `2/2` | `0` | `0` | `22.3500` | `1186.00` |
| `random_lcg_i64_small` | `compiled` | `2/2` | `0` | `0` | `0.0900` | `5.50` |
| `regex_is_match_small` | `compiled` | `2/2` | `0` | `0` | `0.2300` | `11.00` |
| `string_builder_small` | `compiled` | `2/2` | `0` | `0` | `0.1800` | `12.00` |
| `sum_u32_small` | `compiled` | `2/2` | `0` | `0` | `0.0600` | `6.00` |
| `word_count_small` | `compiled` | `2/2` | `0` | `0` | `0.0900` | `8.00` |
| `zigzag_char_small` | `compiled` | `2/2` | `0` | `0` | `0.1000` | `16.00` |
