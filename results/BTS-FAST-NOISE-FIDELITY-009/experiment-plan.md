# BTS-FAST-NOISE-FIDELITY-009 execution plan

Status: `ZERO_BOOTSTRAP_PREFLIGHT_PENDING`

## Frozen source and profile

- Primary measurement implementation commits: `369ce3a84eb072392bacf2359d651413c159e5e8` and zero-call Hook-state gate `42b9840a6ea42bb6d9f4ba368b7c161617d4a02d`. The build will record the exact clean Primary VCS revision after the refreshed plan checkpoint is committed; its linked measurement-source SHA-256 values must match these implementation commits, and no Go source may change after building.
- Genuine Standard Lattigo: `5dbffbdea05394de2ca3a432ed5318aa832e3f40` (clean detached checkout).
- Fast Lattigo: `experiment/fast-bootstrap-noise-fidelity@be9c42ff23662300769b77f1eead3f12fc53adef` (clean, pushed dedicated worktree).
- Original BTS manifest SHA-256: `53205b37abd1fb57a960394256b958f5d5e5b1a66f527bfb7cbadabfe3cf0cda`.
- Effective profile is frozen to the BTS-008 Q30/P5, LogN16/E32 source parameters, input L3/Scale 2^32, Mod1 L26/Scale 2^42, output L18/Scale 2^32, `BTSModePlain`, BTS seed 0, amplitude 1, H=32, σ `4.004280012097432`, Noise root seed `20261027`, and the existing deterministic input SHA. No profile regeneration or retuning is allowed.

## Gates before a Bootstrap reservation

1. `cmd/perfprobe bts-build --task BTS-FAST-NOISE-FIDELITY-009` builds a Standard binary from the clean pinned Standard checkout and one Fast binary from the clean pinned Fast checkout. Build manifests must prove Primary VCS revision, `vcs.modified=false`, backend source revision/root, original BTS manifest, measurement adapter hashes and binary SHA-256. The same Fast binary is used for Fast OFF and Fast+Noise.
2. Run exactly three `BTS_PREFLIGHT_ONLY=1` invocations into fresh task-owned directories: Standard/OFF, Fast/OFF, Fast/ON. These construct BTS/evaluator state, generate/encrypt the same frozen input, and record the effective profile, but reserve and execute zero Bootstrap calls.
3. Require each preflight status `ZERO_BOOTSTRAP_PREFLIGHT_OK`, bootstrap reservations `0`, matching input SHA and effective Q/P/profile, correct source and binary identities, and Fast Hook disabled/enabled with the frozen σ/root seed as requested. Any mismatch stops before CP-C.
4. Version and push the preflight manifests/summary and updated ledger with counters still `0/0/0` before starting a formal arm.

## Formal arm order and immutable budget

| Order | Arm | Attempt IDs | Calls | Binary |
|---:|---|---|---:|---|
| 1 | Fast+Noise ON | `FN-01` cold, `FN-02`–`FN-06` warm 01–05 | 6 | shared Fast binary |
| 2 | Genuine Standard | `S-01` cold, `S-02`–`S-06` warm 01–05 | 6 | pinned Standard binary |
| 3 | Fast Noise OFF | `F-01` cold, `F-02`–`F-06` warm 01–05 | 6 | same shared Fast binary |

Each arm is one process and one evaluator/BTS lifecycle. It constructs the same deterministic original plaintext and encrypts it through that backend's native CKKS lifecycle; Standard and Fast ciphertext coefficients/keys are not required to match. Each arm performs one first cold Bootstrap followed by five warm Bootstraps, each on a fresh copy of that arm's original encrypted input. There is no hidden warmup or extra Bootstrap. Each invocation reserves durably through the existing `BootstrapAttemptBudget` before entering the operation. A failed or interrupted call remains consumed; stop immediately with no retry and do not continue to another arm.

The task ledger is the authoritative cumulative counter (maximum 6 per arm, 18 total); historical 002, 003 and 008 captures are not reusable. Preserve per-call run metadata, attempt sidecars, decoded `f64le` vectors, Noise event evidence, output/evidence SHA-256 and the final comparator outputs under this task directory. Inspect artifact sizes before manual review; do not paste vectors or event arrays into summary prose.

## Analysis

Use the maintained `perfprobe bts-report` three-arm mode and shared `internal/perfmeasure` / `internal/numericalmetrics`. Validate the strict BTS source/profile/input/provenance/capacity and event evidence before computing metrics. Report cold separately from the five warm samples; show individual timings, allocations, RMSE, global `-log2(RMSE)` precision, per-slot median precision, MaxAbs, p99 and p99.9; report warm mean/median/sample standard deviation/min/max, speedup, and Fast+Noise time/allocation overhead. Compare all three arms against the same deterministic BTS plaintext, and label BTS-008 timing strictly as historical context.
