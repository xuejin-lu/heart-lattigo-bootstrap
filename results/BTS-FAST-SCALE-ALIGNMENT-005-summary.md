# BTS-FAST-SCALE-ALIGNMENT-005 — source audit checkpoint

**Status: `NEEDS_WEB_REVIEW`**
**Batch state:** `BATCH_BLOCKED_NEEDS_WEB_REVIEW`
**Bootstrap / BTS stage calls:** 0 / 0

## Result

The requested T3 fast dispatch is not safe to implement under the current Q-prefix contract without a trusted bound source at the ordinary public API boundary. The accepted `T1_BOUND_PROVEN_OFFLINE` applies to the captured LogN16 seed-0 `BTSModePlain`, amp-1 chain only. No such bound or provenance reaches a general `ckks.Evaluator.Sub` call. I stopped before changing Secondary production or test code; no captured bound was hard-coded and no centered residue was treated as proof of historical no-wrap.

The four acceptance questions remain distinct: Standard's modular integer-alignment semantics were source-audited; an independent numeric oracle for the new different-Scale path was not run because no safe implementation exists yet; the existing equal-Scale regression proves only its existing Fast dispatch, not a different-Scale dispatch; and Q0123 capacity is proven only for the captured T3 chain, not for arbitrary public inputs.

## Source-backed reuse and dispatch audit

| Concern | Verified source behavior |
|---|---|
| Ordinary public dispatch | `schemes/ckks/evaluator.go:193-215` calls `fastCKKSZeroSecretAddSubEligible`; that predicate in `evaluator_fast_zero_secret_add.go:60-89` requires equal scales. Different-scale compact inputs fail eligibility and proceed only after full-active-row validation; absent q4+ therefore returns an error rather than reading dormant rows. |
| Existing Fast scalar kernel | `fastcore.ScalarWorkspace.MulRows` / `MulAtScaleRows` (`schemes/ckks/internal/fastcore/scalar.go:121-165`) uses scale 1 for integer scalars, multiplies exactly the requested Q-prefix rows, preserves the input Scale in output metadata, and has no capacity-bound input/check. Thus the caller could set the temp Scale only after the actual multiplication, but this does not establish a valid integer lift. |
| Existing Fast Add/Sub kernel | `fastcore.AddSubCiphertextsRows` (`schemes/ckks/internal/fastcore/addsub.go:103-185`) rejects unequal scales, validates prefix backing before writes, applies modular rows, preserves the higher Degree component for `opOut==op0`, and sets output metadata from op0. It can be reused after a legitimate alignment, but cannot supply the missing bound. |
| Genuine Standard semantics | Pinned unmodified Standard `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, `schemes/ckks/evaluator.go:221-394`: for `opOut==c0` and `Scale(c0)>Scale(c1)`, compute `ratioFlo=c0.Scale/c1.Scale`, truncate via `ratioFlo.Int(nil)`, multiply the lower-scale c1 by that integer into a temporary, evaluate common components, preserve c0's higher-degree component in-place, and set output Scale to `max(c0.Scale,c1.Scale)`. For the captured T3, k is `4398038122510`. |
| Bound availability | Ordinary `ckks.Evaluator` stores kernel workspaces but no coefficient-bound/provenance state (`schemes/ckks/evaluator.go:17-25`). The optional observer is only on explicit `schemes/ckks/fast.Evaluator`; `ObserveQPrefixCapacity` reconstructs the current centered prefix (`schemes/ckks/fast/qprefix_capacity.go:34-88`) and cannot prove prior no-wrap. |

## Accepted proof carried forward

The fixed-path values are documented in the Batch spec and backed by the immutable source artifacts listed there: `B_T1 <= 22208883997622`, `B_T2_pre <= 64649255451857143382418104254464`, `B_T2_post <= 14699510789662016947`, and `B_T3_preSub <= 42789729724024999377720936179651527668`. Exact C2S raw bounds are `1202504720161674545536166`, `18087138025836700338678205`, and `48686836780655299035616753`; the resulting stage post-bounds are `273418643178`, `4112528265648`, and `11070082260443`. They fit the existing Q0123 bound for the analyzed path. The original diagnostic artifact remains unchanged with its earlier `T1_SOURCE_EVIDENCE_MISSING` label; the later offline review corrected the proof using the exact q0-canonical ModUp lift, audited C2S L1 recurrence, and modular Level-1 Rescale argument. This acceptance does not generalize the captured bound to arbitrary ciphertexts.

## Verification performed

Commands were run in the clean independent Secondary clone at `d1b6dc9963b0df95a7c149306fb62660e1020f7b`. The default Go build-cache write was denied, so the successful reruns used a writable task-local cache under `/private/tmp`.

- PASS — `GOCACHE=/private/tmp/bts-fast-scale-align-005.47bVSR/gocache go test ./schemes/ckks -run '^TestPublicFastZeroSecretAddSubUsesSharedQPrefixCoreAndDecodes$' -count=1`
- PASS — `GOCACHE=/private/tmp/bts-fast-scale-align-005.47bVSR/gocache go test ./schemes/ckks/internal/fastcore -count=1`
- PASS — `GOCACHE=/private/tmp/bts-fast-scale-align-005.47bVSR/gocache go vet ./schemes/ckks ./schemes/ckks/internal/fastcore`
- No T3 regression was added or run because doing so would imply unsupported capacity semantics. No Bootstrap-capable suite, Bootstrap, BTS stage, benchmark, or Standard BTS was run.

## Repository state and decision requested

- Primary task bootstrap commit `04d5c198d8f3528b37d1ca13b13b5430b18cf114` was normally pushed before Secondary source work.
- Secondary remote base is verified at `d1b6dc9963b0df95a7c149306fb62660e1020f7b`; the independent clone is clean. No Secondary commit exists because no Secondary file was changed.
- The original Secondary checkout remains at `0b922dc48315128a27db811a5f237e0a6c9813d6` with the same five pre-existing dirty files.
- Batch journal remains at 0 Bootstrap, 0 Standard BTS, 0 S2C, 0 C2S, and 0 EvalMod calls.

Web decision needed: how should ordinary public `Add`/`Sub` obtain a trusted bound/certificate for compact scale alignment without (a) treating current centered residues as historical proof, (b) hard-coding the captured seed's bounds, or (c) adding an unapproved bound-propagation/representation contract? Until resolved, compact different-Scale operands must remain rejected; complete-Q inputs can use the existing full-Q route.
