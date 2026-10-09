# Current Task

Task: FAST-DROPIN-PUBLIC-ROTATE-BOOTSTRAP-COMPOSITION-010
Status: READY_FOR_CODEX
Class: bounded E integration verification; no Secondary source changes

**Executable spec:** `specs/FAST-DROPIN-PUBLIC-ROTATE-BOOTSTRAP-COMPOSITION-010.md`
**Accepted 009 review:** `results/FAST-DROPIN-SHARED-AUTOMORPHISM-BRIDGE-009-web-review.md`
**Long-term goal:** `docs/FAST_DROPIN_ZERO_SECRET_GOAL.md`

## Accepted current architecture

Review of `FAST-DROPIN-SHARED-AUTOMORPHISM-BRIDGE-009` ACCEPTED. Fast Secondary `00ac70ba136d190fa31bbb26c2f51d003a221634` extracted the existing automorphism kernel to `schemes/ckks/internal/fastcore`; ordinary public CKKS Rotate (P0 full-Q zero-secret) and explicit Fast Rotate (C0 compact-Q) invoke the same moved core while keeping distinct row contracts. Public Fast Rotate is keyless, fully handles all Q rows at its logical Level and fails closed for nonzero c1, invalid full backing or domains. The fixed frontend public Add→MulRelin→Rescale→Rotate numerical run passed, with Fast Rotate RMSE `5.508361443991879e-15`. That result was run on a dirty working tree with the subsequently committed same production source; final focused/package tests were rerun. There was no Bootstrap or benchmark.

## Current bounded task

Test **composition, not another algorithm implementation**: same ordinary CKKS frontend source compiled against genuine pinned Standard and pinned Fast, fixed canonical LogN13 E32 profile, public KeyGen/EncryptNew → public `ckks.Evaluator.RotateNew` → public `bootstrapping.NewEvaluator/B​​ootstrap` → public DecryptNew/Decode. Run inexpensive separated preflights and check genuine slot oracle and source constraints before **at most one Bootstrap per backend**. Distinct public/explicit representation contracts remain: P0 full-Q is source-proven; never send compact/stale C0 output to public native full-Q operations. Do not change Secondary, existing 004/009 sources, frontend parameters, or old Standard baseline. No benchmark. Report precise parameters, Q/P hashes, full-Q/zero-c1 and numeric/error provenance and failures. See spec for stop rules and artifact details.

Fast Secondary baseline `00ac70ba136d190fa31bbb26c2f51d003a221634`; pinned unchanged Standard `5dbffbdea05394de2ca3a432ed5318aa832e3f40`. Former 007 remains canceled. `FAST-STANDARD-PERF-REBASELINE-003` remains BLOCKED.

Return `FAST_DROPIN_COMPOSITION_COMPLETE_PENDING_WEB_REVIEW`, `FAST_DROPIN_COMPOSITION_PARTIAL`, or `FAST_DROPIN_COMPOSITION_BLOCKED`, plus `READY_FOR_WEB_REVIEW` / `NEEDS_WEB_REVIEW`.
