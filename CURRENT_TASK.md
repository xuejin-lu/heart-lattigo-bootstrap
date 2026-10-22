# CURRENT TASK — FIXED-SIGMA-DIRECT-MUL-004

Task: `FIXED-SIGMA-DIRECT-MUL-004`
Status: `RMS_SCREEN_PASS_DECOMPOSITION_ROUNDOFF_REVIEW_REQUIRED` — the authorized 16/16 Fast and 2/2 Standard chains remain captured and validated. The original comparator and its fail-closed `gamma_64` result are preserved. An authorized offline high-precision review finds the exact identity closes for all slots and an operation-derived binary64 roundoff envelope covers all observed closures; final Web acceptance is pending. No new CKKS computation is authorized. See the task summary, original analysis, roundoff review, and journal.
**Spec:** `specs/FIXED-SIGMA-DIRECT-MUL-004.md`
**Journal:** `results/FIXED-SIGMA-DIRECT-MUL-004/journal.md`
**Results:** `results/FIXED-SIGMA-DIRECT-MUL-004/summary.md`

Prior 003 Web review: `results/FIXED-SIGMA-PROPAGATION-003/web-review.md`. Accepted as `ACCEPTED_SCOPED_PROPAGATION_ABLATION`, with `INSUFFICIENT_EVIDENCE` for general frozen-sigma transferability. Preserve all 001–003 historical results, accepted source SHA and budgets.

**New goal:** User approved replacing the historical Add-containing `(A+B)×C×C` test circuit with a **separate versioned** pure ciphertext×ciphertext direct-mul `A×B×C` circuit. The frontend must perform exactly two ordered `MulRelinNew→Rescale` operations, no `AddNew`. Stage1 encrypt A at `Δ=2^45`, B at effective `q5`, output Level4/Δ. Stage2 encrypt independent deterministic C at effective `q4`, output Level3/Δ. Introduce own workload hash and capacity plan; don't mutate legacy fixture.

**Fixed coefficient sigma:** `4.004280012097432`, unchanged and unretuned; manual Fast-only opt-in c0 after each completed operation; retain `OFF/OFF`, `ON/OFF`, `OFF/ON`, `ON/ON` ablation on two preregistered variants `direct-base` and `direct-c4`. Stagewise incremental errors, total oracle errors, and propagated first-stage injection have separate metrics. Do not mistake RMS match for distribution equivalence.

**Budget:** genuine Standard at most 2 **two-stage chains**; Fast at most 16 **two-stage chains**, including failed attempts; Bootstrap **0** Standard/Fast, no LogN16 workload, no BTS, no large benchmark. Strict centered Q-prefix capacity, source/binary provenance, honest STOP gates. Two repos: Primary `heart-lattigo-bootstrap/main`, Secondary `lattigo/fast-qprefix`; genuine Standard original `5dbffbdea05394de2ca3a432ed5318aa832e3f40`. Secondary initially `1e7828293c36f9845e30723ba4e9500d84bb50d2`.

Codex: mandatory safe branch/status/fetch/path check and conditional ff-only, read `AGENTS.md`, `docs/MEASUREMENT_PLATFORM.md`, scientific workflow, task spec; preserve existing local work, reuse mature platform, read file sizes before large logs. Follow CP-A→CP-D, journal every checkpoint; no task diversion to old A=0 or obsolete standard fixture. Complete and normal push with `BATCH_COMPLETE_READY_FOR_WEB_REVIEW`, or stop `BATCH_BLOCKED_NEEDS_WEB_REVIEW` for a new math/representation decision.
