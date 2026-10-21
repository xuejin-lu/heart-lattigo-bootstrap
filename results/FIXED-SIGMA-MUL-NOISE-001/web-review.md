# GPT Web independent review — FIXED-SIGMA-MUL-NOISE-001

Date: 2026-10-22 (Asia/Taipei)
Decision: **ACCEPTED_SCOPED_RMSE_ONLY**, not full scientific validation.
Reviewed Primary: `167bd0c9c1d61a6f2ff01e37f62565eb123de5dd`
Reviewed Secondary: `9ef0315feaf59dbee03bf0852b72c76fdeb6b800`
Genuine Standard pin: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.

## Accepted verified evidence

- GitHub branch heads were independently read; Primary advanced six commits, Secondary one. The Secondary delta adds an opt-in CKKS Noise helper, tests and a Q-prefix capacity checker; there is no Standard algorithm edit in this diff.
- CP-B calibration `sigma=4.004280012097432` from a single calibration pair; two further pairs were held out from **sigma fitting**. CP-C three single-operation tests had fixed sigma and documented distinct sampling seeds, source/binary SHA and provenance.
- CP-C paired decoded incremental RMS discrepancy relative to genuine Standard was +1.047%, -0.598%, +1.423%. Decoded injection-delta variance matched `N Var(eta) / Delta^2` within 0.019% at the tested output Scale; 8192 coefficient samples, consistent Q-prefix residues and c1/metadata invariants are tested.
- The Secondary source explicitly states this diagnostic helper **is not automatically called by ordinary CKKS operations**. It is called explicitly after one isolated `MulRelinNew -> Rescale` by Primary perfprobe; no Bootstrap-wide injection, generic per-call interception or automatic mul-trigger behavior has been demonstrated.
- Historical Standard A=0 branch is not used as the genuine Standard comparator. Explicit zero-Bootstrap evidence is present.

## Scientific limits and non-acceptance

1. **Only one workload/level/scale**: every CP-B/CP-C pair uses the same canonical workload SHA. Held out means random trials not used in fitting; **not unseen plaintext inputs, multiple levels or consecutive multiplication chains**. No general fixed-sigma applicability has been demonstrated.
2. RMS/variance calibration is **moment-matching**, not equality of sampled CKKS noise distributions or cryptographic error sources. Differences in paired variance signs are small and do not by themselves prove the Gaussian hypothesis false; the sample is insufficient to establish Gaussianity or independence. Cross-backend small spatial covariance cannot prove independence.
3. **Tail mismatch**: Standard per-vector maximum incremental residual ~4.54e-11–5.54e-11, Fast+Noise ~2.97e-11–3.37e-11. RMSE closeness does not imply tail/max agreement.
4. Single-run diagnostic injection timing ~1.305–2.716 ms against multiplication+rescale ~2.269–5.365 ms, or ~50.6–58.4% operation overhead in these observations. This is not a controlled performance benchmark or proof of optimized overhead; diagnostic NTT/capacity-check cost requires decomposition.
5. The original experimental goal “automatically inject on every complete ct×ct+Relin+Rescale” remains **UNIMPLEMENTED**. Current helper is manual and restricted to LogN13 Level4, Scale 2^45, Q-prefix rows 4. A lazy/deferred Bootstrap implementation need not have a one-to-one three-step boundary. Do not inject on arbitrary Rescale.
6. No repeated-ct multiplication or Bootstrap validation; PK unsupported in selected Fast zero-secret shape. Full Standard package tests have reported unrelated historical/modfile failures; targeted task tests passed. This review inspected source, concise recorded artifacts, and repository history, **not an independent local rebuild or all multi-megabyte vector contents**.
7. Committing many multi-megabyte JSON vectors is an artifact-size concern. Next phase should preserve existing committed evidence and avoid unnecessary duplication; keep future summaries bounded, sample raw data only as needed, and hash artifacts.

## Result classification and next boundary

**ACCEPTED_SCOPED_RMSE_ONLY** for one isolated operation / fixed profile and matching coefficient-to-slot gain. **NOT_ACCEPTED** as general Gaussian equivalent, automatic injection, multi-level behavior, performance advantage or Bootstrap result.

Next charter `specs/FIXED-SIGMA-CHAIN-002.md` authorizes only zero-Bootstrap cheap two-step multiplication/scale propagation and overhead attribution with one frozen sigma, gated by math/Q-prefix capacity; it does not authorize new security or universal per-operation interception semantics. Prior experiment remains immutable evidence.
