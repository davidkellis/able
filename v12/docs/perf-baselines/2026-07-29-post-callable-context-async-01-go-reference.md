# Pinned Go Reference Refresh

- Generated: `2026-07-29T12:06:46Z`
- Go: `go version go1.26.5 linux/amd64`
- Go toolchain selector: `go1.26.5`
- Runs: `5`
- Timeout: `90s`
- CPU pool: `7-10`
- GOMAXPROCS: `catalog CPU budget per row`

| Benchmark | Go status | Validation | Go real (s) | Source SHA-256 | Stdout SHA-256 |
| --- | --- | --- | ---: | --- | --- |
| `channel_rollup` | 5/5 (timeouts 0, failures 0) | verified | 0.0073 | ce97e496db8d6e776af7e6714d2ff51c0c55c3e7fcf098fa0db48c616f060ef7 | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 |
| `future_pipeline` | 5/5 (timeouts 0, failures 0) | verified | 0.0068 | 6b8d6b83a6086d48e1c62ec05f192bb2bc7a488f8e3ac6143cbb4a6b5c3df73c | 3db937fdd21b5719ab57c25679a1944e8f30128695c8edbfae4940aba7aa6d98 |
| `future_await_race` | 5/5 (timeouts 0, failures 0) | verified | 0.0049 | 04d296c4183b7e00c93369218788cac51325c8ba3a3a2cb97b64c81a5c2ac94b | 33297798eeb96d6d471a6afe97ec8a8bf09eda0ce30e13917aed1db40fb930e4 |
