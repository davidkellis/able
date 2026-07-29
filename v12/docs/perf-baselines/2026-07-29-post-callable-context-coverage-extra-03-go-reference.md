# Pinned Go Reference Refresh

- Generated: `2026-07-29T12:15:40Z`
- Go: `go version go1.26.5 linux/amd64`
- Go toolchain selector: `go1.26.5`
- Runs: `5`
- Timeout: `90s`
- CPU pool: `7-10`
- GOMAXPROCS: `catalog CPU budget per row`

| Benchmark | Go status | Validation | Go real (s) | Source SHA-256 | Stdout SHA-256 |
| --- | --- | --- | ---: | --- | --- |
| `regex_set_audit` | 5/5 (timeouts 0, failures 0) | verified | 0.0082 | ce6d01f4d9331f3ee52ff73bcca39637e226c14f3c045942f5a4765e0c3995ed | 3d8f861a312416f95b95d59be62614f0ffc7918e86fb984fe6035ca7b7b28da2 |
| `regex_stream_audit` | 5/5 (timeouts 0, failures 0) | verified | 0.0064 | 1f45fe2a76d146be18659ac9dd6eeab88548c5afe22e2a3ccc5ac3c8c0d82be9 | dd7801f0b104d8bd47aaef64d685bc65a06263851f12c8df8cf75260d09e717b |
| `log_routing_redaction` | 5/5 (timeouts 0, failures 0) | verified | 0.0058 | 4b9ee58bfba7e36b5c2bd530cc0b3977eb5a8ce22e14c1d5efaa656e224ce855 | 0d9585b01f83904fdf11d47b2902678c1718c8442ed1d84410d61d5d90f60bf4 |
| `array_slice_window` | 5/5 (timeouts 0, failures 0) | verified | 0.0062 | b05ef51494077012bf4dd3822d8b36edaf10331d7de66ea6532a02f86e5d0402 | 155f89122475c7b282637dbf2ecba6d19771d396e801b581cb1d1b0cef64103e |
