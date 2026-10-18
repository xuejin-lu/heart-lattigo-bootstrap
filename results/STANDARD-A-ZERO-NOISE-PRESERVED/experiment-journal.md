# Standard A=0 experiment journal

- Primary SHA at experiment start: `b93fe439d59e95d6ef5d561d74d7fe247c4dfdfe`
- Secondary base SHA: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`
- Secondary experiment branch: `experiment/standard-a-zero-noise-preserved`
- Bootstrap budget: at most one complete Bootstrap at each of LogN13 and LogN16.
- Complete Bootstrap calls used: LogN13 `1`; LogN16 `0` (LogN16 has one durable reservation and has not started).
- LogN13 reservation 1: reserved at `2026-10-18T16:11:05Z`; consumed by exactly one successful Bootstrap call. The tool's attempt journal records the irrevocable reservation and returned call.
- LogN13: output `logn13-output.json`; config `configs/bootstrap_config.logN13.json` SHA-256 `919a2d9409b8ddeb720458d3112855f87d7a69e0cc39aad825b0ddf794769c98`; deterministic input SHA-256 `d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285`; pre-Bootstrap decoded input SHA-256 `f5e0a2149494defcf7f28ff1b56ae4adadb9a2e1edfe12e9615528132ca6b4f0`.
- LogN13 result: `SUCCEEDED`, `VALIDATED`, diagnostic-only, quality assessment `UNASSESSED`; output vs original input Complex RMSE `1.1903314766124451e-9`, Max Complex Error `5.804203475530148e-8`, SNR `144.7138388863248 dB`; output vs decoded pre-Bootstrap input Complex RMSE `1.1902077135565897e-9`, Max Complex Error `5.803616375993631e-8`, SNR `144.71474203659363 dB`.
- LogN13 provenance: Primary `7b7002c9b71a8807732a993d5aad7be34c2f4df6`; Secondary `d7cf81ebaf03b8c6a20d92845a4bb7967b29a2f4` on `experiment/standard-a-zero-noise-preserved`.
- LogN16 reservation 1: reserved at `2026-10-18T16:12:14Z`; actual calls `0` before invocation. This reservation is consumed even if the process fails or is interrupted. Available resources at reservation: 16 GiB RAM with 52% system-wide free memory; 13 GiB free disk on `/private/tmp`.
- LogN16 planned output: `logn16-output.json`; config `configs/bootstrap_config.logN16.json`; one Bootstrap call maximum.
- Formal performance measurements / warmups: not authorized and not run.
