# FAST-DROPIN-KEYLESS-MULRELIN-006

**Status:** READY_FOR_CODEX
**Task class:** M-approved limited algebra / I implementation / bounded E verification.
**Evidence & authority:** `results/FAST-DROPIN-CKKS-PRIMITIVE-API-AUDIT-005-web-review.md`; `docs/FAST_DROPIN_ZERO_SECRET_GOAL.md`; `docs/RESEARCH_ENGINEERING_WORKFLOW.md`.
**Previous task:** `FAST-DROPIN-CKKS-PRIMITIVE-API-AUDIT-005`, accepted strictly as `PARTIAL`.
**Pinned Standard:** unmodified `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
**Secondary Fast starting point:** `eb7793f1e449702590cdba3d9603261623354f80`.

## Purpose and frozen mathematical invariant

Implement **one real Fast-optimized primitive behind the unchanged Standard public CKKS evaluator type**. A caller using only `ckks.NewEvaluator(params, evk)` and `MulRelin` or `MulRelinNew` should obtain keyless, zero-secret multiplication **in the Fast library**, while the identical source links to genuine Standard and uses its normal Relinearization KeySwitch.

For both freshly created degree-one zero-secret CKKS ciphertexts `(a0, 0)`, `(b0, 0)` in valid NTT/Montgomery representations:

```text
(a0,0) * (b0,0) = (a0*b0, 0, 0).
Relin under effective s=0 -> (a0*b0, 0), without reading ANY RelinearizationKey or native GadgetProduct.
Scale_out = Scale_a * Scale_b, Level_out = min(Level_a,Level_b, destination's supported level).
```

- **Must not** silently clear `c1` of a native nonzero-c1 ciphertext, or use the result as a valid zero-secret representation.
- **Must preserve every logically active Q limb** for this first pilot. Do **not** produce an output with stale/dormant higher Q residues while leaving Level unchanged. Do not silently materialize or fall back from compact Q-prefix inputs that lack complete active rows.
- Must retain degrees, exact scale semantics, NTT and Montgomery flags, slot dimensions, and all required metadata; allow valid in-place aliasing or explicitly reject unsupported aliasing before any mutation, consistent with the ordinary API's established behavior.
- Preserve all public CKKS constructor and method signatures. `ckks.NewEvaluator` returns `*ckks.Evaluator`, not `*fast.Evaluator`. Existing `schemes/ckks/fast` already imports `schemes/ckks`; do not create an import cycle or replace the Go concrete return type.
- Detect the Fast CKKS simulation branch through an internal source-proven capability as already employed by `rlwe.NewEncryptor` for `ckks.Parameters`. Do not activate the path for ordinary RLWE, BFV, or BGV; do not change Standard pinned sources. No new frontend flags, special constructors, env variables or parameter substitutions.
- Preserve all existing `KeyLayoutFast` rejection guards in native KeySwitch. The new Fast keyless branch must **never call native GadgetProduct**, even when the ordinary frontend still supplies a normal Standard-layout eval key set for parity.

## Scope and steps

1. Safe-sync Primary main and Secondary fast-qprefix using updated `AGENTS.md` (one fetch, then ff-only merge), require clean worktrees, reread `CURRENT_TASK.md`, this spec, Secondary `docs/FAST_QPREFIX_SPEC.md`, source and tests. If Secondary HEAD has advanced beyond this task's known baseline for unrelated reasons, stop and request Web review.
2. Perform a **read-only** preflight of public `MulRelin` / `MulRelinNew` call signatures, zero-secret marker scope, ordinary metadata/degree handling, Q-level backing, aliasing and NTT/Montgomery semantics. Reuse existing ckks-internal arithmetic infrastructure where correct. If an import-cycle-free minimal branch with full active Q authority is mathematically or structurally incompatible, STOP `NEEDS_WEB_REVIEW` without an improvised broad architecture change.
3. Change **only Fast Secondary source** and focused tests to add a narrowly gated keyless MulRelin route in the public `*ckks.Evaluator` implementation (a logically equivalent minimal mechanism with the exact same public API is acceptable if source-proven). Do not modify explicit `schemes/ckks/fast` Q-prefix arithmetic, Bootstrap, key-generator layout or other native primitive paths unless a mechanically necessary and within-scope correction is documented. Do not globally disable normal KeySwitch or change unrelated scheme behavior.
4. Tests for Fast using ordinary public CKKS APIs:
   - Fresh independent degree-one secret-key `EncryptNew` inputs with c1=0; c0 product and output c1=0 at two distinct valid Levels with full Q active; output degree=1.
   - `MulRelin` and `MulRelinNew`, distinct-input multiplication and squaring, in-place/output aliasing where supported, both recognized NTT/Montgomery forms or explicit safe rejection of unsupported form **before input mutation**, Level/Scale/dimensions preservation, ciphertext output original-input oracle after DecryptNew/Decode.
   - A no-key (or deliberately failing keyset mock) diagnostic demonstrates that Fast zero-secret `MulRelin` works without *accessing* a RelinearizationKey, and that native GadgetProduct is not used.
   - A nonzero-c1 negative test must fail clearly without modifying input/output; do not silently convert normally encrypted ciphertexts.
   - Tests for unsupported degree/level/key/layout and for ordinary native `KeyLayoutFast` rejection remain effective.
   - Targeted regressions `go test ./schemes/ckks ./schemes/ckks/fast ./core/rlwe` on the Fast branch; investigate regressions honestly and do not broaden source modifications.
5. **Same unchanged frontend**: reuse `tools/fast-dropin-ckks-primitive-api-audit-005/main.go` (same SHA-256 `62e1a069d2ce3c0106e34594cb799388d9d6fbebbb72dd55b0ef5989f02d224e`), without editing its source or config, build and run once against pinned Standard and new Fast dependency. Final fixed profile from audit: LogN13, Q four primes, P one prime, initial Level3, scale 2^45, LogSlots4, independent inputs. Read the **current actual source** and only accept this reuse if the first MulRelin result and all remaining native checkpoints are proven full-Q-authoritative. No bootstrap and no performance tests. Measure cleartext RMSE/max error and c1 invariant at each checkpoint. Standard's native Relinearization KeySwitch and result must remain unchanged.
6. Because subsequent public Rotate in this frontend still uses ordinary native keys, **do not claim the Rotate/Rescale optimized path was dispatched**. A surviving c1=0 with full Q-authority permits the pre-existing native path in this *fully materialized profile*; no disguised native fallback is allowed on compact Q-prefix ciphertexts. If subsequent semantics are uncertain, stop and report the safe subset rather than forcing the full pipeline.
7. Preserve exact input/profile/backend commits/hashes and execution count. No Bootstrap calls, timing benchmark, LogN16, parameter tuning, full CNN or unbounded test run.

## Results / handoff

- Secondary production/test commits, focused tests and safe ff-only push if authorized.
- A new concise Primary report `results/FAST-DROPIN-KEYLESS-MULRELIN-006-summary.md` with explicit proof of optimized branch selection (key-access mock or equivalent), source map, both backend SHA/identical frontend hash, metrics, c1/Level/Scale evidence, tests, and isolated limitations. Do not rewrite predecessor 005 report; do not commit keys or bulky console output.
- `git diff --check` in affected repos; one bounded Codex self-review/repair pass, no unrelated refactor. If unexpectedly failing baseline tests require architecture choice, classify BLOCKED and return for Web review.
- Classifications: `FAST_DROPIN_KEYLESS_MULRELIN_COMPLETE_PENDING_WEB_REVIEW`, `FAST_DROPIN_KEYLESS_MULRELIN_PARTIAL`, or `FAST_DROPIN_KEYLESS_MULRELIN_BLOCKED`; follow with `READY_FOR_WEB_REVIEW` or `NEEDS_WEB_REVIEW`.
- Do not claim general Fast evaluator transparency, complete CNN compatibility, cryptographic security or speedup. `FAST-STANDARD-PERF-REBASELINE-003` stays BLOCKED.
