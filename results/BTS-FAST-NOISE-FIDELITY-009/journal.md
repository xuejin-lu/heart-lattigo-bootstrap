# BTS-FAST-NOISE-FIDELITY-009 journal

Status: `PLANNED_NOT_STARTED`

## Authorization and separation

- Authorized by the user-provided 009 task specification on 2026-10-27. This is a new, independent 18-call budget.
- Prior Primary pointer was `FAST-BOOTSTRAP-NOISE-FIDELITY-003-LOGN16`; its original 18/18 Bootstrap calls remain exhausted and untouched. Its spec permits a separately authorized new task and budget.
- BTS-008 remains historical: its Standard/Fast captures and results are not 009 samples.
- No 009 Bootstrap has been reserved or executed.

## Initial synchronized source state

- Primary `main`: `3df02885e9ab99f8760f2319a7577fba5f338ad6`, clean and equal to fetched `origin/main`.
- Genuine Standard pin: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
- Secondary `experiment/fast-bootstrap-noise-fidelity`: `325e501709c42ab5995bf7f36fb8d19621d34131`, clean and equal to fetched origin. `fast-qprefix` is not used or modified.
- Original `external/BTS` source manifest SHA-256 verified as `53205b37abd1fb57a960394256b958f5d5e5b1a66f527bfb7cbadabfe3cf0cda`; original BTS files are not modified.

## Frozen experiment profile

Use the exact BTS-008 effective Q/P primes and profile, LogN16 / E32, input L3 / Scale 2^32, Mod1 L26 / Scale 2^42, output L18 / Scale 2^32, `BTSModePlain`, seed 0, amplitude 1, σ `4.004280012097432`, root seed `20261027`. The authoritative full Q/P arrays are retained in the BTS-008 Standard/Fast run manifests; no regenerated profile is accepted.

## Planned gates

1. Minimal Secondary hook wiring into Fast public CKKS ciphertext×ciphertext `Mul` and `MulRelin` dispatch, default off; preserve wrappers and arithmetic semantics.
2. Targeted zero-Bootstrap tests at the BTS effective Q/P, Levels and Scales; N13–N16 regression tests.
3. Extend the existing `cmd/perfprobe` BTS adapter/report and shared `internal/perfmeasure` / `internal/numericalmetrics` owners only for evidenced capability gaps.
4. Pin final Fast source, verify provenance and shared Fast OFF/Noise binary, perform zero-call preflight, and commit/push the plan before any Bootstrap.
5. Only after all gates pass, run at most 6 calls per arm (18 total), no warmup or retry, recording each reservation and outcome.

No tests or Bootstrap calls were run during this initial checkpoint.
