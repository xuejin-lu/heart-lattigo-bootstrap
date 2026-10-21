# FIXED-SIGMA-CHAIN-002 — two-step noise compensation report

**Status: IN_PROGRESS_CP_C.** CP-A passed; CP-B has one provenance-clean matched pair and passed its Level/Scale/oracle gates. The fixed-σ CP-C chain has not yet run.

## Stagewise results

| Stage / metric | Genuine Standard | Original Fast (`9ef0315`) | Fast + fixed σ (`b5c55e0`) |
|---|---|---|---|
| Stage 1: Level5→4 incremental RMS / variance / max | `1.03186208e-11` / `1.06400717e-22` / `4.23940370e-11` | `7.43727623e-13` / `5.53130768e-25` / `2.20541281e-12` | NOT_RUN |
| Stage 1: decoded total-oracle RMS | `1.03216615e-11` | `7.44206624e-13` | NOT_RUN |
| Stage 2: Level4→3 incremental RMS / variance / max | `1.03435118e-11` / `1.06978460e-22` / `5.91369577e-11` | `7.39085148e-13` / `5.46152329e-25` / `2.13308582e-12` | NOT_RUN |
| Stage 2: propagated total-oracle RMS | `1.03649044e-11` | `7.39395197e-13` | NOT_RUN |
| Logical complete ct×ct operations | 2 | 2 | NOT_RUN (target 2) |
| Noise injections | 0 | 0 | NOT_RUN (target exactly 2) |
| Frozen coefficient σ | N/A | N/A | `4.004280012097432` (not run) |
| Helper overhead | N/A | N/A | NOT_RUN |

CP-B evidence: [compact matched pair](cp-b-pair-01.json), SHA-256 `95a858ad8c893242ffcb1b18c19e0a98f078601b4f1272231ada7d62f7f970cb`. The three-pair CP-B attempt allowance is exhausted; pairs 02/03 are excluded because their runtime Primary commit did not match their binaries' embedded build revision. See `journal.md` for hashes and conditions. Bootstrap calls remain 0; no LogN16 or large benchmark was run.
