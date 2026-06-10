# Pinned Go Reference Refresh

- Generated: `2026-07-22T01:17:32Z`
- Go: `go version go1.26.4 linux/amd64`
- Runs: `5`
- Timeout: `55s`
- CPU pool: `0-15`
- GOMAXPROCS: `catalog CPU budget per row`

| Benchmark | Go status | Validation | Go real (s) | Source SHA-256 | Stdout SHA-256 |
| --- | --- | --- | ---: | --- | --- |
| `config_validation_extraction` | 5/5 (timeouts 0, failures 0) | verified | 0.0038 | dd0a7582e9ac0b83d571d57000fa30cf2c09fbed5f9bd0988283755aeef74f3e | c1aa99b9a13bb6e0c7731cb2aea77e300cd3cecc695df7fd4af90036939341d1 |
| `log_routing_redaction` | 5/5 (timeouts 0, failures 0) | verified | 0.0050 | 4b9ee58bfba7e36b5c2bd530cc0b3977eb5a8ce22e14c1d5efaa656e224ce855 | 0d9585b01f83904fdf11d47b2902678c1718c8442ed1d84410d61d5d90f60bf4 |
| `policy_record_dispatch` | 5/5 (timeouts 0, failures 0) | verified | 0.0059 | a31c758f4f52941e200b527621253bf7559e55b3da8561671aab5dab0c813ed5 | f15c66e53e4c89650ae12aff6c2f466f2cf0c7adb6ca3fe5be0127bfba217c05 |
| `regex_set_audit` | 5/5 (timeouts 0, failures 0) | verified | 0.0063 | ce6d01f4d9331f3ee52ff73bcca39637e226c14f3c045942f5a4765e0c3995ed | 3d8f861a312416f95b95d59be62614f0ffc7918e86fb984fe6035ca7b7b28da2 |
| `regex_stream_audit` | 5/5 (timeouts 0, failures 0) | verified | 0.0054 | 1f45fe2a76d146be18659ac9dd6eeab88548c5afe22e2a3ccc5ac3c8c0d82be9 | dd7801f0b104d8bd47aaef64d685bc65a06263851f12c8df8cf75260d09e717b |
| `regex_suffix_audit` | 5/5 (timeouts 0, failures 0) | verified | 0.0057 | 4424b20eadfabba04dc90132f1606bdcd33be424fbe37b99041a2d10b356ca98 | b5d5ccfabbfd4dc5952406cb1c42d62b807f75828661c4c3774b251abe38380f |
