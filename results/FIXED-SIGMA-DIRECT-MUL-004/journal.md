# FIXED-SIGMA-DIRECT-MUL-004 — durable experiment journal

**Status:** `PLANNED_NOT_RUN`. Established 2026-10-22 Asia/Taipei from user's approved direct-multiplication redesign; no experiments or source changes executed in the documentation commit.
**Prior:** `FIXED-SIGMA-PROPAGATION-003` results remain unchanged; Web review in `../FIXED-SIGMA-PROPAGATION-003/web-review.md`.
**Authoritative spec:** `../../specs/FIXED-SIGMA-DIRECT-MUL-004.md`.

## Frozen decisions

- Replace `(A+B)×C×C` only in a **new versioned fixture** with `A×B×C`; no `AddNew` in the new measured workload. Original 001–003 files/fixtures remain valid historical evidence.
- Two full ct×ct MulRelin→Rescale operations, Stage1 Level5→4, Stage2 Level4→3, both output Scale `2^45`. B Stage1 `q5` Scale; independently defined C Stage2 `q4` Scale, q values from runtime constructed Q chain.
- New deterministic C formula and `direct-base`/`direct-c4` workloads predefined in spec. Both require new SHA/fingerprint, preflight conservative and runtime strict Q-prefix capacity.
- Frozen signed coefficient `σ=4.004280012097432`, not refit to direct multiplication. Four arms OFF/OFF, ON/OFF, OFF/ON, ON/ON, distinct deterministic seeds per stage.
- Genuine original Standard pin `5dbffbdea05394de2ca3a432ed5318aa832e3f40` (read-only); Secondary active `fast-qprefix@1e7828293c36f9845e30723ba4e9500d84bb50d2`. Primary initial reviewed base `574a1aee5ce5b0e329096095de3eda1b2df7bd1b`.

## Bounded counters (persist across sessions)

| Metric | Used | Maximum |
|---|---:|---:|
| Fast two-stage chain runs | 0 | 16 |
| Genuine Standard two-stage chains | 0 | 2 |
| Standard Bootstrap calls | 0 | 0 |
| Fast Bootstrap calls | 0 | 0 |
| LogN16 workload runs | 0 | 0 |
| BTS | 0 | 0 |

| Checkpoint | Status |
|---|---|
| CP-A new-workload source/math/capacity gates | NOT_STARTED |
| CP-B direct A×B→×C baseline | NOT_STARTED |
| CP-C fixed σ four-arm experiment | NOT_STARTED |
| CP-D full analysis/provenance/commit | NOT_STARTED |

All numerical metrics, exact source hashes, binaries, attempts, failed gates and original source build evidence are **NOT_RUN**. Codex must record them instead of guessing. Avoid broad `go test` selecting unrelated LogN16 parameter construction; test exact focused Go patterns. No Bootstrap calls are authorized.
