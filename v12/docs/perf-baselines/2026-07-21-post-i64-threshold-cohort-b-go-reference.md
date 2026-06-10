# Pinned Go Reference Refresh

- Generated: `2026-07-21T12:08:31Z`
- Go: `go version go1.26.4 linux/amd64`
- Runs: `5`
- Timeout: `55s`
- CPU pool: `0-15`
- GOMAXPROCS: `catalog CPU budget per row`

| Benchmark | Go status | Validation | Go real (s) | Source SHA-256 | Stdout SHA-256 |
| --- | --- | --- | ---: | --- | --- |
| `fib` | 5/5 (timeouts 0, failures 0) | verified | 3.2893 | 79a2f518ad5cefdb74f7fe7fffd343031b162fbc790cf1ed2ab98dbfc1bd088e | 231e154363ef69683f6dca862f5a448e786ccc90485636bc20a86e2e470a2188 |
| `base64` | 5/5 (timeouts 0, failures 0) | verified | 2.5534 | 61a41a0ec45d3b3a8c890853d0a7839ce371b758ded8d8b8d0129a5b28390af6 | 5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316 |
| `matrixmultiply` | 5/5 (timeouts 0, failures 0) | verified | 1.0581 | 4b77c1e4f0add1763c830c9da22a9d92c43d14e48432bab021e2146e3ccb1e42 | 0dfcf69f5c73589f22465d7054ec20cd1aa43a7a1829c57673b147a49290fc13 |
| `monte_carlo_pi` | 5/5 (timeouts 0, failures 0) | verified-nondeterministic | 0.2839 | 34b6ad655126f97e453b17350d49552144c2bcd332fb3b3ca7192382554e3877 | 1b39386d0ed54ac136d4731c0262fab4ee41af1ba773082fb1e71c91c1ca4207,511279c9b6f954c3c5bdd55fa8e5e1835fe034cce0523743188ba722bedcb2f4,7d421f091f9ceac316835defc9a9fe9e0a96ae8d8a836a3463ec9acb633dc7e9,c4758829aa38cd029d849a51ac405a82faaa3c16fb4cbbf29bedf385ce7595ec,e28065b2b49c98eda257863e1ab4a636bab9938b698abec10a575f6f5bdfdb22 |
