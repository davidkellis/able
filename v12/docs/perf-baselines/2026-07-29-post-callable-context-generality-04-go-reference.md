# Pinned Go Reference Refresh

- Generated: `2026-07-29T12:00:06Z`
- Go: `go version go1.26.5 linux/amd64`
- Go toolchain selector: `go1.26.5`
- Runs: `5`
- Timeout: `90s`
- CPU pool: `7-10`
- GOMAXPROCS: `catalog CPU budget per row`

| Benchmark | Go status | Validation | Go real (s) | Source SHA-256 | Stdout SHA-256 |
| --- | --- | --- | ---: | --- | --- |
| `monte_carlo_pi` | 5/5 (timeouts 0, failures 0) | verified-nondeterministic | 0.2858 | 34b6ad655126f97e453b17350d49552144c2bcd332fb3b3ca7192382554e3877 | 00d8d015e7f048021cf9e6451ce7cd1e35bad5debaf590f988726daafe379d2d,35bd496503d7ec80e36be2e63a8c704714d9b96d971f1a142e7d7cb0c8e0f77e,6f25f91aa22ab8bd5b9838ed9d9cce6dbed21f2e3ad0254dfa7f95f6f3cd8b1a,788a85b12bafede5bf1b137c55362fa02ac6a7614ee223333f3df9509e0d2f87,9182b00eaa589854d78e1a2a86975dd046595d023550ef43aee5fed255eb4ce7 |
| `pidigits` | 5/5 (timeouts 0, failures 0) | verified | 1.4025 | c8669a71e52ce32ed4e6852547efec42f67e8b8f5656e88653720d74b37da58c | bdfa7b6c756d96492f472f97aee9cc139bee954d271eacedfd7ace5d2875f06c |
| `mandelbrot` | 5/5 (timeouts 0, failures 0) | verified | 0.0549 | c0a81d428de3e5b86ec9980b514441c8388dadfc550ed345a1f6271dcd8f0b4c | d1560cadc49ab9dca30a4cfbe4123c1fcc0240ed80de27ad37d7fcc97159173e |
