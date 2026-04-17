# Benchmark Suite Report

- Generated: `2026-04-16T15:53:38Z`
- Suite: `bytecode-core`
- Git: `c10f85fbad8cc24e326495ac7ee5ef925a78245f` (dirty)
- Machine: `Linux 6.19.11-200.fc43.x86_64 x86_64` on `davidlinux`
- CPU: `11th Gen Intel(R) Core(TM) i7-11800H @ 2.30GHz` (`16` logical cores)
- Go: `go version go1.26.1 linux/amd64`
- Runs per row: `1`
- Timeout: `90s`
- Build timeout: `300s`

| Benchmark | Mode | OK/Runs | Timeouts | Errors | Real Avg (s) | GC Avg |
| --- | --- | --- | --- | --- | --- | --- |
| `quicksort` | `compiled` | `1/1` | `0` | `0` | `0.2000` | `10.00` |
| `quicksort` | `treewalker` | `1/1` | `0` | `0` | `2.8400` | `79.00` |
| `quicksort` | `bytecode` | `1/1` | `0` | `0` | `3.9800` | `34.00` |
| `future_yield_i32_small` | `compiled` | `1/1` | `0` | `0` | `0.0100` | `1.00` |
| `future_yield_i32_small` | `treewalker` | `1/1` | `0` | `0` | `0.0600` | `4.00` |
| `future_yield_i32_small` | `bytecode` | `1/1` | `0` | `0` | `0.0500` | `3.00` |
| `sum_u32_small` | `compiled` | `1/1` | `0` | `0` | `0.0200` | `2.00` |
| `sum_u32_small` | `treewalker` | `1/1` | `0` | `0` | `34.0100` | `130.00` |
| `sum_u32_small` | `bytecode` | `1/1` | `0` | `0` | `11.2800` | `71.00` |
